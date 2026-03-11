# Phase 1: Deploy to Staging - Execution Plan
## pgAnalytics v3.3.0 - Staging Environment Setup

**Date**: 11 de Março de 2026
**Phase**: 1 - Staging Deployment
**Duration**: Week 1 (5 working days)
**Environment**: Staging (Non-Production)
**Status**: 🟢 READY TO START

---

## Executive Summary

This phase focuses on deploying pgAnalytics v3.3.0 to a **staging environment** to validate the complete installation, configuration, and operational procedures before production deployment.

**Objectives:**
- ✅ Deploy all components to staging
- ✅ Run comprehensive smoke tests
- ✅ Validate security configuration
- ✅ Test monitoring and alerting
- ✅ Verify backup/recovery procedures
- ✅ Train operations team
- ✅ Document any issues/learnings

**Timeline**: 1 week
**Go-Live Decision**: End of Week 1

---

## Phase 1 Timeline

```
MONDAY (Day 1)
  └─ Environment Preparation & Prerequisites

TUESDAY (Day 2)
  └─ Database Setup & Configuration

WEDNESDAY (Day 3)
  └─ API Server & Collector Deployment

THURSDAY (Day 4)
  └─ Monitoring & Testing

FRIDAY (Day 5)
  └─ Validation & Sign-off
```

---

## Pre-Requisites Check

### Infrastructure Requirements

Before starting, ensure you have:

- [ ] **Staging Server(s)**
  - [ ] Linux server (Ubuntu 20.04+ or CentOS 8+)
  - [ ] 4+ CPU cores
  - [ ] 16GB+ RAM
  - [ ] 100GB+ disk space
  - [ ] Internet connectivity (for package downloads)

- [ ] **Database Server**
  - [ ] PostgreSQL 16.12+ installed
  - [ ] SSL/TLS enabled
  - [ ] Backup procedure configured
  - [ ] Remote connectivity enabled (if separate from API server)

- [ ] **Network**
  - [ ] SSH access to all servers
  - [ ] Required ports open (443, 8080, 5432, 3000, 9090)
  - [ ] DNS resolution working
  - [ ] Time synchronization (NTP)

### Software Requirements

- [ ] Docker 20.10+ (if using Docker Compose method)
- [ ] Docker Compose 2.0+ (if using Docker Compose method)
- [ ] Git installed (for cloning repository)
- [ ] curl/wget installed
- [ ] openssl installed (for certificate generation)

### Access & Credentials

- [ ] Root or sudo access on servers
- [ ] Git repository access (SSH key configured)
- [ ] DNS admin access (if creating staging subdomain)
- [ ] Firewall admin access (to open ports)

---

## MONDAY - Day 1: Environment Preparation

### Task 1.1: Gather Infrastructure Details

**Subtasks:**
- [ ] Document staging server IP addresses
- [ ] Document database server details
- [ ] Identify network/firewall contacts
- [ ] Confirm DNS setup requirements
- [ ] Verify internet connectivity

**Deliverable:** Infrastructure spreadsheet with all IPs and hostnames

```bash
# Example inventory
Staging API Server 1:     staging-api-1.example.com (192.168.1.100)
Database Server:          staging-db.example.com (192.168.1.101)
Grafana Server:           staging-mon.example.com (192.168.1.102)
Collector Test Server:    staging-col-1.example.com (192.168.1.103)
```

### Task 1.2: Prepare Configuration

**Steps:**
1. Copy configuration template
```bash
cp /Users/glauco.torres/git/pganalytics-v3/DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md ~/.env.staging
chmod 600 ~/.env.staging
```

2. Edit with staging values
```bash
nano ~/.env.staging
```

3. Required variables for staging:
```bash
# Environment
export ENVIRONMENT="staging"
export DEPLOYMENT_MODE="docker"  # Using docker-compose for staging

# Database
export DB_HOST="staging-db.example.com"
export DB_PORT="5432"
export DB_NAME="pganalytics_staging"
export DB_USER="pganalytics"
export DB_PASSWORD="$(openssl rand -base64 32)"

# API Server
export API_HOST_1="staging-api-1.example.com"
export API_PORT="8080"
export API_TLS_CERT="/etc/pganalytics/certs/staging.crt"
export API_TLS_KEY="/etc/pganalytics/keys/staging.key"

# Secrets
export JWT_SECRET="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_ENCRYPTION_KEY="$(openssl rand -base64 32)"

# Monitoring
export GRAFANA_HOST="staging-mon.example.com"
export PROMETHEUS_HOST="staging-mon.example.com"
export PROMETHEUS_PORT="9090"
```

