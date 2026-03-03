package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// SCHEMA METRICS MODELS
// ============================================================================

// SchemaTable represents a database table schema
type SchemaTable struct {
	ID           string    `json:"id" db:"id"`
	CollectorID  uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName string    `json:"database_name" db:"database_name"`
	SchemaName   string    `json:"schema_name" db:"schema_name"`
	TableName    string    `json:"table_name" db:"table_name"`
	TableType    string    `json:"table_type" db:"table_type"` // BASE TABLE, VIEW, etc.
	Timestamp    time.Time `json:"timestamp" db:"time"`
}

// SchemaColumn represents a column in a database table
type SchemaColumn struct {
	ID                  string    `json:"id" db:"id"`
	CollectorID         uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName        string    `json:"database_name" db:"database_name"`
	SchemaName          string    `json:"schema_name" db:"schema_name"`
	TableName           string    `json:"table_name" db:"table_name"`
	ColumnName          string    `json:"column_name" db:"column_name"`
	DataType            string    `json:"data_type" db:"data_type"`
	IsNullable          bool      `json:"is_nullable" db:"is_nullable"`
	ColumnDefault       *string   `json:"column_default" db:"column_default"`
	OrdinalPosition     int       `json:"ordinal_position" db:"ordinal_position"`
	CharacterMaxLength  *int      `json:"character_max_length" db:"character_max_length"`
	NumericPrecision    *int      `json:"numeric_precision" db:"numeric_precision"`
	NumericScale        *int      `json:"numeric_scale" db:"numeric_scale"`
	Timestamp           time.Time `json:"timestamp" db:"time"`
}

// SchemaConstraint represents a constraint on a table
type SchemaConstraint struct {
	ID              string    `json:"id" db:"id"`
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName    string    `json:"database_name" db:"database_name"`
	SchemaName      string    `json:"schema_name" db:"schema_name"`
	TableName       string    `json:"table_name" db:"table_name"`
	ConstraintName  string    `json:"constraint_name" db:"constraint_name"`
	ConstraintType  string    `json:"constraint_type" db:"constraint_type"` // PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK
	Columns         string    `json:"columns" db:"columns"`                 // Comma-separated
	Timestamp       time.Time `json:"timestamp" db:"time"`
}

// SchemaForeignKey represents a foreign key relationship
type SchemaForeignKey struct {
	ID              string    `json:"id" db:"id"`
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName    string    `json:"database_name" db:"database_name"`
	SourceSchema    string    `json:"source_schema" db:"source_schema"`
	SourceTable     string    `json:"source_table" db:"source_table"`
	SourceColumn    string    `json:"source_column" db:"source_column"`
	TargetSchema    string    `json:"target_schema" db:"target_schema"`
	TargetTable     string    `json:"target_table" db:"target_table"`
	TargetColumn    string    `json:"target_column" db:"target_column"`
	UpdateRule      string    `json:"update_rule" db:"update_rule"`
	DeleteRule      string    `json:"delete_rule" db:"delete_rule"`
	Timestamp       time.Time `json:"timestamp" db:"time"`
}

// ============================================================================
// LOCK METRICS MODELS
// ============================================================================

// Lock represents an active database lock
type Lock struct {
	ID              string    `json:"id" db:"id"`
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName    string    `json:"database_name" db:"database_name"`
	PID             int       `json:"pid" db:"pid"`
	LockType        string    `json:"locktype" db:"locktype"`
	Mode            string    `json:"mode" db:"mode"`
	Granted         bool      `json:"granted" db:"granted"`
	RelationID      *int      `json:"relation_id" db:"relation_id"`
	PageNumber      *int      `json:"page_number" db:"page_number"`
	TupleID         *int      `json:"tuple_id" db:"tuple_id"`
	Username        *string   `json:"username" db:"username"`
	SessionState    *string   `json:"session_state" db:"session_state"`
	LockAgeSeconds  *float64  `json:"lock_age_seconds" db:"lock_age_seconds"`
	Query           *string   `json:"query" db:"query"`
	Timestamp       time.Time `json:"timestamp" db:"time"`
}

// LockWait represents a lock wait scenario
type LockWait struct {
	ID                string    `json:"id" db:"id"`
	CollectorID       uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName      string    `json:"database_name" db:"database_name"`
	BlockedPID        int       `json:"blocked_pid" db:"blocked_pid"`
	BlockingPID       int       `json:"blocking_pid" db:"blocking_pid"`
	BlockedUsername   string    `json:"blocked_username" db:"blocked_username"`
	BlockingUsername  string    `json:"blocking_username" db:"blocking_username"`
	BlockedQuery      string    `json:"blocked_query" db:"blocked_query"`
	BlockingQuery     string    `json:"blocking_query" db:"blocking_query"`
	WaitTimeSeconds   *float64  `json:"wait_time_seconds" db:"wait_time_seconds"`
	BlockedApplication string   `json:"blocked_application" db:"blocked_application"`
	BlockingApplication string  `json:"blocking_application" db:"blocking_application"`
	Timestamp         time.Time `json:"timestamp" db:"time"`
}

