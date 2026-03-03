# Guia de Implementação Técnica Detalhada
## Como Construir o Melhor PostgreSQL Monitoring Tool

**Data**: 3 de março de 2026
**Objetivo**: Plano técnico passo-a-passo para implementação
**Status**: Pronto para execução imediata

---

## PARTE 1: ARQUITETURA GERAL PROPOSTA

### 1.1 Stack Tecnológico (Recomendado)

```
FRONTEND
├─ React 18+ (já tem)
├─ TypeScript (já tem)
├─ Tailwind CSS (já tem)
├─ Recharts para charts (já tem)
│
├─ ADICIONAR:
├─ Zustand para state management (vs current)
├─ React Query para data fetching
├─ Visx para advanced visualizations
├─ Monaco Editor para query/plan viewing
├─ D3.js para custom graphs (lock contention)
├─ WebSocket para real-time updates
└─ Plotly.js para flame graphs

BACKEND
├─ Go 1.21+ (já tem)
├─ PostgreSQL 13+ (já tem)
├─ TimescaleDB 2.10+ (já tem)
│
├─ ADICIONAR:
├─ goreleaser para releases
├─ Grafana Loki para logs
├─ Prometheus para metrics interno
├─ WebSocket server (gorilla/websocket)
├─ gRPC para collector communication
├─ Machine Learning (goml ou scikit-learn via Python bridge)
└─ pgx driver (melhor performance)

INFRASTRUCTURE
├─ Docker & Docker Compose (production-ready)
├─ Kubernetes manifests (Helm charts)
├─ Terraform modules (AWS, GCP, Azure)
├─ GitHub Actions para CI/CD
├─ SonarQube para code quality
└─ Dependabot para security updates

DATA LAYER
├─ TimescaleDB (séries temporais)
├─ Redis (cache + real-time updates)
├─ PostgreSQL (metadata, users, config)
├─ Prometheus (internal metrics)
└─ Grafana Loki (centralized logs)
```

### 1.2 Estrutura de Diretórios Proposta

