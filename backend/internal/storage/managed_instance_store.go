package storage

import (
	"context"
	"database/sql"
	"fmt"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// HealthCheckInstance represents a managed instance for health checks
type HealthCheckInstance struct {
	ID       int
	Name     string
	Endpoint string
	Port     int
	SSLMode  string
	Status   string
}

// CreateManagedInstance creates a new RDS instance record
func (p *PostgresDB) CreateManagedInstance(ctx context.Context, instance *models.CreateManagedInstanceRequest, secretID *int, userID int) (*models.ManagedInstance, error) {
	var id int
	var createdAt, updatedAt sql.NullTime

	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.managed_instances (
			name, aws_region, rds_endpoint, port,
			environment, ssl_enabled, ssl_mode,
			multi_az, backup_retention_days,
			is_active, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			true, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) RETURNING id, created_at, updated_at`,
		instance.Name, instance.AWSRegion,
		instance.Endpoint, instance.Port,
		instance.Environment, instance.SSLEnabled, instance.SSLMode,
		instance.MultiAZ, instance.BackupRetentionDays,
		userID,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, apperrors.DatabaseError("create managed instance", err.Error())
	}

	result := &models.ManagedInstance{
		ID:                  id,
		Name:                instance.Name,
		AWSRegion:           instance.AWSRegion,
		Endpoint:            instance.Endpoint,
		Port:                instance.Port,
		Environment:         instance.Environment,
		SSLEnabled:          instance.SSLEnabled,
		SSLMode:             instance.SSLMode,
		IsActive:            true,
		MultiAZ:             instance.MultiAZ,
		BackupRetentionDays: ptrInt(instance.BackupRetentionDays),
		CreatedAt:           createdAt.Time,
		UpdatedAt:           updatedAt.Time,
		CreatedBy:           &userID,
	}

	return result, nil
}

// GetManagedInstance retrieves an RDS instance by ID
func (p *PostgresDB) GetManagedInstance(ctx context.Context, id int) (*models.ManagedInstance, error) {
	instance := &models.ManagedInstance{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, name, aws_region, rds_endpoint, port, engine_version, db_instance_class,
			ssl_enabled, ssl_mode, is_active, last_connection_status,
			last_heartbeat, last_error_message, environment, multi_az,
			backup_retention_days, created_by, created_at, updated_at
		FROM pganalytics.managed_instances WHERE id = $1`,
		id,
	).Scan(
		&instance.ID, &instance.Name, &instance.AWSRegion, &instance.Endpoint, &instance.Port,
		&instance.EngineVersion, &instance.DBInstanceClass,
		&instance.SSLEnabled, &instance.SSLMode, &instance.IsActive, &instance.LastConnectionStatus,
		&instance.LastHeartbeat, &instance.LastErrorMessage, &instance.Environment, &instance.MultiAZ,
		&instance.BackupRetentionDays, &instance.CreatedBy, &instance.CreatedAt, &instance.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("RDS instance not found", fmt.Sprintf("ID: %d", id))
		}
		return nil, apperrors.DatabaseError("get managed instance", err.Error())
	}

	return instance, nil
}

