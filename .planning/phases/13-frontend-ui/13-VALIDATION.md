---
phase: 13
slug: frontend-ui
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-05-15"
---

# Phase 13 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Vitest 1.0.0 |
| **Config file** | frontend/vite.config.ts |
| **Quick run command** | `npm run test --prefix frontend` |
| **Full suite command** | `npm run test:coverage --prefix frontend` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `npm run test --prefix frontend`
- **After every plan wave:** Run `npm run test:coverage --prefix frontend`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 13-01-01 | 01 | 1 | REP-06, UI-01 | unit | `vitest run src/components/topology/` | Wave 0 | pending |
| 13-01-02 | 01 | 1 | REP-06 | unit | `vitest run src/api/replicationApi.test.ts` | Wave 0 | pending |
| 13-02-01 | 02 | 1 | UI-03 | unit | `vitest run src/pages/DataClassificationPage.test.tsx` | Wave 0 | pending |
| 13-02-02 | 02 | 1 | UI-03 | unit | `vitest run src/api/classificationApi.test.ts` | Wave 0 | pending |
| 13-03-01 | 03 | 2 | UI-04 | unit | `vitest run src/pages/HostInventoryPage.test.tsx` | Wave 0 | pending |
| 13-03-02 | 03 | 2 | UI-04 | unit | `vitest run src/api/hostApi.test.ts` | Wave 0 | pending |

*Status: pending | green | red | flaky*

---

## Wave 0 Requirements

- [ ] `frontend/src/components/topology/TopologyGraph.test.tsx` — topology rendering tests
- [ ] `frontend/src/components/topology/TopologyNode.test.tsx` — custom node tests
- [ ] `frontend/src/pages/ReplicationTopologyPage.test.tsx` — page integration tests
- [ ] `frontend/src/pages/DataClassificationPage.test.tsx` — classification page tests
- [ ] `frontend/src/pages/HostInventoryPage.test.tsx` — host inventory page tests
- [ ] `frontend/src/api/replicationApi.test.ts` — replication API tests
- [ ] `frontend/src/api/hostApi.test.ts` — host API tests
- [ ] `frontend/src/api/classificationApi.test.ts` — classification API tests
- [ ] `frontend/src/types/replication.ts` — replication type definitions
- [ ] `frontend/src/types/host.ts` — host type definitions
- [ ] `frontend/src/types/classification.ts` — classification type definitions

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Graph drag/zoom interactions | REP-06 | Complex user interactions | Drag nodes, zoom with mouse wheel, verify smooth performance |
| Classification drill-down navigation | UI-03 | Multi-step user flow | Click database → schema → table, verify breadcrumb navigation |
| Responsive layout on mobile | UI-01, UI-03, UI-04 | Device-specific testing | Resize browser to 375px, verify all dashboards usable |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending