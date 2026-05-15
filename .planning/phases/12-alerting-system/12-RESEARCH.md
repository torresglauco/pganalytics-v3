# Phase 12: Alerting System - Research

**Researched:** 2026-05-14
**Domain:** Alerting, Notifications, Escalation Policies, Multi-channel Delivery
**Confidence:** HIGH

## Summary

Phase 12 builds upon a substantial existing alerting foundation. The codebase already contains:
- Alert rule engine with threshold, anomaly, change, and composite condition types
- Notification service with Slack, Email, Webhook, PagerDuty, and Jira channel implementations
- Escalation policies with multi-step workflows and acknowledgment tracking
- Silence management with auto-expiration
- Circuit breaker pattern for resilient notification delivery

The primary work involves completing stubbed implementations (email SMTP), wiring existing handlers to routes, building the alert rules CRUD API, and creating the frontend UI for alert configuration and notification channel management.

**Primary recommendation:** Leverage existing infrastructure; focus on completing gaps (email delivery, API endpoints) and building the UI layer for user-configurable alert rules.

<user_constraints>
## User Constraints

No CONTEXT.md exists for this phase. The following requirements from REQUIREMENTS.md define the scope:

### Phase Requirements (from REQUIREMENTS.md)
- **REP-05**: User can view replication lag alerts when thresholds exceeded
- **HOST-05**: User can configure host-level alert thresholds
- **ALERT-01**: User can configure alert rules based on metric thresholds
- **ALERT-02**: User can receive email notifications for alerts
- **ALERT-03**: User can receive Slack notifications via webhook
- **ALERT-04**: User can configure generic webhooks for alert notifications
- **ALERT-05**: User can integrate with PagerDuty/OpsGenie for incident management
- **ALERT-06**: User can view alert history with timestamps
- **ALERT-07**: User can acknowledge and silence alerts
- **ALERT-08**: User can configure alert escalation policies
- **UI-02**: User can configure alert rules via UI
- **UI-05**: User can manage notification channels

### Success Criteria
1. User can configure alert rules based on metric thresholds
2. User receives email notifications for triggered alerts
3. User receives Slack notifications via webhook integration
4. User can view alert history and acknowledge/silence active alerts
5. User can configure escalation policies for critical alerts
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| REP-05 | Replication lag alerts when thresholds exceeded | Existing `AlertRuleEngine` supports metric-based rules; add replication lag metric type |
| HOST-05 | Host-level alert thresholds | Extend `AlertRule` model for host metrics; use existing `metrics` table |
| ALERT-01 | Configure alert rules based on metric thresholds | `AlertRule` model exists with `metric_type`, `condition_type`, `condition_value`; needs CRUD API |
| ALERT-02 | Email notifications for alerts | `EmailChannel` exists but stubbed; implement SMTP with `net/smtp` or `go-mail` |
| ALERT-03 | Slack notifications via webhook | `SlackChannel` fully implemented with circuit breaker; ready to use |
| ALERT-04 | Generic webhooks for alerts | `WebhookChannel` fully implemented; supports auth headers, custom payloads |
| ALERT-05 | PagerDuty/OpsGenie integration | `PagerDutyChannel` implemented; OpsGenie needs similar implementation |
| ALERT-06 | View alert history with timestamps | `alert_triggers` table exists; needs API endpoint and UI |
| ALERT-07 | Acknowledge and silence alerts | `EscalationService.AcknowledgeAlert` and `SilenceService` implemented; needs UI |
| ALERT-08 | Configure escalation policies | `EscalationPolicy` model and service exist; needs CRUD API and UI |
| UI-02 | Configure alert rules via UI | Build on existing `AlertsPage.tsx` with form components |
| UI-05 | Manage notification channels | `NotificationChannelsPage.tsx` exists with full CRUD UI |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| net/smtp | stdlib | Email delivery | Go standard library, no external dependencies |
| net/http | stdlib | Webhook delivery | Already used throughout codebase |
| github.com/gorilla/websocket | v1.5.3 | Real-time alert events | Already in project for WebSocket broadcasting |
| go.uber.org/zap | v1.27.0 | Structured logging | Already used for alert logging |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| encoding/json | stdlib | Alert payload serialization | Channel configurations |
| context | stdlib | Request cancellation | Notification timeouts |
| time | stdlib | Scheduling, TTL | Escalation delays, silence expiration |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| net/smtp | github.com/go-gomail/gomail | gomail simpler API but adds dependency; net/smtp sufficient |
| Custom circuit breaker | github.com/sony/gobreaker | Custom implementation matches existing patterns, no migration needed |
| Custom escalation worker | github.com/hibiken/asynq | Asynq adds Redis dependency; current approach uses database state |

