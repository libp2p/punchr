package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
)

var (
	CommunicationTimeout = 15 * time.Second
	RetryCount           = 3
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
	)
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	h.Host = libp2pHost

	libp2pHost.Network().Notify(h)

	return h, nil
}

func (h *Host) logEntry(remoteID peer.ID) *log.Entry {
	return log.WithFields(log.Fields{
		"remoteID": util.FmtPeerID(remoteID),
		"hostID":   util.FmtPeerID(h.ID()),
	})
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

// WaitForPublicAddr blocks execution until the host has identified its public address.
// As we currently don't have an event like this, just check our observed addresses
// regularly (exponential backoff starting at 250 ms, capped at 5s).
// TODO: There should be an event here that fires when identify discovers a new address
func (h *Host) WaitForPublicAddr(ctx context.Context) error {
	logEntry := log.WithField("hostID", util.FmtPeerID(h.ID()))
	logEntry.Infoln("Waiting for public address...")

	timeout := time.NewTimer(CommunicationTimeout)

	duration := 250 * time.Millisecond
	const maxDuration = 5 * time.Second
	t := time.NewTimer(duration)
	defer t.Stop()
	for {
		if util.ContainsPublicAddr(h.Host.Addrs()) {
			logEntry.Debug("Found >= 1 public addresses!")
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			duration *= 2
			if duration > maxDuration {
				duration = maxDuration
			}
			t.Reset(duration)
		case <-timeout.C:
			return fmt.Errorf("timeout wait for public addrs")
		}
	}
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
	defer func() { hpState.EndedAt = time.Now() }()

	// Track open connections after the hole punch
	defer func() {
		for _, conn := range h.Network().ConnsToPeer(addrInfo.ID) {
			hpState.OpenMaddrs = append(hpState.OpenMaddrs, conn.RemoteMultiaddr())
		}
		hpState.HasDirectConns = h.hasDirectConnToPeer(addrInfo.ID)
	}()

	// connect to the remote peer via relay
	hpState.ConnectStartedAt = time.Now()
	if err := h.Connect(ctx, addrInfo); err != nil {
		h.logEntry(addrInfo.ID).WithError(err).Warnln("Error connecting to remote peer")
		hpState.ConnectEndedAt = time.Now()
		hpState.Error = err.Error()
		hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_CONNECTION
		return hpState
	}
	hpState.ConnectEndedAt = time.Now()
	h.logEntry(addrInfo.ID).Infoln("Connected!")

	// we were able to connect to the remote peer.
	for i := 0; i < RetryCount; i++ {
		// wait for the DCUtR stream to be opened
		select {
		case <-h.WaitForDCUtRStream(addrInfo.ID):
			// pass
		case <-time.After(CommunicationTimeout):
			// Stream was not opened in time by the remote.
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_STREAM
			hpState.Error = "/libp2p/dcutr stream was not opened after " + CommunicationTimeout.String()
			return hpState
		case <-ctx.Done():
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_CANCELLED
			hpState.Error = ctx.Err().Error()
			return hpState
		}
		// stream was opened! Now, wait for the first hole punch event.

		hpa := hpState.TrackHolePunch(ctx, addrInfo.ID, evtChan)
		hpState.HolePunchAttempts = append(hpState.HolePunchAttempts, &hpa)

		switch hpa.Outcome {
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_PROTOCOL_ERROR:
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_FAILED
			hpState.Error = hpa.Error
			return hpState
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_DIRECT_DIAL:
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_SUCCESS
			return hpState
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_UNKNOWN:
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_UNKNOWN
			hpState.Error = "unknown hole punch attempt outcome"
			return hpState
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_CANCELLED:
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_CANCELLED
			hpState.Error = hpa.Error
			return hpState
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_TIMEOUT:
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_FAILED
			hpState.Error = hpa.Error
			return hpState
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_SUCCESS:
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_SUCCESS
			return hpState
		}
	}

	hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_FAILED
	hpState.Error = fmt.Sprintf("none of the %d attempts succeeded", RetryCount)

	return hpState
}

func (hps HolePunchState) TrackHolePunch(ctx context.Context, remoteID peer.ID, evtChan <-chan *holepunch.Event) HolePunchAttempt {
	hps.logEntry(remoteID).Infoln("Waiting for hole punch events...")
	hpa := &HolePunchAttempt{
		HostID:   hps.HostID,
		RemoteID: remoteID,
		OpenedAt: time.Now(),
	}

	for {
		select {
		case evt := <-evtChan:
			switch event := evt.Evt.(type) {
			case *holepunch.StartHolePunchEvt:
				hpa.handleStartHolePunchEvt(evt, event)
			case *holepunch.EndHolePunchEvt:
				hpa.handleEndHolePunchEvt(evt, event)
				return *hpa
			case *holepunch.HolePunchAttemptEvt:
				hpa.handleHolePunchAttemptEvt(event)
			case *holepunch.ProtocolErrorEvt:
				hpa.handleProtocolErrorEvt(evt, event)
				return *hpa
			case *holepunch.DirectDialEvt:
				hpa.handleDirectDialEvt(evt, event)
				if event.Success {
					return *hpa
				}
			default:
				panic(fmt.Sprintf("unexpected event %T", evt.Evt))
			}
		case <-time.After(CommunicationTimeout):
			hpa.handleHolePunchTimeout()
			return *hpa
		case <-ctx.Done():
			hpa.handleHolePunchCancelled(ctx.Err())
			return *hpa
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
				// If not found, it was already closed by the open stream handler
				if _, found := h.streamOpenPeers.LoadAndDelete(pid); !found {
					return evtChan
				}
				close(evtChan)
				return evtChan
			}
		}
	}

	h.logEntry(pid).Infoln("Waiting for /libp2p/dcutr stream...")
	return evtChan
}

