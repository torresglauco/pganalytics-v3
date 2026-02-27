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
	ID           int        `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	FullName     string     `db:"full_name" json:"full_name,omitempty"`
	Role         string     `db:"role" json:"role"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	LastLogin    *time.Time `db:"last_login" json:"last_login,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
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
// RDS MONITORING MODELS
// ============================================================================

// ManagedInstanceCluster represents a group of RDS instances (master + replicas)
type ManagedInstanceCluster struct {
	ID          int                   `db:"id" json:"id"`
	Name        string                `db:"name" json:"name"`
	Description *string               `db:"description" json:"description,omitempty"`
	ClusterType string                `db:"cluster_type" json:"cluster_type"`
	Environment string                `db:"environment" json:"environment"`
	Status      string                `db:"status" json:"status"`
	IsActive    bool                  `db:"is_active" json:"is_active"`
	Tags        map[string]interface{} `db:"tags" json:"tags,omitempty"`
	CreatedAt   time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time             `db:"updated_at" json:"updated_at"`
	CreatedBy   *int                  `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy   *int                  `db:"updated_by" json:"updated_by,omitempty"`
	Instances   []*ManagedInstance        `db:"-" json:"instances,omitempty"`
}

// ManagedInstance represents an AWS RDS PostgreSQL instance
type ManagedInstance struct {
	ID                      int        `db:"id" json:"id"`
	Name                    string     `db:"name" json:"name"`
	Description             *string    `db:"description" json:"description,omitempty"`
	ClusterID               *int       `db:"cluster_id" json:"cluster_id,omitempty"`
	InstanceRole            string     `db:"instance_role" json:"instance_role"`
	AWSRegion               string     `db:"aws_region" json:"aws_region"`
	Endpoint             string     `db:"endpoint" json:"endpoint"`
	Port                    int        `db:"port" json:"port"`
	EngineVersion           *string    `db:"engine_version" json:"engine_version,omitempty"`
	DBInstanceClass         *string    `db:"db_instance_class" json:"db_instance_class,omitempty"`
	AllocatedStorageGB      *int       `db:"allocated_storage_gb" json:"allocated_storage_gb,omitempty"`
	Environment             string     `db:"environment" json:"environment"`
	MasterUsername          string     `db:"master_username" json:"master_username"`
	SecretID                *int       `db:"secret_id" json:"secret_id,omitempty"`
	EnableEnhancedMonitoring bool      `db:"enable_enhanced_monitoring" json:"enable_enhanced_monitoring"`
	MonitoringInterval      int        `db:"monitoring_interval" json:"monitoring_interval"`
	SSLEnabled              bool       `db:"ssl_enabled" json:"ssl_enabled"`
	SSLMode                 string     `db:"ssl_mode" json:"ssl_mode"`
	ConnectionTimeout       int        `db:"connection_timeout" json:"connection_timeout"`
	IsActive                bool       `db:"is_active" json:"is_active"`
	Status                  string     `db:"status" json:"status"`
	LastHeartbeat           *time.Time `db:"last_heartbeat" json:"last_heartbeat,omitempty"`
	LastConnectionStatus    string     `db:"last_connection_status" json:"last_connection_status"`
	LastErrorMessage        *string    `db:"last_error_message" json:"last_error_message,omitempty"`
	LastErrorTime           *time.Time `db:"last_error_time" json:"last_error_time,omitempty"`
	MultiAZ                 bool       `db:"multi_az" json:"multi_az"`
	BackupRetentionDays     *int       `db:"backup_retention_days" json:"backup_retention_days,omitempty"`
	PreferredBackupWindow   *string    `db:"preferred_backup_window" json:"preferred_backup_window,omitempty"`
	PreferredMaintenanceWindow *string `db:"preferred_maintenance_window" json:"preferred_maintenance_window,omitempty"`
	Tags                    map[string]interface{} `db:"tags" json:"tags,omitempty"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
	CreatedBy               *int       `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy               *int       `db:"updated_by" json:"updated_by,omitempty"`
}

