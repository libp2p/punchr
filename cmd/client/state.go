package main

import (
	"time"

	"github.com/dennis-tra/punchr/pkg/util"
	log "github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"

	"github.com/dennis-tra/punchr/pkg/pb"
)

type HolePunchState struct {
	// The host that established the connection to the remote peer via a relay
	HostID peer.ID

	// The remote peer and its multi addresses - usually relayed ones.
	RemoteID     peer.ID
	RemoteMaddrs []multiaddr.Multiaddr

	// Start and end times for the establishment of the relayed connection
	ConnectStartedAt time.Time
	ConnectEndedAt   time.Time

	// Information about each individual hole punch attempt
	HolePunchAttempts []*HolePunchAttempt

	// Multi addresses of the open connections after the hole punch
	OpenMaddrs     []multiaddr.Multiaddr
	HasDirectConns bool

	Error   string
	Outcome pb.HolePunchOutcome
	EndedAt time.Time
}

func NewHolePunchState(hostID peer.ID, remoteID peer.ID, maddrs []multiaddr.Multiaddr) *HolePunchState {
	return &HolePunchState{
		HostID:            hostID,
		RemoteID:          remoteID,
		RemoteMaddrs:      maddrs,
		HolePunchAttempts: []*HolePunchAttempt{},
		OpenMaddrs:        []multiaddr.Multiaddr{},
	}
}

func (hps HolePunchState) logEntry(remoteID peer.ID) *log.Entry {
	return log.WithFields(log.Fields{
		"remoteID": util.FmtPeerID(remoteID),
		"hostID":   util.FmtPeerID(hps.HostID),
	})
}

func (hps HolePunchState) ToProto() (*pb.TrackHolePunchRequest, error) {
	localID, err := hps.HostID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal local peer id")
	}

	remoteID, err := hps.RemoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal remote peer id")
	}

	rMaddrBytes := make([][]byte, len(hps.RemoteMaddrs))
	for i, maddr := range hps.RemoteMaddrs {
		rMaddrBytes[i] = maddr.Bytes()
	}

	oMaddrBytes := make([][]byte, len(hps.OpenMaddrs))
	for i, maddr := range hps.OpenMaddrs {
		oMaddrBytes[i] = maddr.Bytes()
	}

	hpAttempts := make([]*pb.HolePunchAttempt, len(hps.HolePunchAttempts))
	for i, attempt := range hps.HolePunchAttempts {
		hpAttempts[i] = attempt.ToProto()
	}

	return &pb.TrackHolePunchRequest{
		ClientId:             localID,
		RemoteId:             remoteID,
		RemoteMultiAddresses: rMaddrBytes,
		ConnectStartedAt:     toUnixMillis(hps.ConnectStartedAt),
		ConnectEndedAt:       toUnixMillis(hps.ConnectEndedAt),
		HolePunchAttempts:    hpAttempts,
		OpenMultiAddresses:   oMaddrBytes,
		HasDirectConns:       &hps.HasDirectConns,
		Outcome:              &hps.Outcome,
		EndedAt:              toUnixMillis(hps.EndedAt),
	}, nil
}

type HolePunchAttempt struct {
	HostID          peer.ID
	RemoteID        peer.ID
	OpenedAt        time.Time
	StartedAt       time.Time
	EndedAt         time.Time
	StartRTT        time.Duration
	ElapsedTime     time.Duration
	Error           string
	DirectDialError string
	Outcome         pb.HolePunchAttemptOutcome
}

func (hpa *HolePunchAttempt) handleError(err error) {
	hpa.EndedAt = time.Now()
	hpa.ElapsedTime = hpa.EndedAt.Sub(hpa.StartedAt)
	hpa.Error = err.Error()
	if errors.Is(err, ErrHolePunchTimeout) {
		hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_TIMEOUT
	}
}

func (hpa *HolePunchAttempt) handleStartHolePunchEvt(event *holepunch.Event, evt *holepunch.StartHolePunchEvt) {
	hpa.logEntry().Infoln("Hole punch started")
	hpa.StartedAt = time.Unix(0, event.Timestamp)
	hpa.StartRTT = evt.RTT
}

func (hpa *HolePunchAttempt) handleEndHolePunchEvt(event *holepunch.Event, evt *holepunch.EndHolePunchEvt) {
	hpa.logEntry().WithField("success", evt.Success).Infoln("Hole punch ended")
	hpa.EndedAt = time.Unix(0, event.Timestamp)
	hpa.Error = evt.Error
	hpa.ElapsedTime = evt.EllapsedTime
	if evt.Success {
		hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_SUCCESS
	} else {
		hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_FAILED
	}
}

func (hpa *HolePunchAttempt) handleHolePunchAttemptEvt(evt *holepunch.HolePunchAttemptEvt) {
	hpa.logEntry().Infoln("Hole punch attempt")
}

func (hpa *HolePunchAttempt) handleProtocolErrorEvt(event *holepunch.Event, evt *holepunch.ProtocolErrorEvt) {
	hpa.logEntry().WithField("err", evt.Error).Infoln("Hole punching protocol error :/")
	hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_PROTOCOL_ERROR
	hpa.Error = evt.Error
	hpa.ElapsedTime = time.Since(hpa.StartedAt)
	hpa.EndedAt = time.Unix(0, event.Timestamp)
}

func (hpa *HolePunchAttempt) handleDirectDialEvt(event *holepunch.Event, evt *holepunch.DirectDialEvt) {
	hpa.logEntry().WithField("success", evt.Success).Warnln("Hole punch direct dial event")
	if evt.Success {
		hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_DIRECT_DIAL
		hpa.ElapsedTime = evt.EllapsedTime
		hpa.EndedAt = time.Unix(0, event.Timestamp)
	} else {
		hpa.DirectDialError = evt.Error
	}
}

func (hpa HolePunchAttempt) logEntry() *log.Entry {
	return log.WithFields(log.Fields{
		"remoteID": util.FmtPeerID(hpa.RemoteID),
		"hostID":   util.FmtPeerID(hpa.HostID),
	})
}

func (hpa HolePunchAttempt) ToProto() *pb.HolePunchAttempt {
	return &pb.HolePunchAttempt{
		OpenedAt:        toUnixMillis(hpa.OpenedAt),
		StartedAt:       toUnixMillis(hpa.StartedAt),
		EndedAt:         toUnixMillis(hpa.EndedAt),
		StartRtt:        toSeconds(hpa.StartRTT),
		ElapsedTime:     toSeconds(hpa.ElapsedTime),
		Error:           &hpa.Error,
		DirectDialError: &hpa.DirectDialError,
		Outcome:         &hpa.Outcome,
	}
}

func toUnixMillis(t time.Time) *uint64 {
	millis := uint64(t.UnixMilli())
	return &millis
}

func toSeconds(dur time.Duration) *float32 {
	if dur == 0 {
		return nil
	}

	s := float32(dur.Seconds())
	return &s
}
