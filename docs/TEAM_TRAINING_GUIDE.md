# pgAnalytics v3 - Team Training & Operational Guide

**Version**: 1.0
**Date**: March 3, 2026
**Audience**: Database Engineering Team, On-Call DBAs, Operations
**Training Duration**: 4-6 hours total

---

## Table of Contents

1. System Overview & Architecture
2. Alert System Training
3. Incident Response Workflow
4. Automation System Usage
5. Hands-On Lab Exercises
6. On-Call Procedures
7. Escalation & Communication

---

## Part 1: System Overview & Architecture

### 1.1 What is pgAnalytics v3?

pgAnalytics v3 is a comprehensive PostgreSQL monitoring and automation system that:

- **Collects** 11+ database metrics every 60 seconds
- **Stores** time-series data in TimescaleDB
- **Visualizes** dashboards in Grafana
- **Alerts** on 11 alert rules with 9 notification channels
- **Auto-remediates** common issues (locks, bloat, connections)
- **Correlates** related alerts into incidents
- **Analyzes** root causes for faster resolution

### 1.2 Five-Phase Architecture

```
Phase 1: Collectors (✅ Complete)
  - 12 collector plugins on C++
  - Collects metrics from PostgreSQL
  - Sends data to backend API

Phase 2: Storage (✅ Complete)
  - PostgreSQL + TimescaleDB
  - Stores time-series metrics
  - Retention: 30 days default

Phase 3: REST API (✅ Complete)
  - Go/Gin web service
  - 50+ endpoints
  - Powers frontend and integrations

Phase 4: Dashboards (✅ Complete)
  - 7 Grafana dashboards
  - Real-time visualization
  - Query performance, bloat, cache, etc.

Phase 5: Alerting & Automation (✅ Complete)
  - Alert rules (11 total)
  - Notification channels (9 total)
  - Auto-remediation (6 actions)
  - Incident correlation (4 groups)
```

### 1.3 Key Components You'll Use

| Component | Purpose | Access |
|-----------|---------|--------|
| Grafana | View dashboards & alerts | http://grafana.internal |
| Alert Rules | Define detection thresholds | Grafana → Alerting |
| Notification Channels | Route alerts | Slack, PagerDuty, Email |
| Automation Engine | Execute remediation | Port 5002 (internal) |
| Incident Correlation | Group related alerts | Port 5003 (internal) |

---

## Part 2: Alert System Training

### 2.1 Understanding the 11 Alert Rules

**Critical Alerts (Page On-Call)**:

1. **Lock Contention** - Active locks > 10 for 5m
   - Indicator: App timeouts, slow queries
   - Auto-action: Kill blocking locks
   - Runbook: RUNBOOK_LOCK_CONTENTION.md

2. **Blocking Transaction** - Lock wait > 300s for 5m
   - Indicator: App timeout errors
   - Auto-action: Kill blocking locks
   - Runbook: RUNBOOK_LOCK_CONTENTION.md

3. **Collection Failure** - No metrics for 15m
   - Indicator: Grafana data gap
   - Auto-action: Restart collectors
   - Runbook: RUNBOOK_COLLECTOR_FAILURE.md

**Warning Alerts (Alert & Create Ticket)**:

4. **Idle-in-Transaction** - Count > 5 for 2m
   - Indicator: Locks held by idle connections
   - Auto-action: Close idle-txn connections
   - Runbook: RUNBOOK_CONNECTIONS.md

5. **High Table Bloat** - Dead ratio > 50% for 10m
   - Indicator: Slow queries, disk space
   - Auto-action: VACUUM ANALYZE
   - Runbook: RUNBOOK_TABLE_BLOAT.md

6. **Low Cache Hit Ratio** - Ratio < 80% for 30m
   - Indicator: Disk I/O, slow queries
   - Auto-action: Analysis only
   - Runbook: RUNBOOK_CACHE_OPTIMIZATION.md

7. **High Connection Count** - > 150 for 10m
   - Indicator: App connection pool issues
   - Auto-action: Close idle connections
   - Runbook: RUNBOOK_CONNECTIONS.md

8. **Max Lock Age** - Lock age > 300s for 5m
   - Indicator: Long-held locks
   - Auto-action: None (manual only)
   - Runbook: RUNBOOK_LOCK_CONTENTION.md

**Info Alerts (Log & Digest)**:

9. **Schema Growth** - Schema change detected
   - Indicator: New tables/indexes
   - Auto-action: Log to Slack
   - Runbook: None (informational)

10. **Unused Index** - Not scanned for 7+ days
    - Indicator: Unused indexes
    - Auto-action: Log to Slack
    - Runbook: None (informational)

