# Phase 4.5.3: Parameter Optimization Recommendations - Testing Guide

**Date**: February 20, 2026
**Status**: Testing Procedures Ready
**Duration**: Estimated 2-3 hours for full test suite

---

## Quick Test (5-10 minutes)

### Setup: Insert Test Data
```sql
-- Insert test queries into metrics table
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, calls_per_sec, database_name
)
VALUES
(3001, 'SELECT * FROM events WHERE type = $1', 15.0, 250, 1, 0.5, 'testdb'),
(3002, 'SELECT * FROM orders ORDER BY created_at DESC', 450.0, 50, 10000, 0.1, 'testdb'),
(3003, 'SELECT * FROM products WHERE active = true', 180.0, 5000, 15000, 10.0, 'testdb')
ON CONFLICT (query_hash) DO UPDATE SET
    mean_exec_time_ms = EXCLUDED.mean_exec_time_ms,
    calls = EXCLUDED.calls,
    rows = EXCLUDED.rows;
```

### Test 1: Generate LIMIT Recommendations
```bash
# Generate suggestions for query with large result set
curl -X POST http://localhost:8080/api/v1/queries/3003/parameter-optimization/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# Expected Response (200 OK):
# {
#   "query_hash": 3003,
#   "suggestions_count": 1,
#   "suggestion_types": ["LIMIT"],
#   "generated_at": "2026-02-20T..."
# }
```

### Test 2: Get Parameter Suggestions
```bash
curl "http://localhost:8080/api/v1/queries/3003/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN"

# Expected Response (200 OK):
# {
#   "query_hash": 3003,
#   "suggestions": [
#     {
#       "parameter_name": "LIMIT",
#       "recommended_value": "LIMIT 1500",
#       "confidence_score": 0.95,
#       "estimated_improvement_percent": 85.0
#     }
#   ],
#   "count": 1,
#   "parameter_types": {"LIMIT": 1}
# }
```

---

## Database Level Tests (5 test cases)

### Test 1: LIMIT Recommendation for Large Result Sets

**Objective**: Verify LIMIT suggestions are created for queries returning large result sets

**Setup**:
```sql
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (
    3101,
    'SELECT * FROM events WHERE timestamp > now() - interval ''7 days''',
    250.0,
    50,
    25000,
    'testdb'
)
ON CONFLICT (query_hash) DO UPDATE SET
    mean_exec_time_ms = EXCLUDED.mean_exec_time_ms,
    rows = EXCLUDED.rows;
```

**Execute**:
```sql
SELECT * FROM optimize_parameters(3101);

-- Verify suggestion was created
SELECT parameter_name, recommended_value, confidence_score, estimated_improvement_percent
FROM parameter_tuning_suggestions
WHERE query_hash = 3101 AND parameter_name = 'LIMIT';
```

**Expected Results**:
- ✓ 1 suggestion created with parameter_name = 'LIMIT'
- ✓ recommended_value = 'LIMIT 2500' (25000 / 10)
- ✓ confidence_score >= 0.90 (large result set)
- ✓ estimated_improvement_percent >= 80.0

**Failure Indicators**:
- ✗ No suggestion created
- ✗ confidence_score < 0.85
- ✗ estimated_improvement_percent < 50

---

### Test 2: work_mem Optimization for Sort Operations

**Objective**: Verify work_mem recommendations are generated for queries with ORDER BY/GROUP BY

**Setup**:
```sql
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (
    3102,
    'SELECT category, COUNT(*) as cnt FROM products GROUP BY category',
    450.0,
    75,
    150,
    'testdb'
)
ON CONFLICT (query_hash) DO UPDATE SET
    mean_exec_time_ms = EXCLUDED.mean_exec_time_ms,
    calls = EXCLUDED.calls;
```

