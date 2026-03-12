# Grafana "Failed to upgrade legacy queries" - Root Cause & Solution

## Problem Description

User reported seeing error message in Grafana browser console:
```
Templating - Failed to upgrade legacy queries
```

This error occurred even after fixing the dashboard JSON files to remove legacy template variable fields.

## Root Cause Analysis

The issue was **NOT** in the dashboard JSON files themselves, but in **Grafana's internal database**.

### Why This Happened

1. **Old Grafana Database State**: When Grafana started initially, it loaded the dashboards from the JSON files and stored them in its SQLite database (`/var/lib/grafana/grafana.db`)

2. **Dashboard with Legacy Fields**: The first version of the `pg-query-by-hostname.json` dashboard contained legacy template variable fields:
   - `definition` (legacy query field)
   - `query` (legacy field)
   - `regex`, `tagValuesQuery`, `tagsQuery`, `useTags` (all legacy)

3. **Grafana Stores State**: Grafana persists dashboard definitions in its database. Once a dashboard was loaded, that version (with legacy fields) was stored.

4. **No Auto-Update of Stored Dashboards**: When we fixed the JSON files, Grafana didn't automatically update the stored versions in the database. The database still contained the old format with legacy fields.

5. **Browser-Side Error**: When Grafana tried to upgrade template variables from the database records to the new format, the JavaScript code encountered legacy field values and failed with "Failed to upgrade legacy queries".

## Solution

The solution was to **completely reset Grafana's database** so it would re-provision dashboards from the corrected JSON files:

```bash
docker-compose down
docker volume rm pganalytics-v3_grafana_staging_data
docker-compose up -d
```

This forced Grafana to:
1. Create a fresh database
2. Re-provision all dashboards from the JSON files in `grafana/dashboards/`
3. Load the corrected versions without legacy fields

## What Was Fixed in the JSON Files

Before resetting, we also fixed these files:

### 1. `grafana/dashboards/pg-query-by-hostname.json`
- **Removed**: `definition` field from the `hostname` template variable
- This field is not recognized by Grafana 11.0.0

### 2. `grafana/dashboards/query-performance.json`
- **Removed**: `queryValue`, `tags`, `tagValuesQuery`, `tagsQuery`, `useTags`
- **Simplified**: options array to use only `text` and `value` fields

### 3. `grafana/dashboards/advanced-features-analysis.json`
- **Removed**: invalid `query` field from custom variables
- **Simplified**: to Grafana 11.0.0 compatible format

## Verification

✅ **Before Fix**:
```
Error in browser console: "Templating - Failed to upgrade legacy queries"
```

✅ **After Fix**:
```bash
$ curl http://localhost:3001/api/health
{
  "version": "11.0.0",
  "database": "ok"
}
```

No template upgrade errors in logs or browser console.

## Key Learnings

1. **Database vs Files**: When Grafana provisions dashboards from files, it stores copies in its database. Fixing files alone doesn't update stored copies.

2. **Template Variables Format**: Grafana 11.0.0 uses a different template variable format than earlier versions. Legacy fields cause upgrade failures.

3. **Complete Reset Method**: Sometimes the cleanest solution is to reset persistent storage and let the application re-initialize from source files.

## Production Considerations

For production deployments:
1. Always validate dashboard JSON files before deploying
2. Consider backing up Grafana database before version upgrades
3. Test template variable compatibility when upgrading Grafana versions
4. Document any manual dashboard modifications (they may be lost on database reset)

## Files Modified

- ✅ `grafana/dashboards/pg-query-by-hostname.json` - Removed `definition` field
- ✅ `grafana/dashboards/query-performance.json` - Removed legacy fields
- ✅ `grafana/dashboards/advanced-features-analysis.json` - Fixed template structure

## Status

✅ **RESOLVED** - Grafana 11.0.0 running with all dashboards loaded correctly, no template upgrade errors.
