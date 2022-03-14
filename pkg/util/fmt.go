package util

import "github.com/libp2p/go-libp2p-core/peer"

// IDLength is here as a variable so that it can be decreased for tests with mocknet where IDs are way shorter.
// The call to FmtPeerID would panic if this value stayed at 16.
var IDLength = 16

func FmtPeerID(id peer.ID) string {
	if len(id.Pretty()) <= IDLength {
		return id.Pretty()
	}
	return id.Pretty()[:IDLength]
}
