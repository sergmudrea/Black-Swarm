package recon

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// ... остальной код без изменений

// SubdomainScanner discovers subdomains using a wordlist.
type SubdomainScanner struct {
	wordlist    []string
	concurrency int
	timeout     time.Duration
}

// NewSubdomainScanner creates a new SubdomainScanner.
func NewSubdomainScanner(wordlist []string, concurrency int, timeout time.Duration) *SubdomainScanner {
	if concurrency <= 0 {
		concurrency = 50
	}
	return &SubdomainScanner{
		wordlist:    wordlist,
		concurrency: concurrency,
		timeout:     timeout,
	}
}

// DefaultSubdomainWordlist returns a small built-in wordlist.
func DefaultSubdomainWordlist() []string {
	return []string{
		"www", "mail", "ftp", "smtp", "pop", "imap",
		"admin", "api", "dev", "staging", "test",
		"blog", "shop", "store", "portal",
		"vpn", "remote", "gateway",
		"db", "database", "mysql", "postgres",
		"jenkins", "gitlab", "github", "bitbucket",
		"jira", "confluence", "wiki",
		"monitor", "monitoring", "grafana", "kibana",
		"docker", "kubernetes", "k8s",
		"cdn", "static", "assets", "media",
	}
}

// Scan performs subdomain discovery on the given domains.
func (ss *SubdomainScanner) Scan(ctx context.Context, domains []string) ([]protocol.Finding, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("domains must not be empty")
	}
	if len(ss.wordlist) == 0 {
		return nil, fmt.Errorf("wordlist is empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
		sem      = make(chan struct{}, ss.concurrency)
	)

	for _, domain := range domains {
		for _, prefix := range ss.wordlist {
			select {
			case <-ctx.Done():
				return findings, ctx.Err()
			default:
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(d, p string) {
				defer wg.Done()
				defer func() { <-sem }()

				subdomain := fmt.Sprintf("%s.%s", p, d)
				ips, err := net.LookupIP(subdomain)
				if err != nil || len(ips) == 0 {
					return
				}

				ipStrs := make([]string, len(ips))
				for i, ip := range ips {
					ipStrs[i] = ip.String()
				}

				finding := protocol.Finding{
					Target:      d,
					Protocol:    "dns",
					Title:       fmt.Sprintf("Subdomain found: %s", subdomain),
					Description: fmt.Sprintf("Resolved to: %s", strings.Join(ipStrs, ", ")),
					Severity:    "info",
					Timestamp:   time.Now().UnixMilli(),
				}

				mu.Lock()
				findings = append(findings, finding)
				mu.Unlock()
			}(domain, prefix)
		}
	}

	wg.Wait()
	return findings, nil
}
