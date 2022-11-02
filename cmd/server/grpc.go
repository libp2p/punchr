package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/lib/pq"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/dennis-tra/punchr/pkg/pb"
)

type Server struct {
	pb.UnimplementedPunchrServiceServer
	DBClient    *db.Client
	apiKeyCache *lru.Cache
}

func (s Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	authID, err := s.checkApiKey(ctx, req.ApiKey)
	if err != nil {
		return nil, err
	}

	clientID, err := peer.IDFromBytes(req.ClientId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from client ID")
	}

	// Start a database transaction
	txn, err := s.DBClient.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin txn")
	}
	defer db.DeferRollback(txn)

	dbPeer, err := s.DBClient.UpsertPeer(ctx, txn, clientID, req.AgentVersion, req.Protocols)
	if err != nil {
		return nil, err
	}

	clientExists, err := models.Clients(
		models.ClientWhere.PeerID.EQ(dbPeer.ID),
		models.ClientWhere.AuthorizationID.EQ(authID),
	).Exists(ctx, txn)
	if err != nil {
		return nil, errors.Wrap(err, "checking authorization for api key")
	}

	if clientExists {
		return &pb.RegisterResponse{DbPeerId: &dbPeer.ID}, txn.Commit()
	}

	client := models.Client{
		PeerID:          dbPeer.ID,
		AuthorizationID: authID,
	}

	err = client.Insert(ctx, txn, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "inserting client")
	}

	return &pb.RegisterResponse{DbPeerId: &dbPeer.ID}, txn.Commit()
}

func (s Server) GetAddrInfo(ctx context.Context, req *pb.GetAddrInfoRequest) (*pb.GetAddrInfoResponse, error) {
	_, err := s.checkApiKey(ctx, req.ApiKey)
	if err != nil {
		return nil, err
	}

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

	if len(dbHosts) == 0 {
		return nil, fmt.Errorf("not registered, please restart the client")
	}

	dbHostIDs := make([]string, len(dbHosts))
	for i, dbHost := range dbHosts {
		dbHostIDs[i] = strconv.FormatInt(dbHost.ID, 10)
	}

	var resp *pb.GetAddrInfoResponse
	if rand.Float32() > 0.5 {
		resp, err = s.querySingleMaddr(ctx, dbHostIDs)
	} else {
		resp, err = s.queryAllMaddrs(ctx, dbHostIDs)
	}
	if err != nil {
		return nil, err
	}

	r := rand.Float32()
	if r < 0.15 {
		resp.Protocols = []int32{multiaddr.P_IP4, multiaddr.P_TCP}
	} else if r < 0.3 {
		resp.Protocols = []int32{multiaddr.P_IP4, multiaddr.P_QUIC}
	} else if r < 0.45 {
		resp.Protocols = []int32{multiaddr.P_IP6, multiaddr.P_TCP}
	} else if r < 0.6 {
		resp.Protocols = []int32{multiaddr.P_IP6, multiaddr.P_QUIC}
	} else {
		resp.Protocols = []int32{}
	}

	return resp, nil
}

