package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackpal/gateway"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/dennis-tra/punchr/pkg/key"
	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
)

// Punchr is responsible for fetching information from the server,
// distributing the work load to different hosts and then reporting
// the results back.
type Punchr struct {
	hosts              []*Host
	apiKey             string
	privKeyFile        string
	client             pb.PunchrServiceClient
	clientConn         *grpc.ClientConn
	routerHTML         string
	disableRouterCheck bool
}

func NewPunchr(c *cli.Context) (*Punchr, error) {
	// Dial gRPC server
	addr := fmt.Sprintf("%s:%s", c.String("server-host"), c.String("server-port"))
	log.WithField("addr", addr).Infoln("Dial server")

	// Derive transport credentials from configuration
	var tc credentials.TransportCredentials
	if c.Bool("server-ssl") {
		config := &tls.Config{InsecureSkipVerify: c.Bool("server-ssl-skip-verify")}
		tc = credentials.NewTLS(config)
	} else {
		tc = insecure.NewCredentials()
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(tc))
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial")
	}

	return &Punchr{
		hosts:              make([]*Host, c.Int("host-count")),
		apiKey:             c.String("api-key"),
		privKeyFile:        c.String("key-file"),
		client:             pb.NewPunchrServiceClient(conn),
		clientConn:         conn,
		disableRouterCheck: c.Bool("disable-router-check"),
	}, nil
}

func (p Punchr) InitHosts(c *cli.Context) error {
	privKeys, err := key.Load(p.privKeyFile)
	if err != nil {
		privKeys, err = key.Add(p.privKeyFile, len(p.hosts))
		if err != nil {
			return errors.Wrap(err, "create new key pairs")
		}
	} else if len(p.hosts) > len(privKeys) {
		// we have more hosts than keys, generate remaining
		additionalPrivKeys, err := key.Add(p.privKeyFile, len(p.hosts)-len(privKeys))
		if err != nil {
			return errors.Wrap(err, "create new key pairs")
		}
		privKeys = append(privKeys, additionalPrivKeys...)
	}

	for i := range p.hosts {
		h, err := InitHost(c, privKeys[i])
		if err != nil {
			return errors.Wrap(err, "init host")
		}
		p.hosts[i] = h
	}

	return nil
}

// UpdateRouterHTML discovers the default gateway address and fetches the
// home HTML page to get a sense which router is used.
func (p Punchr) UpdateRouterHTML() error {
	router, err := gateway.DiscoverGateway()
	if err != nil {
		return errors.Wrap(err, "discover gateway")
	}

	u := url.URL{Scheme: "http", Host: router.String()}
	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read response body")
	}

	p.routerHTML = string(html)

	return nil
}

// Bootstrap loops through all hosts, connects each of them to the canonical bootstrap nodes, and
// waits until they have identified their public address(es).
func (p Punchr) Bootstrap(ctx context.Context) error {
	var wg sync.WaitGroup
	var successes int32

	for _, h := range p.hosts {
		wg.Add(1)
		h2 := h
		go func() {
			defer wg.Done()
			log.WithField("hostID", util.FmtPeerID(h2.ID())).Info("Bootstrapping host...")
			if err := h2.Bootstrap(ctx); err != nil {
				log.Warnf("bootstrapping host %s: %s\n", util.FmtPeerID(h2.ID()), err)
				return
			}

			if err := h2.WaitForPublicAddr(ctx); err != nil {
				log.Warnf("waiting for public addr host %s: %s\n", util.FmtPeerID(h2.ID()), err)
				return
			}
			atomic.AddInt32(&successes, 1)
		}()
	}
	wg.Wait()

	if successes >= 3 || successes == int32(len(p.hosts)) {
		return nil
	} else {
		return fmt.Errorf("could not connect to enough hosts")
	}
}

// Register makes all hosts known to the server.
func (p Punchr) Register(c *cli.Context) error {
	for i, h := range p.hosts {
		log.WithField("hostID", util.FmtPeerID(h.ID())).WithField("hostNum", i).Infoln("Registering host at Punchr server")

		bytesLocalPeerID, err := h.ID().Marshal()
		if err != nil {
			return errors.Wrap(err, "marshal peer id")
		}

		av := "punchr/go-client/" + c.App.Version
		apiKey := c.String("api-key")
		req := &pb.RegisterRequest{
			ClientId:     bytesLocalPeerID,
			AgentVersion: &av,
			ApiKey:       &apiKey,
			Protocols:    h.GetProtocols(h.ID()),
		}
		if _, err = p.client.Register(c.Context, req); err != nil {
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

		if !p.disableRouterCheck {
			// Check if the multi addresses have changed - if that's the case we have switched networks
			for _, maddr := range h.Addrs() {
				if _, found := h.maddrs[maddr]; found {
					continue
				}

				log.Infoln("Found new multi addresses - fetching Router Login")
				if err = p.UpdateRouterHTML(); err != nil {
					log.WithError(err).Warnln("Could not get router HTML page")
				} else {
					hpState.RouterHTML = p.routerHTML
				}

				// Update list of multi addresses
				h.maddrs = map[multiaddr.Multiaddr]struct{}{}
				for _, newMaddr := range h.Addrs() {
					h.maddrs[newMaddr] = struct{}{}
				}

				break
			}
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
		ApiKey:     &p.apiKey,
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

	req, err := hps.ToProto(p.apiKey)
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
