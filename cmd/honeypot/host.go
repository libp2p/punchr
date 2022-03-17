package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/crawler"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
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
	pm       *pb.ProtocolMessenger
}

func InitHost(c *cli.Context, port string, dbClient *db.Client) (*Host, error) {
	log.Info("Starting libp2p host...")

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
	var dht *kaddht.IpfsDHT
	agentVersion := "punchr/honeypot/" + c.App.Version
	libp2pHost, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.UserAgent(agentVersion),
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

	// Create new protocol messenger to have access to low level DHT RPC calls
	pm, err := pb.NewProtocolMessenger(&msgSender{
		h:         libp2pHost,
		protocols: kaddht.DefaultProtocols,
		timeout:   time.Minute,
	})
	if err != nil {
		return nil, errors.Wrap(err, "new protocol messenger")
	}

	h := &Host{
		ctx:      c.Context,
		Host:     libp2pHost,
		DBClient: dbClient,
		DHT:      dht,
		pm:       pm,
	}

	h.DBPeer, err = h.DBClient.UpsertPeer(c.Context, h.DBClient, h.ID(), &agentVersion, h.GetProtocols(h.ID()))
	if err != nil {
		return nil, errors.Wrap(err, "save new host identity")
	}

	// Register for all network notifications
	h.Network().Notify(h)

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
func (h *Host) WalkDHT() {
	c, err := crawler.New(h, crawler.WithParallelism(100))
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
		}

		log.Infoln("Start walking the DHT...")

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
			if errors.Is(h.ctx.Err(), context.Canceled) {
				return
			}
			log.WithError(err).WithField("remoteID", util.FmtPeerID(p)).Infoln("Done crawling peer")
			crawledPeers.With(prometheus.Labels{"status": "error"}).Inc()
		}

		c.Run(h.ctx, seedPeers, handleSuccess, handleFail)

		log.Infoln("Done walking the DHT!")
		completedWalks.Inc()
	}
}
