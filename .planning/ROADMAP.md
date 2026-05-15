# Roadmap: pganalytics-v3

## Milestones

- ✅ **v1.0 Security & E2E Testing** - Phases 01-04 (shipped 2026-04-22)
- ✅ **v1.1 Testing & Validation** - Phase 05 (shipped 2026-04-30)
- ✅ **v1.2 Performance Optimization** - Phases 06-09 (shipped 2026-05-13)
- 🚧 **v1.3 Monitoring & Alerting Platform** - Phases 10-14 (in progress)

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

<details>
<summary>✅ v1.2 Performance Optimization (Phases 06-09) - SHIPPED 2026-05-13</summary>

### Phase 06: Query Optimization Foundation
**Goal**: Users experience faster query execution with optimized connection pooling and performance visibility
**Depends on**: Phase 05 (CI/CD Infrastructure complete)
**Requirements**: QRY-01, QRY-02, QRY-05, IDX-01, API-02, API-03, API-04, MON-01, MON-02, MON-03
**Plans**: 4 plans

Plans:
- [x] 06-01: Migrate from lib/pq to pgx v5 with connection pooling and read-only pool
- [x] 06-02: Implement slow query identification and timeline
- [x] 06-03: Enable pprof and Prometheus histograms
- [x] 06-04: Add metrics middleware and API endpoints

### Phase 07: Caching Infrastructure
**Goal**: Users experience faster API responses through intelligent caching
**Depends on**: Phase 06 (Query Optimization Foundation)
**Requirements**: API-01, MON-04
**Plans**: 2 plans

Plans:
- [x] 07-01: Implement API response caching
- [x] 07-02: Add cache metrics and invalidation

### Phase 08: Dashboard Optimization
**Goal**: Users see instant dashboard loads through pre-computed aggregations
**Depends on**: Phase 07 (Caching Infrastructure)
**Requirements**: DASH-01, DASH-02, DASH-03, DASH-04
**Plans**: 3 plans

Plans:
- [x] 08-01: Create TimescaleDB continuous aggregates
- [x] 08-02: Implement dashboard pre-computation worker
- [x] 08-03: Wire aggregate queries to API handlers

### Phase 09: Index Intelligence
**Goal**: Users receive instant, actionable index recommendations with impact estimation
**Depends on**: Phase 08 (Dashboard Optimization)
**Requirements**: QRY-03, QRY-04, IDX-02, IDX-03, IDX-04
**Plans**: 2 plans

Plans:
- [x] 09-01: Implement query plan analysis and fingerprinting
- [x] 09-02: Add index recommendation engine with impact estimation

</details>

### 🚧 v1.3 Monitoring & Alerting Platform (In Progress)

**Milestone Goal:** Comprehensive monitoring and alerting for PostgreSQL replication, host health, data classification, and multi-version support.

- [ ] **Phase 10: Collector & Backend Foundation** - Replication monitoring, host monitoring, database inventory, collector architecture, multi-version support
- [x] **Phase 11: Data Classification & Health Analysis** - PII/PCI detection, host health scores, scalability infrastructure (completed 2026-05-14)
- [x] **Phase 12: Alerting System** - Alert rules, notifications, escalation policies, notification channel management (completed 2026-05-15)
- [x] **Phase 13: Frontend UI** - Dashboards for replication topology, data classification, host inventory (completed 2026-05-15)
- [x] **Phase 14: Testing & Quality** - Comprehensive test coverage for all new features (completed 2026-05-15)

## Phase Details

### Phase 10: Collector & Backend Foundation
**Goal**: Users can monitor PostgreSQL replication, host status, and database inventory through a flexible collector architecture
**Depends on**: Phase 09 (Index Intelligence complete)
**Requirements**: REP-01, REP-02, REP-03, REP-04, HOST-01, HOST-02, HOST-03, INV-01, INV-02, INV-03, INV-04, INV-05, VER-01, VER-02, VER-04, COLL-01, COLL-02, COLL-03, COLL-04, COLL-05
**Success Criteria** (what must be TRUE):
  1. User can view streaming replication status with write/flush/replay lag metrics
  2. User can view logical replication subscriptions and publications
  3. User can view host up/down status and OS metrics (CPU, memory, disk, network)
  4. User can view complete table and index inventory with usage statistics
  5. Collector supports both decentralized (co-located) and centralized (remote) deployment modes
**Plans**: TBD

### Phase 11: Data Classification & Health Analysis
**Goal**: Users can identify sensitive data and understand host/database health through automated analysis
**Depends on**: Phase 10 (Collector & Backend Foundation)
**Requirements**: HOST-04, DATA-01, DATA-02, DATA-03, DATA-04, DATA-05, VER-03, SCALE-01, SCALE-02, SCALE-03, SCALE-04
**Success Criteria** (what must be TRUE):
  1. User can view PII detection results for sensitive data patterns (CPF, CNPJ, email, phone)
  2. User can view PCI detection results for credit card numbers
  3. User can view host health score based on resource utilization
  4. System supports 2000+ PostgreSQL clusters with multi-tenancy isolation
  5. User can view version-specific health checks for PostgreSQL 11-17
