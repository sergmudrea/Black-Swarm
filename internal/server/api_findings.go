package server

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// handleFindings dispatches requests to /api/findings based on HTTP method.
func (s *Server) handleFindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.listFindings(w, r)
}

func (s *Server) listFindings(w http.ResponseWriter, r *http.Request) {
	if s.FindingStore == nil {
		http.Error(w, "finding store not initialised", http.StatusInternalServerError)
		return
	}

	query := r.URL.Query()
	target := query.Get("target")
	severity := query.Get("severity")
	taskID := query.Get("task_id")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit := 100
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	offset := 0
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	findings, err := s.FindingStore.Query(target, severity, taskID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(findings)
}
