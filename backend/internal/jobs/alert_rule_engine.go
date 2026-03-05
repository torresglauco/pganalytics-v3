package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// ALERT RULE ENGINE TYPES
// ============================================================================

// AlertRuleEngineJob manages automated alert rule evaluation
type AlertRuleEngineJob struct {
	db *sql.DB
	mu sync.RWMutex

	// Job configuration
	enabled              bool
	checkIntervalSeconds int
	maxConcurrentRules   int

	// Control channels
	stopChan chan struct{}
	done     chan struct{}

	// Metrics
	lastRun   time.Time
	lastError error
	runCount  int64

	// Rule cache for performance
	ruleCache       map[int64]*AlertRule
	ruleCacheTTL    time.Duration
	lastCacheUpdate time.Time
}

// AlertRule represents an alert rule definition
type AlertRule struct {
	ID                   int64
	UserID               int
	Name                 string
	Description          string
	RuleType             string // "threshold", "change", "anomaly", "composite"
	DatabaseID           *int
	QueryID              *int
	MetricName           string
	Condition            json.RawMessage // JSON condition definition
	AlertSeverity        string          // "low", "medium", "high", "critical"
	EvaluationInterval   int             // seconds
	ForDurationSeconds   int             // trigger only if true for N seconds
	NotificationEnabled  bool
	NotificationChannels []int64         // Channel IDs to notify
	IsEnabled            bool
	IsPaused             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// RuleCondition is the interface for different condition types
type RuleCondition interface {
	Evaluate(ctx context.Context, db *sql.DB, rule *AlertRule) (bool, interface{}, error)
	Type() string
}

// ThresholdCondition represents a simple threshold rule
type ThresholdCondition struct {
	Metric   string      `json:"metric"`
	Operator string      `json:"operator"` // "==", "!=", ">", ">=", "<", "<="
	Value    float64     `json:"value"`
	Unit     string      `json:"unit,omitempty"`
}

// AnomalyCondition represents a rule triggered by anomalies
type AnomalyCondition struct {
	Severity string `json:"severity"` // "low", "medium", "high", "critical"
	Within   int    `json:"within"`   // minutes
}

// ChangeCondition represents change detection rule
type ChangeCondition struct {
	Metric           string  `json:"metric"`
	ChangePercent    float64 `json:"change_percent"`
	ComparisonPeriod string  `json:"comparison_period"` // "5m", "1h", "1d"
}

// CompositeCondition represents combination of rules
type CompositeCondition struct {
	Operator string           `json:"operator"` // "AND", "OR"
	Rules    []json.RawMessage `json:"rules"`
}

// RuleEvaluationResult contains evaluation outcome
type RuleEvaluationResult struct {
	RuleID         int64
	ConditionMet   bool
	CurrentValue   float64
	ThresholdValue float64
	ExecutionMs    int64
	ErrorMessage   string
	EvaluatedAt    time.Time
}

// FiredAlert represents an alert that was triggered
type FiredAlert struct {
	ID             int64
	RuleID         int64
	Title          string
	Description    string
	Severity       string
	DatabaseID     *int
	QueryID        *int
	Context        json.RawMessage
	Status         string // "firing", "alerting", "resolved", "acknowledged"
	Fingerprint    string
	FiredAt        time.Time
	NotificationID *int64
}

// ============================================================================
// CONSTRUCTOR & LIFECYCLE
// ============================================================================

// NewAlertRuleEngineJob creates a new alert rule engine job
func NewAlertRuleEngineJob(db *sql.DB) *AlertRuleEngineJob {
	return &AlertRuleEngineJob{
		db:                   db,
		enabled:              true,
		checkIntervalSeconds: 300, // Default: check every 5 minutes
		maxConcurrentRules:   10,  // Max 10 concurrent rule evaluations
		stopChan:             make(chan struct{}),
		done:                 make(chan struct{}),
		ruleCache:            make(map[int64]*AlertRule),
		ruleCacheTTL:         5 * time.Minute,
	}
}

// ============================================================================
// JOB LIFECYCLE METHODS
// ============================================================================

// Start begins the alert rule evaluation job
func (e *AlertRuleEngineJob) Start(ctx context.Context) error {
	log.Println("[AlertEngine] Starting alert rule engine job")

	e.mu.Lock()
	e.enabled = true
	e.mu.Unlock()

	// Run initial evaluation
	go e.evaluateRules(ctx)

	// Schedule periodic checks
	ticker := time.NewTicker(time.Duration(e.checkIntervalSeconds) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-e.stopChan:
				close(e.done)
				log.Println("[AlertEngine] Stopping alert rule engine job")
				return
			case <-ticker.C:
				e.evaluateRules(ctx)
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the alert rule engine
func (e *AlertRuleEngineJob) Stop() error {
	e.mu.Lock()
	e.enabled = false
	e.mu.Unlock()

	select {
	case e.stopChan <- struct{}{}:
	default:
	}

	select {
	case <-e.done:
	case <-time.After(5 * time.Second):
		log.Println("[AlertEngine] Stop timeout")
	}

	return nil
}

// ============================================================================
// MAIN EVALUATION WORKFLOW
// ============================================================================

// evaluateRules evaluates all enabled alert rules
func (e *AlertRuleEngineJob) evaluateRules(ctx context.Context) {
	startTime := time.Now()
	e.mu.Lock()
	e.lastRun = startTime
	e.runCount++
	e.mu.Unlock()

	log.Printf("[AlertEngine] Running evaluation cycle #%d\n", e.runCount)

	// Step 1: Load or refresh rule cache
	rules, err := e.loadRules(ctx)
	if err != nil {
		log.Printf("[AlertEngine] Error loading rules: %v\n", err)
		e.mu.Lock()
		e.lastError = err
		e.mu.Unlock()
		return
	}

	if len(rules) == 0 {
		log.Println("[AlertEngine] No active rules to evaluate")
		return
	}

	log.Printf("[AlertEngine] Evaluating %d rules\n", len(rules))

	// Step 2: Evaluate rules in parallel with concurrency control
	semaphore := make(chan struct{}, e.maxConcurrentRules)
	wg := sync.WaitGroup{}
	resultsChannel := make(chan *RuleEvaluationResult, len(rules))

	for _, rule := range rules {
		wg.Add(1)
		go func(r *AlertRule) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			result := e.evaluateRule(ctx, r)
			resultsChannel <- result
		}(rule)
	}

	// Step 3: Collect results and store evaluations
	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	firedAlerts := 0
	for result := range resultsChannel {
		if err := e.storeEvaluation(ctx, result); err != nil {
			log.Printf("[AlertEngine] Error storing evaluation for rule %d: %v\n", result.RuleID, err)
			continue
		}

		// If condition met, fire alert
		if result.ConditionMet {
			if err := e.fireAlert(ctx, result); err != nil {
				log.Printf("[AlertEngine] Error firing alert for rule %d: %v\n", result.RuleID, err)
				continue
			}
			firedAlerts++
		}
	}

	duration := time.Since(startTime)
	log.Printf("[AlertEngine] Evaluation cycle completed in %dms (%d alerts fired)\n", duration.Milliseconds(), firedAlerts)

	e.mu.Lock()
	e.lastError = nil
	e.mu.Unlock()
}

// ============================================================================
// RULE LOADING & CACHING
// ============================================================================

// loadRules fetches enabled rules from database with caching
func (e *AlertRuleEngineJob) loadRules(ctx context.Context) ([]*AlertRule, error) {
	// Check cache validity
	e.mu.RLock()
	cacheValid := time.Since(e.lastCacheUpdate) < e.ruleCacheTTL
	cachedRules := e.ruleCache
	e.mu.RUnlock()

	if cacheValid && len(cachedRules) > 0 {
		rules := make([]*AlertRule, 0, len(cachedRules))
		for _, r := range cachedRules {
			if r.IsEnabled && !r.IsPaused {
				rules = append(rules, r)
			}
		}
		return rules, nil
	}

	// Load from database
	query := `
		SELECT id, user_id, name, description, rule_type,
		       database_id, query_id, metric_name, condition,
		       alert_severity, evaluation_interval_seconds, for_duration_seconds,
		       notification_enabled, is_enabled, is_paused,
		       created_at, updated_at
		FROM alert_rules
		WHERE is_enabled = TRUE
		  AND is_paused = FALSE
		  AND deleted_at IS NULL
		LIMIT 1000
	`

	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}
	defer rows.Close()

	rules := make([]*AlertRule, 0)
	newCache := make(map[int64]*AlertRule)

	for rows.Next() {
		rule := &AlertRule{}
		if err := rows.Scan(
			&rule.ID, &rule.UserID, &rule.Name, &rule.Description, &rule.RuleType,
			&rule.DatabaseID, &rule.QueryID, &rule.MetricName, &rule.Condition,
			&rule.AlertSeverity, &rule.EvaluationInterval, &rule.ForDurationSeconds,
			&rule.NotificationEnabled, &rule.IsEnabled, &rule.IsPaused,
			&rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			log.Printf("[AlertEngine] Error scanning rule: %v\n", err)
			continue
		}

		rules = append(rules, rule)
		newCache[rule.ID] = rule
	}

	// Update cache
	e.mu.Lock()
	e.ruleCache = newCache
	e.lastCacheUpdate = time.Now()
	e.mu.Unlock()

	return rules, rows.Err()
}

// ============================================================================
// RULE EVALUATION
// ============================================================================

// evaluateRule evaluates a single alert rule
func (e *AlertRuleEngineJob) evaluateRule(ctx context.Context, rule *AlertRule) *RuleEvaluationResult {
	startTime := time.Now()
	result := &RuleEvaluationResult{
		RuleID:      rule.ID,
		EvaluatedAt: startTime,
	}

	// Parse condition based on rule type
	condition, err := e.parseCondition(rule)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Parse condition: %v", err)
		return result
	}

	// Evaluate condition
	conditionMet, contextData, err := condition.Evaluate(ctx, e.db, rule)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("Evaluate condition: %v", err)
		return result
	}

	result.ConditionMet = conditionMet
	result.ExecutionMs = time.Since(startTime).Milliseconds()

	// Extract numeric values from context if available
	if ctxMap, ok := contextData.(map[string]interface{}); ok {
		if curr, ok := ctxMap["current"]; ok {
			if v, ok := curr.(float64); ok {
				result.CurrentValue = v
			}
		}
		if thresh, ok := ctxMap["threshold"]; ok {
			if v, ok := thresh.(float64); ok {
				result.ThresholdValue = v
			}
		}
	}

	log.Printf("[AlertEngine] Rule %d (%s): condition=%v, exectime=%dms\n",
		rule.ID, rule.Name, conditionMet, result.ExecutionMs)

	return result
}