**Installation:** No new packages required. All functionality uses existing dependencies or Go standard library.

**Version verification:** All versions verified from `/Users/glauco.torres/git/pganalytics-v3/go.mod`.

## Architecture Patterns

### Recommended Project Structure
```
backend/
├── internal/
│   ├── notifications/       # Channel implementations (exists)
│   │   ├── notification_service.go
│   │   └── channels.go      # Slack, Email, Webhook, PagerDuty, Jira
│   └── jobs/
│       └── alert_rule_engine.go  # Evaluation loop (exists)
├── pkg/
│   ├── handlers/
│   │   ├── alerts.go        # Alert rules CRUD (needs creation)
│   │   ├── escalations.go   # Exists - wire to routes
│   │   ├── silences.go      # Exists - wire to routes
│   │   └── conditions.go    # Exists - wire to routes
│   └── services/
│       ├── alert_worker.go       # Trigger creation (exists)
│       ├── notification_worker.go # Delivery loop (exists)
│       ├── escalation_service.go  # Policy management (exists)
│       ├── escalation_worker.go   # Step execution (exists)
│       └── silence_service.go     # Silence management (exists)
frontend/
├── src/
│   ├── pages/
│   │   ├── AlertRulesPage.tsx     # Exists - enhance
│   │   ├── AlertsPage.tsx         # Exists - enhance
│   │   └── NotificationChannelsPage.tsx  # Exists - complete
│   ├── components/
│   │   └── alerts/
│   │       └── AlertsViewer.tsx   # Exists - enhance
│   └── api/
│       └── notificationsApi.ts    # Exists - ensure complete
```

### Pattern 1: Alert Rule Lifecycle
**What:** Alert rules are evaluated periodically, triggering notifications when conditions match.
**When to use:** All alert rule implementations.
**Example:**
```go
// Source: backend/internal/jobs/alert_rule_engine.go (existing)
type AlertRule struct {
    ID                   int64
    UserID               int
    Name                 string
    RuleType             string // "threshold", "change", "anomaly", "composite"
    MetricName           string
    Condition            json.RawMessage
    AlertSeverity        string // "low", "medium", "high", "critical"
    EvaluationInterval   int    // seconds
    ForDurationSeconds   int    // trigger only if true for N seconds
    NotificationEnabled  bool
    NotificationChannels []int64
    IsEnabled            bool
}

// Evaluation flow (existing in alert_rule_engine.go):
// 1. loadRules() - fetch enabled rules from database
// 2. evaluateRule() - check condition against current metrics
// 3. storeEvaluation() - record evaluation result
// 4. fireAlert() - create alert and trigger notifications
```

### Pattern 2: Notification Channel Interface
**What:** All notification channels implement a common interface for consistent delivery.
**When to use:** Adding new notification channels (e.g., OpsGenie).
**Example:**
```go
// Source: backend/internal/notifications/notification_service.go (existing)
type NotificationChannel interface {
    Type() string
    Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error)
    Validate(config ChannelConfig) error
    Test(ctx context.Context, config ChannelConfig) error
}

// Slack implementation uses webhook (no SDK needed):
// Source: backend/internal/notifications/channels.go
func (s *SlackChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
    // Check circuit breaker first
    if s.circuitBreaker.IsOpen() {
        return &DeliveryResult{Success: false, ErrorMsg: "Circuit open"}, nil
    }

    // POST to webhook URL
    resp, err := s.httpClient.Do(req)
    // Handle response, update circuit breaker state
}
```

### Pattern 3: Escalation Policy Execution
**What:** Multi-step notification escalation with acknowledgment support.
**When to use:** Critical alerts requiring multiple notification attempts.
**Example:**
```go
// Source: backend/pkg/services/escalation_worker.go (existing)
func (ew *EscalationWorker) processEscalation(state *models.EscalationState) error {
    policy, _ := ew.db.GetPolicy(state.PolicyID)

    // Get current step configuration
    step := policy.Steps[state.CurrentStep]

    // Send notification
    notifReq := &NotificationRequest{
        AlertTriggerID: state.AlertTriggerID,
        Channel:        step.ChannelType,
        Config:         step.ChannelConfig,
    }
    ew.notifier.SendNotification(notifReq)

    // Advance state
    state.CurrentStep++
    if state.CurrentStep < len(policy.Steps) {
        nextStep := policy.Steps[state.CurrentStep]
        nextTime := now.Add(time.Duration(nextStep.DelayMinutes) * time.Minute)
        state.NextEscalationAt = &nextTime
    }
}
```