11. **Extension Installation** - New extension detected
    - Indicator: Extension added
    - Auto-action: Log to Slack
    - Runbook: None (informational)

### 2.2 Alert Notification Channels

**Slack Channels** (3 channels):
```
#critical-alerts   → Critical alerts (real-time)
#database-alerts   → Warning alerts (batched 5m)
#database-info     → Info alerts (batched 1h)
```

**PagerDuty** (2 services):
```
Critical Service     → Critical alerts (immediate page)
Warning Service      → Warning alerts (standard escalation)
```

**Email** (2 recipients):
```
dba-team@company.com         → 1-hour digest (critical + warning)
operations-team@company.com  → Daily digest (all alerts)
```

**Webhooks** (2 receivers):
```
/correlation/correlate    → Incident grouping
/automation/remediate     → Auto-remediation
```

### 2.3 Alert Response Expectations

| Severity | Response Time | Action | Escalation |
|----------|---------------|--------|------------|
| CRITICAL | < 2 min | Acknowledge | Immediate if not acknowledged |
| WARNING | < 10 min | Investigate | 30 min if unresolved |
| INFO | < 1 hour | Review | No escalation |

---

## Part 3: Incident Response Workflow

### 3.1 What Happens When Alert Fires

```
1. Alert Condition Detected
   └─ Grafana evaluates rule every 60s

2. Notification Sent
   └─ Slack, PagerDuty, Email receive alert

3. Incident Correlation
   └─ System groups related alerts

4. Auto-Remediation (Optional)
   └─ If enabled, system attempts fix

5. On-Call Action
   └─ You investigate and respond

6. Resolution
   └─ Issue fixed, alert clears

7. Post-Incident Review
   └─ Document and prevent recurrence
```

### 3.2 Your Response Workflow

**Step 1: Acknowledge (0-2 min)**
```
- Receive alert notification
- Click link to incident
- Acknowledge in PagerDuty
- Note start time
```

**Step 2: Triage (2-5 min)**
```
- Open relevant runbook
- Assess severity
- Determine action needed
- Check if auto-remediation ran
```

**Step 3: Diagnose (5-15 min)**
```
- Run diagnostic queries
- Check recent changes
- Analyze root cause
- Determine remediation option
```

**Step 4: Remediate (15-30 min)**
```
- Execute remediation action
- Monitor results
- Verify resolution
- Document actions taken
```

**Step 5: Verify (30-45 min)**
```
- Confirm alert has cleared
- Verify application recovered
- Check system stability
- Resolve incident
```

**Step 6: Document (Next business day)**
```
- Update incident ticket
- Note root cause
- Document prevention
- Schedule follow-up if needed
```

### 3.3 Runbook Navigation

```
Critical Lock Issues
  → RUNBOOK_LOCK_CONTENTION.md
  → Flow: Diagnose → Kill Locks → Verify

Table Bloat Issues
  → RUNBOOK_TABLE_BLOAT.md
  → Flow: Diagnose → VACUUM → Verify

Connection Pool Issues
  → RUNBOOK_CONNECTIONS.md
  → Flow: Diagnose → Close Idle → Verify

Collection Failure
  → RUNBOOK_COLLECTOR_FAILURE.md
  → Flow: Check Collector → Restart → Verify

Cache Issues
  → RUNBOOK_CACHE_OPTIMIZATION.md
  → Flow: Analyze → Optimize → Verify
```

---

## Part 4: Automation System Usage

### 4.1 How Auto-Remediation Works

```
Alert Fires
  ↓
Decision Engine Evaluates
  ├─ Is auto-remediation enabled?
  ├─ Is this alert type configured?
  ├─ Is remediation already running?
  └─ Does threshold meet criteria?
  ↓
Remediation Action Executes
  └─ Lock killing, VACUUM, close connections, etc.
  ↓
Result Tracked
  └─ Success/failure logged
  ↓
On-Call Reviews Results
  └─ No action needed if successful
```

### 4.2 Dry-Run Mode (Testing)

**Enable safe testing without making changes**:

```bash
# Start automation engine in dry-run mode
DRY_RUN=true AUTO_REMEDIATE=false python automation_engine.py &

# All actions logged but not executed
# Perfect for testing and validation
```

### 4.3 Staged Rollout Recommendation

```
Week 1: DRY_RUN=true, AUTO_REMEDIATE=false
  └─ Only log intended actions
  └─ Review logs for accuracy
  └─ Duration: Full week

Week 2: DRY_RUN=false, AUTO_REMEDIATE=false
  └─ Actions logged, not executed
  └─ Manual review and approval
  └─ Duration: 3-5 days

Week 3: DRY_RUN=false, AUTO_REMEDIATE=true
  └─ Full automation enabled
  └─ Monitor results closely
  └─ Duration: Ongoing
```