// ManagedInstanceDatabase represents a database within an RDS instance
type ManagedInstanceDatabase struct {
	ID           int        `db:"id" json:"id"`
	ManagedInstanceID int       `db:"rds_instance_id" json:"rds_instance_id"`
	Name         string     `db:"name" json:"name"`
	Owner        *string    `db:"owner" json:"owner,omitempty"`
	SizeBytes    *int64     `db:"size_bytes" json:"size_bytes,omitempty"`
	IsTemplate   bool       `db:"is_template" json:"is_template"`
	IsActive     bool       `db:"is_active" json:"is_active"`
	LastAnalyzed *time.Time `db:"last_analyzed" json:"last_analyzed,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// ManagedInstanceMetric represents a CloudWatch metric for an RDS instance
type ManagedInstanceMetric struct {
	ID                int        `db:"id" json:"id"`
	ManagedInstanceID     int        `db:"rds_instance_id" json:"rds_instance_id"`
	MetricTimestamp   time.Time  `db:"metric_timestamp" json:"metric_timestamp"`
	MetricType        string     `db:"metric_type" json:"metric_type"`
	MetricValue       float64    `db:"metric_value" json:"metric_value"`
	MetricUnit        *string    `db:"metric_unit" json:"metric_unit,omitempty"`
	Dimensions        map[string]interface{} `db:"dimensions" json:"dimensions,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
}

// CreateManagedInstanceRequest represents a request to add a new RDS instance
// Requires minimum: name, endpoint, port, environment, master_username, master_password
type CreateManagedInstanceRequest struct {
	Name                      string `json:"name" binding:"required,min=3"`
	AWSRegion                 string `json:"aws_region"`                 // Optional - defaults to us-east-1
	Endpoint               string `json:"endpoint" binding:"required"`
	Port                      int    `json:"port" binding:"required,min=1,max=65535"`
	Environment               string `json:"environment"`               // Optional - defaults to development
	MasterUsername            string `json:"master_username" binding:"required"`
	MasterPassword            string `json:"master_password" binding:"required"`
	Description               string `json:"description"`
	EngineVersion             string `json:"engine_version"`
	DBInstanceClass           string `json:"db_instance_class"`
	AllocatedStorageGB        int    `json:"allocated_storage_gb"`
	EnableEnhancedMonitoring  bool   `json:"enable_enhanced_monitoring"`
	MonitoringInterval        int    `json:"monitoring_interval" binding:"min=0"`
	SSLEnabled                bool   `json:"ssl_enabled"`
	SSLMode                   string `json:"ssl_mode"`
	ConnectionTimeout         int    `json:"connection_timeout" binding:"min=0"`
	MultiAZ                   bool   `json:"multi_az"`
	BackupRetentionDays       int    `json:"backup_retention_days"`
	PreferredBackupWindow     string `json:"preferred_backup_window"`
	PreferredMaintenanceWindow string `json:"preferred_maintenance_window"`
	Tags                      map[string]interface{} `json:"tags"`
}

// UpdateManagedInstanceRequest represents a request to update an RDS instance
type UpdateManagedInstanceRequest struct {
	Name                      string `json:"name" binding:"required,min=3"`
	AWSRegion                 string `json:"aws_region"`                 // Optional
	Endpoint               string `json:"endpoint" binding:"required"`
	Port                      int    `json:"port" binding:"required,min=1,max=65535"`
	Environment               string `json:"environment"`               // Optional
	MasterUsername            string `json:"master_username" binding:"required"`
	MasterPassword            string `json:"master_password" binding:"required"`
	Description               string `json:"description"`
	Status                    string `json:"status" binding:"required,oneof=registering registered monitoring paused"`
	EngineVersion             string `json:"engine_version"`
	DBInstanceClass           string `json:"db_instance_class"`
	AllocatedStorageGB        int    `json:"allocated_storage_gb"`
	EnableEnhancedMonitoring  bool   `json:"enable_enhanced_monitoring"`
	MonitoringInterval        int    `json:"monitoring_interval" binding:"min=0"`
	SSLEnabled                bool   `json:"ssl_enabled"`
	SSLMode                   string `json:"ssl_mode"`
	ConnectionTimeout         int    `json:"connection_timeout" binding:"min=0"`
	MultiAZ                   bool   `json:"multi_az"`
	BackupRetentionDays       int    `json:"backup_retention_days"`
	PreferredBackupWindow     string `json:"preferred_backup_window"`
	PreferredMaintenanceWindow string `json:"preferred_maintenance_window"`
	Tags                      map[string]interface{} `json:"tags"`
}

// CreateManagedInstanceClusterRequest represents a request to create an RDS cluster
type CreateManagedInstanceClusterRequest struct {
	Name        string                 `json:"name" binding:"required,min=3"`
	Description string                 `json:"description"`
	ClusterType string                 `json:"cluster_type" binding:"required,oneof=single-az multi-az aurora custom"`
	Environment string                 `json:"environment" binding:"required,oneof=production staging development test"`
	Tags        map[string]interface{} `json:"tags"`
}

