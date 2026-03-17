# Staging Environment Access Guide
## pgAnalytics v3.3.0 - Local Sandbox

**Environment**: Docker Compose (localhost)
**Updated**: March 11, 2026
**Status**: 🟢 ACTIVE

---

## Quick Access URLs

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3001 | admin / staging_admin |
| **Prometheus** | http://localhost:9090 | (no auth) |
| **PostgreSQL** | localhost:5432 | postgres / staging_password |
| **TimescaleDB** | localhost:5433 | postgres / staging_password |
| **API Backend** | https://localhost:8080 | (coming Wednesday) |
| **Frontend** | http://localhost:3000 | (coming Wednesday) |

---

## Database Access

### PostgreSQL Main Database
```bash
# Connection details
Host: localhost (or postgres-staging via Docker network)
Port: 5432
Username: postgres
Password: staging_password
Database: pganalytics_staging

# Connect locally
psql -h localhost -U postgres -d pganalytics_staging

# Connect via Docker
docker exec pganalytics-staging-postgres psql -U postgres -d pganalytics_staging

# Verify connection
docker exec pganalytics-staging-postgres pg_isready -U postgres
```

### TimescaleDB Metrics Database
```bash
# Connection details
Host: localhost (or timescale-staging via Docker network)
Port: 5433
Username: postgres
Password: staging_password
Database: metrics_staging

# Connect locally
psql -h localhost -p 5433 -U postgres -d metrics_staging

# Connect via Docker
docker exec pganalytics-staging-timescale psql -U postgres -d metrics_staging
```

### Current Schema
```sql
-- Main pganalytics database
Schema: pganalytics

Tables:
- servers (id, name, hostname, port, created_at, updated_at)
  └── Example: staging_local, localhost, 5432

- collectors (id, name, status, last_heartbeat, created_at, updated_at)
  └── Example: staging-collector-1, online

-- Metrics database (TimescaleDB)
Database: metrics_staging (empty, ready for time-series data)
```

---

## Monitoring Stack

### Prometheus
```bash
# Web UI
http://localhost:9090

# Query Metrics
curl http://localhost:9090/api/v1/query?query=up

# Scrape Targets
http://localhost:9090/targets

# Configuration
File: ./monitoring/prometheus.staging.yml
Reload: docker-compose restart prometheus-staging
```

### Grafana
```bash
# Web UI
http://localhost:3001

# Login
Username: admin
Password: staging_admin

# Datasources (auto-configured)
- PostgreSQL Staging (postgres-staging:5432)
- TimescaleDB Staging (timescale-staging:5432)
- Prometheus (prometheus-staging:9090)

# Dashboards (provisioned)
- Advanced Features Analysis
- System Metrics Breakdown
- Query Performance
- Infrastructure Stats
- Query Stats Performance
- PostgreSQL Query by Hostname
- Replication Advanced Analytics
- Multi-Collector Monitor
- Replication Health Monitor

Location: ./grafana/dashboards/
```

---

## Docker Network

### Container Connectivity
All containers run on isolated Docker network: `pganalytics-staging` (172.21.0.0/16)

| Service | Internal IP | External Port |
|---------|------------|---------------|
| PostgreSQL | 172.21.0.10 | 5432 |
| TimescaleDB | 172.21.0.11 | 5433 |
| Backend (soon) | 172.21.0.20 | 8080 |
| Frontend (soon) | 172.21.0.30 | 3000 |
| Prometheus | 172.21.0.40 | 9090 |
| Grafana | 172.21.0.41 | 3001 |

### DNS within Docker
Services can reference each other by container name:
```bash
# From backend container to PostgreSQL
DATABASE_URL=postgres://postgres:staging_password@postgres-staging:5432/pganalytics_staging

# From frontend to backend
REACT_APP_API_URL=https://backend-staging:8080
```

---

## Resource Limits

### CPU & Memory Allocation
```yaml
PostgreSQL:   2 CPU / 2GB memory (limit), 1 CPU / 1GB (reservation)
TimescaleDB:  2 CPU / 2GB memory (limit), 1 CPU / 1GB (reservation)
Backend:      2 CPU / 1GB memory (limit), 1 CPU / 512MB (reservation)
Frontend:     1 CPU / 512MB memory (limit), 0.5 CPU / 256MB (reservation)
Prometheus:   1 CPU / 512MB memory (limit), 0.5 CPU / 256MB (reservation)
Grafana:      1 CPU / 512MB memory (limit), 0.5 CPU / 256MB (reservation)

Total: ~8 CPU / 6GB memory allocated
Typical Usage: <1 CPU / 1.2GB memory (idle)
```

