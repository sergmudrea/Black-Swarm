package port

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// ServiceBanner contains the result of a service detection probe.
type ServiceBanner struct {
	Port    int
	Service string
	Banner  string
}

// ServiceDetector probes open ports to identify running services.
type ServiceDetector struct {
	timeout     time.Duration
	concurrency int
}

// NewServiceDetector creates a new ServiceDetector.
func NewServiceDetector(timeout time.Duration, concurrency int) *ServiceDetector {
	if concurrency <= 0 {
		concurrency = 50
	}
	return &ServiceDetector{
		timeout:     timeout,
		concurrency: concurrency,
	}
}

// Detect takes a list of targets with known open ports and attempts to identify
// the service running on each port.
func (sd *ServiceDetector) Detect(ctx context.Context, targets []string, openPorts map[string][]int) ([]protocol.Finding, error) {
	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
		sem      = make(chan struct{}, sd.concurrency)
	)

	for _, target := range targets {
		ports, exists := openPorts[target]
		if !exists {
			continue
		}
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

				banner := sd.grabBanner(t, p)
				if banner == nil {
					return
				}

				finding := protocol.Finding{
					Target:      t,
					Port:        p,
					Protocol:    "tcp",
					Service:     banner.Service,
					Title:       fmt.Sprintf("Service detected: %s", banner.Service),
					Description: fmt.Sprintf("Banner: %s", banner.Banner),
					Severity:    "info",
					Timestamp:   time.Now().UnixMilli(),
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

func (sd *ServiceDetector) grabBanner(target string, port int) *ServiceBanner {
	address := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("tcp", address, sd.timeout)
	if err != nil {
		return nil
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(sd.timeout))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil
	}

	banner := strings.TrimSpace(string(buf[:n]))
	if banner == "" {
		return nil
	}

	service := identifyService(port, banner)
	return &ServiceBanner{
		Port:    port,
		Service: service,
		Banner:  banner,
	}
}

func identifyService(port int, banner string) string {
	bannerUpper := strings.ToUpper(banner)

	switch {
	case strings.Contains(bannerUpper, "SSH"):
		return "ssh"
	case strings.Contains(bannerUpper, "HTTP"):
		return "http"
	case strings.Contains(bannerUpper, "HTTPS") || strings.Contains(bannerUpper, "SSL"):
		return "https"
	case strings.Contains(bannerUpper, "FTP"):
		return "ftp"
	case strings.Contains(bannerUpper, "SMTP"):
		return "smtp"
	case strings.Contains(bannerUpper, "MYSQL"):
		return "mysql"
	case strings.Contains(bannerUpper, "POSTGRESQL"):
		return "postgresql"
	case strings.Contains(bannerUpper, "REDIS"):
		return "redis"
	case strings.Contains(bannerUpper, "MONGODB"):
		return "mongodb"
	case port == 3389 || strings.Contains(bannerUpper, "RDP"):
		return "rdp"
	case port == 445 || strings.Contains(bannerUpper, "SMB"):
		return "smb"
	default:
		return "unknown"
	}
}
