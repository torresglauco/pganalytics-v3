# Requirements: pganalytics-v3

**Defined:** 2026-04-28 (v1.1), 2026-05-11 (v1.2), 2026-05-13 (v1.3)
**Core Value:** Enable database teams to proactively identify and fix performance issues before they impact production systems.

---

## v1.3 Requirements (Monitoring & Alerting Platform)

Requirements for comprehensive monitoring and alerting platform. Extends existing collector and backend capabilities.

### Replication Monitoring

- [x] **REP-01**: User can view streaming replication status with write/flush/replay lag metrics
- [ ] **REP-02**: User can view logical replication subscriptions and publications
- [ ] **REP-03**: User can view cascading replication topology (primary → standby → standby)
- [x] **REP-04**: User can view replication slots with WAL retention and active status
- [ ] **REP-05**: User can view replication lag alerts when thresholds exceeded
- [ ] **REP-06**: User can view replication topology graph visualization

### Host Monitoring

- [ ] **HOST-01**: User can view host up/down status for all monitored instances
- [ ] **HOST-02**: User can view OS metrics (CPU, memory, disk, network I/O)
- [ ] **HOST-03**: User can view host inventory (OS version, hardware, PostgreSQL config)
- [x] **HOST-04**: User can view host health score based on resource utilization
- [ ] **HOST-05**: User can configure host-level alert thresholds

### Database Inventory

- [x] **INV-01**: User can view complete table inventory with row counts and sizes
- [x] **INV-02**: User can view column inventory with data types and nullability
- [x] **INV-03**: User can view index inventory with usage statistics
- [x] **INV-04**: User can view extension inventory with versions
- [x] **INV-05**: User can track schema changes over time

### Data Classification

- [x] **DATA-01**: User can view PII detection results (CPF, CNPJ, email, phone, names)
- [x] **DATA-02**: User can view PCI detection results (credit card numbers)
- [x] **DATA-03**: User can view LGPD/GDPR regulated data identification
- [x] **DATA-04**: User can configure custom detection patterns
- [x] **DATA-05**: User can view data classification reports by database/table

### Alerting System

- [ ] **ALERT-01**: User can configure alert rules based on metric thresholds
- [ ] **ALERT-02**: User can receive email notifications for alerts
- [ ] **ALERT-03**: User can receive Slack notifications via webhook
- [ ] **ALERT-04**: User can configure generic webhooks for alert notifications
- [ ] **ALERT-05**: User can integrate with PagerDuty/OpsGenie for incident management
- [ ] **ALERT-06**: User can view alert history with timestamps
- [ ] **ALERT-07**: User can acknowledge and silence alerts
- [ ] **ALERT-08**: User can configure alert escalation policies

### Multi-Version Support

- [ ] **VER-01**: System supports PostgreSQL 13, 14, 15, 16, 17 (actively supported)
- [ ] **VER-02**: System supports PostgreSQL 11, 12 (EOL but critical for migration)
- [ ] **VER-03**: User can view version-specific health checks
- [ ] **VER-04**: System adapts queries based on PostgreSQL version

### Scalability

- [x] **SCALE-01**: System supports 2000+ PostgreSQL clusters
- [x] **SCALE-02**: System supports 5000+ monitored hosts
- [x] **SCALE-03**: System supports sharding/partitioning by tenant/cluster
- [x] **SCALE-04**: System supports multi-tenancy with logical isolation

### Collector Architecture

- [ ] **COLL-01**: Collector can run decentralized (same host as PostgreSQL)
- [ ] **COLL-02**: Collector can run centralized (remote connection to RDS/cloud)
- [ ] **COLL-03**: System supports mixed deployment (decentralized + centralized)
- [ ] **COLL-04**: Collector has low resource footprint for co-location with PostgreSQL
- [ ] **COLL-05**: Collector uses secure communication (TLS, authentication)

### Frontend

- [ ] **UI-01**: User can view replication topology graph
- [ ] **UI-02**: User can configure alert rules via UI
- [ ] **UI-03**: User can view data classification reports
- [ ] **UI-04**: User can view host inventory dashboards
- [ ] **UI-05**: User can manage notification channels

