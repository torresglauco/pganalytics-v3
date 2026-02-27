# pgAnalytics Deployment Standards

**Last Updated:** 2026-02-27
**Status:** APPROVED & ADOPTED

---

## Deployment Philosophy

**100% Containerization:** All pgAnalytics services run within Docker containers. Zero services run directly on the host machine. This ensures environment consistency, easy scaling, and clean local development.

---

## Network Configuration Standards

### Docker Network
- **Name Pattern:** `pganalytics_<network-name>`
- **Type:** Custom bridge driver
- **Subnet:** `172.20.0.0/16`
- **Gateway:** `172.20.0.1`
- **Isolation:** All services communicate via Docker internal DNS, not host ports

### Service IP Assignment Table
Standard IP assignments for consistent service discovery:

| Service | Container Name | Internal IP | Internal Port | External Port |
|---------|---------------|------------|---------------|---------------|
| PostgreSQL | pganalytics-postgres | 172.20.0.10 | 5432 | 5432 |
| TimescaleDB | pganalytics-timescale | 172.20.0.11 | 5432 | 5433 |
| Backend API | pganalytics-backend | 172.20.0.20 | 8080 | 8080 |
| C++ Collector | pganalytics-collector-demo | 172.20.0.30 | (internal) | 9090 |
| Grafana | pganalytics-grafana | 172.20.0.40 | 3000 | 3000 |
| Frontend UI | pganalytics-frontend | 172.20.0.60 | 3000 | 4000 |

**Note:** Frontend uses internal port 3000 but maps to external port **4000** to avoid conflict with Grafana.

---

## Port Mapping Standard

### Mandatory Port Assignments
- **3000** → Grafana dashboards (RESERVED)
- **4000** → Frontend UI (STANDARD for frontend external access)
- **5432** → PostgreSQL (STANDARD)
- **5433** → TimescaleDB (STANDARD - different port to run alongside PostgreSQL)
- **8080** → Backend API (STANDARD)

### Port Conflict Resolution
When two services want the same internal port:
- Assign different external ports
- Document in docker-compose.yml clearly
- Frontend example: internal 3000 → external 4000

---

## Service Dependency Standards

### Health Check Dependencies
Use Docker Compose `depends_on` with `condition: service_healthy` for critical dependencies:

```yaml
# Example: Backend requires both databases healthy before starting
depends_on:
  postgres:
    condition: service_healthy
  timescale:
    condition: service_healthy
```

### Dependency Chain
```
Frontend → Backend → (PostgreSQL + TimescaleDB)
Collector → Backend → (PostgreSQL + TimescaleDB)
Grafana → (PostgreSQL + TimescaleDB)
```

---

## Health Check Standards

### Standard Health Check Configurations

**PostgreSQL / TimescaleDB:**
```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U postgres"]
  interval: 5s
  timeout: 5s
  retries: 5
```

**Backend API:**
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/api/v1/health"]
  interval: 10s
  timeout: 5s
  retries: 3
  start_period: 10s
```

**Grafana:**
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
  interval: 10s
  timeout: 5s
  retries: 3
```

**Frontend:**
```yaml
healthcheck:
  test: ["CMD", "curl", "-s", "http://localhost:3000"]
  interval: 30s
  timeout: 5s
  start_period: 10s
  retries: 3
```

---

## Docker Image Standards

### Image Naming Convention
- **Format:** `pganalytics-v3-<service>:latest`
- **Examples:**
  - pganalytics-v3-backend
  - pganalytics-v3-collector
  - pganalytics-v3-frontend

### Multi-Stage Build Strategy
**Backend (Go):**
- Stage 1: `golang:1.22-alpine` - compilation
- Stage 2: `alpine:3.19` - runtime (minimal)

**Collector (C/C++):**
- Stage 1: `ubuntu:22.04` - build environment
- Stage 2: `ubuntu:22.04` - runtime with only necessary libs

**Frontend (Node.js):**
- Stage 1: `node:18-alpine` - build & npm install
- Stage 2: `node:18-alpine` - runtime with `serve` package

### Build Context
- **Always use:** Project root (`.`) as build context
- **COPY paths:** Use relative paths from root (e.g., `COPY frontend/package*.json ./`)
- **Build consistency:** Use `--no-cache` for fresh rebuilds

---

## Environment Variables Standards

### Backend Configuration
```env
DATABASE_URL=postgres://postgres:pganalytics@postgres:5432/pganalytics?sslmode=disable
TIMESCALE_URL=postgres://postgres:pganalytics@timescale:5432/metrics?sslmode=disable
JWT_SECRET=demo-secret-key-change-in-production
JWT_EXPIRATION=900
TLS_CERT=/etc/pganalytics/tls/server.crt
TLS_KEY=/etc/pganalytics/tls/server.key
LOG_LEVEL=debug
PORT=8080
```

