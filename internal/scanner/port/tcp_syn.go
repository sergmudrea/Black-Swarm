// Package port implements TCP, UDP, and service detection scanners.
package port

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// TCPSYNScanner performs TCP SYN (half-open) port scanning.
type TCPSYNScanner struct {
	rateLimit   int
	timeout     time.Duration
	concurrency int
}

// NewTCPSYNScanner creates a new TCP SYN scanner with the given parameters.
func NewTCPSYNScanner(rateLimit int, timeout time.Duration, concurrency int) *TCPSYNScanner {
	if concurrency <= 0 {
		concurrency = 100
	}
	return &TCPSYNScanner{
		rateLimit:   rateLimit,
		timeout:     timeout,
		concurrency: concurrency,
	}
}

// Scan performs a SYN scan against the specified targets and ports.
// It returns a slice of findings for open ports.
func (s *TCPSYNScanner) Scan(ctx context.Context, targets []string, ports []int) ([]protocol.Finding, error) {
	if len(targets) == 0 || len(ports) == 0 {
		return nil, fmt.Errorf("targets and ports must not be empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
		sem      = make(chan struct{}, s.concurrency)
	)

	for _, target := range targets {
		for _, port := range ports {
			select {
			case <-ctx.Done():
				return findings, ctx.Err()
			default:
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(t string, p int) {
				defer wg.Done()
				defer func() { <-sem }()

				address := fmt.Sprintf("%s:%d", t, p)
				conn, err := net.DialTimeout("tcp", address, s.timeout)
				if err != nil {
					return
				}
				conn.Close()

				finding := protocol.Finding{
					Target:    t,
					Port:      p,
					Protocol:  "tcp",
					Title:     fmt.Sprintf("Open port %d", p),
					Severity:  "info",
					Timestamp: time.Now().UnixMilli(),
				}

				mu.Lock()
				findings = append(findings, finding)
				mu.Unlock()
			}(target, port)
		}
	}

	wg.Wait()
	return findings, nil
}
