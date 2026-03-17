# Phase 1 - Daily Execution Checklist
## Staging Deployment - Week of March 11, 2026

**Overall Status**: 🟢 READY TO START
**Phase 1 Duration**: 5 working days (Mon-Fri)
**Expected Completion**: March 15, 2026

---

## 📋 MONDAY - March 11 (Day 1)

### Morning (9:00 AM - 12:00 PM)
**Objective**: Environment Preparation

- [ ] **Team Kickoff Meeting** (30 min)
  - Review phase objectives
  - Assign roles and responsibilities
  - Discuss risks and mitigation
  
- [ ] **Infrastructure Inventory** (1 hour)
  - [ ] Document all server IPs
  - [ ] Verify DNS setup
  - [ ] Confirm network connectivity
  - [ ] Test SSH access to all servers
  
  **Checklist:**
  ```
  Staging API Server:    ___________________
  Database Server:       ___________________
  Grafana Server:        ___________________
  Collector Test Server: ___________________
  ```

- [ ] **Firewall/Network Validation** (30 min)
  - [ ] Confirm required ports are open
  - [ ] Test connectivity matrix
  - [ ] Verify time sync (NTP)

### Afternoon (1:00 PM - 5:00 PM)
**Objective**: Configuration Preparation

- [ ] **Configuration File Setup** (1 hour)
  - [ ] Copy `DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md`
  - [ ] Create `.env.staging`
  - [ ] Fill in all infrastructure values
  - [ ] Secure file permissions (chmod 600)
  - [ ] Generate secrets (JWT, Registration, Backup)
  
  **Secrets Generated:**
  ```
  JWT_SECRET:           ___________________
  REGISTRATION_SECRET:  ___________________
  BACKUP_KEY:           ___________________
  DB_PASSWORD:          ___________________
  ```

- [ ] **Documentation Review** (1 hour)
  - [ ] Read DEPLOYMENT_START_HERE.md
  - [ ] Read PHASE1_STAGING_DEPLOYMENT_PLAN.md
  - [ ] Review runbook
  - [ ] Identify any missing information

- [ ] **Connectivity Testing** (1 hour)
  - [ ] Test SSH to all servers
  - [ ] Ping DNS names
  - [ ] Verify port accessibility
  - [ ] Test database connectivity (if possible)

**End of Day:**
- [ ] Configuration file ready and tested
- [ ] All team members briefed
- [ ] Infrastructure validated
- [ ] Status: 🟢 Day 1 Complete

---

## 📋 TUESDAY - March 12 (Day 2)

### Morning (9:00 AM - 12:00 PM)
**Objective**: Database Setup

- [ ] **SSH to Database Server** (30 min)
  ```bash
  ssh user@staging-db.example.com
  sudo -i
  ```

- [ ] **Create PostgreSQL Role** (30 min)
  - [ ] Connect as postgres user
  - [ ] Create pganalytics role
  - [ ] Set password
  - [ ] Grant pg_monitor role
  - [ ] Verify with `\du pganalytics`

- [ ] **Create Databases** (30 min)
  - [ ] Create pganalytics_staging (with TimescaleDB)
  - [ ] Create pganalytics_staging_main
  - [ ] Verify databases created (`\l`)
  - [ ] Test connection from API server

- [ ] **Database Connectivity Test** (30 min)
  ```bash
  # From staging-api-1:
  psql -h staging-db -U pganalytics -d postgres -c "SELECT version();"
  # Should return PostgreSQL version
  ```

### Afternoon (1:00 PM - 5:00 PM)
**Objective**: PostgreSQL Configuration

- [ ] **Configure PostgreSQL Settings** (1.5 hours)
  - [ ] Edit postgresql.conf
  - [ ] Add monitoring extensions
  - [ ] Configure performance settings
  - [ ] Set logging parameters
  - [ ] Restart PostgreSQL service
  - [ ] Verify restart successful

  **Settings Verified:**
  - [ ] pg_stat_statements enabled
  - [ ] shared_preload_libraries set
  - [ ] Logging enabled
  - [ ] Performance tuned

- [ ] **Configure Backups** (1.5 hours)
  - [ ] Create backup directory
  - [ ] Create backup script
  - [ ] Test backup execution
  - [ ] Verify backup file created
  - [ ] Add to crontab (daily 2 AM)
  
  **Backup Test:**
  ```bash
  ls -lh /var/backups/pganalytics/
  # Should show backup file with current timestamp
  ```

- [ ] **Security Hardening** (1 hour)
  - [ ] Configure SSL/TLS for PostgreSQL
  - [ ] Set pg_hba.conf for authentication
  - [ ] Test remote connection with SSL
  - [ ] Verify password authentication working

