# Operations Guide: High Availability & Disaster Recovery
**Date**: March 11, 2026
**Version**: v3.3.0+
**Status**: Implementation Ready

---

## 📋 Table of Contents

1. [HA Architecture](#ha-architecture)
2. [Load Balancing Setup](#load-balancing-setup)
3. [Database Replication](#database-replication)
4. [Backup Strategy](#backup-strategy)
5. [Disaster Recovery](#disaster-recovery)
6. [Monitoring & Alerting](#monitoring--alerting)
7. [Failover Testing](#failover-testing)
8. [Runbooks](#runbooks)

---

## 🏗️ HA Architecture

### Overview

High Availability (HA) setup for pgAnalytics v3.3+ consists of:

```
┌─────────────────────────────────────────────────────────────┐
│                    Client/User                              │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              Load Balancer (Virtual IP)                     │
│          (HAProxy, Nginx, or Cloud LB)                      │
└─────────────────────────────────────────────────────────────┘
        │                    │                    │
        ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│  API Server  │    │  API Server  │    │  API Server  │
│   Instance 1 │    │   Instance 2 │    │   Instance 3 │
└──────────────┘    └──────────────┘    └──────────────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             ▼
                ┌────────────────────────┐
                │   Shared PostgreSQL    │
                │    (Primary + Read     │
                │    Replicas)           │
                └────────────────────────┘
                             │
                             ▼
                ┌────────────────────────┐
                │   Backup Storage       │
                │   (S3, Azure Blob,     │
                │    On-prem NAS)        │
                └────────────────────────┘
```

### Components

#### 1. Load Balancer
- **Purpose**: Distribute requests across API servers
- **Options**:
  - HAProxy (on-prem)
  - Nginx (on-prem)
  - AWS ALB (AWS)
  - Google Cloud LB (GCP)
  - Azure Load Balancer (Azure)

#### 2. API Servers
- **Requirement**: Stateless design
- **Scale**: 2-N instances (3+ recommended)
- **Health Check**: /api/v1/health endpoint
- **Graceful Shutdown**: Handle SIGTERM

#### 3. PostgreSQL
- **Primary**: Single write instance
- **Replicas**: Read-only replicas for failover
- **Shared Storage**: For data persistence

#### 4. Backup Storage
- **Type**: Geographically redundant storage
- **Retention**: 30+ days
- **Encryption**: At rest and in transit

---

## 🔄 Load Balancing Setup

### HAProxy Configuration (On-Prem)

**File**: `/etc/haproxy/haproxy.cfg`

```haproxy
global
  maxconn 4096
  log 127.0.0.1 local0

defaults
  log global
  mode http
  option httplog
  option dontlognull
  timeout connect 5000
  timeout client 50000
  timeout server 50000

# HTTP Front-end
frontend http_front
  bind *:80
  mode http
  default_backend api_servers

# API Backend
backend api_servers
  mode http
  balance roundrobin

  # Health check
  option httpchk GET /api/v1/health HTTP/1.1
  http-check expect status 200

  # API servers
  server api1 10.0.1.10:8080 check inter 5000 rise 2 fall 2
  server api2 10.0.1.11:8080 check inter 5000 rise 2 fall 2
  server api3 10.0.1.12:8080 check inter 5000 rise 2 fall 2

  # Sticky sessions (optional)
  cookie SERVERID insert indirect nocache

# HTTPS Frontend
frontend https_front
  bind *:443 ssl crt /etc/ssl/certs/pganalytics.pem
  mode http
  default_backend api_servers

  # Security headers
  http-response set-header Strict-Transport-Security max-age=31536000
```

### Nginx Configuration (On-Prem)

**File**: `/etc/nginx/conf.d/pganalytics.conf`

```nginx
upstream pganalytics_api {
  least_conn;
  server 10.0.1.10:8080 max_fails=2 fail_timeout=10s;
  server 10.0.1.11:8080 max_fails=2 fail_timeout=10s;
  server 10.0.1.12:8080 max_fails=2 fail_timeout=10s;

  # Keepalive connections
  keepalive 32;
}

# HTTP redirect to HTTPS
server {
  listen 80;
  server_name api.pganalytics.local;
  return 301 https://$server_name$request_uri;
}

# HTTPS server
server {
  listen 443 ssl http2;
  server_name api.pganalytics.local;

  # SSL certificates
  ssl_certificate /etc/ssl/certs/pganalytics.crt;
  ssl_certificate_key /etc/ssl/private/pganalytics.key;

  # Security headers
  add_header Strict-Transport-Security "max-age=31536000" always;
  add_header X-Content-Type-Options "nosniff" always;
  add_header X-Frame-Options "DENY" always;

  # Rate limiting
  limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
  limit_req zone=api_limit burst=200;

  location / {
    proxy_pass http://pganalytics_api;
    proxy_http_version 1.1;
    proxy_set_header Connection "";

    # Headers
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # Timeouts
    proxy_connect_timeout 5s;
    proxy_send_timeout 30s;
    proxy_read_timeout 30s;
  }

  # Health check endpoint
  location /health {
    access_log off;
    proxy_pass http://pganalytics_api/api/v1/health;
  }
}
```

### AWS ALB Setup

```bash
# Create Target Group
aws elbv2 create-target-group \
  --name pganalytics-api \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-xxxxx \
  --health-check-protocol HTTP \
  --health-check-path /api/v1/health \
  --health-check-interval-seconds 5 \
  --health-check-timeout-seconds 3 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 2

# Register targets (API servers)
aws elbv2 register-targets \
  --target-group-arn arn:aws:elasticloadbalancing:... \
  --targets Id=i-1234567890 Id=i-0987654321 Id=i-abcdefghij

# Create ALB
aws elbv2 create-load-balancer \
  --name pganalytics-alb \
  --subnets subnet-1 subnet-2 \
  --security-groups sg-xxxxx \
  --scheme internet-facing
```

---

## 💾 Database Replication

### PostgreSQL Streaming Replication

#### Primary Server Setup

```sql
-- Edit postgresql.conf
wal_level = replica
max_wal_senders = 5
max_replication_slots = 5
hot_standby_feedback = on

-- Create replication user
CREATE USER replication WITH REPLICATION ENCRYPTED PASSWORD 'secure_password';

-- pg_hba.conf entry
host    replication     replication     10.0.1.0/24     md5
```

#### Replica Server Setup

```bash
# Stop replica if running
sudo systemctl stop postgresql

# Perform base backup
pg_basebackup -h 10.0.1.5 -D /var/lib/postgresql/data \
  -U replication -W -Pv -R

# Start replica
sudo systemctl start postgresql

# Verify replication
psql -c "SELECT slot_name, active FROM pg_replication_slots;"
```

---

## 🔒 Backup Strategy

### Automated Daily Backups

#### Script: `/opt/pganalytics/backup.sh`

```bash
#!/bin/bash

BACKUP_DIR="/backups/pganalytics"
RETENTION_DAYS=30
DB_HOST="localhost"
DB_USER="postgres"
DB_NAME="pganalytics"

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/pganalytics_$DATE.sql.gz"

# Create backup
pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > $BACKUP_FILE

# Verify backup
if gunzip -t "$BACKUP_FILE" > /dev/null 2>&1; then
  echo "Backup successful: $BACKUP_FILE"

  # Upload to S3
  aws s3 cp "$BACKUP_FILE" "s3://pganalytics-backups/$(date +%Y/%m)/"

  # Clean old backups
  find $BACKUP_DIR -name "pganalytics_*.sql.gz" -mtime +$RETENTION_DAYS -delete
else
  echo "Backup verification failed!"
  exit 1
fi
```

#### Cron Job

```bash
# /etc/cron.d/pganalytics-backup
0 2 * * * root /opt/pganalytics/backup.sh >> /var/log/pganalytics/backup.log 2>&1
```

### Backup Verification

```bash
#!/bin/bash

# Test restore to temporary database
BACKUP_FILE=$1
TEMP_DB="pganalytics_test"

createdb $TEMP_DB
gunzip -c "$BACKUP_FILE" | psql -d $TEMP_DB

# Run tests
psql -d $TEMP_DB -c "SELECT COUNT(*) FROM collectors;"
psql -d $TEMP_DB -c "SELECT COUNT(*) FROM metrics;"

# Cleanup
dropdb $TEMP_DB

echo "Backup verified: $BACKUP_FILE"
```

---

## 🚨 Disaster Recovery

### RTO and RPO

- **RTO (Recovery Time Objective)**: < 1 hour
- **RPO (Recovery Point Objective)**: < 5 minutes

### Recovery Procedures

#### Scenario 1: Single API Server Failure

**Detection**: Health check fails, traffic rerouted

**Action**: Automatic (no manual intervention needed)

```
1. Load balancer detects failed health check
2. Traffic automatically rerouted to healthy servers
3. Failed server removed from pool
4. Alert sent to ops team
5. Replace failed server when ready
```

#### Scenario 2: Database Replica Failure

**Detection**: Replication lag alerts

**Action**:

```bash
# 1. Stop replica
sudo systemctl stop postgresql

# 2. Perform new base backup
pg_basebackup -h 10.0.1.5 -D /var/lib/postgresql/data \
  -U replication -W -Pv -R

# 3. Start replica
sudo systemctl start postgresql

# 4. Verify replication
watch 'psql -c "SELECT slot_name, active FROM pg_replication_slots;"'
```

#### Scenario 3: Primary Database Failure

**Detection**: All connections fail

**Action**: Promote replica to primary

```bash
# 1. Verify replica is up to date
ssh replica_server
psql -c "SELECT pg_last_wal_receive_lsn();"

# 2. Promote replica to primary
sudo -u postgres /usr/lib/postgresql/15/bin/pg_ctl promote \
  -D /var/lib/postgresql/data

# 3. Update connection strings
# - Update API servers to point to new primary
# - Update backup scripts

# 4. Create new replica from new primary
# Follow replica setup procedure
```

#### Scenario 4: Complete Data Loss

**Detection**: Corruption detected

**Action**: Restore from backup

```bash
# 1. Stop API servers
ansible all -m service -a "name=pganalytics state=stopped"

# 2. Stop database
sudo systemctl stop postgresql

# 3. Backup corrupted data
sudo mv /var/lib/postgresql/data /var/lib/postgresql/data.bak

# 4. Create new cluster
sudo -u postgres /usr/lib/postgresql/15/bin/initdb \
  -D /var/lib/postgresql/data

# 5. Restore from backup
gunzip -c /backups/pganalytics_LATEST.sql.gz | \
  psql -U postgres

# 6. Start database
sudo systemctl start postgresql

# 7. Start API servers
ansible all -m service -a "name=pganalytics state=started"

# 8. Verify
curl https://api.pganalytics.local/api/v1/health
```

---

## 📊 Monitoring & Alerting

### Key Metrics to Monitor

```
API Servers:
  - Response time (p50, p95, p99)
  - Error rate
  - Active connections
  - CPU usage
  - Memory usage
  - Disk I/O

Database:
  - Connection count
  - Query duration
  - Replication lag
  - Disk usage
  - WAL archiving status
  - Checkpoint frequency

Load Balancer:
  - Active connections
  - Failed backends
  - Request rate
  - Error rate
```

### Prometheus Scrape Config

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'pganalytics-api'
    static_configs:
      - targets: ['10.0.1.10:9090', '10.0.1.11:9090', '10.0.1.12:9090']

  - job_name: 'postgresql'
    static_configs:
      - targets: ['10.0.1.5:9187']  # postgres_exporter

  - job_name: 'haproxy'
    static_configs:
      - targets: ['10.0.1.20:8404']  # HAProxy stats

  - job_name: 'node'
    static_configs:
      - targets: ['10.0.1.10:9100', '10.0.1.11:9100', '10.0.1.12:9100', '10.0.1.5:9100']
```

### Alert Rules

```yaml
groups:
  - name: pganalytics_alerts
    rules:
      - alert: APIServerDown
        expr: up{job="pganalytics-api"} == 0
        for: 2m
        annotations:
          summary: "API server {{ $labels.instance }} is down"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        annotations:
          summary: "Error rate > 5%"

      - alert: DatabaseReplicationLag
        expr: pg_replication_slot_confirmed_flush_lsn_bytes - pg_wal_lsn_last_removed_bytes > 1000000000
        annotations:
          summary: "Replication lag > 1GB"

      - alert: DiskSpaceWarning
        expr: node_filesystem_avail_bytes / node_filesystem_size_bytes < 0.2
        annotations:
          summary: "Disk usage > 80%"
```

---

## 🧪 Failover Testing

### Monthly Failover Drill

#### Test 1: API Server Failure

```bash
# Procedure
1. Identify one API server to "fail"
2. Stop the service: systemctl stop pganalytics
3. Verify traffic reroutes to other servers
4. Measure failover time (target: <5 seconds)
5. Start service: systemctl start pganalytics
6. Verify recovery
7. Document results

# Success Criteria
- No user impact (requests rerouted)
- Failover time < 5 seconds
- Service recovers cleanly
```

#### Test 2: Database Replica Failure

```bash
# Procedure
1. Stop replica: systemctl stop postgresql
2. Verify primary still operational
3. Wait 5 minutes
4. Start replica: systemctl start postgresql
5. Verify replication catches up
6. Check replication lag

# Success Criteria
- Primary continues accepting writes
- Replica recovers and syncs
- Replication lag goes to 0
```

#### Test 3: Database Failover

```bash
# Procedure (in staging only!)
1. Verify replica is current
2. Promote replica to primary
3. Verify promotion successful
4. Update connection strings
5. Restart API servers
6. Verify full functionality
7. Promote old primary back to replica

# Success Criteria
- Failover completes in < 1 minute
- No data loss
- All systems reconnect properly
```

---

## 📖 Runbooks

### Runbook 1: API Server Down

**Symptom**: Load balancer shows server unhealthy

**Steps**:
```
1. Verify server is actually down:
   curl https://10.0.1.10:8080/api/v1/health

2. Check server status:
   ssh 10.0.1.10
   systemctl status pganalytics

3. View logs:
   journalctl -u pganalytics -n 50

4. Restart if needed:
   systemctl restart pganalytics

5. Verify recovery:
   systemctl status pganalytics
   curl https://10.0.1.10:8080/api/v1/health

6. Check load balancer:
   # Should show server as healthy within 1 minute
```

### Runbook 2: Database Connection Issues

**Symptom**: API servers can't connect to database

**Steps**:
```
1. Verify database is running:
   ssh database_server
   systemctl status postgresql

2. Check database logs:
   tail -f /var/log/postgresql/postgresql.log

3. Verify connectivity:
   psql -h 10.0.1.5 -U postgres -c "SELECT 1"

4. Check replication:
   psql -c "SELECT * FROM pg_stat_replication;"

5. Restart if needed:
   systemctl restart postgresql

6. Verify from API server:
   curl https://api.pganalytics.local/api/v1/health
```

### Runbook 3: Disk Space Critical

**Symptom**: Disk usage > 90%

**Steps**:
```
1. Identify what's using space:
   du -sh /var/lib/postgresql/data/*

2. Archive old WAL files:
   sudo -u postgres pg_archivecleanup /wal_archive '0/00000000'

3. Truncate old logs:
   sudo truncate -s 0 /var/log/postgresql/postgresql.log

4. Clean system logs:
   journalctl --vacuum=200M

5. Monitor usage:
   watch 'df -h | grep -E "/$|/var"'
```

---

## ✅ Checklist

### Daily
- [ ] Monitor alert system
- [ ] Check backup logs
- [ ] Verify replication status

### Weekly
- [ ] Review error rates
- [ ] Check disk usage trends
- [ ] Test a backup restore (in staging)

### Monthly
- [ ] Run full failover drill
- [ ] Review and update runbooks
- [ ] Update disaster recovery plan
- [ ] Train ops team on procedures

---

**Created**: March 11, 2026
**Status**: Ready for Implementation
**Next**: Integrate with v3.3.0 release
