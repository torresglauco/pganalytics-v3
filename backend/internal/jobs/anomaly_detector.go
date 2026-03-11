package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// ANOMALY DETECTOR TYPES
// ============================================================================

// AnomalyDetectionJob manages automated anomaly detection across databases
type AnomalyDetectionJob struct {
	db *sql.DB
	mu sync.RWMutex

	// Job configuration
	enabled              bool
	checkIntervalMinutes int
	baselineWindowHours  int

	// Control channels
	stopChan chan struct{}
	done     chan struct{}

	// Metrics
	lastRun   time.Time
	lastError error
	runCount  int64

	// Thresholds
	zscorerThreshold       float64 // Default 2.5 standard deviations
	anomalySeverityLevels AnomalySeverityConfig
}

// AnomalySeverityConfig defines thresholds for different severity levels
type AnomalySeverityConfig struct {
	CriticalZScore float64 // e.g., 3.0 = 3 sigma
	HighZScore     float64 // e.g., 2.5 = 2.5 sigma
	MediumZScore   float64 // e.g., 1.5 = 1.5 sigma
	LowZScore      float64 // e.g., 1.0 = 1 sigma
}

// QueryBaseline represents statistical baseline for a query metric
type QueryBaseline struct {
	ID                 int64
	DatabaseID         int
	QueryID            int
	MetricName         string
	BaselineMean       float64
	BaselineStdDev     float64
	BaselineMin        float64
	BaselineMax        float64
	BaselineMedian     float64
	BaselineP25        float64
	BaselineP75        float64
	BaselineP90        float64
	BaselineP95        float64
	BaselineP99        float64
	BaselineWindowHrs  int
	DataPoints         int
	CalculatedAt       time.Time
	IsEnabled          bool
}

// DetectedAnomaly represents an anomaly detected in query metrics
type DetectedAnomaly struct {
	ID               int64
	DatabaseID       int
	QueryID          int
	BaselineID       int64
	MetricName       string
	CurrentValue     float64
	BaselineValue    float64
	ZScore           float64
	DeviationPercent float64
	Severity         string // "low", "medium", "high", "critical"
	AnomalyType      string // "statistical", "trend", "seasonal", "pattern"
	DetectionMethod  string
	DetectedAt       time.Time
	IsActive         bool
}

// AnomalyDetectionMetrics contains statistics from a detection run
type AnomalyDetectionMetrics struct {
	TotalAnomaliesDetected   int
	NewAnomalies             int
	ResolvedAnomalies        int
	CriticalAnomalies        int
	HighAnomalies            int
	MediumAnomalies          int
	LowAnomalies             int
	BaselineUpdated          int
	ExecutionTimeMs          int64
	AffectedDatabases        int
	AffectedQueries          int
	Timestamp                time.Time
}

// ============================================================================
// CONSTRUCTOR & LIFECYCLE
// ============================================================================

// NewAnomalyDetectionJob creates a new anomaly detection job
func NewAnomalyDetectionJob(db *sql.DB) *AnomalyDetectionJob {
	return &AnomalyDetectionJob{
		db:                   db,
		enabled:              true,
		checkIntervalMinutes: 5, // Default: check every 5 minutes
		baselineWindowHours:  168, // Default: 7-day rolling window
		zscorerThreshold:     2.5,
		anomalySeverityLevels: AnomalySeverityConfig{
			CriticalZScore: 3.0,  // 3 sigma
			HighZScore:     2.5,  // 2.5 sigma
			MediumZScore:   1.5,  // 1.5 sigma
			LowZScore:      1.0,  // 1 sigma
		},
		stopChan: make(chan struct{}),
		done:     make(chan struct{}),
	}
}

// ============================================================================
// JOB LIFECYCLE METHODS
// ============================================================================

