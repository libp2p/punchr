use clap::Parser;
use futures::future::FutureExt;
use futures::stream::StreamExt;
use libp2p::core::multiaddr::{Multiaddr, Protocol};
use libp2p::core::upgrade;
use libp2p::dcutr;
use libp2p::dcutr::behaviour::UpgradeError;
use libp2p::dns::DnsConfig;
use libp2p::identify::{Identify, IdentifyConfig, IdentifyEvent};
use libp2p::noise;
use libp2p::ping::{Ping, PingConfig, PingEvent};
use libp2p::relay::v1::{new_transport_and_behaviour, Relay, RelayConfig};
use libp2p::swarm::dial_opts::DialOpts;
use libp2p::swarm::{ConnectionHandlerUpgrErr, Swarm, SwarmBuilder, SwarmEvent};
use libp2p::tcp::TcpConfig;
use libp2p::Transport;
use libp2p::{identity, NetworkBehaviour, PeerId};
use log::info;
use std::convert::TryInto;
use std::net::Ipv4Addr;
use std::time::{SystemTime, UNIX_EPOCH};

pub mod grpc {
    tonic::include_proto!("_");
}

fn agent_version() -> String {
    format!("punchr/rust-client/{}", env!("CARGO_PKG_VERSION"))
}

#[derive(Parser, Debug)]
struct Opt {
    #[clap(long)]
    url: String,

    /// Fixed value to generate deterministic peer id.
    #[clap(long)]
    secret_key_seed: u8,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let _ = env_logger::try_init();
    let opt = Opt::parse();

    let mut client =
        grpc::punchr_service_client::PunchrServiceClient::connect(opt.url.clone()).await?;

    info!("Connected to server {}.", opt.url);

    let local_key = generate_ed25519(opt.secret_key_seed);
    let local_peer_id = PeerId::from(local_key.public());
    info!("Local peer id: {:?}", local_peer_id);
    let mut swarm = build_swarm(local_key).await?;

    swarm.listen_on(
        Multiaddr::empty()
            .with("0.0.0.0".parse::<Ipv4Addr>()?.into())
            .with(Protocol::Tcp(0)),
    )?;

    // Wait to listen on all interfaces.
    let mut delay = futures_timer::Delay::new(std::time::Duration::from_secs(1)).fuse();
    loop {
        futures::select! {
            event = swarm.select_next_some() => {
                match event {
                    SwarmEvent::NewListenAddr { address, .. } => {
                        info!("Listening on {:?}", address);
                    }
                    event => panic!("{:?}", event),
                }
            }
            _ = delay => {
                // Likely listening on all interfaces now, thus continuing by breaking the loop.
                break;
            }
        }
    }

    let request = tonic::Request::new(grpc::RegisterRequest {
        client_id: local_peer_id.to_bytes(),
        agent_version: agent_version(),
        // TODO: Fill.
        protocols: vec![],
    });

    let _response = client.register(request).await?;

    let request = tonic::Request::new(grpc::GetAddrInfoRequest {
        host_id: local_peer_id.to_bytes(),
        // TODO
        all_host_ids: vec![local_peer_id.to_bytes()],
    });

    let response = client.get_addr_info(request).await?.into_inner();

    let remote_peer_id = PeerId::from_bytes(&response.remote_id)?;
    let remote_addrs = response
        .multi_addresses
        .into_iter()
        .map(Multiaddr::try_from)
        .collect::<Result<Vec<_>, libp2p::multiaddr::Error>>()?;

    info!(
        "Attempting to punch through to {:?} via {:?}.",
        remote_peer_id, remote_addrs
    );

    swarm.dial(
        DialOpts::peer_id(remote_peer_id)
            .addresses(remote_addrs.clone())
            .build(),
    )?;
    let request = drive_hole_punch(local_peer_id, remote_peer_id, remote_addrs, swarm).await;
    client
        .track_hole_punch(tonic::Request::new(request))
        .await?;

    Ok(())
}