```
pganalytics-v3/
├── backend/
│   ├── cmd/
│   │   ├── pganalytics-api/
│   │   │   └── main.go
│   │   ├── pganalytics-collector/      [NEW]
│   │   │   └── main.go
│   │   └── pganalytics-cli/            [NEW]
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── auth/ (já tem)
│   │   ├── storage/ (já tem)
│   │   ├── config/ (já tem)
│   │   ├── cache/ (já tem)
│   │   ├── crypto/ (já tem)
│   │   ├── timescale/ (melhorar)
│   │   ├── metrics/ (expandir)
│   │   │   ├── collector.go           [NEW]
│   │   │   ├── query_stats.go         [NEW]
│   │   │   ├── lock_stats.go          [NEW]
│   │   │   ├── bloat_stats.go         [NEW]
│   │   │   ├── index_stats.go         [NEW]
│   │   │   ├── connection_stats.go    [NEW]
│   │   │   ├── cache_stats.go         [NEW]
│   │   │   └── replication_stats.go   [NEW]
│   │   │
│   │   ├── ml/                         [EXPAND]
│   │   │   ├── features.go
│   │   │   ├── baseline.go            [NEW]
│   │   │   ├── anomaly.go             [NEW]
│   │   │   ├── predict.go             [NEW]
│   │   │   └── model.go               [NEW]
│   │   │
│   │   ├── alerts/                     [NEW]
│   │   │   ├── rules.go
│   │   │   ├── engine.go
│   │   │   ├── notifications.go
│   │   │   ├── slack.go
│   │   │   ├── pagerduty.go
│   │   │   ├── email.go
│   │   │   └── webhooks.go
│   │   │
│   │   ├── recommendations/            [NEW]
│   │   │   ├── query_opt.go
│   │   │   ├── index_opt.go
│   │   │   ├── bloat_cleanup.go
│   │   │   ├── lock_resolution.go
│   │   │   └── cache_tuning.go
│   │   │
│   │   ├── automation/                 [NEW]
│   │   │   ├── executor.go
│   │   │   ├── safety.go
│   │   │   ├── approval.go
│   │   │   └── audit.go
│   │   │
│   │   └── api/                        [NEW]
│   │       ├── handlers.go
│   │       ├── middleware.go
│   │       ├── errors.go
│   │       └── schemas.go
│   │
│   ├── pkg/
│   │   ├── postgres/                   [NEW]
│   │   │   ├── client.go
│   │   │   ├── pool.go
│   │   │   └── queries.go
│   │   │
│   │   ├── timescale/                  [NEW]
│   │   │   ├── schema.go
│   │   │   └── queries.go
│   │   │
│   │   └── utils/                      [NEW]
│   │       ├── logging.go
│   │       ├── metrics.go
│   │       └── helpers.go
│   │
│   └── tests/
│       ├── integration/                [NEW]
│       ├── load/                       [NEW]
│       └── fixtures/                   [NEW]
│
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── common/ (já tem)
│   │   │   ├── cards/ (já tem)
│   │   │   ├── charts/ (já tem)
│   │   │   ├── tables/ (já tem)
│   │   │   ├── forms/ (já tem)
│   │   │   │
│   │   │   ├── visualization/          [NEW]
│   │   │   │   ├── FlameGraph.tsx
│   │   │   │   ├── LockGraph.tsx
│   │   │   │   ├── DependencyGraph.tsx
│   │   │   │   ├── Heatmap.tsx
│   │   │   │   └── Timeline.tsx
│   │   │   │
│   │   │   ├── analysis/               [NEW]
│   │   │   │   ├── QueryAnalyzer.tsx
│   │   │   │   ├── LockAnalyzer.tsx
│   │   │   │   ├── BloatAnalyzer.tsx
│   │   │   │   └── IndexAnalyzer.tsx
│   │   │   │
│   │   │   └── dashboard/              [NEW]
│   │   │       ├── DashboardBuilder.tsx
│   │   │       ├── DashboardTemplate.tsx
│   │   │       └── WidgetLibrary.tsx
│   │   │
│   │   ├── pages/ (já tem)
│   │   │   └─ [Implementar todas 10]
│   │   │
│   │   ├── services/
│   │   │   ├── api.ts (já tem)
│   │   │   ├── websocket.ts            [NEW]
│   │   │   ├── realtime.ts             [NEW]
│   │   │   └── cache.ts                [NEW]
│   │   │
│   │   ├── hooks/                      [NEW]
│   │   │   ├── useMetrics.ts
│   │   │   ├── useAlerts.ts
│   │   │   ├── useRecommendations.ts
│   │   │   ├── useWebSocket.ts
│   │   │   └── useLocalStorage.ts
│   │   │
│   │   ├── store/                      [NEW]
│   │   │   ├── metricsStore.ts
│   │   │   ├── alertsStore.ts
│   │   │   ├── filtersStore.ts
│   │   │   └── uiStore.ts (já tem)
│   │   │
│   │   ├── utils/
│   │   │   ├── constants.ts (já tem)
│   │   │   ├── formatting.ts (já tem)
│   │   │   ├── calculations.ts (já tem)
│   │   │   ├── charts.ts                [NEW]
│   │   │   └── analysis.ts             [NEW]
│   │   │
│   │   └── types/
│   │       ├── alerts.ts (já tem)
│   │       ├── metrics.ts              [NEW]
│   │       ├── recommendations.ts      [NEW]
│   │       └── visualization.ts        [NEW]
│   │
│   ├── public/
│   │   ├── docs/                       [NEW]
│   │   └── schemas/                    [NEW]
│   │
│   └── tests/                          [NEW]
│       ├── unit/
│       ├── integration/
│       └── e2e/
│
├── collector/                          [EXISTING - manter]
│   ├── src/
│   ├── include/
│   ├── config/
│   └── tests/
│
├── docs/
│   ├── ARCHITECTURE.md                 [NEW]
│   ├── API.md                          [NEW]
│   ├── DEPLOYMENT.md                   [NEW]
│   ├── DEVELOPMENT.md                  [NEW]
│   ├── ML_MODELS.md                    [NEW]
│   ├── POSTGRESQL_SETUP.md             [NEW]
│   ├── MIGRATION_GUIDE.md              [NEW]
│   └── TROUBLESHOOTING.md              [NEW]
│
├── infrastructure/                     [NEW]
│   ├── docker/
│   │   ├── Dockerfile
│   │   ├── docker-compose.yml
│   │   └── docker-compose.prod.yml
│   │
│   ├── kubernetes/
│   │   ├── helm/
│   │   ├── manifests/
│   │   └── values/
│   │
│   ├── terraform/
│   │   ├── aws/
│   │   ├── gcp/
│   │   ├── azure/
│   │   └── modules/
│   │
│   └── scripts/
│       ├── setup.sh
│       ├── migrate.sh
│       └── backup.sh
│
├── .github/
│   ├── workflows/
│   │   ├── test.yml                    [EXPAND]
│   │   ├── build.yml                   [EXPAND]
│   │   ├── deploy.yml                  [EXPAND]
│   │   └── security.yml                [NEW]
│   │
│   └── ISSUE_TEMPLATE/
│       ├── bug.md
│       ├── feature.md
│       └── discussion.md
│
└── README.md (melhorar)
```

---

## PARTE 2: IMPLEMENTAÇÃO POR FUNCIONALIDADE

### 2.1 QUERY PERFORMANCE (Semanas 1-8)

#### A. Coleta de Dados (Backend)