**Execute**:
```sql
SELECT * FROM optimize_parameters(3102);

-- Verify work_mem suggestion
SELECT parameter_name, current_value, recommended_value, confidence_score
FROM parameter_tuning_suggestions
WHERE query_hash = 3102 AND parameter_name = 'work_mem';
```

**Expected Results**:
- ✓ 1 suggestion with parameter_name = 'work_mem'
- ✓ current_value matches current PostgreSQL setting (typically '4MB')
- ✓ recommended_value suggests increase (e.g., '6MB')
- ✓ confidence_score = 0.85-0.90 (sort operation + slow execution)
- ✓ estimated_improvement_percent = 25-35%

**Verification**:
```sql
-- Verify recommendation is based on execution time threshold
SELECT mean_exec_time_ms, calls FROM metrics_pg_stats_query WHERE query_hash = 3102;
-- Should have: mean_exec_time_ms > 200 AND calls > 10
```

---

### Test 3: Batch Size Recommendations for High Frequency Queries

**Objective**: Verify batch size suggestions for N+1 pattern queries

**Setup**:
```sql
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (
    3103,
    'SELECT * FROM users WHERE id = $1',
    12.0,
    500,
    1,
    'testdb'
)
ON CONFLICT (query_hash) DO UPDATE SET
    mean_exec_time_ms = EXCLUDED.mean_exec_time_ms,
    calls = EXCLUDED.calls;
```

**Execute**:
```sql
SELECT * FROM optimize_parameters(3103);

-- Verify batch size suggestions
SELECT parameter_name, recommended_value, confidence_score, estimated_improvement_percent
FROM parameter_tuning_suggestions
WHERE query_hash = 3103 AND parameter_name = 'batch_size'
ORDER BY confidence_score DESC;
```

**Expected Results**:
- ✓ 3 suggestions created (batch sizes: 50, 100, 500)
- ✓ Confidence scores: 0.75 (batch 50), 0.73 (batch 100), 0.70 (batch 500)
- ✓ All improvement percentages: 70-75%
- ✓ recommended_value contains integer (e.g., '50', '100', '500')

**Verification**:
```sql
-- Count total suggestions
SELECT COUNT(*) FROM parameter_tuning_suggestions
WHERE query_hash = 3103;
-- Should be 3 (three batch size options)
```

---

### Test 4: Combined Suggestions for Complex Query

**Objective**: Verify multiple suggestion types can be generated for the same query

**Setup**:
```sql
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (
    3104,
    'SELECT * FROM orders WHERE status = $1 ORDER BY created_at',
    280.0,
    200,
    8000,
    'testdb'
)
ON CONFLICT (query_hash) DO UPDATE SET
    mean_exec_time_ms = EXCLUDED.mean_exec_time_ms,
    calls = EXCLUDED.calls,
    rows = EXCLUDED.rows;
```

**Execute**:
```sql
SELECT * FROM optimize_parameters(3104);

-- Verify multiple suggestion types
SELECT DISTINCT parameter_name
FROM parameter_tuning_suggestions
WHERE query_hash = 3104
ORDER BY parameter_name;
```

**Expected Results**:
- ✓ At least 2 suggestion types created
- ✓ Should include: LIMIT (no LIMIT clause + large result set) and work_mem (ORDER BY + slow)
- ✓ May include: batch_size (high call count: 200)
- ✓ Total suggestions: 3-5

**Verification**:
```sql
-- Check suggestion details
SELECT parameter_name, confidence_score, estimated_improvement_percent
FROM parameter_tuning_suggestions
WHERE query_hash = 3104
ORDER BY confidence_score DESC;
```

---

### Test 5: Idempotency - No Duplicates on Repeated Calls

**Objective**: Verify ON CONFLICT handling prevents duplicate suggestions

**Setup**:
```sql
-- Same query as Test 1
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (3105, 'SELECT * FROM users WHERE active = true', 200.0, 100, 12000, 'testdb')
ON CONFLICT (query_hash) DO UPDATE SET mean_exec_time_ms = EXCLUDED.mean_exec_time_ms;
```