func (h *Host) RegisterPeerToTrace(pid peer.ID) <-chan *holepunch.Event {
	evtChan := make(chan *holepunch.Event, 3) // Start, attempt, end -> so that we get the events in this order!
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
			h.logEntry(pid).WithField("evtType", evt.Type).Infoln("Draining event channel")
		default:
			close(evtChan)
			return
		}
	}
}

// logAddrInfo logs address information about the given peer
func (h *Host) logAddrInfo(addrInfo peer.AddrInfo) {
	h.logEntry(addrInfo.ID).WithField("openConns", len(h.Network().ConnsToPeer(addrInfo.ID))).Infoln("Connecting to remote peer...")
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
		h.logEntry(pid).WithError(err).Warnln("Error closing connection")
	}
	h.Peerstore().RemovePeer(pid)
	h.Peerstore().ClearAddrs(pid)
}

// Trace is called during the hole punching process
func (h *Host) Trace(evt *holepunch.Event) {
	val, found := h.holePunchEventsPeers.Load(evt.Remote)
	if !found {
		h.logEntry(evt.Remote).Infoln("Tracer event for untracked peer")
		return
	}

	val.(chan *holepunch.Event) <- evt
}

func (h *Host) Listen(network.Network, multiaddr.Multiaddr)      {}
func (h *Host) ListenClose(network.Network, multiaddr.Multiaddr) {}
func (h *Host) Connected(network.Network, network.Conn)          {}
func (h *Host) Disconnected(network.Network, network.Conn)       {}
func (h *Host) ClosedStream(_ network.Network, stream network.Stream) {
	if stream.Protocol() != holepunch.Protocol {
		return
	}
	h.logEntry(stream.Conn().RemotePeer()).Debugln("/libp2p/dcutr stream closed!")
}

func (h *Host) OpenedStream(_ network.Network, stream network.Stream) {
	// The following is a hack. `stream` does not have the `Protocol` field set yet. So we just check
	// every 5 ms for a total of 15 s.
	go func() {
		timeout := time.After(CommunicationTimeout)
		timer := time.NewTimer(0)
		for {

			select {
			case <-timeout:
				return
			case <-timer.C:
			}

			if stream.Protocol() == "" {
				timer.Reset(5 * time.Millisecond)
				continue
			}

			if stream.Protocol() != holepunch.Protocol {
				return
			}

			break

		}

		val, found := h.streamOpenPeers.LoadAndDelete(stream.Conn().RemotePeer())
		if !found {
			return
		}

		h.logEntry(stream.Conn().RemotePeer()).Debugln("/libp2p/dcutr stream opened!")
		close(val.(chan struct{}))
	}()
}