```go
// internal/metrics/query_stats.go

package metrics

import (
    "context"
    "database/sql"
    "time"
)

type QueryStats struct {
    QueryID         string
    Query           string
    Fingerprint     string          // SHA256 hash
    Calls           int64
    TotalTime       float64         // milliseconds
    MeanTime        float64
    MaxTime         float64
    MinTime         float64
    StddevTime      float64
    RowsReturned    int64
    RowsScanned     int64
    BlkReadTime     float64
    BlkWriteTime    float64
    CollectedAt     time.Time
}

type QueryStatsCollector struct {
    db              *sql.DB
    interval        time.Duration
    collectorID     string
}

// Implementação de coleta
func (c *QueryStatsCollector) Collect(ctx context.Context) ([]QueryStats, error) {
    query := `
    SELECT
        query,
        calls,
        total_time,
        mean_time,
        max_time,
        min_time,
        stddev_time,
        rows,
        100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) as cache_ratio,
        blk_read_time,
        blk_write_time
    FROM pg_stat_statements
    WHERE query NOT LIKE '%pg_stat_statements%'
    ORDER BY total_time DESC
    LIMIT 1000
    `

    rows, err := c.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    stats := make([]QueryStats, 0)
    for rows.Next() {
        var s QueryStats
        if err := rows.Scan(&s.Query, &s.Calls, &s.TotalTime, &s.MeanTime, &s.MaxTime, &s.MinTime, &s.StddevTime, &s.RowsReturned, &s.BlkReadTime, &s.BlkWriteTime); err != nil {
            continue
        }
        s.Fingerprint = sha256Hash(normalizeQuery(s.Query))
        s.QueryID = generateID(s.Fingerprint, s.CollectorID)
        s.CollectedAt = time.Now()
        stats = append(stats, s)
    }

    return stats, rows.Err()
}

// Armazenar em TimescaleDB
func (c *QueryStatsCollector) Store(ctx context.Context, stats []QueryStats) error {
    // INSERT INTO metrics.query_stats
    // VALUES ($1, $2, $3, ...) ON CONFLICT UPDATE
}

// Auto-explain (safe, não-blocking)
func (c *QueryStatsCollector) ExplainPlan(ctx context.Context, query string) (string, error) {
    // SET auto_explain.log_analyze = off;
    // SET auto_explain.log_verbose = off;
    // SET auto_explain.log_min_duration = 1000; // 1 segundo
    // EXPLAIN (FORMAT JSON) SELECT ...
}
```

#### B. Schema TimescaleDB

```sql
-- Create hypertable for query stats
CREATE TABLE IF NOT EXISTS metrics.query_stats (
    time                TIMESTAMPTZ NOT NULL,
    collector_id        TEXT NOT NULL,
    query_id            TEXT NOT NULL,
    query               TEXT,
    fingerprint         TEXT,
    calls               BIGINT,
    total_time          FLOAT8,
    mean_time           FLOAT8,
    max_time            FLOAT8,
    min_time            FLOAT8,
    stddev_time         FLOAT8,
    rows_returned       BIGINT,
    rows_scanned        BIGINT,
    blk_read_time       FLOAT8,
    blk_write_time      FLOAT8,
    cache_ratio         FLOAT8
) PARTITION BY RANGE (time);

SELECT create_hypertable('metrics.query_stats', 'time', if_not_exists => TRUE);

-- Criar índices para queries rápidas
CREATE INDEX IF NOT EXISTS idx_query_stats_fingerprint
    ON metrics.query_stats (fingerprint, time DESC);

CREATE INDEX IF NOT EXISTS idx_query_stats_calls
    ON metrics.query_stats (calls DESC, time DESC);

-- Criar continuous aggregates para histórico
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.query_stats_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS hour,
    collector_id,
    fingerprint,
    AVG(mean_time) AS avg_mean_time,
    MAX(max_time) AS max_time,
    SUM(calls) AS total_calls
FROM metrics.query_stats
GROUP BY hour, collector_id, fingerprint;

-- Alertas on continuous aggregate refresh
SELECT add_continuous_aggregate_policy('metrics.query_stats_hourly',
    start_offset => INTERVAL '2 hours',
    schedule_interval => INTERVAL '1 hour');
```

#### C. API Endpoints

```go
// internal/api/handlers/query_handler.go

// GET /api/v1/queries
// List all unique queries with aggregation
func ListQueries(w http.ResponseWriter, r *http.Request) {
    // Query parameters: sort_by, order, limit, offset, search, time_range
    // Response: []QuerySummary with pagination
}

// GET /api/v1/queries/{fingerprint}
// Get detailed analysis for specific query
func GetQueryDetail(w http.ResponseWriter, r *http.Request) {
    // Response: QueryDetail {
    //   summary: QueryStats
    //   history: []QueryStats (last 7 days)
    //   plans: []ExplainPlan
    //   trends: Trends
    //   anomalies: []Anomaly
    //   recommendations: []Recommendation
    // }
}

// POST /api/v1/queries/{fingerprint}/analyze
// Trigger deeper analysis (auto-explain, plan comparison)
func AnalyzeQuery(w http.ResponseWriter, r *http.Request) {
    // Long-running operation, return task ID
}

// GET /api/v1/queries/trending/slow
// Get trending slow queries
func GetTrendingSlowQueries(w http.ResponseWriter, r *http.Request) {
}

// GET /api/v1/queries/anomalies
// Get queries with performance anomalies
func GetQueryAnomalies(w http.ResponseWriter, r *http.Request) {
}
```

