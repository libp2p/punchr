package main

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/pb"
)

type Server struct {
	pb.UnimplementedPunchrServiceServer
	DBClient *db.Client
}

func (s Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	clientID, err := peer.IDFromBytes(req.ClientId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from client ID")
	}

	_, err = s.DBClient.UpsertPeer(ctx, s.DBClient, clientID, &req.AgentVersion, req.Protocols)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{}, nil
}

func (s Server) GetAddrInfo(ctx context.Context, req *pb.GetAddrInfoRequest) (*pb.GetAddrInfoResponse, error) {
	clientID, err := peer.IDFromBytes(req.ClientId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from client ID")
	}

	dbClientPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(clientID.String())).One(ctx, s.DBClient)
	if err != nil {
		return nil, errors.Wrap(err, "get client peer from db")
	}

	query := `
SELECT p.multi_hash, ma.maddr
FROM connection_events ce
         INNER JOIN connection_events_x_multi_addresses cexma ON ce.id = cexma.connection_event_id
         INNER JOIN multi_addresses ma ON cexma.multi_address_id = ma.id
         INNER JOIN peers p ON ce.remote_id = p.id
WHERE ce.listens_on_relay_multi_address = true
  AND ce.supports_dcutr = true
  AND ce.opened_at > NOW() - '10min'::INTERVAL
  AND ma.is_relay = true
  AND NOT EXISTS(
        SELECT
        FROM hole_punch_results hpr
                 INNER JOIN hole_punch_results_x_multi_addresses hprxma ON hpr.id = hprxma.hole_punch_result_id
        WHERE hpr.remote_id =  ce.remote_id
          AND hpr.client_id = $1
          AND hprxma.multi_address_id =ma.id
          AND hpr.created_at > NOW() - '10min'::INTERVAL
    )
LIMIT 1
`
	rows, err := s.DBClient.QueryContext(ctx, query, dbClientPeer.ID)
	if err != nil {
		return nil, errors.Wrap(err, "query addr infos")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.WithError(err).Warnln("Could not close database query")
		}
	}()

	if !rows.Next() {
		return &pb.GetAddrInfoResponse{}, nil
	}

	var remoteMultiHash string
	var remoteMaddrStr string
	if err = rows.Scan(&remoteMultiHash, &remoteMaddrStr); err != nil {
		return nil, errors.Wrap(err, "map query results")
	}
	remoteID, err := peer.Decode(remoteMultiHash)
	if err != nil {
		return nil, errors.Wrap(err, "query addr infos")
	}

	remoteIDBytes, err := remoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal remote ID to bytes")
	}

	maddr, err := multiaddr.NewMultiaddr(remoteMaddrStr)
	if err != nil {
		return nil, errors.Wrapf(err, "parse multi address %s", remoteMaddrStr)
	}

	resp := &pb.GetAddrInfoResponse{
		RemoteId:       remoteIDBytes,
		MultiAddresses: [][]byte{maddr.Bytes()},
	}

	return resp, nil
}

func (s Server) TrackHolePunch(ctx context.Context, req *pb.TrackHolePunchRequest) (*pb.TrackHolePunchResponse, error) {
	clientID, err := peer.IDFromBytes(req.ClientId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from client ID")
	}

	remoteID, err := peer.IDFromBytes(req.RemoteId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from remote ID")
	}

	dbClientPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(clientID.String())).One(ctx, s.DBClient)
	if err != nil {
		return nil, errors.Wrap(err, "get client peer from db")
	}

	dbRemotePeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(remoteID.String())).One(ctx, s.DBClient)
	if err != nil {
		return nil, errors.Wrap(err, "get remote peer from db")
	}

	endReason := models.HolePunchEndReasonUNKNOWN
	switch req.EndReason {
	case pb.HolePunchEndReason_PROTOCOL_ERROR:
		endReason = models.HolePunchEndReasonPROTOCOL_ERROR
	case pb.HolePunchEndReason_DIRECT_DIAL:
		endReason = models.HolePunchEndReasonDIRECT_DIAL
	case pb.HolePunchEndReason_HOLE_PUNCH:
		endReason = models.HolePunchEndReasonHOLE_PUNCH
	case pb.HolePunchEndReason_NO_CONNECTION:
		endReason = models.HolePunchEndReasonNO_CONNECTION
	}

	// Start a database transaction
	txn, err := s.DBClient.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin txn")
	}
	defer db.DeferRollback(txn)

	hpr := &models.HolePunchResult{
		ClientID:        dbClientPeer.ID,
		RemoteID:        dbRemotePeer.ID,
		StartRTT:        fmt.Sprintf("%fs", req.StartRtt),
		ElapsedTime:     fmt.Sprintf("%fs", req.ElapsedTime),
		EndReason:       endReason,
		Attempts:        int16(req.Attempts),
		Success:         req.Success,
		Error:           null.NewString(req.Error, req.Error != ""),
		DirectDialError: null.NewString(req.DirectDialError, req.DirectDialError != ""),
	}

	if err = hpr.Insert(ctx, txn, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "insert hole punch result")
	}

	maddrs := make([]multiaddr.Multiaddr, len(req.RemoteMultiAddresses))
	for i, maddrBytes := range req.RemoteMultiAddresses {
		maddr, err := multiaddr.NewMultiaddrBytes(maddrBytes)
		if err != nil {
			return nil, errors.Wrap(err, "multi addr from bytes")
		}
		maddrs[i] = maddr
	}

	dbMaddrs, err := s.DBClient.UpsertMultiAddresses(ctx, txn, maddrs)
	if err != nil {
		return nil, errors.Wrap(err, "upsert multi addresses")
	}

	if err = hpr.SetMultiAddresses(ctx, txn, false, dbMaddrs...); err != nil {
		return nil, errors.Wrap(err, "set multi addresses to hole punch result")
	}

	return &pb.TrackHolePunchResponse{}, txn.Commit()
}
