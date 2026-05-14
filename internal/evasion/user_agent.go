package evasion

import (
	"math/rand"
	"sync"
)

// UserAgentRotator manages a pool of User-Agent strings and selects them
// randomly or sequentially to avoid fingerprinting based on static headers.
type UserAgentRotator struct {
	mu         sync.Mutex
	agents     []string
	current    int
	sequential bool
}

// DefaultUserAgents returns a curated list of recent, diverse User-Agent strings.
func DefaultUserAgents() []string {
	return []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64; rv:127.0) Gecko/20100101 Firefox/127.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:127.0) Gecko/20100101 Firefox/127.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/125.0.0.0",
	}
}

// NewUserAgentRotator creates a new UserAgentRotator with the given list.
// If the list is empty, DefaultUserAgents is used.
func NewUserAgentRotator(agents []string) *UserAgentRotator {
	if len(agents) == 0 {
		agents = DefaultUserAgents()
	}
	return &UserAgentRotator{
		agents: agents,
		current: rand.Intn(len(agents)),
	}
}

// Next returns the next User-Agent in round‑robin order.
func (r *UserAgentRotator) Next() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.agents) == 0 {
		return ""
	}
	ua := r.agents[r.current]
	r.current = (r.current + 1) % len(r.agents)
	return ua
}

// Random returns a random User-Agent from the pool.
func (r *UserAgentRotator) Random() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.agents) == 0 {
		return ""
	}
	idx := rand.Intn(len(r.agents))
	return r.agents[idx]
}

// Add appends a new User-Agent string to the pool.
func (r *UserAgentRotator) Add(ua string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents = append(r.agents, ua)
}