#### D. Frontend Implementation

```typescript
// frontend/src/pages/QueryPerformance.tsx

import React, { useState, useEffect } from 'react';
import { useMetrics } from '../hooks/useMetrics';
import { LineChart, BarChart } from '../components/charts';
import { DataTable } from '../components/tables/DataTable';
import { MetricCard } from '../components/cards/MetricCard';
import { PageWrapper } from '../components/common/PageWrapper';

export const QueryPerformance: React.FC = () => {
    const [selectedQuery, setSelectedQuery] = useState<string | null>(null);
    const [timeRange, setTimeRange] = useState('24h');

    const { queries, loading, error } = useMetrics(
        `/api/v1/queries?time_range=${timeRange}`,
        1000 // refresh every 1 second
    );

    return (
        <PageWrapper title="Query Performance">
            {/* Summary Cards */}
            <div className="grid grid-cols-4 gap-4 mb-6">
                <MetricCard title="Total Queries" value={queries.length} />
                <MetricCard title="Avg Exec Time" value="45.2ms" />
                <MetricCard title="Slow Queries" value="23" status="warning" />
                <MetricCard title="Anomalies" value="5" status="critical" />
            </div>

            {/* Main Content */}
            <div className="grid grid-cols-3 gap-6">
                {/* Query List */}
                <div className="col-span-2 bg-white rounded-lg shadow p-6">
                    <h3 className="text-lg font-semibold mb-4">All Queries</h3>
                    <DataTable
                        data={queries}
                        columns={[
                            { key: 'fingerprint', label: 'Query Pattern' },
                            { key: 'calls', label: 'Calls' },
                            { key: 'total_time', label: 'Total Time' },
                            { key: 'mean_time', label: 'Avg Time' },
                            { key: 'max_time', label: 'Max Time' }
                        ]}
                        onRowClick={(row) => setSelectedQuery(row.fingerprint)}
                    />
                </div>

                {/* Query Detail Sidebar */}
                {selectedQuery && (
                    <div className="bg-white rounded-lg shadow p-6">
                        <h3 className="text-lg font-semibold mb-4">Analysis</h3>
                        <QueryDetailPanel fingerprint={selectedQuery} />
                    </div>
                )}
            </div>

            {/* Timeline Chart */}
            <div className="bg-white rounded-lg shadow p-6 mt-6">
                <h3 className="text-lg font-semibold mb-4">Execution Time Trend</h3>
                <LineChart
                    data={queries.history}
                    dataKey="mean_time"
                    name="Avg Execution Time"
                />
            </div>
        </PageWrapper>
    );
};

// QueryDetailPanel component
interface QueryDetailPanelProps {
    fingerprint: string;
}

const QueryDetailPanel: React.FC<QueryDetailPanelProps> = ({ fingerprint }) => {
    const { detail } = useMetrics(`/api/v1/queries/${fingerprint}`);

    if (!detail) return <div>Loading...</div>;

    return (
        <div className="space-y-4">
            {/* Query Text */}
            <div>
                <h4 className="font-semibold text-sm text-gray-600">Query</h4>
                <code className="bg-gray-50 p-2 rounded text-xs block overflow-x-auto">
                    {detail.query}
                </code>
            </div>

            {/* Performance Metrics */}
            <div className="grid grid-cols-2 gap-2">
                <div>
                    <p className="text-xs text-gray-600">Calls</p>
                    <p className="text-lg font-bold">{detail.calls}</p>
                </div>
                <div>
                    <p className="text-xs text-gray-600">Avg Time</p>
                    <p className="text-lg font-bold">{detail.mean_time}ms</p>
                </div>
            </div>

            {/* Recommendations */}
            <div>
                <h4 className="font-semibold text-sm text-gray-600 mb-2">Recommendations</h4>
                {detail.recommendations.map((rec, i) => (
                    <div key={i} className="text-sm p-2 bg-yellow-50 rounded mb-2">
                        <p className="font-semibold text-yellow-900">{rec.title}</p>
                        <p className="text-yellow-700">{rec.description}</p>
                    </div>
                ))}
            </div>

            {/* Execution Plans */}
            <div>
                <h4 className="font-semibold text-sm text-gray-600 mb-2">Execution Plans</h4>
                {detail.plans.map((plan, i) => (
                    <button
                        key={i}
                        className="w-full text-left p-2 border rounded mb-2 hover:bg-gray-50"
                        onClick={() => showPlanDetail(plan)}
                    >
                        <p className="text-xs text-gray-600">{plan.timestamp}</p>
                        <p className="text-sm">{plan.summary}</p>
                    </button>
                ))}
            </div>
        </div>
    );
};
```

#### E. ML Integration (Query Baseline & Anomaly)

