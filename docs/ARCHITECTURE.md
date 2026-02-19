# pgAnalytics v3.0 Architecture

## Overview

pgAnalytics v3.0 is a modern, scalable PostgreSQL monitoring platform built with cutting-edge technologies. The architecture follows a distributed collector + centralized backend pattern with strong emphasis on security, performance, and reliability.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│         CENTRAL INFRASTRUCTURE (Single Instance)    │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────────┐  ┌──────────────┐               │
│  │ PostgreSQL   │  │ TimescaleDB  │               │
│  │ (Metadata)   │  │ (Metrics)    │               │
│  └──────┬───────┘  └──────┬───────┘               │
│         │                 │                       │
│         └────────────┬────┘                       │
│                      │                            │
│         ┌────────────▼──────────────┐             │
│         │                           │             │
│         │   Go Backend API          │             │
│         │  (Port 8080, TLS 1.3)     │             │
│         │                           │             │
│         └──────────────┬────────────┘             │
│                        │                          │
│         ┌──────────────▼──────────────┐           │
│         │                            │           │
│         │   Grafana Dashboards       │           │
│         │  (Port 3000)               │           │
│         │                            │           │
│         └────────────────────────────┘           │
│                                                   │
└─────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────┐
│     DISTRIBUTED COLLECTORS (Multiple Locations)     │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Server 1          Server 2          Server N       │
│  ┌─────────┐      ┌─────────┐      ┌─────────┐    │
│  │PgSQL    │      │PgSQL    │      │PgSQL    │    │
│  │5432     │      │5432     │      │5432     │    │
│  └────┬────┘      └────┬────┘      └────┬────┘    │
│       │                │                │         │
│  ┌────▼─────────┐ ┌────▼─────────┐ ┌────▼──────┐ │
│  │C++ Collector │ │C++ Collector │ │C++ Collect│ │
│  │ • TLS 1.3    │ │ • TLS 1.3    │ │ • TLS 1.3 │ │
│  │ • mTLS       │ │ • mTLS       │ │ • mTLS    │ │
│  │ • JWT Auth   │ │ • JWT Auth   │ │ • JWT Auth│ │
│  │ • Buffering  │ │ • Buffering  │ │ • Buffer. │ │
│  └────┬─────────┘ └────┬─────────┘ └────┬──────┘ │
│       │                │                │         │
│       └────────────────┼────────────────┘         │
│                        │                          │
│    HTTPS (TLS 1.3 + mTLS + JWT + gzip)           │
│    POST /api/v1/metrics/push every 60s           │
│    GET /api/v1/config/{id} every 5min            │
│                        │                          │
│                        │                          │
└────────────────────────┼──────────────────────────┘
                         │
                ┌────────▼────────┐
                │  Backend API    │
                │  Load Balancer  │
                └─────────────────┘
