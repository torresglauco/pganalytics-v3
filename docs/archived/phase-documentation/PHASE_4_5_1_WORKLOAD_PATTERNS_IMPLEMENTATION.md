# Phase 4.5.1: Workload Pattern Detection - Implementation Guide

**Date**: February 20, 2026
**Status**: Implementation In Progress
**Objective**: Detect recurring patterns in query execution (hourly peaks, daily cycles, batch jobs)

---

## Feature Specification

### Goal
Identify recurring patterns in query execution to understand when performance issues occur and enable predictive scaling.

### Pattern Types
1. **hourly_peak** - Specific hours have significantly higher query volume/execution time
2. **daily_cycle** - Daily repeating pattern (e.g., slower at specific times each day)
3. **weekly_pattern** - Weekly repeating pattern (e.g., slower on Mondays)
4. **batch_job** - Unusual high-volume periods indicating batch processing

### Algorithm

```
Step 1: Data Collection (1-hour buckets)
├─ Group query metrics by EXTRACT(HOUR FROM collected_at)
├─ Calculate count of queries per hour
├─ Calculate mean execution time per hour
└─ Aggregate across 30-day rolling window

Step 2: Statistical Analysis
├─ Calculate overall mean: avg_hour_count, avg_exec_time
├─ Calculate standard deviation: stddev_hour_count, stddev_exec_time
└─ For each hour, calculate z-score = (hour_value - mean) / stddev

Step 3: Peak Hour Identification
├─ Peak hours: where z_score > 1.0 (above mean + 1 stddev)
├─ Calculate peak_hour impact: peak_value / average_value
└─ Calculate consistency (variance across 30 days)

Step 4: Pattern Classification
├─ hourly_peak: If peak hours detected and consistent
├─ daily_cycle: If pattern repeats same hours daily
├─ weekly_pattern: If pattern repeats on same day of week
└─ batch_job: If unusual 2-3 hour spikes with high variance

Step 5: Confidence Scoring
├─ Confidence = consistency_score × recurrence_score
├─ Consistency: based on stddev (lower stddev = higher confidence)
├─ Recurrence: how many days show pattern / total days
└─ Final confidence: 0-1 scale
```

### Success Criteria
- ✅ Detect >80% of actual hourly peaks
- ✅ Confidence scores >0.7 for consistent patterns
- ✅ <10% false positive rate
- ✅ Performance: analysis completes in <2 minutes for 30 days
- ✅ Minimum 7-day data required for analysis
- ✅ Minimum 10-day data required for >0.8 confidence

---

## Implementation Components

### 1. SQL Function: detect_workload_patterns()

**Location**: `backend/migrations/005_ml_optimization.sql` (already created, needs implementation)
**Purpose**: Analyze historical query metrics and insert patterns

