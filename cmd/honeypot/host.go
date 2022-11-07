package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"

	"github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/crawler"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/key"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/util"
)

// Host holds information of the honeypot libp2p host.
type Host struct {
	host.Host

	ctx      context.Context
	DBPeer   *models.Peer
	DBClient *db.Client
	DHT      *kaddht.IpfsDHT
	crawlers int
}

func InitHost(c *cli.Context, port string, dbClient *db.Client) (*Host, error) {
	log.Info("Starting libp2p host...")

	// Load private key data from file or create a new identity
	privKeyFile := c.String("key")
	privKeys, err := key.Load(privKeyFile)
	if err != nil || len(privKeys) < 1 {
		privKeys, err = key.Add(privKeyFile, 1)
		if err != nil {
			return nil, errors.Wrap(err, "load or create key pair")
		}
	}

	// Configure the resource manager to not limit anything
	limiter := rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits)
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, errors.Wrap(err, "new resource manager")
	}

	// Configure new libp2p host
	var dht *kaddht.IpfsDHT
	agentVersion := "punchr/honeypot/" + c.App.Version
	libp2pHost, err := libp2p.New(
		libp2p.Identity(privKeys[0]),
		libp2p.UserAgent(agentVersion),
		libp2p.ResourceManager(rm),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%s/quic", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/udp/%s/quic", port)),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			dht, err = kaddht.New(c.Context, h, kaddht.Mode(kaddht.ModeServer))
			return dht, err
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host")
	}

	h := &Host{
		ctx:      c.Context,
		Host:     libp2pHost,
		DBClient: dbClient,
		DHT:      dht,
		crawlers: c.Int("crawler-count"),
	}

	h.DBPeer, err = h.DBClient.UpsertPeer(c.Context, h.DBClient, h.ID(), &agentVersion, h.GetProtocols(h.ID()))
	if err != nil {
		return nil, errors.Wrap(err, "save new host identity")
	}

	// Register for all network notifications
	h.Network().Notify(h)

	return h, nil
}

func (h *Host) Close() error {
	h.Network().StopNotify(h)
	return h.Host.Close()
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

// GetMultiAddresses returns a list of multi addresses for the given peer.
func (h *Host) GetMultiAddresses(pid peer.ID) []ma.Multiaddr {
	return h.Peerstore().Addrs(pid)
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

// WalkDHT slowly enumerates the whole DHT to announce ourselves to the network.
func (h *Host) WalkDHT(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		log.Infoln("Start walking the DHT...")

		c, err := crawler.New(h, crawler.WithParallelism(h.crawlers), crawler.WithConnectTimeout(5*time.Second), crawler.WithMsgTimeout(5*time.Second))
		if err != nil {
			log.WithError(err).Infoln("Could not create crawler")
			time.Sleep(10 * time.Second)
			continue
		}

		bps := kaddht.GetDefaultBootstrapPeerAddrInfos()
		seedPeers := make([]*peer.AddrInfo, len(bps))
		for i, bp := range bps {
			seedPeers[i] = &bp
		}

		handleSuccess := func(p peer.ID, rtPeers []*peer.AddrInfo) {
			log.WithField("remoteID", util.FmtPeerID(p)).Infoln("Done crawling peer")
			crawledPeers.With(prometheus.Labels{"status": "ok"}).Inc()
		}

		handleFail := func(p peer.ID, err error) {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			log.WithError(err).WithField("remoteID", util.FmtPeerID(p)).Infoln("Done crawling peer")
			crawledPeers.With(prometheus.Labels{"status": "error"}).Inc()
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		c.Run(timeoutCtx, seedPeers, handleSuccess, handleFail)
		cancel()

		if timeoutCtx.Err() == nil {
			log.Infoln("Done walking the DHT!")
		} else {
			log.WithError(timeoutCtx.Err()).Infoln("Done walking the DHT!")
		}
		completedWalks.Inc()

		for _, conn := range h.Network().Conns() {
			if err := conn.Close(); err != nil {
				log.WithError(err).WithField("remoteID", util.FmtPeerID(conn.RemotePeer())).Warnln("Could not close connection to peer.")
			}
		}
	}
}
