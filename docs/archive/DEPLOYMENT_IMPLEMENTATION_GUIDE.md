# pgAnalytics v3.2.0 - Production Deployment Implementation Guide

**Status**: ✅ Production Approval Granted
**Date**: February 26, 2026
**Version**: 3.2.0
**Last Updated**: February 26, 2026

---

## Overview

This guide provides step-by-step instructions for deploying pgAnalytics v3.2.0 to production following the audit approval (PRODUCTION_APPROVAL.md).

**Approval Status**: ✅ Approved by Glauco Torres on February 26, 2026
**Supported Configuration**: 1-50 concurrent collectors

---

## Pre-Deployment Checklist

### Week Before Deployment

**Planning & Preparation**:
- [ ] Notify stakeholders of deployment date and maintenance window
- [ ] Schedule deployment window (recommend off-peak hours)
- [ ] Assign deployment team members
- [ ] Create deployment runbook (use template below)
- [ ] Plan rollback procedure
- [ ] Get team sign-offs on deployment plan

**Infrastructure Preparation**:
- [ ] Verify production servers are available
- [ ] Check storage capacity (30GB minimum recommended)
- [ ] Verify network connectivity
- [ ] Confirm firewall rules allow necessary ports
- [ ] Test backup systems
- [ ] Reserve IP addresses if needed

**Security Preparation**:
- [ ] Generate JWT_SECRET: `openssl rand -base64 32`
- [ ] Generate REGISTRATION_SECRET: `openssl rand -base64 32`
- [ ] Generate BACKUP_KEY: `openssl rand -base64 32`
- [ ] Store secrets securely (use secret management system)
- [ ] Obtain TLS certificate from CA
- [ ] Verify TLS certificate chain
- [ ] Test TLS configuration locally

**Database Preparation**:
- [ ] Create PostgreSQL database (pganalytics)
- [ ] Create TimescaleDB database (metrics)
- [ ] Verify database access and connectivity
- [ ] Test database migrations
- [ ] Create database backup
- [ ] Test backup restore

**Monitoring Preparation**:
- [ ] Set up monitoring infrastructure
- [ ] Configure dashboards
- [ ] Set up alerting rules
- [ ] Test alerts
- [ ] Prepare on-call documentation
- [ ] Brief on-call team

### Day Before Deployment

**Final Verification**:
- [ ] Run `make test-backend` - all tests pass
- [ ] Run `make test-integration` - all tests pass
- [ ] Verify all configuration values ready
- [ ] Test TLS setup locally
- [ ] Test database migrations on copy of production DB
- [ ] Verify backup systems working
- [ ] Confirm team availability

**Documentation**:
- [ ] Print deployment runbook
- [ ] Print rollback procedures
- [ ] Print monitoring dashboard URLs
- [ ] Prepare team communication template
- [ ] Review incident response procedures

---

## Phase 1: Pre-Deployment (3-4 hours before deployment window)

### 1.1 Final Environment Check

```bash
# Verify system resources
free -h                    # Check memory
df -h                      # Check disk space
nproc                      # Check CPU cores

# Verify network connectivity
ping <database_host>
ping <monitoring_host>

# Verify ports are available
netstat -tlnp | grep -E ':(8080|3000|5432|5433)'

# Verify git status
git status
git log --oneline -5
```

### 1.2 Database Backup

```bash
# Backup PostgreSQL (metadata database)
pg_dump -h <db_host> -U postgres pganalytics > pganalytics_backup_$(date +%Y%m%d_%H%M%S).sql

# Backup TimescaleDB (metrics database)
pg_dump -h <ts_host> -U postgres metrics > metrics_backup_$(date +%Y%m%d_%H%M%S).sql

# Verify backups
ls -lah *_backup_*.sql
```

### 1.3 Prepare Deployment Package

```bash
# Clone repository (if fresh deployment)
git clone https://github.com/torresglauco/pganalytics-v3.git /opt/pganalytics

# Or update existing
cd /opt/pganalytics
git pull origin main
git log --oneline -1

# Verify audit approval exists
test -f PRODUCTION_APPROVAL.md && echo "✅ Approval document found"
```

### 1.4 Prepare Environment Variables

Create `/opt/pganalytics/.env.production`:

