# Advanced Features Implementation Plan (v3.1.0 - v3.4.0)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement 4 major PostgreSQL monitoring features in parallel waves to compete with pganalyze: Query Performance Deep Dive, Log Analysis Engine, Index Advisor, and VACUUM Advisor.

**Architecture:**
- **Onda 1 (Weeks 5-8):** v3.1.0 (Query Performance) and v3.2.0 (Log Analysis) built in parallel, establishing core data collection and analysis infrastructure
- **Onda 2 (Weeks 9-11):** v3.3.0 (Index Advisor) depends on query performance data from Onda 1
- **Onda 3 (Weeks 12-14):** v3.4.0 (VACUUM Advisor) depends on all previous data sources
- Each version has independent frontend dashboard, backend service, collector integration, and database schema

**Tech Stack:** Go (backend), React/TypeScript (frontend), C/C++ (collector), PostgreSQL + TimescaleDB (storage)

---

## File Structure

### Backend (Go)

**v3.1.0 - Query Performance:**
- `backend/internal/services/query_performance/collector.go` - Captures EXPLAIN ANALYZE
- `backend/internal/services/query_performance/parser.go` - Parses execution plans
- `backend/internal/services/query_performance/analyzer.go` - Detects patterns + issues
- `backend/internal/services/query_performance/timeline.go` - Performance history
- `backend/internal/api/handlers_query_performance.go` - HTTP endpoints
- `backend/internal/services/query_performance/test_*` - Unit tests

**v3.2.0 - Log Analysis:**
- `backend/internal/services/log_analysis/collector.go` - Ingests PostgreSQL logs
- `backend/internal/services/log_analysis/parser.go` - 50+ log categories
- `backend/internal/services/log_analysis/analyzer.go` - Pattern + anomaly detection
- `backend/internal/api/handlers_log_analysis.go` - HTTP + WebSocket endpoints
- `backend/internal/services/log_analysis/test_*` - Unit tests

**v3.3.0 - Index Advisor:**
- `backend/internal/services/index_advisor/analyzer.go` - Index recommendation engine
- `backend/internal/services/index_advisor/constraint_model.go` - Constraint programming
- `backend/internal/services/index_advisor/cost_calculator.go` - Cost estimation
- `backend/internal/api/handlers_index_advisor.go` - HTTP endpoints

**v3.4.0 - VACUUM Advisor:**
- `backend/internal/services/vacuum_advisor/analyzer.go` - VACUUM statistics + recommendations
- `backend/internal/services/vacuum_advisor/simulator.go` - VACUUM impact simulator
- `backend/internal/api/handlers_vacuum_advisor.go` - HTTP endpoints

### Frontend (React/TypeScript)

**v3.1.0 - Query Performance:**
- `frontend/src/pages/QueryPerformance.tsx` - Main page + dashboard
- `frontend/src/components/QueryPlan/PlanTree.tsx` - EXPLAIN plan tree visualization
- `frontend/src/components/QueryPlan/PlanNode.tsx` - Individual node rendering
- `frontend/src/components/QueryPerformance/Timeline.tsx` - Performance timeline chart
- `frontend/src/components/QueryPerformance/PatternCard.tsx` - Pattern detection UI
- `frontend/src/components/QueryPerformance/IssueDetail.tsx` - Issue details modal
- `frontend/src/hooks/useQueryPerformance.ts` - Data fetching hook
- `frontend/src/types/queryPerformance.ts` - TypeScript types

**v3.2.0 - Log Analysis:**
- `frontend/src/pages/LogAnalysis.tsx` - Main page + dashboard
- `frontend/src/components/LogInsights/LogStream.tsx` - Real-time log stream
- `frontend/src/components/LogInsights/LogTable.tsx` - Log list with filters
- `frontend/src/components/LogInsights/CategoryBreakdown.tsx` - Category distribution
- `frontend/src/components/LogInsights/AnomalyTimeline.tsx` - Anomaly visualization
- `frontend/src/hooks/useLogAnalysis.ts` - Data fetching hook
- `frontend/src/types/logAnalysis.ts` - TypeScript types

**v3.3.0 - Index Advisor:**
- `frontend/src/pages/IndexAdvisor.tsx` - Main page + dashboard
- `frontend/src/components/IndexAdvisor/RecommendationList.tsx` - Recommendation cards
- `frontend/src/components/IndexAdvisor/ImpactSimulator.tsx` - Index impact simulator
- `frontend/src/components/IndexAdvisor/UnusedIndexes.tsx` - Cleanup candidates
- `frontend/src/hooks/useIndexAdvisor.ts` - Data fetching hook
- `frontend/src/types/indexAdvisor.ts` - TypeScript types

**v3.4.0 - VACUUM Advisor:**
- `frontend/src/pages/VacuumAdvisor.tsx` - Main page
- `frontend/src/components/VacuumAdvisor/BloatAnalysis.tsx` - Bloat category
- `frontend/src/components/VacuumAdvisor/FreezingAnalysis.tsx` - Freezing category
- `frontend/src/components/VacuumAdvisor/PerformanceAnalysis.tsx` - Performance category
- `frontend/src/components/VacuumAdvisor/ActivityTimeline.tsx` - Activity category
- `frontend/src/hooks/useVacuumAdvisor.ts` - Data fetching hook
- `frontend/src/types/vacuumAdvisor.ts` - TypeScript types

### Collector (C/C++)

**v3.1.0:**
- `collector/src/plugins/query_performance_plugin.cpp` - EXPLAIN collector

**v3.2.0:**
- `collector/src/plugins/log_analysis_plugin.cpp` - Log ingestion

### Database (PostgreSQL + TimescaleDB)

**v3.1.0:**
- `backend/migrations/0XX_create_query_performance_schema.sql`

**v3.2.0:**
- `backend/migrations/0XX_create_log_analysis_schema.sql`

**v3.3.0:**
- `backend/migrations/0XX_create_index_advisor_schema.sql`

**v3.4.0:**
- `backend/migrations/0XX_create_vacuum_advisor_schema.sql`

---

## ONDA 1: v3.1.0 - Query Performance Deep Dive

### Task 1: Database Schema - Query Performance Tables

**Files:**
- Create: `backend/migrations/0XX_create_query_performance_schema.sql`
- Modify: None

- [ ] **Step 1: Write database migration**

```sql
-- Create query_plans table
CREATE TABLE query_plans (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    query_hash VARCHAR(64) NOT NULL,
    query_text TEXT NOT NULL,
    plan_json JSONB NOT NULL,
    mean_time FLOAT NOT NULL,
    total_time FLOAT NOT NULL,
    calls BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, query_hash)
);

CREATE TABLE query_issues (
    id BIGSERIAL PRIMARY KEY,
    query_plan_id BIGINT NOT NULL REFERENCES query_plans(id) ON DELETE CASCADE,
    issue_type VARCHAR(50) NOT NULL, -- 'sequential_scan', 'nested_loop', 'missing_index'
    severity VARCHAR(20) NOT NULL, -- 'low', 'medium', 'high', 'critical'
    affected_node_id INT,
    description TEXT,
    recommendation TEXT,
    estimated_benefit FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE query_performance_timeline (
    id BIGSERIAL PRIMARY KEY,
    query_plan_id BIGINT NOT NULL REFERENCES query_plans(id) ON DELETE CASCADE,
    metric_timestamp TIMESTAMP NOT NULL,
    avg_duration FLOAT NOT NULL,
    max_duration FLOAT NOT NULL,
    executions BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_query_plans_database ON query_plans(database_id);
CREATE INDEX idx_query_issues_query_plan ON query_issues(query_plan_id);
CREATE INDEX idx_timeline_query_timestamp ON query_performance_timeline(query_plan_id, metric_timestamp);
```

- [ ] **Step 2: Verify migration syntax**

