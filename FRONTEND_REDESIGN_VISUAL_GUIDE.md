# Frontend Redesign - Visual Guide & Mockups
## pgAnalytics v3 UI/UX Transformation

**Date**: March 3, 2026

---

## Application Layout Overview

```
┌────────────────────────────────────────────────────────────────────────┐
│ pgAnalytics Logo │ Dashboard  │ 🔍 Search...  │ 🔔 (3) │ 👤 Menu ▼    │
├──────────────────┬────────────────────────────────────────────────────┤
│                  │                                                    │
│  📊 Overview     │  Welcome to pgAnalytics v3                         │
│  🚨 Alerts       │  ┌──────────────────────────────────────┐         │
│  ⚡ Queries      │  │ System Health                        │         │
│  🔒 Locks        │  │ ████████████░░░░░░ 85/100  ↗ +5%   │         │
│  🧹 Bloat        │  └──────────────────────────────────────┘         │
│  📡 Connections  │                                                    │
│  💾 Cache        │  ┌─────────────┬─────────────┬─────────────┐     │
│  📐 Schema       │  │ Critical    │ Warnings    │ Info        │     │
│  🔄 Replication  │  │ 2           │ 7           │ 12          │     │
│  💪 Health       │  └─────────────┴─────────────┴─────────────┘     │
│  ⚙️ Extensions   │                                                    │
│  🖥️ Collectors   │  ┌──────────────────────────────────────┐         │
│  ⚙️ Settings     │  │ Top Issues                           │         │
│                  │  │ 1. Lock contention on prod-db-1      │         │
│                  │  │ 2. Table bloat > 50% on events       │         │
│                  │  │ 3. Cache hit < 70% on users table    │         │
│                  │  └──────────────────────────────────────┘         │
│                  │                                                    │
└──────────────────┴────────────────────────────────────────────────────┘
```

---

## Page Designs

### 1. Overview Dashboard

