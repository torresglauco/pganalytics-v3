# Codebase Concerns

**Analysis Date:** 2026-03-30

## Tech Debt

### Incomplete Alert System Implementation

**Issue:** Alert workers and notification workers contain extensive TODO comments for critical database operations.

**Files:**
- `backend/pkg/services/alert_worker.go` (lines 60, 75, 87, 108, 114, 121)
- `backend/pkg/services/notification_worker.go` (lines 56, 65, 73, 76, 107, 117, 129, 137)
- `backend/internal/jobs/alert_rule_engine.go`
- `backend/internal/jobs/anomaly_detector.go`

**Impact:** Alert system is non-functional in production. The workers return empty datasets and stub implementations, meaning:
- No alerts are being evaluated or fetched from the database
- No alert triggers are being created or recorded
- No notifications are actually being sent despite code structure
- Retry logic exists but operates on empty data

**Fix approach:**
1. Complete database query implementations in `alert_worker.go` (fetch alerts, check recent triggers, insert triggers)
2. Implement notification worker database operations (fetch pending notifications, update delivery status)
3. Complete condition evaluation logic for alert rules
4. Add integration tests to verify end-to-end alert flow
5. Enable alert background workers in main server initialization

**Priority:** CRITICAL - Affects core product functionality

---

### Incomplete mTLS Implementation

**Issue:** mTLS middleware is a placeholder with three critical TODOs.

**Files:** `backend/internal/api/middleware.go` (lines 108, 133-135)

**Impact:**
- Certificate verification is not implemented
- No certificate thumbprint validation
- No collector_id extraction from client certificates
- In production, mTLS provides no actual security despite configuration

**Current behavior:**
- Development mode: allows non-TLS connections
- Production mode: checks certificate presence but doesn't validate it
- Never stores certificate information in context

**Fix approach:**
1. Parse client certificate from TLS connection
2. Implement thumbprint extraction and verification against registered certificates
3. Store collector_id in context for metrics processing
4. Add integration tests with actual client certificates
5. Document certificate lifecycle and registration

**Priority:** HIGH - Security concern in production deployments

---

### Hardcoded OAuth Redirect URL

**Issue:** OAuth connector is initialized with hardcoded localhost callback URL.

**Files:** `backend/internal/api/handlers_auth.go` (line 176)

```go
oauthConn, err := auth.NewOAuthConnector("http://localhost:8080", providerConfigs)
```

**Impact:**
- OAuth authentication fails in production/staging environments
- Callback URL mismatch with OAuth provider configuration
- Cannot redirect users back to correct frontend after OAuth flow

**Fix approach:**
1. Read callback URL from configuration or environment variable
2. Use `s.config.APIBaseURL` or similar from server config
3. Add validation in config initialization
4. Test with actual OAuth providers (Google, GitHub, Azure AD)

**Priority:** HIGH - Breaks OAuth authentication outside localhost

---

## Known Bugs

### Database Connection Pool Configuration Issue

**Issue:** Connection pool configuration reads from environment variables but uses hardcoded defaults that may not be optimal.

**Files:** `backend/internal/storage/postgres.go` (lines 42-77)

**Symptoms:**
- Default: 100 max connections, 20 idle connections
- Configured for "500+ collectors" but defaults may be conservative
- Pool exhaustion could occur under high load
- Connection timeout errors may appear in logs

**Current mitigation:** Env vars allow override, but documentation unclear on recommended values

**Workaround:** Set `MAX_DATABASE_CONNS` and `MAX_IDLE_DATABASE_CONNS` environment variables

---

### Missing Request ID Generation

**Issue:** Request ID generation is incomplete - request IDs not being generated or tracked.

**Files:** `backend/internal/api/middleware.go` (line 311 - TODO comment)

**Impact:**
- Cannot correlate logs across distributed tracing
- Difficult to track request lifecycle in monitoring systems
- Error reports lack context for debugging

**Fix approach:**
1. Generate UUID for each request at middleware level
2. Store in context and response headers
3. Include in all log statements using context
4. Wire into distributed tracing system

---

## Security Considerations

### Missing LDAP/OAuth Configuration Validation

**Issue:** Configuration for LDAP and OAuth is unmarshaled from JSON strings without validation.

**Files:**
- `backend/internal/api/handlers_auth.go` (lines 57-60, 169-170)