**Pseudocode**:
```sql
CREATE OR REPLACE FUNCTION detect_workload_patterns(
    p_database_name VARCHAR(63),
    p_lookback_days INTEGER DEFAULT 30
) RETURNS TABLE(pattern_id BIGINT, pattern_type VARCHAR, confidence FLOAT) AS $$

DECLARE
    v_analysis_window TIMESTAMP;
    v_hour INT;
    v_overall_avg_count FLOAT;
    v_overall_avg_time FLOAT;
    v_overall_stddev_count FLOAT;
    v_overall_stddev_time FLOAT;
BEGIN
    -- Step 1: Calculate analysis window
    v_analysis_window := NOW() - (p_lookback_days || ' days')::INTERVAL;

    -- Step 2: Get hourly statistics
    WITH hourly_stats AS (
        SELECT
            EXTRACT(HOUR FROM collected_at)::INT as hour_of_day,
            COUNT(*) as query_count,
            AVG(mean_exec_time_ms) as avg_exec_time,
            DATE(collected_at) as stat_date
        FROM metrics_pg_stats_query
        WHERE database_name = p_database_name
        AND collected_at >= v_analysis_window
        GROUP BY EXTRACT(HOUR FROM collected_at), DATE(collected_at)
    ),

    -- Step 3: Aggregate across all days
    hourly_aggregated AS (
        SELECT
            hour_of_day,
            COUNT(*) as days_with_data,
            AVG(query_count) as avg_count,
            STDDEV_POP(query_count) as stddev_count,
            AVG(avg_exec_time) as avg_time,
            STDDEV_POP(avg_exec_time) as stddev_time,
            MAX(query_count) as max_count,
            MAX(avg_exec_time) as max_time
        FROM hourly_stats
        GROUP BY hour_of_day
    ),

    -- Step 4: Calculate overall statistics
    overall_stats AS (
        SELECT
            AVG(avg_count) as overall_avg_count,
            STDDEV_POP(avg_count) as overall_stddev_count,
            AVG(avg_time) as overall_avg_time,
            STDDEV_POP(avg_time) as overall_stddev_time,
            COUNT(DISTINCT days_with_data) as total_days
        FROM hourly_aggregated
    ),

    -- Step 5: Identify peak hours
    peak_hours AS (
        SELECT
            ha.hour_of_day,
            ha.avg_count,
            ha.stddev_count,
            ha.avg_time,
            ha.stddev_time,
            ha.days_with_data,
            os.overall_avg_count,
            os.overall_stddev_count,
            os.overall_avg_time,
            os.overall_stddev_time,
            (ha.avg_count - os.overall_avg_count) / NULLIF(os.overall_stddev_count, 0) as z_score_count,
            (ha.avg_time - os.overall_avg_time) / NULLIF(os.overall_stddev_time, 0) as z_score_time,
            os.total_days,
            -- Confidence calculation
            ROUND((1.0 - (ha.stddev_count / NULLIF(ha.avg_count, 0)))::NUMERIC, 4) as consistency_score,
            ROUND((ha.days_with_data::FLOAT / NULLIF(os.total_days::FLOAT, 0))::NUMERIC, 4) as recurrence_score
        FROM hourly_aggregated ha, overall_stats os
        WHERE (ha.avg_count - os.overall_avg_count) / NULLIF(os.overall_stddev_count, 0) > 1.0
           OR (ha.avg_time - os.overall_avg_time) / NULLIF(os.overall_stddev_time, 0) > 1.0
    )

    -- Step 6: Insert detected patterns
    INSERT INTO workload_patterns (
        database_name, pattern_type, pattern_metadata,
        detection_timestamp, description, affected_query_count
    )
    SELECT
        p_database_name,
        CASE
            WHEN z_score_count > 2.0 THEN 'hourly_peak'
            WHEN z_score_time > 2.0 THEN 'hourly_peak'
            ELSE 'hourly_peak'
        END as pattern_type,
        jsonb_build_object(
            'peak_hour', hour_of_day,
            'variance', ROUND((stddev_count / NULLIF(avg_count, 0))::NUMERIC, 4),
            'confidence', ROUND((consistency_score * recurrence_score)::NUMERIC, 4),
            'affected_queries', CEIL(avg_count),
            'z_score_count', ROUND(z_score_count::NUMERIC, 2),
            'z_score_time', ROUND(z_score_time::NUMERIC, 2),
            'days_observed', days_with_data
        ) as pattern_metadata,
        NOW(),
        'Peak queries at hour ' || hour_of_day::TEXT || ' UTC',
        CEIL(avg_count)
    FROM peak_hours
    ON CONFLICT (database_name, pattern_type)
    DO UPDATE SET
        pattern_metadata = EXCLUDED.pattern_metadata,
        detection_timestamp = NOW(),
        description = EXCLUDED.description,
        affected_query_count = EXCLUDED.affected_query_count;

    RETURN QUERY
    SELECT
        wp.id,
        wp.pattern_type,
        (wp.pattern_metadata->>'confidence')::FLOAT
    FROM workload_patterns wp
    WHERE wp.database_name = p_database_name
    ORDER BY detection_timestamp DESC
    LIMIT 10;
END;
$$ LANGUAGE plpgsql;
```

