# pgAnalytics v3.3.0 → v3.5.0 Implementation Roadmap

## Executive Summary

This document outlines the complete implementation plan for three major version releases of pgAnalytics, adding **enterprise-grade features, scalability for 500+ collectors, and advanced anomaly detection with intelligent alerting**.

**Total Effort**: 560 hours (~14 weeks with 3 devs, or 8 weeks with 5 devs)

---

## Phase 3 (v3.3.0): Enterprise Features

### Timeline: 4 weeks | Team: 2-3 developers | Effort: 220 hours

Enterprise authentication, encryption, high availability, and audit logging.

#### 3.1 Enterprise Authentication (LDAP/SAML/OAuth/MFA) - 80 hours

**Priority**: 🔴 CRITICAL - Customer demand

**Objective**: Support corporate authentication in enterprise environments.

**Components**:

1. **LDAP/Active Directory** (`/backend/internal/auth/ldap.go`)
   ```
   - LDAPConnector with TLS support
   - Methods: AuthenticateUser(), SyncUserGroups(), GetUserAttributes()
   - Config: LDAP_SERVER_URL, LDAP_BIND_DN, LDAP_BIND_PASSWORD, LDAP_*_SEARCH_BASE
   - Group-to-role mapping from JSON config
   - Middleware integration for transparent LDAP auth
   ```

2. **SAML 2.0 SSO** (`/backend/internal/auth/saml.go`)
   ```
   - Use: github.com/crewjam/saml
   - Endpoints:
     * GET /api/v1/auth/saml/metadata (XML descriptor)
     * GET /api/v1/auth/saml/acs (Assertion Consumer Service)
     * GET /api/v1/auth/saml/sls (Single Logout)
   - Config: SAML_CERT_PATH, SAML_KEY_PATH, SAML_IDP_URL
   ```

3. **OAuth 2.0/OIDC** (`/backend/internal/auth/oauth.go`)
   ```
   - Library: golang.org/x/oauth2
   - Providers: Google, Azure AD, GitHub, custom OIDC
   - Endpoints:
     * GET /api/v1/auth/oauth/:provider/login
     * GET /api/v1/auth/oauth/callback
   - Config: OAUTH_PROVIDERS (JSON per provider)
   ```

4. **Multi-Factor Authentication** (`/backend/internal/auth/mfa.go`)
   ```
   - TOTP: github.com/pquerna/otp
   - SMS: Twilio/AWS SNS integration
   - Backup codes: Generate/validate
   - Endpoints:
     * POST /api/v1/users/mfa/setup
     * POST /api/v1/users/mfa/verify
     * POST /api/v1/auth/mfa/challenge
   ```

5. **Session Management** (`/backend/internal/session/session.go`)
   ```
   - Backend: Redis (distributed)
   - Model: Session {session_id, user_id, expires_at, ip, user_agent}
   - Methods: CreateSession(), ValidateSession(), RevokeSession()
   - Logout: Add to Redis blacklist
   ```

**Database Schema**:
```sql
-- New tables in migration 011_enterprise_auth.sql
user_mfa_methods (id, user_id, type, secret, verified, created_at)
user_backup_codes (id, user_id, code_hash, used, created_at)
user_sessions (id, user_id, session_token_hash, ip, user_agent, expires_at, created_at)
oauth_providers (id, provider_name, client_id_encrypted, client_secret_encrypted, config_json)
```

**Configuration**:
```bash
# LDAP
LDAP_SERVER_URL=ldap://ldap.example.com:389
LDAP_BIND_DN=cn=admin,dc=example,dc=com
LDAP_BIND_PASSWORD_ENCRYPTED=${VAULT_SECRET}
LDAP_USER_SEARCH_BASE=ou=users,dc=example,dc=com
LDAP_GROUP_SEARCH_BASE=ou=groups,dc=example,dc=com
LDAP_GROUP_TO_ROLE_MAPPING={"admin_group":"admin","user_group":"viewer"}

# SAML
SAML_CERT_PATH=/etc/pganalytics/saml_cert.pem
SAML_KEY_PATH=/etc/pganalytics/saml_key.pem
SAML_IDP_URL=https://idp.example.com/sso
SAML_ENTITY_ID=pganalytics.example.com

# OAuth
OAUTH_PROVIDERS=[{"name":"google","client_id":"...","client_secret":"..."}]

# MFA
MFA_TOTP_ENABLED=true
MFA_SMS_ENABLED=true
MFA_SMS_PROVIDER=twilio|sns
```

**Success Criteria**:
- ✅ LDAP login works against test AD instance
- ✅ SAML assertion processing complete
- ✅ OAuth works with Google/Azure/GitHub
- ✅ TOTP setup/verify operational
- ✅ SMS MFA functional
- ✅ Backup codes working
- ✅ Sessions persist in Redis

---

#### 3.2 Encryption at Rest & Key Management - 60 hours

**Priority**: 🔴 CRITICAL - Compliance requirement

**Objective**: Protect sensitive data with transparent encryption and key rotation.

**Components**:

1. **Column-Level Encryption** (`/backend/internal/crypto/column_encryption.go`)
   ```
   - Algorithm: AES-256-GCM (already in crypto.go, extend it)
   - Crypt columns:
     * users.email
     * users.password_hash (extra layer)
     * registration_secrets.secret_value [CRITICAL - currently plaintext!]
     * postgresql_instances.connection_string [CRITICAL - currently plaintext!]
     * api_tokens.token_hash
     * audit_log.changes

   - Migration Strategy:
     1. Add {column}_encrypted columns
     2. Migration script: copy and encrypt data
     3. Update application to read from _encrypted
     4. Deprecate old column after validation period
     5. Drop old column in future release
   ```

2. **Key Management System** (`/backend/internal/crypto/key_manager.go`)
   ```
   - Backends:
     * AWS Secrets Manager (recommended)
     * HashiCorp Vault (fallback)
     * Google Cloud KMS
     * Local keyfile (dev only)

   - Features:
     * Key versioning with metadata
     * Automatic rotation every 90 days
     * Background reencryption job
     * Key retirement tracking

   - Methods:
     GetCurrentKey() -> []byte
     GetKeyByVersion(version) -> []byte
     RotateKey() -> error
     ScheduleReencryption(table, column) -> error
   ```

3. **Backup Encryption** (`/backend/internal/backup/backup.go`)
   ```
   - Encrypt pg_dump output with AES-256-GCM
   - Separate key versioning for backups
   - Restore capability with old key versions
   ```

**Database Schema**:
```sql
-- In migration 012_encryption_schema.sql
encryption_keys (
  version INTEGER PRIMARY KEY,
  algorithm VARCHAR(50),
  created_at TIMESTAMP,
  retired_at TIMESTAMP,
  key_material_encrypted BYTEA
)

-- Add encrypted columns to tables
ALTER TABLE users ADD COLUMN email_encrypted BYTEA;
ALTER TABLE registration_secrets ADD COLUMN secret_value_encrypted BYTEA;
ALTER TABLE postgresql_instances ADD COLUMN connection_string_encrypted BYTEA;
-- ... etc
```

**Configuration**:
```bash
ENCRYPTION_KEY_BACKEND=aws|vault|gcp|local
AWS_SECRETS_MANAGER_ARN=arn:aws:secretsmanager:...
VAULT_ADDR=https://vault.example.com
VAULT_TOKEN=${VAULT_SECRET}
GCP_KMS_KEY_NAME=projects/.../locations/.../keyRings/.../cryptoKeys/...
KEY_ROTATION_INTERVAL_DAYS=90
ENCRYPTION_ALGORITHM=aes-256-gcm
```

**Success Criteria**:
- ✅ All sensitive data encrypted
- ✅ No query performance degradation
- ✅ Key rotation works without downtime
- ✅ Backups encrypted and restorable
- ✅ Migration scripts tested with real data
- ✅ Encryption/decryption working correctly

---

#### 3.3 High Availability & Failover - 50 hours

**Priority**: 🟠 HIGH - Production requirement

**Objective**: Achieve 99.9% uptime with automatic failover.

**Components**:

1. **PostgreSQL Replication**
   ```
   - Setup: Primary + 1+ Standby with streaming replication
   - Tool: pg_auto_failover or Patroni for orchestration
   - RTO: < 2 seconds
   - Connection string uses failover endpoint (not primary directly)
   - Replication slots for log retention
   ```

2. **Database Connection Optimization**
   ```
   - Increase pool: MaxOpenConns 50 → 200, MaxIdleConns 15 → 60
   - Read replica routing: send SELECT queries to replicas
   - Connection retry logic with exponential backoff
   ```

3. **Distributed Session State**
   ```
   - Redis Sentinel setup: 3+ replicas with quorum
   - AOF persistence enabled
   - Replication lag monitoring
   - Sessions auto-persist across Redis failover
   ```

4. **API Graceful Shutdown** (`/backend/internal/api/server.go`)
   ```go
   // On SIGTERM:
   1. Set readiness probe to false (stops new requests)
   2. Wait for in-flight requests (timeout 30s)
   3. Close database connections
   4. Close Redis connections
   5. Exit cleanly
   ```

5. **Enhanced Health Checks**
   ```
   GET /api/v1/health returns:
   {
     "status": "ok|degraded|unavailable",
     "database_ok": true,
     "replica_lag_ms": 150,
     "redis_ok": true,
     "timestamp": "2026-03-05T..."
   }
   ```

**Helm Chart Updates**:
```yaml
# New files
/helm/pganalytics/templates/postgresql-primary-statefulset.yaml
/helm/pganalytics/templates/postgresql-replica-statefulset.yaml
/helm/pganalytics/templates/postgresql-failover-service.yaml
/helm/pganalytics/templates/redis-sentinel-statefulset.yaml

# Updated
/helm/pganalytics/values-prod.yaml
  postgresql:
    replication:
      enabled: true
      slots: 3
```

**Success Criteria**:
- ✅ Primary/replica replication working
- ✅ Automatic failover < 2 seconds RTO
- ✅ Sessions persist across failover
- ✅ Graceful shutdown without request loss
- ✅ Health checks accurate
- ✅ Load balanced reads across replicas

---

#### 3.4 Audit Logging & Compliance - 30 hours

**Priority**: 🟠 HIGH - Compliance/security

**Objective**: Immutable tracking of all user actions for compliance.

**Components**:

1. **Audit Log System** (`/backend/internal/audit/audit.go`)
   ```go
   type AuditLog struct {
     ID           int
     UserID       int
     Action       string // create, update, delete, login, logout
     ResourceType string // user, collector, alert_rule, etc.
     ResourceID   string
     ChangesBefore interface{} // JSONB
     ChangesAfter  interface{} // JSONB
     IPAddress    string
     UserAgent    string
     Timestamp    time.Time
   }

   LogAction(ctx, action, resource, before, after) error
   GetHistory(filters) []AuditLog
   ExportAuditLog(format string) ([]byte, error)
   ```

2. **Audit Integration Points**
   ```
   Track in handlers:
   - POST /api/v1/users (Create user)
   - PUT /api/v1/users/:id (Update user)
   - DELETE /api/v1/users/:id (Delete user)
   - POST /api/v1/collectors (Register)
   - DELETE /api/v1/collectors/:id (Unregister)
   - PUT /api/v1/collectors/:id/config (Config change)
   - POST /api/v1/auth/login
   - POST /api/v1/auth/logout
   - POST /api/v1/users/:id/change-password
   - POST /api/v1/auth/refresh-token
   ```

3. **Audit APIs** (`/backend/internal/api/handlers_audit.go`)
   ```
   GET /api/v1/audit-logs
     ?user_id=X&resource_type=Y&action=Z&date_from=T1&date_to=T2
   GET /api/v1/audit-logs/:id
   GET /api/v1/audit-logs/export?format=csv|json
   GET /api/v1/audit-logs/stats
   ```

**Database Schema**:
```sql
-- In migration 013_audit_logs.sql
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  action VARCHAR(50) NOT NULL,
  resource_type VARCHAR(100) NOT NULL,
  resource_id VARCHAR(255),
  changes_before JSONB,
  changes_after JSONB,
  ip_address INET,
  user_agent TEXT,
  timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Make immutable
CREATE TRIGGER audit_logs_immutable
  BEFORE UPDATE OR DELETE ON audit_logs
  FOR EACH ROW EXECUTE FUNCTION prevent_update_delete();

-- Archival
CREATE TABLE audit_logs_archive (LIKE audit_logs);
-- Auto-archive logs > 1 year old to S3 or filesystem
```

**Configuration**:
```bash
AUDIT_ENABLED=true
AUDIT_RETENTION_DAYS=365
AUDIT_ARCHIVE_PATH=s3://bucket/audit/ or /archive/audit/
AUDIT_ARCHIVE_SCHEDULE=daily # Daily archival of old logs
```

**Success Criteria**:
- ✅ All critical actions logged
- ✅ Logs immutable (no update/delete)
- ✅ Export functionality working
- ✅ Archive process functional
- ✅ Queries fast (proper indexing)
- ✅ No performance impact

---

### Phase 3 Summary

| Component | Hours | Difficulty | Risk |
|-----------|-------|-----------|------|
| Enterprise Auth | 80 | Medium | Medium |
| Encryption | 60 | High | High |
| HA/Failover | 50 | Medium | Low |
| Audit Logging | 30 | Low | Low |
| **Total** | **220** | - | - |

**Team**: 2-3 developers (4 weeks parallelized)

---

## Phase 4 (v3.4.0): Collector Scalability

### Timeline: 4 weeks | Team: 1-2 developers | Effort: 130 hours

Optimize for 500+ concurrent collectors with improved networking and threading.

#### 4.1 Backend Optimization for 500+ Collectors - 40 hours

**Objective**: Scale API to handle 500+ metric submissions per minute.

**Components**:

1. **Rate Limiting Enhancement** (`/backend/internal/api/ratelimit.go`)
   ```go
   // Update token bucket limits
   routes := map[string]int{
     "/api/v1/metrics/push":               10_000, // 10K req/min
     "/api/v1/config/refresh":             1_000,  // 1K req/min
     "/api/v1/collectors/refresh-token":   500,    // 500 req/min
   }
   ```

2. **Database Connection Pool** (`/backend/internal/storage/postgres.go`)
   ```go
   // In config initialization:
   MaxOpenConns:    200,  // was 50
   MaxIdleConns:    60,   // was 15
   MaxConnLifetime: 5 * time.Minute,
   MaxConnIdleTime: 2 * time.Minute,

   // Add read replica support
   readReplicaConns *sql.DB
   ```

3. **Collector Auto-Cleanup** (`/backend/internal/jobs/collector_cleanup.go`)
   ```go
   // Daily job to remove stale collectors
   CleanupOfflineCollectors(maxOfflineDays int) error

   // Deletes collectors offline > 7 days (configurable)
   // SQL: DELETE FROM collectors WHERE last_seen < NOW() - interval '7 days'
   ```

4. **Configuration Caching** (`/backend/internal/cache/config_cache.go`)
   ```go
   type ConfigCache struct {
     // Redis backend
     // Key: collector:{collector_id}:config:{version}
     // TTL: 5 minutes
   }

   GetCollectorConfig(collectorID uuid.UUID) (*CollectorConfig, error)
   // Uses Redis cache, fallback to DB
   // Reduces DB queries by ~50% for active collectors
   ```

**Success Criteria**:
- ✅ API handles 10K req/min without queue buildup
- ✅ Connection pool stable
- ✅ Config caching reduces DB load
- ✅ Cleanup job removes stale collectors
- ✅ Latency p95 < 500ms under sustained load

