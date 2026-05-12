package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/torresglauco/pganalytics-v3/backend/internal/cache"
	"github.com/torresglauco/pganalytics-v3/backend/internal/timescale"
	"go.uber.org/zap"
)

// DashboardAggregationWorker manages aggregate freshness and cache invalidation
type DashboardAggregationWorker struct {
	db           *timescale.TimescaleDB
	cacheManager *cache.Manager
	logger       *zap.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex
	isRunning    bool
	tickInterval time.Duration
}

// NewDashboardAggregationWorker creates a new aggregation worker
func NewDashboardAggregationWorker(
	db *timescale.TimescaleDB,
	cacheManager *cache.Manager,
	logger *zap.Logger,
) *DashboardAggregationWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &DashboardAggregationWorker{
		db:           db,
		cacheManager: cacheManager,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		tickInterval: 30 * time.Second,
	}
}

// Start begins the aggregation worker
func (w *DashboardAggregationWorker) Start() error {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return fmt.Errorf("worker already running")
	}
	w.isRunning = true
	w.mu.Unlock()

	w.logger.Info("Starting dashboard aggregation worker",
		zap.Duration("interval", w.tickInterval),
	)

	w.wg.Add(1)
	go w.run()

	return nil
}

// Stop gracefully shuts down the worker
func (w *DashboardAggregationWorker) Stop(timeout time.Duration) error {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return fmt.Errorf("worker not running")
	}
	w.mu.Unlock()

	w.logger.Info("Stopping dashboard aggregation worker")
	w.cancel()

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("Dashboard aggregation worker stopped")
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("worker shutdown timeout exceeded")
	}
}

// IsRunning returns whether the worker is currently running
func (w *DashboardAggregationWorker) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isRunning
}

// run is the main worker loop
func (w *DashboardAggregationWorker) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.tickInterval)
	defer ticker.Stop()

	// Initial check on startup
	w.checkAggregateHealth()

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("Dashboard aggregation worker context canceled")
			return

		case <-ticker.C:
			w.checkAggregateHealth()
		}
	}
}

// checkAggregateHealth verifies aggregates are refreshing correctly
func (w *DashboardAggregationWorker) checkAggregateHealth() {
	ctx, cancel := context.WithTimeout(w.ctx, 10*time.Second)
	defer cancel()

	jobs, err := w.db.GetAggregateJobStatus(ctx)
	if err != nil {
		w.logger.Error("Failed to check aggregate job status", zap.Error(err))
		return
	}

	// Handle case where TimescaleDB is not available (nil jobs, nil error)
	if jobs == nil {
		w.logger.Debug("TimescaleDB continuous aggregates not available")
		return
	}

	failedCount := 0
	for _, job := range jobs {
		if job.LastRunStatus != "" && job.LastRunStatus != "Success" {
			w.logger.Warn("Aggregate job may have issues",
				zap.Int("job_id", job.JobID),
				zap.String("job_name", job.JobName),
				zap.String("hypertable", job.Hypertable),
				zap.String("status", job.LastRunStatus),
				zap.Int("total_failures", job.TotalFailures),
			)
			failedCount++
		} else {
			w.logger.Debug("Aggregate job healthy",
				zap.Int("job_id", job.JobID),
				zap.String("job_name", job.JobName),
				zap.String("hypertable", job.Hypertable),
			)
		}
	}

	if len(jobs) > 0 {
		w.logger.Debug("Aggregate health check complete",
			zap.Int("total_jobs", len(jobs)),
			zap.Int("failed_jobs", failedCount),
		)
	}
}

// InvalidateCollectorCache clears cache entries for a specific collector
func (w *DashboardAggregationWorker) InvalidateCollectorCache(collectorID uuid.UUID) {
	w.logger.Info("Invalidating dashboard cache for collector",
		zap.String("collector_id", collectorID.String()),
	)

	if w.cacheManager != nil {
		// Clear response cache entries for this collector
		// The cache middleware will repopulate on next request
		// Note: This is a simplified approach - in production, you might want
		// to clear only specific keys related to this collector
		w.cacheManager.ClearResponseCache(collectorID.String())
	}
}
