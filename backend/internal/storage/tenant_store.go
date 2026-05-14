package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// TENANT OPERATIONS (SCALE-01, SCALE-02, SCALE-03, SCALE-04)
// ============================================================================

// GetTenantByUserID retrieves the tenant associated with a user
// For single-tenant mode, returns the first active tenant for the user
func (p *PostgresDB) GetTenantByUserID(ctx context.Context, userID uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.created_at, t.updated_at, t.is_active
		FROM tenants t
		JOIN tenant_users tu ON t.id = tu.tenant_id
		WHERE tu.user_id = $1 AND t.is_active = TRUE
		ORDER BY t.created_at
		LIMIT 1
	`

	tenant := &models.Tenant{}
	err := p.db.QueryRowContext(ctx, query, userID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.IsActive,
	)
	if err != nil {
		return nil, apperrors.DatabaseError("query tenant by user id", err.Error())
	}

	return tenant, nil
}

// SetTenantSessionVariable sets the tenant context for RLS policies
// This must be called after establishing a connection to enable tenant isolation
func (p *PostgresDB) SetTenantSessionVariable(ctx context.Context, tenantID uuid.UUID) error {
	query := `SELECT set_tenant_context($1)`

	_, err := p.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return apperrors.DatabaseError("set tenant context", err.Error())
	}

	return nil
}

// GetCollectorsByTenantID retrieves all collectors assigned to a tenant
func (p *PostgresDB) GetCollectorsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.Collector, error) {
	query := `
		SELECT id, name, description, hostname, address, version, last_seen, created_at, tenant_id
		FROM collectors
		WHERE tenant_id = $1
		ORDER BY hostname
	`

	rows, err := p.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, apperrors.DatabaseError("query collectors by tenant", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var collectors []*models.Collector
	for rows.Next() {
		c := &models.Collector{}
		var address, version *string
		var lastSeen *time.Time
		var tenantIDPtr *uuid.UUID

		err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Hostname, &address, &version,
			&lastSeen, &c.CreatedAt, &tenantIDPtr,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan collector", err.Error())
		}

		c.Address = address
		c.Version = version
		c.LastSeen = lastSeen
		collectors = append(collectors, c)
	}

	return collectors, nil
}

// CreateTenant creates a new tenant in the database
func (p *PostgresDB) CreateTenant(ctx context.Context, tenant *models.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}

	query := `
		INSERT INTO tenants (id, name, slug, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := p.db.QueryRowContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.IsActive,
	).Scan(&tenant.CreatedAt, &tenant.UpdatedAt)

	if err != nil {
		return apperrors.DatabaseError("create tenant", err.Error())
	}

	return nil
}

// AddUserToTenant adds a user to a tenant with a specified role
func (p *PostgresDB) AddUserToTenant(ctx context.Context, tenantID, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO tenant_users (tenant_id, user_id, role, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET role = $3
	`

	_, err := p.db.ExecContext(ctx, query, tenantID, userID, role)
	if err != nil {
		return apperrors.DatabaseError("add user to tenant", err.Error())
	}

	return nil
}

// AssignCollectorToTenant assigns a collector to a tenant
func (p *PostgresDB) AssignCollectorToTenant(ctx context.Context, tenantID, collectorID uuid.UUID) error {
	query := `
		UPDATE collectors
		SET tenant_id = $1
		WHERE id = $2
	`

	result, err := p.db.ExecContext(ctx, query, tenantID, collectorID)
	if err != nil {
		return apperrors.DatabaseError("assign collector to tenant", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("get rows affected", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("collector not found", fmt.Sprintf("id: %s", collectorID))
	}

	return nil
}

// GetTenantByID retrieves a tenant by its ID
func (p *PostgresDB) GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*models.Tenant, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at, is_active
		FROM tenants
		WHERE id = $1
	`

	tenant := &models.Tenant{}
	err := p.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.IsActive,
	)
	if err != nil {
		return nil, apperrors.DatabaseError("query tenant by id", err.Error())
	}

	return tenant, nil
}

// GetTenantsByUserID retrieves all tenants a user belongs to
func (p *PostgresDB) GetTenantsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Tenant, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.created_at, t.updated_at, t.is_active
		FROM tenants t
		JOIN tenant_users tu ON t.id = tu.tenant_id
		WHERE tu.user_id = $1 AND t.is_active = TRUE
		ORDER BY t.name
	`

	rows, err := p.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, apperrors.DatabaseError("query tenants by user", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var tenants []*models.Tenant
	for rows.Next() {
		t := &models.Tenant{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.Slug,
			&t.CreatedAt, &t.UpdatedAt, &t.IsActive,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan tenant", err.Error())
		}
		tenants = append(tenants, t)
	}

	return tenants, nil
}

// RemoveUserFromTenant removes a user from a tenant
func (p *PostgresDB) RemoveUserFromTenant(ctx context.Context, tenantID, userID uuid.UUID) error {
	query := `
		DELETE FROM tenant_users
		WHERE tenant_id = $1 AND user_id = $2
	`

	_, err := p.db.ExecContext(ctx, query, tenantID, userID)
	if err != nil {
		return apperrors.DatabaseError("remove user from tenant", err.Error())
	}

	return nil
}

// GetUserRoleInTenant gets the role of a user in a specific tenant
func (p *PostgresDB) GetUserRoleInTenant(ctx context.Context, tenantID, userID uuid.UUID) (string, error) {
	query := `
		SELECT role FROM tenant_users
		WHERE tenant_id = $1 AND user_id = $2
	`

	var role string
	err := p.db.QueryRowContext(ctx, query, tenantID, userID).Scan(&role)
	if err != nil {
		return "", apperrors.DatabaseError("query user role", err.Error())
	}

	return role, nil
}
