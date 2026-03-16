# Phase 4 Staging Environment - Quick Reference

## 🚀 Start/Stop Staging Environment

```bash
# Start all services
docker-compose -f docker-compose.staging.yml up -d

# Stop all services
docker-compose -f docker-compose.staging.yml down

# View logs
docker-compose -f docker-compose.staging.yml logs -f

# View specific service logs
docker-compose -f docker-compose.staging.yml logs -f api
docker-compose -f docker-compose.staging.yml logs -f frontend

# Check service status
docker-compose -f docker-compose.staging.yml ps
```

## 🌐 Access Services

| Service | URL | Purpose |
|---------|-----|---------|
| Frontend | http://localhost:3000 | React UI for log management |
| API | http://localhost:8000/api/v1/ | REST API endpoints |
| API Health | http://localhost:8000/api/v1/health | Check API status |
| Prometheus | http://localhost:9090 | Metrics collection |
| Grafana | http://localhost:3001 | Dashboard visualization |
| PostgreSQL | localhost:5432 | Database (user: pganalytics) |
| Redis | localhost:6379 | Cache layer |

## 📝 Environment Configuration

Configuration file: `.env.staging`

```env
DB_PASSWORD=staging-1773664938
JWT_SECRET=staging-jwt-1773664938
GRAFANA_PASSWORD=grafana-1773664938
VITE_API_URL=http://localhost:8000
```

## 🧪 Test API Endpoints

### Health Check
```bash
curl http://localhost:8000/api/v1/health
```

### Through Frontend Proxy
```bash
curl http://localhost:3000/api/v1/health
```

### Get All Alerts (requires authentication)
```bash
TOKEN="your-jwt-token"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8000/api/v1/alerts
```

## 🐳 Docker Commands

### Rebuild specific service
```bash
docker-compose -f docker-compose.staging.yml build --no-cache api
docker-compose -f docker-compose.staging.yml build --no-cache frontend
```

### Restart specific service
```bash
docker-compose -f docker-compose.staging.yml restart api
docker-compose -f docker-compose.staging.yml restart frontend
```

### View container details
```bash
docker ps
docker inspect pganalytics-staging-api
docker stats pganalytics-staging-api
```

## 📊 Monitoring

### View Prometheus metrics
http://localhost:9090/graph

### Create Grafana dashboard
1. Visit http://localhost:3001
2. Login with admin / grafana-1773664938
3. Add Prometheus as data source: http://prometheus:9090
4. Create dashboards

## 🔧 Troubleshooting

### API not responding
```bash
# Check API logs
docker-compose -f docker-compose.staging.yml logs api | tail -50

# Check database connection
curl http://localhost:8000/api/v1/health
```

### Frontend not loading
```bash
# Check frontend logs
docker-compose -f docker-compose.staging.yml logs frontend | tail -20

# Clear browser cache and reload
# Or hard refresh: Ctrl+Shift+R (Windows/Linux) or Cmd+Shift+R (Mac)
```

### Database connection issues
```bash
# Check PostgreSQL logs
docker-compose -f docker-compose.staging.yml logs postgres

# Test connection
psql -h localhost -U pganalytics -d pganalytics_staging
```

## 📦 Database Management

### Connect to PostgreSQL
```bash
psql -h localhost -U pganalytics -d pganalytics_staging
# Password: staging-1773664938
```

### View database contents
```sql
-- List tables
\dt

-- Check alert tables
SELECT * FROM alerts;
SELECT * FROM silences;
SELECT * FROM escalation_policies;
```

## 🔍 Service Health Checks

### Verify all services are healthy
```bash
# Check PostgreSQL
docker-compose -f docker-compose.staging.yml ps postgres

# Check API
docker-compose -f docker-compose.staging.yml ps api

# Check Frontend
docker-compose -f docker-compose.staging.yml ps frontend

# Check Redis
docker-compose -f docker-compose.staging.yml ps redis

# Check Prometheus
docker-compose -f docker-compose.staging.yml ps prometheus

# Check Grafana
docker-compose -f docker-compose.staging.yml ps grafana
```

## 🚨 Common Issues & Solutions

### "Port 3000 already in use"
```bash
# Find and kill process using port 3000
lsof -i :3000
kill -9 <PID>
```

### "Database connection failed"
```bash
# Ensure .env.staging is correct
cat .env.staging

# Restart database
docker-compose -f docker-compose.staging.yml restart postgres
```

### "API routes not loading"
```bash
# Rebuild API without cache
docker-compose -f docker-compose.staging.yml build --no-cache api

# Restart API
docker-compose -f docker-compose.staging.yml restart api
```

### "Frontend not connecting to API"
```bash
# Check proxy logs
docker-compose -f docker-compose.staging.yml logs frontend | grep -i proxy

# Test API health
curl http://localhost:3000/api/v1/health
```

## 📄 Key Files

| File | Purpose |
|------|---------|
| `docker-compose.staging.yml` | Service orchestration |
| `.env.staging` | Environment variables |
| `frontend/Dockerfile` | Frontend build & proxy config |
| `backend/Dockerfile` | Backend build config |
| `STAGING_DEPLOYMENT_PLAN.md` | Deployment documentation |
| `STAGING_DEPLOYMENT_COMPLETE.md` | Deployment status |
| `API_TEST_REPORT.md` | Test results |

## 🎯 Phase 4 Features to Test

1. **Custom Alert Conditions**
   - Endpoint: GET/POST `/api/v1/alerts`
   - Test at: http://localhost:3000 → Alerts menu

2. **Alert Silencing**
   - Endpoint: POST `/api/v1/alerts/:id/silence`
   - Test at: http://localhost:3000 → Silence alerts

3. **Escalation Policies**
   - Endpoint: POST/GET `/api/v1/escalation-policies`
   - Test at: http://localhost:3000 → Escalation settings

## 📞 Support

For issues or questions:
1. Check logs: `docker-compose -f docker-compose.staging.yml logs -f`
2. Review documentation: `API_TEST_REPORT.md`, `STAGING_DEPLOYMENT_COMPLETE.md`
3. Check service health: `docker-compose -f docker-compose.staging.yml ps`

---

**Last Updated**: 2026-03-16
**Version**: Phase 4 v4.0.0