```bash
# Security - MUST CHANGE FROM DEFAULTS
JWT_SECRET="<32-byte-random-string>"
REGISTRATION_SECRET="<32-byte-random-string>"
BACKUP_KEY="<32-byte-random-string>"

# TLS Configuration
TLS_ENABLED="true"
TLS_CERT_PATH="/etc/pganalytics/cert.pem"
TLS_KEY_PATH="/etc/pganalytics/key.pem"

# Environment
ENVIRONMENT="production"
LOG_LEVEL="info"
PORT="8080"

# Database URLs
DATABASE_URL="postgres://user:password@db.example.com:5432/pganalytics?sslmode=require"
TIMESCALE_URL="postgres://user:password@ts.example.com:5432/metrics?sslmode=require"

# CORS Configuration
CORS_ALLOWED_ORIGINS="https://monitoring.example.com,https://dashboards.example.com"

# Optional tuning
COLLECTION_INTERVAL="120"          # 120 seconds instead of 60 for stability
MAX_CONCURRENT_COLLECTORS="50"     # Hard limit
JWT_EXPIRATION="86400"             # 24 hours
```

**IMPORTANT**: Never commit `.env.production` to git. Use secret management system.

### 1.5 Pre-flight Checks

```bash
# Test configuration loading
cd /opt/pganalytics
export $(cat .env.production | xargs)

# Test database migrations
go run main.go migrate up 2>&1 | head -20

# Verify backend builds
make build

# Check binary size
ls -lh pganalytics-api
```

---

## Phase 2: Deployment Window (3-4 hours)

### 2.1 Notification

Send team notification:

```
DEPLOYMENT STARTING: pgAnalytics v3.2.0 Production Deployment
Time: [Deployment time]
Duration: ~3-4 hours
Impact: New metrics monitoring system being deployed
Status: Will communicate progress every 30 minutes
Contact: [On-call contact]
```

### 2.2 Start Backend Service

```bash
# Create systemd service file /etc/systemd/system/pganalytics.service
[Unit]
Description=pgAnalytics Backend API
After=network.target postgresql.service

[Service]
Type=simple
User=pganalytics
WorkingDirectory=/opt/pganalytics
EnvironmentFile=/opt/pganalytics/.env.production
ExecStart=/opt/pganalytics/pganalytics-api
Restart=on-failure
RestartSec=10s
StandardOutput=append:/var/log/pganalytics/api.log
StandardError=append:/var/log/pganalytics/api.log

[Install]
WantedBy=multi-user.target

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable pganalytics.service
sudo systemctl start pganalytics.service
sudo systemctl status pganalytics.service

# Check logs
sudo tail -f /var/log/pganalytics/api.log
```

### 2.3 Verify Backend Health

Wait 30 seconds for service to start, then:

```bash
# Check service is running
sudo systemctl is-active pganalytics.service

# Test health endpoint
curl -k https://localhost:8080/api/v1/health

# Expected response:
# {
#   "status": "healthy",
#   "version": "3.2.0",
#   "timestamp": "2026-02-26T12:00:00Z"
# }

# Check logs for errors
sudo journalctl -u pganalytics.service -n 50 | grep -i error
```

### 2.4 Deploy Collectors (Gradually)

Start with 1-2 collectors, gradually increase to final count:

```bash
# Deploy first collector
export COLLECTOR_ID="col_prod_001"
export BACKEND_URL="https://api.example.com:8080"
export POSTGRES_HOST="db.example.com"
export POSTGRES_PORT="5432"
export POSTGRES_DATABASES="postgres,production_db"

docker-compose -f collector-docker-compose.yml up -d

# Wait 5 minutes and verify metrics are flowing
curl https://api.example.com:8080/api/v1/metrics/count

# Check logs
docker logs pganalytics-collector | tail -20

# If successful, deploy additional collectors
# Deploy 2-3 more collectors before full deployment
```

### 2.5 Verify Metrics Flow

```bash
# Check metrics are being ingested
curl -k https://localhost:8080/api/v1/metrics \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.count'

# Check database has records
psql $TIMESCALE_URL -c "SELECT COUNT(*) FROM metrics LIMIT 1;"

# Check Grafana dashboards
# Open https://grafana.example.com
# Verify data appears in dashboards
```

