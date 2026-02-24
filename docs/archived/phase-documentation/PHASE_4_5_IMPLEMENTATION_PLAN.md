# Phase 4.5: ML-Based Query Optimization Suggestions - Implementation Plan

**Date**: February 20, 2026
**Status**: Planning Complete - Ready for Implementation
**Objective**: Implement machine learning models for automated query optimization recommendations and predictive analytics

---

## Quick Reference: 10 Implementation Tasks

| Task ID | Title | Status | Priority |
|---------|-------|--------|----------|
| 1 | Phase 4.5.1: Workload Pattern Detection | Pending | High |
| 2 | Phase 4.5.2: Query Rewrite Suggestions | Pending | High |
| 3 | Phase 4.5.3: Parameter Optimization | Pending | High |
| 4 | Phase 4.5.4: ML-Powered Workflow | Pending | High |
| 5 | Phase 4.5.5: Python ML Service | Pending | High |
| 6 | Phase 4.5.6: Predictive Modeling | Pending | High |
| 7 | Phase 4.5.7: Database Migration 005 | Pending | Critical |
| 8 | Phase 4.5.8: Go Model Structs | Pending | Critical |
| 9 | Phase 4.5.9: Handlers & Storage | Pending | Critical |
| 10 | Phase 4.5.10: Testing & Verification | Pending | Critical |

---

## Implementation Strategy

### Recommended Execution Order

1. **Foundation (Days 1-2)**
   - Task 7: Create database migration 005_ml_optimization.sql
   - Task 8: Add Go model structs to models.go
   - Task 9: Implement handlers and storage methods

2. **Feature Implementation (Days 3-5)**
   - Task 1: Workload Pattern Detection
   - Task 2: Query Rewrite Suggestions
   - Task 3: Parameter Optimization

3. **ML & Integration (Days 6-8)**
   - Task 5: Python ML Service
   - Task 6: Predictive Performance Modeling
   - Task 4: ML-Powered Optimization Workflow

4. **Verification (Days 9-10)**
   - Task 10: Integration Testing and Verification

---

## Architecture Overview

### Database Tables (Migration 005)

```
workload_patterns
├── Pattern type classification (hourly_peak, daily_cycle, etc)
├── Metadata (peak hour, variance, confidence)
└── Detection timestamps

query_rewrite_suggestions
├── Suggestion type (n_plus_one, subquery_opt, etc)
├── SQL examples (original, suggested)
├── Estimated improvement %
└── Confidence scores

parameter_tuning_suggestions
├── Parameter name (work_mem, sort_mem, LIMIT)
├── Current vs recommended values
├── Estimated improvement %
└── Confidence scores

optimization_recommendations
├── Aggregated from all sources
├── ROI scoring (confidence × impact × urgency)
├── Ranking by priority
└── Creation tracking

optimization_implementations
├── Implementation tracking
├── Pre/post metrics snapshots
├── Actual improvement measurement
└── Status tracking (pending/implemented/reverted)

query_performance_models
├── Trained model storage
├── Feature specifications
├── Model accuracy metrics (R²)
└── Version tracking
```

### API Endpoints (9 Total)

#### Workload Pattern Detection
- `POST /api/v1/workload-patterns/analyze` - Trigger pattern detection
- `GET /api/v1/workload-patterns` - List detected patterns

#### Query Rewrite Suggestions
- `POST /api/v1/queries/{hash}/rewrite-suggestions/generate` - Generate suggestions
- `GET /api/v1/queries/{hash}/rewrite-suggestions` - List suggestions

#### Parameter Optimization
- `GET /api/v1/queries/{hash}/parameter-optimization` - Get tuning recommendations

#### Performance Prediction
- `POST /api/v1/queries/{hash}/predict-performance` - Predict execution time
- `GET /api/v1/optimization-recommendations` - List top recommendations

#### Optimization Workflow
- `POST /api/v1/optimization-recommendations/{id}/implement` - Track implementation
- `GET /api/v1/optimization-results` - Measure results

---

## Phase 4.5 Feature Details

### Feature 1: Workload Pattern Detection
**Goal**: Identify recurring patterns in query execution (time-of-day effects, batch jobs, peak loads)

