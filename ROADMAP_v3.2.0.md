# pgAnalytics-v3 v3.2.0 Roadmap

**Planned Release:** Q2 2026 (April-June 2026)  
**Status:** Planning Phase  
**Type:** Performance Optimization & Enhanced Monitoring

---

## Overview

v3.2.0 focuses on addressing the critical performance bottlenecks identified in the v3.1.0 load testing phase. This release will implement batch query processing, optimize JSON serialization, enhance monitoring capabilities, and prepare for multi-collector scalability.

**Key Goals:**
- Resolve critical performance bottlenecks from 90% data loss issue
- Improve metrics processing capacity from 100 to 1000+ queries per collection cycle
- Reduce CPU overhead by 40% through JSON serialization optimization
- Add comprehensive audit logging for security monitoring
- Implement mTLS for collector communication (Phase 2)

---

## Performance Optimization (CRITICAL)

### 1. Implement Batch Query Processing

**Problem Identified in v3.1.0:**
- Sequential query processing causes bottleneck
- Each query = 1 database round-trip
- Baseline: 285-870ms per 60-second cycle

**Solution:**
- Use `pgx.Batch` API for concurrent execution
- Group queries into batches of 10-50
- Expected improvement: 3-5x faster processing

**Implementation Details:**
```go
// Instead of:
for _, queryInfo := range db.Queries {
    InsertQueryStats(...)  // 1 round-trip per query
}

// Use batch:
batch := &pgx.Batch{}
for _, queryInfo := range db.Queries {
    batch.Queue("INSERT INTO query_stats ...", ...)
}
results := conn.SendBatch(ctx, batch)  // Single round-trip for all
```

**Acceptance Criteria:**
- ✅ Batch query processing implemented
- ✅ 3-5x performance improvement verified
- ✅ Load test confirms 150-250ms per cycle (from 285-870ms)
- ✅ No data loss at 500 queries

**Effort:** 6-8 hours

---

### 2. Remove Hard-Coded 100-Query Limit

**Problem Identified in v3.1.0:**
- Hard-coded limit causes 90% data loss at scale
- Silent metric discarding without visibility
- No operator awareness of lost data

**Solution:**
- Make query limit configurable via environment variable
- Increase default from 100 to 1000 queries
- Add metrics for discarded queries
- Implement alert when discard rate exceeds threshold

**Configuration:**
```bash
# Environment variables (new in v3.2.0)
export MAX_QUERIES_PER_DB=1000          # Default: 1000
export QUERY_LIMIT_ALERT_THRESHOLD=0.05 # Alert if >5% discarded
```

**Implementation:**
- Update collector C++ code to read from config
- Log discarded query count per collection cycle
- Expose metrics via `/api/v1/metrics/health`
- Alert integration with monitoring system

**Acceptance Criteria:**
- ✅ Query limit configurable (environment variable)
- ✅ Default increased to 1000 queries
- ✅ Discarded query metrics exposed
- ✅ Alert triggers when >5% data loss detected
- ✅ No silent data loss

**Effort:** 4-6 hours

---

### 3. Optimize JSON Serialization

**Problem Identified in v3.1.0:**
- Double/triple JSON serialization overhead
- 30-50% CPU wasted on redundant operations
- Metrics serialized → deserialized → reserialized

**Current Flow (INEFFICIENT):**
```go
metricsJSON, _ := json.Marshal(metric)        // Serialize 1
var singleDB models.QueryStatsDB
json.Unmarshal(metricsJSON, &singleDB)        // Deserialize
// Process...
json.Marshal(...)                              // Serialize 2
```

**Optimized Flow:**
```go
// Parse once, keep as structured data
var singleDB models.QueryStatsDB
json.Unmarshal(reqBytes, &singleDB)           // Single deserialization
// Process & insert directly
InsertQueryStats(singleDB)
```

**Expected Improvements:**
- 40% CPU reduction in JSON operations
- Single pass through metric data
- Faster handler processing

**Acceptance Criteria:**
- ✅ Double serialization removed
- ✅ 40% CPU improvement in JSON operations verified
- ✅ Load test confirms improvement
- ✅ Handler response time <150ms

**Effort:** 3-4 hours

---

### 4. Tune Connection Pool

**Problem Identified in v3.1.0:**
- MaxDatabaseConns: 50 (too small for concurrent collectors)
- Response time degradation with 5+ parallel collectors
- Potential connection exhaustion

**Current Config:**
```go
MaxDatabaseConns: 50
MaxIdleDatabaseConns: 15
```