**End of Day:**
- [ ] Database fully configured
- [ ] Backups running
- [ ] PostgreSQL restarted successfully
- [ ] Status: 🟢 Day 2 Complete

**Sign-off Required**: Database Admin

---

## 📋 WEDNESDAY - March 13 (Day 3)

### Morning (9:00 AM - 12:00 PM)
**Objective**: Deploy with Docker Compose

- [ ] **Prepare Docker Environment** (1 hour)
  - [ ] SSH to staging-api-1
  - [ ] Clone repository
  - [ ] Copy docker-compose.production.yml
  - [ ] Create .env.staging file
  - [ ] Verify Docker installed and running

  ```bash
  docker --version  # Should be 20.10+
  docker-compose --version  # Should be 2.0+
  ```

- [ ] **Start Services** (1.5 hours)
  - [ ] Run docker-compose up -d
  - [ ] Wait for services to initialize (5-10 min)
  - [ ] Check service status: `docker-compose ps`
  - [ ] Monitor logs: `docker-compose logs -f`
  
  **Services Expected:**
  - [ ] PostgreSQL container running
  - [ ] API container running
  - [ ] Grafana container running
  - [ ] Prometheus container running

- [ ] **Health Check** (30 min)
  ```bash
  curl -k https://staging-api-1.example.com:8080/api/v1/health
  # Should return: {"status":"healthy","version":"3.3.0"}
  ```

### Afternoon (1:00 PM - 5:00 PM)
**Objective**: Collector Deployment

- [ ] **Prepare Collector Server** (1 hour)
  - [ ] SSH to staging-col-1
  - [ ] Copy pganalytics binary
  - [ ] Create /etc/pganalytics/collector.toml
  - [ ] Generate TLS certificates

- [ ] **Configure Collector** (1 hour)
  - [ ] Set collector ID (staging-col-1)
  - [ ] Configure PostgreSQL connection
  - [ ] Configure API endpoint
  - [ ] Set collection interval (60 seconds)

- [ ] **Start Collector** (30 min)
  - [ ] Start collector service
  - [ ] Monitor logs
  - [ ] Verify connection to API
  - [ ] Confirm metrics being sent

- [ ] **Register Collector** (1 hour)
  ```bash
  curl -k -X POST \
    https://staging-api-1:8080/api/v1/collectors/register \
    -H "X-Registration-Secret: $REGISTRATION_SECRET" \
    -d '{"id":"staging-col-1","hostname":"staging-col-1"}'
  ```
  - [ ] Collector registered in API
  - [ ] Collector appears in /api/v1/collectors list

**End of Day:**
- [ ] All services deployed
- [ ] API healthy
- [ ] Collector running
- [ ] Metrics flowing
- [ ] Status: 🟢 Day 3 Complete

**Sign-off Required**: Operations Lead

---

## 📋 THURSDAY - March 14 (Day 4)

### Morning (9:00 AM - 12:00 PM)
**Objective**: Testing & Validation

- [ ] **Smoke Tests** (1 hour)
  - [ ] Health check passing
  - [ ] List collectors endpoint
  - [ ] Metrics endpoint
  - [ ] Grafana health check
  
  **Test Results:**
  ```
  Health Check:    ✅ / ❌
  Collectors:      ✅ / ❌
  Metrics:         ✅ / ❌
  Grafana:         ✅ / ❌
  ```

- [ ] **Security Validation** (1.5 hours)
  - [ ] SSL certificate valid
  - [ ] JWT token generation works
  - [ ] Protected endpoints require auth
  - [ ] RBAC working correctly
  
  **Security Tests:**
  ```
  SSL Certificate:  ✅ / ❌
  JWT Auth:         ✅ / ❌
  RBAC:             ✅ / ❌
  ```

- [ ] **Performance Baseline** (30 min)
  ```bash
  ab -n 10 -c 2 https://staging-api-1:8080/api/v1/health
  # Record response times
  # Average Response Time: _____ ms
  # Requests/sec: _____
  ```

### Afternoon (1:00 PM - 5:00 PM)
**Objective**: Monitoring Setup

