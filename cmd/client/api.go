package main

import (
	"context"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/pb"
)

func (h *Host) RegisterHost(ctx context.Context) error {
	log.Infoln("Registering at Punchr server")

	bytesLocalPeerID, err := h.ID().Marshal()
	if err != nil {
		return errors.Wrap(err, "marshal peer id")
	}

	_, err = h.client.Register(ctx, &pb.RegisterRequest{
		ClientId: bytesLocalPeerID,
		// AgentVersion: *h.GetAgentVersion(h.ID()),
		AgentVersion: "punchr/go-client/0.1.0",
		Protocols:    h.GetProtocols(h.ID()),
	})
	if err != nil {
		return errors.Wrap(err, "register client")
	}
	return nil
}

func (h *Host) RequestAddrInfo(ctx context.Context) (*peer.AddrInfo, error) {
	log.Infoln("Requesting peer to hole punch from server...")

	bytesLocalPeerID, err := h.ID().Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal peer id")
	}

	res, err := h.client.GetAddrInfo(ctx, &pb.GetAddrInfoRequest{ClientId: bytesLocalPeerID})
	if err != nil {
		return nil, errors.Wrap(err, "get addr info RPC")
	}

	// If not remote ID is given the server does not have a peer to hole punch
	if res.GetRemoteId() == nil {
		return nil, nil
	}

	remoteID, err := peer.IDFromBytes(res.RemoteId)
	if err != nil {
		return nil, errors.Wrap(err, "peer ID from bytes")
	}

	maddrs := make([]multiaddr.Multiaddr, len(res.MultiAddresses))
	for i, maddrBytes := range res.MultiAddresses {
		maddr, err := multiaddr.NewMultiaddrBytes(maddrBytes)
		if err != nil {
			return nil, errors.Wrap(err, "multi address from bytes")
		}
		maddrs[i] = maddr
	}

	return &peer.AddrInfo{ID: remoteID, Addrs: maddrs}, nil
}

func (h *Host) TrackHolePunchResult(ctx context.Context, hps *HolePunchState) error {
	req, err := hps.ToProto(h.ID())
	if err != nil {
		return err
	}

	_, err = h.client.TrackHolePunch(ctx, req)
	return err
}