### Anti-Patterns to Avoid
- **Blocking notification delivery:** Never block the alert evaluation loop waiting for SMTP/webhook responses. Use the existing async `NotificationWorker`.
- **Missing circuit breaker:** All external HTTP calls must go through the existing `CircuitBreaker` pattern to prevent cascading failures.
- **Duplicate alerts:** Use the existing `generateFingerprint()` and `findExistingAlert()` to deduplicate alerts before firing.
- **Unbounded retry:** The existing implementation limits retries to 3 with exponential backoff (5s, 30s, 300s) - do not change without consideration.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Email delivery | Custom SMTP client | `net/smtp` with existing `EmailChannel` stub | Standard library handles TLS, authentication |
| Circuit breaker | New implementation | Existing `CircuitBreaker` in `channels.go` | Already integrated with all channels |
| Alert deduplication | Manual tracking | `generateFingerprint()` + `alerts.fingerprint` column | Database-enforced uniqueness |
| Escalation timing | Custom scheduler | `EscalationWorker` with `next_escalation_at` column | Database-driven scheduling |

**Key insight:** The alerting infrastructure is 80% complete. Focus on filling gaps (email SMTP, API endpoints) rather than building new systems.

## Common Pitfalls

### Pitfall 1: Missing Tenant Context in Alert Rules
**What goes wrong:** Alert rules created without tenant_id won't appear for users in multi-tenant mode.
**Why it happens:** Phase 11 added multi-tenancy but alert_rules table may not have tenant_id column.
**How to avoid:** Add `tenant_id` column to `alert_rules` table and set RLS policies. Use `app.current_tenant` session variable.
**Warning signs:** Alert rules visible across tenants; rules created by one user visible to others.

### Pitfall 2: Email Configuration Missing Environment Variables
**What goes wrong:** Email notifications fail silently when SMTP_* variables not set.
**Why it happens:** Current `EmailChannel.Send()` logs but doesn't actually send email.
**How to avoid:**
1. Add validation in `EmailChannel.Validate()` to check for required SMTP settings
2. Implement actual SMTP delivery in `Send()` method
3. Add health check endpoint that verifies SMTP connectivity
**Warning signs:** Email notifications marked "delivered" in logs but recipients don't receive them.

### Pitfall 3: Unbounded Alert Trigger Growth
**What goes wrong:** `alert_triggers` table grows without cleanup, impacting query performance.
**Why it happens:** No retention policy for old triggers.
**How to avoid:**
1. Add TimescaleDB hypertable with retention policy for `alert_triggers`
2. Implement cleanup job similar to existing log cleanup
3. Consider partitioning by `triggered_at` date
**Warning signs:** Slow queries on alert history; storage growth without bound.

### Pitfall 4: Missing Notification Channel Verification
**What goes wrong:** Users add invalid webhook URLs or expired Slack webhooks, notifications fail silently.
**Why it happens:** Channels added without verification step.
**How to avoid:** Use existing `TestChannel()` method to verify channels before activation. Set `is_verified = true` only after successful test.
**Warning signs:** High "failed" notification rate; users unaware their channels are broken.

## Code Examples

### Email SMTP Implementation (Completion)
```go
// Complete the stubbed EmailChannel.Send() method
// Source: backend/internal/notifications/channels.go (enhance existing)

func (e *EmailChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
    if e.circuitBreaker.IsOpen() {
        return &DeliveryResult{Success: false, ErrorMsg: "Circuit open"}, nil
    }

    var emailConfig EmailConfig
    json.Unmarshal(config.Config, &emailConfig)

    // Build email
    from := os.Getenv("SMTP_FROM")
    to := emailConfig.Recipients
    subject := fmt.Sprintf("[%s] %s", strings.ToUpper(alert.Severity), alert.Title)
    body := FormatAlertHTML(alert)

    // Compose message
    msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
        from, strings.Join(to, ","), subject, body)

    // Send via SMTP
    auth := smtp.PlainAuth("",
        os.Getenv("SMTP_USER"),
        os.Getenv("SMTP_PASSWORD"),
        os.Getenv("SMTP_HOST"))

    addr := fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))
    err := smtp.SendMail(addr, auth, from, to, []byte(msg))

    if err != nil {
        e.circuitBreaker.RecordFailure()
        return &DeliveryResult{Success: false, ErrorMsg: err.Error()}, nil
    }

    e.circuitBreaker.RecordSuccess()
    return &DeliveryResult{Success: true, MessageID: fmt.Sprintf("email_%d", alert.AlertID)}, nil
}
```

