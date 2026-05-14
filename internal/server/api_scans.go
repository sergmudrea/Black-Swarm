```go
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/blackswarm/siege/internal/protocol"
)

// handleScans dispatches requests to /api/scans based on HTTP method.
func (s *Server) handleScans(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listScans(w, r)
	case http.MethodPost:
		s.createScan(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listScans(w http.ResponseWriter, r *http.Request) {
	if s.Scheduler == nil {
		http.Error(w, "scheduler not initialised", http.StatusInternalServerError)
		return
	}
	tasks := s.Scheduler.PendingTasks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) createScan(w http.ResponseWriter, r *http.Request) {
	if s.Scheduler == nil {
		http.Error(w, "scheduler not initialised", http.StatusInternalServerError)
		return
	}
	var req struct {
		Targets []string `json:"targets"`
		Ports   []int    `json:"ports,omitempty"`
		Modules []string `json:"modules,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if len(req.Targets) == 0 {
		http.Error(w, "at least one target is required", http.StatusBadRequest)
		return
	}

	task := &protocol.TaskMsg{
		TaskID:   generateTaskID(),
		Targets:  req.Targets,
		Ports:    req.Ports,
		Modules:  req.Modules,
		Priority: 5,
	}
	if err := s.Scheduler.Submit(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}