**Execute**:
```sql
-- Generate suggestions first time
SELECT * FROM optimize_parameters(3105);

-- Count suggestions
SELECT COUNT(*) as count_after_first FROM parameter_tuning_suggestions WHERE query_hash = 3105;
-- Record count: should be N (e.g., 2 or 3)

-- Generate suggestions second time (should NOT duplicate)
SELECT * FROM optimize_parameters(3105);

-- Count suggestions again
SELECT COUNT(*) as count_after_second FROM parameter_tuning_suggestions WHERE query_hash = 3105;
-- Should still be N (same as after first call)

-- Verify timestamps were updated
SELECT parameter_name, created_at, updated_at
FROM parameter_tuning_suggestions
WHERE query_hash = 3105
ORDER BY parameter_name;
-- updated_at should be more recent than created_at
```

**Expected Results**:
- ✓ Count remains the same after second call (no duplicates)
- ✓ updated_at timestamp changes (ON CONFLICT DO UPDATE executed)
- ✓ No ERROR or constraint violations
- ✓ Idempotent: can call repeatedly with same results

---

## API Level Tests (5 test cases)

### Test 1: Generate Parameter Suggestions - Success

**Endpoint**: `POST /api/v1/queries/{query_hash}/parameter-optimization/generate`

**Setup**:
```sql
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time_ms, calls, rows, database_name)
VALUES (4001, 'SELECT * FROM logs', 300.0, 200, 20000, 'testdb')
ON CONFLICT (query_hash) DO UPDATE SET mean_exec_time_ms = EXCLUDED.mean_exec_time_ms;
```

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/queries/4001/parameter-optimization/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

**Expected Response** (200 OK):
```json
{
  "query_hash": 4001,
  "suggestions_count": 2,
  "suggestion_types": ["LIMIT", "batch_size"],
  "generated_at": "2026-02-20T15:30:45Z",
  "message": "Generated 2 parameter optimization suggestions"
}
```

**Verification**:
- ✓ Status code 200
- ✓ suggestions_count > 0
- ✓ suggestion_types is non-empty array
- ✓ All types match generated suggestions in database

---

### Test 2: Get Parameter Suggestions - Success

**Endpoint**: `GET /api/v1/queries/{query_hash}/parameter-optimization?limit=10`

**Request**:
```bash
curl "http://localhost:8080/api/v1/queries/4001/parameter-optimization?limit=5" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response** (200 OK):
```json
{
  "query_hash": 4001,
  "suggestions": [
    {
      "id": 201,
      "query_hash": 4001,
      "parameter_name": "LIMIT",
      "current_value": "NOT SET",
      "recommended_value": "LIMIT 2000",
      "reasoning": "Query returns 20000 rows...",
      "estimated_improvement_percent": 85.0,
      "confidence_score": 0.95,
      "created_at": "2026-02-20T15:30:45Z"
    }
  ],
  "count": 2,
  "parameter_types": {
    "LIMIT": 1,
    "batch_size": 1
  },
  "timestamp": "2026-02-20T15:31:00Z"
}
```

**Verification**:
- ✓ Status code 200
- ✓ suggestions array matches query
- ✓ All fields populated correctly
- ✓ Suggestions ordered by confidence_score DESC
- ✓ parameter_types map accurate counts

---

### Test 3: Invalid Query Hash - Negative

**Endpoint**: `GET /api/v1/queries/{query_hash}/parameter-optimization`

**Request**:
```bash
curl "http://localhost:8080/api/v1/queries/-1/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response** (400 Bad Request):
```json
{
  "error": "query_hash must be positive"
}
```

**Verification**:
- ✓ Status code 400
- ✓ Clear error message
- ✓ No database error occurs

---

### Test 4: Invalid Query Hash - Zero

**Endpoint**: `POST /api/v1/queries/{query_hash}/parameter-optimization/generate`

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/queries/0/parameter-optimization/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