// querySingleMaddr queries a single peer with a single multi address to hole punch as opposed to
// queryAllMaddrs that queries a single peer with all its multi addresses.
// TODO: Remove redundancy between querySingleMaddr and queryAllMaddrs
func (s Server) querySingleMaddr(ctx context.Context, dbHostIDs []string) (*pb.GetAddrInfoResponse, error) {
	query := `
-- select all peers that connected to the honeypot within the last 10 mins, listen on a relay address,
-- and support dcutr. Then also select all of their relay multi addresses. But only if:
--   1. the peer has not been hole punched more than 10 times in the last minute AND
--   2. the peer/maddr combination was not hole-punched by the same client in the last 30 mins.
-- then only return ONE random peer/maddr combination!
SELECT p.multi_hash, ma.maddr
FROM connection_events ce
         INNER JOIN connection_events_x_multi_addresses cexma ON ce.id = cexma.connection_event_id
         INNER JOIN multi_addresses ma ON cexma.multi_address_id = ma.id
         INNER JOIN peers p ON ce.remote_id = p.id
WHERE ma.is_relay = true
  AND ce.opened_at > NOW() - '10min'::INTERVAL -- peer connected to honeypot within last 10min
  AND ( -- prevent DoS. Exclude peers that were hole-punched >= 10 times in the last minute
          SELECT count(*)
          FROM hole_punch_results hpr
          WHERE hpr.remote_id = ce.remote_id
            AND hpr.created_at > NOW() - '1min'::INTERVAL
      ) < 10
  AND NOT EXISTS( -- Exclude peer/maddr combinations that were hole-punched from the same client within the last 30 min
        SELECT
        FROM hole_punch_results hpr
                 INNER JOIN hole_punch_results_x_multi_addresses hprxma ON hpr.id = hprxma.hole_punch_result_id
        WHERE hpr.remote_id = ce.remote_id
          AND hpr.local_id IN (%s)
          AND hprxma.multi_address_id = ma.id
          AND hprxma.relationship = 'INITIAL'
          AND hpr.created_at > NOW() - '30min'::INTERVAL
    )
ORDER BY random() -- get random peer/maddr combination
LIMIT 1
`
	start := time.Now()
	query = fmt.Sprintf(query, strings.Join(dbHostIDs, ","))
	rows, err := s.DBClient.QueryContext(ctx, query)
	if err != nil {
		allocationQueryDurationHistogram.WithLabelValues("single", "false").Observe(time.Since(start).Seconds())
		return nil, errors.Wrap(err, "query addr infos")
	}
	allocationQueryDurationHistogram.WithLabelValues("single", "true").Observe(time.Since(start).Seconds())

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
		return nil, errors.Wrap(err, "decode remote multi hash")
	}

	remoteIDBytes, err := remoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal remote ID to bytes")
	}

	maddr, err := multiaddr.NewMultiaddr(remoteMaddrStr)
	if err != nil {
		return nil, errors.Wrapf(err, "parse multi address %s", remoteMaddrStr)
	}

	return &pb.GetAddrInfoResponse{
		RemoteId:       remoteIDBytes,
		MultiAddresses: [][]byte{maddr.Bytes()},
	}, nil
}