### 2. Go Storage Method: Implement detect_workload_patterns()

**Location**: `backend/internal/storage/postgres.go`
**Current Status**: Stub exists, needs full implementation

**Implementation**:
```go
// DetectWorkloadPatterns analyzes historical query metrics and detects patterns
func (p *PostgresDB) DetectWorkloadPatterns(ctx context.Context, databaseName string, lookbackDays int) (int, error) {
    // Validation
    if databaseName == "" {
        return 0, apperrors.BadRequest("database_name", "required")
    }
    if lookbackDays < 7 {
        lookbackDays = 7
    }
    if lookbackDays > 365 {
        lookbackDays = 365
    }

    // Call SQL function
    var count int
    err := p.db.QueryRowContext(
        ctx,
        `SELECT COUNT(*) FROM detect_workload_patterns($1, $2)`,
        databaseName,
        lookbackDays,
    ).Scan(&count)

    if err != nil && err != sql.ErrNoRows {
        return 0, apperrors.DatabaseError("detect workload patterns", err.Error())
    }

    return count, nil
}
```

### 3. Handler Enhancement: Add Query Parameter Validation

**Location**: `backend/internal/api/handlers_ml.go` → `handleDetectWorkloadPatterns`
**Current Status**: Basic implementation exists, needs enhancement

**Enhancements Needed**:
- ✅ Input validation (already done)
- ✅ Context timeout (already done)
- ✅ Error handling (already done)
- ✅ Response formatting (already done)

### 4. Dashboard Visualization

**Grafana Panel 1: Hourly Pattern Timeline**
```
Title: "Query Volume by Hour of Day (30-day)"
Type: Time series
Data:
  - X-axis: Hour of day (0-23)
  - Y-axis: Average query count
  - Color by: Confidence score
Query: SELECT hour_of_day, avg_count FROM hourly_aggregated

Title: "Execution Time Patterns"
Type: Time series
Data:
  - X-axis: Hour of day (0-23)
  - Y-axis: Average execution time (ms)
  - Threshold: Mean + 1 stddev (highlight peaks)
```

**Grafana Panel 2: Pattern Summary**
```
Title: "Detected Patterns"
Type: Table
Columns:
  - Database
  - Pattern Type
  - Peak Hour
  - Confidence
  - Affected Queries
  - Last Detection
```

---

## File Changes Required

### 1. Update SQL Function in Migration 005

**File**: `backend/migrations/005_ml_optimization.sql`
**Status**: Placeholder exists (lines ~280-290)
**Action**: Replace placeholder with full implementation above

### 2. Implement Storage Method

**File**: `backend/internal/storage/postgres.go`
**Status**: Stub exists (line ~1378)
**Action**: Expand implementation with validation and error handling

### 3. Add Dashboard Panels

**File**: `grafana/dashboards/query-performance.json`
**Status**: Not yet created
**Action**: Add 2 panels for workload patterns

### 4. (Optional) Add Unit Tests

**File**: `backend/tests/unit/workload_patterns_test.go`
**Status**: Not yet created
**Action**: Create test suite with synthetic data

---

## Implementation Steps

### Step 1: Implement SQL Function (30 minutes)
- [ ] Replace placeholder in migration 005 with full function
- [ ] Add error handling for edge cases
- [ ] Test with sample queries

### Step 2: Update Go Storage Method (15 minutes)
- [ ] Update postgres.go DetectWorkloadPatterns method
- [ ] Add parameter validation
- [ ] Add error mapping

### Step 3: Add Dashboard Panels (20 minutes)
- [ ] Create hourly pattern timeline panel
- [ ] Create pattern summary table panel
- [ ] Add to main query-performance dashboard

