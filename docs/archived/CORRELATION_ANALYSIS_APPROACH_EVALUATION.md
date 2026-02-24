# Correlation Analysis Approach Evaluation: AI vs Graph vs Hybrid

**Date**: February 22, 2026
**Project**: pganalytics-v3
**Status**: Strategic Evaluation for Architecture Decision

---

## Executive Summary

For PostgreSQL monitoring and anomaly correlation, **a Hybrid approach combining Graph-based causal analysis with AI-driven pattern recognition is recommended**. Pure AI alone lacks explainability for production databases. Pure graphs require manual relationship definition. The hybrid approach provides both real-time pattern detection AND causality understanding.

**Recommendation: Implement Phase 1 as Graph+Rules, Phase 2 add Lightweight ML for pattern refinement**

---

## The Problem Context

We need to:
1. **Identify correlations** between PostgreSQL metrics (CPU ↔ IO latency ↔ lock contention)
2. **Detect causal relationships** (which metric caused the issue, not just coincidence)
3. **Root cause analysis** (CPU spike → caused by missing index → full table scans → high IO)
4. **Real-time operation** (streaming metrics, decisions in <100ms)
5. **Explainability** (production DBAs must understand WHY an issue occurred)

---

## Approach 1: Pure AI (Statistical ML Models)

### What It Is
Use statistical methods and machine learning models to learn patterns from historical data and predict/detect anomalies autonomously.

**Techniques**:
- Pearson/Spearman correlation coefficients
- Time-series forecasting (ARIMA, exponential smoothing)
- Clustering (K-means for grouping similar query patterns)
- Isolation Forests for anomaly detection
- LSTM neural networks for sequence patterns
- Causal inference models (if well-trained)

### Strengths ✅

1. **Pattern Recognition**: Discovers non-obvious correlations humans miss
   - Example: `slow_queries` ↔ `cache_hit_ratio` with 3-minute lag
   - Example: `disk_io` spikes correlate with `vacuum_duration` 50% of the time

2. **Adaptive Learning**: Improves with more data, learns seasonal patterns
   - Differentiates between normal business hours volatility and actual problems
   - Learns that Sundays have different query patterns

3. **Real-time Detection**: Once trained, predictions are instant (<1ms)
   - Can detect emerging issues as they develop
   - Can correlate across 50+ metrics simultaneously

4. **No Manual Configuration**: Doesn't require DBAs to pre-define relationships
   - Set it and forget it
   - Works across diverse PostgreSQL deployments

5. **Scalability**: Handles high-dimensional data (many metrics) naturally
   - Works with 100+ metrics without degradation
   - Can correlate across multiple servers simultaneously

### Weaknesses ❌

1. **Black Box Problem**: Cannot explain WHY correlations exist
   - ML says "CPU and IO are correlated" but cannot say "it's because of missing index"
   - In production databases, DBAs MUST understand the cause
   - Risk of implementing wrong fixes based on correlation without causality

2. **Data Quality Dependency**: Requires clean historical data
   - Garbage in → garbage out
   - If your metrics are inconsistent or gaps exist, model fails
   - Needs 2-4 weeks of clean baseline data before being useful

3. **False Positives**: Correlation ≠ Causation
   - May detect spurious correlations (both metrics spike at noon = lunch traffic, unrelated)
   - Difficult to tune false positive rate without extensive validation

4. **Training & Tuning Overhead**:
   - Requires ML expertise to implement properly
   - Must continuously retrain (monthly/quarterly) as baselines shift
   - Hyperparameter tuning is non-trivial (which window size? which threshold?)

5. **Cold Start Problem**:
   - Useless for first 2-4 weeks of deployment
   - Cannot help with new query patterns until they appear in training data
   - Difficult with ephemeral environments or rapidly changing workloads

6. **Drift Handling**: Model degrades over time as system changes
   - Adding new tables, indexes, or applications changes baseline
   - Requires retraining and validation each time

### Best For
- ✅ Large organizations with dedicated ML/data teams
- ✅ Mature systems with stable workloads and 6+ months history
- ✅ Situations where pattern discovery > causal understanding
- ✅ High-volume anomaly detection at scale (1000+ servers)

