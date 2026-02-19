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
	ID        int       `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	FullName  string    `db:"full_name" json:"full_name,omitempty"`
	Role      string    `db:"role" json:"role"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	LastLogin *time.Time `db:"last_login" json:"last_login,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
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
	ID         int       `db:"id" json:"id"`
	CollectorID uuid.UUID `db:"collector_id" json:"collector_id"`
	Version    int       `db:"version" json:"version"`
	Config     string    `db:"config" json:"config"` // JSON string
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedBy  *int      `db:"updated_by" json:"updated_by,omitempty"`
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
	ID                    int        `db:"id" json:"id"`
	ServerID              int        `db:"server_id" json:"server_id"`
	Name                  string     `db:"name" json:"name"`
	Version               *string    `db:"version" json:"version,omitempty"`
	Port                  int        `db:"port" json:"port"`
	ConnectionString      *string    `db:"connection_string" json:"connection_string,omitempty"`
	MaintenanceDatabase   string     `db:"maintenance_database" json:"maintenance_database"`
	MonitoringRole        string     `db:"monitoring_role" json:"monitoring_role"`
	IsActive              bool       `db:"is_active" json:"is_active"`
	LastConnected         *time.Time `db:"last_connected" json:"last_connected,omitempty"`
	ReplicationRole       *string    `db:"replication_role" json:"replication_role,omitempty"` // primary, standby, unknown
	CreatedAt             time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at" json:"updated_at"`
}

// Database represents a database within a PostgreSQL instance
type Database struct {
	ID            int        `db:"id" json:"id"`
	InstanceID    int        `db:"instance_id" json:"instance_id"`
	Name          string     `db:"name" json:"name"`
	Owner         *string    `db:"owner" json:"owner,omitempty"`
	SizeBytes     *int64     `db:"size_bytes" json:"size_bytes,omitempty"`
	IsTemplate    bool       `db:"is_template" json:"is_template"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	LastAnalyzed  *time.Time `db:"last_analyzed" json:"last_analyzed,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// AUTHENTICATION MODELS
// ============================================================================

// APIToken represents an authentication token
type APIToken struct {
	ID         int        `db:"id" json:"id"`
	CollectorID *uuid.UUID `db:"collector_id" json:"collector_id,omitempty"`
	UserID     *int       `db:"user_id" json:"user_id,omitempty"`
	TokenHash  string     `db:"token_hash" json:"token_hash"`
	Description string    `db:"description" json:"description,omitempty"`
	LastUsed   *time.Time `db:"last_used" json:"last_used,omitempty"`
	ExpiresAt  *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

// Secret represents an encrypted secret
type Secret struct {
	ID                int       `db:"id" json:"id"`
	Name              string    `db:"name" json:"name"`
	SecretEncrypted   []byte    `db:"secret_encrypted" json:"-"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// ALERT MODELS
// ============================================================================

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID                   int        `db:"id" json:"id"`
	Name                 string     `db:"name" json:"name"`
	Description          string     `db:"description" json:"description,omitempty"`
	MetricType           string     `db:"metric_type" json:"metric_type"`
	ConditionType        string     `db:"condition_type" json:"condition_type"` // threshold, change, anomaly
	ConditionValue       string     `db:"condition_value" json:"condition_value"`
	Severity             string     `db:"severity" json:"severity"` // info, warning, critical
	Enabled              bool       `db:"enabled" json:"enabled"`
	NotificationChannel  string     `db:"notification_channel" json:"notification_channel"`
	EvaluationInterval   int        `db:"evaluation_interval" json:"evaluation_interval"`
	CreatedBy            *int       `db:"created_by" json:"created_by,omitempty"`
	CreatedAt            time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updated_at"`
}

// Alert represents an alert instance
type Alert struct {
	ID           int        `db:"id" json:"id"`
	CollectorID  *uuid.UUID `db:"collector_id" json:"collector_id,omitempty"`
	RuleID       *int       `db:"rule_id" json:"rule_id,omitempty"`
	ServerID     *int       `db:"server_id" json:"server_id,omitempty"`
	DatabaseID   *int       `db:"database_id" json:"database_id,omitempty"`
	MetricType   string     `db:"metric_type" json:"metric_type"`
	MetricValue  string     `db:"metric_value" json:"metric_value"`
	Severity     string     `db:"severity" json:"severity"` // info, warning, critical
	Message      string     `db:"message" json:"message"`
	IsAcknowledged bool     `db:"is_acknowledged" json:"is_acknowledged"`
	AcknowledgedBy *int     `db:"acknowledged_by" json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `db:"acknowledged_at" json:"acknowledged_at,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	ResolvedAt   *time.Time `db:"resolved_at" json:"resolved_at,omitempty"`
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
	Name     string `json:"name" binding:"required"`
	Hostname string `json:"hostname" binding:"required"`
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
	CollectorID    string        `json:"collector_id" binding:"required"`
	Hostname       string        `json:"hostname" binding:"required"`
	Timestamp      time.Time     `json:"timestamp" binding:"required"`
	Version        string        `json:"version,omitempty"`
	MetricsCount   int           `json:"metrics_count"`
	Metrics        []interface{} `json:"metrics"` // Flexible metric structure
}

// MetricsPushResponse represents the response to a metrics push
type MetricsPushResponse struct {
	Status              string    `json:"status"` // success, error
	CollectorID         string    `json:"collector_id"`
	MetricsInserted     int       `json:"metrics_inserted"`
	BytesReceived       int       `json:"bytes_received"`
	ProcessingTimeMs    int64     `json:"processing_time_ms"`
	NextConfigVersion   int       `json:"next_config_version"`
	NextCheckInSeconds  int       `json:"next_check_in_seconds"`
}

// HealthResponse represents system health status
type HealthResponse struct {
	Status       string    `json:"status"` // ok, degraded, error
	Version      string    `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
	Uptime       int64     `json:"uptime"`
	DatabaseOk   bool      `json:"database_ok"`
	TimescaleOk  bool      `json:"timescale_ok"`
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