### Alert Rule CRUD Handler
```go
// Create new file: backend/pkg/handlers/alerts.go

package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/torresglauco/pganalytics-v3/backend/pkg/models"
    "github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

type AlertRulesHandler struct {
    service *services.AlertRulesService
}

// CreateAlertRule handles POST /api/v1/alert-rules
func (h *AlertRulesHandler) CreateAlertRule(w http.ResponseWriter, r *http.Request) {
    var rule models.AlertRule
    if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate condition using existing ConditionValidator
    // Set tenant_id from context
    // Insert and return created rule
}

// ListAlertRules handles GET /api/v1/alert-rules
func (h *AlertRulesHandler) ListAlertRules(w http.ResponseWriter, r *http.Request) {
    // Fetch rules for current tenant
    // Return paginated list
}

// GetAlertHistory handles GET /api/v1/alerts/history
func (h *AlertRulesHandler) GetAlertHistory(w http.ResponseWriter, r *http.Request) {
    // Query alert_triggers with joins to alert_rules
    // Support filtering by severity, status, date range
    // Return list of AlertTrigger with rule details
}
```

### OpsGenie Channel Implementation
```go
// Add to: backend/internal/notifications/channels.go

type OpsGenieChannel struct {
    *BaseChannel
    httpClient *http.Client
}

type OpsGenieConfig struct {
    APIKey  string `json:"api_key"`
    Region  string `json:"region,omitempty"` // "us" or "eu"
    TeamID  string `json:"team_id,omitempty"`
}

type OpsGenieAlert struct {
    Message     string            `json:"message"`
    Alias       string            `json:"alias"`
    Description string            `json:"description"`
    Priority    string            `json:"priority"`
    Tags        []string          `json:"tags"`
    Details     map[string]string `json:"details"`
}

func (o *OpsGenieChannel) Type() string {
    return "opsgenie"
}

func (o *OpsGenieChannel) Send(ctx context.Context, alert *AlertNotification, config ChannelConfig) (*DeliveryResult, error) {
    if o.circuitBreaker.IsOpen() {
        return &DeliveryResult{Success: false, ErrorMsg: "Circuit open"}, nil
    }

    var ogConfig OpsGenieConfig
    json.Unmarshal(config.Config, &ogConfig)

    baseURL := "https://api.opsgenie.com"
    if ogConfig.Region == "eu" {
        baseURL = "https://api.eu.opsgenie.com"
    }

    priority := "P3"
    switch alert.Severity {
    case "critical": priority = "P1"
    case "high": priority = "P2"
    case "medium": priority = "P3"
    case "low": priority = "P4"
    }

    payload := OpsGenieAlert{
        Message:     alert.Title,
        Alias:       fmt.Sprintf("pganalytics_%d", alert.AlertID),
        Description: alert.Description,
        Priority:    priority,
        Tags:        []string{"pganalytics", alert.Severity},
    }

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/v2/alerts", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "GenieKey "+ogConfig.APIKey)

    resp, err := o.httpClient.Do(req)
    if err != nil {
        o.circuitBreaker.RecordFailure()
        return &DeliveryResult{Success: false, ErrorMsg: err.Error()}, nil
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        o.circuitBreaker.RecordFailure()
        return &DeliveryResult{Success: false, ErrorMsg: fmt.Sprintf("HTTP %d", resp.StatusCode)}, nil
    }

    o.circuitBreaker.RecordSuccess()
    return &DeliveryResult{Success: true}, nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Synchronous notification delivery | Async NotificationWorker with retry | Phase 3 | Non-blocking alert evaluation |
| Hardcoded notification channels | Database-configured channels | Phase 3 | User-managed integrations |
| Simple threshold alerts | Threshold + Change + Anomaly + Composite | Phase 4 | Flexible alert conditions |
| No deduplication | Fingerprint-based deduplication | Phase 4 | Reduced alert noise |
| Manual escalation | EscalationPolicy with steps | Phase 4 | Automated escalation workflows |

**Deprecated/outdated:**
- Direct SMTP in handlers: Replaced by async `NotificationWorker` processing queue
- In-memory alert state: Replaced by database-backed `escalation_state` table

## Open Questions

1. **Email Template Customization**
   - What we know: `FormatAlertHTML()` provides default HTML template
   - What's unclear: Should users be able to customize email templates?
   - Recommendation: Defer to v2. Use default templates for Phase 12.

2. **Alert Rule Scope**
   - What we know: Rules can target specific databases or instances
   - What's unclear: Should rules support label/tag-based targeting for dynamic scoping?
   - Recommendation: Use existing `database_id`, `query_id` columns. Add label-based targeting in future if needed.

3. **Notification Rate Limiting**
   - What we know: 5-minute deduplication window exists
   - What's unclear: Should there be global rate limits per channel?
   - Recommendation: Implement per-channel rate limiting as configuration option. Default: 10 notifications per minute per channel.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package + testify |
| Config file | None - uses build tags |
| Quick run command | `go test ./pkg/... -short` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ALERT-01 | Configure alert rules | unit | `go test ./pkg/handlers/... -run TestAlertRules -v` | Partial - needs creation |
| ALERT-02 | Email notifications | unit | `go test ./internal/notifications/... -run TestEmailChannel -v` | Partial - exists |
| ALERT-03 | Slack notifications | unit | `go test ./internal/notifications/... -run TestSlackChannel -v` | Yes |
| ALERT-04 | Webhook notifications | unit | `go test ./internal/notifications/... -run TestWebhookChannel -v` | Yes |
| ALERT-05 | PagerDuty/OpsGenie | unit | `go test ./internal/notifications/... -run TestPagerDutyChannel -v` | Partial |
| ALERT-06 | Alert history | integration | `go test ./tests/integration/... -run TestAlertHistory -v` | No - Wave 0 |
| ALERT-07 | Acknowledge/silence | unit | `go test ./pkg/services/... -run TestSilence -v` | Yes |
| ALERT-08 | Escalation policies | unit | `go test ./pkg/services/... -run TestEscalation -v` | Yes |
| REP-05 | Replication lag alerts | integration | `go test ./tests/integration/... -run TestReplicationAlerts -v` | No - Wave 0 |
| HOST-05 | Host-level thresholds | unit | `go test ./pkg/handlers/... -run TestHostAlerts -v` | No - Wave 0 |
| UI-02 | Alert rules UI | e2e | Playwright tests | No - Wave 0 |
| UI-05 | Notification channels UI | e2e | Playwright tests | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./pkg/... ./internal/... -short` (fast unit tests only)
- **Per wave merge:** `go test ./... -count=1` (full suite)
- **Phase gate:** Full suite green + integration tests passing

