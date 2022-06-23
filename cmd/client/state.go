package main

import (
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
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

// onlyRelayRemoteAddrs returns true if the hole punch was attempted with ONLY relayed addresses
// (we didn't have a direct connection prior the hole punch)
func (hps HolePunchState) onlyRelayRemoteAddrs() bool {
	for _, maddr := range hps.RemoteMaddrs {
		if !util.IsRelayedMaddr(maddr) {
			return false
		}
	}
	return true
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

	var errStr *string
	if hps.Error != "" {
		errStr = &hps.Error
	}

	return &pb.TrackHolePunchRequest{
		ClientId:             localID,
		RemoteId:             remoteID,
		RemoteMultiAddresses: rMaddrBytes,
		ConnectStartedAt:     toUnixNanos(hps.ConnectStartedAt),
		ConnectEndedAt:       toUnixNanos(hps.ConnectEndedAt),
		HolePunchAttempts:    hpAttempts,
		OpenMultiAddresses:   oMaddrBytes,
		HasDirectConns:       &hps.HasDirectConns,
		Outcome:              &hps.Outcome,
		Error:                errStr,
		EndedAt:              toUnixNanos(hps.EndedAt),
	}, nil
}

type HolePunchAttempt struct {
	HostID   peer.ID
	RemoteID peer.ID

	// Time when the /libp2p/dcutr stream was opened
	OpenedAt time.Time
	// Time when we received a hole punch started event
	StartedAt time.Time
	// Time when this hole punch attempt stopped (failure, cancel, timeout)
	EndedAt time.Time
	// The measured round trip time from the holepunch start event
	StartRTT time.Duration

	ElapsedTime     time.Duration
	Error           string
	DirectDialError string
	Outcome         pb.HolePunchAttemptOutcome
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
		hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_SUCCESS
	} else {
		hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_FAILED
	}
}

func (hpa *HolePunchAttempt) handleHolePunchAttemptEvt(evt *holepunch.HolePunchAttemptEvt) {
	hpa.logEntry().Infoln("Hole punch attempt")
}

func (hpa *HolePunchAttempt) handleProtocolErrorEvt(event *holepunch.Event, evt *holepunch.ProtocolErrorEvt) {
	hpa.logEntry().WithField("err", evt.Error).Infoln("Hole punching protocol error :/")
	hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_PROTOCOL_ERROR
	hpa.Error = evt.Error
	hpa.EndedAt = time.Unix(0, event.Timestamp)
	if hpa.StartedAt.IsZero() {
		hpa.ElapsedTime = hpa.EndedAt.Sub(hpa.OpenedAt)
	} else {
		hpa.ElapsedTime = time.Since(hpa.StartedAt)
	}
}

func (hpa *HolePunchAttempt) handleDirectDialEvt(event *holepunch.Event, evt *holepunch.DirectDialEvt) {
	hpa.logEntry().WithField("success", evt.Success).Warnln("Hole punch direct dial event")
	if evt.Success {
		hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_DIRECT_DIAL
		hpa.ElapsedTime = evt.EllapsedTime
		hpa.EndedAt = time.Unix(0, event.Timestamp)
	} else {
		hpa.DirectDialError = evt.Error
	}
}

func (hpa *HolePunchAttempt) handleHolePunchTimeout() {
	hpa.logEntry().Infoln("no hole punch event after %s", CommunicationTimeout)
	hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_TIMEOUT
	hpa.EndedAt = time.Now()
	hpa.ElapsedTime = hpa.EndedAt.Sub(hpa.OpenedAt)
	hpa.Error = fmt.Sprintf("no hole punch event after %s", CommunicationTimeout)
}

func (hpa *HolePunchAttempt) handleHolePunchCancelled(err error) {
	hpa.logEntry().WithField("err", err).Infoln("Hole punch was cancelled")
	hpa.Outcome = pb.HolePunchAttemptOutcome_HOLE_PUNCH_ATTEMPT_OUTCOME_CANCELLED
	hpa.EndedAt = time.Now()
	if hpa.StartedAt.IsZero() {
		hpa.ElapsedTime = hpa.EndedAt.Sub(hpa.OpenedAt)
	} else {
		hpa.ElapsedTime = hpa.EndedAt.Sub(hpa.StartedAt)
	}
	hpa.Error = err.Error()
}

func (hpa HolePunchAttempt) logEntry() *log.Entry {
	return log.WithFields(log.Fields{
		"remoteID": util.FmtPeerID(hpa.RemoteID),
		"hostID":   util.FmtPeerID(hpa.HostID),
	})
}

func (hpa HolePunchAttempt) ToProto() *pb.HolePunchAttempt {
	var startedAt *uint64
	if !hpa.StartedAt.IsZero() {
		startedAt = toUnixNanos(hpa.StartedAt)
	}
	return &pb.HolePunchAttempt{
		OpenedAt:        toUnixNanos(hpa.OpenedAt),
		StartedAt:       startedAt,
		EndedAt:         toUnixNanos(hpa.EndedAt),
		StartRtt:        toSeconds(hpa.StartRTT),
		ElapsedTime:     toSeconds(hpa.ElapsedTime),
		Error:           &hpa.Error,
		DirectDialError: &hpa.DirectDialError,
		Outcome:         &hpa.Outcome,
	}
}

func toUnixNanos(t time.Time) *uint64 {
	nanos := uint64(t.UnixNano())
	return &nanos
}

func toSeconds(dur time.Duration) *float32 {
	if dur == 0 {
		return nil
	}

	s := float32(dur.Seconds())
	return &s
}
