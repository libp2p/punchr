package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	pb "github.com/libp2p/go-libp2p-kad-dht/pb"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/maxmind"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/util"
)

// Host holds information of the honeypot
// libp2p host.
type Host struct {
	host.Host

	ctx      context.Context
	DBPeer   *models.Peer
	DBClient *db.Client
	MMClient *maxmind.Client
	DHT      *kaddht.IpfsDHT
	pm       *pb.ProtocolMessenger
}

func InitHost(ctx context.Context, version string, port string, dbClient *db.Client) (*Host, error) {
	log.Info("Starting libp2p host...")

	// Load private key data from file or create a new identity
	key, err := loadOrCreateKeyPair()
	if err != nil {
		return nil, errors.Wrap(err, "load or create key pair")
	}

	// Configure connection manager to allow unlimited incoming connections
	cmgr, err := connmgr.NewConnManager(0, math.MaxInt)
	if err != nil {
		return nil, errors.Wrap(err, "new connection manager")
	}

	// Configure new libp2p host
	var dht *kaddht.IpfsDHT
	agentVersion := "punchr/honeypot/" + version
	libp2pHost, err := libp2p.New(
		libp2p.Identity(key),
		libp2p.UserAgent(agentVersion),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%s/quic", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/tcp/%s", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip6/::/udp/%s/quic", port)),
		libp2p.ConnectionManager(cmgr),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			var err error
			dht, err = kaddht.New(ctx, h, kaddht.Mode(kaddht.ModeServer))
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

	// Create maxmind client to derive geo information
	mmClient, err := maxmind.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "new maxmind client")
	}

	h := &Host{
		ctx:      ctx,
		Host:     libp2pHost,
		DBClient: dbClient,
		MMClient: mmClient,
		DHT:      dht,
		pm:       pm,
	}

	h.DBPeer, err = h.saveIdentity(ctx, agentVersion)
	if err != nil {
		return nil, errors.Wrap(err, "save new host identity")
	}

	// Register for all network notifications
	h.Network().Notify(h)

	return h, nil
}

func loadOrCreateKeyPair() (crypto.PrivKey, error) {
	keyFile := "./honeypot.key"

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

func (h *Host) saveIdentity(ctx context.Context, agentVersion string) (*models.Peer, error) {
	dbPeer := &models.Peer{
		MultiHash:     h.ID().String(),
		AgentVersion:  null.StringFrom(agentVersion),
		Protocols:     h.GetProtocols(h.ID()),
		SupportsDcutr: false,
	}

	err := dbPeer.Upsert(
		ctx,
		h.DBClient,
		true,
		[]string{models.PeerColumns.MultiHash},
		boil.Whitelist( // which columns to update on conflict
			models.PeerColumns.UpdatedAt,
			models.PeerColumns.AgentVersion,
			models.PeerColumns.Protocols,
			models.PeerColumns.SupportsDcutr,
		),
		boil.Infer(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "upsert host peer")
	}

	return dbPeer, nil
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
	for {
		log.Infoln("Start walking the DHT...")

		dialed := map[peer.ID]peer.AddrInfo{}
		toDial := map[peer.ID]peer.AddrInfo{}

		for _, bp := range kaddht.GetDefaultBootstrapPeerAddrInfos() {
			toDial[bp.ID] = bp
		}

		for {
			select {
			case <-h.ctx.Done():
				return
			default:
			}

			p, ok := anyAddrInfo(toDial)
			if !ok {
				break
			}

			log.WithField("remoteID", util.FmtPeerID(p.ID)).Infoln("Crawling remote peer")

			delete(toDial, p.ID)
			dialed[p.ID] = p

			connCtx, cancel := context.WithTimeout(h.ctx, 10*time.Second) // Be aggressive
			if err := h.Connect(connCtx, p); err != nil {
				log.Warnln(errors.Wrapf(err, "connect to peer"))
				cancel()
				continue
			}
			cancel()

			rt, err := kbucket.NewRoutingTable(20, kbucket.ConvertPeerID(p.ID), time.Hour, nil, time.Hour, nil)
			if err != nil {
				log.Warnln(errors.Wrapf(err, "create new routing table"))
				continue
			}

			for i := uint(0); i <= 15; i++ { // 15 is maximum
				// Generate a peer with the given common prefix length
				rpi, err := rt.GenRandPeerID(i)
				if err != nil {
					log.Warnln(errors.Wrapf(err, "generating random peer ID with CPL %d", i))
					break
				}

				neighbors, err := h.pm.GetClosestPeers(h.ctx, p.ID, rpi)
				if err != nil {
					log.Warnln(errors.Wrapf(err, "getting closest peer with CPL %d", i))
					break
				}

				for _, n := range neighbors {
					if _, found := dialed[n.ID]; !found {
						toDial[n.ID] = *n
					}
				}
			}
		}

		log.Infoln("Done walking the DHT!")
	}
}

// anyAddrInfo returns a "random" address information struct from the given map.
func anyAddrInfo(m map[peer.ID]peer.AddrInfo) (peer.AddrInfo, bool) {
	for _, addrInfo := range m {
		return addrInfo, true
	}
	return peer.AddrInfo{}, false
}
