package gossip

import (
	"context"

	"github.com/blackswarm/siege/internal/node"
	"github.com/blackswarm/siege/internal/protocol"
)

// PeerManager implements the node.PeerManager interface using the gossip Protocol.
type PeerManager struct {
	proto *Protocol
}

// NewPeerManager creates a PeerManager that wraps the given gossip Protocol.
func NewPeerManager(proto *Protocol) *PeerManager {
	return &PeerManager{proto: proto}
}

// Start begins the gossip protocol. The context is ignored (the protocol manages
// its own lifecycle); cancellation is achieved via Stop.
func (pm *PeerManager) Start(_ context.Context) error {
	return pm.proto.Start()
}

// Stop terminates the gossip protocol.
func (pm *PeerManager) Stop() error {
	return pm.proto.Stop()
}

// Peers returns a snapshot of the currently known peers.
func (pm *PeerManager) Peers() []protocol.PeerInfo {
	return pm.proto.Peers()
}

// Broadcast sends an envelope to all known peers (best‑effort).
func (pm *PeerManager) Broadcast(msg *protocol.Envelope) error {
	return pm.proto.Broadcast(msg)
}

// Send delivers an envelope to a specific peer by its node ID.
func (pm *PeerManager) Send(peerID string, msg *protocol.Envelope) error {
	return pm.proto.Send(peerID, msg)
}

// OnMessage registers a callback for every incoming message.
func (pm *PeerManager) OnMessage(handler func(*protocol.Envelope)) {
	pm.proto.OnMessage(handler)
}
