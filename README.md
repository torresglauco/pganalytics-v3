# 🦉 pgAnalytics v3.3.0

**PostgreSQL Database Observability & Monitoring Platform**

pgAnalytics is a comprehensive observability platform designed to monitor, analyze, and optimize PostgreSQL databases in production environments. Real-time metrics, intelligent alerting, and deep performance insights—all in one unified dashboard.

**Status**: ✅ Production Ready | v3.3.0

## 📦 PostgreSQL Support

**Supported Versions**: PostgreSQL 14, 15, 16, 17, 18

pgAnalytics v3 is fully compatible with all currently supported PostgreSQL versions:
- ✅ **PG 14** - Baseline version, all migrations compatible
- ✅ **PG 15** - Enhanced performance and ICU collations
- ✅ **PG 16** - Recommended for new deployments
- ✅ **PG 17** - Latest stable with latest features
- ✅ **PG 18** - Future-ready version support

See **[POSTGRES_COMPATIBILITY.md](POSTGRES_COMPATIBILITY.md)** for detailed compatibility matrix and version-specific deployment guidance.

## 🎯 Key Features

### 📊 **Real-Time Monitoring**
- Live database metrics and health indicators
- Connection pool monitoring and management
- Query performance tracking and analysis
- Transaction monitoring and deadlock detection
- Replication lag monitoring

### 🚨 **Intelligent Alerting**
- Configurable alert rules with severity levels
- Incident grouping and root cause analysis
- Multi-channel notifications (Email, Slack, webhooks)
- Alert suppression rules and maintenance windows
- AI-powered optimization suggestions

### 🔍 **Deep Performance Analysis**
- Slow query detection and optimization hints
- Table bloat analysis and vacuum recommendations
- Index usage and fragmentation analysis
- Lock contention and blocking transaction detection
- Cache hit ratio optimization

### 🛠️ **Collector Management**
- Multi-database monitoring across infrastructure
- Lightweight C/C++ collectors with minimal overhead
- Auto-registration and self-healing capabilities
- Real-time metric collection and aggregation

### 👥 **Team & Access Control**
- Role-based access control (Admin, User, Viewer)
- API token generation for programmatic access
- User management and password reset
- Audit logging for compliance

### 📈 **Data & Visualization**
- TimescaleDB-backed time-series storage
- Interactive dashboards and charts
- Custom metric views and drill-downs
- Export capabilities for reporting
- **[ENTERPRISE_INSTALLATION.md](ENTERPRISE_INSTALLATION.md)** - Multi-server installation guide (separate PostgreSQL, API, Collectors, Grafana)
- **[docs/COLLECTOR_REGISTRATION_GUIDE.md](docs/COLLECTOR_REGISTRATION_GUIDE.md)** - Collector registration and JWT authentication

**Complete Documentation Suite:**
- **[INTEGRATION.md](INTEGRATION.md)** - Component architecture and integration guide
- **[TESTING.md](TESTING.md)** - Comprehensive testing guide (unit, integration, E2E)
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Deployment procedures (Docker, K8s, On-premise)
- **[POSTGRES_COMPATIBILITY.md](POSTGRES_COMPATIBILITY.md)** - PostgreSQL version support matrix (14, 15, 16, 17, 18)

**Reference Documentation:**
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System architecture and design
- **[docs/API_SECURITY_REFERENCE.md](docs/API_SECURITY_REFERENCE.md)** - API specifications and security
- **[docs/REPLICATION_COLLECTOR_GUIDE.md](docs/REPLICATION_COLLECTOR_GUIDE.md)** - PostgreSQL replication metrics
- **[SECURITY.md](SECURITY.md)** - Security requirements and best practices