// parseCondition parses rule condition based on type
func (e *AlertRuleEngineJob) parseCondition(rule *AlertRule) (RuleCondition, error) {
	switch rule.RuleType {
	case "threshold":
		var cond ThresholdCondition
		if err := json.Unmarshal(rule.Condition, &cond); err != nil {
			return nil, fmt.Errorf("unmarshal threshold: %w", err)
		}
		return &cond, nil

	case "anomaly":
		var cond AnomalyCondition
		if err := json.Unmarshal(rule.Condition, &cond); err != nil {
			return nil, fmt.Errorf("unmarshal anomaly: %w", err)
		}
		return &cond, nil

	case "change":
		var cond ChangeCondition
		if err := json.Unmarshal(rule.Condition, &cond); err != nil {
			return nil, fmt.Errorf("unmarshal change: %w", err)
		}
		return &cond, nil

	case "composite":
		var cond CompositeCondition
		if err := json.Unmarshal(rule.Condition, &cond); err != nil {
			return nil, fmt.Errorf("unmarshal composite: %w", err)
		}
		return &cond, nil

	default:
		return nil, fmt.Errorf("unknown rule type: %s", rule.RuleType)
	}
}

// ============================================================================
// CONDITION EVALUATION IMPLEMENTATIONS
// ============================================================================

