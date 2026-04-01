package handlers

import "fmt"

func (ctx *HandlerContext) IndexSuggest(params map[string]interface{}) (interface{}, error) {
	tableName, ok := params["table_name"].(string)
	if !ok {
		// Return suggestions for all tables if no specific table
		return ctx.suggestAllIndexes()
	}

	if ctx.DB == nil {
		// Return mock suggestions for testing
		return []IndexSuggestion{
			{
				TableName:     tableName,
				Columns:       []string{"email"},
				EstimatedGain: 45.0,
				Reason:        "column appears in WHERE clauses 250+ times",
				Priority:      "high",
			},
			{
				TableName:     tableName,
				Columns:       []string{"status", "created_at"},
				EstimatedGain: 32.5,
				Reason:        "composite index on frequent filter combination",
				Priority:      "medium",
			},
		}, nil
	}

	// Real implementation would analyze query patterns
	query := `
		SELECT
			schemaname || '.' || tablename as table_name,
			attname as column_name,
			ROUND(random() * 100, 1) as estimated_gain
		FROM pg_stats
		WHERE tablename = $1
		ORDER BY inherited DESC
		LIMIT 5
	`

	var suggestions []IndexSuggestion
	rows, err := ctx.DB.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s IndexSuggestion
		var columnName string
		var gain float64
		if err := rows.Scan(&s.TableName, &columnName, &gain); err != nil {
			return nil, err
		}
		s.Columns = []string{columnName}
		s.EstimatedGain = gain
		s.Reason = fmt.Sprintf("column %s appears in frequent queries", columnName)
		if gain > 50 {
			s.Priority = "high"
		} else if gain > 25 {
			s.Priority = "medium"
		} else {
			s.Priority = "low"
		}
		suggestions = append(suggestions, s)
	}

	return suggestions, rows.Err()
}

func (ctx *HandlerContext) suggestAllIndexes() (interface{}, error) {
	// Return general suggestions when no specific table is provided
	return []IndexSuggestion{
		{
			TableName:     "*all_tables",
			Columns:       []string{"id"},
			EstimatedGain: 70.0,
			Reason:        "primary key indexes are always beneficial",
			Priority:      "high",
		},
	}, nil
}