```
┌─────────────────────────────────────────────────────────────────────┐
│ pgAnalytics › Overview Dashboard                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Quick Status Cards (4 columns)                                     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────┐  │
│  │ 💪 Health    │ │ 🚨 Critical  │ │ ⚠️ Warnings  │ │ ℹ️ Info  │  │
│  │ 85/100 ↗    │ │ 2 alerts     │ │ 7 alerts     │ │ 12       │  │
│  └──────────────┘ └──────────────┘ └──────────────┘ └──────────┘  │
│                                                                     │
│  Health Score Gauge              Connection Usage                   │
│  ┌──────────────────────┐        ┌──────────────────────┐          │
│  │   Health Score       │        │ Connections: 142/200 │          │
│  │                      │        │ ████████████░░░░░░   │          │
│  │     ████░░░░░░       │        │ 71%                  │          │
│  │       85/100         │        │ Status: Healthy      │          │
│  │                      │        └──────────────────────┘          │
│  │ Trend: ↗ improving  │                                           │
│  └──────────────────────┘                                          │
│                                                                     │
│  Top Issues                      Recent Incidents                  │
│  ┌──────────────────────┐        ┌──────────────────────┐          │
│  │ 1. [CRITICAL]        │        │ Lock Contention      │          │
│  │    Lock contention   │        │ Status: Active       │          │
│  │    prod-db-1         │        │ Severity: Critical   │          │
│  │    → View Details    │        │ Alerts: 3            │          │
│  │                      │        │ → View Incident      │          │
│  │ 2. [WARNING]         │        │                      │          │
│  │    Table bloat       │        │ Table Bloat          │          │
│  │    events table      │        │ Status: Monitoring   │          │
│  │    → View Details    │        │ Severity: Warning    │          │
│  │                      │        │ Alerts: 2            │          │
│  │ 3. [WARNING]         │        │ → View Incident      │          │
│  │    Cache miss        │        │                      │          │
│  │    users table       │        └──────────────────────┘          │
│  │    → View Details    │                                          │
│  └──────────────────────┘                                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 2. Alerts & Incidents Page

```
┌─────────────────────────────────────────────────────────────────────┐
│ pgAnalytics › Alerts & Incidents                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌───────────┬───────────┬───────────┬──────────────┐               │
│  │ Severity: │ Status:   │ Type:     │ Collector:   │               │
│  │ [All ▼]   │ [All ▼]   │ [All ▼]   │ [All ▼]      │ [Search]    │
│  └───────────┴───────────┴───────────┴──────────────┘               │
│                                                                     │
│  Summary Cards                                                      │
│  ┌─────────────┬─────────────┬────────────┬────────────┐           │
│  │ 🔴 Critical │ 🟡 Warning  │ 🔵 Info    │ ✓ Resolved │           │
│  │ 2           │ 7           │ 12         │ 24         │           │
│  └─────────────┴─────────────┴────────────┴────────────┘           │
│                                                                     │
│  Alert List                                                         │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Severity │ Alert           │ Status │ Fired   │ Actions      │  │
│  ├──────────┼─────────────────┼────────┼─────────┼──────────────┤  │
│  │ 🔴 CRIT  │ Lock Contention │ Active │ 2 min   │ [Ack] [View] │  │
│  │ 🔴 CRIT  │ Collection Fail  │ Active │ 5 min   │ [Ack] [View] │  │
│  │ 🟡 WARN  │ Table Bloat      │ Active │ 10 min  │ [Ack] [View] │  │
│  │ 🟡 WARN  │ Cache Miss       │ Active │ 15 min  │ [Ack] [View] │  │
│  │ 🟡 WARN  │ High Conn Count  │ Active │ 18 min  │ [Ack] [View] │  │
│  │ 🔵 INFO  │ Bloat Detected   │ Muted  │ 30 min  │ [Unmute]     │  │
│  │ ✓ CRIT   │ Lock Resolved    │ Closed │ 45 min  │ [Details]    │  │
│  └──────────┴─────────────────┴────────┴─────────┴──────────────┘  │
│                                                                     │
│  Incident Groups                                                    │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 🔴 Lock Contention Group        │ 85% Confidence            │  │
│  │   Related Alerts: 3              │ Status: Active            │  │
│  │   • Lock Contention (Critical)   │ Suggested Actions:        │  │
│  │   • High Lock Wait Time (Warning)│ • Kill blocking TXN       │  │
│  │   • Connection Timeout (Warning) │ • Review blocking query   │  │
│  │                                  │ [View Runbook]            │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3. Query Performance Page