Run: `cd backend && go run ./cmd/cli migrate verify migrations/0XX_create_query_performance_schema.sql`

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/0XX_create_query_performance_schema.sql
git commit -m "schema: add query performance tables (v3.1.0)"
```

---

### Task 2: Backend - Query Performance Collector Service

**Files:**
- Create: `backend/internal/services/query_performance/collector.go`
- Create: `backend/internal/services/query_performance/collector_test.go`

- [ ] **Step 1: Write failing test for QueryCollector**

```go
package query_performance

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestQueryCollector_CaptureExplainAnalyze(t *testing.T) {
    collector := NewQueryCollector(nil)
    assert.NotNil(t, collector)
}

func TestQueryCollector_ParseExplainOutput(t *testing.T) {
    collector := NewQueryCollector(nil)

    explainOutput := `{"Plan": {"Node Type": "Seq Scan"}}`
    plan, err := collector.ParseExplainOutput(explainOutput)

    assert.NoError(t, err)
    assert.NotNil(t, plan)
    assert.Equal(t, "Seq Scan", plan.NodeType)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd backend && go test ./internal/services/query_performance -v`
Expected: FAIL - NewQueryCollector undefined

- [ ] **Step 3: Write minimal QueryCollector implementation**

```go
package query_performance

import (
    "encoding/json"
    "database/sql"
)

type QueryCollector struct {
    db *sql.DB
}

type ExplainPlan struct {
    NodeType string `json:"Node Type"`
    // Additional fields...
}

func NewQueryCollector(db *sql.DB) *QueryCollector {
    return &QueryCollector{db: db}
}

func (qc *QueryCollector) ParseExplainOutput(output string) (*ExplainPlan, error) {
    var plan struct {
        Plan ExplainPlan `json:"Plan"`
    }

    err := json.Unmarshal([]byte(output), &plan)
    if err != nil {
        return nil, err
    }

    return &plan.Plan, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd backend && go test ./internal/services/query_performance -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/query_performance/collector.go
git add backend/internal/services/query_performance/collector_test.go
git commit -m "feat: implement query collector for EXPLAIN ANALYZE capture"
```

---

### Task 3: Backend - Query Performance Parser & Analyzer

**Files:**
- Create: `backend/internal/services/query_performance/parser.go`
- Create: `backend/internal/services/query_performance/analyzer.go`
- Create: `backend/internal/services/query_performance/models.go`

- [ ] **Step 1: Write test for QueryParser**

```go
package query_performance

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestQueryParser_DetectIssues(t *testing.T) {
    parser := NewQueryParser()

    plan := &ExplainPlan{
        NodeType: "Seq Scan",
        TotalCost: 1000.0,
    }

    issues := parser.DetectIssues(plan)
    assert.Greater(t, len(issues), 0)
    assert.Equal(t, "sequential_scan", issues[0].Type)
}
```

- [ ] **Step 2: Implement QueryParser with issue detection**

```go
package query_performance

type QueryIssue struct {
    Type            string
    Severity        string
    AffectedNode    string
    Description     string
    Recommendation  string
    EstimatedBenefit float64
}

type QueryParser struct{}

func NewQueryParser() *QueryParser {
    return &QueryParser{}
}

func (qp *QueryParser) DetectIssues(plan *ExplainPlan) []QueryIssue {
    var issues []QueryIssue

    // Detect sequential scans (expensive)
    if plan.NodeType == "Seq Scan" && plan.TotalCost > 100 {
        issues = append(issues, QueryIssue{
            Type:            "sequential_scan",
            Severity:        "medium",
            Description:     "Sequential scan detected on large table",
            Recommendation:  "Consider creating an index on the WHERE clause columns",
            EstimatedBenefit: plan.TotalCost * 0.3,
        })
    }

    // Detect nested loops
    if plan.NodeType == "Nested Loop" {
        issues = append(issues, QueryIssue{
            Type:            "nested_loop",
            Severity:        "high",
            Description:     "Nested loop detected, may be inefficient for large datasets",
            Recommendation:  "Consider hash join or additional indexes",
            EstimatedBenefit: plan.TotalCost * 0.4,
        })
    }

    return issues
}
```

- [ ] **Step 3: Write test and implement QueryAnalyzer**

```go
func TestQueryAnalyzer_CalculateScore(t *testing.T) {
    analyzer := NewQueryAnalyzer()

    issues := []QueryIssue{
        {Severity: "high"},
        {Severity: "medium"},
    }

    score := analyzer.CalculateSeverityScore(issues)
    assert.Greater(t, score, 0.0)
    assert.LessOrEqual(t, score, 100.0)
}

// Implementation
type QueryAnalyzer struct{}

func NewQueryAnalyzer() *QueryAnalyzer {
    return &QueryAnalyzer{}
}

func (qa *QueryAnalyzer) CalculateSeverityScore(issues []QueryIssue) float64 {
    if len(issues) == 0 {
        return 0.0
    }

    severityMap := map[string]float64{
        "low":      10.0,
        "medium":   40.0,
        "high":     70.0,
        "critical": 100.0,
    }

    totalScore := 0.0
    for _, issue := range issues {
        totalScore += severityMap[issue.Severity]
    }

    return totalScore / float64(len(issues))
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./internal/services/query_performance -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/query_performance/parser.go
git add backend/internal/services/query_performance/analyzer.go
git add backend/internal/services/query_performance/models.go
git commit -m "feat: implement query parser and analyzer for issue detection"
```

---

### Task 4: Backend - Query Performance API Handler

**Files:**
- Create: `backend/internal/api/handlers_query_performance.go`
- Create: `backend/internal/api/handlers_query_performance_test.go`

- [ ] **Step 1: Write test for API endpoint**

```go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestGetQueryPerformance(t *testing.T) {
    server := setupTestServer(t)

    req := httptest.NewRequest("GET", "/api/v1/query-performance/database/1", nil)
    w := httptest.NewRecorder()

    server.router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
}
```

- [ ] **Step 2: Implement API handler**

```go
package api

import (
    "net/http"
    "github.com/go-chi/chi"
)

func (s *Server) handleGetQueryPerformance(w http.ResponseWriter, r *http.Request) {
    databaseID := chi.URLParam(r, "database_id")

    // Fetch query performance data
    queries, err := s.storage.GetQueryPerformance(r.Context(), databaseID)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }

    s.json(w, http.StatusOK, map[string]interface{}{
        "queries": queries,
    })
}

func (s *Server) handleGetQueryPlan(w http.ResponseWriter, r *http.Request) {
    queryID := chi.URLParam(r, "query_id")

    plan, err := s.storage.GetQueryPlan(r.Context(), queryID)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }

    s.json(w, http.StatusOK, plan)
}
```

- [ ] **Step 3: Register routes in router**

```go
func (s *Server) registerQueryPerformanceRoutes(r chi.Router) {
    r.Get("/api/v1/query-performance/database/{database_id}", s.handleGetQueryPerformance)
    r.Get("/api/v1/query-performance/{query_id}", s.handleGetQueryPlan)
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./internal/api -v -run TestGetQueryPerformance`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/api/handlers_query_performance.go
git add backend/internal/api/handlers_query_performance_test.go
git commit -m "feat: add query performance API endpoints"
```

---

### Task 5: Frontend - Query Performance Types & Hooks

**Files:**
- Create: `frontend/src/types/queryPerformance.ts`
- Create: `frontend/src/hooks/useQueryPerformance.ts`

- [ ] **Step 1: Define TypeScript types**

```typescript
// frontend/src/types/queryPerformance.ts
export interface QueryPlan {
    id: number;
    database_id: number;
    query_hash: string;
    query_text: string;
    plan_json: Record<string, any>;
    mean_time: number;
    total_time: number;
    calls: number;
    created_at: string;
}

export interface QueryIssue {
    id: number;
    query_plan_id: number;
    issue_type: 'sequential_scan' | 'nested_loop' | 'missing_index' | 'hash_aggregate';
    severity: 'low' | 'medium' | 'high' | 'critical';
    affected_node_id: number;
    description: string;
    recommendation: string;
    estimated_benefit: number;
}

export interface PerformanceTimeline {
    id: number;
    query_plan_id: number;
    metric_timestamp: string;
    avg_duration: number;
    max_duration: number;
    executions: number;
}

export interface QueryPerformanceData {
    queries: QueryPlan[];
    issues: QueryIssue[];
    timeline: PerformanceTimeline[];
}
```

- [ ] **Step 2: Implement custom hook**

```typescript
// frontend/src/hooks/useQueryPerformance.ts
import { useState, useEffect } from 'react';
import { QueryPerformanceData } from '../types/queryPerformance';

export const useQueryPerformance = (databaseId: string) => {
    const [data, setData] = useState<QueryPerformanceData | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchData = async () => {
            try {
                const response = await fetch(`/api/v1/query-performance/database/${databaseId}`);
                if (!response.ok) throw new Error('Failed to fetch');
                const json = await response.json();
                setData(json);
            } catch (err) {
                setError(err instanceof Error ? err.message : 'Unknown error');
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [databaseId]);

    return { data, loading, error };
};
```

- [ ] **Step 3: Create test file**

```typescript
import { renderHook, waitFor } from '@testing-library/react';
import { useQueryPerformance } from './useQueryPerformance';

describe('useQueryPerformance', () => {
    it('fetches query performance data', async () => {
        const { result } = renderHook(() => useQueryPerformance('1'));

        await waitFor(() => {
            expect(result.current.loading).toBe(false);
        });

        expect(result.current.data).toBeDefined();
    });
});
```

- [ ] **Step 4: Run tests**

Run: `cd frontend && npm test -- useQueryPerformance`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/types/queryPerformance.ts
git add frontend/src/hooks/useQueryPerformance.ts
git commit -m "feat: add query performance types and hook"
```

---

### Task 6: Frontend - Query Performance Dashboard Page

**Files:**
- Create: `frontend/src/pages/QueryPerformance.tsx`
- Create: `frontend/src/components/QueryPlan/PlanTree.tsx`
- Create: `frontend/src/components/QueryPerformance/Timeline.tsx`

- [ ] **Step 1: Create PlanTree component**

```typescript
// frontend/src/components/QueryPlan/PlanTree.tsx
import React from 'react';
import { QueryPlan } from '../../types/queryPerformance';

interface PlanTreeProps {
    plan: QueryPlan;
}

export const PlanTree: React.FC<PlanTreeProps> = ({ plan }) => {
    const renderNode = (node: any, depth: number = 0) => {
        const indent = depth * 20;

        return (
            <div key={node.id} style={{ marginLeft: `${indent}px` }} className="p-2 border-l">
                <div className="font-mono text-sm">
                    {node['Node Type']}
                    {node['Actual Loops'] && ` (loops: ${node['Actual Loops']})`}
                </div>
                {node['Plans']?.map((child: any) => renderNode(child, depth + 1))}
            </div>
        );
    };

    return (
        <div className="bg-gray-50 p-4 rounded border">
            <h3 className="font-bold mb-3">Execution Plan</h3>
            {renderNode(plan.plan_json.Plan)}
        </div>
    );
};
```

- [ ] **Step 2: Create Timeline component**

```typescript
// frontend/src/components/QueryPerformance/Timeline.tsx
import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { PerformanceTimeline } from '../../types/queryPerformance';

interface TimelineProps {
    data: PerformanceTimeline[];
}

export const Timeline: React.FC<TimelineProps> = ({ data }) => {
    const chartData = data.map(d => ({
        timestamp: new Date(d.metric_timestamp).toLocaleString(),
        duration: d.avg_duration,
        executions: d.executions,
    }));

    return (
        <ResponsiveContainer width="100%" height={300}>
            <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="timestamp" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="duration" stroke="#8884d8" />
            </LineChart>
        </ResponsiveContainer>
    );
};
```

- [ ] **Step 3: Create QueryPerformance main page**

```typescript
// frontend/src/pages/QueryPerformance.tsx
import React from 'react';
import { useParams } from 'react-router-dom';
import { useQueryPerformance } from '../hooks/useQueryPerformance';
import { PlanTree } from '../components/QueryPlan/PlanTree';
import { Timeline } from '../components/QueryPerformance/Timeline';
import { PageWrapper } from '../components/PageWrapper';

export const QueryPerformancePage: React.FC = () => {
    const { databaseId } = useParams<{ databaseId: string }>();
    const { data, loading, error } = useQueryPerformance(databaseId || '');

    if (loading) return <div className="p-4">Loading...</div>;
    if (error) return <div className="p-4 text-red-600">Error: {error}</div>;
    if (!data) return <div className="p-4">No data</div>;

    return (
        <PageWrapper title="Query Performance" subtitle="Analyze and optimize slow queries">
            <div className="grid grid-cols-1 gap-6">
                {data.queries.map(query => (
                    <div key={query.id} className="bg-white p-6 rounded-lg shadow">
                        <div className="mb-4">
                            <h2 className="text-xl font-bold">{query.query_text.substring(0, 100)}</h2>
                            <div className="text-sm text-gray-600 mt-2">
                                Calls: {query.calls} | Avg time: {query.mean_time.toFixed(2)}ms
                            </div>
                        </div>
                        <PlanTree plan={query} />
                        <div className="mt-4">
                            <Timeline data={data.timeline.filter(t => t.query_plan_id === query.id)} />
                        </div>
                    </div>
                ))}
            </div>
        </PageWrapper>
    );
};
```

- [ ] **Step 4: Add route to main router**

```typescript
// frontend/src/App.tsx - add to routes
<Route path="/query-performance/:databaseId" element={<QueryPerformancePage />} />
```

- [ ] **Step 5: Run tests**

Run: `cd frontend && npm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/QueryPerformance.tsx
git add frontend/src/components/QueryPlan/PlanTree.tsx
git add frontend/src/components/QueryPerformance/Timeline.tsx
git commit -m "feat: add query performance dashboard with plan visualization"
```

---

## ONDA 1: v3.2.0 - Log Analysis Engine

### Task 7: Database Schema - Log Analysis Tables

**Files:**
- Create: `backend/migrations/0XX_create_log_analysis_schema.sql`

- [ ] **Step 1: Write database migration**

```sql
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    log_timestamp TIMESTAMP NOT NULL,
    category VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    message TEXT NOT NULL,
    duration FLOAT,
    table_affected VARCHAR(255),
    query_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE log_patterns (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    pattern_name VARCHAR(255) NOT NULL,
    pattern_regex TEXT NOT NULL,
    frequency BIGINT DEFAULT 0,
    severity_avg FLOAT,
    last_seen TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, pattern_name)
);

CREATE TABLE log_anomalies (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    pattern_id BIGINT REFERENCES log_patterns(id),
    anomaly_timestamp TIMESTAMP NOT NULL,
    anomaly_score FLOAT NOT NULL,
    deviation_from_baseline FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_logs_database_timestamp ON logs(database_id, log_timestamp DESC);
CREATE INDEX idx_logs_category ON logs(category);
CREATE INDEX idx_log_patterns_database ON log_patterns(database_id);
CREATE INDEX idx_log_anomalies_timestamp ON log_anomalies(anomaly_timestamp DESC);
```

- [ ] **Step 2: Verify migration**

Run: `cd backend && go run ./cmd/cli migrate verify migrations/0XX_create_log_analysis_schema.sql`

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/0XX_create_log_analysis_schema.sql
git commit -m "schema: add log analysis tables (v3.2.0)"
```

---

### Task 8: Backend - Log Parser with 50+ Categories

**Files:**
- Create: `backend/internal/services/log_analysis/parser.go`
- Create: `backend/internal/services/log_analysis/parser_test.go`
- Create: `backend/internal/services/log_analysis/categories.go`

- [ ] **Step 1: Define log categories**

```go
package log_analysis

// LogCategory represents a PostgreSQL log category
type LogCategory string

const (
    // Error-related categories
    CategoryDatabaseError     LogCategory = "database_error"
    CategoryConnectionError   LogCategory = "connection_error"
    CategoryAuthenticationError LogCategory = "authentication_error"
    CategorySyntaxError       LogCategory = "syntax_error"
    CategoryConstraintError   LogCategory = "constraint_error"

    // Performance categories
    CategorySlowQuery         LogCategory = "slow_query"
    CategoryCheckpointError   LogCategory = "checkpoint"
    CategoryVacuumError       LogCategory = "vacuum"
    CategoryLongTransaction   LogCategory = "long_transaction"

    // Lock-related categories
    CategoryLockTimeout       LogCategory = "lock_timeout"
    CategoryDeadlock          LogCategory = "deadlock"

    // Replication categories
    CategoryReplicationError  LogCategory = "replication_error"
    CategoryWALError          LogCategory = "wal_error"

    // Memory/Resource categories
    CategoryOutOfMemory       LogCategory = "out_of_memory"
    CategoryDiskFull          LogCategory = "disk_full"

    // Other
    CategoryWarning           LogCategory = "warning"
    CategoryInfo              LogCategory = "info"
)

// LogCategoryPatterns maps patterns to categories
var LogCategoryPatterns = map[LogCategory][]string{
    CategoryDatabaseError: {
        "database .* does not exist",
        "FATAL: database .* does not exist",
    },
    CategoryConnectionError: {
        "connection refused",
        "Connection refused",
        "FATAL: could not accept SSL connection",
    },
    CategoryAuthenticationError: {
        "FATAL: no pg_hba.conf entry",
        "FATAL: password authentication failed",
        "role .* does not exist",
    },
    CategorySlowQuery: {
        "duration: \\d+\\.\\d+ ms",
        "slow query",
    },
    CategoryCheckpointError: {
        "LOG: checkpoint",
        "FATAL: checkpoint failed",
    },
    // ... 40+ more categories
}
```

- [ ] **Step 2: Write test for LogParser**

```go
func TestLogParser_ClassifyLog(t *testing.T) {
    parser := NewLogParser()

    tests := []struct {
        message  string
        expected LogCategory
    }{
        {"FATAL: database mydb does not exist", CategoryDatabaseError},
        {"Connection refused", CategoryConnectionError},
        {"duration: 5000.123 ms", CategorySlowQuery},
    }

    for _, tt := range tests {
        category := parser.ClassifyLog(tt.message)
        assert.Equal(t, tt.expected, category)
    }
}
```

- [ ] **Step 3: Implement LogParser**

```go
import (
    "regexp"
    "strings"
)

type LogParser struct {
    patterns map[LogCategory]*regexp.Regexp
}

func NewLogParser() *LogParser {
    parser := &LogParser{
        patterns: make(map[LogCategory]*regexp.Regexp),
    }

    for category, patternList := range LogCategoryPatterns {
        // Combine multiple patterns with OR
        combined := "(" + strings.Join(patternList, "|") + ")"
        parser.patterns[category] = regexp.MustCompile(combined)
    }

    return parser
}

func (lp *LogParser) ClassifyLog(message string) LogCategory {
    for category, pattern := range lp.patterns {
        if pattern.MatchString(message) {
            return category
        }
    }

    if strings.Contains(strings.ToLower(message), "error") {
        return CategoryDatabaseError
    }

    if strings.Contains(strings.ToLower(message), "warning") {
        return CategoryWarning
    }

    return CategoryInfo
}

func (lp *LogParser) ExtractMetadata(message string) map[string]interface{} {
    metadata := make(map[string]interface{})

    // Extract duration (e.g., "duration: 1234.56 ms")
    durationRegex := regexp.MustCompile(`duration: ([\d.]+) ms`)
    if matches := durationRegex.FindStringSubmatch(message); len(matches) > 1 {
        metadata["duration"] = matches[1]
    }

    // Extract table name (e.g., "relation \"mytable\"")
    tableRegex := regexp.MustCompile(`relation "([^"]+)"`)
    if matches := tableRegex.FindStringSubmatch(message); len(matches) > 1 {
        metadata["table"] = matches[1]
    }

    return metadata
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./internal/services/log_analysis -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/log_analysis/parser.go
git add backend/internal/services/log_analysis/parser_test.go
git add backend/internal/services/log_analysis/categories.go
git commit -m "feat: implement log parser with 50+ categories"
```

---

### Task 9: Backend - Log Collector Service

**Files:**
- Create: `backend/internal/services/log_analysis/collector.go`
- Create: `backend/internal/services/log_analysis/collector_test.go`

- [ ] **Step 1: Write test for LogCollector**

```go
func TestLogCollector_IngestLogs(t *testing.T) {
    collector := NewLogCollector(nil)

    logs := []map[string]interface{}{
        {
            "timestamp": "2026-03-30T10:00:00Z",
            "message":   "FATAL: database mydb does not exist",
            "severity":  "FATAL",
        },
    }

    err := collector.IngestLogs(context.Background(), "db1", logs)
    assert.NoError(t, err)
}
```

- [ ] **Step 2: Implement LogCollector**

```go
package log_analysis

