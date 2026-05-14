package intel

import (
	"context"
	"encoding/json"
	"fmt"
  "os"
	"net/http"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// ShodanClient queries the Shodan and Censys APIs for exposed services.
type ShodanClient struct {
	client     *http.Client
	shodanKey  string
	censysID   string
	censysSecret string
}

// NewShodanClient creates a new ShodanClient.
func NewShodanClient(shodanKey, censysID, censysSecret string, timeout time.Duration) *ShodanClient {
	if shodanKey == "" {
		shodanKey = os.Getenv("SHODAN_API_KEY")
	}
	if censysID == "" {
		censysID = os.Getenv("CENSYS_API_ID")
	}
	if censysSecret == "" {
		censysSecret = os.Getenv("CENSYS_API_SECRET")
	}
	return &ShodanClient{
		client: &http.Client{
			Timeout: timeout,
		},
		shodanKey:    shodanKey,
		censysID:     censysID,
		censysSecret: censysSecret,
	}
}

// Search performs a Shodan host search for the given IPs or domains.
func (sc *ShodanClient) Search(ctx context.Context, targets []string) ([]protocol.Finding, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("targets must not be empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
	)

	for _, target := range targets {
		select {
		case <-ctx.Done():
			return findings, ctx.Err()
		default:
		}

		wg.Add(1)
		go func(t string) {
			defer wg.Done()

			results, err := sc.queryShodan(ctx, t)
			if err != nil {
				return
			}

			mu.Lock()
			findings = append(findings, results...)
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	return findings, nil
}

type shodanHostResponse struct {
	IP        string   `json:"ip_str"`
	Ports     []int    `json:"ports"`
	Hostnames []string `json:"hostnames"`
	Org       string   `json:"org"`
	OS        string   `json:"os"`
	Data      []struct {
		Port     int    `json:"port"`
		Transport string `json:"transport"`
		Product  string `json:"product"`
		Version  string `json:"version"`
	} `json:"data"`
}

func (sc *ShodanClient) queryShodan(ctx context.Context, ip string) ([]protocol.Finding, error) {
	if sc.shodanKey == "" {
		return nil, nil
	}

	url := fmt.Sprintf("https://api.shodan.io/shodan/host/%s?key=%s", ip, sc.shodanKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("shodan API returned %d", resp.StatusCode)
	}

	var host shodanHostResponse
	if err := json.NewDecoder(resp.Body).Decode(&host); err != nil {
		return nil, err
	}

	var findings []protocol.Finding

	// General host information
	findings = append(findings, protocol.Finding{
		Target:      ip,
		Protocol:    "tcp",
		Title:       fmt.Sprintf("Shodan: %s — %s", host.IP, host.Org),
		Description: fmt.Sprintf("OS: %s, Hostnames: %v, Open ports: %d", host.OS, host.Hostnames, len(host.Ports)),
		Severity:    "info",
		Timestamp:   time.Now().UnixMilli(),
	})

	// Service information
	for _, data := range host.Data {
		findings = append(findings, protocol.Finding{
			Target:      ip,
			Port:        data.Port,
			Protocol:    data.Transport,
			Service:     data.Product,
			Title:       fmt.Sprintf("Shodan service: %s %s", data.Product, data.Version),
			Description: fmt.Sprintf("Port %d/%s: %s %s", data.Port, data.Transport, data.Product, data.Version),
			Severity:    "info",
			Timestamp:   time.Now().UnixMilli(),
		})
	}

	return findings, nil
}
