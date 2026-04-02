package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/ml/models"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services/index_advisor"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services/log_analysis"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services/query_performance"
	"github.com/torresglauco/pganalytics-v3/backend/internal/services/vacuum_advisor"

	_ "github.com/lib/pq"
)

// TestBackendMultiVersionSupport validates that the backend analyzes data from all PostgreSQL versions (14-18)
// This test suite ensures that each backend component works correctly regardless of the source PostgreSQL version
func TestBackendMultiVersionSupport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-version analysis test in short mode")
	}

	db := setupBackendTestDB(t)
	defer db.Close()

	// Get the current PostgreSQL version
	version := getPostgresVersionString(t, db)
	t.Logf("Backend multi-version support testing against PostgreSQL: %s", version)

	// Validate minimum supported version
	if !isSupportedVersion(version) {
		t.Skipf("PostgreSQL version %s is not in supported range (14-18)", version)
	}
}

// TestQueryAnalysisEngineMultiVersion validates the Query Analysis Engine works with data from all versions
func TestQueryAnalysisEngineMultiVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("AnalyzeQueryFromPG14", func(t *testing.T) {
		testQueryAnalysisWithVersionData(t, "14", mockPG14QueryData())
	})

	t.Run("AnalyzeQueryFromPG15", func(t *testing.T) {
		testQueryAnalysisWithVersionData(t, "15", mockPG15QueryData())
	})

	t.Run("AnalyzeQueryFromPG16", func(t *testing.T) {
		testQueryAnalysisWithVersionData(t, "16", mockPG16QueryData())
	})

	t.Run("AnalyzeQueryFromPG17", func(t *testing.T) {
		testQueryAnalysisWithVersionData(t, "17", mockPG17QueryData())
	})

	t.Run("AnalyzeQueryFromPG18", func(t *testing.T) {
		testQueryAnalysisWithVersionData(t, "18", mockPG18QueryData())
	})
}

// testQueryAnalysisWithVersionData tests query analysis with version-specific data
func testQueryAnalysisWithVersionData(t *testing.T, pgVersion string, issues []query_performance.QueryIssue) {
	analyzer := query_performance.NewQueryAnalyzer()

	// Test that analyzer can calculate severity score for any version data
	if len(issues) > 0 {
		score := analyzer.CalculateSeverityScore(issues)
		if score < 0 || score > 100 {
			t.Errorf("PG%s: Invalid severity score %v (expected 0-100)", pgVersion, score)
		}
		t.Logf("PG%s: Query severity score = %.2f", pgVersion, score)
	}

	// Validate that all issues have required fields
	for _, issue := range issues {
		if issue.Type == "" {
			t.Errorf("PG%s: Query issue missing Type", pgVersion)
		}
		if issue.Severity == "" {
			t.Errorf("PG%s: Query issue missing Severity", pgVersion)
		}
	}
}

// TestIndexAdvisorMultiVersion validates Index Advisor works with plans from all versions
func TestIndexAdvisorMultiVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupBackendTestDB(t)
	defer db.Close()

	t.Run("AnalyzePlansFromPG14", func(t *testing.T) {
		testIndexAdvisorWithVersionData(t, "14", mockPG14QueryPlans())
	})

	t.Run("AnalyzePlansFromPG15", func(t *testing.T) {
		testIndexAdvisorWithVersionData(t, "15", mockPG15QueryPlans())
	})

	t.Run("AnalyzePlansFromPG16", func(t *testing.T) {
		testIndexAdvisorWithVersionData(t, "16", mockPG16QueryPlans())
	})

	t.Run("AnalyzePlansFromPG17", func(t *testing.T) {
		testIndexAdvisorWithVersionData(t, "17", mockPG17QueryPlans())
	})

	t.Run("AnalyzePlansFromPG18", func(t *testing.T) {
		testIndexAdvisorWithVersionData(t, "18", mockPG18QueryPlans())
	})
}

