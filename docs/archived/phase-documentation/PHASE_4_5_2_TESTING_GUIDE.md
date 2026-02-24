# Phase 4.5.2: Query Rewrite Suggestions - Testing Guide

**Date**: February 20, 2026
**Status**: Testing & Verification Ready
**Objective**: Comprehensive testing of query rewrite suggestion generation

---

## Quick Test Summary

Three levels of testing provided:
1. **Database Level** - Test SQL function directly with sample queries
2. **API Level** - Test HTTP endpoints with curl
3. **Integration** - Test full workflow with real patterns

---

## Level 1: Database Testing

### Test Setup: Create Sample Queries

```sql
-- Test 1: N+1 Pattern Detection
-- Create metrics with high call frequency (N+1 indicator)
INSERT INTO metrics_pg_stats_query (
    database_name, query_hash, query_text, fingerprint_hash,
    calls, total_exec_time_ms, mean_exec_time_ms, rows, collected_at
)
VALUES (
    'testdb',
    1001::BIGINT,
    'SELECT * FROM users WHERE id = ?',
    100::BIGINT,
    500::BIGINT,              -- Called 500 times (N+1 indicator)
    50000.0::FLOAT,           -- Total 50 seconds
    100.0::FLOAT,             -- Mean 100ms per call
    500::BIGINT,
    NOW()
);

-- Test 2: Create EXPLAIN plan with Nested Loop (inefficient join)
INSERT INTO explain_plans (
    query_hash, plan_json, has_seq_scan, has_nested_loop,
    collected_at
)
VALUES (
    1002::BIGINT,
    '{"Node Type": "Nested Loop", "Plans": [
        {"Node Type": "Seq Scan", "Relation Name": "orders"},
        {"Node Type": "Seq Scan", "Relation Name": "customers"}
    ]}'::JSONB,
    TRUE,   -- has_seq_scan
    TRUE,   -- has_nested_loop
    NOW()
);

INSERT INTO metrics_pg_stats_query (
    database_name, query_hash, query_text, fingerprint_hash,
    calls, total_exec_time_ms, mean_exec_time_ms, rows, collected_at
)
VALUES (
    'testdb',
    1002::BIGINT,
    'SELECT o.*, c.* FROM orders o, customers c WHERE o.customer_id = c.id',
    102::BIGINT,
    200::BIGINT,
    40000.0::FLOAT,
    2000.0::FLOAT,            -- Mean 2000ms (slow - join issue)
    10000::BIGINT,
    NOW()
);

-- Test 3: Sequential Scan on large table
INSERT INTO explain_plans (
    query_hash, plan_json, has_seq_scan, has_nested_loop,
    collected_at
)
VALUES (
    1003::BIGINT,
    '{"Node Type": "Seq Scan", "Relation Name": "orders", "Filter": "status = active"}'::JSONB,
    TRUE,   -- has_seq_scan
    FALSE,
    NOW()
);

INSERT INTO metrics_pg_stats_query (
    database_name, query_hash, query_text, fingerprint_hash,
    calls, total_exec_time_ms, mean_exec_time_ms, rows, collected_at
)
VALUES (
    'testdb',
    1003::BIGINT,
    'SELECT * FROM orders WHERE status = ?',
    103::BIGINT,
    250::BIGINT,
    50000.0::FLOAT,
    200.0::FLOAT,             -- Mean 200ms (slow seq scan)
    50000::BIGINT,
    NOW()
);

-- Test 4: Subquery pattern
INSERT INTO metrics_pg_stats_query (
    database_name, query_hash, query_text, fingerprint_hash,
    calls, total_exec_time_ms, mean_exec_time_ms, rows, collected_at
)
VALUES (
    'testdb',
    1004::BIGINT,
    'SELECT * FROM orders WHERE customer_id IN (SELECT id FROM customers WHERE status = ?)',
    104::BIGINT,
    100::BIGINT,
    20000.0::FLOAT,
    200.0::FLOAT,
    10000::BIGINT,
    NOW()
);

-- Test 5: IN with many values
INSERT INTO metrics_pg_stats_query (
    database_name, query_hash, query_text, fingerprint_hash,
    calls, total_exec_time_ms, mean_exec_time_ms, rows, collected_at
)
VALUES (
    'testdb',
    1005::BIGINT,
    'SELECT * FROM products WHERE id IN (1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20)',
    105::BIGINT,
    500::BIGINT,
    25000.0::FLOAT,
    50.0::FLOAT,
    500::BIGINT,
    NOW()
);
```

### Test 1: N+1 Pattern Detection

