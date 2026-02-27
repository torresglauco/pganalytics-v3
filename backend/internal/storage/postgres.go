package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/lib/pq"
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
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(15)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Set search_path to include pganalytics schema
	if _, err := db.ExecContext(ctx, "SET search_path TO pganalytics, public"); err != nil {
		return nil, apperrors.DatabaseError("set search_path", err.Error())
	}

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
		`SELECT id, username, email, password_hash, full_name, role, is_active, last_login, created_at, updated_at
		 FROM pganalytics.users WHERE username = $1`,
		username,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
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
		`SELECT id, username, email, password_hash, full_name, role, is_active, last_login, created_at, updated_at
		 FROM pganalytics.users WHERE id = $1`,
		userID,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
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

// CreateUser creates a new user with hashed password (role defaults to 'user')
func (p *PostgresDB) CreateUser(ctx context.Context, username, email, passwordHash, fullName string) (*models.User, error) {
	return p.CreateUserWithRole(ctx, username, email, passwordHash, fullName, "user")
}

// CreateUserWithRole creates a new user with specified role
func (p *PostgresDB) CreateUserWithRole(ctx context.Context, username, email, passwordHash, fullName, role string) (*models.User, error) {
	user := &models.User{}

	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.users (username, email, password_hash, full_name, role, is_active)
		 VALUES ($1, $2, $3, $4, $5, true)
		 RETURNING id, username, email, password_hash, full_name, role, is_active, last_login, created_at, updated_at`,
		username, email, passwordHash, fullName, role,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.Role, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"" ||
			err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return nil, apperrors.BadRequest("User already exists", "Username or email already in use")
		}
		return nil, apperrors.DatabaseError("create user", err.Error())
	}

	return user, nil
}

// ListUsers retrieves all users from the database
func (p *PostgresDB) ListUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT id, username, email, password_hash, full_name, role, is_active, last_login, created_at, updated_at
		 FROM pganalytics.users
		 ORDER BY created_at DESC`,
	)

	if err != nil {
		return nil, apperrors.DatabaseError("list users", err.Error())
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
			&user.Role, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan user", err.Error())
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.DatabaseError("list users", err.Error())
	}

	return users, nil
}

// UpdateUser updates user information
func (p *PostgresDB) UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) (*models.User, error) {
	user := &models.User{}

	// Build dynamic UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if role, ok := updates["role"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, role)
		argIndex++
	}

	if isActive, ok := updates["is_active"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, isActive)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil, apperrors.BadRequest("No fields to update", "")
	}

	// Add WHERE clause
	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID)

	query := fmt.Sprintf(
		`UPDATE pganalytics.users
		 SET %s
		 WHERE id::text = $%d
		 RETURNING id, username, email, password_hash, full_name, role, is_active, last_login, created_at, updated_at`,
		strings.Join(setClauses[:len(setClauses)-1], ", "),
		argIndex,
	)

	err := p.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.Role, &user.IsActive, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("User not found", "")
	}
	if err != nil {
		return nil, apperrors.DatabaseError("update user", err.Error())
	}

	return user, nil
}

