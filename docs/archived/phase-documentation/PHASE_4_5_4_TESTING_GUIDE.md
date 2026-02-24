# Phase 4.5.4: ML-Powered Optimization Workflow - Testing Guide

**Date**: February 20, 2026
**Status**: Testing Procedures Ready
**Duration**: Estimated 3-4 hours for full test suite

---

## Quick Test (10-15 minutes)

### Setup: Insert Test Data
```sql
-- Insert test recommendations
INSERT INTO optimization_recommendations (
    query_hash, source_type, recommendation_text, estimated_improvement_percent,
    confidence_score, urgency_score, roi_score, implementation_complexity
)
VALUES
(4001, 'rewrite', 'Rewrite N+1 query to use IN clause', 95.0, 0.92, 0.5, 0.437, 'medium'),
(4001, 'parameter', 'Set LIMIT = LIMIT 2000', 85.0, 0.95, 0.5, 0.401, 'low'),
(4001, 'parameter', 'Set work_mem = 6MB', 25.0, 0.85, 0.5, 0.106, 'medium')
ON CONFLICT DO NOTHING;
```

### Test 1: Get Top Recommendations
```bash
curl "http://localhost:8080/api/v1/optimization-recommendations?limit=10&min_impact=20" \
  -H "Authorization: Bearer $TOKEN"

# Expected: 3 recommendations sorted by ROI score descending
```

### Test 2: Implement Recommendation
```bash
curl -X POST http://localhost:8080/api/v1/optimization-recommendations/1/implement \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query_hash": 4001, "implementation_notes": "Applied N+1 fix"}'

# Expected: implementation_id, status = 'pending'
```

### Test 3: Get Results (simulated after waiting)
```bash
curl "http://localhost:8080/api/v1/optimization-results?status=pending" \
  -H "Authorization: Bearer $TOKEN"

# Expected: Implementation records with pre_metrics captured
```

---

## Database Level Tests (5 test cases)

### Test 1: Aggregate Recommendations for Query

**Objective**: Verify aggregate_recommendations_for_query() function creates recommendations

**Setup**:
```sql
-- Ensure we have query metrics
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (
    4101,
    'SELECT * FROM orders WHERE customer_id = $1 ORDER BY created_at',
    450.0,
    500,
    10000,
    'testdb'
)
ON CONFLICT (query_hash) DO UPDATE SET
    mean_exec_time_ms = EXCLUDED.mean_exec_time_ms;

-- Ensure we have suggestions for this query
INSERT INTO query_rewrite_suggestions (
    query_hash, suggestion_type, description, original_query, suggested_rewrite,
    reasoning, estimated_improvement_percent, confidence_score
)
VALUES (
    4101,
    'n_plus_one_detected',
    'N+1 pattern detected',
    'SELECT * FROM users WHERE id = $1',
    'SELECT * FROM users WHERE id = ANY($1)',
    'Batch the queries',
    95.0,
    0.92
)
ON CONFLICT (query_hash, suggestion_type) DO NOTHING;

INSERT INTO parameter_tuning_suggestions (
    query_hash, parameter_name, current_value, recommended_value,
    reasoning, estimated_improvement_percent, confidence_score
)
VALUES (
    4101,
    'LIMIT',
    'NOT SET',
    'LIMIT 1000',
    'Large result set',
    85.0,
    0.95
)
ON CONFLICT (query_hash, parameter_name) DO NOTHING;
```

**Execute**:
```sql
SELECT * FROM aggregate_recommendations_for_query(4101);

-- Verify recommendations created
SELECT source_type, COUNT(*), AVG(roi_score), AVG(confidence_score)
FROM optimization_recommendations
WHERE query_hash = 4101
GROUP BY source_type;
```

**Expected Results**:
- ✓ 2 recommendations created (rewrite + parameter)
- ✓ ROI scores calculated correctly
- ✓ Urgency score = frequency_score × impact_score
- ✓ Source types: ['parameter', 'rewrite']

**ROI Calculation Verification**:
```
Frequency = 500 / 100000 = 0.005 (5% of max)
Impact = 450 / 10000 = 0.045 (4.5% of max)
Urgency = 0.005 × 0.045 = 0.000225

N+1 ROI = 0.92 × 0.95 × 0.000225 ≈ 0.000197
LIMIT ROI = 0.95 × 0.85 × 0.000225 ≈ 0.000182
```

---

### Test 2: ROI Score Calculation

**Objective**: Verify ROI scores are calculated correctly