async fn build_swarm(
    local_key: identity::Keypair,
) -> Result<Swarm<Behaviour>, Box<dyn std::error::Error>> {
    let local_peer_id = PeerId::from(local_key.public());
    let transport = DnsConfig::system(TcpConfig::new().port_reuse(true)).await?;
    let (relay_transport, relay_behaviour) =
        new_transport_and_behaviour(RelayConfig::default(), transport);

    let noise_keys = noise::Keypair::<noise::X25519Spec>::new()
        .into_authentic(&local_key)
        .expect("Signing libp2p-noise static DH keypair failed.");

    let transport = relay_transport
        .upgrade(upgrade::Version::V1)
        .authenticate(noise::NoiseConfig::xx(noise_keys).into_authenticated())
        // TODO: Consider supporting mplex.
        .multiplex(libp2p::yamux::YamuxConfig::default())
        .boxed();

    let identify_config = IdentifyConfig::new("/TODO/0.0.1".to_string(), local_key.public())
        .with_agent_version(agent_version());

    let behaviour = Behaviour {
        relay: relay_behaviour,
        ping: Ping::new(PingConfig::new()),
        identify: Identify::new(identify_config),
        dcutr: dcutr::behaviour::Behaviour::new(),
    };

    let swarm = SwarmBuilder::new(transport, behaviour, local_peer_id)
        .dial_concurrency_factor(10_u8.try_into()?)
        .build();
    Ok(swarm)
}

async fn drive_hole_punch(
    client_id: PeerId,
    remote_id: PeerId,
    remote_multi_addresses: Vec<Multiaddr>,
    mut swarm: libp2p::swarm::Swarm<Behaviour>,
) -> grpc::TrackHolePunchRequest {
    let mut track_request = grpc::TrackHolePunchRequest {
        client_id: client_id.into(),
        remote_id: remote_id.into(),
        remote_multi_addresses: remote_multi_addresses
            .into_iter()
            .map(|a| a.to_vec())
            .collect(),
        open_multi_addresses: Vec::new(),
        has_direct_conns: false,
        connect_started_at: 0,
        connect_ended_at: 0,
        hole_punch_attempts: Vec::new(),
        error: None,
        outcome: grpc::HolePunchOutcome::Unknown.into(),
        ended_at: 0,
    };
    let mut active_holepunch_attempt: Option<HolePunchAttemptState> = None;
    loop {
        match swarm.select_next_some().await {
            SwarmEvent::NewListenAddr { address, .. } => info!("Listening on {:?}", address),
            SwarmEvent::Behaviour(Event::Relay(event)) => info!("{:?}", event),
            SwarmEvent::Behaviour(Event::Dcutr(event)) => {
                info!("{:?}", event);

                match event {
                    dcutr::behaviour::Event::RemoteInitiatedDirectConnectionUpgrade { .. } => {
                        if let Some(attempt) = active_holepunch_attempt.as_mut() {
                            attempt.started_at = Some(unix_time_now());
                        }
                    }
                    dcutr::behaviour::Event::DirectConnectionUpgradeSucceeded { .. } => {
                        track_request.outcome = grpc::HolePunchOutcome::Success.into();
                        if let Some(attempt) = active_holepunch_attempt.take() {
                            let resolved =
                                attempt.resolve(grpc::HolePunchAttemptOutcome::Success, None);
                            track_request.hole_punch_attempts.push(resolved);
                        }
                        break;
                    }
                    dcutr::behaviour::Event::DirectConnectionUpgradeFailed { error, .. } => {
                        track_request.error = Some("none of the 3 attempts succeeded".into());
                        track_request.outcome = grpc::HolePunchOutcome::Failed.into();
                        if let Some(attempt) = active_holepunch_attempt.take() {
                            let (outcome, error) = match error {
                                UpgradeError::Dial => (
                                    grpc::HolePunchAttemptOutcome::Failed,
                                    "failed to establish a direct connection",
                                ),
                                UpgradeError::Handler(ConnectionHandlerUpgrErr::Timeout) => (
                                    grpc::HolePunchAttemptOutcome::Timeout,
                                    "hole-punch timed out",
                                ),
                                UpgradeError::Handler(ConnectionHandlerUpgrErr::Timer) => {
                                    (grpc::HolePunchAttemptOutcome::Timeout, "timer error")
                                }
                                UpgradeError::Handler(ConnectionHandlerUpgrErr::Upgrade(_)) => (
                                    grpc::HolePunchAttemptOutcome::ProtocolError,
                                    "protocol error",
                                ),
                            };
                            let resolved = attempt.resolve(outcome, Some(error.into()));
                            track_request.hole_punch_attempts.push(resolved);
                        }
                        break;
                    }
                    dcutr::behaviour::Event::InitiatedDirectConnectionUpgrade { .. } => {
                        unreachable!()
                    }
                }
            }
            SwarmEvent::Behaviour(Event::Identify(event)) => info!("{:?}", event),
            SwarmEvent::Behaviour(Event::Ping(_)) => {}
            SwarmEvent::ConnectionEstablished {
                peer_id, endpoint, ..
            } => {
                info!("Established connection to {:?} via {:?}", peer_id, endpoint);

                if peer_id.to_bytes() != track_request.remote_id {
                    continue;
                }

                track_request
                    .open_multi_addresses
                    .push(endpoint.get_remote_address().to_vec());

                if !endpoint.is_relayed() {
                    track_request.has_direct_conns = true;
                }

                if active_holepunch_attempt.is_none() && !endpoint.is_relayed() {
                    track_request.outcome = grpc::HolePunchOutcome::ConnectionReversed.into();
                    break;
                }

                // New hole-punch attempt will be started by the DCUtR protocol.
                active_holepunch_attempt = Some(HolePunchAttemptState {
                    opened_at: unix_time_now(),
                    started_at: None,
                });
            }
            SwarmEvent::OutgoingConnectionError { peer_id, error }
                if peer_id == Some(remote_id) =>
            {
                info!("Outgoing connection error to {:?}: {:?}", peer_id, error);

                if !swarm.is_connected(&remote_id) {
                    // Initial connection to the remote failed.
                    track_request.connect_started_at = unix_time_now();
                    track_request.connect_ended_at = unix_time_now();
                    track_request.error =
                        Some("Error connecting to remote peer via relay.".to_string());
                    track_request.outcome = grpc::HolePunchOutcome::NoConnection.into();
                    break;
                }

                if let Some(attempt) = active_holepunch_attempt.take() {
                    // Hole punch attempt failed. Up to 3 attempts may fail before the whole hole-punch
                    // is considered as failed.
                    let outcome = grpc::HolePunchAttemptOutcome::Failed;
                    let error = "failed to establish a direct connection";
                    let resolved = attempt.resolve(outcome, Some(error.into()));
                    track_request.hole_punch_attempts.push(resolved);
                }
            }
            SwarmEvent::ConnectionClosed {
                peer_id, endpoint, ..
            } if peer_id == remote_id => {
                let remote_addr = endpoint.get_remote_address().clone().to_vec();
                track_request
                    .open_multi_addresses
                    .retain(|a| a != &remote_addr);
            }
            _ => {}
        }
    }
    track_request.ended_at = unix_time_now();
    info!(
        "Finished whole hole punching process with outcome: {:?}, error: {:?}",
        track_request.outcome, track_request.error
    );
    track_request
}

