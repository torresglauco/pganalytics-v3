# Phase 4: Advanced UI Features - Design Specification

**Version:** 1.0
**Date:** 2026-03-13
**Status:** Approved
**Author:** Claude Code + Team

---

## Executive Summary

Phase 4 delivers three interconnected features that transform pgAnalytics into an enterprise-grade alerting system with intuitive UI controls:

1. **Custom Alert Conditions UI** - Visual builder for creating alert rules without JSON
2. **Alert Silencing/Snoozing** - Quick controls to suppress alerts during maintenance
3. **Escalation Policies** - Multi-step notification escalation with acknowledgment tracking

**Deliverables:**
- 5 new React components for frontend
- 4 new Go services for backend
- 5 new database tables
- 8 new API endpoints
- 40+ unit/integration tests
- Complete documentation

**Timeline:** 4-5 weeks (23 days effort)
**Status:** Ready for implementation planning

---

## 1. Custom Alert Conditions UI

### 1.1 Purpose

Currently, alert rules require manual JSON creation. This feature provides a visual builder that democratizes alert creation for non-technical users while maintaining full expressiveness for advanced users.

### 1.2 User Experience

**Create New Alert Flow:**
1. User navigates to "Create Alert Rule"
2. Fills basic info: Name, Description, Alert Type (Error Count, Slow Query, etc.)
3. Opens Condition Builder:
   - Drag-drop interface
   - Add conditions: "ERROR_COUNT > 5"
   - Add operators: "AND", "OR", "NOT"
   - Set time windows: "in last 10 minutes"
4. Preview shows human-readable version: "Alert if ERROR_COUNT exceeds 5 in last 10 minutes"
5. Click Save → Backend validates JSON → Creates alert_rule record

### 1.3 Technical Design

**Frontend Components:**

```
AlertRuleBuilder (container)
├── RuleBasics (name, description, alert type)
├── ConditionBuilder (visual editor)
│   ├── ConditionBlock (individual condition)
│   ├── LogicalOperator (AND/OR/NOT selector)
│   ├── ThresholdSelector (dropdown + number input)
│   └── TimeWindowSelector (5 min, 10 min, 1 hour, custom)
├── ConditionPreview (plain English display)
└── SaveButton → API call
```

**Supported Conditions:**
- Metric-based: `ERROR_COUNT > threshold`
- Time-based: `in last X minutes`
- Duration: `for at least X minutes`
- Multiple: Can combine with AND/OR/NOT

**Backend Validation:**
```go
type Condition struct {
  MetricType    string      // "error_count", "slow_query_count", etc
  Operator      string      // ">", "<", "==", "!="
  Threshold     float64
  TimeWindow    int         // minutes
  Duration      int         // minutes (alert must be true for this long)
}

type ConditionValidator struct {}

func (v *ConditionValidator) Validate(condition Condition) error {
  // Ensure metric_type is valid
  // Ensure operator is supported
  // Ensure threshold makes sense for metric type
  // Ensure time windows are reasonable
}
```

**Database Changes:**
```sql
ALTER TABLE alert_rules ADD COLUMN (
  condition_json JSONB,         -- Structured condition
  condition_display TEXT,       -- Human-readable: "ERROR > 5 in 10min"
  builder_version INT DEFAULT 1 -- For schema migrations
);
```

### 1.4 Success Criteria

✅ Users can create rules via UI without touching JSON
✅ Condition preview matches what actually happens
✅ Backend validates all conditions before save
✅ Validation errors shown to user with helpful messages
✅ 90%+ test coverage for condition validator

---

## 2. Alert Silencing/Snoozing

### 2.1 Purpose

During deployments or maintenance, alerts fire but shouldn't notify the team. This feature allows one-click silencing with automatic expiration.

### 2.2 User Experience

**Silence Alert Flow:**
1. Alert fires and notification sent
2. User sees alert in UI with "Silence" button
3. Clicks button → shows options:
   - Quick: "1 hour", "4 hours", "24 hours", "1 week"
   - Custom: Date/time picker
   - Scope: "This alert only", "This rule (all instances)", "All rules"
4. Optional: Add reason ("Deploying feature X")
5. Click Silence → Alert worker skips future notifications until expiration
6. After expiration, notifications resume automatically

