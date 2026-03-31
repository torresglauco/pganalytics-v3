package query_performance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryCollector_CaptureExplainAnalyze(t *testing.T) {
	collector := NewQueryCollector(nil)
	assert.NotNil(t, collector)
}

func TestQueryCollector_ParseExplainOutput(t *testing.T) {
	collector := NewQueryCollector(nil)

	explainOutput := `{"Plan": {"Node Type": "Seq Scan"}}`
	plan, err := collector.ParseExplainOutput(explainOutput)

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "Seq Scan", plan.NodeType)
}
