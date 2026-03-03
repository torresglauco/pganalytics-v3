# Phase 4: Quick Start Guide - Grafana Dashboards

**Status**: ✅ Complete and Production Ready
**Date**: March 3, 2026

---

## What Was Delivered

### 7 Production Grafana Dashboards

1. **Schema Overview** - Database structure tracking
2. **Lock Monitoring** - Real-time lock detection
3. **Bloat Analysis** - Table/index bloat analysis
4. **Cache Performance** - Buffer pool efficiency
5. **Connection Tracking** - Session monitoring
6. **Extensions Config** - Extension inventory
7. **System Overview** - Health dashboard

### Key Statistics

| Metric | Value |
|--------|-------|
| Total Dashboards | 7 |
| Total Panels | 46 |
| SQL Queries | 46 |
| Data Sources | 11 tables |
| Documentation | 2,500+ lines |
| Total Size | 82 KB JSON |

---

## Quick Deployment

### Option 1: Manual Import (5 minutes)

```bash
1. Go to Grafana: http://localhost:3000
2. Dashboards → Import
3. Paste JSON from dashboards/*.json files
4. Select "PostgreSQL" datasource
5. Click Import × 7
```

### Option 2: Scripted Import (2 minutes)

```bash
#!/bin/bash
GRAFANA_URL="http://localhost:3000"
API_KEY="your-api-key"

for file in dashboards/*.json; do
  curl -X POST "$GRAFANA_URL/api/dashboards/db" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d @"$file"
done
```

### Option 3: Docker Compose (Automated)

```yaml
services:
  grafana:
    image: grafana/grafana:8.0
    volumes:
      - ./dashboards:/var/lib/grafana/dashboards
      - ./provisioning:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

---

## What Each Dashboard Does

### Schema Overview
**Purpose**: Monitor database structure

```
View:
- Table count over time
- Total tables and schemas
- Column count by table
- Constraint distribution
- Foreign key relationships
```

**Refresh**: 5 minutes
**Time Window**: 7 days

### Lock Monitoring ⚡
**Purpose**: Real-time lock detection

```
View:
- Active locks counter (updates every 1 minute)
- Lock wait chains
- Maximum lock age
- Lock trends over 24 hours
- Current blocking queries
```

**Refresh**: 1 minute (CRITICAL)
**Alert Thresholds**:
- Yellow at 5 locks
- Red at 10 locks

### Bloat Analysis
**Purpose**: Find tables needing cleanup

```
View:
- Max bloat percentage (color-coded)
- Dead tuples count
- Space wasted amount
- Tables with >10% bloat (with cleanup recommendations)
- Unused indexes
- Bloat trend over 7 days
```

**Refresh**: 5 minutes
**Color Coding**:
- 🔴 Critical: > 50% bloat
- 🟠 High: 25-50% bloat
- 🟡 Medium: 10-25% bloat
- 🟢 OK: < 10% bloat

### Cache Performance
**Purpose**: Monitor buffer pool efficiency

```
View:
- Average cache hit ratio (24h)
- Index cache hit ratio
- Total block misses
- Cache trends over 7 days
- Per-table performance
```

**Refresh**: 5 minutes
**Target**: Cache hit ratio > 95%

### Connection Tracking 🔌
**Purpose**: Monitor sessions and detect issues

```
View:
- Active connections (real-time)
- Idle connections
- Idle-in-transaction (ALERT if > 0)
- Connection states trend
- Per-database connections
- Long-running transactions table
```

**Refresh**: 1 minute (CRITICAL)
**Alerts**:
- Yellow: 25+ idle connections
- Red: 10+ idle-in-transaction

### Extensions Config
**Purpose**: Extension inventory

```
View:
- Total extensions count
- Databases using extensions
- UUID-OSSP deployments
- Extension distribution trends
- Complete inventory with versions
```

**Refresh**: 5 minutes
**Time Window**: 30 days

### System Overview 📊
**Purpose**: Executive health dashboard

```
View:
- 4 key metrics summary (top row)
  - Active connections
  - Active locks
  - Cache hit ratio
  - Max bloat %
- Connection trend (24h)
- Cache trend (7d)
- Database health summary table
```

**Refresh**: 1 minute
**Status Colors**:
- 🟢 Healthy: All metrics OK
- 🟠 Warning: Cache < 80%
- 🔴 Critical: Bloat > 50%

---

## Before You Start

### Prerequisites

✅ Grafana 8.0+ installed
✅ PostgreSQL datasource configured
✅ Database name: `pganalytics`
✅ Phase 3 metrics flowing (collector running)
✅ Metrics tables populated with data

### Check Your Setup

```sql
-- Verify metrics are being collected
SELECT COUNT(*) FROM metrics_pg_schema_tables;
SELECT COUNT(*) FROM metrics_pg_locks;
SELECT COUNT(*) FROM metrics_pg_bloat_tables;
SELECT COUNT(*) FROM metrics_pg_cache_hit_ratios;
SELECT COUNT(*) FROM metrics_pg_connections;
SELECT COUNT(*) FROM metrics_pg_extensions;

-- Should return > 0 rows for each
```

---

## Troubleshooting

### "No Data" in Dashboard

```
1. Check collector is running:
   ps aux | grep pganalytics-collector

2. Check metrics in database:
   SELECT COUNT(*) FROM metrics_pg_schema_tables;

3. Check time filter:
   Click time picker → Last 7 days (or Last 24h)