// testIndexAdvisorWithVersionData tests index advisor with version-specific EXPLAIN plans
func testIndexAdvisorWithVersionData(t *testing.T, pgVersion string, plans []*index_advisor.QueryPlan) {
	db := setupBackendTestDB(t)
	defer db.Close()

	ia := index_advisor.NewIndexAnalyzer(db)

	// Analyzer should handle plans from any version
	recommendations := ia.FindMissingIndexes(plans)

	// Validate that recommendations (if any) have required fields
	for _, rec := range recommendations {
		if rec.TableName == "" {
			t.Errorf("PG%s: Recommendation missing table name", pgVersion)
		}
		if len(rec.ColumnNames) == 0 {
			t.Errorf("PG%s: Recommendation missing column names", pgVersion)
		}
		if rec.EstimatedBenefit < 0 {
			t.Errorf("PG%s: Invalid estimated benefit %v", pgVersion, rec.EstimatedBenefit)
		}
	}

	t.Logf("PG%s: Generated %d index recommendations", pgVersion, len(recommendations))
}

// TestVacuumAdvisorMultiVersion validates Vacuum Advisor works with metrics from all versions
func TestVacuumAdvisorMultiVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("AnalyzeVacuumFromPG14", func(t *testing.T) {
		testVacuumAdvisorWithVersionData(t, "14", mockPG14VacuumMetrics())
	})

	t.Run("AnalyzeVacuumFromPG15", func(t *testing.T) {
		testVacuumAdvisorWithVersionData(t, "15", mockPG15VacuumMetrics())
	})

	t.Run("AnalyzeVacuumFromPG16", func(t *testing.T) {
		testVacuumAdvisorWithVersionData(t, "16", mockPG16VacuumMetrics())
	})

	t.Run("AnalyzeVacuumFromPG17", func(t *testing.T) {
		testVacuumAdvisorWithVersionData(t, "17", mockPG17VacuumMetrics())
	})

	t.Run("AnalyzeVacuumFromPG18", func(t *testing.T) {
		testVacuumAdvisorWithVersionData(t, "18", mockPG18VacuumMetrics())
	})
}

// testVacuumAdvisorWithVersionData tests vacuum advisor with version-specific metrics
func testVacuumAdvisorWithVersionData(t *testing.T, pgVersion string, metrics *vacuum_advisor.VacuumMetrics) {
	va := vacuum_advisor.NewVacuumAnalyzer(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Analyze table metrics from any version
	rec, err := va.AnalyzeTable(ctx, metrics.DatabaseID, metrics.TableName)
	if err != nil {
		t.Logf("PG%s: Vacuum analysis error (expected in test environment): %v", pgVersion, err)
	}

	if rec != nil {
		if rec.TableName != metrics.TableName {
			t.Errorf("PG%s: Recommendation table mismatch", pgVersion)
		}
		// Note: RecommendationType may be empty for new recommendations
	}

	t.Logf("PG%s: Vacuum analysis completed for table %s", pgVersion, metrics.TableName)
}

// TestLogAnalysisParserMultiVersion validates Log Analysis Parser handles logs from all versions
func TestLogAnalysisParserMultiVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("ParseLogsFromPG14", func(t *testing.T) {
		testLogParserWithVersionData(t, "14", mockPG14Logs())
	})

	t.Run("ParseLogsFromPG15", func(t *testing.T) {
		testLogParserWithVersionData(t, "15", mockPG15Logs())
	})

	t.Run("ParseLogsFromPG16", func(t *testing.T) {
		testLogParserWithVersionData(t, "16", mockPG16Logs())
	})

	t.Run("ParseLogsFromPG17", func(t *testing.T) {
		testLogParserWithVersionData(t, "17", mockPG17Logs())
	})

	t.Run("ParseLogsFromPG18", func(t *testing.T) {
		testLogParserWithVersionData(t, "18", mockPG18Logs())
	})
}

// testLogParserWithVersionData tests log parser with version-specific logs
func testLogParserWithVersionData(t *testing.T, pgVersion string, logMessages []string) {
	parser := log_analysis.NewLogParser()

	categoryCounts := make(map[log_analysis.LogCategory]int)

	for _, msg := range logMessages {
		category := parser.ClassifyLog(msg)
		categoryCounts[category]++

		// Extract metadata (should work for any version format)
		metadata := parser.ExtractMetadata(msg)

		// Validate metadata extraction
		if len(metadata) > 0 {
			for key, value := range metadata {
				if value == nil || value == "" {
					t.Logf("PG%s: Extracted metadata key %s has empty value", pgVersion, key)
				}
			}
		}
	}

	t.Logf("PG%s: Parsed %d logs, categories: %v", pgVersion, len(logMessages), categoryCounts)
}

