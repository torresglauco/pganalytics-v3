# Phase 4.5.1: Workload Pattern Detection - Testing Guide

**Date**: February 20, 2026
**Status**: Testing & Verification Ready
**Objective**: Comprehensive testing of workload pattern detection implementation

---

## Quick Test Summary

Three levels of testing provided:
1. **Database Level** - Test SQL function directly
2. **API Level** - Test HTTP endpoints with curl
3. **Integration** - Test full workflow with real data

---

## Level 1: Database Testing

### Test Setup: Create Sample Data

```sql
-- Create 30 days of synthetic query metrics with hour 8 as peak
INSERT INTO metrics_pg_stats_query (
    database_name, query_hash, query_text,
    calls, total_exec_time_ms, mean_exec_time_ms,
    rows, collected_at
)
WITH RECURSIVE dates AS (
    SELECT (NOW() - INTERVAL '30 days')::TIMESTAMP as d
    UNION ALL
    SELECT d + INTERVAL '1 hour' FROM dates
    WHERE d < NOW()
),
hourly_data AS (
    SELECT
        d,
        EXTRACT(HOUR FROM d)::INT as hour_of_day,
        EXTRACT(DATE FROM d) as stat_date,
        CASE
            WHEN EXTRACT(HOUR FROM d)::INT = 8 THEN 500    -- Peak hour: 500 queries
            ELSE 50                                           -- Normal hours: 50 queries
        END as query_count,
        CASE
            WHEN EXTRACT(HOUR FROM d)::INT = 8 THEN 250.0   -- Peak: longer execution
            ELSE 50.0                                         -- Normal: shorter execution
        END as exec_time
    FROM dates
)
SELECT
    'testdb' as database_name,
    1::BIGINT as query_hash,
    'SELECT * FROM users WHERE status = ?'::TEXT,
    query_count::BIGINT,
    (query_count * exec_time)::FLOAT,
    exec_time::FLOAT,
    query_count::BIGINT,
    d
FROM hourly_data
ORDER BY d;

-- Verify data was inserted
SELECT COUNT(*) as total_rows FROM metrics_pg_stats_query
WHERE database_name = 'testdb' AND query_hash = 1;
-- Expected: ~720 rows (30 days × 24 hours)
```

### Test 1: Basic Pattern Detection

```sql
-- Run pattern detection
SELECT * FROM detect_workload_patterns('testdb', 30)
ORDER BY confidence DESC;

-- Expected output:
-- pattern_id | pattern_type | confidence
-- -----------+--------------+------------
--          1 | hourly_peak  |       0.92
-- (1 row)
```

**Expected Results**:
- ✅ 1 pattern detected (hourly_peak)
- ✅ Confidence score between 0.8-0.95
- ✅ Pattern type is 'hourly_peak'
- ✅ Pattern metadata contains peak_hour, confidence, z_scores

### Test 2: Pattern Metadata Validation

```sql
-- Check pattern metadata details
SELECT
    pattern_type,
    (pattern_metadata->>'peak_hour')::INT as peak_hour,
    (pattern_metadata->>'confidence')::FLOAT as confidence,
    (pattern_metadata->>'variance')::FLOAT as variance,
    (pattern_metadata->>'z_score_count')::FLOAT as z_score_count,
    (pattern_metadata->>'affected_queries')::INT as affected_queries
FROM workload_patterns
WHERE database_name = 'testdb'
ORDER BY detection_timestamp DESC;

-- Expected output:
-- pattern_type | peak_hour | confidence | variance | z_score_count | affected_queries
-- --------------+-----------+------------+----------+---------------+------------------
-- hourly_peak  |         8 |       0.92 |     0.15 |          2.45 |              500
```

**Expected Results**:
- ✅ peak_hour = 8 (correct hour)
- ✅ confidence ≈ 0.92 (high confidence)
- ✅ variance ≈ 0.15 (consistent)
- ✅ z_score_count > 2.0 (statistically significant)
- ✅ affected_queries ≈ 500 (correct count)

### Test 3: Multiple Peak Hours

