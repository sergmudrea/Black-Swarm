package utils

import (
	"net"
	"net/url"
	"strings"
)

// Target represents a scan target with optional port and protocol hints.
type Target struct {
	Host     string   `json:"host"`
	Ports    []int    `json:"ports,omitempty"`
	Protocol string   `json:"protocol,omitempty"` // tcp, udp
	Tags     []string `json:"tags,omitempty"`
}

// Validate checks that the target is well‑formed and returns an error if not.
func (t *Target) Validate() error {
	if t.Host == "" {
		return fmt.Errorf("host must not be empty")
	}

	// Strip scheme if present
	host := t.Host
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	// If it is a URL, extract the hostname
	if u, err := url.Parse("https://" + host); err == nil {
		host = u.Hostname()
	}

	// Validate hostname or IP
	if ip := net.ParseIP(host); ip != nil {
		return nil
	}

	// DNS name validation (simplified)
	if len(host) < 1 || len(host) > 253 {
		return fmt.Errorf("host name length invalid")
	}

	for _, part := range strings.Split(host, ".") {
		if len(part) < 1 || len(part) > 63 {
			return fmt.Errorf("host label length invalid")
		}
	}

	return nil
}

// Normalize returns a clean copy of the target.
func (t *Target) Normalize() Target {
	host := t.Host
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	if u, err := url.Parse("https://" + host); err == nil {
		host = u.Hostname()
	}

	if t.Protocol == "" {
		t.Protocol = "tcp"
	}

	return Target{
		Host:     host,
		Ports:    t.Ports,
		Protocol: t.Protocol,
		Tags:     t.Tags,
	}
}

// String returns a human-readable representation of the target.
func (t *Target) String() string {
	if len(t.Ports) == 0 {
		return t.Host
	}
	portStrs := make([]string, len(t.Ports))
	for i, p := range t.Ports {
		portStrs[i] = fmt.Sprintf("%d", p)
	}
	return fmt.Sprintf("%s:%s", t.Host, strings.Join(portStrs, ","))
}
