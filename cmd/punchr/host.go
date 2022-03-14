package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dennis-tra/punchr/pkg/util"

	"github.com/multiformats/go-multiaddr"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/maxmind"
	"github.com/dennis-tra/punchr/pkg/pb"
)

// Host holds information of the honeypot
// libp2p host.
type Host struct {
	host.Host

	ctx       context.Context
	MMClient  *maxmind.Client
	APIClient pb.PunchrServiceClient

	hpStatesLk sync.RWMutex
	hpStates   map[peer.ID]HolePunchState
}

func InitHost(ctx context.Context, version string, port string) (*Host, error) {
	log.Info("Starting libp2p host...")

	// Create maxmind client to derive geo information
	mmClient, err := maxmind.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "new maxmind client")
	}

	h := &Host{
		ctx:      ctx,
		MMClient: mmClient,
		hpStates: map[peer.ID]HolePunchState{},
	}

	// Load private key data from file or create a new identity
	key, err := loadOrCreateKeyPair()
	if err != nil {
		return nil, errors.Wrap(err, "load or create key pair")
	}

	// Configure new libp2p host
	libp2pHost, err := libp2p.New(
		libp2p.Identity(key),
		libp2p.UserAgent("punchr/"+version),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%s/quic", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/udp/%s/quic", port)),
		libp2p.EnableHolePunching(holepunch.WithTracer(h)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	h.Host = libp2pHost

	return h, nil
}

func loadOrCreateKeyPair() (crypto.PrivKey, error) {
	keyFile := "./punchr.key"

	if _, err := os.Stat(keyFile); err == nil {
		dat, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, errors.Wrap(err, "reading key data")
		}

		key, err := crypto.UnmarshalPrivateKey(dat)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal private key")
		}

		return key, err
	}

	key, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	if err != nil {
		return nil, errors.Wrap(err, "generate key pair")
	}

	keyDat, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "marshal private key")
	}

	if err = os.WriteFile(keyFile, keyDat, 0o644); err != nil {
		return nil, errors.Wrap(err, "write marshaled private key")
	}

	return key, nil
}

func (h *Host) StartHolePunching() {
	for {
		log.Infoln("Requesting addr infos from server")

		remoteID, remoteMaddr, err := h.RequestAddrInfo()
		if err != nil {
			log.WithError(err).Warnln("Error requesting addr info - sleeping 10s")
			time.Sleep(10 * time.Second)
			continue
		}
		logEntry := log.WithFields(log.Fields{
			"remoteID":    util.FmtPeerID(remoteID),
			"remoteMaddr": remoteMaddr.String(),
		})

		logEntry.Infoln("Connecting to peer")
		hpState := h.NewHolePunchState(remoteID, remoteMaddr)

		if err = h.Connect(h.ctx, peer.AddrInfo{ID: remoteID, Addrs: []multiaddr.Multiaddr{remoteMaddr}}); err != nil {
			logEntry.WithError(err).Warnln("Error connecting to remote peer")
		}
		logEntry.Infoln("Successfully connected to peer!")

		h.Peerstore().RemovePeer(hpState.PeerID)
		if err = h.Network().ClosePeer(hpState.PeerID); err != nil {
			logEntry.WithError(err).Warnln("Error closing connection")
		}

		hpState = h.DeleteHolePunchState(remoteID)

		// Persist it
		_ = hpState
	}
}

func (h *Host) NewHolePunchState(remoteID peer.ID, maddr multiaddr.Multiaddr) HolePunchState {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	h.hpStates[remoteID] = HolePunchState{
		PeerID:              remoteID,
		ConnectionStartedAt: time.Now(),
		RemoteMaddr:         maddr,
	}

	return h.hpStates[remoteID]
}

func (h *Host) DeleteHolePunchState(remoteID peer.ID) HolePunchState {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	hpState := h.hpStates[remoteID]
	delete(h.hpStates, remoteID)

	return hpState
}

func (h *Host) RequestAddrInfo() (peer.ID, multiaddr.Multiaddr, error) {
	bytesLocalPeerID, err := h.ID().Marshal()
	if err != nil {
		return "", nil, errors.Wrap(err, "marshal peer id")
	}

	res, err := h.APIClient.GetAddrInfo(h.ctx, &pb.GetAddrInfoRequest{PeerId: bytesLocalPeerID})
	if err != nil {
		return "", nil, errors.Wrap(err, "get addr info RPC")
	}

	peerID, err := peer.IDFromBytes(res.PeerId)
	if err != nil {
		return "", nil, errors.Wrap(err, "peer ID from bytes")
	}

	maddr, err := multiaddr.NewMultiaddrBytes(res.MultiAddress)
	if err != nil {
		return "", nil, errors.Wrap(err, "multi address from bytes")
	}

	return peerID, maddr, nil
}

type EndReason string

const (
	EndReasonDirectDial    EndReason = "DIRECT_DIAL"
	EndReasonProtocolError EndReason = "PROTOCOL_ERROR"
	EndReasonHolePunchEnd  EndReason = "HOLE_PUNCH"
)

type HolePunchState struct {
	PeerID              peer.ID
	ConnectionStartedAt time.Time
	RemoteMaddr         multiaddr.Multiaddr
	StartRTT            time.Duration
	ElapsedTime         time.Duration
	EndReason           EndReason
	Attempts            int
	Success             bool
	Error               string
	DirectDialError     string
}

func (h *Host) Trace(evt *holepunch.Event) {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	hpState, found := h.hpStates[evt.Remote]
	if !found {
		return
	}

	switch event := evt.Evt.(type) {
	case *holepunch.StartHolePunchEvt:
		hpState.StartRTT = event.RTT
	case *holepunch.EndHolePunchEvt:
		hpState.EndReason = EndReasonHolePunchEnd
		hpState.Error = event.Error
		hpState.Success = event.Success
		hpState.ElapsedTime = event.EllapsedTime
	case *holepunch.HolePunchAttemptEvt:
		hpState.Attempts = event.Attempt
	case *holepunch.ProtocolErrorEvt:
		hpState.EndReason = EndReasonProtocolError
		hpState.Error = event.Error
		hpState.Success = false
		hpState.ElapsedTime = time.Since(hpState.ConnectionStartedAt)
	case *holepunch.DirectDialEvt:
		if event.Success {
			hpState.EndReason = EndReasonDirectDial
			hpState.Success = event.Success
			hpState.ElapsedTime = event.EllapsedTime
		} else {
			hpState.DirectDialError = event.Error
		}
	}
}