```go
// internal/ml/query_anomaly.go

package ml

import (
    "time"
    "github.com/pganalytics/pkg/stats"
)

type QueryBaseline struct {
    Fingerprint string
    MeanTime    float64
    StdDev      float64
    P95Time     float64
    P99Time     float64
    UpdatedAt   time.Time
}

type QueryAnomaly struct {
    Fingerprint string
    DetectedAt  time.Time
    Type        string      // "slower", "faster", "pattern_change"
    Score       float64     // 0-100
    Message     string
}

// Calculate baseline from historical data (first 7 days)
func CalculateQueryBaseline(stats []QueryStats) QueryBaseline {
    times := extractMeanTimes(stats)

    baseline := QueryBaseline{
        MeanTime: stats.Mean(times),
        StdDev:   stats.StdDev(times),
        P95Time:  stats.Percentile(times, 0.95),
        P99Time:  stats.Percentile(times, 0.99),
        UpdatedAt: time.Now(),
    }

    return baseline
}

// Detect anomalies in current query execution
func DetectQueryAnomaly(current QueryStats, baseline QueryBaseline) *QueryAnomaly {
    // Check if current execution time significantly deviates from baseline

    // Anomaly scoring:
    // > P99: Score 50-70 (warning)
    // > P99 + 2σ: Score 70-85 (critical)
    // > P99 + 3σ: Score 85-100 (severe)

    zScore := (current.MeanTime - baseline.MeanTime) / baseline.StdDev

    if zScore < 2 {
        return nil // Not anomalous
    }

    return &QueryAnomaly{
        Fingerprint: current.Fingerprint,
        DetectedAt: current.CollectedAt,
        Type: "slower",
        Score: min(100, 50 + (zScore * 10)),
        Message: fmt.Sprintf(
            "Query execution time is %.1f%% slower than baseline (%.2fms vs %.2fms)",
            (current.MeanTime/baseline.MeanTime - 1) * 100,
            current.MeanTime,
            baseline.MeanTime,
        ),
    }
}

// Predict query performance impact
func PredictQueryOptimizationImpact(query string, recommendation Recommendation) (ImpactPrediction, error) {
    // Use historical data + ML model to predict:
    // - Query speedup percentage (with 85%+ confidence)
    // - Benefit vs risk score
    // - Implementation cost estimate
}
```

---

### 2.2 LOCK CONTENTION (Semanas 3-4)

#### A. Lock Collection

```go
// internal/metrics/lock_stats.go

package metrics

import (
    "context"
    "database/sql"
    "time"
)

type LockEvent struct {
    LockID          string
    DatabaseID      uint32
    TableID         uint32
    TransactionID   uint64
    LockType        string          // AccessShareLock, RowExclusiveLock, etc
    Granted         bool
    PID             int32
    Query           string
    ApplicationName string
    CollectedAt     time.Time
}

type LockWait struct {
    BlockerPID      int32
    BlockerQuery    string
    WaiterPID       int32
    WaiterQuery     string
    WaitDuration    float64         // seconds
    LockType        string
    DetectedAt      time.Time
}

type LockStatsCollector struct {
    db              *sql.DB
    interval        time.Duration
    collectorID     string
}

// Coletar locks em tempo real
func (c *LockStatsCollector) CollectLocks(ctx context.Context) ([]LockEvent, error) {
    query := `
    SELECT
        l.locktype,
        l.database,
        l.relation,
        l.transactionid,
        l.mode,
        l.granted,
        l.pid,
        a.query,
        a.application_name
    FROM pg_locks l
    LEFT JOIN pg_stat_activity a ON l.pid = a.pid
    WHERE l.locktype IN ('relation', 'transactionid')
    `

    rows, err := c.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    locks := make([]LockEvent, 0)
    for rows.Next() {
        var lock LockEvent
        if err := rows.Scan(
            &lock.LockType, &lock.DatabaseID, &lock.TableID,
            &lock.TransactionID, &lock.LockType, &lock.Granted,
            &lock.PID, &lock.Query, &lock.ApplicationName,
        ); err != nil {
            continue
        }
        lock.LockID = generateID(lock.DatabaseID, lock.TableID, lock.PID)
        lock.CollectedAt = time.Now()
        locks = append(locks, lock)
    }

    return locks, rows.Err()
}

// Detectar blocking chains
func (c *LockStatsCollector) DetectBlockingChains(ctx context.Context) ([]LockWait, error) {
    query := `
    SELECT
        blocked_locks.pid AS blocked_pid,
        blocked_activity.query AS blocked_query,
        blocking_locks.pid AS blocking_pid,
        blocking_activity.query AS blocking_query,
        blocked_activity.query_start
    FROM pg_catalog.pg_locks blocked_locks
    JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
    JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
        AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
        AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
        AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
        AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
        AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
        AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
        AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
        AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
        AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
        AND blocking_locks.pid != blocked_locks.pid
    JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
    WHERE NOT blocked_locks.granted
    `

    rows, err := c.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    waits := make([]LockWait, 0)
    for rows.Next() {
        var wait LockWait
        var queryStart time.Time
        if err := rows.Scan(
            &wait.WaiterPID, &wait.WaiterQuery,
            &wait.BlockerPID, &wait.BlockerQuery,
            &queryStart,
        ); err != nil {
            continue
        }
        wait.WaitDuration = time.Since(queryStart).Seconds()
        wait.DetectedAt = time.Now()
        waits = append(waits, wait)
    }

    return waits, rows.Err()
}

// Armazenar em TimescaleDB
func (c *LockStatsCollector) StoreLocks(ctx context.Context, locks []LockEvent, waits []LockWait) error {
    // INSERT into metrics.lock_events
    // INSERT into metrics.lock_waits
}
```

