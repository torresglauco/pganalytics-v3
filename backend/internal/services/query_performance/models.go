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

// FullExplainPlan represents EXPLAIN (FORMAT JSON) output
type FullExplainPlan struct {
	Plan          *PlanNode `json:"Plan"`
	PlanningTime  float64   `json:"Planning Time"`
	ExecutionTime float64   `json:"Execution Time"`
}

// PlanNode represents a node in the query plan tree
type PlanNode struct {
	NodeType     string      `json:"Node Type"`
	TotalCost    float64     `json:"Total Cost"`
	StartupCost  float64     `json:"Startup Cost"`
	PlanRows     int64       `json:"Plan Rows"`
	PlanWidth    int         `json:"Plan Width"`
	ActualRows   int64       `json:"Actual Rows"`
	ActualLoops  int64       `json:"Actual Loops"`
	RelationName string      `json:"Relation Name"`
	IndexName    string      `json:"Index Name"`
	Filter       string      `json:"Filter"`
	HashCond     string      `json:"Hash Cond"`
	JoinType     string      `json:"Join Type"`
	Plans        []*PlanNode `json:"Plans"`
}
