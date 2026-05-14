// Package genetics implements the genetic algorithm for scan strategy optimisation.
package genetics

import (
	"encoding/json"
	"math/rand"
)

// ScanStrategy represents a chromosome encoding a complete scan configuration.
type ScanStrategy struct {
	// Port selection
	Ports      []int   `json:"ports"`
	PortRange  [2]int  `json:"port_range"` // min, max
	TopPorts   int     `json:"top_ports"`  // number of top ports to scan

	// Timing
	RateLimit    int     `json:"rate_limit"`     // packets per second
	Timeout      float64 `json:"timeout"`        // seconds per probe
	Concurrency  int     `json:"concurrency"`    // parallel workers
	ScanDelay    float64 `json:"scan_delay"`     // delay between scans

	// Module selection
	EnableTCP       bool `json:"enable_tcp"`
	EnableUDP       bool `json:"enable_udp"`
	EnableService   bool `json:"enable_service"`
	EnableWeb       bool `json:"enable_web"`
	EnableVuln      bool `json:"enable_vuln"`
	EnableDNS       bool `json:"enable_dns"`
	EnableSubdomain bool `json:"enable_subdomain"`

	// Web scanning
	DirBustDepth   int      `json:"dir_bust_depth"`
	FuzzEnabled    bool     `json:"fuzz_enabled"`
	WordlistSize   int      `json:"wordlist_size"`

	// Evasion
	UserAgentRotation bool  `json:"user_agent_rotation"`
	IPRotation        bool  `json:"ip_rotation"`
	Jitter            float64 `json:"jitter"` // random delay factor
}

// RandomStrategy creates a new ScanStrategy with random parameters.
func RandomStrategy() *ScanStrategy {
	return &ScanStrategy{
		Ports:       randomPorts(10),
		PortRange:   [2]int{1, 65535},
		TopPorts:    100 + rand.Intn(900),
		RateLimit:   100 + rand.Intn(2000),
		Timeout:     0.5 + rand.Float64()*5.0,
		Concurrency: 10 + rand.Intn(100),
		ScanDelay:   rand.Float64() * 2.0,

		EnableTCP:       rand.Float64() < 0.9,
		EnableUDP:       rand.Float64() < 0.3,
		EnableService:   rand.Float64() < 0.8,
		EnableWeb:       rand.Float64() < 0.6,
		EnableVuln:      rand.Float64() < 0.7,
		EnableDNS:       rand.Float64() < 0.8,
		EnableSubdomain: rand.Float64() < 0.7,

		DirBustDepth:   1 + rand.Intn(3),
		FuzzEnabled:    rand.Float64() < 0.4,
		WordlistSize:   100 + rand.Intn(1000),

		UserAgentRotation: rand.Float64() < 0.5,
		IPRotation:        rand.Float64() < 0.3,
		Jitter:            rand.Float64(),
	}
}

// Clone returns a deep copy of the strategy.
func (s *ScanStrategy) Clone() *ScanStrategy {
	clone := *s
	clone.Ports = make([]int, len(s.Ports))
	copy(clone.Ports, s.Ports)
	return &clone
}

// Serialize marshals the strategy to JSON.
func (s *ScanStrategy) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize unmarshals a strategy from JSON.
func Deserialize(data []byte) (*ScanStrategy, error) {
	var s ScanStrategy
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func randomPorts(count int) []int {
	ports := make([]int, count)
	for i := range ports {
		ports[i] = 1 + rand.Intn(65535)
	}
	return ports
}
