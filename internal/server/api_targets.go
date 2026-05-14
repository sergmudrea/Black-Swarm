```go
package server

import (
	"encoding/json"
	"net/http"
)

// handleTargets dispatches requests to /api/targets based on HTTP method.
func (s *Server) handleTargets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTargets(w, r)
	case http.MethodPost:
		s.addTarget(w, r)
	case http.MethodDelete:
		s.removeTarget(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listTargets(w http.ResponseWriter, r *http.Request) {
	if s.Targets == nil {
		http.Error(w, "target manager not initialised", http.StatusInternalServerError)
		return
	}
	targets := s.Targets.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(targets)
}

func (s *Server) addTarget(w http.ResponseWriter, r *http.Request) {
	if s.Targets == nil {
		http.Error(w, "target manager not initialised", http.StatusInternalServerError)
		return
	}
	var req struct {
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Target == "" {
		http.Error(w, "target is required", http.StatusBadRequest)
		return
	}
	if err := s.Targets.Add(req.Target); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "added", "target": req.Target})
}

func (s *Server) removeTarget(w http.ResponseWriter, r *http.Request) {
	if s.Targets == nil {
		http.Error(w, "target manager not initialised", http.StatusInternalServerError)
		return
	}
	var req struct {
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Target == "" {
		http.Error(w, "target is required", http.StatusBadRequest)
		return
	}
	if err := s.Targets.Remove(req.Target); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "removed", "target": req.Target})
}
