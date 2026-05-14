// Package gossip implements a lightweight UDP gossip protocol for peer discovery,
// state dissemination, and best‑effort broadcast within the swarm.
package gossip

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/config"
	"github.com/blackswarm/siege/internal/protocol"
)

// Protocol is the core gossip service that manages peer membership and message routing.
type Protocol struct {
	nodeID       string
	bindAddr     string
	advertiseAddr string
	conn         *net.UDPConn

	// Peer management
	peersMu sync.RWMutex
	peers   map[string]*peerState // keyed by node ID

	// Handlers for incoming messages (callbacks registered via OnMessage)
	handlers []func(*protocol.Envelope)

	logger *slog.Logger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// peerState holds the information about a known peer and its last seen timestamp.
type peerState struct {
	info     protocol.PeerInfo
	lastSeen time.Time
}

// NewProtocol creates a new gossip Protocol instance.
func NewProtocol(nodeID, bindAddr, advertiseAddr string, logger *slog.Logger) *Protocol {
	ctx, cancel := context.WithCancel(context.Background())
	return &Protocol{
		nodeID:        nodeID,
		bindAddr:      bindAddr,
		advertiseAddr: advertiseAddr,
		peers:         make(map[string]*peerState),
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start binds the UDP socket and begins background loops.
func (p *Protocol) Start() error {
	addr, err := net.ResolveUDPAddr("udp", p.bindAddr)
	if err != nil {
		return fmt.Errorf("gossip: resolve bind address: %w", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("gossip: listen UDP: %w", err)
	}
	p.conn = conn

	p.wg.Add(2)
	go p.readLoop()
	go p.gossipLoop()

	p.logger.Info("gossip protocol started", "bind", p.bindAddr, "advertise", p.advertiseAddr)
	return nil
}

// Stop gracefully shuts down the gossip service.
func (p *Protocol) Stop() error {
	p.cancel()
	if p.conn != nil {
		p.conn.Close()
	}
	p.wg.Wait()
	p.logger.Info("gossip protocol stopped")
	return nil
}

// OnMessage registers a callback that is invoked for every incoming envelope.
// Handlers must be short‑lived and non‑blocking; they receive a copy of the envelope.
func (p *Protocol) OnMessage(handler func(*protocol.Envelope)) {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()
	p.handlers = append(p.handlers, handler)
}

// Broadcast sends an envelope to all known peers (best‑effort).
func (p *Protocol) Broadcast(msg *protocol.Envelope) error {
	if msg.SenderID == "" {
		msg.SenderID = p.nodeID
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	p.peersMu.RLock()
	defer p.peersMu.RUnlock()

	for _, peer := range p.peers {
		p.sendRaw(peer.info.Address, data)
	}
	return nil
}

// Send delivers an envelope to a specific peer by its node ID.
func (p *Protocol) Send(peerID string, msg *protocol.Envelope) error {
	p.peersMu.RLock()
	peer, ok := p.peers[peerID]
	p.peersMu.RUnlock()
	if !ok {
		return fmt.Errorf("gossip: unknown peer %s", peerID)
	}

	if msg.SenderID == "" {
		msg.SenderID = p.nodeID
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	p.sendRaw(peer.info.Address, data)
	return nil
}

// Peers returns a snapshot of the current peer list.
func (p *Protocol) Peers() []protocol.PeerInfo {
	p.peersMu.RLock()
	defer p.peersMu.RUnlock()
	list := make([]protocol.PeerInfo, 0, len(p.peers))
	for _, peer := range p.peers {
		list = append(list, peer.info)
	}
	return list
}

// AddPeer adds or updates a peer in the local membership.
func (p *Protocol) AddPeer(peer protocol.PeerInfo) {
	p.peersMu.Lock()
	defer p.peersMu.Unlock()
	if existing, ok := p.peers[peer.ID]; ok {
		existing.info = peer
		existing.lastSeen = time.Now()
	} else {
		p.peers[peer.ID] = &peerState{
			info:     peer,
			lastSeen: time.Now(),
		}
	}
}

// ---------------------------------------------------------------------------
// Internal loops
// ---------------------------------------------------------------------------

func (p *Protocol) readLoop() {
	defer p.wg.Done()
	buf := make([]byte, 65535)
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
		}
		n, remoteAddr, err := p.conn.ReadFromUDP(buf)
		if err != nil {
			if p.ctx.Err() != nil {
				return
			}
			p.logger.Error("gossip read error", "error", err)
			continue
		}

		var envelope protocol.Envelope
		if err := json.Unmarshal(buf[:n], &envelope); err != nil {
			p.logger.Warn("gossip: malformed packet", "from", remoteAddr.String())
			continue
		}

		p.handleEnvelope(&envelope, remoteAddr)
	}
}

func (p *Protocol) gossipLoop() {
	defer p.wg.Done()
	ticker := time.NewTicker(config.DefaultGossipInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.emitGossip()
			p.purgeDeadPeers()
		}
	}
}

func (p *Protocol) emitGossip() {
	msg := &protocol.GossipMessage{
		Peers: p.Peers(),
		Round: uint64(time.Now().UnixNano()),
	}
	envelope, err := protocol.NewEnvelope(protocol.TypeGossip, p.nodeID, msg)
	if err != nil {
		p.logger.Error("gossip: failed to create gossip envelope", "error", err)
		return
	}
	_ = p.Broadcast(envelope)
}

func (p *Protocol) purgeDeadPeers() {
	threshold := time.Now().Add(-config.DefaultPeerTimeout)
	p.peersMu.Lock()
	defer p.peersMu.Unlock()
	for id, peer := range p.peers {
		if peer.lastSeen.Before(threshold) {
			delete(p.peers, id)
			p.logger.Info("gossip: peer removed", "id", id)
		}
	}
}

// ---------------------------------------------------------------------------
// Packet handling
// ---------------------------------------------------------------------------

func (p *Protocol) handleEnvelope(env *protocol.Envelope, remoteAddr *net.UDPAddr) {
	switch env.Type {
	case protocol.TypeGossip:
		var gm protocol.GossipMessage
		if err := json.Unmarshal(env.Payload, &gm); err != nil {
			p.logger.Warn("gossip: bad gossip payload", "error", err)
			return
		}
		p.mergePeers(gm.Peers)

	case protocol.TypeStateSync:
		// Full state synchronisation is handled by the coordination layer;
		// we simply forward it to registered handlers.
	default:
		// Other message types are passed to handlers.
	}

	// Deliver a copy to every registered handler.
	p.peersMu.RLock()
	handlers := p.handlers
	p.peersMu.RUnlock()
	for _, h := range handlers {
		// Call in a new goroutine to avoid blocking the read loop.
		go h(env)
	}

	// Update last‑seen time for the sender if we know this peer.
	p.peersMu.Lock()
	if peer, ok := p.peers[env.SenderID]; ok {
		peer.lastSeen = time.Now()
		// Update address if it changed
		peer.info.Address = remoteAddr.String()
	}
	p.peersMu.Unlock()
}

func (p *Protocol) mergePeers(incoming []protocol.PeerInfo) {
	for _, peer := range incoming {
		if peer.ID == p.nodeID {
			continue
		}
		// Overwrite address with the one we saw (from envelope handling).
		// If the peer already exists, the lastSeen update happens in handleEnvelope.
		p.AddPeer(peer)
	}
}

func (p *Protocol) sendRaw(address string, data []byte) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		p.logger.Warn("gossip: cannot resolve peer address", "address", address, "error", err)
		return
	}
	if _, err := p.conn.WriteToUDP(data, addr); err != nil {
		p.logger.Warn("gossip: send failed", "address", address, "error", err)
	}
}
