package storage

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/google/uuid"
)

// ============================================================================
// SCHEMA METRICS OPERATIONS
// ============================================================================

// StoreSchemaMetrics inserts schema metrics into the database
func (p *PostgresDB) StoreSchemaMetrics(ctx context.Context, tables []*models.SchemaTable, columns []*models.SchemaColumn, constraints []*models.SchemaConstraint, fkeys []*models.SchemaForeignKey) error {
	if len(tables) == 0 && len(columns) == 0 && len(constraints) == 0 && len(fkeys) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer tx.Rollback()

	// Insert tables
	if len(tables) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_schema_tables (time, collector_id, database_name, schema_name, table_name, table_type)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare schema tables insert", err.Error())
		}
		defer stmt.Close()

		for _, t := range tables {
			if _, err := stmt.ExecContext(ctx, time.Now(), t.CollectorID, t.DatabaseName, t.SchemaName, t.TableName, t.TableType); err != nil {
				return apperrors.DatabaseError("insert schema table", err.Error())
			}
		}
	}

	// Insert columns
	if len(columns) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_schema_columns (time, collector_id, database_name, schema_name, table_name, column_name, data_type, is_nullable, column_default, ordinal_position, character_max_length, numeric_precision, numeric_scale)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare schema columns insert", err.Error())
		}
		defer stmt.Close()

		for _, c := range columns {
			if _, err := stmt.ExecContext(ctx, time.Now(), c.CollectorID, c.DatabaseName, c.SchemaName, c.TableName, c.ColumnName, c.DataType, c.IsNullable, c.ColumnDefault, c.OrdinalPosition, c.CharacterMaxLength, c.NumericPrecision, c.NumericScale); err != nil {
				return apperrors.DatabaseError("insert schema column", err.Error())
			}
		}
	}

	// Insert constraints
	if len(constraints) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_schema_constraints (time, collector_id, database_name, schema_name, table_name, constraint_name, constraint_type, columns)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare constraints insert", err.Error())
		}
		defer stmt.Close()

		for _, c := range constraints {
			if _, err := stmt.ExecContext(ctx, time.Now(), c.CollectorID, c.DatabaseName, c.SchemaName, c.TableName, c.ConstraintName, c.ConstraintType, c.Columns); err != nil {
				return apperrors.DatabaseError("insert constraint", err.Error())
			}
		}
	}

	// Insert foreign keys
	if len(fkeys) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_schema_foreign_keys (time, collector_id, database_name, source_schema, source_table, source_column, target_schema, target_table, target_column, update_rule, delete_rule)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare foreign keys insert", err.Error())
		}
		defer stmt.Close()

		for _, fk := range fkeys {
			if _, err := stmt.ExecContext(ctx, time.Now(), fk.CollectorID, fk.DatabaseName, fk.SourceSchema, fk.SourceTable, fk.SourceColumn, fk.TargetSchema, fk.TargetTable, fk.TargetColumn, fk.UpdateRule, fk.DeleteRule); err != nil {
				return apperrors.DatabaseError("insert foreign key", err.Error())
			}
		}
	}

	return tx.Commit()
}

