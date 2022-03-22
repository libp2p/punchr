use futures::executor::block_on;
use futures::future::FutureExt;
use futures::stream::StreamExt;
use libp2p::core::multiaddr::{Multiaddr, Protocol};
use libp2p::core::transport::OrTransport;
use libp2p::core::upgrade;
use libp2p::dcutr;
use libp2p::dns::DnsConfig;
use libp2p::identify::{Identify, IdentifyConfig, IdentifyEvent};
use libp2p::noise;
use libp2p::ping::{Ping, PingConfig, PingEvent};
use libp2p::relay::v2::client::{self, Client};
use libp2p::swarm::dial_opts::DialOpts;
use libp2p::swarm::{SwarmBuilder, SwarmEvent};
use libp2p::tcp::TcpConfig;
use libp2p::Transport;
use libp2p::{identity, NetworkBehaviour, PeerId};
use log::info;
use std::convert::TryInto;

use std::net::Ipv4Addr;

use std::time::{SystemTime, UNIX_EPOCH};
use structopt::StructOpt;

pub mod grpc {
    tonic::include_proto!("_");
}

#[derive(StructOpt)]
struct Opt {
    #[structopt(long)]
    url: String,

    /// Fixed value to generate deterministic peer id.
    #[structopt(long)]
    secret_key_seed: u8,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let _ = env_logger::try_init();
    let opt = Opt::from_args();

    let mut client =
        grpc::punchr_service_client::PunchrServiceClient::connect(opt.url.clone()).await?;

    info!("Connected to server {}.", opt.url);

    let local_key = generate_ed25519(opt.secret_key_seed);
    let local_peer_id = PeerId::from(local_key.public());
    info!("Local peer id: {:?}", local_peer_id);

    let (relay_transport, relay_client) = Client::new_transport_and_behaviour(local_peer_id);

    let noise_keys = noise::Keypair::<noise::X25519Spec>::new()
        .into_authentic(&local_key)
        .expect("Signing libp2p-noise static DH keypair failed.");

    let transport = OrTransport::new(
        relay_transport,
        block_on(DnsConfig::system(TcpConfig::new().port_reuse(true))).unwrap(),
    )
    .upgrade(upgrade::Version::V1)
    .authenticate(noise::NoiseConfig::xx(noise_keys).into_authenticated())
    // TODO: Consider supporting mplex.
    .multiplex(libp2p::yamux::YamuxConfig::default())
    .boxed();

    let identify_config = IdentifyConfig::new("/TODO/0.0.1".to_string(), local_key.public());

    let behaviour = Behaviour {
        relay_client,
        ping: Ping::new(PingConfig::new()),
        identify: Identify::new(identify_config.clone()),
        dcutr: dcutr::behaviour::Behaviour::new(),
    };

    let mut swarm = SwarmBuilder::new(transport, behaviour, local_peer_id)
        .dial_concurrency_factor(10_u8.try_into().unwrap())
        .build();

    swarm
        .listen_on(
            Multiaddr::empty()
                .with("0.0.0.0".parse::<Ipv4Addr>().unwrap().into())
                .with(Protocol::Tcp(0)),
        )
        .unwrap();

