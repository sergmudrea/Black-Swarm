// Package server implements the HTTP/WebSocket server for the Siege dashboard and API.
package server

import (
	"context"
	"crypto/tls"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

  "github.com/sergmudrea/black-swarm/siege/internal/protocol"
)

//go:embed web/build
var webAssets embed.FS

// Server provides the operator dashboard and REST API for the Siege swarm.
type Server struct {
	addr    string
	logger  *slog.Logger
	httpSrv *http.Server

	// Services (injected before Start)
	FindingStore FindingStore
	Scheduler    TaskScheduler
	Targets      TargetManager
	Reporter     Reporter
}

// FindingStore is the interface for querying scan findings.
type FindingStore interface {
	Query(target, severity, taskID string, limit, offset int) ([]*protocol.Finding, error)
	Stats() (map[string]interface{}, error)
}

// TaskScheduler is the interface for submitting and tracking tasks.
type TaskScheduler interface {
	Submit(task *protocol.TaskMsg) error
	PendingTasks() []protocol.TaskMsg
}

// TargetManager is the interface for managing scan targets.
type TargetManager interface {
	Add(target string) error
	List() []string
	Remove(target string) error
}

// Reporter generates operational reports.
type Reporter interface {
	Generate() (string, error)
}

// NewServer creates a new Server instance.
func NewServer(addr string, logger *slog.Logger) *Server {
	return &Server{
		addr:   addr,
		logger: logger,
	}
}

// Start begins listening for HTTP/HTTPS connections.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// REST API endpoints
	mux.HandleFunc("/api/targets", s.handleTargets)
	mux.HandleFunc("/api/scans", s.handleScans)
	mux.HandleFunc("/api/findings", s.handleFindings)
	mux.HandleFunc("/api/reports", s.handleReports)
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Serve embedded React dashboard
	staticFS, err := fs.Sub(webAssets, "web/build")
	if err != nil {
		s.logger.Warn("embedded web assets not found, serving fallback page", "error", err)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Siege Server Active"))
		})
	} else {
		fileServer := http.FileServer(http.FS(staticFS))
		mux.Handle("/", fileServer)
	}

	s.httpSrv = &http.Server{
		Addr:         s.addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Try TLS, fallback to plain HTTP
	certFile := "server.crt"
	keyFile := "server.key"
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return err
			}
			s.httpSrv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS13,
			}
			s.logger.Info("starting TLS server", "addr", s.addr)
			return s.httpSrv.ListenAndServeTLS("", "")
		}
	}
	s.logger.Warn("no TLS credentials found, starting plain HTTP server", "addr", s.addr)
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

// handleWebSocket upgrades an HTTP connection to WebSocket (placeholder for future real-time updates).
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Not implemented in this version
	http.Error(w, "WebSocket not yet implemented", http.StatusNotImplemented)
}
