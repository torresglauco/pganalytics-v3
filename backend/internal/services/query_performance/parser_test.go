package query_performance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test existing DetectIssues still works
func TestQueryParser_DetectIssues_SequentialScan(t *testing.T) {
	parser := NewQueryParser()

	plan := &ExplainPlan{
		NodeType:  "Seq Scan",
		TotalCost: 1000.0,
	}

	issues := parser.DetectIssues(plan)
	assert.Greater(t, len(issues), 0)
	assert.Equal(t, "sequential_scan", issues[0].Type)
	assert.Equal(t, "medium", issues[0].Severity)
}

func TestQueryParser_DetectIssues_NestedLoop(t *testing.T) {
	parser := NewQueryParser()

	plan := &ExplainPlan{
		NodeType:  "Nested Loop",
		TotalCost: 500.0,
	}

	issues := parser.DetectIssues(plan)
	assert.Greater(t, len(issues), 0)
	assert.Equal(t, "nested_loop", issues[0].Type)
	assert.Equal(t, "high", issues[0].Severity)
}

func TestQueryParser_DetectIssues_NoIssues(t *testing.T) {
	parser := NewQueryParser()

	plan := &ExplainPlan{
		NodeType:  "Index Scan",
		TotalCost: 50.0,
	}

	issues := parser.DetectIssues(plan)
	assert.Equal(t, 0, len(issues))
}

func TestQueryParser_DetectIssues_LowCostSequentialScan(t *testing.T) {
	parser := NewQueryParser()

	plan := &ExplainPlan{
		NodeType:  "Seq Scan",
		TotalCost: 50.0,
	}

	issues := parser.DetectIssues(plan)
	assert.Equal(t, 0, len(issues))
}

// Test DetectIssuesFull for recursive plan analysis

func TestQueryParser_DetectIssuesFull_NestedSeqScan(t *testing.T) {
	parser := NewQueryParser()

	// EXPLAIN JSON with Seq Scan nested inside a Hash Join
	explainJSON := `[
		{
			"Plan": {
				"Node Type": "Hash Join",
				"Total Cost": 1500.0,
				"Plans": [
					{
						"Node Type": "Seq Scan",
						"Total Cost": 1200.0,
						"Relation Name": "orders",
						"Plans": []
					},
					{
						"Node Type": "Hash",
						"Total Cost": 300.0,
						"Plans": []
					}
				]
			}
		}
	]`

	issues, err := parser.DetectIssuesFull(explainJSON)
	assert.NoError(t, err)
	assert.Greater(t, len(issues), 0, "Should detect Seq Scan in nested plan")

	// Find the sequential scan issue
	var seqScanIssue *QueryIssue
	for i := range issues {
		if issues[i].Type == "sequential_scan" {
			seqScanIssue = &issues[i]
			break
		}
	}
	assert.NotNil(t, seqScanIssue, "Should find sequential_scan issue")
	if seqScanIssue != nil {
		assert.Equal(t, "orders", seqScanIssue.AffectedNode)
		assert.Equal(t, "high", seqScanIssue.Severity)
	}
}

func TestQueryParser_DetectIssuesFull_NestedLoopHighRows(t *testing.T) {
	parser := NewQueryParser()

	// EXPLAIN JSON with Nested Loop iterating many rows
	explainJSON := `[
		{
			"Plan": {
				"Node Type": "Nested Loop",
				"Total Cost": 5000.0,
				"Actual Rows": 5000,
				"Plans": [
					{
						"Node Type": "Seq Scan",
						"Total Cost": 100.0,
						"Relation Name": "users",
						"Plans": []
					}
				]
			}
		}
	]`

	issues, err := parser.DetectIssuesFull(explainJSON)
	assert.NoError(t, err)

	// Should detect both nested loop high rows and nested loop
	var nestedLoopHighRows *QueryIssue
	for i := range issues {
		if issues[i].Type == "nested_loop_high_rows" {
			nestedLoopHighRows = &issues[i]
			break
		}
	}
	assert.NotNil(t, nestedLoopHighRows, "Should find nested_loop_high_rows issue")
	if nestedLoopHighRows != nil {
		assert.Equal(t, "high", nestedLoopHighRows.Severity)
		assert.Contains(t, nestedLoopHighRows.Description, "Nested loop")
	}
}