// Start begins the anomaly detection job with periodic scheduling
func (a *AnomalyDetectionJob) Start(ctx context.Context) error {
	log.Println("[AnomalyDetector] Starting anomaly detection job")

	a.mu.Lock()
	a.enabled = true
	a.mu.Unlock()

	// Run initial detection immediately
	go a.runDetection(ctx)

	// Schedule periodic checks
	ticker := time.NewTicker(time.Duration(a.checkIntervalMinutes) * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-a.stopChan:
				close(a.done)
				log.Println("[AnomalyDetector] Stopping anomaly detection job")
				return
			case <-ticker.C:
				a.runDetection(ctx)
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the anomaly detection job
func (a *AnomalyDetectionJob) Stop() error {
	a.mu.Lock()
	a.enabled = false
	a.mu.Unlock()

	select {
	case a.stopChan <- struct{}{}:
	default:
	}

	// Wait for goroutine to finish
	select {
	case <-a.done:
	case <-time.After(5 * time.Second):
		log.Println("[AnomalyDetector] Stop timeout")
	}

	return nil
}

// ============================================================================
// MAIN DETECTION WORKFLOW
// ============================================================================

// runDetection executes the full anomaly detection pipeline
func (a *AnomalyDetectionJob) runDetection(ctx context.Context) {
	startTime := time.Now()
	a.mu.Lock()
	a.lastRun = startTime
	a.runCount++
	a.mu.Unlock()

	log.Printf("[AnomalyDetector] Running detection cycle #%d\n", a.runCount)

	// Get all active databases
	databases, err := a.getActiveDatabases(ctx)
	if err != nil {
		log.Printf("[AnomalyDetector] Error fetching databases: %v\n", err)
		a.mu.Lock()
		a.lastError = err
		a.mu.Unlock()
		return
	}

	if len(databases) == 0 {
		log.Println("[AnomalyDetector] No active databases to check")
		return
	}

	metrics := &AnomalyDetectionMetrics{
		Timestamp: startTime,
	}

	// Process each database in parallel
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, 5) // Max 5 concurrent database checks

	for _, db := range databases {
		wg.Add(1)
		go func(database *models.Database) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			a.processDatabase(ctx, database, metrics)
		}(db)
	}

	wg.Wait()

	// Calculate execution time
	duration := time.Since(startTime)
	metrics.ExecutionTimeMs = duration.Milliseconds()
	metrics.AffectedDatabases = len(databases)

	log.Printf("[AnomalyDetector] Detection cycle completed in %dms\n", metrics.ExecutionTimeMs)
	log.Printf("[AnomalyDetector] Summary: %d anomalies (C:%d H:%d M:%d L:%d)\n",
		metrics.TotalAnomaliesDetected,
		metrics.CriticalAnomalies,
		metrics.HighAnomalies,
		metrics.MediumAnomalies,
		metrics.LowAnomalies)

	a.mu.Lock()
	a.lastError = nil
	a.mu.Unlock()
}

// ============================================================================
// DATABASE & QUERY PROCESSING
// ============================================================================

// getActiveDatabases retrieves all active databases for monitoring
func (a *AnomalyDetectionJob) getActiveDatabases(ctx context.Context) ([]*models.Database, error) {
	query := `
		SELECT id, instance_id, name, created_at, updated_at
		FROM databases
		WHERE is_active = TRUE
		LIMIT 1000
	`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query databases: %w", err)
	}
	defer rows.Close()

	var databases []*models.Database
	for rows.Next() {
		db := &models.Database{}
		if err := rows.Scan(&db.ID, &db.InstanceID, &db.Name, &db.CreatedAt, &db.UpdatedAt); err != nil {
			log.Printf("[AnomalyDetector] Error scanning database: %v\n", err)
			continue
		}
		databases = append(databases, db)
	}

	return databases, rows.Err()
}

// processDatabase runs anomaly detection for a single database
func (a *AnomalyDetectionJob) processDatabase(ctx context.Context, database *models.Database, metrics *AnomalyDetectionMetrics) {
	// Step 1: Update baselines
	updatedCount, err := a.updateBaselines(ctx, database.ID)
	if err != nil {
		log.Printf("[AnomalyDetector] Error updating baselines for DB %d: %v\n", database.ID, err)
		return
	}
	metrics.BaselineUpdated += updatedCount

	// Step 2: Detect anomalies using Z-score method
	anomalies, err := a.detectAnomaliesZScore(ctx, database.ID)
	if err != nil {
		log.Printf("[AnomalyDetector] Error detecting anomalies for DB %d: %v\n", database.ID, err)
		return
	}

	// Step 3: Store anomalies and update status
	for _, anomaly := range anomalies {
		if err := a.storeAnomaly(ctx, anomaly); err != nil {
			log.Printf("[AnomalyDetector] Error storing anomaly: %v\n", err)
			continue
		}

		// Increment severity counter
		switch anomaly.Severity {
		case "critical":
			metrics.CriticalAnomalies++
		case "high":
			metrics.HighAnomalies++
		case "medium":
			metrics.MediumAnomalies++
		case "low":
			metrics.LowAnomalies++
		}
		metrics.TotalAnomaliesDetected++
	}

	// Step 4: Resolve old anomalies
	resolvedCount, err := a.resolveOldAnomalies(ctx, database.ID)
	if err != nil {
		log.Printf("[AnomalyDetector] Error resolving old anomalies for DB %d: %v\n", database.ID, err)
	}
	metrics.ResolvedAnomalies += resolvedCount
}

