# Staging Deployment & Extended Load Test Guide
**Phase 5 (v3.5.0) Production Readiness Validation**
**Date**: March 5, 2026
**Target Duration**: 2-3 days (setup + testing)

---

## Executive Summary

This guide provides step-by-step instructions for deploying pgAnalytics Phase 5 to a staging environment and running an 8-hour sustained load test to validate production readiness.

**Scope**:
- Deploy Phase 5 (v3.5.0) with all features enabled
- Run baseline, medium, full-scale, and sustained load tests
- Validate anomaly detection, alert rules, and notifications
- Collect performance metrics and optimize configuration
- Generate production deployment recommendations

**Prerequisites**:
- Kubernetes cluster (1.24+) with 8+ CPU cores and 16GB+ RAM
- kubectl configured to target staging cluster
- Helm 3.10+
- Docker registry access for pganalytics images
- Slack/Email configured for test notifications

---

## Part 1: Environment Setup

### Step 1.1: Prepare Staging Cluster

```bash
# Verify cluster connectivity
kubectl cluster-info
kubectl get nodes

# Create namespace
kubectl create namespace pganalytics-staging

# Label namespace for monitoring
kubectl label namespace pganalytics-staging environment=staging monitoring=enabled

# Set default namespace
kubectl config set-context --current --namespace=pganalytics-staging
```

### Step 1.2: Install Required Components

```bash
# Install cert-manager for TLS
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Wait for cert-manager to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=300s

# Create staging certificate issuer
cat <<'EOF' | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: admin@pganalytics-staging.io
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

# Install NGINX Ingress Controller
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx --create-namespace

# Wait for ingress to be ready
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=ingress-nginx -n ingress-nginx --timeout=300s
```

### Step 1.3: Setup Storage

```bash
# For AWS EKS, use AWS EBS driver
# For GKE, use Google Persistent Disk
# For local testing, use local storage

# Verify storage classes
kubectl get storageclass

# Create EBS storage class if using AWS EKS
cat <<'EOF' | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: gp3
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
EOF
```

### Step 1.4: Setup Container Registry Credentials

```bash
# Create Docker registry secret for pulling images
kubectl create secret docker-registry pganalytics-registry \
  --docker-server=docker.io \
  --docker-username=YOUR_USERNAME \
  --docker-password=YOUR_PASSWORD \
  --docker-email=YOUR_EMAIL

# Verify secret
kubectl get secret pganalytics-registry
```

---

## Part 2: Deploy Phase 5 to Staging

### Step 2.1: Add Helm Repository

```bash
# If using internal Helm repo, add it
# helm repo add pganalytics https://helm.pganalytics.io

# Or build from local chart
cd /Users/glauco.torres/git/pganalytics-v3/helm/pganalytics

# Verify chart
helm lint .
```

### Step 2.2: Validate Configuration

```bash
# Review staging values
cat values-staging.yaml

# Dry-run to validate deployment
helm install pganalytics-staging . \
  -f values-staging.yaml \
  --namespace pganalytics-staging \
  --dry-run \
  --debug > /tmp/helm-dry-run.yaml

# Check for errors
cat /tmp/helm-dry-run.yaml | head -100
```

### Step 2.3: Deploy Helm Chart

```bash
# Deploy to staging
helm install pganalytics-staging . \
  -f values-staging.yaml \
  --namespace pganalytics-staging \
  --wait \
  --timeout 10m

# Wait for deployment
kubectl rollout status statefulset/pganalytics-staging-postgresql-primary -n pganalytics-staging
kubectl rollout status statefulset/pganalytics-staging-postgresql-replica -n pganalytics-staging
kubectl rollout status deployment/pganalytics-staging-backend -n pganalytics-staging
```

### Step 2.4: Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n pganalytics-staging

# Expected output:
# - pganalytics-staging-postgresql-primary-0      Running
# - pganalytics-staging-postgresql-replica-0      Running
# - pganalytics-staging-backend-0,1,2             Running
# - pganalytics-staging-redis-0                   Running
# - pganalytics-staging-redis-sentinel-0,1,2      Running
# - pganalytics-staging-grafana-0                 Running

# Check logs for errors
kubectl logs -n pganalytics-staging -l app=backend --tail=50

# Verify services
kubectl get services -n pganalytics-staging

# Verify ingress
kubectl get ingress -n pganalytics-staging

# Wait for ingress to get external IP
kubectl get ingress -n pganalytics-staging -w
```

### Step 2.5: Configure Database

```bash
# Port-forward to PostgreSQL
kubectl port-forward -n pganalytics-staging svc/pganalytics-staging-postgresql 5432:5432 &

# Run migrations
# Using your database migration tool (flyway, migrate, etc.)
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -f backend/migrations/017_anomaly_detection.sql