---

#### 4.2 Collector C++ Optimization - 60 hours

**Objective**: Optimize collector for 500+ concurrent connections.

**Components**:

1. **Thread Pool Lock-Free Queue** (`/collector/include/thread_pool.h`)
   ```cpp
   // Replace std::queue + std::mutex
   #include <boost/lockfree/queue.hpp>

   class ThreadPool {
   private:
     boost::lockfree::queue<Task> task_queue;
     // Reduces lock contention by 90%
   };

   // Dynamic sizing
   size_t optimal_threads = std::thread::hardware_concurrency();
   // Min 2, Max 32 threads
   // Configurable: COLLECTOR_THREAD_POOL_SIZE env var
   ```

2. **Task Timeout Management**
   ```cpp
   struct Task {
     std::function<void()> fn;
     std::chrono::seconds timeout; // default 30s
     std::chrono::steady_clock::time_point deadline;
   };

   // Detect and terminate hung tasks
   ```

3. **HTTP/2 Connection Pooling** (`/collector/src/sender.cpp`)
   ```cpp
   // Use libcurl with HTTP/2 multiplexing
   // Maintain 4-8 persistent connections
   // Multiplexed streams reduce overhead

   curl_easy_setopt(handle, CURLOPT_HTTP_VERSION, CURL_HTTP_VERSION_2_0);
   // Pool management for connection reuse
   ```

4. **Binary Protocol Default**
   ```cpp
   // /collector/src/sender.cpp
   // Default: binary protocol (70% bandwidth savings vs JSON)
   // Auto-fallback to JSON on error
   // Configurable via COLLECTOR_PROTOCOL env var
   ```

5. **Metrics Buffer Optimization** (`/collector/include/metrics_buffer.h`)
   ```cpp
   // Increase batch size: 1-2 → 5-10 metrics per push
   // Compress batches with zstd
   // Smart batching: send when buffer full OR 10s timer
   ```

**Success Criteria**:
- ✅ Lock-free queue reduces contention by 90%
- ✅ HTTP/2 multiplexing working correctly
- ✅ Binary protocol reduces bandwidth by 70%
- ✅ Throughput increases by 150%+
- ✅ Memory usage optimized
- ✅ 500 concurrent connections stable

---

#### 4.3 Load Testing & Validation - 30 hours

**Objective**: Validate 500+ collector performance.

**Components**:

1. **Stress Test Suite** (`/backend/tests/load/load_test.go`)
   ```go
   // Simulate 500 collectors:
   // - Register collectors
   // - Push metrics every 5 seconds
   // - Run for 8 hours
   // - Measure:
   //   * Latency p95 < 500ms
   //   * Error rate < 0.1%
   //   * Memory stable
   //   * CPU usage reasonable
   ```

2. **Collector Simulator**
   ```go
   // tool/collector-simulator/main.go
   // Simulates N collectors registering and pushing metrics
   // HTTP benchmark with customizable payload
   ```

3. **Documentation**
   - `SCALABILITY_GUIDE.md`: How to configure for 500+ collectors
   - Capacity planning guidelines
   - Performance tuning recommendations

**Success Criteria**:
- ✅ 500 collectors sustained for 8+ hours
- ✅ Latency p95 < 500ms consistently
- ✅ Error rate < 0.1%
- ✅ Memory stable (no leaks)
- ✅ CPU utilization reasonable

---

### Phase 4 Summary

| Component | Hours | Difficulty | Risk |
|-----------|-------|-----------|------|
| Backend Optimization | 40 | Low | Low |
| Collector C++ Optimization | 60 | High | Medium |
| Load Testing | 30 | Medium | Low |
| **Total** | **130** | - | - |

**Team**: 1-2 developers (4 weeks)

---

## Phase 5 (v3.5.0): Advanced Analytics & Alerting

### Timeline: 4 weeks | Team: 2 developers | Effort: 210 hours

Intelligent anomaly detection, advanced alerting, and real-time notifications.

#### 5.1 Anomaly Detection Engine - 50 hours

**Priority**: 🟢 MEDIUM - Competitive differentiator

**Objective**: Automatically detect performance anomalies using ML and statistics.

**Components**:

1. **Anomaly Detector Job** (`/backend/internal/jobs/anomaly_detector.go`)
   ```go
   // Scheduler: every 5 minutes
   type AnomalyDetector struct {
     db *sql.DB
     mlClient *ml.Client
   }

   func (ad *AnomalyDetector) Run(ctx context.Context) error {
     // 1. Get all active monitored queries
     // 2. Calculate baselines
     // 3. Detect anomalies in parallel
     // 4. Store results in DB
   }
   ```

2. **Detection Algorithms**

   **Statistical Z-Score**:
   ```
   - Baseline: 7-day rolling window
   - Calculate: mean, stddev, min, max for each metric
   - Detect: |value - mean| > threshold * stddev
   - Thresholds:
     * ±2.5 stddev → CRITICAL severity
     * ±1.5 stddev → WARNING severity
   ```

   **ML-Based Detection** (via ML service):
   ```
   - Isolation Forest for outlier detection
   - Seasonal decomposition (SARIMA) for trend analysis
   - LSTM for time-series prediction
   - Ensemble voting for final anomaly score
   ```

   **Seasonal Decomposition**:
   ```
   - Detect cyclic patterns (hourly peaks, weekly patterns)
   - Compare against seasonal baseline
   - Configurable seasonal periods
   ```

