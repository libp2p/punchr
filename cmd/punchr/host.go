package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/multiformats/go-multiaddr"
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

	ctx       context.Context
	APIClient pb.PunchrServiceClient

	hpStatesLk sync.RWMutex
	hpStates   map[peer.ID]*HolePunchState
}

func InitHost(c *cli.Context, port string) (*Host, error) {
	log.Info("Starting libp2p host...")

	h := &Host{
		ctx:      c.Context,
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

func (h *Host) StartHolePunching() error {
	for {
		select {
		case <-h.ctx.Done():
			return h.ctx.Err()
		default:
		}

		addrInfo, err := h.RequestAddrInfo()
		if err != nil {
			log.WithError(err).Warnln("Error requesting addr info")
		}
		if addrInfo == nil {
			log.Infoln("No peer to hole punch received waiting 10s")
			select {
			case <-h.ctx.Done():
				return h.ctx.Err()
			case <-time.After(10 * time.Second):
				continue
			}
		}
		logEntry := log.WithFields(log.Fields{
			"remoteID": util.FmtPeerID(addrInfo.ID),
		})

		log.Infoln("Connecting to", addrInfo.ID.String())
		for i, maddr := range addrInfo.Addrs {
			log.Infoln("\t["+strconv.Itoa(i)+"]", maddr.String())
		}
		hpState := h.NewHolePunchState(addrInfo.ID, addrInfo.Addrs)
		if err = h.Connect(h.ctx, *addrInfo); err != nil {
			logEntry.WithError(err).Warnln("Error connecting to remote peer")
			hpState.Error = err.Error()
			hpState.EndReason = pb.HolePunchEndReason_NO_CONNECTION
		} else {
			if err = hpState.WaitForHolePunch(h.ctx); err != nil {
				logEntry.WithError(err).Warnln("Hole punch (initiation) timeout")
				hpState.EndReason = pb.HolePunchEndReason_NOT_INITIATED
				hpState.Error = err.Error()
			}
		}

		hpState = h.DeleteHolePunchState(addrInfo.ID)

		h.Peerstore().RemovePeer(addrInfo.ID)
		if err = h.Network().ClosePeer(addrInfo.ID); err != nil {
			logEntry.WithError(err).Warnln("Error closing connection")
		}

		// Persist it
		logEntry.WithFields(log.Fields{
			"attempts":  hpState.Attempts,
			"success":   hpState.Success,
			"duration":  hpState.ElapsedTime,
			"endReason": hpState.EndReason,
		}).Infoln("Tracking hole punch result")
		if err = h.TrackHolePunchResult(hpState); err != nil {
			logEntry.WithError(err).Warnln("Error tracking hole punch result")
		}
	}
}

func (h *Host) RegisterHost() error {
	log.Infoln("Registering at API server")

	bytesLocalPeerID, err := h.ID().Marshal()
	if err != nil {
		return errors.Wrap(err, "marshal peer id")
	}

	_, err = h.APIClient.Register(h.ctx, &pb.RegisterRequest{
		ClientId: bytesLocalPeerID,
		// AgentVersion: *h.GetAgentVersion(h.ID()),
		AgentVersion: "punchr/go-client/0.1.0",
		Protocols:    h.GetProtocols(h.ID()),
	})
	if err != nil {
		return errors.Wrap(err, "register client")
	}
	return nil
}

func (h *Host) RequestAddrInfo() (*peer.AddrInfo, error) {
	log.Infoln("Requesting addr infos from server...")

	bytesLocalPeerID, err := h.ID().Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal peer id")
	}

	res, err := h.APIClient.GetAddrInfo(h.ctx, &pb.GetAddrInfoRequest{ClientId: bytesLocalPeerID})
	if err != nil {
		return nil, errors.Wrap(err, "get addr info RPC")
	}
	if res.GetRemoteId() == nil {
		return nil, nil
	}

	remoteID, err := peer.IDFromBytes(res.RemoteId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from bytes")
	}

	maddrs := make([]multiaddr.Multiaddr, len(res.MultiAddresses))
	for i, maddrBytes := range res.MultiAddresses {
		maddr, err := multiaddr.NewMultiaddrBytes(maddrBytes)
		if err != nil {
			return nil, errors.Wrap(err, "multi address from bytes")
		}
		maddrs[i] = maddr
	}

	return &peer.AddrInfo{ID: remoteID, Addrs: maddrs}, nil
}

func (h *Host) TrackHolePunchResult(hps *HolePunchState) error {
	req, err := hps.ToProto(h.ID())
	if err != nil {
		return err
	}

	_, err = h.APIClient.TrackHolePunch(h.ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (h *Host) NewHolePunchState(remoteID peer.ID, maddrs []multiaddr.Multiaddr) *HolePunchState {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	h.hpStates[remoteID] = &HolePunchState{
		RemoteID:            remoteID,
		ConnectionStartedAt: time.Now(),
		RemoteMaddrs:        maddrs,
		holePunchStarted:    make(chan struct{}),
		holePunchFinished:   make(chan struct{}),
	}

	return h.hpStates[remoteID]
}

func (h *Host) DeleteHolePunchState(remoteID peer.ID) *HolePunchState {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	hpState := h.hpStates[remoteID]
	delete(h.hpStates, remoteID)

	return hpState
}

type HolePunchState struct {
	RemoteID            peer.ID
	ConnectionStartedAt time.Time
	RemoteMaddrs        []multiaddr.Multiaddr
	StartRTT            time.Duration
	ElapsedTime         time.Duration
	EndReason           pb.HolePunchEndReason
	Attempts            int
	Success             bool
	Error               string
	DirectDialError     string
	holePunchStarted    chan struct{}
	holePunchFinished   chan struct{}
}

func (hps HolePunchState) ToProto(peerID peer.ID) (*pb.TrackHolePunchRequest, error) {
	localID, err := peerID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal local peer id")
	}

	remoteID, err := hps.RemoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal remote peer id")
	}

	maddrsBytes := make([][]byte, len(hps.RemoteMaddrs))
	for i, maddr := range hps.RemoteMaddrs {
		maddrsBytes[i] = maddr.Bytes()
	}

	return &pb.TrackHolePunchRequest{
		ClientId:             localID,
		RemoteId:             remoteID,
		Success:              hps.Success,
		StartedAt:            hps.ConnectionStartedAt.UnixMilli(),
		RemoteMultiAddresses: maddrsBytes,
		Attempts:             int32(hps.Attempts),
		Error:                hps.Error,
		DirectDialError:      hps.DirectDialError,
		StartRtt:             float32(hps.StartRTT.Seconds()),
		ElapsedTime:          float32(hps.ElapsedTime.Seconds()),
		EndReason:            hps.EndReason,
	}, nil
}

func (hps HolePunchState) WaitForHolePunch(ctx context.Context) error {
	logEntry := log.WithFields(log.Fields{"remoteID": util.FmtPeerID(hps.RemoteID)})
	logEntry.WithField("waitDur", 15*time.Second).Infoln("Waiting for hole punch initiation...")

	select {
	case <-hps.holePunchStarted:
	case <-time.After(15 * time.Second):
		return errors.New("hole punch was not initiated")
	case <-ctx.Done():
		return ctx.Err()
	}

	logEntry.Infoln("Hole punch initiated!")
	logEntry.WithField("waitDur", time.Minute).Infoln("Waiting to finish...")

	// Then wait for the hole punch to finish
	select {
	case <-hps.holePunchFinished:
		return nil
	case <-time.After(time.Minute):
		return errors.New("hole punch did not finish in time")
	case <-ctx.Done():
		return ctx.Err()
	}
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
		close(hpState.holePunchStarted)
	case *holepunch.EndHolePunchEvt:
		hpState.EndReason = pb.HolePunchEndReason_HOLE_PUNCH
		hpState.Error = event.Error
		hpState.Success = event.Success
		hpState.ElapsedTime = event.EllapsedTime
		close(hpState.holePunchFinished)
	case *holepunch.HolePunchAttemptEvt:
		hpState.Attempts += 1 // event.Attempt <-- does not count correctly if hole punching with same peer happens shortly after one another.
	case *holepunch.ProtocolErrorEvt:
		hpState.EndReason = pb.HolePunchEndReason_PROTOCOL_ERROR
		hpState.Error = event.Error
		hpState.Success = false
		hpState.ElapsedTime = time.Since(hpState.ConnectionStartedAt)
		close(hpState.holePunchFinished)
	case *holepunch.DirectDialEvt:
		log.Warnln("Unexpected event DirectDialEvt")
		if event.Success {
			hpState.EndReason = pb.HolePunchEndReason_DIRECT_DIAL
			hpState.Success = event.Success
			hpState.ElapsedTime = event.EllapsedTime
			close(hpState.holePunchFinished)
		} else {
			hpState.DirectDialError = event.Error
		}
	}
}
