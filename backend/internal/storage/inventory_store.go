package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// TABLE INVENTORY OPERATIONS (INV-01)
// ============================================================================

// StoreTableInventory inserts table inventory data into the database
func (p *PostgresDB) StoreTableInventory(ctx context.Context, tables []*models.TableInventory) error {
	if len(tables) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_table_inventory (
			time, collector_id, database_name, schema_name, table_name, table_type,
			row_count, total_size_mb, table_size_mb, index_size_mb, toast_size_mb,
			has_oids, table_oid
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare table inventory insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, t := range tables {
		_, err := stmt.ExecContext(ctx,
			t.Time, t.CollectorID, t.DatabaseName, t.SchemaName, t.TableName, t.TableType,
			t.RowCount, t.TotalSizeMb, t.TableSizeMb, t.IndexSizeMb, t.ToastSizeMb,
			t.HasOids, t.TableOid,
		)
		if err != nil {
			return apperrors.DatabaseError("insert table inventory", err.Error())
		}
	}

	return tx.Commit()
}

// GetTableInventory retrieves table inventory for a collector
func (p *PostgresDB) GetTableInventory(ctx context.Context, collectorID uuid.UUID, database *string, schema *string, limit int, offset int) ([]*models.TableInventory, error) {
	query := `
		SELECT time, collector_id, database_name, schema_name, table_name, table_type,
			row_count, total_size_mb, table_size_mb, index_size_mb, toast_size_mb,
			has_oids, table_oid
		FROM metrics_table_inventory
		WHERE collector_id = $1
	`
	args := []interface{}{collectorID}
	argNum := 2

	if database != nil {
		query += fmt.Sprintf(" AND database_name = $%d", argNum)
		args = append(args, *database)
		argNum++
	}

	if schema != nil {
		query += fmt.Sprintf(" AND schema_name = $%d", argNum)
		args = append(args, *schema)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY total_size_mb DESC, schema_name, table_name LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query table inventory", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var tables []*models.TableInventory
	for rows.Next() {
		t := &models.TableInventory{}
		err := rows.Scan(
			&t.Time, &t.CollectorID, &t.DatabaseName, &t.SchemaName, &t.TableName, &t.TableType,
			&t.RowCount, &t.TotalSizeMb, &t.TableSizeMb, &t.IndexSizeMb, &t.ToastSizeMb,
			&t.HasOids, &t.TableOid,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan table inventory", err.Error())
		}
		tables = append(tables, t)
	}

	return tables, nil
}

// ============================================================================
// COLUMN INVENTORY OPERATIONS (INV-02)
// ============================================================================

// StoreColumnInventory inserts column inventory data into the database
func (p *PostgresDB) StoreColumnInventory(ctx context.Context, columns []*models.ColumnInventory) error {
	if len(columns) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_column_inventory (
			time, collector_id, database_name, schema_name, table_name, column_name,
			data_type, is_nullable, column_default, ordinal_position,
			character_max_length, numeric_precision, numeric_scale,
			is_primary_key, is_foreign_key
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare column inventory insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, c := range columns {
		_, err := stmt.ExecContext(ctx,
			c.Time, c.CollectorID, c.DatabaseName, c.SchemaName, c.TableName, c.ColumnName,
			c.DataType, c.IsNullable, c.ColumnDefault, c.OrdinalPosition,
			c.CharacterMaxLength, c.NumericPrecision, c.NumericScale,
			c.IsPrimaryKey, c.IsForeignKey,
		)
		if err != nil {
			return apperrors.DatabaseError("insert column inventory", err.Error())
		}
	}

	return tx.Commit()
}

// GetColumnInventory retrieves column inventory for a collector
func (p *PostgresDB) GetColumnInventory(ctx context.Context, collectorID uuid.UUID, database *string, table *string, limit int, offset int) ([]*models.ColumnInventory, error) {
	query := `
		SELECT time, collector_id, database_name, schema_name, table_name, column_name,
			data_type, is_nullable, column_default, ordinal_position,
			character_max_length, numeric_precision, numeric_scale,
			is_primary_key, is_foreign_key
		FROM metrics_column_inventory
		WHERE collector_id = $1
	`
	args := []interface{}{collectorID}
	argNum := 2

	if database != nil {
		query += fmt.Sprintf(" AND database_name = $%d", argNum)
		args = append(args, *database)
		argNum++
	}

	if table != nil {
		query += fmt.Sprintf(" AND table_name = $%d", argNum)
		args = append(args, *table)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY schema_name, table_name, ordinal_position LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query column inventory", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var columns []*models.ColumnInventory
	for rows.Next() {
		c := &models.ColumnInventory{}
		err := rows.Scan(
			&c.Time, &c.CollectorID, &c.DatabaseName, &c.SchemaName, &c.TableName, &c.ColumnName,
			&c.DataType, &c.IsNullable, &c.ColumnDefault, &c.OrdinalPosition,
			&c.CharacterMaxLength, &c.NumericPrecision, &c.NumericScale,
			&c.IsPrimaryKey, &c.IsForeignKey,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan column inventory", err.Error())
		}
		columns = append(columns, c)
	}

	return columns, nil
}

// ============================================================================
// INDEX INVENTORY OPERATIONS (INV-03)
// ============================================================================

// StoreIndexInventory inserts index inventory data into the database
func (p *PostgresDB) StoreIndexInventory(ctx context.Context, indexes []*models.IndexInventory) error {
	if len(indexes) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_index_inventory (
			time, collector_id, database_name, schema_name, table_name, index_name,
			index_definition, index_size_mb, idx_scan, idx_tup_read, idx_tup_fetch,
			usage_status, is_primary, is_unique, index_oid
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare index inventory insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, idx := range indexes {
		_, err := stmt.ExecContext(ctx,
			idx.Time, idx.CollectorID, idx.DatabaseName, idx.SchemaName, idx.TableName, idx.IndexName,
			idx.IndexDefinition, idx.IndexSizeMb, idx.IdxScan, idx.IdxTupRead, idx.IdxTupFetch,
			idx.UsageStatus, idx.IsPrimary, idx.IsUnique, idx.IndexOid,
		)
		if err != nil {
			return apperrors.DatabaseError("insert index inventory", err.Error())
		}
	}

	return tx.Commit()
}

// GetIndexInventory retrieves index inventory for a collector
func (p *PostgresDB) GetIndexInventory(ctx context.Context, collectorID uuid.UUID, database *string, table *string, limit int, offset int) ([]*models.IndexInventory, error) {
	query := `
		SELECT time, collector_id, database_name, schema_name, table_name, index_name,
			index_definition, index_size_mb, idx_scan, idx_tup_read, idx_tup_fetch,
			usage_status, is_primary, is_unique, index_oid
		FROM metrics_index_inventory
		WHERE collector_id = $1
	`
	args := []interface{}{collectorID}
	argNum := 2

	if database != nil {
		query += fmt.Sprintf(" AND database_name = $%d", argNum)
		args = append(args, *database)
		argNum++
	}

	if table != nil {
		query += fmt.Sprintf(" AND table_name = $%d", argNum)
		args = append(args, *table)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY idx_scan DESC, schema_name, table_name, index_name LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query index inventory", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var indexes []*models.IndexInventory
	for rows.Next() {
		idx := &models.IndexInventory{}
		err := rows.Scan(
			&idx.Time, &idx.CollectorID, &idx.DatabaseName, &idx.SchemaName, &idx.TableName, &idx.IndexName,
			&idx.IndexDefinition, &idx.IndexSizeMb, &idx.IdxScan, &idx.IdxTupRead, &idx.IdxTupFetch,
			&idx.UsageStatus, &idx.IsPrimary, &idx.IsUnique, &idx.IndexOid,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan index inventory", err.Error())
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// ============================================================================
// EXTENSION INVENTORY OPERATIONS (INV-04)
// ============================================================================

// StoreExtensionInventory inserts extension inventory data into the database
func (p *PostgresDB) StoreExtensionInventory(ctx context.Context, extensions []*models.ExtensionInventory) error {
	if len(extensions) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_extension_inventory (
			time, collector_id, database_name, extension_name, extension_version,
			extension_owner, extension_schema, is_relocatable, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare extension inventory insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, ext := range extensions {
		_, err := stmt.ExecContext(ctx,
			ext.Time, ext.CollectorID, ext.DatabaseName, ext.ExtensionName, ext.ExtensionVersion,
			ext.ExtensionOwner, ext.ExtensionSchema, ext.IsRelocatable, ext.Description,
		)
		if err != nil {
			return apperrors.DatabaseError("insert extension inventory", err.Error())
		}
	}

	return tx.Commit()
}

// GetExtensionInventory retrieves extension inventory for a collector
func (p *PostgresDB) GetExtensionInventory(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) ([]*models.ExtensionInventory, error) {
	query := `
		SELECT time, collector_id, database_name, extension_name, extension_version,
			extension_owner, extension_schema, is_relocatable, description
		FROM metrics_extension_inventory
		WHERE collector_id = $1
	`
	args := []interface{}{collectorID}
	argNum := 2

	if database != nil {
		query += fmt.Sprintf(" AND database_name = $%d", argNum)
		args = append(args, *database)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY extension_name LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query extension inventory", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var extensions []*models.ExtensionInventory
	for rows.Next() {
		ext := &models.ExtensionInventory{}
		err := rows.Scan(
			&ext.Time, &ext.CollectorID, &ext.DatabaseName, &ext.ExtensionName, &ext.ExtensionVersion,
			&ext.ExtensionOwner, &ext.ExtensionSchema, &ext.IsRelocatable, &ext.Description,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan extension inventory", err.Error())
		}
		extensions = append(extensions, ext)
	}

	return extensions, nil
}

// ============================================================================
// SCHEMA VERSION OPERATIONS (INV-05)
// ============================================================================

// StoreSchemaVersion inserts schema version change data into the database
func (p *PostgresDB) StoreSchemaVersion(ctx context.Context, version *models.SchemaVersion) error {
	query := `
		INSERT INTO metrics_schema_versions (
			time, collector_id, database_name, version_hash, change_type,
			object_type, object_name, change_details, previous_value, new_value
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT DO NOTHING
	`

	_, err := p.db.ExecContext(ctx, query,
		version.Time, version.CollectorID, version.DatabaseName, version.VersionHash, version.ChangeType,
		version.ObjectType, version.ObjectName, version.ChangeDetails, version.PreviousValue, version.NewValue,
	)
	if err != nil {
		return apperrors.DatabaseError("insert schema version", err.Error())
	}

	return nil
}

// GetSchemaVersions retrieves schema version changes for a collector
func (p *PostgresDB) GetSchemaVersions(ctx context.Context, collectorID uuid.UUID, database *string, limit int) ([]*models.SchemaVersion, error) {
	query := `
		SELECT time, collector_id, database_name, version_hash, change_type,
			object_type, object_name, change_details, previous_value, new_value
		FROM metrics_schema_versions
		WHERE collector_id = $1
	`
	args := []interface{}{collectorID}
	argNum := 2

	if database != nil {
		query += fmt.Sprintf(" AND database_name = $%d", argNum)
		args = append(args, *database)
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY time DESC LIMIT $%d", argNum)
	args = append(args, limit)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query schema versions", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var versions []*models.SchemaVersion
	for rows.Next() {
		v := &models.SchemaVersion{}
		err := rows.Scan(
			&v.Time, &v.CollectorID, &v.DatabaseName, &v.VersionHash, &v.ChangeType,
			&v.ObjectType, &v.ObjectName, &v.ChangeDetails, &v.PreviousValue, &v.NewValue,
		)
		if err != nil {
			return nil, apperrors.DatabaseError("scan schema version", err.Error())
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// CalculateSchemaHash calculates an MD5 hash of all object signatures for change detection
func (p *PostgresDB) CalculateSchemaHash(ctx context.Context, collectorID uuid.UUID, database string) (string, error) {
	// Build a comprehensive hash from all inventory objects
	var combined string

	// Get tables
	tableQuery := `
		SELECT table_oid, schema_name, table_name, table_type, row_count, total_size_mb
		FROM metrics_table_inventory
		WHERE collector_id = $1 AND database_name = $2
		ORDER BY schema_name, table_name
	`
	rows, err := p.db.QueryContext(ctx, tableQuery, collectorID, database)
	if err != nil {
		return "", apperrors.DatabaseError("query tables for hash", err.Error())
	}
	for rows.Next() {
		var oid uint64
		var schema, name, ttype string
		var rowCount, sizeMb int64
		if err := rows.Scan(&oid, &schema, &name, &ttype, &rowCount, &sizeMb); err == nil {
			combined += fmt.Sprintf("T:%d:%s:%s:%s:%d:%d|", oid, schema, name, ttype, rowCount, sizeMb)
		}
	}
	_ = rows.Close()

	// Get columns
	colQuery := `
		SELECT schema_name, table_name, column_name, data_type, is_nullable, ordinal_position
		FROM metrics_column_inventory
		WHERE collector_id = $1 AND database_name = $2
		ORDER BY schema_name, table_name, ordinal_position
	`
	rows, err = p.db.QueryContext(ctx, colQuery, collectorID, database)
	if err != nil {
		return "", apperrors.DatabaseError("query columns for hash", err.Error())
	}
	for rows.Next() {
		var schema, table, col, dtype string
		var nullable bool
		var pos int
		if err := rows.Scan(&schema, &table, &col, &dtype, &nullable, &pos); err == nil {
			combined += fmt.Sprintf("C:%s:%s:%s:%s:%v:%d|", schema, table, col, dtype, nullable, pos)
		}
	}
	_ = rows.Close()

	// Get indexes
	idxQuery := `
		SELECT index_oid, schema_name, table_name, index_name, index_definition
		FROM metrics_index_inventory
		WHERE collector_id = $1 AND database_name = $2
		ORDER BY schema_name, table_name, index_name
	`
	rows, err = p.db.QueryContext(ctx, idxQuery, collectorID, database)
	if err != nil {
		return "", apperrors.DatabaseError("query indexes for hash", err.Error())
	}
	for rows.Next() {
		var oid uint64
		var schema, table, name, def string
		if err := rows.Scan(&oid, &schema, &table, &name, &def); err == nil {
			combined += fmt.Sprintf("I:%d:%s:%s:%s:%s|", oid, schema, table, name, def)
		}
	}
	_ = rows.Close()

	// Get extensions
	extQuery := `
		SELECT extension_name, extension_version
		FROM metrics_extension_inventory
		WHERE collector_id = $1 AND database_name = $2
		ORDER BY extension_name
	`
	rows, err = p.db.QueryContext(ctx, extQuery, collectorID, database)
	if err != nil {
		return "", apperrors.DatabaseError("query extensions for hash", err.Error())
	}
	for rows.Next() {
		var name, version string
		if err := rows.Scan(&name, &version); err == nil {
			combined += fmt.Sprintf("E:%s:%s|", name, version)
		}
	}
	_ = rows.Close()

	// Calculate MD5 hash
	hash := md5.Sum([]byte(combined))
	return hex.EncodeToString(hash[:]), nil
}

// TrackSchemaChanges compares current snapshot with previous and records changes
func (p *PostgresDB) TrackSchemaChanges(ctx context.Context, collectorID uuid.UUID, database string) ([]*models.SchemaVersion, error) {
	// Get current hash
	currentHash, err := p.CalculateSchemaHash(ctx, collectorID, database)
	if err != nil {
		return nil, err
	}

	// Get previous hash
	var previousHash string
	err = p.db.QueryRowContext(ctx, `
		SELECT version_hash FROM metrics_schema_versions
		WHERE collector_id = $1 AND database_name = $2
		ORDER BY time DESC LIMIT 1
	`, collectorID, database).Scan(&previousHash)

	if err != nil {
		// No previous version - this is an initial snapshot
		version := &models.SchemaVersion{
			Time:          time.Now(),
			CollectorID:   collectorID,
			DatabaseName:  database,
			VersionHash:   currentHash,
			ChangeType:    "INIT",
			ObjectType:    "SCHEMA",
			ObjectName:    database,
			ChangeDetails: "Initial schema snapshot",
		}
		if err := p.StoreSchemaVersion(ctx, version); err != nil {
			return nil, err
		}
		return []*models.SchemaVersion{version}, nil
	}

	// Compare hashes
	if currentHash == previousHash {
		return nil, nil // No changes
	}

	// Detect and record changes (simplified - full implementation would compare each object)
	changes := []*models.SchemaVersion{}

	// For now, record a generic schema change
	// TODO: Implement detailed change detection comparing individual objects
	change := &models.SchemaVersion{
		Time:          time.Now(),
		CollectorID:   collectorID,
		DatabaseName:  database,
		VersionHash:   currentHash,
		ChangeType:    "SCHEMA_MODIFIED",
		ObjectType:    "SCHEMA",
		ObjectName:    database,
		ChangeDetails: "Schema hash changed",
		PreviousValue: previousHash,
		NewValue:      currentHash,
	}
	if err := p.StoreSchemaVersion(ctx, change); err != nil {
		return nil, err
	}
	changes = append(changes, change)

	return changes, nil
}