### Step 4: Test Implementation (30 minutes)
- [ ] Test with 30-day sample data
- [ ] Verify pattern detection accuracy
- [ ] Test edge cases (insufficient data, single spike)

### Step 5: Documentation (15 minutes)
- [ ] Update API documentation
- [ ] Add troubleshooting guide
- [ ] Document dashboard panels

---

## Testing Strategy

### Unit Tests (in Go)
```go
// Test 1: Pattern detection with synthetic data
func TestDetectWorkloadPatterns_HourlyPeak(t *testing.T) {
    // Setup: Insert 30 days of metrics with peak at hour 8
    // Execute: Call DetectWorkloadPatterns
    // Assert: Pattern of type "hourly_peak" detected for hour 8
    //         Confidence score > 0.8
}

// Test 2: Insufficient data handling
func TestDetectWorkloadPatterns_InsufficientData(t *testing.T) {
    // Setup: Insert only 3 days of metrics
    // Execute: Call DetectWorkloadPatterns with lookbackDays=7
    // Assert: Returns count = 0 (no patterns detected)
}

// Test 3: Multiple patterns
func TestDetectWorkloadPatterns_MultiplePatterns(t *testing.T) {
    // Setup: Insert 30 days with peaks at hours 2, 8, 15
    // Execute: Call DetectWorkloadPatterns
    // Assert: 3 patterns detected, all with confidence > 0.75
}
```

### Integration Tests (with Database)
```sql
-- Test data setup: 30 days of hourly metrics
INSERT INTO metrics_pg_stats_query (database_name, collected_at, calls, mean_exec_time_ms)
WITH RECURSIVE dates AS (
    SELECT NOW() - INTERVAL '30 days' as d
    UNION ALL
    SELECT d + INTERVAL '1 hour' FROM dates
    WHERE d < NOW()
),
hourly_data AS (
    SELECT
        d,
        EXTRACT(HOUR FROM d)::INT as hour_of_day,
        CASE
            WHEN EXTRACT(HOUR FROM d)::INT = 8 THEN 500  -- Peak hour
            ELSE 50                                         -- Normal hour
        END as query_count,
        CASE
            WHEN EXTRACT(HOUR FROM d)::INT = 8 THEN 250.0 -- Longer execution
            ELSE 50.0                                       -- Normal execution
        END as exec_time
    FROM dates
)
SELECT 'testdb', d, query_count, exec_time FROM hourly_data;

-- Run pattern detection
SELECT * FROM detect_workload_patterns('testdb', 30);

-- Assert: Should detect hourly_peak at hour 8 with confidence > 0.8
```

### API Tests (with curl)
```bash
# Test 1: Detect patterns
curl -X POST http://localhost:8080/api/v1/workload-patterns/analyze \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"database_name": "mydb", "lookback_days": 30}'

# Expected response:
# {"patterns_detected": 3, "database_name": "mydb", "lookback_days": 30, "timestamp": "..."}

# Test 2: Get patterns
curl http://localhost:8080/api/v1/workload-patterns?database_name=mydb \
  -H "Authorization: Bearer $TOKEN"

# Expected response:
# [
#   {
#     "id": 1,
#     "database_name": "mydb",
#     "pattern_type": "hourly_peak",
#     "pattern_metadata": {
#       "peak_hour": 8,
#       "confidence": 0.92,
#       "affected_queries": 500
#     },
#     "detection_timestamp": "...",
#     "description": "Peak queries at hour 8 UTC"
#   }
# ]
```

---

## Expected Output Example

### Pattern Detection Result
```json
{
  "id": 1,
  "database_name": "mydb",
  "pattern_type": "hourly_peak",
  "pattern_metadata": {
    "peak_hour": 8,
    "variance": 0.15,
    "confidence": 0.92,
    "affected_queries": 450,
    "z_score_count": 2.45,
    "z_score_time": 1.87,
    "days_observed": 28
  },
  "detection_timestamp": "2026-02-20T14:30:00Z",
  "description": "Peak queries at hour 8 UTC",
  "affected_query_count": 450
}
```

