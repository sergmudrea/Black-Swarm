package port

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// UDPScanner performs UDP port scanning.
type UDPScanner struct {
	timeout     time.Duration
	concurrency int
}

// NewUDPScanner creates a new UDP scanner with the given parameters.
func NewUDPScanner(timeout time.Duration, concurrency int) *UDPScanner {
	if concurrency <= 0 {
		concurrency = 50
	}
	return &UDPScanner{
		timeout:     timeout,
		concurrency: concurrency,
	}
}

// Scan performs a UDP scan against the specified targets and ports.
func (s *UDPScanner) Scan(ctx context.Context, targets []string, ports []int) ([]protocol.Finding, error) {
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
				conn, err := net.DialTimeout("udp", address, s.timeout)
				if err != nil {
					return
				}
				defer conn.Close()

				// Send a simple probe
				conn.SetWriteDeadline(time.Now().Add(s.timeout))
				conn.Write([]byte{0})

				// Try to read a response
				buf := make([]byte, 512)
				conn.SetReadDeadline(time.Now().Add(s.timeout))
				n, _, readErr := conn.ReadFromUDP(buf)

				finding := protocol.Finding{
					Target:    t,
					Port:      p,
					Protocol:  "udp",
					Title:     fmt.Sprintf("Open UDP port %d", p),
					Severity:  "info",
					Timestamp: time.Now().UnixMilli(),
				}

				if readErr == nil && n > 0 {
					finding.Description = fmt.Sprintf("Received %d bytes", n)
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