// DeleteUser deletes a user from the database
func (p *PostgresDB) DeleteUser(ctx context.Context, userID string) error {
	result, err := p.db.ExecContext(
		ctx,
		`DELETE FROM pganalytics.users WHERE id::text = $1`,
		userID,
	)

	if err != nil {
		return apperrors.DatabaseError("delete user", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("delete user", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("User not found", "")
	}

	return nil
}

// ResetUserPassword updates a user's password (admin action)
func (p *PostgresDB) ResetUserPassword(ctx context.Context, userID int, newPasswordHash string) error {
	result, err := p.db.ExecContext(
		ctx,
		`UPDATE pganalytics.users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		newPasswordHash,
		userID,
	)

	if err != nil {
		return apperrors.DatabaseError("reset user password", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("reset user password", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("User not found", fmt.Sprintf("User ID %d not found", userID))
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

// ListCollectorsWithTotal lists collectors with pagination and returns total count
func (p *PostgresDB) ListCollectorsWithTotal(ctx context.Context, page, pageSize int) ([]*models.Collector, int, error) {
	// Get total count
	var total int
	countErr := p.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pganalytics.collectors`).Scan(&total)
	if countErr != nil {
		return nil, 0, apperrors.DatabaseError("count collectors", countErr.Error())
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get collectors for this page
	collectors, err := p.ListCollectors(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// Return collectors (could be empty slice if query didn't find anything)
	if collectors == nil {
		collectors = []*models.Collector{}
	}

	return collectors, total, nil
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

// DeleteCollector deletes a collector from the database
func (p *PostgresDB) DeleteCollector(ctx context.Context, collectorID string) error {
	result, err := p.db.ExecContext(
		ctx,
		`DELETE FROM pganalytics.collectors WHERE id::text = $1`,
		collectorID,
	)

	if err != nil {
		return apperrors.DatabaseError("delete collector", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("delete collector", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("Collector not found", "")
	}

	return nil
}

// GetCollectorConfig retrieves the latest configuration for a collector
func (p *PostgresDB) GetCollectorConfig(ctx context.Context, collectorID string) (*models.CollectorConfig, error) {
	config := &models.CollectorConfig{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, collector_id, version, config, created_at, updated_by
		 FROM pganalytics.collector_config
		 WHERE collector_id::text = $1
		 ORDER BY version DESC
		 LIMIT 1`,
		collectorID,
	).Scan(
		&config.ID, &config.CollectorID, &config.Version, &config.Config,
		&config.CreatedAt, &config.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("collector configuration", collectorID)
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get collector config", err.Error())
	}

	return config, nil
}

// CreateCollectorConfig creates a new configuration version for a collector
func (p *PostgresDB) CreateCollectorConfig(ctx context.Context, config *models.CollectorConfig) error {
	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO pganalytics.collector_config (collector_id, config, version, updated_by)
		 VALUES ($1, $2, COALESCE((SELECT MAX(version) FROM pganalytics.collector_config WHERE collector_id = $1), 0) + 1, $3)
		 RETURNING id, version, created_at`,
		config.CollectorID, config.Config, config.UpdatedBy,
	).Scan(&config.ID, &config.Version, &config.CreatedAt)

	if err != nil {
		return apperrors.DatabaseError("create collector config", err.Error())
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

// ============================================================================
// QUERY STATISTICS METHODS
// ============================================================================

// InsertQueryStats inserts query statistics from collector
func (p *PostgresDB) InsertQueryStats(ctx context.Context, collectorID string, stats []*models.QueryStats) error {
	if len(stats) == 0 {
		return nil
	}

	// Prepare bulk insert statement
	query := `INSERT INTO metrics_pg_stats_query (
		time, collector_id, database_name, user_name, query_hash, query_text,
		calls, total_time, mean_time, min_time, max_time, stddev_time, rows,
		shared_blks_hit, shared_blks_read, shared_blks_dirtied, shared_blks_written,
		local_blks_hit, local_blks_read, local_blks_dirtied, local_blks_written,
		temp_blks_read, temp_blks_written, blk_read_time, blk_write_time,
		wal_records, wal_fpi, wal_bytes, query_plan_time, query_exec_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
	         $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)`

	// Use transaction for bulk insert
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return apperrors.DatabaseError("prepare statement", err.Error())
	}
	defer stmt.Close()

	for _, stat := range stats {
		if _, err := stmt.ExecContext(ctx,
			stat.Time, collectorID, stat.DatabaseName, stat.UserName, stat.QueryHash,
			stat.QueryText, stat.Calls, stat.TotalTime, stat.MeanTime, stat.MinTime,
			stat.MaxTime, stat.StddevTime, stat.Rows, stat.SharedBlksHit, stat.SharedBlksRead,
			stat.SharedBlksDirtied, stat.SharedBlksWritten, stat.LocalBlksHit, stat.LocalBlksRead,
			stat.LocalBlksDirtied, stat.LocalBlksWritten, stat.TempBlksRead, stat.TempBlksWritten,
			stat.BlkReadTime, stat.BlkWriteTime, stat.WalRecords, stat.WalFpi, stat.WalBytes,
			stat.QueryPlanTime, stat.QueryExecTime,
		); err != nil {
			return apperrors.DatabaseError("insert query stats", err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return apperrors.DatabaseError("commit transaction", err.Error())
	}

	return nil
}

// GetTopSlowQueries returns the N slowest queries in a time range
func (p *PostgresDB) GetTopSlowQueries(ctx context.Context, collectorID string, limit int, since time.Time) ([]*models.QueryStats, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT time, collector_id, database_name, user_name, query_hash, query_text,
		       calls, total_time, mean_time, min_time, max_time, stddev_time, rows,
		       shared_blks_hit, shared_blks_read, shared_blks_dirtied, shared_blks_written,
		       local_blks_hit, local_blks_read, local_blks_dirtied, local_blks_written,
		       temp_blks_read, temp_blks_written, blk_read_time, blk_write_time,
		       wal_records, wal_fpi, wal_bytes, query_plan_time, query_exec_time
		FROM metrics_pg_stats_query
		WHERE collector_id = $1 AND time >= $2
		ORDER BY max_time DESC
		LIMIT $3
	`, collectorID, since, limit)

	if err != nil {
		return nil, apperrors.DatabaseError("query top slow queries", err.Error())
	}
	defer rows.Close()

	var queries []*models.QueryStats
	for rows.Next() {
		q := &models.QueryStats{}
		err := rows.Scan(
			&q.Time, &q.CollectorID, &q.DatabaseName, &q.UserName, &q.QueryHash, &q.QueryText,
			&q.Calls, &q.TotalTime, &q.MeanTime, &q.MinTime, &q.MaxTime, &q.StddevTime, &q.Rows,
			&q.SharedBlksHit, &q.SharedBlksRead, &q.SharedBlksDirtied, &q.SharedBlksWritten,
			&q.LocalBlksHit, &q.LocalBlksRead, &q.LocalBlksDirtied, &q.LocalBlksWritten,
			&q.TempBlksRead, &q.TempBlksWritten, &q.BlkReadTime, &q.BlkWriteTime,
			&q.WalRecords, &q.WalFpi, &q.WalBytes, &q.QueryPlanTime, &q.QueryExecTime,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan slow query row", err.Error())
		}
		queries = append(queries, q)
	}

	return queries, rows.Err()
}

// GetTopFrequentQueries returns the N most frequently executed queries
func (p *PostgresDB) GetTopFrequentQueries(ctx context.Context, collectorID string, limit int, since time.Time) ([]*models.QueryStats, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT time, collector_id, database_name, user_name, query_hash, query_text,
		       calls, total_time, mean_time, min_time, max_time, stddev_time, rows,
		       shared_blks_hit, shared_blks_read, shared_blks_dirtied, shared_blks_written,
		       local_blks_hit, local_blks_read, local_blks_dirtied, local_blks_written,
		       temp_blks_read, temp_blks_written, blk_read_time, blk_write_time,
		       wal_records, wal_fpi, wal_bytes, query_plan_time, query_exec_time
		FROM metrics_pg_stats_query
		WHERE collector_id = $1 AND time >= $2
		ORDER BY calls DESC
		LIMIT $3
	`, collectorID, since, limit)

	if err != nil {
		return nil, apperrors.DatabaseError("query top frequent queries", err.Error())
	}
	defer rows.Close()

	var queries []*models.QueryStats
	for rows.Next() {
		q := &models.QueryStats{}
		err := rows.Scan(
			&q.Time, &q.CollectorID, &q.DatabaseName, &q.UserName, &q.QueryHash, &q.QueryText,
			&q.Calls, &q.TotalTime, &q.MeanTime, &q.MinTime, &q.MaxTime, &q.StddevTime, &q.Rows,
			&q.SharedBlksHit, &q.SharedBlksRead, &q.SharedBlksDirtied, &q.SharedBlksWritten,
			&q.LocalBlksHit, &q.LocalBlksRead, &q.LocalBlksDirtied, &q.LocalBlksWritten,
			&q.TempBlksRead, &q.TempBlksWritten, &q.BlkReadTime, &q.BlkWriteTime,
			&q.WalRecords, &q.WalFpi, &q.WalBytes, &q.QueryPlanTime, &q.QueryExecTime,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan frequent query row", err.Error())
		}
		queries = append(queries, q)
	}

	return queries, rows.Err()
}

// GetQueryTimeline returns time-series data for a specific query
func (p *PostgresDB) GetQueryTimeline(ctx context.Context, queryHash int64, since time.Time) ([]*models.QueryStats, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT time, collector_id, database_name, user_name, query_hash, query_text,
		       calls, total_time, mean_time, min_time, max_time, stddev_time, rows,
		       shared_blks_hit, shared_blks_read, shared_blks_dirtied, shared_blks_written,
		       local_blks_hit, local_blks_read, local_blks_dirtied, local_blks_written,
		       temp_blks_read, temp_blks_written, blk_read_time, blk_write_time,
		       wal_records, wal_fpi, wal_bytes, query_plan_time, query_exec_time
		FROM metrics_pg_stats_query_1h
		WHERE query_hash = $1 AND time >= $2
		ORDER BY time ASC
	`, queryHash, since)

	if err != nil {
		return nil, apperrors.DatabaseError("query timeline", err.Error())
	}
	defer rows.Close()

	var queries []*models.QueryStats
	for rows.Next() {
		q := &models.QueryStats{}
		err := rows.Scan(
			&q.Time, &q.CollectorID, &q.DatabaseName, &q.UserName, &q.QueryHash, &q.QueryText,
			&q.Calls, &q.TotalTime, &q.MeanTime, &q.MinTime, &q.MaxTime, &q.StddevTime, &q.Rows,
			&q.SharedBlksHit, &q.SharedBlksRead, &q.SharedBlksDirtied, &q.SharedBlksWritten,
			&q.LocalBlksHit, &q.LocalBlksRead, &q.LocalBlksDirtied, &q.LocalBlksWritten,
			&q.TempBlksRead, &q.TempBlksWritten, &q.BlkReadTime, &q.BlkWriteTime,
			&q.WalRecords, &q.WalFpi, &q.WalBytes, &q.QueryPlanTime, &q.QueryExecTime,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan timeline row", err.Error())
		}
		queries = append(queries, q)
	}

	return queries, rows.Err()
}

// ============================================================================
// PHASE 4.4: ADVANCED QUERY ANALYSIS METHODS
// ============================================================================

// GetQueryFingerprints returns grouped queries by fingerprint
func (p *PostgresDB) GetQueryFingerprints(ctx context.Context, limit int) ([]*models.QueryFingerprintResponse, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}

	query := `
	SELECT
		fingerprint_hash,
		normalized_text,
		total_calls,
		avg_execution_time,
		first_seen,
		last_seen
	FROM query_fingerprints
	WHERE total_calls > 0
	ORDER BY total_calls DESC
	LIMIT $1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("query fingerprints", err.Error())
	}
	defer rows.Close()

	var fingerprints []*models.QueryFingerprintResponse
	for rows.Next() {
		fp := &models.QueryFingerprintResponse{}
		err := rows.Scan(
			&fp.FingerprintHash,
			&fp.NormalizedQuery,
			&fp.TotalCalls,
			&fp.AvgExecutionTime,
			&fp.FirstSeen,
			&fp.LastSeen,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan fingerprint row", err.Error())
		}
		fingerprints = append(fingerprints, fp)
	}

	return fingerprints, rows.Err()
}

// GetQueriesByFingerprint returns all individual queries for a specific fingerprint
func (p *PostgresDB) GetQueriesByFingerprint(ctx context.Context, fingerprintHash int64, limit int) ([]*models.QueryStats, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}

	query := `
	SELECT
		time, collector_id, database_name, user_name, query_hash, query_text,
		calls, total_time, mean_time, min_time, max_time, stddev_time, rows,
		shared_blks_hit, shared_blks_read, shared_blks_dirtied, shared_blks_written,
		local_blks_hit, local_blks_read, local_blks_dirtied, local_blks_written,
		temp_blks_read, temp_blks_written, blk_read_time, blk_write_time,
		wal_records, wal_fpi, wal_bytes, query_plan_time, query_exec_time
	FROM metrics_pg_stats_query
	WHERE fingerprint_query_hash(normalize_query_text(query_text)) = $1
	ORDER BY time DESC
	LIMIT $2
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, fingerprintHash, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("query by fingerprint", err.Error())
	}
	defer rows.Close()

	var queries []*models.QueryStats
	for rows.Next() {
		q := &models.QueryStats{}
		err := rows.Scan(
			&q.Time, &q.CollectorID, &q.DatabaseName, &q.UserName, &q.QueryHash, &q.QueryText,
			&q.Calls, &q.TotalTime, &q.MeanTime, &q.MinTime, &q.MaxTime, &q.StddevTime, &q.Rows,
			&q.SharedBlksHit, &q.SharedBlksRead, &q.SharedBlksDirtied, &q.SharedBlksWritten,
			&q.LocalBlksHit, &q.LocalBlksRead, &q.LocalBlksDirtied, &q.LocalBlksWritten,
			&q.TempBlksRead, &q.TempBlksWritten, &q.BlkReadTime, &q.BlkWriteTime,
			&q.WalRecords, &q.WalFpi, &q.WalBytes, &q.QueryPlanTime, &q.QueryExecTime,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan query row", err.Error())
		}
		queries = append(queries, q)
	}

	return queries, rows.Err()
}

// GetExplainPlan returns the latest EXPLAIN plan for a query
func (p *PostgresDB) GetExplainPlan(ctx context.Context, queryHash int64) (*models.ExplainPlan, error) {
	query := `
	SELECT
		id, query_hash, query_fingerprint_hash, collected_at, plan_json, plan_text,
		rows_expected, rows_actual, plan_duration_ms, execution_duration_ms,
		has_seq_scan, has_index_scan, has_bitmap_scan, has_nested_loop,
		total_buffers_read, total_buffers_hit
	FROM explain_plans
	WHERE query_hash = $1
	ORDER BY collected_at DESC
	LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	plan := &models.ExplainPlan{}
	err := p.db.QueryRowContext(ctx, query, queryHash).Scan(
		&plan.ID, &plan.QueryHash, &plan.QueryFingerprintHash, &plan.CollectedAt, &plan.PlanJSON, &plan.PlanText,
		&plan.RowsExpected, &plan.RowsActual, &plan.PlanDurationMs, &plan.ExecutionDurationMs,
		&plan.HasSeqScan, &plan.HasIndexScan, &plan.HasBitmapScan, &plan.HasNestedLoop,
		&plan.TotalBuffersRead, &plan.TotalBuffersHit,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No explain plan found
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get explain plan", err.Error())
	}

	return plan, nil
}

// GetQueryAnomalies returns detected anomalies for a query
func (p *PostgresDB) GetQueryAnomalies(ctx context.Context, queryHash int64, days int) ([]*models.QueryAnomaly, error) {
	if days > 30 {
		days = 30
	}
	if days < 1 {
		days = 7
	}

	query := `
	SELECT
		id, query_hash, query_fingerprint_hash, anomaly_type, severity, detected_at,
		metric_name, metric_value, baseline_value, deviation_stddev, z_score,
		raw_metrics_json, resolved, resolved_at
	FROM query_anomalies
	WHERE query_hash = $1 AND detected_at >= NOW() - INTERVAL '1 day' * $2
	ORDER BY detected_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, queryHash, days)
	if err != nil {
		return nil, apperrors.DatabaseError("get query anomalies", err.Error())
	}
	defer rows.Close()

	var anomalies []*models.QueryAnomaly
	for rows.Next() {
		anomaly := &models.QueryAnomaly{}
		err := rows.Scan(
			&anomaly.ID, &anomaly.QueryHash, &anomaly.QueryFingerprintHash, &anomaly.AnomalyType,
			&anomaly.Severity, &anomaly.DetectedAt, &anomaly.MetricName, &anomaly.MetricValue,
			&anomaly.BaselineValue, &anomaly.DeviationStddev, &anomaly.ZScore,
			&anomaly.RawMetricsJSON, &anomaly.Resolved, &anomaly.ResolvedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan anomaly row", err.Error())
		}
		anomalies = append(anomalies, anomaly)
	}

	return anomalies, rows.Err()
}

// CreatePerformanceSnapshot creates a new baseline snapshot of query metrics
func (p *PostgresDB) CreatePerformanceSnapshot(ctx context.Context, name string, description *string, snapshotType string, createdBy *string) (int64, error) {
	var snapshotID int64
	query := `
	SELECT create_performance_snapshot($1, $2, $3, $4)
	`

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second) // Longer timeout for snapshot capture
	defer cancel()

	err := p.db.QueryRowContext(ctx, query, name, description, snapshotType, createdBy).Scan(&snapshotID)
	if err != nil {
		return 0, apperrors.DatabaseError("create snapshot", err.Error())
	}

	return snapshotID, nil
}

// GetPerformanceSnapshots returns all performance snapshots with metadata
func (p *PostgresDB) GetPerformanceSnapshots(ctx context.Context, limit int) ([]*models.PerformanceSnapshot, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}

	query := `
	SELECT id, name, description, snapshot_type, created_at, created_by, metadata_json
	FROM performance_snapshots
	ORDER BY created_at DESC
	LIMIT $1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("get snapshots", err.Error())
	}
	defer rows.Close()

	var snapshots []*models.PerformanceSnapshot
	for rows.Next() {
		snap := &models.PerformanceSnapshot{}
		err := rows.Scan(&snap.ID, &snap.Name, &snap.Description, &snap.SnapshotType, &snap.CreatedAt, &snap.CreatedBy, &snap.MetadataJSON)
		if err != nil {
			return nil, apperrors.DatabaseError("scan snapshot row", err.Error())
		}
		snapshots = append(snapshots, snap)
	}

	return snapshots, rows.Err()
}

// CompareSnapshots compares metrics between two snapshots
func (p *PostgresDB) CompareSnapshots(ctx context.Context, beforeSnapshotID int64, afterSnapshotID int64, limit int) ([]*models.SnapshotComparison, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}

	query := `
	SELECT
		query_hash, database_name, before_calls, after_calls, calls_change, calls_change_percent,
		before_mean_time, after_mean_time, mean_time_change, mean_time_change_percent,
		before_max_time, after_max_time, max_time_change,
		before_cache_hits, after_cache_hits, before_cache_reads, after_cache_reads,
		improvement_status
	FROM compare_snapshots($1, $2)
	LIMIT $3
	`

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, beforeSnapshotID, afterSnapshotID, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("compare snapshots", err.Error())
	}
	defer rows.Close()

	var comparisons []*models.SnapshotComparison
	for rows.Next() {
		comp := &models.SnapshotComparison{}
		err := rows.Scan(
			&comp.QueryHash, &comp.DatabaseName, &comp.BeforeCalls, &comp.AfterCalls,
			&comp.CallsChange, &comp.CallsChangePercent, &comp.BeforeMeanTime, &comp.AfterMeanTime,
			&comp.MeanTimeChange, &comp.MeanTimeChangePercent, &comp.BeforeMaxTime, &comp.AfterMaxTime,
			&comp.MaxTimeChange, &comp.BeforeCacheHits, &comp.AfterCacheHits,
			&comp.BeforeCacheReads, &comp.AfterCacheReads, &comp.ImprovementStatus,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan comparison row", err.Error())
		}
		comparisons = append(comparisons, comp)
	}

	return comparisons, rows.Err()
}

// ============================================================================
// PHASE 4.4.3: INDEX RECOMMENDATIONS
// ============================================================================

// GetIndexRecommendations returns recommended indexes for a specific database
func (p *PostgresDB) GetIndexRecommendations(ctx context.Context, databaseName string, limit int) ([]*models.IndexRecommendation, error) {
	if limit > 50 {
		limit = 50
	}
	if limit < 1 {
		limit = 20
	}

	query := `
	SELECT
		id, collector_id, database_name, schema_name, table_name, column_names,
		create_statement, estimated_improvement_percent, affected_query_count,
		affected_total_time_ms, frequency_score, impact_score, confidence_score,
		dismissed, dismissed_at, dismissed_reason, created_at
	FROM index_recommendations
	WHERE database_name = $1
		AND NOT dismissed
	ORDER BY confidence_score DESC, impact_score DESC
	LIMIT $2
	`

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, databaseName, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("get index recommendations", err.Error())
	}
	defer rows.Close()

	var recommendations []*models.IndexRecommendation
	for rows.Next() {
		rec := &models.IndexRecommendation{}
		err := rows.Scan(
			&rec.ID, &rec.CollectorID, &rec.DatabaseName, &rec.SchemaName, &rec.TableName, pq.Array(&rec.ColumnNames),
			&rec.CreateStatement, &rec.EstimatedImprovementPct, &rec.AffectedQueryCount,
			&rec.AffectedTotalTimeMs, &rec.FrequencyScore, &rec.ImpactScore, &rec.ConfidenceScore,
			&rec.Dismissed, &rec.DismissedAt, &rec.DismissedReason, &rec.CreatedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan index recommendation row", err.Error())
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations, rows.Err()
}

// StoreIndexRecommendations stores index recommendations (upsert)
func (p *PostgresDB) StoreIndexRecommendations(ctx context.Context, recommendations []*models.IndexRecommendation) error {
	if len(recommendations) == 0 {
		return nil
	}

	query := `
	INSERT INTO index_recommendations (
		collector_id, database_name, schema_name, table_name, column_names, column_names_str,
		create_statement, estimated_improvement_percent, affected_query_count,
		affected_total_time_ms, frequency_score, impact_score, confidence_score, created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW())
	ON CONFLICT (database_name, schema_name, table_name, column_names_str) DO UPDATE SET
		estimated_improvement_percent = EXCLUDED.estimated_improvement_percent,
		affected_query_count = EXCLUDED.affected_query_count,
		affected_total_time_ms = EXCLUDED.affected_total_time_ms,
		frequency_score = EXCLUDED.frequency_score,
		impact_score = EXCLUDED.impact_score,
		confidence_score = EXCLUDED.confidence_score
	RETURNING id
	`

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	for _, rec := range recommendations {
		// Convert column names array to string for unique constraint
		var columnNamesStr string
		if cols, ok := rec.ColumnNames.([]string); ok {
			columnNamesStr = strings.Join(cols, ",")
		} else if cols, ok := rec.ColumnNames.([]interface{}); ok {
			var colStrs []string
			for _, c := range cols {
				if s, ok := c.(string); ok {
					colStrs = append(colStrs, s)
				}
			}
			columnNamesStr = strings.Join(colStrs, ",")
		}

		_, err := p.db.ExecContext(ctx, query,
			rec.CollectorID, rec.DatabaseName, rec.SchemaName, rec.TableName, pq.Array(rec.ColumnNames), columnNamesStr,
			rec.CreateStatement, rec.EstimatedImprovementPct, rec.AffectedQueryCount,
			rec.AffectedTotalTimeMs, rec.FrequencyScore, rec.ImpactScore, rec.ConfidenceScore,
		)
		if err != nil {
			return apperrors.DatabaseError("store index recommendation", err.Error())
		}
	}

	return nil
}

// DismissIndexRecommendation marks a recommendation as dismissed
func (p *PostgresDB) DismissIndexRecommendation(ctx context.Context, recommendationID int64, reason *string) error {
	query := `
	UPDATE index_recommendations
	SET dismissed = TRUE, dismissed_at = NOW(), dismissed_reason = $2
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := p.db.ExecContext(ctx, query, recommendationID, reason)
	if err != nil {
		return apperrors.DatabaseError("dismiss index recommendation", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("check rows affected", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("index recommendation", fmt.Sprintf("%d", recommendationID))
	}

	return nil
}

// GetIndexRecommendationByID retrieves a specific index recommendation by ID
func (p *PostgresDB) GetIndexRecommendationByID(ctx context.Context, recommendationID int64) (*models.IndexRecommendation, error) {
	query := `
	SELECT
		id, collector_id, database_name, schema_name, table_name, column_names,
		create_statement, estimated_improvement_percent, affected_query_count,
		affected_total_time_ms, frequency_score, impact_score, confidence_score,
		dismissed, dismissed_at, dismissed_reason, created_at
	FROM index_recommendations
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rec := &models.IndexRecommendation{}
	err := p.db.QueryRowContext(ctx, query, recommendationID).Scan(
		&rec.ID, &rec.CollectorID, &rec.DatabaseName, &rec.SchemaName, &rec.TableName, pq.Array(&rec.ColumnNames),
		&rec.CreateStatement, &rec.EstimatedImprovementPct, &rec.AffectedQueryCount,
		&rec.AffectedTotalTimeMs, &rec.FrequencyScore, &rec.ImpactScore, &rec.ConfidenceScore,
		&rec.Dismissed, &rec.DismissedAt, &rec.DismissedReason, &rec.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NotFound("index recommendation", fmt.Sprintf("%d", recommendationID))
		}
		return nil, apperrors.DatabaseError("get index recommendation", err.Error())
	}

	return rec, nil
}

// GenerateIndexRecommendations analyzes recent EXPLAIN plans to generate recommendations
// This is called periodically (e.g., via background job or manual trigger)
func (p *PostgresDB) GenerateIndexRecommendations(ctx context.Context, databaseName string, collectorID *string) (int, error) {
	// Phase 4.4.3 Implementation:
	// Analyze EXPLAIN plans from the last 24 hours to identify:
	// 1. Seq Scan operations (missing indexes)
	// 2. Frequency of the pattern (how many times)
	// 3. Estimated improvement (query count * mean time savings)
	// 4. Confidence score based on frequency and impact

	query := `
	INSERT INTO index_recommendations (
		collector_id, database_name, schema_name, table_name, column_names, column_names_str,
		create_statement, estimated_improvement_percent, affected_query_count,
		affected_total_time_ms, frequency_score, impact_score, confidence_score, created_at
	)
	SELECT
		COALESCE($1::uuid, ep.collector_id),
		$2,
		'public',
		json_extract_path_text(ep.plan_json, 'Plan', 'Relation Name')::varchar(63),
		ARRAY[json_extract_path_text(ep.plan_json, 'Plan', 'Filter')],
		json_extract_path_text(ep.plan_json, 'Plan', 'Filter'),
		'CREATE INDEX CONCURRENTLY idx_' || json_extract_path_text(ep.plan_json, 'Plan', 'Relation Name') ||
			'_' || REPLACE(json_extract_path_text(ep.plan_json, 'Plan', 'Filter'), ' ', '_') ||
			' ON ' || json_extract_path_text(ep.plan_json, 'Plan', 'Relation Name') ||
			' (' || json_extract_path_text(ep.plan_json, 'Plan', 'Filter') || ')',
		50.0,
		COUNT(DISTINCT ep.query_hash),
		SUM(COALESCE(mq.total_time, 0)),
		0.7,
		0.8,
		0.8,
		NOW()
	FROM explain_plans ep
	JOIN metrics_pg_stats_query mq ON ep.query_hash = mq.query_hash
	WHERE ep.has_seq_scan = true
		AND ep.collected_at >= NOW() - INTERVAL '24 hours'
		AND mq.mean_time > 100  -- Only for queries taking >100ms
	GROUP BY
		json_extract_path_text(ep.plan_json, 'Plan', 'Relation Name'),
		json_extract_path_text(ep.plan_json, 'Plan', 'Filter')
	ON CONFLICT (database_name, schema_name, table_name, column_names_str) DO UPDATE SET
		affected_query_count = EXCLUDED.affected_query_count,
		affected_total_time_ms = EXCLUDED.affected_total_time_ms,
		frequency_score = EXCLUDED.frequency_score,
		confidence_score = EXCLUDED.confidence_score
	`

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := p.db.ExecContext(ctx, query, collectorID, databaseName)
	if err != nil {
		return 0, apperrors.DatabaseError("generate index recommendations", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, apperrors.DatabaseError("get rows affected", err.Error())
	}

	return int(rowsAffected), nil
}

// ============================================================================
// PHASE 4.4.4: ANOMALY DETECTION
// ============================================================================

// GetAnomaliesBySeverity returns anomalies filtered by severity level
func (p *PostgresDB) GetAnomaliesBySeverity(ctx context.Context, severity string, limit int) ([]*models.QueryAnomaly, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 50
	}

	// Validate severity value
	if severity != "low" && severity != "medium" && severity != "high" {
		severity = "high"
	}

	query := `
	SELECT
		id, query_hash, query_fingerprint_hash, anomaly_type, severity, detected_at,
		metric_name, metric_value, baseline_value, deviation_stddev, z_score,
		raw_metrics_json, resolved, resolved_at
	FROM query_anomalies
	WHERE severity = $1
		AND NOT resolved
	ORDER BY detected_at DESC
	LIMIT $2
	`

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, severity, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("get anomalies by severity", err.Error())
	}
	defer rows.Close()

	var anomalies []*models.QueryAnomaly
	for rows.Next() {
		anomaly := &models.QueryAnomaly{}
		err := rows.Scan(
			&anomaly.ID, &anomaly.QueryHash, &anomaly.QueryFingerprintHash, &anomaly.AnomalyType,
			&anomaly.Severity, &anomaly.DetectedAt, &anomaly.MetricName, &anomaly.MetricValue,
			&anomaly.BaselineValue, &anomaly.DeviationStddev, &anomaly.ZScore,
			&anomaly.RawMetricsJSON, &anomaly.Resolved, &anomaly.ResolvedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan anomaly row", err.Error())
		}
		anomalies = append(anomalies, anomaly)
	}

	return anomalies, rows.Err()
}

// StoreAnomalies stores detected anomalies in the database
func (p *PostgresDB) StoreAnomalies(ctx context.Context, anomalies []*models.QueryAnomaly) error {
	if len(anomalies) == 0 {
		return nil
	}

	query := `
	INSERT INTO query_anomalies (
		query_hash, query_fingerprint_hash, anomaly_type, severity, detected_at,
		metric_name, metric_value, baseline_value, deviation_stddev, z_score,
		raw_metrics_json, resolved
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT (query_hash, anomaly_type, metric_name, detected_at) DO NOTHING
	`

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return apperrors.DatabaseError("prepare statement", err.Error())
	}
	defer stmt.Close()

	for _, anomaly := range anomalies {
		_, err := stmt.ExecContext(ctx,
			anomaly.QueryHash, anomaly.QueryFingerprintHash, anomaly.AnomalyType, anomaly.Severity,
			anomaly.DetectedAt, anomaly.MetricName, anomaly.MetricValue, anomaly.BaselineValue,
			anomaly.DeviationStddev, anomaly.ZScore, anomaly.RawMetricsJSON, anomaly.Resolved,
		)
		if err != nil {
			return apperrors.DatabaseError("insert anomaly", err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return apperrors.DatabaseError("commit transaction", err.Error())
	}

	return nil
}

// CalculateBaselineAndDetectAnomalies executes the database function to calculate baselines and detect anomalies
func (p *PostgresDB) CalculateBaselineAndDetectAnomalies(ctx context.Context) error {
	query := `SELECT calculate_baselines_and_anomalies()`

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // Allow up to 60 seconds for calculation
	defer cancel()

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return apperrors.DatabaseError("calculate baselines and anomalies", err.Error())
	}

	return nil
}

// GetQueryBaseline retrieves the baseline metrics for a query
func (p *PostgresDB) GetQueryBaseline(ctx context.Context, queryHash int64, metricName string) (*models.QueryBaseline, error) {
	query := `
	SELECT
		id, query_hash, metric_name, baseline_value, stddev_value,
		baseline_period_days, last_updated, min_value, max_value
	FROM query_baselines
	WHERE query_hash = $1 AND metric_name = $2
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	baseline := &models.QueryBaseline{}
	err := p.db.QueryRowContext(ctx, query, queryHash, metricName).Scan(
		&baseline.ID, &baseline.QueryHash, &baseline.MetricName, &baseline.BaselineValue,
		&baseline.StddevValue, &baseline.BaselinePeriodDays, &baseline.LastUpdated,
		&baseline.MinValue, &baseline.MaxValue,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No baseline found, not an error
		}
		return nil, apperrors.DatabaseError("get query baseline", err.Error())
	}

	return baseline, nil
}

// GetQueryBaselines retrieves all baselines for a query
func (p *PostgresDB) GetQueryBaselines(ctx context.Context, queryHash int64) ([]*models.QueryBaseline, error) {
	query := `
	SELECT
		id, query_hash, metric_name, baseline_value, stddev_value,
		baseline_period_days, last_updated, min_value, max_value
	FROM query_baselines
	WHERE query_hash = $1
	ORDER BY metric_name ASC
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := p.db.QueryContext(ctx, query, queryHash)
	if err != nil {
		return nil, apperrors.DatabaseError("get query baselines", err.Error())
	}
	defer rows.Close()

	var baselines []*models.QueryBaseline
	for rows.Next() {
		baseline := &models.QueryBaseline{}
		err := rows.Scan(
			&baseline.ID, &baseline.QueryHash, &baseline.MetricName, &baseline.BaselineValue,
			&baseline.StddevValue, &baseline.BaselinePeriodDays, &baseline.LastUpdated,
			&baseline.MinValue, &baseline.MaxValue,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan baseline row", err.Error())
		}
		baselines = append(baselines, baseline)
	}

	return baselines, rows.Err()
}

// ResolveAnomaly marks an anomaly as resolved
func (p *PostgresDB) ResolveAnomaly(ctx context.Context, anomalyID int64) error {
	query := `
	UPDATE query_anomalies
	SET resolved = TRUE, resolved_at = NOW()
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := p.db.ExecContext(ctx, query, anomalyID)
	if err != nil {
		return apperrors.DatabaseError("resolve anomaly", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("check rows affected", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("anomaly", fmt.Sprintf("%d", anomalyID))
	}

	return nil
}

// ============================================================================
// PHASE 4.5: ML-BASED QUERY OPTIMIZATION SUGGESTIONS
// ============================================================================

// DetectWorkloadPatterns detects recurring patterns in query execution
// Analyzes 30-day rolling window by default, minimum 7 days, maximum 365 days
func (p *PostgresDB) DetectWorkloadPatterns(ctx context.Context, databaseName string, lookbackDays int) (int, error) {
	// Validate database name
	if databaseName == "" {
		return 0, apperrors.BadRequest("database_name", "Database name is required")
	}

	// Validate lookback days
	if lookbackDays < 7 {
		lookbackDays = 7
	}
	if lookbackDays > 365 {
		lookbackDays = 365
	}

	// Call PostgreSQL function to analyze patterns
	// Returns number of patterns detected
	query := `SELECT COUNT(*) FROM detect_workload_patterns($1, $2)`

	var count int
	err := p.db.QueryRowContext(
		ctx,
		query,
		databaseName,
		lookbackDays,
	).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, apperrors.DatabaseError("detect workload patterns", err.Error())
	}

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return count, nil
}

// GetWorkloadPatterns retrieves detected workload patterns
func (p *PostgresDB) GetWorkloadPatterns(ctx context.Context, databaseName, patternType string, limit int) ([]models.WorkloadPattern, error) {
	query := `
		SELECT id, database_name, pattern_type, pattern_metadata, detection_timestamp, description, affected_query_count
		FROM workload_patterns
		WHERE 1=1`

	args := []interface{}{}
	argNum := 1

	if databaseName != "" {
		query += fmt.Sprintf(` AND database_name = $%d`, argNum)
		args = append(args, databaseName)
		argNum++
	}

	if patternType != "" {
		query += fmt.Sprintf(` AND pattern_type = $%d`, argNum)
		args = append(args, patternType)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY detection_timestamp DESC LIMIT $%d`, argNum)
	args = append(args, limit)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("get workload patterns", err.Error())
	}
	defer rows.Close()

	var patterns []models.WorkloadPattern
	for rows.Next() {
		var p models.WorkloadPattern
		var metadata interface{}

		err := rows.Scan(
			&p.ID,
			&p.DatabaseName,
			&p.PatternType,
			&metadata,
			&p.DetectionTimestamp,
			&p.Description,
			&p.AffectedQueryCount,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan pattern", err.Error())
		}

		// Convert metadata to map
		if metadata != nil {
			p.PatternMetadata = metadata.(map[string]interface{})
		}

		patterns = append(patterns, p)
	}

	return patterns, rows.Err()
}

// GenerateRewriteSuggestions analyzes a query and generates rewrite suggestions
// Detects anti-patterns: N+1, inefficient joins, missing indexes, subqueries, IN vs ANY
func (p *PostgresDB) GenerateRewriteSuggestions(ctx context.Context, queryHash int64) (int, error) {
	// Validate query_hash
	if queryHash <= 0 {
		return 0, apperrors.BadRequest("query_hash", "Must be a positive integer")
	}

	// Call PostgreSQL function to generate suggestions
	// Returns count of suggestions generated for this query
	query := `SELECT COUNT(*) FROM generate_rewrite_suggestions($1)`

	var count int
	err := p.db.QueryRowContext(
		ctx,
		query,
		queryHash,
	).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return 0, apperrors.DatabaseError("generate rewrite suggestions", err.Error())
	}

	if err == sql.ErrNoRows {
		return 0, nil
	}

	return count, nil
}

