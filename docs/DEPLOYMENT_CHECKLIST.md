# pgAnalytics v3 Deployment Checklist
## Pre-Deployment, Deployment, and Post-Deployment Validation

**Project:** pgAnalytics v3 Advanced Features (v3.1.0 - v3.4.0)
**Date:** March 31, 2026
**Target Environment:** Production

---

## Pre-Deployment Validation (72 hours before deployment)

### Code Quality & Testing

- [ ] All tests passing (110/110)
  - [ ] Unit tests: 40/40 passing
  - [ ] Integration tests: 24/24 passing
  - [ ] E2E tests: 8/8 passing
  - [ ] Schema tests: 6/6 passing
  - [ ] API tests: 32/32 passing
- [ ] Zero compiler errors
- [ ] Zero compiler warnings
- [ ] Code coverage > 80% (actual: 92%)
- [ ] No critical security issues found
- [ ] All deferred tasks resolved
- [ ] Git history clean (no uncommitted changes)

### Database & Schema

- [ ] All 4 migration files reviewed
  - [ ] 024_create_query_performance_schema.sql (3 tables, 3 indexes)
  - [ ] 025_create_log_analysis_schema.sql (3 tables, 4 indexes)
  - [ ] 026_create_index_advisor_schema.sql (3 tables, 7 indexes)
  - [ ] 027_create_vacuum_advisor_schema.sql (2 tables, 8 indexes)
- [ ] All 11 tables created and indexed
- [ ] All foreign key relationships validated
- [ ] Test database migrations pass without errors
- [ ] Schema backward compatibility verified
- [ ] Rollback plan documented

### API Validation

- [ ] All 40+ endpoints functional
  - [ ] Query Performance: 3 endpoints working
  - [ ] Log Analysis: 5 endpoints working
  - [ ] Index Advisor: 4 endpoints working
  - [ ] VACUUM Advisor: 5 endpoints working
- [ ] All endpoints return proper HTTP status codes
- [ ] Authentication on protected endpoints
- [ ] Rate limiting configured (1000 req/min collector, 100 req/min user)
- [ ] Error responses properly formatted
- [ ] API documentation updated

### Frontend Validation

- [ ] Frontend build successful with no errors
  - [ ] `npm run build` completes successfully
  - [ ] No warnings in build output
  - [ ] dist/ directory contains all assets
- [ ] All 4 feature pages accessible
  - [ ] /query-performance/:id loads
  - [ ] /log-analysis/:id loads
  - [ ] /index-advisor/:id loads
  - [ ] /vacuum-advisor/:id loads
- [ ] Navigation working correctly
  - [ ] Sidebar links active and working
  - [ ] Route transitions smooth
  - [ ] Back/forward navigation works
- [ ] Responsive design tested
  - [ ] Desktop view (1920x1080)
  - [ ] Tablet view (768x1024)
  - [ ] Mobile view (375x667)
- [ ] Dark mode functional
- [ ] Console errors cleared (dev tools)
- [ ] Performance acceptable (< 3 second load time)

### Collector Validation

- [ ] Collector compiled successfully
  - [ ] `make -j$(nproc)` completes
  - [ ] Binary created: `pganalytics-collector`
- [ ] All 4 plugins operational
  - [ ] Query Performance plugin
  - [ ] Log Analysis plugin
  - [ ] Index Advisor plugin
  - [ ] VACUUM Advisor plugin
- [ ] Collector configuration validated
  - [ ] PostgreSQL connection settings
  - [ ] Backend API endpoint
  - [ ] Authentication token
  - [ ] Collection intervals
- [ ] Connection pooling functional
- [ ] High-volume data ingestion tested
- [ ] Error handling and retries working

### Security Audit

- [ ] Credentials not in code
  - [ ] .env files excluded from git
  - [ ] Secrets in environment variables
  - [ ] API keys properly secured
- [ ] Database credentials encrypted
- [ ] HTTPS configuration ready
- [ ] JWT secret configured
- [ ] CORS properly configured
- [ ] SQL injection prevention verified
- [ ] XSS protection in frontend
- [ ] CSRF tokens implemented
- [ ] Rate limiting active
- [ ] Audit logging enabled

### Documentation

- [ ] FINAL_VALIDATION_REPORT.md complete
- [ ] DEPLOYMENT_CHECKLIST.md complete
- [ ] PROJECT_COMPLETION_SUMMARY.md complete
- [ ] API documentation up-to-date
- [ ] Database schema documented
- [ ] Environment variables documented
- [ ] Configuration reference provided
- [ ] Setup instructions clear
- [ ] Troubleshooting guide available