```
┌─────────────────────────────────────────────────────────────────────┐
│ pgAnalytics › Query Performance                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Top Metrics                                                        │
│  ┌──────────────┬──────────────┬──────────────┬──────────────┐     │
│  │ Total Queries│ Avg Time     │ Slow Queries │ Very Slow    │     │
│  │ 1,234,567    │ 125ms        │ 42           │ 8            │     │
│  └──────────────┴──────────────┴──────────────┴──────────────┘     │
│                                                                     │
│  Query Duration Timeline                                            │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ ms│                                                       │      │
│  │  5│    ╭╮  ╭╮                                             │      │
│  │  4│    ││  ││                                             │      │
│  │  3│╭╮  ││  ││      ╭╮                                     │      │
│  │  2│││╭─╮╰╮ ││╭─────╯│                                     │      │
│  │  1│╰╯│ │ │ ╰╯│                                            │      │
│  │   └──┴─┴─┴───┴─────────────────────────────────────────┘      │
│  │      12:00    12:30    13:00    13:30    14:00               │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  Slow Queries (Sorted by Avg Time)                                 │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Query Text          │ Calls │ Avg  │ Max  │ Rows │ Plan     │  │
│  ├─────────────────────┼───────┼──────┼──────┼──────┼──────────┤  │
│  │ SELECT * FROM users │ 1,230 │ 850m │ 5.2s │ 500K │ [Seq]    │  │
│  │ where status='ac..  │       │      │      │      │ [Analyze]│  │
│  │                     │       │      │      │      │          │  │
│  │ SELECT * FROM ord.. │  542  │ 320m │ 1.8s │ 12K  │ [Index]  │  │
│  │ where date > now()  │       │      │      │      │          │  │
│  │                     │       │      │      │      │          │  │
│  │ SELECT COUNT(*) ... │  89   │ 210m │ 850m │ 1    │ [Seq]    │  │
│  │ FROM transactions..  │       │      │      │      │ [Analyze]│  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 4. Lock Contention Page

```
┌─────────────────────────────────────────────────────────────────────┐
│ pgAnalytics › Lock Contention                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Current Status                                                     │
│  ┌──────────────────┬──────────────────┬──────────────────┐        │
│  │ Active Locks     │ Blocked TXNs     │ Max Wait Time    │        │
│  │ 12 locks         │ 3 transactions   │ 2m 35s           │        │
│  │ Status: ⚠️ WARN │ Status: 🔴 CRIT │ Status: 🔴 CRIT  │        │
│  └──────────────────┴──────────────────┴──────────────────┘        │
│                                                                     │
│  Lock Wait Chain Visualization                                      │
│  ┌─────────────────────────────────────────────────────────┐       │
│  │                                                         │       │
│  │  TXN 4521                 TXN 4589                     │       │
│  │  (WAITING)      ←──────   (WAITING)                    │       │
│  │  UPDATE users   (LOCK)    SELECT orders                │       │
│  │                                    ↓                   │       │
│  │                              TXN 4523                  │       │
│  │                              (BLOCKING)                │       │
│  │                              UPDATE orders             │       │
│  │                              since 2m 35s              │       │
│  │                                                         │       │
│  │  RECOMMENDATION: Kill TXN 4523 (slow query detected)  │       │
│  │  [Kill Transaction] [View Query] [View Runbook]       │       │
│  │                                                         │       │
│  └─────────────────────────────────────────────────────────┘       │
│                                                                     │
│  Active Locks Detail                                                │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Blocking │ Lock Type    │ Duration │ Query (first 50 chars) │  │
│  │ PID      │              │          │                        │  │
│  ├──────────┼──────────────┼──────────┼────────────────────────┤  │
│  │ 4523     │ AccessExcl.  │ 2m 35s   │ UPDATE orders SET...   │  │
│  │ 4521     │ RowExcl.Wait │ 1m 20s   │ UPDATE users SET...    │  │
│  │ 4589     │ RowExcl.Wait │ 45s      │ SELECT * FROM orders.. │  │
│  │ 4601     │ ShareLock    │ 30s      │ CREATE INDEX on users  │  │
│  └──────────┴──────────────┴──────────┴────────────────────────┘  │
│                                                                     │
│  Auto-Remediation History                                           │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Action              │ Status   │ Time      │ Result           │  │
│  ├─────────────────────┼──────────┼───────────┼──────────────────┤  │
│  │ Kill TXN 4523       │ ✓ Success│ 10 min ago│ Lock released    │  │
│  │ Auto-remediation    │          │           │ Alert cleared    │  │
│  │                     │          │           │                  │  │
│  │ Kill TXN 4521       │ ✓ Success│ 5 min ago │ Query completed  │  │
│  │ Manual action       │          │           │ No side effects  │  │
│  └─────────────────────┴──────────┴───────────┴──────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 5. Table Bloat Analysis Page

