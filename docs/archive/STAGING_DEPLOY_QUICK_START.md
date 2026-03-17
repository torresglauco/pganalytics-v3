# Staging Deployment Quick Start - Phase 4 v4.0.0

> **⚡ TL;DR:** Run one command to deploy Phase 4 to staging with full testing and monitoring

---

## 🚀 One-Command Deployment

```bash
cd /Users/glauco.torres/git/pganalytics-v3
./scripts/deploy-staging.sh
```

**That's it!** The script will:
- ✅ Verify prerequisites (Docker, Docker Compose)
- ✅ Create secure environment variables
- ✅ Build Docker images
- ✅ Start all services (PostgreSQL, API, Frontend, Redis, Prometheus, Grafana)
- ✅ Apply database migrations
- ✅ Run health checks
- ✅ Execute smoke tests
- ✅ Show access information

**Time:** ~5-10 minutes (including Docker image builds)

---

## 📊 After Deployment

### Access the Application

| Service | URL | Purpose |
|---------|-----|---------|
| **Frontend** | http://localhost:3000 | Main UI application |
| **API** | http://localhost:8000 | REST API endpoints |
| **API Health** | http://localhost:8000/health | Health check |
| **Prometheus** | http://localhost:9090 | Metrics collection |
| **Grafana** | http://localhost:3001 | Dashboards (admin/password in .env.staging) |

### Check Service Status

```bash
cd /Users/glauco.torres/git/pganalytics-v3
docker-compose -f docker-compose.staging.yml ps
```

**Expected Output:**
```
NAME                              STATUS
pganalytics-staging-db           Up (healthy)
pganalytics-staging-api          Up (healthy)
pganalytics-staging-frontend     Up
pganalytics-staging-redis        Up
pganalytics-staging-prometheus   Up
pganalytics-staging-grafana      Up
```

### View Logs

```bash
# API logs
docker-compose -f docker-compose.staging.yml logs api -f

# Database logs
docker-compose -f docker-compose.staging.yml logs postgres

# All logs
docker-compose -f docker-compose.staging.yml logs -f
```

---

## 🧪 Manual Smoke Tests

If you want to test features manually:

### 1. API Health Check

```bash
curl http://localhost:8000/health | jq .
```

### 2. Create an Alert Rule

```bash
curl -X POST http://localhost:8000/api/v1/alert-rules \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -H "X-Instance-ID: test-instance" \
  -d '{
    "name": "High Error Count",
    "conditions": [{
      "metric_type": "error_count",
      "operator": ">",
      "threshold": 10,
      "time_window": 5,
      "duration": 300
    }]
  }' | jq .
```

### 3. Create an Alert Silence

```bash
curl -X POST http://localhost:8000/api/v1/alert-silences \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -H "X-Instance-ID: test-instance" \
  -d '{
    "alert_rule_id": "uuid-from-previous-step",
    "duration_seconds": 3600,
    "reason": "Maintenance window"
  }' | jq .
```

### 4. Create an Escalation Policy

```bash
curl -X POST http://localhost:8000/api/v1/escalation-policies \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -H "X-Instance-ID: test-instance" \
  -d '{
    "name": "Standard Escalation",
    "steps": [{
      "step_number": 1,
      "wait_minutes": 5,
      "notification_channel": "email",
      "channel_config": {"email": "test@example.com"}
    }]
  }' | jq .
```

---

## 📈 Monitoring

### Prometheus Metrics

Access http://localhost:9090 to query metrics:

```
# API request rate
rate(http_requests_total[1m])

# Error rate
rate(http_request_errors_total[1m])

# Database connections
pg_stat_activity

# API latency
http_request_duration_seconds
```

### Grafana Dashboards

Access http://localhost:3001:
- **Username:** admin
- **Password:** Check `.env.staging` file

Dashboards available:
- API Performance
- Database Metrics
- System Resources

---

## ⏹️ Stop Services

```bash
# Stop all services (keep data)
docker-compose -f docker-compose.staging.yml stop

# Stop and remove containers (keep data)
docker-compose -f docker-compose.staging.yml down

# Stop and delete everything (including data)
docker-compose -f docker-compose.staging.yml down -v
```

---

## 🔄 Restart Services

```bash
# Restart all services
docker-compose -f docker-compose.staging.yml restart

# Restart specific service
docker-compose -f docker-compose.staging.yml restart api
```

