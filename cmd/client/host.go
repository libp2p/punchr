package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"

	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p-core/crypto"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
)

// Host holds information of the honeypot libp2p host.
type Host struct {
	host.Host

	holePunchEventsPeers sync.Map
	streamOpenPeers      sync.Map
}

func InitHost(ctx context.Context, privKey crypto.PrivKey) (*Host, error) {
	log.Info("Starting libp2p host...")

	h := &Host{
		holePunchEventsPeers: sync.Map{},
		streamOpenPeers:      sync.Map{},
	}

	// Configure new libp2p host
	libp2pHost, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.UserAgent("punchr/go-client/0.1.0"),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/udp/0/quic"),
		libp2p.ListenAddrStrings("/ip6/::/tcp/0"),
		libp2p.ListenAddrStrings("/ip6/::/udp/0/quic"),
		libp2p.EnableHolePunching(holepunch.WithTracer(h)),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return kaddht.New(ctx, h, kaddht.Mode(kaddht.ModeClient))
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	h.Host = libp2pHost

	libp2pHost.Network().Notify(h)

	return h, nil
}

// Bootstrap connects this host to bootstrap peers.
func (h *Host) Bootstrap(ctx context.Context) error {
	for _, bp := range kaddht.GetDefaultBootstrapPeerAddrInfos() {
		log.WithField("remoteID", util.FmtPeerID(bp.ID)).Info("Connecting to bootstrap peer...")
		if err := h.Connect(ctx, bp); err != nil {
			return errors.Wrap(err, "connecting to bootstrap peer")
		}
	}
	return nil
}

// GetAgentVersion pulls the agent version from the peer store. Returns nil if no information is available.
func (h *Host) GetAgentVersion(pid peer.ID) *string {
	if value, err := h.Peerstore().Get(pid, "AgentVersion"); err == nil {
		av := value.(string)
		return &av
	} else {
		return nil
	}
}

// GetProtocols pulls the supported protocols of a peer from the peer store. Returns nil if no information is available.
func (h *Host) GetProtocols(pid peer.ID) []string {
	protocols, err := h.Peerstore().GetProtocols(pid)
	if err != nil {
		log.WithError(err).Warnln("Could not get protocols from peerstore")
		return nil
	}
	sort.Strings(protocols)
	return protocols
}

func (h *Host) Close() error {
	h.Host.Network().StopNotify(h)
	return h.Host.Close()
}

// HolePunch .
func (h *Host) HolePunch(ctx context.Context, addrInfo peer.AddrInfo) *HolePunchState {
	// we received a new peer to hole punch -> log its information
	h.logAddrInfo(addrInfo)

	// sanity operation -> clean up all resources before and after
	h.prunePeer(addrInfo.ID)
	defer h.prunePeer(addrInfo.ID)

	// register for tracer events for this particular peer
	evtChan := h.RegisterPeerToTrace(addrInfo.ID)
	defer h.UnregisterPeerToTrace(addrInfo.ID)

	// initialize a new hole punch state
	hpState := NewHolePunchState(h.ID(), addrInfo.ID, addrInfo.Addrs)

	// connect to the remote peer via relay
	hpState.ConnectStartedAt = time.Now()
	if err := h.Connect(ctx, addrInfo); err != nil {
		log.WithField("remoteID", util.FmtPeerID(addrInfo.ID)).WithError(err).Warnln("Error connecting to remote peer")
		hpState.ConnectEndedAt = time.Now()
		hpState.Error = err.Error()
		hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_CONNECTION
		return hpState
	}
	hpState.ConnectEndedAt = time.Now()
	defer func() { hpState.EndedAt = time.Now() }()
	log.WithField("remoteID", util.FmtPeerID(addrInfo.ID)).Infoln("Connected!")

	// we were able to connect to the remote peer.
	for i := 0; i < 3; i++ {
		// wait for the DCUtR stream to be opened
		select {
		case <-h.WaitForDCUtRStream(addrInfo.ID):
			// pass
		case <-time.After(15 * time.Second):
			// Stream was not opened in time by the remote.
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_STREAM
			hpState.Error = "/libp2p/dcutr stream was not opened in time"
			return hpState
		case <-ctx.Done():
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_CANCELLED
			hpState.Error = ctx.Err().Error()
			return hpState
		}
		// stream was opened! Now, wait for the first hole punch event.
		log.WithField("remoteID", util.FmtPeerID(addrInfo.ID)).Infoln("/libp2p/dcutr stream opened!")

		hpa, err := hpState.WaitForHolePunchAttempt(ctx, addrInfo.ID, evtChan)
		if err != nil {
			hpa.handleError(err)
			break
		}

		hpState.HolePunchAttempts = append(hpState.HolePunchAttempts, hpa)
		if hpa.Success {
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_SUCCESS
			break
		}
	}

	return hpState
}