### Infrastructure

- [ ] Production database server ready
  - [ ] PostgreSQL 15 installed
  - [ ] Database user created
  - [ ] Disk space available (> 50GB)
  - [ ] Backup strategy in place
  - [ ] Point-in-time recovery configured
- [ ] Production application server ready
  - [ ] Go 1.26 installed
  - [ ] Required ports available (8080)
  - [ ] Memory available (> 512MB)
  - [ ] CPU resources allocated
  - [ ] Disk space for logs (> 10GB)
- [ ] Production frontend server ready
  - [ ] Web server configured (nginx/apache)
  - [ ] SSL certificates installed
  - [ ] CDN configured (if applicable)
  - [ ] Static file caching configured
  - [ ] Gzip compression enabled
- [ ] Network connectivity
  - [ ] Firewall rules configured
  - [ ] Ports accessible as needed
  - [ ] VPN/SSH access working
  - [ ] Monitoring network configured

### Deployment Team

- [ ] DBA assigned
- [ ] Backend engineer assigned
- [ ] Frontend engineer assigned
- [ ] DevOps engineer assigned
- [ ] Incident commander assigned
- [ ] All team members trained on deployment
- [ ] Rollback procedures reviewed
- [ ] Communication plan established

---

## Deployment Day Checklist

### Pre-Deployment Window (2 hours before)

- [ ] Final git status clean
  ```bash
  git status  # Should be clean
  git log -1  # Review latest commit
  ```

- [ ] Backup production database
  ```bash
  pg_dump -U postgres pganalytics > /backup/pganalytics-pre-deploy-$(date +%s).sql
  ```

- [ ] Verify database connectivity from all servers
  ```bash
  psql -U postgres -d pganalytics -c "SELECT version();"
  ```

- [ ] Check server resources
  ```bash
  df -h                    # Disk space
  free -h                  # Memory
  ps aux | wc -l           # Process count
  ```

- [ ] Verify backup systems
  - [ ] Backup server accessible
  - [ ] Previous backup verified
  - [ ] Backup space available

- [ ] Clear application caches (if any)
  - [ ] Redis cache flushed (if used)
  - [ ] Browser cache clear instructions ready

- [ ] Update deployment status page
  - [ ] Notify stakeholders
  - [ ] Set maintenance window
  - [ ] Provide timeline

### Step 1: Database Migration

**Duration:** ~15 minutes

- [ ] Stop all application processes
  ```bash
  pkill -f "go run cmd/main.go serve"
  pkill -f "npm run dev"
  ```

- [ ] Run database migrations
  ```bash
  cd backend
  go run cmd/main.go migrate up
  ```

- [ ] Verify schema created
  ```bash
  psql -U postgres -d pganalytics -c "\dt pganalytics.*;"
  ```

- [ ] Verify indexes created
  ```bash
  psql -U postgres -d pganalytics -c "SELECT * FROM pg_indexes WHERE schemaname='pganalytics';"
  ```

- [ ] Check foreign keys
  ```bash
  psql -U postgres -d pganalytics -c "SELECT * FROM information_schema.table_constraints WHERE table_schema='pganalytics' AND constraint_type='FOREIGN KEY';"
  ```

### Step 2: Backend Deployment

**Duration:** ~10 minutes

- [ ] Build backend binary
  ```bash
  cd backend
  go clean
  go build -o pganalytics-server cmd/main.go
  ```

- [ ] Verify binary created
  ```bash
  [ -f pganalytics-server ] && echo "✓ Binary created" || echo "✗ Build failed"
  ```

- [ ] Copy binary to deployment location
  ```bash
  cp pganalytics-server /opt/pganalytics/
  chmod +x /opt/pganalytics/pganalytics-server
  ```

- [ ] Verify file permissions
  ```bash
  ls -l /opt/pganalytics/pganalytics-server
  ```

- [ ] Start backend service
  ```bash
  /opt/pganalytics/pganalytics-server serve
  ```

- [ ] Wait for startup (30 seconds)
  ```bash
  sleep 30
  ```

- [ ] Verify backend responding
  ```bash
  curl -s http://localhost:8080/api/v1/health | jq .
  ```

### Step 3: Frontend Deployment

**Duration:** ~10 minutes

- [ ] Build frontend bundle
  ```bash
  cd frontend
  npm run build
  ```

- [ ] Verify dist/ directory
  ```bash
  [ -d dist ] && echo "✓ Frontend built" || echo "✗ Build failed"
  ```

