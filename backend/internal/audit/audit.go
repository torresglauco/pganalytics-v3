package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// AuditAction represents an audited action
type AuditAction string

const (
	ActionUserCreate        AuditAction = "user_create"
	ActionUserUpdate        AuditAction = "user_update"
	ActionUserDelete        AuditAction = "user_delete"
	ActionUserLogin         AuditAction = "user_login"
	ActionUserLogout        AuditAction = "user_logout"
	ActionPasswordChange    AuditAction = "password_change"
	ActionTokenRefresh      AuditAction = "token_refresh"
	ActionCollectorRegister AuditAction = "collector_register"
	ActionCollectorDelete   AuditAction = "collector_delete"
	ActionConfigChange      AuditAction = "config_change"
	ActionAlertRuleCreate   AuditAction = "alert_rule_create"
	ActionAlertRuleUpdate   AuditAction = "alert_rule_update"
	ActionAlertRuleDelete   AuditAction = "alert_rule_delete"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID             int64           `json:"id" db:"id"`
	UserID         *int            `json:"user_id" db:"user_id"`
	Action         AuditAction     `json:"action" db:"action"`
	ResourceType   string          `json:"resource_type" db:"resource_type"`
	ResourceID     *string         `json:"resource_id" db:"resource_id"`
	ChangesBefore  json.RawMessage `json:"changes_before,omitempty" db:"changes_before"`
	ChangesAfter   json.RawMessage `json:"changes_after,omitempty" db:"changes_after"`
	IPAddress      *net.IP         `json:"ip_address" db:"ip_address"`
	UserAgent      *string         `json:"user_agent" db:"user_agent"`
	AdditionalData json.RawMessage `json:"additional_data,omitempty" db:"additional_data"`
	Timestamp      time.Time       `json:"timestamp" db:"timestamp"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}

// AuditLogger handles audit logging
type AuditLogger struct {
	db *sql.DB
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{
		db: db,
	}
}

// LogAction logs an action
func (al *AuditLogger) LogAction(ctx context.Context, log *AuditLog) (int64, error) {
	query := `
		INSERT INTO audit_logs (
			user_id, action, resource_type, resource_id,
			changes_before, changes_after, ip_address, user_agent,
			additional_data, timestamp
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var id int64
	err := al.db.QueryRowContext(
		ctx,
		query,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.ChangesBefore,
		log.ChangesAfter,
		log.IPAddress,
		log.UserAgent,
		log.AdditionalData,
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to log action: %w", err)
	}

	return id, nil
}

// AuditFilter represents filter criteria for audit logs
type AuditFilter struct {
	UserID       *int
	Action       *string
	ResourceType *string
	ResourceID   *string
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	Offset       int
}

// GetHistory retrieves audit log history
func (al *AuditLogger) GetHistory(ctx context.Context, filter *AuditFilter) ([]AuditLog, error) {
	query := `
		SELECT
			id, user_id, action, resource_type, resource_id,
			changes_before, changes_after, ip_address, user_agent,
			additional_data, timestamp, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := make([]interface{}, 0)
	argCount := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, *filter.Action)
		argCount++
	}

	if filter.ResourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", argCount)
		args = append(args, *filter.ResourceType)
		argCount++
	}

	if filter.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argCount)
		args = append(args, *filter.ResourceID)
		argCount++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argCount)
		args = append(args, *filter.DateFrom)
		argCount++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argCount)
		args = append(args, *filter.DateTo)
		argCount++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := al.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
			&log.ChangesBefore, &log.ChangesAfter, &log.IPAddress, &log.UserAgent,
			&log.AdditionalData, &log.Timestamp, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading audit logs: %w", err)
	}

	return logs, nil
}

// GetByID retrieves an audit log by ID
func (al *AuditLogger) GetByID(ctx context.Context, id int64) (*AuditLog, error) {
	query := `
		SELECT
			id, user_id, action, resource_type, resource_id,
			changes_before, changes_after, ip_address, user_agent,
			additional_data, timestamp, created_at
		FROM audit_logs
		WHERE id = $1
	`

	var log AuditLog
	err := al.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID,
		&log.ChangesBefore, &log.ChangesAfter, &log.IPAddress, &log.UserAgent,
		&log.AdditionalData, &log.Timestamp, &log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query audit log: %w", err)
	}

	return &log, nil
}

// GetStats retrieves audit log statistics
func (al *AuditLogger) GetStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(DISTINCT action) as action_types,
			MAX(timestamp) as last_action
		FROM audit_logs
	`

	var total, uniqueUsers, actionTypes int
	var lastAction *time.Time

	err := al.db.QueryRowContext(ctx, query).Scan(&total, &uniqueUsers, &actionTypes, &lastAction)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit stats: %w", err)
	}

	return map[string]interface{}{
		"total":        total,
		"unique_users": uniqueUsers,
		"action_types": actionTypes,
		"last_action":  lastAction,
	}, nil
}

// ExportFormat represents export format
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// Export exports audit logs
func (al *AuditLogger) Export(ctx context.Context, filter *AuditFilter, format ExportFormat) ([]byte, error) {
	logs, err := al.GetHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	switch format {
	case ExportFormatJSON:
		return json.MarshalIndent(logs, "", "  ")

	case ExportFormatCSV:
		// Simple CSV export
		csv := "ID,User ID,Action,Resource Type,Resource ID,Timestamp,IP Address,User Agent\n"
		for _, log := range logs {
			userID := ""
			if log.UserID != nil {
				userID = fmt.Sprintf("%d", *log.UserID)
			}
			resourceID := ""
			if log.ResourceID != nil {
				resourceID = *log.ResourceID
			}
			ipAddr := ""
			if log.IPAddress != nil {
				ipAddr = log.IPAddress.String()
			}
			userAgent := ""
			if log.UserAgent != nil {
				userAgent = *log.UserAgent
			}

			csv += fmt.Sprintf(
				"%d,%s,%s,%s,%s,%s,%s,%s\n",
				log.ID, userID, log.Action, log.ResourceType, resourceID,
				log.Timestamp.Format(time.RFC3339), ipAddr, userAgent,
			)
		}
		return []byte(csv), nil

	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// Cleanup removes old audit logs (archival)
func (al *AuditLogger) Cleanup(ctx context.Context, retentionDays int) (int64, error) {
	query := `
		DELETE FROM audit_logs
		WHERE created_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := al.db.ExecContext(ctx, query, retentionDays)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup audit logs: %w", err)
	}

	return result.RowsAffected()
}

// GetActionCounts gets count of actions
func (al *AuditLogger) GetActionCounts(ctx context.Context, since time.Time) (map[string]int, error) {
	query := `
		SELECT action, COUNT(*) as count
		FROM audit_logs
		WHERE timestamp > $1
		GROUP BY action
		ORDER BY count DESC
	`

	rows, err := al.db.QueryContext(ctx, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query action counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			return nil, err
		}
		counts[action] = count
	}

	return counts, rows.Err()
}

// GetUserActions gets all actions by a user
func (al *AuditLogger) GetUserActions(ctx context.Context, userID int, limit int) ([]AuditLog, error) {
	filter := &AuditFilter{
		UserID: &userID,
		Limit:  limit,
	}
	return al.GetHistory(ctx, filter)
}

// GetResourceHistory gets all actions on a specific resource
func (al *AuditLogger) GetResourceHistory(ctx context.Context, resourceType, resourceID string) ([]AuditLog, error) {
	filter := &AuditFilter{
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
	}
	return al.GetHistory(ctx, filter)
}