### Wave 0 Gaps
- [ ] `backend/pkg/handlers/alerts_test.go` - covers ALERT-01 CRUD operations
- [ ] `backend/internal/notifications/email_test.go` - covers ALERT-02 SMTP integration
- [ ] `backend/tests/integration/alert_history_test.go` - covers ALERT-06 history queries
- [ ] `backend/tests/integration/replication_alerts_test.go` - covers REP-05
- [ ] `backend/pkg/handlers/host_alerts_test.go` - covers HOST-05
- [ ] Playwright e2e tests for UI-02, UI-05

*(If no gaps: "None - existing test infrastructure covers all phase requirements")*

## Sources

### Primary (HIGH confidence)
- `/Users/glauco.torres/git/pganalytics-v3/backend/internal/jobs/alert_rule_engine.go` - Alert rule evaluation implementation
- `/Users/glauco.torres/git/pganalytics-v3/backend/internal/notifications/notification_service.go` - Multi-channel notification service
- `/Users/glauco.torres/git/pganalytics-v3/backend/internal/notifications/channels.go` - Channel implementations (Slack, Email, Webhook, PagerDuty, Jira)
- `/Users/glauco.torres/git/pganalytics-v3/backend/migrations/022_realtime_tables.sql` - Alert tables schema
- `/Users/glauco.torres/git/pganalytics-v3/backend/migrations/023_phase4_tables.sql` - Escalation and silence tables schema

### Secondary (MEDIUM confidence)
- `/Users/glauco.torres/git/pganalytics-v3/backend/pkg/services/escalation_service.go` - Escalation policy management
- `/Users/glauco.torres/git/pganalytics-v3/backend/pkg/services/silence_service.go` - Silence management
- `/Users/glauco.torres/git/pganalytics-v3/backend/pkg/models/models.go` (lines 1015-1142) - Alert model definitions
- `/Users/glauco.torres/git/pganalytics-v3/frontend/src/types/notifications.ts` - Frontend type definitions

### Tertiary (LOW confidence)
- None - all findings verified against source code

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - verified from existing go.mod and codebase
- Architecture: HIGH - comprehensive existing implementation analyzed
- Pitfalls: HIGH - derived from code review and existing patterns

**Research date:** 2026-05-14
**Valid until:** 30 days (stable architecture, existing implementation)