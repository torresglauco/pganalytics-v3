package storage

// PoolMetrics contains connection pool statistics
type PoolMetrics struct {
	OpenConns    int32 `json:"open_connections"`
	IdleConns    int32 `json:"idle_connections"`
	InUseConns   int32 `json:"in_use_connections"`
	MaxOpenConns int32 `json:"max_open_connections"`
	WaitCount    int64 `json:"wait_count"`
	WaitDuration int64 `json:"wait_duration_ms"`
}

// PoolMetricsProvider interface for pool implementations
type PoolMetricsProvider interface {
	GetPoolMetrics() PoolMetrics
}
