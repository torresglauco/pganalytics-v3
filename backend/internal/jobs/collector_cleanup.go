package jobs

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"go.uber.org/zap"
)

// CollectorCleanupJob removes offline collectors and stale data
type CollectorCleanupJob struct {
	db             *storage.PostgresDB
	logger         *zap.Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	mu             sync.RWMutex
	isRunning      bool
	tickInterval   time.Duration
	jitterFactor   float64
	offlineTimeout time.Duration // Time to consider collector offline (default: 7 days)
}

// NewCollectorCleanupJob creates a new collector cleanup job
func NewCollectorCleanupJob(
	db *storage.PostgresDB,
	logger *zap.Logger,
) *CollectorCleanupJob {
	ctx, cancel := context.WithCancel(context.Background())
	return &CollectorCleanupJob{
		db:             db,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
		tickInterval:   24 * time.Hour, // Run daily
		jitterFactor:   0.1,            // 10% randomization
		offlineTimeout: 7 * 24 * time.Hour, // 7 days offline = inactive
		isRunning:      false,
	}
}

// SetOfflineTimeout sets the time after which a collector is considered offline
func (ccj *CollectorCleanupJob) SetOfflineTimeout(duration time.Duration) {
	ccj.mu.Lock()
	defer ccj.mu.Unlock()
	ccj.offlineTimeout = duration
}

// Start begins the collector cleanup job
func (ccj *CollectorCleanupJob) Start() {
	ccj.mu.Lock()
	if ccj.isRunning {
		ccj.mu.Unlock()
		return
	}
	ccj.isRunning = true
	ccj.mu.Unlock()

	ccj.wg.Add(1)
	go ccj.run()
	ccj.logger.Info("Collector cleanup job started", zap.Duration("interval", ccj.tickInterval))
}

// Stop stops the collector cleanup job
func (ccj *CollectorCleanupJob) Stop() {
	ccj.mu.Lock()
	defer ccj.mu.Unlock()

	if !ccj.isRunning {
		return
	}

	ccj.isRunning = false
	ccj.cancel()
	ccj.wg.Wait()
	ccj.logger.Info("Collector cleanup job stopped")
}

// run executes the cleanup logic
func (ccj *CollectorCleanupJob) run() {
	defer ccj.wg.Done()

	// Add initial jitter to stagger job starts
	initialDelay := time.Duration(
		float64(ccj.tickInterval) * ccj.jitterFactor * rand.Float64(),
	)
	timer := time.NewTimer(initialDelay)

	for {
		select {
		case <-ccj.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			// Execute cleanup
			cleanupCtx, cancel := context.WithTimeout(
				context.Background(),
				5*time.Minute,
			)
			err := ccj.cleanup(cleanupCtx)
			cancel()

			if err != nil {
				ccj.logger.Error(
					"Collector cleanup failed",
					zap.Error(err),
					zap.Time("next_run", time.Now().Add(ccj.tickInterval)),
				)
			}

			// Add jitter to next execution
			jitter := time.Duration(
				float64(ccj.tickInterval) * ccj.jitterFactor * (2*rand.Float64() - 1),
			)
			nextInterval := ccj.tickInterval + jitter
			if nextInterval < ccj.tickInterval/2 {
				nextInterval = ccj.tickInterval / 2
			}

			timer.Reset(nextInterval)
		}
	}
}

// cleanup performs the actual cleanup operations
func (ccj *CollectorCleanupJob) cleanup(ctx context.Context) error {
	startTime := time.Now()
	ccj.logger.Info("Starting collector cleanup task")

	// Step 1: Find and mark offline collectors
	offlineCount, err := ccj.markOfflineCollectors(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark offline collectors: %w", err)
	}

	if offlineCount > 0 {
		ccj.logger.Info(
			"Marked collectors as offline",
			zap.Int("count", offlineCount),
		)
	}

	// Step 2: Delete very old offline collectors (older than 30 days)
	deletedCount, err := ccj.deleteOldOfflineCollectors(ctx, 30*24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to delete old collectors: %w", err)
	}

	if deletedCount > 0 {
		ccj.logger.Info(
			"Deleted old offline collectors",
			zap.Int("count", deletedCount),
		)
	}

	// Step 3: Cleanup stale metrics for deleted collectors
	if deletedCount > 0 {
		metricsDeleted, err := ccj.cleanupOrphanMetrics(ctx)
		if err != nil {
			ccj.logger.Warn(
				"Failed to cleanup orphan metrics",
				zap.Error(err),
			)
			// Don't fail the job for this
		} else if metricsDeleted > 0 {
			ccj.logger.Info(
				"Cleaned up orphan metrics",
				zap.Int("count", metricsDeleted),
			)
		}
	}

	duration := time.Since(startTime)
	ccj.logger.Info(
		"Collector cleanup task completed",
		zap.Duration("duration", duration),
		zap.Int("offline_marked", offlineCount),
		zap.Int("deleted", deletedCount),
	)

	return nil
}

// markOfflineCollectors marks collectors as offline if no heartbeat in offlineTimeout
func (ccj *CollectorCleanupJob) markOfflineCollectors(ctx context.Context) (int, error) {
	query := `
		UPDATE collectors
		SET status = 'offline', updated_at = NOW()
		WHERE status != 'offline'
		AND last_heartbeat < NOW() - INTERVAL '1 second' * $1
		AND last_heartbeat IS NOT NULL
	`

	result, err := ccj.db.db.ExecContext(
		ctx,
		query,
		int(ccj.offlineTimeout.Seconds()),
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

// deleteOldOfflineCollectors removes collectors offline for longer than duration
func (ccj *CollectorCleanupJob) deleteOldOfflineCollectors(
	ctx context.Context,
	duration time.Duration,
) (int, error) {
	query := `
		DELETE FROM collectors
		WHERE status = 'offline'
		AND updated_at < NOW() - INTERVAL '1 second' * $1
	`

	result, err := ccj.db.db.ExecContext(
		ctx,
		query,
		int(duration.Seconds()),
	)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

// cleanupOrphanMetrics removes metrics for deleted collectors
func (ccj *CollectorCleanupJob) cleanupOrphanMetrics(ctx context.Context) (int, error) {
	// This query assumes a metrics table with a collector_id foreign key
	query := `
		DELETE FROM metrics
		WHERE collector_id NOT IN (SELECT id FROM collectors)
		AND created_at < NOW() - INTERVAL '7 days'
	`

	result, err := ccj.db.db.ExecContext(ctx, query)
	if err != nil {
		// If table doesn't exist, just log and continue
		return 0, nil
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

// IsRunning returns whether the job is currently running
func (ccj *CollectorCleanupJob) IsRunning() bool {
	ccj.mu.RLock()
	defer ccj.mu.RUnlock()
	return ccj.isRunning
}