**Algorithm**:
1. Group query metrics by 1-hour buckets
2. Calculate mean, stddev for each hour across 30 days
3. Identify peak hours where volume/time > mean + 1 stddev
4. Detect recurring patterns using autocorrelation
5. Classify: hourly_peak, daily_cycle, weekly_pattern, batch_job
6. Store with confidence score

**Benefits**:
- Understand when performance issues occur
- Plan maintenance windows
- Predictively scale resources
- Alert on unusual deviations

---

### Feature 2: Query Rewrite Suggestions
**Goal**: Recommend SQL rewrites that improve performance based on EXPLAIN analysis

**Anti-Pattern Detection**:
- N+1 query patterns (same fingerprint, tight timeframe)
- Inefficient joins (Nested Loop → Hash Join)
- Missing indexes (Seq Scans on large tables)
- Subquery optimization opportunities
- IN vs ANY clause optimization

**Deliverables**:
- Specific SQL rewrite examples
- Scoring by estimated improvement potential
- Confidence levels for recommendations
- Reasoning explanation for each suggestion

---

### Feature 3: Parameter Optimization Recommendations
**Goal**: Suggest optimal query parameters (LIMIT, batch size, work_mem, etc.)

**Recommendation Types**:
- work_mem: Based on sort operations and memory availability
- sort_mem: Based on ORDER BY frequency
- LIMIT: Avoiding large result sets
- Batch size: Optimizing processing

**Algorithm**:
1. Group queries by fingerprint
2. Collect parameter variants with metrics
3. Correlation analysis: parameter value vs execution time
4. Find optimal parameter value
5. Confidence = consistency across conditions

---

### Feature 4: Predictive Performance Modeling
**Goal**: Predict query performance after optimization using ML models

**Model Types** (Python ML Service):
- Linear Regression: execution_time = a×table_size + b×index_count + ...
- Decision Tree Regressor: Hierarchical rules
- Random Forest: Ensemble of decision trees
- Gradient Boosting (XGBoost): Advanced non-linear models

**Features Used**:
- Query characteristics: fingerprint, scan_type, join_type
- Table statistics: row_count, table_size_mb, index_count
- Historical patterns: avg_calls_per_hour, peak_hour_impact
- System state: concurrent_queries_avg

**Deliverables**:
- Predicted execution time with confidence interval
- Model accuracy metrics (R²)
- Feature importance analysis
- Periodic model retraining (weekly)

---

### Feature 5: ML-Powered Optimization Workflow
**Goal**: End-to-end workflow for discovering, testing, and implementing optimizations

**Recommendation Ranking**:
```
roi_score = confidence_score × estimated_improvement_percent × urgency_score
urgency_score = frequency × current_impact
```

**Implementation Tracking**:
1. User marks recommendation as implemented
2. System captures pre-optimization metrics
3. Wait 24-48 hours for post-implementation data
4. Compare metrics and actual improvement
5. Update model confidence based on accuracy

**Learning Loop**:
- Track prediction accuracy
- Measure actual vs predicted improvements
- Retrain models with new data
- Improve recommendations iteratively

---

## Python ML Service Architecture

### Components

**ml-service/app.py**
- Flask/FastAPI initialization
- REST endpoint handlers
- Request validation
- Error handling with fallbacks

**ml-service/models/performance_predictor.py**
- PerformanceModel class
- Model training pipeline
- Feature engineering
- Prediction with confidence intervals
- Model serialization/deserialization

**ml-service/models/pattern_detector.py**
- PatternDetector class
- Autocorrelation analysis
- Cycle detection
- Confidence scoring

**ml-service/utils/db_connection.py**
- PostgreSQL connection pool
- Feature extraction queries
- Model storage/retrieval

### Key ML Service Endpoints

```
POST /api/train/performance-model
  → Train execution time prediction model
  → Input: {database_name, lookback_days}
  → Output: {model_id, r_squared, feature_count}

POST /api/predict/query-execution
  → Predict query execution time
  → Input: {query_hash, parameters, scenario}
  → Output: {predicted_ms, confidence, range}

POST /api/detect/patterns
  → Detect workload patterns
  → Input: {database_name, lookback_days}
  → Output: {patterns: [{type, confidence, metadata}]}

GET /api/models/{id}
  → Retrieve trained model metadata
```

