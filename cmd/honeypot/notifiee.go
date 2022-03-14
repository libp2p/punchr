package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/util"
)

func (h *Host) Listen(network network.Network, multiaddr ma.Multiaddr)      {}
func (h *Host) ListenClose(network network.Network, multiaddr ma.Multiaddr) {}
func (h *Host) OpenedStream(network network.Network, stream network.Stream) {}
func (h *Host) ClosedStream(network network.Network, stream network.Stream) {}

func (h *Host) Connected(_ network.Network, conn network.Conn) {
	connOpen.With(prometheus.Labels{"direction": conn.Stat().Direction.String()}).Inc()
	openConns.With(prometheus.Labels{"direction": conn.Stat().Direction.String()}).Inc()

	if err := h.handleConnected(conn); err != nil {
		log.WithError(err).Warnln("An error occurred when handling connection event")
	}
}

func (h *Host) Disconnected(network network.Network, conn network.Conn) {
	connClose.With(prometheus.Labels{"direction": conn.Stat().Direction.String()}).Inc()
	openConns.With(prometheus.Labels{"direction": conn.Stat().Direction.String()}).Dec()
}

func (h *Host) handleConnected(conn network.Conn) error {
	// We can do expensive things here as it's called within a go-routine by swarm
	start := time.Now()

	// Start a database transaction
	txn, err := h.DBClient.BeginTx(h.ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin txn")
	}
	defer deferRollback(txn)

	log.WithFields(log.Fields{
		"remoteID":    util.FmtPeerID(conn.RemotePeer()),
		"remoteMaddr": conn.RemoteMultiaddr(),
		"direction":   conn.Stat().Direction,
	}).Infoln("Opened Connection")

	dbIPAddresses, country, continent, asn, err := h.saveIPAddresses(h.ctx, txn, conn.RemoteMultiaddr())
	if err != nil {
		return errors.Wrap(err, "save ip addresses")
	}

	dbMaddr := &models.MultiAddress{
		Maddr:          conn.RemoteMultiaddr().String(),
		Country:        null.StringFromPtr(country),
		Continent:      null.StringFromPtr(continent),
		Asn:            null.IntFromPtr(asn),
		IsPublic:       manet.IsPublicAddr(conn.RemoteMultiaddr()),
		IPAddressCount: len(dbIPAddresses),
	}
	if err = dbMaddr.Upsert(h.ctx, txn, true, []string{models.MultiAddressColumns.Maddr}, boil.Whitelist(models.MultiAddressColumns.UpdatedAt), boil.Infer()); err != nil {
		return errors.Wrap(err, "upsert multi address")
	}

	if err := dbMaddr.SetIPAddresses(h.ctx, txn, false); err != nil {
		return errors.Wrap(err, "removing ip address to multi address association")
	}

	if err = dbMaddr.AddIPAddresses(h.ctx, txn, false, dbIPAddresses...); err != nil {
		return errors.Wrap(err, "adding ip addresses to multi address")
	}

	agentVersion := h.GetAgentVersion(conn.RemotePeer())
	protocols := h.GetProtocols(conn.RemotePeer())
	dbPeer := &models.Peer{
		MultiHash:     conn.RemotePeer().String(),
		AgentVersion:  null.StringFromPtr(agentVersion),
		Protocols:     protocols,
		SupportsDcutr: doesSupportDCUtR(protocols),
	}
	err = dbPeer.Upsert(
		h.ctx,
		txn,
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
		return errors.Wrap(err, "upsert peer")
	}

	dbConnEvt := &models.ConnectionEvent{
		LocalID:        h.DBPeer.ID,
		RemoteID:       dbPeer.ID,
		MultiAddressID: dbMaddr.ID,
		Direction:      mapNetDirection(conn),
		Relayed:        isRelayedMaddr(conn.RemoteMultiaddr()),
		OpenedAt:       conn.Stat().Opened,
	}
	if err = dbConnEvt.Insert(h.ctx, txn, boil.Infer()); err != nil {
		return errors.Wrap(err, "insert connection event")
	}

	if err = txn.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	handleConnDur.Observe(time.Since(start).Seconds())

	log.WithField("duration", time.Since(start).String()).Infoln("Done handling")

	// Hack incoming: If we detect that the information is missing we
	// sleep for a bit and try to access again.
	if agentVersion == nil || protocols == nil {
		for i := 0; i < 20; i++ {
			av := h.GetAgentVersion(conn.RemotePeer())
			prots := h.GetProtocols(conn.RemotePeer())

			if h.GetAgentVersion(conn.RemotePeer()) == nil || h.GetProtocols(conn.RemotePeer()) == nil {
				time.Sleep(100 * time.Millisecond * time.Duration(1+i))
				continue
			}

			log.Debugln("Found after iteration", i)
			dbPeer.AgentVersion = null.StringFromPtr(av)
			dbPeer.Protocols = prots
			if _, err = dbPeer.Update(h.ctx, h.DBClient, boil.Infer()); err != nil {
				return errors.Wrap(err, "update peer")
			}
			break
		}
	}

	return nil
}