    // Wait to listen on all interfaces.
    block_on(async {
        let mut delay = futures_timer::Delay::new(std::time::Duration::from_secs(1)).fuse();
        loop {
            futures::select! {
                event = swarm.next() => {
                    match event.unwrap() {
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
    });

    let request = tonic::Request::new(grpc::RegisterRequest {
        client_id: local_peer_id.to_bytes(),
        agent_version: identify_config.agent_version,
        // TODO: Fill.
        protocols: vec![],
    });

    let _response = client.register(request).await?;

    let request = tonic::Request::new(grpc::GetAddrInfoRequest {
        client_id: local_peer_id.to_bytes(),
    });

    let response = client.get_addr_info(request).await?.into_inner();

    let remote_peer_id = PeerId::from_bytes(&response.remote_id)?;
    let remote_addrs = response
        .multi_addresses
        .into_iter()
        .map(|a| Multiaddr::try_from(a))
        .collect::<Result<Vec<_>, libp2p::multiaddr::Error>>()?;

    info!(
        "Attempting to punch through to {:?} via {:?}.",
        remote_peer_id, remote_addrs
    );

    let start_time = SystemTime::now();

    let result = if remote_addrs
        .iter()
        .all(|a| a.iter().any(|p| p == libp2p::multiaddr::Protocol::Quic))
    {
        Err(())
    } else {
        swarm.dial(
            DialOpts::peer_id(remote_peer_id)
                .addresses(remote_addrs.clone())
                .build(),
        )?;
        drive_hole_punch(swarm, remote_peer_id).await
    };

    let finish_time = SystemTime::now();
    let elapsed_time = (finish_time.duration_since(UNIX_EPOCH).unwrap()
        - start_time.duration_since(UNIX_EPOCH).unwrap())
    .as_secs_f32();
    let request = tonic::Request::new(grpc::TrackHolePunchRequest {
        client_id: local_peer_id.to_bytes(),
        remote_id: remote_peer_id.to_bytes(),
        success: result.is_ok(),
        started_at: start_time
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_millis()
            .try_into()
            .unwrap(),
        remote_multi_addresses: remote_addrs.into_iter().map(|a| a.to_vec()).collect(),
        // TODO: Set proper attempts
        attempts: 0,
        // TODO
        error: String::new(),
        // TODO
        direct_dial_error: String::new(),
        // TODO
        start_rtt: 0.0,
        elapsed_time,
        end_reason: grpc::HolePunchEndReason::Unknown.into(),
    });

    client.track_hole_punch(request).await?;

    Ok(())
}

async fn drive_hole_punch(
    mut swarm: libp2p::swarm::Swarm<Behaviour>,
    remote_peer_id: PeerId,
) -> Result<(), ()> {
    loop {
        match swarm.next().await.unwrap() {
            SwarmEvent::NewListenAddr { address, .. } => {
                info!("Listening on {:?}", address);
            }
            SwarmEvent::Behaviour(Event::Relay(event)) => {
                info!("{:?}", event);

                match event {
                    client::Event::ReservationReqAccepted {
                        relay_peer_id: _,
                        renewal: _,
                        limit: _,
                    } => unreachable!(),
                    client::Event::ReservationReqFailed {
                        relay_peer_id: _,
                        renewal: _,
                        error: _,
                    } => unreachable!(),
                    client::Event::OutboundCircuitEstablished {
                        relay_peer_id: _,
                        limit: _,
                    } => {}
                    client::Event::OutboundCircuitReqFailed {
                        relay_peer_id: _,
                        error: _,
                    } => break Err(()),
                    client::Event::InboundCircuitEstablished {
                        src_peer_id: _,
                        limit: _,
                    } => {
                        unreachable!()
                    }
                    client::Event::InboundCircuitReqFailed {
                        relay_peer_id: _,
                        error: _,
                    } => unreachable!(),
                    client::Event::InboundCircuitReqDenied { src_peer_id: _ } => unreachable!(),
                    client::Event::InboundCircuitReqDenyFailed {
                        src_peer_id: _,
                        error: _,
                    } => {
                        unreachable!()
                    }
                };
            }
            SwarmEvent::Behaviour(Event::Dcutr(event)) => {
                info!("{:?}", event);

                match event {
                    dcutr::behaviour::Event::InitiatedDirectConnectionUpgrade {
                        remote_peer_id: _,
                        local_relayed_addr: _,
                    } => unreachable!(),
                    dcutr::behaviour::Event::RemoteInitiatedDirectConnectionUpgrade {
                        remote_peer_id: _,
                        remote_relayed_addr: _,
                    } => {}
                    dcutr::behaviour::Event::DirectConnectionUpgradeSucceeded {
                        remote_peer_id: _,
                    } => break Ok(()),
                    dcutr::behaviour::Event::DirectConnectionUpgradeFailed {
                        remote_peer_id: _,
                        error: _,
                    } => todo!(),
                }
            }
            SwarmEvent::Behaviour(Event::Identify(event)) => {
                info!("{:?}", event)
            }
            SwarmEvent::Behaviour(Event::Ping(_)) => {}
            SwarmEvent::ConnectionEstablished {
                peer_id, endpoint, ..
            } => {
                info!("Established connection to {:?} via {:?}", peer_id, endpoint);

                if peer_id == remote_peer_id && !endpoint.is_relayed() {
                    break Ok(());
                }
            }
            SwarmEvent::OutgoingConnectionError { peer_id, error } => {
                info!("Outgoing connection error to {:?}: {:?}", peer_id, error);
            }
            _ => {}
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
    relay_client: Client,
    ping: Ping,
    identify: Identify,
    dcutr: dcutr::behaviour::Behaviour,
}

#[derive(Debug)]
enum Event {
    Ping(PingEvent),
    Identify(IdentifyEvent),
    Relay(client::Event),
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

impl From<client::Event> for Event {
    fn from(e: client::Event) -> Self {
        Event::Relay(e)
    }
}

impl From<dcutr::behaviour::Event> for Event {
    fn from(e: dcutr::behaviour::Event) -> Self {
        Event::Dcutr(e)
    }
}