**Visual Feedback:**
- Silenced alert shows "Silenced until 3:00 PM" with countdown
- Silence badge in alert list
- WebSocket notification to all users when silenced/unsil​enced

### 2.3 Technical Design

**Database Schema:**
```sql
CREATE TABLE alert_silences (
  id BIGSERIAL PRIMARY KEY,
  alert_rule_id BIGINT NOT NULL REFERENCES alert_rules(id),
  instance_id INT,              -- NULL = all instances
  silenced_until TIMESTAMP NOT NULL,
  silence_type VARCHAR(20),     -- 'rule', 'instance', 'all'
  reason TEXT,
  created_by INT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_alert_silences_lookup
  ON alert_silences(alert_rule_id, instance_id, silenced_until);
```

**Backend Service:**
```go
type SilenceService struct {
  db *sql.DB
  ws *WebSocketManager
}

func (s *SilenceService) CreateSilence(ruleID int, minutes int, silenceType string) error {
  silence := AlertSilence{
    AlertRuleID: ruleID,
    SilencedUntil: time.Now().Add(time.Duration(minutes) * time.Minute),
    SilenceType: silenceType,
  }

  err := s.db.SaveSilence(silence)
  if err != nil {
    return err
  }

  // Broadcast to frontend via WebSocket
  s.ws.Broadcast("silence:created", silence)
  return nil
}

func (s *SilenceService) IsSilenced(ruleID, instanceID int) bool {
  // Check if active silence exists for this rule/instance
  return s.db.CheckActiveSilence(ruleID, instanceID)
}
```

**Integration with Alert Worker:**
```go
// In alert_worker.go, before sending notification:
if s.silenceService.IsSilenced(alert.RuleID, alert.InstanceID) {
  log.Printf("Alert %d silenced, skipping notification", alert.ID)
  return nil
}

// Send notification
return s.sendNotification(alert)
```

**Frontend Component:**
```tsx
// SilenceModal.tsx
export const SilenceModal: React.FC<{ alertRuleID: number }> = ({ alertRuleID }) => {
  const [duration, setDuration] = useState(60) // minutes
  const [reason, setReason] = useState('')
  const [silenceType, setSilenceType] = useState('rule')

  const handleSilence = async () => {
    const response = await fetch(`/api/v1/alerts/${alertRuleID}/silence`, {
      method: 'POST',
      body: JSON.stringify({ duration, reason, silenceType })
    })

    if (response.ok) {
      toast.success(`Alert silenced for ${duration} minutes`)
      onClose()
    }
  }

  return (
    <Modal>
      <h2>Silence Alert</h2>

      <div>
        <label>Duration:</label>
        <div>
          <button onClick={() => setDuration(60)}>1 hour</button>
          <button onClick={() => setDuration(240)}>4 hours</button>
          <button onClick={() => setDuration(1440)}>24 hours</button>
          <button onClick={() => setDuration(10080)}>1 week</button>
        </div>
      </div>

      <div>
        <label>Scope:</label>
        <select value={silenceType} onChange={(e) => setSilenceType(e.target.value)}>
          <option value="alert">This alert only</option>
          <option value="rule">This rule (all instances)</option>
          <option value="all">All rules</option>
        </select>
      </div>

      <div>
        <label>Reason (optional):</label>
        <textarea
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Deploying feature X..."
        />
      </div>

      <button onClick={handleSilence}>Silence</button>
    </Modal>
  )
}
```

### 2.4 Success Criteria

✅ Silence prevents notifications from being sent
✅ Auto-expiration works without manual intervention
✅ Can silence at rule level or instance level
✅ UI shows countdown until silence expires
✅ WebSocket broadcasts silence state to all users
✅ Silence reason tracked for audit

---

## 3. Escalation Policies

### 3.1 Purpose

Critical alerts need to reach the right person at the right time. Escalation policies automatically escalate notifications through multiple channels (Slack → PagerDuty → SMS → etc) if not acknowledged.

### 3.2 User Experience

**Create Escalation Policy Flow:**
1. Navigate to "Create Escalation Policy"
2. Add steps (drag-drop to reorder):
   - **Step 1 (Immediate):** Slack to #critical-alerts
   - **Step 2 (5 min if no ACK):** PagerDuty page to on-call
   - **Step 3 (15 min if still no ACK):** SMS to manager
   - **Step 4 (30 min if still no ACK):** Call escalation lead