func (h *Host) saveIPAddresses(ctx context.Context, exec boil.ContextExecutor, maddr ma.Multiaddr) (models.IPAddressSlice, *string, *string, *int, error) {
	addrInfos, err := h.MMClient.MaddrInfo(ctx, maddr)
	if err != nil {
		log.WithError(err).Warnln("Could not get multi address information")
	}

	var countries []string
	var continents []string
	var asns []int
	var dbIPAddresses []*models.IPAddress

	for ipAddress, addrInfo := range addrInfos {
		if addrInfo.Country != "" {
			countries = append(countries, addrInfo.Country)
		}

		if addrInfo.Continent != "" {
			continents = append(continents, addrInfo.Continent)
		}

		if addrInfo.ASN != 0 {
			asns = append(asns, int(addrInfo.ASN))
		}

		dbIPAddress := &models.IPAddress{
			Address:   ipAddress,
			Country:   null.NewString(addrInfo.Country, addrInfo.Country != ""),
			Continent: null.NewString(addrInfo.Continent, addrInfo.Continent != ""),
			Asn:       null.NewInt(int(addrInfo.ASN), addrInfo.ASN != 0),
			IsPublic:  manet.IsPublicAddr(maddr),
		}

		err = dbIPAddress.Upsert(
			ctx,
			exec,
			true,
			[]string{models.IPAddressColumns.Address},
			boil.Whitelist(models.IPAddressColumns.UpdatedAt),
			boil.Infer(),
		)
		if err != nil {
			return nil, nil, nil, nil, errors.Wrap(err, "upsert ip address")
		}

		dbIPAddresses = append(dbIPAddresses, dbIPAddress)
	}

	return dbIPAddresses, util.UniqueStr(countries), util.UniqueStr(continents), util.UniqueInt(asns), nil
}

func mapNetDirection(conn network.Conn) string {
	switch conn.Stat().Direction {
	case network.DirInbound:
		return models.ConnectionDirectionINBOUND
	case network.DirOutbound:
		return models.ConnectionDirectionOUTBOUND
	default:
		return models.ConnectionDirectionUNKNOWN
	}
}

func deferRollback(txn *sql.Tx) {
	if err := txn.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
		log.WithError(err).Warnln("An error occurred when rolling back transaction")
	}
}

func isRelayedMaddr(maddr ma.Multiaddr) bool {
	_, err := maddr.ValueForProtocol(ma.P_CIRCUIT)
	if err == nil {
		return true
	} else if errors.Is(err, ma.ErrProtocolNotFound) {
		return false
	} else {
		log.WithError(err).WithField("maddr", maddr).Warnln("Unexpected error while parsing multi address")
		return false
	}
}

func doesSupportDCUtR(protocols []string) bool {
	for _, p := range protocols {
		if p == "/libp2p/dcutr" {
			return true
		}
	}
	return false
}