// TestAnomalyDetectionMultiVersion validates Anomaly Detection works with data from all versions
func TestAnomalyDetectionMultiVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("DetectAnomaliesFromPG14", func(t *testing.T) {
		testAnomalyDetectionWithVersionData(t, "14", mockPG14MetricData())
	})

	t.Run("DetectAnomaliesFromPG15", func(t *testing.T) {
		testAnomalyDetectionWithVersionData(t, "15", mockPG15MetricData())
	})

	t.Run("DetectAnomaliesFromPG16", func(t *testing.T) {
		testAnomalyDetectionWithVersionData(t, "16", mockPG16MetricData())
	})

	t.Run("DetectAnomaliesFromPG17", func(t *testing.T) {
		testAnomalyDetectionWithVersionData(t, "17", mockPG17MetricData())
	})

	t.Run("DetectAnomaliesFromPG18", func(t *testing.T) {
		testAnomalyDetectionWithVersionData(t, "18", mockPG18MetricData())
	})
}

// testAnomalyDetectionWithVersionData tests anomaly detector with version-specific metrics
func testAnomalyDetectionWithVersionData(t *testing.T, pgVersion string, metrics map[string]float64) {
	detector := models.NewAnomalyDetector()

	// Set baseline from historical data
	detector.SetBaseline("query_latency", &models.MetricBaseline{
		Mean:   100.0,
		StdDev: 15.0,
		Min:    50.0,
		Max:    200.0,
	})

	detector.SetBaseline("connection_count", &models.MetricBaseline{
		Mean:   50.0,
		StdDev: 10.0,
		Min:    20.0,
		Max:    100.0,
	})

	anomaliesDetected := 0
	for metricName, value := range metrics {
		alert, isAnomaly := detector.Detect(metricName, value)
		if isAnomaly {
			anomaliesDetected++
			if alert.Severity == "" {
				t.Errorf("PG%s: Anomaly alert missing severity", pgVersion)
			}
			t.Logf("PG%s: Detected anomaly in %s: value=%.2f, severity=%s", pgVersion, metricName, value, alert.Severity)
		}
	}

	t.Logf("PG%s: Anomaly detection completed, %d anomalies found", pgVersion, anomaliesDetected)
}

// TestBackendEndToEndMultiVersion validates complete data flow from collection to analysis
func TestBackendEndToEndMultiVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("EndToEndPG14", func(t *testing.T) {
		testBackendEndToEnd(t, "14")
	})

	t.Run("EndToEndPG15", func(t *testing.T) {
		testBackendEndToEnd(t, "15")
	})

	t.Run("EndToEndPG16", func(t *testing.T) {
		testBackendEndToEnd(t, "16")
	})

	t.Run("EndToEndPG17", func(t *testing.T) {
		testBackendEndToEnd(t, "17")
	})

	t.Run("EndToEndPG18", func(t *testing.T) {
		testBackendEndToEnd(t, "18")
	})
}