3. Each step shows: channel, who/where it goes, delay, requires ack?
4. Preview shows timeline: "Step 1 (now) → Step 2 (5 min) → Step 3 (15 min)"
5. Save → Associates with alert rules
6. When alert triggers:
   - Step 1 sent immediately
   - 5 min timer starts for Step 2
   - If user ack's in Slack → stops escalation
   - If no ack → Step 2 sent
   - Continue until ack'd or all steps exhausted

### 3.3 Technical Design

**Database Schema:**
```sql
CREATE TABLE escalation_policies (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  is_active BOOLEAN DEFAULT TRUE,
  created_by INT REFERENCES users(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE escalation_policy_steps (
  id BIGSERIAL PRIMARY KEY,
  policy_id BIGINT NOT NULL REFERENCES escalation_policies(id),
  step_order INT NOT NULL,
  channel_type VARCHAR(50),         -- 'slack', 'pagerduty', 'email', 'sms', 'webhook'
  channel_config JSONB,             -- channel-specific config (webhook URL, SMS number, etc)
  delay_minutes INT NOT NULL,       -- 0 = immediate, 5 = 5 min delay
  requires_acknowledgment BOOLEAN,  -- does this step need ACK to stop escalation?
  created_at TIMESTAMP DEFAULT NOW(),

  UNIQUE(policy_id, step_order),
  CONSTRAINT valid_delay CHECK (delay_minutes >= 0)
);

CREATE TABLE alert_rule_escalation_policies (
  alert_rule_id BIGINT NOT NULL REFERENCES alert_rules(id),
  policy_id BIGINT NOT NULL REFERENCES escalation_policies(id),
  PRIMARY KEY (alert_rule_id, policy_id)
);

CREATE TABLE escalation_state (
  id BIGSERIAL PRIMARY KEY,
  alert_trigger_id BIGINT NOT NULL REFERENCES alert_triggers(id),
  policy_id BIGINT NOT NULL REFERENCES escalation_policies(id),
  current_step INT NOT NULL DEFAULT 0,
  ack_received BOOLEAN DEFAULT FALSE,
  ack_by INT REFERENCES users(id),
  ack_at TIMESTAMP,
  last_escalated_at TIMESTAMP,
  next_escalation_at TIMESTAMP,
  status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'acknowledged', 'resolved'
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_escalation_state_pending
  ON escalation_state(next_escalation_at)
  WHERE status = 'pending';
```

**Backend Services:**
```go
type EscalationService struct {
  db        *sql.DB
  notifier  NotificationService
  logger    Logger
}

func (s *EscalationService) StartEscalation(triggerID int64, policyID int64) error {
  // Create escalation state tracking record
  state := EscalationState{
    AlertTriggerID: triggerID,
    PolicyID: policyID,
    CurrentStep: 0,
    Status: "pending",
    NextEscalationAt: time.Now(), // Start immediately
  }

  return s.db.SaveEscalationState(state)
}

func (s *EscalationService) AcknowledgeAlert(triggerID int64, userID int) error {
  // Find escalation state for this trigger
  state, err := s.db.GetEscalationState(triggerID)
  if err != nil {
    return err
  }

  // Mark as acknowledged
  state.AckReceived = true
  state.AckBy = userID
  state.AckAt = time.Now()
  state.Status = "acknowledged"

  // Stop escalation (don't send more steps)
  return s.db.UpdateEscalationState(state)
}
```