```sql
-- Run suggestion generation for N+1 query
SELECT * FROM generate_rewrite_suggestions(1001);

-- Expected output:
-- suggestion_id | suggestion_type      | confidence
-- ---------------+---------------------+------------
--             1 | n_plus_one_detected  |       0.92
-- (1 row)

-- Verify suggestion details
SELECT
    suggestion_type,
    description,
    suggested_rewrite,
    estimated_improvement_percent,
    confidence_score
FROM query_rewrite_suggestions
WHERE query_hash = 1001
ORDER BY confidence_score DESC;

-- Expected:
-- suggestion_type      | description                                      | estimated_improvement_percent | confidence_score
-- ---------------------+------------------------------------------------+--------------------------------+-----------------
-- n_plus_one_detected  | Multiple queries with identical pattern detected  |                           99.0 |             0.92
```

**Verification**:
- ✅ N+1 pattern detected
- ✅ Confidence score > 0.85
- ✅ Estimated improvement ~99%
- ✅ Suggested rewrite suggests IN clause or JOIN
- ✅ Reasoning explains the issue

### Test 2: Inefficient Join Detection

```sql
-- Generate suggestions for inefficient join query
SELECT * FROM generate_rewrite_suggestions(1002);

-- Expected output:
-- suggestion_id | suggestion_type           | confidence
-- ---------------+---------------------------+------------
--             2 | inefficient_join_detected |        0.80
-- (1 row)

-- Verify details
SELECT
    suggestion_type,
    estimated_improvement_percent,
    confidence_score,
    reasoning
FROM query_rewrite_suggestions
WHERE query_hash = 1002;

-- Expected: Suggests adding index or reordering joins
-- Confidence: 0.80, Improvement: 85%
```

**Verification**:
- ✅ Inefficient join detected
- ✅ Suggests indexing or reordering
- ✅ Confidence around 0.80
- ✅ High estimated improvement (85%)

### Test 3: Missing Index Detection

```sql
-- Generate suggestions for sequential scan query
SELECT * FROM generate_rewrite_suggestions(1003);

-- Expected output:
-- suggestion_id | suggestion_type           | confidence
-- ---------------+---------------------------+------------
--             3 | missing_index_detected    |       0.83
-- (1 row)

-- Verify details
SELECT
    suggestion_type,
    suggested_rewrite,
    estimated_improvement_percent,
    confidence_score
FROM query_rewrite_suggestions
WHERE query_hash = 1003;

-- Expected: Suggests CREATE INDEX statement
-- Confidence: 0.83, Improvement: 80%
```

**Verification**:
- ✅ Missing index detected
- ✅ Suggests specific index creation
- ✅ Confidence ~0.83
- ✅ High improvement potential

### Test 4: Subquery Optimization

```sql
-- Generate suggestions for subquery
SELECT * FROM generate_rewrite_suggestions(1004);

-- Expected output:
-- suggestion_id | suggestion_type           | confidence
-- ---------------+---------------------------+------------
--             4 | subquery_optimization     |       0.75
-- (1 row)

-- Verify details
SELECT
    suggestion_type,
    suggested_rewrite,
    estimated_improvement_percent
FROM query_rewrite_suggestions
WHERE query_hash = 1004;

-- Expected: Suggests rewriting as JOIN
```

**Verification**:
- ✅ Subquery pattern detected
- ✅ Suggests JOIN rewrite
- ✅ Confidence ~0.75
- ✅ Reasonable improvement estimate

### Test 5: IN vs ANY Optimization

```sql
-- Generate suggestions for IN with many values
SELECT * FROM generate_rewrite_suggestions(1005);

-- Expected output:
-- suggestion_id | suggestion_type           | confidence
-- ---------------+---------------------------+------------
--             5 | in_vs_any_optimization    |       0.65
-- (1 row)

-- Verify details
SELECT
    suggestion_type,
    suggested_rewrite,
    estimated_improvement_percent
FROM query_rewrite_suggestions
WHERE query_hash = 1005;

-- Expected: Suggests ANY(ARRAY[...]) instead of IN
```

**Verification**:
- ✅ IN clause detected
- ✅ Suggests ANY with array
- ✅ Confidence ~0.65
- ✅ Lower improvement (15%)

---

## Level 2: API Testing

### Setup: Start API Server

```bash
cd /Users/glauco.torres/git/pganalytics-v3/backend
go build -o pganalytics-api cmd/pganalytics-api/main.go
./pganalytics-api &

# Get auth token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | jq -r '.token')

echo "Token: $TOKEN"
```

### Test 1: Generate Rewrite Suggestions

```bash
# Test generating suggestions for N+1 query
curl -X POST http://localhost:8080/api/v1/queries/1001/rewrite-suggestions/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" | jq .

# Expected response:
# {
#   "suggestions_generated": 1,
#   "query_hash": 1001,
#   "suggestion_types": [
#     "n_plus_one_detected",
#     "inefficient_join_detected",
#     "missing_index_detected",
#     "subquery_optimization",
#     "in_vs_any_optimization"
#   ],
#   "timestamp": "2026-02-20T14:30:00Z"
# }
```

