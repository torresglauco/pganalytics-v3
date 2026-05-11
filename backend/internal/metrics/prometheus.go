package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// APIResponseTimeHistogram tracks API response times by method and path
var APIResponseTimeHistogram = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "pganalytics_api_response_time_seconds",
		Help:    "API response time distribution in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	},
	[]string{"method", "path", "status"},
)

// QueryDurationHistogram tracks database query durations
var QueryDurationHistogram = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "pganalytics_query_duration_seconds",
		Help:    "Database query duration distribution in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	},
	[]string{"database", "query_type"},
)

// QueryCounter counts queries by type
var QueryCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "pganalytics_queries_total",
		Help: "Total number of database queries",
	},
	[]string{"database", "query_type", "status"},
)

// RecordAPIResponseTime records an API response time
func RecordAPIResponseTime(method, path string, status int, duration time.Duration) {
	APIResponseTimeHistogram.WithLabelValues(
		method,
		path,
		statusToString(status),
	).Observe(duration.Seconds())
}

// RecordQueryDuration records a database query duration
func RecordQueryDuration(database, queryType string, duration time.Duration) {
	QueryDurationHistogram.WithLabelValues(database, queryType).Observe(duration.Seconds())
}

// IncrementQueryCount increments the query counter
func IncrementQueryCount(database, queryType, status string) {
	QueryCounter.WithLabelValues(database, queryType, status).Inc()
}

// HistogramBuckets returns the bucket boundaries for documentation
func HistogramBuckets() []float64 {
	return []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
}

// PercentileLabels returns human-readable percentile labels
func PercentileLabels() map[float64]string {
	return map[float64]string{
		0.001:  "1ms",
		0.005:  "5ms",
		0.01:   "10ms",
		0.025:  "25ms",
		0.05:   "50ms",
		0.1:    "100ms",
		0.25:   "250ms",
		0.5:    "500ms",
		1.0:    "1s",
		2.5:    "2.5s",
		5.0:    "5s",
		10.0:   "10s",
	}
}

// statusToString converts HTTP status code to string
func statusToString(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}