// GetRewriteSuggestions retrieves rewrite suggestions for a query
func (p *PostgresDB) GetRewriteSuggestions(ctx context.Context, queryHash int64, limit int) ([]models.QueryRewriteSuggestion, error) {
	query := `
		SELECT id, query_hash, fingerprint_hash, suggestion_type, description, original_query,
		       suggested_rewrite, reasoning, estimated_improvement_percent, confidence_score,
		       dismissed, implemented, implementation_notes, created_at, updated_at
		FROM query_rewrite_suggestions
		WHERE query_hash = $1 AND dismissed = FALSE
		ORDER BY confidence_score DESC, estimated_improvement_percent DESC
		LIMIT $2`

	rows, err := p.db.QueryContext(ctx, query, queryHash, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("get rewrite suggestions", err.Error())
	}
	defer rows.Close()

	var suggestions []models.QueryRewriteSuggestion
	for rows.Next() {
		var s models.QueryRewriteSuggestion
		err := rows.Scan(
			&s.ID, &s.QueryHash, &s.FingerprintHash, &s.SuggestionType, &s.Description,
			&s.OriginalQuery, &s.SuggestedRewrite, &s.Reasoning, &s.EstimatedImprovementPct,
			&s.ConfidenceScore, &s.Dismissed, &s.Implemented, &s.ImplementationNotes,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan suggestion", err.Error())
		}
		suggestions = append(suggestions, s)
	}

	return suggestions, rows.Err()
}

