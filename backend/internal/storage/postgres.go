package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dextra/pganalytics-v3/backend/pkg/models"
	apperrors "github.com/dextra/pganalytics-v3/backend/pkg/errors"
	_ "github.com/lib/pq"
)

// PostgresDB wraps a PostgreSQL database connection
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(connString string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, apperrors.DatabaseError("open connection", err.Error())
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, apperrors.DatabaseError("ping database", err.Error())
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// Health checks the database health
func (p *PostgresDB) Health(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return p.db.PingContext(ctx) == nil
}

// ============================================================================
// USER OPERATIONS
// ============================================================================

// GetUserByUsername retrieves a user by username
func (p *PostgresDB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, full_name, role, is_active, last_login, created_at, updated_at
		 FROM pganalytics.users WHERE username = $1`,
		username,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.Role, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.UserNotFound(username)
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get user", err.Error())
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (p *PostgresDB) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	user := &models.User{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, full_name, role, is_active, last_login, created_at, updated_at
		 FROM pganalytics.users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.Role, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("User not found", fmt.Sprintf("User ID %d not found", userID))
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get user by id", err.Error())
	}

	return user, nil
}

// UpdateUserLastLogin updates the last login timestamp for a user
func (p *PostgresDB) UpdateUserLastLogin(ctx context.Context, userID int) error {
	_, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.users SET last_login = CURRENT_TIMESTAMP WHERE id = $1`,
		userID,
	)

	if err != nil {
		return apperrors.DatabaseError("update user last_login", err.Error())
	}

	return nil
}

// ============================================================================
// COLLECTOR OPERATIONS
// ============================================================================

// CreateCollector creates a new collector
func (p *PostgresDB) CreateCollector(ctx context.Context, collector *models.Collector) error {
	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.collectors
		 (name, description, hostname, address, version, status, certificate_thumbprint, certificate_expires_at, config_version, health_check_interval)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, created_at, updated_at`,
		collector.Name, collector.Description, collector.Hostname, collector.Address,
		collector.Version, collector.Status, collector.CertificateThumbprint,
		collector.CertificateExpiresAt, collector.ConfigVersion, collector.HealthCheckInterval,
	).Scan(&collector.ID, &collector.CreatedAt, &collector.UpdatedAt)

	if err != nil {
		return apperrors.DatabaseError("create collector", err.Error())
	}

	return nil
}

// GetCollectorByID retrieves a collector by ID
func (p *PostgresDB) GetCollectorByID(ctx context.Context, collectorID string) (*models.Collector, error) {
	collector := &models.Collector{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, name, description, hostname, address, version, status, last_seen,
		        certificate_thumbprint, certificate_expires_at, config_version, metrics_count_total,
		        metrics_count_24h, health_check_interval, created_at, updated_at
		 FROM pganalytics.collectors WHERE id::text = $1`,
		collectorID,
	).Scan(
		&collector.ID, &collector.Name, &collector.Description, &collector.Hostname, &collector.Address,
		&collector.Version, &collector.Status, &collector.LastSeen,
		&collector.CertificateThumbprint, &collector.CertificateExpiresAt, &collector.ConfigVersion,
		&collector.MetricsCountTotal, &collector.MetricsCount24h, &collector.HealthCheckInterval,
		&collector.CreatedAt, &collector.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.CollectorNotFound(collectorID)
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get collector", err.Error())
	}

	return collector, nil
}

// GetCollectorByHostname retrieves a collector by hostname
func (p *PostgresDB) GetCollectorByHostname(ctx context.Context, hostname string) (*models.Collector, error) {
	collector := &models.Collector{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, name, description, hostname, address, version, status, last_seen,
		        certificate_thumbprint, certificate_expires_at, config_version, metrics_count_total,
		        metrics_count_24h, health_check_interval, created_at, updated_at
		 FROM pganalytics.collectors WHERE hostname = $1`,
		hostname,
	).Scan(
		&collector.ID, &collector.Name, &collector.Description, &collector.Hostname, &collector.Address,
		&collector.Version, &collector.Status, &collector.LastSeen,
		&collector.CertificateThumbprint, &collector.CertificateExpiresAt, &collector.ConfigVersion,
		&collector.MetricsCountTotal, &collector.MetricsCount24h, &collector.HealthCheckInterval,
		&collector.CreatedAt, &collector.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found, but not an error
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get collector by hostname", err.Error())
	}

	return collector, nil
}

// ListCollectors lists all collectors with pagination
func (p *PostgresDB) ListCollectors(ctx context.Context, offset, limit int) ([]*models.Collector, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT id, name, description, hostname, address, version, status, last_seen,
		        certificate_thumbprint, certificate_expires_at, config_version, metrics_count_total,
		        metrics_count_24h, health_check_interval, created_at, updated_at
		 FROM pganalytics.collectors
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)

	if err != nil {
		return nil, apperrors.DatabaseError("list collectors", err.Error())
	}
	defer rows.Close()

	var collectors []*models.Collector
	for rows.Next() {
		collector := &models.Collector{}
		if err := rows.Scan(
			&collector.ID, &collector.Name, &collector.Description, &collector.Hostname, &collector.Address,
			&collector.Version, &collector.Status, &collector.LastSeen,
			&collector.CertificateThumbprint, &collector.CertificateExpiresAt, &collector.ConfigVersion,
			&collector.MetricsCountTotal, &collector.MetricsCount24h, &collector.HealthCheckInterval,
			&collector.CreatedAt, &collector.UpdatedAt,
		); err != nil {
			return nil, apperrors.DatabaseError("scan collector row", err.Error())
		}
		collectors = append(collectors, collector)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("iterate collectors", err.Error())
	}

	return collectors, nil
}

// UpdateCollectorStatus updates the status of a collector
func (p *PostgresDB) UpdateCollectorStatus(ctx context.Context, collectorID, status string) error {
	_, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.collectors SET status = $1, last_seen = CURRENT_TIMESTAMP WHERE id::text = $2`,
		status, collectorID,
	)

	if err != nil {
		return apperrors.DatabaseError("update collector status", err.Error())
	}

	return nil
}