# Verify schema
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;"

# Expected tables include:
# - query_baselines
# - query_anomalies
# - alert_rules
# - alerts
# - notification_channels
# - notification_deliveries
```

### Step 2.6: Configure Notifications (Optional)

```bash
# For staging, you can skip real notifications
# Or configure test channels:

# Port-forward to API
kubectl port-forward -n pganalytics-staging svc/pganalytics-staging-backend 8080:8080 &

# Create test notification channel (email to yourself)
curl -X POST http://localhost:8080/api/v1/notification-channels \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Staging Test Email",
    "channel_type": "email",
    "config": {
      "recipients": ["your-email@example.com"]
    }
  }'

# Test channel
curl -X POST http://localhost:8080/api/v1/notification-channels/{channel-id}/test
```

---

## Part 3: Extended Load Test

### Step 3.1: Prepare Load Test Environment

```bash
# Build load test tool
cd /Users/glauco.torres/git/pganalytics-v3/backend/tests/load
go build -o load-test-runner main.go

# Verify build
./load-test-runner -h
```

### Step 3.2: Run Baseline Test (100 collectors, 5 minutes)

```bash
echo "=== BASELINE TEST START ===" && date

./load-test-runner \
  --url http://localhost:8080 \
  --collectors 100 \
  --metrics 10 \
  --interval 5 \
  --duration 5 \
  --concurrent 10 \
  --verbose \
  > /tmp/baseline_test.log 2>&1

echo "=== BASELINE TEST COMPLETE ===" && date
cat /tmp/baseline_test.log | tail -50
```

**Expected Results**:
- Success rate: >99%
- p95 latency: <200ms
- Error rate: <0.05%
- Cache hit rate: >80%

### Step 3.3: Run Medium Load Test (300 collectors, 10 minutes)

```bash
echo "=== MEDIUM LOAD TEST START ===" && date

./load-test-runner \
  --url http://localhost:8080 \
  --collectors 300 \
  --metrics 10 \
  --interval 5 \
  --duration 10 \
  --concurrent 15 \
  --verbose \
  > /tmp/medium_test.log 2>&1

echo "=== MEDIUM LOAD TEST COMPLETE ===" && date
cat /tmp/medium_test.log | tail -50
```

**Expected Results**:
- Success rate: >99%
- p95 latency: <250ms
- Error rate: <0.05%
- Cache hit rate: >83%

### Step 3.4: Run Full-Scale Test (500 collectors, 30 minutes)

```bash
echo "=== FULL-SCALE TEST START ===" && date

./load-test-runner \
  --url http://localhost:8080 \
  --collectors 500 \
  --metrics 10 \
  --interval 5 \
  --duration 30 \
  --concurrent 20 \
  --verbose \
  > /tmp/fullscale_test.log 2>&1

echo "=== FULL-SCALE TEST COMPLETE ===" && date
cat /tmp/fullscale_test.log | tail -50

# Monitor in real-time in separate terminals:
# Terminal 1: kubectl logs -f -n pganalytics-staging -l app=backend
# Terminal 2: kubectl top nodes
# Terminal 3: kubectl top pods -n pganalytics-staging
```

**Expected Results** (from Phase 4 validation):
- Success rate: >99.8%
- p95 latency: 185ms
- Error rate: 0.06%
- Cache hit rate: 85%+
- Memory stable: <0.15%/min growth

### Step 3.5: Run Extended Sustained Load Test (500 collectors, 60 minutes)

```bash
echo "=== SUSTAINED LOAD TEST START ===" && date
date > /tmp/sustained_test_start.log

./load-test-runner \
  --url http://localhost:8080 \
  --collectors 500 \
  --metrics 10 \
  --interval 5 \
  --duration 60 \
  --concurrent 20 \
  --verbose \
  > /tmp/sustained_test.log 2>&1

echo "=== SUSTAINED LOAD TEST COMPLETE ===" && date
date > /tmp/sustained_test_end.log
cat /tmp/sustained_test.log | tail -100
```

**Critical Validation**:
- ✅ No memory leaks (compare start vs end memory)
- ✅ Consistent latency (no degradation over time)
- ✅ Connection pool not exhausted
- ✅ Rate limiter working fairly
- ✅ Cache hit rate maintained

---

## Part 4: Validate Phase 5 Features

### Step 4.1: Verify Anomaly Detection

```bash
# Check anomaly detection job is running
kubectl logs -n pganalytics-staging -l app=backend | grep -i "anomaly"

# Expected log output:
# "[AnomalyDetector] Starting anomaly detection job"
# "[AnomalyDetector] Running detection cycle #1"
# "[AnomalyDetector] Detection cycle completed in Xms"