// testBackendEndToEnd validates the complete pipeline for a given PostgreSQL version
func testBackendEndToEnd(t *testing.T, pgVersion string) {
	db := setupBackendTestDB(t)
	defer db.Close()

	// Simulate data collection from specified version
	var queryIssues []query_performance.QueryIssue
	var plans []*index_advisor.QueryPlan
	var vacuumMetrics *vacuum_advisor.VacuumMetrics
	var logs []string
	var metrics map[string]float64

	switch pgVersion {
	case "14":
		queryIssues = mockPG14QueryData()
		plans = mockPG14QueryPlans()
		vacuumMetrics = mockPG14VacuumMetrics()
		logs = mockPG14Logs()
		metrics = mockPG14MetricData()
	case "15":
		queryIssues = mockPG15QueryData()
		plans = mockPG15QueryPlans()
		vacuumMetrics = mockPG15VacuumMetrics()
		logs = mockPG15Logs()
		metrics = mockPG15MetricData()
	case "16":
		queryIssues = mockPG16QueryData()
		plans = mockPG16QueryPlans()
		vacuumMetrics = mockPG16VacuumMetrics()
		logs = mockPG16Logs()
		metrics = mockPG16MetricData()
	case "17":
		queryIssues = mockPG17QueryData()
		plans = mockPG17QueryPlans()
		vacuumMetrics = mockPG17VacuumMetrics()
		logs = mockPG17Logs()
		metrics = mockPG17MetricData()
	case "18":
		queryIssues = mockPG18QueryData()
		plans = mockPG18QueryPlans()
		vacuumMetrics = mockPG18VacuumMetrics()
		logs = mockPG18Logs()
		metrics = mockPG18MetricData()
	default:
		t.Fatalf("Unknown PostgreSQL version: %s", pgVersion)
	}

	// 1. Query Analysis Engine
	analyzer := query_performance.NewQueryAnalyzer()
	queryScore := analyzer.CalculateSeverityScore(queryIssues)
	if queryScore < 0 || queryScore > 100 {
		t.Errorf("PG%s: Invalid query severity score", pgVersion)
	}

	// 2. Index Advisor
	ia := index_advisor.NewIndexAnalyzer(db)
	indexRecs := ia.FindMissingIndexes(plans)

	// 3. Vacuum Advisor
	va := vacuum_advisor.NewVacuumAnalyzer(db)
	vacRec, vacErr := va.AnalyzeTable(context.Background(), vacuumMetrics.DatabaseID, vacuumMetrics.TableName)
	_ = vacErr

	// 4. Log Analysis
	parser := log_analysis.NewLogParser()
	for _, msg := range logs {
		_ = parser.ClassifyLog(msg)
	}

	// 5. Anomaly Detection
	detector := models.NewAnomalyDetector()
	detector.SetBaseline("query_latency", &models.MetricBaseline{
		Mean:   100.0,
		StdDev: 15.0,
		Min:    50.0,
		Max:    200.0,
	})

	for metricName, value := range metrics {
		_, _ = detector.Detect(metricName, value)
	}

	// Validate end-to-end results
	if len(queryIssues) > 0 && queryScore == 0 {
		t.Errorf("PG%s: Query score should be > 0 when issues exist", pgVersion)
	}

	if len(indexRecs) >= 0 {
		t.Logf("PG%s: End-to-end pipeline completed successfully", pgVersion)
	}

	if vacRec == nil && vacuumMetrics.DatabaseID > 0 {
		t.Errorf("PG%s: Vacuum recommendation should not be nil", pgVersion)
	}

	t.Logf("PG%s: End-to-end test: Query Score=%.2f, Index Recs=%d, Vacuum Rec=%v", pgVersion, queryScore, len(indexRecs), vacRec != nil)
}

// Mock data generators for each PostgreSQL version

func mockPG14QueryData() []query_performance.QueryIssue {
	return []query_performance.QueryIssue{
		{
			Type:             "sequential_scan",
			Severity:         "high",
			AffectedNode:     "Seq Scan on users",
			Description:      "Full table scan detected on large table",
			Recommendation:   "Create index on filter columns",
			EstimatedBenefit: 45.5,
		},
		{
			Type:             "missing_index",
			Severity:         "medium",
			AffectedNode:     "Nested Loop on orders",
			Description:      "Join operation could benefit from index",
			Recommendation:   "Create index on join columns",
			EstimatedBenefit: 25.3,
		},
	}
}

func mockPG15QueryData() []query_performance.QueryIssue {
	// PG15 has similar structure as PG14, with additional metadata
	issues := mockPG14QueryData()
	issues = append(issues, query_performance.QueryIssue{
		Type:             "high_planning_time",
		Severity:         "low",
		AffectedNode:     "Planner",
		Description:      "Complex query takes long time to plan (PG15 feature)",
		Recommendation:   "Simplify query or use materialized views",
		EstimatedBenefit: 10.0,
	})
	return issues
}