### Example: What It Would Detect
```
Raw Metrics:
  [12:00] CPU=45%, IO_latency=2ms, active_queries=50
  [12:05] CPU=52%, IO_latency=3ms, active_queries=55
  [12:10] CPU=68%, IO_latency=8ms, active_queries=75
  [12:15] CPU=85%, IO_latency=15ms, active_queries=120
  [12:20] CPU=92%, IO_latency=22ms, active_queries=150

ML Detects: "CPU and IO_latency correlation = 0.98, correlation shifted from baseline"
DBAs See: "Alert: CPU and IO are correlated differently than usual"
DBAs Need: Why? Is it lock contention? Missing index? Vacuum? Replication lag? → UNKNOWN
```

---

## Approach 2: Pure Graph-Based (Knowledge Graph + Rule Engine)

### What It Is
Build an explicit knowledge graph of PostgreSQL relationships and run rule-based queries to detect anomalies and trace causality.

**Techniques**:
- Knowledge graph of PostgreSQL entities (tables, indexes, queries, locks, connections)
- Entity relationships (query → uses_index → on_table, table → has_lock → connection)
- Rule engine with Prolog-like queries or graph pattern matching
- Causal graphs (missing_index → full_table_scan → high_io → cpu_spike)
- Temporal reasoning (event A happened 2 min before B)

### Strengths ✅

1. **Explainability**: Every conclusion has a traceable path
   - Alert: "High CPU (92%) caused by missing index on users.email → full table scan (5M rows) → IO spike → blocking other queries"
   - Path: `query_slow` ← `missing_index` ← `table_users` ← `full_table_scan`
   - DBAs immediately understand and can fix

2. **Causality**: Distinguishes causation from correlation
   - Graph explicitly encodes: "missing index CAUSES full table scan CAUSES high IO"
   - Not just "correlation = 0.98"
   - Can validate against domain knowledge

3. **Deterministic**: No randomness, fully auditable logic
   - Every alert can be reproduced and explained
   - Regulatory/compliance teams can audit the reasoning
   - No "the model thinks..." uncertainty

4. **No Training Required**: Works immediately with PostgreSQL knowledge
   - Use best practices encoded as rules
   - Doesn't need historical data baseline
   - Useful on day 1

5. **Customizable**: Easy to add domain-specific rules
   - "If vacuum_duration > 30min AND active_connections > 100, then potential blocking"
   - "If query_plan contains SeqScan on large table, recommend index"
   - New rules added as business learns

6. **Real-time & Lightweight**: Graph queries are fast
   - Checking relationship paths: <5ms
   - No model loading/inference overhead
   - Can run on modest hardware

### Weaknesses ❌

1. **Manual Relationship Definition**: Must encode all relationships upfront
   - Every correlation must be explicitly modeled
   - Easy to miss subtle relationships
   - Requires domain expertise to define correctly

2. **Knowledge Graph Maintenance**: Graph must be kept accurate
   - Schema changes → graph updates
   - New indexes created → graph updates
   - Relationships between entities → manual maintenance

3. **Cannot Discover New Patterns**: Only finds what's explicitly modeled
   - If you didn't think to model "vacuum+high_io correlation", you won't find it
   - Reactive (fixes known problems) vs proactive (discovers new issues)

4. **Scalability Challenges**: Graphs can become massive
   - PostgreSQL with 10,000+ tables → graph has millions of nodes/edges
   - Graph queries become expensive at scale
   - Need sophisticated graph DB and indexing

5. **Temporal Reasoning Complexity**: Handling time-lagged correlations is difficult
   - "IO spike happens 2-5 minutes AFTER missing index detection" → hard to express
   - Requires complex temporal logic programming

6. **False Negatives**: Miss issues that don't match pre-defined patterns
   - New type of issue appears → no rule exists → goes undetected
   - Degraded until someone manually adds the rule

### Best For
- ✅ Regulatory/compliance requirements (need explainability)
- ✅ Well-understood domains (PostgreSQL is mature, relationships well-known)
- ✅ Small-to-medium deployments (manageable graph size)
- ✅ Critical systems where false positives are expensive
- ✅ Initial MVP or POC (works immediately)