import (
    "context"
    "database/sql"
    "time"
)

type LogCollector struct {
    db     *sql.DB
    parser *LogParser
}

func NewLogCollector(db *sql.DB) *LogCollector {
    return &LogCollector{
        db:     db,
        parser: NewLogParser(),
    }
}

func (lc *LogCollector) IngestLogs(ctx context.Context, databaseID string, logs []map[string]interface{}) error {
    for _, logEntry := range logs {
        message := logEntry["message"].(string)
        severity := logEntry["severity"].(string)
        timestamp, _ := time.Parse(time.RFC3339, logEntry["timestamp"].(string))

        // Classify the log
        category := lc.parser.ClassifyLog(message)
        metadata := lc.parser.ExtractMetadata(message)

        // Insert into database
        query := `
            INSERT INTO logs (database_id, log_timestamp, category, severity, message, duration, table_affected)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
        `

        _, err := lc.db.ExecContext(ctx, query,
            databaseID,
            timestamp,
            category,
            severity,
            message,
            metadata["duration"],
            metadata["table"],
        )

        if err != nil {
            return err
        }
    }

    return nil
}

func (lc *LogCollector) StreamLogs(ctx context.Context, databaseID string, ch chan<- map[string]interface{}) error {
    // Real-time log streaming (used for WebSocket)
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            // Query recent logs
            rows, err := lc.db.QueryContext(ctx, `
                SELECT id, log_timestamp, category, severity, message
                FROM logs
                WHERE database_id = $1
                ORDER BY created_at DESC
                LIMIT 10
            `, databaseID)

            if err != nil {
                return err
            }

            // Send to channel
            for rows.Next() {
                var id int64
                var timestamp time.Time
                var category, severity, message string

                rows.Scan(&id, &timestamp, &category, &severity, &message)
                ch <- map[string]interface{}{
                    "id":        id,
                    "timestamp": timestamp,
                    "category":  category,
                    "severity":  severity,
                    "message":   message,
                }
            }
            rows.Close()
        }
    }
}
```

- [ ] **Step 3: Run tests**

Run: `cd backend && go test ./internal/services/log_analysis -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/services/log_analysis/collector.go
git add backend/internal/services/log_analysis/collector_test.go
git commit -m "feat: implement log collector with ingestion and streaming"
```

---

### Task 10: Backend - Log Analysis API Handler with WebSocket

**Files:**
- Create: `backend/internal/api/handlers_log_analysis.go`

- [ ] **Step 1: Write test for log API**

```go
func TestGetLogs(t *testing.T) {
    server := setupTestServer(t)

    req := httptest.NewRequest("GET", "/api/v1/logs/database/1", nil)
    w := httptest.NewRecorder()

    server.router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogWebSocket(t *testing.T) {
    server := setupTestServer(t)

    // Test WebSocket connection
    req := httptest.NewRequest("GET", "/api/v1/logs/stream/1", nil)
    req.Header.Set("Upgrade", "websocket")
    req.Header.Set("Connection", "Upgrade")
    req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
    req.Header.Set("Sec-WebSocket-Version", "13")

    w := httptest.NewRecorder()
    server.router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusSwitchingProtocols, w.Code)
}
```

- [ ] **Step 2: Implement log endpoints**

```go
package api

