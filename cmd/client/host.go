package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/punchr/pkg/key"
	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
)

// Host holds information of the honeypot libp2p host.
type Host struct {
	host.Host

	client pb.PunchrServiceClient

	hpStatesLk sync.RWMutex
	hpStates   map[peer.ID]*HolePunchState
}

func InitHost(c *cli.Context, port string) (*Host, error) {
	log.Info("Starting libp2p host...")

	h := &Host{
		hpStates: map[peer.ID]*HolePunchState{},
	}

	// Load private key data from file or create a new identity
	privKeyFile := c.String("key")
	privKey, err := key.Load(privKeyFile)
	if err != nil {
		privKey, err = key.Create(privKeyFile)
		if err != nil {
			return nil, errors.Wrap(err, "load or create key pair")
		}
	}

	// Configure new libp2p host
	libp2pHost, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.UserAgent("punchr/go-client/"+c.App.Version),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%s/quic", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/udp/%s/quic", port)),
		libp2p.EnableHolePunching(holepunch.WithTracer(h)),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			return kaddht.New(c.Context, h, kaddht.Mode(kaddht.ModeClient))
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	h.Host = libp2pHost

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

func (h *Host) StartHolePunching(ctx context.Context) error {
	for {
		// Give a cancelled context precedence
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// request new peer from server
		addrInfo, err := h.RequestAddrInfo(ctx)
		if addrInfo == nil {
			if err != nil {
				log.WithError(err).Warnln("Error requesting addr info")
			} else {
				log.Infoln("No peer to hole punch received waiting 10s")
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Second):
				continue
			}
		}

		// we received a new peer -> log its information
		h.logAddrInfo(addrInfo)

		// sanity operation -> clean up all resources
		h.prunePeer(addrInfo.ID)

		// initialize a new hole punch state
		// hpStart gets closed when the hole punch process was started and hpFinish when it finished
		hpStart, hpFinish := h.NewHolePunchState(addrInfo.ID, addrInfo.Addrs)

		// connect to the remote peer via relay
		if err = h.Connect(ctx, *addrInfo); err != nil {
			h.handleConnectionError(addrInfo.ID, err)
		} else {
			// wait for the hole punch to be started by the remote
			if err = h.WaitForHolePunch(ctx, addrInfo.ID, hpStart, hpFinish); err != nil {
				h.handleHolePunchWaitError(addrInfo.ID, err)
			}
		}

		// Hole punch process is done -> clean up resources
		hpState := h.DeleteHolePunchState(addrInfo.ID)
		h.prunePeer(addrInfo.ID)

		// Log hole punch result and report it back to the server
		log.WithFields(log.Fields{
			"remoteID":  util.FmtPeerID(addrInfo.ID),
			"attempts":  hpState.Attempts,
			"success":   hpState.Success,
			"duration":  hpState.ElapsedTime,
			"endReason": hpState.EndReason,
		}).Infoln("Tracking hole punch result")
		if err = h.TrackHolePunchResult(ctx, hpState); err != nil {
			log.WithError(err).Warnln("Error tracking hole punch result")
		}
	}
}

// logAddrInfo logs address information about the given peer
func (h *Host) logAddrInfo(addrInfo *peer.AddrInfo) {
	log.WithField("openConns", len(h.Network().ConnsToPeer(addrInfo.ID))).
		Infoln("Connecting to remote peer:", addrInfo.ID.String())
	for i, maddr := range addrInfo.Addrs {
		log.Infoln("  ["+strconv.Itoa(i)+"]", maddr.String())
	}
}

// handleConnectionError is called if we could not connect to the remote peer via a relay connection.
// It updates the internal hole punch state with connection error information.
func (h *Host) handleConnectionError(pid peer.ID, err error) {
	log.WithField("remoteID", util.FmtPeerID(pid)).WithError(err).Warnln("Error connecting to remote peer")

	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	hpState, found := h.hpStates[pid]
	if !found {
		return
	}

	hpState.Error = err.Error()
	hpState.EndReason = pb.HolePunchEndReason_NO_CONNECTION
	hpState.ElapsedTime = time.Since(hpState.ConnectionStartedAt)
}

// handleHolePunchWaitError is called if we could detect a hole punch initiation from the remote peer.
// This could happen we (for whatever reason) have a direct connection to it.
// It updates the internal hole punch state with correct error information.
func (h *Host) handleHolePunchWaitError(pid peer.ID, err error) {
	log.WithField("remoteID", util.FmtPeerID(pid)).WithError(err).Warnln("Hole punch (initiation) timeout. Open connections:")

	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	hpState, found := h.hpStates[pid]
	if !found {
		return
	}

	if h.hasDirectConnToPeer(pid) {
		hpState.EndReason = pb.HolePunchEndReason_DIRECT_CONNECTION
		hpState.Error = "Client has at least one direct connection to the remote peer:\n"
		for i, conn := range h.Network().ConnsToPeer(pid) {
			hpState.Error += fmt.Sprintf("  [%d] %s: %s\n", i, conn.Stat().Direction, conn.RemoteMultiaddr())
		}
		hpState.Error = strings.TrimSpace(hpState.Error)
	} else {
		hpState.EndReason = pb.HolePunchEndReason_NOT_INITIATED
		hpState.Error = err.Error()
	}
	hpState.ElapsedTime = time.Since(hpState.ConnectionStartedAt)
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
