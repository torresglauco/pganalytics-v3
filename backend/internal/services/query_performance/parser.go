package query_performance

// QueryParser detects performance issues in query plans
type QueryParser struct{}

// NewQueryParser creates a new QueryParser instance
func NewQueryParser() *QueryParser {
	return &QueryParser{}
}

// DetectIssues analyzes an explain plan and returns detected issues
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
