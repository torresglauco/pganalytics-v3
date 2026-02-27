package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// CreateRegistrationSecret creates a new registration secret
func (s *PostgresStore) CreateRegistrationSecret(
	ctx context.Context,
	name string,
	description string,
	expiresAt *time.Time,
	createdBy int,
) (*models.CreateRegistrationSecretResponse, error) {
	// Generate a cryptographically secure random secret
	secretValue, err := generateSecureSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	id := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO registration_secrets (id, name, secret_value, description, active, created_by, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, description, active, created_at
	`

	var createdSecret models.RegistrationSecret
	err = s.db.QueryRowContext(ctx, query,
		id, name, secretValue, description, true, createdBy, now, now, expiresAt,
	).Scan(
		&createdSecret.ID, &createdSecret.Name, &createdSecret.Description,
		&createdSecret.Active, &createdSecret.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create registration secret: %w", err)
	}

	return &models.CreateRegistrationSecretResponse{
		ID:          createdSecret.ID,
		Name:        createdSecret.Name,
		SecretValue: secretValue, // Only returned once
		Description: createdSecret.Description,
		Active:      createdSecret.Active,
		CreatedAt:   createdSecret.CreatedAt,
		Message:     "Registration secret created successfully. Save the secret value securely as it won't be shown again.",
	}, nil
}

// GetRegistrationSecret retrieves a registration secret by ID
func (s *PostgresStore) GetRegistrationSecret(ctx context.Context, id string) (*models.RegistrationSecret, error) {
	query := `
		SELECT rs.id, rs.name, rs.description, rs.active, rs.created_by, u.username,
		       rs.created_at, rs.updated_at, rs.expires_at, rs.total_registrations, rs.last_used_at
		FROM registration_secrets rs
		LEFT JOIN users u ON rs.created_by = u.id
		WHERE rs.id = $1
	`

	var secret models.RegistrationSecret
	var username *string
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&secret.ID, &secret.Name, &secret.Description, &secret.Active, &secret.CreatedBy,
		&username, &secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt,
		&secret.TotalRegistrations, &secret.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get registration secret: %w", err)
	}

	if username != nil {
		secret.CreatedByUsername = *username
	}

	return &secret, nil
}

// ListRegistrationSecrets retrieves all registration secrets
func (s *PostgresStore) ListRegistrationSecrets(ctx context.Context) ([]models.RegistrationSecret, error) {
	query := `
		SELECT rs.id, rs.name, rs.description, rs.active, rs.created_by, u.username,
		       rs.created_at, rs.updated_at, rs.expires_at, rs.total_registrations, rs.last_used_at
		FROM registration_secrets rs
		LEFT JOIN users u ON rs.created_by = u.id
		ORDER BY rs.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list registration secrets: %w", err)
	}
	defer rows.Close()

	var secrets []models.RegistrationSecret
	for rows.Next() {
		var secret models.RegistrationSecret
		var username *string
		err := rows.Scan(
			&secret.ID, &secret.Name, &secret.Description, &secret.Active, &secret.CreatedBy,
			&username, &secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt,
			&secret.TotalRegistrations, &secret.LastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration secret: %w", err)
		}

		if username != nil {
			secret.CreatedByUsername = *username
		}

		secrets = append(secrets, secret)
	}

	return secrets, rows.Err()
}

// UpdateRegistrationSecret updates a registration secret
func (s *PostgresStore) UpdateRegistrationSecret(
	ctx context.Context,
	id string,
	name *string,
	description *string,
	active *bool,
) (*models.RegistrationSecret, error) {
	query := `
		UPDATE registration_secrets
		SET name = COALESCE($2, name),
		    description = COALESCE($3, description),
		    active = COALESCE($4, active),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, active, created_by, created_at, updated_at, expires_at, total_registrations, last_used_at
	`

	var secret models.RegistrationSecret
	err := s.db.QueryRowContext(ctx, query, id, name, description, active).Scan(
		&secret.ID, &secret.Name, &secret.Description, &secret.Active, &secret.CreatedBy,
		&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.TotalRegistrations,
		&secret.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update registration secret: %w", err)
	}

	return &secret, nil
}

// DeleteRegistrationSecret deletes a registration secret
func (s *PostgresStore) DeleteRegistrationSecret(ctx context.Context, id string) error {
	query := `DELETE FROM registration_secrets WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete registration secret: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("registration secret not found")
	}

	return nil
}

// ValidateRegistrationSecret validates a registration secret
func (s *PostgresStore) ValidateRegistrationSecret(ctx context.Context, secretValue string) (*models.RegistrationSecret, error) {
	query := `
		SELECT id, name, description, active, created_by, created_at, updated_at, expires_at, total_registrations, last_used_at
		FROM registration_secrets
		WHERE secret_value = $1 AND active = true
		AND (expires_at IS NULL OR expires_at > NOW())
	`

	var secret models.RegistrationSecret
	err := s.db.QueryRowContext(ctx, query, secretValue).Scan(
		&secret.ID, &secret.Name, &secret.Description, &secret.Active, &secret.CreatedBy,
		&secret.CreatedAt, &secret.UpdatedAt, &secret.ExpiresAt, &secret.TotalRegistrations,
		&secret.LastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired registration secret")
	}

	return &secret, nil
}

// RecordRegistrationSecretUsage records the usage of a registration secret
func (s *PostgresStore) RecordRegistrationSecretUsage(
	ctx context.Context,
	secretID string,
	collectorID string,
	collectorName string,
	status string,
	errorMessage *string,
	ipAddress string,
) error {
	// Update the secret's usage stats
	updateQuery := `
		UPDATE registration_secrets
		SET total_registrations = total_registrations + 1,
		    last_used_at = NOW()
		WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, updateQuery, secretID)
	if err != nil {
		return fmt.Errorf("failed to update secret usage: %w", err)
	}

	// Record in audit table
	auditQuery := `
		INSERT INTO registration_secret_audit (secret_id, collector_id, collector_name, status, error_message, ip_address, used_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err = s.db.ExecContext(ctx, auditQuery, secretID, collectorID, collectorName, status, errorMessage, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to record audit: %w", err)
	}

	return nil
}

// generateSecureSecret generates a cryptographically secure random secret
func generateSecureSecret() (string, error) {
	// Generate 32 bytes of random data
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode to base64 for a URL-safe string
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
