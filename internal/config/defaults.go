// Package config provides configuration constants and loading/saving functions.
package config

import "time"

const (
	// DefaultGossipPort is the UDP port used for peer discovery and gossip.
	DefaultGossipPort = 7946

	// DefaultHTTPPort is the TCP port for the embedded HTTP/WebSocket server.
	DefaultHTTPPort = 8443

	// DefaultScanTimeout is the maximum duration for a single scan task.
	DefaultScanTimeout = 30 * time.Minute

	// DefaultGossipInterval is the period between gossip rounds.
	DefaultGossipInterval = 5 * time.Second

	// DefaultPeerTimeout is the duration after which a silent peer is considered dead.
	DefaultPeerTimeout = 30 * time.Second

	// DefaultMaxPeers is the maximum number of peers a node will maintain.
	DefaultMaxPeers = 50

	// DefaultRateLimit is the maximum packets per second for scanning.
	DefaultRateLimit = 1000

	// DefaultPopulationSize is the number of chromosomes in the genetic algorithm.
	DefaultPopulationSize = 100

	// DefaultGenerations is the number of generations per evolution run.
	DefaultGenerations = 50

	// DefaultMutationRate is the initial mutation probability.
	DefaultMutationRate = 0.05

	// DefaultCrossoverRate is the probability of crossover between two parents.
	DefaultCrossoverRate = 0.7

	// DefaultEliteCount is the number of top chromosomes preserved unchanged.
	DefaultEliteCount = 5

	// DefaultDBPath is the SQLite database file for strategic nodes.
	DefaultDBPath = "./siege.db"
)

// DefaultPortRanges defines the standard port ranges for scanning.
var DefaultPortRanges = []string{
	"1-1024",
	"3306",
	"3389",
	"5432",
	"6379",
	"8080",
	"8443",
	"27017",
}

// DefaultModules lists the scanning modules enabled by default.
var DefaultModules = []string{
	"tcp_syn",
	"service_detect",
	"dns_recon",
	"subdomain",
	"cve_match",
}
