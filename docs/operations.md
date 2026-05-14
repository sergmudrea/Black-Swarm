# Siege — Operations Manual

**Version:** 3.0.0
**Audience:** Red Team Operators
**Last Updated:** 2026-05-14

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Building](#building)
3. [Configuration](#configuration)
4. [Running Locally with Docker Compose](#running-locally-with-docker-compose)
5. [Deploying to Production](#deploying-to-production)
6. [Using the Dashboard](#using-the-dashboard)
7. [Interpreting Results](#interpreting-results)
8. [Genetic Algorithm Tuning](#genetic-algorithm-tuning)
9. [Troubleshooting](#troubleshooting)

## 1. Prerequisites
- Go 1.22+
- Node.js 18+ with npm
- Docker & Docker Compose (for local testing)
- Ansible 2.14+ (for production deployment)

## 2. Building
```bash
git clone https://github.com/sergmudrea/Black-Swarm.git
cd Black-Swarm
make build

This compiles the web dashboard and then the Go binary into bin/swarm.
3. Configuration

    Generate a 32‑byte encryption key:
    bash

export SIEGE_CONFIG_KEY=$(openssl rand -base64 32)

Copy configs/node_config.json and edit the placeholders:

    node_id: unique name (e.g., strategic-1).

    mode: strategic, scanner, or hybrid.

    seed_peers: list of initial peer addresses (host:7946).

Encrypt the configuration file:
bash

./bin/swarm -encrypt-in node_config.json -encrypt-out node_config.enc

4. Running Locally with Docker Compose

The provided docker-compose.yml starts one strategic and two scanner nodes.
bash

docker-compose up -d

Access the dashboard at http://localhost:8443.
5. Deploying to Production

Use the Ansible playbooks:

    scripts/deploy_strategic.yml — for strategic nodes.

    scripts/deploy_scanner.yml — for scanner nodes.
    Set the SIEGE_CONFIG_KEY environment variable on the control machine.

6. Using the Dashboard

    Targets: Add IPs or domains in the Scan Configuration panel.

    Modules: Select the scanning modules to run.

    Ports: Specify a comma‑separated list.

    Start Scan: Click the button to dispatch the task.

    Findings: View results in the table, filter by severity or search text.

    Swarm Map: 3D view of connected peers (color‑coded by mode).

7. Interpreting Results

Findings are severity‑rated:

    Critical: Immediate risk, e.g., RCE.

    High: Significant weakness, e.g., SQLi.

    Medium: Potential risk, e.g., exposed configuration.

    Low: Minor issue, e.g., information disclosure.

    Info: Informational, e.g., open port.

8. Genetic Algorithm Tuning

The GA evolves scan strategies across generations. Adjust parameters in the configuration:

    population_size: Number of chromosomes (default 100).

    generations: Number of generations per evolution (default 50).

    mutation_rate: Probability of random gene change (default 0.05).

    crossover_rate: Probability of combining two parents (default 0.7).

    elite_count: Number of best chromosomes preserved (default 5).
    Higher generations and population sizes improve optimization but take longer.

9. Troubleshooting

    Node not connecting: Check seed_peers and network connectivity.

    Dashboard not loading: Ensure the web/build directory exists (rebuild with make build-web).

    Scan timeout: Increase scan_timeout_seconds in config.

    Rate limiting: Adjust rate_limit to avoid triggering IDS/IPS.