**Escalation Worker (background job, runs every 30 seconds):**
```go
type EscalationWorker struct {
  db        *sql.DB
  notifier  NotificationService
  policies  *PolicyService
  logger    Logger
}

func (w *EscalationWorker) Process() error {
  // Find all pending escalations where next_escalation_at <= NOW
  pending, err := w.db.GetPendingEscalations()
  if err != nil {
    w.logger.Error("Failed to fetch pending escalations", err)
    return err
  }

  for _, state := range pending {
    if time.Now().Before(state.NextEscalationAt) {
      continue // Not time yet
    }

    if state.Status == "acknowledged" {
      // User acknowledged, don't escalate further
      continue
    }

    // Get policy and current step
    policy, err := w.policies.GetPolicy(state.PolicyID)
    if err != nil {
      w.logger.Error("Failed to get policy", err)
      continue
    }

    if state.CurrentStep >= len(policy.Steps) {
      // All steps exhausted
      state.Status = "exhausted"
      w.db.UpdateEscalationState(state)
      continue
    }

    step := policy.Steps[state.CurrentStep]

    // Send this step's notification
    err = w.notifier.SendNotification(NotificationRequest{
      AlertTriggerID: state.AlertTriggerID,
      Channel: step.ChannelType,
      Config: step.ChannelConfig,
      StepNumber: state.CurrentStep + 1,
    })

    if err != nil {
      w.logger.Error("Failed to send escalation notification", err)
      // Retry next cycle
      continue
    }

    // Schedule next step
    nextStep := state.CurrentStep + 1
    nextEscalationTime := time.Now().Add(time.Duration(step.DelayMinutes) * time.Minute)

    state.CurrentStep = nextStep
    state.LastEscalatedAt = time.Now()
    state.NextEscalationAt = nextEscalationTime

    w.db.UpdateEscalationState(state)
  }

  return nil
}
```

**Frontend Component:**
```tsx
// EscalationPolicyBuilder.tsx
export const EscalationPolicyBuilder: React.FC = () => {
  const [policy, setPolicy] = useState<EscalationPolicy>({
    name: '',
    description: '',
    steps: []
  })

  const addStep = () => {
    setPolicy({
      ...policy,
      steps: [...policy.steps, {
        stepOrder: policy.steps.length + 1,
        channelType: 'slack',
        delayMinutes: 0,
        requiresAcknowledgment: true,
        channelConfig: {}
      }]
    })
  }

  const removeStep = (index: number) => {
    setPolicy({
      ...policy,
      steps: policy.steps.filter((_, i) => i !== index)
    })
  }

  const updateStep = (index: number, field: string, value: any) => {
    const newSteps = [...policy.steps]
    newSteps[index] = { ...newSteps[index], [field]: value }
    setPolicy({ ...policy, steps: newSteps })
  }

  const handleSave = async () => {
    const response = await fetch('/api/v1/escalation-policies', {
      method: 'POST',
      body: JSON.stringify(policy)
    })

    if (response.ok) {
      toast.success('Escalation policy created')
      navigate('/escalation-policies')
    }
  }

  return (
    <div>
      <h1>Create Escalation Policy</h1>

      <input
        value={policy.name}
        onChange={(e) => setPolicy({ ...policy, name: e.target.value })}
        placeholder="Policy name (e.g., 'Critical Alert Escalation')"
      />

      <textarea
        value={policy.description}
        onChange={(e) => setPolicy({ ...policy, description: e.target.value })}
        placeholder="Description"
      />

      <h3>Escalation Steps</h3>
      <EscalationTimeline steps={policy.steps} />

      {policy.steps.map((step, index) => (
        <EscalationStepEditor
          key={index}
          step={step}
          stepNumber={index + 1}
          onUpdate={(field, value) => updateStep(index, field, value)}
          onRemove={() => removeStep(index)}
        />
      ))}

      <button onClick={addStep}>Add Step</button>
      <button onClick={handleSave}>Save Policy</button>
    </div>
  )
}

// EscalationStepEditor.tsx
const EscalationStepEditor: React.FC<StepEditorProps> = ({
  step,
  stepNumber,
  onUpdate,
  onRemove
}) => {
  return (
    <div style={{ border: '1px solid #ddd', padding: '16px', margin: '12px 0' }}>
      <h4>Step {stepNumber}</h4>

      <label>Delay (minutes):</label>
      <input
        type="number"
        value={step.delayMinutes}
        onChange={(e) => onUpdate('delayMinutes', parseInt(e.target.value))}
      />

      <label>Channel:</label>
      <select
        value={step.channelType}
        onChange={(e) => onUpdate('channelType', e.target.value)}
      >
        <option value="slack">Slack</option>
        <option value="pagerduty">PagerDuty</option>
        <option value="email">Email</option>
        <option value="sms">SMS</option>
        <option value="webhook">Webhook</option>
      </select>

      <label>
        <input
          type="checkbox"
          checked={step.requiresAcknowledgment}
          onChange={(e) => onUpdate('requiresAcknowledgment', e.target.checked)}
        />
        Requires Acknowledgment
      </label>

      <ChannelConfigEditor
        channelType={step.channelType}
        config={step.channelConfig}
        onChange={(config) => onUpdate('channelConfig', config)}
      />

      <button onClick={onRemove}>Remove Step</button>
    </div>
  )
}

// EscalationTimeline.tsx
const EscalationTimeline: React.FC<{ steps: Step[] }> = ({ steps }) => {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '16px', margin: '16px 0' }}>
      {steps.map((step, index) => (
        <div key={index}>
          <div style={{ textAlign: 'center', fontWeight: 'bold' }}>
            {step.delayMinutes === 0 ? 'Now' : `+${step.delayMinutes}m`}
          </div>
          <div style={{ fontSize: '12px', textAlign: 'center' }}>
            {step.channelType}
          </div>
          {index < steps.length - 1 && <div>→</div>}
        </div>
      ))}
    </div>
  )
}
```