func TestQueryParser_DetectIssuesFull_LargeSort(t *testing.T) {
	parser := NewQueryParser()

	// EXPLAIN JSON with Sort on large dataset
	explainJSON := `[
		{
			"Plan": {
				"Node Type": "Sort",
				"Total Cost": 3000.0,
				"Plan Rows": 50000,
				"Plans": [
					{
						"Node Type": "Seq Scan",
						"Total Cost": 500.0,
						"Relation Name": "products",
						"Plans": []
					}
				]
			}
		}
	]`

	issues, err := parser.DetectIssuesFull(explainJSON)
	assert.NoError(t, err)

	// Should detect both large sort and sequential scan
	assert.GreaterOrEqual(t, len(issues), 1, "Should detect at least large_sort issue")

	var largeSortIssue *QueryIssue
	for i := range issues {
		if issues[i].Type == "large_sort" {
			largeSortIssue = &issues[i]
			break
		}
	}
	assert.NotNil(t, largeSortIssue, "Should find large_sort issue")
	if largeSortIssue != nil {
		assert.Equal(t, "medium", largeSortIssue.Severity)
		assert.Contains(t, largeSortIssue.Recommendation, "index")
	}
}

func TestQueryParser_DetectIssuesFull_MultipleIssues(t *testing.T) {
	parser := NewQueryParser()

	// Complex EXPLAIN JSON with multiple issues at different levels
	explainJSON := `[
		{
			"Plan": {
				"Node Type": "Hash Join",
				"Total Cost": 5000.0,
				"Plans": [
					{
						"Node Type": "Seq Scan",
						"Total Cost": 1500.0,
						"Relation Name": "orders",
						"Plans": []
					},
					{
						"Node Type": "Sort",
						"Total Cost": 2000.0,
						"Plan Rows": 25000,
						"Plans": [
							{
								"Node Type": "Seq Scan",
								"Total Cost": 800.0,
								"Relation Name": "customers",
								"Plans": []
							}
						]
					}
				]
			}
		}
	]`

	issues, err := parser.DetectIssuesFull(explainJSON)
	assert.NoError(t, err)

	// Should detect at least 3 issues: 2 seq scans + 1 large sort
	assert.GreaterOrEqual(t, len(issues), 3, "Should detect multiple issues")

	// Count issue types
	issueTypes := make(map[string]int)
	for _, issue := range issues {
		issueTypes[issue.Type]++
	}

	assert.Equal(t, 2, issueTypes["sequential_scan"], "Should find 2 sequential scans")
	assert.Equal(t, 1, issueTypes["large_sort"], "Should find 1 large sort")
}

func TestQueryParser_DetectIssuesFull_CostThresholds(t *testing.T) {
	parser := NewQueryParser()

	t.Run("medium cost seq scan", func(t *testing.T) {
		explainJSON := `[
			{
				"Plan": {
					"Node Type": "Seq Scan",
					"Total Cost": 500.0,
					"Relation Name": "test_table",
					"Plans": []
				}
			}
		]`

		issues, err := parser.DetectIssuesFull(explainJSON)
		assert.NoError(t, err)
		assert.Greater(t, len(issues), 0)
		assert.Equal(t, "medium", issues[0].Severity)
	})

	t.Run("high cost seq scan", func(t *testing.T) {
		explainJSON := `[
			{
				"Plan": {
					"Node Type": "Seq Scan",
					"Total Cost": 2000.0,
					"Relation Name": "test_table",
					"Plans": []
				}
			}
		]`

		issues, err := parser.DetectIssuesFull(explainJSON)
		assert.NoError(t, err)
		assert.Greater(t, len(issues), 0)
		assert.Equal(t, "high", issues[0].Severity)
	})
}

func TestQueryParser_DetectIssuesFull_InvalidJSON(t *testing.T) {
	parser := NewQueryParser()

	_, err := parser.DetectIssuesFull("not valid json")
	assert.Error(t, err)
}

func TestQueryParser_DetectIssuesFull_EmptyPlans(t *testing.T) {
	parser := NewQueryParser()

	explainJSON := `[
		{
			"Plan": {
				"Node Type": "Index Scan",
				"Total Cost": 50.0,
				"Plans": []
			}
		}
	]`

	issues, err := parser.DetectIssuesFull(explainJSON)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(issues), "Index Scan should not produce issues")
}