// GetSchemaMetrics retrieves schema metrics for a collector
func (p *PostgresDB) GetSchemaMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.SchemaMetricsResponse, error) {
	resp := &models.SchemaMetricsResponse{}

	// Get tables
	query := `SELECT database_name, schema_name, table_name, table_type FROM metrics_pg_schema_tables WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query schema tables", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		t := &models.SchemaTable{CollectorID: collectorID}
		if err := rows.Scan(&t.DatabaseName, &t.SchemaName, &t.TableName, &t.TableType); err != nil {
			return nil, apperrors.DatabaseError("scan schema table", err.Error())
		}
		resp.Tables = append(resp.Tables, t)
	}

	return resp, nil
}

// ============================================================================
// LOCK METRICS OPERATIONS
// ============================================================================

// StoreLockMetrics inserts lock metrics into the database
func (p *PostgresDB) StoreLockMetrics(ctx context.Context, locks []*models.Lock, waits []*models.LockWait) error {
	if len(locks) == 0 && len(waits) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer tx.Rollback()

	// Insert locks
	if len(locks) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_locks (time, collector_id, database_name, pid, locktype, mode, granted, relation_id, page_number, tuple_id, username, session_state, lock_age_seconds, query)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare locks insert", err.Error())
		}
		defer stmt.Close()

		for _, l := range locks {
			if _, err := stmt.ExecContext(ctx, time.Now(), l.CollectorID, l.DatabaseName, l.PID, l.LockType, l.Mode, l.Granted, l.RelationID, l.PageNumber, l.TupleID, l.Username, l.SessionState, l.LockAgeSeconds, l.Query); err != nil {
				return apperrors.DatabaseError("insert lock", err.Error())
			}
		}
	}

	// Insert lock waits
	if len(waits) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_lock_waits (time, collector_id, database_name, blocked_pid, blocking_pid, blocked_username, blocking_username, blocked_query, blocking_query, wait_time_seconds, blocked_application, blocking_application)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare lock waits insert", err.Error())
		}
		defer stmt.Close()

		for _, w := range waits {
			if _, err := stmt.ExecContext(ctx, time.Now(), w.CollectorID, w.DatabaseName, w.BlockedPID, w.BlockingPID, w.BlockedUsername, w.BlockingUsername, w.BlockedQuery, w.BlockingQuery, w.WaitTimeSeconds, w.BlockedApplication, w.BlockingApplication); err != nil {
				return apperrors.DatabaseError("insert lock wait", err.Error())
			}
		}
	}

	return tx.Commit()
}

// GetLockMetrics retrieves lock metrics for a collector
func (p *PostgresDB) GetLockMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.LockMetricsResponse, error) {
	resp := &models.LockMetricsResponse{}

	// Get active locks
	query := `SELECT pid, locktype, mode, granted, username, session_state, lock_age_seconds FROM metrics_pg_locks WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query locks", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		l := &models.Lock{CollectorID: collectorID}
		if err := rows.Scan(&l.PID, &l.LockType, &l.Mode, &l.Granted, &l.Username, &l.SessionState, &l.LockAgeSeconds); err != nil {
			return nil, apperrors.DatabaseError("scan lock", err.Error())
		}
		resp.ActiveLocks = append(resp.ActiveLocks, l)
	}

	return resp, nil
}

// ============================================================================
// BLOAT METRICS OPERATIONS
// ============================================================================

// StoreBloatMetrics inserts bloat metrics into the database
func (p *PostgresDB) StoreBloatMetrics(ctx context.Context, tableBloat []*models.TableBloat, indexBloat []*models.IndexBloat) error {
	if len(tableBloat) == 0 && len(indexBloat) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer tx.Rollback()

	// Insert table bloat
	if len(tableBloat) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_bloat_tables (time, collector_id, database_name, schema_name, table_name, dead_tuples, live_tuples, dead_ratio_percent, table_size, space_wasted_percent, last_vacuum, last_autovacuum, vacuum_count, autovacuum_count)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare table bloat insert", err.Error())
		}
		defer stmt.Close()

		for _, tb := range tableBloat {
			if _, err := stmt.ExecContext(ctx, time.Now(), tb.CollectorID, tb.DatabaseName, tb.SchemaName, tb.TableName, tb.DeadTuples, tb.LiveTuples, tb.DeadRatioPercent, tb.TableSize, tb.SpaceWastedPercent, tb.LastVacuum, tb.LastAutovacuum, tb.VacuumCount, tb.AutovacuumCount); err != nil {
				return apperrors.DatabaseError("insert table bloat", err.Error())
			}
		}
	}

	// Insert index bloat
	if len(indexBloat) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_bloat_indexes (time, collector_id, database_name, schema_name, table_name, index_name, index_scans, tuples_read, tuples_fetched, index_size, usage_status, recommendation)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare index bloat insert", err.Error())
		}
		defer stmt.Close()

		for _, ib := range indexBloat {
			if _, err := stmt.ExecContext(ctx, time.Now(), ib.CollectorID, ib.DatabaseName, ib.SchemaName, ib.TableName, ib.IndexName, ib.IndexScans, ib.TuplesRead, ib.TuplesFetched, ib.IndexSize, ib.UsageStatus, ib.Recommendation); err != nil {
				return apperrors.DatabaseError("insert index bloat", err.Error())
			}
		}
	}

	return tx.Commit()
}

