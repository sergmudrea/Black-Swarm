package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// Fuzzer performs parameter fuzzing on web endpoints.
type Fuzzer struct {
	client      *http.Client
	payloads    []string
	concurrency int
}

// NewFuzzer creates a new Fuzzer with the given payloads and concurrency.
func NewFuzzer(payloads []string, timeout time.Duration, concurrency int) *Fuzzer {
	if concurrency <= 0 {
		concurrency = 10
	}
	return &Fuzzer{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		payloads:    payloads,
		concurrency: concurrency,
	}
}

// DefaultPayloads returns a small set of fuzzing payloads.
func DefaultPayloads() []string {
	return []string{
		"' OR '1'='1",
		"<script>alert(1)</script>",
		"../../etc/passwd",
		"| whoami",
		"; cat /etc/passwd",
		"`id`",
		"$(id)",
		"{{7*7}}",
		"${7*7}",
		"<img src=x onerror=alert(1)>",
	}
}

// Scan performs fuzzing against the given endpoints.
// endpoints is a map of base URL to a list of parameter names to fuzz.
func (f *Fuzzer) Scan(ctx context.Context, endpoints map[string][]string) ([]protocol.Finding, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("endpoints must not be empty")
	}
	if len(f.payloads) == 0 {
		return nil, fmt.Errorf("payloads list is empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
		sem      = make(chan struct{}, f.concurrency)
	)

	for baseURL, params := range endpoints {
		for _, param := range params {
			for _, payload := range f.payloads {
				select {
				case <-ctx.Done():
					return findings, ctx.Err()
				default:
				}

				wg.Add(1)
				sem <- struct{}{}
				go func(base, p, pl string) {
					defer wg.Done()
					defer func() { <-sem }()

					testURL := fmt.Sprintf("%s?%s=%s", base, p, url.QueryEscape(pl))
					req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
					if err != nil {
						return
					}

					resp, err := f.client.Do(req)
					if err != nil {
						return
					}
					defer resp.Body.Close()

					// Check for common vulnerability indicators
					indicators := f.checkIndicators(resp)
					for _, indicator := range indicators {
						finding := protocol.Finding{
							Target:      base,
							Protocol:    "http",
							Title:       fmt.Sprintf("Potential vulnerability: %s", indicator),
							Description: fmt.Sprintf("Parameter: %s, Payload: %s", p, pl),
							Severity:    "medium",
							Evidence:    testURL,
							Timestamp:   time.Now().UnixMilli(),
						}
						mu.Lock()
						findings = append(findings, finding)
						mu.Unlock()
					}
				}(baseURL, param, payload)
			}
		}
	}

	wg.Wait()
	return findings, nil
}

func (f *Fuzzer) checkIndicators(resp *http.Response) []string {
	var indicators []string

	// Check for SQL error messages
	sqlErrors := []string{
		"SQL syntax", "mysql_fetch", "ORA-", "PostgreSQL", "SQLite",
		"Unclosed quotation mark", "Microsoft OLE DB",
	}
	body := make([]byte, 4096)
	n, _ := resp.Body.Read(body)
	bodyStr := strings.ToLower(string(body[:n]))

	for _, err := range sqlErrors {
		if strings.Contains(bodyStr, strings.ToLower(err)) {
			indicators = append(indicators, "Possible SQL Injection")
			break
		}
	}

	// Check for XSS reflection
	payloads := []string{"<script>alert(1)</script>", "<img src=x onerror=alert(1)>"}
	for _, payload := range payloads {
		if strings.Contains(bodyStr, payload) {
			indicators = append(indicators, "Reflected XSS")
			break
		}
	}

	return indicators
}
