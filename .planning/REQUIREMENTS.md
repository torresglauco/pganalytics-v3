# Requirements: pganalytics-v3

**Defined:** 2026-04-28 (v1.1), 2026-05-11 (v1.2)
**Core Value:** Enable database teams to proactively identify and fix performance issues before they impact production systems.

---

## v1.2 Requirements (Performance Optimization)

Requirements for Performance Optimization milestone. Each maps to roadmap phases.

### Query Optimization

- [ ] **QRY-01**: User can view top N slow queries by mean_time from pg_stat_statements
- [ ] **QRY-02**: User can see query performance timeline with historical trends
- [ ] **QRY-03**: User receives automated detection of query plan anti-patterns (Seq Scan, nested loops)
- [ ] **QRY-04**: User can view grouped similar queries with different parameters (fingerprinting)
- [ ] **QRY-05**: User can view query execution statistics (calls, total_time, rows, mean_time)

### Index Intelligence

- [ ] **IDX-01**: User can view index usage statistics from pg_stat_user_indexes
- [ ] **IDX-02**: User can see unused indexes that may be candidates for removal
- [ ] **IDX-03**: User receives index impact estimation before creating new indexes
- [ ] **IDX-04**: User can view recommended indexes with estimated benefit scores

### API Performance

- [ ] **API-01**: User experiences faster API responses through response caching
- [x] **API-02**: System uses pgx v5 connection pooling for 2-3x query performance
- [x] **API-03**: Dashboard queries use dedicated read-only connection pool
- [x] **API-04**: User can monitor connection pool metrics (open, idle, in-use connections)

### Dashboard Optimization

- [ ] **DASH-01**: User sees instant dashboard loads through pre-computed aggregations
- [ ] **DASH-02**: System uses TimescaleDB continuous aggregates for time-series queries
- [ ] **DASH-03**: User can view historical metrics without full table scans
- [ ] **DASH-04**: Background worker pre-computes dashboard metrics on schedule

### Performance Monitoring

- [ ] **MON-01**: User can access pprof endpoints for on-demand performance profiling
- [ ] **MON-02**: User can view Prometheus metrics for API response time histograms
- [ ] **MON-03**: User can monitor query duration percentiles (P50, P95, P99)
- [ ] **MON-04**: User can view cache hit/miss rates for performance tuning

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
- [ ] **TEST-18**: Test failures block deployment (pipeline gates)
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
| Redis mandatory integration | Single-instance deployments don't need distributed cache, keep optional |
| Multi-tenant performance isolation | Not a current deployment scenario, revisit if needed |
| Mobile app testing | Mobile app not in scope for v1 |

---

## Traceability

### v1.2 Requirements (Performance Optimization)

| Requirement | Phase | Status |
|-------------|-------|--------|
| QRY-01 | Phase 06 | Pending |
| QRY-02 | Phase 06 | Pending |
| QRY-03 | Phase 09 | Pending |
| QRY-04 | Phase 09 | Pending |
| QRY-05 | Phase 06 | Pending |
| IDX-01 | Phase 06 | Pending |
| IDX-02 | Phase 09 | Pending |
| IDX-03 | Phase 09 | Pending |
| IDX-04 | Phase 09 | Pending |
| API-01 | Phase 07 | Pending |
| API-02 | Phase 06 | Complete |
| API-03 | Phase 06 | Complete |
| API-04 | Phase 06 | Complete |
| DASH-01 | Phase 08 | Pending |
| DASH-02 | Phase 08 | Pending |
| DASH-03 | Phase 08 | Pending |
| DASH-04 | Phase 08 | Pending |
| MON-01 | Phase 06 | Pending |
| MON-02 | Phase 06 | Pending |
| MON-03 | Phase 06 | Pending |
| MON-04 | Phase 07 | Pending |

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
| TEST-18 | Phase 5 | Pending |
| TEST-19 | Phase 5 | Complete |
| TEST-20 | Phase 5 | Complete |
| TEST-21 | Phase 2 | Complete |

**v1.1 Coverage:**
- v1.1 requirements: 27 total
- Mapped to phases: 27
- Unmapped: 0 ✓

---

*Requirements defined: 2026-04-28 (v1.1), 2026-05-11 (v1.2)*
*Last updated: 2026-05-11 after v1.2 milestone definition*