### Example: What It Would Detect
```
Knowledge Graph:
  query_slow_users_query
    → [uses_table] → users
    → [has_index] → users_pkey (no email index)
    → [causes_full_scan] → SeqScan(5M rows)
    → [causes_io] → disk_io_latency
    → [causes_blocking] → other_queries_wait

Rule Engine:
  IF query IS SeqScan AND table_size > 1M THEN missing_index
  IF missing_index THEN recommend_create_index(table, columns)
  IF SeqScan(5M rows) AND high_io THEN io_bottleneck

Alert Generated:
  [ROOT CAUSE] Missing index on users(email)
  [CONSEQUENCE] SeqScan reading 5M rows
  [IMPACT] IO latency 22ms, blocking 47 queries
  [RECOMMENDATION] CREATE INDEX users_email_idx ON users(email)
  [Confidence] 100% (deterministic rule match)
```

---

## Approach 3: Hybrid (Graph + Lightweight AI)

### What It Is
**Primary**: Rule-based graph engine for known correlations and causality
**Secondary**: Lightweight ML models for pattern discovery and anomaly detection
**Integration**: Graph provides structure, ML refines thresholds and learns exceptions

**Techniques**:
- Knowledge graph for PostgreSQL structure and known relationships
- Statistical anomaly detection (Z-score, IQR) for metric-level anomalies
- Correlation analysis (Pearson) to discover new relationships
- Graph pattern matching to validate discovered correlations
- Rules to trace causality

### Strengths ✅

1. **Best of Both Worlds**:
   - Graph provides **explainability** + **causality** (critical for production)
   - ML provides **pattern discovery** + **adaptive learning** (finds hidden issues)
   - Together: "We detected this pattern, validated it against known causality, here's why"

2. **Explainable Anomalies**: AI detects, graph explains
   - ML: "IO latency is 3σ above baseline"
   - Graph: "Because: missing_index → full_table_scan → high_io"
   - DBAs understand both "what" and "why"

3. **Practical Training Path**: No cold start
   - Day 1: Graph rules work immediately
   - Week 1: ML models train in background
   - Week 2: ML starts providing pattern refinement
   - Graceful degradation if ML fails (graph still works)

4. **Reduced False Positives**: Graph validates ML findings
   - ML detects correlation: `metric_A ↔ metric_B`
   - Graph checks: "Do these metrics have causal path?" → validates
   - Or: "Is this a known spurious correlation?" → filters
   - Only alerts if validated

5. **Continuous Learning with Safety**: ML learns, graph supervises
   - ML: "I noticed queries ending with EXCEPT are 40% slower"
   - Graph: "Check if this is validated by optimizer knowledge"
   - Domain expert: Approves the rule or rejects it
   - Added to graph only if validated

6. **Scalable Both Ways**:
   - Can start small (10 rules) and grow graph gradually
   - Can start with 1 ML model (anomaly detection) and add more later
   - Each component is independently scalable

### Weaknesses ⚠️

1. **Complexity**: More moving parts to maintain
   - Graph relationships to define
   - ML models to train and monitor
   - Integration logic between them
   - Requires broader team skills

2. **Hybrid Validation Overhead**: Must define when each is used
   - Which issues are graph-only? Which need ML validation?
   - How much confidence from ML is needed?
   - Rules for combining evidence from both systems

3. **More Data Required**: Graph + ML both need configuration
   - Initial setup more involved (define rules + train models)
   - More test data needed to validate hybrid logic

### Best For
- ✅ **Production PostgreSQL monitoring (RECOMMENDED for pganalytics)**
- ✅ Organizations wanting both explainability AND pattern discovery
- ✅ Teams with moderate ML experience (not experts, but not beginners)
- ✅ Systems that need to grow from MVP to mature platform
- ✅ Situations where false positives have high cost (production databases)

