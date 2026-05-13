// Package vuln implements vulnerability detection modules.
package vuln

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/blackswarm/siege/internal/protocol"
)

// CVEEntry represents a known CVE with its matching criteria.
type CVEEntry struct {
	CVE         string
	CVSS        float64
	Severity    string
	Description string
	Service     string
	Version     string
	Pattern     string
}

// CVEMatcher matches service banners against known CVEs.
type CVEMatcher struct {
	database []CVEEntry
}

// NewCVEMatcher creates a new CVEMatcher with a built-in database.
func NewCVEMatcher() *CVEMatcher {
	return &CVEMatcher{
		database: loadCVEDatabase(),
	}
}

func loadCVEDatabase() []CVEEntry {
	return []CVEEntry{
		{
			CVE:         "CVE-2021-44228",
			CVSS:        10.0,
			Severity:    "critical",
			Description: "Log4Shell — Remote code execution in Apache Log4j2",
			Service:     "http",
			Pattern:     "Apache",
		},
		{
			CVE:         "CVE-2017-5638",
			CVSS:        10.0,
			Severity:    "critical",
			Description: "Apache Struts2 Remote Code Execution",
			Service:     "http",
			Pattern:     "Struts",
		},
		{
			CVE:         "CVE-2019-0708",
			CVSS:        9.8,
			Severity:    "critical",
			Description: "BlueKeep — Remote Desktop Services Remote Code Execution",
			Service:     "rdp",
			Pattern:     "RDP",
		},
		{
			CVE:         "CVE-2020-0796",
			CVSS:        10.0,
			Severity:    "critical",
			Description: "SMBGhost — SMBv3 Remote Code Execution",
			Service:     "smb",
			Pattern:     "SMB",
		},
		{
			CVE:         "CVE-2021-41773",
			CVSS:        7.5,
			Severity:    "high",
			Description: "Apache HTTP Server Path Traversal",
			Service:     "http",
			Pattern:     "Apache",
		},
		{
			CVE:         "CVE-2022-22965",
			CVSS:        9.8,
			Severity:    "critical",
			Description: "Spring4Shell — Spring Framework RCE",
			Service:     "http",
			Pattern:     "Spring",
		},
		{
			CVE:         "CVE-2023-23397",
			CVSS:        9.8,
			Severity:    "critical",
			Description: "Microsoft Outlook Privilege Escalation",
			Service:     "smtp",
			Pattern:     "Microsoft",
		},
	}
}

// Match takes service findings and returns potential CVE matches.
func (m *CVEMatcher) Match(ctx context.Context, serviceFindings []protocol.Finding) ([]protocol.Finding, error) {
	var (
		results []protocol.Finding
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for _, sf := range serviceFindings {
		wg.Add(1)
		go func(f protocol.Finding) {
			defer wg.Done()

			for _, cve := range m.database {
				if strings.EqualFold(cve.Service, f.Service) ||
					strings.Contains(strings.ToLower(f.Description), strings.ToLower(cve.Pattern)) {
					result := protocol.Finding{
						Target:      f.Target,
						Port:        f.Port,
						Protocol:    f.Protocol,
						Service:     f.Service,
						Title:       fmt.Sprintf("%s: %s", cve.CVE, cve.Description),
						Description: fmt.Sprintf("Potential %s vulnerability on %s", cve.CVE, f.Service),
						Severity:    cve.Severity,
						CVE:         cve.CVE,
						CVSS:        cve.CVSS,
					}
					mu.Lock()
					results = append(results, result)
					mu.Unlock()
				}
			}
		}(sf)
	}

	wg.Wait()
	return results, nil
}