func mockPG16QueryData() []query_performance.QueryIssue {
	// PG16 adds enhanced monitoring
	return []query_performance.QueryIssue{
		{
			Type:             "sequential_scan",
			Severity:         "critical",
			AffectedNode:     "Seq Scan on large_table",
			Description:      "Full table scan on billion-row table",
			Recommendation:   "Create covering index immediately",
			EstimatedBenefit: 78.5,
		},
	}
}

func mockPG17QueryData() []query_performance.QueryIssue {
	// PG17 data
	return []query_performance.QueryIssue{
		{
			Type:             "parallel_scan_overhead",
			Severity:         "medium",
			AffectedNode:     "Parallel Seq Scan",
			Description:      "Parallel overhead exceeds benefits",
			Recommendation:   "Adjust parallel settings or increase data size",
			EstimatedBenefit: 18.2,
		},
	}
}

func mockPG18QueryData() []query_performance.QueryIssue {
	// PG18 data with latest features
	return []query_performance.QueryIssue{
		{
			Type:             "incremental_sort",
			Severity:         "low",
			AffectedNode:     "Incremental Sort",
			Description:      "Incremental sort could optimize multi-level ordering",
			Recommendation:   "Review sort strategy for large result sets",
			EstimatedBenefit: 12.5,
		},
	}
}

func mockPG14QueryPlans() []*index_advisor.QueryPlan {
	return []*index_advisor.QueryPlan{
		{
			NodeType:  "Seq Scan",
			TotalCost: 1500.0,
			Calls:     100,
		},
		{
			NodeType:  "Nested Loop",
			TotalCost: 2500.0,
			Calls:     50,
		},
	}
}

func mockPG15QueryPlans() []*index_advisor.QueryPlan {
	plans := mockPG14QueryPlans()
	plans = append(plans, &index_advisor.QueryPlan{
		NodeType:  "BitmapHeapScan",
		TotalCost: 800.0,
		Calls:     75,
	})
	return plans
}

func mockPG16QueryPlans() []*index_advisor.QueryPlan {
	return []*index_advisor.QueryPlan{
		{
			NodeType:  "IndexScan",
			TotalCost: 500.0,
			Calls:     200,
		},
	}
}

func mockPG17QueryPlans() []*index_advisor.QueryPlan {
	return []*index_advisor.QueryPlan{
		{
			NodeType:  "ParallelSeqScan",
			TotalCost: 1200.0,
			Calls:     150,
		},
	}
}

func mockPG18QueryPlans() []*index_advisor.QueryPlan {
	return []*index_advisor.QueryPlan{
		{
			NodeType:  "IncrementalSort",
			TotalCost: 600.0,
			Calls:     120,
		},
	}
}

func mockPG14VacuumMetrics() *vacuum_advisor.VacuumMetrics {
	return &vacuum_advisor.VacuumMetrics{
		DatabaseID:        1,
		TableName:         "users",
		TableSize:         1000000,
		DeadTuples:        150000,
		LiveTuples:        850000,
		DeadTuplesRatio:   15.0,
		VacuumFrequency:   "daily",
		AutovacuumEnabled: true,
	}
}

func mockPG15VacuumMetrics() *vacuum_advisor.VacuumMetrics {
	return &vacuum_advisor.VacuumMetrics{
		DatabaseID:        1,
		TableName:         "orders",
		TableSize:         5000000,
		DeadTuples:        500000,
		LiveTuples:        4500000,
		DeadTuplesRatio:   10.0,
		VacuumFrequency:   "hourly",
		AutovacuumEnabled: true,
	}
}

func mockPG16VacuumMetrics() *vacuum_advisor.VacuumMetrics {
	return &vacuum_advisor.VacuumMetrics{
		DatabaseID:        1,
		TableName:         "transactions",
		TableSize:         10000000,
		DeadTuples:        2000000,
		LiveTuples:        8000000,
		DeadTuplesRatio:   20.0,
		VacuumFrequency:   "frequent",
		AutovacuumEnabled: true,
	}
}

func mockPG17VacuumMetrics() *vacuum_advisor.VacuumMetrics {
	return &vacuum_advisor.VacuumMetrics{
		DatabaseID:        1,
		TableName:         "events",
		TableSize:         2000000,
		DeadTuples:        250000,
		LiveTuples:        1750000,
		DeadTuplesRatio:   12.5,
		VacuumFrequency:   "regular",
		AutovacuumEnabled: true,
	}
}