### Example: What It Would Detect
```
Stage 1 - Graph detects known issue:
  Schema change: users.email index created → graph updated
  Query monitoring: seqscan detected on users → triggers rule
  Rule matches: "seqscan on users_pkey" → missing email index previously
  Alert: "Missing index on users(email) detected in slow queries"

Stage 2 - ML detects emerging pattern:
  Historical: connection_count and io_latency usually uncorrelated (r=-0.05)
  Current: connection_count ↑ 200% AND io_latency ↑ 300% (r=0.87)
  ML Alert: "Anomaly: new correlation detected, 3σ deviation"

Stage 3 - Hybrid validation:
  Graph checks: "Does high connections CAUSE high IO?"
  Path found: connections → contention_for_locks → lock_wait → io_stall
  Validates: ML finding matches causal path
  Confidence: 95% (ML evidence 85%, graph validation +10%)

Final Alert:
  [SEVERITY] High
  [TYPE] Performance degradation
  [DETECTED BY] ML (correlation spike) + Graph (causal validation)
  [ROOT CAUSE] Lock contention from 200+ concurrent connections
  [EVIDENCE]
    - Connections jumped 200% (statistical anomaly)
    - Causal path confirmed: connections→locks→io_wait
    - Correlates with io_latency spike (r=0.87)
  [RECOMMENDATION] Implement connection pooling (pgBouncer)
  [CONFIDENCE] 95%
```

---

## Detailed Comparison Matrix

| Criterion | Pure AI | Pure Graph | Hybrid |
|-----------|---------|-----------|--------|
| **Explainability** | ⭐⭐ (low) | ⭐⭐⭐⭐⭐ (high) | ⭐⭐⭐⭐⭐ (high) |
| **Causality Detection** | ⭐⭐ (probabilistic) | ⭐⭐⭐⭐⭐ (explicit) | ⭐⭐⭐⭐⭐ (explicit) |
| **Pattern Discovery** | ⭐⭐⭐⭐⭐ (autonomous) | ⭐⭐ (manual only) | ⭐⭐⭐⭐ (both) |
| **Day 1 Usability** | ❌ (needs training) | ✅ (immediate) | ✅ (immediate) |
| **False Positives** | ⭐⭐⭐ (moderate) | ⭐⭐⭐⭐ (low) | ⭐⭐⭐⭐⭐ (lowest) |
| **Implementation Complexity** | ⭐⭐⭐⭐ (moderate-high) | ⭐⭐ (low-moderate) | ⭐⭐⭐ (moderate) |
| **Maintenance Burden** | ⭐⭐⭐⭐ (continuous retraining) | ⭐⭐⭐ (schema tracking) | ⭐⭐⭐ (balanced) |
| **ML Expertise Required** | ⭐⭐⭐⭐⭐ (high) | ❌ (none) | ⭐⭐ (low-moderate) |
| **Scalability to 1000s Servers** | ⭐⭐⭐⭐⭐ (excellent) | ⭐⭐ (poor) | ⭐⭐⭐⭐ (good) |
| **Real-time Performance** | ⭐⭐⭐⭐ (once trained) | ⭐⭐⭐⭐⭐ (excellent) | ⭐⭐⭐⭐⭐ (excellent) |
| **Regulatory/Compliance** | ⭐⭐ (audit trail weak) | ⭐⭐⭐⭐⭐ (full audit) | ⭐⭐⭐⭐⭐ (full audit) |
| **Cost to Implement** | 💰💰💰 (high) | 💰 (low) | 💰💰 (moderate) |

---

## PostgreSQL-Specific Considerations

### Why Graph Wins for PostgreSQL
1. **Well-defined domain**: PostgreSQL structure and best practices are well-documented
2. **Relationships are known**: Indexes → query performance, locks → contention, vacuum → autovacuum bloat
3. **Causality is deterministic**: "Missing index" ALWAYS causes sequential scan (not probabilistic)
4. **Must have explainability**: DBA must understand why an alert fired to fix production issue

### Why ML Adds Value
1. **Workload-specific patterns**: Each application has unique query patterns and baseline
2. **Anomaly detection**: Threshold-based rules fail; statistical anomaly detection works better
3. **Temporal correlations**: "Vacuum 5 minutes ago caused slow queries now" requires learning
4. **Drift handling**: System changes (new schema, new app) automatically adjusted by ML

### Why Hybrid is Perfect for pganalytics
1. **Launch quickly**: Graph rules work immediately (day 1)
2. **Mature gradually**: Add ML capabilities as platform grows
3. **Production-safe**: Graph prevents ML false positives in critical system
4. **Competitive advantage**: "Explainable AI" beats both pure approaches

