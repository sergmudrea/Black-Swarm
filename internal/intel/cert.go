package intel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// CertScanner queries Certificate Transparency logs for subdomains.
type CertScanner struct {
	client *http.Client
}

// NewCertScanner creates a new CertScanner.
func NewCertScanner(timeout time.Duration) *CertScanner {
	return &CertScanner{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Search queries crt.sh for certificates issued to the given domains.
func (cs *CertScanner) Search(ctx context.Context, domains []string) ([]protocol.Finding, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("domains must not be empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
	)

	for _, domain := range domains {
		select {
		case <-ctx.Done():
			return findings, ctx.Err()
		default:
		}

		wg.Add(1)
		go func(d string) {
			defer wg.Done()

			results, err := cs.queryCrtSh(ctx, d)
			if err != nil {
				return
			}

			seen := make(map[string]bool)
			for _, entry := range results {
				name := entry.NameValue
				if seen[name] || name == "" {
					continue
				}
				seen[name] = true

				finding := protocol.Finding{
					Target:      d,
					Protocol:    "https",
					Title:       fmt.Sprintf("Certificate Transparency: %s", name),
					Description: fmt.Sprintf("Found in CT logs for %s (issuer: %s)", d, entry.IssuerName),
					Severity:    "info",
					Timestamp:   time.Now().UnixMilli(),
				}

				mu.Lock()
				findings = append(findings, finding)
				mu.Unlock()
			}
		}(domain)
	}

	wg.Wait()
	return findings, nil
}

type crtShEntry struct {
	NameValue  string `json:"name_value"`
	IssuerName string `json:"issuer_name"`
}

func (cs *CertScanner) queryCrtSh(ctx context.Context, domain string) ([]crtShEntry, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%.%s&output=json", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := cs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh returned %d", resp.StatusCode)
	}

	var entries []crtShEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	return entries, nil
}
