package server

import (
	"net/http"
)

// handleReports dispatches requests to /api/reports based on HTTP method.
func (s *Server) handleReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.generateReport(w, r)
}

func (s *Server) generateReport(w http.ResponseWriter, r *http.Request) {
	if s.Reporter == nil {
		http.Error(w, "reporter not initialised", http.StatusInternalServerError)
		return
	}

	report, err := s.Reporter.Generate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(report))
}
