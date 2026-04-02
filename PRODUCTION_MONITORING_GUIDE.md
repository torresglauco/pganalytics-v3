# Production Monitoring & Validation Guide

**Version:** v3.1.0
**Date:** April 2, 2026
**Status:** All features tested and production-ready

---

## 🎯 Overview

This guide provides comprehensive monitoring and validation procedures for pgAnalytics v3.1.0 in production environments. It covers:

- Pre-deployment validation
- Health monitoring during operations
- Feature validation checklists
- Performance monitoring
- Alert configuration
- Incident response procedures

---

## 📋 Pre-Deployment Checklist

### System Requirements Validation

- [ ] **PostgreSQL Version Check**
  ```bash
  psql --version
  # Expected: PostgreSQL 14, 15, 16, 17, or 18
  ```

- [ ] **Docker Installation**
  ```bash
  docker --version
  # Expected: 20.10 or higher
  ```

- [ ] **Available Resources**
  - CPU: Minimum 2 cores, 4+ recommended
  - Memory: Minimum 4 GB, 8+ GB recommended
  - Disk: Minimum 10 GB, 50+ GB for production

- [ ] **Network Connectivity**
  ```bash
  # Test connection to PostgreSQL instances
  psql -h <pg-host> -U postgres -d postgres -c "SELECT version();"
  ```

### Configuration Validation

- [ ] **Environment Variables**
  ```bash
  # Required environment variables
  - DATABASE_URL: PostgreSQL connection string
  - PGANALYTICS_API_PORT: API server port (default: 8080)
  - PGANALYTICS_MCP_ENABLED: Enable MCP server (true/false)
  - PGANALYTICS_MCP_PORT: MCP server port (default: 9000)
  - FRONTEND_URL: Frontend deployment URL
  - COLLECTOR_SECRET: Collector registration secret
  ```

- [ ] **TLS/SSL Certificates**
  ```bash
  # Verify certificate validity
  openssl x509 -in /path/to/cert.pem -text -noout
  # Check expiration: notAfter field should be in future
  ```

- [ ] **Database Migrations**
  ```bash
  # Verify all migrations applied
  psql -h <pg-host> -U postgres -d pganalytics -c "\dt" | grep "pganalytics"
  # Should show: users, api_tokens, collectors, servers, etc.
  ```

---

## 🚀 Deployment & Startup Validation

### Docker Compose Deployment

```bash
# Step 1: Clone and checkout v3.1.0
git clone https://github.com/torresglauco/pganalytics-v3
cd pganalytics-v3
git checkout v3.1.0

# Step 2: Setup environment
cp .env.example .env
# Edit .env with production values

# Step 3: Start services
docker-compose up -d

# Step 4: Verify all containers running
docker-compose ps
```

Expected output:
```
NAME                    STATUS
pganalytics-backend     Up (healthy)
pganalytics-frontend    Up (healthy)
pganalytics-postgres    Up (healthy)
pganalytics-timescaledb Up (healthy)
```

### Kubernetes Deployment

```bash
# Step 1: Apply Helm values
helm install pganalytics ./helm/pganalytics \
  -f helm/pganalytics/values.yaml \
  -n pganalytics --create-namespace

# Step 2: Verify pods
kubectl get pods -n pganalytics
# All pods should be Running

# Step 3: Check services
kubectl get services -n pganalytics
# Backend, Frontend, and MCP services should be listed
```

---

## ✅ Feature Validation Checklist

### 1. Backend API Validation

```bash
# Health check
curl -X GET http://localhost:8080/health
# Expected response: {"status": "healthy"}

# Metrics endpoint
curl -X GET http://localhost:8080/metrics
# Expected: Prometheus-format metrics output

# API version
curl -X GET http://localhost:8080/api/v1/version
# Expected: {"version": "3.1.0"}
```

**Validation Points:**
- [ ] Health endpoint responds with 200 OK
- [ ] Metrics endpoint returns valid Prometheus format
- [ ] API version matches v3.1.0
- [ ] Response times < 100ms for health checks

### 2. Frontend Application Validation

```bash
# Frontend accessibility
curl -X GET http://localhost:3000
# Expected: HTML content for React application

# API connectivity from frontend
# Check browser console (DevTools F12) for no CORS errors
# Check network tab for successful API calls to /api/v1/*
```

**Validation Points:**
- [ ] Frontend loads successfully (200 OK)
- [ ] No CORS errors in browser console
- [ ] API calls to backend succeed
- [ ] Dashboard displays without errors
- [ ] All UI components render correctly

