# Siege — Architecture Document

**Version:** 3.0.0
**Status:** Approved
**Audience:** Developers, Red Team Operators
**Last Updated:** 2026-05-14

## Table of Contents

1. [System Overview](#system-overview)
2. [Design Principles](#design-principles)
3. [Component Architecture](#component-architecture)
   - 3.1 [Node](#node)
   - 3.2 [Gossip Protocol](#gossip-protocol)
   - 3.3 [Scheduler](#scheduler)
   - 3.4 [Scanner Modules](#scanner-modules)
   - 3.5 [OSINT Modules](#osint-modules)
   - 3.6 [Genetic Algorithm](#genetic-algorithm)
   - 3.7 [Transport Layer](#transport-layer)
   - 3.8 [Server & Dashboard](#server--dashboard)
4. [Communication Protocol](#communication-protocol)
5. [Deployment Topology](#deployment-topology)
6. [Security Model](#security-model)
7. [Future Evolution](#future-evolution)
8. [Glossary](#glossary)

## 1. System Overview
Siege is a distributed reconnaissance system that coordinates a swarm of scanning nodes using a genetic algorithm to optimize scan strategies. Nodes communicate via a UDP gossip protocol, discover each other automatically, and can take on three roles:
- **Strategic:** Coordinates the swarm, runs the genetic algorithm, serves the dashboard.
- **Scanner:** Executes scanning tasks assigned by strategic nodes.
- **Hybrid:** Acts as both strategic and scanner.

## 2. Design Principles
- **Decentralization:** No single point of failure; any strategic node can coordinate.
- **Self-organization:** Nodes discover each other via seed nodes and multicast.
- **Evolutionary optimization:** Scan parameters (ports, modules, timing, evasion) are treated as chromosomes and optimized through a genetic algorithm using scan results as fitness feedback.
- **Modularity:** Scanners, OSINT modules, and evasion techniques are pluggable.
- **Operator visibility:** A real‑time 3D dashboard displays swarm status and findings.

## 3. Component Architecture

### 3.1 Node
The `Node` struct (`internal/node/node.go`) manages the lifecycle. It starts subsystems according to its mode: gossip, scheduler (if strategic), and outbox processing (if scanner).

### 3.2 Gossip Protocol
UDP-based gossip (`internal/coordination/gossip/protocol.go`) handles peer discovery, membership, and message broadcasting. Peers are periodically probed; dead peers are removed after a timeout. Discovery uses seed nodes and optional multicast.

### 3.3 Scheduler
The `Scheduler` (`internal/coordination/scheduler/scheduler.go`) distributes tasks to scanner nodes based on load. A deduplicator (`dedup.go`) prevents duplicate findings.

### 3.4 Scanner Modules
- **Port scanning:** TCP SYN (`tcp_syn.go`), UDP (`udp.go`), service detection (`service_detect.go`).
- **Web scanning:** Directory busting (`dirbuster.go`), parameter fuzzing (`fuzzer.go`).
- **Vulnerability detection:** CVE matching (`cve_matcher.go`), Nuclei runner (`nuclei_runner.go`).
- **Reconnaissance:** DNS (`dns.go`), subdomain enumeration (`subdomain.go`).

### 3.5 OSINT Modules
- **GitHub dorking:** Searches for exposed secrets (`intel/github.go`).
- **Shodan/Censys:** Queries internet-wide scan data (`intel/shodan.go`).
- **Certificate Transparency:** Discovers subdomains via crt.sh (`intel/cert.go`).

### 3.6 Genetic Algorithm
Located in `internal/genetics/`. A population of `ScanStrategy` chromosomes evolves over generations. Fitness is a composite of coverage, severity, efficiency, diversity, and stealth. Operators: tournament selection, uniform crossover, random mutation, elitism.

### 3.7 Transport Layer
`internal/transport/` defines a `Transport` interface with WebSocket and DNS tunnel implementations.

### 3.8 Server & Dashboard
The HTTP/WebSocket server (`internal/server/server.go`) exposes a REST API for targets, scans, findings, and reports. A React dashboard (`web/`) visualizes the swarm as a 3D map and displays findings in a sortable table.

## 4. Communication Protocol
Messages are JSON envelopes (`protocol/messages.go`) with types: `task`, `result`, `gossip`, `peer_hello`, `state_sync`, etc. Gossip messages carry peer lists. Scan requests and responses are used internally by the scheduler.

## 5. Deployment Topology
Typically one or more strategic nodes and many scanner nodes. All nodes connect via UDP gossip. The strategic node serves the dashboard on port 8443. Docker Compose provides a local test setup. Ansible playbooks deploy to VPS.

## 6. Security Model
- **TLS 1.3** for WebSocket and HTTP traffic.
- **AES‑256‑GCM** encrypts configuration files.
- Gossip messages are plaintext but can be encrypted at the application layer in future.

## 7. Future Evolution
- Full end‑to‑end encryption of gossip.
- Integration with external C2 frameworks.
- Machine learning for adaptive scanning.

## 8. Glossary
| Term | Definition |
|------|------------|
| **Node** | A single Siege instance. |
| **Strategic node** | Node that coordinates the swarm. |
| **Scanner node** | Node that executes scans. |
| **Gossip** | Decentralized communication protocol. |
| **Chromosome** | Encoding of a scan strategy. |
| **Fitness** | Score measuring strategy effectiveness. |