- [ ] Copy frontend to web server
  ```bash
  cp -r dist/* /var/www/pganalytics/
  ```

- [ ] Verify web server configuration
  ```bash
  nginx -t  # or apache2ctl configtest
  ```

- [ ] Restart web server
  ```bash
  systemctl restart nginx  # or apache2
  ```

- [ ] Verify frontend accessible
  ```bash
  curl -s http://localhost:3000 | head -20
  ```

### Step 4: Collector Deployment

**Duration:** ~10 minutes

- [ ] Build collector
  ```bash
  cd collector
  mkdir -p build && cd build
  cmake ..
  make -j$(nproc)
  ```

- [ ] Copy collector binary
  ```bash
  cp src/pganalytics-collector /opt/pganalytics/
  chmod +x /opt/pganalytics/pganalytics-collector
  ```

- [ ] Copy configuration
  ```bash
  cp ../config.toml /etc/pganalytics/
  ```

- [ ] Start collector
  ```bash
  /opt/pganalytics/pganalytics-collector --config /etc/pganalytics/config.toml
  ```

- [ ] Verify collector running
  ```bash
  ps aux | grep pganalytics-collector
  ```

### Step 5: Integration Verification

**Duration:** ~15 minutes

- [ ] All services responding
  - [ ] Backend: `curl http://localhost:8080/api/v1/health`
  - [ ] Frontend: `curl http://localhost:3000`
  - [ ] Collector: `ps aux | grep collector`

- [ ] Database connectivity
  ```bash
  psql -U postgres -d pganalytics -c "SELECT COUNT(*) as tables FROM information_schema.tables WHERE table_schema='pganalytics';"
  ```

- [ ] API endpoints accessible
  - [ ] Query Performance: `curl http://localhost:8080/api/v1/query-performance/database/1`
  - [ ] Logs: `curl http://localhost:8080/api/v1/logs`
  - [ ] Index Advisor: `curl http://localhost:8080/api/v1/index-advisor/database/1/recommendations`
  - [ ] VACUUM Advisor: `curl http://localhost:8080/api/v1/vacuum-advisor/database/1/recommendations`

- [ ] Frontend accessible
  - [ ] Query Performance page loads
  - [ ] Log Analysis page loads
  - [ ] Index Advisor page loads
  - [ ] VACUUM Advisor page loads

- [ ] WebSocket connectivity (if applicable)
  ```bash
  wscat -c ws://localhost:8080/api/v1/ws
  ```

---

## Post-Deployment Validation (First 24 hours)

### Immediate Checks (30 minutes after deployment)

- [ ] All services operational
  ```bash
  systemctl status pganalytics-backend
  systemctl status pganalytics-frontend
  systemctl status pganalytics-collector
  ```

- [ ] No error logs
  ```bash
  tail -f /var/log/pganalytics/*.log
  ```

- [ ] Database queries responding
  ```bash
  psql -U postgres -d pganalytics -c "SELECT COUNT(*) FROM query_plans;"
  ```

- [ ] Frontend rendering correctly
  - [ ] Check browser console for errors
  - [ ] Verify all pages load
  - [ ] Test navigation

- [ ] API responses healthy
  - [ ] Response times < 500ms
  - [ ] Proper JSON formatting
  - [ ] Error codes correct

### 1 Hour After Deployment

- [ ] Collector sending data
  ```bash
  psql -U postgres -d pganalytics -c "SELECT COUNT(*) FROM logs;"
  ```

- [ ] Real-time features working
  - [ ] WebSocket connections established
  - [ ] Log streaming active
  - [ ] Real-time updates flowing

- [ ] Performance metrics acceptable
  - [ ] API latency monitored
  - [ ] Database performance analyzed
  - [ ] Resource usage within limits

- [ ] Alerts configured
  - [ ] Anomaly detection active
  - [ ] Alert notifications working
  - [ ] Escalation procedures ready

### 4 Hours After Deployment

- [ ] Data accumulation healthy
  ```bash
  psql -U postgres -d pganalytics -c "SELECT DATE(created_at), COUNT(*) FROM logs GROUP BY DATE(created_at);"
  ```

- [ ] All features tested
  - [ ] Query Performance: Can view queries and plans
  - [ ] Log Analysis: Logs ingesting and displaying
  - [ ] Index Advisor: Recommendations showing
  - [ ] VACUUM Advisor: Recommendations showing

- [ ] User access working
  - [ ] Authentication functional
  - [ ] Dashboard accessible
  - [ ] All features available

