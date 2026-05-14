// Package scheduler implements distributed task scheduling for the swarm.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// Scheduler distributes scanning tasks across the swarm.
type Scheduler struct {
	nodeID string
	logger *slog.Logger

	// Peers provider
	peerProvider PeerProvider

	// Pending tasks
	tasksMu sync.RWMutex
	tasks   map[string]*protocol.TaskMsg

	// Results
	resultsMu sync.RWMutex
	results   map[string][]*protocol.ResultMsg

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// PeerProvider is the interface for obtaining the current list of peers.
type PeerProvider interface {
	Peers() []protocol.PeerInfo
}

// NewScheduler creates a new Scheduler.
func NewScheduler(nodeID string, peerProvider PeerProvider, logger *slog.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		nodeID:       nodeID,
		peerProvider: peerProvider,
		logger:       logger,
		tasks:        make(map[string]*protocol.TaskMsg),
		results:      make(map[string][]*protocol.ResultMsg),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins the scheduler's background loops.
func (s *Scheduler) Start(_ context.Context) error {
	s.logger.Info("scheduler started")
	return nil
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() error {
	s.cancel()
	s.wg.Wait()
	s.logger.Info("scheduler stopped")
	return nil
}

// Submit adds a new task and triggers its distribution.
func (s *Scheduler) Submit(task *protocol.TaskMsg) error {
	if task.TaskID == "" {
		return fmt.Errorf("task ID is required")
	}

	s.tasksMu.Lock()
	s.tasks[task.TaskID] = task
	s.tasksMu.Unlock()

	s.logger.Info("task submitted", "task_id", task.TaskID, "targets", len(task.Targets))
	go s.dispatchTask(task)
	return nil
}

// PendingTasks returns all currently pending tasks.
func (s *Scheduler) PendingTasks() []protocol.TaskMsg {
	s.tasksMu.RLock()
	defer s.tasksMu.RUnlock()
	list := make([]protocol.TaskMsg, 0, len(s.tasks))
	for _, t := range s.tasks {
		list = append(list, *t)
	}
	return list
}

// RecordResult stores a received result.
func (s *Scheduler) RecordResult(result *protocol.ResultMsg) {
	s.resultsMu.Lock()
	defer s.resultsMu.Unlock()
	s.results[result.TaskID] = append(s.results[result.TaskID], result)
}

// ResultsForTask returns all results for a given task ID.
func (s *Scheduler) ResultsForTask(taskID string) []*protocol.ResultMsg {
	s.resultsMu.RLock()
	defer s.resultsMu.RUnlock()
	return s.results[taskID]
}

// dispatchTask selects the best peer(s) to execute a task and sends them a ScanRequest.
func (s *Scheduler) dispatchTask(task *protocol.TaskMsg) {
	peers := s.peerProvider.Peers()
	if len(peers) == 0 {
		s.logger.Warn("no peers available for task", "task_id", task.TaskID)
		return
	}

	// Simple round‑robin / load‑aware selection: pick the peer with the lowest load.
	best := peers[0]
	for _, p := range peers[1:] {
		if p.Load < best.Load {
			best = p
		}
	}

	req := &protocol.ScanRequest{Task: *task}
	env, err := protocol.NewEnvelope(protocol.TypeScanRequest, s.nodeID, req)
	if err != nil {
		s.logger.Error("failed to create scan request", "error", err)
		return
	}

	// In a real implementation we would use the peer manager to send directly.
	// Here we rely on the caller injecting a suitable transport.
	s.logger.Info("dispatching task", "task_id", task.TaskID, "peer", best.ID)
	_ = env // placeholder
}