**Setup**:
```sql
-- Different urgency scenarios
INSERT INTO optimization_recommendations (
    query_hash, source_type, recommendation_text, estimated_improvement_percent,
    confidence_score, urgency_score, roi_score, implementation_complexity
)
VALUES
-- High urgency: frequent, slow query
(4201, 'rewrite', 'Fix high-impact query', 80.0, 0.8, 1.0, 0.64, 'medium'),
-- Medium urgency
(4202, 'parameter', 'Tune medium-impact query', 50.0, 0.7, 0.1, 0.035, 'low'),
-- Low urgency: infrequent, fast
(4203, 'parameter', 'Tune low-impact query', 30.0, 0.6, 0.01, 0.0018, 'low')
ON CONFLICT DO NOTHING;
```

**Execute**:
```sql
SELECT query_hash, roi_score, urgency_score, confidence_score, estimated_improvement_percent
FROM optimization_recommendations
WHERE query_hash IN (4201, 4202, 4203)
ORDER BY roi_score DESC;
```

**Expected Results**:
- ✓ Highest ROI: 0.64 (query 4201)
- ✓ Medium ROI: 0.035 (query 4202)
- ✓ Lowest ROI: 0.0018 (query 4203)
- ✓ Correctly ordered by ROI descending

---

### Test 3: Record Implementation

**Objective**: Verify record_recommendation_implementation() captures metrics

**Setup**:
```sql
-- Ensure recommendation exists
INSERT INTO optimization_recommendations (
    query_hash, source_type, recommendation_text, estimated_improvement_percent,
    confidence_score, urgency_score, roi_score, implementation_complexity
)
VALUES (4301, 'rewrite', 'Test recommendation', 80.0, 0.9, 0.5, 0.36, 'medium')
ON CONFLICT DO NOTHING;

-- Ensure query metrics exist
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (4301, 'SELECT * FROM test', 300.0, 100, 1000, 'testdb')
ON CONFLICT (query_hash) DO UPDATE SET mean_exec_time_ms = EXCLUDED.mean_exec_time_ms;
```

**Execute**:
```sql
-- Get recommendation ID
SELECT id FROM optimization_recommendations WHERE query_hash = 4301;

-- Record implementation (using the ID from above)
SELECT impl_id, status, pre_snapshot
FROM record_recommendation_implementation(1, 4301, 'Test implementation notes');

-- Verify implementation record created
SELECT id, status, pre_optimization_stats
FROM optimization_implementations
WHERE query_hash = 4301;
```

**Expected Results**:
- ✓ Implementation record created
- ✓ Status = 'pending'
- ✓ Pre-metrics snapshot captured (JSON with mean_exec_time_ms, calls, rows, etc.)
- ✓ Recommendation marked as is_dismissed = TRUE

---

### Test 4: Measure Implementation Results

**Objective**: Verify measure_implementation_results() calculates improvement

**Setup**:
```sql
-- Create test scenario: pre-implementation already recorded
INSERT INTO optimization_implementations (
    recommendation_id, query_hash, implementation_notes, pre_optimization_stats, status
)
VALUES (
    1,
    4401,
    'Test measurement',
    jsonb_build_object(
        'mean_exec_time_ms', 400.0,
        'calls', 100
    ),
    'pending'
)
RETURNING id;

-- Record query_id for result lookup
-- Note: Save the returned ID for next step

-- Update metrics to simulate post-optimization
INSERT INTO metrics_pg_stats_query (
    query_hash, query_text, mean_exec_time_ms, calls, rows, database_name
)
VALUES (4401, 'SELECT * FROM test', 200.0, 100, 1000, 'testdb')
ON CONFLICT (query_hash) DO UPDATE SET mean_exec_time_ms = EXCLUDED.mean_exec_time_ms;
```

**Execute**:
```sql
-- Measure results (substitute impl_id from setup)
SELECT impl_id, actual_improvement_percent, predicted_improvement_percent,
       status, accuracy_score
FROM measure_implementation_results(2);

-- Verify update
SELECT actual_improvement_percent, post_optimization_stats, status
FROM optimization_implementations
WHERE id = 2;
```

**Expected Results**:
- ✓ actual_improvement_percent = (400 - 200) / 400 × 100 = 50%
- ✓ accuracy_score calculated (difference between actual and predicted)
- ✓ Status updated to 'implemented'
- ✓ Post-metrics captured

---

### Test 5: Get Top Recommendations Ranking

**Objective**: Verify recommendations ranked by ROI correctly

