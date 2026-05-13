package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"

	"github.com/blackswarm/siege/internal/crypto"
)

// Config holds all tunable parameters for a Siege node.
type Config struct {
	// Identity
	NodeID string `json:"node_id"`
	Mode   string `json:"mode"` // "strategic", "scanner", "hybrid"

	// Network
	GossipPort int    `json:"gossip_port"`
	HTTPPort   int    `json:"http_port"`
	BindAddr   string `json:"bind_addr"`
	AdvertiseAddr string `json:"advertise_addr,omitempty"`

	// Peers
	SeedPeers []string `json:"seed_peers,omitempty"`
	MaxPeers  int      `json:"max_peers"`

	// Scanning
	RateLimit    int      `json:"rate_limit"`
	ScanTimeout  int      `json:"scan_timeout_seconds"`
	DefaultPorts []int    `json:"default_ports,omitempty"`
	Modules      []string `json:"modules,omitempty"`

	// Genetics
	PopulationSize int     `json:"population_size"`
	Generations    int     `json:"generations"`
	MutationRate   float64 `json:"mutation_rate"`
	CrossoverRate  float64 `json:"crossover_rate"`
	EliteCount     int     `json:"elite_count"`

	// Storage
	DBPath string `json:"db_path"`

	// Secrets (never serialised directly)
	SecretKey []byte `json:"-"`
}

// configFile is the JSON‑friendly representation of Config.
type configFile struct {
	NodeID         string   `json:"node_id"`
	Mode           string   `json:"mode"`
	GossipPort     int      `json:"gossip_port"`
	HTTPPort       int      `json:"http_port"`
	BindAddr       string   `json:"bind_addr"`
	AdvertiseAddr  string   `json:"advertise_addr,omitempty"`
	SeedPeers      []string `json:"seed_peers,omitempty"`
	MaxPeers       int      `json:"max_peers"`
	RateLimit      int      `json:"rate_limit"`
	ScanTimeout    int      `json:"scan_timeout_seconds"`
	DefaultPorts   []int    `json:"default_ports,omitempty"`
	Modules        []string `json:"modules,omitempty"`
	PopulationSize int      `json:"population_size"`
	Generations    int      `json:"generations"`
	MutationRate   float64  `json:"mutation_rate"`
	CrossoverRate  float64  `json:"crossover_rate"`
	EliteCount     int      `json:"elite_count"`
	DBPath         string   `json:"db_path"`
	SecretKey      string   `json:"secret_key"` // base64-encoded
}

// LoadConfig reads and decrypts a configuration file.
func LoadConfig(path string, key []byte) (*Config, error) {
	if len(key) != 32 {
		return nil, errors.New("config: key must be 32 bytes")
	}

	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	plain, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		return nil, err
	}

	var cf configFile
	if err := json.Unmarshal(plain, &cf); err != nil {
		return nil, err
	}

	secretKey, err := base64.StdEncoding.DecodeString(cf.SecretKey)
	if err != nil {
		return nil, errors.New("config: invalid secret key encoding")
	}

	return &Config{
		NodeID:         cf.NodeID,
		Mode:           cf.Mode,
		GossipPort:     cf.GossipPort,
		HTTPPort:       cf.HTTPPort,
		BindAddr:       cf.BindAddr,
		AdvertiseAddr:  cf.AdvertiseAddr,
		SeedPeers:      cf.SeedPeers,
		MaxPeers:       cf.MaxPeers,
		RateLimit:      cf.RateLimit,
		ScanTimeout:    cf.ScanTimeout,
		DefaultPorts:   cf.DefaultPorts,
		Modules:        cf.Modules,
		PopulationSize: cf.PopulationSize,
		Generations:    cf.Generations,
		MutationRate:   cf.MutationRate,
		CrossoverRate:  cf.CrossoverRate,
		EliteCount:     cf.EliteCount,
		DBPath:         cf.DBPath,
		SecretKey:      secretKey,
	}, nil
}

// SaveConfig encrypts and writes a configuration file.
func SaveConfig(path string, cfg *Config, key []byte) error {
	if len(key) != 32 {
		return errors.New("config: key must be 32 bytes")
	}

	cf := configFile{
		NodeID:         cfg.NodeID,
		Mode:           cfg.Mode,
		GossipPort:     cfg.GossipPort,
		HTTPPort:       cfg.HTTPPort,
		BindAddr:       cfg.BindAddr,
		AdvertiseAddr:  cfg.AdvertiseAddr,
		SeedPeers:      cfg.SeedPeers,
		MaxPeers:       cfg.MaxPeers,
		RateLimit:      cfg.RateLimit,
		ScanTimeout:    cfg.ScanTimeout,
		DefaultPorts:   cfg.DefaultPorts,
		Modules:        cfg.Modules,
		PopulationSize: cfg.PopulationSize,
		Generations:    cfg.Generations,
		MutationRate:   cfg.MutationRate,
		CrossoverRate:  cfg.CrossoverRate,
		EliteCount:     cfg.EliteCount,
		DBPath:         cfg.DBPath,
		SecretKey:      base64.StdEncoding.EncodeToString(cfg.SecretKey),
	}

	plain, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return err
	}

	ciphertext, err := crypto.Encrypt(plain, key)
	if err != nil {
		return err
	}

	return os.WriteFile(path, ciphertext, 0600)
}

// DefaultConfig returns a Config populated with safe defaults.
func DefaultConfig() *Config {
	return &Config{
		NodeID:         "",
		Mode:           "scanner",
		GossipPort:     DefaultGossipPort,
		HTTPPort:       DefaultHTTPPort,
		BindAddr:       "0.0.0.0",
		MaxPeers:       DefaultMaxPeers,
		RateLimit:      DefaultRateLimit,
		ScanTimeout:    int(DefaultScanTimeout.Seconds()),
		PopulationSize: DefaultPopulationSize,
		Generations:    DefaultGenerations,
		MutationRate:   DefaultMutationRate,
		CrossoverRate:  DefaultCrossoverRate,
		EliteCount:     DefaultEliteCount,
		DBPath:         DefaultDBPath,
	}
}