// OptimizeParameters analyzes a query and generates parameter optimization suggestions
// Detects: missing LIMIT, work_mem optimization opportunities, batch size recommendations
func (p *PostgresDB) OptimizeParameters(ctx context.Context, queryHash int64) ([]models.ParameterTuningSuggestion, error) {
	// Validate query_hash
	if queryHash <= 0 {
		return nil, apperrors.BadRequest("query_hash", "Must be a positive integer")
	}

	// Call PostgreSQL function to generate suggestions
	// Returns list of (suggestion_count, parameter_types[])
	_, err := p.db.ExecContext(
		ctx,
		`SELECT optimize_parameters($1)`,
		queryHash,
	)
	if err != nil {
		return nil, apperrors.DatabaseError("optimize parameters", err.Error())
	}

	// Retrieve all generated suggestions for this query
	query := `
		SELECT id, query_hash, fingerprint_hash, parameter_name, current_value, recommended_value,
		       reasoning, estimated_improvement_percent, confidence_score, created_at, updated_at
		FROM parameter_tuning_suggestions
		WHERE query_hash = $1
		ORDER BY confidence_score DESC, estimated_improvement_percent DESC`

	rows, err := p.db.QueryContext(ctx, query, queryHash)
	if err != nil {
		return nil, apperrors.DatabaseError("get parameter suggestions", err.Error())
	}
	defer rows.Close()

	var suggestions []models.ParameterTuningSuggestion
	for rows.Next() {
		var s models.ParameterTuningSuggestion
		err := rows.Scan(
			&s.ID, &s.QueryHash, &s.FingerprintHash, &s.ParameterName, &s.CurrentValue,
			&s.RecommendedValue, &s.Reasoning, &s.EstimatedImprovementPct, &s.ConfidenceScore,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			continue
		}
		suggestions = append(suggestions, s)
	}
	return suggestions, rows.Err()
}

