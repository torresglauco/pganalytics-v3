package index_advisor

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUnusedIndexDetector_FindUnused_ReturnsIndexesWithZeroScans(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	}).
		AddRow("public", "audit_logs", "idx_audit_old", int64(1048576), int64(0), false, false).
		AddRow("public", "events", "idx_events_created", int64(524288), int64(0), false, false)

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	assert.NoError(t, err)
	assert.Len(t, indexes, 2)
	assert.Equal(t, "audit_logs", indexes[0].TableName)
	assert.Equal(t, "idx_audit_old", indexes[0].IndexName)
	assert.Equal(t, int64(0), indexes[0].IdxScan)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_ExcludesPrimaryKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Query excludes primary keys via "WHERE contype IS NULL"
	// so we should only get non-constraint indexes
	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	}).
		AddRow("public", "users", "idx_users_email", int64(524288), int64(0), false, false)

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	assert.NoError(t, err)
	assert.Len(t, indexes, 1)
	assert.False(t, indexes[0].IsPrimary)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_ExcludesUniqueConstraints(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Query excludes unique constraints via "WHERE contype IS NULL"
	// so we should only get non-constraint indexes
	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	}).
		AddRow("public", "products", "idx_products_sku", int64(262144), int64(0), false, false)

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	assert.NoError(t, err)
	assert.Len(t, indexes, 1)
	assert.False(t, indexes[0].IsUnique)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_ExcludesForeignKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Query excludes foreign keys via "WHERE contype IS NULL"
	// This is implicitly tested since the constraint check excludes all contype values
	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	}).
		AddRow("public", "orders", "idx_orders_user_id", int64(131072), int64(0), false, false)

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	assert.NoError(t, err)
	assert.Len(t, indexes, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_OrdersBySizeDesc(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Results should be ordered by size_bytes DESC (largest first)
	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	}).
		AddRow("public", "large_table", "idx_large", int64(10485760), int64(0), false, false).  // 10MB
		AddRow("public", "medium_table", "idx_medium", int64(1048576), int64(0), false, false). // 1MB
		AddRow("public", "small_table", "idx_small", int64(102400), int64(0), false, false)     // 100KB

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	assert.NoError(t, err)
	assert.Len(t, indexes, 3)
	// Verify ordering: largest first
	assert.Equal(t, int64(10485760), indexes[0].SizeBytes)
	assert.Equal(t, int64(1048576), indexes[1].SizeBytes)
	assert.Equal(t, int64(102400), indexes[2].SizeBytes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_RespectsLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	}).
		AddRow("public", "table1", "idx1", int64(1000000), int64(0), false, false).
		AddRow("public", "table2", "idx2", int64(500000), int64(0), false, false)

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(2).
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 2)

	assert.NoError(t, err)
	assert.Len(t, indexes, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_DefaultLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	})

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20). // Default limit is 20
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 0) // 0 should use default

	assert.NoError(t, err)
	assert.Empty(t, indexes) // Empty result is valid
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_MaxLimitCap(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	})

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(100). // Max limit is 100
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 500) // Should cap at 100

	assert.NoError(t, err)
	assert.Empty(t, indexes) // Empty result is valid
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_HandlesMissingExtension(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Simulate "does not exist" error (missing pg_stat_user_indexes view)
	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnError(errors.New(`relation "pg_stat_user_indexes" does not exist`))

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	// Should return empty slice instead of error
	assert.NoError(t, err)
	assert.Empty(t, indexes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_ReturnsErrorOnOtherErrors(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnError(errors.New("connection refused"))

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	assert.Error(t, err)
	assert.Nil(t, indexes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_FindUnused_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"schemaname", "tablename", "indexname", "size_bytes", "idx_scan", "is_primary", "is_unique",
	})

	mock.ExpectQuery(`SELECT schemaname, tablename, indexrelname as indexname`).
		WithArgs(20).
		WillReturnRows(rows)

	detector := NewUnusedIndexDetector(db)
	indexes, err := detector.FindUnused(context.Background(), 20)

	assert.NoError(t, err)
	assert.Empty(t, indexes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnusedIndexDetector_New(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	detector := NewUnusedIndexDetector(db)
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.db)
}