// queryAllMaddrs queries a single peer from the database and all its multi addresses to hole punch
// as opposed to querySingleMaddr that quries a single peer with a single multi address.
// TODO: Remove redundancy between querySingleMaddr and queryAllMaddrs
func (s Server) queryAllMaddrs(ctx context.Context, dbHostIDs []string) (*pb.GetAddrInfoResponse, error) {
	query := `
-- select all peers that connected to the honeypot within the last 10 mins, listen on a relay address,
-- and support dcutr. Then also select all of their relay multi addresses. But only if:
--   1. the peer has not been hole punched more than 10 times in the last minute AND
--   2. the peer/maddr combination was not hole-punched by the same client in the last 30 mins.
-- then only return ONE random peer/maddr combination!
SELECT p.multi_hash, array_agg(DISTINCT ma.maddr)
FROM connection_events ce
         INNER JOIN connection_events_x_multi_addresses cexma ON ce.id = cexma.connection_event_id
         INNER JOIN multi_addresses ma ON cexma.multi_address_id = ma.id
         INNER JOIN peers p ON ce.remote_id = p.id
WHERE ma.is_relay = true
  AND ce.opened_at > NOW() - '10min'::INTERVAL -- peer connected to honeypot within last 10min
  AND ( -- prevent DoS. Exclude peers that were hole-punched >= 10 times in the last minute
          SELECT count(*)
          FROM hole_punch_results hpr
          WHERE hpr.remote_id = ce.remote_id
            AND hpr.created_at > NOW() - '1min'::INTERVAL
      ) < 10
  AND NOT EXISTS(
        SELECT
        FROM hole_punch_results hpr
                 INNER JOIN hole_punch_results_x_multi_addresses hprxma ON hpr.id = hprxma.hole_punch_result_id
        WHERE hpr.remote_id = ce.remote_id
          AND hpr.local_id IN (%s)
          AND hprxma.multi_address_id = ma.id
          AND hprxma.relationship = 'INITIAL'
          AND hpr.created_at > NOW() - '10min'::INTERVAL
    )
GROUP BY p.id
ORDER BY random() -- get random peer/maddr combination
LIMIT 1
`
	start := time.Now()
	query = fmt.Sprintf(query, strings.Join(dbHostIDs, ","))
	rows, err := s.DBClient.QueryContext(ctx, query)
	if err != nil {
		allocationQueryDurationHistogram.WithLabelValues("all", "false").Observe(time.Since(start).Seconds())
		return nil, errors.Wrap(err, "query addr infos")
	}
	allocationQueryDurationHistogram.WithLabelValues("all", "true").Observe(time.Since(start).Seconds())

	defer func() {
		if err := rows.Close(); err != nil {
			log.WithError(err).Warnln("Could not close database query")
		}
	}()

	if !rows.Next() {
		return nil, status.Error(codes.NotFound, "no peers to hole punch")
	}

	var remoteMultiHash string
	var remoteMaddrStrs []string
	if err = rows.Scan(&remoteMultiHash, pq.Array(&remoteMaddrStrs)); err != nil {
		return nil, errors.Wrap(err, "map query results")
	}

	remoteID, err := peer.Decode(remoteMultiHash)
	if err != nil {
		return nil, errors.Wrap(err, "decode remote multi hash")
	}

	remoteIDBytes, err := remoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal remote ID to bytes")
	}

	maddrBytes := make([][]byte, len(remoteMaddrStrs))
	for i, remoteMaddrStr := range remoteMaddrStrs {
		maddr, err := multiaddr.NewMultiaddr(remoteMaddrStr)
		if err != nil {
			return nil, errors.Wrapf(err, "parse multi address %s", remoteMaddrStr)
		}
		maddrBytes[i] = maddr.Bytes()
	}

	return &pb.GetAddrInfoResponse{
		RemoteId:       remoteIDBytes,
		MultiAddresses: maddrBytes,
	}, nil
}