---

## Recommended Implementation: Hybrid Approach

### Phase 1: Graph Foundation (Week 1-2)
**Goal**: Deploy production-ready correlation engine on day 1

**What to Build**:
```
1. PostgreSQL Knowledge Graph
   - Entities: tables, indexes, queries, connections, locks, autovacuum
   - Relationships: has_index, depends_on, locks, is_blocking, uses_disk_io

2. Rule Engine (50-100 rules)
   - Missing index rules (detect seqscans on large tables)
   - Lock contention rules (high wait_time → blocking chains)
   - Autovacuum rules (vacuum_delay → bloat → seqscan)
   - Connection rules (too_many_connections → lock_contention)
   - Replication rules (replication_lag → query_slowness)

3. Correlation Engine
   - Template correlations: missing_index → seqscan → high_io
   - Lock cascade: connection_spike → lock_queue → blocked_queries
   - Autovacuum impact: table_growth → vacuum_duration → io_spike
```

**Code Structure**:
```go
// backend/internal/analytics/graph/
├── entity.go              // Node types (Table, Index, Query, etc.)
├── relationship.go        // Edge types (HasIndex, Depends, Blocks, etc.)
├── graph.go              // Graph construction and traversal
└── rules.go              // Rule engine implementation

// backend/internal/analytics/correlation/
├── detector.go           // Correlation detection
├── causal_path.go        // Find causal paths in graph
└── explainer.go          // Generate human-readable explanations
```

**Deliverable**: Explainable root cause alerts for 15+ common PostgreSQL issues

### Phase 2: Lightweight ML (Week 3-4)
**Goal**: Add pattern discovery and threshold learning

**What to Add**:
```
1. Anomaly Detection
   - Z-score based (how many σ from baseline?)
   - Dynamic baseline (learn what's normal per hour/day)
   - Per-metric anomalies (CPU spike vs IO spike handled differently)

2. Correlation Discovery
   - Compute pairwise correlations between all metrics
   - Detect new correlations not in graph
   - Validate discovered correlations via causal path checking

3. Threshold Learning
   - ML learns "optimal" alert threshold for each metric
   - Replaces hardcoded rules like "CPU > 80%"
   - Adapts to workload patterns
```

**Code Structure**:
```go
// backend/internal/analytics/ml/
├── anomaly_detector.go       // Z-score + dynamic baselines
├── correlation_calculator.go // Pearson correlation
├── threshold_learner.go      // Optimal thresholds
└── pattern_validator.go      // Validate against graph
```

**Integration**:
```go
// Hybrid validation
type HybridAnalyzer struct {
  graph        *Graph
  mlDetector   *AnomalyDetector
  correlations *CorrelationEngine
}

func (h *HybridAnalyzer) Analyze(metrics MetricSnapshot) []Alert {
  // Stage 1: Graph detects known issues
  graphAlerts := h.graph.DetectIssues(metrics)

  // Stage 2: ML detects anomalies
  mlAnomalies := h.mlDetector.Detect(metrics)

  // Stage 3: Validate via causal paths
  for _, anomaly := range mlAnomalies {
    if path := h.graph.FindCausalPath(anomaly); path != nil {
      // Confirmed by graph → high confidence
      alert := anomaly.WithConfidence(0.95)
      graphAlerts = append(graphAlerts, alert)
    }
  }

  return graphAlerts
}
```

**Deliverable**: Adaptive anomaly detection with <5% false positive rate

### Phase 3: Advanced Features (Weeks 5-6)
- Time-lagged correlation detection (vacuum 5min ago → queries slow now)
- Causal inference models (which metric CAUSES slowness)
- Forecasting (predict CPU spike 5 minutes in advance)
- User-defined rules (DBAs can add custom graph rules)

---

## Architecture Diagram: Hybrid Approach

