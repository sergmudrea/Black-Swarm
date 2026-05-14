package evasion

import (
	"math/rand"
	"net"
	"sync"
)

// IPRotator manages a pool of source IP addresses and rotates them for outgoing
// connections, helping to avoid IP‑based rate limiting.
type IPRotator struct {
	mu      sync.Mutex
	ips     []net.IP
	current int
}

// NewIPRotator creates an IPRotator with the given list of IP addresses.
// Each address must be a valid IPv4 or IPv6 string.
func NewIPRotator(addrs []string) *IPRotator {
	ips := make([]net.IP, 0, len(addrs))
	for _, a := range addrs {
		if ip := net.ParseIP(a); ip != nil {
			ips = append(ips, ip)
		}
	}
	return &IPRotator{
		ips:     ips,
		current: rand.Intn(len(ips)),
	}
}

// Next returns the next IP address from the pool in round‑robin order.
func (r *IPRotator) Next() net.IP {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.ips) == 0 {
		return nil
	}
	ip := r.ips[r.current]
	r.current = (r.current + 1) % len(r.ips)
	return ip
}

// Random returns a random IP from the pool.
func (r *IPRotator) Random() net.IP {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.ips) == 0 {
		return nil
	}
	idx := rand.Intn(len(r.ips))
	return r.ips[idx]
}

// Add adds a new IP to the pool.
func (r *IPRotator) Add(addr string) {
	if ip := net.ParseIP(addr); ip != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.ips = append(r.ips, ip)
	}
}
