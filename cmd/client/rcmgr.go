package main

import (
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
)

type ResourceManager struct {
	network.ResourceManager

	notifieesLk sync.RWMutex
	notifiees   map[peer.ID]chan struct{}
}

var _ network.ResourceManager = (*ResourceManager)(nil)

func NewResourceManager() (*ResourceManager, error) {
	// Start with the default scaling limits.
	scalingLimits := rcmgr.DefaultLimits

	// Add limits around included libp2p protocols
	libp2p.SetDefaultServiceLimits(&scalingLimits)

	// Turn the scaling limits into a static set of limits using `.AutoScale`. This
	// scales the limits proportional to your system memory.
	limits := scalingLimits.AutoScale()

	// The resource manager expects a limiter, se we create one from our limits.
	limiter := rcmgr.NewFixedLimiter(limits)

	mgr, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, err
	}

	return &ResourceManager{ResourceManager: mgr, notifiees: map[peer.ID]chan struct{}{}}, nil
}

func (r *ResourceManager) Register(pid peer.ID) <-chan struct{} {
	r.notifieesLk.Lock()
	defer r.notifieesLk.Unlock()

	dcutrOpenedChan := make(chan struct{})
	r.notifiees[pid] = dcutrOpenedChan
	return dcutrOpenedChan
}

func (r *ResourceManager) Unregister(pid peer.ID) {
	r.notifieesLk.Lock()
	delete(r.notifiees, pid)
	r.notifieesLk.Unlock()
}

func (r *ResourceManager) OpenStream(p peer.ID, dir network.Direction) (network.StreamManagementScope, error) {
	r.notifieesLk.RLock()
	defer r.notifieesLk.RUnlock()

	if dcutrOpenedChan, ok := r.notifiees[p]; ok && dir == network.DirInbound {
		sms, err := r.ResourceManager.OpenStream(p, dir)
		return &StreamManagementScope{
			StreamManagementScope: sms,
			dcutrOpenedChan:       dcutrOpenedChan,
		}, err
	} else {
		return r.ResourceManager.OpenStream(p, dir)
	}
}

type StreamManagementScope struct {
	network.StreamManagementScope

	notifiedLk sync.RWMutex
	notified   bool

	dcutrOpenedChan chan struct{}
}

var _ network.StreamManagementScope = (*StreamManagementScope)(nil)

func (s *StreamManagementScope) SetProtocol(proto protocol.ID) error {
	s.notifiedLk.Lock()
	if proto != holepunch.Protocol || s.notified {
		s.notifiedLk.Unlock()
		return s.StreamManagementScope.SetProtocol(proto)
	}

	close(s.dcutrOpenedChan)
	s.notified = true
	s.notifiedLk.Unlock()

	return s.StreamManagementScope.SetProtocol(proto)
}