**Deliverable:** `.env.staging` file configured

### Task 1.3: Network & Security Validation

**Steps:**
1. Test connectivity to all servers
```bash
# From your machine
for host in staging-api-1 staging-db staging-mon staging-col-1; do
  echo "Testing $host..."
  ping -c 1 $host || echo "❌ Cannot reach $host"
  ssh -o ConnectTimeout=5 user@$host "echo OK" && echo "✅ SSH OK" || echo "❌ SSH Failed"
done
```

2. Verify ports are accessible
```bash
# From staging-api-1
netstat -tuln | grep LISTEN
# Should show ports: 22 (SSH), 443 (HTTPS), 8080 (API)
```

3. Test database connectivity
```bash
# From staging-api-1
psql -h staging-db -U pganalytics -d postgres -c "SELECT version();"
```

**Deliverable:** Connectivity test report

### Task 1.4: Create Deployment Runbook

**Create file:** `STAGING_DEPLOYMENT_RUNBOOK.md`

Structure:
```markdown
# Staging Deployment Runbook

## Day-by-Day Schedule
## Server Access Details
## Rollback Procedures
## Emergency Contacts
## Health Check Procedures
## Monitoring & Alerting Setup
```

**Deliverable:** Runbook document shared with team

---

## TUESDAY - Day 2: Database Setup

### Task 2.1: Create Staging Database

**SSH to database server:**
```bash
ssh user@staging-db.example.com
sudo -i
```

**Create PostgreSQL role and database:**
```sql
-- Connect as postgres user
sudo -u postgres psql

-- Create role
CREATE ROLE pganalytics WITH LOGIN INHERIT;
ALTER ROLE pganalytics WITH PASSWORD 'your-secure-password-here';
GRANT pg_monitor TO pganalytics;

-- Create TimescaleDB extension database
CREATE DATABASE pganalytics_staging OWNER pganalytics;
\c pganalytics_staging
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create main database
CREATE DATABASE pganalytics_staging_main OWNER pganalytics;
\c pganalytics_staging_main
-- (Schema will be created by migrations)

-- Verify
\l pganalytics_staging
\du pganalytics
```

**Deliverable:** Verified database connection

### Task 2.2: Configure PostgreSQL for pgAnalytics

**Settings to apply:**
```bash
# As root on database server

# Edit postgresql.conf
sudo nano /etc/postgresql/16/main/postgresql.conf

# Add these settings:
# Monitoring
shared_preload_libraries = 'pg_stat_statements'
pg_stat_statements.track = all
pg_stat_statements.max = 10000

# Performance
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
random_page_cost = 1.1

# Replication (if HA planned)
wal_level = replica
max_wal_senders = 3
max_replication_slots = 3

# Logging
log_statement = 'all'
log_duration = on
log_min_duration_statement = 1000

# Restart PostgreSQL
sudo systemctl restart postgresql
```

**Deliverable:** PostgreSQL configured and restarted

### Task 2.3: Enable Backups

**Set up automated backups:**
```bash
# As root on database server

# Create backup directory
sudo mkdir -p /var/backups/pganalytics
sudo chown postgres:postgres /var/backups/pganalytics
sudo chmod 700 /var/backups/pganalytics

# Create backup script
sudo tee /usr/local/bin/pganalytics-backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/var/backups/pganalytics"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/pganalytics_$TIMESTAMP.sql.gz"

/usr/bin/pg_dump -U pganalytics pganalytics_staging | \
  gzip > "$BACKUP_FILE"

# Keep only last 7 days
find $BACKUP_DIR -name "pganalytics_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE"
EOF

sudo chmod +x /usr/local/bin/pganalytics-backup.sh

# Add to crontab (daily at 2 AM)
sudo -u postgres crontab -e
# Add: 0 2 * * * /usr/local/bin/pganalytics-backup.sh
```

**Test backup:**
```bash
sudo -u postgres /usr/local/bin/pganalytics-backup.sh
ls -lh /var/backups/pganalytics/
```

**Deliverable:** Backup procedure tested and running