// UpdateManagedInstanceClusterRequest represents a request to update an RDS cluster
type UpdateManagedInstanceClusterRequest struct {
	Name        string                 `json:"name" binding:"required,min=3"`
	Description string                 `json:"description"`
	ClusterType string                 `json:"cluster_type" binding:"required,oneof=single-az multi-az aurora custom"`
	Environment string                 `json:"environment" binding:"required,oneof=production staging development test"`
	Status      string                 `json:"status" binding:"required,oneof=registering registered monitoring paused"`
	Tags        map[string]interface{} `json:"tags"`
}

// CreateManagedInstanceWithClusterRequest extends CreateManagedInstanceRequest with cluster info
type CreateManagedInstanceWithClusterRequest struct {
	CreateManagedInstanceRequest
	ClusterID    *int   `json:"cluster_id"`
	InstanceRole string `json:"instance_role" binding:"required,oneof=master read-replica standby standalone"`
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

// SignupRequest represents a user signup request
type SignupRequest struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"max=255"`
}

// CreateUserRequest represents an admin request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"max=255"`
	Role     string `json:"role" binding:"required,oneof=admin user viewer"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user"`
}

// ChangePasswordRequest represents a user request to change their own password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPasswordResponse represents the response from a password reset
type ResetPasswordResponse struct {
	Username    string `json:"username"`
	TempPassword string `json:"temp_password"`
	Message     string `json:"message"`
}

// ============================================================================
// RDS MODELS
// ============================================================================

// TestConnectionRequest represents a request to test RDS connection (for existing instance)
type TestConnectionRequest struct {
	Username string `json:"username"` // Optional - uses stored username from instance if not provided
	Password string `json:"password"` // Optional - uses decrypted password from secret if not provided
}

// TestManagedInstanceConnectionRequest represents a request to test RDS connection with endpoint details
type TestManagedInstanceConnectionRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
	Port        int    `json:"port" binding:"required,min=1,max=65535"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// TestConnectionResponse represents the response from a connection test
type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
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

// ============================================================================
// PHASE 4.5: ML-BASED QUERY OPTIMIZATION SUGGESTIONS MODELS
// ============================================================================

// WorkloadPattern represents a detected recurring pattern in query execution
type WorkloadPattern struct {
	ID                 int64                  `db:"id" json:"id"`
	DatabaseName       string                 `db:"database_name" json:"database_name"`
	PatternType        string                 `db:"pattern_type" json:"pattern_type"`         // hourly_peak, daily_cycle, weekly_pattern, batch_job
	PatternMetadata    map[string]interface{} `db:"pattern_metadata" json:"pattern_metadata"` // peak_hour, variance, confidence, affected_queries
	DetectionTimestamp time.Time              `db:"detection_timestamp" json:"detection_timestamp"`
	Description        *string                `db:"description" json:"description,omitempty"`
	AffectedQueryCount int                    `db:"affected_query_count" json:"affected_query_count"`
}

// QueryRewriteSuggestion represents a recommended SQL rewrite
type QueryRewriteSuggestion struct {
	ID                      int64     `db:"id" json:"id"`
	QueryHash               int64     `db:"query_hash" json:"query_hash"`
	FingerprintHash         *int64    `db:"fingerprint_hash" json:"fingerprint_hash,omitempty"`
	SuggestionType          string    `db:"suggestion_type" json:"suggestion_type"` // n_plus_one_detected, subquery_optimization, join_reorder, missing_limit
	Description             string    `db:"description" json:"description"`
	OriginalQuery           *string   `db:"original_query" json:"original_query,omitempty"`
	SuggestedRewrite        string    `db:"suggested_rewrite" json:"suggested_rewrite"`
	Reasoning               *string   `db:"reasoning" json:"reasoning,omitempty"`
	EstimatedImprovementPct float64   `db:"estimated_improvement_percent" json:"estimated_improvement_percent"`
	ConfidenceScore         float64   `db:"confidence_score" json:"confidence_score"`
	Dismissed               bool      `db:"dismissed" json:"dismissed"`
	Implemented             bool      `db:"implemented" json:"implemented"`
	ImplementationNotes     *string   `db:"implementation_notes" json:"implementation_notes,omitempty"`
	CreatedAt               time.Time `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time `db:"updated_at" json:"updated_at"`
}

