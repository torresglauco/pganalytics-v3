package query_performance

// QueryIssue represents a detected issue in a query plan
type QueryIssue struct {
	Type             string
	Severity         string
	AffectedNode     string
	Description      string
	Recommendation   string
	EstimatedBenefit float64
}
