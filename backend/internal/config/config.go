package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port        int
	Environment string
	LogLevel    string

	// Databases
	DatabaseURL  string
	TimescaleURL string

	// JWT
	JWTSecret            string
	JWTExpiration        time.Duration
	JWTRefreshExpiration time.Duration

	// Security
	RegistrationSecret string

	// TLS
	TLSCertPath string
	TLSKeyPath  string
	TLSEnabled  bool

	// Timeouts
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration

	// ML Service
	MLServiceURL     string
	MLServiceTimeout time.Duration
	MLServiceEnabled bool

	// Caching
	CacheEnabled         bool
	CacheMaxSize         int
	FeatureCacheTTL      time.Duration
	PredictionCacheTTL   time.Duration
	QueryResultsCacheTTL time.Duration

	// Connection Pooling
	MaxDatabaseConns     int
	MaxIdleDatabaseConns int
	MaxHTTPConns         int
	MaxHTTPConnsPerHost  int

	// Retry Policy
	RetryMaxAttempts       int
	RetryBackoffMultiplier float64
	RetryInitialBackoff    time.Duration

	// Enterprise Authentication
	// LDAP Configuration
	LDAPEnabled         bool
	LDAPServerURL       string
	LDAPBindDN          string
	LDAPBindPassword    string
	LDAPUserSearchBase  string
	LDAPGroupSearchBase string
	LDAPGroupToRoleJSON string // JSON map of LDAP groups to roles

	// SAML Configuration
	SAMLEnabled        bool
	SAMLCertPath       string
	SAMLKeyPath        string
	SAMLIDPMetadataURL string
	SAMLEntityID       string

	// OAuth Configuration
	OAuthEnabled       bool
	OAuthProvidersJSON string // JSON config for OAuth providers (Google, Azure, GitHub, custom OIDC)

	// MFA Configuration
	MFAEnabled         bool
	MFADefaultType     string // totp, sms, email
	MFAToTPIssuer      string // Issuer name for TOTP (displayed in authenticator apps)
	MFASMSProvider     string // twilio, sns, custom
	MFABackupCodeCount int    // Number of backup codes to generate

	// Encryption Configuration
	EncryptionEnabled         bool
	EncryptionKeyBackend      string // local, aws, vault, gcp
	EncryptionAlgorithm       string // aes-256-gcm
	EncryptionKeyRotationDays int    // Default 90 days

	// AWS Encryption Backend
	AWSSecretsManagerARN string

	// Vault Encryption Backend
	VaultAddr  string
	VaultToken string
	VaultPath  string

	// GCP Encryption Backend
	GCPKMSKeyName string

	// Audit Configuration
	AuditEnabled       bool
	AuditRetentionDays int    // Default 365
	AuditArchivePath   string // S3 path or filesystem path
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:                   getIntEnv("PORT", 8080),
		Environment:            getEnv("ENVIRONMENT", "development"),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		TimescaleURL:           getEnv("TIMESCALE_URL", ""),
		JWTSecret:              getEnv("JWT_SECRET", "default-insecure-secret"),
		JWTExpiration:          time.Duration(getIntEnv("JWT_EXPIRATION", 900)) * time.Second,
		JWTRefreshExpiration:   time.Duration(getIntEnv("JWT_REFRESH_EXPIRATION", 86400)) * time.Second,
		RegistrationSecret:     getEnv("REGISTRATION_SECRET", "change-me-in-production"),
		TLSCertPath:            getEnv("TLS_CERT", ""),
		TLSKeyPath:             getEnv("TLS_KEY", ""),
		TLSEnabled:             getBoolEnv("TLS_ENABLED", false),
		RequestTimeout:         time.Duration(getIntEnv("REQUEST_TIMEOUT", 30)) * time.Second,
		ShutdownTimeout:        time.Duration(getIntEnv("SHUTDOWN_TIMEOUT", 10)) * time.Second,
		MLServiceURL:           getEnv("ML_SERVICE_URL", "http://localhost:8081"),
		MLServiceTimeout:       time.Duration(getIntEnv("ML_SERVICE_TIMEOUT", 5)) * time.Second,
		MLServiceEnabled:       getBoolEnv("ML_SERVICE_ENABLED", true),
		CacheEnabled:           getBoolEnv("CACHE_ENABLED", true),
		CacheMaxSize:           getIntEnv("CACHE_MAX_SIZE", 10000),
		FeatureCacheTTL:        time.Duration(getIntEnv("FEATURE_CACHE_TTL", 900)) * time.Second,
		PredictionCacheTTL:     time.Duration(getIntEnv("PREDICTION_CACHE_TTL", 300)) * time.Second,
		QueryResultsCacheTTL:   time.Duration(getIntEnv("QUERY_RESULTS_CACHE_TTL", 600)) * time.Second,
		MaxDatabaseConns:       getIntEnv("MAX_DATABASE_CONNS", 50),
		MaxIdleDatabaseConns:   getIntEnv("MAX_IDLE_DATABASE_CONNS", 15),
		MaxHTTPConns:           getIntEnv("MAX_HTTP_CONNS", 10),
		MaxHTTPConnsPerHost:    getIntEnv("MAX_HTTP_CONNS_PER_HOST", 5),
		RetryMaxAttempts:       getIntEnv("RETRY_MAX_ATTEMPTS", 3),
		RetryBackoffMultiplier: getFloatEnv("RETRY_BACKOFF_MULTIPLIER", 2.0),
		RetryInitialBackoff:    time.Duration(getIntEnv("RETRY_INITIAL_BACKOFF", 100)) * time.Millisecond,
		// Enterprise Authentication
		LDAPEnabled:               getBoolEnv("LDAP_ENABLED", false),
		LDAPServerURL:             getEnv("LDAP_SERVER_URL", ""),
		LDAPBindDN:                getEnv("LDAP_BIND_DN", ""),
		LDAPBindPassword:          getEnv("LDAP_BIND_PASSWORD", ""),
		LDAPUserSearchBase:        getEnv("LDAP_USER_SEARCH_BASE", ""),
		LDAPGroupSearchBase:       getEnv("LDAP_GROUP_SEARCH_BASE", ""),
		LDAPGroupToRoleJSON:       getEnv("LDAP_GROUP_TO_ROLE_MAPPING", "{}"),
		SAMLEnabled:               getBoolEnv("SAML_ENABLED", false),
		SAMLCertPath:              getEnv("SAML_CERT_PATH", ""),
		SAMLKeyPath:               getEnv("SAML_KEY_PATH", ""),
		SAMLIDPMetadataURL:        getEnv("SAML_IDP_METADATA_URL", ""),
		SAMLEntityID:              getEnv("SAML_ENTITY_ID", ""),
		OAuthEnabled:              getBoolEnv("OAUTH_ENABLED", false),
		OAuthProvidersJSON:        getEnv("OAUTH_PROVIDERS", "{}"),
		MFAEnabled:                getBoolEnv("MFA_ENABLED", false),
		MFADefaultType:            getEnv("MFA_DEFAULT_TYPE", "totp"),
		MFAToTPIssuer:             getEnv("MFA_TOTP_ISSUER", "pgAnalytics"),
		MFASMSProvider:            getEnv("MFA_SMS_PROVIDER", "twilio"),
		MFABackupCodeCount:        getIntEnv("MFA_BACKUP_CODE_COUNT", 8),
		EncryptionEnabled:         getBoolEnv("ENCRYPTION_ENABLED", false),
		EncryptionKeyBackend:      getEnv("ENCRYPTION_KEY_BACKEND", "local"),
		EncryptionAlgorithm:       getEnv("ENCRYPTION_ALGORITHM", "aes-256-gcm"),
		EncryptionKeyRotationDays: getIntEnv("ENCRYPTION_KEY_ROTATION_DAYS", 90),
		AWSSecretsManagerARN:      getEnv("AWS_SECRETS_MANAGER_ARN", ""),
		VaultAddr:                 getEnv("VAULT_ADDR", ""),
		VaultToken:                getEnv("VAULT_TOKEN", ""),
		VaultPath:                 getEnv("VAULT_PATH", "/secret/pganalytics"),
		GCPKMSKeyName:             getEnv("GCP_KMS_KEY_NAME", ""),
		AuditEnabled:              getBoolEnv("AUDIT_ENABLED", true),
		AuditRetentionDays:        getIntEnv("AUDIT_RETENTION_DAYS", 365),
		AuditArchivePath:          getEnv("AUDIT_ARCHIVE_PATH", ""),
	}

	return cfg
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return NewConfigError("DATABASE_URL is required")
	}
	if c.TimescaleURL == "" {
		return NewConfigError("TIMESCALE_URL is required")
	}
	if c.JWTSecret == "default-insecure-secret" && c.Environment == "production" {
		return NewConfigError("JWT_SECRET must be set in production")
	}
	if c.RegistrationSecret == "change-me-in-production" && c.Environment == "production" {
		return NewConfigError("REGISTRATION_SECRET must be set in production")
	}
	if c.TLSEnabled && (c.TLSCertPath == "" || c.TLSKeyPath == "") {
		return NewConfigError("TLS_CERT and TLS_KEY must be set when TLS_ENABLED is true")
	}
	return nil
}

// IsProduction checks if running in production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment checks if running in development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// ConfigError represents a configuration error
type ConfigError struct {
	message string
}

func NewConfigError(message string) *ConfigError {
	return &ConfigError{message: message}
}

func (e *ConfigError) Error() string {
	return e.message
}
