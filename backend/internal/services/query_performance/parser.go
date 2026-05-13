package query_performance

import (
	"encoding/json"
)

// QueryParser detects performance issues in query plans
type QueryParser struct{}

// NewQueryParser creates a new QueryParser instance
func NewQueryParser() *QueryParser {
	return &QueryParser{}
}

// DetectIssues analyzes an explain plan and returns detected issues (backward compatible)
func (qp *QueryParser) DetectIssues(plan *ExplainPlan) []QueryIssue {
	var issues []QueryIssue

	// Detect sequential scans (expensive)
	if plan.NodeType == "Seq Scan" && plan.TotalCost > 100 {
		issues = append(issues, QueryIssue{
			Type:             "sequential_scan",
			Severity:         "medium",
			AffectedNode:     plan.NodeType,
			Description:      "Sequential scan detected on large table",
			Recommendation:   "Consider creating an index on the WHERE clause columns",
			EstimatedBenefit: plan.TotalCost * 0.3,
		})
	}

	// Detect nested loops
	if plan.NodeType == "Nested Loop" {
		issues = append(issues, QueryIssue{
			Type:             "nested_loop",
			Severity:         "high",
			AffectedNode:     plan.NodeType,
			Description:      "Nested loop detected, may be inefficient for large datasets",
			Recommendation:   "Consider hash join or additional indexes",
			EstimatedBenefit: plan.TotalCost * 0.4,
		})
	}

	return issues
}

// DetectIssuesFull recursively analyzes EXPLAIN JSON for anti-patterns
func (qp *QueryParser) DetectIssuesFull(explainJSON string) ([]QueryIssue, error) {
	var plans []FullExplainPlan
	if err := json.Unmarshal([]byte(explainJSON), &plans); err != nil {
		return nil, err
	}

	var issues []QueryIssue
	for _, plan := range plans {
		qp.walkPlan(plan.Plan, &issues)
	}
	return issues, nil
}

// walkPlan recursively walks the query plan tree and detects issues
func (qp *QueryParser) walkPlan(node *PlanNode, issues *[]QueryIssue) {
	if node == nil {
		return
	}

	// Detect Seq Scan on expensive operations
	if node.NodeType == "Seq Scan" && node.TotalCost > 100 {
		severity := "medium"
		if node.TotalCost > 1000 {
			severity = "high"
		}
		*issues = append(*issues, QueryIssue{
			Type:             "sequential_scan",
			Severity:         severity,
			AffectedNode:     node.RelationName,
			Description:      "Sequential scan on table with high cost",
			Recommendation:   "Consider creating an index on filtered columns",
			EstimatedBenefit: node.TotalCost * 0.3,
		})
	}

	// Detect Nested Loop with high row counts
	if node.NodeType == "Nested Loop" && node.ActualRows > 1000 {
		*issues = append(*issues, QueryIssue{
			Type:             "nested_loop_high_rows",
			Severity:         "high",
			AffectedNode:     "Nested Loop",
			Description:      "Nested loop iterating over many rows",
			Recommendation:   "Consider hash join or merge join via index optimization",
			EstimatedBenefit: float64(node.ActualRows) * 0.01,
		})
	}

	// Detect Sort on large datasets
	if node.NodeType == "Sort" && node.PlanRows > 10000 {
		*issues = append(*issues, QueryIssue{
			Type:             "large_sort",
			Severity:         "medium",
			AffectedNode:     "Sort",
			Description:      "Sorting large dataset without index",
			Recommendation:   "Consider creating index on ORDER BY columns",
			EstimatedBenefit: node.TotalCost * 0.2,
		})
	}

	// Recursively walk child plans
	for _, child := range node.Plans {
		qp.walkPlan(child, issues)
	}
}