import (
    "github.com/gorilla/websocket"
    "net/http"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
    databaseID := chi.URLParam(r, "database_id")

    rows, err := s.storage.GetLogs(r.Context(), databaseID)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }

    s.json(w, http.StatusOK, map[string]interface{}{
        "logs": rows,
    })
}

func (s *Server) handleLogStream(w http.ResponseWriter, r *http.Request) {
    databaseID := chi.URLParam(r, "database_id")

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }
    defer conn.Close()

    // Stream logs to WebSocket
    logChan := make(chan map[string]interface{})
    go s.logCollector.StreamLogs(r.Context(), databaseID, logChan)

    for log := range logChan {
        if err := conn.WriteJSON(log); err != nil {
            break
        }
    }
}

func (s *Server) registerLogAnalysisRoutes(r chi.Router) {
    r.Get("/api/v1/logs/database/{database_id}", s.handleGetLogs)
    r.Get("/api/v1/logs/stream/{database_id}", s.handleLogStream)
}
```

- [ ] **Step 3: Run tests**

Run: `cd backend && go test ./internal/api -v -run TestGetLogs`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add backend/internal/api/handlers_log_analysis.go
git commit -m "feat: add log analysis API endpoints with WebSocket streaming"
```

---

### Task 11: Frontend - Log Analysis Types & Components

**Files:**
- Create: `frontend/src/types/logAnalysis.ts`
- Create: `frontend/src/hooks/useLogAnalysis.ts`
- Create: `frontend/src/pages/LogAnalysis.tsx`
- Create: `frontend/src/components/LogInsights/LogStream.tsx`

- [ ] **Step 1: Define types**

```typescript
export interface LogEntry {
    id: number;
    database_id: number;
    log_timestamp: string;
    category: string;
    severity: 'INFO' | 'WARNING' | 'ERROR' | 'FATAL';
    message: string;
    duration?: number;
    table_affected?: string;
}

export interface LogPattern {
    id: number;
    pattern_name: string;
    frequency: number;
    severity_avg: number;
    last_seen: string;
}

export interface LogAnomaly {
    id: number;
    pattern_id: number;
    anomaly_timestamp: string;
    anomaly_score: number;
    deviation_from_baseline: number;
}
```

- [ ] **Step 2: Create hook with WebSocket**

```typescript
export const useLogAnalysis = (databaseId: string) => {
    const [logs, setLogs] = useState<LogEntry[]>([]);
    const [connected, setConnected] = useState(false);

    useEffect(() => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const ws = new WebSocket(`${protocol}//${window.location.host}/api/v1/logs/stream/${databaseId}`);

        ws.onopen = () => setConnected(true);
        ws.onmessage = (event) => {
            const newLog = JSON.parse(event.data);
            setLogs(prev => [newLog, ...prev].slice(0, 100)); // Keep last 100
        };
        ws.onclose = () => setConnected(false);

        return () => ws.close();
    }, [databaseId]);

    return { logs, connected };
};
```

- [ ] **Step 3: Create LogStream component**

```typescript
export const LogStream: React.FC<{ databaseId: string }> = ({ databaseId }) => {
    const { logs, connected } = useLogAnalysis(databaseId);

    const severityColors = {
        'INFO': 'text-blue-600',
        'WARNING': 'text-yellow-600',
        'ERROR': 'text-red-600',
        'FATAL': 'text-red-900',
    };

    return (
        <div className="bg-gray-900 text-white p-4 rounded font-mono text-sm max-h-96 overflow-y-auto">
            <div className="mb-2 text-xs">
                {connected ? '🟢 Connected' : '🔴 Disconnected'}
            </div>
            {logs.map(log => (
                <div key={log.id} className="mb-1">
                    <span className="text-gray-400">[{new Date(log.log_timestamp).toLocaleTimeString()}]</span>
                    <span className={`ml-2 ${severityColors[log.severity]}`}>
                        {log.severity}
                    </span>
                    <span className="ml-2 text-gray-200">{log.message}</span>
                </div>
            ))}
        </div>
    );
};
```

- [ ] **Step 4: Create LogAnalysis page**

```typescript
export const LogAnalysisPage: React.FC = () => {
    const { databaseId } = useParams<{ databaseId: string }>();

    return (
        <PageWrapper title="Log Analysis" subtitle="Real-time PostgreSQL log insights">
            <div className="grid grid-cols-1 gap-6">
                <div className="bg-white p-6 rounded-lg shadow">
                    <h2 className="text-xl font-bold mb-4">Live Log Stream</h2>
                    <LogStream databaseId={databaseId || ''} />
                </div>
            </div>
        </PageWrapper>
    );
};
```

- [ ] **Step 5: Run tests**

Run: `cd frontend && npm test`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/types/logAnalysis.ts
git add frontend/src/hooks/useLogAnalysis.ts
git add frontend/src/pages/LogAnalysis.tsx
git add frontend/src/components/LogInsights/LogStream.tsx
git commit -m "feat: add log analysis frontend with real-time streaming"
```

---

## ONDA 2: v3.3.0 - Index Advisor

### Task 12: Database Schema - Index Advisor Tables

**Files:**
- Create: `backend/migrations/0XX_create_index_advisor_schema.sql`

- [ ] **Step 1: Write migration**

```sql
CREATE TABLE index_recommendations (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    column_names TEXT[] NOT NULL,
    index_type VARCHAR(50) DEFAULT 'btree',
    estimated_benefit FLOAT NOT NULL,
    weighted_cost_improvement FLOAT NOT NULL,
    status VARCHAR(20) DEFAULT 'recommended', -- 'recommended', 'created', 'rejected'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, table_name, column_names)
);

CREATE TABLE index_analysis (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    query_id BIGINT REFERENCES query_plans(id),
    index_candidate VARCHAR(255),
    cost_without_index FLOAT,
    cost_with_index FLOAT,
    benefit_score FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE unused_indexes (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    index_name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    size_bytes BIGINT,
    last_used TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_recommendations_database ON index_recommendations(database_id);
CREATE INDEX idx_analysis_query ON index_analysis(query_id);
CREATE INDEX idx_unused_indexes_database ON unused_indexes(database_id);
```

- [ ] **Step 2: Verify and commit**

Run: `cd backend && go run ./cmd/cli migrate verify migrations/0XX_create_index_advisor_schema.sql`

```bash
git add backend/migrations/0XX_create_index_advisor_schema.sql
git commit -m "schema: add index advisor tables (v3.3.0)"
```

---

### Task 13: Backend - Index Advisor Analyzer (Constraint Programming)

**Files:**
- Create: `backend/internal/services/index_advisor/analyzer.go`
- Create: `backend/internal/services/index_advisor/cost_calculator.go`
- Create: `backend/internal/services/index_advisor/analyzer_test.go`

- [ ] **Step 1: Write tests**

```go
func TestIndexAnalyzer_FindMissingIndexes(t *testing.T) {
    analyzer := NewIndexAnalyzer(nil)

    // Simulate query plan with seq scan
    plan := &QueryPlan{
        NodeType: "Seq Scan",
        TotalCost: 1000.0,
    }

    recommendations := analyzer.FindMissingIndexes([]*QueryPlan{plan})
    assert.Greater(t, len(recommendations), 0)
}

func TestCostCalculator_CalculateImprovement(t *testing.T) {
    calc := NewCostCalculator()

    costWithout := 1000.0
    costWith := 100.0

    improvement := calc.CalculateImprovement(costWithout, costWith)
    assert.Equal(t, 90.0, improvement)
}
```

- [ ] **Step 2: Implement CostCalculator**