---

## WEDNESDAY - Day 3: Deploy & Configure Services

### Task 3.1: Deploy Using Docker Compose (Recommended for Staging)

**On staging-api-1 server:**
```bash
# Clone repository
cd /opt
git clone https://github.com/torresglauco/pganalytics-v3.git pganalytics
cd pganalytics

# Copy production compose file
cp docker-compose.production.yml docker-compose.staging.yml

# Create .env file for Docker
source ~/.env.staging
cat > .env.staging << EOF
ENVIRONMENT=staging
DB_HOST=$DB_HOST
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
JWT_SECRET=$JWT_SECRET
REGISTRATION_SECRET=$REGISTRATION_SECRET
API_HOST=$API_HOST_1
API_PORT=$API_PORT
GRAFANA_HOST=$GRAFANA_HOST
EOF

# Start services
docker-compose -f docker-compose.staging.yml --env-file .env.staging up -d

# Verify services
docker-compose -f docker-compose.staging.yml ps
```

**Deliverable:** All services running

### Task 3.2: Verify API Health

**Health check:**
```bash
# From your machine
curl -k https://staging-api-1.example.com:8080/api/v1/health

# Expected response:
# {"status":"healthy","version":"3.3.0","timestamp":"2026-03-11T10:00:00Z"}
```

**Deliverable:** Health check passing

### Task 3.3: Deploy Collector

**On staging-col-1 server:**
```bash
# Copy collector binary
scp pganalytics-api user@staging-col-1:/opt/pganalytics/

# Create collector config
cat > /etc/pganalytics/collector.toml << EOF
[collector]
id = "staging-col-1"
hostname = "staging-col-1"
interval = 60

[backend]
url = "https://staging-api-1.example.com:8080"
tls_cert = "/etc/pganalytics/certs/collector.crt"
tls_key = "/etc/pganalytics/keys/collector.key"
tls_ca = "/etc/pganalytics/certs/ca.crt"

[postgres]
host = "staging-db.example.com"
port = 5432
databases = ["postgres"]
EOF

# Start collector
/opt/pganalytics/pganalytics-api collector start
```

**Deliverable:** Collector running and connecting to API

---

## THURSDAY - Day 4: Testing & Validation

### Task 4.1: Smoke Tests

**Run basic functionality tests:**
```bash
#!/bin/bash

API_BASE="https://staging-api-1.example.com:8080"

echo "=== Health Check ==="
curl -k -s "$API_BASE/api/v1/health" | jq .

echo "=== List Collectors ==="
curl -k -s "$API_BASE/api/v1/collectors" | jq .

echo "=== Check Metrics ==="
curl -k -s "$API_BASE/api/v1/servers/*/metrics" | jq .

echo "=== Grafana Health ==="
curl -s "https://staging-mon.example.com:3000/api/health" | jq .
```

**Deliverable:** Smoke test report

### Task 4.2: Security Validation

**SSL/TLS Certificate Test:**
```bash
openssl s_client -connect staging-api-1.example.com:443 -showcerts
# Verify certificate is valid and not self-signed
```

**JWT Token Test:**
```bash
# Get token
TOKEN=$(curl -k -s -X POST \
  https://staging-api-1.example.com:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "X-Registration-Secret: $REGISTRATION_SECRET" \
  -d '{"username":"staging-user","password":"test-password"}' | jq -r .token)

echo "Token: $TOKEN"

# Test protected endpoint
curl -k -s -H "Authorization: Bearer $TOKEN" \
  https://staging-api-1.example.com:8080/api/v1/servers
```

**Deliverable:** Security test report

### Task 4.3: Performance Baseline

**Run baseline performance test:**
```bash
# Simple load test (10 requests)
ab -n 10 -c 2 https://staging-api-1.example.com:8080/api/v1/health

# Monitor resource usage
ssh user@staging-api-1
top -b -n 1 | head -15
df -h
free -h
```

**Deliverable:** Performance baseline documented

### Task 4.4: Setup Monitoring & Alerting

**Configure Grafana:**
```bash
# Access Grafana
https://staging-mon.example.com:3000

# Login (default: admin/admin)
# Change password immediately

# Add Prometheus datasource
# URL: http://staging-mon:9090

# Import pgAnalytics dashboards
# Go to Dashboards > Import
# Upload dashboards from /grafana/dashboards/
```

