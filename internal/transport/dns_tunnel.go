package transport

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
  "encoding/json"

	"github.com/blackswarm/siege/internal/protocol"
	"github.com/miekg/dns"
)

// DNSTunnel implements a covert communication channel using DNS TXT records.
// It is designed for environments where outbound UDP/53 is permitted.
type DNSTunnel struct {
	domain   string
	client   *dns.Client
	resolver string
}

// NewDNSTunnel creates a new DNS tunnel.
// domain is the authoritative domain under which data is exchanged
// (e.g., "tunnel.example.com").
func NewDNSTunnel(domain, resolver string) *DNSTunnel {
	if resolver == "" {
		resolver = "8.8.8.8:53"
	}
	return &DNSTunnel{
		domain:   domain,
		client:   &dns.Client{Timeout: 5 * time.Second},
		resolver: resolver,
	}
}

// Send encodes an envelope and sends it as a DNS TXT query.
// The data is split into chunks if necessary and sent as multiple queries.
func (dt *DNSTunnel) Send(ctx context.Context, address string, env *protocol.Envelope) error {
	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("dns tunnel: marshal: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(data)
	chunks := chunkString(encoded, 200) // DNS label size limit

	for i, chunk := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		query := fmt.Sprintf("%s.%d.%s", chunk, i, dt.domain)
		if err := dt.sendQuery(query); err != nil {
			return fmt.Errorf("dns tunnel: chunk %d: %w", i, err)
		}
	}
	return nil
}

// Receive listens for incoming DNS queries and decodes them into envelopes.
// This is a server‑side function; the tunnel must be configured on the receiver
// to handle incoming queries to the authoritative domain.
func (dt *DNSTunnel) Receive(ctx context.Context) (*protocol.Envelope, error) {
	// In a real implementation this would be a DNS server listening on port 53.
	// For now, return a not‑implemented sentinel.
	return nil, fmt.Errorf("dns tunnel: receive not implemented (requires DNS server)")
}

// Close releases resources.
func (dt *DNSTunnel) Close() error {
	return nil
}

func (dt *DNSTunnel) sendQuery(query string) error {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(query), dns.TypeTXT)
	msg.RecursionDesired = true

	_, _, err := dt.client.Exchange(msg, dt.resolver)
	return err
}

func chunkString(s string, chunkSize int) []string {
	if len(s) == 0 {
		return []string{}
	}
	var chunks []string
	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}
