package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	dbPeer, err := s.DBClient.UpsertPeer(ctx, s.DBClient, clientID, req.AgentVersion, req.Protocols)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{DbPeerId: &dbPeer.ID}, nil
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
		return nil, status.Error(codes.NotFound, "no peers to hole punch")
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
	start := time.Now()
	defer func() { log.WithField("dur", time.Since(start).String()).Infoln("Tracked result") }()

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

	dbAttempts := make([]*models.HolePunchAttempt, len(req.HolePunchAttempts))
	for i, hpa := range req.HolePunchAttempts {
		hpaOutcome := models.HolePunchAttemptOutcomeUNKNOWN
		switch *hpa.Outcome {
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_DIRECT_DIAL:
			hpaOutcome = models.HolePunchAttemptOutcomeDIRECT_DIAL
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_PROTOCOL_ERROR:
			hpaOutcome = models.HolePunchAttemptOutcomePROTOCOL_ERROR
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_CANCELLED:
			hpaOutcome = models.HolePunchAttemptOutcomeCANCELLED
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_TIMEOUT:
			hpaOutcome = models.HolePunchAttemptOutcomeTIMEOUT
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_FAILED:
			hpaOutcome = models.HolePunchAttemptOutcomeFAILED
		case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_SUCCESS:
			hpaOutcome = models.HolePunchAttemptOutcomeSUCCESS
		}

		if hpa.OpenedAt == nil {
			return nil, errors.Wrapf(err, "opened at in attempt %d is nil", i)
		}

		if hpa.StartedAt == nil {
			return nil, errors.Wrapf(err, "started at in attempt %d is nil", i)
		}

		if hpa.EndedAt == nil {
			return nil, errors.Wrapf(err, "ended at in attempt %d is nil", i)
		}

		if hpa.ElapsedTime == nil {
			return nil, errors.Wrapf(err, "elapsed time in attempt %d is nil", i)
		}

		startRtt := ""
		if hpa.StartRtt != nil {
			startRtt = fmt.Sprintf("%fs", *hpa.StartRtt)
		}

		dbAttempts[i] = &models.HolePunchAttempt{
			OpenedAt:        time.UnixMilli(int64(*hpa.OpenedAt)),
			StartedAt:       time.UnixMilli(int64(*hpa.StartedAt)),
			EndedAt:         time.UnixMilli(int64(*hpa.EndedAt)),
			StartRTT:        null.NewString(startRtt, startRtt != ""),
			ElapsedTime:     fmt.Sprintf("%fs", *hpa.ElapsedTime),
			Outcome:         hpaOutcome,
			Error:           null.NewString(hpa.GetError(), hpa.GetError() != ""),
			DirectDialError: null.NewString(hpa.GetDirectDialError(), hpa.GetDirectDialError() != ""),
		}
	}

	outcome := models.HolePunchOutcomeUNKNOWN
	switch *req.Outcome {
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_CONNECTION:
		outcome = models.HolePunchOutcomeNO_CONNECTION
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_STREAM:
		outcome = models.HolePunchOutcomeNO_STREAM
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_CANCELLED:
		outcome = models.HolePunchOutcomeCANCELLED
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_FAILED:
		outcome = models.HolePunchOutcomeFAILED
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_SUCCESS:
		outcome = models.HolePunchOutcomeSUCCESS
	}

	// Start a database transaction
	txn, err := s.DBClient.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin txn")
	}
	defer db.DeferRollback(txn)

	if req.ConnectStartedAt == nil {
		return nil, errors.Wrap(err, "connect started at is nil")
	}
	if req.ConnectEndedAt == nil {
		return nil, errors.Wrap(err, "connect ended at is nil")
	}
	if req.EndedAt == nil {
		return nil, errors.Wrap(err, "ended at is nil")
	}
	if req.HasDirectConns == nil {
		return nil, errors.Wrap(err, "has direct conns is nil")
	}

	hpr := &models.HolePunchResult{
		ClientID:         dbClientPeer.ID,
		RemoteID:         dbRemotePeer.ID,
		ConnectStartedAt: time.UnixMilli(int64(*req.ConnectStartedAt)),
		ConnectEndedAt:   time.UnixMilli(int64(*req.ConnectEndedAt)),
		HasDirectConns:   *req.HasDirectConns,
		Outcome:          outcome,
		EndedAt:          time.UnixMilli(int64(*req.EndedAt)),
	}

	if err = hpr.Insert(ctx, txn, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "insert hole punch result")
	}

	if err = hpr.AddHolePunchAttempts(ctx, txn, true, dbAttempts...); err != nil {
		return nil, errors.Wrap(err, "add attempts to hole punch result")
	}

	rmaddrs := make([]multiaddr.Multiaddr, len(req.RemoteMultiAddresses))
	for i, maddrBytes := range req.RemoteMultiAddresses {
		maddr, err := multiaddr.NewMultiaddrBytes(maddrBytes)
		if err != nil {
			return nil, errors.Wrap(err, "multi addr from bytes")
		}
		rmaddrs[i] = maddr
	}

	dbRMaddrs, err := s.DBClient.UpsertMultiAddresses(ctx, txn, rmaddrs)
	if err != nil {
		return nil, errors.Wrap(err, "upsert multi addresses")
	}

	omaddrs := make([]multiaddr.Multiaddr, len(req.OpenMultiAddresses))
	for i, maddrBytes := range req.OpenMultiAddresses {
		maddr, err := multiaddr.NewMultiaddrBytes(maddrBytes)
		if err != nil {
			return nil, errors.Wrap(err, "multi addr from bytes")
		}
		omaddrs[i] = maddr
	}

	dbOMaddrs, err := s.DBClient.UpsertMultiAddresses(ctx, txn, omaddrs)
	if err != nil {
		return nil, errors.Wrap(err, "upsert multi addresses")
	}

	for _, dbRMaddr := range dbRMaddrs {
		hprxma := models.HolePunchResultsXMultiAddress{
			HolePunchResultID: hpr.ID,
			MultiAddressID:    dbRMaddr.ID,
			Relationship:      models.HolePunchMultiAddressRelationshipREMOTE,
		}
		if err = hprxma.Insert(ctx, txn, boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "insert remote HolePunchResultsXMultiAddress")
		}
	}

	for _, dbOMaddr := range dbOMaddrs {
		hprxma := models.HolePunchResultsXMultiAddress{
			HolePunchResultID: hpr.ID,
			MultiAddressID:    dbOMaddr.ID,
			Relationship:      models.HolePunchMultiAddressRelationshipOPEN,
		}
		if err = hprxma.Insert(ctx, txn, boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "insert remote HolePunchResultsXMultiAddress")
		}
	}

	return &pb.TrackHolePunchResponse{}, txn.Commit()
}