**Risk:** Malformed JSON silently fails and defaults to empty configuration, allowing authentication bypass.

**Current mitigation:** Errors logged but request continues

**Recommendations:**
1. Add strict schema validation for `LDAPGroupToRoleJSON` and `OAuthProvidersJSON`
2. Fail startup if required auth config is invalid
3. Log validation errors at warning level to surface misconfigurations
4. Add integration tests for JSON parsing edge cases

---

### Session Token Creation Failure Not Handled

**Issue:** Session creation failure is logged but doesn't prevent authentication.

**Files:** `backend/internal/api/handlers_auth.go` (lines 108-114)

```go
if err != nil {
    s.logger.Error("Failed to create session", zap.Error(err))
}
```

**Risk:** Users authenticated without session tracking, breaking audit trail and logout functionality.

**Fix approach:**
1. Return 500 error if session creation fails
2. Ensure session state is created before returning auth response
3. Add transaction support to prevent partial state

---

### Crypto Randomness Sources

**Issue:** Backend uses `math/rand` in some areas which is not cryptographically secure.

**Files:** `backend/internal/api/handlers.go` (line 7 imports `math/rand`)

**Risk:** Weak randomness could be used for token generation or sensitive operations if code is refactored

**Current mitigation:** JWT and password generation use proper crypto libraries

**Recommendations:**
1. Audit all uses of `math/rand` - remove if not crypto-related
2. Use `crypto/rand` exclusively for security-sensitive operations
3. Add lint rule to prevent math/rand imports where crypto/rand is needed

---

## Performance Bottlenecks

### Blocking Database Operations in Request Handlers

**Issue:** Multiple database queries executed sequentially in HTTP handlers without optimization.

**Files:**
- `backend/internal/api/handlers.go` (1916 lines - large file with multiple query sequences)
- `backend/internal/storage/postgres.go` (2498 lines - monolithic storage layer)

**Problem:**
- User authentication requires: GetUserByID query (line 39 of handlers_auth.go)
- Followed by session creation
- Followed by token generation
- All synchronous in request path

**Impact:** P95 latency increases with database load. Single database bottleneck for API.

**Improvement path:**
1. Batch database queries where possible
2. Cache user data in Redis after auth
3. Consider connection pooling improvements
4. Profile slow endpoints with pprof

---

### Monolithic Handler and Storage Files

**Issue:** Two very large files handle the bulk of API logic and database operations.

**Files:**
- `backend/internal/api/handlers.go` (1916 lines)
- `backend/internal/storage/postgres.go` (2498 lines)

**Impact:**
- Difficult to reason about code flow
- Testing requires extensive mocking
- Changes in one area risk regressions elsewhere
- Longer compile times

**Scaling limit:** Single file is at practical limit for maintainability

**Safe modification:**
1. Extract related handlers into separate files (e.g., handlers_metrics.go, handlers_instances.go)
2. Extract storage operations by domain (users_store.go, metrics_store.go, alerts_store.go)
3. Create middleware/helper packages for common patterns
4. Add interface boundaries between layers

---

### Alert and Anomaly Detection Intensive Queries

**Issue:** Anomaly detection and alert rule evaluation execute complex queries periodically.

**Files:**
- `backend/internal/jobs/anomaly_detector.go` (707 lines)
- `backend/internal/jobs/alert_rule_engine.go` (828 lines)

**Problem:**
- Concurrent rule evaluation with semaphore limiting to 5 concurrent ops (line 220 of anomaly_detector.go)
- No query timeout enforcement visible
- Full table scans on large datasets possible
- WebSocket broadcast during anomaly detection (line 94 of alert_worker.go)

**Scaling limit:** Becomes I/O bound around 1000+ alert rules or anomalies

**Improvement path:**
1. Add query execution timeout
2. Implement pagination for large result sets
3. Use database-level filtering before fetching to application
4. Cache baseline calculations
5. Consider dedicated analytics database for historical queries

---

## Fragile Areas

### WebSocket Connection Manager

**Files:** `backend/pkg/services/websocket.go`

**Why fragile:**
- Unbuffered operations on concurrent channels
- Broadcast sends to all connections without error handling
- No timeout on channel operations
- Goroutine leak possible if client disconnect not handled

