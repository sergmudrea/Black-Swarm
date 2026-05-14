// Package transport defines the interface for sending and receiving messages
// between swarm nodes, and provides a registry of available transports.
package transport

import (
	"context"

	"github.com/blackswarm/siege/internal/protocol"
)

// Transport is the interface that all communication transports must implement.
type Transport interface {
	// Send delivers an envelope to the given address.
	// The address format is transport‑specific (e.g., "host:port" for UDP,
	// "wss://host/path" for WebSocket).
	Send(ctx context.Context, address string, env *protocol.Envelope) error

	// Receive blocks until an envelope is received, then returns it.
	// The caller should call Receive in a loop.
	Receive(ctx context.Context) (*protocol.Envelope, error)

	// Close releases any resources held by the transport.
	Close() error
}