```
┌──────────────────────────────────────────────────────────────┐
│               PostgreSQL Metrics Stream                      │
│  (CPU, IO latency, query count, lock waits, connection count)│
└────────────────────────┬─────────────────────────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
    ┌───▼───┐      ┌──────▼──────┐  ┌────▼──────┐
    │ Graph │      │ ML Anomaly  │  │ Baseline  │
    │Engine │      │ Detector    │  │ Learning  │
    │       │      │             │  │           │
    │ Rules │      │ Z-score:    │  │ Dynamic   │
    │ 50-100│      │ is 15σ dev? │  │ thresholds│
    └───┬───┘      └──────┬──────┘  └────┬──────┘
        │                │               │
        └────────────────┼───────────────┘
                         │
                    ┌────▼──────────────┐
                    │ Correlation       │
                    │ Discovery Layer   │
                    │ (Validate against │
                    │  causal paths)    │
                    └────┬──────────────┘
                         │
              ┌──────────┴──────────┐
              │                     │
         ┌────▼────┐         ┌──────▼────┐
         │ Graph   │         │ ML        │
         │ Explainer         │ Confidence│
         │         │         │ Score     │
         └────┬────┘         └──────┬────┘
              │                     │
              └──────────┬──────────┘
                         │
              ┌──────────▼──────────┐
              │ Alert with         │
              │ - Root cause       │
              │ - Causal path      │
              │ - Confidence       │
              │ - Recommendations  │
              └────────────────────┘
```

---

## Implementation Recommendation for pganalytics-v3

### Recommended Path: **Hybrid Approach**