```
┌─────────────────────────────────────────────────────────────────────┐
│ pgAnalytics › Table Bloat Analysis                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Overall Bloat Metrics                                              │
│  ┌──────────────────┬──────────────────┬──────────────────┐        │
│  │ Avg Bloat Ratio  │ Bloated Tables   │ Reclaimable      │        │
│  │ ████████░░ 28%   │ 12 / 47 tables   │ 2.3 GB           │        │
│  │ Status: ⚠️ WARN  │ Status: ⚠️ WARN  │ (Can be freed)   │        │
│  └──────────────────┴──────────────────┴──────────────────┘        │
│                                                                     │
│  Bloat Trend (30 days)                                              │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ %│                                                       │      │
│  │35│      ╭╮                                               │      │
│  │30│╭────╮││╭─────╮                                       │      │
│  │25││    ││╰╯     │                                       │      │
│  │20│╰────╯       ╰───────────────────────────────────    │      │
│  │15│                                                     │      │
│  │ └─────────────────────────────────────────────────────│      │
│  │  1w ago    Now                                          │      │
│  │  Trend: ↗ Increasing (needs maintenance)              │      │
│  └──────────────────────────────────────────────────────────┘      │
│                                                                     │
│  Bloated Tables (Sorted by Bloat Ratio)                             │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Table        │ Bloat % │ Dead Tup│ Reclaim  │ Recommend.   │  │
│  ├──────────────┼─────────┼─────────┼──────────┼──────────────┤  │
│  │ events       │ ███████░│ 52%     │ 125.3M   │ 850K tuples  │  │
│  │ (CRITICAL)   │         │         │          │              │  │
│  │              │         │         │          │ [VACUUM FULL]│  │
│  │              │         │         │          │ (sched. 2am) │  │
│  │                                                             │  │
│  │ audit_log    │ █████░░░│ 34%     │ 89.2M    │ 420K tuples  │  │
│  │ (WARNING)    │         │         │          │              │  │
│  │              │         │         │          │ [VACUUM]     │  │
│  │              │         │         │          │ (regular)    │  │
│  │                                                             │  │
│  │ sessions     │ ████░░░░│ 28%     │ 56.1M    │ 280K tuples  │  │
│  │ (MONITORING) │         │         │          │              │  │
│  │              │         │         │          │ [MONITOR]    │  │
│  │              │         │         │          │ (no action)  │  │
│  └──────────────┴─────────┴─────────┴──────────┴──────────────┘  │
│                                                                     │
│  Maintenance Recommendations                                        │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1. URGENT: VACUUM FULL events table                         │  │
│  │    Estimated duration: 45 minutes (schedule for 2:00 AM)    │  │
│  │    [Schedule Maintenance] [View Runbook]                    │  │
│  │                                                              │  │
│  │ 2. SOON: VACUUM audit_log table                             │  │
│  │    Can run during business hours (< 5 minutes)              │  │
│  │    [Execute Now] [Schedule] [View Runbook]                 │  │
│  │                                                              │  │
│  │ 3. MONITOR: sessions table                                  │  │
│  │    Current bloat acceptable, check monthly                 │  │
│  │    [Monitor] [View Autovacuum Config]                       │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 6. Database Health Page

```
┌─────────────────────────────────────────────────────────────────────┐
│ pgAnalytics › Database Health                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Overall Health Score                                               │
│  ┌─────────────────────────────────────────────────────────┐       │
│  │                                                         │       │
│  │           HEALTH SCORE: 85/100                         │       │
│  │                                                         │       │
│  │           ████████████░░░░░░░░░░░░░░░░                │       │
│  │                                                         │       │
│  │           Status: GOOD ↗                               │       │
│  │           Trend: Improving (+5% in 7 days)            │       │
│  │           Last Updated: 2 minutes ago                 │       │
│  │                                                         │       │
│  └─────────────────────────────────────────────────────────┘       │
│                                                                     │
│  Health Component Breakdown                                         │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐        │
│  │ Lock Health │ Bloat Health│ Query Perf. │ Cache Eff.  │        │
│  │ ████████░░░│ ███░░░░░░░░░│ █████████░░ │ ██████░░░░░│        │
│  │ 80/100      │ 45/100      │ 90/100      │ 60/100      │        │
│  │ ✓ Healthy   │ ⚠️ Warning  │ ✓ Healthy   │ ⚠️ Warning  │        │
│  └─────────────┴─────────────┴─────────────┴─────────────┘        │
│                                                                     │
│  ┌─────────────┬─────────────┬──────────────┐                     │
│  │ Connections │ Replication │ Extensions   │                     │
│  │ ██████████░ │ ██████████░ │ ███████████░ │                     │
│  │ 85/100      │ 92/100      │ 95/100       │                     │
│  │ ✓ Healthy   │ ✓ Healthy   │ ✓ Healthy    │                     │
│  └─────────────┴─────────────┴──────────────┘                     │
│                                                                     │
│  Health Score History (30 days)                                     │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │ Score│                                                  │      │
│  │ 100 │╭────────────────────────────────────────────╮   │      │
│  │  85 ││  ╭─╮                                        │   │      │
│  │  70 ││  │ ╰─╮                                      │   │      │
│  │  55 ││  │   ╰──────╮                               │   │      │
│  │  40 ││  │          ╰──────────────────────────    │   │      │
│  │    │╰──────────────────────────────────────────────╯   │      │
│  │    └────────────────────────────────────────────────────┘      │
│  │     1w ago            Now                                      │
│  └──────────────────────────────────────────────────────────┘      │
│                                                                     │
│  Top Issues by Component                                            │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ [⚠️ BLOAT HEALTH - 45/100]                                    │  │
│  │ • events table: 52% bloat (125.3 MB reclaimable)             │  │
│  │ • audit_log: 34% bloat (89.2 MB reclaimable)                 │  │
│  │ Action: Run VACUUM FULL on events table                      │  │
│  │ [View Bloat Details] [Schedule Maintenance]                  │  │
│  │                                                               │  │
│  │ [⚠️ CACHE EFFICIENCY - 60/100]                                │  │
│  │ • users table: 45% cache hit ratio (should be >80%)          │  │
│  │ • orders table: 52% cache hit ratio (high disk I/O)          │  │
│  │ Action: Add indexes on frequently filtered columns           │  │
│  │ [View Cache Analysis] [See Slow Queries]                     │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Component Visual Examples