**Acknowledgment Flow:**
```tsx
// AckButton.tsx - shown in alert modal
export const AckButton: React.FC<{ alertTriggerID: number }> = ({ alertTriggerID }) => {
  const [isLoading, setIsLoading] = useState(false)
  const [isAck'd, setIsAck'd] = useState(false)

  const handleAck = async () => {
    setIsLoading(true)
    try {
      const response = await fetch(
        `/api/v1/alerts/${alertTriggerID}/acknowledge`,
        { method: 'POST' }
      )

      if (response.ok) {
        setIsAck'd(true)
        toast.success('Alert acknowledged - escalation stopped')
      }
    } finally {
      setIsLoading(false)
    }
  }

  if (isAck'd) {
    return <div>✓ Acknowledged</div>
  }

  return (
    <button onClick={handleAck} disabled={isLoading}>
      {isLoading ? 'Acknowledging...' : 'Acknowledge'}
    </button>
  )
}
```

### 3.4 Success Criteria

✅ Escalation steps execute in correct order with correct delays
✅ Acknowledgment stops escalation immediately
✅ WebSocket notifies frontend when acknowledged
✅ Multiple escalation policies can be created and managed
✅ Alert timeline shows which steps were escalated
✅ Background worker is reliable and handles failures gracefully
✅ 90%+ test coverage including background job logic

---

## 4. API Endpoints

### 4.1 Custom Conditions API

```
POST /api/v1/alert-rules/validate
  Request: { condition_json: {...} }
  Response: { valid: true/false, errors?: [...] }
  Purpose: Real-time validation of condition JSON

POST /api/v1/alert-rules
  Request: { name, description, condition_json, condition_display, policy_ids: [...] }
  Response: { id, created_at, ... }
  Purpose: Create new alert rule with conditions

PATCH /api/v1/alert-rules/{id}
  Request: { name?, description?, condition_json?, condition_display? }
  Response: { id, updated_at, ... }
  Purpose: Update existing rule conditions
```

### 4.2 Silencing API

```
POST /api/v1/alerts/{rule_id}/silence
  Request: { duration: 60, reason?: "...", silence_type: "rule" }
  Response: { id, silenced_until, ... }
  Purpose: Create new silence

DELETE /api/v1/silences/{id}
  Response: { success: true }
  Purpose: Cancel silence early

GET /api/v1/silences
  Query: ?active=true&rule_id=123
  Response: [{ id, rule_id, silenced_until, ... }]
  Purpose: List silences
```

### 4.3 Escalation API

```
POST /api/v1/escalation-policies
  Request: { name, description, steps: [{channel_type, delay_minutes, ...}] }
  Response: { id, created_at, ... }
  Purpose: Create escalation policy

PATCH /api/v1/escalation-policies/{id}
  Request: { name?, description?, steps? }
  Response: { id, updated_at, ... }
  Purpose: Update policy

POST /api/v1/escalation-policies/{id}/steps
  Request: { step_order, channel_type, delay_minutes, ... }
  Response: { id, ... }
  Purpose: Add step to policy

DELETE /api/v1/escalation-policies/{id}/steps/{step_id}
  Response: { success: true }
  Purpose: Remove step from policy

GET /api/v1/escalation-policies
  Response: [{ id, name, steps: [...], ... }]
  Purpose: List all policies

POST /api/v1/alerts/{trigger_id}/acknowledge
  Request: { reason?: "..." }
  Response: { success: true, ack_at, ... }
  Purpose: Acknowledge alert (stops escalation)

GET /api/v1/alerts/{trigger_id}/timeline
  Response: [
    { event: 'triggered', at: '...', ... },
    { event: 'escalated_step_1', at: '...', ... },
    { event: 'acknowledged', at: '...', by: 'user@...', ... }
  ]
  Purpose: Get audit trail of alert lifecycle
```

