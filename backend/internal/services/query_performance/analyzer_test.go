package query_performance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryAnalyzer_CalculateScore_MultipleIssues(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	issues := []QueryIssue{
		{Severity: "high"},
		{Severity: "medium"},
	}

	score := analyzer.CalculateSeverityScore(issues)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 100.0)
	// Expected: (70 + 40) / 2 = 55
	assert.Equal(t, 55.0, score)
}

func TestQueryAnalyzer_CalculateScore_NoIssues(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	issues := []QueryIssue{}

	score := analyzer.CalculateSeverityScore(issues)
	assert.Equal(t, 0.0, score)
}

func TestQueryAnalyzer_CalculateScore_SingleLowIssue(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	issues := []QueryIssue{
		{Severity: "low"},
	}

	score := analyzer.CalculateSeverityScore(issues)
	assert.Equal(t, 10.0, score)
}

func TestQueryAnalyzer_CalculateScore_CriticalIssue(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	issues := []QueryIssue{
		{Severity: "critical"},
	}

	score := analyzer.CalculateSeverityScore(issues)
	assert.Equal(t, 100.0, score)
}

func TestQueryAnalyzer_CalculateScore_MixedSeverities(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	issues := []QueryIssue{
		{Severity: "high"},
		{Severity: "high"},
		{Severity: "low"},
	}

	score := analyzer.CalculateSeverityScore(issues)
	// Expected: (70 + 70 + 10) / 3 = 50
	assert.Equal(t, 50.0, score)
}