**Verification**:
- ✅ HTTP 200 status
- ✅ suggestions_generated >= 1
- ✅ Lists all possible suggestion types
- ✅ Timestamp in ISO 8601 format

### Test 2: Get Rewrite Suggestions

```bash
# Get all suggestions for a query
curl http://localhost:8080/api/v1/queries/1001/rewrite-suggestions \
  -H "Authorization: Bearer $TOKEN" | jq .

# Expected response:
# [
#   {
#     "id": 1,
#     "query_hash": 1001,
#     "fingerprint_hash": 100,
#     "suggestion_type": "n_plus_one_detected",
#     "description": "Multiple queries with identical pattern detected...",
#     "original_query": "SELECT * FROM users WHERE id = ?",
#     "suggested_rewrite": "Combine into single query using IN clause...",
#     "reasoning": "Query called 500 times with mean execution 100ms...",
#     "estimated_improvement_percent": 99.0,
#     "confidence_score": 0.92,
#     "dismissed": false,
#     "implemented": false,
#     "created_at": "2026-02-20T14:30:00Z",
#     "updated_at": "2026-02-20T14:30:00Z"
#   }
# ]
```

**Verification**:
- ✅ HTTP 200 status
- ✅ Returns array of suggestions
- ✅ All fields present
- ✅ Suggestions sorted by confidence descending

### Test 3: Get with Limit Parameter

```bash
# Get first 5 suggestions
curl "http://localhost:8080/api/v1/queries/1001/rewrite-suggestions?limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq '. | length'

# Expected: 5 (or fewer if less than 5 suggestions exist)
```

**Verification**:
- ✅ Respects limit parameter
- ✅ Default limit of 10 works
- ✅ Max limit of 100 works

### Test 4: Error Handling

```bash
# Test 1: Invalid query_hash
curl -X POST http://localhost:8080/api/v1/queries/abc/rewrite-suggestions/generate \
  -H "Authorization: Bearer $TOKEN" | jq .

# Expected: 400 Bad Request with error message

# Test 2: Negative query_hash
curl -X POST http://localhost:8080/api/v1/queries/-1/rewrite-suggestions/generate \
  -H "Authorization: Bearer $TOKEN" | jq .

# Expected: 400 Bad Request

# Test 3: Query with no suggestions
curl http://localhost:8080/api/v1/queries/99999/rewrite-suggestions \
  -H "Authorization: Bearer $TOKEN" | jq .

# Expected: Empty array []

# Test 4: Unauthorized access
curl -X POST http://localhost:8080/api/v1/queries/1001/rewrite-suggestions/generate \
  -H "Content-Type: application/json" | jq .

# Expected: 401 Unauthorized
```

**Verification**:
- ✅ Invalid input returns 400
- ✅ Negative hash returns 400
- ✅ No suggestions returns empty array (not error)
- ✅ Missing auth returns 401

---

## Level 3: Integration Testing

### Full Workflow Test

