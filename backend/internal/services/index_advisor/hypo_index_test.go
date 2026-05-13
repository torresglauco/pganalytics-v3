package index_advisor

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHypoIndexTester_CheckHypoPGAvailable_True(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(rows)

	tester := NewHypoIndexTester(db, zap.NewNop())
	result := tester.CheckHypoPGAvailable(context.Background())

	assert.True(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHypoIndexTester_CheckHypoPGAvailable_False(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(rows)

	tester := NewHypoIndexTester(db, zap.NewNop())
	result := tester.CheckHypoPGAvailable(context.Background())

	assert.False(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHypoIndexTester_EstimateImpact_ReturnsImprovement(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock hypopg availability check
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Mock baseline EXPLAIN cost
	baselinePlan := []map[string]interface{}{
		{"Plan": map[string]interface{}{"Total Cost": float64(1500.0)}},
	}
	baselineJSON, _ := json.Marshal(baselinePlan)
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(baselineJSON))

	// Mock hypopg_create_index
	mock.ExpectQuery(`SELECT hypopg_create_index`).
		WillReturnRows(sqlmock.NewRows([]string{"index_name"}).AddRow("btree_my_table_my_column_idx"))

	// Mock EXPLAIN with hypothetical index
	indexPlan := []map[string]interface{}{
		{"Plan": map[string]interface{}{"Total Cost": float64(150.0)}},
	}
	indexJSON, _ := json.Marshal(indexPlan)
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(indexJSON))

	// Mock hypopg_drop_index
	mock.ExpectExec(`SELECT hypopg_drop_index`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "SELECT * FROM my_table WHERE my_column = 1", "my_table", []string{"my_column"})

	assert.NoError(t, err)
	assert.NotNil(t, impact)
	assert.Equal(t, "my_table", impact.TableName)
	assert.Equal(t, []string{"my_column"}, impact.Columns)
	assert.Equal(t, 1500.0, impact.CostWithout)
	assert.Equal(t, 150.0, impact.CostWith)
	assert.InDelta(t, 90.0, impact.ImprovementPct, 0.01) // (1500-150)/1500 * 100 = 90%
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHypoIndexTester_EstimateImpact_ErrorWhenHypopgNotInstalled(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock hypopg not available
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "SELECT * FROM my_table", "my_table", []string{"col"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hypopg extension not installed")
	assert.Nil(t, impact)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHypoIndexTester_EstimateImpact_CalculatesCorrectImprovement(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock hypopg availability
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Baseline cost: 2000
	baselinePlan := []map[string]interface{}{
		{"Plan": map[string]interface{}{"Total Cost": float64(2000.0)}},
	}
	baselineJSON, _ := json.Marshal(baselinePlan)
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(baselineJSON))

	// Create hypothetical index
	mock.ExpectQuery(`SELECT hypopg_create_index`).
		WillReturnRows(sqlmock.NewRows([]string{"index_name"}).AddRow("btree_test_idx"))

	// Cost with index: 500
	indexPlan := []map[string]interface{}{
		{"Plan": map[string]interface{}{"Total Cost": float64(500.0)}},
	}
	indexJSON, _ := json.Marshal(indexPlan)
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(indexJSON))

	// Cleanup
	mock.ExpectExec(`SELECT hypopg_drop_index`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "SELECT * FROM test WHERE x = 1", "test", []string{"x"})

	assert.NoError(t, err)
	assert.Equal(t, 75.0, impact.ImprovementPct) // (2000-500)/2000 * 100 = 75%
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHypoIndexTester_EstimateImpact_EmptyColumns(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "SELECT * FROM test", "test", []string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "columns cannot be empty")
	assert.Nil(t, impact)
}

func TestHypoIndexTester_EstimateImpact_NilColumns(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "SELECT * FROM test", "test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "columns cannot be empty")
	assert.Nil(t, impact)
}

func TestHypoIndexTester_EstimateImpact_CleanupOnFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock hypopg availability
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Baseline cost query
	baselinePlan := []map[string]interface{}{
		{"Plan": map[string]interface{}{"Total Cost": float64(1000.0)}},
	}
	baselineJSON, _ := json.Marshal(baselinePlan)
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(baselineJSON))

	// Create hypothetical index
	mock.ExpectQuery(`SELECT hypopg_create_index`).
		WillReturnRows(sqlmock.NewRows([]string{"index_name"}).AddRow("btree_test_idx"))

	// EXPLAIN with index fails
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnError(errors.New("query failed"))

	// Cleanup should still happen
	mock.ExpectExec(`SELECT hypopg_drop_index`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "SELECT * FROM test WHERE x = 1", "test", []string{"x"})

	// Should return result with baseline cost as fallback
	assert.NoError(t, err)
	assert.NotNil(t, impact)
	assert.Equal(t, 1000.0, impact.CostWith) // Falls back to baseline
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHypoIndexTester_ExtractTotalCost_ValidJSON(t *testing.T) {
	tester := NewHypoIndexTester(nil, zap.NewNop())

	jsonData := []byte(`[{"Plan": {"Total Cost": 1234.56}}]`)
	cost, err := tester.extractTotalCost(jsonData)

	assert.NoError(t, err)
	assert.Equal(t, 1234.56, cost)
}

func TestHypoIndexTester_ExtractTotalCost_InvalidJSON(t *testing.T) {
	tester := NewHypoIndexTester(nil, zap.NewNop())

	jsonData := []byte(`invalid json`)
	cost, err := tester.extractTotalCost(jsonData)

	assert.Error(t, err)
	assert.Equal(t, 0.0, cost)
}

func TestHypoIndexTester_ExtractTotalCost_EmptyPlan(t *testing.T) {
	tester := NewHypoIndexTester(nil, zap.NewNop())

	jsonData := []byte(`[]`)
	cost, err := tester.extractTotalCost(jsonData)

	assert.Error(t, err)
	assert.Equal(t, 0.0, cost)
}

func TestHypoIndexTester_New(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger := zap.NewNop()
	tester := NewHypoIndexTester(db, logger)

	assert.NotNil(t, tester)
	assert.NotNil(t, tester.db)
}

func TestHypoIndexTester_EstimateImpact_BaselineQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock hypopg availability
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Baseline EXPLAIN fails
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnError(errors.New("syntax error"))

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "INVALID SQL", "test", []string{"x"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get baseline cost")
	assert.Nil(t, impact)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHypoIndexTester_EstimateImpact_ZeroBaselineCost(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock hypopg availability
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM pg_extension WHERE extname = 'hypopg'\)`).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Baseline cost: 0 (edge case)
	baselinePlan := []map[string]interface{}{
		{"Plan": map[string]interface{}{"Total Cost": float64(0)}},
	}
	baselineJSON, _ := json.Marshal(baselinePlan)
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(baselineJSON))

	// Create hypothetical index
	mock.ExpectQuery(`SELECT hypopg_create_index`).
		WillReturnRows(sqlmock.NewRows([]string{"index_name"}).AddRow("btree_test_idx"))

	// Cost with index
	indexPlan := []map[string]interface{}{
		{"Plan": map[string]interface{}{"Total Cost": float64(0)}},
	}
	indexJSON, _ := json.Marshal(indexPlan)
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow(indexJSON))

	// Cleanup
	mock.ExpectExec(`SELECT hypopg_drop_index`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tester := NewHypoIndexTester(db, zap.NewNop())
	impact, err := tester.EstimateImpact(context.Background(), "SELECT 1", "test", []string{"x"})

	assert.NoError(t, err)
	assert.Equal(t, 0.0, impact.ImprovementPct) // 0 baseline means 0% improvement
	assert.NoError(t, mock.ExpectationsWereMet())
}
