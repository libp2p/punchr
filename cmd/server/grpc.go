package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/pb"
)

var allocationQueryDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "db_allocation_query_duration_seconds",
	Help: "Histogram of database query times for client allocations",
}, []string{"success"})

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
	hostIDs := make([]string, len(req.AllHostIds))
	for i, bytesHostID := range req.AllHostIds {
		hostID, err := peer.IDFromBytes(bytesHostID)
		if err != nil {
			return nil, errors.Wrap(err, "peer ID from client ID")
		}
		hostIDs[i] = hostID.String()
	}

	dbHosts, err := models.Peers(models.PeerWhere.MultiHash.IN(hostIDs)).All(ctx, s.DBClient)
	if err != nil {
		return nil, errors.Wrap(err, "get client peer from db")
	}

	dbHostIDs := make([]string, len(dbHosts))
	for i, dbHost := range dbHosts {
		dbHostIDs[i] = strconv.FormatInt(dbHost.ID, 10)
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
          AND hpr.client_id IN (%s)
          AND hprxma.multi_address_id = ma.id
          AND hprxma.relationship = 'INITIAL'
          AND hpr.created_at > NOW() - '10min'::INTERVAL
    )
LIMIT 1
`
	start := time.Now()
	query = fmt.Sprintf(query, strings.Join(dbHostIDs, ","))
	rows, err := s.DBClient.QueryContext(ctx, query)
	if err != nil {
		allocationQueryDurationHistogram.WithLabelValues("false").Observe(time.Since(start).Seconds())
		return nil, errors.Wrap(err, "query addr infos")
	}
	allocationQueryDurationHistogram.WithLabelValues("true").Observe(time.Since(start).Seconds())

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

		var startedAt *time.Time
		if hpa.StartedAt != nil {
			t := time.Unix(0, int64(*hpa.StartedAt))
			startedAt = &t
		}

		dbAttempts[i] = &models.HolePunchAttempt{
			OpenedAt:        time.Unix(0, int64(*hpa.OpenedAt)),
			StartedAt:       null.TimeFromPtr(startedAt),
			EndedAt:         time.Unix(0, int64(*hpa.EndedAt)),
			StartRTT:        null.NewString(startRtt, startRtt != ""),
			ElapsedTime:     fmt.Sprintf("%fs", *hpa.ElapsedTime),
			Outcome:         s.mapHolePunchAttemptOutcome(hpa),
			Error:           null.NewString(hpa.GetError(), hpa.GetError() != ""),
			DirectDialError: null.NewString(hpa.GetDirectDialError(), hpa.GetDirectDialError() != ""),
		}
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
		ConnectStartedAt: time.Unix(0, int64(*req.ConnectStartedAt)),
		ConnectEndedAt:   time.Unix(0, int64(*req.ConnectEndedAt)),
		HasDirectConns:   *req.HasDirectConns,
		Outcome:          s.mapHolePunchOutcome(req),
		Error:            null.StringFromPtr(req.Error),
		EndedAt:          time.Unix(0, int64(*req.EndedAt)),
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
			Relationship:      models.HolePunchMultiAddressRelationshipINITIAL,
		}
		if err = hprxma.Insert(ctx, txn, boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "insert remote HolePunchResultsXMultiAddress")
		}
	}

	for _, dbOMaddr := range dbOMaddrs {
		hprxma := models.HolePunchResultsXMultiAddress{
			HolePunchResultID: hpr.ID,
			MultiAddressID:    dbOMaddr.ID,
			Relationship:      models.HolePunchMultiAddressRelationshipFINAL,
		}
		if err = hprxma.Insert(ctx, txn, boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "insert remote HolePunchResultsXMultiAddress")
		}
	}

	return &pb.TrackHolePunchResponse{}, txn.Commit()
}

func (s Server) mapHolePunchAttemptOutcome(hpa *pb.HolePunchAttempt) string {
	switch *hpa.Outcome {
	case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_DIRECT_DIAL:
		return models.HolePunchAttemptOutcomeDIRECT_DIAL
	case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_PROTOCOL_ERROR:
		return models.HolePunchAttemptOutcomePROTOCOL_ERROR
	case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_CANCELLED:
		return models.HolePunchAttemptOutcomeCANCELLED
	case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_TIMEOUT:
		return models.HolePunchAttemptOutcomeTIMEOUT
	case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_FAILED:
		return models.HolePunchAttemptOutcomeFAILED
	case pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_SUCCESS:
		return models.HolePunchAttemptOutcomeSUCCESS
	default:
		return models.HolePunchAttemptOutcomeUNKNOWN
	}
}

func (s Server) mapHolePunchOutcome(req *pb.TrackHolePunchRequest) string {
	switch *req.Outcome {
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_CONNECTION:
		return models.HolePunchOutcomeNO_CONNECTION
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_NO_STREAM:
		return models.HolePunchOutcomeNO_STREAM
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_CANCELLED:
		return models.HolePunchOutcomeCANCELLED
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_CONNECTION_REVERSED:
		return models.HolePunchOutcomeCONNECTION_REVERSED
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_FAILED:
		return models.HolePunchOutcomeFAILED
	case pb.HolePunchOutcome_HOLE_PUNCH_OUTCOME_SUCCESS:
		return models.HolePunchOutcomeSUCCESS
	default:
		return models.HolePunchOutcomeUNKNOWN
	}
}
