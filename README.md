# pgAnalytics v3.0 - Modernized PostgreSQL Monitoring

A modern, scalable PostgreSQL monitoring and analytics platform built with cutting-edge technologies.

## Architecture Overview

**pgAnalytics v3** is a complete redesign of the PostgreSQL monitoring system with:

- **🚀 High-Performance Backend**: Written in Go with REST API + Swagger documentation
- **📊 Distributed Collectors**: C/C++ agents running on each database server with TLS 1.3 + mTLS security
- **📈 Time-Series Storage**: TimescaleDB for efficient metrics storage and queries
- **🎨 Grafana Integration**: Pre-built dashboards for visualization and alerting
- **🔐 Enterprise Security**: JWT tokens, mutual TLS, encrypted credentials
- **🐳 Cloud-Ready**: Docker Compose for demo, Kubernetes-ready architecture
- **✅ Comprehensive Testing**: Unit, integration, and E2E tests with >70% coverage
- **📚 Full Documentation**: Architecture guide, API reference, deployment guide

## Quick Start (Demo Environment)

### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+
- Make

### Run Demo
```bash
# Clone and setup
git clone https://github.com/dextra/pganalytics-v3.git
cd pganalytics-v3

# Start all services
docker-compose up -d

# Check services
curl http://localhost:8080/api/v1/health      # Backend API
curl http://localhost:3000/api/health         # Grafana
```

### Access Points
- **Backend API**: http://localhost:8080
- **API Swagger UI**: http://localhost:8080/swagger
- **Grafana**: http://localhost:3000 (admin/admin)
- **PostgreSQL**: localhost:5432 (postgres/pganalytics)

## Project Structure

```
pganalytics-v3/
├── backend/                    # Go backend API
│   ├── cmd/pganalytics-api/    # API server entry point
│   ├── internal/               # Core packages
│   │   ├── api/               # HTTP handlers
│   │   ├── auth/              # JWT + mTLS
│   │   ├── collector/         # Collector management
│   │   ├── metrics/           # Metrics ingestion
│   │   ├── storage/           # Database layer
│   │   └── timescale/         # TimescaleDB setup
│   ├── migrations/            # Database migrations
│   └── tests/                 # Integration tests
│
├── collector/                  # C/C++ distributed collector
│   ├── src/                   # Source files
│   ├── include/               # Headers
│   ├── tests/                 # Unit + integration tests
│   └── CMakeLists.txt         # Build configuration
│
├── grafana/                    # Grafana dashboards
│   ├── dashboards/            # Pre-built dashboards
│   └── datasources/           # Data source configs
│
├── docs/                       # Documentation
│   ├── ARCHITECTURE.md
│   ├── API.md
│   ├── SECURITY.md
│   ├── DEPLOYMENT.md
│   └── EXAMPLES.md
│
└── docker-compose.yml         # Demo environment

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

### Building Backend
```bash
cd backend
go build -o pganalytics-api ./cmd/pganalytics-api
```

### Building Collector
```bash
cd collector
mkdir build && cd build
cmake ..
make
```

### Running Tests
```bash
make test-backend       # Go tests
make test-collector     # C++ tests
make test-integration   # E2E tests (requires docker-compose)
```

## Configuration

### Backend (Environment Variables)
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

### Docker
```bash
docker-compose up -d
```

### Kubernetes (Ready)
Helm charts available in `deployments/helm/`

### Standalone
1. Build binaries
2. Set environment variables
3. Run migrations
4. Start backend: `./pganalytics-api`
5. Configure and run collectors

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed instructions.

## Performance

- Backend handles 100+ concurrent collectors
- ~1000 metrics per push (~60s interval)
- Target latency: p95 < 500ms
- TimescaleDB optimized for time-series (100K inserts/sec)

## Supported PostgreSQL Versions

- PostgreSQL 12, 13, 14, 15, 16
- Versions < 12 reach EOL

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Licensed under the same terms as pgAnalytics v2.

## Support & Community

- Issues: GitHub Issues
- Documentation: `/docs` directory
- Examples: [docs/EXAMPLES.md](docs/EXAMPLES.md)

## Roadmap

- [ ] Kubernetes/Helm support
- [ ] React-based custom UI
- [ ] Machine learning anomaly detection
- [ ] Query performance monitoring (pg_stat_statements)
- [ ] Backup/recovery tracking
- [ ] Multi-region replication tracking

---

**pgAnalytics v3.0** - Modern PostgreSQL Monitoring for the Cloud Era