3. **Baseline Calculation** (SQL function)
   ```sql
   -- In migration 014_anomaly_detection.sql

   CREATE OR REPLACE FUNCTION calculate_baselines_and_anomalies()
   RETURNS TABLE (query_id INT, anomalies INT) AS $$
   BEGIN
     -- 1. Calculate baselines for all queries (7-day window)
     INSERT INTO query_baselines (query_id, metric_name, mean, stddev, min, max, window_start)
     SELECT
       q.id,
       qm.metric_name,
       AVG(qm.value),
       STDDEV(qm.value),
       MIN(qm.value),
       MAX(qm.value),
       NOW() - INTERVAL '7 days'
     FROM queries q
     JOIN query_metrics qm ON q.id = qm.query_id
     WHERE qm.timestamp > NOW() - INTERVAL '7 days'
     GROUP BY q.id, qm.metric_name;

     -- 2. Detect anomalies
     INSERT INTO query_anomalies (query_id, metric_name, value, z_score, severity, detected_at)
     SELECT
       q.id,
       qm.metric_name,
       qm.value,
       (qm.value - qb.mean) / NULLIF(qb.stddev, 0) as z_score,
       CASE
         WHEN ABS((qm.value - qb.mean) / NULLIF(qb.stddev, 0)) > 2.5 THEN 'CRITICAL'
         WHEN ABS((qm.value - qb.mean) / NULLIF(qb.stddev, 0)) > 1.5 THEN 'WARNING'
         ELSE NULL
       END as severity,
       NOW()
     FROM queries q
     JOIN query_metrics qm ON q.id = qm.query_id
     JOIN query_baselines qb ON q.id = qb.query_id AND qm.metric_name = qb.metric_name
     WHERE qm.timestamp > NOW() - INTERVAL '5 minutes'
     AND qb.mean IS NOT NULL;

     RETURN QUERY SELECT q.id, COUNT(*)::INT FROM query_anomalies a
     JOIN queries q ON a.query_id = q.id
     WHERE a.detected_at > NOW() - INTERVAL '5 minutes'
     GROUP BY q.id;
   END;
   $$ LANGUAGE plpgsql;
   ```

4. **Database Schema**
   ```sql
   CREATE TABLE query_baselines (
     id BIGSERIAL PRIMARY KEY,
     query_id INT REFERENCES queries(id),
     metric_name VARCHAR(255),
     mean FLOAT,
     stddev FLOAT,
     min FLOAT,
     max FLOAT,
     window_start TIMESTAMP,
     created_at TIMESTAMP DEFAULT NOW(),
     UNIQUE(query_id, metric_name, window_start)
   );

   CREATE TABLE query_anomalies (
     id BIGSERIAL PRIMARY KEY,
     query_id INT REFERENCES queries(id),
     metric_name VARCHAR(255),
     value FLOAT,
     z_score FLOAT,
     severity VARCHAR(20), -- LOW, MEDIUM, HIGH, CRITICAL
     detected_at TIMESTAMP,
     acknowledged BOOLEAN DEFAULT FALSE
   );
   ```

**Success Criteria**:
- ✅ Baselines calculated correctly every 5 minutes
- ✅ Anomalies detected with >90% precision
- ✅ False positive rate < 5%
- ✅ Performance impact < 5% CPU
- ✅ ML models training and inferencing correctly

---

#### 5.2 Alert Rule Execution Engine - 40 hours

**Priority**: 🟢 MEDIUM - Core alerting

**Objective**: Evaluate alert rules and manage alert lifecycle.

**Components**:

1. **Alert Rule Engine** (`/backend/internal/jobs/alert_rule_engine.go`)
   ```go
   type AlertRuleEngine struct {
     db *sql.DB
     maxConcurrent int // 10
   }

   func (are *AlertRuleEngine) Run(ctx context.Context) error {
     // 1. Fetch all enabled rules
     // 2. Evaluate in parallel (max concurrency 10)
     // 3. Manage state transitions
     // 4. Trigger notifications for state changes
   }
   ```

2. **Rule Types**
   ```go
   // Threshold: metric > value
   type ThresholdRule struct {
     MetricName string
     Operator   string // >, <, >=, <=, ==, !=
     Threshold  float64
     Duration   time.Duration // Must exceed threshold for this duration
   }

   // Change: % variation between readings
   type ChangeRule struct {
     MetricName string
     ChangePercent float64
     Period time.Duration
   }

   // Anomaly: triggered by anomaly detector
   type AnomalyRule struct {
     QueryID int
     Severity string // LOW, MEDIUM, HIGH, CRITICAL
   }

   // Composite: AND/OR combinations
   type CompositeRule struct {
     Operator string // AND, OR
     Rules []AlertRule
   }
   ```

3. **State Machine**
   ```
   FIRING (rule condition true)
     ↓
   ALERTING (if persistent > threshold)
     ↓
   RESOLVED (condition becomes false)

   Config:
   - resolve_time_threshold: 5 minutes (alert must be resolved for this long)
   - alert_window: 1 hour (how far back to check for state changes)
   ```

4. **Deduplication**
   ```go
   // Fingerprint = hash(rule_id + database + severity)
   // Dedup window: 5 minutes (configurable)
   // Prevents notification storms for the same condition
   ```

