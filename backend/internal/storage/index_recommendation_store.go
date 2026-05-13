package storage

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

// IndexRecommendationStore handles storage for index recommendations
type IndexRecommendationStore struct {
	db *PostgresDB
}

// NewIndexRecommendationStore creates a new store instance
func NewIndexRecommendationStore(db *PostgresDB) *IndexRecommendationStore {
	return &IndexRecommendationStore{db: db}
}

// SaveIndexRecommendation saves a recommendation to the database
// Uses upsert to avoid duplicates
func (s *IndexRecommendationStore) SaveIndexRecommendation(
	ctx context.Context,
	databaseID int,
	tableName string,
	columns []string,
	estimatedBenefit float64,
	costImprovement float64,
) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO index_recommendations (database_id, table_name, column_names, estimated_benefit, weighted_cost_improvement)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (database_id, table_name, column_names)
		DO UPDATE SET
			estimated_benefit = $4,
			weighted_cost_improvement = $5,
			status = 'recommended',
			created_at = CURRENT_TIMESTAMP
		RETURNING id
	`, databaseID, tableName, pq.Array(columns), estimatedBenefit, costImprovement).Scan(&id)
	return id, err
}

// GetUnusedIndexes retrieves unused indexes for a database from the unused_indexes table
func (s *IndexRecommendationStore) GetUnusedIndexes(ctx context.Context, databaseID int, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT
			index_name,
			table_name,
			size_bytes,
			last_used,
			created_at
		FROM unused_indexes
		WHERE database_id = $1
		ORDER BY size_bytes DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, databaseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []map[string]interface{}
	for rows.Next() {
		var indexName, tableName string
		var sizeBytes int64
		var lastUsed sql.NullTime
		var createdAt sql.NullTime

		err := rows.Scan(&indexName, &tableName, &sizeBytes, &lastUsed, &createdAt)
		if err != nil {
			return nil, err
		}

		idx := map[string]interface{}{
			"index_name": indexName,
			"table_name": tableName,
			"size_bytes": sizeBytes,
		}
		if lastUsed.Valid {
			idx["last_used"] = lastUsed.Time
		}
		if createdAt.Valid {
			idx["created_at"] = createdAt.Time
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// StoreUnusedIndex stores a detected unused index
func (s *IndexRecommendationStore) StoreUnusedIndex(
	ctx context.Context,
	databaseID int,
	indexName string,
	tableName string,
	sizeBytes int64,
) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO unused_indexes (database_id, index_name, table_name, size_bytes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (database_id, index_name)
		DO UPDATE SET size_bytes = $4
	`, databaseID, indexName, tableName, sizeBytes)
	return err
}

// ClearUnusedIndexes removes all unused index records for a database
func (s *IndexRecommendationStore) ClearUnusedIndexes(ctx context.Context, databaseID int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM unused_indexes WHERE database_id = $1`, databaseID)
	return err
}

// GetDatabaseConnection retrieves a database connection for the given database ID
// Returns the monitored database connection, not the pganalytics database
func (s *IndexRecommendationStore) GetDatabaseConnection(ctx context.Context, databaseID int) (*sql.DB, error) {
	// Get database connection info from the databases table
	var connectionStr string
	err := s.db.QueryRowContext(ctx,
		`SELECT connection_string FROM databases WHERE id = $1`,
		databaseID,
	).Scan(&connectionStr)
	if err != nil {
		return nil, err
	}

	// Open connection to the monitored database
	return sql.Open("postgres", connectionStr)
}
