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
