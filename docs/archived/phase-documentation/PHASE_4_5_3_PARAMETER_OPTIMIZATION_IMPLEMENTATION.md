# Phase 4.5.3: Parameter Optimization Recommendations - Implementation Specification

**Date**: February 20, 2026
**Status**: Implementation In Progress
**Duration**: 3-4 days estimated
**Lines of Code**: 300+ (SQL + Go enhancements)

---

## What Will Be Implemented

### 1. SQL Function: optimize_parameters()
**File**: `backend/migrations/005_ml_optimization.sql`
**Status**: Implementation In Progress

**Features**:
- ✅ work_mem optimization detection
- ✅ sort_mem optimization detection
- ✅ LIMIT value optimization
- ✅ Batch size recommendations
- ✅ Confidence scoring for each parameter
- ✅ Improvement percentage estimation
- ✅ Detailed reasoning for each suggestion
- ✅ CONFLICT handling (don't create duplicates)

**Parameter Optimization Strategies**:

#### 1. work_mem Optimization
**Detection Logic**:
```
- High sort operations: Has ORDER BY, GROUP BY, DISTINCT
- Slow execution with high memory usage
- Current work_mem from pg_settings
- Correlation: More work_mem → faster sorts (but limits connections)
```

**Recommendation Rules**:
```
- Query has sorts/aggregates AND mean_exec_time > 500ms
  → Recommend increase by 50% (current × 1.5)

- Multiple simultaneous queries with sorts
  → Recommend conservative increase (current × 1.2)

- Very small queries (< 50ms)
  → Recommend decrease (current × 0.8) to free memory
```

**Confidence Calculation**:
```
- Has clear SORT/GROUP BY nodes: confidence = 0.85
- Slow execution (> 500ms): confidence += 0.05
- Frequent execution (> 100 calls): confidence += 0.05
- Actual memory available: confidence -= (if constrained)
```

#### 2. sort_mem Optimization
**Detection Logic**:
```
- Dedicated parameter for sort operations (PostgreSQL 16+)
- Similar to work_mem but specialized for sorts
- Detects queries with heavy sorting requirements
```

**Recommendation Rules**:
```
- Heavy sort operations (ORDER BY with large result sets)
  → Recommend sort_mem increase

- Multiple sorts in same query
  → More aggressive recommendation
```

**Confidence**: 0.75-0.90

#### 3. LIMIT Value Optimization
**Detection Logic**:
```
- Analyze query execution patterns
- Query retrieves N rows but client uses first M rows
- LIMIT M would significantly reduce execution time
```

**Recommendation Rules**:
```
- Query without LIMIT returning > 1000 rows: HIGH confidence
  → Recommend LIMIT 100 or LIMIT 1000 based on typical usage

- Query with LIMIT already: Review if value is optimal
  → Analyze row patterns to optimize LIMIT value

- Queries with ORDER BY + no LIMIT
  → Recommend LIMIT to stop early
```

**Confidence**: 0.70-0.95 (depends on result set size consistency)

#### 4. Batch Size Recommendations
**Detection Logic**:
```
- Queries called frequently in tight loops (N+1 pattern)
- Calculate optimal batch size from:
  - Call frequency
  - Memory constraints
  - Network latency
  - Connection pool capacity
```

**Recommendation Rules**:
```
- N+1 pattern detected with 100+ calls/hour
  → Recommend batch sizes: 10, 50, 100, 500
  → Show performance impact for each batch size

- High frequency query (> 1000 calls/hour)
  → Recommend optimal batch size: frequency / 60 / 10
```

**Confidence**: 0.70-0.80

---

## Algorithm Overview

### detect_missing_limits() Function
**Purpose**: Find queries that would benefit from LIMIT clause

**Algorithm**:
```sql
1. Find queries without LIMIT clause
2. Filter: mean_exec_time > 100ms (slow queries)
3. For each query:
   a. Estimate result set size from metrics
   b. Calculate typical rows used percentage
   c. IF typical_usage < 20% of result set
      → LIMIT = typical_usage * 1.2
      → confidence = 0.85
   d. IF typical_usage < 5% of result set
      → confidence = 0.95
   e. Insert suggestion

Example:
- Query: SELECT * FROM users WHERE active = true
- Typical exec: Returns 10,000 rows but client uses first 100
- Suggestion: ADD LIMIT 120
- Confidence: 0.95
- Improvement: 85% (stops scanning after 120 rows instead of 10,000)
```

### detect_sort_optimization() Function
**Purpose**: Find queries that would benefit from work_mem/sort_mem increase

**Algorithm**:
```sql
1. Find queries with SORT nodes in EXPLAIN plan
2. Filter: mean_exec_time > 200ms
3. For each query:
   a. Detect sort type: ORDER BY, GROUP BY, DISTINCT
   b. Calculate current work_mem (from pg_settings)
   c. Estimate sort operation cost
   d. IF execution_time > threshold AND has_sort
      → Estimate new work_mem = current * 1.5
      → confidence = 0.80 + (exec_time_factor)
   e. Insert suggestion

Example:
- Query: SELECT * FROM orders ORDER BY created_at
- Mean exec: 450ms
- Current work_mem: 4MB
- Suggestion: SET work_mem = '6MB'
- Confidence: 0.85
- Improvement: 30% (faster sorting with more memory)
```

### detect_batch_optimization() Function
**Purpose**: Find N+1 patterns and recommend batch sizes

**Algorithm**:
```sql
1. Find queries with call_count > 50
2. Filter fingerprints that appear to be N+1 pattern
3. For each pattern:
   a. Calculate current call frequency
   b. Recommend batch sizes: 10, 50, 100, 500
   c. For each batch size:
      → Estimate new execution time
      → Calculate reduction in total_time
      → confidence = 0.70 + (pattern_consistency)
   d. Insert suggestions for top 3 batch sizes

Example:
- Query: SELECT * FROM users WHERE id = ?
- Call frequency: 500/hour
- Suggestion 1: Use LIMIT 50, confidence 0.75, improvement 75%
- Suggestion 2: Use LIMIT 100, confidence 0.73, improvement 70%
- Suggestion 3: Use LIMIT 10, confidence 0.72, improvement 80%
```

---

## Detection Pattern Details

### Pattern 1: Missing LIMIT Clauses
**Detection**: `query_text NOT ILIKE '%LIMIT%' AND mean_exec_time > 100ms`
**Example**: SELECT * FROM orders WHERE status = 'active'
**Suggestion**: ADD LIMIT 100 (or appropriate value based on usage)
**Confidence**: 0.70-0.95 (higher with consistent result set sizes)
**Improvement**: 50-90% (depends on typical usage vs result set)

### Pattern 2: Inefficient work_mem Settings
**Detection**: `has_sort_nodes = TRUE AND mean_exec_time > 200ms AND call_count > 10`
**Example**: SELECT * FROM orders ORDER BY created_at DESC (with 450ms execution)
**Suggestion**: SET work_mem = '6MB' (increase from 4MB)
**Confidence**: 0.80-0.90
**Improvement**: 20-40% (faster sort operations)

### Pattern 3: sort_mem Optimization (PostgreSQL 16+)
**Detection**: `has_sort_nodes = TRUE AND mean_exec_time > 300ms`
**Example**: Large ORDER BY with GROUP BY
**Suggestion**: SET sort_mem = '16MB'
**Confidence**: 0.75-0.85
**Improvement**: 15-30%

### Pattern 4: Batch Size Opportunities
**Detection**: `call_count > 100 AND fingerprint_pattern = 'parameterized_lookup'`
**Example**: 500 calls/hour of SELECT * FROM users WHERE id = $1
**Suggestions**:
- Batch 10: improvement 70%
- Batch 50: improvement 75%
- Batch 100: improvement 72%
**Confidence**: 0.70-0.80
**Improvement**: 70-75%

### Pattern 5: Result Set Limiting
**Detection**: `query_text NOT ILIKE '%LIMIT%' AND result_rows > 1000`
**Example**: SELECT * FROM events WHERE timestamp > now() - '7 days'
**Suggestion**: ADD LIMIT 1000 (scan optimization)
**Confidence**: 0.85-0.95
**Improvement**: 40-80%

---

## SQL Implementation

### Table: parameter_tuning_suggestions
Already defined in 005_ml_optimization.sql with columns:
- id, query_hash, parameter_name, current_value, recommended_value
- reasoning, estimated_improvement_percent, confidence_score
- created_at, updated_at

### Function: optimize_parameters()
**Signature**:
```sql
CREATE OR REPLACE FUNCTION optimize_parameters(
    p_query_hash BIGINT
)
RETURNS TABLE (
    suggestion_count INT,
    parameter_types TEXT[]
) AS $$
```

**Implementation**:
```sql
DECLARE
    v_query_hash BIGINT;
    v_query_text TEXT;
    v_has_sort BOOLEAN;
    v_has_limit BOOLEAN;
    v_mean_exec_time FLOAT;
    v_call_count INT;
    v_current_work_mem VARCHAR;
    v_result_rows INT;
BEGIN
    -- 1. Get query details from metrics table
    SELECT mpsq.query_hash, mpsq.query_text, mpsq.mean_exec_time,
           mpsq.calls, mpsq.rows
    INTO v_query_hash, v_query_text, v_mean_exec_time, v_call_count, v_result_rows
    FROM metrics_pg_stats_query mpsq
    WHERE mpsq.query_hash = p_query_hash;

    -- 2. Validate query exists
    IF v_query_hash IS NULL THEN
        RETURN QUERY SELECT 0::INT, ARRAY[]::TEXT[];
        RETURN;
    END IF;

    -- 3. Analyze query characteristics
    v_has_sort := (v_query_text ILIKE '%ORDER BY%' OR
                   v_query_text ILIKE '%GROUP BY%' OR
                   v_query_text ILIKE '%DISTINCT%');
    v_has_limit := v_query_text ILIKE '%LIMIT%';
    v_current_work_mem := current_setting('work_mem');

    -- 4. Generate LIMIT recommendations
    IF NOT v_has_limit AND v_mean_exec_time > 100 AND v_result_rows > 100 THEN
        INSERT INTO parameter_tuning_suggestions
            (query_hash, parameter_name, current_value, recommended_value,
             reasoning, estimated_improvement_percent, confidence_score)
        VALUES (
            v_query_hash,
            'LIMIT',
            'NOT SET',
            'LIMIT ' || (v_result_rows / 10)::INT,
            'Query returns ' || v_result_rows || ' rows. Consider adding LIMIT to stop scanning early.',
            CASE
                WHEN v_result_rows > 5000 THEN 85.0
                WHEN v_result_rows > 1000 THEN 75.0
                ELSE 50.0
            END,
            CASE
                WHEN v_result_rows > 5000 THEN 0.95
                WHEN v_result_rows > 1000 THEN 0.85
                ELSE 0.70
            END
        )
        ON CONFLICT (query_hash, parameter_name) DO UPDATE SET
            updated_at = NOW();
    END IF;

    -- 5. Generate work_mem recommendations
    IF v_has_sort AND v_mean_exec_time > 200 AND v_call_count > 10 THEN
        INSERT INTO parameter_tuning_suggestions
            (query_hash, parameter_name, current_value, recommended_value,
             reasoning, estimated_improvement_percent, confidence_score)
        VALUES (
            v_query_hash,
            'work_mem',
            v_current_work_mem,
            (CAST(split_part(v_current_work_mem, 'M', 1) AS INT) * 1.5)::TEXT || 'MB',
            'Query has sort/group operations taking ' || ROUND(v_mean_exec_time, 0)::TEXT || 'ms. More work_mem would speed up sorting.',
            CASE
                WHEN v_mean_exec_time > 500 THEN 35.0
                WHEN v_mean_exec_time > 300 THEN 25.0
                ELSE 15.0
            END,
            CASE
                WHEN v_mean_exec_time > 500 THEN 0.90
                WHEN v_mean_exec_time > 300 THEN 0.85
                ELSE 0.80
            END
        )
        ON CONFLICT (query_hash, parameter_name) DO UPDATE SET
            updated_at = NOW();
    END IF;

    -- 6. Generate batch size recommendations
    IF v_call_count > 100 THEN
        -- Recommend batch sizes: 50, 100, 500
        FOR i IN ARRAY[50, 100, 500] LOOP
            INSERT INTO parameter_tuning_suggestions
                (query_hash, parameter_name, current_value, recommended_value,
                 reasoning, estimated_improvement_percent, confidence_score)
            VALUES (
                v_query_hash,
                'batch_size',
                '1',
                i::TEXT,
                'Query called ' || v_call_count || ' times. Consider batching with size ' || i::TEXT || '.',
                CASE
                    WHEN i = 50 THEN 75.0
                    WHEN i = 100 THEN 72.0
                    ELSE 70.0
                END,
                CASE
                    WHEN i = 50 THEN 0.75
                    WHEN i = 100 THEN 0.73
                    ELSE 0.70
                END
            )
            ON CONFLICT (query_hash, parameter_name) DO UPDATE SET
                updated_at = NOW();
        END LOOP;
    END IF;

    -- 7. Return results
    RETURN QUERY
    SELECT
        COUNT(*)::INT,
        ARRAY_AGG(DISTINCT parameter_name)
    FROM parameter_tuning_suggestions
    WHERE query_hash = v_query_hash
      AND created_at > NOW() - INTERVAL '1 minute';
END;
$$ LANGUAGE plpgsql;
```

**Key Features**:
- Validates query_hash exists
- Analyzes query text for patterns (ORDER BY, GROUP BY, LIMIT)
- Detects sort operations
- Generates work_mem recommendations based on execution time
- Generates LIMIT recommendations for large result sets
- Generates batch size recommendations for frequently called queries
- Uses ON CONFLICT for idempotency
- Returns suggestion count and types for API response

---

## Go Implementation

### Storage Method: optimize_parameters()
**File**: `backend/internal/storage/postgres.go`

**Signature**:
```go
func (p *PostgresDB) OptimizeParameters(
    ctx context.Context,
    queryHash int64,
) ([]models.ParameterTuningSuggestion, error)
```

**Implementation**:
```go
func (p *PostgresDB) OptimizeParameters(
    ctx context.Context,
    queryHash int64,
) ([]models.ParameterTuningSuggestion, error) {
    // Validate input
    if queryHash <= 0 {
        return nil, apperrors.ValidationError("query_hash must be positive")
    }

    // Call SQL function to generate suggestions
    _, err := p.db.ExecContext(
        ctx,
        `SELECT optimize_parameters($1)`,
        queryHash,
    )
    if err != nil {
        p.logger.Warnf("Failed to generate parameter suggestions: %v", err)
        return nil, apperrors.DatabaseError("optimize parameters", err.Error())
    }

    // Retrieve generated suggestions
    rows, err := p.db.QueryContext(
        ctx,
        `SELECT id, query_hash, parameter_name, current_value, recommended_value,
                reasoning, estimated_improvement_percent, confidence_score, created_at
         FROM parameter_tuning_suggestions
         WHERE query_hash = $1
         ORDER BY confidence_score DESC, estimated_improvement_percent DESC`,
        queryHash,
    )
    if err != nil {
        p.logger.Errorf("Failed to retrieve parameter suggestions: %v", err)
        return nil, apperrors.DatabaseError("get parameter suggestions", err.Error())
    }
    defer rows.Close()

    var suggestions []models.ParameterTuningSuggestion
    for rows.Next() {
        var s models.ParameterTuningSuggestion
        var createdAt time.Time

        err := rows.Scan(
            &s.ID,
            &s.QueryHash,
            &s.ParameterName,
            &s.CurrentValue,
            &s.RecommendedValue,
            &s.Reasoning,
            &s.EstimatedImprovement,
            &s.ConfidenceScore,
            &createdAt,
        )
        if err != nil {
            p.logger.Errorf("Failed to scan parameter suggestion: %v", err)
            continue
        }

        s.CreatedAt = createdAt
        suggestions = append(suggestions, s)
    }

    p.logger.Infof("Generated %d parameter optimization suggestions for query %d", len(suggestions), queryHash)
    return suggestions, nil
}
```

### Storage Method: GetParameterTuningSuggestions()
**Signature**:
```go
func (p *PostgresDB) GetParameterTuningSuggestions(
    ctx context.Context,
    queryHash int64,
    limit int,
) ([]models.ParameterTuningSuggestion, error)
```

**Implementation**:
```go
func (p *PostgresDB) GetParameterTuningSuggestions(
    ctx context.Context,
    queryHash int64,
    limit int,
) ([]models.ParameterTuningSuggestion, error) {
    // Validate inputs
    if queryHash <= 0 {
        return nil, apperrors.ValidationError("query_hash must be positive")
    }

    if limit <= 0 || limit > 100 {
        limit = 10
    }

    rows, err := p.db.QueryContext(
        ctx,
        `SELECT id, query_hash, parameter_name, current_value, recommended_value,
                reasoning, estimated_improvement_percent, confidence_score, created_at
         FROM parameter_tuning_suggestions
         WHERE query_hash = $1
         ORDER BY confidence_score DESC, estimated_improvement_percent DESC
         LIMIT $2`,
        queryHash,
        limit,
    )
    if err != nil {
        p.logger.Errorf("Failed to query parameter suggestions: %v", err)
        return nil, apperrors.DatabaseError("get parameter suggestions", err.Error())
    }
    defer rows.Close()

    var suggestions []models.ParameterTuningSuggestion
    for rows.Next() {
        var s models.ParameterTuningSuggestion
        var createdAt time.Time

        err := rows.Scan(
            &s.ID,
            &s.QueryHash,
            &s.ParameterName,
            &s.CurrentValue,
            &s.RecommendedValue,
            &s.Reasoning,
            &s.EstimatedImprovement,
            &s.ConfidenceScore,
            &createdAt,
        )
        if err != nil {
            p.logger.Errorf("Failed to scan parameter suggestion: %v", err)
            continue
        }

        s.CreatedAt = createdAt
        suggestions = append(suggestions, s)
    }

    if len(suggestions) == 0 {
        p.logger.Debugf("No parameter suggestions found for query %d", queryHash)
        return []models.ParameterTuningSuggestion{}, nil
    }

    p.logger.Infof("Retrieved %d parameter optimization suggestions for query %d", len(suggestions), queryHash)
    return suggestions, nil
}
```

---

## API Handler Enhancements

### Handler: handleGetParameterOptimization()
**Endpoint**: `GET /api/v1/queries/{query_hash}/parameter-optimization`

**Enhanced Implementation**:
```go
func (s *Server) handleGetParameterOptimization(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    queryHashStr := c.Param("query_hash")
    queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
    if err != nil {
        s.logger.Warnf("Invalid query_hash format: %s", queryHashStr)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
        return
    }

    if queryHash <= 0 {
        s.logger.Warnf("Invalid query_hash value: %d", queryHash)
        c.JSON(http.StatusBadRequest, gin.H{"error": "query_hash must be positive"})
        return
    }

    // Parse optional limit parameter
    limit := 10
    if limitStr := c.Query("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
            limit = l
        }
    }

    s.logger.Debugf("Fetching parameter optimization suggestions for query %d (limit: %d)", queryHash, limit)

    // Get suggestions from database
    suggestions, err := s.postgres.GetParameterTuningSuggestions(ctx, queryHash, limit)
    if err != nil {
        s.logger.Errorf("Failed to get parameter suggestions: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve parameter suggestions"})
        return
    }

    if len(suggestions) == 0 {
        s.logger.Infof("No parameter optimization suggestions available for query %d", queryHash)
        c.JSON(http.StatusOK, gin.H{
            "query_hash":   queryHash,
            "suggestions":  []models.ParameterTuningSuggestion{},
            "count":        0,
            "message":      "No parameter optimization suggestions available",
        })
        return
    }

    // Group suggestions by parameter type
    paramTypes := make(map[string]int)
    for _, s := range suggestions {
        paramTypes[s.ParameterName]++
    }

    s.logger.Infof("Retrieved %d parameter suggestions for query %d", len(suggestions), queryHash)

    c.JSON(http.StatusOK, gin.H{
        "query_hash":      queryHash,
        "suggestions":     suggestions,
        "count":           len(suggestions),
        "parameter_types": paramTypes,
        "timestamp":       time.Now().UTC(),
    })
}
```

### Handler: handleOptimizeParameters()
**Endpoint**: `POST /api/v1/queries/{query_hash}/parameter-optimization/generate`

**Implementation**:
```go
func (s *Server) handleOptimizeParameters(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
    defer cancel()

    queryHashStr := c.Param("query_hash")
    queryHash, err := strconv.ParseInt(queryHashStr, 10, 64)
    if err != nil {
        s.logger.Warnf("Invalid query_hash format: %s", queryHashStr)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query_hash format"})
        return
    }

    if queryHash <= 0 {
        s.logger.Warnf("Invalid query_hash value: %d", queryHash)
        c.JSON(http.StatusBadRequest, gin.H{"error": "query_hash must be positive"})
        return
    }

    s.logger.Infof("Generating parameter optimization suggestions for query %d", queryHash)

    // Generate suggestions using SQL function
    suggestions, err := s.postgres.OptimizeParameters(ctx, queryHash)
    if err != nil {
        s.logger.Errorf("Failed to generate parameter suggestions: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate parameter suggestions"})
        return
    }

    // Extract parameter types
    paramTypes := make(map[string]bool)
    for _, s := range suggestions {
        paramTypes[s.ParameterName] = true
    }

    typesList := make([]string, 0, len(paramTypes))
    for paramType := range paramTypes {
        typesList = append(typesList, paramType)
    }

    s.logger.Infof("Generated %d parameter optimization suggestions for query %d", len(suggestions), queryHash)

    c.JSON(http.StatusOK, gin.H{
        "query_hash":          queryHash,
        "suggestions_count":   len(suggestions),
        "suggestion_types":    typesList,
        "generated_at":        time.Now().UTC(),
        "message":             fmt.Sprintf("Generated %d parameter optimization suggestions", len(suggestions)),
    })
}
```

---

## Testing Strategy

### Database Level Tests

**Test 1: LIMIT Recommendation for Large Result Sets**
```sql
-- Setup: Create query with large result set
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time, calls, rows)
VALUES (2001, 'SELECT * FROM events WHERE type = $1', 250.0, 50, 5000);

-- Execute: Generate suggestions
SELECT optimize_parameters(2001);

-- Verify: LIMIT suggestion created
SELECT * FROM parameter_tuning_suggestions
WHERE query_hash = 2001 AND parameter_name = 'LIMIT';

-- Expected: confidence >= 0.85, improvement >= 50
```

**Test 2: work_mem Recommendation for Sort Operations**
```sql
-- Setup
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time, calls, rows)
VALUES (2002, 'SELECT * FROM orders ORDER BY created_at DESC', 450.0, 150, 10000);

-- Execute
SELECT optimize_parameters(2002);

-- Verify
SELECT * FROM parameter_tuning_suggestions
WHERE query_hash = 2002 AND parameter_name = 'work_mem';

-- Expected: confidence >= 0.85, improvement >= 25
```

**Test 3: Batch Size Recommendations**
```sql
-- Setup: High frequency query
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time, calls, rows)
VALUES (2003, 'SELECT * FROM users WHERE id = $1', 15.0, 500, 1);

-- Execute
SELECT optimize_parameters(2003);

-- Verify
SELECT * FROM parameter_tuning_suggestions
WHERE query_hash = 2003 AND parameter_name = 'batch_size'
ORDER BY confidence_score DESC;

-- Expected: 3 suggestions with confidence 0.70-0.75
```

**Test 4: GROUP BY work_mem Optimization**
```sql
-- Setup
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time, calls, rows)
VALUES (2004, 'SELECT category, COUNT(*) FROM products GROUP BY category', 350.0, 100, 50);

-- Execute
SELECT optimize_parameters(2004);

-- Verify: work_mem recommendation
SELECT * FROM parameter_tuning_suggestions
WHERE query_hash = 2004 AND parameter_name = 'work_mem';
```

**Test 5: No Recommendations for Optimal Query**
```sql
-- Setup: Fast query already optimized
INSERT INTO metrics_pg_stats_query (query_hash, query_text, mean_exec_time, calls, rows)
VALUES (2005, 'SELECT * FROM users WHERE id = $1 LIMIT 1', 5.0, 10, 1);

-- Execute
SELECT optimize_parameters(2005);

-- Verify: Few or no suggestions (already optimal)
SELECT COUNT(*) FROM parameter_tuning_suggestions WHERE query_hash = 2005;

-- Expected: 0 suggestions
```

### API Level Tests

**Test 1: Generate Parameter Suggestions**
```bash
curl -X POST http://localhost:8080/api/v1/queries/2001/parameter-optimization/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# Expected: 200 OK with suggestion count
# Response: {
#   "query_hash": 2001,
#   "suggestions_count": 3,
#   "suggestion_types": ["LIMIT", "work_mem", "batch_size"],
#   "generated_at": "2026-02-20T..."
# }
```

**Test 2: Get Parameter Suggestions with Limit**
```bash
curl "http://localhost:8080/api/v1/queries/2001/parameter-optimization?limit=5" \
  -H "Authorization: Bearer $TOKEN"

# Expected: 200 OK with suggestions array
# Response: {
#   "query_hash": 2001,
#   "suggestions": [
#     {
#       "id": 1,
#       "parameter_name": "LIMIT",
#       "recommended_value": "LIMIT 500",
#       "confidence_score": 0.95,
#       "estimated_improvement_percent": 85.0
#     }
#   ],
#   "count": 3
# }
```

**Test 3: Invalid Query Hash**
```bash
curl "http://localhost:8080/api/v1/queries/0/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN"

# Expected: 400 Bad Request
# Response: {"error": "query_hash must be positive"}
```

**Test 4: Non-existent Query**
```bash
curl "http://localhost:8080/api/v1/queries/99999/parameter-optimization" \
  -H "Authorization: Bearer $TOKEN"

# Expected: 200 OK with empty suggestions array
# Response: {
#   "query_hash": 99999,
#   "suggestions": [],
#   "count": 0,
#   "message": "No parameter optimization suggestions available"
# }
```

**Test 5: Parameter Type Grouping**
```bash
curl "http://localhost:8080/api/v1/queries/2001/parameter-optimization?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.parameter_types'

# Expected: Group count by parameter type
# {
#   "LIMIT": 1,
#   "work_mem": 1,
#   "batch_size": 3
# }
```

### Integration Tests

**Test 1: Full Workflow - Detection to Recommendations**
```
1. Insert slow query with large result set into metrics
2. POST to /parameter-optimization/generate
3. Verify suggestions created with appropriate confidence
4. GET /parameter-optimization to retrieve
5. Verify ranking by confidence_score
6. Verify suggestions are actionable (specific parameter values)
```

**Test 2: Multiple Query Types**
```
1. Insert 5 different query types:
   - Simple lookup (batch opportunity)
   - Large result set (LIMIT opportunity)
   - Sort operation (work_mem opportunity)
   - Aggregate (work_mem opportunity)
   - Already optimized (no suggestions)

2. Generate suggestions for all 5 queries
3. Verify each gets appropriate recommendations
4. Verify confidence scores differ by query type
5. Verify improvement percentages are reasonable
```

**Test 3: Idempotency**
```
1. Generate suggestions for query X
2. Retrieve and count suggestions (should be N)
3. Generate suggestions again for same query X
4. Retrieve again - should still be N (not duplicated)
5. Verify ON CONFLICT worked correctly
```

---

## Success Criteria

✅ **Criteria 1**: LIMIT recommendations generated for queries with large result sets
- Confidence scores 0.70-0.95 based on result set size
- Improvement percentages 50-90%
- Actionable recommendations with specific LIMIT values

✅ **Criteria 2**: work_mem recommendations for sort operations
- Detected via ORDER BY, GROUP BY, DISTINCT in query text
- Confidence scores 0.80-0.90
- Improvement percentages 15-40%
- Reasonable multipliers (1.5x - 2.0x current work_mem)

✅ **Criteria 3**: Batch size recommendations for N+1 patterns
- Generated for queries with call_count > 100
- Multiple recommendations (10, 50, 100, 500 batch sizes)
- Confidence scores 0.70-0.75
- Improvement percentages 70-75%

✅ **Criteria 4**: Confidence scoring works correctly
- All confidence scores in 0.0-1.0 range
- Higher confidence for larger result sets (LIMIT recommendations)
- Higher confidence for more frequent slow operations (work_mem)
- Lower confidence for batch size (less certain benefit)

✅ **Criteria 5**: API endpoints fully functional
- POST /parameter-optimization/generate works correctly
- GET /parameter-optimization returns suggestions
- Proper validation of query_hash
- Correct HTTP status codes
- Proper error handling for edge cases

✅ **Criteria 6**: Database function idempotent
- ON CONFLICT handling prevents duplicates
- Repeated calls don't multiply suggestions
- updated_at timestamp updated on conflicts

✅ **Criteria 7**: Suggestions ranked by confidence and improvement
- Suggestions ordered by confidence_score DESC
- Secondary ordering by estimated_improvement_percent DESC
- API returns suggestions in priority order

✅ **Criteria 8**: Error handling complete
- Invalid query_hash rejected
- Non-existent queries return empty results (not error)
- Database errors logged and handled gracefully
- Context timeouts working (15s for generation, 10s for retrieval)

---

## Known Considerations

### work_mem Recommendations
- Assumes increasing work_mem will improve sort performance
- May not help if bottleneck is I/O or CPU, not memory
- Should track actual improvement after recommendation implemented
- May be limited by total available system memory

### LIMIT Recommendations
- Assumes client doesn't need full result set
- Should be validated by analyzing actual row usage patterns
- May not work for all query types (aggregates, JOINs)
- Confidence varies by result set consistency

### Batch Size Recommendations
- Assumes application can be modified to batch queries
- Requires transaction isolation consideration
- Benefits depend on network latency (higher latency = larger batches better)
- May conflict with connection pool size

### Parameter Changes Impact
- work_mem changes affect all sorts globally (session or user level)
- LIMIT changes require code modifications
- Batch size changes require application logic changes
- All should be tested before production deployment

---

## Integration with Previous Phases

### Phase 4.5.1 (Workload Patterns)
- Workload patterns can trigger parameter optimization review
- High peak hour can suggest batch size recommendations
- Identify patterns → optimize parameters for those patterns

### Phase 4.5.2 (Query Rewrites)
- May complement rewrite suggestions
- Both target performance improvement but different approaches
- Rewrite: change query structure
- Parameter: tune execution parameters

### Phase 4.5.4 (ML Workflow)
- Parameter suggestions feed into recommendation ranking
- Combined with other suggestion types for ROI scoring
- Implementation tracking measures actual improvements
- Confidence scores inform recommendation priority

---

## Deployment Checklist

- [ ] SQL function optimize_parameters() implemented in 005_ml_optimization.sql
- [ ] Go storage methods OptimizeParameters() and GetParameterTuningSuggestions() added
- [ ] API handlers handleOptimizeParameters() and handleGetParameterOptimization() implemented
- [ ] Routes registered in server.go with AuthMiddleware
- [ ] Error handling complete with apperrors mapping
- [ ] Logging added at info/debug levels
- [ ] Input validation on all parameters
- [ ] Testing guide created with 12+ test cases
- [ ] Database migration verified
- [ ] Code verified with go fmt
- [ ] No breaking changes to existing endpoints

---

**Status**: Ready for implementation