5. **Database Schema**
   ```sql
   CREATE TABLE alert_rules (
     id BIGSERIAL PRIMARY KEY,
     name VARCHAR(255) NOT NULL,
     description TEXT,
     rule_type VARCHAR(50), -- threshold, change, anomaly, composite
     rule_config JSONB,
     enabled BOOLEAN DEFAULT TRUE,
     severity VARCHAR(20), -- LOW, MEDIUM, HIGH, CRITICAL
     created_by INT REFERENCES users(id),
     created_at TIMESTAMP DEFAULT NOW(),
     updated_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE alerts (
     id BIGSERIAL PRIMARY KEY,
     rule_id BIGINT REFERENCES alert_rules(id),
     status VARCHAR(50), -- FIRING, ALERTING, RESOLVED
     severity VARCHAR(20),
     message TEXT,
     context JSONB, -- metric values, etc.
     first_triggered_at TIMESTAMP,
     last_triggered_at TIMESTAMP,
     resolved_at TIMESTAMP,
     acknowledged BOOLEAN DEFAULT FALSE,
     acknowledged_by INT REFERENCES users(id),
     acknowledged_at TIMESTAMP,
     acknowledged_note TEXT
   );

   CREATE TABLE alert_history (
     id BIGSERIAL PRIMARY KEY,
     alert_id BIGINT REFERENCES alerts(id),
     old_status VARCHAR(50),
     new_status VARCHAR(50),
     changed_at TIMESTAMP DEFAULT NOW()
   );
   ```

**Success Criteria**:
- ✅ Rules evaluated every 5 minutes reliably
- ✅ State transitions working correctly
- ✅ Deduplication prevents false alert storms
- ✅ Composite rules working (AND/OR)
- ✅ Acknowledge/resolve functionality
- ✅ Performance: < 100ms per rule

---

#### 5.3 Notification Delivery System - 45 hours

**Priority**: 🟢 MEDIUM - User experience critical

**Objective**: Multi-channel alert notifications with retry logic.

**Components**:

1. **Notification Service** (`/backend/internal/notifications/notification_service.go`)
   ```go
   type NotificationChannel interface {
     Name() string
     Send(ctx context.Context, notification *Notification) error
     Test(ctx context.Context, config interface{}) error
   }

   // Implementations:
   // - SlackNotifier (via slack-go/slack)
   // - EmailNotifier (SMTP + templates)
   // - WebhookNotifier (HTTP POST)
   // - PagerDutyNotifier (PagerDuty API)
   // - JiraNotifier (Jira API)
   ```

2. **Message Templates** (`/backend/templates/alerts/`)
   ```
   slack.tmpl      - Markdown with context
   email.html.tmpl - Rich HTML email
   email.txt.tmpl  - Plain text fallback
   webhook.json    - JSON payload structure
   pagerduty.json  - PagerDuty-specific format
   ```

3. **Retry Logic**
   ```go
   // Exponential backoff
   backoffs := []time.Duration{
     1 * time.Second,
     2 * time.Second,
     4 * time.Second,
     8 * time.Second,
     16 * time.Second,
   }

   // Max 5 retries, then move to DLQ (Dead Letter Queue) in Redis
   // Monitor DLQ and allow manual retry
   ```

4. **Rate Limiting**
   ```
   - Max 10 notifications per rule per hour
   - Throttling for same rule + severity
   - Configurable burst allowance
   ```

5. **Configuration**
   ```sql
   CREATE TABLE notification_channels (
     id BIGSERIAL PRIMARY KEY,
     name VARCHAR(255) NOT NULL,
     type VARCHAR(50), -- slack, email, webhook, pagerduty, jira
     config JSONB, -- encrypted credentials
     enabled BOOLEAN DEFAULT TRUE,
     created_by INT REFERENCES users(id),
     created_at TIMESTAMP DEFAULT NOW()
   );

   CREATE TABLE alert_rule_channels (
     alert_rule_id BIGINT REFERENCES alert_rules(id),
     channel_id BIGINT REFERENCES notification_channels(id),
     PRIMARY KEY (alert_rule_id, channel_id)
   );

   CREATE TABLE notification_delivery_log (
     id BIGSERIAL PRIMARY KEY,
     alert_id BIGINT REFERENCES alerts(id),
     channel_id BIGINT REFERENCES notification_channels(id),
     status VARCHAR(50), -- sent, failed, retrying
     attempts INT,
     last_error TEXT,
     delivered_at TIMESTAMP,
     created_at TIMESTAMP DEFAULT NOW()
   );
   ```

**Success Criteria**:
- ✅ All 5 notification channels working
- ✅ Retry mechanism functioning
- ✅ 99%+ delivery success rate
- ✅ Rate limiting prevents storms
- ✅ Message formatting correct
- ✅ DLQ capturing failed messages

---

#### 5.4 Backend Alert APIs - 35 hours

**Priority**: 🟢 MEDIUM - API completeness

**Objective**: Complete REST API for alerts and rules.

**Components**:

1. **Alert Management APIs** (`/backend/internal/api/handlers_alerts.go`)
   ```
   GET    /api/v1/alerts
     ?severity=CRITICAL&status=FIRING&limit=50&offset=0
     Returns: []Alert with pagination

   GET    /api/v1/alerts/:id
     Returns: Alert with full history

   POST   /api/v1/alerts/:id/acknowledge
     Body: {"note": "acknowledged by ops team"}
     Updates: acknowledged, acknowledged_by, acknowledged_at

   DELETE /api/v1/alerts/:id
     Manually resolves alert

   GET    /api/v1/alerts/stats
     Returns: {total, critical, high, medium, low}
   ```

2. **Rule Management APIs**
   ```
   POST   /api/v1/alert-rules
     Body: {name, rule_type, rule_config, severity, enabled}
     Returns: AlertRule with ID

   GET    /api/v1/alert-rules
     Returns: []AlertRule

   GET    /api/v1/alert-rules/:id
     Returns: AlertRule with full config

   PUT    /api/v1/alert-rules/:id
     Body: {name, rule_config, severity, enabled, ...}
     Returns: Updated AlertRule

   DELETE /api/v1/alert-rules/:id

   POST   /api/v1/alert-rules/:id/test
     Body: {test_metric_values}
     Returns: {would_trigger: true/false}
   ```

3. **Notification Channel APIs**
   ```
   POST   /api/v1/notification-channels
     Body: {name, type, config}
     Returns: NotificationChannel

   GET    /api/v1/notification-channels

   PUT    /api/v1/notification-channels/:id

   DELETE /api/v1/notification-channels/:id

   POST   /api/v1/notification-channels/:id/test
     Sends test notification to channel

   GET    /api/v1/notification-channels/:id/history
     Returns delivery history
   ```

**Success Criteria**:
- ✅ All CRUD operations working
- ✅ Filtering and pagination working
- ✅ Test endpoints functional
- ✅ Proper error handling and validation
- ✅ Authentication and authorization enforced

---

#### 5.5 Frontend Alert UI - 40 hours

**Priority**: 🟢 MEDIUM - User experience

**Objective**: Comprehensive alert management dashboard.

**Components**:

1. **Alerts Dashboard** (`/frontend/src/pages/AlertsIncidents.tsx`)
   ```tsx
   Features:
   - Real-time alert list with WebSocket/polling
   - Filter by: severity, status, type, date range
   - Bulk actions: acknowledge, resolve
   - Visual status indicators (red/yellow/green)
   - Sort by: severity, triggered_at, last_change
   - Search by: rule name, metric, database
   ```

2. **Alert Detail Page**
   ```tsx
   - Timeline of state changes
   - Associated metrics graph with anomaly highlight
   - Context data (metric values at trigger time)
   - Runbook links (configurable per rule)
   - Suggested remediation actions
   - Acknowledge form with modal
   - History of previous occurrences
   ```

3. **Alert Rules Management Page**
   ```tsx
   - CRUD interface for rules
   - Rule template library (suggestions)
   - Visual condition builder (drag-drop optional)
   - Rule enablement toggle
   - Test rule form with sample data
   - Rule cloning
   - Import/export rules
   ```

4. **Notification Channel Setup**
   ```tsx
   - Configuration forms per channel type
   - Test delivery button
   - Credential storage UI
   - Delivery history table
   - Enable/disable channels
   - Delete with confirmation
   ```

5. **Real-time Updates**
   ```tsx
   - WebSocket connection for live updates
   - Fallback to polling (15s interval)
   - Connection state indicator
   - Auto-reconnect with exponential backoff
   ```

6. **Analytics Dashboard**
   ```tsx
   - Alert metrics: total, by severity
   - False positive tracking
   - MTTF (Mean Time To Failure)
   - MTTR (Mean Time To Resolution)
   - Rule effectiveness analysis
   ```

**File Structure**:
```
/frontend/src/
  pages/
    AlertsIncidents.tsx       # Dashboard
    AlertDetail.tsx           # Detail page (new)
    AlertRulesManagement.tsx  # Rules management (new)
  components/
    AlertList.tsx             # Reusable list component
    AlertCard.tsx             # Alert card component
    RuleBuilder.tsx           # Visual rule builder
    NotificationChannelForm.tsx
  types/
    alerts.ts                 # Type definitions
  store/
    alertStore.ts             # State management
  hooks/
    useAlerts.ts              # Custom hooks for alerts
    useWebSocket.ts           # WebSocket management
  services/
    alertService.ts           # API client for alerts
```

**Success Criteria**:
- ✅ Dashboard loads in < 2 seconds
- ✅ Real-time updates working smoothly
- ✅ Rule creation/editing/deletion working
- ✅ Notifications delivering successfully
- ✅ Responsive design (mobile + desktop)
- ✅ Performance optimized (virtualization for long lists)

---

### Phase 5 Summary

| Component | Hours | Difficulty | Risk |
|-----------|-------|-----------|------|
| Anomaly Detection | 50 | High | Medium |
| Alert Rules Engine | 40 | Medium | Low |
| Notifications | 45 | Medium | Low |
| APIs + Frontend | 75 | High | Low |
| **Total** | **210** | - | - |

**Team**: 2 developers (4 weeks parallelized)

---

## Overall Summary

### Timeline & Effort