## Quick Start

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [mise](https://mise.jdx.dev/getting-started.html)

### Setup

```bash
# Clone the repo
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3

# Install runtimes (Node.js, Go)
mise install

# Bootstrap the project (idempotent, safe to re-run)
mise run setup

# Start all services
mise run dev
```

### Available Commands

| Command | Description |
|---------|-------------|
| `mise run setup` | Bootstrap project (install deps, generate .env, TLS certs) |
| `mise run dev` | Start all services via Docker Compose |
| `mise run dev-frontend` | Start frontend with hot reload (Vite) |
| `mise run down` | Stop all services |
| `mise run logs` | Follow logs from all services |
| `mise run reset` | Remove all data and start fresh |
| `mise run test` | Run frontend tests |
| `mise run lint` | Run frontend linters |
| `mise run ps` | Show service status |
| `mise run test:postgres:compatibility` | Test migrations on all PostgreSQL versions (14-18) |
| `mise run test:postgres:16` | Test against PostgreSQL 16 (recommended) |

### Access Points
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Grafana**: http://localhost:3001 (admin/Th101327!!!)
- **PostgreSQL**: localhost:5432 (postgres/pganalytics)

## Project Structure

```
pganalytics-v3/
├── mise.toml                   # Runtimes (Node, Go) + dev tasks
├── .env.example                # Environment variables template
├── docker-compose.yml          # Local development services
│
├── backend/                    # Go backend API
│   ├── cmd/pganalytics-api/    # API server entry point
│   ├── internal/               # Core packages
│   ├── migrations/             # Database migrations
│   └── tests/                  # Integration tests
│
├── frontend/                   # React + Vite + TypeScript UI
│   ├── src/                    # Source code
│   ├── package.json            # Dependencies
│   └── vite.config.ts          # Vite configuration
│
├── collector/                  # C/C++ distributed collector
│   ├── src/                    # Source files
│   ├── include/                # Headers
│   └── tests/                  # Unit + integration tests
│
├── grafana/                    # Grafana dashboards & provisioning
├── scripts/                    # Setup and utility scripts
└── docs/                       # Documentation
```

## Key Features

### Backend
- REST API with OpenAPI 3.0 documentation
- Collector registration & certificate management
- Metrics ingestion (JSON + gzip compression)
- Time-series queries and dashboards
- Alert rules with webhooks
- User authentication (JWT tokens)
- Prometheus metrics export

### Collector
- PostgreSQL statistics collection
- System metrics (CPU, memory, disk, I/O)
- Log file processing
- Dynamic configuration (pulled from backend)
- Secure transmission (TLS 1.3 + mTLS)
- Local metrics buffering & retry logic

### Observability
- Grafana dashboards (performance, health, trends)
- Alert rules integrated with Grafana
- Prometheus metrics for both backend and collectors
- Structured JSON logging

## Development

All development tasks are managed via [mise](https://mise.jdx.dev). Run `mise tasks` to see all available commands.

```bash
# Start everything
mise run dev

# Frontend with hot reload (requires backend running)
mise run dev-frontend

# Run tests and linters
mise run test
mise run lint
mise run typecheck

# View logs and service status
mise run logs
mise run ps

# Reset everything (removes databases, volumes)
mise run reset
```

### Building Backend (standalone)
```bash
cd backend
go build -o pganalytics-api ./cmd/pganalytics-api
```

### Building Collector (standalone)
```bash
cd collector
mkdir build && cd build
cmake ..
make
```

## Configuration

### Deployment Configuration (Production)

pgAnalytics v3.2.0 uses environment variables for infrastructure-agnostic deployment. Choose based on your needs:

**For most users:**
```bash
# Copy and fill the open configuration template
cp DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md ~/.env.pganalytics
nano ~/.env.pganalytics         # Fill in YOUR infrastructure details
source ~/.env.pganalytics
bash scripts/phase1_automated_setup.sh
```

**Available configuration templates:**
- `DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md` - **Start here** (no infrastructure assumptions)
- `DEPLOYMENT_CONFIG_TEMPLATE.md` - Full template with all parameters
- `DEPLOYMENT_CONFIG_ENTERPRISE_SCALE.md` - For massive distributed deployments

See **[DEPLOYMENT_START_HERE.md](DEPLOYMENT_START_HERE.md)** for complete deployment instructions.

### Backend (Development Environment Variables)
```bash
DATABASE_URL=postgres://user:pass@localhost/pganalytics
TIMESCALE_URL=postgres://user:pass@localhost/timescale
JWT_SECRET=your-secret-key-here
TLS_CERT=/path/to/cert.pem
TLS_KEY=/path/to/key.pem
PORT=8080
```

### Collector (TOML Configuration)
```toml
[collector]
id = "col_001"
hostname = "db-server-01"
interval = 60

[backend]
url = "https://api.pganalytics.local"
tls_cert = "/etc/pganalytics/collector.crt"
tls_key = "/etc/pganalytics/collector.key"

[postgres]
host = "localhost"
port = 5432
databases = ["postgres", "app_db"]
```

## API Reference

### Collector Management
- `POST /api/v1/collectors/register` - Register new collector
- `GET /api/v1/collectors` - List collectors
- `GET /api/v1/collectors/{id}` - Get collector details

### Metrics
- `POST /api/v1/metrics/push` - Push metrics (secured with mTLS + JWT)
- `GET /api/v1/servers/{id}/metrics` - Query historical metrics

### Configuration
- `GET /api/v1/config/{collector_id}` - Pull collector config
- `PUT /api/v1/config/{collector_id}` - Update config (admin)

### Health
- `GET /api/v1/health` - System health check

Full API documentation available at `/swagger` endpoint.

## Security

- **TLS 1.3**: All connections encrypted
- **mTLS**: Mutual TLS authentication for collectors
- **JWT**: Token-based user authentication
- **SQL Injection Protection**: Prepared statements via SQLC
- **Certificate Management**: Auto-rotation support

See [docs/SECURITY.md](docs/SECURITY.md) for detailed security guidelines.

## Deployment

### Development
```bash
mise run setup    # first time only
mise run dev      # start all services
```

### Production Deployment

**Step 1: Choose your configuration template**
```bash
cp DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md ~/.env.pganalytics
nano ~/.env.pganalytics  # Fill with YOUR infrastructure values
```

**Step 2: Run deployment automation**
```bash
source ~/.env.pganalytics
bash scripts/phase1_automated_setup.sh
```

**Step 3: Follow Phase 1 checklist**
```bash
cat PHASE1_EXECUTION_CHECKLIST_V2.md
```

**Deployment Options:**
- **Docker** - `docker-compose up -d` (development/testing)
- **Standalone** - Manual deployment on physical servers or VMs
- **AWS EC2** - Deploy using configuration template + scripts
- **On-Premises** - Deploy using configuration template + scripts
- **Kubernetes** - Deploy using configuration template + K8s manifests
- **Hybrid** - Mix and match across regions and infrastructure types

**Works with ANY infrastructure:**
- ✅ AWS (EC2, RDS, any region)
- ✅ On-premises (physical machines)
- ✅ Kubernetes (any cluster)
- ✅ Docker (local or remote)
- ✅ Hybrid deployments (mix of everything)

See **[DEPLOYMENT_PLAN_v3.2.0.md](DEPLOYMENT_PLAN_v3.2.0.md)** for complete 4-phase timeline and procedures.

## Performance

- Backend handles 100+ concurrent collectors
- ~1000 metrics per push (~60s interval)
- Target latency: p95 < 500ms
- TimescaleDB optimized for time-series (100K inserts/sec)

## PostgreSQL Version Support

**Currently Supported:**
- PostgreSQL 9.4 - 16 (Full support with all metrics)

**Partial Support (Phase 2):**
- PostgreSQL 17 (Missing I/O context stats - roadmap available)
- PostgreSQL 18 (Missing compression stats - roadmap available)

For version-specific implementation details, see: `docs/POSTGRESQL_VERSION_COMPATIBILITY_REPORT.md`

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Licensed under the same terms as pgAnalytics v2.

## Support & Community

- Issues: GitHub Issues
- Documentation: `/docs` directory
- Examples: [docs/EXAMPLES.md](docs/EXAMPLES.md)

## What's Included in v3.2.0

### ✅ Phase 1 (Production Ready)
- **PostgreSQL Replication Metrics Collector**: 25+ metrics (C++, 1,251 lines)
- **ML-Powered Optimization**: Performance prediction, query rewrites, parameter tuning
- **Enterprise API**: 50+ REST endpoints with JWT authentication, RBAC, rate limiting
- **Grafana Dashboards**: 9 pre-built production-ready dashboards
- **Security**: TLS/mTLS, BCrypt hashing, SQL injection prevention
- **Load Tested**: Validated to 5x scale (5 collectors, 1000+ metrics/cycle)
- **PostgreSQL Support**: Versions 9.4-16 (17-18 roadmap available)

### 🚀 Phase 2 (Roadmap - Future)
- [ ] Kubernetes/Helm support
- [ ] React-based custom UI (optional - REST API is production-ready)
- [ ] PostgreSQL 17/18 full support (analysis and code samples provided)
- [ ] Advanced anomaly detection dashboard
- [ ] Token blacklist implementation
- [ ] CORS origin whitelisting

---

**pgAnalytics v3.0** - Modern PostgreSQL Monitoring for the Cloud Era