// GetBloatMetrics retrieves bloat metrics for a collector
func (p *PostgresDB) GetBloatMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.BloatMetricsResponse, error) {
	resp := &models.BloatMetricsResponse{}

	query := `SELECT schema_name, table_name, dead_tuples, live_tuples, dead_ratio_percent, table_size FROM metrics_pg_bloat_tables WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query bloat", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		tb := &models.TableBloat{CollectorID: collectorID}
		if err := rows.Scan(&tb.SchemaName, &tb.TableName, &tb.DeadTuples, &tb.LiveTuples, &tb.DeadRatioPercent, &tb.TableSize); err != nil {
			return nil, apperrors.DatabaseError("scan bloat", err.Error())
		}
		resp.TableBloat = append(resp.TableBloat, tb)
	}

	return resp, nil
}

// ============================================================================
// CACHE METRICS OPERATIONS
// ============================================================================

// StoreCacheMetrics inserts cache metrics into the database
func (p *PostgresDB) StoreCacheMetrics(ctx context.Context, tableCacheHit []*models.TableCacheHit, indexCacheHit []*models.IndexCacheHit) error {
	if len(tableCacheHit) == 0 && len(indexCacheHit) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer tx.Rollback()

	// Insert table cache hit
	if len(tableCacheHit) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_cache_tables (time, collector_id, database_name, schema_name, table_name, heap_blks_hit, heap_blks_read, heap_cache_hit_ratio, idx_blks_hit, idx_blks_read, idx_cache_hit_ratio, toast_blks_hit, toast_blks_read, tidx_blks_hit, tidx_blks_read)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare table cache insert", err.Error())
		}
		defer stmt.Close()

		for _, tch := range tableCacheHit {
			if _, err := stmt.ExecContext(ctx, time.Now(), tch.CollectorID, tch.DatabaseName, tch.SchemaName, tch.TableName, tch.HeapBlksHit, tch.HeapBlksRead, tch.HeapCacheHitRatio, tch.IdxBlksHit, tch.IdxBlksRead, tch.IdxCacheHitRatio, tch.ToastBlksHit, tch.ToastBlksRead, tch.TidxBlksHit, tch.TidxBlksRead); err != nil {
				return apperrors.DatabaseError("insert table cache", err.Error())
			}
		}
	}

	// Insert index cache hit
	if len(indexCacheHit) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_cache_indexes (time, collector_id, database_name, schema_name, table_name, index_name, blks_hit, blks_read, cache_hit_ratio)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare index cache insert", err.Error())
		}
		defer stmt.Close()

		for _, ich := range indexCacheHit {
			if _, err := stmt.ExecContext(ctx, time.Now(), ich.CollectorID, ich.DatabaseName, ich.SchemaName, ich.TableName, ich.IndexName, ich.BlksHit, ich.BlksRead, ich.CacheHitRatio); err != nil {
				return apperrors.DatabaseError("insert index cache", err.Error())
			}
		}
	}

	return tx.Commit()
}

// GetCacheMetrics retrieves cache metrics for a collector
func (p *PostgresDB) GetCacheMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.CacheMetricsResponse, error) {
	resp := &models.CacheMetricsResponse{}

	query := `SELECT schema_name, table_name, heap_cache_hit_ratio, idx_cache_hit_ratio FROM metrics_pg_cache_tables WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query cache", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		tch := &models.TableCacheHit{CollectorID: collectorID}
		if err := rows.Scan(&tch.SchemaName, &tch.TableName, &tch.HeapCacheHitRatio, &tch.IdxCacheHitRatio); err != nil {
			return nil, apperrors.DatabaseError("scan cache", err.Error())
		}
		resp.TableCacheHit = append(resp.TableCacheHit, tch)
	}

	return resp, nil
}

// ============================================================================
// CONNECTION METRICS OPERATIONS
// ============================================================================