```sql
-- Create data with peaks at hours 2, 8, and 15
INSERT INTO metrics_pg_stats_query (database_name, query_hash, calls, mean_exec_time_ms, collected_at)
WITH RECURSIVE dates AS (
    SELECT (NOW() - INTERVAL '30 days')::TIMESTAMP as d
    UNION ALL
    SELECT d + INTERVAL '1 hour' FROM dates WHERE d < NOW()
),
hourly_data AS (
    SELECT
        d,
        EXTRACT(HOUR FROM d)::INT as hour_of_day,
        CASE
            WHEN EXTRACT(HOUR FROM d)::INT IN (2, 8, 15) THEN 500
            ELSE 50
        END as query_count,
        CASE
            WHEN EXTRACT(HOUR FROM d)::INT IN (2, 8, 15) THEN 250.0
            ELSE 50.0
        END as exec_time
    FROM dates
)
SELECT 'multidb'::VARCHAR, 1::BIGINT, query_count::BIGINT, exec_time::FLOAT, d
FROM hourly_data;

-- Run detection
SELECT COUNT(*) as pattern_count FROM detect_workload_patterns('multidb', 30);
-- Expected: 3 patterns (one for each peak hour)
```

**Expected Results**:
- ✅ 3 patterns detected (hours 2, 8, 15)
- ✅ All with confidence > 0.8
- ✅ Each pattern has correct peak_hour in metadata

### Test 4: Insufficient Data

```sql
-- Test with only 3 days of data
INSERT INTO metrics_pg_stats_query (database_name, query_hash, calls, mean_exec_time_ms, collected_at)
WITH RECURSIVE dates AS (
    SELECT (NOW() - INTERVAL '3 days')::TIMESTAMP as d
    UNION ALL
    SELECT d + INTERVAL '1 hour' FROM dates WHERE d < NOW()
),
hourly_data AS (
    SELECT
        d,
        CASE WHEN EXTRACT(HOUR FROM d)::INT = 8 THEN 500 ELSE 50 END,
        CASE WHEN EXTRACT(HOUR FROM d)::INT = 8 THEN 250.0 ELSE 50.0 END
    FROM dates
)
SELECT 'shortdb'::VARCHAR, 1::BIGINT, * FROM hourly_data;

-- Run detection (should handle gracefully)
SELECT COUNT(*) FROM detect_workload_patterns('shortdb', 3);
-- Expected: 0 patterns (minimum 7 days required)
```

**Expected Results**:
- ✅ 0 patterns detected (insufficient data)
- ✅ No errors raised
- ✅ Function completes successfully

### Test 5: Edge Case - No Peaks

```sql
-- Create uniform data with no peaks
INSERT INTO metrics_pg_stats_query (database_name, query_hash, calls, mean_exec_time_ms, collected_at)
WITH RECURSIVE dates AS (
    SELECT (NOW() - INTERVAL '30 days')::TIMESTAMP as d
    UNION ALL
    SELECT d + INTERVAL '1 hour' FROM dates WHERE d < NOW()
)
SELECT 'uniformdb'::VARCHAR, 1::BIGINT, 100::BIGINT, 50.0::FLOAT, d
FROM dates;

-- Run detection
SELECT COUNT(*) FROM detect_workload_patterns('uniformdb', 30);
-- Expected: 0 patterns (no significant peaks)
```

**Expected Results**:
- ✅ 0 patterns detected
- ✅ Correctly identifies uniform load
- ✅ No false positives

---

## Level 2: API Testing

### Setup: Start API Server

```bash
cd /Users/glauco.torres/git/pganalytics-v3/backend
go build -o pganalytics-api cmd/pganalytics-api/main.go
./pganalytics-api

# In another terminal, get auth token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | jq -r '.token')

echo "Token: $TOKEN"
```

### Test 1: Detect Patterns API Call

```bash
# Test 1: Basic pattern detection
curl -X POST http://localhost:8080/api/v1/workload-patterns/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "database_name": "testdb",
    "lookback_days": 30
  }' | jq .

# Expected response:
# {
#   "patterns_detected": 1,
#   "database_name": "testdb",
#   "lookback_days": 30,
#   "timestamp": "2026-02-20T14:30:00Z"
# }
```

**Verification**:
- ✅ HTTP 200 status
- ✅ patterns_detected > 0
- ✅ database_name matches input
- ✅ timestamp is ISO 8601 format

### Test 2: Get Patterns API Call

```bash
# Get detected patterns
curl http://localhost:8080/api/v1/workload-patterns?database_name=testdb \
  -H "Authorization: Bearer $TOKEN" | jq .

# Expected response:
# [
#   {
#     "id": 1,
#     "database_name": "testdb",
#     "pattern_type": "hourly_peak",
#     "pattern_metadata": {
#       "peak_hour": 8,
#       "confidence": 0.92,
#       "variance": 0.15,
#       "affected_queries": 500,
#       "z_score_count": 2.45,
#       "z_score_time": 1.87,
#       "days_observed": 28,
#       "consistency_score": 0.85,
#       "recurrence_score": 0.9333
#     },
#     "detection_timestamp": "2026-02-20T14:30:00Z",
#     "description": "Peak load detected at hour 8 UTC (92.0% confidence)",
#     "affected_query_count": 500
#   }
# ]
```