// ============================================================================
// BASELINE CALCULATION
// ============================================================================

// updateBaselines recalculates statistical baselines for query metrics
func (a *AnomalyDetectionJob) updateBaselines(ctx context.Context, databaseID int) (int, error) {
	// Get all queries in this database
	query := `
		SELECT DISTINCT q.id, q.name
		FROM queries q
		WHERE q.database_id = $1
		  AND q.is_deleted = FALSE
		LIMIT 500
	`

	rows, err := a.db.QueryContext(ctx, query, databaseID)
	if err != nil {
		return 0, fmt.Errorf("fetch queries: %w", err)
	}
	defer rows.Close()

	updatedCount := 0
	for rows.Next() {
		var queryID int
		var queryName string
		if err := rows.Scan(&queryID, &queryName); err != nil {
			log.Printf("[AnomalyDetector] Error scanning query: %v\n", err)
			continue
		}

		// Update baseline for each metric
		for _, metricName := range []string{"execution_time", "calls", "rows_returned"} {
			if err := a.updateQueryBaseline(ctx, databaseID, queryID, metricName); err != nil {
				log.Printf("[AnomalyDetector] Error updating baseline for Q%d.%s: %v\n", queryID, metricName, err)
				continue
			}
			updatedCount++
		}
	}

	return updatedCount, rows.Err()
}

// updateQueryBaseline calculates and stores baseline for a specific query metric
func (a *AnomalyDetectionJob) updateQueryBaseline(ctx context.Context, databaseID, queryID int, metricName string) error {
	baseline, err := a.calculateBaseline(ctx, databaseID, queryID, metricName)
	if err != nil {
		return fmt.Errorf("calculate baseline: %w", err)
	}

	if baseline == nil {
		// Not enough data
		return nil
	}

	// Upsert baseline
	query := `
		INSERT INTO query_baselines(
			database_id, query_id, metric_name,
			baseline_mean, baseline_stddev, baseline_min, baseline_max,
			baseline_median, baseline_p25, baseline_p75, baseline_p90, baseline_p95, baseline_p99,
			baseline_window_hours, baseline_data_points, baseline_calculated_at, is_enabled
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT(database_id, query_id, metric_name) DO UPDATE SET
			baseline_mean = EXCLUDED.baseline_mean,
			baseline_stddev = EXCLUDED.baseline_stddev,
			baseline_min = EXCLUDED.baseline_min,
			baseline_max = EXCLUDED.baseline_max,
			baseline_median = EXCLUDED.baseline_median,
			baseline_p25 = EXCLUDED.baseline_p25,
			baseline_p75 = EXCLUDED.baseline_p75,
			baseline_p90 = EXCLUDED.baseline_p90,
			baseline_p95 = EXCLUDED.baseline_p95,
			baseline_p99 = EXCLUDED.baseline_p99,
			baseline_data_points = EXCLUDED.baseline_data_points,
			baseline_calculated_at = EXCLUDED.baseline_calculated_at
	`

	_, err = a.db.ExecContext(ctx, query,
		databaseID, queryID, metricName,
		baseline.Mean, baseline.StdDev, baseline.Min, baseline.Max,
		baseline.Median, baseline.P25, baseline.P75, baseline.P90, baseline.P95, baseline.P99,
		a.baselineWindowHours, baseline.DataPoints, time.Now())

	return err
}

