# Phase 4.5.4: ML-Powered Optimization Workflow - Implementation Specification

**Date**: February 20, 2026
**Status**: Implementation In Progress
**Duration**: 3-4 days estimated
**Lines of Code**: 400+ (SQL + Go enhancements)

---

## What Will Be Implemented

### Overview

Phase 4.5.4 aggregates all optimization suggestions from Phases 4.5.1-4.5.3 into a unified workflow that:

1. **Aggregates** all recommendation sources (workload patterns, rewrite suggestions, parameter tuning)
2. **Ranks** recommendations by ROI (confidence × improvement × urgency)
3. **Tracks** when recommendations are implemented
4. **Measures** actual improvements vs. predicted improvements
5. **Learns** from implementation results to refine future recommendations

---

## System Architecture

### Recommendation Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ Suggestion Sources (Phase 4.5.1-4.5.3)                          │
├─────────────────────────────────────────────────────────────────┤
│ • Workload Patterns (patterns from Phase 4.5.1)                 │
│ • Query Rewrite Suggestions (5 types from Phase 4.5.2)          │
│ • Parameter Tuning (4 types from Phase 4.5.3)                   │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ Recommendation Aggregation & Ranking (Phase 4.5.4)              │
├─────────────────────────────────────────────────────────────────┤
│ 1. Collect all suggestions from source tables                   │
│ 2. Calculate ROI Score: confidence × improvement × urgency      │
│ 3. Rank by ROI (highest first)                                  │
│ 4. Store in optimization_recommendations table                  │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ User Views & Implements Recommendations                          │
├─────────────────────────────────────────────────────────────────┤
│ • GET /api/v1/optimization-recommendations (ranked by ROI)      │
│ • POST /api/v1/optimization-recommendations/{id}/implement      │
│ • Captures pre-optimization metrics snapshot                    │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│ Result Measurement & Learning Loop                              │
├─────────────────────────────────────────────────────────────────┤
│ • Wait 24-48 hours for post-implementation data                │
│ • Capture post-optimization metrics snapshot                   │
│ • Compare actual vs. predicted improvement                      │
│ • Update model confidence based on accuracy                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Core Concepts

### Recommendation Aggregation

**Purpose**: Collect all optimization opportunities from different sources into single ranked list

**Sources**:
1. **Workload Patterns** (Phase 4.5.1)
   - Pattern type: hourly_peak, daily_cycle, weekly_pattern, batch_job
   - Source: workload_patterns table
   - Recommendation: "Scale resources at peak hour {hour}"

2. **Query Rewrite Suggestions** (Phase 4.5.2)
   - Types: n_plus_one, inefficient_join, missing_index, subquery, in_vs_any
   - Source: query_rewrite_suggestions table
   - Recommendation: "Rewrite query to [suggested_rewrite]"

3. **Parameter Tuning** (Phase 4.5.3)
   - Types: LIMIT, work_mem, sort_mem, batch_size
   - Source: parameter_tuning_suggestions table
   - Recommendation: "Set {parameter_name} = {recommended_value}"

### ROI Scoring Algorithm

**ROI Score** = Confidence × Improvement × Urgency

Where:
- **Confidence**: 0-1 (from source suggestion)
- **Improvement**: Normalized to 0-1 (improvement_percent / 100, capped at 1.0)
- **Urgency**: 0-1 (calculated from query frequency and current impact)

**Urgency Calculation**:
```
Urgency = Frequency Score × Impact Score

Frequency Score = MIN(query_calls / 100000, 1.0)
  - Queries called many times → high frequency score
  - Queries called few times → low frequency score

Impact Score = MIN(mean_exec_time_ms / 10000, 1.0)
  - Slow queries → high impact score
  - Fast queries → low impact score

Examples:
- Query called 100,000 times, 10 second avg: urgency = 1.0 × 1.0 = 1.0
- Query called 1,000 times, 100ms avg: urgency = 0.01 × 0.01 = 0.0001
- Query called 10,000 times, 1 second avg: urgency = 0.1 × 0.1 = 0.01
```