**Create basic alert:**
```bash
# In Grafana:
# 1. Create alert rule: "API Down"
# 2. Condition: Health check fails for 5 min
# 3. Action: Send to webhook or email
```

**Deliverable:** Monitoring configured and tested

---

## FRIDAY - Day 5: Validation & Sign-off

### Task 5.1: 24-Hour Monitoring Period

**Run all services for 24 hours and monitor:**

- [ ] API availability (target: 100%)
- [ ] Response time (target: <500ms p95)
- [ ] Error rate (target: <0.1%)
- [ ] Collector data flowing (check metrics)
- [ ] Backups running successfully
- [ ] Logs being collected properly

**Monitoring dashboard:**
```
CREATE GRAFANA PANEL:
- API Response Time (last 24h)
- Request Success Rate (last 24h)
- Collector Health Status
- Database Connections
- Disk Usage Trend
```

**Deliverable:** 24-hour monitoring report

### Task 5.2: Documentation & Runbook Update

**Update documentation:**
- [ ] Record actual deployment time
- [ ] Document any issues encountered
- [ ] Update deployment runbook with actual steps
- [ ] Document environment-specific settings
- [ ] Create troubleshooting guide

**Deliverable:** Updated runbook and issue log

### Task 5.3: Team Sign-off

**Get approvals from:**
- [ ] **Operations Lead**: Infrastructure and deployment
- [ ] **Security Officer**: Security controls verification
- [ ] **Database Admin**: Database configuration review
- [ ] **Architecture Lead**: System design validation
- [ ] **Project Manager**: Timeline and quality sign-off

**Sign-off Template:**
```markdown
## Staging Deployment Sign-off

Date: ___________
Environment: Staging

### Operations
- [ ] Approved by: ________________
- [ ] Notes: ____________________

### Security
- [ ] Approved by: ________________
- [ ] Notes: ____________________

### Database
- [ ] Approved by: ________________
- [ ] Notes: ____________________

### Architecture
- [ ] Approved by: ________________
- [ ] Notes: ____________________

### Project
- [ ] Approved by: ________________
- [ ] Notes: ____________________

## Recommendation for Production

🟢 **APPROVED FOR PRODUCTION** / 🟡 **CONDITIONAL** / 🔴 **BLOCKED**
```

**Deliverable:** All signatures obtained

### Task 5.4: Issues & Remediation Planning

**Document any issues found:**

| Issue | Severity | Resolution | Timeline |
|-------|----------|-----------|----------|
| Example | High | Fix before prod | ASAP |

**Create remediation tasks in JIRA/GitHub Issues**

**Deliverable:** Issue tracking with owners assigned

---

## Success Criteria

- ✅ All services deployed and running
- ✅ Health checks passing 100%
- ✅ Smoke tests passing
- ✅ Security validation complete
- ✅ Monitoring & alerting configured
- ✅ 24-hour stability period completed
- ✅ All team approvals obtained
- ✅ Documentation updated
- ✅ No critical issues outstanding

---

## Rollback Plan (If Needed)

If critical issues occur during staging:

```bash
# Stop all services
docker-compose -f docker-compose.staging.yml down

# Restore from backup
pg_restore -U pganalytics -d pganalytics_staging < /var/backups/pganalytics/latest.sql

# Restart services
docker-compose -f docker-compose.staging.yml up -d
```

---

## Next Steps After Phase 1

If sign-off is approved:
1. ✅ Schedule Phase 2 (Production) deployment
2. ✅ Create production configuration
3. ✅ Brief all operations staff
4. ✅ Prepare rollback procedures
5. ✅ Set up production monitoring

If issues remain:
1. 🟡 Address critical issues in staging
2. 🟡 Re-test after fixes
3. 🟡 Get new sign-offs
4. 🟡 Then proceed to production

---

## Contact & Escalation

**Deployment Lead:** ___________________
**Operations Lead:** ___________________
**Security Lead:** ___________________
**Database Lead:** ___________________
**Escalation Contact:** ___________________

---

**Phase 1 Status: 🟢 READY TO START**

Start Date: Week of March 11, 2026
Expected Completion: March 15, 2026
Go-Live Decision: End of Week

---

Generated: 11 de Março de 2026
Template: pgAnalytics v3.3.0 Staging Deployment Plan