### 3. PostgreSQL Compatibility Validation

Test against each supported version:

```bash
# PostgreSQL 14
psql -h postgres-14 -U postgres -d postgres -c "SELECT version();"
# Expected: PostgreSQL 14.x

# PostgreSQL 15
psql -h postgres-15 -U postgres -d postgres -c "SELECT version();"
# Expected: PostgreSQL 15.x

# PostgreSQL 16
psql -h postgres-16 -U postgres -d postgres -c "SELECT version();"
# Expected: PostgreSQL 16.x

# PostgreSQL 17
psql -h postgres-17 -U postgres -d postgres -c "SELECT version();"
# Expected: PostgreSQL 17.x

# PostgreSQL 18
psql -h postgres-18 -U postgres -d postgres -c "SELECT version();"
# Expected: PostgreSQL 18.x
```

**Validation Points:**
- [ ] Backend can connect to PG14
- [ ] Backend can connect to PG15
- [ ] Backend can connect to PG16
- [ ] Backend can connect to PG17
- [ ] Backend can connect to PG18
- [ ] Query analysis works on all versions
- [ ] Index recommendations work on all versions
- [ ] VACUUM analysis works on all versions

### 4. CLI Tools Validation

```bash
# Query Analysis
pganalytics-cli query analyze <query-id>
# Expected: Performance metrics and recommendations

# Index Suggestions
pganalytics-cli index suggest
# Expected: List of recommended indexes

# VACUUM Analysis
pganalytics-cli vacuum analyze
# Expected: Bloat analysis and recommendations
```

**Validation Points:**
- [ ] CLI tools installed and accessible
- [ ] Query analysis returns valid results
- [ ] Index suggestions include reasoning
- [ ] VACUUM recommendations are actionable
- [ ] JSON output formats correctly
- [ ] Table output formats correctly

### 5. MCP (Model Context Protocol) Validation

```bash
# MCP Server health
curl -X GET http://localhost:9000/health
# Expected: {"status": "healthy"}

# MCP tools availability
# Connect with Claude or compatible MCP client
# Verify 4 tools are registered:
#   1. table_stats
#   2. query_analysis
#   3. index_suggest
#   4. anomaly_detect
```

**Validation Points:**
- [ ] MCP server responds to health checks
- [ ] All 4 tools are registered
- [ ] Tools return valid JSON-RPC 2.0 responses
- [ ] Recommendations include severity levels
- [ ] Integration with AI models succeeds

### 6. ML/AI Service Validation

```bash
# Latency prediction
curl -X POST http://localhost:8000/predict \
  -H "Content-Type: application/json" \
  -d '{"query_complexity": 5, "table_size": 1000000}'
# Expected: {"predicted_latency": XXX, "confidence": XX%}

# Anomaly detection
curl -X POST http://localhost:8000/detect-anomaly \
  -H "Content-Type: application/json" \
  -d '{"metrics": [...], "baseline": [...]}'
# Expected: {"anomaly_score": XX, "severity": "LOW|MEDIUM|HIGH"}
```

**Validation Points:**
- [ ] ML service responds correctly
- [ ] Predictions have >90% accuracy
- [ ] Anomaly scores match historical patterns
- [ ] Inference time < 100ms
- [ ] Model weights loaded successfully

### 7. Collector Integration Validation

```bash
# Collector registration
curl -X POST http://localhost:8080/api/v1/collectors/register \
  -H "Content-Type: application/json" \
  -d '{"name": "collector-1", "host": "localhost", "port": 9000}'
# Expected: Collector ID and registration token

# Metrics ingestion
# Verify metrics arrive in backend
curl -X GET http://localhost:8080/api/v1/metrics \
  -H "Authorization: Bearer <token>"
# Expected: Recent metrics from collectors
```

**Validation Points:**
- [ ] Collectors register successfully
- [ ] Metrics ingestion works
- [ ] Multi-collector support functions
- [ ] Connection pooling handles load
- [ ] No metric loss during ingestion

---

## 📊 Production Monitoring

### Real-Time Health Dashboard

Monitor these key metrics continuously:

```bash
# Backend service health
watch -n 5 'curl -s http://localhost:8080/health | jq .'

# Database connection pool
curl -s http://localhost:8080/metrics | grep db_connections

# API response time (p95)
curl -s http://localhost:8080/metrics | grep http_request_duration_seconds

# Memory usage
docker stats pganalytics-backend | grep MEMORY
```

### Key Performance Indicators (KPIs)