// ParameterTuningSuggestion represents a recommended parameter optimization
type ParameterTuningSuggestion struct {
	ID                      int64     `db:"id" json:"id"`
	QueryHash               int64     `db:"query_hash" json:"query_hash"`
	FingerprintHash         *int64    `db:"fingerprint_hash" json:"fingerprint_hash,omitempty"`
	ParameterName           string    `db:"parameter_name" json:"parameter_name"` // work_mem, sort_mem, limit, batch_size
	CurrentValue            *string   `db:"current_value" json:"current_value,omitempty"`
	RecommendedValue        string    `db:"recommended_value" json:"recommended_value"`
	Reasoning               *string   `db:"reasoning" json:"reasoning,omitempty"`
	EstimatedImprovementPct float64   `db:"estimated_improvement_percent" json:"estimated_improvement_percent"`
	ConfidenceScore         float64   `db:"confidence_score" json:"confidence_score"`
	CreatedAt               time.Time `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time `db:"updated_at" json:"updated_at"`
}

// OptimizationRecommendation represents an aggregated recommendation with ROI scoring
type OptimizationRecommendation struct {
	ID                       int64     `db:"id" json:"id"`
	QueryHash                int64     `db:"query_hash" json:"query_hash"`
	SourceType               string    `db:"source_type" json:"source_type"` // index, rewrite, parameter, workload
	SourceID                 *int64    `db:"source_id" json:"source_id,omitempty"`
	RecommendationText       string    `db:"recommendation_text" json:"recommendation_text"`
	DetailedExplanation      *string   `db:"detailed_explanation" json:"detailed_explanation,omitempty"`
	EstimatedImprovementPct  float64   `db:"estimated_improvement_percent" json:"estimated_improvement_percent"`
	ConfidenceScore          float64   `db:"confidence_score" json:"confidence_score"`
	UrgencyScore             float64   `db:"urgency_score" json:"urgency_score"`
	ROIScore                 float64   `db:"roi_score" json:"roi_score"`                                           // confidence × improvement × urgency
	ImplementationComplexity *string   `db:"implementation_complexity" json:"implementation_complexity,omitempty"` // low, medium, high
	DismissalReason          *string   `db:"dismissal_reason" json:"dismissal_reason,omitempty"`
	IsDismissed              bool      `db:"is_dismissed" json:"is_dismissed"`
	CreatedAt                time.Time `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time `db:"updated_at" json:"updated_at"`
}

// OptimizationImplementation tracks when a recommendation is applied and measures results
type OptimizationImplementation struct {
	ID                       int64                  `db:"id" json:"id"`
	RecommendationID         int64                  `db:"recommendation_id" json:"recommendation_id"`
	QueryHash                int64                  `db:"query_hash" json:"query_hash"`
	ImplementationTimestamp  time.Time              `db:"implementation_timestamp" json:"implementation_timestamp"`
	ImplementationNotes      *string                `db:"implementation_notes" json:"implementation_notes,omitempty"`
	PreOptimizationStats     map[string]interface{} `db:"pre_optimization_stats" json:"pre_optimization_stats,omitempty"`   // mean_time, calls, total_time
	PostOptimizationStats    map[string]interface{} `db:"post_optimization_stats" json:"post_optimization_stats,omitempty"` // measured after implementation
	ActualImprovementPct     *float64               `db:"actual_improvement_percent" json:"actual_improvement_percent,omitempty"`
	ActualImprovementSeconds *float64               `db:"actual_improvement_seconds" json:"actual_improvement_seconds,omitempty"`
	Status                   string                 `db:"status" json:"status"` // pending, implemented, reverted, failed
	ErrorMessage             *string                `db:"error_message" json:"error_message,omitempty"`
	MeasuredAt               *time.Time             `db:"measured_at" json:"measured_at,omitempty"`
}