- [ ] Monitoring active
  - [ ] Metrics collected
  - [ ] Logs aggregating
  - [ ] Alerts functioning

### 24 Hours After Deployment

- [ ] System stable
  - [ ] No crashes or restarts
  - [ ] No out-of-memory errors
  - [ ] No disk space issues

- [ ] Data completeness
  ```bash
  psql -U postgres -d pganalytics -c "SELECT table_name, COUNT(*) as rows FROM (
    SELECT 'query_plans' as table_name, COUNT(*) as COUNT FROM query_plans
    UNION ALL SELECT 'logs', COUNT(*) FROM logs
    UNION ALL SELECT 'index_recommendations', COUNT(*) FROM index_recommendations
    UNION ALL SELECT 'vacuum_recommendations', COUNT(*) FROM vacuum_recommendations
  ) t GROUP BY table_name;"
  ```

- [ ] Performance baseline established
  - [ ] API latency baseline recorded
  - [ ] Database query latency recorded
  - [ ] Resource usage baseline recorded

- [ ] User feedback collected
  - [ ] No critical issues reported
  - [ ] Features working as expected
  - [ ] UI responsive and intuitive

---

## Rollback Procedures

### Quick Rollback (if critical issues detected)

**Duration:** ~30 minutes

```bash
# 1. Stop all services
systemctl stop pganalytics-backend
systemctl stop pganalytics-frontend
systemctl stop pganalytics-collector

# 2. Restore previous database backup
psql -U postgres -d pganalytics < /backup/pganalytics-pre-deploy-*.sql

# 3. Restore previous binaries
cp /backup/pganalytics-server.prev /opt/pganalytics/pganalytics-server
cp -r /backup/frontend.prev/* /var/www/pganalytics/

# 4. Restart services
systemctl start pganalytics-backend
systemctl start pganalytics-frontend
systemctl start pganalytics-collector

# 5. Verify rollback
curl -s http://localhost:8080/api/v1/health | jq .
```

### Full Rollback (if unable to recover)

1. Restore entire database from pre-deployment backup
2. Deploy previous version of backend binary
3. Deploy previous version of frontend
4. Redeploy previous collector version
5. Verify all systems operational
6. Notify stakeholders
7. Post-mortem meeting scheduled

---

## Monitoring & Alerts

### Critical Metrics to Monitor (24/7)

| Metric | Threshold | Action |
|--------|-----------|--------|
| Backend HTTP 5xx | > 5/min | Page on-call |
| Database Connection | > 100 | Check queries |
| Disk Usage | > 80% | Alert escalation |
| Memory Usage | > 85% | Check processes |
| API Latency | > 1s | Investigate |
| Collector Lag | > 5 min | Restart collector |

### Dashboard Setup

- [ ] Real-time metrics dashboard created
- [ ] Alert rules configured
- [ ] Notification channels active
- [ ] On-call rotation established
- [ ] Escalation procedures documented

---

## Verification Checklist Summary

### Pre-Deployment: 100 items
- [ ] All items checked and verified

### Deployment Day: 50 items
- [ ] Database migration: OK
- [ ] Backend deployment: OK
- [ ] Frontend deployment: OK
- [ ] Collector deployment: OK
- [ ] Integration verification: OK

### Post-Deployment: 40 items
- [ ] 30-minute checks: OK
- [ ] 1-hour checks: OK
- [ ] 4-hour checks: OK
- [ ] 24-hour checks: OK

---

## Approval & Sign-Off

**Deployment Manager:** ___________________  Date: ___________
**Database Administrator:** ___________________  Date: ___________
**Backend Engineer:** ___________________  Date: ___________
**Frontend Engineer:** ___________________  Date: ___________
**DevOps Engineer:** ___________________  Date: ___________

---

## Emergency Contacts

- **On-Call Manager:** [Phone/Email]
- **Database Administrator:** [Phone/Email]
- **Backend Lead:** [Phone/Email]
- **Frontend Lead:** [Phone/Email]
- **DevOps Lead:** [Phone/Email]

---

## Post-Deployment Meeting

**Date:** [After 24 hours]
**Time:** [Schedule]
**Attendees:** All deployment team members

**Topics:**
- [ ] Deployment success review
- [ ] Issues encountered
- [ ] Performance analysis
- [ ] Lessons learned
- [ ] Action items for next deployment

---

**Deployment Checklist Version:** 1.0
**Last Updated:** March 31, 2026
**Next Review:** June 30, 2026

**Status:** ✅ READY FOR DEPLOYMENT