**Rationale**:
1. ✅ **Production-ready day 1** (graph rules work immediately)
2. ✅ **Explainability** (DBAs understand every alert)
3. ✅ **Pattern discovery** (ML finds hidden correlations)
4. ✅ **Competitive advantage** (explainable AI ≫ black-box ML)
5. ✅ **Team friendly** (doesn't require ML experts)
6. ✅ **Growth path** (can add more ML features later)
7. ✅ **False positive control** (graph validates ML)

### Why Not Pure Approaches

**Pure AI**:
- ❌ 2-4 week delay before useful (need training data)
- ❌ Black box (DBAs won't trust alerts they can't understand)
- ❌ Overkill for well-known PostgreSQL relationships

**Pure Graph**:
- ❌ Can't discover new patterns DBAs didn't think of
- ❌ Manual maintenance as schema evolves
- ❌ Reactive (fixes known issues, misses new ones)

### Timeline
- **Week 1-2**: Deploy Phase 1 (graph + 50-100 rules) → production ready
- **Week 3-4**: Add Phase 2 (ML anomaly detection) → enhanced detection
- **Week 5-6**: Add Phase 3 (advanced features) → competitive differentiation

### Success Metrics
- **Day 1**: 15+ known PostgreSQL issues detected with 100% confidence
- **Week 2**: <2% false positive rate
- **Week 4**: +20% new issues discovered by ML (not in original rules)
- **Week 6**: <100ms alert latency, 95%+ confidence scoring

---

## Code Example: Hybrid System Structure

```go
// backend/internal/analytics/hybrid_analyzer.go

package analytics

import (
  "github.com/torresglauco/pganalytics-v3/backend/internal/analytics/graph"
  "github.com/torresglauco/pganalytics-v3/backend/internal/analytics/ml"
)

type HybridAnalyzer struct {
  // Graph-based correlation
  causalGraph *graph.PostgresGraph
  ruleEngine  *graph.RuleEngine

  // ML-based anomaly detection
  anomalyDetector *ml.AnomalyDetector
  correlationML   *ml.CorrelationEngine

  logger *zap.Logger
}

// Analyze performs hybrid correlation analysis
func (h *HybridAnalyzer) Analyze(
  ctx context.Context,
  metrics *MetricSnapshot,
) ([]CorrelationAlert, error) {

  // Phase 1: Detect known issues via graph rules
  graphAlerts := h.detectViaGraph(metrics)

  // Phase 2: Detect anomalies via ML
  mlAlerts := h.detectViaML(metrics)

  // Phase 3: Validate and merge
  validatedAlerts := h.validateAndMerge(graphAlerts, mlAlerts)

  return validatedAlerts, nil
}

// detectViaGraph finds known correlations in graph
func (h *HybridAnalyzer) detectViaGraph(
  metrics *MetricSnapshot,
) []CorrelationAlert {

  alerts := []CorrelationAlert{}

  // Rule 1: Missing index detection
  if metrics.SequentialScans > 0 && metrics.TableSize > 1_000_000 {
    // Trace causal path in graph
    path := h.causalGraph.FindPath(
      "sequential_scan",
      "missing_index",
    )

    if path != nil {
      alerts = append(alerts, CorrelationAlert{
        Type:        "MissingIndex",
        Severity:    "high",
        Confidence:  1.0, // 100% from rule
        CausalPath:  path,
        Explanation: fmt.Sprintf(
          "Sequential scan detected on %d-row table - likely missing index",
          metrics.TableSize,
        ),
      })
    }
  }

  // Rule 2: Lock contention detection
  if metrics.LockWaitTime > 100*time.Millisecond &&
     metrics.ActiveConnections > 50 {

    path := h.causalGraph.FindPath(
      "high_connections",
      "lock_contention",
    )

    if path != nil {
      alerts = append(alerts, CorrelationAlert{
        Type:        "LockContention",
        Severity:    "critical",
        Confidence:  1.0, // 100% from rule
        CausalPath:  path,
        Explanation: "High connection count causing lock contention",
      })
    }
  }

  return alerts
}

// detectViaMLC detects anomalies via ML
func (h *HybridAnalyzer) detectViaMLC(
  metrics *MetricSnapshot,
) []CorrelationAlert {

  alerts := []CorrelationAlert{}

  // Anomaly detection: is CPU 3σ above baseline?
  cpuAnomaly := h.anomalyDetector.Detect("cpu_usage", metrics.CPU)
  if cpuAnomaly != nil {
    alerts = append(alerts, CorrelationAlert{
      Type:       "AnomalyDetected",
      Severity:   "medium",
      Confidence: cpuAnomaly.Confidence, // 60-95% from statistics
      Explanation: fmt.Sprintf(
        "CPU usage is %.1f standard deviations above baseline (%.1f%% vs %.1f%%)",
        cpuAnomaly.StdDevs,
        metrics.CPU,
        cpuAnomaly.BaselineAvg,
      ),
    })
  }

  // Correlation discovery: find new correlations
  newCorrelations := h.correlationML.DiscoverCorrelations(metrics)
  for _, corr := range newCorrelations {
    if corr.Coefficient > 0.7 { // Pearson correlation > 0.7
      alerts = append(alerts, CorrelationAlert{
        Type:       "NewCorrelation",
        Severity:   "low",
        Confidence: 0.7,
        Explanation: fmt.Sprintf(
          "New correlation detected: %s ↔ %s (r=%.2f)",
          corr.MetricA,
          corr.MetricB,
          corr.Coefficient,
        ),
      })
    }
  }

  return alerts
}

// validateAndMerge combines graph and ML alerts
func (h *HybridAnalyzer) validateAndMerge(
  graphAlerts, mlAlerts []CorrelationAlert,
) []CorrelationAlert {

  final := []CorrelationAlert{}

  // Add all graph alerts (100% confidence)
  final = append(final, graphAlerts...)

  // Validate ML alerts via graph
  for _, mlAlert := range mlAlerts {
    // Check if causal path exists in graph
    if path := h.causalGraph.FindCausalPath(mlAlert); path != nil {
      // ML finding validated by graph → boost confidence
      mlAlert.Confidence = 0.95
      mlAlert.CausalPath = path
      mlAlert.ValidationSource = "graph"
      final = append(final, mlAlert)
    } else if mlAlert.Confidence > 0.85 {
      // High-confidence ML finding even without graph validation
      final = append(final, mlAlert)
    }
    // Else: discard low-confidence ML findings without graph support
  }

  return final
}
```

---

## Conclusion

**For pganalytics-v3, implement the Hybrid approach**:
- **Graph as primary** (explainability, causality)
- **ML as secondary** (pattern discovery, threshold learning)
- **Validation layer** between them (reduce false positives)

This combines the strengths of both approaches while mitigating their weaknesses, creating a production-ready monitoring system that is both intelligent and transparent.

---

**Next Step**: Proceed with Phase 4.5.12 implementation using this hybrid architecture.
