package main

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dennis-tra/punchr/pkg/key"
	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
)

// Punchr is responsible for fetching information from the server,
// distributing the work load to different hosts and then reporting
// the results back.
type Punchr struct {
	hosts         []*Host
	privKeyPrefix string
	client        pb.PunchrServiceClient
	clientConn    *grpc.ClientConn
}

func NewPunchr(c *cli.Context) (*Punchr, error) {
	// Dial gRPC server
	addr := fmt.Sprintf("%s:%s", c.String("server-host"), c.String("server-port"))
	log.WithField("addr", addr).Infoln("Dial server")
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial")
	}

	return &Punchr{
		hosts:         make([]*Host, c.Int("host-count")),
		privKeyPrefix: c.String("key-prefix"),
		client:        pb.NewPunchrServiceClient(conn),
		clientConn:    conn,
	}, nil
}

func (p Punchr) InitHosts(ctx context.Context) error {
	for i := range p.hosts {
		// Load private key data from file or create a new identity
		privKeyFile := fmt.Sprintf("%s-%d.key", p.privKeyPrefix, i)
		privKey, err := key.Load(privKeyFile)
		if err != nil {
			privKey, err = key.Create(privKeyFile)
			if err != nil {
				return errors.Wrap(err, "load or create key pair")
			}
		}

		h, err := InitHost(ctx, privKey)
		if err != nil {
			return errors.Wrap(err, "init host")
		}
		p.hosts[i] = h
	}

	return nil
}

// Bootstrap loops through all hosts and connects each of them to the canonical bootstrap nodes.
func (p Punchr) Bootstrap(ctx context.Context) error {
	for i, h := range p.hosts {
		log.WithField("hostNum", i).Info("Bootstrapping host...")
		if err := h.Bootstrap(ctx); err != nil {
			return errors.Wrapf(err, "bootstrapping host %d", i)
		}
	}
	return nil
}

// Register makes all hosts known to the server.
func (p Punchr) Register(ctx context.Context) error {
	for i, h := range p.hosts {
		log.WithField("hostID", util.FmtPeerID(h.ID())).WithField("hostNum", i).Infoln("Registering host at Punchr server")

		bytesLocalPeerID, err := h.ID().Marshal()
		if err != nil {
			return errors.Wrap(err, "marshal peer id")
		}

		av := "punchr/go-client/0.1.0"
		req := &pb.RegisterRequest{
			ClientId:     bytesLocalPeerID,
			AgentVersion: &av,
			// AgentVersion: *h.GetAgentVersion(h.ID()),
			Protocols: h.GetProtocols(h.ID()),
		}
		if _, err = p.client.Register(ctx, req); err != nil {
			return errors.Wrapf(err, "registering host %d", i)
		}
	}

	return nil
}

// StartHolePunching requests a peer from the server, chooses a host to perform a hole punch,
// and then reports back the result.
func (p Punchr) StartHolePunching(ctx context.Context) error {
	i := 0
	for {
		// Give a cancelled context precedence
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Choose host for this round
		h := p.hosts[i]

		// Request peer to hole punch
		addrInfo, err := p.RequestAddrInfo(ctx, h.ID())
		if addrInfo == nil {
			if err != nil {
				log.WithError(err).Warnln("Error requesting addr info")
			} else {
				log.Infoln("No peer to hole punch received waiting 10s")
			}

			// Wait 10s until next request in either case
			select {
			case <-time.After(10 * time.Second):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Instruct the i-th host to hole punch
		hpState := h.HolePunch(ctx, *addrInfo)

		// Conditions for a connection reversal:
		//   1. /libp2p/dcutr stream was not opened.
		//   2. We connected to the remote peer via a relay
		//   3. We have a direct connection to the remote peer after we have waited for the libp2p/dcutr stream.
		if hpState.Outcome == pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_STREAM && hpState.onlyRelayRemoteAddrs() && hpState.HasDirectConns {
			hpState.Outcome = pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_CONNECTION_REVERSED
		}

		// Tell the server about the hole punch outcome
		if err = p.TrackHolePunchResult(ctx, hpState); err != nil {
			log.WithError(err).Warnln("Error tracking hole punch result")
		}

		// Choose next host in our list or roll over to the beginning
		i += 1
		i %= len(p.hosts)
	}
}

// RequestAddrInfo calls the hole punching server for a new peer + multi address to hole punch.
func (p Punchr) RequestAddrInfo(ctx context.Context, clientID peer.ID) (*peer.AddrInfo, error) {
	log.Infoln("Requesting peer to hole punch from server...")

	// Marshal client ID
	hostID, err := clientID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal client id")
	}

	allHostIDs := [][]byte{}
	for _, h := range p.hosts {
		marshalled, err := h.ID().Marshal()
		if err != nil {
			return nil, errors.Wrap(err, "marshal client id")
		}
		allHostIDs = append(allHostIDs, marshalled)
	}

	// Request address information
	req := &pb.GetAddrInfoRequest{
		HostId:     hostID,
		AllHostIds: allHostIDs,
	}

	res, err := p.client.GetAddrInfo(ctx, req)
	if st, ok := status.FromError(err); ok && st != nil {
		if st.Code() == codes.NotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "get addr info RPC")
	}

	// If not remote ID is given the server does not have a peer to hole punch
	if res.GetRemoteId() == nil {
		return nil, nil
	}

	// Parse response
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

func (p Punchr) TrackHolePunchResult(ctx context.Context, hps *HolePunchState) error {
	// Log hole punch result and report it back to the server
	log.WithFields(log.Fields{
		"hostID":    util.FmtPeerID(hps.HostID),
		"remoteID":  util.FmtPeerID(hps.RemoteID),
		"attempts":  len(hps.HolePunchAttempts),
		"endReason": hps.Outcome,
	}).Infoln("Tracking hole punch result")

	req, err := hps.ToProto()
	if err != nil {
		return err
	}
	_, err = p.client.TrackHolePunch(ctx, req)
	return err
}

func (p Punchr) Close() error {
	if err := p.clientConn.Close(); err != nil {
		log.WithError(err).Warnln("Closing gRPC server connection")
	}
	for _, h := range p.hosts {
		if err := h.Close(); err != nil {
			log.WithError(err).Warnln("Could not close host")
		}
	}
	return nil
}