### Metric Card Component

```
┌─────────────────────────────┐
│ 🔴                          │
│        15 Active Locks      │
│                             │
│  Status: ⚠️ WARN  ↗ +2     │
└─────────────────────────────┘

Color variations:
Healthy:   Green background, ✓ icon
Warning:   Yellow background, ⚠️ icon
Critical:  Red background, 🔴 icon
```

### Status Badge Component

```
┌──────────────────┐
│ 🔴 CRITICAL      │
└──────────────────┘

┌──────────────────┐
│ 🟡 WARNING       │
└──────────────────┘

┌──────────────────┐
│ 🔵 INFO          │
└──────────────────┘

┌──────────────────┐
│ ✓ RESOLVED       │
└──────────────────┘
```

### Chart Variations

```
Line Chart (Trends)          Gauge Chart (Health/Usage)
┌─────────────────────┐      ┌─────────────────┐
│  ms│                │      │       ↑ 100%    │
│ 500│   ╭─╮           │      │    ╭───────╮   │
│ 400│  ╱ ╰─╮          │      │   ╱         ╲  │
│ 300│ ╱    ╰─╮        │      │  │  85/100   │ │
│ 200│────────╰───     │      │   ╲         ╱  │
│    └─────────────────┘      │    ╰───────╯   │
│                             │       ↓ 0%     │
│                             └─────────────────┘

Heatmap (Time Series)        Bar Chart (Comparisons)
┌──────────────────────┐      ┌───────────────┐
│                ████  │      │ █████████   90│
│            ██████░   │      │ ███████░░   70│
│        ██████░░░░░   │      │ █████░░░░░ 50│
│    ████░░░░░░░░░░    │      │ ██░░░░░░░░ 20│
│                      │      └───────────────┘
└──────────────────────┘
```

### Data Table Example

```
┌──────────────┬─────────────┬─────────────┬──────────────┐
│ ↑ Status     │ Alert Type  │ Severity ↓  │ Last Update  │
├──────────────┼─────────────┼─────────────┼──────────────┤
│ 🔴 Active    │ Lock        │ Critical    │ 2 min ago    │
│ 🔴 Active    │ Bloat       │ Critical    │ 5 min ago    │
│ 🟡 Active    │ Cache Miss  │ Warning     │ 10 min ago   │
│ ✓ Resolved   │ Conn Pool   │ Warning     │ 45 min ago   │
└──────────────┴─────────────┴─────────────┴──────────────┘

Features:
• Click column header to sort
• Type to search/filter
• Select rows with checkboxes
• Responsive on mobile
```

