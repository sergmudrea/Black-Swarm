package intel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// GitHubDorker performs GitHub code search for exposed secrets and sensitive files.
type GitHubDorker struct {
	client  *http.Client
	token   string
	queries []string
}

// NewGitHubDorker creates a new GitHubDorker.
func NewGitHubDorker(token string, timeout time.Duration) *GitHubDorker {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	return &GitHubDorker{
		client: &http.Client{
			Timeout: timeout,
		},
		token:   token,
		queries: defaultDorkQueries(),
	}
}

func defaultDorkQueries() []string {
	return []string{
		`"%s" password`,
		`"%s" secret`,
		`"%s" api_key`,
		`"%s" private_key`,
		`"%s" token`,
		`"%s" .env`,
		`"%s" config.yml`,
		`"%s" database.yml`,
	}
}

// Search performs GitHub dorking for the given targets.
func (gd *GitHubDorker) Search(ctx context.Context, targets []string) ([]protocol.Finding, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("targets must not be empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
	)

	for _, target := range targets {
		for _, queryTemplate := range gd.queries {
			select {
			case <-ctx.Done():
				return findings, ctx.Err()
			default:
			}

			wg.Add(1)
			go func(t, q string) {
				defer wg.Done()

				query := fmt.Sprintf(q, t)
				results, err := gd.searchGitHub(ctx, query)
				if err != nil {
					return
				}

				for _, result := range results {
					finding := protocol.Finding{
						Target:      t,
						Protocol:    "https",
						Title:       fmt.Sprintf("GitHub exposure: %s", query),
						Description: fmt.Sprintf("File: %s — %s", result.Path, result.URL),
						Severity:    "high",
						Evidence:    result.URL,
						Timestamp:   time.Now().UnixMilli(),
					}

					mu.Lock()
					findings = append(findings, finding)
					mu.Unlock()
				}
			}(target, queryTemplate)
		}
	}

	wg.Wait()
	return findings, nil
}

type githubSearchResult struct {
	Path string `json:"path"`
	URL  string `json:"html_url"`
}

func (gd *GitHubDorker) searchGitHub(ctx context.Context, query string) ([]githubSearchResult, error) {
	encodedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://api.github.com/search/code?q=%s&per_page=10", encodedQuery)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if gd.token != "" {
		req.Header.Set("Authorization", "Bearer "+gd.token)
	}

	resp, err := gd.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}

	var searchResp struct {
		Items []githubSearchResult `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	return searchResp.Items, nil
}
