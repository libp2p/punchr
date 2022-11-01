package util

import (
	"context"
	"github.com/jackpal/gateway"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
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

func ExtractRelayMaddr(maddr ma.Multiaddr) (*peer.AddrInfo, error) {
	circProt := ma.ProtocolWithCode(ma.P_CIRCUIT)
	circComp, err := ma.NewComponent(circProt.Name, "")
	if err != nil {
		return nil, errors.Wrap(err, "new circuit component")
	}

	return peer.AddrInfoFromP2pAddr(maddr.Decapsulate(circComp))
}

// DefaultGatewayHTML discovers the default gateway address and fetches the
// home HTML page to get a sense which router is used.
func DefaultGatewayHTML(ctx context.Context) (string, error) {
	log.Infoln("Checking router HTML...")
	defer log.Infoln("Checking router HTML - Done!")

	router, err := gateway.DiscoverGateway()
	if err != nil {
		return "", errors.Wrap(err, "discover gateway")
	}

	u := url.URL{Scheme: "https", Host: router.String()}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", errors.Wrap(err, "new https request with context")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		u := url.URL{Scheme: "http", Host: router.String()}
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return "", errors.Wrap(err, "new http request with context")
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return "", errors.Wrap(err, "get http")
		}
	}

	html, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read response body")
	}

	return string(html), nil
}