// ListManagedInstances retrieves all active RDS instances
func (p *PostgresDB) ListManagedInstances(ctx context.Context) ([]*models.ManagedInstance, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT id, name, aws_region, rds_endpoint, port, engine_version, db_instance_class,
			ssl_enabled, ssl_mode, is_active, last_connection_status,
			last_heartbeat, last_error_message, environment, multi_az,
			backup_retention_days, created_by, created_at, updated_at
		FROM pganalytics.managed_instances WHERE is_active = true
		ORDER BY name ASC`,
	)

	if err != nil {
		return nil, apperrors.DatabaseError("list managed instances", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var instances []*models.ManagedInstance
	for rows.Next() {
		instance := &models.ManagedInstance{}

		err := rows.Scan(
			&instance.ID, &instance.Name, &instance.AWSRegion, &instance.Endpoint, &instance.Port,
			&instance.EngineVersion, &instance.DBInstanceClass,
			&instance.SSLEnabled, &instance.SSLMode, &instance.IsActive, &instance.LastConnectionStatus,
			&instance.LastHeartbeat, &instance.LastErrorMessage, &instance.Environment, &instance.MultiAZ,
			&instance.BackupRetentionDays, &instance.CreatedBy, &instance.CreatedAt, &instance.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan RDS instance", err.Error())
		}

		instances = append(instances, instance)
	}

	return instances, rows.Err()
}

// UpdateManagedInstanceStatus updates the connection status and last heartbeat
func (p *PostgresDB) UpdateManagedInstanceStatus(ctx context.Context, id int, status string, errorMsg *string) error {
	result, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.managed_instances
		SET last_connection_status = $1, last_heartbeat = CURRENT_TIMESTAMP,
		    last_error_message = $2::text, last_error_time = CASE WHEN $2 IS NOT NULL THEN CURRENT_TIMESTAMP ELSE last_error_time END,
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

// UpdateManagedInstance updates an RDS instance
func (p *PostgresDB) UpdateManagedInstance(ctx context.Context, id int, instance *models.UpdateManagedInstanceRequest, userID int) (*models.ManagedInstance, error) {
	updatedInstance := &models.ManagedInstance{}

	err := p.db.QueryRowContext(
		ctx,
		`UPDATE pganalytics.managed_instances SET
			name = $1, aws_region = $2, rds_endpoint = $3, port = $4,
			environment = $5, ssl_enabled = $6, ssl_mode = $7,
			multi_az = $8, backup_retention_days = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
		RETURNING id, name, aws_region, rds_endpoint, port,
			environment, ssl_enabled, ssl_mode,
			multi_az, backup_retention_days,
			is_active, last_heartbeat, last_connection_status,
			last_error_message, engine_version, db_instance_class,
			created_at, updated_at, created_by`,
		instance.Name, instance.AWSRegion, instance.Endpoint, instance.Port,
		instance.Environment, instance.SSLEnabled, instance.SSLMode,
		instance.MultiAZ, instance.BackupRetentionDays,
		id,
	).Scan(
		&updatedInstance.ID, &updatedInstance.Name, &updatedInstance.AWSRegion, &updatedInstance.Endpoint, &updatedInstance.Port,
		&updatedInstance.Environment, &updatedInstance.SSLEnabled, &updatedInstance.SSLMode,
		&updatedInstance.MultiAZ, &updatedInstance.BackupRetentionDays,
		&updatedInstance.IsActive, &updatedInstance.LastHeartbeat, &updatedInstance.LastConnectionStatus,
		&updatedInstance.LastErrorMessage, &updatedInstance.EngineVersion, &updatedInstance.DBInstanceClass,
		&updatedInstance.CreatedAt, &updatedInstance.UpdatedAt, &updatedInstance.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("RDS instance not found", fmt.Sprintf("ID: %d", id))
		}
		return nil, apperrors.DatabaseError("update managed instance", err.Error())
	}

	return updatedInstance, nil
}

// DeleteManagedInstance deletes an RDS instance (soft delete)
func (p *PostgresDB) DeleteManagedInstance(ctx context.Context, id int) error {
	result, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.managed_instances SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`,
		id,
	)

	if err != nil {
		return apperrors.DatabaseError("delete managed instance", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("delete managed instance", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("RDS instance not found", fmt.Sprintf("ID: %d", id))
	}

	return nil
}

// ListManagedInstancesForHealthCheck retrieves all managed instances for background health checks
func (p *PostgresDB) ListManagedInstancesForHealthCheck(ctx context.Context) ([]*HealthCheckInstance, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT m.id, m.name, m.rds_endpoint, m.port,
		        m.ssl_mode, m.last_connection_status
		 FROM pganalytics.managed_instances m
		 WHERE m.is_active = true
		 ORDER BY m.id ASC`,
	)
	if err != nil {
		return nil, apperrors.DatabaseError("list managed instances", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var instances []*HealthCheckInstance
	for rows.Next() {
		instance := &HealthCheckInstance{}
		err := rows.Scan(
			&instance.ID, &instance.Name, &instance.Endpoint, &instance.Port,
			&instance.SSLMode, &instance.Status,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan managed instance", err.Error())
		}
		instances = append(instances, instance)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("iterate managed instances", err.Error())
	}

	return instances, nil
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
