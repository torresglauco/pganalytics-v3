package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// CreateRDSInstance creates a new RDS instance record
func (p *PostgresDB) CreateRDSInstance(ctx context.Context, instance *models.CreateRDSInstanceRequest, secretID *int, userID int) (*models.RDSInstance, error) {
	var id int
	var createdAt, updatedAt sql.NullTime

	tagsJSON, _ := json.Marshal(instance.Tags)

	// Convert nil pointer to interface{} for proper NULL handling
	var secretIDValue interface{} = secretID
	if secretID == nil {
		secretIDValue = nil
	}

	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.rds_instances (
			name, description, aws_region, rds_endpoint, port,
			engine_version, db_instance_class, allocated_storage_gb,
			environment, master_username, secret_id,
			enable_enhanced_monitoring, monitoring_interval,
			ssl_enabled, ssl_mode, connection_timeout,
			multi_az, backup_retention_days, preferred_backup_window,
			preferred_maintenance_window, tags, is_active, status,
			created_by, updated_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13,
			$14, $15, $16,
			$17, $18, $19,
			$20, $21, true, 'registered',
			$22, $22, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id, created_at, updated_at`,
		instance.Name, instance.Description, instance.AWSRegion,
		instance.RDSEndpoint, instance.Port,
		instance.EngineVersion, instance.DBInstanceClass, instance.AllocatedStorageGB,
		instance.Environment, instance.MasterUsername, secretIDValue,
		instance.EnableEnhancedMonitoring, instance.MonitoringInterval,
		instance.SSLEnabled, instance.SSLMode, instance.ConnectionTimeout,
		instance.MultiAZ, instance.BackupRetentionDays,
		instance.PreferredBackupWindow, instance.PreferredMaintenanceWindow,
		tagsJSON, userID,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, apperrors.DatabaseError("create RDS instance", err.Error())
	}

	result := &models.RDSInstance{
		ID:                      id,
		Name:                    instance.Name,
		Description:             ptrString(instance.Description),
		AWSRegion:               instance.AWSRegion,
		RDSEndpoint:             instance.RDSEndpoint,
		Port:                    instance.Port,
		EngineVersion:           ptrString(instance.EngineVersion),
		DBInstanceClass:         ptrString(instance.DBInstanceClass),
		AllocatedStorageGB:      ptrInt(instance.AllocatedStorageGB),
		Environment:             instance.Environment,
		MasterUsername:          instance.MasterUsername,
		SecretID:                secretID,
		EnableEnhancedMonitoring: instance.EnableEnhancedMonitoring,
		MonitoringInterval:      instance.MonitoringInterval,
		SSLEnabled:              instance.SSLEnabled,
		SSLMode:                 instance.SSLMode,
		ConnectionTimeout:       instance.ConnectionTimeout,
		IsActive:                true,
		Status:                  "registered",
		MultiAZ:                 instance.MultiAZ,
		BackupRetentionDays:     ptrInt(instance.BackupRetentionDays),
		PreferredBackupWindow:   ptrString(instance.PreferredBackupWindow),
		PreferredMaintenanceWindow: ptrString(instance.PreferredMaintenanceWindow),
		Tags:                    instance.Tags,
		CreatedAt:               createdAt.Time,
		UpdatedAt:               updatedAt.Time,
		CreatedBy:               &userID,
		UpdatedBy:               &userID,
	}

	return result, nil
}

// GetRDSInstance retrieves an RDS instance by ID
func (p *PostgresDB) GetRDSInstance(ctx context.Context, id int) (*models.RDSInstance, error) {
	instance := &models.RDSInstance{}
	var tags json.RawMessage

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, name, description, aws_region, rds_endpoint, port,
			engine_version, db_instance_class, allocated_storage_gb,
			environment, master_username, secret_id,
			enable_enhanced_monitoring, monitoring_interval,
			ssl_enabled, ssl_mode, connection_timeout,
			is_active, status, last_heartbeat, last_connection_status,
			last_error_message, last_error_time,
			multi_az, backup_retention_days, preferred_backup_window,
			preferred_maintenance_window, tags,
			created_at, updated_at, created_by, updated_by
		FROM pganalytics.rds_instances WHERE id = $1`,
		id,
	).Scan(
		&instance.ID, &instance.Name, &instance.Description, &instance.AWSRegion,
		&instance.RDSEndpoint, &instance.Port, &instance.EngineVersion,
		&instance.DBInstanceClass, &instance.AllocatedStorageGB,
		&instance.Environment, &instance.MasterUsername, &instance.SecretID,
		&instance.EnableEnhancedMonitoring, &instance.MonitoringInterval,
		&instance.SSLEnabled, &instance.SSLMode, &instance.ConnectionTimeout,
		&instance.IsActive, &instance.Status, &instance.LastHeartbeat, &instance.LastConnectionStatus,
		&instance.LastErrorMessage, &instance.LastErrorTime,
		&instance.MultiAZ, &instance.BackupRetentionDays,
		&instance.PreferredBackupWindow, &instance.PreferredMaintenanceWindow,
		&tags,
		&instance.CreatedAt, &instance.UpdatedAt, &instance.CreatedBy, &instance.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("RDS instance not found", fmt.Sprintf("ID: %d", id))
		}
		return nil, apperrors.DatabaseError("get RDS instance", err.Error())
	}

	// Parse tags
	if len(tags) > 0 {
		_ = json.Unmarshal(tags, &instance.Tags)
	}

	return instance, nil
}