// Type returns the condition type
func (t *ThresholdCondition) Type() string {
	return "threshold"
}

// Evaluate checks if metric meets threshold
func (t *ThresholdCondition) Evaluate(ctx context.Context, db *sql.DB, rule *AlertRule) (bool, interface{}, error) {
	if rule.QueryID == nil || rule.DatabaseID == nil {
		return false, nil, fmt.Errorf("threshold rule missing database or query")
	}

	// Get latest metric value
	var currentValue float64
	query := `
		SELECT execution_time_ms
		FROM query_history
		WHERE database_id = $1 AND query_id = $2
		ORDER BY collected_at DESC
		LIMIT 1
	`

	err := db.QueryRowContext(ctx, query, *rule.DatabaseID, *rule.QueryID).Scan(&currentValue)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil // No data yet
		}
		return false, nil, fmt.Errorf("query latest metric: %w", err)
	}

	// Compare with threshold
	conditionMet := evaluateOperator(currentValue, t.Value, t.Operator)

	return conditionMet, map[string]interface{}{
		"current":   currentValue,
		"threshold": t.Value,
	}, nil
}

// Type returns the condition type
func (a *AnomalyCondition) Type() string {
	return "anomaly"
}

// Evaluate checks for recent anomalies
func (a *AnomalyCondition) Evaluate(ctx context.Context, db *sql.DB, rule *AlertRule) (bool, interface{}, error) {
	query := `
		SELECT COUNT(*)
		FROM query_anomalies
		WHERE is_active = TRUE
		  AND severity >= $1
		  AND detected_at > NOW() - INTERVAL '1 minute' * $2
	`

	var count int
	severityLevel := mapSeverityLevel(a.Severity)

	err := db.QueryRowContext(ctx, query, severityLevel, a.Within).Scan(&count)
	if err != nil {
		return false, nil, fmt.Errorf("query anomalies: %w", err)
	}

	return count > 0, map[string]interface{}{
		"anomaly_count": count,
	}, nil
}

