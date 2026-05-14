package genetics

import "math/rand"

// CrossoverOperator combines two parent strategies to produce a child.
type CrossoverOperator struct {
	rate float64
}

// NewCrossoverOperator creates a new CrossoverOperator with the given rate.
func NewCrossoverOperator(rate float64) *CrossoverOperator {
	return &CrossoverOperator{rate: rate}
}

// Crossover performs uniform crossover between two parents.
// Returns a new child strategy.
func (c *CrossoverOperator) Crossover(parent1, parent2 *ScanStrategy) *ScanStrategy {
	child := &ScanStrategy{}

	if rand.Float64() < c.rate {
		// Uniform crossover: for each gene, randomly choose from either parent
		if rand.Float64() < 0.5 {
			child.Ports = make([]int, len(parent1.Ports))
			copy(child.Ports, parent1.Ports)
		} else {
			child.Ports = make([]int, len(parent2.Ports))
			copy(child.Ports, parent2.Ports)
		}

		child.PortRange = parent1.PortRange
		if rand.Float64() < 0.5 {
			child.PortRange = parent2.PortRange
		}

		child.TopPorts = parent1.TopPorts
		if rand.Float64() < 0.5 {
			child.TopPorts = parent2.TopPorts
		}

		child.RateLimit = parent1.RateLimit
		if rand.Float64() < 0.5 {
			child.RateLimit = parent2.RateLimit
		}

		child.Timeout = parent1.Timeout
		if rand.Float64() < 0.5 {
			child.Timeout = parent2.Timeout
		}

		child.Concurrency = parent1.Concurrency
		if rand.Float64() < 0.5 {
			child.Concurrency = parent2.Concurrency
		}

		child.ScanDelay = parent1.ScanDelay
		if rand.Float64() < 0.5 {
			child.ScanDelay = parent2.ScanDelay
		}

		child.EnableTCP = parent1.EnableTCP
		if rand.Float64() < 0.5 {
			child.EnableTCP = parent2.EnableTCP
		}
		child.EnableUDP = parent1.EnableUDP
		if rand.Float64() < 0.5 {
			child.EnableUDP = parent2.EnableUDP
		}
		child.EnableService = parent1.EnableService
		if rand.Float64() < 0.5 {
			child.EnableService = parent2.EnableService
		}
		child.EnableWeb = parent1.EnableWeb
		if rand.Float64() < 0.5 {
			child.EnableWeb = parent2.EnableWeb
		}
		child.EnableVuln = parent1.EnableVuln
		if rand.Float64() < 0.5 {
			child.EnableVuln = parent2.EnableVuln
		}
		child.EnableDNS = parent1.EnableDNS
		if rand.Float64() < 0.5 {
			child.EnableDNS = parent2.EnableDNS
		}
		child.EnableSubdomain = parent1.EnableSubdomain
		if rand.Float64() < 0.5 {
			child.EnableSubdomain = parent2.EnableSubdomain
		}

		child.DirBustDepth = parent1.DirBustDepth
		if rand.Float64() < 0.5 {
			child.DirBustDepth = parent2.DirBustDepth
		}
		child.FuzzEnabled = parent1.FuzzEnabled
		if rand.Float64() < 0.5 {
			child.FuzzEnabled = parent2.FuzzEnabled
		}
		child.WordlistSize = parent1.WordlistSize
		if rand.Float64() < 0.5 {
			child.WordlistSize = parent2.WordlistSize
		}

		child.UserAgentRotation = parent1.UserAgentRotation
		if rand.Float64() < 0.5 {
			child.UserAgentRotation = parent2.UserAgentRotation
		}
		child.IPRotation = parent1.IPRotation
		if rand.Float64() < 0.5 {
			child.IPRotation = parent2.IPRotation
		}
		child.Jitter = parent1.Jitter
		if rand.Float64() < 0.5 {
			child.Jitter = parent2.Jitter
		}
	} else {
		// No crossover: just clone one parent
		child = parent1.Clone()
	}

	return child
}