### Testing

- [ ] **TEST-01**: All new collector plugins have C++ unit tests
- [ ] **TEST-02**: All new backend services have Go unit tests
- [ ] **TEST-03**: All new API endpoints have integration tests
- [ ] **TEST-04**: All new frontend components have tests
- [ ] **TEST-05**: End-to-end tests cover critical user flows

---

## v1.2 Requirements (Performance Optimization) ✓ Complete

Requirements for Performance Optimization milestone. Each maps to roadmap phases.

### Query Optimization

- [x] **QRY-01**: User can view top N slow queries by mean_time from pg_stat_statements
- [x] **QRY-02**: User can see query performance timeline with historical trends
- [x] **QRY-03**: User receives automated detection of query plan anti-patterns (Seq Scan, nested loops)
- [x] **QRY-04**: User can view grouped similar queries with different parameters (fingerprinting)
- [x] **QRY-05**: User can view query execution statistics (calls, total_time, rows, mean_time)

### Index Intelligence

- [x] **IDX-01**: User can view index usage statistics from pg_stat_user_indexes
- [x] **IDX-02**: User can see unused indexes that may be candidates for removal
- [x] **IDX-03**: User receives index impact estimation before creating new indexes
- [x] **IDX-04**: User can view recommended indexes with estimated benefit scores

### API Performance

- [x] **API-01**: User experiences faster API responses through response caching
- [x] **API-02**: System uses pgx v5 connection pooling for 2-3x query performance
- [x] **API-03**: Dashboard queries use dedicated read-only connection pool
- [x] **API-04**: User can monitor connection pool metrics (open, idle, in-use connections)

### Dashboard Optimization

- [x] **DASH-01**: User sees instant dashboard loads through pre-computed aggregations
- [x] **DASH-02**: System uses TimescaleDB continuous aggregates for time-series queries
- [x] **DASH-03**: User can view historical metrics without full table scans
- [x] **DASH-04**: Background worker pre-computes dashboard metrics on schedule

### Performance Monitoring

- [x] **MON-01**: User can access pprof endpoints for on-demand performance profiling
- [x] **MON-02**: User can view Prometheus metrics for API response time histograms
- [x] **MON-03**: User can monitor query duration percentiles (P50, P95, P99)
- [x] **MON-04**: User can view cache hit/miss rates for performance tuning

---

## v1.1 Requirements (Testing & Validation) ✓ Complete

Requirements for Testing & Validation milestone. Focuses on comprehensive test coverage and code quality.

### Backend Integration Testing

- [x] **TEST-01**: All API endpoints have integration tests covering happy path and error cases
- [x] **TEST-02**: Authentication boundary tests validate token validation, expiration, and revocation
- [x] **TEST-03**: Collector endpoints tested with boundary validation (invalid IDs, missing fields, SQL injection attempts)
- [x] **TEST-04**: Instance endpoints tested with various PostgreSQL versions and configuration combinations
- [x] **TEST-05**: User management endpoints tested with permission boundaries (admin vs regular user)
- [x] **TEST-06**: All HTTP status codes (200, 400, 401, 403, 404, 500) tested for appropriate endpoints

### Database Testing

- [x] **TEST-07**: Database transaction handling verified (commits, rollbacks, nested transactions)
- [x] **TEST-08**: Query validation tested with edge cases (empty results, null values, large datasets)
- [x] **TEST-09**: Connection pool management tested under concurrent load
- [x] **TEST-10**: Database schema migrations validated (no data loss, backward compatibility)
- [x] **TEST-11**: Time-series data handling tested (timestamp ordering, timezone conversions)

### Frontend Integration Testing

- [x] **TEST-12**: Dashboard components render correctly with API data
- [x] **TEST-13**: Form components validate input and display errors correctly
- [x] **TEST-14**: Navigation between pages maintains state properly
- [x] **TEST-15**: API error responses handled gracefully in UI
- [x] **TEST-16**: Authentication state persists across page refresh with httpOnly cookies

### Code Quality