- [ ] **Configure Grafana** (2 hours)
  - [ ] Access Grafana (https://staging-mon:3000)
  - [ ] Change default password
  - [ ] Add Prometheus datasource
  - [ ] Import pgAnalytics dashboards
  - [ ] Verify metrics displayed
  
  **Dashboards Imported:**
  - [ ] System Overview
  - [ ] PostgreSQL Metrics
  - [ ] API Performance
  - [ ] Collector Health

- [ ] **Setup Alerting** (1.5 hours)
  - [ ] Create alert rule: "API Down"
  - [ ] Create alert rule: "High Error Rate"
  - [ ] Create alert rule: "DB Connection Pool Low"
  - [ ] Configure notification channels
  - [ ] Test alert firing
  
  **Alerts Created:**
  - [ ] API Health Alert
  - [ ] Error Rate Alert
  - [ ] Database Alert

- [ ] **Log Aggregation** (30 min)
  - [ ] Verify logs being collected
  - [ ] Check log rotation configured
  - [ ] Test log search functionality

**End of Day:**
- [ ] Monitoring fully configured
- [ ] Dashboards displaying metrics
- [ ] Alerts configured and tested
- [ ] Status: 🟢 Day 4 Complete

**Sign-off Required**: Monitoring/Security Lead

---

## 📋 FRIDAY - March 15 (Day 5)

### Morning (9:00 AM - 12:00 PM)
**Objective**: 24-Hour Monitoring & Validation

- [ ] **Review 24-Hour Metrics** (1 hour)
  - [ ] API availability: _____ %
  - [ ] Average response time: _____ ms
  - [ ] Error rate: _____ %
  - [ ] Collector data flowing: ✅ / ❌
  - [ ] No critical alerts: ✅ / ❌

- [ ] **Backup Verification** (30 min)
  - [ ] Daily backup completed: ✅ / ❌
  - [ ] Backup size: _____
  - [ ] Test restore procedure
  - [ ] Verify data integrity

- [ ] **Documentation Update** (30 min)
  - [ ] Record actual deployment steps
  - [ ] Document any deviations
  - [ ] Update runbook
  - [ ] Note any manual steps

### Afternoon (1:00 PM - 5:00 PM)
**Objective**: Sign-off & Approval

- [ ] **Final Review Meeting** (1 hour)
  - [ ] Present 24-hour monitoring report
  - [ ] Review any issues encountered
  - [ ] Discuss production readiness
  - [ ] Get team feedback

- [ ] **Sign-off Review** (1.5 hours)
  
  **Operations Lead:**
  - [ ] Infrastructure stable
  - [ ] Deployment procedures validated
  - [ ] Rollback tested
  - [ ] **Sign-off**: _____ Date: _____

  **Security Officer:**
  - [ ] Security controls verified
  - [ ] No unauthorized access attempts
  - [ ] SSL/TLS working correctly
  - [ ] **Sign-off**: _____ Date: _____

  **Database Admin:**
  - [ ] Database performing well
  - [ ] Backups running successfully
  - [ ] No connection pool issues
  - [ ] **Sign-off**: _____ Date: _____

  **Architecture Lead:**
  - [ ] System design validated
  - [ ] Performance meets expectations
  - [ ] Monitoring adequate
  - [ ] **Sign-off**: _____ Date: _____

  **Project Manager:**
  - [ ] Phase 1 complete
  - [ ] No critical issues
  - [ ] Ready for Phase 2
  - [ ] **Sign-off**: _____ Date: _____

- [ ] **Phase 2 Planning** (1 hour)
  - [ ] Schedule production deployment
  - [ ] Assign Phase 2 tasks
  - [ ] Brief all stakeholders
  - [ ] Confirm go-live date

**End of Day:**
- [ ] All sign-offs obtained
- [ ] Phase 1 complete
- [ ] Phase 2 scheduled
- [ ] Status: 🟢 PHASE 1 COMPLETE

---

## Summary Metrics

### Deployment Timeline
- **Start Date**: March 11, 2026
- **Completion Date**: March 15, 2026
- **Total Duration**: 5 days
- **Status**: 🟢 On Track

### Availability
- **Target Uptime**: 100%
- **Actual Uptime**: _____%
- **Issues Encountered**: ______

### Performance
- **API Response Time (p95)**: _____ ms (target: <500ms)
- **Error Rate**: ____% (target: <0.1%)
- **Successful Deployments**: ___/5

### Sign-offs
- [ ] Operations Lead
- [ ] Security Officer
- [ ] Database Admin
- [ ] Architecture Lead
- [ ] Project Manager

---

## Issues & Notes

| Issue | Severity | Status | Resolution |
|-------|----------|--------|-----------|
| | | | |
| | | | |

---

## Next Phase: Production Deployment

**When**: [Schedule]
**Duration**: 2-3 days
**Objective**: Deploy to production environment
**Success Criteria**: Same as staging, with production monitoring

---

**Phase 1 Status**: 🟢 COMPLETE

Generated: 11 de Março de 2026
Ready to deploy: Now