// QueryPerformanceModel represents a trained ML model for performance prediction
type QueryPerformanceModel struct {
	ID                 int64                  `db:"id" json:"id"`
	ModelType          string                 `db:"model_type" json:"model_type"` // linear_regression, decision_tree, random_forest, xgboost
	ModelName          *string                `db:"model_name" json:"model_name,omitempty"`
	DatabaseName       *string                `db:"database_name" json:"database_name,omitempty"`
	ModelBinary        []byte                 `db:"model_binary" json:"-"`                  // Serialized model (not exposed in JSON)
	ModelJSON          map[string]interface{} `db:"model_json" json:"model_json,omitempty"` // JSON representation
	FeatureNames       []string               `db:"feature_names" json:"feature_names"`
	TrainingSampleSize *int                   `db:"training_sample_size" json:"training_sample_size,omitempty"`
	RSquared           *float64               `db:"r_squared" json:"r_squared,omitempty"`
	RMSE               *float64               `db:"rmse" json:"rmse,omitempty"`
	MAE                *float64               `db:"mae" json:"mae,omitempty"`
	FeatureImportance  map[string]interface{} `db:"feature_importance" json:"feature_importance,omitempty"`
	TrainingTimestamp  time.Time              `db:"training_timestamp" json:"training_timestamp"`
	LastUpdated        time.Time              `db:"last_updated" json:"last_updated"`
	Version            int                    `db:"version" json:"version"`
	IsActive           bool                   `db:"is_active" json:"is_active"`
	Metrics            map[string]interface{} `db:"metrics" json:"metrics,omitempty"` // Additional metrics
}

// PerformancePrediction represents a performance prediction for a query
type PerformancePrediction struct {
	QueryHash            int64                  `json:"query_hash"`
	PredictedExecutionMs float64                `json:"predicted_execution_time_ms"`
	ConfidenceScore      float64                `json:"confidence"`
	PredictionRange      PredictionRange        `json:"range"`
	ModelVersion         *string                `json:"model_version,omitempty"`
	Features             map[string]interface{} `json:"features,omitempty"`
	Timestamp            time.Time              `json:"timestamp"`
}

// PredictionRange represents the min/max range for a prediction
type PredictionRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// OptimizationResult represents the result view of optimization impact
type OptimizationResult struct {
	ImplementationID     int64      `json:"implementation_id"`
	RecommendationID     int64      `json:"recommendation_id"`
	QueryHash            int64      `json:"query_hash"`
	RecommendationText   string     `json:"recommendation_text"`
	EstimatedImprovement float64    `json:"estimated_improvement"`
	ActualImprovement    *float64   `json:"actual_improvement,omitempty"`
	PredictionErrorPct   *float64   `json:"prediction_error_percent,omitempty"`
	Status               string     `json:"status"`
	ImplementationTime   time.Time  `json:"implementation_timestamp"`
	MeasuredAt           *time.Time `json:"measured_at,omitempty"`
	TimeToMeasurement    *string    `json:"time_to_measurement,omitempty"`
	ConfidenceScore      float64    `json:"confidence_score"`
	ActualImprovementSec *float64   `json:"actual_improvement_seconds,omitempty"`
}

// WorkloadPatternSummary represents a summary of patterns
type WorkloadPatternSummary struct {
	DatabaseName    string        `json:"database_name"`
	PatternType     string        `json:"pattern_type"`
	Occurrences     int           `json:"occurrences"`
	AvgConfidence   float64       `json:"avg_confidence"`
	LatestDetection time.Time     `json:"latest_detection"`
	AllMetadata     []interface{} `json:"all_metadata"`
}

// ============================================================================
// REGISTRATION SECRET MODELS
// ============================================================================

// RegistrationSecret represents a secret for collector self-registration
type RegistrationSecret struct {
	ID                 string     `db:"id" json:"id"`
	Name               string     `db:"name" json:"name"`
	SecretValue        string     `db:"secret_value" json:"secret_value,omitempty"` // Only returned on creation
	Description        string     `db:"description" json:"description,omitempty"`
	Active             bool       `db:"active" json:"active"`
	CreatedBy          *int       `db:"created_by" json:"created_by,omitempty"`
	CreatedByUsername  string     `json:"created_by_username,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
	ExpiresAt          *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	TotalRegistrations int        `db:"total_registrations" json:"total_registrations"`
	LastUsedAt         *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
}

// CreateRegistrationSecretRequest is the request to create a new registration secret
type CreateRegistrationSecretRequest struct {
	Name        string     `json:"name" binding:"required,min=1,max=255"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// UpdateRegistrationSecretRequest is the request to update a registration secret
type UpdateRegistrationSecretRequest struct {
	Name        string `json:"name" binding:"min=1,max=255"`
	Description string `json:"description"`
	Active      *bool  `json:"active,omitempty"`
}

// CreateRegistrationSecretResponse is the response when creating a registration secret
type CreateRegistrationSecretResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SecretValue string    `json:"secret_value"` // Only returned on creation
	Description string    `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	Message     string    `json:"message"`
}
