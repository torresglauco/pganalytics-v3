package query_performance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
