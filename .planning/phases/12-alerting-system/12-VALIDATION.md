---
phase: 12
slug: alerting-system
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-05-14
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go: testing + testify |
| **Config file** | None (tests self-contained) |
| **Quick run command** | `go build ./backend/...` (compilation check) |
| **Full suite command** | `go test ./... -cover -race` |
| **Estimated runtime** | ~90 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go build ./backend/...` (compilation verification)
- **After every plan wave:** Run `go test ./... -cover -race`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | Status |
|---------|------|------|-------------|-----------|-------------------|--------|
| 12-01-01 | 01 | 1 | ALERT-01 | build | `go build ./backend/pkg/models` | ⬜ |
| 12-01-02 | 01 | 1 | ALERT-01 | build | `go build ./backend/internal/storage` | ⬜ |
| 12-01-03 | 01 | 1 | ALERT-01 | build | `go build ./backend/internal/api` | ⬜ |
| 12-02-01 | 02 | 1 | ALERT-02 | build | `grep -c "smtp\|Send" backend/internal/notifications/email.go` | ⬜ |
| 12-02-02 | 02 | 1 | ALERT-03 | build | `grep -c "slack\|webhook" backend/internal/notifications/*.go` | ⬜ |
| 12-03-01 | 03 | 2 | ALERT-04 | build | `go build ./backend/internal/api` | ⬜ |
| 12-03-02 | 03 | 2 | ALERT-05 | build | `grep -c "silence\|acknowledge" backend/internal/api/handlers_alerts.go` | ⬜ |
| 12-04-01 | 04 | 2 | ALERT-06 | build | `go build ./backend/internal/services` | ⬜ |
| 12-04-02 | 04 | 2 | ALERT-07 | build | `grep -c "EscalationPolicy\|escalation" backend/pkg/models/alert_models.go` | ⬜ |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

All tasks use build-based verification that does not require pre-existing test infrastructure.

**Existing infrastructure covers**: Alert rule engine, notification channels (Slack, PagerDuty, Jira, Webhook), escalation policies, silence management.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Email delivery | ALERT-02 | Requires SMTP server | Configure SMTP settings, trigger alert, verify email received |
| Slack webhook | ALERT-03 | Requires Slack workspace | Configure webhook URL, trigger alert, verify Slack message |
| Escalation timing | ALERT-07 | Requires timing verification | Create escalation policy, trigger alert, verify timing |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 90s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending