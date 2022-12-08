use clap::Parser;
use futures::future::FutureExt;
use futures::stream::StreamExt;
use libp2p::core::either::EitherOutput;
use libp2p::core::multiaddr::{Multiaddr, Protocol};
use libp2p::core::muxing::StreamMuxerBox;
use libp2p::core::transport::OrTransport;
use libp2p::core::{upgrade, ConnectedPoint, ProtocolName, UpgradeInfo};
use libp2p::dcutr::behaviour::UpgradeError;
use libp2p::dns::DnsConfig;
use libp2p::relay::v2::client::{self, Client};
use libp2p::swarm::dial_opts::DialOpts;
use libp2p::swarm::{
    ConnectionHandlerUpgrErr, DialError, IntoConnectionHandler, NetworkBehaviour, Swarm,
    SwarmBuilder, SwarmEvent,
};
use libp2p::{dcutr, identify, identity, noise, ping, quic, tcp};
use libp2p::{PeerId, Transport};
use log::{info, warn};
use std::collections::HashSet;
use std::convert::TryInto;
use std::env;
use std::net::{Ipv4Addr, Ipv6Addr};
use std::num::NonZeroU32;
use std::ops::ControlFlow;
use std::time::{SystemTime, UNIX_EPOCH};
use tonic::transport::{Certificate, ClientTlsConfig, Endpoint};

#[allow(clippy::derive_partial_eq_without_eq)]
pub mod grpc {
    tonic::include_proto!("_");
}

fn agent_version() -> String {
    format!("punchr/rust-client/{}", env!("CARGO_PKG_VERSION"))
}

#[derive(Parser, Debug)]
#[clap(
    name = "Rust Punchr Client",
    version,
    after_help = "Note: The api key for authentication is read from env value \"API_KEY\"."
)]
struct Opt {
    /// URL and port of the punchr server. Note that the scheme ist required.
    #[clap(
        long,
        name = "SERVER_URL",
        default_value = "https://punchr.dtrautwein.eu:443"
    )]
    server: String,

    /// Path to PEM encoded CA certificate against which the server's TLS certificate is verified
    /// [default: hardcoded CA certificate for punchr.dtrautwein.eu]
    #[clap(long, name = "PATH_TO_PEM_FILE")]
    pem: Option<String>,

    /// Fixed value to generate a deterministic peer id.
    #[clap(long, name = "SECRET_KEY_SEED")]
    seed: Option<u8>,

    /// Only run a fixed number of rounds.
    #[clap(long, name = "NUMBER_OF_ROUNDS")]
    rounds: Option<usize>,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let _ = env_logger::try_init();

    // Api-key used to authenticate our client at the server.
    let api_key = match env::var("API_KEY") {
        Ok(k) => k,
        Err(env::VarError::NotPresent) => {
            warn!("No value for env variable \"API_KEY\" found. If the server enforces authorization it will reject our requests.");
            String::new()
        }
        Err(e) => return Err(e.into()),
    };

    let opt = Opt::parse();

    let pem = match opt.pem {
        Some(path) => tokio::fs::read(path).await?,
        None => CERTIFICATE.into(),
    };

    // We have to manually set the CA certificate again which the server certificate is
    // verified, otherwise the rustls WebPKI verifier will reject the server certificate.
    let tls = ClientTlsConfig::new().ca_certificate(Certificate::from_pem(pem));
    let channel = Endpoint::from_shared(opt.server.clone())?
        .tls_config(tls)?
        .connect()
        .await?;
    let mut client = grpc::punchr_service_client::PunchrServiceClient::new(channel);

    info!("Connected to server {}.", opt.server);

    let seed = opt.seed.unwrap_or_else(rand::random);
    let local_key = generate_ed25519(seed);
    let local_peer_id = PeerId::from(local_key.public());
    info!("Local peer id: {:?}", local_peer_id);

    let mut protocols = None;

    for _ in 0..opt.rounds.unwrap_or(usize::MAX) {
        let mut swarm = init_swarm(local_key.clone()).await?;
        if protocols.is_none() {
            let supported_protocols = swarm
                .behaviour_mut()
                .new_handler()
                .inbound_protocol()
                .protocol_info()
                .into_iter()
                .map(|info| String::from_utf8(info.protocol_name().to_vec()))
                .collect::<Result<Vec<String>, _>>()?;
            let _ = protocols.insert(supported_protocols);
        }

        let request = tonic::Request::new(grpc::RegisterRequest {
            client_id: local_peer_id.to_bytes(),
            agent_version: agent_version(),
            protocols: protocols.clone().unwrap(),
            api_key: api_key.clone(),
        });

        client.register(request).await?;

        let request = tonic::Request::new(grpc::GetAddrInfoRequest {
            host_id: local_peer_id.to_bytes(),
            all_host_ids: vec![local_peer_id.to_bytes()],
            api_key: api_key.clone(),
        });

        let response = client.get_addr_info(request).await?.into_inner();

        let remote_peer_id = PeerId::from_bytes(&response.remote_id)?;
        let remote_addrs = response
            .multi_addresses
            .clone()
            .into_iter()
            .filter_map(|a| Multiaddr::try_from(a).ok())
            .collect::<Vec<_>>();

        let state = HolePunchState::new(
            local_peer_id,
            swarm.listeners(),
            remote_peer_id,
            remote_addrs.iter().map(|a| a.to_vec()).collect(),
            api_key.clone(),
        );

        if remote_addrs.is_empty() {
            let request = state.cancel("list of remote addresses is empty");
            client
                .track_hole_punch(tonic::Request::new(request))
                .await?;
            continue;
        }

        info!(
            "Attempting to punch through to {:?} via {:?}.",
            remote_peer_id, remote_addrs
        );

        swarm.dial(
            DialOpts::peer_id(remote_peer_id)
                .addresses(remote_addrs)
                .build(),
        )?;

        let request = state.drive_hole_punch(&mut swarm).await;

        client
            .track_hole_punch(tonic::Request::new(request))
            .await?;

        futures_timer::Delay::new(std::time::Duration::from_secs(1)).await;
    }

    Ok(())
}

async fn init_swarm(
    local_key: identity::Keypair,
) -> Result<Swarm<Behaviour>, Box<dyn std::error::Error>> {
    let local_peer_id = PeerId::from(local_key.public());

    let noise_keys = noise::Keypair::<noise::X25519Spec>::new()
        .into_authentic(&local_key)
        .expect("Signing libp2p-noise static DH keypair failed.");

    let (relay_transport, relay_behaviour) = Client::new_transport_and_behaviour(local_peer_id);

    let tcp_relay_transport = OrTransport::new(
        relay_transport,
        tcp::async_io::Transport::new(tcp::Config::new().port_reuse(true)),
    )
    .upgrade(upgrade::Version::V1)
    .authenticate(noise::NoiseConfig::xx(noise_keys).into_authenticated())
    .multiplex(libp2p::yamux::YamuxConfig::default());

    let mut quic_config = quic::Config::new(&local_key);
    quic_config.support_draft_29 = true;
    let quic_transport = quic::async_std::Transport::new(quic_config);

    let transport = DnsConfig::system(OrTransport::new(quic_transport, tcp_relay_transport).map(
        |either_output, _| match either_output {
            EitherOutput::First((peer_id, muxer)) => (peer_id, StreamMuxerBox::new(muxer)),
            EitherOutput::Second((peer_id, muxer)) => (peer_id, StreamMuxerBox::new(muxer)),
        },
    ))
    .await?
    .boxed();

    let identify_config = identify::Config::new("/ipfs/0.1.0".to_string(), local_key.public())
        .with_agent_version(agent_version());

    let behaviour = Behaviour {
        relay: relay_behaviour,
        ping: ping::Behaviour::new(ping::Config::new()),
        identify: identify::Behaviour::new(identify_config),
        dcutr: dcutr::behaviour::Behaviour::new(),
    };

    let mut swarm = SwarmBuilder::with_async_std_executor(transport, behaviour, local_peer_id)
        .dial_concurrency_factor(10_u8.try_into()?)
        .build();

    for ip_addr_version in [
        Protocol::from(Ipv4Addr::UNSPECIFIED),
        Protocol::from(Ipv6Addr::UNSPECIFIED),
    ]
    .into_iter()
    {
        swarm.listen_on(
            Multiaddr::empty()
                .with(ip_addr_version.clone())
                .with(Protocol::Tcp(0)),
        )?;

        swarm.listen_on(
            Multiaddr::empty()
                .with(ip_addr_version)
                .with(Protocol::Udp(0))
                .with(Protocol::Quic),
        )?;
    }

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

    let mut nodes = [
        "/ip4/139.178.91.71/tcp/4001/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
        "/ip6/2604:1380:4602:5c00::3/tcp/4001/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
        "/ip6/2604:1380:40e1:9c00::1/udp/4001/quic/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
        "/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
    ].into_iter().map(|ipfs_bootstrap_node| {
        let mut a: Multiaddr = ipfs_bootstrap_node.parse()?;
        swarm.dial(a.clone())?;
        let peer_id = match a.pop().unwrap() {
            Protocol::P2p(hash) => PeerId::from_multihash(hash).unwrap(),
            _ => unreachable!("All ipfs bootstrap node multiaddresses include the peer-id.")
        };
        Ok(peer_id)
    }).collect::<Result<HashSet<_>,  Box<dyn std::error::Error>>>()?;

    while !nodes.is_empty() {
        match swarm.select_next_some().await {
            SwarmEvent::ConnectionEstablished { endpoint, .. } => {
                info!("Connection established via {:?}", endpoint);
            }
            SwarmEvent::Behaviour(Event::Identify(identify::Event::Received { peer_id, info })) => {
                info!("{:?}", info);
                nodes.remove(&peer_id);
            }
            SwarmEvent::OutgoingConnectionError {
                peer_id: Some(peer_id),
                error,
            } => {
                // `Swarm::dial` extracts the PeerId from the multiaddr.
                nodes.remove(&peer_id);
                log::info!("dial to {} failed: {:?}", peer_id, error);
            }
            SwarmEvent::ConnectionClosed { peer_id, .. } => {
                nodes.remove(&peer_id);
            }
            _ => {}
        }
    }

    Ok(swarm)
}