| Metric | Threshold | Action |
|--------|-----------|--------|
| API Response Time (p95) | < 500ms | Alert if exceeds |
| Database Connection Pool Utilization | < 80% | Alert if exceeds |
| Error Rate | < 0.1% | Alert if exceeds |
| Uptime | > 99.9% | Monitor for degradation |
| Memory Usage | < 80% of limit | Alert if exceeds |
| Disk Space | > 10% free | Alert if below |

### Logging & Alerting

**Application Logs:**
```bash
# Backend logs
docker logs -f pganalytics-backend --tail=100

# Frontend logs
docker logs -f pganalytics-frontend --tail=100

# Collector logs
docker logs -f pganalytics-collector --tail=100
```

**Set up alerts for:**
- [ ] Service unavailability (response timeout)
- [ ] Database connection failures
- [ ] High error rates (> 0.1%)
- [ ] Out-of-memory conditions
- [ ] Disk space critically low
- [ ] API latency degradation
- [ ] Failed authentication attempts

---

## 🔍 Daily Validation Checklist

Run daily in production:

### Morning (8:00 AM)

- [ ] **Service Health**
  ```bash
  curl http://localhost:8080/health
  curl http://localhost:3000
  ```
  - All services should respond with 200 OK

- [ ] **Database Connectivity**
  ```bash
  psql -h <pg-host> -U postgres -d pganalytics -c "SELECT COUNT(*) FROM users;"
  ```
  - Database should be reachable

- [ ] **Collector Status**
  - Check collectors are active
  - Verify metrics ingestion rate

### Afternoon (2:00 PM)

- [ ] **Performance Metrics**
  ```bash
  curl http://localhost:8080/metrics | grep histogram_quantile
  ```
  - p95 response time should be < 500ms

- [ ] **Error Logs Review**
  - Check for any error patterns
  - Investigate 5xx responses

- [ ] **Alert Review**
  - Check for active alerts
  - Verify alert thresholds appropriate

### Evening (5:00 PM)

- [ ] **Capacity Check**
  ```bash
  docker stats --no-stream | grep pganalytics
  ```
  - Memory/CPU utilization normal
  - Disk space adequate

- [ ] **Backup Verification**
  - Confirm automated backups ran
  - Verify backup integrity

- [ ] **Test Alert Channels**
  - Send test alert to Slack
  - Verify email notifications work

---

## 🚨 Incident Response Procedures

### Scenario 1: Backend Service Down

```bash
# Step 1: Check service status
docker ps | grep pganalytics-backend

# Step 2: Review logs for errors
docker logs pganalytics-backend --tail=50

# Step 3: Check database connectivity
psql -h <pg-host> -U postgres -d pganalytics -c "SELECT 1;"

# Step 4: Restart service
docker-compose restart pganalytics-backend

# Step 5: Verify recovery
curl http://localhost:8080/health
```

### Scenario 2: High Database Load

```bash
# Step 1: Check slow queries
psql -h <pg-host> -U postgres -d pganalytics -c "SELECT query, calls, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# Step 2: Check active connections
psql -h <pg-host> -U postgres -d pganalytics -c "SELECT datname, usename, count(*) FROM pg_stat_activity GROUP BY datname, usename;"

# Step 3: Check table bloat
pganalytics-cli vacuum analyze

# Step 4: Run VACUUM if needed
psql -h <pg-host> -U postgres -d pganalytics -c "VACUUM ANALYZE;"

# Step 5: Monitor recovery
watch -n 5 'psql -h <pg-host> -U postgres -d pganalytics -c "SELECT load_average FROM pg_stat_database LIMIT 1;"'
```

### Scenario 3: Collector Connection Issues

```bash
# Step 1: Verify collector is running
docker ps | grep collector

# Step 2: Check collector logs
docker logs <collector-container> --tail=50

# Step 3: Test network connectivity
telnet <collector-host> 9000

# Step 4: Verify authentication token
curl -X GET http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer <token>"

# Step 5: Re-register collector if needed
curl -X POST http://localhost:8080/api/v1/collectors/register \
  -H "Content-Type: application/json" \
  -d '{"name": "collector-1", "host": "<collector-host>"}'
```

---

## 🔒 Security Validation

### Authentication & Authorization

- [ ] **API Token Validation**
  ```bash
  # Test invalid token
  curl -H "Authorization: Bearer invalid-token" \
    http://localhost:8080/api/v1/collectors
  # Expected: 401 Unauthorized

  # Test valid token
  curl -H "Authorization: Bearer <valid-token>" \
    http://localhost:8080/api/v1/collectors
  # Expected: 200 OK with data
  ```

