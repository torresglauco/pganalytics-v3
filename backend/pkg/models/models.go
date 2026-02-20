package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// USER MODELS
// ============================================================================

// User represents a user in the system
type User struct {
	ID        int        `db:"id" json:"id"`
	Username  string     `db:"username" json:"username"`
	Email     string     `db:"email" json:"email"`
	FullName  string     `db:"full_name" json:"full_name,omitempty"`
	Role      string     `db:"role" json:"role"`
	IsActive  bool       `db:"is_active" json:"is_active"`
	LastLogin *time.Time `db:"last_login" json:"last_login,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// COLLECTOR MODELS
// ============================================================================

// Collector represents a distributed data collector
type Collector struct {
	ID                    uuid.UUID  `db:"id" json:"id"`
	Name                  string     `db:"name" json:"name"`
	Description           string     `db:"description" json:"description,omitempty"`
	Hostname              string     `db:"hostname" json:"hostname"`
	Address               *string    `db:"address" json:"address,omitempty"`
	Version               *string    `db:"version" json:"version,omitempty"`
	Status                string     `db:"status" json:"status"` // registered, active, offline, error
	LastSeen              *time.Time `db:"last_seen" json:"last_seen,omitempty"`
	CertificateThumbprint *string    `db:"certificate_thumbprint" json:"certificate_thumbprint,omitempty"`
	CertificateExpiresAt  *time.Time `db:"certificate_expires_at" json:"certificate_expires_at,omitempty"`
	ConfigVersion         int        `db:"config_version" json:"config_version"`
	MetricsCountTotal     int64      `db:"metrics_count_total" json:"metrics_count_total"`
	MetricsCount24h       int64      `db:"metrics_count_24h" json:"metrics_count_24h"`
	HealthCheckInterval   int        `db:"health_check_interval" json:"health_check_interval"`
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// CollectorConfig represents dynamic configuration for a collector
type CollectorConfig struct {
	ID          int       `db:"id" json:"id"`
	CollectorID uuid.UUID `db:"collector_id" json:"collector_id"`
	Version     int       `db:"version" json:"version"`
	Config      string    `db:"config" json:"config"` // JSON string
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedBy   *int      `db:"updated_by" json:"updated_by,omitempty"`
}

// ============================================================================
// SERVER & INSTANCE MODELS
// ============================================================================

// Server represents a monitored database server
type Server struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description,omitempty"`
	Hostname    string     `db:"hostname" json:"hostname"`
	Address     string     `db:"address" json:"address"`
	Environment string     `db:"environment" json:"environment"` // production, staging, development, test
	CollectorID *uuid.UUID `db:"collector_id" json:"collector_id,omitempty"`
	IsActive    bool       `db:"is_active" json:"is_active"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// PostgreSQLInstance represents a PostgreSQL database instance
type PostgreSQLInstance struct {
	ID                  int        `db:"id" json:"id"`
	ServerID            int        `db:"server_id" json:"server_id"`
	Name                string     `db:"name" json:"name"`
	Version             *string    `db:"version" json:"version,omitempty"`
	Port                int        `db:"port" json:"port"`
	ConnectionString    *string    `db:"connection_string" json:"connection_string,omitempty"`
	MaintenanceDatabase string     `db:"maintenance_database" json:"maintenance_database"`
	MonitoringRole      string     `db:"monitoring_role" json:"monitoring_role"`
	IsActive            bool       `db:"is_active" json:"is_active"`
	LastConnected       *time.Time `db:"last_connected" json:"last_connected,omitempty"`
	ReplicationRole     *string    `db:"replication_role" json:"replication_role,omitempty"` // primary, standby, unknown
	CreatedAt           time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}

// Database represents a database within a PostgreSQL instance
type Database struct {
	ID           int        `db:"id" json:"id"`
	InstanceID   int        `db:"instance_id" json:"instance_id"`
	Name         string     `db:"name" json:"name"`
	Owner        *string    `db:"owner" json:"owner,omitempty"`
	SizeBytes    *int64     `db:"size_bytes" json:"size_bytes,omitempty"`
	IsTemplate   bool       `db:"is_template" json:"is_template"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	LastAnalyzed *time.Time `db:"last_analyzed" json:"last_analyzed,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// AUTHENTICATION MODELS
// ============================================================================

// APIToken represents an authentication token
type APIToken struct {
	ID          int        `db:"id" json:"id"`
	CollectorID *uuid.UUID `db:"collector_id" json:"collector_id,omitempty"`
	UserID      *int       `db:"user_id" json:"user_id,omitempty"`
	TokenHash   string     `db:"token_hash" json:"token_hash"`
	Description string     `db:"description" json:"description,omitempty"`
	LastUsed    *time.Time `db:"last_used" json:"last_used,omitempty"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

// Secret represents an encrypted secret
type Secret struct {
	ID              int       `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	SecretEncrypted []byte    `db:"secret_encrypted" json:"-"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// ALERT MODELS
// ============================================================================

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID                  int       `db:"id" json:"id"`
	Name                string    `db:"name" json:"name"`
	Description         string    `db:"description" json:"description,omitempty"`
	MetricType          string    `db:"metric_type" json:"metric_type"`
	ConditionType       string    `db:"condition_type" json:"condition_type"` // threshold, change, anomaly
	ConditionValue      string    `db:"condition_value" json:"condition_value"`
	Severity            string    `db:"severity" json:"severity"` // info, warning, critical
	Enabled             bool      `db:"enabled" json:"enabled"`
	NotificationChannel string    `db:"notification_channel" json:"notification_channel"`
	EvaluationInterval  int       `db:"evaluation_interval" json:"evaluation_interval"`
	CreatedBy           *int      `db:"created_by" json:"created_by,omitempty"`
	CreatedAt           time.Time `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time `db:"updated_at" json:"updated_at"`
}

// Alert represents an alert instance
type Alert struct {
	ID             int        `db:"id" json:"id"`
	CollectorID    *uuid.UUID `db:"collector_id" json:"collector_id,omitempty"`
	RuleID         *int       `db:"rule_id" json:"rule_id,omitempty"`
	ServerID       *int       `db:"server_id" json:"server_id,omitempty"`
	DatabaseID     *int       `db:"database_id" json:"database_id,omitempty"`
	MetricType     string     `db:"metric_type" json:"metric_type"`
	MetricValue    string     `db:"metric_value" json:"metric_value"`
	Severity       string     `db:"severity" json:"severity"` // info, warning, critical
	Message        string     `db:"message" json:"message"`
	IsAcknowledged bool       `db:"is_acknowledged" json:"is_acknowledged"`
	AcknowledgedBy *int       `db:"acknowledged_by" json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `db:"acknowledged_at" json:"acknowledged_at,omitempty"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	ResolvedAt     *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
}

// ============================================================================
// AUDIT LOG MODELS
// ============================================================================

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int       `db:"id" json:"id"`
	UserID       *int      `db:"user_id" json:"user_id,omitempty"`
	Action       string    `db:"action" json:"action"`
	ResourceType string    `db:"resource_type" json:"resource_type"`
	ResourceID   string    `db:"resource_id" json:"resource_id"`
	Changes      string    `db:"changes" json:"changes,omitempty"` // JSON string
	IPAddress    *string   `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// ============================================================================
// QUERY STATISTICS MODELS
// ============================================================================

// QueryStats represents query-level performance statistics from pg_stat_statements
type QueryStats struct {
	ID                int64     `db:"id" json:"id"`
	Time              time.Time `db:"time" json:"time"`
	CollectorID       uuid.UUID `db:"collector_id" json:"collector_id"`
	DatabaseName      string    `db:"database_name" json:"database_name"`
	UserName          string    `db:"user_name" json:"user_name"`
	QueryHash         int64     `db:"query_hash" json:"query_hash"`
	QueryText         string    `db:"query_text" json:"query_text"`
	Calls             int64     `db:"calls" json:"calls"`
	TotalTime         float64   `db:"total_time" json:"total_time"`   // milliseconds
	MeanTime          float64   `db:"mean_time" json:"mean_time"`     // milliseconds
	MinTime           float64   `db:"min_time" json:"min_time"`       // milliseconds
	MaxTime           float64   `db:"max_time" json:"max_time"`       // milliseconds
	StddevTime        float64   `db:"stddev_time" json:"stddev_time"` // milliseconds
	Rows              int64     `db:"rows" json:"rows"`
	SharedBlksHit     int64     `db:"shared_blks_hit" json:"shared_blks_hit"`
	SharedBlksRead    int64     `db:"shared_blks_read" json:"shared_blks_read"`
	SharedBlksDirtied int64     `db:"shared_blks_dirtied" json:"shared_blks_dirtied"`
	SharedBlksWritten int64     `db:"shared_blks_written" json:"shared_blks_written"`
	LocalBlksHit      int64     `db:"local_blks_hit" json:"local_blks_hit"`
	LocalBlksRead     int64     `db:"local_blks_read" json:"local_blks_read"`
	LocalBlksDirtied  int64     `db:"local_blks_dirtied" json:"local_blks_dirtied"`
	LocalBlksWritten  int64     `db:"local_blks_written" json:"local_blks_written"`
	TempBlksRead      int64     `db:"temp_blks_read" json:"temp_blks_read"`
	TempBlksWritten   int64     `db:"temp_blks_written" json:"temp_blks_written"`
	BlkReadTime       float64   `db:"blk_read_time" json:"blk_read_time"`   // milliseconds
	BlkWriteTime      float64   `db:"blk_write_time" json:"blk_write_time"` // milliseconds
	WalRecords        *int64    `db:"wal_records" json:"wal_records,omitempty"`
	WalFpi            *int64    `db:"wal_fpi" json:"wal_fpi,omitempty"`
	WalBytes          *int64    `db:"wal_bytes" json:"wal_bytes,omitempty"`
	QueryPlanTime     *float64  `db:"query_plan_time" json:"query_plan_time,omitempty"`
	QueryExecTime     *float64  `db:"query_exec_time" json:"query_exec_time,omitempty"`
}

// QueryStatsRequest represents query stats data from collector
type QueryStatsRequest struct {
	Type      string         `json:"type"` // "pg_query_stats"
	Timestamp time.Time      `json:"timestamp"`
	Databases []QueryStatsDB `json:"databases"`
}

// QueryStatsDB represents query stats for a single database
type QueryStatsDB struct {
	Database string      `json:"database"`
	Queries  []QueryInfo `json:"queries"`
}

// QueryInfo represents a single query's statistics
type QueryInfo struct {
	Hash              int64    `json:"hash"`
	Text              string   `json:"text"`
	Calls             int64    `json:"calls"`
	TotalTime         float64  `json:"total_time"`
	MeanTime          float64  `json:"mean_time"`
	MinTime           float64  `json:"min_time"`
	MaxTime           float64  `json:"max_time"`
	StddevTime        float64  `json:"stddev_time"`
	Rows              int64    `json:"rows"`
	SharedBlksHit     int64    `json:"shared_blks_hit"`
	SharedBlksRead    int64    `json:"shared_blks_read"`
	SharedBlksDirtied int64    `json:"shared_blks_dirtied"`
	SharedBlksWritten int64    `json:"shared_blks_written"`
	LocalBlksHit      int64    `json:"local_blks_hit"`
	LocalBlksRead     int64    `json:"local_blks_read"`
	LocalBlksDirtied  int64    `json:"local_blks_dirtied"`
	LocalBlksWritten  int64    `json:"local_blks_written"`
	TempBlksRead      int64    `json:"temp_blks_read"`
	TempBlksWritten   int64    `json:"temp_blks_written"`
	BlkReadTime       float64  `json:"blk_read_time"`
	BlkWriteTime      float64  `json:"blk_write_time"`
	WalRecords        *int64   `json:"wal_records,omitempty"`
	WalFpi            *int64   `json:"wal_fpi,omitempty"`
	WalBytes          *int64   `json:"wal_bytes,omitempty"`
	QueryPlanTime     *float64 `json:"query_plan_time,omitempty"`
	QueryExecTime     *float64 `json:"query_exec_time,omitempty"`
}

// TopQueriesResponse represents the response for top queries API
type TopQueriesResponse struct {
	ServerID  uuid.UUID     `json:"server_id"`
	QueryType string        `json:"type"` // "slow", "frequent"
	Hours     int           `json:"hours"`
	Count     int           `json:"count"`
	Queries   []*QueryStats `json:"queries"`
}

// QueryTimelineResponse represents the response for query timeline API
type QueryTimelineResponse struct {
	QueryHash int64         `json:"query_hash"`
	Data      []*QueryStats `json:"data"`
}

// ============================================================================
// RESPONSE MODELS
// ============================================================================

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user"`
}

// CollectorRegisterRequest represents a collector registration request
type CollectorRegisterRequest struct {
	Name     string  `json:"name" binding:"required"`
	Hostname string  `json:"hostname" binding:"required"`
	Address  *string `json:"address,omitempty"`
	Version  *string `json:"version,omitempty"`
}

// CollectorRegisterResponse represents a successful registration response
type CollectorRegisterResponse struct {
	CollectorID uuid.UUID `json:"collector_id"`
	Token       string    `json:"token"`
	Certificate string    `json:"certificate"` // PEM format
	PrivateKey  string    `json:"private_key"` // PEM format
	ExpiresAt   time.Time `json:"expires_at"`
}

// MetricsPushRequest represents metrics being pushed by a collector
type MetricsPushRequest struct {
	CollectorID  string        `json:"collector_id" binding:"required"`
	Hostname     string        `json:"hostname" binding:"required"`
	Timestamp    time.Time     `json:"timestamp" binding:"required"`
	Version      string        `json:"version,omitempty"`
	MetricsCount int           `json:"metrics_count"`
	Metrics      []interface{} `json:"metrics"` // Flexible metric structure
}

// MetricsPushResponse represents the response to a metrics push
type MetricsPushResponse struct {
	Status             string `json:"status"` // success, error
	CollectorID        string `json:"collector_id"`
	MetricsInserted    int    `json:"metrics_inserted"`
	BytesReceived      int    `json:"bytes_received"`
	ProcessingTimeMs   int64  `json:"processing_time_ms"`
	NextConfigVersion  int    `json:"next_config_version"`
	NextCheckInSeconds int    `json:"next_check_in_seconds"`
}

// ExplainPlanRequest represents a request to store an EXPLAIN plan
type ExplainPlanRequest struct {
	QueryHash    int64  `json:"query_hash" binding:"required"`
	QueryText    string `json:"query_text" binding:"required"`
	DatabaseName string `json:"database_name" binding:"required"`
}

// HealthResponse represents system health status
type HealthResponse struct {
	Status      string    `json:"status"` // ok, degraded, error
	Version     string    `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
	Uptime      int64     `json:"uptime"`
	DatabaseOk  bool      `json:"database_ok"`
	TimescaleOk bool      `json:"timescale_ok"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// ============================================================================
// PAGINATION MODELS
// ============================================================================

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// ============================================================================
// PHASE 4.4: ADVANCED QUERY ANALYSIS MODELS
// ============================================================================

// QueryFingerprint represents a grouped set of similar queries
type QueryFingerprint struct {
	ID               int64      `db:"id" json:"id"`
	FingerprintHash  int64      `db:"fingerprint_hash" json:"fingerprint_hash"`
	NormalizedText   string     `db:"normalized_text" json:"normalized_text"`
	SampleQueryText  *string    `db:"sample_query_text" json:"sample_query_text,omitempty"`
	CollectorID      *uuid.UUID `db:"collector_id" json:"collector_id,omitempty"`
	DatabaseName     *string    `db:"database_name" json:"database_name,omitempty"`
	TotalCalls       int64      `db:"total_calls" json:"total_calls"`
	AvgExecutionTime float64    `db:"avg_execution_time" json:"avg_execution_time"`
	FirstSeen        time.Time  `db:"first_seen" json:"first_seen"`
	LastSeen         time.Time  `db:"last_seen" json:"last_seen"`
}

// QueryFingerprintResponse groups queries by fingerprint with statistics
type QueryFingerprintResponse struct {
	FingerprintHash  int64     `json:"fingerprint_hash"`
	NormalizedQuery  string    `json:"normalized_query"`
	TotalCalls       int64     `json:"total_calls"`
	UniqueQueryCount int       `json:"unique_query_count"`
	AvgExecutionTime float64   `json:"avg_execution_time"`
	MaxExecutionTime float64   `json:"max_execution_time"`
	MinExecutionTime float64   `json:"min_execution_time"`
	SampleQueries    []string  `json:"sample_queries"`
	FirstSeen        time.Time `json:"first_seen"`
	LastSeen         time.Time `json:"last_seen"`
}

// ExplainPlan represents a stored EXPLAIN plan output
type ExplainPlan struct {
	ID                   int64       `db:"id" json:"id"`
	QueryHash            int64       `db:"query_hash" json:"query_hash"`
	QueryFingerprintHash *int64      `db:"query_fingerprint_hash" json:"query_fingerprint_hash,omitempty"`
	CollectedAt          time.Time   `db:"collected_at" json:"collected_at"`
	PlanJSON             interface{} `db:"plan_json" json:"plan_json"` // JSONB
	PlanText             *string     `db:"plan_text" json:"plan_text,omitempty"`
	RowsExpected         *int64      `db:"rows_expected" json:"rows_expected,omitempty"`
	RowsActual           *int64      `db:"rows_actual" json:"rows_actual,omitempty"`
	PlanDurationMs       *float64    `db:"plan_duration_ms" json:"plan_duration_ms,omitempty"`
	ExecutionDurationMs  *float64    `db:"execution_duration_ms" json:"execution_duration_ms,omitempty"`
	HasSeqScan           bool        `db:"has_seq_scan" json:"has_seq_scan"`
	HasIndexScan         bool        `db:"has_index_scan" json:"has_index_scan"`
	HasBitmapScan        bool        `db:"has_bitmap_scan" json:"has_bitmap_scan"`
	HasNestedLoop        bool        `db:"has_nested_loop" json:"has_nested_loop"`
	TotalBuffersRead     *int64      `db:"total_buffers_read" json:"total_buffers_read,omitempty"`
	TotalBuffersHit      *int64      `db:"total_buffers_hit" json:"total_buffers_hit,omitempty"`
}

// IndexRecommendation represents a recommended index
type IndexRecommendation struct {
	ID                      int64       `db:"id" json:"id"`
	CollectorID             *uuid.UUID  `db:"collector_id" json:"collector_id,omitempty"`
	DatabaseName            string      `db:"database_name" json:"database_name"`
	SchemaName              string      `db:"schema_name" json:"schema_name"`
	TableName               string      `db:"table_name" json:"table_name"`
	ColumnNames             interface{} `db:"column_names" json:"column_names"` // TEXT[]
	CreateStatement         string      `db:"create_statement" json:"create_statement"`
	EstimatedImprovementPct float64     `db:"estimated_improvement_percent" json:"estimated_improvement_percent"`
	AffectedQueryCount      int64       `db:"affected_query_count" json:"affected_query_count"`
	AffectedTotalTimeMs     float64     `db:"affected_total_time_ms" json:"affected_total_time_ms"`
	FrequencyScore          float64     `db:"frequency_score" json:"frequency_score"`
	ImpactScore             float64     `db:"impact_score" json:"impact_score"`
	ConfidenceScore         float64     `db:"confidence_score" json:"confidence_score"`
	Dismissed               bool        `db:"dismissed" json:"dismissed"`
	DismissedAt             *time.Time  `db:"dismissed_at" json:"dismissed_at,omitempty"`
	DismissedReason         *string     `db:"dismissed_reason" json:"dismissed_reason,omitempty"`
	CreatedAt               time.Time   `db:"created_at" json:"created_at"`
}

// QueryAnomaly represents a detected anomaly in query performance
type QueryAnomaly struct {
	ID                   int64       `db:"id" json:"id"`
	QueryHash            int64       `db:"query_hash" json:"query_hash"`
	QueryFingerprintHash *int64      `db:"query_fingerprint_hash" json:"query_fingerprint_hash,omitempty"`
	AnomalyType          string      `db:"anomaly_type" json:"anomaly_type"` // execution_time_spike, cache_degradation, etc.
	Severity             string      `db:"severity" json:"severity"`         // low, medium, high
	DetectedAt           time.Time   `db:"detected_at" json:"detected_at"`
	MetricName           *string     `db:"metric_name" json:"metric_name,omitempty"`
	MetricValue          *float64    `db:"metric_value" json:"metric_value,omitempty"`
	BaselineValue        *float64    `db:"baseline_value" json:"baseline_value,omitempty"`
	DeviationStddev      *float64    `db:"deviation_stddev" json:"deviation_stddev,omitempty"`
	ZScore               *float64    `db:"z_score" json:"z_score,omitempty"`                   // Standard deviations from mean
	RawMetricsJSON       interface{} `db:"raw_metrics_json" json:"raw_metrics_json,omitempty"` // JSONB
	Resolved             bool        `db:"resolved" json:"resolved"`
	ResolvedAt           *time.Time  `db:"resolved_at" json:"resolved_at,omitempty"`
}

// QueryBaseline stores baseline metrics for anomaly detection
type QueryBaseline struct {
	ID                 int64     `db:"id" json:"id"`
	QueryHash          int64     `db:"query_hash" json:"query_hash"`
	MetricName         string    `db:"metric_name" json:"metric_name"`
	BaselineValue      float64   `db:"baseline_value" json:"baseline_value"`
	StddevValue        float64   `db:"stddev_value" json:"stddev_value"`
	BaselinePeriodDays int       `db:"baseline_period_days" json:"baseline_period_days"`
	LastUpdated        time.Time `db:"last_updated" json:"last_updated"`
	MinValue           *float64  `db:"min_value" json:"min_value,omitempty"`
	MaxValue           *float64  `db:"max_value" json:"max_value,omitempty"`
}

// PerformanceSnapshot captures a point-in-time snapshot of all query metrics
type PerformanceSnapshot struct {
	ID           int64       `db:"id" json:"id"`
	Name         string      `db:"name" json:"name"`
	Description  *string     `db:"description" json:"description,omitempty"`
	SnapshotType string      `db:"snapshot_type" json:"snapshot_type"` // manual, scheduled, pre_deploy, post_deploy
	CreatedAt    time.Time   `db:"created_at" json:"created_at"`
	CreatedBy    *string     `db:"created_by" json:"created_by,omitempty"`
	MetadataJSON interface{} `db:"metadata_json" json:"metadata_json,omitempty"` // JSONB
}

// QueryPerformanceSnapshot stores query metrics for a specific snapshot
type QueryPerformanceSnapshot struct {
	ID                   int64    `db:"id" json:"id"`
	SnapshotID           int64    `db:"snapshot_id" json:"snapshot_id"`
	QueryHash            int64    `db:"query_hash" json:"query_hash"`
	QueryFingerprintHash *int64   `db:"query_fingerprint_hash" json:"query_fingerprint_hash,omitempty"`
	DatabaseName         *string  `db:"database_name" json:"database_name,omitempty"`
	Calls                int64    `db:"calls" json:"calls"`
	TotalTime            float64  `db:"total_time" json:"total_time"`
	MeanTime             float64  `db:"mean_time" json:"mean_time"`
	MaxTime              float64  `db:"max_time" json:"max_time"`
	MinTime              float64  `db:"min_time" json:"min_time"`
	StddevTime           *float64 `db:"stddev_time" json:"stddev_time,omitempty"`
	Rows                 int64    `db:"rows" json:"rows"`
	SharedBlksHit        int64    `db:"shared_blks_hit" json:"shared_blks_hit"`
	SharedBlksRead       int64    `db:"shared_blks_read" json:"shared_blks_read"`
	BlkReadTime          float64  `db:"blk_read_time" json:"blk_read_time"`
	BlkWriteTime         float64  `db:"blk_write_time" json:"blk_write_time"`
	// Optional PG13+ fields
	WalRecords    *int64   `db:"wal_records" json:"wal_records,omitempty"`
	WalFpi        *int64   `db:"wal_fpi" json:"wal_fpi,omitempty"`
	WalBytes      *int64   `db:"wal_bytes" json:"wal_bytes,omitempty"`
	QueryPlanTime *float64 `db:"query_plan_time" json:"query_plan_time,omitempty"`
	QueryExecTime *float64 `db:"query_exec_time" json:"query_exec_time,omitempty"`
}

// SnapshotComparison represents a comparison between two snapshots
type SnapshotComparison struct {
	QueryHash             int64    `json:"query_hash"`
	DatabaseName          *string  `json:"database_name,omitempty"`
	BeforeCalls           int64    `json:"before_calls"`
	AfterCalls            int64    `json:"after_calls"`
	CallsChange           int64    `json:"calls_change"`
	CallsChangePercent    *float64 `json:"calls_change_percent,omitempty"`
	BeforeMeanTime        float64  `json:"before_mean_time"`
	AfterMeanTime         float64  `json:"after_mean_time"`
	MeanTimeChange        float64  `json:"mean_time_change"`
	MeanTimeChangePercent *float64 `json:"mean_time_change_percent,omitempty"`
	BeforeMaxTime         float64  `json:"before_max_time"`
	AfterMaxTime          float64  `json:"after_max_time"`
	MaxTimeChange         float64  `json:"max_time_change"`
	BeforeCacheHits       int64    `json:"before_cache_hits"`
	AfterCacheHits        int64    `json:"after_cache_hits"`
	BeforeCacheReads      int64    `json:"before_cache_reads"`
	AfterCacheReads       int64    `json:"after_cache_reads"`
	ImprovementStatus     string   `json:"improvement_status"` // improved, degraded, unchanged
}
