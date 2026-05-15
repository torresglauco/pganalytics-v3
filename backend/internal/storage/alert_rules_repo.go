package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// AlertRule represents an alert rule definition
// Matches the struct in backend/internal/jobs/alert_rule_engine.go
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
	NotificationChannels []int64 // Channel IDs to notify
	IsEnabled            bool
	IsPaused             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// AlertTrigger represents a triggered alert
type AlertTrigger struct {
	ID          int64
	AlertID     int64
	InstanceID  int
	TriggeredAt time.Time
	Status      string // "firing", "acknowledged", "resolved"
	CreatedAt   time.Time
}

// AlertRulesRepository provides CRUD operations for alert rules
type AlertRulesRepository struct {
	db *sql.DB
}

// NewAlertRulesRepository creates a new AlertRulesRepository
func NewAlertRulesRepository(db *sql.DB) *AlertRulesRepository {
	return &AlertRulesRepository{db: db}
}

// CreateRule inserts a new alert rule and returns the generated ID
func (r *AlertRulesRepository) CreateRule(rule *AlertRule) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO alert_rules (
			user_id, name, description, rule_type, database_id, query_id,
			metric_name, condition, alert_severity, evaluation_interval_seconds,
			for_duration_seconds, notification_enabled, notification_channels,
			is_enabled, is_paused, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id
	`

	now := time.Now()
	var id int64

	err := r.db.QueryRowContext(ctx, query,
		rule.UserID,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.DatabaseID,
		rule.QueryID,
		rule.MetricName,
		rule.Condition,
		rule.AlertSeverity,
		rule.EvaluationInterval,
		rule.ForDurationSeconds,
		rule.NotificationEnabled,
		pq.Array(rule.NotificationChannels),
		rule.IsEnabled,
		rule.IsPaused,
		now,
		now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("create alert rule: %w", err)
	}

	rule.ID = id
	rule.CreatedAt = now
	rule.UpdatedAt = now

	return id, nil
}

// GetRuleByID retrieves an alert rule by its ID
func (r *AlertRulesRepository) GetRuleByID(id int64) (*AlertRule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, name, description, rule_type, database_id, query_id,
			   metric_name, condition, alert_severity, evaluation_interval_seconds,
			   for_duration_seconds, notification_enabled, notification_channels,
			   is_enabled, is_paused, created_at, updated_at
		FROM alert_rules
		WHERE id = $1 AND deleted_at IS NULL
	`

	rule := &AlertRule{}
	var notificationChannels pq.Int64Array

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.UserID,
		&rule.Name,
		&rule.Description,
		&rule.RuleType,
		&rule.DatabaseID,
		&rule.QueryID,
		&rule.MetricName,
		&rule.Condition,
		&rule.AlertSeverity,
		&rule.EvaluationInterval,
		&rule.ForDurationSeconds,
		&rule.NotificationEnabled,
		&notificationChannels,
		&rule.IsEnabled,
		&rule.IsPaused,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("alert rule not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get alert rule by id: %w", err)
	}

	rule.NotificationChannels = []int64(notificationChannels)
	return rule, nil
}

// ListRules retrieves alert rules for a user with pagination
func (r *AlertRulesRepository) ListRules(userID int, limit, offset int) ([]*AlertRule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, name, description, rule_type, database_id, query_id,
			   metric_name, condition, alert_severity, evaluation_interval_seconds,
			   for_duration_seconds, notification_enabled, notification_channels,
			   is_enabled, is_paused, created_at, updated_at
		FROM alert_rules
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list alert rules: %w", err)
	}
	defer rows.Close()

	rules := make([]*AlertRule, 0)
	for rows.Next() {
		rule := &AlertRule{}
		var notificationChannels pq.Int64Array

		if err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.Name,
			&rule.Description,
			&rule.RuleType,
			&rule.DatabaseID,
			&rule.QueryID,
			&rule.MetricName,
			&rule.Condition,
			&rule.AlertSeverity,
			&rule.EvaluationInterval,
			&rule.ForDurationSeconds,
			&rule.NotificationEnabled,
			&notificationChannels,
			&rule.IsEnabled,
			&rule.IsPaused,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert rule: %w", err)
		}

		rule.NotificationChannels = []int64(notificationChannels)
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate alert rules: %w", err)
	}

	return rules, nil
}

// UpdateRule updates an existing alert rule
func (r *AlertRulesRepository) UpdateRule(rule *AlertRule) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE alert_rules SET
			name = $2,
			description = $3,
			rule_type = $4,
			database_id = $5,
			query_id = $6,
			metric_name = $7,
			condition = $8,
			alert_severity = $9,
			evaluation_interval_seconds = $10,
			for_duration_seconds = $11,
			notification_enabled = $12,
			notification_channels = $13,
			is_enabled = $14,
			is_paused = $15,
			updated_at = $16
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		rule.ID,
		rule.Name,
		rule.Description,
		rule.RuleType,
		rule.DatabaseID,
		rule.QueryID,
		rule.MetricName,
		rule.Condition,
		rule.AlertSeverity,
		rule.EvaluationInterval,
		rule.ForDurationSeconds,
		rule.NotificationEnabled,
		pq.Array(rule.NotificationChannels),
		rule.IsEnabled,
		rule.IsPaused,
		now,
	)

	if err != nil {
		return fmt.Errorf("update alert rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert rule not found or deleted: %d", rule.ID)
	}

	rule.UpdatedAt = now
	return nil
}

// DeleteRule performs a soft delete on an alert rule
func (r *AlertRulesRepository) DeleteRule(id int64, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE alert_rules
		SET deleted_at = $3, updated_at = $3
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, id, userID, now)
	if err != nil {
		return fmt.Errorf("delete alert rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert rule not found, already deleted, or not owned by user: %d", id)
	}

	return nil
}

// GetAlertHistory retrieves alert trigger history for a rule
func (r *AlertRulesRepository) GetAlertHistory(ruleID int64, limit, offset int) ([]*AlertTrigger, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT at.id, at.alert_id, at.instance_id, at.triggered_at,
			   COALESCE(a.status, 'firing') as status, at.created_at
		FROM alert_triggers at
		LEFT JOIN alerts a ON a.rule_id = at.alert_id
		WHERE at.alert_id IN (SELECT id FROM alerts WHERE rule_id = $1)
		ORDER BY at.triggered_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, ruleID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get alert history: %w", err)
	}
	defer rows.Close()

	triggers := make([]*AlertTrigger, 0)
	for rows.Next() {
		trigger := &AlertTrigger{}
		if err := rows.Scan(
			&trigger.ID,
			&trigger.AlertID,
			&trigger.InstanceID,
			&trigger.TriggeredAt,
			&trigger.Status,
			&trigger.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert trigger: %w", err)
		}
		triggers = append(triggers, trigger)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate alert triggers: %w", err)
	}

	return triggers, nil
}

// AcknowledgeAlert marks an alert trigger as acknowledged
func (r *AlertRulesRepository) AcknowledgeAlert(triggerID int64, userID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update the alert status via the alerts table
	query := `
		UPDATE alerts
		SET status = 'acknowledged', updated_at = $3
		WHERE id = (
			SELECT alert_id FROM alert_triggers WHERE id = $1
		)
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, triggerID, userID, now)
	if err != nil {
		return fmt.Errorf("acknowledge alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert trigger not found: %d", triggerID)
	}

	return nil
}