// Package node implements the lifecycle and core behaviour of a Siege node.
package node

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/config"
	"github.com/blackswarm/siege/internal/protocol"
)

// Node represents a single Siege swarm node.
type Node struct {
	cfg    *config.Config
	logger *slog.Logger

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Components (lazy initialised)
	httpServer *http.Server
	peerMgr    PeerManager
	scheduler  TaskScheduler

	// Outbound message queue
	outbox chan *protocol.Envelope
}

// PeerManager is the interface for peer discovery and gossip.
type PeerManager interface {
	Start(ctx context.Context) error
	Stop() error
	Peers() []protocol.PeerInfo
	Broadcast(msg *protocol.Envelope) error
	Send(peerID string, msg *protocol.Envelope) error
	OnMessage(handler func(*protocol.Envelope))
}

// TaskScheduler is the interface for distributed task scheduling.
type TaskScheduler interface {
	Start(ctx context.Context) error
	Stop() error
	Submit(task *protocol.TaskMsg) error
	PendingTasks() []protocol.TaskMsg
}

// NewNode creates a new Node from the given configuration.
func NewNode(cfg *config.Config, logger *slog.Logger) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		cfg:    cfg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		outbox: make(chan *protocol.Envelope, 256),
	}
}

// Start initialises and starts all node subsystems.
func (n *Node) Start() error {
	n.logger.Info("starting node",
		"node_id", n.cfg.NodeID,
		"mode", n.cfg.Mode,
		"bind", n.cfg.BindAddr,
	)

	// Start subsystems according to mode
	switch n.cfg.Mode {
	case "strategic":
		n.startStrategic()
	case "scanner":
		n.startScanner()
	case "hybrid":
		n.startStrategic()
		n.startScanner()
	default:
		n.logger.Warn("unknown mode, starting as scanner", "mode", n.cfg.Mode)
		n.startScanner()
	}

	n.logger.Info("node started")
	return nil
}

func (n *Node) startStrategic() {
	// Peer manager + scheduler for coordination
	if n.peerMgr != nil {
		n.wg.Add(1)
		go func() {
			defer n.wg.Done()
			if err := n.peerMgr.Start(n.ctx); err != nil {
				n.logger.Error("peer manager failed", "error", err)
			}
		}()
	}
	if n.scheduler != nil {
		n.wg.Add(1)
		go func() {
			defer n.wg.Done()
			if err := n.scheduler.Start(n.ctx); err != nil {
				n.logger.Error("scheduler failed", "error", err)
			}
		}()
	}
}

func (n *Node) startScanner() {
	// Scanner nodes listen for incoming tasks via the outbox
	n.wg.Add(1)
	go n.processOutbox()
}

func (n *Node) processOutbox() {
	defer n.wg.Done()
	for {
		select {
		case <-n.ctx.Done():
			return
		case msg := <-n.outbox:
			n.logger.Debug("outbound message", "type", msg.Type, "sender", msg.SenderID)
		}
	}
}

// Shutdown gracefully stops the node.
func (n *Node) Shutdown(ctx context.Context) error {
	n.logger.Info("shutting down node")
	n.cancel()

	done := make(chan struct{})
	go func() {
		n.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		n.logger.Info("node stopped cleanly")
	case <-ctx.Done():
		n.logger.Warn("shutdown timed out")
	}

	return nil
}

// SetPeerManager injects a peer manager implementation.
func (n *Node) SetPeerManager(pm PeerManager) {
	n.peerMgr = pm
}

// SetScheduler injects a task scheduler implementation.
func (n *Node) SetScheduler(s TaskScheduler) {
	n.scheduler = s
}

// Config returns the node configuration.
func (n *Node) Config() *config.Config {
	return n.cfg
}

// Logger returns the node logger.
func (n *Node) Logger() *slog.Logger {
	return n.logger
}