---

## Service Management

### Start All Services
```bash
docker-compose -f docker-compose.staging.yml up -d
```

### Stop All Services
```bash
docker-compose -f docker-compose.staging.yml down
```

### Stop and Remove Volumes (clean slate)
```bash
docker-compose -f docker-compose.staging.yml down -v
```

### View Logs
```bash
# All services
docker-compose -f docker-compose.staging.yml logs -f

# Specific service
docker-compose -f docker-compose.staging.yml logs -f postgres-staging
docker-compose -f docker-compose.staging.yml logs -f grafana-staging
docker-compose -f docker-compose.staging.yml logs -f prometheus-staging
```

### Status
```bash
docker-compose -f docker-compose.staging.yml ps
docker-compose -f docker-compose.staging.yml ps -a
```

---

## Backend API (Coming Wednesday)

### Expected Endpoints
```bash
# Health check
curl -k https://localhost:8080/api/v1/health

# Get collectors
curl -k -H "Authorization: Bearer $TOKEN" https://localhost:8080/api/v1/collectors

# Get servers
curl -k -H "Authorization: Bearer $TOKEN" https://localhost:8080/api/v1/servers

# Get metrics
curl -k https://localhost:8080/api/v1/metrics/prometheus
```

### Authentication
```bash
# JWT Secret (staging)
JWT_SECRET=staging-jwt-secret-change-in-production
JWT_EXPIRATION=3600 (1 hour)

# Registration Secret
REGISTRATION_SECRET=staging-registration-secret

# TLS Certificate
Self-signed certificate will be generated
Path: ./tls/server.crt and ./tls/server.key
```

---

## Frontend Application (Coming Wednesday)

### Build & Deploy
```bash
# The frontend will be automatically built by Docker Compose
# Access via: http://localhost:3000

# Environment Variables (configured in docker-compose.staging.yml)
REACT_APP_API_URL=https://localhost:8080
REACT_APP_ENVIRONMENT=staging
```

### Expected Features
- Real-time PostgreSQL monitoring
- Query performance analytics
- Alert management and notifications
- Dashboard customization
- Data export capabilities

---

## Troubleshooting

### PostgreSQL Connection Issues
```bash
# Check if container is running
docker-compose -f docker-compose.staging.yml ps postgres-staging

# Check logs
docker-compose -f docker-compose.staging.yml logs postgres-staging

# Test connectivity
docker exec pganalytics-staging-postgres pg_isready -U postgres -h localhost

# Manual connection test
psql -h localhost -U postgres -d postgres -c "SELECT version();"
```

### Prometheus Not Scraping Metrics
```bash
# Check configuration
curl http://localhost:9090/api/v1/status/config | jq .

# Check targets
curl http://localhost:9090/api/v1/targets | jq .

# Reload configuration (after editing prometheus.staging.yml)
docker-compose restart prometheus-staging
```

### Grafana Dashboards Not Showing
```bash
# Verify datasources are connected
- Login to http://localhost:3001
- Go to Configuration > Data Sources
- Click on each datasource and click "Test"

# Check dashboard provisioning
- Go to Dashboards > Browse
- Verify dashboards are listed under "pgAnalytics" folder

# Reload provisioning
docker-compose restart grafana-staging
```

### Out of Memory
```bash
# Check memory usage
docker stats

# Increase Docker desktop memory limit:
# Docker Desktop Settings > Resources > Memory

# Reduce container limits in docker-compose.staging.yml
# (Not recommended - already optimized for staging)
```

---

## Next Steps

### Wednesday (March 13)
- Deploy Backend API server
- Deploy Frontend React application
- Test API endpoints
- Configure health checks

### Thursday (March 14)
- Run smoke tests
- Security validation
- Performance baseline testing
- Setup alerting rules

### Friday (March 15)
- 24-hour monitoring period
- Documentation finalization
- Team sign-offs
- Preparation for production deployment

---

## Environment Details

**Created**: March 11, 2026
**Maintainer**: pgAnalytics Team
**Version**: v3.3.0
**Docker Compose**: 2.0+
**Docker**: 20.10+