**Plans**: 4 plans

Plans:
- [ ] 11-01: Data Classification Backend (PII/PCI detection models, storage, API)
- [ ] 11-02: Host Health Scoring (weighted calculation, persistence, API)
- [ ] 11-03: Multi-Tenancy Infrastructure (tenant_id, RLS, middleware, API)
- [ ] 11-04: Version-Specific Health Checks (adaptive queries for PG 11-17)

### Phase 12: Alerting System
**Goal**: Users receive timely notifications when metrics breach configured thresholds
**Depends on**: Phase 11 (Data Classification & Health Analysis)
**Requirements**: REP-05, HOST-05, ALERT-01, ALERT-02, ALERT-03, ALERT-04, ALERT-05, ALERT-06, ALERT-07, ALERT-08, UI-02, UI-05
**Success Criteria** (what must be TRUE):
  1. User can configure alert rules based on metric thresholds
  2. User receives email notifications for triggered alerts
  3. User receives Slack notifications via webhook integration
  4. User can view alert history and acknowledge/silence active alerts
  5. User can configure escalation policies for critical alerts
**Plans**: 4 plans

Plans:
- [x] 12-01: Alert Rules Repository & Handler Wiring
- [x] 12-02: SMTP Email Delivery Implementation
- [x] 12-03: Alert Rules CRUD API and OpsGenie Channel
- [x] 12-04: Multi-Tenancy and UI Enhancement

### Phase 13: Frontend UI
**Goal**: Users can visualize monitoring data through intuitive dashboards and topology views
**Depends on**: Phase 12 (Alerting System)
**Requirements**: REP-06, UI-01, UI-03, UI-04
**Success Criteria** (what must be TRUE):
  1. User can view replication topology as an interactive graph
  2. User can view data classification reports with drill-down by database/table
  3. User can view host inventory dashboards with status and metrics
**Plans**: 3 plans

Plans:
- [x] 13-01: Replication Topology Graph - Interactive visualization with @xyflow/react
- [ ] 13-02: Data Classification Reports - Drill-down UI with filters and charts
- [ ] 13-03: Host Inventory Dashboard - Status table and metrics visualization

### Phase 14: Testing & Quality
**Goal**: All new features have comprehensive test coverage ensuring reliability
**Depends on**: Phase 13 (Frontend UI)
**Requirements**: TEST-01, TEST-02, TEST-03, TEST-04, TEST-05
**Success Criteria** (what must be TRUE):
  1. All new collector plugins have C++ unit tests passing
  2. All new backend services have Go unit tests passing
  3. All new API endpoints have integration tests covering happy path and error cases
  4. All new frontend components have tests passing
  5. End-to-end tests cover critical user flows for monitoring features
**Plans**: 4 plans

Plans:
- [x] 14-01: Backend Unit Tests - Tenant middleware, health calculator, collector plugins (TEST-01, TEST-02)
- [x] 14-02: Backend Integration Tests - Replication, host monitoring, alert rules APIs (TEST-03)
- [ ] 14-03: Frontend Unit Tests - Components and pages for topology, classification, hosts (TEST-04)
- [x] 14-04: E2E Tests - Playwright tests for critical user flows (TEST-05)

## Progress

**Execution Order:**
Phases execute in numeric order: 10 → 11 → 12 → 13 → 14

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 01. Security Fixes | v1.0 | 3/3 | Complete | 2026-04-22 |
| 02. Authentication Hardening | v1.0 | 2/2 | Complete | 2026-04-22 |
| 03. E2E Test Infrastructure | v1.0 | 2/2 | Complete | 2026-04-22 |
| 04. Core E2E Tests | v1.0 | 3/3 | Complete | 2026-04-22 |
| 05. CI/CD Infrastructure | v1.1 | 3/3 | Complete | 2026-04-30 |
| 06. Query Optimization Foundation | v1.2 | 4/4 | Complete | 2026-05-11 |
| 07. Caching Infrastructure | v1.2 | 2/2 | Complete | 2026-05-12 |
| 08. Dashboard Optimization | v1.2 | 3/3 | Complete | 2026-05-12 |
| 09. Index Intelligence | v1.2 | 2/2 | Complete | 2026-05-13 |
| 10. Collector & Backend Foundation | v1.3 | 5/5 | Complete | 2026-05-13 |
| 11. Data Classification & Health Analysis | v1.3 | 4/4 | Complete | 2026-05-14 |
| 12. Alerting System | v1.3 | 4/4 | Complete | 2026-05-15 |
| 13. Frontend UI | v1.3 | Complete    | 2026-05-15 | 2026-05-15 |
| 14. Testing & Quality | 4/4 | Complete   | 2026-05-15 | 2026-05-15 |

---

*Roadmap created: 2026-04-28*
*Last updated: 2026-05-15 for Phase 14 planning*