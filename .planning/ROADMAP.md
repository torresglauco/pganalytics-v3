# Roadmap: pganalytics-v3

## Milestones

- ✅ **v1.0 Security & E2E Testing** - Phases 01-04 (shipped 2026-04-22)
- ✅ **v1.1 Testing & Validation** - Phase 05 (shipped 2026-04-30)
- 🚧 **v1.2 Performance Optimization** - Phases 06-09 (in progress)

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

<details>
<summary>✅ v1.0 MVP (Phases 01-04) - SHIPPED 2026-04-22</summary>

### Phase 1: Security Fixes
**Goal**: Fix authentication vulnerabilities
**Plans**: 3 plans

Plans:
- [x] 01-01: Implement secure cookie handling
- [x] 01-02: Add CSRF protection
- [x] 01-03: Fix token validation issues

### Phase 2: Authentication Hardening
**Goal**: Harden authentication boundaries
**Plans**: 2 plans

Plans:
- [x] 02-01: Implement token expiration handling
- [x] 02-02: Add audit logging for auth events

### Phase 3: E2E Test Infrastructure
**Goal**: Establish E2E testing foundation
**Plans**: 2 plans

Plans:
- [x] 03-01: Configure Playwright test environment
- [x] 03-02: Create base test utilities

### Phase 4: Core E2E Tests
**Goal**: Cover critical user flows with E2E tests
**Plans**: 3 plans

Plans:
- [x] 04-01: Authentication flow tests
- [x] 04-02: Dashboard navigation tests
- [x] 04-03: Instance management tests

</details>

<details>
<summary>✅ v1.1 Testing & Validation (Phase 05) - SHIPPED 2026-04-30</summary>

### Phase 05: CI/CD Infrastructure
**Goal**: Achieve enterprise-grade reliability through comprehensive testing and CI/CD automation
**Plans**: 3 plans

Plans:
- [x] 05-01: CI Quality Gate with Coverage Reporting
- [x] 05-02: GitHub Actions E2E Testing Pipeline
- [x] 05-03: Branch Protection Configuration

</details>

### v1.2 Performance Optimization (In Progress)

**Milestone Goal:** Accelerate query/API response times for dashboard, query analysis, and index advisor operations.

- [x] **Phase 06: Query Optimization Foundation** - Establish optimized query infrastructure with pgx v5, connection pooling, and performance monitoring (completed 2026-05-11)
- [x] **Phase 07: Caching Infrastructure** - Add response caching for faster API responses and reduced database load (completed 2026-05-12)
- [ ] **Phase 08: Dashboard Optimization** - Implement TimescaleDB continuous aggregates for instant dashboard loads
- [ ] **Phase 09: Index Intelligence** - Add background index analysis with impact estimation

## Phase Details

### Phase 06: Query Optimization Foundation
**Goal**: Users experience faster query execution with optimized connection pooling and performance visibility
**Depends on**: Phase 05 (CI/CD Infrastructure complete)
**Requirements**: QRY-01, QRY-02, QRY-05, IDX-01, API-02, API-03, API-04, MON-01, MON-02, MON-03
**Success Criteria** (what must be TRUE):
  1. User can view top slow queries ranked by mean execution time
  2. User can see query performance trends over time through timeline visualization
  3. System uses pgx v5 connection pooling for all database operations
  4. User can monitor connection pool status showing open, idle, and in-use connections
  5. User can profile application performance on-demand via pprof endpoints
**Plans**: 4 plans

Plans:
- [x] 06-01: Migrate from lib/pq to pgx v5 with connection pooling and read-only pool
- [x] 06-02: Implement slow query identification and timeline
- [x] 06-03: Enable pprof and Prometheus histograms
- [ ] 06-04: Add metrics middleware and API endpoints

### Phase 07: Caching Infrastructure
**Goal**: Users experience faster API responses through intelligent caching
**Depends on**: Phase 06 (Query Optimization Foundation)
**Requirements**: API-01, MON-04
**Success Criteria** (what must be TRUE):
  1. User sees faster dashboard API responses through response caching
  2. User can view cache hit/miss rates to understand caching effectiveness
**Plans**: 2 plans

Plans:
- [x] 07-01: Implement API response caching
- [x] 07-02: Add cache metrics and invalidation

### Phase 08: Dashboard Optimization
**Goal**: Users see instant dashboard loads through pre-computed aggregations
**Depends on**: Phase 07 (Caching Infrastructure)
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04
**Success Criteria** (what must be TRUE):
  1. User sees instant dashboard loads without waiting for on-demand aggregations
  2. System uses TimescaleDB continuous aggregates for time-series queries
  3. User can view historical metrics without triggering slow full-table scans
  4. Dashboard metrics are pre-computed by background worker on schedule
**Plans**: 2 plans

Plans:
- [ ] 08-01: Create TimescaleDB continuous aggregates
- [ ] 08-02: Implement dashboard pre-computation worker

### Phase 09: Index Intelligence
**Goal**: Users receive instant, actionable index recommendations with impact estimation
**Depends on**: Phase 08 (Dashboard Optimization)
**Requirements**: QRY-03, QRY-04, IDX-02, IDX-03, IDX-04
**Success Criteria** (what must be TRUE):
  1. User receives automated detection of query plan anti-patterns (Seq Scan, nested loops)
  2. User can view grouped similar queries with different parameters (fingerprinting)
  3. User can see unused indexes that are candidates for removal
  4. User receives index impact estimation before creating new indexes
**Plans**: 2 plans

Plans:
- [ ] 09-01: Implement query plan analysis and fingerprinting
- [ ] 09-02: Add index recommendation engine with impact estimation

## Progress

**Execution Order:**
Phases execute in numeric order: 06 → 07 → 08 → 09

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 06. Query Optimization Foundation | v1.2 | Complete | 2026-05-11 | 2026-05-11 |
| 07. Caching Infrastructure | v1.2 | Complete | 2026-05-12 | 2026-05-12 |
| 08. Dashboard Optimization | v1.2 | 0/2 | Not started | - |
| 09. Index Intelligence | v1.2 | 0/2 | Not started | - |

---

*Roadmap created: 2026-05-11*
*Last updated: 2026-05-12*