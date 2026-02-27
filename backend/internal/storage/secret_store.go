package storage

import (
	"context"
	"database/sql"
	"fmt"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// CreateSecret creates a new secret in the database
// Returns the secret ID
func (p *PostgresDB) CreateSecret(ctx context.Context, name, encryptedValue string) (int, error) {
	var id int

	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.secrets (name, secret_encrypted, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id`,
		name, encryptedValue,
	).Scan(&id)

	if err != nil {
		return 0, apperrors.DatabaseError("create secret", err.Error())
	}

	return id, nil
}

// GetSecret retrieves a secret from the database
func (p *PostgresDB) GetSecret(ctx context.Context, id int) (*models.Secret, error) {
	secret := &models.Secret{}
	var encryptedValue []byte

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, name, secret_encrypted, created_at, updated_at
		FROM pganalytics.secrets
		WHERE id = $1`,
		id,
	).Scan(
		&secret.ID,
		&secret.Name,
		&encryptedValue,
		&secret.CreatedAt,
		&secret.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("Secret not found", fmt.Sprintf("ID: %d", id))
		}
		return nil, apperrors.DatabaseError("get secret", err.Error())
	}

	secret.SecretEncrypted = encryptedValue
	return secret, nil
}

// DeleteSecret deletes a secret from the database
func (p *PostgresDB) DeleteSecret(ctx context.Context, id int) error {
	result, err := p.db.ExecContext(
		ctx,
		`DELETE FROM pganalytics.secrets WHERE id = $1`,
		id,
	)

	if err != nil {
		return apperrors.DatabaseError("delete secret", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("delete secret", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("Secret not found", fmt.Sprintf("ID: %d", id))
	}

	return nil
}

// UpdateSecretValue updates the encrypted value of a secret
func (p *PostgresDB) UpdateSecretValue(ctx context.Context, id int, encryptedValue string) error {
	result, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.secrets SET secret_encrypted = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		encryptedValue, id,
	)

	if err != nil {
		return apperrors.DatabaseError("update secret", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("update secret", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("Secret not found", fmt.Sprintf("ID: %d", id))
	}

	return nil
}