---

## File Changes Summary

### New Files
- `backend/migrations/005_ml_optimization.sql` (300+ lines)
- `backend/internal/api/handlers_ml.go` (400+ lines)
- `ml-service/app.py` (200 lines)
- `ml-service/models/performance_predictor.py` (300 lines)
- `ml-service/models/pattern_detector.py` (150 lines)
- `ml-service/utils/db_connection.py` (100 lines)
- `ml-service/requirements.txt`
- `ml-service/Dockerfile`
- `ml-service/config.yaml`

### Modified Files
- `backend/internal/api/server.go` (+30 lines) - Register ML endpoints
- `backend/internal/storage/postgres.go` (+400 lines) - Storage methods
- `backend/pkg/models/models.go` (+350 lines) - Model structs
- `grafana/dashboards/query-performance.json` - 5 new panels

---

## Success Criteria

✅ Workload patterns detected with >80% accuracy
✅ Rewrite suggestions applicable to >50% of slow queries
✅ Parameter recommendations improve performance by >5% on average
✅ Performance predictions within 20% of actual execution time
✅ Top 10 recommendations have ROI score >50
✅ All 9 API endpoints fully functional
✅ Dashboard shows optimization pipeline with metrics
✅ Learning loop: actual vs predicted improvements tracked
✅ <10% false positive rate in pattern detection
✅ Documentation complete with ML methodology

---

## Known Constraints & Mitigations

| Constraint | Mitigation |
|-----------|------------|
| Pattern detection needs historical data | Require 30-day minimum before analysis |
| ML models need representative data | Start with conservative confidence thresholds |
| Rewrite suggestions may not work for all queries | Rate suggestions by confidence, allow feedback |
| Performance predictions can be inaccurate | Show confidence intervals, track accuracy |
| Optimization recommendations may not apply | Allow dismissal, track actual implementation success |

---

## Timeline Estimate

**Phase 4.5.1**: Workload Pattern Detection (3-4 days)
**Phase 4.5.2**: Query Rewrite Suggestions (3-4 days)
**Phase 4.5.3**: Parameter Optimization (2-3 days)
**Phase 4.5.4**: ML-Powered Workflow (2-3 days)
**Phase 4.5.5**: Python ML Service (3-4 days)

**Total**: ~3 weeks with concurrent development

---

## Integration with Previous Phases

**Phase 4.4 Integration**:
- Uses query fingerprints for ML features
- Analyzes EXPLAIN plans for rewrite patterns
- Index recommendations inform parameter tuning
- Performance snapshots provide training data
- Anomalies trigger suggestion generation

**Phase 4.1-4.3 Integration**:
- Uses query metrics from pg_stat_statements
- Leverages historical data from TimescaleDB
- Builds on existing time-series aggregation

---

## Deployment Notes

### Docker Deployment
```
Services:
- api-backend (Go) - unchanged
- ml-service (Python) - new
- postgres - unchanged
- grafana - unchanged

Environment variables:
- ML_SERVICE_URL=http://ml-service:8081
- ML_SERVICE_TIMEOUT=5s
```

### Configuration
```
Backend:
- ML_SERVICE_ENABLED=true (allow disable)
- ML_SERVICE_TIMEOUT=5s
- CIRCUIT_BREAKER_THRESHOLD=5 failures

ML Service:
- DATABASE_URL=postgresql://...
- LOG_LEVEL=INFO
- MODEL_REFRESH_INTERVAL=7d
- MAX_WORKERS=4
```

---

## Next Steps

1. ✅ Plan created and approved
2. → Start Task 7: Create database migration
3. → Start Task 8: Add Go model structs
4. → Start Task 9: Implement handlers and storage
5. → Continue with feature implementation
6. → Create Python ML service
7. → Integration testing
8. → Dashboard integration
9. → Documentation
10. → Production deployment

---

**Ready to start implementation. Use `TaskList` to view all tasks and `TaskUpdate` to track progress.**