func (s Server) TrackHolePunch(ctx context.Context, req *pb.TrackHolePunchRequest) (*pb.TrackHolePunchResponse, error) {
	_, err := s.checkApiKey(ctx, req.ApiKey)
	if err != nil {
		return nil, err
	}

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
	if len(req.ListenMultiAddresses) == 0 {
		return nil, errors.Wrap(err, "no listen multi addresses given")
	}

	start := time.Now()
	defer func() { log.WithField("dur", time.Since(start).String()).Infoln("Tracked result") }()

	clientID, err := peer.IDFromBytes(req.ClientId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from client ID")
	}

	dbLocalPeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(clientID.String())).One(ctx, s.DBClient)
	if err != nil {
		return nil, errors.Wrap(err, "get client peer from db")
	}

	clientExists, err := models.Clients(models.ClientWhere.PeerID.EQ(dbLocalPeer.ID)).Exists(ctx, s.DBClient)
	if !clientExists {
		return nil, fmt.Errorf("client not registered")
	}

	remoteID, err := peer.IDFromBytes(req.RemoteId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from remote ID")
	}

	dbRemotePeer, err := models.Peers(models.PeerWhere.MultiHash.EQ(remoteID.String())).One(ctx, s.DBClient)
	if err != nil {
		return nil, errors.Wrap(err, "get remote peer from db")
	}

	// Start a database transaction
	txn, err := s.DBClient.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin txn")
	}
	defer db.DeferRollback(txn)

	dbhpas := make([]*models.HolePunchAttempt, len(req.HolePunchAttempts))
	for i, hpa := range req.HolePunchAttempts {
		if hpa.OpenedAt == nil {
			return nil, errors.Wrapf(err, "opened at in attempt %d is nil", i)
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

		dbhpas[i] = &models.HolePunchAttempt{
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

	lmaddrs := make([]multiaddr.Multiaddr, len(req.ListenMultiAddresses))
	for i, maddrBytes := range req.ListenMultiAddresses {
		maddr, err := multiaddr.NewMultiaddrBytes(maddrBytes)
		if err != nil {
			return nil, errors.Wrap(err, "multi addr from bytes")
		}
		lmaddrs[i] = maddr
	}

	dbLMaddrs, err := s.DBClient.UpsertMultiAddresses(ctx, txn, lmaddrs)
	if err != nil {
		return nil, errors.Wrap(err, "upsert listen multi addresses")
	}

	maddrSetID, err := s.DBClient.UpsertMultiAddressesSet(ctx, txn, dbLMaddrs)
	if err != nil {
		return nil, errors.Wrap(err, "upsert multi addresses set")
	}

	filters := make([]int64, len(req.Protocols))
	for i, p := range req.Protocols {
		filters[i] = int64(p)
	}

	hpr := &models.HolePunchResult{
		LocalID:                   dbLocalPeer.ID,
		ListenMultiAddressesSetID: maddrSetID,
		RemoteID:                  dbRemotePeer.ID,
		ConnectStartedAt:          time.Unix(0, int64(*req.ConnectStartedAt)),
		ConnectEndedAt:            time.Unix(0, int64(*req.ConnectEndedAt)),
		HasDirectConns:            *req.HasDirectConns,
		ProtocolFilters:           filters,
		Outcome:                   s.mapHolePunchOutcome(req),
		Error:                     null.StringFromPtr(req.Error),
		EndedAt:                   time.Unix(0, int64(*req.EndedAt)),
	}

	if err = hpr.Insert(ctx, txn, boil.Infer()); err != nil {
		return nil, errors.Wrap(err, "insert hole punch result")
	}

	if err = hpr.AddHolePunchAttempts(ctx, txn, true, dbhpas...); err != nil {
		return nil, errors.Wrap(err, "add attempts to hole punch result")
	}

	for i, hpa := range req.HolePunchAttempts {
		hpamaddrs := make([]multiaddr.Multiaddr, len(hpa.MultiAddresses))
		for j, maddrBytes := range hpa.MultiAddresses {
			maddr, err := multiaddr.NewMultiaddrBytes(maddrBytes)
			if err != nil {
				return nil, errors.Wrap(err, "hole punch attempt multi addr from bytes")
			}
			hpamaddrs[j] = maddr
		}

		dbHPAMaddrs, err := s.DBClient.UpsertMultiAddresses(ctx, txn, hpamaddrs)
		if err != nil {
			return nil, errors.Wrap(err, "hole punch attempt multi addresses")
		}

		if err = dbhpas[i].SetMultiAddresses(ctx, txn, false, dbHPAMaddrs...); err != nil {
			return nil, errors.Wrap(err, "upsert hole punch attempt multi addresses")
		}
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
		return nil, errors.Wrap(err, "upsert open multi addresses")
	}

	for _, dbOMaddr := range dbOMaddrs {
		hprxma := models.HolePunchResultsXMultiAddress{
			HolePunchResultID: hpr.ID,
			MultiAddressID:    dbOMaddr.ID,
			Relationship:      models.HolePunchMultiAddressRelationshipFINAL,
		}
		if err = hprxma.Insert(ctx, txn, boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "insert open HolePunchResultsXMultiAddress")
		}
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
		return nil, errors.Wrap(err, "upsert remote multi addresses")
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

	if req.NetworkInformation != nil {
		dbNetInfo := models.NetworkInformation{
			PeerID:            dbLocalPeer.ID,
			SupportsIpv6:      null.BoolFromPtr(req.NetworkInformation.SupportsIpv6),
			SupportsIpv6Error: null.StringFromPtr(req.NetworkInformation.SupportsIpv6Error),
			RouterHTML:        null.StringFromPtr(req.NetworkInformation.RouterLoginHtml),
			RouterHTMLError:   null.StringFromPtr(req.NetworkInformation.RouterLoginHtmlError),
		}

		if err = dbNetInfo.SetPeer(ctx, txn, false, dbLocalPeer); err != nil {
			return nil, errors.Wrap(err, "set client peer of router information")
		}

		if err = dbNetInfo.Insert(ctx, txn, boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "insert router information")
		}
	}

	for _, lm := range req.LatencyMeasurements {

		maddr, err := multiaddr.NewMultiaddrBytes(lm.MultiAddress)
		if err != nil {
			return nil, errors.Wrap(err, "multi addr from bytes")
		}

		dbMaddr, err := s.DBClient.UpsertMultiAddress(ctx, txn, maddr)
		if err != nil {
			return nil, errors.Wrap(err, "upsert latency measurement multi address")
		}

		lmRemoteID, err := peer.IDFromBytes(lm.RemoteId)
		if err != nil {
			return nil, errors.Wrap(err, "peer ID from remote ID")
		}
		dbLmRemotePeer, err := s.DBClient.UpsertPeer(ctx, txn, lmRemoteID, lm.AgentVersion, lm.Protocols)
		if err != nil {
			return nil, errors.Wrap(err, "could not upsert latency measurement peer")
		}

		dbMtype, err := mapMeasurementType(lm.Mtype)
		if err != nil {
			return nil, errors.Wrap(err, "map measurement type")
		}

		dbRtts := make(types.Float64Array, len(lm.Rtts))
		for i, rtt := range lm.Rtts {
			dbRtts[i] = float64(rtt)
		}

		dblm := models.LatencyMeasurement{
			RemoteID:          dbLmRemotePeer.ID,
			HolePunchResultID: hpr.ID,
			MultiAddressID:    dbMaddr.ID,
			Mtype:             dbMtype,
			RTTS:              dbRtts,
			RTTAvg:            stat.Mean(dbRtts, nil),
			RTTMax:            floats.Max(dbRtts),
			RTTMin:            floats.Min(dbRtts),
			RTTSTD:            stat.StdDev(dbRtts, nil),
		}

		if err := dblm.Insert(ctx, txn, boil.Infer()); err != nil {
			return nil, errors.Wrap(err, "insert latency measurement")
		}
	}

	return &pb.TrackHolePunchResponse{}, txn.Commit()
}

func mapMeasurementType(mtype *pb.LatencyMeasurementType) (string, error) {
	if mtype == nil {
		return "", fmt.Errorf("latency measurement type is nil")
	}

	switch *mtype {
	case pb.LatencyMeasurementType_TO_RELAY:
		return models.LatencyMeasurementTypeTO_RELAY, nil
	case pb.LatencyMeasurementType_TO_REMOTE_THROUGH_RELAY:
		return models.LatencyMeasurementTypeTO_REMOTE_THROUGH_RELAY, nil
	case pb.LatencyMeasurementType_TO_REMOTE_AFTER_HOLE_PUNCH:
		return models.LatencyMeasurementTypeTO_REMOTE_AFTER_HOLEPUNCH, nil
	default:
		return "", fmt.Errorf("unsupportet latency measurement type: %s", mtype)
	}
}

func (s Server) checkApiKey(ctx context.Context, apiKey *string) (int, error) {
	if apiKey == nil || *apiKey == "" {
		return 0, fmt.Errorf("API key is missing")
	}

	iAuthID, found := s.apiKeyCache.Get(*apiKey)
	authID, ok := iAuthID.(int)
	if found && !ok {
		s.apiKeyCache.Remove(*apiKey)
	}

	if found {
		return authID, nil
	}

	dbAuthorization, err := s.DBClient.GetAuthorization(ctx, s.DBClient, *apiKey)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("unauthorized")
	} else if err != nil {
		return 0, errors.Wrap(err, "checking authorization for api key")
	}

	s.apiKeyCache.Add(*apiKey, dbAuthorization.ID)

	return dbAuthorization.ID, nil
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