// calculateBaseline computes statistical metrics from historical query data
func (a *AnomalyDetectionJob) calculateBaseline(ctx context.Context, databaseID, queryID int, metricName string) (*BaselineStats, error) {
	// Map metric names to query_history columns
	var columnName string
	switch metricName {
	case "execution_time":
		columnName = "execution_time_ms"
	case "calls":
		columnName = "calls"
	case "rows_returned":
		columnName = "rows_returned"
	case "rows_affected":
		columnName = "rows_affected"
	case "mean_time":
		columnName = "mean_exec_time"
	default:
		columnName = metricName
	}

	// Fetch historical data within baseline window
	query := fmt.Sprintf(`
		SELECT
			AVG(%[1]s)::NUMERIC,
			STDDEV_POP(%[1]s)::NUMERIC,
			MIN(%[1]s)::NUMERIC,
			MAX(%[1]s)::NUMERIC,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY %[1]s)::NUMERIC,
			PERCENTILE_CONT(0.25) WITHIN GROUP (ORDER BY %[1]s)::NUMERIC,
			PERCENTILE_CONT(0.75) WITHIN GROUP (ORDER BY %[1]s)::NUMERIC,
			PERCENTILE_CONT(0.90) WITHIN GROUP (ORDER BY %[1]s)::NUMERIC,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY %[1]s)::NUMERIC,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY %[1]s)::NUMERIC,
			COUNT(*)::INTEGER
		FROM query_history
		WHERE database_id = $1
		  AND query_id = $2
		  AND collected_at > NOW() - INTERVAL '1 hour' * $3
		  AND %[1]s IS NOT NULL
	`, columnName)

	var stats BaselineStats
	var count int

	err := a.db.QueryRowContext(ctx, query, databaseID, queryID, a.baselineWindowHours).Scan(
		&stats.Mean, &stats.StdDev, &stats.Min, &stats.Max,
		&stats.Median, &stats.P25, &stats.P75, &stats.P90, &stats.P95, &stats.P99,
		&count)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not enough data
		}
		return nil, fmt.Errorf("query baseline: %w", err)
	}

	if count < 10 {
		return nil, nil // Need at least 10 data points
	}

	stats.DataPoints = count
	return &stats, nil
}

// BaselineStats holds calculated statistical metrics
type BaselineStats struct {
	Mean       float64
	StdDev     float64
	Min        float64
	Max        float64
	Median     float64
	P25        float64
	P75        float64
	P90        float64
	P95        float64
	P99        float64
	DataPoints int
}

// ============================================================================
// ANOMALY DETECTION METHODS
// ============================================================================

// detectAnomaliesZScore detects anomalies using statistical Z-score method
func (a *AnomalyDetectionJob) detectAnomaliesZScore(ctx context.Context, databaseID int) ([]*DetectedAnomaly, error) {
	query := `
		SELECT
			qb.id,
			qb.query_id,
			qb.metric_name,
			qh.execution_time_ms,
			qb.baseline_mean,
			qb.baseline_stddev
		FROM query_history qh
		JOIN query_baselines qb ON qh.query_id = qb.query_id AND qh.database_id = qb.database_id
		WHERE qh.database_id = $1
		  AND qh.collected_at > NOW() - INTERVAL '1 hour'
		  AND qb.is_enabled = TRUE
		  AND qb.baseline_stddev > 0
		LIMIT 5000
	`

	rows, err := a.db.QueryContext(ctx, query, databaseID)
	if err != nil {
		return nil, fmt.Errorf("query anomalies: %w", err)
	}
	defer rows.Close()

	var anomalies []*DetectedAnomaly
	for rows.Next() {
		var baselineID int64
		var queryID int
		var metricName string
		var currentValue float64
		var baselineMean float64
		var baselineStdDev float64

		if err := rows.Scan(&baselineID, &queryID, &metricName, &currentValue, &baselineMean, &baselineStdDev); err != nil {
			log.Printf("[AnomalyDetector] Error scanning row: %v\n", err)
			continue
		}

		// Calculate Z-score
		var zScore float64
		if baselineStdDev > 0 {
			zScore = (currentValue - baselineMean) / baselineStdDev
		}

		// Check if anomalous
		if math.Abs(zScore) < 1.0 {
			continue // Not anomalous
		}

		// Determine severity
		severity := a.classifySeverity(math.Abs(zScore))

		// Calculate deviation percentage
		var deviationPercent float64
		if baselineMean != 0 {
			deviationPercent = ((currentValue - baselineMean) / baselineMean) * 100
		}

		anomalies = append(anomalies, &DetectedAnomaly{
			DatabaseID:       databaseID,
			QueryID:          queryID,
			BaselineID:       baselineID,
			MetricName:       metricName,
			CurrentValue:     currentValue,
			BaselineValue:    baselineMean,
			ZScore:           zScore,
			DeviationPercent: deviationPercent,
			Severity:         severity,
			AnomalyType:      "statistical",
			DetectionMethod:  "z-score",
			DetectedAt:       time.Now(),
		})
	}

	return anomalies, rows.Err()
}