fn unix_time_now() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis()
        .try_into()
        .unwrap()
}

struct HolePunchAttemptState {
    opened_at: u64,
    started_at: Option<u64>,
}

impl HolePunchAttemptState {
    fn resolve(
        self,
        outcome: grpc::HolePunchAttemptOutcome,
        error: Option<String>,
    ) -> grpc::HolePunchAttempt {
        let ended_at = unix_time_now();
        info!(
            "Finished hole punching attempt with outcome: {:?}, error: {:?}",
            outcome, error
        );
        grpc::HolePunchAttempt {
            opened_at: self.opened_at,
            started_at: self.started_at,
            ended_at,
            // TODO
            start_rtt: None,
            elapsed_time: (ended_at - self.started_at.unwrap_or(self.opened_at)) as f32,
            outcome: outcome.into(),
            error,
            direct_dial_error: None,
        }
    }
}

fn generate_ed25519(secret_key_seed: u8) -> identity::Keypair {
    let mut bytes = [0u8; 32];
    bytes[0] = secret_key_seed;

    let secret_key = identity::ed25519::SecretKey::from_bytes(&mut bytes)
        .expect("this returns `Err` only if the length is wrong; the length is correct; qed");
    identity::Keypair::Ed25519(secret_key.into())
}

#[derive(NetworkBehaviour)]
#[behaviour(out_event = "Event", event_process = false)]
struct Behaviour {
    relay: Relay,
    ping: Ping,
    identify: Identify,
    dcutr: dcutr::behaviour::Behaviour,
}

#[allow(clippy::large_enum_variant)]
#[derive(Debug)]
enum Event {
    Ping(PingEvent),
    Identify(IdentifyEvent),
    Relay(()),
    Dcutr(dcutr::behaviour::Event),
}

impl From<PingEvent> for Event {
    fn from(e: PingEvent) -> Self {
        Event::Ping(e)
    }
}

impl From<IdentifyEvent> for Event {
    fn from(e: IdentifyEvent) -> Self {
        Event::Identify(e)
    }
}

impl From<()> for Event {
    fn from(e: ()) -> Self {
        Event::Relay(e)
    }
}

impl From<dcutr::behaviour::Event> for Event {
    fn from(e: dcutr::behaviour::Event) -> Self {
        Event::Dcutr(e)
    }
}
