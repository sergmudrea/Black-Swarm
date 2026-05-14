// Package state provides a thread‑safe shared state for the swarm.
package state

import (
	"sync"

	"github.com/blackswarm/siege/internal/protocol"
)

// State holds the current shared knowledge of the swarm.
type State struct {
	mu       sync.RWMutex
	tasks    map[string]*protocol.TaskMsg
	findings map[string][]*protocol.Finding // keyed by task ID
	peers    []protocol.PeerInfo
	seqNum   uint64
}

// NewState creates a new empty State.
func NewState() *State {
	return &State{
		tasks:    make(map[string]*protocol.TaskMsg),
		findings: make(map[string][]*protocol.Finding),
		peers:    make([]protocol.PeerInfo, 0),
	}
}

// AddTask adds a task to the state.
func (s *State) AddTask(task *protocol.TaskMsg) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.TaskID] = task
}

// GetTask returns a task by its ID.
func (s *State) GetTask(taskID string) *protocol.TaskMsg {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[taskID]
}

// AddFindings appends findings for a given task.
func (s *State) AddFindings(taskID string, findings []*protocol.Finding) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.findings[taskID] = append(s.findings[taskID], findings...)
}

// GetFindings returns all findings for a given task.
func (s *State) GetFindings(taskID string) []*protocol.Finding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.findings[taskID]
}

// AllFindings returns all findings across all tasks.
func (s *State) AllFindings() []*protocol.Finding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var all []*protocol.Finding
	for _, f := range s.findings {
		all = append(all, f...)
	}
	return all
}

// UpdatePeers replaces the current peer list.
func (s *State) UpdatePeers(peers []protocol.PeerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers = peers
}

// Peers returns a copy of the current peer list.
func (s *State) Peers() []protocol.PeerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := make([]protocol.PeerInfo, len(s.peers))
	copy(cp, s.peers)
	return cp
}

// IncSeqNum atomically increments and returns the new sequence number.
func (s *State) IncSeqNum() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seqNum++
	return s.seqNum
}

// SeqNum returns the current sequence number.
func (s *State) SeqNum() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.seqNum
}