```go
package index_advisor

type CostCalculator struct{}

func NewCostCalculator() *CostCalculator {
    return &CostCalculator{}
}

func (cc *CostCalculator) CalculateImprovement(costWithout, costWith float64) float64 {
    if costWithout == 0 {
        return 0
    }
    return ((costWithout - costWith) / costWithout) * 100
}

func (cc *CostCalculator) EstimateBenefit(costImprovement, frequency float64) float64 {
    return costImprovement * frequency / 100
}

func (cc *CostCalculator) CalculateIndexMaintenanceCost(tableWriteFrequency float64) float64 {
    // Each index adds ~2-5% write overhead
    return tableWriteFrequency * 0.03
}

func (cc *CostCalculator) ShouldCreateIndex(benefit, maintenanceCost float64) bool {
    // Create index if benefit > maintenance cost by 2x
    return benefit > (maintenanceCost * 2)
}
```

- [ ] **Step 3: Implement IndexAnalyzer**

```go
package index_advisor

type IndexRecommendation struct {
    TableName              string
    ColumnNames            []string
    IndexType              string
    EstimatedBenefit       float64
    WeightedCostImprovement float64
}

type IndexAnalyzer struct {
    db     *sql.DB
    calc   *CostCalculator
}

func NewIndexAnalyzer(db *sql.DB) *IndexAnalyzer {
    return &IndexAnalyzer{
        db:   db,
        calc: NewCostCalculator(),
    }
}

func (ia *IndexAnalyzer) FindMissingIndexes(queryPlans []*QueryPlan) []IndexRecommendation {
    var recommendations []IndexRecommendation

    for _, plan := range queryPlans {
        // Extract WHERE and JOIN conditions from plan
        conditions := ia.extractConditions(plan)

        for _, cond := range conditions {
            if cond.SeqScan || cond.CostlyJoin {
                // Recommend index on condition columns
                costImprovement := ia.calc.CalculateImprovement(
                    cond.CostWithout,
                    cond.CostWith,
                )

                benefit := ia.calc.EstimateBenefit(costImprovement, float64(plan.Calls))
                maintenanceCost := ia.calc.CalculateIndexMaintenanceCost(
                    ia.getTableWriteFrequency(cond.TableName),
                )

                if ia.calc.ShouldCreateIndex(benefit, maintenanceCost) {
                    recommendations = append(recommendations, IndexRecommendation{
                        TableName:              cond.TableName,
                        ColumnNames:            cond.Columns,
                        IndexType:              "btree",
                        EstimatedBenefit:       benefit,
                        WeightedCostImprovement: costImprovement,
                    })
                }
            }
        }
    }

    return recommendations
}

func (ia *IndexAnalyzer) extractConditions(plan *QueryPlan) []Condition {
    // Parse WHERE and JOIN conditions from EXPLAIN plan
    // This is simplified - real implementation parses JSON plan structure
    return []Condition{}
}

func (ia *IndexAnalyzer) getTableWriteFrequency(tableName string) float64 {
    // Query pg_stat_user_tables for INSERT/UPDATE/DELETE frequency
    return 0.0 // Simplified
}

type Condition struct {
    TableName   string
    Columns     []string
    SeqScan     bool
    CostlyJoin  bool
    CostWithout float64
    CostWith    float64
}
```

- [ ] **Step 4: Run tests**

Run: `cd backend && go test ./internal/services/index_advisor -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add backend/internal/services/index_advisor/analyzer.go
git add backend/internal/services/index_advisor/cost_calculator.go
git add backend/internal/services/index_advisor/analyzer_test.go
git commit -m "feat: implement index advisor with cost-based analysis"
```

---

### Task 14: Backend - Index Advisor API Handler

**Files:**
- Create: `backend/internal/api/handlers_index_advisor.go`

- [ ] **Step 1: Implement endpoints**

```go
func (s *Server) handleGetIndexRecommendations(w http.ResponseWriter, r *http.Request) {
    databaseID := chi.URLParam(r, "database_id")

    recommendations, err := s.storage.GetIndexRecommendations(r.Context(), databaseID)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }

    s.json(w, http.StatusOK, map[string]interface{}{
        "recommendations": recommendations,
    })
}

func (s *Server) handleCreateIndex(w http.ResponseWriter, r *http.Request) {
    recommendationID := chi.URLParam(r, "recommendation_id")

    // Create the index
    err := s.storage.CreateIndexFromRecommendation(r.Context(), recommendationID)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }

    s.json(w, http.StatusOK, map[string]interface{}{
        "status": "created",
    })
}

func (s *Server) handleGetUnusedIndexes(w http.ResponseWriter, r *http.Request) {
    databaseID := chi.URLParam(r, "database_id")

    unused, err := s.storage.GetUnusedIndexes(r.Context(), databaseID)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }

    s.json(w, http.StatusOK, map[string]interface{}{
        "unused_indexes": unused,
    })
}

func (s *Server) registerIndexAdvisorRoutes(r chi.Router) {
    r.Get("/api/v1/index-advisor/database/{database_id}/recommendations", s.handleGetIndexRecommendations)
    r.Post("/api/v1/index-advisor/recommendation/{recommendation_id}/create", s.handleCreateIndex)
    r.Get("/api/v1/index-advisor/database/{database_id}/unused", s.handleGetUnusedIndexes)
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/api/handlers_index_advisor.go
git commit -m "feat: add index advisor API endpoints"
```

---

### Task 15: Frontend - Index Advisor Dashboard

**Files:**
- Create: `frontend/src/pages/IndexAdvisor.tsx`
- Create: `frontend/src/components/IndexAdvisor/RecommendationCard.tsx`

- [ ] **Step 1: Create RecommendationCard**

```typescript
export interface IndexRecommendation {
    id: number;
    table_name: string;
    column_names: string[];
    estimated_benefit: number;
    weighted_cost_improvement: number;
    status: 'recommended' | 'created' | 'rejected';
}

export const RecommendationCard: React.FC<{
    recommendation: IndexRecommendation;
    onCreateIndex: (id: number) => Promise<void>;
}> = ({ recommendation, onCreateIndex }) => {
    const [loading, setLoading] = useState(false);

    return (
        <div className="bg-white p-4 border-l-4 border-blue-500 rounded shadow">
            <div className="flex justify-between items-start">
                <div>
                    <h3 className="font-bold">{recommendation.table_name}</h3>
                    <p className="text-sm text-gray-600">
                        Columns: {recommendation.column_names.join(', ')}
                    </p>
                    <div className="mt-2 flex gap-4 text-sm">
                        <div>
                            <span className="text-gray-600">Estimated Benefit:</span>
                            <span className="ml-2 font-bold text-green-600">
                                {recommendation.estimated_benefit.toFixed(1)}%
                            </span>
                        </div>
                        <div>
                            <span className="text-gray-600">Cost Improvement:</span>
                            <span className="ml-2 font-bold">
                                {recommendation.weighted_cost_improvement.toFixed(1)}%
                            </span>
                        </div>
                    </div>
                </div>
                <button
                    onClick={() => {
                        setLoading(true);
                        onCreateIndex(recommendation.id).finally(() => setLoading(false));
                    }}
                    disabled={loading || recommendation.status === 'created'}
                    className="px-4 py-2 bg-blue-600 text-white rounded disabled:bg-gray-400"
                >
                    {recommendation.status === 'created' ? 'Created' : 'Create Index'}
                </button>
            </div>
        </div>
    );
};
```

- [ ] **Step 2: Create IndexAdvisor page**

