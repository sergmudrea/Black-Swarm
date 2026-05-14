package scheduler

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"github.com/blackswarm/siege/internal/protocol"
)

// Deduplicator removes duplicate findings from a stream.
type Deduplicator struct {
	mu   sync.Mutex
	seen map[string]bool
}

// NewDeduplicator creates a new Deduplicator.
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		seen: make(map[string]bool),
	}
}

// Deduplicate filters out findings that have already been seen.
// A finding is considered duplicate if it has the same target, port, protocol, and title.
func (d *Deduplicator) Deduplicate(findings []protocol.Finding) []protocol.Finding {
	d.mu.Lock()
	defer d.mu.Unlock()

	unique := make([]protocol.Finding, 0, len(findings))
	for _, f := range findings {
		key := d.makeKey(f)
		if d.seen[key] {
			continue
		}
		d.seen[key] = true
		unique = append(unique, f)
	}
	return unique
}

// Reset clears the deduplication state.
func (d *Deduplicator) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[string]bool)
}

func (d *Deduplicator) makeKey(f protocol.Finding) string {
	data := fmt.Sprintf("%s|%d|%s|%s", f.Target, f.Port, f.Protocol, f.Title)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