**Expected Response** (400 Bad Request):
```json
{
  "error": "query_hash must be positive"
}
```

**Verification**:
- ✓ Status code 400
- ✓ Validation prevents invalid input
- ✓ No database operations attempted

---

### Test 5: Non-existent Query - Empty Results

**Endpoint**: `GET /api/v1/queries/{query_hash}/parameter-optimization`

**Request**:
```bash
curl "http://localhost:8080/api/v1/queries/99999/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response** (200 OK with empty results):
```json
{
  "query_hash": 99999,
  "suggestions": [],
  "count": 0,
  "message": "No parameter optimization suggestions available"
}
```

**Verification**:
- ✓ Status code 200 (not 404)
- ✓ Empty suggestions array
- ✓ Count = 0
- ✓ Friendly message explaining no suggestions available

---

## Integration Tests (3 test cases)

### Test 1: Full Workflow - Generation to Retrieval

**Objective**: Complete end-to-end flow of generating and retrieving suggestions

**Steps**:
```bash
# Step 1: Insert test query
psql -c "
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time_ms, calls, rows)
VALUES (5001, 'SELECT * FROM events WHERE date > now() - 30 days', 350.0, 80, 50000);
"

# Step 2: Generate suggestions
curl -X POST http://localhost:8080/api/v1/queries/5001/parameter-optimization/generate \
  -H "Authorization: Bearer $TOKEN"
# Response should have suggestions_count > 0

# Step 3: Get all suggestions with limit=20
curl "http://localhost:8080/api/v1/queries/5001/parameter-optimization?limit=20" \
  -H "Authorization: Bearer $TOKEN" | jq '.count'
# Should match suggestions_count from step 2

# Step 4: Verify suggestions are actionable
curl "http://localhost:8080/api/v1/queries/5001/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN" | jq '.suggestions[] | {parameter_name, recommended_value}'
# Each suggestion should have specific recommended_value (e.g., "LIMIT 5000")
```

**Expected Outcomes**:
- ✓ Suggestion generation succeeds
- ✓ Retrieved count matches generated count
- ✓ All suggestions have specific, actionable recommendations
- ✓ Suggestions ordered by confidence score

---

### Test 2: Confidence Scoring Verification

**Objective**: Verify confidence scores reflect query characteristics correctly

**Setup**:
```sql
-- Query 1: VERY large result set (should have high LIMIT confidence)
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time_ms, calls, rows)
VALUES (5101, 'SELECT * FROM huge_table', 500.0, 10, 100000);

-- Query 2: Moderate result set
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time_ms, calls, rows)
VALUES (5102, 'SELECT * FROM medium_table', 200.0, 10, 5000);

-- Query 3: Fast query (little optimization opportunity)
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time_ms, calls, rows)
VALUES (5103, 'SELECT id FROM small_table WHERE id = $1', 5.0, 10, 1);
```

**Execute**:
```bash
# Generate for all three
for hash in 5101 5102 5103; do
  curl -X POST http://localhost:8080/api/v1/queries/$hash/parameter-optimization/generate \
    -H "Authorization: Bearer $TOKEN"
done

# Retrieve and compare confidence scores
for hash in 5101 5102 5103; do
  echo "Query $hash:"
  curl "http://localhost:8080/api/v1/queries/$hash/parameter-optimization" \
    -H "Authorization: Bearer $TOKEN" | jq '.suggestions[] | {param: .parameter_name, confidence: .confidence_score}'
done
```

**Expected Results**:
- ✓ Query 5101 (100K rows): confidence >= 0.95 for LIMIT
- ✓ Query 5102 (5K rows): confidence >= 0.85 for LIMIT
- ✓ Query 5103 (1 row): confidence <= 0.70 or fewer suggestions
- ✓ Confidence scores decrease as result set size decreases

---

### Test 3: Limit Parameter Validation

**Objective**: Verify limit parameter is properly validated and used

**Setup**:
```sql
-- Query with 5+ parameter suggestions
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time_ms, calls, rows)
VALUES (5201, 'SELECT * FROM data WHERE active = true GROUP BY type ORDER BY date DESC', 500.0, 150, 25000);

