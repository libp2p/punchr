package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
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
	hpStates   map[peer.ID]HolePunchState
}

func InitHost(c *cli.Context, port string) (*Host, error) {
	log.Info("Starting libp2p host...")

	h := &Host{
		ctx:      c.Context,
		hpStates: map[peer.ID]HolePunchState{},
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
	)
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	h.Host = libp2pHost

	return h, nil
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

		log.Infoln("Requesting addr infos from server")

		addrInfo, err := h.RequestAddrInfo()
		if err != nil {
			log.WithError(err).Warnln("Error requesting addr info - sleeping 10s")
			time.Sleep(10 * time.Second)
			select {
			case <-h.ctx.Done():
				return h.ctx.Err()
			case <-time.After(10 * time.Second):
				continue
			}
		} else if addrInfo == nil {
			log.Infoln("No peer to hole punch received")
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

		logEntry.Infoln("Connecting to peer")
		hpState := h.NewHolePunchState(addrInfo.ID, addrInfo.Addrs)

		if err = h.Connect(h.ctx, *addrInfo); err != nil {
			logEntry.WithError(err).Warnln("Error connecting to remote peer")
		} else {
			logEntry.Infoln("Successfully connected to peer!")
		}

		h.Peerstore().RemovePeer(hpState.RemoteID)
		if err = h.Network().ClosePeer(hpState.RemoteID); err != nil {
			logEntry.WithError(err).Warnln("Error closing connection")
		}

		hpState = h.DeleteHolePunchState(addrInfo.ID)

		// Persist it
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

func (h *Host) TrackHolePunchResult(hps HolePunchState) error {
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

func (h *Host) NewHolePunchState(remoteID peer.ID, maddrs []multiaddr.Multiaddr) HolePunchState {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	h.hpStates[remoteID] = HolePunchState{
		RemoteID:            remoteID,
		ConnectionStartedAt: time.Now(),
		RemoteMaddrs:        maddrs,
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

type EndReason string

const (
	EndReasonDirectDial    EndReason = "DIRECT_DIAL"
	EndReasonProtocolError EndReason = "PROTOCOL_ERROR"
	EndReasonHolePunchEnd  EndReason = "HOLE_PUNCH"
)

func (er EndReason) ToProto() pb.HolePunchEndReason {
	switch er {
	case EndReasonDirectDial:
		return pb.HolePunchEndReason_DIRECT_DIAL
	case EndReasonProtocolError:
		return pb.HolePunchEndReason_PROTOCOL_ERROR
	case EndReasonHolePunchEnd:
		return pb.HolePunchEndReason_HOLE_PUNCH
	default:
		return pb.HolePunchEndReason_UNKNOWN
	}
}

type HolePunchState struct {
	RemoteID            peer.ID
	ConnectionStartedAt time.Time
	RemoteMaddrs        []multiaddr.Multiaddr
	StartRTT            time.Duration
	ElapsedTime         time.Duration
	EndReason           EndReason
	Attempts            int
	Success             bool
	Error               string
	DirectDialError     string
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
		EndReason:            hps.EndReason.ToProto(),
	}, nil
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