// ListRDSInstances retrieves all active RDS instances
func (p *PostgresDB) ListRDSInstances(ctx context.Context) ([]*models.RDSInstance, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT id, name, description, aws_region, rds_endpoint, port,
			engine_version, db_instance_class, allocated_storage_gb,
			environment, master_username, secret_id,
			enable_enhanced_monitoring, monitoring_interval,
			ssl_enabled, ssl_mode, connection_timeout,
			is_active, status, last_heartbeat, last_connection_status,
			last_error_message, last_error_time,
			multi_az, backup_retention_days, preferred_backup_window,
			preferred_maintenance_window, tags,
			created_at, updated_at, created_by, updated_by
		FROM pganalytics.rds_instances WHERE is_active = true
		ORDER BY name ASC`,
	)

	if err != nil {
		return nil, apperrors.DatabaseError("list RDS instances", err.Error())
	}
	defer rows.Close()

	var instances []*models.RDSInstance
	for rows.Next() {
		instance := &models.RDSInstance{}
		var tags json.RawMessage

		err := rows.Scan(
			&instance.ID, &instance.Name, &instance.Description, &instance.AWSRegion,
			&instance.RDSEndpoint, &instance.Port, &instance.EngineVersion,
			&instance.DBInstanceClass, &instance.AllocatedStorageGB,
			&instance.Environment, &instance.MasterUsername, &instance.SecretID,
			&instance.EnableEnhancedMonitoring, &instance.MonitoringInterval,
			&instance.SSLEnabled, &instance.SSLMode, &instance.ConnectionTimeout,
			&instance.IsActive, &instance.Status, &instance.LastHeartbeat, &instance.LastConnectionStatus,
			&instance.LastErrorMessage, &instance.LastErrorTime,
			&instance.MultiAZ, &instance.BackupRetentionDays,
			&instance.PreferredBackupWindow, &instance.PreferredMaintenanceWindow,
			&tags,
			&instance.CreatedAt, &instance.UpdatedAt, &instance.CreatedBy, &instance.UpdatedBy,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan RDS instance", err.Error())
		}

		if len(tags) > 0 {
			_ = json.Unmarshal(tags, &instance.Tags)
		}

		instances = append(instances, instance)
	}

	return instances, rows.Err()
}

// UpdateRDSInstanceStatus updates the connection status and last heartbeat
func (p *PostgresDB) UpdateRDSInstanceStatus(ctx context.Context, id int, status string, errorMsg *string) error {
	result, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.rds_instances
		SET last_connection_status = $1, last_heartbeat = CURRENT_TIMESTAMP,
		    last_error_message = $2, last_error_time = CASE WHEN $2 IS NOT NULL THEN CURRENT_TIMESTAMP ELSE last_error_time END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $3`,
		status, errorMsg, id,
	)

	if err != nil {
		return apperrors.DatabaseError("update RDS instance status", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("update RDS instance status", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("RDS instance not found", fmt.Sprintf("ID: %d", id))
	}

	return nil
}

// UpdateRDSInstance updates an RDS instance
func (p *PostgresDB) UpdateRDSInstance(ctx context.Context, id int, instance *models.UpdateRDSInstanceRequest, userID int) (*models.RDSInstance, error) {
	tagsJSON, _ := json.Marshal(instance.Tags)

	// Convert nil pointer to interface{} for proper NULL handling
	var secretIDValue interface{} = nil

	updatedInstance := &models.RDSInstance{}
	var tags json.RawMessage

	err := p.db.QueryRowContext(
		ctx,
		`UPDATE pganalytics.rds_instances SET
			name = $1, description = $2, aws_region = $3, rds_endpoint = $4, port = $5,
			engine_version = $6, db_instance_class = $7, allocated_storage_gb = $8,
			environment = $9, master_username = $10, secret_id = $11,
			enable_enhanced_monitoring = $12, monitoring_interval = $13,
			ssl_enabled = $14, ssl_mode = $15, connection_timeout = $16,
			multi_az = $17, backup_retention_days = $18, preferred_backup_window = $19,
			preferred_maintenance_window = $20, tags = $21, status = $22,
			updated_by = $23, updated_at = CURRENT_TIMESTAMP
		WHERE id = $24
		RETURNING id, name, description, aws_region, rds_endpoint, port,
			engine_version, db_instance_class, allocated_storage_gb,
			environment, master_username, secret_id,
			enable_enhanced_monitoring, monitoring_interval,
			ssl_enabled, ssl_mode, connection_timeout,
			is_active, status, last_heartbeat, last_connection_status,
			last_error_message, last_error_time,
			multi_az, backup_retention_days, preferred_backup_window,
			preferred_maintenance_window, tags,
			created_at, updated_at, created_by, updated_by`,
		instance.Name, instance.Description, instance.AWSRegion, instance.RDSEndpoint, instance.Port,
		instance.EngineVersion, instance.DBInstanceClass, instance.AllocatedStorageGB,
		instance.Environment, instance.MasterUsername, secretIDValue,
		instance.EnableEnhancedMonitoring, instance.MonitoringInterval,
		instance.SSLEnabled, instance.SSLMode, instance.ConnectionTimeout,
		instance.MultiAZ, instance.BackupRetentionDays, instance.PreferredBackupWindow,
		instance.PreferredMaintenanceWindow, tagsJSON, instance.Status,
		userID, id,
	).Scan(
		&updatedInstance.ID, &updatedInstance.Name, &updatedInstance.Description, &updatedInstance.AWSRegion,
		&updatedInstance.RDSEndpoint, &updatedInstance.Port, &updatedInstance.EngineVersion,
		&updatedInstance.DBInstanceClass, &updatedInstance.AllocatedStorageGB,
		&updatedInstance.Environment, &updatedInstance.MasterUsername, &updatedInstance.SecretID,
		&updatedInstance.EnableEnhancedMonitoring, &updatedInstance.MonitoringInterval,
		&updatedInstance.SSLEnabled, &updatedInstance.SSLMode, &updatedInstance.ConnectionTimeout,
		&updatedInstance.IsActive, &updatedInstance.Status, &updatedInstance.LastHeartbeat, &updatedInstance.LastConnectionStatus,
		&updatedInstance.LastErrorMessage, &updatedInstance.LastErrorTime,
		&updatedInstance.MultiAZ, &updatedInstance.BackupRetentionDays,
		&updatedInstance.PreferredBackupWindow, &updatedInstance.PreferredMaintenanceWindow,
		&tags,
		&updatedInstance.CreatedAt, &updatedInstance.UpdatedAt, &updatedInstance.CreatedBy, &updatedInstance.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("RDS instance not found", fmt.Sprintf("ID: %d", id))
		}
		return nil, apperrors.DatabaseError("update RDS instance", err.Error())
	}

	if len(tags) > 0 {
		_ = json.Unmarshal(tags, &updatedInstance.Tags)
	}

	return updatedInstance, nil
}

// DeleteRDSInstance deletes an RDS instance (soft delete)
func (p *PostgresDB) DeleteRDSInstance(ctx context.Context, id int) error {
	result, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.rds_instances SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		id,
	)

	if err != nil {
		return apperrors.DatabaseError("delete RDS instance", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("delete RDS instance", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("RDS instance not found", fmt.Sprintf("ID: %d", id))
	}

	return nil
}

// Helper functions for pointer conversion
func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrInt(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