// classifySeverity determines severity based on Z-score magnitude
func (a *AnomalyDetectionJob) classifySeverity(absZScore float64) string {
	if absZScore >= a.anomalySeverityLevels.CriticalZScore {
		return "critical"
	} else if absZScore >= a.anomalySeverityLevels.HighZScore {
		return "high"
	} else if absZScore >= a.anomalySeverityLevels.MediumZScore {
		return "medium"
	}
	return "low"
}

// ============================================================================
// ANOMALY STORAGE & STATE MANAGEMENT
// ============================================================================

// storeAnomaly saves a detected anomaly or updates existing one
func (a *AnomalyDetectionJob) storeAnomaly(ctx context.Context, anomaly *DetectedAnomaly) error {
	query := `
		INSERT INTO query_anomalies(
			database_id, query_id, baseline_id, metric_name,
			current_value, baseline_value, z_score, deviation_percent,
			severity, anomaly_type, detection_method, is_active,
			detected_at, first_seen_at, last_seen_at
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT(database_id, query_id, baseline_id) DO UPDATE SET
			current_value = EXCLUDED.current_value,
			z_score = EXCLUDED.z_score,
			deviation_percent = EXCLUDED.deviation_percent,
			severity = EXCLUDED.severity,
			last_seen_at = EXCLUDED.last_seen_at,
			is_active = TRUE,
			detected_at = EXCLUDED.detected_at
		WHERE query_anomalies.is_active = TRUE
	`

	_, err := a.db.ExecContext(ctx, query,
		anomaly.DatabaseID, anomaly.QueryID, anomaly.BaselineID, anomaly.MetricName,
		anomaly.CurrentValue, anomaly.BaselineValue, anomaly.ZScore, anomaly.DeviationPercent,
		anomaly.Severity, anomaly.AnomalyType, anomaly.DetectionMethod, true,
		anomaly.DetectedAt, anomaly.DetectedAt, anomaly.DetectedAt)

	return err
}

// resolveOldAnomalies marks anomalies as resolved if condition no longer true
func (a *AnomalyDetectionJob) resolveOldAnomalies(ctx context.Context, databaseID int) (int, error) {
	// Mark anomalies as resolved if they haven't been seen in last 2 hours
	query := `
		UPDATE query_anomalies
		SET is_active = FALSE, resolved_at = NOW()
		WHERE database_id = $1
		  AND is_active = TRUE
		  AND last_seen_at < NOW() - INTERVAL '2 hours'
	`

	result, err := a.db.ExecContext(ctx, query, databaseID)
	if err != nil {
		return 0, fmt.Errorf("resolve anomalies: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

// ============================================================================
// STATUS & MONITORING
// ============================================================================

// GetStatus returns current job status and metrics
func (a *AnomalyDetectionJob) GetStatus() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":       a.enabled,
		"last_run":      a.lastRun,
		"run_count":     a.runCount,
		"interval_mins": a.checkIntervalMinutes,
	}

	if a.lastError != nil {
		status["last_error"] = a.lastError.Error()
	}

	return status
}

// SetCheckInterval updates the anomaly detection check interval
func (a *AnomalyDetectionJob) SetCheckInterval(minutes int) {
	if minutes < 1 {
		minutes = 1
	}
	if minutes > 60 {
		minutes = 60
	}

	a.mu.Lock()
	a.checkIntervalMinutes = minutes
	a.mu.Unlock()
}

// SetBaselineWindow updates the baseline calculation window
func (a *AnomalyDetectionJob) SetBaselineWindow(hours int) {
	if hours < 24 {
		hours = 24
	}
	if hours > 730 { // 30 days
		hours = 730
	}

	a.mu.Lock()
	a.baselineWindowHours = hours
	a.mu.Unlock()
}

// SetZScoreThreshold updates the Z-score threshold for anomaly detection
func (a *AnomalyDetectionJob) SetZScoreThreshold(threshold float64) {
	if threshold < 0.5 {
		threshold = 0.5
	}
	if threshold > 5.0 {
		threshold = 5.0
	}

	a.mu.Lock()
	a.zscorerThreshold = threshold
	a.mu.Unlock()
}