---

## 🐛 Troubleshooting

### Services not starting?

```bash
# Check Docker Compose configuration
docker-compose -f docker-compose.staging.yml config

# Check service logs
docker-compose -f docker-compose.staging.yml logs postgres
docker-compose -f docker-compose.staging.yml logs api
docker-compose -f docker-compose.staging.yml logs frontend
```

### Database connection failing?

```bash
# Test database connection
docker-compose -f docker-compose.staging.yml exec postgres \
  pg_isready -U pganalytics -d pganalytics_staging

# Check database size
docker-compose -f docker-compose.staging.yml exec postgres \
  psql -U pganalytics -d pganalytics_staging -c "\l"
```

### API not responding?

```bash
# Test API connectivity
docker-compose -f docker-compose.staging.yml exec api \
  curl -s http://localhost:8000/health

# Check API port bindings
docker-compose -f docker-compose.staging.yml port api
```

### Frontend showing API errors?

```bash
# Check frontend build
docker-compose -f docker-compose.staging.yml logs frontend

# Verify API URL configuration
grep VITE_API_URL .env.staging
```

---

## 🚨 Rollback to Previous Version

```bash
# Stop current deployment
docker-compose -f docker-compose.staging.yml down

# Checkout previous version
git checkout v3.4.0

# Start previous version
docker-compose -f docker-compose.staging.yml up -d
```

---

## 📋 Deployment Options

### Option 1: Automated Docker (Recommended)
```bash
./scripts/deploy-staging.sh
```
Time: ~10 minutes | Complexity: Low

### Option 2: Skip Building Images
```bash
./scripts/deploy-staging.sh --skip-build
```
Time: ~5 minutes | Use when: Images already built

### Option 3: Skip Tests
```bash
./scripts/deploy-staging.sh --skip-tests
```
Time: ~5 minutes | Use when: Testing manually

### Option 4: Manual Docker Commands
```bash
docker-compose -f docker-compose.staging.yml up -d
sleep 30
docker-compose -f docker-compose.staging.yml ps
curl http://localhost:8000/health
```
Time: ~5-10 minutes | Complexity: Low

### Option 5: Manual Deployment
See `docs/PHASE4_STAGING_DEPLOYMENT.md` for detailed manual setup
Time: ~30-45 minutes | Complexity: Medium

---

## 📚 Full Documentation

For detailed deployment procedures, see:

- **Quick Deployment:** This file (you are here)
- **Complete Plan:** `docs/STAGING_DEPLOYMENT_PLAN.md` (5,000+ lines)
- **Manual Setup:** `docs/PHASE4_STAGING_DEPLOYMENT.md` (2,000+ lines)
- **Architecture:** `docs/PHASE4_ADVANCED_UI_IMPLEMENTATION.md`
- **Release Info:** `RELEASE_NOTES_v4.0.0.md`

---

## ✅ Success Indicators

✅ **Deployment successful when:**
- All containers show "Up" status
- `curl http://localhost:8000/health` returns 200
- `http://localhost:3000` loads the frontend
- Smoke tests complete (8/8 passing or skipped)
- Logs show no errors

---

## 🎯 What's Deployed

**Phase 4 Features:**
- ✅ Custom Alert Conditions (flexible metric/operator combinations)
- ✅ Alert Silencing (TTL-based auto-expiration)
- ✅ Escalation Policies (multi-step alert routing)

**Infrastructure:**
- ✅ PostgreSQL (5 new tables, 10 indices)
- ✅ Backend API (5 services, 8 endpoints)
- ✅ Frontend (React/TypeScript, 6+ components)
- ✅ Monitoring (Prometheus + Grafana)

**Quality:**
- ✅ 301 tests (100% passing)
- ✅ 95%+ backend coverage
- ✅ 89%+ frontend coverage
- ✅ Zero build errors

---

## 📞 Next Steps

1. **Immediate:** Deploy using `./scripts/deploy-staging.sh`
2. **First hour:** Run smoke tests and verify all features
3. **First day:** Monitor metrics, test with real data
4. **This week:** Complete user acceptance testing
5. **Next step:** Production deployment planning

---

**Version:** 1.0
**Updated:** 2026-03-14
**Status:** Ready for Immediate Deployment ✅
