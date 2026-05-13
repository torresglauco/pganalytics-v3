package index_advisor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// IndexImpact represents the estimated impact of creating an index
type IndexImpact struct {
	IndexName      string   `json:"index_name,omitempty"`
	TableName      string   `json:"table_name"`
	Columns        []string `json:"columns"`
	CostWithout    float64  `json:"cost_without"`
	CostWith       float64  `json:"cost_with"`
	ImprovementPct float64  `json:"improvement_pct"`
	QueryCount     int      `json:"query_count,omitempty"`
}

// HypoIndexTester tests hypothetical indexes using hypopg
type HypoIndexTester struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewHypoIndexTester creates a new tester instance
func NewHypoIndexTester(db *sql.DB, logger *zap.Logger) *HypoIndexTester {
	return &HypoIndexTester{db: db, logger: logger}
}

// CheckHypoPGAvailable checks if hypopg extension is installed
func (t *HypoIndexTester) CheckHypoPGAvailable(ctx context.Context) bool {
	var exists bool
	err := t.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'hypopg')",
	).Scan(&exists)
	return err == nil && exists
}

// EstimateImpact creates a hypothetical index and measures improvement
func (t *HypoIndexTester) EstimateImpact(
	ctx context.Context,
	queryText string,
	tableName string,
	columns []string,
) (*IndexImpact, error) {
	if len(columns) == 0 {
		return nil, fmt.Errorf("columns cannot be empty")
	}

	// Check hypopg availability
	if !t.CheckHypoPGAvailable(ctx) {
		return nil, fmt.Errorf("hypopg extension not installed")
	}

	// Get baseline cost
	baselineCost, err := t.getQueryCost(ctx, queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline cost: %w", err)
	}

	// Create hypothetical index
	colList := strings.Join(columns, ", ")
	hypoQuery := fmt.Sprintf(
		"SELECT hypopg_create_index('CREATE INDEX ON %s (%s)')",
		tableName, colList,
	)

	var indexName string
	err = t.db.QueryRowContext(ctx, hypoQuery).Scan(&indexName)
	if err != nil {
		return nil, fmt.Errorf("failed to create hypothetical index: %w", err)
	}

	// Ensure cleanup
	defer func() {
		_, _ = t.db.ExecContext(ctx, "SELECT hypopg_drop_index($1)", indexName)
	}()

	// Get cost with hypothetical index
	indexCost, err := t.getQueryCost(ctx, queryText)
	if err != nil {
		if t.logger != nil {
			t.logger.Warn("Failed to get index cost", zap.Error(err))
		}
		indexCost = baselineCost // Fallback
	}

	// Calculate improvement
	var improvement float64
	if baselineCost > 0 {
		improvement = ((baselineCost - indexCost) / baselineCost) * 100
	}

	return &IndexImpact{
		TableName:      tableName,
		Columns:        columns,
		CostWithout:    baselineCost,
		CostWith:       indexCost,
		ImprovementPct: improvement,
	}, nil
}

// getQueryCost extracts total cost from EXPLAIN output
func (t *HypoIndexTester) getQueryCost(ctx context.Context, queryText string) (float64, error) {
	explainQuery := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s", queryText)

	var result []byte
	err := t.db.QueryRowContext(ctx, explainQuery).Scan(&result)
	if err != nil {
		return 0, err
	}

	return t.extractTotalCost(result)
}

// extractTotalCost parses EXPLAIN JSON for total cost
func (t *HypoIndexTester) extractTotalCost(explainJSON []byte) (float64, error) {
	var plans []struct {
		Plan struct {
			TotalCost float64 `json:"Total Cost"`
		} `json:"Plan"`
	}

	if err := json.Unmarshal(explainJSON, &plans); err != nil {
		return 0, err
	}

	if len(plans) == 0 {
		return 0, fmt.Errorf("no plan found in EXPLAIN output")
	}

	return plans[0].Plan.TotalCost, nil
}