- [x] **QUAL-01**: Go code passes go vet, go fmt, and golangci-lint checks
- [x] **QUAL-02**: TypeScript code passes ESLint with strict config
- [x] **QUAL-03**: No hardcoded credentials, secrets, or sensitive data in codebase
- [x] **QUAL-04**: Code comments explain "why" not "what" for complex logic
- [x] **QUAL-05**: Test coverage reports generated and tracked (target 80%+)
- [x] **QUAL-06**: Unused imports, variables, and functions removed

### Testing Infrastructure

- [x] **TEST-17**: Test suite runs in CI/CD pipeline automatically
- [x] **TEST-18**: Test failures block deployment (pipeline gates)
- [x] **TEST-19**: Coverage reports published after each test run
- [x] **TEST-20**: Test execution time documented (identify slow tests)
- [x] **TEST-21**: Mock/stub libraries configured for external dependencies

---

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Real-Time Features

- **REAL-01**: User receives real-time dashboard metrics via WebSocket
- **REAL-02**: User sees live query execution updates
- **REAL-03**: User receives instant alert notifications in browser

### Advanced Optimization

- **ADV-01**: System automatically creates recommended indexes (with user approval)
- **ADV-02**: System rewrites queries for better performance
- **ADV-03**: User receives query optimization suggestions with auto-apply option

### Performance Testing

- **PERF-01**: Load testing with concurrent user simulation (1000+ users)
- **PERF-02**: Query performance benchmarks documented
- **PERF-03**: Memory usage profiling and optimization

### Advanced Testing

- **ADV-TEST-01**: Chaos engineering tests (failure injection)
- **ADV-TEST-02**: Security penetration testing
- **ADV-TEST-03**: Accessibility testing (WCAG compliance)

---

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Automatic index creation | Risk of breaking production, require user approval workflow not in scope |
| Query rewriting | Error-prone, may change query semantics, requires domain expert review |
| Real-time dashboard metrics | Requires WebSocket infrastructure, high complexity, defer to v2 |
| SMS notifications | Email and Slack sufficient for v1.3 |
| Mobile app | Mobile app not in scope for v1 |

---

## Traceability

### v1.3 Requirements (Monitoring & Alerting Platform)

| Requirement | Phase | Status |
|-------------|-------|--------|
| REP-01 | Phase 10 | Complete |
| REP-02 | Phase 10 | Pending |
| REP-03 | Phase 10 | Pending |
| REP-04 | Phase 10 | Complete |
| REP-05 | Phase 12 | Pending |
| REP-06 | Phase 13 | Pending |
| HOST-01 | Phase 10 | Pending |
| HOST-02 | Phase 10 | Pending |
| HOST-03 | Phase 10 | Pending |
| HOST-04 | Phase 11 | Complete |
| HOST-05 | Phase 12 | Pending |
| INV-01 | Phase 10 | Complete |
| INV-02 | Phase 10 | Complete |
| INV-03 | Phase 10 | Complete |
| INV-04 | Phase 10 | Complete |
| INV-05 | Phase 10 | Complete |
| DATA-01 | Phase 11 | Complete |
| DATA-02 | Phase 11 | Complete |
| DATA-03 | Phase 11 | Complete |
| DATA-04 | Phase 11 | Complete |
| DATA-05 | Phase 11 | Complete |
| ALERT-01 | Phase 12 | Pending |
| ALERT-02 | Phase 12 | Pending |
| ALERT-03 | Phase 12 | Pending |
| ALERT-04 | Phase 12 | Pending |
| ALERT-05 | Phase 12 | Pending |
| ALERT-06 | Phase 12 | Pending |
| ALERT-07 | Phase 12 | Pending |
| ALERT-08 | Phase 12 | Pending |
| VER-01 | Phase 10 | Pending |
| VER-02 | Phase 10 | Pending |
| VER-03 | Phase 11 | Pending |
| VER-04 | Phase 10 | Pending |
| SCALE-01 | Phase 11 | Complete |
| SCALE-02 | Phase 11 | Complete |
| SCALE-03 | Phase 11 | Complete |
| SCALE-04 | Phase 11 | Complete |
| COLL-01 | Phase 10 | Pending |
| COLL-02 | Phase 10 | Pending |
| COLL-03 | Phase 10 | Pending |
| COLL-04 | Phase 10 | Pending |
| COLL-05 | Phase 10 | Pending |
| UI-01 | Phase 13 | Pending |
| UI-02 | Phase 12 | Pending |
| UI-03 | Phase 13 | Pending |
| UI-04 | Phase 13 | Pending |
| UI-05 | Phase 12 | Pending |
| TEST-01 | Phase 14 | Pending |
| TEST-02 | Phase 14 | Pending |
| TEST-03 | Phase 14 | Pending |
| TEST-04 | Phase 14 | Pending |
| TEST-05 | Phase 14 | Pending |