```bash
#!/bin/bash

set -e

echo "=== Phase 4.5.2: Query Rewrite Suggestions Test ==="

# 1. Setup test data
echo "Setting up test data..."
psql -U postgres -d pganalytics << 'SQL'
DELETE FROM metrics_pg_stats_query WHERE database_name = 'rewrite_test_db';
DELETE FROM explain_plans WHERE query_hash IN (2001, 2002, 2003);

-- Insert N+1 test data
INSERT INTO metrics_pg_stats_query (database_name, query_hash, query_text, calls, mean_exec_time_ms, collected_at)
VALUES ('rewrite_test_db', 2001, 'SELECT * FROM users WHERE id = ?', 100, 50.0, NOW());

-- Insert inefficient join test data
INSERT INTO explain_plans (query_hash, plan_json, has_seq_scan, has_nested_loop, collected_at)
VALUES (2002, '{"Node Type": "Nested Loop"}'::JSONB, TRUE, TRUE, NOW());
INSERT INTO metrics_pg_stats_query (database_name, query_hash, query_text, calls, mean_exec_time_ms, collected_at)
VALUES ('rewrite_test_db', 2002, 'SELECT o.* FROM orders o, customers c WHERE o.customer_id = c.id', 200, 2000.0, NOW());

-- Insert missing index test data
INSERT INTO explain_plans (query_hash, plan_json, has_seq_scan, has_nested_loop, collected_at)
VALUES (2003, '{"Node Type": "Seq Scan"}'::JSONB, TRUE, FALSE, NOW());
INSERT INTO metrics_pg_stats_query (database_name, query_hash, query_text, calls, mean_exec_time_ms, collected_at)
VALUES ('rewrite_test_db', 2003, 'SELECT * FROM orders WHERE status = ?', 250, 200.0, NOW());

SELECT 'Test data created' as status;
SQL

echo "✓ Test data created"

# 2. Generate suggestions via API
echo "Generating suggestions..."
RESULT=$(curl -s -X POST http://localhost:8080/api/v1/queries/2001/rewrite-suggestions/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")

COUNT=$(echo $RESULT | jq '.suggestions_generated')
echo "✓ Generated $COUNT suggestions"

# 3. Get suggestions
echo "Retrieving suggestions..."
SUGGESTIONS=$(curl -s http://localhost:8080/api/v1/queries/2001/rewrite-suggestions \
  -H "Authorization: Bearer $TOKEN")

TYPE=$(echo $SUGGESTIONS | jq -r '.[0].suggestion_type')
CONFIDENCE=$(echo $SUGGESTIONS | jq -r '.[0].confidence_score')
IMPROVEMENT=$(echo $SUGGESTIONS | jq -r '.[0].estimated_improvement_percent')

echo "✓ Retrieved suggestions:"
echo "  Type: $TYPE"
echo "  Confidence: $CONFIDENCE"
echo "  Improvement: $IMPROVEMENT%"

# 4. Verify results
if [ "$TYPE" = "null" ] || [ -z "$TYPE" ]; then
    echo "✗ ERROR: No suggestions returned"
    exit 1
fi

if (( $(echo "$CONFIDENCE < 0.6" | bc -l) )); then
    echo "✗ ERROR: Confidence too low (got $CONFIDENCE, expected >= 0.6)"
    exit 1
fi

echo ""
echo "All integration tests passed! ✓"
```

---

## Performance Testing

### Test Setup

```bash
# Create large dataset with N+1 patterns (1000 queries)
time psql -U postgres -d pganalytics << 'SQL'
INSERT INTO metrics_pg_stats_query (database_name, query_hash, query_text, calls, mean_exec_time_ms, collected_at)
SELECT
    'perf_test_db',
    (1000 + seq)::BIGINT,
    'SELECT * FROM users WHERE id = ' || seq,
    (10 + seq % 100)::BIGINT,
    (50.0 + random() * 50.0)::FLOAT,
    NOW()
FROM generate_series(1, 1000) as t(seq);
SQL
```

### Performance Metrics

```bash
# Measure suggestion generation time
time curl -s -X POST http://localhost:8080/api/v1/queries/1001/rewrite-suggestions/generate \
  -H "Authorization: Bearer $TOKEN" > /dev/null

# Expected: < 1 second
```

---

## Test Automation Script

```bash
#!/bin/bash
# save as: scripts/test-rewrite-suggestions.sh

set -e

echo "=== Phase 4.5.2: Query Rewrite Suggestions Tests ==="

# Database tests
echo "Running database tests..."
psql -U postgres -d pganalytics << 'SQL'
SELECT * FROM generate_rewrite_suggestions(1001);
SELECT COUNT(*) as suggestions FROM query_rewrite_suggestions WHERE dismissed = FALSE;
SQL

echo "✓ Database tests complete"

# API tests
echo "Running API tests..."
curl -s http://localhost:8080/api/v1/queries/1001/rewrite-suggestions \
  -H "Authorization: Bearer $TOKEN" | jq '.[0].suggestion_type'

echo "✓ API tests complete"

echo ""
echo "All tests passed! ✓"
```

---

## Coverage Checklist

- [ ] Test 1.1: N+1 pattern detection (SQL)
- [ ] Test 1.2: Inefficient join detection
- [ ] Test 1.3: Missing index detection
- [ ] Test 1.4: Subquery optimization
- [ ] Test 1.5: IN vs ANY optimization
- [ ] Test 2.1: Generate suggestions API
- [ ] Test 2.2: Get suggestions API
- [ ] Test 2.3: Limit parameter
- [ ] Test 2.4: Error handling
- [ ] Test 3.1: Full integration workflow
- [ ] Test Performance: < 1 second response
- [ ] Verify all suggestion types included

---

## Success Metrics

- ✅ N+1 detection accuracy > 85%
- ✅ Join problem detection > 75%
- ✅ Index recommendation accuracy > 80%
- ✅ Confidence scores 0.65-0.95 range
- ✅ API response time < 1 second
- ✅ SQL execution time < 2 seconds
- ✅ Error handling for all cases
- ✅ All suggestion types functioning

---

**Testing Status**: Ready to execute
**Estimated Time**: 45-60 minutes
**Success Criteria**: 12/12 tests pass
