package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	pool "github.com/libp2p/go-buffer-pool"
	"io"
	mrand "math/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	log "github.com/sirupsen/logrus"
)

// Ping pings the remote peer until the context is canceled, returning a stream
// of RTTs or errors.
func Ping(ctx context.Context, h host.Host, p peer.ID) (<-chan ping.Result, network.Stream) {
	s, err := h.NewStream(network.WithUseTransient(ctx, "ping"), p, ping.ID)
	if err != nil {
		return pingError(err), nil
	}

	if err := s.Scope().SetService(ping.ServiceName); err != nil {
		log.Debugf("error attaching stream to ping service: %s", err)
		s.Reset()
		return pingError(err), nil
	}

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		log.Errorf("failed to get cryptographic random: %s", err)
		s.Reset()
		return pingError(err), nil
	}
	ra := mrand.New(mrand.NewSource(int64(binary.BigEndian.Uint64(b))))

	ctx, cancel := context.WithCancel(ctx)

	out := make(chan ping.Result)
	go func() {
		defer close(out)
		defer cancel()

		for ctx.Err() == nil {
			var res ping.Result
			res.RTT, res.Error = pingImpl(s, ra)

			// canceled, ignore everything.
			if ctx.Err() != nil {
				return
			}

			// No error, record the RTT.
			if res.Error == nil {
				h.Peerstore().RecordLatency(p, res.RTT)
			}

			select {
			case out <- res:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		// forces the ping to abort.
		<-ctx.Done()
		s.Reset()
	}()

	return out, s
}

func pingError(err error) chan ping.Result {
	ch := make(chan ping.Result, 1)
	ch <- ping.Result{Error: err}
	close(ch)
	return ch
}

func pingImpl(s network.Stream, randReader io.Reader) (time.Duration, error) {
	if err := s.Scope().ReserveMemory(2*ping.PingSize, network.ReservationPriorityAlways); err != nil {
		log.Debugf("error reserving memory for ping stream: %s", err)
		s.Reset()
		return 0, err
	}
	defer s.Scope().ReleaseMemory(2 * ping.PingSize)

	buf := pool.Get(ping.PingSize)
	defer pool.Put(buf)

	if _, err := io.ReadFull(randReader, buf); err != nil {
		return 0, err
	}

	before := time.Now()
	if _, err := s.Write(buf); err != nil {
		return 0, err
	}

	rbuf := pool.Get(ping.PingSize)
	defer pool.Put(rbuf)

	if _, err := io.ReadFull(s, rbuf); err != nil {
		return 0, err
	}

	if !bytes.Equal(buf, rbuf) {
		return 0, errors.New("ping packet was incorrect")
	}

	return time.Since(before), nil
}
