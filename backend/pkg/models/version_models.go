package models

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// POSTGRESQL VERSION MODELS
// ============================================================================

// PostgreSQLVersion represents a PostgreSQL version with lifecycle information
type PostgreSQLVersion struct {
	Major       int       `json:"major" db:"major"`               // Major version (e.g., 16)
	Minor       int       `json:"minor" db:"minor"`               // Minor version (e.g., 2)
	FullVersion string    `json:"full_version" db:"full_version"` // Full version string (e.g., "16.2")
	ReleaseDate time.Time `json:"release_date" db:"release_date"` // Release date
	EOLDate     time.Time `json:"eol_date" db:"eol_date"`         // End of Life date
	IsSupported bool      `json:"is_supported" db:"is_supported"` // Currently supported by community
}

// VersionCapabilities represents feature capabilities for a PostgreSQL version
type VersionCapabilities struct {
	Version               PostgreSQLVersion `json:"version"`
	HasWriteLagColumns    bool              `json:"has_write_lag_columns"`    // PG 13+
	HasWalReceiver        bool              `json:"has_wal_receiver"`         // PG 9.6+
	HasLogicalReplication bool              `json:"has_logical_replication"`  // PG 10+
	HasPublication        bool              `json:"has_publication"`          // PG 10+
	HasStandbySignal      bool              `json:"has_standby_signal"`       // PG 12+
	HasPgStatWal          bool              `json:"has_pg_stat_wal"`          // PG 14+
	HasPgStatSubscription bool              `json:"has_pg_stat_subscription"` // PG 10+
	MinQueryVersion       string            `json:"min_query_version"`        // Minimum version for query compatibility
}

// VersionFeature represents a PostgreSQL feature with version requirements
type VersionFeature struct {
	FeatureName   string `json:"feature_name" db:"feature_name"`     // Feature name
	MinVersion    int    `json:"min_version" db:"min_version"`       // Minimum major version required
	Description   string `json:"description" db:"description"`       // Feature description
	QueryTemplate string `json:"query_template" db:"query_template"` // Version-specific query template
}

// ============================================================================
// COLLECTOR MODE CONFIGURATION
// ============================================================================

// CollectorModeConfig represents the deployment mode configuration for a collector
type CollectorModeConfig struct {
	CollectorID    uuid.UUID `json:"collector_id" db:"collector_id"`
	Mode           string    `json:"mode" db:"mode"`                       // "decentralized" or "centralized"
	ConnectionType string    `json:"connection_type" db:"connection_type"` // "unix_socket" or "tcp"
	UseTLS         bool      `json:"use_tls" db:"use_tls"`                 // TLS enabled
	TLSConfig      TLSConfig `json:"tls_config"`                           // TLS configuration
}

// TLSConfig represents TLS configuration for collector connections
type TLSConfig struct {
	CertFile           string `json:"cert_file" db:"cert_file"`                       // Client certificate file
	KeyFile            string `json:"key_file" db:"key_file"`                         // Client key file
	CAFile             string `json:"ca_file" db:"ca_file"`                           // CA certificate file
	ServerName         string `json:"server_name" db:"server_name"`                   // Server name for SNI
	InsecureSkipVerify bool   `json:"insecure_skip_verify" db:"insecure_skip_verify"` // Skip TLS verification (dev only)
}

// ============================================================================
// VERSION INFO RESPONSE
// ============================================================================

// VersionInfoResponse represents the response for version information queries
type VersionInfoResponse struct {
	CollectorID  uuid.UUID           `json:"collector_id"`
	Version      PostgreSQLVersion   `json:"version"`
	Capabilities VersionCapabilities `json:"capabilities"`
	Mode         CollectorModeConfig `json:"mode"`
	Timestamp    time.Time           `json:"timestamp"`
}

// SupportedVersionsResponse represents the response for listing supported PostgreSQL versions
type SupportedVersionsResponse struct {
	Versions []PostgreSQLVersion `json:"versions"`
	Count    int                 `json:"count"`
}

// CollectorModeResponse represents the response for collector mode queries
type CollectorModeResponse struct {
	CollectorID    uuid.UUID `json:"collector_id"`
	Mode           string    `json:"mode"`
	ConnectionType string    `json:"connection_type"`
	UseTLS         bool      `json:"use_tls"`
	TLSEnabled     bool      `json:"tls_enabled"`
}