// Type returns the condition type
func (c *ChangeCondition) Type() string {
	return "change"
}

// Evaluate checks for metric change
func (c *ChangeCondition) Evaluate(ctx context.Context, db *sql.DB, rule *AlertRule) (bool, interface{}, error) {
	if rule.QueryID == nil || rule.DatabaseID == nil {
		return false, nil, fmt.Errorf("change rule missing database or query")
	}

	// Get comparison period duration
	interval := parseDuration(c.ComparisonPeriod)

	query := `
		SELECT
			(SELECT execution_time_ms FROM query_history
			 WHERE database_id = $1 AND query_id = $2
			 ORDER BY collected_at DESC LIMIT 1) as current,
			(SELECT execution_time_ms FROM query_history
			 WHERE database_id = $1 AND query_id = $2
			 AND collected_at < NOW() - INTERVAL '1' || $3
			 ORDER BY collected_at DESC LIMIT 1) as previous
	`

	var current *float64
	var previous *float64

	err := db.QueryRowContext(ctx, query, *rule.DatabaseID, *rule.QueryID, interval).Scan(&current, &previous)
	if err != nil && err != sql.ErrNoRows {
		return false, nil, fmt.Errorf("query change: %w", err)
	}

	if current == nil || previous == nil {
		return false, nil, nil // Not enough data
	}

	// Calculate percentage change
	changePercent := ((*current - *previous) / *previous) * 100
	conditionMet := changePercent >= c.ChangePercent

	return conditionMet, map[string]interface{}{
		"current":        *current,
		"previous":       *previous,
		"change_percent": changePercent,
	}, nil
}

// Type returns the condition type
func (c *CompositeCondition) Type() string {
	return "composite"
}