**New Config (v3.2.0):**
```go
MaxDatabaseConns: 200        // 4x increase
MaxIdleDatabaseConns: 50     // 3x increase
MaxConnLifetime: 5 * time.Minute
ConnectionTimeout: 5 * time.Second
```

**Rationale:**
- pgx.Batch batches queries, reducing unique connection needs
- 200 connections supports 10+ concurrent collectors
- Idle connections kept warm for faster response
- Lifecycle management prevents stale connections

**Testing Plan:**
- Baseline: 5 collectors × 100 queries each
- Expected: <150ms response time (vs 150-540ms in v3.1.0)
- Stress test: 10 collectors × 100 queries

**Acceptance Criteria:**
- ✅ Connection pool tuned
- ✅ 10+ concurrent collectors supported
- ✅ Response time <150ms consistent
- ✅ No connection exhaustion errors
- ✅ Memory usage reasonable

**Effort:** 2 hours (testing: 4 hours)

---

## Monitoring & Observability

### 5. Implement Comprehensive Audit Logging

**OWASP Gap Identified in v3.1.0:**
- Logging/monitoring scored only partial (TODO)
- No audit trail for sensitive operations
- Difficult to detect security incidents

**Audit Events to Log:**
1. **Authentication Events**
   - Successful login (user_id, timestamp)
   - Failed login attempts (username, IP, timestamp)
   - Token refresh (user_id, timestamp)
   - Password change (user_id, timestamp)

2. **Authorization Events**
   - Permission denied (user_id, endpoint, timestamp)
   - Role change (user_id, old_role, new_role, admin_id)
   - Access control violation (user_id, resource, timestamp)

3. **Configuration Events**
   - Collector registered (collector_id, hostname, timestamp)
   - Collector config updated (collector_id, updated_by, timestamp)
   - API key created/revoked (user_id, timestamp)

4. **Security Events**
   - Rate limit exceeded (client_id, endpoint, timestamp)
   - Authentication bypass attempt (IP, timestamp)
   - Certificate renewal (collector_id, timestamp)
   - Security configuration changed (admin_id, change, timestamp)

**Implementation:**
- Use PostgreSQL audit table: `audit_log`
- Log level: INFO for routine, WARN for failures, ERROR for violations
- Structured logging with JSON format
- Queryable via API: `GET /api/v1/audit-logs`

**Acceptance Criteria:**
- ✅ Audit logging implemented for all sensitive operations
- ✅ 10+ audit event types tracked
- ✅ OWASP A9 (Logging) requirement met
- ✅ Query audit logs API endpoint working
- ✅ Retention policy implemented (90 days default)

**Effort:** 8-10 hours

---

### 6. Add Real-Time Security Alerting

**New Feature:**
- Alert on suspicious activity patterns
- Integration with monitoring systems (Prometheus, AlertManager)
- Configurable alert thresholds

**Alert Rules:**
1. **Failed Authentication Surge** - >10 failed logins in 5 minutes
2. **Rate Limit Abuse** - Same IP exceeds limit 50+ times in 5 minutes
3. **Unusual Access Patterns** - User accessing 20+ different endpoints in 1 minute
4. **Collector Disconnection** - Collector offline >5 minutes
5. **Configuration Changes** - Any config change triggers alert to admins
6. **Token Expiration** - Alert when certificate expires in <7 days

**Implementation:**
- Prometheus metrics for alert rule engine
- AlertManager integration for notifications
- Webhook support for custom integrations
- Admin dashboard for alert management

**Acceptance Criteria:**
- ✅ 6 alert rules configured
- ✅ Prometheus metrics exposed
- ✅ AlertManager integration working
- ✅ Admin dashboard for alert management
- ✅ Notifications sent correctly

**Effort:** 10-12 hours

---

## Security Enhancements

### 7. Implement mTLS for Collectors (Phase 2)

**Current State (v3.1.0):**
- JWT tokens only
- No certificate validation
- Placeholder MTLSMiddleware

**v3.2.0 Goals:**
- Full mTLS handshake for collectors
- Certificate pinning
- Automated certificate rotation
- CRL (Certificate Revocation List) support

**Implementation Plan:**
1. Generate CA certificate (self-signed for dev, proper CA for prod)
2. Issue collector certificates with unique serial numbers
3. Implement mTLS middleware with proper validation
4. Add certificate rotation logic (30-day renewal)
5. Implement revocation checking

**Configuration:**
```bash
export TLS_CA_CERT="/etc/pganalytics/ca.crt"
export TLS_CERT_PATH="/etc/pganalytics/server.crt"
export TLS_KEY_PATH="/etc/pganalytics/server.key"
export CERT_ROTATION_DAYS=30
export ENABLE_MTLS=true
```

