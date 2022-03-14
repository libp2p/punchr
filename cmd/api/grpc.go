package main

import (
	"context"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/dennis-tra/punchr/pkg/models"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/dennis-tra/punchr/pkg/db"
	"github.com/dennis-tra/punchr/pkg/pb"
)

type Server struct {
	pb.UnimplementedPunchrServiceServer
	DBClient *db.Client
}

func (s Server) GetPeerInfo(ctx context.Context, req *pb.GetAddrInfoRequest) (*pb.GetAddrInfoResponse, error) {
	dbPeers, _ := models.Peers(
		models.PeerWhere.SupportsDcutr.EQ(true),
		qm.OrderBy(models.PeerColumns.CreatedAt+" DESC"),
		qm.Limit(100),
	).All(ctx, s.DBClient)

	peerID, err := peer.IDFromString(dbPeers[0].MultiHash)
	if err != nil {
		return nil, err
	}
	bytesPeerID, err := peerID.Marshal()
	if err != nil {
		return nil, err
	}

	bytesMultiAddress := addrInfo.Addrs[0].Bytes()

	return &pb.GetAddrInfoResponse{
		PeerId:       bytesPeerID,
		MultiAddress: bytesMultiAddress,
	}, nil
}

func (s Server) TrackHolePunch(context.Context, *pb.TrackHolePunchRequest) (*pb.TrackHolePunchResponse, error) {
	return nil, nil
}