struct HolePunchState {
    request: grpc::TrackHolePunchRequest,
    active_holepunch_attempt: Option<HolePunchAttemptState>,
    remote_id: PeerId,
}

impl HolePunchState {
    fn new<'a>(
        client_id: PeerId,
        client_listen_addrs: impl Iterator<Item = &'a Multiaddr>,
        remote_id: PeerId,
        remote_multi_addresses: Vec<Vec<u8>>,
        api_key: String,
    ) -> Self {
        let request = grpc::TrackHolePunchRequest {
            client_id: client_id.into(),
            remote_id: remote_id.into(),
            remote_multi_addresses,
            open_multi_addresses: Vec::new(),
            has_direct_conns: false,
            connect_started_at: unix_time_now(),
            connect_ended_at: 0,
            hole_punch_attempts: Vec::new(),
            error: None,
            outcome: grpc::HolePunchOutcome::Unknown.into(),
            ended_at: 0,
            listen_multi_addresses: client_listen_addrs.map(|a| a.to_vec()).collect(),
            api_key,
            protocols: Vec::new(),
            latency_measurements: Vec::new(),
            network_information: None,
            nat_mappings: Vec::new(),
        };
        HolePunchState {
            request,
            remote_id,
            active_holepunch_attempt: None,
        }
    }

    fn cancel(mut self, reason: &str) -> grpc::TrackHolePunchRequest {
        info!(
            "Skipping hole punch through to {:?} via {:?} because {}.",
            self.remote_id,
            self.request
                .remote_multi_addresses
                .iter()
                .filter_map(|a| Multiaddr::try_from(a.clone()).ok())
                .collect::<Vec<_>>(),
            reason
        );
        let unix_time_now = unix_time_now();
        self.request.connect_ended_at = unix_time_now;
        self.request.ended_at = unix_time_now;
        self.request.error = Some(reason.into());
        self.request.outcome = grpc::HolePunchOutcome::Cancelled.into();
        self.request
    }

    async fn drive_hole_punch(
        mut self,
        swarm: &mut libp2p::swarm::Swarm<Behaviour>,
    ) -> grpc::TrackHolePunchRequest {
        let (outcome, error) = loop {
            match swarm.select_next_some().await {
                SwarmEvent::NewListenAddr { address, .. } => info!("Listening on {:?}", address),
                SwarmEvent::Behaviour(Event::Relay(event)) => info!("{:?}", event),
                SwarmEvent::Behaviour(Event::Dcutr(event)) => {
                    info!("{:?}", event);
                    if let ControlFlow::Break((outcome, error)) = self.handle_dcutr_event(event) {
                        break (outcome, error);
                    }
                }
                SwarmEvent::Behaviour(Event::Identify(event)) => info!("{:?}", event),
                SwarmEvent::Behaviour(Event::Ping(_)) => {}
                SwarmEvent::ConnectionEstablished {
                    peer_id,
                    endpoint,
                    num_established,
                    concurrent_dial_errors,
                    ..
                } => {
                    info!("Established connection to {:?} via {:?}", peer_id, endpoint);
                    if peer_id == self.remote_id {
                        let failed = concurrent_dial_errors
                            .map(|e| {
                                e.into_iter()
                                    .map(|(addr, _err)| addr.to_vec())
                                    .collect::<Vec<_>>()
                            })
                            .unwrap_or_default();
                        if let ControlFlow::Break((outcome, error)) =
                            self.handle_established_connection(endpoint, num_established, failed)
                        {
                            break (outcome, error);
                        }
                    }
                }
                SwarmEvent::OutgoingConnectionError { peer_id, error } => {
                    info!("Outgoing connection error to {:?}: {:?}", peer_id, error);
                    let addresses = match error {
                        DialError::Transport(dials) => dials
                            .into_iter()
                            .map(|(addr, _err)| addr.to_vec())
                            .collect(),
                        _ => Default::default(),
                    };
                    if peer_id == Some(self.remote_id) {
                        let is_connected = swarm.is_connected(&self.remote_id);
                        if let ControlFlow::Break((outcome, error)) =
                            self.handle_connection_error(is_connected, addresses)
                        {
                            break (outcome, error);
                        }
                    }
                }
                SwarmEvent::ConnectionClosed {
                    peer_id,
                    endpoint,
                    num_established,
                    ..
                } => {
                    info!("Connection to {:?} via {:?} closed", peer_id, endpoint);
                    if peer_id == self.remote_id {
                        let remote_addr = endpoint.get_remote_address().clone().to_vec();
                        self.request
                            .open_multi_addresses
                            .retain(|a| a != &remote_addr);
                        if num_established == 0 {
                            let error =
                                Some("connection closed without a successful hole-punch".into());
                            break (grpc::HolePunchOutcome::Failed, error);
                        }
                    }
                }
                _ => {}
            }
        };
        self.request.outcome = outcome.into();
        self.request.error = error;
        self.request.ended_at = unix_time_now();
        info!(
            "Finished whole hole punching process: attempts: {:?}, outcome: {:?}, error: {:?}",
            self.request.hole_punch_attempts.len(),
            self.request.outcome,
            self.request.error
        );
        self.request
    }

    fn handle_dcutr_event(
        &mut self,
        event: dcutr::behaviour::Event,
    ) -> ControlFlow<(grpc::HolePunchOutcome, Option<String>)> {
        match event {
            dcutr::behaviour::Event::RemoteInitiatedDirectConnectionUpgrade { .. } => {
                self.active_holepunch_attempt = Some(HolePunchAttemptState {
                    opened_at: self.request.connect_ended_at,
                    started_at: unix_time_now(),
                    addresses: vec![],
                });
            }
            dcutr::behaviour::Event::DirectConnectionUpgradeSucceeded { .. } => {
                if let Some(attempt) = self.active_holepunch_attempt.take() {
                    let resolved = attempt.resolve(grpc::HolePunchAttemptOutcome::Success, None);
                    self.request.hole_punch_attempts.push(resolved);
                }
                return ControlFlow::Break((grpc::HolePunchOutcome::Success, None));
            }
            dcutr::behaviour::Event::DirectConnectionUpgradeFailed { error, .. } => {
                if let Some(attempt) = self.active_holepunch_attempt.take() {
                    let (attempt_outcome, attempt_error) = match error {
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
                    let resolved = attempt.resolve(attempt_outcome, Some(attempt_error.into()));
                    self.request.hole_punch_attempts.push(resolved);
                }
                let error = Some("none of the 3 attempts succeeded".into());
                return ControlFlow::Break((grpc::HolePunchOutcome::Failed, error));
            }
            dcutr::behaviour::Event::InitiatedDirectConnectionUpgrade { .. } => {
                unreachable!()
            }
        }
        ControlFlow::Continue(())
    }

    fn handle_established_connection(
        &mut self,
        endpoint: ConnectedPoint,
        num_established: NonZeroU32,
        mut failed_addrs: Vec<Vec<u8>>,
    ) -> ControlFlow<(grpc::HolePunchOutcome, Option<String>)> {
        if num_established == NonZeroU32::new(1).expect("1 != 0") {
            self.request.connect_ended_at = unix_time_now();
        }

        self.request
            .open_multi_addresses
            .push(endpoint.get_remote_address().to_vec());

        if !endpoint.is_relayed() {
            self.request.has_direct_conns = true;
        }
        match self.active_holepunch_attempt.as_mut() {
            Some(attempt) => {
                failed_addrs.push(endpoint.get_remote_address().to_vec());
                attempt.addresses.extend(failed_addrs)
            }
            None if !endpoint.is_relayed() => {
                // Reverse-dial succeeded.
                return ControlFlow::Break((grpc::HolePunchOutcome::ConnectionReversed, None));
            }
            _ => {}
        }

        ControlFlow::Continue(())
    }

    fn handle_connection_error(
        &mut self,
        is_connected: bool,
        failed_addresses: Vec<Vec<u8>>,
    ) -> ControlFlow<(grpc::HolePunchOutcome, Option<String>)> {
        if !is_connected {
            // Initial connection to the remote failed.
            self.request.connect_ended_at = unix_time_now();
            let error = Some("Error connecting to remote peer via relay.".to_string());
            return ControlFlow::Break((grpc::HolePunchOutcome::NoConnection, error));
        }

        if let Some(mut attempt) = self.active_holepunch_attempt.take() {
            attempt.addresses.extend(failed_addresses);
            // Hole punch attempt failed. Up to 3 attempts may fail before the whole hole-punch
            // is considered as failed.
            let attempt_outcome = grpc::HolePunchAttemptOutcome::Failed;
            let attempt_error = "failed to establish a direct connection";
            let resolved = attempt.resolve(attempt_outcome, Some(attempt_error.into()));
            self.request.hole_punch_attempts.push(resolved);
        }
        ControlFlow::Continue(())
    }
}

