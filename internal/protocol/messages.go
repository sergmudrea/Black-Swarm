// Package protocol defines all message types exchanged between nodes in the Black Swarm.
package protocol

import (
	"encoding/json"
	"time"
)

// Message types used in the "Type" field of an envelope.
const (
	TypeTask         = "task"
	TypeResult       = "result"
	TypeGossip       = "gossip"
	TypePeerHello    = "peer_hello"
	TypePeerBye      = "peer_bye"
	TypeStateSync    = "state_sync"
	TypeFindingsSync = "findings_sync"
	TypeScanRequest  = "scan_request"
	TypeScanResponse = "scan_response"
	TypeError        = "error"
)

// Envelope is the generic wrapper for all messages.
type Envelope struct {
	Type      string          `json:"type"`
	SenderID  string          `json:"sender_id"`
	Timestamp int64           `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// NewEnvelope creates a signed envelope for the given message type and payload.
func NewEnvelope(msgType string, senderID string, payload interface{}) (*Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		Type:      msgType,
		SenderID:  senderID,
		Timestamp: time.Now().UnixMilli(),
		Payload:   data,
	}, nil
}

// TaskMsg describes a scanning task dispatched to one or more scanner nodes.
type TaskMsg struct {
	TaskID    string   `json:"task_id"`
	Targets   []string `json:"targets"`
	Ports     []int    `json:"ports,omitempty"`
	Modules   []string `json:"modules"`
	Priority  int      `json:"priority"`
	Deadline  int64    `json:"deadline,omitempty"`
	Strategy  string   `json:"strategy,omitempty"` // serialised chromosome
}

// ResultMsg contains the outcome of a completed task.
type ResultMsg struct {
	TaskID    string    `json:"task_id"`
	NodeID    string    `json:"node_id"`
	Status    string    `json:"status"` // "ok", "partial", "error"
	Findings  []Finding `json:"findings,omitempty"`
	Error     string    `json:"error,omitempty"`
	Duration  float64   `json:"duration_sec"`
	Completed int64     `json:"completed"`
}

// Finding is a single vulnerability or information item discovered during a scan.
type Finding struct {
	ID          string `json:"id"`
	Target      string `json:"target"`
	Port        int    `json:"port,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Service     string `json:"service,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "critical", "high", "medium", "low", "info"
	CVE         string `json:"cve,omitempty"`
	CVSS        float64 `json:"cvss,omitempty"`
	Evidence    string `json:"evidence,omitempty"`
	Remediation string `json:"remediation,omitempty"`
	Timestamp   int64  `json:"timestamp"`
}

// GossipMessage carries peer information and optional state updates.
type GossipMessage struct {
	Peers     []PeerInfo `json:"peers"`
	StateHash string     `json:"state_hash,omitempty"`
	Round     uint64     `json:"round"`
}

// PeerInfo describes a known node in the swarm.
type PeerInfo struct {
	ID        string   `json:"id"`
	Address   string   `json:"address"`
	Mode      string   `json:"mode"`
	Capabilities []string `json:"capabilities,omitempty"`
	LastSeen  int64    `json:"last_seen"`
	Load      float64  `json:"load"` // 0.0 – 1.0
}

// StateSync carries the full or delta state of the swarm.
type StateSync struct {
	Tasks    []TaskMsg    `json:"tasks,omitempty"`
	Findings []Finding    `json:"findings,omitempty"`
	Peers    []PeerInfo   `json:"peers,omitempty"`
	SeqNum   uint64       `json:"seq_num"`
}

// ScanRequest is sent by the scheduler to assign a task to a specific node.
type ScanRequest struct {
	Task TaskMsg `json:"task"`
}

// ScanResponse acknowledges a scan request.
type ScanResponse struct {
	TaskID  string `json:"task_id"`
	Accepted bool   `json:"accepted"`
	Reason  string `json:"reason,omitempty"`
}
