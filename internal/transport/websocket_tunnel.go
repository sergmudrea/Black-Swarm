
package transport

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
	"github.com/gorilla/websocket"
)

// WebSocketTunnel implements the Transport interface over a WebSocket connection.
// It can be used both as client (dialing a server) and server (accepting a connection).
type WebSocketTunnel struct {
	conn *websocket.Conn
	mu   sync.Mutex // protects writes
}

// NewWebSocketClient creates a WebSocketTunnel by dialing the given URL.
// tlsConfig may be nil for insecure connections.
func NewWebSocketClient(url string, tlsConfig *tls.Config) (*WebSocketTunnel, error) {
	dialer := websocket.Dialer{
		TLSClientConfig:  tlsConfig,
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(url, http.Header{})
	if err != nil {
		return nil, fmt.Errorf("websocket tunnel: dial: %w", err)
	}
	return &WebSocketTunnel{conn: conn}, nil
}

// NewWebSocketServer wraps an existing WebSocket connection (e.g., from an HTTP upgrade).
func NewWebSocketServer(conn *websocket.Conn) *WebSocketTunnel {
	return &WebSocketTunnel{conn: conn}
}

// Send marshals the envelope to JSON and writes it to the WebSocket.
func (w *WebSocketTunnel) Send(ctx context.Context, address string, env *protocol.Envelope) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn == nil {
		return fmt.Errorf("websocket tunnel: not connected")
	}
	return w.conn.WriteJSON(env)
}

// Receive blocks until a JSON envelope is read from the WebSocket.
func (w *WebSocketTunnel) Receive(ctx context.Context) (*protocol.Envelope, error) {
	if w.conn == nil {
		return nil, fmt.Errorf("websocket tunnel: not connected")
	}
	var env protocol.Envelope
	err := w.conn.ReadJSON(&env)
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// Close closes the underlying WebSocket connection.
func (w *WebSocketTunnel) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}