-- Generate suggestions
SELECT * FROM optimize_parameters(5201);
```

**Execute**:
```bash
# Test 1: Default limit (10)
curl "http://localhost:8080/api/v1/queries/5201/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN" | jq '.count'
# Should return max 10

# Test 2: Custom limit=20
curl "http://localhost:8080/api/v1/queries/5201/parameter-optimization?limit=20" \
  -H "Authorization: Bearer $TOKEN" | jq '.count'
# Should return up to 20

# Test 3: Invalid limit (should default to 10)
curl "http://localhost:8080/api/v1/queries/5201/parameter-optimization?limit=1000" \
  -H "Authorization: Bearer $TOKEN" | jq '.count'
# Should return max 10 (limit capped at 100)

# Test 4: Zero limit (should default to 10)
curl "http://localhost:8080/api/v1/queries/5201/parameter-optimization?limit=0" \
  -H "Authorization: Bearer $TOKEN" | jq '.count'
# Should return max 10
```

**Expected Results**:
- ✓ Default limit = 10
- ✓ Custom limits respected (up to 100)
- ✓ Invalid limits default to 10
- ✓ All suggestions returned in priority order

---

## Error Handling Tests (3 test cases)

### Test 1: Missing Authentication

**Request**:
```bash
curl http://localhost:8080/api/v1/queries/5301/parameter-optimization
# No Authorization header
```

**Expected Response** (401 Unauthorized):
```json
{
  "error": "Unauthorized",
  "message": "Authentication required"
}
```

---

### Test 2: Invalid Format Query Hash

**Request**:
```bash
curl "http://localhost:8080/api/v1/queries/not_a_number/parameter-optimization/generate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

**Expected Response** (400 Bad Request):
```json
{
  "error": "Invalid query_hash format"
}
```

---

### Test 3: Database Connection Error Handling

**Objective**: Verify graceful error handling if database becomes unavailable

**Simulate**:
```bash
# Stop PostgreSQL temporarily
systemctl stop postgresql

# Try to generate suggestions
curl -X POST http://localhost:8080/api/v1/queries/5401/parameter-optimization/generate \
  -H "Authorization: Bearer $TOKEN"

# Expected: 500 error with clear message
# {"error": "Failed to generate parameter suggestions"}

# Restart PostgreSQL
systemctl start postgresql
```

---

## Performance Tests

### Test 1: Response Time - Generate Suggestions

**Objective**: Verify generation completes within timeout

**Execute**:
```bash
time curl -X POST http://localhost:8080/api/v1/queries/6001/parameter-optimization/generate \
  -H "Authorization: Bearer $TOKEN"
# Should complete in < 15 seconds (handler timeout)
```

**Expected**:
- ✓ Response time < 3 seconds (typical)
- ✓ Never exceeds 15 second timeout

---

### Test 2: Response Time - Get Suggestions

**Objective**: Verify retrieval with limit=100 still responds quickly

**Setup**:
```sql
-- Insert query with many potential suggestions
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time_ms, calls, rows)
VALUES (6101, 'COMPLEX QUERY', 600.0, 500, 100000);

SELECT * FROM optimize_parameters(6101);
```

**Execute**:
```bash
time curl "http://localhost:8080/api/v1/queries/6101/parameter-optimization?limit=100" \
  -H "Authorization: Bearer $TOKEN"
# Should complete in < 1 second
```

**Expected**:
- ✓ Response time < 500ms (typical)
- ✓ Never exceeds 10 second timeout

---

## Test Automation Script

**File**: `test_parameter_optimization.sh`