// ============================================================================
// BLOAT METRICS MODELS
// ============================================================================

// TableBloat represents table bloat metrics
type TableBloat struct {
	ID                 string    `json:"id" db:"id"`
	CollectorID        uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName       string    `json:"database_name" db:"database_name"`
	SchemaName         string    `json:"schema_name" db:"schema_name"`
	TableName          string    `json:"table_name" db:"table_name"`
	DeadTuples         int64     `json:"dead_tuples" db:"dead_tuples"`
	LiveTuples         int64     `json:"live_tuples" db:"live_tuples"`
	DeadRatioPercent   float64   `json:"dead_ratio_percent" db:"dead_ratio_percent"`
	TableSize          string    `json:"table_size" db:"table_size"`
	SpaceWastedPercent *float64  `json:"space_wasted_percent" db:"space_wasted_percent"`
	LastVacuum         *time.Time `json:"last_vacuum" db:"last_vacuum"`
	LastAutovacuum     *time.Time `json:"last_autovacuum" db:"last_autovacuum"`
	VacuumCount        int64     `json:"vacuum_count" db:"vacuum_count"`
	AutovacuumCount    int64     `json:"autovacuum_count" db:"autovacuum_count"`
	Timestamp          time.Time `json:"timestamp" db:"time"`
}

// IndexBloat represents index bloat metrics
type IndexBloat struct {
	ID               string    `json:"id" db:"id"`
	CollectorID      uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName     string    `json:"database_name" db:"database_name"`
	SchemaName       string    `json:"schema_name" db:"schema_name"`
	TableName        string    `json:"table_name" db:"table_name"`
	IndexName        string    `json:"index_name" db:"index_name"`
	IndexScans       int64     `json:"index_scans" db:"index_scans"`
	TuplesRead       int64     `json:"tuples_read" db:"tuples_read"`
	TuplesFetched    int64     `json:"tuples_fetched" db:"tuples_fetched"`
	IndexSize        string    `json:"index_size" db:"index_size"`
	UsageStatus      string    `json:"usage_status" db:"usage_status"` // UNUSED, RARELY_USED, ACTIVE
	Recommendation   string    `json:"recommendation" db:"recommendation"` // CONSIDER_DROPPING, IN_USE
	Timestamp        time.Time `json:"timestamp" db:"time"`
}

// ============================================================================
// CACHE METRICS MODELS
// ============================================================================

// TableCacheHit represents table cache metrics
type TableCacheHit struct {
	ID                    string    `json:"id" db:"id"`
	CollectorID           uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName          string    `json:"database_name" db:"database_name"`
	SchemaName            string    `json:"schema_name" db:"schema_name"`
	TableName             string    `json:"table_name" db:"table_name"`
	HeapBlksHit           int64     `json:"heap_blks_hit" db:"heap_blks_hit"`
	HeapBlksRead          int64     `json:"heap_blks_read" db:"heap_blks_read"`
	HeapCacheHitRatio     float64   `json:"heap_cache_hit_ratio" db:"heap_cache_hit_ratio"` // Percentage
	IdxBlksHit            int64     `json:"idx_blks_hit" db:"idx_blks_hit"`
	IdxBlksRead           int64     `json:"idx_blks_read" db:"idx_blks_read"`
	IdxCacheHitRatio      float64   `json:"idx_cache_hit_ratio" db:"idx_cache_hit_ratio"` // Percentage
	ToastBlksHit          int64     `json:"toast_blks_hit" db:"toast_blks_hit"`
	ToastBlksRead         int64     `json:"toast_blks_read" db:"toast_blks_read"`
	TidxBlksHit           int64     `json:"tidx_blks_hit" db:"tidx_blks_hit"`
	TidxBlksRead          int64     `json:"tidx_blks_read" db:"tidx_blks_read"`
	Timestamp             time.Time `json:"timestamp" db:"time"`
}

// IndexCacheHit represents index cache metrics
type IndexCacheHit struct {
	ID              string    `json:"id" db:"id"`
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName    string    `json:"database_name" db:"database_name"`
	SchemaName      string    `json:"schema_name" db:"schema_name"`
	TableName       string    `json:"table_name" db:"table_name"`
	IndexName       string    `json:"index_name" db:"index_name"`
	BlksHit         int64     `json:"blks_hit" db:"blks_hit"`
	BlksRead        int64     `json:"blks_read" db:"blks_read"`
	CacheHitRatio   float64   `json:"cache_hit_ratio" db:"cache_hit_ratio"` // Percentage
	Timestamp       time.Time `json:"timestamp" db:"time"`
}

// ============================================================================
// CONNECTION METRICS MODELS
// ============================================================================

// ConnectionSummary represents connection state summary
type ConnectionSummary struct {
	ID               string    `json:"id" db:"id"`
	CollectorID      uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName     string    `json:"database_name" db:"database_name"`
	ConnectionState  string    `json:"connection_state" db:"connection_state"`
	ConnectionCount  int       `json:"connection_count" db:"connection_count"`
	MaxAgeSeconds    *float64  `json:"max_age_seconds" db:"max_age_seconds"`
	MinAgeSeconds    *float64  `json:"min_age_seconds" db:"min_age_seconds"`
	Timestamp        time.Time `json:"timestamp" db:"time"`
}