```

## Component Description

### 1. Backend API (Go)

**Location**: `backend/`

**Purpose**: Central API server that:
- Receives metrics from collectors
- Manages collector lifecycle (registration, auth, config)
- Stores metrics in TimescaleDB
- Provides query APIs for dashboards
- Handles user authentication
- Exports Prometheus metrics

**Key Packages**:
- `internal/api` - HTTP handlers and routes
- `internal/auth` - JWT tokens and mTLS verification
- `internal/collector` - Collector registration and management
- `internal/metrics` - Metrics ingestion and processing
- `internal/storage` - Database layer (PostgreSQL + TimescaleDB)
- `internal/timescale` - TimescaleDB-specific setup

**Technology**:
- Go 1.22+
- Gin web framework (high performance)
- SQLC for type-safe database queries
- Zap for structured logging
- Prometheus client for metrics

**API Endpoints**:
```
POST   /api/v1/collectors/register       # Register collector + get cert
GET    /api/v1/collectors                # List collectors
GET    /api/v1/collectors/{id}           # Get collector details
POST   /api/v1/metrics/push              # Ingest metrics (secured)
GET    /api/v1/config/{collector_id}    # Pull configuration
GET    /api/v1/servers/{id}/metrics     # Query metrics
GET    /api/v1/health                    # Health check
```

### 2. PostgreSQL (Metadata Database)

**Location**: Docker service `postgres`

**Purpose**:
- Store user accounts and permissions
- Collector registry (metadata, certificates, configuration)
- Alert rules and definitions
- Audit logs

**Schema**: `pganalytics`

**Key Tables**:
- `users` - User accounts and roles
- `collectors` - Collector metadata and status
- `servers` - Monitored database servers
- `postgresql_instances` - PostgreSQL instance info
- `databases` - Database list per instance
- `alert_rules` - Alert configuration
- `alerts` - Active/resolved alerts
- `api_tokens` - Authentication tokens

### 3. TimescaleDB (Time-Series Database)

**Location**: Docker service `timescale`

**Purpose**:
- Store time-series metrics with high compression
- Fast queries for historical data
- Automatic data retention/cleanup

**Hypertables**:
- `metrics_pg_stats_table` - Table-level statistics (7-day retention)
- `metrics_pg_stats_index` - Index statistics (7-day retention)
- `metrics_pg_stats_database` - Database-level metrics (7-day retention)
- `metrics_sysstat` - System CPU, memory, I/O (7-day retention)
- `metrics_disk_usage` - Disk usage (30-day retention)
- `metrics_pg_log` - PostgreSQL logs (7-day retention)
- `metrics_replication` - Replication lag (7-day retention)

**Continuous Aggregates**:
- `metrics_pg_stats_table_1h` - Hourly rollups (for dashboards)

### 4. Distributed Collectors (C/C++)

**Location**: `collector/`

**Purpose**:
- Run on each database server
- Gather PostgreSQL metrics
- Collect system statistics
- Process logs
- Send metrics to backend securely

**Architecture**:
```
┌────────────────────────────────────┐
│      Collector Manager             │
│  (Orchestrates all collectors)     │
├────────────────────────────────────┤
│ ┌──────────────────────────────┐   │
│ │  PgStatsCollector            │   │
│ │  • Table/index/DB stats      │   │
│ │  • Query PostgreSQL views    │   │
│ └──────────────────────────────┘   │
│ ┌──────────────────────────────┐   │
│ │  SysstatCollector            │   │
│ │  • CPU, memory, I/O          │   │
│ │  • Load average              │   │
│ └──────────────────────────────┘   │
│ ┌──────────────────────────────┐   │
│ │  DiskUsageCollector          │   │
│ │  • Filesystem usage          │   │
│ │  • Inode count               │   │
│ └──────────────────────────────┘   │
│ ┌──────────────────────────────┐   │
│ │  PgLogCollector              │   │
│ │  • Parse pg_log files        │   │
│ │  • Extract checkpoints, etc  │   │
│ └──────────────────────────────┘   │
├────────────────────────────────────┤
│  MetricsBuffer                     │
│  • Circular buffer                 │
│  • Gzip compression                │
│  • Flush every 60s                 │
├────────────────────────────────────┤
│  ConfigManager                     │
│  • Pull config from backend        │
│  • Every 5 minutes                 │
│  • Dynamic enable/disable          │
├────────────────────────────────────┤
│  Sender                            │
│  • libcurl + TLS 1.3               │
│  • mTLS authentication             │
│  • JWT tokens                      │
│  • Retry with exponential backoff  │
└────────────────────────────────────┘
```

**Key Features**:
- **High Performance**: Reutilizes 70% of v2 C++ codebase
- **Security**: TLS 1.3 + mTLS + JWT authentication
- **Resilience**: Local buffering, retry logic, graceful degradation
- **Modern C++**: C++17, nlohmann/json, spdlog
- **Configuration**: TOML-based, pulled from backend

### 5. Grafana (Visualization)

**Location**: Docker service `grafana`

**Purpose**:
- Visualize metrics via dashboards
- Create alerts based on metric thresholds
- Provide user-friendly monitoring interface

**Pre-built Dashboards**:
- PostgreSQL Overview (key metrics at a glance)
- Performance Analysis (slow queries, cache hit ratios)
- System Health (CPU, memory, I/O)
- Replication Monitoring (lag, status)
- Disk Usage Trends

**Data Sources**:
- PostgreSQL (for metadata queries)
- TimescaleDB (for time-series metrics)

## Security Model

### Authentication & Authorization

**Collector ↔ Backend**:
1. **mTLS (Mutual TLS 1.3)**
   - Client cert: Issued at registration
   - Server cert: Self-signed for demo, CA-signed for production
   - Enforces TLS 1.3 only

2. **JWT Tokens**
   - Issued at registration
   - Included in Authorization header
   - Expires every 15 minutes (configurable)
   - Refresh tokens for long-running collectors

3. **Flow**:
   ```
   Collector                          Backend
   ─────────────────────────────────────────────
   POST /api/v1/collectors/register
        ├─ TLS handshake (mTLS)      ✓ Verify cert
        └─ Send collector_id         ├─ Create collector record
                                     ├─ Issue JWT token
                                     └─ Return cert + token

   Then, every 60s:
   POST /api/v1/metrics/push
        ├─ TLS handshake (mTLS)      ✓ Verify cert
        ├─ Authorization: Bearer {JWT}
        └─ gzip compressed JSON      ├─ Verify JWT (signature, exp)
                                     ├─ Decompress
                                     └─ Insert to TimescaleDB
   ```

**User ↔ Backend**:
1. **JWT Tokens** (HTTP/HTTPS)
2. **Refresh tokens** (long-lived)
3. **RBAC** (Role-Based Access Control)
   - Admin: Full access
   - User: Read/write to assigned resources
   - Viewer: Read-only access

### Encryption

- **At Rest**: PostgreSQL native encryption (pgcrypto extension)
- **In Transit**: TLS 1.3, enforced minimum
- **Secrets**: Stored encrypted in database, environment variables for deployment

## Data Flow

### Metrics Collection & Ingestion

```
1. Collector wakes up (every 60s)
   ↓