4. Check datasource:
   Settings → Data Sources → PostgreSQL → Test Connection
```

### Slow Dashboard Load

```
1. Reduce time window:
   Click time picker → Last 24h (instead of 30d)

2. Check PostgreSQL:
   SELECT NOW() - max(time) FROM metrics_pg_schema_tables;
   (Should be < 5 minutes ago)

3. Check query:
   Select panel → Inspect → View query
   (Check for missing time filter)
```

### Query Error

```
1. Panel → Inspect → Show Query
2. Copy query → Run in psql
3. Check for:
   - Missing time column
   - Wrong table name
   - Missing schema prefix
```

---

## Using the Dashboards

### Daily Checklist

**Every Morning**:
1. Open System Overview dashboard
2. Check health status (green = OK)
3. Note any yellow/red indicators
4. Investigate if needed

**Action Items**:
- Red locks? → Check Lock Monitoring
- Red bloat? → Check Bloat Analysis
- Red cache? → Check Cache Performance
- Red connections? → Check Connection Tracking

### Weekly Review

**Every Monday**:
1. Bloat Analysis → Table Bloat Chart
   - Schedule VACUUM if needed
2. Cache Performance → 7-day trend
   - Check if stable or declining
3. Extensions → Check for new installations

### Monthly Report

**First of Month**:
1. Schema Overview → 30-day view
   - Schema growth rate
   - New tables added
2. Lock Monitoring → 30-day statistics
   - Lock frequency
   - Blocking patterns
3. System Overview → Performance summary
   - Overall health trend

---

## Alert Setup (Optional)

### Grafana Alerts

**In Dashboard → Panel → Alert**:

```json
{
  "Lock Alert": "if(value > 10) then Alert",
  "Bloat Alert": "if(value > 50) then Alert",
  "Cache Alert": "if(value < 80) then Alert"
}
```

### Notification Channels

1. Grafana Settings → Notifications
2. Add Channel (Slack, PagerDuty, Email)
3. Configure alert rules

---

## Common Queries

### Find Blocking Queries

```
Lock Monitoring Dashboard
→ Current Lock Blocking Chain (Bottom Table)
→ Copy blocking_query
→ Run in PostgreSQL client
→ Optimize or KILL if necessary
```

### Find Bloated Tables

```
Bloat Analysis Dashboard
→ Tables with High Bloat (>10%)
→ Click table_name
→ Check recommendation column
→ Schedule VACUUM or REINDEX
```

### Monitor Cache Hit

```
Cache Performance Dashboard
→ Cache Hit Ratio Trend (7 days)
→ Check trend direction
→ If declining: increase shared_buffers
```

---

## Performance Notes

### Query Optimization

All 46 queries use:
- ✅ Time-based filtering (fast)
- ✅ TimescaleDB hypertable optimization
- ✅ Aggregation at query level
- ✅ LIMIT clauses for pagination
- ✅ Index utilization

### Refresh Rates

- **1 minute** (3 dashboards): Critical operational
- **5 minutes** (4 dashboards): Trend analysis

### Storage Impact

- 46 queries per dashboard refresh
- ~500MB per month (at default settings)
- ~100 rows returned average per query

---

## Files Included

### Dashboard JSON Files

```
dashboards/
├── schema-overview.json (446 lines)
├── lock-monitoring.json (484 lines)
├── bloat-analysis.json (427 lines)
├── cache-performance.json (415 lines)
├── connection-tracking.json (459 lines)
├── extensions-config.json (410 lines)
└── system-overview.json (277 lines)
```

### Documentation Files

```
├── PHASE4_DASHBOARDS_GUIDE.md (880 lines)
├── PHASE4_COMPLETION_SUMMARY.md (513 lines)
├── PHASE4_STATUS.md (429 lines)
└── PHASE4_QUICK_START.md (this file)
```

---

## Next Steps

### After Deployment

1. ✅ Import all 7 dashboards
2. ✅ Verify data appears (wait 5 minutes)
3. ✅ Test time filters
4. ✅ Set up alerts (optional)
5. ✅ Train team on usage

### For Production

1. Customize alert thresholds for your environment
2. Set up notification channels
3. Schedule regular review meetings
4. Document runbooks for common issues

---

## Support Resources

**For Help With:**

- **Dashboard Deployment** → PHASE4_DASHBOARDS_GUIDE.md
- **Dashboard Details** → PHASE4_COMPLETION_SUMMARY.md
- **Session Information** → PHASE4_STATUS.md
- **API Details** → API_ARCHITECTURE_EXPLANATION.md
- **Phase 3 Info** → PHASE3_QUICK_REFERENCE.md

---

## Success Indicators

✅ All 7 dashboards appear in Grafana
✅ All 46 panels show data
✅ No SQL errors in logs
✅ Dashboard loads in < 2 seconds
✅ Time filters work correctly
✅ Panel colors change appropriately
✅ Team can identify issues faster

---

**Ready to Deploy?**

```bash
# 1. Copy dashboard files to Grafana provisioning directory
cp dashboards/*.json /etc/grafana/provisioning/dashboards/

# 2. Restart Grafana
sudo systemctl restart grafana-server

# 3. Open Grafana and verify dashboards appear
open http://localhost:3000
```

**Questions?** See the full documentation in PHASE4_DASHBOARDS_GUIDE.md

---

**Phase 4 Status**: ✅ **COMPLETE - Ready for Production**