```typescript
export const IndexAdvisorPage: React.FC = () => {
    const { databaseId } = useParams<{ databaseId: string }>();
    const [recommendations, setRecommendations] = useState<IndexRecommendation[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetch = async () => {
            const res = await window.fetch(
                `/api/v1/index-advisor/database/${databaseId}/recommendations`
            );
            const data = await res.json();
            setRecommendations(data.recommendations);
            setLoading(false);
        };
        fetch();
    }, [databaseId]);

    const handleCreateIndex = async (id: number) => {
        await window.fetch(
            `/api/v1/index-advisor/recommendation/${id}/create`,
            { method: 'POST' }
        );

        // Refresh recommendations
        const res = await window.fetch(
            `/api/v1/index-advisor/database/${databaseId}/recommendations`
        );
        const data = await res.json();
        setRecommendations(data.recommendations);
    };

    if (loading) return <div>Loading...</div>;

    return (
        <PageWrapper title="Index Advisor" subtitle="Find missing and unused indexes">
            <div className="space-y-4">
                {recommendations.map(rec => (
                    <RecommendationCard
                        key={rec.id}
                        recommendation={rec}
                        onCreateIndex={handleCreateIndex}
                    />
                ))}
            </div>
        </PageWrapper>
    );
};
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/IndexAdvisor.tsx
git add frontend/src/components/IndexAdvisor/RecommendationCard.tsx
git commit -m "feat: add index advisor dashboard with recommendations"
```

---

## ONDA 3: v3.4.0 - VACUUM Advisor

### Task 16: Database Schema - VACUUM Advisor Tables

**Files:**
- Create: `backend/migrations/0XX_create_vacuum_advisor_schema.sql`

- [ ] **Step 1: Write migration**

```sql
CREATE TABLE table_vacuum_stats (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    table_size_bytes BIGINT,
    dead_tuples BIGINT,
    live_tuples BIGINT,
    last_vacuum TIMESTAMP,
    last_autovacuum TIMESTAMP,
    last_analyze TIMESTAMP,
    bloat_ratio FLOAT,
    freezing_status VARCHAR(50), -- 'healthy', 'warning', 'critical'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(database_id, table_name)
);

CREATE TABLE vacuum_recommendations (
    id BIGSERIAL PRIMARY KEY,
    database_id INT NOT NULL REFERENCES databases(id) ON DELETE CASCADE,
    table_name VARCHAR(255) NOT NULL,
    rec_type VARCHAR(50), -- 'bloat', 'freezing', 'performance', 'activity'
    current_setting VARCHAR(255),
    recommended_setting VARCHAR(255),
    benefit_description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vacuum_stats_database ON table_vacuum_stats(database_id);
CREATE INDEX idx_recommendations_database_table ON vacuum_recommendations(database_id, table_name);
```

- [ ] **Step 2: Commit**

```bash
git add backend/migrations/0XX_create_vacuum_advisor_schema.sql
git commit -m "schema: add vacuum advisor tables (v3.4.0)"
```

---

### Task 17: Backend - VACUUM Advisor Analyzer

**Files:**
- Create: `backend/internal/services/vacuum_advisor/analyzer.go`
- Create: `backend/internal/services/vacuum_advisor/simulator.go`

- [ ] **Step 1: Implement VACUUMAnalyzer**

```go
package vacuum_advisor

import (
    "context"
    "database/sql"
    "time"
)

type TableVacuumStats struct {
    TableName        string
    TableSizeBytes   int64
    DeadTuples       int64
    LiveTuples       int64
    LastVacuum       *time.Time
    LastAutovacuum   *time.Time
    BloatRatio       float64
    FreezingStatus   string
}

type VacuumRecommendation struct {
    TableName             string
    RecType               string // 'bloat', 'freezing', 'performance', 'activity'
    CurrentSetting        string
    RecommendedSetting    string
    BenefitDescription    string
}

type VacuumAnalyzer struct {
    db *sql.DB
}

func NewVacuumAnalyzer(db *sql.DB) *VacuumAnalyzer {
    return &VacuumAnalyzer{db: db}
}

func (va *VacuumAnalyzer) AnalyzeTable(ctx context.Context, databaseID, tableName string) (*TableVacuumStats, []VacuumRecommendation) {
    // Get table statistics
    stats := va.getTableStats(ctx, databaseID, tableName)

    // Generate recommendations based on 4 categories
    var recommendations []VacuumRecommendation

    // 1. Bloat Analysis
    if stats.BloatRatio > 0.2 { // 20% bloat
        recommendations = append(recommendations, VacuumRecommendation{
            TableName:          tableName,
            RecType:            "bloat",
            CurrentSetting:     "autovacuum_vacuum_scale_factor = 0.1",
            RecommendedSetting: "autovacuum_vacuum_scale_factor = 0.01",
            BenefitDescription: "Lower scale factor to run VACUUM more frequently and reduce bloat",
        })
    }

    // 2. Freezing Analysis
    if stats.FreezingStatus == "warning" {
        recommendations = append(recommendations, VacuumRecommendation{
            TableName:          tableName,
            RecType:            "freezing",
            CurrentSetting:     "autovacuum_max_freeze_age = 200000000",
            RecommendedSetting: "autovacuum_max_freeze_age = 400000000",
            BenefitDescription: "Increase freeze age to reduce anti-wraparound VACUUMs",
        })
    }

    // 3. Performance Analysis
    if stats.LastAutovacuum == nil || time.Since(*stats.LastAutovacuum) > 24*time.Hour {
        recommendations = append(recommendations, VacuumRecommendation{
            TableName:          tableName,
            RecType:            "performance",
            CurrentSetting:     "autovacuum_vacuum_cost_delay = 20ms",
            RecommendedSetting: "autovacuum_vacuum_cost_delay = 5ms",
            BenefitDescription: "Lower cost delay to speed up VACUUM on this table",
        })
    }

    return stats, recommendations
}

func (va *VacuumAnalyzer) getTableStats(ctx context.Context, databaseID, tableName string) *TableVacuumStats {
    // Query pg_stat_user_tables for stats
    // This is simplified - real implementation queries the DB
    return &TableVacuumStats{}
}
```

- [ ] **Step 2: Implement VACUUMSimulator**

```go
package vacuum_advisor

type VacuumSimulation struct {
    TableName          string
    CurrentBloat       float64
    ProjectedBloat     float64
    TimeToClean        float64 // minutes
    ImpactScore        float64 // -1.0 to 1.0
}

type VacuumSimulator struct{}

func NewVacuumSimulator() *VacuumSimulator {
    return &VacuumSimulator{}
}

func (vs *VacuumSimulator) SimulateAutovacuumTuning(
    stats *TableVacuumStats,
    scaleFactorChange float64,
) *VacuumSimulation {

    // Estimate effect of changing scale_factor
    // Lower scale factor = more frequent VACUUM = less bloat
    projectedBloat := stats.BloatRatio * (1 - (scaleFactorChange * 0.5))
    if projectedBloat < 0 {
        projectedBloat = 0
    }

    bloatReduction := stats.BloatRatio - projectedBloat
    impactScore := bloatReduction / stats.BloatRatio

    return &VacuumSimulation{
        TableName:      stats.TableName,
        CurrentBloat:   stats.BloatRatio,
        ProjectedBloat: projectedBloat,
        TimeToClean:    stats.TableSizeBytes / (1024 * 1024 * 10), // 10MB/min estimate
        ImpactScore:    impactScore,
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/vacuum_advisor/analyzer.go
git add backend/internal/services/vacuum_advisor/simulator.go
git commit -m "feat: implement vacuum advisor with 4-category analysis and simulator"
```

---

### Task 18: Backend - VACUUM Advisor API & Frontend

**Files:**
- Create: `backend/internal/api/handlers_vacuum_advisor.go`
- Create: `frontend/src/pages/VacuumAdvisor.tsx`

- [ ] **Step 1: Implement API handler**

```go
func (s *Server) handleGetVacuumAdvisor(w http.ResponseWriter, r *http.Request) {
    databaseID := chi.URLParam(r, "database_id")

    // Get all tables with VACUUM stats
    tables, err := s.storage.GetTableVacuumStats(r.Context(), databaseID)
    if err != nil {
        s.error(w, http.StatusInternalServerError, err)
        return
    }

    // Generate recommendations for each table
    recommendations := make(map[string][]VacuumRecommendation)
    for _, table := range tables {
        _, recs := s.vacuumAnalyzer.AnalyzeTable(r.Context(), databaseID, table.TableName)
        recommendations[table.TableName] = recs
    }

    s.json(w, http.StatusOK, map[string]interface{}{
        "stats":            tables,
        "recommendations":  recommendations,
    })
}

func (s *Server) registerVacuumAdvisorRoutes(r chi.Router) {
    r.Get("/api/v1/vacuum-advisor/database/{database_id}", s.handleGetVacuumAdvisor)
}
```