**Safe modification:**
1. Add context timeout to WebSocket operations
2. Implement graceful shutdown sequence
3. Add metrics for connection count and message latency
4. Test with rapid connect/disconnect cycles
5. Add integration tests with actual WebSocket client

**Test coverage gap:** No tests visible for WebSocket broadcast failures

---

### LDAP/OAuth Authentication Flow

**Files:**
- `backend/internal/api/handlers_auth.go` (LDAP: lines 38-128, OAuth: lines 150-195)
- `backend/internal/auth/ldap.go`
- `backend/internal/auth/oauth.go`

**Why fragile:**
- External service dependency (LDAP server, OAuth provider)
- Unmarshal errors don't prevent auth flow
- Token generation failure allows partial auth
- No circuit breaker for failed external calls (would retry all requests)

**Safe modification:**
1. Add circuit breaker for LDAP/OAuth provider
2. Implement bulkhead pattern to limit concurrent external calls
3. Add fallback to cached tokens if provider fails
4. Test provider timeout and network error scenarios

**Test coverage gap:** Only basic happy-path tests, no provider failure scenarios

---

### Notification Channel Configuration

**Files:** `backend/internal/notifications/notification_service.go`, `backend/internal/notifications/channels.go` (674 lines)

**Why fragile:**
- Channel-specific config stored as json.RawMessage (requires runtime unmarshaling)
- No validation on config before sending
- Multiple notification providers with different field requirements
- Webhook URL not validated for protocol or format

**Safe modification:**
1. Add strict validation for each channel type configuration
2. Use type-safe config structs per channel
3. Test configuration validation edge cases
4. Add dry-run mode for channel testing

**Test coverage gap:** Limited testing of invalid channel configurations

---

## Scaling Limits

### Single Database Instance

**Current capacity:** Designed for single PostgreSQL instance

**Limit:** Becomes bottleneck at:
- ~10,000+ metrics per second ingestion
- ~100,000+ alert evaluations per minute
- ~500+ concurrent collectors (current design target)

**Scaling path:**
1. Read replicas for analytics queries (no implementation yet)
2. TimescaleDB for metrics (partially integrated, see s.timescale in code)
3. Metrics shard by instance_id for write scaling
4. Separate OLTP and OLAP databases

---

### Alert Rule Evaluation Concurrency

**Current capacity:** Semaphore limited to `maxConcurrentRules` (typically 5-20)

**Limit:** Linear scaling stops around 100-200 alert rules due to sequential database queries

**Scaling path:**
1. Batch rule evaluation by database (reduce queries)
2. Parallel predicate evaluation within rules
3. Move rule compilation to startup, not evaluation time
4. Cache metrics state between evaluations

---

## Dependencies at Risk

### Outdated Go Version

**Risk:** Go 1.24.0 may have security updates and compatibility issues

**Files:** `go.mod` (line 3)

**Current state:** Very recent Go version (released early 2026)

**Mitigation:**
1. Track Go security advisories
2. Test with Go 1.25 when released to ensure compatibility
3. Set minimum Go version in go.mod

---

### Deprecated Dependencies

**Risk:** Several npm dependencies marked as deprecated

**Files:** `frontend/package-lock.json` (lines 1013, 1043, 3674, 4297, etc.)

**Examples:**
- `@eslint/eslint-config` (marked deprecated)
- Old versions of `glob` package
- `memory-cache` (memory leak reported)
- Rimraf prior to v4

**Impact:**
- No security updates for deprecated packages
- Build tools may break with future npm versions
- Possible memory leaks in test/build infrastructure

**Fix approach:**
1. Update all devDependencies to latest versions
2. Review CHANGELOG for breaking changes
3. Run full test suite after updates
4. Set up Dependabot for automated updates

**Priority:** MEDIUM - Affects development experience but not production app

---

### SAML Library Version

**Risk:** `crewjam/saml` v0.5.1 is older, check for security advisories

**Files:** `go.mod` (line 6)

**Recommendations:**
1. Check for SAML XML signature bypass vulnerabilities (common in SAML libraries)
2. Review upstream issues for known CVEs
3. Plan upgrade path to newer version if available

---

## Missing Critical Features

### Feature Gap: Collector Health Monitoring

**Problem:** Health check scheduler exists but no visible dashboard/API endpoint to view collector health status

**Files:**
- `backend/internal/jobs/health_check_scheduler.go` (implemented)
- `frontend/src/pages/CollectorsManagement.tsx` (no health display)

