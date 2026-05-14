package state

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/blackswarm/siege/internal/protocol"
	_ "modernc.org/sqlite"
)

// FindingStore provides persistent storage for scan findings backed by SQLite.
type FindingStore struct {
	mu   sync.RWMutex
	db   *sql.DB
	path string
}

// NewFindingStore opens (or creates) the SQLite database at the given path
// and ensures the schema exists.
func NewFindingStore(path string) (*FindingStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("findings store: open: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("findings store: ping: %w", err)
	}

	store := &FindingStore{db: db, path: path}
	if err := store.migrate(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *FindingStore) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS findings (
		id          TEXT PRIMARY KEY,
		task_id     TEXT NOT NULL,
		target      TEXT NOT NULL,
		port        INTEGER DEFAULT 0,
		protocol    TEXT DEFAULT '',
		service     TEXT DEFAULT '',
		title       TEXT NOT NULL,
		description TEXT DEFAULT '',
		severity    TEXT DEFAULT 'info',
		cve         TEXT DEFAULT '',
		cvss        REAL DEFAULT 0.0,
		evidence    TEXT DEFAULT '',
		remediation TEXT DEFAULT '',
		timestamp   INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_findings_task ON findings(task_id);
	CREATE INDEX IF NOT EXISTS idx_findings_target ON findings(target);
	CREATE INDEX IF NOT EXISTS idx_findings_severity ON findings(severity);
	`
	_, err := s.db.Exec(query)
	return err
}

// Insert stores a single finding. If a finding with the same ID already exists
// it is silently ignored.
func (s *FindingStore) Insert(f *protocol.Finding) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT OR IGNORE INTO findings
			(id, task_id, target, port, protocol, service, title, description,
			 severity, cve, cvss, evidence, remediation, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		f.ID, f.Target, f.Target, f.Port, f.Protocol, f.Service,
		f.Title, f.Description, f.Severity, f.CVE, f.CVSS,
		f.Evidence, f.Remediation, f.Timestamp,
	)
	return err
}

// InsertBatch stores multiple findings efficiently.
func (s *FindingStore) InsertBatch(findings []*protocol.Finding) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO findings
			(id, task_id, target, port, protocol, service, title, description,
			 severity, cve, cvss, evidence, remediation, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		_, err = stmt.Exec(
			f.ID, f.Target, f.Target, f.Port, f.Protocol, f.Service,
			f.Title, f.Description, f.Severity, f.CVE, f.CVSS,
			f.Evidence, f.Remediation, f.Timestamp,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Query returns findings filtered by optional criteria.
// Zero values for target/severity/taskID mean "any".
func (s *FindingStore) Query(target, severity, taskID string, limit, offset int) ([]*protocol.Finding, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := "SELECT * FROM findings WHERE 1=1"
	args := make([]interface{}, 0)

	if target != "" {
		query += " AND target = ?"
		args = append(args, target)
	}
	if severity != "" {
		query += " AND severity = ?"
		args = append(args, severity)
	}
	if taskID != "" {
		query += " AND task_id = ?"
		args = append(args, taskID)
	}

	query += " ORDER BY timestamp DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFindings(rows)
}

// Stats returns basic statistics about stored findings.
func (s *FindingStore) Stats() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]interface{})

	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM findings").Scan(&total); err != nil {
		return nil, err
	}
	stats["total"] = total

	rows, err := s.db.Query("SELECT severity, COUNT(*) FROM findings GROUP BY severity")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bySeverity := make(map[string]int)
	for rows.Next() {
		var sev string
		var cnt int
		if err := rows.Scan(&sev, &cnt); err != nil {
			return nil, err
		}
		bySeverity[sev] = cnt
	}
	stats["by_severity"] = bySeverity

	return stats, nil
}

// Close closes the database.
func (s *FindingStore) Close() error {
	return s.db.Close()
}

// scanFindings converts SQL rows into Finding pointers.
func scanFindings(rows *sql.Rows) ([]*protocol.Finding, error) {
	var findings []*protocol.Finding
	for rows.Next() {
		var f protocol.Finding
		err := rows.Scan(
			&f.ID, &f.Target, &f.Target, &f.Port, &f.Protocol, &f.Service,
			&f.Title, &f.Description, &f.Severity, &f.CVE, &f.CVSS,
			&f.Evidence, &f.Remediation, &f.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		findings = append(findings, &f)
	}
	return findings, rows.Err()
}
