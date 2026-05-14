package genetics

import "math/rand"

// Mutator applies random changes to a scan strategy.
type Mutator struct {
	rate float64 // probability of mutating each gene
}

// NewMutator creates a new Mutator with the given mutation rate.
func NewMutator(rate float64) *Mutator {
	return &Mutator{rate: rate}
}

// Mutate applies mutation operators to the given strategy in-place.
func (m *Mutator) Mutate(s *ScanStrategy) {
	if rand.Float64() < m.rate {
		m.mutatePorts(s)
	}
	if rand.Float64() < m.rate {
		m.mutateRateLimit(s)
	}
	if rand.Float64() < m.rate {
		m.mutateTimeout(s)
	}
	if rand.Float64() < m.rate {
		m.mutateConcurrency(s)
	}
	if rand.Float64() < m.rate {
		m.mutateModules(s)
	}
	if rand.Float64() < m.rate {
		m.mutateWeb(s)
	}
	if rand.Float64() < m.rate {
		m.mutateEvasion(s)
	}
}

func (m *Mutator) mutatePorts(s *ScanStrategy) {
	// Add or remove a port
	if rand.Float64() < 0.5 && len(s.Ports) > 1 {
		idx := rand.Intn(len(s.Ports))
		s.Ports = append(s.Ports[:idx], s.Ports[idx+1:]...)
	} else {
		newPort := 1 + rand.Intn(65535)
		s.Ports = append(s.Ports, newPort)
	}
}

func (m *Mutator) mutateRateLimit(s *ScanStrategy) {
	delta := rand.Intn(200) - 100
	s.RateLimit += delta
	if s.RateLimit < 10 {
		s.RateLimit = 10
	}
	if s.RateLimit > 2000 {
		s.RateLimit = 2000
	}
}

func (m *Mutator) mutateTimeout(s *ScanStrategy) {
	delta := rand.Float64()*0.5 - 0.25
	s.Timeout += delta
	if s.Timeout < 0.1 {
		s.Timeout = 0.1
	}
	if s.Timeout > 10.0 {
		s.Timeout = 10.0
	}
}

func (m *Mutator) mutateConcurrency(s *ScanStrategy) {
	delta := rand.Intn(20) - 10
	s.Concurrency += delta
	if s.Concurrency < 1 {
		s.Concurrency = 1
	}
	if s.Concurrency > 200 {
		s.Concurrency = 200
	}
}

func (m *Mutator) mutateModules(s *ScanStrategy) {
	switch rand.Intn(7) {
	case 0:
		s.EnableTCP = !s.EnableTCP
	case 1:
		s.EnableUDP = !s.EnableUDP
	case 2:
		s.EnableService = !s.EnableService
	case 3:
		s.EnableWeb = !s.EnableWeb
	case 4:
		s.EnableVuln = !s.EnableVuln
	case 5:
		s.EnableDNS = !s.EnableDNS
	case 6:
		s.EnableSubdomain = !s.EnableSubdomain
	}
}

func (m *Mutator) mutateWeb(s *ScanStrategy) {
	if rand.Float64() < 0.5 {
		s.DirBustDepth = 1 + rand.Intn(5)
	} else {
		s.FuzzEnabled = !s.FuzzEnabled
	}
}

func (m *Mutator) mutateEvasion(s *ScanStrategy) {
	switch rand.Intn(3) {
	case 0:
		s.UserAgentRotation = !s.UserAgentRotation
	case 1:
		s.IPRotation = !s.IPRotation
	case 2:
		s.Jitter = rand.Float64()
	}
}