### 4.4 Checking Remediation History

```bash
# View recent auto-remediation actions
curl -s http://localhost:5002/automation/history | jq '.results[] | {
  remediation_id,
  alert_name,
  action,
  status,
  timestamp
}'

# Check specific alert remediation
curl -s "http://localhost:5002/automation/history?alert_name=lock_contention_critical" | jq
```

---

## Part 5: Hands-On Lab Exercises

### Lab 1: Lock Contention Simulation

**Objective**: Understand lock detection and remediation

**Steps**:
1. Open two PostgreSQL connections
2. Connection 1: `BEGIN; UPDATE table SET col = 1 WHERE id = 1;`
3. Connection 2: `UPDATE table SET col = 2 WHERE id = 1;` (will hang)
4. Run lock diagnostic query from runbook
5. Identify blocking PID
6. Kill blocking transaction
7. Verify Connection 2 completes

**Expected Outcome**: Successfully identified and resolved lock contention

### Lab 2: Table Bloat Analysis

**Objective**: Understand bloat detection and VACUUM

**Steps**:
1. Create test table: `CREATE TABLE test_bloat AS SELECT * FROM pg_catalog.pg_class;`
2. Generate bloat: `UPDATE test_bloat SET relname = relname || 'x' FOR 10000 rows;`
3. Run bloat diagnostic query
4. Note dead tuples and bloat ratio
5. Execute: `VACUUM ANALYZE test_bloat;`
6. Re-run diagnostic query
7. Note recovery

**Expected Outcome**: Saw bloat accumulation and recovery via VACUUM

### Lab 3: Incident Correlation

**Objective**: Understand incident grouping

**Steps**:
1. Send two related alerts to correlation engine:
   ```bash
   curl -X POST http://localhost:5003/correlation/correlate \
     -H 'Content-Type: application/json' \
     -d '{"alert_name": "lock_contention_critical", "database": "production"}'

   curl -X POST http://localhost:5003/correlation/correlate \
     -H 'Content-Type: application/json' \
     -d '{"alert_name": "blocking_transaction_critical", "database": "production"}'
   ```
2. Get incident ID from first response
3. Check incident details
4. Verify both alerts in same incident
5. Get root cause analysis
6. Resolve incident

**Expected Outcome**: Related alerts grouped into single incident with analysis

### Lab 4: Alert Simulation

**Objective**: Test alert notification channels

**Steps**:
1. Trigger test alert in Grafana
2. Verify Slack notification in #database-alerts
3. Check PagerDuty event created
4. Verify email receipt (may be delayed)
5. Resolve alert in Grafana
6. Verify notification of resolution

**Expected Outcome**: All channels received alert and resolution

---

## Part 6: On-Call Procedures

### 6.1 Starting Your On-Call Shift

**Pre-shift Checklist** (15 minutes before shift start):

1. **Access Verification**
   ```bash
   # Test Grafana access
   curl -s http://grafana.internal/api/health | jq .

   # Test database access
   pg_isready -h production.db.internal -p 5432

   # Test automation engine
   curl -s http://localhost:5002/automation/health | jq .
   ```

2. **Notification Setup**
   - [ ] PagerDuty mobile app on phone
   - [ ] Slack desktop notifications enabled
   - [ ] Email notifications working
   - [ ] Phone ringer on

3. **Documentation Ready**
   - [ ] Runbooks accessible
   - [ ] Escalation contacts visible
   - [ ] Runbook quick links bookmarked

4. **Baseline Review**
   ```bash
   # Check current system status
   curl -s http://localhost:5003/correlation/incidents | jq '.incidents | length'
   # Should be 0-2 active incidents max
   ```

### 6.2 During Your Shift

**Every Hour**:
- [ ] Spot-check Grafana dashboards
- [ ] Review any resolved incidents
- [ ] Monitor alert frequency

**When Alert Fires**:
- [ ] Acknowledge within 2 minutes (critical) / 10 minutes (warning)
- [ ] Follow runbook steps
- [ ] Update ticket with progress
- [ ] Keep team informed via Slack

**Escalation**:
```
If stuck > 5 minutes (critical) or > 15 minutes (warning):
1. Reach out to team in Slack
2. Notify team lead
3. Follow escalation policy in runbook
```

### 6.3 Ending Your On-Call Shift

**Handoff Checklist**:

1. **Current Status**
   ```bash
   # List active incidents
   curl -s http://localhost:5003/correlation/incidents?state=active | jq

   # Document any in-progress work
   # Provide detailed update to next on-call
   ```