// GetParameterOptimizationSuggestions retrieves parameter tuning suggestions for a query
func (p *PostgresDB) GetParameterOptimizationSuggestions(ctx context.Context, queryHash int64, limit int) ([]models.ParameterTuningSuggestion, error) {
	// Validate inputs
	if queryHash <= 0 {
		return nil, apperrors.BadRequest("query_hash", "Must be a positive integer")
	}

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	query := `
		SELECT id, query_hash, fingerprint_hash, parameter_name, current_value, recommended_value,
		       reasoning, estimated_improvement_percent, confidence_score, created_at, updated_at
		FROM parameter_tuning_suggestions
		WHERE query_hash = $1
		ORDER BY confidence_score DESC, estimated_improvement_percent DESC
		LIMIT $2`

	rows, err := p.db.QueryContext(ctx, query, queryHash, limit)
	if err != nil {
		return nil, apperrors.DatabaseError("get parameter suggestions", err.Error())
	}
	defer rows.Close()

	var suggestions []models.ParameterTuningSuggestion
	for rows.Next() {
		var s models.ParameterTuningSuggestion
		err := rows.Scan(
			&s.ID, &s.QueryHash, &s.FingerprintHash, &s.ParameterName, &s.CurrentValue,
			&s.RecommendedValue, &s.Reasoning, &s.EstimatedImprovementPct, &s.ConfidenceScore,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			continue
		}
		suggestions = append(suggestions, s)
	}

	if len(suggestions) == 0 {
		return []models.ParameterTuningSuggestion{}, nil
	}
	return suggestions, rows.Err()
}