**v1.3 Coverage:**
- v1.3 requirements: 49 total
- Mapped to phases: 49
- Unmapped: 0 ✓

### v1.2 Requirements (Performance Optimization)

| Requirement | Phase | Status |
|-------------|-------|--------|
| QRY-01 | Phase 06 | Complete |
| QRY-02 | Phase 06 | Complete |
| QRY-03 | Phase 09 | Complete |
| QRY-04 | Phase 09 | Complete |
| QRY-05 | Phase 06 | Complete |
| IDX-01 | Phase 06 | Complete |
| IDX-02 | Phase 09 | Complete |
| IDX-03 | Phase 09 | Complete |
| IDX-04 | Phase 09 | Complete |
| API-01 | Phase 07 | Complete |
| API-02 | Phase 06 | Complete |
| API-03 | Phase 06 | Complete |
| API-04 | Phase 06 | Complete |
| DASH-01 | Phase 08 | Complete |
| DASH-02 | Phase 08 | Complete |
| DASH-03 | Phase 08 | Complete |
| DASH-04 | Phase 08 | Complete |
| MON-01 | Phase 06 | Complete |
| MON-02 | Phase 06 | Complete |
| MON-03 | Phase 06 | Complete |
| MON-04 | Phase 07 | Complete |

**v1.2 Coverage:**
- v1.2 requirements: 21 total
- Mapped to phases: 21
- Unmapped: 0 ✓

### v1.1 Requirements (Testing & Validation)

| Requirement | Phase | Status |
|-------------|-------|--------|
| TEST-01 | Phase 2 | Complete |
| TEST-02 | Phase 2 | Complete |
| TEST-03 | Phase 2 | Complete |
| TEST-04 | Phase 2 | Complete |
| TEST-05 | Phase 2 | Complete |
| TEST-06 | Phase 2 | Complete |
| TEST-07 | Phase 3 | Complete |
| TEST-08 | Phase 3 | Complete |
| TEST-09 | Phase 3 | Complete |
| TEST-10 | Phase 3 | Complete |
| TEST-11 | Phase 3 | Complete |
| TEST-12 | Phase 4 | Complete |
| TEST-13 | Phase 4 | Complete |
| TEST-14 | Phase 4 | Complete |
| TEST-15 | Phase 4 | Complete |
| TEST-16 | Phase 4 | Complete |
| QUAL-01 | Phase 2 | Complete |
| QUAL-02 | Phase 4 | Complete |
| QUAL-03 | Phase 2 | Complete |
| QUAL-04 | Phase 4 | Complete |
| QUAL-05 | Phase 5 | Complete |
| QUAL-06 | Phase 5 | Complete |
| TEST-17 | Phase 5 | Complete |
| TEST-18 | Phase 5 | Complete |
| TEST-19 | Phase 5 | Complete |
| TEST-20 | Phase 5 | Complete |
| TEST-21 | Phase 2 | Complete |

**v1.1 Coverage:**
- v1.1 requirements: 27 total
- Mapped to phases: 27
- Unmapped: 0 ✓

---

*Requirements defined: 2026-04-28 (v1.1), 2026-05-11 (v1.2), 2026-05-13 (v1.3)*
*Last updated: 2026-05-13 after v1.3 milestone definition*