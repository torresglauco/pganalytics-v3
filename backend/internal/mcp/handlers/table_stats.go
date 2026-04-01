package handlers

import "fmt"

func (ctx *HandlerContext) TableStats(params map[string]interface{}) (interface{}, error) {
	tableName, ok := params["table_name"].(string)
	if !ok || tableName == "" {
		return nil, fmt.Errorf("table_name parameter required")
	}

	if ctx.DB == nil {
		// Return mock data for testing
		return []TableStats{
			{
				TableName:         tableName,
				RowCount:          1000000,
				SizeBytes:         10485760,
				IndexCount:        3,
				LastAutovacuum:    "2026-04-01T10:00:00Z",
				DeadRowsPercent:   2.5,
				TableBloatPercent: 5.0,
			},
		}, nil
	}

	// Real implementation would query database
	query := `
		SELECT
			schemaname || '.' || tablename as table_name,
			n_live_tup as row_count,
			pg_total_relation_size(schemaname || '.' || tablename) as size_bytes,
			(SELECT count(*) FROM pg_indexes WHERE tablename = t.tablename) as index_count,
			last_autovacuum,
			ROUND(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_rows_percent,
			ROUND(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0) * 0.5, 2) as table_bloat_percent
		FROM pg_stat_user_tables t
		WHERE tablename = $1
	`

	var stats []TableStats
	rows, err := ctx.DB.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s TableStats
		if err := rows.Scan(&s.TableName, &s.RowCount, &s.SizeBytes, &s.IndexCount, &s.LastAutovacuum, &s.DeadRowsPercent, &s.TableBloatPercent); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}
