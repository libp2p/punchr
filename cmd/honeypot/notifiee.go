package main

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/util"
)

func (h *Host) Listen(network.Network, ma.Multiaddr)         {}
func (h *Host) ListenClose(network.Network, ma.Multiaddr)    {}
func (h *Host) OpenedStream(network.Network, network.Stream) {}
func (h *Host) ClosedStream(network.Network, network.Stream) {}
func (h *Host) Disconnected(network.Network, network.Conn)   {}
func (h *Host) Connected(_ network.Network, conn network.Conn) {
	if conn.Stat().Direction != network.DirInbound {
		return
	}

	if err := h.handleNewConnection(conn); err != nil {
		log.WithError(err).Warnln("An error occurred while handling the new connection")
		handledConns.With(prometheus.Labels{"status": "error"}).Inc()
	}
}

// handleNewConnection handles the new connection establishment.
// We can do expensive things here as it's called within a go-routine by swarm.
func (h *Host) handleNewConnection(conn network.Conn) error {
	defer log.WithFields(log.Fields{"remoteID": util.FmtPeerID(conn.RemotePeer())}).Infoln("Handled connection")

	// Wait for the "identify" protocol to complete
	if err := h.IdentifyWait(h.ctx, conn.RemotePeer()); err != nil {
		return errors.Wrap(err, "identify wait")
	}

	// Grab all peer infos from the peer store
	agentVersion := h.GetAgentVersion(conn.RemotePeer())
	protocols := h.GetProtocols(conn.RemotePeer())
	maddrs := h.GetMultiAddresses(conn.RemotePeer())

	if !util.SupportDCUtR(protocols) {
		// don't save peer as it doesn't support DCUtR
		log.Debugln("Incoming connection, peer does not support DCUtR")
		return nil
	}

	// Check if the remote peer only has relay addresses
	for _, maddr := range maddrs {
		if !manet.IsPrivateAddr(maddr) && !util.IsRelayedMaddr(maddr) {
			log.Debugln("Incoming connection, has a public non-relay address")
			return nil
		}
	}

	// It can happen that the `conn.RemoteMultiaddr()` is not part of the peer store maddrs.
	found := false
	for _, maddr := range maddrs {
		if maddr.Equal(conn.RemoteMultiaddr()) {
			found = true
			break
		}
	}
	if !found {
		maddrs = append(maddrs, conn.RemoteMultiaddr())
	}

	// Start a database transaction
	txn, err := h.DBClient.BeginTx(h.ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin txn")
	}
	defer db.DeferRollback(txn)

	// Save all multi addresses of the connected peer
	dbMaddrs, err := h.DBClient.UpsertMultiAddresses(h.ctx, txn, maddrs)
	if err != nil {
		return errors.Wrap(err, "upsert multi addresses")
	}

	// Save the connected peer
	dbPeer, err := h.DBClient.UpsertPeer(h.ctx, txn, conn.RemotePeer(), agentVersion, protocols)
	if err != nil {
		return errors.Wrap(err, "upsert peer")
	}

	// Determine if there is at least one relay multi address and determine the database
	// TODO: redundant with above
	var connMaddrID int64
	var advertisedMaddrs []*models.MultiAddress
	for _, dbMaddr := range dbMaddrs {
		if dbMaddr.Maddr == conn.RemoteMultiaddr().String() {
			connMaddrID = dbMaddr.ID
		} else {
			advertisedMaddrs = append(advertisedMaddrs, dbMaddr)
		}
	}

	// Save this connection event
	dbConnEvt := &models.ConnectionEvent{
		LocalID:                  h.DBPeer.ID,
		RemoteID:                 dbPeer.ID,
		ConnectionMultiAddressID: connMaddrID,
		OpenedAt:                 conn.Stat().Opened,
	}
	if err = dbConnEvt.Insert(h.ctx, txn, boil.Infer()); err != nil {
		return errors.Wrap(err, "insert connection event")
	}

	// Associate multi addresses with this connection event
	if err = dbConnEvt.SetMultiAddresses(h.ctx, txn, false, advertisedMaddrs...); err != nil {
		return errors.Wrap(err, "set connection event multi addresses")
	}

	if err = txn.Commit(); err != nil {
		return errors.Wrap(err, "commit txn")
	}

	handledConns.With(prometheus.Labels{"status": "ok"}).Inc()

	return nil
}

// IdentifyWait waits for the "identify" protocol to complete.
func (h *Host) IdentifyWait(ctx context.Context, pid peer.ID) error {
	eventTypes := []interface{}{
		new(event.EvtPeerIdentificationCompleted),
		new(event.EvtPeerIdentificationFailed),
	}

	sub, err := h.EventBus().Subscribe(eventTypes)
	if err != nil {
		return errors.Wrap(err, "subscribing to event bus")
	}
	defer func() {
		if err := sub.Close(); err != nil {
			log.WithError(err).Warnln("Error closing event bus subscription")
		}
	}()

	for {
		var evtPeer peer.ID
		var err error
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-sub.Out():
			switch evt := e.(type) {
			case event.EvtPeerIdentificationCompleted:
				evtPeer = evt.Peer
				err = nil
			case event.EvtPeerIdentificationFailed:
				evtPeer = evt.Peer
				err = evt.Reason
			default:
				panic(fmt.Sprintf("unexpected event type %T", e))
			}
		}
		if evtPeer != pid {
			continue
		}
		return err
	}
}