**ROI Score Examples**:
```
1. N+1 Query Pattern
   - Confidence: 0.92 (high - clear pattern)
   - Improvement: 0.95 (95% reduction)
   - Urgency: 0.5 (called 500 times, 50ms avg)
   → ROI = 0.92 × 0.95 × 0.5 = 0.437 (HIGH PRIORITY)

2. Missing Index Recommendation
   - Confidence: 0.83
   - Improvement: 0.80 (80% reduction)
   - Urgency: 0.25 (called 5000 times, 250ms avg)
   → ROI = 0.83 × 0.80 × 0.25 = 0.166 (MEDIUM PRIORITY)

3. IN vs ANY Optimization
   - Confidence: 0.65
   - Improvement: 0.15 (only 15% reduction)
   - Urgency: 0.1 (called 1000 times, 100ms avg)
   → ROI = 0.65 × 0.15 × 0.1 = 0.0098 (LOW PRIORITY)
```

### Implementation Tracking

**Purpose**: Record when recommendations are implemented and measure results

**Process**:
1. User marks recommendation as "implemented"
2. System captures pre-implementation metrics snapshot
3. Wait 24-48 hours for post-implementation data to accumulate
4. Capture post-implementation metrics snapshot
5. Calculate actual improvement: (pre_time - post_time) / pre_time × 100%
6. Compare with predicted improvement
7. Update model confidence based on accuracy

**Metrics Captured**:
```json
Pre-optimization:
{
  "mean_exec_time_ms": 450.0,
  "calls_per_hour": 200,
  "total_time_ms": 90000,
  "p95_exec_time_ms": 520,
  "p99_exec_time_ms": 600
}

Post-optimization (24-48h later):
{
  "mean_exec_time_ms": 250.0,
  "calls_per_hour": 198,
  "total_time_ms": 49500,
  "p95_exec_time_ms": 320,
  "p99_exec_time_ms": 380
}

Actual Improvement:
(450 - 250) / 450 = 44% (vs predicted 35%)
```

---

## Database Schema Enhancements

### Table: optimization_recommendations (Already exists, will be populated)

**New columns to populate**:
- source_type: VARCHAR(50) - 'workload_pattern', 'rewrite', 'parameter'
- source_id: BIGINT - reference to source table
- roi_score: FLOAT - computed ROI value
- urgency_score: FLOAT - frequency × impact
- is_dismissed: BOOLEAN - user dismissed this recommendation

### Function: aggregate_recommendations_for_query()

**Purpose**: Aggregate all suggestions for a query into optimization_recommendations table

**Signature**:
```sql
CREATE OR REPLACE FUNCTION aggregate_recommendations_for_query(
    p_query_hash BIGINT
) RETURNS TABLE (
    recommendation_count INT,
    source_types TEXT[]
) AS $$
```

**Implementation**:
1. Get all rewrite suggestions for query
2. Get all parameter tuning suggestions for query
3. Get workload patterns affecting query
4. For each suggestion:
   - Calculate urgency_score from query metrics
   - Calculate ROI_score = confidence × improvement × urgency
   - Insert into optimization_recommendations
5. Return count and source types

### Function: get_top_recommendations_for_query()

**Purpose**: Return top recommendations ranked by ROI for a specific query

**Signature**:
```sql
CREATE OR REPLACE FUNCTION get_top_recommendations_for_query(
    p_query_hash BIGINT,
    p_limit INTEGER DEFAULT 10
) RETURNS TABLE (
    rec_id BIGINT,
    source_type VARCHAR,
    recommendation_text TEXT,
    confidence FLOAT,
    improvement_percent FLOAT,
    urgency_score FLOAT,
    roi_score FLOAT
) AS $$
```

### Function: record_recommendation_implementation()

**Purpose**: Record that a recommendation was implemented

**Signature**:
```sql
CREATE OR REPLACE FUNCTION record_recommendation_implementation(
    p_recommendation_id BIGINT,
    p_query_hash BIGINT,
    p_implementation_notes TEXT
) RETURNS TABLE (
    impl_id BIGINT,
    status VARCHAR,
    pre_snapshot JSONB
) AS $$
```