// Evaluate recursively evaluates composite conditions
func (c *CompositeCondition) Evaluate(ctx context.Context, db *sql.DB, rule *AlertRule) (bool, interface{}, error) {
	results := make([]bool, len(c.Rules))

	for i, conditionJSON := range c.Rules {
		// Recursively evaluate each sub-condition
		// This is simplified - in production, would need full parsing
		var cond map[string]interface{}
		if err := json.Unmarshal(conditionJSON, &cond); err != nil {
			return false, nil, fmt.Errorf("unmarshal sub-condition: %w", err)
		}

		// For now, return true if any condition - in production would evaluate each
		results[i] = true
	}

	var conditionMet bool
	if c.Operator == "AND" {
		conditionMet = true
		for _, r := range results {
			conditionMet = conditionMet && r
		}
	} else { // OR
		for _, r := range results {
			conditionMet = conditionMet || r
		}
	}

	return conditionMet, map[string]interface{}{
		"operator": c.Operator,
		"results":  results,
	}, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// evaluateOperator compares values using the specified operator
func evaluateOperator(current, threshold float64, operator string) bool {
	switch operator {
	case "==":
		return current == threshold
	case "!=":
		return current != threshold
	case ">":
		return current > threshold
	case ">=":
		return current >= threshold
	case "<":
		return current < threshold
	case "<=":
		return current <= threshold
	default:
		return false
	}
}

// mapSeverityLevel converts severity string to comparable level
func mapSeverityLevel(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

// parseDuration converts duration string to interval
func parseDuration(s string) string {
	switch s {
	case "5m":
		return "5 minutes"
	case "1h":
		return "1 hour"
	case "1d":
		return "1 day"
	default:
		return "1 hour"
	}
}

// ============================================================================
// EVALUATION STORAGE
// ============================================================================

// storeEvaluation stores the rule evaluation result in database
func (e *AlertRuleEngineJob) storeEvaluation(ctx context.Context, result *RuleEvaluationResult) error {
	query := `
		INSERT INTO alert_rule_evaluations(
			rule_id, condition_met, current_value, threshold_value,
			evaluation_timestamp, execution_time_ms, error_message
		) VALUES($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := e.db.ExecContext(ctx, query,
		result.RuleID, result.ConditionMet,
		nullableFloat(result.CurrentValue),
		nullableFloat(result.ThresholdValue),
		result.EvaluatedAt,
		result.ExecutionMs,
		nullableString(result.ErrorMessage))

	return err
}

// nullableFloat returns sql.NullFloat64
func nullableFloat(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}

// nullableString returns sql.NullString
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// ============================================================================
// ALERT FIRING
// ============================================================================

// fireAlert creates an alert from a fired rule
func (e *AlertRuleEngineJob) fireAlert(ctx context.Context, result *RuleEvaluationResult) error {
	// Fetch the rule for context
	rule, ok := e.ruleCache[result.RuleID]
	if !ok {
		return fmt.Errorf("rule not found in cache: %d", result.RuleID)
	}

	// Generate alert fingerprint for deduplication
	fingerprint := generateFingerprint(rule, result)

	// Check if identical alert already firing
	existingAlert, err := e.findExistingAlert(ctx, fingerprint)
	if err == nil && existingAlert != nil {
		// Update last_notified if needed
		return nil // Already firing
	}

	// Create alert
	alert := &FiredAlert{
		RuleID:      rule.ID,
		Title:       fmt.Sprintf("Alert: %s", rule.Name),
		Description: rule.Description,
		Severity:    rule.AlertSeverity,
		DatabaseID:  rule.DatabaseID,
		QueryID:     rule.QueryID,
		Status:      "firing",
		FiredAt:     time.Now(),
		Fingerprint: fingerprint,
	}

	// Store context as JSON
	contextMap := map[string]interface{}{
		"rule_id":          rule.ID,
		"rule_name":        rule.Name,
		"current_value":    result.CurrentValue,
		"threshold_value":  result.ThresholdValue,
		"evaluation_time":  result.EvaluatedAt,
	}
	contextJSON, _ := json.Marshal(contextMap)
	alert.Context = contextJSON

	// Insert alert
	query := `
		INSERT INTO alerts(
			rule_id, title, description, severity, database_id, query_id,
			context, status, fingerprint, fired_at
		) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	err = e.db.QueryRowContext(ctx, query,
		alert.RuleID, alert.Title, alert.Description, alert.Severity,
		alert.DatabaseID, alert.QueryID, alert.Context, alert.Status,
		alert.Fingerprint, alert.FiredAt).Scan(&alert.ID)

	if err != nil {
		return fmt.Errorf("insert alert: %w", err)
	}

	log.Printf("[AlertEngine] Alert fired: rule=%d, severity=%s, fingerprint=%s\n",
		rule.ID, alert.Severity, fingerprint)

	return nil
}

// findExistingAlert checks if alert with fingerprint already exists
func (e *AlertRuleEngineJob) findExistingAlert(ctx context.Context, fingerprint string) (*FiredAlert, error) {
	query := `
		SELECT id, rule_id, title, description, severity, status, fired_at
		FROM alerts
		WHERE fingerprint = $1
		  AND status IN ('firing', 'alerting')
		LIMIT 1
	`

	alert := &FiredAlert{}
	err := e.db.QueryRowContext(ctx, query, fingerprint).Scan(
		&alert.ID, &alert.RuleID, &alert.Title, &alert.Description,
		&alert.Severity, &alert.Status, &alert.FiredAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query existing alert: %w", err)
	}

	return alert, nil
}

// generateFingerprint creates a unique fingerprint for alert deduplication
func generateFingerprint(rule *AlertRule, result *RuleEvaluationResult) string {
	// Simple fingerprint: hash of rule_id + severity
	// In production, would use crypto/sha256
	return fmt.Sprintf("%d_%s_%v", rule.ID, rule.AlertSeverity, result.ConditionMet)
}

// ============================================================================
// STATUS & MONITORING
// ============================================================================

// GetStatus returns current engine status
func (e *AlertRuleEngineJob) GetStatus() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return map[string]interface{}{
		"enabled":              e.enabled,
		"last_run":             e.lastRun,
		"run_count":            e.runCount,
		"check_interval_secs":  e.checkIntervalSeconds,
		"cached_rules":         len(e.ruleCache),
		"max_concurrent_rules": e.maxConcurrentRules,
	}
}

// SetCheckInterval updates the evaluation check interval
func (e *AlertRuleEngineJob) SetCheckInterval(seconds int) {
	if seconds < 60 {
		seconds = 60
	}
	if seconds > 3600 {
		seconds = 3600
	}

	e.mu.Lock()
	e.checkIntervalSeconds = seconds
	e.mu.Unlock()
}

// SetMaxConcurrentRules updates max concurrent rule evaluations
func (e *AlertRuleEngineJob) SetMaxConcurrentRules(count int) {
	if count < 1 {
		count = 1
	}
	if count > 100 {
		count = 100
	}

	e.mu.Lock()
	e.maxConcurrentRules = count
	e.mu.Unlock()
}
