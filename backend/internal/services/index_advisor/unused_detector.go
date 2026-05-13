package index_advisor

import (
	"context"
	"database/sql"
	"strings"
)

// UnusedIndex represents an index that is not being used
type UnusedIndex struct {
	SchemaName string `json:"schema_name"`
	TableName  string `json:"table_name"`
	IndexName  string `json:"index_name"`
	SizeBytes  int64  `json:"size_bytes"`
	IdxScan    int64  `json:"idx_scan"`
	IsPrimary  bool   `json:"is_primary"`
	IsUnique   bool   `json:"is_unique"`
}

// UnusedIndexDetector finds indexes that are not being used
type UnusedIndexDetector struct {
	db *sql.DB
}

// NewUnusedIndexDetector creates a new detector instance
func NewUnusedIndexDetector(db *sql.DB) *UnusedIndexDetector {
	return &UnusedIndexDetector{db: db}
}

// FindUnused returns indexes with zero scans, excluding constraints
func (d *UnusedIndexDetector) FindUnused(ctx context.Context, limit int) ([]UnusedIndex, error) {
	// Apply default and bounds
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT
			schemaname,
			tablename,
			indexrelname as indexname,
			pg_relation_size(indexrelid) as size_bytes,
			COALESCE(idx_scan, 0) as idx_scan,
			contype = 'p' as is_primary,
			contype = 'u' as is_unique
		FROM pg_stat_user_indexes psui
		LEFT JOIN pg_constraint c ON c.conindid = psui.indexrelid
		WHERE idx_scan = 0
		  AND contype IS NULL  -- Exclude primary keys, unique, foreign keys
		ORDER BY pg_relation_size(indexrelid) DESC
		LIMIT $1
	`

	rows, err := d.db.QueryContext(ctx, query, limit)
	if err != nil {
		// Handle missing pg_stat_user_indexes gracefully
		if strings.Contains(err.Error(), "does not exist") {
			return []UnusedIndex{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var indexes []UnusedIndex
	for rows.Next() {
		var idx UnusedIndex
		err := rows.Scan(
			&idx.SchemaName,
			&idx.TableName,
			&idx.IndexName,
			&idx.SizeBytes,
			&idx.IdxScan,
			&idx.IsPrimary,
			&idx.IsUnique,
		)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}