### Collector Configuration
```env
COLLECTOR_ID=col_demo_001
COLLECTOR_NAME=Demo Collector
BACKEND_URL=http://backend:8080
BACKEND_TLS_VERIFY=false
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=pganalytics
POSTGRES_DATABASES=postgres,pganalytics
COLLECTION_INTERVAL=60
LOG_LEVEL=debug
```

### Frontend Configuration
```env
VITE_API_BASE_URL=http://backend:8080
```

### Grafana Configuration
```env
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=<CHANGE_IN_PRODUCTION>
GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-worldmap-panel
GF_USERS_ALLOW_SIGN_UP=false
```

### Internal Service DNS
Services use Docker container names for communication:
- `postgres:5432` - PostgreSQL metadata
- `timescale:5432` - TimescaleDB metrics
- `backend:8080` - Backend API
- `grafana:3000` - Grafana dashboards

---

## Volume Standards

### Named Volumes
Use named volumes for persistent data, not host mounts:

```yaml
volumes:
  postgres_data:
    driver: local
  timescale_data:
    driver: local
  grafana_data:
    driver: local
  collector_data:
    driver: local
```

### Mount Paths
| Service | Mount Point | Purpose |
|---------|------------|---------|
| PostgreSQL | /var/lib/postgresql/data | Database files |
| TimescaleDB | /var/lib/postgresql/data | Time-series data |
| Grafana | /var/lib/grafana | Dashboards & config |
| Collector | /var/lib/pganalytics | Collector cache |

### Read-Only Mounts
Use read-only mounts for configuration and migrations:
```yaml
volumes:
  - ./backend/migrations:/migrations:ro
  - ./grafana/provisioning:/etc/grafana/provisioning:ro
  - ./tls:/etc/pganalytics/tls:ro
```

---

## Database Initialization Standards

### PostgreSQL Init Scripts
- Location: `./backend/migrations/`
- Scripts run automatically on container first start
- Database created: `pganalytics`
- User/password: `postgres`/`pganalytics`

### TimescaleDB Init
- Image: `postgres:16-bullseye` (standard PostgreSQL)
- TimescaleDB extension installed at startup
- Database created: `metrics`
- Configured for time-series data storage

---

## Frontend Build Standards

### Vite Configuration
```javascript
// vite.config.ts
export default defineConfig({
  root: './',
  publicDir: 'public',
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser'
  }
})
```

### Dockerfile Standards
```dockerfile
# Stage 1: Build
FROM node:18-alpine AS builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend /app
RUN npm run build

# Stage 2: Runtime
FROM node:18-alpine
WORKDIR /app
RUN npm install -g serve
COPY --from=builder /app/dist ./dist
EXPOSE 3000
CMD ["serve", "-s", "dist", "-l", "3000"]
```

### API Configuration
- **Base URL Pattern:** `VITE_API_BASE_URL=http://backend:8080`
- **Environment:** Set at build time via docker-compose environment
- **Proxy:** Development mode can use vite proxy to `http://localhost:8080`

---

## Deployment Checklist

Use this checklist for every fresh deployment:

### Pre-Deployment
- [ ] Clean slate: `docker-compose down -v`
- [ ] Verify no host processes: `lsof -i :3000 :4000 :5432 :5433 :8080`
- [ ] Docker daemon running: `docker ps`

### Build Phase
- [ ] Build backend: Verify `./pganalytics-api` binary created
- [ ] Build collector: Verify C++ compilation successful
- [ ] Build frontend: Verify `dist/` folder with assets created
- [ ] Check image sizes: Backend ~60MB, Collector ~120MB, Frontend ~200MB

### Deployment Phase
- [ ] Start services: `docker-compose up -d`
- [ ] Wait 30-60 seconds for health checks
- [ ] Verify all containers running: `docker-compose ps`

### Post-Deployment Validation
- [ ] PostgreSQL: `docker-compose exec postgres pg_isready -U postgres`
- [ ] TimescaleDB: `docker-compose exec timescale pg_isready -U postgres`
- [ ] Backend health: `curl http://localhost:8080/api/v1/health`
- [ ] Frontend: `curl http://localhost:4000` (should return HTML)
- [ ] Grafana: `curl http://localhost:3000` (should return 302)
- [ ] Collector logs: `docker logs pganalytics-collector-demo` (should show metrics)
- [ ] No host processes: Verify all services containerized

### Verification Commands
```bash
# Full stack health check
docker-compose ps
curl http://localhost:4000
curl http://localhost:3000
curl http://localhost:8080/api/v1/health

# Verify isolation
ps aux | grep -E "serve|npm|node" | grep -v grep
lsof -i :3000 :4000 :5432 :5433 :8080

# View logs
docker-compose logs -f
docker logs pganalytics-collector-demo
```

---

## Production Deployment Considerations

### Before Going to Production

1. **Credentials & Secrets**
   - Change all default passwords
   - Generate new `JWT_SECRET` (minimum 32 characters)
   - Update Grafana admin password
   - Use secrets management (Docker Secrets, Kubernetes, etc.)