**Acceptance Criteria:**
- ✅ mTLS handshake working
- ✅ Certificate validation enforced
- ✅ Certificate rotation automated
- ✅ CRL checking implemented
- ✅ Collector requires valid certificate

**Effort:** 12-14 hours

---

## Test Coverage Enhancement

### 8. Expand Integration & Load Testing

**Current State (v3.1.0):**
- Manual load testing scripts
- Limited integration tests
- No performance regression suite

**v3.2.0 Improvements:**
1. **Automated Performance Tests**
   - Baseline performance tracking
   - Regression detection (fail if >10% slower)
   - CI/CD integration

2. **Extended Load Testing**
   - 10+ collector stress test
   - 5000+ queries per cycle
   - 24-hour soak test

3. **Integration Tests**
   - End-to-end security tests
   - RBAC validation
   - Rate limiting verification

4. **Benchmark Suite**
   - Batch processing benchmark
   - JSON serialization benchmark
   - Connection pool efficiency

**Acceptance Criteria:**
- ✅ 50+ integration tests passing
- ✅ Performance regression suite in CI/CD
- ✅ Load test suite automated
- ✅ Benchmark baseline established
- ✅ All tests documented

**Effort:** 8-10 hours

---

## Documentation Updates

### 9. Enhance Security & Deployment Documentation

**Improvements:**
1. **Operational Security Guide**
   - Certificate management procedures
   - Key rotation procedures
   - Incident response playbooks

2. **Performance Tuning Guide**
   - Connection pool configuration
   - Query batch size optimization
   - Memory/CPU trade-offs

3. **Monitoring & Alerting Guide**
   - Setting up alerts
   - Interpreting metrics
   - Troubleshooting guide

4. **Migration Guide**
   - v3.1.0 → v3.2.0 upgrade path
   - Breaking changes (none planned)
   - Rollback procedure

**Acceptance Criteria:**
- ✅ Operations guide completed
- ✅ Performance tuning documented
- ✅ Migration guide created
- ✅ All guides reviewed
- ✅ Examples provided

**Effort:** 6-8 hours

---

## Release Metrics & Goals

### Performance Targets

| Metric | v3.1.0 | v3.2.0 Goal | Improvement |
|--------|--------|------------|-------------|
| Single Collector (100 queries) | 85ms avg | <80ms | 6% |
| Scale (1000 queries) | 90% loss | 0% loss | 100% |
| Multi-Collector (5×100) | 150-540ms | <150ms | 70% |
| CPU Overhead | 5-15% | 3-8% | 40% |
| Max Queries/Cycle | 100 | 1000+ | 10x |
| Concurrent Collectors | 3-4 | 10+ | 3x |

### Security Goals

- ✅ OWASP Top 10: 8/10 → 9/10 (logging/monitoring added)
- ✅ mTLS implementation (Phase 2 begins)
- ✅ Audit logging: 100% coverage
- ✅ Real-time alerting: 6 rule types
- ✅ Certificate lifecycle management

### Code Quality Goals

- ✅ Unit tests: 80%+ coverage
- ✅ Integration tests: 50+ test cases
- ✅ Load tests: Automated regression detection
- ✅ Code review: All security concerns addressed
- ✅ Documentation: Comprehensive (security, ops, performance)

---

## Timeline

### Estimated Duration: 8-10 weeks

**Phase 1: Performance Optimization (Weeks 1-3)**
- Batch query processing
- Hard-coded limit removal
- JSON serialization optimization
- Connection pool tuning
- Performance testing

**Phase 2: Monitoring & Alerting (Weeks 4-5)**
- Audit logging implementation
- Real-time alerting setup
- Prometheus metrics
- Dashboard updates

**Phase 3: Security & Testing (Weeks 6-8)**
- mTLS implementation begins (Phase 2)
- Extended test suite
- Performance regression suite
- Documentation updates

**Phase 4: Testing & Release (Weeks 8-10)**
- Full system testing
- Load testing validation
- Documentation review
- Release preparation
- v3.2.0 release

---

## Dependencies & Prerequisites

### v3.1.0 Requirements
- ✅ All security vulnerabilities fixed
- ✅ RBAC implemented
- ✅ Rate limiting active
- ✅ Security headers present

### External Dependencies
- PostgreSQL 13+ (pgx.Batch support)
- Prometheus (for metrics)
- AlertManager (for alerting)
- OpenSSL (for mTLS)

### Team Capacity
- 1-2 backend engineers
- 1 DevOps/SRE engineer
- 1 QA/testing engineer