```bash
#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

TOKEN="your_token_here"
BASE_URL="http://localhost:8080"

pass_count=0
fail_count=0

log_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((pass_count++))
}

log_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((fail_count++))
}

log_test() {
    echo -e "${YELLOW}→ TEST${NC}: $1"
}

# Test 1: Invalid negative query_hash
log_test "Invalid negative query_hash"
response=$(curl -s "$BASE_URL/api/v1/queries/-1/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN")
if echo "$response" | grep -q "must be positive"; then
    log_pass "Negative query_hash properly rejected"
else
    log_fail "Negative query_hash validation"
fi

# Test 2: Generate suggestions for existing query
log_test "Generate parameter suggestions"
response=$(curl -s -X POST "$BASE_URL/api/v1/queries/4001/parameter-optimization/generate" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json")
if echo "$response" | grep -q '"suggestions_count"'; then
    log_pass "Parameter suggestions generated successfully"
else
    log_fail "Parameter suggestion generation"
fi

# Test 3: Get suggestions with default limit
log_test "Get suggestions with default limit"
response=$(curl -s "$BASE_URL/api/v1/queries/4001/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN")
count=$(echo "$response" | jq '.count' 2>/dev/null)
if [ "$count" -le 10 ] 2>/dev/null; then
    log_pass "Default limit (10) applied correctly"
else
    log_fail "Default limit validation"
fi

# Test 4: Get suggestions with custom limit
log_test "Get suggestions with custom limit"
response=$(curl -s "$BASE_URL/api/v1/queries/4001/parameter-optimization?limit=20" \
  -H "Authorization: Bearer $TOKEN")
count=$(echo "$response" | jq '.count' 2>/dev/null)
if [ "$count" -le 20 ] 2>/dev/null; then
    log_pass "Custom limit applied correctly"
else
    log_fail "Custom limit validation"
fi

# Print summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test Summary:"
echo -e "  ${GREEN}Passed: $pass_count${NC}"
echo -e "  ${RED}Failed: $fail_count${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $fail_count -eq 0 ]; then
    exit 0
else
    exit 1
fi
```

---

## Success Criteria Checklist

### Database Tests
- [ ] Test 1: LIMIT recommendations created with correct confidence
- [ ] Test 2: work_mem recommendations for sort operations
- [ ] Test 3: Batch size recommendations (3 sizes recommended)
- [ ] Test 4: Multiple suggestion types can be combined
- [ ] Test 5: Idempotency verified (no duplicates on repeated calls)

### API Tests
- [ ] Test 1: Generate suggestions returns 200 with correct count
- [ ] Test 2: Get suggestions returns sorted array
- [ ] Test 3: Negative query_hash properly rejected (400)
- [ ] Test 4: Zero query_hash properly rejected (400)
- [ ] Test 5: Non-existent query returns empty results (200)

### Integration Tests
- [ ] Test 1: Full workflow generation → retrieval works correctly
- [ ] Test 2: Confidence scores vary by query characteristics
- [ ] Test 3: Limit parameter properly validated and applied

### Error Handling
- [ ] Missing authentication returns 401
- [ ] Invalid format query_hash returns 400
- [ ] Database errors handled gracefully (500 with message)

### Performance
- [ ] Generation completes within 15 seconds
- [ ] Retrieval with limit=100 completes within 10 seconds

---

## Known Test Limitations

1. **ML Predictions**: Phase 4.5.3 doesn't include ML service, so predictive tests are not included
2. **Real Data**: Tests use synthetic data; real production data may show different patterns
3. **Concurrent Load**: Single-threaded tests; concurrent scenarios not tested here
4. **Network Latency**: Tests assume low-latency local database

---

**Total Test Cases**: 16 (5 database + 5 API + 3 integration + 3 error handling)
**Estimated Duration**: 2-3 hours for complete test suite
**Critical Path**: DB Tests 1-5 → API Tests 1-2 → Integration Test 1