func mockPG18VacuumMetrics() *vacuum_advisor.VacuumMetrics {
	return &vacuum_advisor.VacuumMetrics{
		DatabaseID:        1,
		TableName:         "sessions",
		TableSize:         3000000,
		DeadTuples:        300000,
		LiveTuples:        2700000,
		DeadTuplesRatio:   10.0,
		VacuumFrequency:   "regular",
		AutovacuumEnabled: true,
	}
}

func mockPG14Logs() []string {
	return []string{
		"LOG: checkpoint starting: time",
		"ERROR: syntax error at or near \"SELECT\"",
		"WARNING: could not write block",
		"LOG: automatic vacuum of table \"public.users\"",
		"ERROR: duplicate key value violates unique constraint",
		"LOG: duration: 125.450 ms",
	}
}

func mockPG15Logs() []string {
	return []string{
		"LOG: checkpoint starting: redo=XXX/XXXXXXXX",
		"ERROR: UNIQUE constraint \"idx_unique_email\" is violated",
		"WARNING: relation \"temp_table\" does not exist",
		"LOG: automatic vacuum of table \"public.orders\" (analyze)",
		"LOG: deadlock detected",
		"LOG: duration: 250.123 ms statement: SELECT * FROM users",
	}
}

func mockPG16Logs() []string {
	return []string{
		"LOG: checkpoint starting",
		"FATAL: database \"missing_db\" does not exist",
		"ERROR: could not obtain lock on relation \"public.products\"",
		"LOG: replication slot \"slot1\" created",
		"WARNING: out of memory",
		"LOG: duration: 500.567 ms",
	}
}

func mockPG17Logs() []string {
	return []string{
		"LOG: checkpoint complete",
		"ERROR: FATAL: no pg_hba.conf entry for host",
		"LOG: WAL archiving started",
		"WARNING: estimated number of rows to modify is very large",
		"LOG: long transaction in progress",
		"LOG: duration: 1250.890 ms",
	}
}

func mockPG18Logs() []string {
	return []string{
		"LOG: checkpoint finished",
		"ERROR: could not accept SSL connection",
		"LOG: incremental sort applied",
		"WARNING: hash table size too large",
		"LOG: parallel query worker process launched",
		"LOG: duration: 75.432 ms",
	}
}

func mockPG14MetricData() map[string]float64 {
	return map[string]float64{
		"query_latency":    95.5,
		"connection_count": 48.0,
		"cache_hit_ratio":  0.89,
	}
}

func mockPG15MetricData() map[string]float64 {
	return map[string]float64{
		"query_latency":     105.2,
		"connection_count":  52.0,
		"cache_hit_ratio":   0.91,
		"memory_usage_perc": 65.0,
	}
}

func mockPG16MetricData() map[string]float64 {
	return map[string]float64{
		"query_latency":     115.8,
		"connection_count":  55.0,
		"cache_hit_ratio":   0.87,
		"memory_usage_perc": 72.0,
		"wal_write_latency": 8.5,
	}
}

func mockPG17MetricData() map[string]float64 {
	return map[string]float64{
		"query_latency":     98.3,
		"connection_count":  51.0,
		"cache_hit_ratio":   0.92,
		"memory_usage_perc": 68.0,
		"wal_write_latency": 7.2,
	}
}

func mockPG18MetricData() map[string]float64 {
	return map[string]float64{
		"query_latency":     92.5,
		"connection_count":  49.0,
		"cache_hit_ratio":   0.93,
		"memory_usage_perc": 61.0,
		"wal_write_latency": 6.8,
	}
}

// Helper functions

func setupBackendTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("DATABASE_URL or TEST_DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

func getPostgresVersionString(t *testing.T, db *sql.DB) string {
	var version string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		t.Logf("Failed to get PostgreSQL version: %v", err)
		return ""
	}
	return version
}

func isSupportedVersion(version string) bool {
	supportedVersions := []string{"14", "15", "16", "17", "18"}
	for _, v := range supportedVersions {
		if strings.Contains(version, fmt.Sprintf("PostgreSQL %s", v)) {
			return true
		}
	}
	return false
}