// PredictQueryPerformance predicts query execution time
func (p *PostgresDB) PredictQueryPerformance(ctx context.Context, queryHash int64, parameters map[string]interface{}, scenario string) (*models.PerformancePrediction, error) {
	// TODO: Implement ML model prediction logic
	// For now, return nil to trigger fallback behavior
	// In Phase 4.5.5, this will call the Python ML service

	return nil, nil
}

// UpdateOptimizationResults updates implementation with post-optimization metrics
func (p *PostgresDB) UpdateOptimizationResults(ctx context.Context, implementationID int64, postStats map[string]interface{}, actualImprovementPct, actualImprovementSec float64) error {
	postStatsJSON := "{}"
	if postStats != nil {
		if b, err := json.Marshal(postStats); err == nil {
			postStatsJSON = string(b)
		}
	}

	query := `
		UPDATE optimization_implementations
		SET post_optimization_stats = $1::jsonb,
		    actual_improvement_percent = $2,
		    actual_improvement_seconds = $3,
		    status = 'implemented',
		    measured_at = NOW()
		WHERE id = $4`

	result, err := p.db.ExecContext(ctx, query, postStatsJSON, actualImprovementPct, actualImprovementSec, implementationID)
	if err != nil {
		return apperrors.DatabaseError("update implementation results", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("check rows affected", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("implementation", fmt.Sprintf("%d", implementationID))
	}

	return nil
}

// GetOptimizationResults retrieves implementation results
func (p *PostgresDB) GetOptimizationResults(ctx context.Context, recommendationID *int64, status string, limit int) ([]models.OptimizationResult, error) {
	query := `
		SELECT i.id, r.id, r.query_hash, r.recommendation_text, r.estimated_improvement_percent,
		       i.actual_improvement_percent, i.status, i.implementation_timestamp, i.measured_at,
		       r.confidence_score, i.actual_improvement_seconds
		FROM optimization_implementations i
		JOIN optimization_recommendations r ON i.recommendation_id = r.id
		WHERE 1=1`

	args := []interface{}{}
	argNum := 1

	if recommendationID != nil {
		query += fmt.Sprintf(` AND r.id = $%d`, argNum)
		args = append(args, recommendationID)
		argNum++
	}

	if status != "" {
		query += fmt.Sprintf(` AND i.status = $%d`, argNum)
		args = append(args, status)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY i.implementation_timestamp DESC LIMIT $%d`, argNum)
	args = append(args, limit)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("get optimization results", err.Error())
	}
	defer rows.Close()

	var results []models.OptimizationResult
	for rows.Next() {
		var r models.OptimizationResult
		err := rows.Scan(
			&r.ImplementationID, &r.RecommendationID, &r.QueryHash, &r.RecommendationText,
			&r.EstimatedImprovement, &r.ActualImprovement, &r.Status, &r.ImplementationTime,
			&r.MeasuredAt, &r.ConfidenceScore, &r.ActualImprovementSec,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan result", err.Error())
		}

		// Calculate prediction error if both values available
		if r.ActualImprovement != nil && r.EstimatedImprovement != 0 {
			errorPct := (*r.ActualImprovement - r.EstimatedImprovement) / r.EstimatedImprovement * 100
			r.PredictionErrorPct = &errorPct
		}

		results = append(results, r)
	}

	return results, rows.Err()
}

// DismissOptimizationRecommendation marks a recommendation as dismissed
func (p *PostgresDB) DismissOptimizationRecommendation(ctx context.Context, recommendationID int64, reason string) error {
	query := `
		UPDATE optimization_recommendations
		SET is_dismissed = TRUE, dismissal_reason = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := p.db.ExecContext(ctx, query, reason, recommendationID)
	if err != nil {
		return apperrors.DatabaseError("dismiss recommendation", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return apperrors.DatabaseError("check rows affected", err.Error())
	}

	if rowsAffected == 0 {
		return apperrors.NotFound("recommendation", fmt.Sprintf("%d", recommendationID))
	}

	return nil
}

// GetRecommendationByID retrieves a specific recommendation
func (p *PostgresDB) GetRecommendationByID(ctx context.Context, recommendationID int64) (*models.OptimizationRecommendation, error) {
	query := `
		SELECT id, query_hash, source_type, source_id, recommendation_text, detailed_explanation,
		       estimated_improvement_percent, confidence_score, urgency_score, roi_score,
		       implementation_complexity, dismissal_reason, is_dismissed, created_at, updated_at
		FROM optimization_recommendations
		WHERE id = $1`

	var r models.OptimizationRecommendation
	err := p.db.QueryRowContext(ctx, query, recommendationID).Scan(
		&r.ID, &r.QueryHash, &r.SourceType, &r.SourceID, &r.RecommendationText,
		&r.DetailedExplanation, &r.EstimatedImprovementPct, &r.ConfidenceScore,
		&r.UrgencyScore, &r.ROIScore, &r.ImplementationComplexity, &r.DismissalReason,
		&r.IsDismissed, &r.CreatedAt, &r.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("recommendation", fmt.Sprintf("%d", recommendationID))
	}
	if err != nil {
		return nil, apperrors.DatabaseError("get recommendation", err.Error())
	}

	return &r, nil
}

// PHASE 4.5.4: ML-POWERED OPTIMIZATION WORKFLOW
// ============================================================================

// AggregateRecommendationsForQuery aggregates all suggestions into optimization_recommendations table
func (p *PostgresDB) AggregateRecommendationsForQuery(ctx context.Context, queryHash int64) (int, []string, error) {
	// Validate input
	if queryHash <= 0 {
		return 0, nil, apperrors.BadRequest("query_hash", "Must be a positive integer")
	}


	// Call SQL function to aggregate recommendations
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT recommendation_count, source_types FROM aggregate_recommendations_for_query($1)`,
		queryHash,
	)
	if err != nil {
		return 0, nil, apperrors.DatabaseError("aggregate recommendations", err.Error())
	}
	defer rows.Close()

	var count int
	var sourceTypes pq.StringArray

	if rows.Next() {
		err := rows.Scan(&count, &sourceTypes)
		if err != nil {
			return 0, nil, apperrors.DatabaseError("scan results", err.Error())
		}
	}

	return count, sourceTypes, nil
}

// GetOptimizationRecommendations retrieves top recommendations ranked by ROI
func (p *PostgresDB) GetOptimizationRecommendations(
	ctx context.Context,
	limit int,
	minImpact float64,
	sourceType string,
) ([]models.OptimizationRecommendation, error) {
	// Validate limits
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if minImpact < 0 {
		minImpact = 5.0
	}

	query := `
		SELECT id, query_hash, source_type, source_id, recommendation_text, detailed_explanation,
		       estimated_improvement_percent, confidence_score, urgency_score, roi_score,
		       implementation_complexity, dismissal_reason, is_dismissed, created_at, updated_at
		FROM optimization_recommendations
		WHERE is_dismissed = FALSE
		AND estimated_improvement_percent >= $1`

	args := []interface{}{minImpact}
	argNum := 2

	if sourceType != "" {
		query += fmt.Sprintf(` AND source_type = $%d`, argNum)
		args = append(args, sourceType)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY roi_score DESC, confidence_score DESC LIMIT $%d`, argNum)
	args = append(args, limit)


	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("get recommendations", err.Error())
	}
	defer rows.Close()

	var recommendations []models.OptimizationRecommendation
	for rows.Next() {
		var r models.OptimizationRecommendation
		err := rows.Scan(
			&r.ID, &r.QueryHash, &r.SourceType, &r.SourceID, &r.RecommendationText,
			&r.DetailedExplanation, &r.EstimatedImprovementPct, &r.ConfidenceScore,
			&r.UrgencyScore, &r.ROIScore, &r.ImplementationComplexity, &r.DismissalReason,
			&r.IsDismissed, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			continue
		}
		recommendations = append(recommendations, r)
	}

	if len(recommendations) == 0 {
		return []models.OptimizationRecommendation{}, nil
	}

	return recommendations, rows.Err()
}

// ImplementRecommendation records that a recommendation was implemented
func (p *PostgresDB) ImplementRecommendation(
	ctx context.Context,
	recommendationID int64,
	queryHash int64,
	notes string,
) (*models.OptimizationImplementation, error) {
	// Validate inputs
	if recommendationID <= 0 {
		return nil, apperrors.BadRequest("recommendation_id", "Must be a positive integer")
	}

	if queryHash <= 0 {
		return nil, apperrors.BadRequest("query_hash", "Must be a positive integer")
	}


	// Call SQL function to record implementation
	var implID int64
	var status string
	var preMetadata interface{}

	err := p.db.QueryRowContext(
		ctx,
		`SELECT impl_id, status, pre_snapshot FROM record_recommendation_implementation($1, $2, $3)`,
		recommendationID,
		queryHash,
		notes,
	).Scan(&implID, &status, &preMetadata)

	if err != nil {
		return nil, apperrors.DatabaseError("record implementation", err.Error())
	}


	t := time.Now().UTC()
	return &models.OptimizationImplementation{
		ID:                       implID,
		RecommendationID:         recommendationID,
		QueryHash:                queryHash,
		Status:                   status,
		ImplementationNotes:      &notes,
		ImplementationTimestamp:  t,
	}, nil
}

// MeasureImplementationResults measures actual improvement from implementation
func (p *PostgresDB) MeasureImplementationResults(
	ctx context.Context,
	implementationID int64,
) (*models.OptimizationResult, error) {
	// Validate input
	if implementationID <= 0 {
		return nil, apperrors.BadRequest("implementation_id", "Must be a positive integer")
	}


	// Call SQL function to measure results
	var implID int64
	var actualImprovement float64
	var predictedImprovement float64
	var finalStatus string
	var accuracyScore float64

	err := p.db.QueryRowContext(
		ctx,
		`SELECT impl_id, actual_improvement_percent, predicted_improvement_percent, status, accuracy_score
		 FROM measure_implementation_results($1)`,
		implementationID,
	).Scan(&implID, &actualImprovement, &predictedImprovement, &finalStatus, &accuracyScore)

	if err != nil && err != sql.ErrNoRows {
		return nil, apperrors.DatabaseError("measure results", err.Error())
	}

	if err == sql.ErrNoRows {
		return nil, apperrors.NotFound("implementation", fmt.Sprintf("%d", implementationID))
	}

	t := time.Now().UTC()
	return &models.OptimizationResult{
		ImplementationID:   implID,
		ActualImprovement:  &actualImprovement,
		PredictionErrorPct: &predictedImprovement,
		ConfidenceScore:    accuracyScore,
		Status:             finalStatus,
		MeasuredAt:         &t,
	}, nil
}

// TrainPerformanceModel trains a new ML model for performance prediction
func (p *PostgresDB) TrainPerformanceModel(ctx context.Context, databaseName string, lookbackDays int) error {
	// TODO: Implement model training logic
	// This will be called from Python ML service or Go backend
	return nil
}