// StoreConnectionMetrics inserts connection metrics into the database
func (p *PostgresDB) StoreConnectionMetrics(ctx context.Context, connSummary []*models.ConnectionSummary, longRunning []*models.LongRunningTransaction, idle []*models.IdleTransaction) error {
	if len(connSummary) == 0 && len(longRunning) == 0 && len(idle) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer tx.Rollback()

	// Insert connection summary
	if len(connSummary) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_connections_summary (time, collector_id, database_name, connection_state, connection_count, max_age_seconds, min_age_seconds)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare connection summary insert", err.Error())
		}
		defer stmt.Close()

		for _, cs := range connSummary {
			if _, err := stmt.ExecContext(ctx, time.Now(), cs.CollectorID, cs.DatabaseName, cs.ConnectionState, cs.ConnectionCount, cs.MaxAgeSeconds, cs.MinAgeSeconds); err != nil {
				return apperrors.DatabaseError("insert connection summary", err.Error())
			}
		}
	}

	// Insert long-running transactions
	if len(longRunning) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_long_running_transactions (time, collector_id, database_name, pid, username, session_state, query, query_start, duration_seconds, application_name, client_address)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare long running insert", err.Error())
		}
		defer stmt.Close()

		for _, lr := range longRunning {
			if _, err := stmt.ExecContext(ctx, time.Now(), lr.CollectorID, lr.DatabaseName, lr.PID, lr.Username, lr.SessionState, lr.Query, lr.QueryStart, lr.DurationSeconds, lr.ApplicationName, lr.ClientAddress); err != nil {
				return apperrors.DatabaseError("insert long running", err.Error())
			}
		}
	}

	// Insert idle transactions
	if len(idle) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO metrics_pg_idle_transactions (time, collector_id, database_name, pid, username, query_start, state_change, idle_time_seconds, application_name, client_address)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT DO NOTHING
		`)
		if err != nil {
			return apperrors.DatabaseError("prepare idle insert", err.Error())
		}
		defer stmt.Close()

		for _, it := range idle {
			if _, err := stmt.ExecContext(ctx, time.Now(), it.CollectorID, it.DatabaseName, it.PID, it.Username, it.QueryStart, it.StateChange, it.IdleTimeSeconds, it.ApplicationName, it.ClientAddress); err != nil {
				return apperrors.DatabaseError("insert idle", err.Error())
			}
		}
	}

	return tx.Commit()
}

// GetConnectionMetrics retrieves connection metrics for a collector
func (p *PostgresDB) GetConnectionMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.ConnectionMetricsResponse, error) {
	resp := &models.ConnectionMetricsResponse{}

	query := `SELECT connection_state, connection_count FROM metrics_pg_connections_summary WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query connections", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		cs := &models.ConnectionSummary{CollectorID: collectorID}
		if err := rows.Scan(&cs.ConnectionState, &cs.ConnectionCount); err != nil {
			return nil, apperrors.DatabaseError("scan connection", err.Error())
		}
		resp.ConnectionSummary = append(resp.ConnectionSummary, cs)
	}

	return resp, nil
}

// ============================================================================
// EXTENSION METRICS OPERATIONS
// ============================================================================

// StoreExtensionMetrics inserts extension metrics into the database
func (p *PostgresDB) StoreExtensionMetrics(ctx context.Context, extensions []*models.Extension) error {
	if len(extensions) == 0 {
		return nil
	}

	stmt, err := p.db.PrepareContext(ctx, `
		INSERT INTO metrics_pg_extensions (time, collector_id, database_name, extension_name, extension_version, extension_owner, extension_schema, is_relocatable, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare extension insert", err.Error())
	}
	defer stmt.Close()

	for _, ext := range extensions {
		if _, err := stmt.ExecContext(ctx, time.Now(), ext.CollectorID, ext.DatabaseName, ext.ExtensionName, ext.ExtensionVersion, ext.ExtensionOwner, ext.ExtensionSchema, ext.IsRelocatable, ext.Description); err != nil {
			return apperrors.DatabaseError("insert extension", err.Error())
		}
	}

	return nil
}

// GetExtensionMetrics retrieves extension metrics for a collector
func (p *PostgresDB) GetExtensionMetrics(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) (*models.ExtensionMetricsResponse, error) {
	resp := &models.ExtensionMetricsResponse{}

	query := `SELECT extension_name, extension_version, extension_owner, extension_schema FROM metrics_pg_extensions WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query extensions", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		ext := &models.Extension{CollectorID: collectorID}
		if err := rows.Scan(&ext.ExtensionName, &ext.ExtensionVersion, &ext.ExtensionOwner, &ext.ExtensionSchema); err != nil {
			return nil, apperrors.DatabaseError("scan extension", err.Error())
		}
		resp.Extensions = append(resp.Extensions, ext)
	}

	return resp, nil
}