// UpdateCollectorMetricsCount updates metrics counters
func (p *PostgresDB) UpdateCollectorMetricsCount(ctx context.Context, collectorID string, count int) error {
	_, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.collectors
		 SET metrics_count_total = metrics_count_total + $1,
		     metrics_count_24h = metrics_count_24h + $1,
		     last_seen = CURRENT_TIMESTAMP
		 WHERE id::text = $2`,
		count, collectorID,
	)

	if err != nil {
		return apperrors.DatabaseError("update collector metrics count", err.Error())
	}

	return nil
}

// ============================================================================
// API TOKEN OPERATIONS
// ============================================================================

// CreateAPIToken creates a new API token
func (p *PostgresDB) CreateAPIToken(ctx context.Context, token *models.APIToken) error {
	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.api_tokens (collector_id, user_id, token_hash, description, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		token.CollectorID, token.UserID, token.TokenHash, token.Description, token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)

	if err != nil {
		return apperrors.DatabaseError("create api token", err.Error())
	}

	return nil
}

// GetAPITokenByHash retrieves an API token by hash
func (p *PostgresDB) GetAPITokenByHash(ctx context.Context, tokenHash string) (*models.APIToken, error) {
	token := &models.APIToken{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, collector_id, user_id, token_hash, description, last_used, expires_at, created_at
		 FROM pganalytics.api_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(
		&token.ID, &token.CollectorID, &token.UserID, &token.TokenHash, &token.Description,
		&token.LastUsed, &token.ExpiresAt, &token.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.InvalidToken("Token not found or invalid")
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get api token", err.Error())
	}

	return token, nil
}

// UpdateAPITokenLastUsed updates the last used timestamp
func (p *PostgresDB) UpdateAPITokenLastUsed(ctx context.Context, tokenID int) error {
	_, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.api_tokens SET last_used = CURRENT_TIMESTAMP WHERE id = $1`,
		tokenID,
	)

	if err != nil {
		return apperrors.DatabaseError("update token last used", err.Error())
	}

	return nil
}

// ============================================================================
// AUDIT LOG OPERATIONS
// ============================================================================

// CreateAuditLog creates an audit log entry
func (p *PostgresDB) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.audit_log (user_id, action, resource_type, resource_id, changes, ip_address)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		log.UserID, log.Action, log.ResourceType, log.ResourceID, log.Changes, log.IPAddress,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return apperrors.DatabaseError("create audit log", err.Error())
	}

	return nil
}

// ============================================================================
// SERVER & DATABASE OPERATIONS
// ============================================================================

// GetServerByID retrieves a server by ID
func (p *PostgresDB) GetServerByID(ctx context.Context, serverID int) (*models.Server, error) {
	server := &models.Server{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, name, description, hostname, address, environment, collector_id, is_active, created_at, updated_at
		 FROM pganalytics.servers WHERE id = $1`,
		serverID,
	).Scan(
		&server.ID, &server.Name, &server.Description, &server.Hostname, &server.Address,
		&server.Environment, &server.CollectorID, &server.IsActive, &server.CreatedAt, &server.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("Server not found", fmt.Sprintf("Server ID %d not found", serverID))
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get server", err.Error())
	}

	return server, nil
}

// ListServers lists all active servers
func (p *PostgresDB) ListServers(ctx context.Context, offset, limit int) ([]*models.Server, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT id, name, description, hostname, address, environment, collector_id, is_active, created_at, updated_at
		 FROM pganalytics.servers
		 WHERE is_active = true
		 ORDER BY created_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)

	if err != nil {
		return nil, apperrors.DatabaseError("list servers", err.Error())
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		server := &models.Server{}
		if err := rows.Scan(
			&server.ID, &server.Name, &server.Description, &server.Hostname, &server.Address,
			&server.Environment, &server.CollectorID, &server.IsActive, &server.CreatedAt, &server.UpdatedAt,
		); err != nil {
			return nil, apperrors.DatabaseError("scan server row", err.Error())
		}
		servers = append(servers, server)
	}

	return servers, rows.Err()
}