2. **Pending Issues**
   - [ ] Note any unresolved issues
   - [ ] Provide diagnostic findings
   - [ ] List next steps

3. **Communication**
   ```
   Slack message to next on-call:

   On-Call Handoff: [Your name] → [Next person]
   Time: [current time]

   Active Issues:
   - [Issue 1]: Status, next steps
   - [Issue 2]: Status, next steps

   Alerts to watch:
   - [Alert 1]: Known to fire frequently
   - [Alert 2]: Monitor this one

   Helpful tips:
   - [Tip 1]
   - [Tip 2]
   ```

4. **Documentation Update**
   - [ ] Update any runbooks with findings
   - [ ] Document new error patterns
   - [ ] Note any false positives

---

## Part 7: Escalation & Communication

### 7.1 Escalation Policy

**Level 1: On-Call DBA** (You)
- Respond to alerts
- Execute runbook steps
- Attempt auto-remediation
- Gather diagnostics

**Level 2: Team Lead**
- Triggered if: Unresolved after 5 min (critical) / 15 min (warning)
- Actions: Provide guidance, approve actions
- Contact: [Lead phone/Slack]

**Level 3: Database Architect**
- Triggered if: Systemic issue, recurring problem
- Actions: Root cause analysis, long-term fix
- Contact: [Architect phone/Slack]

**Level 4: VP Engineering**
- Triggered if: Data loss risk, major outage
- Actions: Executive decision, resource allocation
- Contact: [VP contact]

### 7.2 Communication Templates

**Initial Acknowledgment**:
```
🚨 Incident Acknowledged

Alert: [Alert Name]
Severity: [CRITICAL/WARNING/INFO]
Time: [HH:MM UTC]

Investigating now...
Updates every 5 minutes.
```

**Progress Update**:
```
🔍 Incident Update

Alert: [Alert Name]
Duration: [elapsed time]

Findings:
- [Finding 1]
- [Finding 2]

Next Action:
- [Action 1]

ETA: [estimated resolution time]
```

**Resolution**:
```
✅ Incident Resolved

Alert: [Alert Name]
Duration: [total time]
Root Cause: [brief description]

Resolution:
- [Action taken 1]
- [Action taken 2]

Prevention:
- [Planned fix 1]

Closing ticket. Monitoring for recurrence.
```

### 7.3 Communication Channels

| Situation | Channel | Urgency |
|-----------|---------|---------|
| Critical issue < 2 min response | PagerDuty + Slack | IMMEDIATE |
| Critical issue, acknowledged | Slack #critical-alerts | HIGH |
| Warning issue | Slack #database-alerts | NORMAL |
| Info issue | Email + Slack #database-info | LOW |
| Non-production issue | Email + Slack #database-dev | LOW |

---

## Part 8: Quick Reference

### Alert Priority Matrix

```
Severity  | Response  | Action              | Escalate
----------|-----------|---------------------|----------
CRITICAL  | < 2 min   | Kill locks/restart  | 5 min
WARNING   | < 10 min  | VACUUM/close conn   | 15 min
INFO      | < 1 hour  | Review/log          | None
```

### Useful Commands Cheat Sheet

```bash
# Check system health
curl -s http://localhost:5002/automation/health | jq .

# View active incidents
curl -s http://localhost:5003/correlation/incidents | jq .

# Check remediation history
curl -s http://localhost:5002/automation/history | jq .

# Get incident analysis
curl -s http://localhost:5003/correlation/incident/[ID] | jq .

# Connect to database
psql -h production.db.internal -U pganalytics -d pganalytics

# Quick lock status
SELECT COUNT(*) FROM pg_locks WHERE NOT granted;

# Quick bloat check
SELECT tablename, dead_ratio FROM pg_stat_user_tables ORDER BY dead_ratio DESC;

# Connection status
SELECT state, COUNT(*) FROM pg_stat_activity GROUP BY state;
```

---

## Conclusion

**You are now trained to**:
- ✅ Understand the pgAnalytics architecture
- ✅ Respond to alerts using runbooks
- ✅ Use automation and correlation systems
- ✅ Follow on-call procedures
- ✅ Escalate appropriately
- ✅ Communicate effectively

**Next Steps**:
1. Schedule hands-on lab exercises
2. Shadow current on-call (1-2 shifts)
3. Take on-call independently
4. Provide feedback and improvement ideas

**Questions?**
- Team Slack: #database-oncall
- Wiki: [Team Wiki Link]
- Email: [Team Email]

---

Generated: March 3, 2026
Author: Claude Opus 4.6
Training Version: 1.0