2. **TLS/HTTPS**
   - Generate certificates: `./tls/server.crt` and `./tls/server.key`
   - Set `BACKEND_TLS_VERIFY=true` in collector config
   - Configure reverse proxy (nginx, Caddy) for HTTPS

3. **Network Security**
   - Restrict external access to necessary ports only
   - Use firewall rules to limit port exposure
   - Consider VPN for remote access

4. **Data Persistence**
   - Configure regular backups of volumes
   - Test restore procedures
   - Document backup retention policies

5. **Monitoring & Logging**
   - Set up centralized logging (ELK, Loki, etc.)
   - Configure alerts for service failures
   - Monitor resource usage (CPU, memory, disk)

6. **Performance**
   - Monitor collector lag and adjust `COLLECTION_INTERVAL`
   - Configure database indexes for query performance
   - Set up caching layer (Redis) if needed

---

## Troubleshooting Standards

### Common Issues & Solutions

**Frontend health check timeout:**
- Service still functional if accessible at http://localhost:4000
- May indicate slow startup on resource-constrained systems
- Increase `start_period` if needed

**Collector logs show pg_stat_statements warnings:**
- Normal and non-blocking
- Optional PostgreSQL extension
- Core functionality unaffected

**Database connection refused:**
- Verify services healthy: `docker-compose ps`
- Check logs: `docker-compose logs postgres`
- Ensure correct credentials and database names

**API returns 502/503:**
- Backend may still be starting
- Check health: `docker logs pganalytics-backend`
- Verify database connectivity

**Port already in use:**
- Stop conflicting container: `docker-compose down`
- Or use different external port in docker-compose.yml

---

## Git Commit Standards

### Deployment Commits
```
feat: Complete Docker deployment with all services

- Deploy backend, collector, frontend, databases
- Configure isolated network and health checks
- Verify all services operational and accessible
- Confirm zero host processes

Deployment Status: SUCCESSFUL

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

### Configuration Updates
```
feat: Update service configuration

- Change environment variables
- Add new volumes or mounts
- Update health check parameters

Testing: Verified with docker-compose up and health checks

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
```

---

## File Structure Standards

```
pganalytics-v3/
├── docker-compose.yml          # Single source of truth for deployment
├── backend/
│   ├── Dockerfile             # Multi-stage Go build
│   └── migrations/            # SQL initialization scripts
├── collector/
│   ├── Dockerfile             # Multi-stage C++ build
│   ├── entrypoint.sh          # Dynamic config generation
│   ├── config.toml.sample     # Configuration template
│   └── src/                   # C++ source
├── frontend/
│   ├── Dockerfile             # Multi-stage Node.js build
│   ├── vite.config.ts         # Build configuration
│   ├── package.json           # Dependencies
│   └── src/                   # React source
├── grafana/
│   ├── provisioning/          # Data sources & dashboards
│   └── dashboards/            # Dashboard definitions
├── tls/                        # SSL/TLS certificates
├── docs/
│   ├── DEPLOYMENT_STANDARDS.md  # This file
│   └── deployment-reports/      # Historical deployment records
└── scripts/                    # Utility scripts
```

---

## Monitoring & Observability Standards

### Service Health Endpoints
- Backend: `GET /api/v1/health` - Returns JSON with status
- Grafana: `GET /api/health` - Returns JSON with database status
- Frontend: `GET /` - Returns HTML (status 200)
- PostgreSQL: `pg_isready` command
- TimescaleDB: `pg_isready` command

### Recommended Monitoring Setup
1. Docker container monitoring (memory, CPU, I/O)
2. Application-level metrics (latency, errors, throughput)
3. Database performance (query times, connection pool)
4. Collector metrics (collection lag, metric count)
5. System logs aggregation (centralized logging)

---

## Change Management

### Updating Service Configuration
1. Modify docker-compose.yml or relevant Dockerfile
2. Test locally: `docker-compose up -d && docker-compose ps`
3. Verify health checks pass
4. Commit changes with descriptive message
5. Deploy to production following checklist

### Zero-Downtime Deployments
```bash
# Rolling update approach
docker-compose up -d <service-name>  # Redeploys single service
docker-compose ps                     # Verify health
```

### Rollback Strategy
```bash
# Revert to previous commit
git checkout <previous-commit>
docker-compose down -v
docker-compose up -d
```

---

## References & Related Documentation

- Docker Compose Documentation: https://docs.docker.com/compose/
- Grafana Dashboard Documentation: https://grafana.com/docs/
- PostgreSQL Official: https://www.postgresql.org/docs/
- TimescaleDB Documentation: https://docs.timescale.com/
- Vite Build Tool: https://vitejs.dev/

---

**These standards are mandatory for all pgAnalytics deployments going forward.**

*Document Owner: Development Team*
*Last Review: 2026-02-27*
*Next Review: Quarterly*