2. PgStatsCollector queries PostgreSQL views
   ↓
3. SysstatCollector reads /proc, run df command
   ↓
4. DiskUsageCollector parses filesystem usage
   ↓
5. PgLogCollector incremental log parsing
   ↓
6. MetricsBuffer combines all metrics
   ↓
7. Compress with gzip
   ↓
8. Sender performs TLS handshake (mTLS)
   ↓
9. POST to /api/v1/metrics/push with JWT
   ↓
10. Backend verifies TLS cert + JWT
    ↓
11. Decompress JSON
    ↓
12. Validate schema
    ↓
13. Insert to TimescaleDB (hypertables)
    ↓
14. Return 200 OK with config version
```

## Scalability Considerations

### Vertical Scaling
- Backend can run on a single large instance
- TimescaleDB compression keeps database size manageable
- PostgreSQL for metadata (small dataset)

### Horizontal Scaling (Future)
- Backend behind load balancer
- TimescaleDB with PostgreSQL streaming replication
- Collectors distributed across organizations

### Performance Targets
- Backend: Handles 100+ concurrent collectors
- Metrics: 1000+ metrics per push, 60s interval
- Latency: p95 < 500ms for metrics ingestion
- Storage: ~7 days of metrics

## Deployment Models

### Development/Demo
- Docker Compose (single machine)
- All services in one container network
- Self-signed TLS certificates

### Production - Standalone
- Same containers
- TLS certificates from Let's Encrypt
- PostgreSQL with backups
- Monitoring of monitoring

### Production - High Availability (Future)
- Backend cluster behind load balancer
- TimescaleDB streaming replication
- Redis for caching/session store
- Prometheus + Alertmanager for backend monitoring

## Monitoring the Monitoring System

Backend and collectors export Prometheus metrics:

**Backend**:
- HTTP request latency
- Database connection pool
- Collector registration count
- Metrics ingestion rate
- API error rates

**Collector**:
- Collection duration per metric type
- Network send success rate
- Local buffer utilization
- Configuration updates count

These metrics are scraped and visualized in a self-monitoring dashboard.

## Technology Stack Summary

| Component | Technology | Version | Why |
|-----------|-----------|---------|-----|
| Backend | Go | 1.22+ | Performance, concurrency, single binary |
| Framework | Gin | Latest | Fast HTTP framework |
| Database | PostgreSQL | 14+ | ACID, mature, familiar |
| TimeSeries | TimescaleDB | Latest | PostgreSQL extension, SQL standard |
| Collector | C/C++ | 17 | Performance, existing codebase reuse |
| JSON | nlohmann | 3.11+ | Modern, header-only |
| HTTP | libcurl | 7.85+ | TLS 1.3, mTLS support |
| TLS | OpenSSL | 3.0+ | TLS 1.3, cryptography |
| Compression | zlib | 1.2+ | Standard, performant |
| Logging | spdlog | 1.12+ | High performance |
| Testing | Google Test | Latest | Standard C++ testing |
| Dashboard | Grafana | 11+ | Industry standard |
| Container | Docker | 20.10+ | Standard |
| Orchestration | Docker Compose | 2.0+ | Simple for demo/dev |

## Next Steps (Roadmap)

### v3.0 (Current)
- ✅ Foundation complete
- Core backend API
- Basic collectors
- Grafana integration

### v3.1 (Future)
- Query performance monitoring (pg_stat_statements)
- Advanced alerting rules
- Dashboard customization

### v3.2+
- Kubernetes support (Helm)
- React-based custom UI
- Machine learning anomaly detection
- Multi-region support

---

For detailed setup and deployment, see [DEPLOYMENT.md](DEPLOYMENT.md) and other docs.