**Verification**:
- ✅ HTTP 200 status
- ✅ Returns array of patterns
- ✅ pattern_metadata contains all expected fields
- ✅ Confidence in 0.7-1.0 range
- ✅ peak_hour is correct

### Test 3: Filter by Pattern Type

```bash
# Filter specific pattern type
curl "http://localhost:8080/api/v1/workload-patterns?database_name=testdb&pattern_type=hourly_peak" \
  -H "Authorization: Bearer $TOKEN" | jq '.[] | .pattern_type'

# Expected: hourly_peak
```

**Verification**:
- ✅ Returns only hourly_peak patterns
- ✅ Filtering works correctly

### Test 4: Pagination (Limit)

```bash
# Test limit parameter
curl "http://localhost:8080/api/v1/workload-patterns?limit=1" \
  -H "Authorization: Bearer $TOKEN" | jq '. | length'

# Expected: 1 (respects limit)
```

**Verification**:
- ✅ Returns at most 1 result
- ✅ Default limit of 50 works
- ✅ Max limit of 1000 works

### Test 5: Error Handling

```bash
# Test 1: Missing database_name
curl -X POST http://localhost:8080/api/v1/workload-patterns/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"lookback_days": 30}' | jq .

# Expected: 400 Bad Request with error message

# Test 2: Invalid lookback_days
curl -X POST http://localhost:8080/api/v1/workload-patterns/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"database_name": "testdb", "lookback_days": 3}' | jq .

# Expected: 400 Bad Request (minimum 7 days)

# Test 3: Unauthorized access
curl -X POST http://localhost:8080/api/v1/workload-patterns/analyze \
  -H "Content-Type: application/json" \
  -d '{"database_name": "testdb"}' | jq .

# Expected: 401 Unauthorized
```

**Verification**:
- ✅ Missing required field returns 400
- ✅ Invalid lookback_days returns 400 with message
- ✅ Missing auth token returns 401
- ✅ Error messages are informative

---

## Level 3: Integration Testing

### Full Workflow Test

```bash
#!/bin/bash

set -e

# 1. Create test database data
psql -U postgres -d pganalytics << 'SQL'
-- Truncate test data
DELETE FROM metrics_pg_stats_query WHERE database_name = 'integration_test_db';

-- Insert 30 days of peak data at hour 9
INSERT INTO metrics_pg_stats_query (database_name, query_hash, calls, mean_exec_time_ms, collected_at)
WITH RECURSIVE dates AS (
    SELECT (NOW() - INTERVAL '30 days')::TIMESTAMP as d
    UNION ALL
    SELECT d + INTERVAL '1 hour' FROM dates WHERE d < NOW()
),
hourly_data AS (
    SELECT
        d,
        EXTRACT(HOUR FROM d)::INT as hour_of_day,
        CASE WHEN EXTRACT(HOUR FROM d)::INT = 9 THEN 600 ELSE 60 END,
        CASE WHEN EXTRACT(HOUR FROM d)::INT = 9 THEN 300.0 ELSE 60.0 END
    FROM dates
)
SELECT 'integration_test_db'::VARCHAR, 1::BIGINT, * FROM hourly_data;

SELECT COUNT(*) as rows_inserted FROM metrics_pg_stats_query
WHERE database_name = 'integration_test_db';
SQL

echo "✓ Test data created (720 rows)"

# 2. Call API to detect patterns
PATTERNS=$(curl -s -X POST http://localhost:8080/api/v1/workload-patterns/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"database_name": "integration_test_db", "lookback_days": 30}')

PATTERN_COUNT=$(echo $PATTERNS | jq '.patterns_detected')
echo "✓ Detected $PATTERN_COUNT patterns"

if [ "$PATTERN_COUNT" -lt 1 ]; then
    echo "✗ ERROR: Expected at least 1 pattern"
    exit 1
fi

# 3. Retrieve patterns
PATTERNS_LIST=$(curl -s http://localhost:8080/api/v1/workload-patterns?database_name=integration_test_db \
  -H "Authorization: Bearer $TOKEN")

PEAK_HOUR=$(echo $PATTERNS_LIST | jq -r '.[0].pattern_metadata.peak_hour')
CONFIDENCE=$(echo $PATTERNS_LIST | jq -r '.[0].pattern_metadata.confidence')

echo "✓ Retrieved patterns"
echo "  - Peak Hour: $PEAK_HOUR (expected: 9)"
echo "  - Confidence: $CONFIDENCE (expected: >0.8)"

# 4. Verify peak hour
if [ "$PEAK_HOUR" != "9" ]; then
    echo "✗ ERROR: Peak hour mismatch (got $PEAK_HOUR, expected 9)"
    exit 1
fi

# 5. Verify confidence
CONF_NUM=$(echo "$CONFIDENCE" | bc)
if (( $(echo "$CONF_NUM < 0.8" | bc -l) )); then
    echo "✗ ERROR: Confidence too low (got $CONFIDENCE, expected >0.8)"
    exit 1
fi

echo "✓ All integration tests passed!"
```