**Setup**:
```sql
-- Clear previous test data
DELETE FROM optimization_recommendations WHERE query_hash >= 4500 AND query_hash < 4600;

-- Insert test recommendations with varying ROI
INSERT INTO optimization_recommendations (
    query_hash, source_type, recommendation_text, estimated_improvement_percent,
    confidence_score, urgency_score, roi_score, implementation_complexity, is_dismissed
)
VALUES
(4501, 'rewrite', 'High ROI recommendation', 80.0, 0.9, 0.5, 0.36, 'medium', FALSE),
(4502, 'parameter', 'Low ROI recommendation', 10.0, 0.6, 0.1, 0.006, 'low', FALSE),
(4503, 'rewrite', 'Medium ROI recommendation', 50.0, 0.8, 0.25, 0.1, 'high', FALSE),
(4504, 'parameter', 'Dismissed recommendation', 70.0, 0.85, 0.4, 0.238, 'low', TRUE)
ON CONFLICT DO NOTHING;
```

**Execute**:
```sql
SELECT query_hash, roi_score, recommendation_text
FROM get_top_recommendations(10, 5.0)
ORDER BY roi_score DESC;
```

**Expected Results**:
- ✓ First: query 4501 (ROI: 0.36)
- ✓ Second: query 4503 (ROI: 0.1)
- ✓ Third: query 4502 (ROI: 0.006)
- ✓ query 4504 excluded (is_dismissed = TRUE)
- ✓ All have estimated_improvement >= 5%

---

## API Level Tests (4 test cases)

### Test 1: Aggregate Recommendations

**Endpoint**: `POST /api/v1/recommendations/aggregate`

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/recommendations/aggregate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query_hash": 4001, "min_confidence": 0.7}'
```

**Expected Response** (200 OK):
```json
{
  "recommendations_aggregated": 2,
  "source_types": ["parameter", "rewrite"],
  "aggregated_at": "2026-02-20T...",
  "message": "Aggregated 2 recommendations"
}
```

**Verification**:
- ✓ Status code 200
- ✓ recommendations_aggregated > 0
- ✓ source_types contains expected values

---

### Test 2: Get Optimization Recommendations

**Endpoint**: `GET /api/v1/optimization-recommendations?limit=10&min_impact=20&source_type=rewrite`

**Request**:
```bash
curl "http://localhost:8080/api/v1/optimization-recommendations?limit=5&min_impact=30&source_type=rewrite" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response** (200 OK):
```json
{
  "recommendations": [
    {
      "id": 201,
      "query_hash": 4001,
      "source_type": "rewrite",
      "recommendation_text": "Rewrite N+1 query",
      "estimated_improvement_percent": 95.0,
      "confidence_score": 0.92,
      "urgency_score": 0.5,
      "roi_score": 0.437
    }
  ],
  "count": 1,
  "total_roi_potential": 0.437,
  "timestamp": "2026-02-20T..."
}
```

**Verification**:
- ✓ Only recommendations with source_type = 'rewrite'
- ✓ Only recommendations with improvement >= 30%
- ✓ Count <= 5 (limit parameter)
- ✓ Sorted by roi_score DESC

---

### Test 3: Implement Recommendation

**Endpoint**: `POST /api/v1/optimization-recommendations/{id}/implement`

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/optimization-recommendations/201/implement \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query_hash": 4001,
    "implementation_notes": "Applied N+1 fix using IN clause"
  }'
```

**Expected Response** (200 OK):
```json
{
  "implementation_id": 501,
  "recommendation_id": 201,
  "query_hash": 4001,
  "status": "pending",
  "timestamp": "2026-02-20T...",
  "message": "Implementation recorded. Check results after 24-48 hours."
}
```

**Verification**:
- ✓ Status code 200
- ✓ implementation_id returned
- ✓ Status = 'pending'
- ✓ Recommendation marked as dismissed

---

### Test 4: Get Optimization Results

**Endpoint**: `GET /api/v1/optimization-results?status=implemented&limit=10`

**Request**:
```bash
curl "http://localhost:8080/api/v1/optimization-results?status=implemented&limit=5" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response** (200 OK):
```json
{
  "results": [
    {
      "implementation_id": 501,
      "recommendation_id": 201,
      "query_hash": 4001,
      "implementation_notes": "Applied N+1 fix",
      "pre_metrics": {"mean_exec_time_ms": 450.0, "calls": 500},
      "post_metrics": {"mean_exec_time_ms": 225.0},
      "actual_improvement_percent": 50.0,
      "status": "implemented",
      "implemented_at": "2026-02-20T...",
      "measured_at": "2026-02-21T..."
    }
  ],
  "count": 1,
  "total_actual_improvement": 50.0,
  "timestamp": "2026-02-20T..."
}
```

---

## Integration Tests (3 test cases)

### Test 1: Full Workflow - Suggest → Implement → Measure

**Objective**: Complete end-to-end optimization workflow