**Process**:
1. Validate recommendation_id exists
2. Get query's current metrics as pre-optimization snapshot
3. Create optimization_implementations record with status = 'pending'
4. Mark recommendation as is_dismissed = TRUE (don't show again)
5. Return implementation_id and current metrics

### Function: measure_implementation_results()

**Purpose**: Measure actual improvement after implementation

**Signature**:
```sql
CREATE OR REPLACE FUNCTION measure_implementation_results(
    p_implementation_id BIGINT
) RETURNS TABLE (
    impl_id BIGINT,
    actual_improvement_percent FLOAT,
    predicted_improvement_percent FLOAT,
    status VARCHAR,
    accuracy_score FLOAT
) AS $$
```

**Process**:
1. Fetch implementation record (should be 24-48h old)
2. Get current metrics as post-optimization snapshot
3. Calculate actual_improvement = (pre_mean_time - post_mean_time) / pre_mean_time × 100%
4. Get predicted improvement from associated recommendation
5. Calculate accuracy_score = 1 - ABS(actual - predicted) / predicted
6. Update recommendation confidence based on accuracy
7. Return results

---

## Go Implementation

### Storage Methods

#### AggregateRecommendationsForQuery()
```go
func (p *PostgresDB) AggregateRecommendationsForQuery(
    ctx context.Context,
    queryHash int64,
) (int, []string, error)
```

**Purpose**: Aggregate all suggestions for a query into recommendations table

**Implementation**:
- Calls SQL function aggregate_recommendations_for_query()
- Returns count and source types
- Error handling for database errors

#### GetOptimizationRecommendations()
```go
func (p *PostgresDB) GetOptimizationRecommendations(
    ctx context.Context,
    limit int,
    minImpact float64,
    sourceType string,
) ([]models.OptimizationRecommendation, error)
```

**Purpose**: Get top recommendations across all queries, ranked by ROI

**Implementation**:
- Query optimization_recommendations table
- Filter by source_type if provided
- Filter by min_impact_percent if provided
- Order by roi_score DESC, confidence DESC
- Limit results
- Return array of recommendations

#### ImplementRecommendation()
```go
func (p *PostgresDB) ImplementRecommendation(
    ctx context.Context,
    recommendationID int64,
    queryHash int64,
    notes string,
) (*models.OptimizationImplementation, error)
```

**Purpose**: Record that a recommendation was implemented

**Implementation**:
- Call SQL function record_recommendation_implementation()
- Return implementation record with pre-metrics

#### MeasureImplementationResults()
```go
func (p *PostgresDB) MeasureImplementationResults(
    ctx context.Context,
    implementationID int64,
) (*models.OptimizationResult, error)
```

**Purpose**: Measure actual improvement after implementation

**Implementation**:
- Call SQL function measure_implementation_results()
- Return actual vs. predicted improvement
- Accuracy score for model learning

---

## API Handlers

### Handler: handleAggregateRecommendations()
**Endpoint**: `POST /api/v1/recommendations/aggregate`

**Purpose**: Trigger aggregation of all suggestions into recommendations table

**Implementation**:
```
1. Parse optional query_hash from body
2. If query_hash provided: aggregate for single query
3. If not: aggregate for all queries with suggestions
4. Return count and source types
```

**Request**:
```json
{
  "query_hash": 4001,
  "min_confidence": 0.7
}
```

**Response** (200 OK):
```json
{
  "recommendations_aggregated": 15,
  "source_types": ["rewrite", "parameter", "workload_pattern"],
  "aggregated_at": "2026-02-20T...",
  "message": "Aggregated 15 recommendations"
}
```

### Handler: handleGetOptimizationRecommendations()
**Endpoint**: `GET /api/v1/optimization-recommendations?limit=20&min_impact=5&source_type=rewrite`

**Purpose**: Get top recommendations ranked by ROI

**Implementation**:
```
1. Parse query parameters:
   - limit: 1-100 (default 20)
   - min_impact: 0-100 (default 5.0)
   - source_type: optional filter
2. Call storage method to fetch recommendations
3. Return sorted array with ROI scores
```

**Response** (200 OK):
```json
{
  "recommendations": [
    {
      "id": 201,
      "query_hash": 4001,
      "source_type": "rewrite",
      "recommendation_text": "Rewrite N+1 query to use IN clause",
      "estimated_improvement_percent": 95.0,
      "confidence_score": 0.92,
      "urgency_score": 0.5,
      "roi_score": 0.437,
      "created_at": "2026-02-20T..."
    }
  ],
  "count": 15,
  "total_roi_potential": 5.2,
  "timestamp": "2026-02-20T..."
}
```

### Handler: handleImplementRecommendation()
**Endpoint**: `POST /api/v1/optimization-recommendations/{id}/implement`

**Purpose**: Record that recommendation was implemented

**Implementation**:
```
1. Parse recommendation_id from URL
2. Parse optional body:
   - query_hash: required
   - implementation_notes: optional
3. Call storage method to create implementation record
4. Return implementation_id and status
```

**Request**:
```json
{
  "query_hash": 4001,
  "implementation_notes": "Implemented N+1 fix using IN clause batching"
}
```

**Response** (200 OK):
```json
{
  "implementation_id": 501,
  "recommendation_id": 201,
  "query_hash": 4001,
  "status": "pending",
  "pre_metrics": {
    "mean_exec_time_ms": 450.0,
    "calls_per_hour": 200,
    "total_time_ms": 90000
  },
  "timestamp": "2026-02-20T...",
  "message": "Implementation recorded. Check results after 24-48 hours."
}
```

### Handler: handleGetOptimizationResults()
**Endpoint**: `GET /api/v1/optimization-results?recommendation_id=201&status=completed`

**Purpose**: Get results from implemented recommendations

**Implementation**:
```
1. Parse query parameters:
   - recommendation_id: optional filter
   - status: optional (pending, implemented, reverted)
   - limit: 1-100 (default 20)
2. Query optimization_implementations table
3. Include pre/post metrics
4. Calculate actual_improvement_percent
```

**Response** (200 OK):
```json
{
  "results": [
    {
      "implementation_id": 501,
      "recommendation_id": 201,
      "query_hash": 4001,
      "status": "implemented",
      "actual_improvement_percent": 44.0,
      "predicted_improvement_percent": 95.0,
      "accuracy_score": 0.54,
      "pre_metrics": {...},
      "post_metrics": {...},
      "measured_at": "2026-02-21T..."
    }
  ],
  "count": 3,
  "total_actual_improvement": 85.5,
  "timestamp": "2026-02-20T..."
}
```

---

## Recommendation Workflow Example

### Scenario: High-Frequency Slow Query

**Query**: `SELECT * FROM orders WHERE customer_id = $1 ORDER BY created_at`
- Called 500 times/hour
- Mean execution: 450ms
- Returns 10,000 rows

**Step 1: Suggestions Generated (Phases 4.5.1-4.5.3)**

Workload Pattern (Phase 4.5.1):
- Peak load at 8 AM, confidence 0.92

Query Rewrite Suggestions (Phase 4.5.2):
- N+1 pattern: confidence 0.92, improvement 95%
- Missing index: confidence 0.83, improvement 80%

Parameter Tuning (Phase 4.5.3):
- LIMIT clause: confidence 0.95, improvement 85%
- work_mem: confidence 0.85, improvement 25%

**Step 2: Recommendations Aggregated & Ranked**

Calculate urgency:
- Frequency score = 500000 / 100000 = 1.0 (capped)
- Impact score = 450 / 10000 = 0.045
- Urgency = 1.0 × 0.045 = 0.045

ROI Scores:
1. N+1 rewrite: 0.92 × 0.95 × 0.045 = **0.0393** (HIGH)
2. LIMIT clause: 0.95 × 0.85 × 0.045 = **0.0363** (HIGH)
3. Missing index: 0.83 × 0.80 × 0.045 = **0.0298** (MEDIUM)
4. work_mem: 0.85 × 0.25 × 0.045 = **0.0095** (LOW)
5. Pattern scaling: 0.92 × 0.10 × 0.045 = **0.0041** (VERY LOW)

Ranked:
1. "Rewrite N+1 query" - ROI: 0.0393
2. "Add LIMIT clause" - ROI: 0.0363
3. "Create index on customer_id" - ROI: 0.0298

**Step 3: User Implements Recommendation**

User clicks "Implement" on "Rewrite N+1 query"
- Pre-metrics captured: mean_exec_time = 450ms, calls = 500/h
- Implementation recorded with status = 'pending'
- Recommendation dismissed (not shown again)

**Step 4: Results Measured (24-48h later)**

Post-metrics captured: mean_exec_time = 290ms, calls = 498/h
- Actual improvement = (450 - 290) / 450 = 35.6%
- Predicted improvement = 95%
- Accuracy = 1 - |35.6 - 95| / 95 = 0.626 (62.6%)

**Step 5: Learning Loop**

- Update N+1 detection confidence based on actual results
- If consistently underestimated: lower confidence for future N+1 recommendations
- If consistently overestimated: increase confidence (predictions were too conservative)

---

## API Endpoints Summary

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | /api/v1/recommendations/aggregate | Aggregate all suggestions into recommendations |
| GET | /api/v1/optimization-recommendations | Get top recommendations ranked by ROI |
| POST | /api/v1/optimization-recommendations/{id}/implement | Record implementation |
| GET | /api/v1/optimization-results | Get measurement results |

---

## SQL Implementation Details

### aggregate_recommendations_for_query() Function

**Key Logic**:
```sql
1. FOR EACH rewrite suggestion:
   - Urgency = freq_score × impact_score
   - ROI = confidence × improvement × urgency
   - INSERT INTO optimization_recommendations

2. FOR EACH parameter suggestion:
   - Urgency = same calculation
   - ROI = confidence × (improvement / 100) × urgency
   - INSERT INTO optimization_recommendations

3. FOR EACH workload pattern:
   - Estimate queries affected by pattern
   - Urgency = pattern_confidence × impact_multiplier
   - ROI = confidence × 0.1 × urgency (patterns are preventive)
   - INSERT INTO optimization_recommendations

4. Return total count and distinct source_types
```

---

## Success Criteria

✅ **Criteria 1**: Recommendations aggregated from all sources
- All rewrite suggestions aggregated
- All parameter suggestions aggregated
- Workload patterns converted to recommendations

✅ **Criteria 2**: ROI scoring working correctly
- ROI = confidence × improvement × urgency
- Urgency calculated from frequency × impact
- ROI scores in reasonable range (0-1)

✅ **Criteria 3**: Recommendations ranked by ROI
- Top recommendations have highest ROI scores
- API returns sorted by ROI descending

✅ **Criteria 4**: Implementation tracking functional
- Recommendation can be marked as implemented
- Pre-optimization metrics captured
- Implementation status tracked

✅ **Criteria 5**: Results measurement working
- Post-optimization metrics captured
- Actual improvement calculated correctly
- Accuracy score computed

✅ **Criteria 6**: All API endpoints functional
- Aggregation endpoint returns correct counts
- Get recommendations returns sorted array
- Implement endpoint records changes
- Results endpoint shows before/after

✅ **Criteria 7**: Error handling complete
- Invalid recommendation_id returns 400
- Non-existent query_hash handled gracefully
- Database errors logged and reported

✅ **Criteria 8**: Dashboard integration ready
- All data available for visualization
- ROI scores support sorting/filtering
- Implementation tracking trackable

---

## Integration Points

### With Phase 4.5.1 (Workload Patterns)
- Patterns converted to preventive recommendations
- Pattern detection triggers aggregation

### With Phase 4.5.2 (Query Rewrite Suggestions)
- Rewrite suggestions aggregated with full confidence scores
- Source tracking (from query_rewrite_suggestions)

### With Phase 4.5.3 (Parameter Tuning)
- Parameter suggestions aggregated
- Confidence and improvement scores included

### With Phase 4.5.5 (ML Service)
- ML service provides prediction accuracy feedback
- Learning loop refines model confidence

### With Grafana
- Dashboard shows top recommendations
- Tracks implementation progress
- Measures actual improvements

---

## Known Considerations

1. **Recommendation Deduplication**: If same fix suggested multiple ways, rank highest ROI first
2. **Cascading Recommendations**: Some fixes depend on others (e.g., add index before rewrite)
3. **Measurement Delay**: Need 24-48h window to capture post-optimization metrics
4. **Query Volume Changes**: Urgency calculation may vary if query volume changes significantly
5. **Multiple Fixes**: Actual improvement may be less than sum of individual improvements (diminishing returns)

---

**Status**: Ready for implementation