#### B. Lock Analysis Engine

```go
// internal/metrics/lock_analysis.go

type LockGraph struct {
    Nodes       []Node
    Edges       []Edge
    CycleDetected bool
}

type Node struct {
    PID     int32
    Query   string
    Status  string      // "waiting", "holding"
}

type Edge struct {
    From    int32
    To      int32
    Type    string      // "waits_for", "blocks"
    Weight  float64     // wait duration
}

// Build lock dependency graph
func BuildLockGraph(locks []LockEvent, waits []LockWait) LockGraph {
    graph := LockGraph{
        Nodes: make([]Node, 0),
        Edges: make([]Edge, 0),
    }

    // Create nodes from locks
    pidMap := make(map[int32]bool)
    for _, lock := range locks {
        if !pidMap[lock.PID] {
            graph.Nodes = append(graph.Nodes, Node{
                PID: lock.PID,
                Query: lock.Query,
                Status: map[bool]string{true: "holding", false: "waiting"}[lock.Granted],
            })
            pidMap[lock.PID] = true
        }
    }

    // Create edges from wait relationships
    for _, wait := range waits {
        graph.Edges = append(graph.Edges, Edge{
            From: wait.WaiterPID,
            To: wait.BlockerPID,
            Type: "waits_for",
            Weight: wait.WaitDuration,
        })
    }

    // Detect cycles (deadlock potential)
    graph.CycleDetected = detectCycle(graph)

    return graph
}

// Predict deadlock risk
func PredictDeadlockRisk(graph LockGraph) (risk float64, reason string) {
    if graph.CycleDetected {
        return 1.0, "Deadlock cycle detected"
    }

    // Analyze pattern:
    // - Multiple waiters on same lock: 0.3-0.5
    // - Long wait chains: 0.2-0.4
    // - Frequent lock pattern changes: 0.1-0.3

    return risk, reason
}
```

#### C. Frontend Lock Visualization

```typescript
// frontend/src/components/visualization/LockGraph.tsx

import React, { useEffect, useRef } from 'react';
import * as d3 from 'd3';

interface LockGraphProps {
    nodes: Array<{ id: string; label: string; status: string }>;
    edges: Array<{ source: string; target: string; weight: number }>;
    onNodeClick?: (nodeId: string) => void;
}

export const LockGraph: React.FC<LockGraphProps> = ({ nodes, edges, onNodeClick }) => {
    const svgRef = useRef<SVGSVGElement>(null);

    useEffect(() => {
        if (!svgRef.current || nodes.length === 0) return;

        const width = svgRef.current.clientWidth;
        const height = 400;

        // Create D3 simulation
        const simulation = d3.forceSimulation(nodes)
            .force("link", d3.forceLink(edges).id((d: any) => d.id).distance(100))
            .force("charge", d3.forceManyBody().strength(-300))
            .force("center", d3.forceCenter(width / 2, height / 2));

        // Clear previous
        d3.select(svgRef.current).selectAll("*").remove();

        const svg = d3.select(svgRef.current)
            .attr('width', width)
            .attr('height', height);

        // Draw edges
        const links = svg.selectAll("line")
            .data(edges)
            .enter()
            .append("line")
            .attr("stroke", "#999")
            .attr("stroke-width", (d: any) => Math.sqrt(d.weight))
            .attr("marker-end", "url(#arrowhead)");

        // Draw nodes
        const nodeElements = svg.selectAll("circle")
            .data(nodes)
            .enter()
            .append("circle")
            .attr("r", 20)
            .attr("fill", (d: any) => d.status === "blocked" ? "#f43f5e" : "#06b6d4")
            .attr("cursor", "pointer")
            .on("click", (event, d) => onNodeClick?.(d.id));

        // Draw labels
        const labels = svg.selectAll("text")
            .data(nodes)
            .enter()
            .append("text")
            .text((d: any) => `PID ${d.id}`)
            .attr("font-size", "10px")
            .attr("text-anchor", "middle")
            .attr("dy", "0.3em");

        // Update positions
        simulation.on("tick", () => {
            links
                .attr("x1", (d: any) => d.source.x)
                .attr("y1", (d: any) => d.source.y)
                .attr("x2", (d: any) => d.target.x)
                .attr("y2", (d: any) => d.target.y);

            nodeElements
                .attr("cx", (d: any) => d.x)
                .attr("cy", (d: any) => d.y);

            labels
                .attr("x", (d: any) => d.x)
                .attr("y", (d: any) => d.y);
        });

    }, [nodes, edges]);

    return <svg ref={svgRef} style={{ width: '100%', border: '1px solid #eee' }} />;
};
```

