# pgAnalytics Deployment Quick Reference

**Quick lookup guide for standard deployment procedures and configurations.**

---

## Port Mappings (Standard)

| Service | External | Internal | Purpose |
|---------|----------|----------|---------|
| Grafana | **3000** | 3000 | Dashboards |
| Frontend | **4000** | 3000 | React UI |
| Backend | **8080** | 8080 | REST API |
| PostgreSQL | **5432** | 5432 | Metadata DB |
| TimescaleDB | **5433** | 5432 | Metrics DB |
| Collector | (internal) | 9090 | Health only |

---

## Service IPs (Internal Docker Network)

```
172.20.0.10  = postgres
172.20.0.11  = timescale
172.20.0.20  = backend
172.20.0.30  = collector
172.20.0.40  = grafana
172.20.0.60  = frontend
```

---

## Access URLs

```
Frontend:  http://localhost:4000
Grafana:   http://localhost:3000
Backend:   http://localhost:8080/api/v1/health
Database:  localhost:5432 (PostgreSQL)
Database:  localhost:5433 (TimescaleDB)
```

---

## Default Credentials

| Service | Username | Password | Notes |
|---------|----------|----------|-------|
| Grafana | admin | Th101327!!! | CHANGE IN PRODUCTION |
| PostgreSQL | postgres | pganalytics | CHANGE IN PRODUCTION |
| TimescaleDB | postgres | pganalytics | CHANGE IN PRODUCTION |

---

## Essential Commands

### Start Everything
```bash
docker-compose up -d
```

### Stop Everything
```bash
docker-compose down
```

### Stop & Remove Volumes (Fresh Start)
```bash
docker-compose down -v
```

### Check Status
```bash
docker-compose ps
```

### View Logs
```bash
docker-compose logs -f              # All services
docker-compose logs pganalytics-backend     # Specific service
```

### Health Checks
```bash
# Backend health
curl http://localhost:8080/api/v1/health

# Grafana health
curl http://localhost:3000/api/health

# Frontend accessibility
curl http://localhost:4000

# Database connectivity
docker-compose exec postgres pg_isready -U postgres
docker-compose exec timescale pg_isready -U postgres
```

### Collector Status
```bash
docker logs pganalytics-collector-demo | tail -30
```

### Execute Commands in Container
```bash
docker-compose exec postgres psql -U postgres -d pganalytics
docker-compose exec postgres pg_dump -U postgres pganalytics > backup.sql
```

---

## Deployment Workflow

### 1. Fresh Deployment
```bash
# Clean state
docker-compose down -v

# Build images
docker-compose build --no-cache

# Start services
docker-compose up -d

# Wait 30 seconds, then verify
sleep 30
docker-compose ps
```

### 2. Verify Deployment
```bash
# All services healthy?
docker-compose ps

# Can connect to APIs?
curl http://localhost:8080/api/v1/health
curl http://localhost:3000/api/health
curl http://localhost:4000

# No host processes?
ps aux | grep serve | grep -v grep
```

### 3. Commit Changes
```bash
git add docker-compose.yml
git commit -m "Deploy: Fresh containerized stack"
git push origin main
```

---

## Configuration Files

| File | Purpose | Edit For |
|------|---------|----------|
| `docker-compose.yml` | Service definitions | Port changes, volumes, networks |
| `frontend/Dockerfile` | Frontend build | Build process, dependencies |
| `backend/Dockerfile` | Backend build | Build process, runtime |
| `collector/Dockerfile` | Collector build | Build process, libraries |
| `.env` | Environment variables | Secrets, credentials (DO NOT COMMIT) |

---

## Environment Variables Quick Set

### Frontend
```env
VITE_API_BASE_URL=http://backend:8080
```

### Backend
```env
DATABASE_URL=postgres://postgres:pganalytics@postgres:5432/pganalytics?sslmode=disable
TIMESCALE_URL=postgres://postgres:pganalytics@timescale:5432/metrics?sslmode=disable
JWT_SECRET=<change-in-production>
LOG_LEVEL=debug
```

