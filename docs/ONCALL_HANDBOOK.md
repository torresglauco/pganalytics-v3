# On-Call DBA Handbook

**Version**: 1.0
**Last Updated**: March 3, 2026
**Purpose**: Quick reference for on-call database engineers

---

## Table of Contents

1. Quick Start (First 5 minutes)
2. Alert Response Flowchart
3. Emergency Contacts
4. Common Issues & Fixes
5. System Access
6. Safety Rules

---

## 1. Quick Start (First 5 Minutes)

### When You Receive an Alert

```
1. ACKNOWLEDGE (Within 2 min for critical, 10 min for warning)
   PagerDuty → Click "Acknowledge"

2. OPEN INCIDENT
   Slack → Click alert link → Opens Grafana

3. ASSESS SEVERITY
   Look at alert details:
   - Severity: CRITICAL / WARNING / INFO
   - Duration: How long it's been firing
   - Database: Which system affected
   - Indicator: What's wrong

4. FIND RUNBOOK
   All runbooks in: /Users/glauco.torres/git/pganalytics-v3/docs/

   Lock issues      → RUNBOOK_LOCK_CONTENTION.md
   Bloat issues     → RUNBOOK_TABLE_BLOAT.md
   Connection issues→ RUNBOOK_CONNECTIONS.md
   Collection fail  → RUNBOOK_COLLECTOR_FAILURE.md

5. FOLLOW RUNBOOK
   Execute steps in order
   Don't skip steps!
```

---

## 2. Alert Response Flowchart

```
┌─ Alert Received
│
├─ CRITICAL? (lock, blocking, collection failure)
│  ├─ YES: Acknowledge within 2 minutes
│  └─ NO: Acknowledge within 10 minutes
│
├─ Check Auto-Remediation History
│  ├─ Already resolved? → Monitor & close ticket
│  └─ Not resolved? → Continue with runbook
│
├─ Execute Runbook Steps
│  ├─ Diagnose (run queries)
│  ├─ Decide (manual or auto)
│  ├─ Execute (remediation action)
│  └─ Verify (confirm resolution)
│
├─ Verify Alert Cleared
│  ├─ Alert gone? → Yes, continue
│  └─ Still firing? → Runbook step 5 (root cause)
│
├─ Root Cause Analysis
│  ├─ Found cause? → Escalate if needed
│  └─ Can fix? → Implement prevention
│
└─ Close & Document
   ├─ Update ticket with findings
   ├─ Document root cause
   └─ Note prevention steps
```

---

## 3. Emergency Contacts

### Team Escalation

```
Level 1: You (On-Call DBA)
  - Response: 0-2 minutes
  - Phone: Your phone (silent → ring on now!)

Level 2: Team Lead
  - Name: [Team Lead Name]
  - Phone: [+1-XXX-XXX-XXXX]
  - Slack: @[team-lead-slack]
  - When: Alert unresolved > 5 min (critical) / 15 min (warning)

Level 3: Database Architect
  - Name: [Architect Name]
  - Phone: [+1-XXX-XXX-XXXX]
  - Slack: @[architect-slack]
  - When: Systemic issue, recurring problem

Level 4: VP Engineering
  - Name: [VP Name]
  - Phone: [+1-XXX-XXX-XXXX]
  - Email: [vp-email]
  - When: Data loss risk, major outage (> 15 min)
```

### External Escalation

```
Application Team Lead
  - Name: [App Lead]
  - Phone: [+1-XXX-XXX-XXXX]
  - When: Connection pool issues, app-caused locks

Infrastructure Team
  - Name: [Infra Contact]
  - Phone: [+1-XXX-XXX-XXXX]
  - When: Network issues, host problems

AWS/Cloud Support
  - Account ID: [Account ID]
  - Support Plan: Business
  - Portal: https://console.aws.amazon.com/support
  - When: Infrastructure issues
```

### Communication Channels

```
Immediate
  → PagerDuty mobile app (highest priority)
  → Personal phone call

Within 5 minutes
  → Slack @database-oncall
  → SMS if critical

For escalation
  → Email [team-distribution-list]
  → Weekly on-call summary
```