**Steps**:
```bash
# Step 1: Ensure recommendations exist
# (Insert test data from database tests)

# Step 2: Get top recommendations
curl "http://localhost:8080/api/v1/optimization-recommendations?limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq '.recommendations[0].id'
# Save the ID as REC_ID

# Step 3: Implement recommendation
IMPL=$(curl -X POST http://localhost:8080/api/v1/optimization-recommendations/$REC_ID/implement \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query_hash": 4001}' | jq '.implementation_id')

# Step 4: Simulate 48 hours passing (in practice)
# Update query metrics to show improvement
psql -c "
UPDATE metrics_pg_stats_query SET mean_exec_time_ms = 200
WHERE query_hash = 4001;
"

# Step 5: Measure results
curl "http://localhost:8080/api/v1/optimization-results?status=implemented" \
  -H "Authorization: Bearer $TOKEN" | jq '.results[0]'

# Expected: actual_improvement_percent should show measured improvement
```

---

### Test 2: ROI Scoring with Real Metrics

**Objective**: Verify ROI scores reflect actual query impact

**Scenario**: Compare recommendations for different query frequencies

```bash
# Query A: High frequency, medium execution time
# Expected urgency: HIGH
# Expected ROI ranking: HIGH (even if improvement is moderate)

# Query B: Low frequency, high execution time
# Expected urgency: MEDIUM
# Expected ROI ranking: LOWER (despite high improvement %)

# Query C: High frequency, short execution time
# Expected urgency: LOW
# Expected ROI ranking: LOWEST (despite high improvement %)
```

---

### Test 3: Recommendation Filtering and Pagination

**Objective**: Verify filtering and pagination parameters work correctly

```bash
# Test 1: Filter by source type
curl "http://localhost:8080/api/v1/optimization-recommendations?source_type=parameter" \
  -H "Authorization: Bearer $TOKEN" | jq '.recommendations[] | .source_type' | sort -u
# Expected: Only "parameter"

# Test 2: Filter by minimum impact
curl "http://localhost:8080/api/v1/optimization-recommendations?min_impact=50" \
  -H "Authorization: Bearer $TOKEN" | jq '.recommendations[] | .estimated_improvement_percent' | sort -n
# Expected: All >= 50.0

# Test 3: Pagination with limit
curl "http://localhost:8080/api/v1/optimization-recommendations?limit=3" \
  -H "Authorization: Bearer $TOKEN" | jq '.count'
# Expected: 3

# Test 4: Combined filters
curl "http://localhost:8080/api/v1/optimization-recommendations?source_type=rewrite&min_impact=80&limit=5" \
  -H "Authorization: Bearer $TOKEN" | jq '.count'
# Expected: <= 5, all rewrite, all >= 80%
```

---

## Error Handling Tests (3 test cases)

### Test 1: Invalid Recommendation ID

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/optimization-recommendations/0/implement \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query_hash": 4001}'
```

**Expected Response** (400 Bad Request):
```json
{
  "error": "Invalid recommendation_id format"
}
```

---

### Test 2: Missing Required Query Hash

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/optimization-recommendations/201/implement \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Expected Response** (400 Bad Request):
```json
{
  "error": "query_hash is required"
}
```

---

### Test 3: No Recommendations Available

**Request**:
```bash
curl "http://localhost:8080/api/v1/optimization-recommendations?min_impact=99999" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response** (200 OK with empty results):
```json
{
  "recommendations": [],
  "count": 0,
  "total_roi_potential": 0.0,
  "message": "No recommendations available"
}
```

---

## Success Criteria Checklist

### Database Tests
- [ ] Test 1: Aggregate function creates recommendations
- [ ] Test 2: ROI scores calculated correctly
- [ ] Test 3: Implementation records pre-metrics
- [ ] Test 4: Results measurement works
- [ ] Test 5: Ranking by ROI verified

### API Tests
- [ ] Test 1: Aggregate endpoint works
- [ ] Test 2: Get recommendations with filters
- [ ] Test 3: Implement endpoint records changes
- [ ] Test 4: Results endpoint shows before/after

### Integration Tests
- [ ] Test 1: Full workflow functional
- [ ] Test 2: ROI reflects actual impact
- [ ] Test 3: Filtering and pagination work

### Error Handling
- [ ] Invalid recommendation_id rejected
- [ ] Missing query_hash validation
- [ ] Empty results handled gracefully

---

**Total Test Cases**: 12 (5 database + 4 API + 3 integration)
**Estimated Duration**: 3-4 hours for complete test suite
**Critical Path**: DB Tests 1-3 → API Tests 1-2 → Integration Test 1