// LongRunningTransaction represents a long-running transaction
type LongRunningTransaction struct {
	ID               string    `json:"id" db:"id"`
	CollectorID      uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName     string    `json:"database_name" db:"database_name"`
	PID              int       `json:"pid" db:"pid"`
	Username         string    `json:"username" db:"username"`
	SessionState     *string   `json:"session_state" db:"session_state"`
	Query            *string   `json:"query" db:"query"`
	QueryStart       *time.Time `json:"query_start" db:"query_start"`
	DurationSeconds  *float64  `json:"duration_seconds" db:"duration_seconds"`
	ApplicationName  *string   `json:"application_name" db:"application_name"`
	ClientAddress    *string   `json:"client_address" db:"client_address"`
	Timestamp        time.Time `json:"timestamp" db:"time"`
}

// IdleTransaction represents an idle transaction
type IdleTransaction struct {
	ID              string    `json:"id" db:"id"`
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName    string    `json:"database_name" db:"database_name"`
	PID             int       `json:"pid" db:"pid"`
	Username        string    `json:"username" db:"username"`
	QueryStart      *time.Time `json:"query_start" db:"query_start"`
	StateChange     *time.Time `json:"state_change" db:"state_change"`
	IdleTimeSeconds *float64  `json:"idle_time_seconds" db:"idle_time_seconds"`
	ApplicationName *string   `json:"application_name" db:"application_name"`
	ClientAddress   *string   `json:"client_address" db:"client_address"`
	Timestamp       time.Time `json:"timestamp" db:"time"`
}

// ============================================================================
// EXTENSION METRICS MODELS
// ============================================================================

// Extension represents an installed database extension
type Extension struct {
	ID              string    `json:"id" db:"id"`
	CollectorID     uuid.UUID `json:"collector_id" db:"collector_id"`
	DatabaseName    string    `json:"database_name" db:"database_name"`
	ExtensionName   string    `json:"extension_name" db:"extension_name"`
	ExtensionVersion string   `json:"extension_version" db:"extension_version"`
	ExtensionOwner  *string   `json:"extension_owner" db:"extension_owner"`
	ExtensionSchema string    `json:"extension_schema" db:"extension_schema"`
	IsRelocatable   bool      `json:"is_relocatable" db:"is_relocatable"`
	Description     *string   `json:"description" db:"description"`
	Timestamp       time.Time `json:"timestamp" db:"time"`
}

// ============================================================================
// API REQUEST/RESPONSE MODELS
// ============================================================================

// MetricsQueryRequest represents a request to query metrics
type MetricsQueryRequest struct {
	CollectorID  uuid.UUID  `json:"collector_id"`
	DatabaseName *string    `json:"database_name,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// MetricsResponse wraps metrics data with metadata
type MetricsResponse struct {
	MetricType  string      `json:"metric_type"`
	Count       int         `json:"count"`
	Timestamp   time.Time   `json:"timestamp"`
	Data        interface{} `json:"data"`
}

// SchemaMetricsResponse contains all schema-related metrics
type SchemaMetricsResponse struct {
	Tables        []*SchemaTable        `json:"tables,omitempty"`
	Columns       []*SchemaColumn       `json:"columns,omitempty"`
	Constraints   []*SchemaConstraint   `json:"constraints,omitempty"`
	ForeignKeys   []*SchemaForeignKey   `json:"foreign_keys,omitempty"`
}

// LockMetricsResponse contains all lock-related metrics
type LockMetricsResponse struct {
	ActiveLocks     []*Lock      `json:"active_locks,omitempty"`
	LockWaitChains  []*LockWait  `json:"lock_wait_chains,omitempty"`
	BlockingQueries []*Lock      `json:"blocking_queries,omitempty"`
}

// BloatMetricsResponse contains all bloat-related metrics
type BloatMetricsResponse struct {
	TableBloat  []*TableBloat  `json:"table_bloat,omitempty"`
	IndexBloat  []*IndexBloat  `json:"index_bloat,omitempty"`
}

// CacheMetricsResponse contains all cache-related metrics
type CacheMetricsResponse struct {
	TableCacheHit []*TableCacheHit `json:"table_cache_hit,omitempty"`
	IndexCacheHit []*IndexCacheHit `json:"index_cache_hit,omitempty"`
}

// ConnectionMetricsResponse contains all connection-related metrics
type ConnectionMetricsResponse struct {
	ConnectionSummary       []*ConnectionSummary        `json:"connection_summary,omitempty"`
	LongRunningTransactions []*LongRunningTransaction   `json:"long_running_transactions,omitempty"`
	IdleTransactions        []*IdleTransaction          `json:"idle_transactions,omitempty"`
}

// ExtensionMetricsResponse contains all extension-related metrics
type ExtensionMetricsResponse struct {
	Extensions []*Extension `json:"extensions,omitempty"`
}
