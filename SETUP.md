# pgAnalytics v3.0 - Quick Setup Guide

## Phase 1: Foundation ✅ COMPLETED

### Project Structure Created
```
pganalytics-v3/
├── backend/
│   ├── cmd/pganalytics-api/main.go          ✅ Entry point
│   ├── internal/                             (Stubs for Phase 2)
│   ├── migrations/
│   │   ├── 001_init.sql                     ✅ Core schema
│   │   └── 002_timescale.sql                ✅ Time-series setup
│   ├── Dockerfile                           ✅ Multi-stage build
│   ├── go.mod                               ✅ Dependencies
│   └── go.sum
│
├── collector/
│   ├── src/
│   │   ├── main.cpp                         ✅ Entry point (stub)
│   │   ├── collector.cpp                    ✅ Core logic (stub)
│   │   └── 7 other .cpp files               ✅ Stubs
│   ├── include/
│   │   ├── collector.h                      ✅ Interfaces
│   │   └── sender.h                         ✅ HTTP/TLS interface
│   ├── CMakeLists.txt                       ✅ Build config
│   ├── vcpkg.json                           ✅ Dependencies
│   ├── config.toml.sample                   ✅ Config template
│   └── Dockerfile                           ✅ Multi-stage build
│
├── grafana/
│   ├── dashboards/                          (Phase 4)
│   └── datasources/                         (Phase 4)
│
├── docs/
│   └── ARCHITECTURE.md                      ✅ Design docs
│
├── docker-compose.yml                       ✅ Demo environment
├── Makefile                                 ✅ Build/test targets
├── README.md                                ✅ Project README
├── .gitignore                               ✅ Git configuration
└── go.mod/go.sum                            ✅ Go dependencies
```

### Database Schema
- **PostgreSQL** (Metadata):
  - users, collectors, servers, postgresql_instances
  - databases, api_tokens, alert_rules, alerts
  - audit_log, secrets

- **TimescaleDB** (Metrics):
  - metrics_pg_stats_table (7d retention)
  - metrics_pg_stats_index (7d retention)
  - metrics_pg_stats_database (7d retention)
  - metrics_sysstat (7d retention)
  - metrics_disk_usage (30d retention)
  - metrics_pg_log (7d retention)
  - metrics_replication (7d retention)

### Services Configured (docker-compose.yml)
- PostgreSQL 16 (port 5432)
- TimescaleDB 16 (port 5433)
- Go Backend (port 8080)
- Grafana (port 3000)
- Redis (port 6379, for future use)
- C++ Collector (integrated with backend)

### Configuration
- Environment variables for all services
- Database connection pooling ready
- Self-signed TLS certificates (for demo)
- JWT secret configuration

## What's Next: Phase 2 (Backend Core)

Tasks:
1. Implement Go API handlers (collectors, metrics, auth)
2. JWT token generation and validation
3. mTLS certificate verification
4. Metrics ingestion endpoint
5. TimescaleDB data insertion
6. Unit tests + integration tests
7. Swagger documentation

## Getting Started

### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+
- Go 1.22+ (for local development)
- C++17 compiler (for collector)

### Quick Check
```bash
# Verify directory structure
ls -la backend/ collector/ grafana/ docs/

# Check database migrations
cat backend/migrations/001_init.sql

# Review architecture
cat docs/ARCHITECTURE.md

# Check docker-compose config
docker-compose config
```

### Next Phase
```bash
# Phase 2 implementation starts with:
make build-backend    # Will fail until Phase 2 implemented
make docker-up        # Start services (may fail on first run due to migrations)
```

---

**Status**: Foundation complete! Backend and collector structure ready for Phase 2 implementation.

**Commits**: Foundation phase is ready for git init and first commit.