---

## Color Palette

```
Primary Blue        Accent Cyan         Success Emerald
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ #1e3a8a     │    │ #06b6d4     │    │ #10b981     │
│ Professional│    │ Modern      │    │ Healthy     │
│ Trustworthy │    │ Data-driven │    │ Good status │
└─────────────┘    └─────────────┘    └─────────────┘

Warning Amber       Danger Rose         Neutral Slate
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ #f59e0b     │    │ #f43f5e     │    │ #64748b     │
│ Caution     │    │ Critical    │    │ Text/Border │
│ Needs Action│    │ Urgent      │    │ Secondary   │
└─────────────┘    └─────────────┘    └─────────────┘
```

---

## Responsive Design

### Desktop (1920x1080)
```
Full sidebar + main content + right panels
All charts visible side-by-side
Tables with all columns
```

### Tablet (1024x768)
```
Collapsible sidebar
Stacked cards 2-column layout
Tables with horizontal scroll
```

### Mobile (375x812)
```
Full-screen sidebar overlay
Cards stacked 1-column
Tables horizontal scroll (swipeable)
Bottom navigation tabs
Simplified charts
```

---

## User Journey Examples

### New User First Visit
```
1. Landing → Overview Dashboard (see system health immediately)
2. See "Top Issues" section
3. Click on issue → Detailed analysis page
4. See recommendations
5. Take action (acknowledge alert, run maintenance, etc)
```

### On-Call Engineer Alert Response
```
1. Alert notification received
2. Click notification → Alerts page
3. View incident group & correlation
4. Click "View Runbook" → See incident response steps
5. Click "Auto-Remediation History" → See what was already done
6. Click "Take Action" → Execute remediation or investigation
7. Alert clears → Mark as resolved
```

### Database Administrator Monitoring
```
1. Login → Overview Dashboard
2. Check Health Score (85/100) - all good
3. Scan recent incidents - none active
4. Check Bloat Analysis page - see 30-day trend
5. Schedule maintenance if needed
6. Check Replication Status - lag < 1MB - good
7. Set suppression rule for monthly maintenance alert
```

---

## Design System Specifications

### Typography
```
Headlines:    Inter Bold, 32px (page titles)
Subheadlines: Inter SemiBold, 20px (section titles)
Body:         Inter Regular, 14px (content)
Monospace:    Fira Code, 12px (queries, JSON)
```

### Spacing
```
xs: 2px    (tight spacing within components)
sm: 4px    (compact spacing)
md: 8px    (default spacing)
lg: 16px   (comfortable spacing)
xl: 24px   (large spacing)
2xl: 32px  (extra large spacing)
```

### Border Radius
```
sm: 4px    (small elements, inputs)
md: 8px    (cards, buttons)
lg: 12px   (large components)
full: 9999px (badges, avatars)
```

### Shadows
```
sm: 0 1px 2px rgba(0,0,0,0.05)
md: 0 4px 6px rgba(0,0,0,0.1)
lg: 0 10px 15px rgba(0,0,0,0.1)
xl: 0 20px 25px rgba(0,0,0,0.15)
```

---

## Interactive Elements

### Button States
```
Default:   Blue background, white text
Hover:     Darker blue, pointer cursor
Active:    Even darker blue, slight scale down
Disabled:  Gray background, faded text, no cursor

Danger:    Red background (for destructive actions)
Success:   Green background (for confirmations)
```

### Form Elements
```
Input:     Light gray border, white background
Focus:     Cyan border, blue shadow
Error:     Red border, red text below
Disabled:  Gray background, faded text

Checkbox:  Blue when checked, uses accent color
Radio:     Blue when selected, uses accent color
Toggle:    Blue when on, gray when off
```

### Loading States
```
Skeleton:  Light gray pulse animation
Spinner:   Blue rotating circle
Progress:  Blue bar from 0-100%
```

---

**Visual Guide Complete!**

All components ready for implementation in React + Tailwind CSS.

Generated: March 3, 2026