```
PHASE 3 (v3.3.0): 220 hours = 4 weeks (2-3 devs)
  ├─ Week 1-2: Enterprise Auth (80h)
  ├─ Week 2-3: Encryption (60h)
  ├─ Week 3-4: HA/Failover (50h)
  └─ Week 4: Audit (30h)

PHASE 4 (v3.4.0): 130 hours = 4 weeks (1-2 devs)
  ├─ Week 1-2: Backend Optimization (40h)
  ├─ Week 2-3: Collector C++ (60h)
  └─ Week 4: Load Testing (30h)

PHASE 5 (v3.5.0): 210 hours = 4 weeks (2 devs)
  ├─ Week 1-2: Anomaly Detection (50h)
  ├─ Week 2-3: Alert Rules (40h) + Notifications (45h)
  └─ Week 3-4: APIs (35h) + Frontend (40h)

TOTAL: 560 hours
```

### Recommended Staffing

**Option 1**: 3-4 developers (12 weeks)
- Optimal parallelization
- Better quality review
- Lower risk of critical bugs

**Option 2**: 5 developers (8 weeks)
- Aggressive timeline
- Requires good coordination
- Higher risk of integration issues

### Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|-----------|
| LDAP integration breaks login | Critical | Feature flag, local JWT fallback |
| Encryption impacts performance | High | Background encryption, keep old columns |
| HA failover > 2s | High | Pre-test with chaos engineering |
| Collector scaling side effects | Medium | Load test before merge, gradual rollout |
| Anomaly false positives | Medium | Configurable thresholds, manual validation |
| Alert notification storms | Medium | Deduplication and rate limiting |

### Testing Strategy

1. **Unit Tests**: Every new module must have >80% coverage
2. **Integration Tests**: API + Database + External services
3. **Load Tests**: 500 collectors for 8+ hours
4. **E2E Tests**: Complete workflows (anomaly → alert → notify → resolve)
5. **Chaos Engineering**: Failover scenarios, connection failures
6. **Security Tests**: Authentication/authorization, encryption key rotation
7. **Performance Tests**: Latency targets (p95 < 500ms), memory stability

### Deployment Strategy

**Phase 3 (Enterprise Features)**:
1. Deploy auth modules with feature flags disabled
2. Enable LDAP first, then OAuth, then SAML, then MFA
3. Encryption: shadow mode (encrypt new data only) then full migration
4. HA/Failover: test in staging, gradual canary in production
5. Audit: enable on all systems immediately (low risk)

**Phase 4 (Scalability)**:
1. Backend: deploy config changes, monitor impact
2. Collector C++: new version with backward compatibility
3. Load test: validate targets before general release

**Phase 5 (Advanced Analytics)**:
1. Deploy anomaly detector in read-only mode
2. Collect baseline data for 2+ weeks
3. Enable alerts with conservative thresholds
4. Gradually adjust thresholds based on feedback

### Success Metrics

**Phase 3**:
- Enterprise auth options used by customers
- Zero encryption migration issues
- HA failover RTO < 2 seconds
- 100% audit trail completeness

**Phase 4**:
- Support 500+ collectors
- Latency p95 < 500ms
- Memory stable over 24 hours
- No collector registration rejections

**Phase 5**:
- Anomaly detection precision > 90%
- False positive rate < 5%
- 99%+ notification delivery success
- MTTR reduced by 40%+

---

## Task Checklist

### Phase 3 Tasks
- [ ] Task #1: Enterprise Authentication (LDAP/SAML/OAuth/MFA) - 80h
- [ ] Task #2: Encryption at Rest & Key Management - 60h
- [ ] Task #3: High Availability & Failover - 50h
- [ ] Task #4: Audit Logging & Compliance - 30h
- [ ] Task #11: Database migrations - 20h (parallel)

### Phase 4 Tasks
- [ ] Task #5: Backend Scalability - 40h
- [ ] Task #6: Collector C++ Optimization - 60h

### Phase 5 Tasks
- [ ] Task #7: Anomaly Detection Engine - 50h
- [ ] Task #8: Alert Rule Engine - 40h
- [ ] Task #9: Notification Delivery - 45h
- [ ] Task #10: Frontend Alert UI - 40h

### Testing & Integration
- [ ] Task #12: Integration testing for all phases - 80h (throughout)

---

## Next Steps

1. **Finalize Requirements** (Week 1):
   - Stakeholder review of feature priorities
   - Define exact success criteria for each component
   - Identify any dependencies with other teams

2. **Infrastructure Setup** (Week 1-2):
   - Staging environment with PostgreSQL HA
   - Redis Sentinel setup
   - Load testing environment
   - Key management (AWS Secrets Manager / Vault)

3. **Begin Phase 3 Implementation** (Week 2):
   - Start with LDAP (highest priority)
   - Parallel work on Encryption and HA setup

4. **Weekly Progress Tracking**:
   - Technical sync every Monday
   - Risk review every Friday
   - Stakeholder updates every other week

---

## References

- PostgreSQL HA: https://www.postgresql.org/docs/current/warm-standby.html
- Patroni: https://github.com/zalando/patroni
- pg_auto_failover: https://github.com/citusdata/pg_auto_failover
- SAML: https://github.com/crewjam/saml
- OAuth: https://golang.org/x/oauth2
- Key Management: AWS Secrets Manager, HashiCorp Vault, GCP KMS
- Monitoring: Prometheus + Grafana for performance metrics
- Testing: k6 for load testing, chaos-monkey for resilience testing