### 2.6 Security Verification

```bash
# Test authentication is working
curl -k https://localhost:8080/api/v1/metrics/push \
  -d '{}' -H "Content-Type: application/json"
# Should return 401 Unauthorized

# Test with valid token
TOKEN=$(curl -k https://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"..."}'  | jq -r '.token')

curl -k https://localhost:8080/api/v1/metrics/push \
  -d '{"collector_id":"col_001","metrics":[]}' \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
# Should return 200 OK

# Test rate limiting
for i in {1..150}; do
  curl -k https://localhost:8080/api/v1/health -s > /dev/null
done
# After 100 requests, should see 429 responses
```

### 2.7 Monitoring Verification

```bash
# Verify all monitoring is active
# Check dashboards showing data
# Check alerts are triggered (at least one test alert)
# Verify log aggregation working

# Test alert delivery
# Trigger test alert and verify notification received
```

---

## Phase 3: Stabilization (2-4 hours post-deployment)

### 3.1 Monitor System Health

Every 30 minutes for first 2 hours, then hourly:

```bash
# Check service is still running
sudo systemctl status pganalytics.service

# Check no error spikes in logs
sudo journalctl -u pganalytics.service --since "30 min ago" | grep -i error | wc -l

# Check metrics are flowing
curl -k https://localhost:8080/api/v1/health

# Check system resources
free -h
df -h
top -bn1 | head -20
```

### 3.2 Baseline Performance

Establish baseline metrics for alerting:

```bash
# Record current metrics
date > /var/log/pganalytics/baseline_$(date +%Y%m%d_%H%M%S).txt

# Average latency (should be <300ms at 50 collectors)
curl -k https://localhost:8080/api/v1/metrics -w "\nLatency: %{time_total}s\n"

# Database size
psql $DATABASE_URL -c "SELECT pg_size_pretty(pg_database_size(current_database()));"

# Collector count
curl -k https://localhost:8080/api/v1/collectors -H "Authorization: Bearer $TOKEN" | jq '.count'
```

### 3.3 Team Notification

After 2 hours of successful operation:

```
DEPLOYMENT SUCCESSFUL: pgAnalytics v3.2.0 is now in production
Status: ✅ All systems operational
Metrics: Baseline established, monitoring active
Next: Transition to normal operations
```

---

## Phase 4: Post-Deployment (Days 1-7)

### Daily Monitoring (Day 1)

```bash
# Daily health check script
#!/bin/bash
echo "=== pgAnalytics Daily Health Check ==="
echo "Date: $(date)"

# Service status
echo -n "Service: "
sudo systemctl is-active pganalytics.service || echo "ERROR: Service not running"

# Metrics flow
echo -n "Metrics ingestion: "
METRIC_COUNT=$(curl -k https://localhost:8080/api/v1/metrics -H "Authorization: Bearer $TOKEN" 2>/dev/null | jq '.count' 2>/dev/null || echo "0")
echo "$METRIC_COUNT records"

# System resources
echo "System Resources:"
free -h | grep Mem
df -h | grep pganalytics

# Error count in logs
echo "Errors in past 24h:"
sudo journalctl -u pganalytics.service --since "24 hours ago" | grep -i error | wc -l

# Alert status
echo "Active alerts: [Check monitoring system]"
```

### Weekly Review (Day 7)

- [ ] Review all metrics and trends
- [ ] Check for any recurring errors
- [ ] Verify backups are working
- [ ] Review security logs
- [ ] Plan for optimization improvements
- [ ] Update documentation with findings

### Monthly Review (Day 30)

- [ ] Full audit review meeting
- [ ] Performance analysis
- [ ] Security posture review
- [ ] Capacity planning
- [ ] Optimization recommendations
- [ ] Schedule next improvements

---

## Rollback Procedure

If critical issues occur during deployment:

### Immediate Rollback (First 2 Hours)

```bash
# Stop current service
sudo systemctl stop pganalytics.service

# Restore database from backup
psql -h localhost -U postgres < pganalytics_backup_TIMESTAMP.sql
psql -h localhost -U postgres < metrics_backup_TIMESTAMP.sql

# Start previous version (if available)
# Or redeploy from known-good commit

# Notify stakeholders
```

