package client

import (
	"github.com/dennis-tra/punchr/pkg/util"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExtractRelayMaddr(t *testing.T) {
	relayedMaddr, err := multiaddr.NewMultiaddr("/ip4/185.130.47.68/udp/2296/quic/p2p/12D3KooWMNqypRn921xoSU6rEJBa1RVPwuHnFwtSMQZeGfafQzSg/p2p-circuit")
	require.NoError(t, err)

	pi, err := util.ExtractRelayMaddr(relayedMaddr)
	require.NoError(t, err)

	assert.Equal(t, "12D3KooWMNqypRn921xoSU6rEJBa1RVPwuHnFwtSMQZeGfafQzSg", pi.ID.String())
	assert.Equal(t, "/ip4/185.130.47.68/udp/2296/quic", pi.Addrs[0].String())

	relayedMaddr, err = multiaddr.NewMultiaddr("/ip4/185.130.47.68/udp/2296/quic/p2p/12D3KooWMNqypRn921xoSU6rEJBa1RVPwuHnFwtSMQZeGfafQzSg")
	require.NoError(t, err)

	pi, err = util.ExtractRelayMaddr(relayedMaddr)
	require.NoError(t, err)

	assert.Equal(t, "12D3KooWMNqypRn921xoSU6rEJBa1RVPwuHnFwtSMQZeGfafQzSg", pi.ID.String())
	assert.Equal(t, "/ip4/185.130.47.68/udp/2296/quic", pi.Addrs[0].String())
}