# Query anomalies from database
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT COUNT(*) as active_anomalies FROM query_anomalies WHERE is_active = TRUE;"

# Expected: Some anomalies detected (depends on baseline data)

# Check anomaly baselines were calculated
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT COUNT(*) as baselines FROM query_baselines WHERE is_enabled = TRUE;"

# Expected: >100 baselines (one per query metric)
```

### Step 4.2: Verify Alert Rules

```bash
# Check alert rule engine is running
kubectl logs -n pganalytics-staging -l app=backend | grep -i "alert"

# Expected log output:
# "[AlertEngine] Starting alert rule engine job"
# "[AlertEngine] Running evaluation cycle #1"
# "[AlertEngine] Evaluating N rules"

# Create a test alert rule
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Execution Time Test",
    "rule_type": "threshold",
    "metric_name": "execution_time",
    "condition": {
      "type": "threshold",
      "metric": "execution_time",
      "operator": ">",
      "value": 100
    },
    "alert_severity": "high"
  }' | jq .

# Check if rule was created
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT COUNT(*) as alert_rules FROM alert_rules WHERE is_enabled = TRUE;"
```

### Step 4.3: Verify Notifications

```bash
# Check notification service logs
kubectl logs -n pganalytics-staging -l app=backend | grep -i "notification"

# Expected log output:
# "[Notifications] Channel created"
# "[Notifications] Sending alert through N channels"

# List notification channels
curl http://localhost:8080/api/v1/notification-channels | jq .

# Check delivery status
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT delivery_status, COUNT(*) FROM notification_deliveries GROUP BY delivery_status;"

# Expected:
# - Most should be "sent" or "delivered"
# - Few or none should be "failed"
```

### Step 4.4: Verify Phase 4 Optimizations

```bash
# Check rate limiting is working
kubectl logs -n pganalytics-staging -l app=backend | grep -i "rate"

# Check cache statistics
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT COUNT(*) as total, \
           COUNT(*) FILTER (WHERE is_active = TRUE) as active \
           FROM query_baselines;"

# Check connection pool health
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT datname, count(*) FROM pg_stat_activity GROUP BY datname;"
```

---

## Part 5: Performance Analysis

### Step 5.1: Collect Metrics

```bash
# Extract from load test logs
grep "THROUGHPUT:" /tmp/sustained_test.log
grep "LATENCY" /tmp/sustained_test.log
grep "Cache" /tmp/sustained_test.log
grep "Memory" /tmp/sustained_test.log

# From Kubernetes
kubectl top nodes -n pganalytics-staging
kubectl top pods -n pganalytics-staging --sort-by=memory

# From Prometheus (if installed)
# Query: http_request_duration_seconds_bucket (for latency)
# Query: rate(http_requests_total[5m]) (for throughput)
# Query: anomaly_detection_execution_time_ms (for detection speed)
```

### Step 5.2: Generate Test Report

Create `/tmp/STAGING_LOAD_TEST_REPORT.md`:

```bash
cat > /tmp/STAGING_LOAD_TEST_REPORT.md <<'EOF'
# Staging Load Test Report - Phase 5 (v3.5.0)
**Date**: $(date)
**Environment**: Staging (pganalytics-staging)
**Duration**: 2+ hours of testing

## Test Results Summary

### Baseline Test (100 collectors, 5 min)
- Success Rate: ___%
- p95 Latency: ___ms
- Error Rate: ___%
- Cache Hit: ___%

### Medium Load Test (300 collectors, 10 min)
- Success Rate: ___%
- p95 Latency: ___ms
- Error Rate: ___%
- Cache Hit: ___%

### Full-Scale Test (500 collectors, 30 min)
- Success Rate: ___%
- p95 Latency: ___ms
- Error Rate: ___%
- Cache Hit: ___%

### Sustained Load Test (500 collectors, 60 min)
- Success Rate: ___%
- p95 Latency: ___ms (should be consistent)
- Error Rate: ___%
- Memory Growth: ___%/min (should be <0.15%/min)

## Phase 5 Validation

- [ ] Anomaly Detection: Working
- [ ] Alert Rules: Evaluated correctly
- [ ] Notifications: Delivered successfully
- [ ] Rate Limiting: Fair distribution
- [ ] Caching: Hit rate >80%

## Issues & Recommendations

List any issues found and recommendations for optimization.

## Conclusion

Production Readiness: [YES/NO/WITH CONDITIONS]

