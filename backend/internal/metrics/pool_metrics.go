package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
)

var (
	poolOpenConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pganalytics_pool_open_connections",
		Help: "Number of open connections in the pool",
	}, []string{"database", "pool_type"})

	poolIdleConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pganalytics_pool_idle_connections",
		Help: "Number of idle connections in the pool",
	}, []string{"database", "pool_type"})

	poolInUseConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pganalytics_pool_in_use_connections",
		Help: "Number of connections in use",
	}, []string{"database", "pool_type"})

	poolMaxConns = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pganalytics_pool_max_connections",
		Help: "Maximum connections allowed in the pool",
	}, []string{"database", "pool_type"})

	poolWaitCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pganalytics_pool_wait_count",
		Help: "Number of connections that had to wait",
	}, []string{"database", "pool_type"})

	poolWaitDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pganalytics_pool_wait_duration_ms",
		Help: "Total duration of connection waits in milliseconds",
	}, []string{"database", "pool_type"})
)

// UpdatePoolMetrics updates Prometheus gauges from pool metrics
func UpdatePoolMetrics(database, poolType string, metrics storage.PoolMetrics) {
	poolOpenConns.WithLabelValues(database, poolType).Set(float64(metrics.OpenConns))
	poolIdleConns.WithLabelValues(database, poolType).Set(float64(metrics.IdleConns))
	poolInUseConns.WithLabelValues(database, poolType).Set(float64(metrics.InUseConns))
	poolMaxConns.WithLabelValues(database, poolType).Set(float64(metrics.MaxOpenConns))
	poolWaitCount.WithLabelValues(database, poolType).Set(float64(metrics.WaitCount))
	poolWaitDuration.WithLabelValues(database, poolType).Set(float64(metrics.WaitDuration))
}

// UpdateAllPoolMetrics updates Prometheus gauges for all pools
func UpdateAllPoolMetrics(postgresMetrics, timescaleMetrics map[string]storage.PoolMetrics) {
	// Update postgres pool metrics
	for poolType, m := range postgresMetrics {
		UpdatePoolMetrics("postgres", poolType, m)
	}

	// Update timescale pool metrics
	UpdatePoolMetrics("timescale", "primary", timescaleMetrics["primary"])
}