---

## Success Criteria

### Must Have (v3.2.0 Release)
- ✅ Batch query processing implemented
- ✅ 0% data loss at 1000 queries
- ✅ Query limit configurable
- ✅ Audit logging working
- ✅ All performance targets met

### Should Have
- ✅ Real-time alerting
- ✅ Prometheus metrics
- ✅ Performance regression suite
- ✅ Extended documentation

### Nice to Have
- mTLS implementation (Phase 2 continuation)
- Advanced monitoring dashboards
- CLI tools for operational tasks

---

## Risk Assessment

### Technical Risks

**Risk: Batch Processing Regression**
- Probability: Low
- Impact: High (data loss)
- Mitigation: Comprehensive testing, rollback plan

**Risk: Performance Regression**
- Probability: Medium
- Impact: Medium (slower operations)
- Mitigation: Benchmark suite, CI/CD checks

**Risk: mTLS Complexity**
- Probability: Medium
- Impact: High (certificate management)
- Mitigation: Phased approach, documentation

### Mitigation Strategies
1. Comprehensive test coverage (80%+ unit, 50+ integration)
2. Staged rollout (dev → staging → production)
3. Monitoring & alerting for early detection
4. Rollback plan for each component

---

## Known Issues & Technical Debt

### From v3.1.0 Load Testing
1. Hard-coded 100-query limit (CRITICAL) → **Addressed in v3.2.0**
2. Sequential query processing → **Addressed in v3.2.0**
3. JSON serialization overhead → **Addressed in v3.2.0**
4. Connection pool too small → **Addressed in v3.2.0**

### Deferred to v3.3.0+
1. mTLS full implementation (Phase 2 continuation)
2. Advanced ML-based anomaly detection
3. API key authentication
4. Data encryption at rest

---

## Comparison: v3.1.0 vs v3.2.0

| Feature | v3.1.0 | v3.2.0 |
|---------|--------|--------|
| Security Vulnerabilities Fixed | 6 | 6 (maintained) |
| Max Queries/Cycle | 100 | 1000+ |
| Processing Latency | 85-540ms | <80-150ms |
| Audit Logging | None | Comprehensive |
| Real-Time Alerting | None | 6 rule types |
| mTLS | Placeholder | Phase 2 begins |
| OWASP Top 10 | 8/10 PASS | 9/10 PASS |
| Load Test Coverage | Baseline | Automated regression |
| Documentation | 3,200+ lines | 4,000+ lines |

---

## Contact & Questions

**v3.2.0 Planning Lead:** Backend Team  
**Security Lead:** Security Team  
**Release Manager:** DevOps Team

For questions or concerns about the v3.2.0 roadmap, please open an issue or contact the team.

---

## Appendix: Detailed Implementation Specifications

### Batch Processing Implementation Pseudocode

```go
// v3.2.0: Batch query processing
func (s *Server) handleMetricsPush(c *gin.Context) {
    // ... authentication & validation ...

    // Create batch
    batch := &pgx.Batch{}
    insertCount := 0

    for _, queryInfo := range db.Queries {
        stat := &models.QueryStats{
            // ... populate from queryInfo ...
        }
        
        // Queue in batch instead of executing immediately
        batch.Queue(
            "INSERT INTO query_stats (...) VALUES (...)",
            stat.Time, stat.CollectorID, // ... all fields ...
        )
        insertCount++
    }

    // Execute entire batch in single round-trip
    results := conn.SendBatch(ctx, batch)
    defer results.Close()
    
    // Process results
    for i := 0; i < insertCount; i++ {
        _, err := results.Exec()
        if err != nil {
            s.logger.Error("Batch insert failed", zap.Error(err))
        }
    }

    // Expected: 3-5x faster than sequential processing
}
```

### Environment Variables (v3.2.0)

```bash
# Performance tuning
MAX_QUERIES_PER_DB=1000
BATCH_SIZE=50
MAX_DATABASE_CONNS=200
MAX_IDLE_DATABASE_CONNS=50

# Monitoring & alerting
ENABLE_AUDIT_LOGGING=true
AUDIT_LOG_RETENTION_DAYS=90
ALERT_FAILED_LOGIN_THRESHOLD=10
ALERT_RATE_LIMIT_THRESHOLD=50

# Security
ENABLE_MTLS=true
CERT_ROTATION_DAYS=30
TLS_CA_CERT=/etc/pganalytics/ca.crt
```

---

**Document Version:** 1.0  
**Last Updated:** February 24, 2026  
**Status:** Planning Phase - Ready for Implementation
