package vuln

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// NucleiRunner executes Nuclei templates against targets.
type NucleiRunner struct {
	binaryPath  string
	templateDir string
	concurrency int
	timeout     time.Duration
}

// NewNucleiRunner creates a new NucleiRunner.
func NewNucleiRunner(binaryPath, templateDir string, concurrency int, timeout time.Duration) *NucleiRunner {
	if binaryPath == "" {
		binaryPath = "nuclei"
	}
	return &NucleiRunner{
		binaryPath:  binaryPath,
		templateDir: templateDir,
		concurrency: concurrency,
		timeout:     timeout,
	}
}

// Run executes Nuclei against the given targets and returns findings.
func (nr *NucleiRunner) Run(ctx context.Context, targets []string, severityFilter []string) ([]protocol.Finding, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("targets must not be empty")
	}

	// Check if nuclei binary is available
	if _, err := exec.LookPath(nr.binaryPath); err != nil {
		return nil, fmt.Errorf("nuclei binary not found: %s", nr.binaryPath)
	}

	args := []string{
		"-silent",
		"-json",
		"-timeout", fmt.Sprintf("%d", int(nr.timeout.Seconds())),
		"-concurrency", fmt.Sprintf("%d", nr.concurrency),
	}

	if nr.templateDir != "" {
		args = append(args, "-templates", nr.templateDir)
	}

	if len(severityFilter) > 0 {
		args = append(args, "-severity", strings.Join(severityFilter, ","))
	}

	args = append(args, "-target")
	args = append(args, strings.Join(targets, ","))

	cmd := exec.CommandContext(ctx, nr.binaryPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("nuclei scan timed out")
		}
		return nil, fmt.Errorf("nuclei execution failed: %w", err)
	}

	// Parse JSON output (one JSON object per line)
	return parseNucleiOutput(string(output)), nil
}

func parseNucleiOutput(output string) []protocol.Finding {
	var findings []protocol.Finding
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Simplified parsing: extract fields from the JSON line
		// In production, use a proper JSON unmarshaller
		finding := protocol.Finding{
			Protocol:  "http",
			Title:     extractField(line, "name"),
			Severity:  extractField(line, "severity"),
			Timestamp: time.Now().UnixMilli(),
		}

		if finding.Title == "" {
			continue
		}

		findings = append(findings, finding)
	}

	return findings
}

func extractField(line, field string) string {
	key := fmt.Sprintf(`"%s":"`, field)
	start := strings.Index(line, key)
	if start == -1 {
		return ""
	}
	start += len(key)
	end := strings.Index(line[start:], `"`)
	if end == -1 {
		return ""
	}
	return line[start : start+end]
}