fn unix_time_now() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_nanos()
        .try_into()
        .unwrap()
}

struct HolePunchAttemptState {
    opened_at: u64,
    started_at: u64,
    addresses: Vec<Vec<u8>>,
}

impl HolePunchAttemptState {
    fn resolve(
        self,
        attempt_outcome: grpc::HolePunchAttemptOutcome,
        attempt_error: Option<String>,
    ) -> grpc::HolePunchAttempt {
        let ended_at = unix_time_now();
        info!(
            "Finished hole punching attempt with outcome: {:?}, error: {:?}",
            attempt_outcome, attempt_error
        );
        grpc::HolePunchAttempt {
            opened_at: self.opened_at,
            started_at: Some(self.started_at),
            ended_at,
            start_rtt: None,
            elapsed_time: (ended_at - self.started_at) as f32 / 1000f32,
            outcome: attempt_outcome.into(),
            error: attempt_error,
            direct_dial_error: None,
            multi_addresses: self.addresses,
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
    relay: Client,
    ping: ping::Behaviour,
    identify: identify::Behaviour,
    dcutr: dcutr::behaviour::Behaviour,
}

#[allow(clippy::large_enum_variant)]
#[derive(Debug)]
enum Event {
    Ping(ping::Event),
    Identify(identify::Event),
    Relay(client::Event),
    Dcutr(dcutr::behaviour::Event),
}

impl From<ping::Event> for Event {
    fn from(e: ping::Event) -> Self {
        Event::Ping(e)
    }
}

impl From<identify::Event> for Event {
    fn from(e: identify::Event) -> Self {
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

/// PEM encoded X509 certificate from ISRG Root X1.
/// This is the root CA certificate for punchr.dtrautwein.eu.
const CERTIFICATE: &str = "-----BEGIN CERTIFICATE-----
MIIFYDCCBEigAwIBAgIQQAF3ITfU6UK47naqPGQKtzANBgkqhkiG9w0BAQsFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTIxMDEyMDE5MTQwM1oXDTI0MDkzMDE4MTQwM1ow
TzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh
cmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwggIiMA0GCSqGSIb3DQEB
AQUAA4ICDwAwggIKAoICAQCt6CRz9BQ385ueK1coHIe+3LffOJCMbjzmV6B493XC
ov71am72AE8o295ohmxEk7axY/0UEmu/H9LqMZshftEzPLpI9d1537O4/xLxIZpL
wYqGcWlKZmZsj348cL+tKSIG8+TA5oCu4kuPt5l+lAOf00eXfJlII1PoOK5PCm+D
LtFJV4yAdLbaL9A4jXsDcCEbdfIwPPqPrt3aY6vrFk/CjhFLfs8L6P+1dy70sntK
4EwSJQxwjQMpoOFTJOwT2e4ZvxCzSow/iaNhUd6shweU9GNx7C7ib1uYgeGJXDR5
bHbvO5BieebbpJovJsXQEOEO3tkQjhb7t/eo98flAgeYjzYIlefiN5YNNnWe+w5y
sR2bvAP5SQXYgd0FtCrWQemsAXaVCg/Y39W9Eh81LygXbNKYwagJZHduRze6zqxZ
Xmidf3LWicUGQSk+WT7dJvUkyRGnWqNMQB9GoZm1pzpRboY7nn1ypxIFeFntPlF4
FQsDj43QLwWyPntKHEtzBRL8xurgUBN8Q5N0s8p0544fAQjQMNRbcTa0B7rBMDBc
SLeCO5imfWCKoqMpgsy6vYMEG6KDA0Gh1gXxG8K28Kh8hjtGqEgqiNx2mna/H2ql
PRmP6zjzZN7IKw0KKP/32+IVQtQi0Cdd4Xn+GOdwiK1O5tmLOsbdJ1Fu/7xk9TND
TwIDAQABo4IBRjCCAUIwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMCAQYw
SwYIKwYBBQUHAQEEPzA9MDsGCCsGAQUFBzAChi9odHRwOi8vYXBwcy5pZGVudHJ1
c3QuY29tL3Jvb3RzL2RzdHJvb3RjYXgzLnA3YzAfBgNVHSMEGDAWgBTEp7Gkeyxx
+tvhS5B1/8QVYIWJEDBUBgNVHSAETTBLMAgGBmeBDAECATA/BgsrBgEEAYLfEwEB
ATAwMC4GCCsGAQUFBwIBFiJodHRwOi8vY3BzLnJvb3QteDEubGV0c2VuY3J5cHQu
b3JnMDwGA1UdHwQ1MDMwMaAvoC2GK2h0dHA6Ly9jcmwuaWRlbnRydXN0LmNvbS9E
U1RST09UQ0FYM0NSTC5jcmwwHQYDVR0OBBYEFHm0WeZ7tuXkAXOACIjIGlj26Ztu
MA0GCSqGSIb3DQEBCwUAA4IBAQAKcwBslm7/DlLQrt2M51oGrS+o44+/yQoDFVDC
5WxCu2+b9LRPwkSICHXM6webFGJueN7sJ7o5XPWioW5WlHAQU7G75K/QosMrAdSW
9MUgNTP52GE24HGNtLi1qoJFlcDyqSMo59ahy2cI2qBDLKobkx/J3vWraV0T9VuG
WCLKTVXkcGdtwlfFRjlBz4pYg1htmf5X6DYO8A4jqv2Il9DjXA6USbW1FzXSLr9O
he8Y4IWS6wY7bCkjCWDcRQJMEhg76fsO3txE+FiYruq9RUWhiF1myv4Q6W+CyBFC
Dfvp7OOGAN6dEOM4+qR9sdjoSYKEBpsr6GtPAQw4dy753ec5
-----END CERTIFICATE-----";