---

## 4. Common Issues & Fixes

### Issue 1: Lock Contention

**Quick Diagnosis**:
```sql
SELECT COUNT(*) FROM pg_locks WHERE NOT granted;
-- If > 0: You have locks
```

**Quick Fix**:
```bash
# Let auto-remediation handle it (if enabled)
# Or manually kill blocking locks:
curl -X POST http://localhost:5002/automation/remediate \
  -H 'Content-Type: application/json' \
  -d '{"alert_name": "lock_contention_critical", "severity": "critical", "database": "production"}'
```

**Prevention**:
- Add connection timeouts to app
- Monitor query execution times
- Review blocking query patterns

---

### Issue 2: Table Bloat

**Quick Diagnosis**:
```sql
SELECT tablename, bloat_ratio FROM pg_stat_user_tables ORDER BY bloat_ratio DESC LIMIT 5;
```

**Quick Fix**:
```bash
# Let auto-remediation handle it
# Or manually trigger VACUUM:
psql -h $DB_HOST -U pganalytics -d pganalytics -c "
VACUUM ANALYZE public.bloated_table_name;
"
```

**Prevention**:
- Tune autovacuum settings
- Monitor bloat trends
- Schedule VACUUM FULL during maintenance

---

### Issue 3: High Connection Count

**Quick Diagnosis**:
```sql
SELECT COUNT(*) FROM pg_stat_activity;
-- Over 150? That's the alert threshold
```

**Quick Fix**:
```bash
# Let auto-remediation close idle connections
# Or check for idle-in-transaction:
SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'idle in transaction';
-- Close these if > 10 minutes old
```

**Prevention**:
- Reduce connection pool size in app
- Implement connection timeout
- Monitor connection patterns

---

### Issue 4: Collection Failure

**Quick Diagnosis**:
```bash
# Check if collector is running
curl -s http://localhost:5002/automation/health | jq .

# Check collection lag
curl -s http://grafana.internal/api/datasources | jq .
```

**Quick Fix**:
```bash
# Restart collectors
curl -X POST http://localhost:5002/automation/remediate \
  -H 'Content-Type: application/json' \
  -d '{"alert_name": "metrics_collection_failure"}'
```

**Prevention**:
- Monitor collector process
- Verify database connectivity
- Check resource usage

---

## 5. System Access

### Database Access

```bash
# Connection details
HOST: production.db.internal
PORT: 5432
USER: pganalytics
PASSWORD: (in vault)
DATABASE: pganalytics

# Connect
psql -h production.db.internal -U pganalytics -d pganalytics

# Or with .pgpass file:
echo "production.db.internal:5432:pganalytics:pganalytics:PASSWORD" >> ~/.pgpass
chmod 600 ~/.pgpass
psql -h production.db.internal
```

### Grafana

```
URL: http://grafana.internal
User: on-call (shared account)
Password: (in vault)

Key Dashboards:
- Overview Dashboard (current state)
- Lock Monitoring (lock issues)
- Table Bloat (bloat issues)
- Connections (connection pool)
```

### Automation Engines

```
Automation Engine API
  - URL: http://localhost:5002
  - Port: 5002
  - Health: GET /automation/health
  - Remediate: POST /automation/remediate
  - History: GET /automation/history

Correlation Engine API
  - URL: http://localhost:5003
  - Port: 5003
  - Health: GET /correlation/health
  - List Incidents: GET /correlation/incidents
  - Get Incident: GET /correlation/incident/{id}
```

### Slack Channels

```
#critical-alerts     - CRITICAL alerts only (WATCH THIS)
#database-alerts     - WARNING level alerts
#database-info       - INFO level alerts
#database-oncall     - Team chat
#incidents           - Incident summaries
```

---

## 6. Safety Rules

### DO ✅

```
✅ Read the runbook completely BEFORE taking action
✅ Verify the blocking transaction details
✅ Get approval before killing critical processes
✅ Monitor the effect of your changes
✅ Document what you did
✅ Communicate status to team
✅ Ask for help if unsure
✅ Escalate if unresolved after 5 min (critical)
```

### DON'T ❌

