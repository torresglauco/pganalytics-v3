package timescale

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
)

// AggregateJobStatus represents the status of a TimescaleDB aggregate refresh job
type AggregateJobStatus struct {
	JobID         int       `json:"job_id"`
	JobName       string    `json:"job_name"`
	Hypertable    string    `json:"hypertable"`
	LastRun       time.Time `json:"last_run"`
	LastRunStatus string    `json:"last_run_status"`
	NextRun       time.Time `json:"next_run"`
	TotalRuns     int       `json:"total_runs"`
	TotalFailures int       `json:"total_failures"`
}

// GetAggregateJobStatus queries timescaledb_information.jobs for aggregate refresh policies
func (t *TimescaleDB) GetAggregateJobStatus(ctx context.Context) ([]AggregateJobStatus, error) {
	query := `
		SELECT
			j.job_id,
			j.application_name,
			cagg.user_view_name as hypertable,
			j.last_run_started_at,
			j.last_run_status,
			j.next_start,
			j.total_runs,
			j.total_failures
		FROM timescaledb_information.jobs j
		LEFT JOIN timescaledb_information.continuous_aggregates cagg
			ON cagg.user_view_name = j.hypertable_name
		WHERE j.proc_name = 'policy_refresh_continuous_aggregate'
		ORDER BY j.job_id
	`

	rows, err := t.db.QueryContext(ctx, query)
	if err != nil {
		// TimescaleDB extension may not be available
		if isTimescaleNotAvailable(err) {
			return nil, nil
		}
		return nil, apperrors.DatabaseError("query aggregate job status", err.Error())
	}
	defer rows.Close()

	var jobs []AggregateJobStatus
	for rows.Next() {
		var job AggregateJobStatus
		var lastRun, nextRun sql.NullTime
		var status, hypertable sql.NullString
		var totalRuns, totalFailures sql.NullInt64

		if err := rows.Scan(
			&job.JobID,
			&job.JobName,
			&hypertable,
			&lastRun,
			&status,
			&nextRun,
			&totalRuns,
			&totalFailures,
		); err != nil {
			return nil, apperrors.DatabaseError("scan aggregate job row", err.Error())
		}

		if lastRun.Valid {
			job.LastRun = lastRun.Time
		}
		if status.Valid {
			job.LastRunStatus = status.String
		}
		if nextRun.Valid {
			job.NextRun = nextRun.Time
		}
		if hypertable.Valid {
			job.Hypertable = hypertable.String
		}
		if totalRuns.Valid {
			job.TotalRuns = int(totalRuns.Int64)
		}
		if totalFailures.Valid {
			job.TotalFailures = int(totalFailures.Int64)
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// isTimescaleNotAvailable checks if the error indicates TimescaleDB is not available
func isTimescaleNotAvailable(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "timescaledb_information") ||
		strings.Contains(errMsg, "does not exist") ||
		(strings.Contains(errMsg, "relation") && strings.Contains(errMsg, "does not exist"))
}

// RefreshAggregate manually refreshes a continuous aggregate (for testing)
func (t *TimescaleDB) RefreshAggregate(ctx context.Context, viewName string) error {
	query := fmt.Sprintf("CALL refresh_continuous_aggregate('%s', NULL, NULL)", viewName)
	_, err := t.db.ExecContext(ctx, query)
	if err != nil {
		return apperrors.DatabaseError("refresh continuous aggregate", err.Error())
	}
	return nil
}