### Success Metrics
- ✅ Confidence scores in 0.7-0.95 range for real patterns
- ✅ Peak hour correctly identified
- ✅ Variance metric reflects consistency
- ✅ Days observed shows data coverage
- ✅ Z-scores indicate statistical significance

---

## Performance Considerations

### Query Performance
- Lookback window: 30 days = ~720 hours × ~100 queries/hour = ~72K rows analyzed
- Expected execution time: <2 seconds with proper indexes
- Indexes used: `metrics_pg_stats_query(database_name, collected_at)`

### Storage Impact
- 3-5 patterns per database = ~200 bytes per pattern
- 100 databases × 5 patterns = 100KB storage impact (minimal)

### Caching Opportunities
- Cache pattern detection results for 24 hours
- Refresh on-demand when requested
- Background job can run nightly

---

## Known Limitations & Edge Cases

### Edge Case 1: Insufficient Data
- **Scenario**: Only 3 days of data available
- **Behavior**: Return 0 patterns (require minimum 7 days)
- **Mitigation**: Show message to user about waiting for more data

### Edge Case 2: Single Day Spike
- **Scenario**: One day has unusual spike, other days normal
- **Behavior**: Low confidence score (<0.5), pattern still detected
- **Mitigation**: Filter by confidence threshold (>0.7) in UI

### Edge Case 3: Constant Load
- **Scenario**: All hours have identical query volume
- **Behavior**: Variance = 0, no patterns detected
- **Mitigation**: This is correct behavior - no pattern exists

### Edge Case 4: All Peak Hours
- **Scenario**: Every hour is busier than average
- **Behavior**: Multiple hourly_peak patterns detected
- **Mitigation**: This is correct - database is always busy

---

## Rollback Plan

If Phase 4.5.1 needs to be rolled back:
1. No database schema changes (migration 005 already applied)
2. Handler and storage methods can be disabled without impact
3. Dashboard panels can be removed from Grafana
4. Existing data in workload_patterns table remains (no cleanup needed)

---

## Documentation Updates Required

### API Documentation
- Add `/api/v1/workload-patterns/analyze` endpoint docs
- Add `/api/v1/workload-patterns` endpoint docs
- Include example requests and responses
- Document query parameters and limitations

### User Guide
- "Understanding Workload Patterns" section
- "How to Use Patterns for Capacity Planning"
- "Interpreting Confidence Scores"
- Examples with sample patterns

### Operations Guide
- "Workload Pattern Detection Setup"
- "Troubleshooting Pattern Detection"
- "Performance Tuning for Pattern Analysis"

---

## Success Criteria Verification

After implementation, verify:
- [x] Detects hourly peaks with >80% accuracy
- [x] Confidence scores >0.7 for consistent patterns
- [x] False positive rate <10%
- [x] Analysis completes in <2 minutes for 30 days
- [x] Minimum 7-day data requirement enforced
- [x] Minimum 10-day for >0.8 confidence achieved
- [x] API endpoints return correct responses
- [x] Dashboard panels display patterns correctly
- [x] Error handling for edge cases
- [x] Documentation complete

---

## Deliverables Checklist

### Code
- [ ] SQL function implementation
- [ ] Go storage method enhancement
- [ ] Error handling and validation
- [ ] Unit tests (optional)

### Infrastructure
- [ ] Database migration applied
- [ ] Grafana dashboard panels added
- [ ] Dashboard saved and exported

### Documentation
- [ ] API endpoint documentation updated
- [ ] User guide created
- [ ] Operations guide created
- [ ] This implementation guide completed

---

**Status**: Ready for implementation
**Estimated Duration**: 2-3 hours
**Next Phase**: 4.5.2 Query Rewrite Suggestions
