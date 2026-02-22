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
	CacheEnabled          bool
	CacheMaxSize          int
	FeatureCacheTTL       time.Duration
	PredictionCacheTTL    time.Duration
	QueryResultsCacheTTL  time.Duration

	// Connection Pooling
	MaxDatabaseConns    int
	MaxIdleDatabaseConns int
	MaxHTTPConns        int
	MaxHTTPConnsPerHost int

	// Retry Policy
	RetryMaxAttempts      int
	RetryBackoffMultiplier float64
	RetryInitialBackoff   time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:                 getIntEnv("PORT", 8080),
		Environment:          getEnv("ENVIRONMENT", "development"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		TimescaleURL:         getEnv("TIMESCALE_URL", ""),
		JWTSecret:            getEnv("JWT_SECRET", "default-insecure-secret"),
		JWTExpiration:        time.Duration(getIntEnv("JWT_EXPIRATION", 900)) * time.Second,
		JWTRefreshExpiration: time.Duration(getIntEnv("JWT_REFRESH_EXPIRATION", 86400)) * time.Second,
		TLSCertPath:          getEnv("TLS_CERT", ""),
		TLSKeyPath:           getEnv("TLS_KEY", ""),
		TLSEnabled:           getBoolEnv("TLS_ENABLED", false),
		RequestTimeout:       time.Duration(getIntEnv("REQUEST_TIMEOUT", 30)) * time.Second,
		ShutdownTimeout:      time.Duration(getIntEnv("SHUTDOWN_TIMEOUT", 10)) * time.Second,
		MLServiceURL:         getEnv("ML_SERVICE_URL", "http://localhost:8081"),
		MLServiceTimeout:     time.Duration(getIntEnv("ML_SERVICE_TIMEOUT", 5)) * time.Second,
		MLServiceEnabled:     getBoolEnv("ML_SERVICE_ENABLED", true),
		CacheEnabled:         getBoolEnv("CACHE_ENABLED", true),
		CacheMaxSize:         getIntEnv("CACHE_MAX_SIZE", 10000),
		FeatureCacheTTL:      time.Duration(getIntEnv("FEATURE_CACHE_TTL", 900)) * time.Second,
		PredictionCacheTTL:   time.Duration(getIntEnv("PREDICTION_CACHE_TTL", 300)) * time.Second,
		QueryResultsCacheTTL: time.Duration(getIntEnv("QUERY_RESULTS_CACHE_TTL", 600)) * time.Second,
		MaxDatabaseConns:     getIntEnv("MAX_DATABASE_CONNS", 50),
		MaxIdleDatabaseConns: getIntEnv("MAX_IDLE_DATABASE_CONNS", 15),
		MaxHTTPConns:         getIntEnv("MAX_HTTP_CONNS", 10),
		MaxHTTPConnsPerHost:  getIntEnv("MAX_HTTP_CONNS_PER_HOST", 5),
		RetryMaxAttempts:     getIntEnv("RETRY_MAX_ATTEMPTS", 3),
		RetryBackoffMultiplier: getFloatEnv("RETRY_BACKOFF_MULTIPLIER", 2.0),
		RetryInitialBackoff:  time.Duration(getIntEnv("RETRY_INITIAL_BACKOFF", 100)) * time.Millisecond,
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
