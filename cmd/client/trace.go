package main

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/dennis-tra/punchr/pkg/pb"
	"github.com/dennis-tra/punchr/pkg/util"
)

// Trace is called during the hole punching process
func (h *Host) Trace(evt *holepunch.Event) {
	h.hpStatesLk.Lock()
	defer h.hpStatesLk.Unlock()

	hpState, found := h.hpStates[evt.Remote]
	if !found {
		return
	}

	logEntry := log.WithField("remoteID", util.FmtPeerID(evt.Remote))

	switch event := evt.Evt.(type) {
	case *holepunch.StartHolePunchEvt:
		logEntry.Infoln("Hole punch started")
		hpState.StartRTT = event.RTT
		close(hpState.holePunchStarted)
	case *holepunch.EndHolePunchEvt:
		logEntry.WithField("success", event.Success).Infoln("Hole punch ended")
		hpState.EndReason = pb.HolePunchEndReason_HOLE_PUNCH
		hpState.Error = event.Error
		hpState.Success = event.Success
		hpState.ElapsedTime = event.EllapsedTime
		close(hpState.holePunchFinished)
	case *holepunch.HolePunchAttemptEvt:
		hpState.Attempts += 1 // event.Attempt <-- does not count correctly if hole punching with same peer happens shortly after one another. GC is not run in time and can't be triggered.
		logEntry.Infoln("Hole punch attempt", hpState.Attempts)
	case *holepunch.ProtocolErrorEvt:
		logEntry.WithField("err", event.Error).Infoln("Hole punching protocol error :/")
		hpState.EndReason = pb.HolePunchEndReason_PROTOCOL_ERROR
		hpState.Error = event.Error
		hpState.Success = false
		hpState.ElapsedTime = time.Since(hpState.ConnectionStartedAt)
		close(hpState.holePunchFinished)
	case *holepunch.DirectDialEvt:
		logEntry.WithField("success", event.Success).Warnln("Hole punch direct dial event")
		if event.Success {
			hpState.EndReason = pb.HolePunchEndReason_DIRECT_DIAL
			hpState.Success = event.Success
			hpState.ElapsedTime = event.EllapsedTime
			close(hpState.holePunchFinished)
		} else {
			hpState.DirectDialError = event.Error
		}
	default:
		panic(fmt.Sprintf("unexpected event %T", evt.Evt))
	}
}

func (h *Host) WaitForHolePunch(ctx context.Context, pid peer.ID, hpStart <-chan struct{}, hpFinish <-chan struct{}) error {
	log.WithFields(log.Fields{
		"remoteID": util.FmtPeerID(pid),
		"waitDur":  15 * time.Second,
	}).Infoln("Waiting for hole punch...")

	select {
	case <-hpStart:
		// pass
	case <-hpFinish:
		return nil
	case <-time.After(15 * time.Second):
		return errors.New("hole punch was not initiated")
	case <-ctx.Done():
		return ctx.Err()
	}

	// Then wait for the hole punch to finish
	select {
	case <-hpFinish:
		return nil
	case <-time.After(time.Minute):
		return errors.New("hole punch did not finish in time")
	case <-ctx.Done():
		return ctx.Err()
	}
}