EOF
cat /tmp/STAGING_LOAD_TEST_REPORT.md
```

### Step 5.3: Compare with Phase 4 Results

| Metric | Phase 4 Target | Staging Result | Status |
|--------|---|---|---|
| Success Rate | >99% | ___% | ✅/⚠️/❌ |
| p95 Latency | <500ms | ___ms | ✅/⚠️/❌ |
| Error Rate | <0.1% | __% | ✅/⚠️/❌ |
| Cache Hit | >75% | __% | ✅/⚠️/❌ |
| Memory Growth | <0.15%/min | __% | ✅/⚠️/❌ |
| Anomaly Detection | Working | Yes/No | ✅/❌ |
| Alert Rules | Working | Yes/No | ✅/❌ |
| Notifications | Delivered | Yes/No | ✅/❌ |

---

## Part 6: Troubleshooting

### Issue: High Latency During Tests

**Diagnosis**:
```bash
# Check pod CPU/memory
kubectl top pods -n pganalytics-staging

# Check database slow queries
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT query, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Check connection pool
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT * FROM pg_stat_database WHERE datname = 'pganalytics';"
```

**Solutions**:
- Increase `MAX_DATABASE_CONNS` if connections are maxed
- Add more replicas if CPU is saturated
- Check for slow queries and optimize

### Issue: Memory Growth

**Diagnosis**:
```bash
# Get pod memory over time
kubectl top pods -n pganalytics-staging -n pganalytics-staging --containers

# Check for memory leaks
kubectl logs -n pganalytics-staging -l app=backend | grep -i "memory\|leak\|gc"
```

**Solutions**:
- Ensure collector cleanup job is running
- Check cache eviction is working
- Increase memory limits if needed

### Issue: Notification Delivery Failures

**Diagnosis**:
```bash
# Check notification logs
kubectl logs -n pganalytics-staging -l app=backend | grep -i "notification"

# Query delivery status
PGPASSWORD=staging-postgres-password \
  psql -h localhost -U pganalytics -d pganalytics -c \
  "SELECT channel_id, delivery_status, COUNT(*) FROM notification_deliveries GROUP BY channel_id, delivery_status;"
```

**Solutions**:
- Verify notification channel configuration
- Check network connectivity to external services
- Enable retry queue

---

## Part 7: Production Deployment Preparation

### If All Tests Pass ✅

1. **Update Docker images**:
   ```bash
   # Tag images as production
   docker tag pganalytics/api:3.5.0-staging pganalytics/api:3.5.0
   docker tag pganalytics/collector:3.5.0-staging pganalytics/collector:3.5.0
   docker push pganalytics/api:3.5.0
   docker push pganalytics/collector:3.5.0
   ```

2. **Create production values file**:
   ```bash
   cp helm/pganalytics/values-staging.yaml helm/pganalytics/values-prod.yaml
   # Edit for production (different secrets, endpoints, etc.)
   ```

3. **Update Helm chart version**:
   ```bash
   # Update Chart.yaml version to 3.5.0
   sed -i 's/appVersion: "3.4.0"/appVersion: "3.5.0"/' helm/pganalytics/Chart.yaml
   ```

4. **Create production deployment runbook** (see PRODUCTION_DEPLOYMENT_RUNBOOK.md)

5. **Schedule production rollout** (canary deployment recommended)

### If Issues Found ⚠️

1. **Document issues found**
2. **Create GitHub issues** for each problem
3. **Fix in dev environment**
4. **Re-test in staging**
5. **Re-run load tests** before production

---

## Checklist for Production Readiness

- [ ] All load tests passed (success rate >99%)
- [ ] Latency acceptable (<500ms p95)
- [ ] Memory stable (no leaks)
- [ ] Anomaly detection working
- [ ] Alert rules evaluating correctly
- [ ] Notifications delivering successfully
- [ ] Rate limiting fair
- [ ] Cache hit rate >75%
- [ ] Database replication working
- [ ] Redis Sentinel HA working
- [ ] Graceful shutdown working
- [ ] TLS certificates valid
- [ ] Monitoring configured
- [ ] Backup tested
- [ ] Documentation complete
- [ ] Operational runbooks written

---

## Post-Deployment Monitoring

After production deployment:

```bash
# Monitor for 24 hours:
- p95 latency (target <500ms)
- Error rate (target <0.1%)
- False positive rate (anomalies)
- Notification delivery success rate
- Resource utilization

# Key alerts:
- High latency (>500ms p95)
- High error rate (>0.1%)
- Memory growth anomaly
- Rate limit rejections excessive
- Notification delivery failures

# Daily:
- Review anomaly detection accuracy
- Check alert rule false positives
- Optimize thresholds based on patterns
```

---

**End of Staging Deployment Guide**

For questions or issues, refer to:
- `PHASE5_ANOMALY_DETECTION.md` - Technical details
- `PROJECT_IMPLEMENTATION_STATUS.md` - Overall status
- GitHub Issues: https://github.com/torresglauco/pganalytics-v3/issues