- [ ] **User Access Control**
  - Admin users can create/delete collectors
  - Regular users can only view assigned collectors
  - Viewers cannot modify any resources

### TLS/SSL Validation

```bash
# Check certificate expiration
openssl s_client -connect localhost:443 -showcerts < /dev/null | \
  openssl x509 -noout -dates

# Certificate should not expire within 30 days
```

### Data Security

- [ ] **Encrypted Credentials**
  - Collector tokens stored encrypted in database
  - No passwords in application logs

- [ ] **Audit Logging**
  - All API calls logged with user info
  - Configuration changes tracked
  - Failed authentication attempts logged

---

## 📈 Scaling & Load Testing

### Baseline Performance

```bash
# Measure baseline response times
ab -n 1000 -c 10 http://localhost:8080/health
# Note p50, p95, p99 response times

# Run against real API endpoints
ab -n 1000 -c 10 http://localhost:8080/api/v1/metrics
```

### Load Testing Scenario

```bash
# Simulate 100 concurrent collectors
# Each sending 1000 metrics per minute
# Expected: < 1% error rate, < 500ms p95 latency

# Monitor during load:
watch -n 1 'curl -s http://localhost:8080/metrics | grep http_request'
```

### Scaling Indicators

Scale up when:
- API p95 latency > 500ms consistently
- Error rate > 0.1%
- Database connection pool > 80% utilized
- Memory usage > 80% of limit
- CPU usage > 70% sustained

---

## 📊 Weekly Review Checklist

Every Friday:

- [ ] **Performance Summary**
  - Review average response times
  - Identify any degradation trends
  - Compare to baselines

- [ ] **Error Analysis**
  - Aggregate error types
  - Identify root causes
  - Create improvement tasks

- [ ] **Capacity Planning**
  - Review growth trends
  - Project when scaling needed
  - Plan infrastructure updates

- [ ] **Security Review**
  - Audit token usage
  - Review failed authentication attempts
  - Check for suspicious patterns

- [ ] **Backup & Disaster Recovery**
  - Test backup restoration
  - Verify RTO/RPO targets met
  - Update runbooks if needed

---

## 🔗 Key Resources

**Monitoring & Observability:**
- [docs/OPERATIONS_HA_DR.md](docs/OPERATIONS_HA_DR.md) - High availability setup
- [docs/ONCALL_HANDBOOK.md](docs/ONCALL_HANDBOOK.md) - On-call procedures
- [docs/FAQ_AND_TROUBLESHOOTING.md](docs/FAQ_AND_TROUBLESHOOTING.md) - Common issues

**Deployment & Configuration:**
- [DEPLOYMENT.md](DEPLOYMENT.md) - Deployment procedures
- [docs/KUBERNETES_DEPLOYMENT.md](docs/KUBERNETES_DEPLOYMENT.md) - K8s deployment
- [docs/HELM_VALUES_REFERENCE.md](docs/HELM_VALUES_REFERENCE.md) - Helm configuration

**Runbooks & Troubleshooting:**
- [docs/RUNBOOK_CONNECTIONS.md](docs/RUNBOOK_CONNECTIONS.md) - Connection issues
- [docs/RUNBOOK_LOCK_CONTENTION.md](docs/RUNBOOK_LOCK_CONTENTION.md) - Lock debugging
- [docs/RUNBOOK_TABLE_BLOAT.md](docs/RUNBOOK_TABLE_BLOAT.md) - Bloat management

---

## 📞 Support & Escalation

**Tier 1 Support (< 15 minutes):**
- Service health checks
- Basic troubleshooting
- Log review

**Tier 2 Support (< 1 hour):**
- Database performance tuning
- Collector troubleshooting
- Configuration changes

**Tier 3 Support (Engineering Team):**
- Complex debugging
- Code-level investigation
- Infrastructure scaling

**Emergency Contacts:**
- Team Lead: [contact info]
- On-Call Engineer: [rotation info]
- External Support: [vendor info]

---

## ✅ Sign-Off

After completing all validations:

**Deployed by:** _________________ **Date:** _________

**Validated by:** ________________ **Date:** _________

**Approved by:** _________________ **Date:** _________

---

**pgAnalytics v3.1.0 is production-ready!**

All features have been tested, validated, and are ready for monitoring in production environments.

For questions or issues, refer to the documentation or contact the support team.

---

*Last Updated: April 2, 2026*
*Version: v3.1.0*
