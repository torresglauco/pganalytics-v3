package query_performance

import (
	"database/sql"
	"encoding/json"
)

type QueryCollector struct {
	db *sql.DB
}

type ExplainPlan struct {
	NodeType  string  `json:"Node Type"`
	TotalCost float64 `json:"Total Cost"`
}

func NewQueryCollector(db *sql.DB) *QueryCollector {
	return &QueryCollector{db: db}
}

func (qc *QueryCollector) ParseExplainOutput(output string) (*ExplainPlan, error) {
	var plan struct {
		Plan ExplainPlan `json:"Plan"`
	}

	err := json.Unmarshal([]byte(output), &plan)
	if err != nil {
		return nil, err
	}

	return &plan.Plan, nil
}