---

## 5. Integration with Phase 3

Phase 4 builds on Phase 3 infrastructure without breaking changes:

- **Alert Worker**: Already evaluates conditions, now also checks silences
- **Notification Worker**: Unchanged, continues delivering notifications
- **WebSocket**: Broadcasts new events: `silence:created`, `silence:expired`, `escalation:stepped`, `alert:acknowledged`
- **Database**: 5 new tables, no schema changes to existing tables (only additions)

---

## 6. Testing Strategy

### 6.1 Unit Tests
- ConditionValidator: Valid/invalid conditions
- SilenceService: Silence logic, expiration
- EscalationService: Policy evaluation
- React components: User interactions, form validation

### 6.2 Integration Tests
- Full flow: condition + silence + escalation
- Alert triggered → Step 1 sent → No ACK → Step 2 sent → ACK → stops escalation
- Multiple rules with different policies

### 6.3 E2E Tests
- UI: Create rule → Save → Alert triggers → Silence → Escalate → Acknowledge
- Timeline shows all events
- WebSocket updates in real-time

### 6.4 Load Tests
- 1000+ concurrent escalations
- Background worker handles high volume
- Database indices optimized

**Target Coverage:** 90%+ code coverage for all new code

---

## 7. Success Metrics

| Metric | Target |
|--------|--------|
| Users can create alerts without JSON | 100% |
| Alert silencing prevents notifications | 100% |
| Escalation steps execute on time | 99%+ |
| Acknowledgment stops escalation | 99%+ |
| Background worker uptime | 99.9%+ |
| API response time (p95) | <500ms |
| Code coverage | 90%+ |

---

## 8. Timeline & Effort

| Phase | Duration | Effort |
|-------|----------|--------|
| Database Schema + Migration | 2 days | 16 hours |
| Backend Services | 5 days | 40 hours |
| EscalationWorker | 3 days | 24 hours |
| Frontend Components | 5 days | 40 hours |
| API Endpoints | 2 days | 16 hours |
| Testing (Unit + Integration + E2E) | 4 days | 32 hours |
| Documentation | 2 days | 16 hours |
| **TOTAL** | **~23 days** | **~184 hours** |

---

## 9. Implementation Dependencies

**Required Before Starting:**
- Phase 3 complete and stable ✅
- Database access ✅
- Go backend running ✅
- React frontend running ✅
- WebSocket infrastructure ✅

**Can be Developed In Parallel:**
- Backend services (3-5 can work in parallel)
- Frontend components (can work independently)
- Tests (after corresponding component)

---

## 10. Known Constraints & Assumptions

**Constraints:**
- All data persisted to PostgreSQL (no Redis yet)
- Single instance deployment (horizontal scaling in Phase 5)
- Escalation policies created by admins (not self-service)

**Assumptions:**
- Alert rules already exist (Phase 3)
- WebSocket connection is reliable (Phase 3)
- Users have proper permissions (admin/user roles)
- Background workers run continuously (no serverless)

---

## 11. Future Enhancements (Phase 5+)

- Event streaming via Kafka for high-scale escalations
- Escalation policy templates (pre-built policies)
- Conditional escalation ("Only escalate if team X is available")
- Escalation metrics and SLA tracking
- Machine learning to optimize escalation timing
- Integration with external incident management (Opsgenie, etc)

---

## 12. Glossary

- **Condition**: Boolean expression that triggers an alert
- **Escalation Policy**: Multi-step notification plan
- **Silence**: Temporary suppression of notifications
- **Acknowledgment**: User confirms they received/are handling alert
- **Step**: Single notification in escalation sequence
- **Escalation Worker**: Background job that executes escalations

---

**Approval:** ✅ Design approved and ready for implementation planning
