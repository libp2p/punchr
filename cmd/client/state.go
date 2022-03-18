package main

import (
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/dennis-tra/punchr/pkg/pb"
)

type HolePunchState struct {
	RemoteID            peer.ID
	ConnectionStartedAt time.Time
	RemoteMaddrs        []multiaddr.Multiaddr
	StartRTT            time.Duration
	ElapsedTime         time.Duration
	EndReason           pb.HolePunchEndReason
	Attempts            int
	Success             bool
	Error               string
	DirectDialError     string
	holePunchStarted    chan struct{}
	holePunchFinished   chan struct{}
}

func (h *Host) NewHolePunchState(remoteID peer.ID, maddrs []multiaddr.Multiaddr) (<-chan struct{}, <-chan struct{}) {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	h.hpStates[remoteID] = &HolePunchState{
		RemoteID:            remoteID,
		ConnectionStartedAt: time.Now(),
		RemoteMaddrs:        maddrs,
		holePunchStarted:    make(chan struct{}),
		holePunchFinished:   make(chan struct{}),
	}

	return h.hpStates[remoteID].holePunchStarted, h.hpStates[remoteID].holePunchFinished
}

func (h *Host) DeleteHolePunchState(remoteID peer.ID) *HolePunchState {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	hpState := h.hpStates[remoteID]
	delete(h.hpStates, remoteID)

	return hpState
}

func (hps HolePunchState) ToProto(peerID peer.ID) (*pb.TrackHolePunchRequest, error) {
	localID, err := peerID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal local peer id")
	}

	remoteID, err := hps.RemoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal remote peer id")
	}

	maddrsBytes := make([][]byte, len(hps.RemoteMaddrs))
	for i, maddr := range hps.RemoteMaddrs {
		maddrsBytes[i] = maddr.Bytes()
	}

	return &pb.TrackHolePunchRequest{
		ClientId:             localID,
		RemoteId:             remoteID,
		Success:              hps.Success,
		StartedAt:            hps.ConnectionStartedAt.UnixMilli(),
		RemoteMultiAddresses: maddrsBytes,
		Attempts:             int32(hps.Attempts),
		Error:                hps.Error,
		DirectDialError:      hps.DirectDialError,
		StartRtt:             float32(hps.StartRTT.Seconds()),
		ElapsedTime:          float32(hps.ElapsedTime.Seconds()),
		EndReason:            hps.EndReason,
	}, nil
}
