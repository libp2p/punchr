package util

import (
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Unique[T comparable](input []T) *T {
	u := make([]T, 0, len(input))
	m := make(map[T]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	if len(u) == 1 {
		return &u[0]
	}

	return nil
}

func IsRelayedMaddr(maddr ma.Multiaddr) bool {
	_, err := maddr.ValueForProtocol(ma.P_CIRCUIT)
	if err == nil {
		return true
	} else if errors.Is(err, ma.ErrProtocolNotFound) {
		return false
	} else {
		log.WithError(err).WithField("maddr", maddr).Warnln("Unexpected error while parsing multi address")
		return false
	}
}

func SupportDCUtR(protocols []string) bool {
	for _, p := range protocols {
		if p == string(holepunch.Protocol) {
			return true
		}
	}
	return false
}

func ContainsPublicAddr(addrs []ma.Multiaddr) bool {
	for _, addr := range addrs {
		if IsRelayedMaddr(addr) || !manet.IsPublicAddr(addr) {
			continue
		}
		return true
	}
	return false
}