---

## Performance Testing

### Test Setup

```bash
# 1. Create large dataset (90 days × 1000 queries/hour)
time psql -U postgres -d pganalytics << 'SQL'
INSERT INTO metrics_pg_stats_query (database_name, query_hash, calls, mean_exec_time_ms, collected_at)
WITH RECURSIVE dates AS (
    SELECT (NOW() - INTERVAL '90 days')::TIMESTAMP as d
    UNION ALL
    SELECT d + INTERVAL '1 minute' FROM dates WHERE d < NOW()
)
SELECT
    'perf_test_db',
    (random() * 1000)::BIGINT,
    (random() * 1000)::BIGINT,
    (random() * 500)::FLOAT,
    d
FROM dates
LIMIT 129600; -- 90 days × 1440 minutes
SQL
```

### Performance Metrics

```bash
# Test execution time
time psql -U postgres -d pganalytics -c "SELECT COUNT(*) FROM detect_workload_patterns('perf_test_db', 90);"

# Expected: <2 seconds for 90 days of data
```

---

## Dashboard Testing

### Add Grafana Panels

```json
{
  "title": "Workload Patterns - Hourly Peak",
  "type": "timeseries",
  "targets": [
    {
      "expr": "SELECT hour_of_day, avg_count FROM workload_patterns WHERE database_name = '$database'"
    }
  ]
}
```

### Manual Verification

1. Navigate to Grafana dashboard: `http://localhost:3000/d/query-performance`
2. Look for "Workload Patterns" panel
3. Verify data displays correctly
4. Verify peak hours highlighted
5. Verify confidence scores show

---

## Regression Testing

### Verify Phase 4.4 Still Works

```bash
# Test that existing endpoints still function
curl http://localhost:8080/api/v1/databases \
  -H "Authorization: Bearer $TOKEN" | jq '.[] | .name'

# Expected: List of databases
```

---

## Test Automation Script

```bash
#!/bin/bash
# save as: scripts/test-workload-patterns.sh

set -e

echo "=== Phase 4.5.1: Workload Pattern Detection Tests ==="

# Test database connectivity
echo "Connecting to database..."
psql -U postgres -d pganalytics -c "SELECT 1;" > /dev/null

# Test SQL function
echo "Testing SQL function..."
RESULT=$(psql -U postgres -d pganalytics -c "SELECT COUNT(*) FROM detect_workload_patterns('testdb', 30);" | tail -1 | tr -d ' ')
echo "✓ SQL function works (result: $RESULT patterns)"

# Test API endpoint
echo "Testing API endpoint..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/workload-patterns/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"database_name": "testdb", "lookback_days": 30}')

STATUS=$(echo $RESPONSE | jq -r '.patterns_detected')
echo "✓ API endpoint works (patterns detected: $STATUS)"

echo ""
echo "All tests passed! ✓"
```

---

## Coverage Checklist

- [ ] Test 1.1: Basic pattern detection (SQL)
- [ ] Test 1.2: Pattern metadata validation
- [ ] Test 1.3: Multiple peak hours
- [ ] Test 1.4: Insufficient data handling
- [ ] Test 1.5: Edge case - no peaks
- [ ] Test 2.1: Detect patterns API
- [ ] Test 2.2: Get patterns API
- [ ] Test 2.3: Filter by pattern type
- [ ] Test 2.4: Pagination and limits
- [ ] Test 2.5: Error handling
- [ ] Test 3.1: Full integration workflow
- [ ] Test 3.2: Performance (< 2 seconds)
- [ ] Test Dashboard: Panel display
- [ ] Test Regression: Phase 4.4 endpoints

---

## Success Metrics

- ✅ All 14 tests pass
- ✅ Pattern detection accuracy > 80%
- ✅ Confidence scores 0.7-0.95 range
- ✅ API response time < 1 second
- ✅ SQL execution time < 2 seconds
- ✅ No false positives (< 10% rate)
- ✅ Error handling for edge cases
- ✅ Dashboard displays patterns

---

**Testing Status**: Ready to execute
**Estimated Time**: 45-60 minutes
**Success Criteria**: 14/14 tests pass
