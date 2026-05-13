// Package web implements web application scanners (directory busting, fuzzing).
package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// DirBuster performs directory and file enumeration on web servers.
type DirBuster struct {
	client      *http.Client
	wordlist    []string
	concurrency int
}

// NewDirBuster creates a new DirBuster with the given wordlist and concurrency.
func NewDirBuster(wordlist []string, timeout time.Duration, concurrency int) *DirBuster {
	if concurrency <= 0 {
		concurrency = 20
	}
	return &DirBuster{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		wordlist:    wordlist,
		concurrency: concurrency,
	}
}

// DefaultWordlist returns a small built-in wordlist for quick scans.
func DefaultWordlist() []string {
	return []string{
		"admin", "login", "wp-admin", "administrator",
		"api", "api/v1", "graphql",
		"backup", "bak", "old",
		"config", "configuration",
		"dashboard", "panel",
		"dev", "development", "staging", "test",
		"index.php", "index.html", "index.jsp",
		"robots.txt", "sitemap.xml",
		".git", ".svn", ".env", ".htaccess",
		"phpmyadmin", "phpinfo.php",
		"uploads", "downloads", "files",
		"static", "assets", "css", "js", "images",
	}
}

// Scan performs directory busting against the given base URLs.
func (db *DirBuster) Scan(ctx context.Context, baseURLs []string) ([]protocol.Finding, error) {
	if len(baseURLs) == 0 {
		return nil, fmt.Errorf("base URLs must not be empty")
	}
	if len(db.wordlist) == 0 {
		return nil, fmt.Errorf("wordlist is empty")
	}

	var (
		findings []protocol.Finding
		mu       sync.Mutex
		wg       sync.WaitGroup
		sem      = make(chan struct{}, db.concurrency)
	)

	for _, baseURL := range baseURLs {
		for _, word := range db.wordlist {
			select {
			case <-ctx.Done():
				return findings, ctx.Err()
			default:
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(base, path string) {
				defer wg.Done()
				defer func() { <-sem }()

				url := fmt.Sprintf("%s/%s", base, path)
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				if err != nil {
					return
				}

				resp, err := db.client.Do(req)
				if err != nil {
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					finding := protocol.Finding{
						Target:      base,
						Protocol:    "http",
						Title:       fmt.Sprintf("Directory found: %s", path),
						Description: fmt.Sprintf("URL: %s — HTTP %d", url, resp.StatusCode),
						Severity:    "info",
						Timestamp:   time.Now().UnixMilli(),
					}
					mu.Lock()
					findings = append(findings, finding)
					mu.Unlock()
				}
			}(baseURL, word)
		}
	}

	wg.Wait()
	return findings, nil
}
