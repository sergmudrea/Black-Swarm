package gossip

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"time"
  "encoding/json"

	"github.com/blackswarm/siege/internal/protocol"
)

// Discovery handles automatic peer discovery using UDP multicast and seed nodes.
type Discovery struct {
	proto *Protocol

	// Seed nodes (static list from configuration)
	seeds []string

	// Multicast configuration
	multicastAddr string
	multicastTTL  int

	// Periodic discovery ticker
	interval time.Duration

	logger *slog.Logger
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDiscovery creates a new Discovery instance.
func NewDiscovery(proto *Protocol, seeds []string, multicastAddr string, logger *slog.Logger) *Discovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &Discovery{
		proto:         proto,
		seeds:         seeds,
		multicastAddr: multicastAddr,
		multicastTTL:  1,
		interval:      30 * time.Second,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the discovery process: connects to seed nodes and listens for multicast.
func (d *Discovery) Start() error {
	// Connect to seed nodes immediately
	for _, seed := range d.seeds {
		d.addSeed(seed)
	}

	d.wg.Add(1)
	go d.discoveryLoop()

	d.logger.Info("peer discovery started", "seeds", len(d.seeds))
	return nil
}

// Stop terminates the discovery service.
func (d *Discovery) Stop() error {
	d.cancel()
	d.wg.Wait()
	return nil
}

// addSeed sends a PeerHello to the given seed address.
func (d *Discovery) addSeed(address string) {
	// Create a temporary peer entry so the seed can be contacted
	seedPeer := protocol.PeerInfo{
		ID:      address, // temporary ID; the real ID is learned from the hello response
		Address: address,
		Mode:    "unknown",
	}
	d.proto.AddPeer(seedPeer)

	hello := &protocol.GossipMessage{
		Peers: d.proto.Peers(),
	}
	env, err := protocol.NewEnvelope(protocol.TypePeerHello, d.proto.nodeID, hello)
	if err != nil {
		d.logger.Error("discovery: failed to create hello", "error", err)
		return
	}
	_ = d.proto.Send(address, env)
}

func (d *Discovery) discoveryLoop() {
	defer d.wg.Done()
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.probeSeeds()
			d.sendMulticast()
		}
	}
}

func (d *Discovery) probeSeeds() {
	for _, seed := range d.seeds {
		d.addSeed(seed)
	}
}

func (d *Discovery) sendMulticast() {
	if d.multicastAddr == "" {
		return
	}

	addr, err := net.ResolveUDPAddr("udp", d.multicastAddr)
	if err != nil {
		d.logger.Warn("discovery: multicast resolve failed", "addr", d.multicastAddr, "error", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		d.logger.Warn("discovery: multicast dial failed", "error", err)
		return
	}
	defer conn.Close()

	hello := &protocol.GossipMessage{
		Peers: d.proto.Peers(),
	}
	env, _ := protocol.NewEnvelope(protocol.TypePeerHello, d.proto.nodeID, hello)
	data, _ := json.Marshal(env)
	conn.Write(data)
}