```
❌ Kill random processes without verification
❌ Run VACUUM FULL during production hours
❌ Ignore auto-remediation results
❌ Change PostgreSQL config without testing
❌ Delete alert history or logs
❌ Disable alerts without approval
❌ Work alone on complex issues
❌ Assume what the problem is (always verify!)
```

---

## 7. Decision Tree

```
Alert fires
  ↓
Is it CRITICAL?
  ├─ YES: Acknowledge within 2 min
  └─ NO: Acknowledge within 10 min
  ↓
Check if already auto-remediated
  ├─ YES: Did it work?
  │   ├─ YES: Monitor for recurrence
  │   └─ NO: Continue to diagnosis
  └─ NO: Continue to diagnosis
  ↓
Which alert type?
  ├─ Lock/Blocking: RUNBOOK_LOCK_CONTENTION.md
  ├─ Bloat: RUNBOOK_TABLE_BLOAT.md
  ├─ Connections: RUNBOOK_CONNECTIONS.md
  ├─ Collection: RUNBOOK_COLLECTOR_FAILURE.md
  └─ Other: Email team, escalate
  ↓
Follow runbook steps
  1. Diagnose (run queries)
  2. Decide (safe to remediate?)
  3. Execute (take action)
  4. Verify (confirm resolution)
  5. Root cause (prevent recurrence)
  ↓
Alert cleared?
  ├─ YES: Great! Document & close
  └─ NO: Escalate (stuck > 5-15 min)
```

---

## 8. Useful One-Liners

### Check System Health

```bash
# All at once
curl -s http://localhost:5002/automation/health && echo "---" && curl -s http://localhost:5003/correlation/health

# Test database
pg_isready -h production.db.internal -p 5432

# Check PostgreSQL status
psql -h production.db.internal -c "SELECT version();"
```

### View Active Issues

```bash
# Active incidents
curl -s http://localhost:5003/correlation/incidents | jq '.incidents[] | {id, group_name, state}'

# Remediation history
curl -s http://localhost:5002/automation/history?limit=10 | jq '.results[] | {action, status, timestamp}'

# Lock status
psql -h production.db.internal -U pganalytics -d pganalytics -c \
"SELECT COUNT(*) as locked_count FROM pg_locks WHERE NOT granted;"
```

### Kill/Fix Things

```bash
# Terminate a PID
psql -h production.db.internal -U pganalytics -d pganalytics -c "SELECT pg_terminate_backend(12345);"

# Vacuum a table
psql -h production.db.internal -U pganalytics -d pganalytics -c "VACUUM ANALYZE public.table_name;"

# Close idle connections
psql -h production.db.internal -U pganalytics -d pganalytics -c \
"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'idle' AND query_start < NOW() - INTERVAL '5 minutes';"
```

---

## 9. Post-Incident Checklist

- [ ] Alert resolved and acknowledged
- [ ] System stable and monitoring normally
- [ ] Root cause identified
- [ ] Incident ticket updated with details
- [ ] Prevention steps documented
- [ ] Team notified of resolution
- [ ] Handoff complete to next on-call

---

## Emergency Commands

```bash
# If database is unresponsive
ping production.db.internal
telnet production.db.internal 5432

# If can't connect
# 1. Check network: ping host
# 2. Check port: telnet host 5432
# 3. Check auth: psql with verbose
#    psql -h host -U user -d db -v ON_ERROR_STOP=on

# If locks are unresponsive
# Last resort: contact infrastructure team
# They can restart PostgreSQL if needed

# If automation engines down
curl -s http://localhost:5002/automation/health
curl -s http://localhost:5003/correlation/health
# If both down, restart:
sudo systemctl restart pganalytics-automation
sudo systemctl restart pganalytics-correlation
```

---

## Quick Links

- **Grafana**: http://grafana.internal
- **Runbooks Folder**: /docs/ (in this repo)
- **Team Wiki**: [Wiki Link]
- **Slack**: [Workspace Link]
- **On-Call Schedule**: [Calendar Link]
- **Escalation List**: [Contact List]

---

**Remember**: It's better to ask for help than to make things worse!

Generated: March 3, 2026
