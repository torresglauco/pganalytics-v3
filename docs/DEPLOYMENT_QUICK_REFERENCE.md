# Deployment Quick Reference Guide

## Pre-Deployment Checklist

- [ ] Code reviewed and tested
- [ ] E2E tests pass (all browsers)
- [ ] Backend tests pass (70%+ coverage)
- [ ] Frontend build succeeds
- [ ] Security scanning clean (0 critical/high)
- [ ] Database backups created
- [ ] Deployment runbook reviewed
- [ ] Team notified of deployment window
- [ ] Rollback plan documented
- [ ] Monitoring alerts configured

---

## Quick Deploy (Docker Compose)

### For Demo/Development

```bash
# 1. Clone repository
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3

# 2. Create .env file
cp .env.example .env

# 3. Start services
docker-compose up -d

# 4. Wait for services to be ready
sleep 10

# 5. Access application
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# Grafana: http://localhost:3000 (admin/admin)
# PostgreSQL: localhost:5432 (postgres/pganalytics)

# 6. View logs
docker-compose logs -f

# 7. Stop services
docker-compose down
```

---

## Post-Deployment Verification

```bash
# Health checks
curl -s http://localhost:8080/api/v1/health | jq .
curl -s http://localhost:3000 | head -20

# Database verification
psql -h localhost -U postgres -d pganalytics -c \
  "SELECT COUNT(*) as collector_count FROM collectors"
```

---

## Monitoring After Deployment

### Key Metrics to Watch

```bash
# Backend response time
curl -w "@curl-timing.txt" -o /dev/null -s http://localhost:8080/api/v1/health

# CPU usage
docker stats --no-stream

# Memory usage
docker stats pganalytics-v3_api_1 --format "{{.MemUsage}}"
```

---

## Troubleshooting Deployment Issues

### Service Won't Start

```bash
# Check logs
docker-compose logs api

# Check port conflicts
lsof -i :8080
lsof -i :3000

# Try rebuilding
docker-compose build --no-cache api
docker-compose up -d api
```

---

## Rollback Procedure

### If Deployment Fails

```bash
# 1. Stop current version
docker-compose down

# 2. Restore database from backup
pg_restore backup.sql

# 3. Checkout previous version
git checkout v3.2.0

# 4. Restart with previous version
docker-compose up -d

# 5. Verify
curl http://localhost:8080/api/v1/health
```

---

## Support & Documentation

- **Full Deployment Guide**: See DEPLOYMENT_PLAN_v3.2.0.md
- **Configuration Guide**: See DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md
- **Operations Guide**: See OPERATIONS_HA_DR.md
- **Troubleshooting**: See FAQ_AND_TROUBLESHOOTING.md

---

**Remember**: Always test in staging first! 🚀
