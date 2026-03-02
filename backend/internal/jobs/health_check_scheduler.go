package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"go.uber.org/zap"
)

// HealthCheckScheduler manages periodic health checks for managed instances
type HealthCheckScheduler struct {
	db              *storage.PostgresDB
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.RWMutex
	isRunning       bool
	tickInterval    time.Duration
	jitterFactor    float64
	maxConcurrency  int
	activeTasks     int
}

// NewHealthCheckScheduler creates a new health check scheduler
func NewHealthCheckScheduler(
	db *storage.PostgresDB,
	logger *zap.Logger,
) *HealthCheckScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthCheckScheduler{
		db:             db,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
		tickInterval:   30 * time.Second, // Check every 30 seconds
		jitterFactor:   0.3,              // 30% randomization
		maxConcurrency: 3,                // Max 3 concurrent health checks
		activeTasks:    0,
	}
}

// Start begins the health check scheduler
func (s *HealthCheckScheduler) Start() error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	s.logger.Info("Starting health check scheduler",
		zap.Duration("interval", s.tickInterval),
		zap.Int("max_concurrency", s.maxConcurrency),
	)

	s.wg.Add(1)
	go s.run()

	return nil
}

// Stop gracefully shuts down the scheduler
func (s *HealthCheckScheduler) Stop(timeout time.Duration) error {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("scheduler not running")
	}
	s.mu.Unlock()

	s.logger.Info("Stopping health check scheduler")
	s.cancel()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Health check scheduler stopped")
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("scheduler shutdown timeout exceeded")
	}
}

// run is the main scheduler loop
func (s *HealthCheckScheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	// Semaphore channel to limit concurrency
	semaphore := make(chan struct{}, s.maxConcurrency)

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Health check scheduler context canceled")
			return

		case <-ticker.C:
			// Get all active managed instances
			instances, err := s.db.ListManagedInstancesForHealthCheck(s.ctx)
			if err != nil {
				s.logger.Error("Failed to list managed instances", zap.Error(err))
				continue
			}

			if len(instances) == 0 {
				continue
			}

			// Randomize order to avoid always checking the same instances first
			shuffleInstances(instances)

			// Schedule health checks with randomized delays
			for _, instance := range instances {
				// Skip if instance is paused
				if instance.Status == "paused" {
					continue
				}

				// Calculate randomized delay to spread out requests
				delay := s.calculateRandomizedDelay()

				// Acquire semaphore slot
				semaphore <- struct{}{}

				s.wg.Add(1)
				go func(inst *storage.ManagedInstance, delayDuration time.Duration) {
					defer s.wg.Done()
					defer func() { <-semaphore }()

					// Wait for randomized delay
					select {
					case <-time.After(delayDuration):
					case <-s.ctx.Done():
						return
					}

					// Perform health check
					s.performHealthCheck(inst)
				}(instance, delay)
			}
		}
	}
}

// performHealthCheck tests the connection to a managed instance
func (s *HealthCheckScheduler) performHealthCheck(instance *storage.ManagedInstance) {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	s.logger.Debug("Performing health check",
		zap.Int("instance_id", instance.ID),
		zap.String("instance_name", instance.Name),
		zap.String("endpoint", instance.Endpoint),
	)

	// Test connection with various SSL modes
	var connErr error
	var lastSSLMode string

	sslModes := []string{"require", "prefer", "disable"}
	for _, sslMode := range sslModes {
		lastSSLMode = sslMode
		connErr = testPostgresConnection(ctx, testConnectionConfig{
			Host:     instance.Endpoint,
			Port:     instance.Port,
			User:     instance.MasterUsername,
			Password: instance.MasterPassword,
			Database: "postgres",
			SSLMode:  sslMode,
			Timeout:  5,
		})

		if connErr == nil {
			break // Connection successful
		}
	}

	// Update instance status
	if connErr == nil {
		status := "connected"
		if err := s.db.UpdateManagedInstanceStatus(ctx, instance.ID, status, nil); err != nil {
			s.logger.Error("Failed to update instance status",
				zap.Int("instance_id", instance.ID),
				zap.Error(err),
			)
		} else {
			s.logger.Debug("Health check passed",
				zap.Int("instance_id", instance.ID),
				zap.String("ssl_mode", lastSSLMode),
			)
		}
	} else {
		errorMsg := connErr.Error()
		if err := s.db.UpdateManagedInstanceStatus(ctx, instance.ID, "error", &errorMsg); err != nil {
			s.logger.Error("Failed to update instance error status",
				zap.Int("instance_id", instance.ID),
				zap.Error(err),
			)
		} else {
			s.logger.Debug("Health check failed",
				zap.Int("instance_id", instance.ID),
				zap.String("error", errorMsg),
			)
		}
	}
}

// calculateRandomizedDelay returns a random delay based on jitter factor
func (s *HealthCheckScheduler) calculateRandomizedDelay() time.Duration {
	// Calculate jitter range: 0 to (tickInterval * jitterFactor)
	maxJitterMs := int64(s.tickInterval.Milliseconds() * int64(math.Floor(s.jitterFactor*100)) / 100)
	jitterMs := rand.Int63n(maxJitterMs + 1)
	return time.Duration(jitterMs) * time.Millisecond
}

// IsRunning returns whether the scheduler is currently running
func (s *HealthCheckScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetActiveTaskCount returns the number of active health checks
func (s *HealthCheckScheduler) GetActiveTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeTasks
}

// shuffleInstances randomizes the order of managed instances
func shuffleInstances(instances []*storage.ManagedInstance) {
	for i := len(instances) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		instances[i], instances[j] = instances[j], instances[i]
	}
}

// testConnectionConfig holds PostgreSQL connection parameters
type testConnectionConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	Timeout  int
}

// testPostgresConnection attempts to connect to a PostgreSQL instance
func testPostgresConnection(ctx context.Context, cfg testConnectionConfig) error {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode, cfg.Timeout,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}
