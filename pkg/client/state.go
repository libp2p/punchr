package client

import (
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
)

type MeasurementType int

const (
	ToRelay MeasurementType = iota + 1
	ToRemoteThroughRelay
	ToRemoteAfterHolePunch
)

func (lm LatencyMeasurement) toProto() (*pb.LatencyMeasurement, error) {
	remoteID, err := lm.remoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal latency measurement remote peer id")
	}

	rtts := []float32{}
	for _, rtt := range lm.rtts {
		if rtt == 0 {
			rtts = append(rtts, float32(-1))
		} else {
			rtts = append(rtts, float32(rtt.Seconds()))
		}
	}

	rttErrs := []string{}
	for _, rttErr := range lm.rttErrs {
		rttErrs = append(rttErrs, rttErr.Error())
	}

	return &pb.LatencyMeasurement{
		RemoteId:     remoteID,
		AgentVersion: lm.agentVersion,
		Protocols:    lm.protocols,
		MultiAddress: lm.conn.Bytes(),
		Rtts:         rtts,
		RttErrs:      rttErrs,
	}, nil
}

func extractRelayInfo(info peer.AddrInfo) map[peer.ID]*peer.AddrInfo {
	relays := make(map[peer.ID]*peer.AddrInfo)

	for _, addr := range info.Addrs {
		relayAddrInfo, err := util.ExtractRelayMaddr(addr)
		if err != nil {
			log.WithError(err).Warnln("error extracting relay multi address")
			continue
		}

		if _, ok := relays[relayAddrInfo.ID]; !ok {
			relays[relayAddrInfo.ID] = relayAddrInfo
			continue
		}

		relays[relayAddrInfo.ID].Addrs = append(relays[relayAddrInfo.ID].Addrs, relayAddrInfo.Addrs...)
	}

	return relays
}

type HolePunchState struct {
	// The host that established the connection to the remote peer via a relay, and it's listening multi addresses
	HostID      peer.ID
	LocalMaddrs []multiaddr.Multiaddr

	// The remote peer and its multi addresses - usually relayed ones. We use that set for the first connection.
	RemoteID     peer.ID
	RemoteMaddrs []multiaddr.Multiaddr

	// Start and end times for the establishment of the relayed connection
	ConnectStartedAt time.Time
	ConnectEndedAt   time.Time

	// Multi addresses of the open connections before the hole punch aka the multi addresses we
	// managed to get a connection to the remote peer.
	OpenMaddrsBefore []multiaddr.Multiaddr

	// Information about each individual hole punch attempt
	HolePunchAttempts []*HolePunchAttempt

	// Multi addresses of the open connections after the hole punch
	OpenMaddrsAfter []multiaddr.Multiaddr
	HasDirectConns  bool

	Error   string
	Outcome pb.HolePunchOutcome
	EndedAt time.Time

	// The login page of the router
	RouterHTML      string
	ProtocolFilters []int

	// Remote Peer data
	RemoteRttAfterHolePunch time.Duration

	// LatencyMeasurements of different types
	LatencyMeasurements []LatencyMeasurement
}

func NewHolePunchState(hostID peer.ID, remoteID peer.ID, rmaddrs []multiaddr.Multiaddr, lmaddrs []multiaddr.Multiaddr) *HolePunchState {
	return &HolePunchState{
		HostID:              hostID,
		RemoteID:            remoteID,
		RemoteMaddrs:        rmaddrs,
		LocalMaddrs:         lmaddrs,
		OpenMaddrsBefore:    []multiaddr.Multiaddr{},
		HolePunchAttempts:   []*HolePunchAttempt{},
		OpenMaddrsAfter:     []multiaddr.Multiaddr{},
		LatencyMeasurements: []LatencyMeasurement{},
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

func (hps HolePunchState) ToProto(apiKey string) (*pb.TrackHolePunchRequest, error) {
	localID, err := hps.HostID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal local peer id")
	}

	remoteID, err := hps.RemoteID.Marshal()
	if err != nil {
		return nil, errors.Wrap(err, "marshal remote peer id")
	}

	lMaddrBytes := make([][]byte, len(hps.LocalMaddrs))
	for i, maddr := range hps.LocalMaddrs {
		lMaddrBytes[i] = maddr.Bytes()
	}

	rMaddrBytes := make([][]byte, len(hps.RemoteMaddrs))
	for i, maddr := range hps.RemoteMaddrs {
		rMaddrBytes[i] = maddr.Bytes()
	}

	oMaddrBytes := make([][]byte, len(hps.OpenMaddrsAfter))
	for i, maddr := range hps.OpenMaddrsAfter {
		oMaddrBytes[i] = maddr.Bytes()
	}

	hpAttempts := make([]*pb.HolePunchAttempt, len(hps.HolePunchAttempts))
	for i, attempt := range hps.HolePunchAttempts {
		hpAttempts[i] = attempt.ToProto()
	}

	var routerLoginHTML *string
	if hps.RouterHTML != "" {
		routerLoginHTML = &hps.RouterHTML
	}

	lms := make([]*pb.LatencyMeasurement, len(hps.LatencyMeasurements))
	for i, lm := range hps.LatencyMeasurements {
		lms[i], err = lm.toProto()
		if err != nil {
			return nil, errors.Wrap(err, "marshal latency measurement")
		}
	}

	filterProtocols := make([]int32, len(hps.ProtocolFilters))
	for i, pf := range hps.ProtocolFilters {
		filterProtocols[i] = int32(pf)
	}

	var errStr *string
	if hps.Error != "" {
		errStr = &hps.Error
	}

	return &pb.TrackHolePunchRequest{
		ApiKey:               &apiKey,
		ClientId:             localID,
		ListenMultiAddresses: lMaddrBytes,
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
		RouterLoginHtml:      routerLoginHTML,
		LatencyMeasurements:  lms,
		Protocols:            filterProtocols,
	}, nil
}

type HolePunchAttempt struct {
	HostID      peer.ID
	RemoteID    peer.ID
	RemoteAddrs []multiaddr.Multiaddr

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

	maddrs := make([]multiaddr.Multiaddr, len(evt.RemoteAddrs))
	for i, addr := range evt.RemoteAddrs {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			log.WithError(err).WithField("maddr", addr).Warn("Could not parse maddr")
		}
		maddrs[i] = maddr
	}
	hpa.RemoteAddrs = maddrs
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
	hpa.logEntry().Infoln("no hole punch event after", CommunicationTimeout)
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

	maddrs := make([][]byte, len(hpa.RemoteAddrs))
	for i, raddr := range hpa.RemoteAddrs {
		maddrs[i] = raddr.Bytes()
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
		MultiAddresses:  maddrs,
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
