package client

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHost_filter(t *testing.T) {
	h := &Host{protocolFilters: []int32{}}

	dummyID := peer.ID("")

	filtered := h.filter(dummyID, nil)
	assert.Len(t, filtered, 0)

	filtered = h.filter(dummyID, []multiaddr.Multiaddr{})
	assert.Len(t, filtered, 0)

	maddr1, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/1234")
	maddr2, err := multiaddr.NewMultiaddr("/ip6/::1/udp/1234/quic")
	require.NoError(t, err)

	filtered = h.filter(dummyID, []multiaddr.Multiaddr{maddr1, maddr2})
	assert.Len(t, filtered, 2)

	h.protocolFilters = []int32{multiaddr.P_IP6}
	filtered = h.filter(dummyID, []multiaddr.Multiaddr{maddr1, maddr2})
	assert.Len(t, filtered, 1)

	h.protocolFilters = []int32{multiaddr.P_IP4}
	filtered = h.filter(dummyID, []multiaddr.Multiaddr{maddr1, maddr2})
	assert.Len(t, filtered, 1)

	h.protocolFilters = []int32{multiaddr.P_IP4, multiaddr.P_TCP}
	filtered = h.filter(dummyID, []multiaddr.Multiaddr{maddr1, maddr2})
	assert.Len(t, filtered, 1)

	h.protocolFilters = []int32{multiaddr.P_IP4, multiaddr.P_QUIC}
	filtered = h.filter(dummyID, []multiaddr.Multiaddr{maddr1, maddr2})
	assert.Len(t, filtered, 0)

	h.protocolFilters = []int32{multiaddr.P_IP6, multiaddr.P_TCP}
	filtered = h.filter(dummyID, []multiaddr.Multiaddr{maddr1, maddr2})
	assert.Len(t, filtered, 0)
}