### Verification After Rollback

```bash
# Verify databases restored
psql $DATABASE_URL -c "SELECT COUNT(*) FROM collectors;"

# Verify service running
sudo systemctl status pganalytics.service

# Test basic functionality
curl -k https://localhost:8080/api/v1/health
```

---

## Operational Runbook (Post-Deployment)

### Daily Operations

**Morning Check (8:00 AM)**:
- Verify service is running
- Check for error spikes overnight
- Verify backups completed
- Check disk space

**Mid-day Review (12:00 PM)**:
- Monitor performance
- Check alert status
- Verify collectors are healthy
- Review new metrics

**Evening Review (5:00 PM)**:
- Performance summary
- Any issues to escalate
- Plan next day items

**Before Shutdown (6:00 PM)**:
- Verify no critical tasks running
- Check backups are scheduled
- Confirm on-call coverage

### Weekly Maintenance

**Monday**:
- Full system health check
- Review metric trends
- Check database growth rate
- Plan scaling if needed

**Friday**:
- Backup integrity verification
- Security log review
- Capacity planning
- Documentation updates

### Monthly Operations

**First Monday**:
- Performance metrics review
- Resource utilization analysis
- Identify optimization opportunities
- Plan improvements for next month

**Last Friday**:
- Disaster recovery drill
- Documentation update
- Team training
- Schedule next review

---

## Configuration Tuning (After Deployment)

### If CPU Usage High (>40%)

```bash
# 1. Check collector count
curl -k https://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer $TOKEN" | jq '.count'

# 2. If >30 collectors, increase COLLECTION_INTERVAL
# Edit .env.production:
# COLLECTION_INTERVAL="120"  # Increase from 60
sudo systemctl restart pganalytics.service

# 3. Consider binary protocol (future version)
```

### If Memory Usage High (>200MB)

```bash
# 1. Check for memory leaks
# Monitor over 1 hour - if steadily increasing, issue exists

# 2. Check query count per collector
# Review LOAD_TEST_REPORT_FEB_2026.md recommendations

# 3. Increase METRICS_BUFFER_SIZE if approaching limit
# Edit .env.production:
# METRICS_BUFFER_SIZE="104857600"  # 100MB
sudo systemctl restart pganalytics.service
```

### If Latency High (>500ms)

```bash
# 1. Check database performance
psql $DATABASE_URL -c "SELECT pid, state, query FROM pg_stat_activity WHERE state='active';"

# 2. Check connection pool
# Look for exhausted connections

# 3. Consider connection pooling (PgBouncer)
# Future improvement
```

---

## Success Criteria

Deployment is successful when all of the following are true:

- [ ] Service started without errors
- [ ] Health check endpoint returns 200 OK
- [ ] At least one collector is pushing metrics
- [ ] Metrics appear in Grafana dashboards
- [ ] Authentication is working (401 for unauthorized requests)
- [ ] Rate limiting is active (429 after limit exceeded)
- [ ] All security headers are present
- [ ] No error spikes in logs
- [ ] CPU usage <40%
- [ ] Memory usage <200MB
- [ ] Disk space sufficient for 30 days of data
- [ ] Backups are working
- [ ] Monitoring and alerting operational
- [ ] Team trained and confident with system
- [ ] Documentation complete and up to date

---

## Support & Escalation

**During Deployment**:
- Primary contact: Glauco Torres
- Technical support: pgAnalytics team
- Database support: DBA team
- Network support: Infrastructure team

**Emergency Contacts**:
- Production database down: DBA on-call
- Service crash: System administrator
- Security incident: Security team
- Monitoring down: Monitoring team

**Escalation Path**:
- Operational issue → Team lead
- Critical issue → Manager
- Security issue → Security officer
- Architectural change → Glauco Torres

---

## Approved For Deployment

**Status**: ✅ Approved
**Approver**: Glauco Torres
**Date**: February 26, 2026
**Authority**: Project Owner

**Approval Document**: PRODUCTION_APPROVAL.md

This deployment guide implements the approval granted in PRODUCTION_APPROVAL.md and must be followed for production deployment of pgAnalytics v3.2.0.

---

*Last Updated: February 26, 2026*
*Next Review: Post-deployment (within 7 days)*