---

### 2.3 TABLE BLOAT (Semanas 5-6)

```go
// internal/metrics/bloat_stats.go

type BloatMetrics struct {
    TableID         uint32
    TableName       string
    TotalSize       int64       // bytes
    LiveTuples      int64
    DeadTuples      int64
    BloatPercent    float64
    BloatSize       int64       // bytes
    LastVacuum      *time.Time
    LastAnalyze     *time.Time
    VacuumCount     int
    AutovacuumCount int
    CollectedAt     time.Time
}

// Collect bloat metrics (safe - uses estimates)
func (c *MetricsCollector) CollectBloatMetrics(ctx context.Context) ([]BloatMetrics, error) {
    query := `
    SELECT
        schemaname,
        tablename,
        pg_total_relation_size(schemaname||'.'||tablename)::bigint as total_size,
        n_live_tup,
        n_dead_tup,
        ROUND(100 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_ratio,
        last_vacuum,
        last_autovacuum,
        vacuum_count,
        autovacuum_count
    FROM pg_stat_user_tables
    WHERE n_dead_tup > 1000
    ORDER BY n_dead_tup DESC
    `

    // Parse results and build BloatMetrics
}

// Predictive bloat analysis
type BloatPrediction struct {
    TableID         uint32
    ProjectedBloat  float64
    TimeToThreshold time.Duration
    VacuumSize      int64
    VacuumDuration  time.Duration
    LockDuration    time.Duration
    Recommendation  string
}

func PredictBloatGrowth(metrics []BloatMetrics, history []BloatMetrics) []BloatPrediction {
    predictions := make([]BloatPrediction, 0)

    for _, metric := range metrics {
        // Find historical data for same table
        growth := calculateGrowthRate(metric, history)

        pred := BloatPrediction{
            TableID: metric.TableID,
            ProjectedBloat: metric.BloatPercent + growth.PercentPerDay*7,
            TimeToThreshold: time.Duration(
                (40.0 - metric.BloatPercent) / growth.PercentPerDay,
            ) * 24 * time.Hour,
            VacuumSize: metric.BloatSize,
            VacuumDuration: estimateVacuumDuration(metric),
            LockDuration: estimateLockDuration(metric),
            Recommendation: generateBloatRecommendation(metric, pred),
        }

        predictions = append(predictions, pred)
    }

    return predictions
}
```

---

## PARTE 3: IMPLEMENTAÇÃO DE ALERTING (Phase 5)

### 3.1 Alert Rule Engine

```go
// internal/alerts/rules.go

type AlertRule struct {
    ID              string
    Name            string
    Description     string
    Condition       string          // SQL query
    Threshold       float64
    Duration        time.Duration   // how long condition must be true
    Severity        string          // critical, warning, info
    Enabled         bool
    NotifyChannels  []string        // slack, pagerduty, email
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type AlertEngine struct {
    db              *sql.DB
    rules           map[string]*AlertRule
    activeAlerts    map[string]*Alert
    notifier        *AlertNotifier
}

// Execute alert rules
func (e *AlertEngine) EvaluateRules(ctx context.Context) error {
    for _, rule := range e.rules {
        if !rule.Enabled {
            continue
        }

        // Query condition
        result, err := e.queryCondition(ctx, rule.Condition)
        if err != nil {
            continue
        }

        // Check threshold
        if result > rule.Threshold {
            // Check duration (must be ongoing for rule.Duration)
            if e.isConditionPersistent(rule.ID, rule.Duration) {
                // Fire alert
                alert := &Alert{
                    ID: generateID(),
                    RuleID: rule.ID,
                    Severity: rule.Severity,
                    Status: "active",
                    FiredAt: time.Now(),
                    Value: result,
                }

                e.activeAlerts[alert.ID] = alert

                // Notify
                for _, channel := range rule.NotifyChannels {
                    e.notifier.Notify(ctx, channel, alert)
                }
            }
        }
    }

    return nil
}

// Pre-configured alert rules
func GetDefaultAlertRules() []AlertRule {
    return []AlertRule{
        {
            Name: "High Lock Contention",
            Condition: `SELECT COUNT(*) FROM metrics.lock_waits WHERE detected_at > NOW() - INTERVAL '5 minutes'`,
            Threshold: 10,
            Duration: 5 * time.Minute,
            Severity: "critical",
            NotifyChannels: []string{"pagerduty", "slack"},
        },
        {
            Name: "Table Bloat Critical",
            Condition: `SELECT MAX(bloat_percent) FROM metrics.bloat_metrics WHERE collected_at > NOW() - INTERVAL '1 hour'`,
            Threshold: 40,
            Duration: 30 * time.Minute,
            Severity: "warning",
            NotifyChannels: []string{"slack"},
        },
        {
            Name: "Query Performance Degradation",
            Condition: `SELECT AVG(mean_time) FROM metrics.query_stats WHERE collected_at > NOW() - INTERVAL '5 minutes'`,
            Threshold: 1000, // milliseconds
            Duration: 10 * time.Minute,
            Severity: "warning",
            NotifyChannels: []string{"slack"},
        },
        {
            Name: "Replication Lag Critical",
            Condition: `SELECT MAX(lag_seconds) FROM metrics.replication_stats WHERE collected_at > NOW() - INTERVAL '1 minute'`,
            Threshold: 5,
            Duration: 2 * time.Minute,
            Severity: "critical",
            NotifyChannels: []string{"pagerduty"},
        },
        {
            Name: "Low Cache Hit Ratio",
            Condition: `SELECT AVG(cache_ratio) FROM metrics.cache_stats WHERE collected_at > NOW() - INTERVAL '1 hour'`,
            Threshold: 0.85,
            Duration: 30 * time.Minute,
            Severity: "warning",
            NotifyChannels: []string{"slack"},
        },
    }
}
```

### 3.2 Notification Channels

```go
// internal/alerts/slack.go

type SlackNotifier struct {
    webhookURL string
    client     *http.Client
}

func (n *SlackNotifier) Send(ctx context.Context, alert *Alert) error {
    message := slack.Message{
        Text: alert.Title,
        Blocks: []slack.Block{
            slack.SectionBlock{
                Type: slack.MBTSection,
                TextBlock: slack.NewTextBlockObject(slack.MarkdownType,
                    fmt.Sprintf("*%s*\n%s\n*Severity:* %s\n*Value:* %.2f",
                        alert.Title, alert.Description, alert.Severity, alert.Value),
                    false, false),
                Accessory: slack.NewButtonBlockElement("", alert.ID, slack.NewTextBlockObject(slack.PlainTextType, "View Details", false, false)),
            },
        },
    }

    return n.send(ctx, message)
}

// internal/alerts/pagerduty.go

type PagerDutyNotifier struct {
    integrationKey string
    client         *http.Client
}

func (n *PagerDutyNotifier) Send(ctx context.Context, alert *Alert) error {
    event := pagerduty.Event{
        RoutingKey: n.integrationKey,
        Action: "trigger",
        Payload: pagerduty.EventPayload{
            Summary: alert.Title,
            Severity: alert.Severity,
            Source: "pgAnalytics",
            Timestamp: alert.FiredAt.Format(time.RFC3339),
            Custom: map[string]interface{}{
                "description": alert.Description,
                "value": alert.Value,
                "rule_id": alert.RuleID,
            },
        },
    }

    return n.send(ctx, event)
}

// internal/alerts/email.go

type EmailNotifier struct {
    smtpHost string
    smtpPort int
    username string
    password string
}

func (n *EmailNotifier) Send(ctx context.Context, alert *Alert, recipients []string) error {
    // Send HTML email with alert details
}
```

---

## PARTE 4: PRIORIZAÇÃO IMEDIATA (Próximas 2-4 semanas)

### Ordem Recomendada de Implementação

```
SEMANA 1: Foundation
├─ [ ] Expand TimescaleDB schema para todas as métricas
├─ [ ] Create metric collectors framework
├─ [ ] Implement query stats collection
└─ [ ] Create frontend hooks para real-time data

SEMANA 2: Query Performance
├─ [ ] Query collection implementation
├─ [ ] Query storage in TimescaleDB
├─ [ ] API endpoints para query listing/detail
├─ [ ] QueryPerformance.tsx page implementation
└─ [ ] Query baseline calculation

SEMANA 3: Lock & Bloat
├─ [ ] Lock metrics collection
├─ [ ] Lock analysis engine (dependency graph)
├─ [ ] Bloat metrics collection
├─ [ ] Bloat prediction model
├─ [ ] LockContention.tsx & TableBloat.tsx pages
└─ [ ] Lock graph visualization

SEMANA 4: Alerting Foundation
├─ [ ] Alert rule engine
├─ [ ] Slack integration (basic)
├─ [ ] Alert persistence
├─ [ ] Alert UI components
└─ [ ] Test end-to-end alert flow

CHECKLIST DE COMMITS
├─ [ ] Commit 1: "feat: Expand TimescaleDB schema for comprehensive metrics"
├─ [ ] Commit 2: "feat: Implement query performance collection and analysis"
├─ [ ] Commit 3: "feat: Add lock contention detection and visualization"
├─ [ ] Commit 4: "feat: Implement table bloat metrics and predictions"
├─ [ ] Commit 5: "feat: Add basic alert rule engine with Slack integration"
└─ [ ] Commit 6: "feat: Complete QueryPerformance, LockContention, TableBloat pages"
```

---

**Próxima Ação**: Começar pela Semana 1 imediatamente
**Duração Total**: 18 semanas para feature-complete
**Equipe Recomendada**: 3-4 full-stack engineers