func (hps HolePunchState) WaitForHolePunchAttempt(ctx context.Context, remoteID peer.ID, evtChan <-chan *holepunch.Event) (*HolePunchAttempt, error) {
	hpa := &HolePunchAttempt{
		RemoteID: remoteID,
		OpenedAt: time.Now(),
	}

	for {
		select {
		case evt := <-evtChan:
			switch event := evt.Evt.(type) {
			case *holepunch.StartHolePunchEvt:
				hpa.handleStartHolePunchEvt(event)
			case *holepunch.EndHolePunchEvt:
				hpa.handleEndHolePunchEvt(event)
				return hpa, nil
			case *holepunch.HolePunchAttemptEvt:
				hpa.handleHolePunchAttemptEvt(event)
			case *holepunch.ProtocolErrorEvt:
				hpa.handleProtocolErrorEvt(event)
				return hpa, nil
			case *holepunch.DirectDialEvt:
				hpa.handleDirectDialEvt(event)
				if event.Success {
					return hpa, nil
				}
			default:
				panic(fmt.Sprintf("unexpected event %T", evt.Evt))
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (h *Host) WaitForDCUtRStream(pid peer.ID) <-chan struct{} {
	evtChan := make(chan struct{})
	h.streamOpenPeers.Store(pid, evtChan)

	// Exit early if the DCUtR stream is already open
	for _, conn := range h.Network().ConnsToPeer(pid) {
		for _, stream := range conn.GetStreams() {
			if stream.Protocol() == holepunch.Protocol {
				close(evtChan)
				h.streamOpenPeers.Delete(pid)
				return evtChan
			}
		}
	}

	log.WithField("remoteID", util.FmtPeerID(pid)).Infoln("Waiting for /libp2p/dcutr stream...")
	return evtChan
}

func (h *Host) RegisterPeerToTrace(pid peer.ID) <-chan *holepunch.Event {
	evtChan := make(chan *holepunch.Event)
	h.holePunchEventsPeers.Store(pid, evtChan)
	return evtChan
}

func (h *Host) UnregisterPeerToTrace(pid peer.ID) {
	val, found := h.holePunchEventsPeers.LoadAndDelete(pid)
	if !found {
		return
	}
	evtChan := val.(chan *holepunch.Event)

	// Drain channel
	for {
		select {
		case evt := <-evtChan:
			log.WithFields(log.Fields{
				"remoteID": util.FmtPeerID(evt.Peer),
				"evtType":  evt.Type,
			}).Infoln("Draining event channel")
		default:
			close(evtChan)
			return
		}
	}
}

// logAddrInfo logs address information about the given peer
func (h *Host) logAddrInfo(addrInfo peer.AddrInfo) {
	log.WithField("openConns", len(h.Network().ConnsToPeer(addrInfo.ID))).
		Infoln("Connecting to remote peer:", addrInfo.ID.String())
	for i, maddr := range addrInfo.Addrs {
		log.Infoln("  ["+strconv.Itoa(i)+"]", maddr.String())
	}
}

// hasDirectConnToPeer returns true if the libp2p host has a direct (non-relay) connection to the given peer.
func (h *Host) hasDirectConnToPeer(pid peer.ID) bool {
	for _, conn := range h.Network().ConnsToPeer(pid) {
		if !util.IsRelayedMaddr(conn.RemoteMultiaddr()) {
			return true
		}
	}
	return false
}

// prunePeer closes all connections to the given peer and removes all information about it from the peer store.
func (h *Host) prunePeer(pid peer.ID) {
	if err := h.Network().ClosePeer(pid); err != nil {
		log.WithError(err).WithField("remoteID", util.FmtPeerID(pid)).Warnln("Error closing connection")
	}
	h.Peerstore().RemovePeer(pid)
	h.Peerstore().ClearAddrs(pid)
}

// Trace is called during the hole punching process
func (h *Host) Trace(evt *holepunch.Event) {
	val, found := h.holePunchEventsPeers.Load(evt.Peer)
	if !found {
		return
	}

	val.(chan *holepunch.Event) <- evt
}

func (h *Host) Listen(network network.Network, m multiaddr.Multiaddr)       {}
func (h *Host) ListenClose(network network.Network, m multiaddr.Multiaddr)  {}
func (h *Host) Connected(network network.Network, conn network.Conn)        {}
func (h *Host) Disconnected(network network.Network, conn network.Conn)     {}
func (h *Host) ClosedStream(network network.Network, stream network.Stream) {}

func (h *Host) OpenedStream(network network.Network, stream network.Stream) {
	if stream.Protocol() != holepunch.Protocol {
		return
	}

	val, found := h.holePunchEventsPeers.LoadAndDelete(stream.Conn().RemotePeer())
	if !found {
		return
	}
	close(val.(chan struct{}))
}