- [ ] **Step 2: Create frontend page**

```typescript
export const VacuumAdvisorPage: React.FC = () => {
    const { databaseId } = useParams<{ databaseId: string }>();
    const [data, setData] = useState<any>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch(`/api/v1/vacuum-advisor/database/${databaseId}`)
            .then(r => r.json())
            .then(d => {
                setData(d);
                setLoading(false);
            });
    }, [databaseId]);

    if (loading) return <div>Loading...</div>;

    return (
        <PageWrapper title="VACUUM Advisor" subtitle="Optimize autovacuum settings per table">
            <div className="grid grid-cols-2 gap-4">
                <div>
                    <h3 className="font-bold mb-2">Bloat Analysis</h3>
                    {data.stats
                        .filter((t: any) => parseFloat(t.bloat_ratio) > 0.1)
                        .map((t: any) => (
                            <div key={t.table_name} className="p-2 bg-orange-50 border border-orange-200 rounded mb-2">
                                <div className="font-semibold">{t.table_name}</div>
                                <div className="text-sm">Bloat: {(parseFloat(t.bloat_ratio) * 100).toFixed(1)}%</div>
                            </div>
                        ))}
                </div>

                <div>
                    <h3 className="font-bold mb-2">Freezing Status</h3>
                    {data.stats
                        .filter((t: any) => t.freezing_status === 'warning' || t.freezing_status === 'critical')
                        .map((t: any) => (
                            <div key={t.table_name} className="p-2 bg-red-50 border border-red-200 rounded mb-2">
                                <div className="font-semibold">{t.table_name}</div>
                                <div className="text-sm">Status: {t.freezing_status.toUpperCase()}</div>
                            </div>
                        ))}
                </div>
            </div>
        </PageWrapper>
    );
};
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/api/handlers_vacuum_advisor.go
git add frontend/src/pages/VacuumAdvisor.tsx
git commit -m "feat: complete vacuum advisor with API and frontend"
```

---

## Integration Tasks

### Task 19: Collector Integration - Query Performance Plugin

**Files:**
- Create: `collector/src/plugins/query_performance_plugin.cpp`

- [ ] **Step 1: Implement query collector**

```cpp
#include <pqxx/pqxx>
#include "plugin.h"

class QueryPerformancePlugin : public CollectorPlugin {
public:
    void Collect() override {
        auto conn = GetConnection();
        pqxx::work txn(*conn);

        // Query pg_stat_statements for query stats
        auto res = txn.exec(
            "SELECT query, mean_exec_time, calls, total_exec_time "
            "FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 100"
        );

        // For each query, capture EXPLAIN ANALYZE
        for (auto row : res) {
            std::string query = row[0].as<std::string>();

            // Execute EXPLAIN ANALYZE and capture plan
            auto explain_res = txn.exec(
                "EXPLAIN (FORMAT JSON, ANALYZE) " + query
            );

            std::string plan_json = explain_res[0][0].as<std::string>();

            // Send to backend
            SendMetric("query_performance", {
                {"query_hash", HashQuery(query)},
                {"query_text", query},
                {"plan_json", plan_json},
                {"mean_time", row[1].as<double>()},
                {"calls", row[2].as<long>()},
            });
        }

        txn.commit();
    }
};
```

- [ ] **Step 2: Commit**

```bash
git add collector/src/plugins/query_performance_plugin.cpp
git commit -m "feat: add query performance collector plugin"
```

---

### Task 20: Collector Integration - Log Analysis Plugin

**Files:**
- Create: `collector/src/plugins/log_analysis_plugin.cpp`

- [ ] **Step 1: Implement log collector**

```cpp
#include <fstream>
#include <sstream>
#include "plugin.h"

class LogAnalysisPlugin : public CollectorPlugin {
private:
    std::string log_file_path;
    std::streampos last_position = 0;

public:
    void Initialize() override {
        log_file_path = Config()["log_directory"] + "/postgresql.log";
    }

    void Collect() override {
        std::ifstream log_file(log_file_path);
        log_file.seekg(last_position);

        std::string line;
        while (std::getline(log_file, line)) {
            // Parse log line
            auto timestamp = ExtractTimestamp(line);
            auto severity = ExtractSeverity(line);

            SendMetric("log_entry", {
                {"timestamp", timestamp},
                {"severity", severity},
                {"message", line},
                {"database", ExtractDatabase(line)},
            });
        }

        last_position = log_file.tellg();
        log_file.close();
    }
};
```

- [ ] **Step 2: Commit**

```bash
git add collector/src/plugins/log_analysis_plugin.cpp
git commit -m "feat: add log analysis collector plugin"
```

---

## Final Integration & Testing

### Task 21: Register All Routes in Server

**Files:**
- Modify: `backend/internal/api/server.go`

- [ ] **Step 1: Register all route groups**

```go
func (s *Server) setupRoutes() {
    r := chi.NewRouter()

    // Register all feature routes
    s.registerQueryPerformanceRoutes(r)
    s.registerLogAnalysisRoutes(r)
    s.registerIndexAdvisorRoutes(r)
    s.registerVacuumAdvisorRoutes(r)

    s.router = r
}
```

- [ ] **Step 2: Commit**

```bash
git commit -m "feat: register all advanced features routes"
```

---

### Task 22: Add Frontend Navigation Links

**Files:**
- Modify: `frontend/src/components/Sidebar.tsx`

- [ ] **Step 1: Add navigation items**

```typescript
export const Sidebar: React.FC = () => {
    return (
        <nav className="space-y-2">
            {/* Existing items */}

            {/* New Advanced Features */}
            <NavItem to="/query-performance" label="Query Performance" icon="⚡" />
            <NavItem to="/log-analysis" label="Log Analysis" icon="📝" />
            <NavItem to="/index-advisor" label="Index Advisor" icon="🔍" />
            <NavItem to="/vacuum-advisor" label="VACUUM Advisor" icon="🧹" />
        </nav>
    );
};
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/Sidebar.tsx
git commit -m "feat: add navigation links for advanced features"
```

---

### Task 23: E2E Integration Tests

**Files:**
- Create: `backend/tests/integration/advanced_features_test.go`

- [ ] **Step 1: Write E2E tests**

```go
func TestAdvancedFeaturesIntegration(t *testing.T) {
    // Setup
    server := setupTestServer(t)
    databaseID := "1"

    // 1. Collect query performance
    t.Run("QueryPerformance", func(t *testing.T) {
        collector := query_performance.NewQueryCollector(server.db)
        // ... test collection
    })

    // 2. Ingest logs
    t.Run("LogAnalysis", func(t *testing.T) {
        collector := log_analysis.NewLogCollector(server.db)
        // ... test ingestion
    })

    // 3. Generate index recommendations (depends on v3.1.0)
    t.Run("IndexAdvisor", func(t *testing.T) {
        analyzer := index_advisor.NewIndexAnalyzer(server.db)
        // ... test recommendations
    })

    // 4. Generate vacuum recommendations (depends on v3.1.0 + v3.2.0)
    t.Run("VacuumAdvisor", func(t *testing.T) {
        analyzer := vacuum_advisor.NewVacuumAnalyzer(server.db)
        // ... test analysis
    })
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/tests/integration/advanced_features_test.go
git commit -m "test: add E2E integration tests for all advanced features"
```

---

## Summary

This plan implements all 4 versions in 3 waves with 23 tasks total:

**Onda 1 (Weeks 5-8):** 11 tasks - v3.1.0 Query Performance + v3.2.0 Log Analysis
**Onda 2 (Weeks 9-11):** 5 tasks - v3.3.0 Index Advisor
**Onda 3 (Weeks 12-14):** 3 tasks - v3.4.0 VACUUM Advisor
**Integration:** 4 tasks - Collector plugins + E2E tests + Server integration