### Collector
```env
BACKEND_URL=http://backend:8080
POSTGRES_HOST=postgres
POSTGRES_USER=postgres
POSTGRES_PASSWORD=pganalytics
```

### Grafana
```env
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=<change-in-production>
```

---

## Troubleshooting Checklist

- [ ] Docker daemon running? `docker ps`
- [ ] All containers up? `docker-compose ps`
- [ ] All healthy? Look for `(healthy)` status
- [ ] Port available? `lsof -i :3000 :4000 :8080`
- [ ] Frontend responsive? `curl http://localhost:4000`
- [ ] Backend responsive? `curl http://localhost:8080/api/v1/health`
- [ ] Can reach databases? `docker-compose exec postgres pg_isready`
- [ ] Check service logs? `docker-compose logs <service>`
- [ ] No host processes? `ps aux | grep serve | grep -v grep`

---

## Common Issues

**Frontend shows "Connection Refused"**
→ Backend not healthy yet, wait 30 seconds

**Port 3000 already in use**
→ Grafana uses 3000, frontend uses 4000

**Database connection failed**
→ Check DATABASE_URL, ensure postgres/timescale healthy

**Collector not collecting**
→ Check BACKEND_URL and POSTGRES_* vars, view logs

**Health check timeout**
→ Normal on slower systems, service likely still functional

---

## Database Backup & Restore

### Backup PostgreSQL
```bash
docker-compose exec postgres pg_dump -U postgres pganalytics > backup.sql
```

### Restore PostgreSQL
```bash
docker-compose exec -T postgres psql -U postgres pganalytics < backup.sql
```

### Backup TimescaleDB
```bash
docker-compose exec timescale pg_dump -U postgres metrics > metrics_backup.sql
```

---

## Performance Tuning

### Collector Interval
**File:** `docker-compose.yml`
```yaml
COLLECTION_INTERVAL: 60  # seconds
```
Increase for lighter load, decrease for more frequent metrics.

### Connection Pooling
**File:** `docker-compose.yml` (Backend config)
```yaml
DATABASE_POOL_SIZE: 10
```

### Frontend Build Optimization
**File:** `frontend/vite.config.ts`
- `minify: 'terser'` - Already optimized
- Source maps disabled in production build

---

## Production Checklist

- [ ] Change all default passwords
- [ ] Generate new JWT_SECRET
- [ ] Deploy TLS certificates to `./tls/`
- [ ] Set `LOG_LEVEL=info` (not debug)
- [ ] Enable database backups
- [ ] Configure monitoring/alerting
- [ ] Set up centralized logging
- [ ] Enable rate limiting on API
- [ ] Configure firewall rules
- [ ] Test disaster recovery plan

---

## Git Workflow

### Check Status
```bash
git status
git diff docker-compose.yml
```

### Commit Deployment
```bash
git add docker-compose.yml
git commit -m "feat: Deploy service updates"
git push origin main
```

### Revert Deployment
```bash
git revert <commit-hash>
docker-compose down -v
docker-compose up -d
```

---

## Useful Docker Commands

```bash
# Remove unused images/volumes
docker system prune -a

# Remove specific image
docker rmi pganalytics-v3-frontend

# Inspect container
docker inspect pganalytics-backend

# View container stats
docker stats

# Copy file from container
docker cp pganalytics-postgres:/tmp/file ./local/path

# Execute SQL in container
docker-compose exec postgres psql -U postgres -c "SELECT * FROM schema.table;"
```

---

## Network Inspection

```bash
# List networks
docker network ls

# Inspect pganalytics network
docker network inspect pganalytics-v3_pganalytics

# Test DNS from container
docker-compose exec backend nslookup postgres
```

---

## Documentation

**Full Standards:** `docs/DEPLOYMENT_STANDARDS.md`
**This Guide:** `docs/DEPLOYMENT_QUICK_REFERENCE.md`
**Recent Reports:** `docs/deployment-reports/`

---

**Updated: 2026-02-27**
**Standards Adopted: ✅ APPROVED**