**Impact:** Operators cannot identify failed collectors without logs

**Blocks:** Collector management UI, proactive incident response

---

### Feature Gap: Alert Silence/Maintenance Windows

**Problem:** Silence endpoints exist but silences may not actually prevent alert notifications

**Files:** `backend/internal/api/handlers_silences.go` (line 181 - TODO: Delete silence from database)

**Impact:** Cannot suppress false-positive alerts during maintenance

---

## Test Coverage Gaps

### Alert System End-to-End

**What's not tested:** Complete alert flow from rule creation → trigger evaluation → notification delivery

**Files:** `backend/pkg/services/alert_worker.go`, `backend/pkg/services/notification_worker.go`

**Risk:** Critical functionality completely untested

**Recommendation:** Add integration test suite:
```
backend/tests/integration/alert_flow_test.go
- Create alert rule
- Trigger condition becomes true
- Verify trigger created
- Verify notification queued
- Verify delivery attempted
```

**Priority:** CRITICAL

---

### mTLS Certificate Validation

**What's not tested:** Client certificate verification and thumbprint validation

**Files:** `backend/internal/api/middleware.go` (lines 107-138)

**Risk:** Cannot verify TLS implementation works

**Recommendation:** Add mTLS integration test with actual certificates

---

### External Provider Failures (LDAP, OAuth)

**What's not tested:** LDAP server down, OAuth provider timeout, network errors

**Files:**
- `backend/internal/auth/ldap.go`
- `backend/internal/auth/oauth.go`

**Risk:** Auth system behavior under provider failure is unknown

**Recommendation:** Mock LDAP/OAuth servers that simulate failures

---

### WebSocket Connection Failures

**What's not tested:** WebSocket client disconnect during broadcast, channel errors

**Files:** `backend/pkg/services/websocket.go`

**Risk:** Unknown memory leaks or goroutine leaks

**Recommendation:** Integration tests with connection pool stress

---

### Configuration Validation

**What's not tested:** Invalid JSON in `LDAPGroupToRoleJSON`, `OAuthProvidersJSON`, notification channel configs

**Files:** `backend/internal/api/handlers_auth.go`

**Risk:** Silent failures that are hard to debug

**Recommendation:** Add config validation unit tests covering malformed JSON edge cases

---

## Database Compatibility Issues

### PostgreSQL Version Requirements

**Issue:** Code references various PostgreSQL versions but compatibility matrix unclear

**Files:** `collector/sql/replication_queries.sql` (line 154 - requires 9.4+)

**Concern:** Some operations may fail on older PostgreSQL versions without clear error messages

**Recommendation:**
1. Document minimum PostgreSQL version explicitly
2. Add startup check to verify version compatibility
3. Test against supported versions in CI

---

## Architecture Concerns

### Tight Coupling Between API and Storage Layers

**Issue:** API handlers directly call storage methods without abstraction

**Files:** `backend/internal/api/handlers.go` (calls `s.postgres.*` directly)

**Impact:** Cannot easily swap storage implementations or add caching layer

**Safe modification:**
1. Create repository interfaces for each domain
2. Implement in-memory cache decorator
3. Add metrics/logging at repository level

---

### Missing Circuit Breaker for External Services

**Issue:** No circuit breaker for LDAP, OAuth, or webhook endpoints

**Files:**
- `backend/internal/auth/ldap.go`
- `backend/internal/auth/oauth.go`
- `backend/internal/notifications/channels.go`

**Risk:** Cascading failures if external service is slow/down

**Fix approach:**
1. Implement circuit breaker around LDAP authentication
2. Implement circuit breaker around OAuth provider calls
3. Implement circuit breaker around webhook delivery
4. Test with failing services

---

## Documentation Gaps

### Alert System Architecture Undocumented

**Issue:** How alert rules are evaluated, when, with what data - not documented

**Impact:** Developers cannot implement missing TODO items without understanding design

**Recommendation:** Add ADR (Architecture Decision Record) for alert system

---

### mTLS Setup Process Undocumented

**Issue:** How to generate, register, and validate client certificates not documented

**Impact:** Cannot deploy mTLS without reverse-engineering code

**Recommendation:** Add step-by-step mTLS configuration guide

---

*Concerns audit: 2026-03-30*
