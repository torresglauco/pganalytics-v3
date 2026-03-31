package query_performance

// QueryAnalyzer analyzes query issues and calculates severity scores
type QueryAnalyzer struct{}

// NewQueryAnalyzer creates a new QueryAnalyzer instance
func NewQueryAnalyzer() *QueryAnalyzer {
	return &QueryAnalyzer{}
}

// CalculateSeverityScore calculates the average severity score for a list of issues
// Scoring: low=10, medium=40, high=70, critical=100
func (qa *QueryAnalyzer) CalculateSeverityScore(issues []QueryIssue) float64 {
	if len(issues) == 0 {
		return 0.0
	}

	severityMap := map[string]float64{
		"low":      10.0,
		"medium":   40.0,
		"high":     70.0,
		"critical": 100.0,
	}

	totalScore := 0.0
	for _, issue := range issues {
		totalScore += severityMap[issue.Severity]
	}

	return totalScore / float64(len(issues))
}
