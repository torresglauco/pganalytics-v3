package query_performance

import (
	"context"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"go.uber.org/zap"
)

// Store interface for query performance data
type Store interface {
	GetSlowQueries(ctx context.Context, databaseID int, limit int) ([]storage.SlowQuery, error)
	GetQueryTimeline(ctx context.Context, queryHash string, since time.Time) ([]storage.QueryTimelinePoint, error)
	GetIndexStats(ctx context.Context, databaseID int) ([]storage.IndexUsageStats, error)
}

// Service provides query performance analysis
type Service struct {
	store  Store
	logger *zap.Logger
}

// NewService creates a new query performance service
func NewService(store Store, logger *zap.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

// NewServiceWithStore creates a new service with a custom store implementation
func NewServiceWithStore(store Store, logger *zap.Logger) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

// SlowQueriesResponse is the API response for slow queries
type SlowQueriesResponse struct {
	Queries    []storage.SlowQuery `json:"queries"`
	Total      int                 `json:"total"`
	Limit      int                 `json:"limit"`
	SortedBy   string              `json:"sorted_by"`
	DatabaseID int                 `json:"database_id"`
}

// QueryTimelineResponse is the API response for query timeline
type QueryTimelineResponse struct {
	QueryHash  string                       `json:"query_hash"`
	TimeRange  string                       `json:"time_range"`
	DataPoints []storage.QueryTimelinePoint `json:"data_points"`
	Statistics TimelineStatistics           `json:"statistics"`
}

// TimelineStatistics contains aggregated timeline statistics
type TimelineStatistics struct {
	AvgDuration float64 `json:"avg_duration_ms"`
	MaxDuration float64 `json:"max_duration_ms"`
	MinDuration float64 `json:"min_duration_ms"`
	TotalCalls  int64   `json:"total_calls"`
}

// IndexStatsResponse is the API response for index statistics
type IndexStatsResponse struct {
	Indexes      []IndexWithCategory `json:"indexes"`
	TotalIndexes int                 `json:"total_indexes"`
	UnusedCount  int                 `json:"unused_count"`
	DatabaseID   int                 `json:"database_id"`
}

// IndexWithCategory adds a usage category to index stats
type IndexWithCategory struct {
	storage.IndexUsageStats
	Category string `json:"category"` // "unused", "low", "normal", "high"
}

// GetSlowQueries retrieves top slow queries for a database
func (s *Service) GetSlowQueries(ctx context.Context, databaseID, limit int) (*SlowQueriesResponse, error) {
	queries, err := s.store.GetSlowQueries(ctx, databaseID, limit)
	if err != nil {
		s.logger.Error("Failed to get slow queries",
			zap.Error(err),
			zap.Int("database_id", databaseID),
		)
		return nil, err
	}

	return &SlowQueriesResponse{
		Queries:    queries,
		Total:      len(queries),
		Limit:      limit,
		SortedBy:   "mean_time",
		DatabaseID: databaseID,
	}, nil
}

// GetQueryTimeline retrieves historical performance for a query
func (s *Service) GetQueryTimeline(ctx context.Context, queryHash string, hours int) (*QueryTimelineResponse, error) {
	if hours <= 0 {
		hours = 24
	}
	if hours > 8760 { // Max 1 year
		hours = 8760
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	points, err := s.store.GetQueryTimeline(ctx, queryHash, since)
	if err != nil {
		s.logger.Error("Failed to get query timeline",
			zap.Error(err),
			zap.String("query_hash", queryHash),
		)
		return nil, err
	}

	// Calculate statistics
	var stats TimelineStatistics
	if len(points) > 0 {
		var totalDuration float64
		stats.MaxDuration = points[0].MaxDuration
		stats.MinDuration = points[0].MaxDuration

		for _, p := range points {
			totalDuration += p.AvgDuration
			stats.TotalCalls += p.Executions
			if p.MaxDuration > stats.MaxDuration {
				stats.MaxDuration = p.MaxDuration
			}
			if p.MaxDuration < stats.MinDuration {
				stats.MinDuration = p.MaxDuration
			}
		}
		stats.AvgDuration = totalDuration / float64(len(points))
	}

	return &QueryTimelineResponse{
		QueryHash:  queryHash,
		TimeRange:  formatTimeRange(hours),
		DataPoints: points,
		Statistics: stats,
	}, nil
}

// GetIndexStats retrieves index usage statistics
func (s *Service) GetIndexStats(ctx context.Context, databaseID int) (*IndexStatsResponse, error) {
	stats, err := s.store.GetIndexStats(ctx, databaseID)
	if err != nil {
		s.logger.Error("Failed to get index stats",
			zap.Error(err),
			zap.Int("database_id", databaseID),
		)
		return nil, err
	}

	var result []IndexWithCategory
	var unusedCount int

	for _, idx := range stats {
		categorized := IndexWithCategory{
			IndexUsageStats: idx,
			Category:        categorizeIndexUsage(idx.IdxScan),
		}
		if categorized.Category == "unused" {
			unusedCount++
		}
		result = append(result, categorized)
	}

	return &IndexStatsResponse{
		Indexes:      result,
		TotalIndexes: len(result),
		UnusedCount:  unusedCount,
		DatabaseID:   databaseID,
	}, nil
}

// categorizeIndexUsage determines usage category based on scan count
func categorizeIndexUsage(scans int64) string {
	switch {
	case scans == 0:
		return "unused"
	case scans < 100:
		return "low"
	case scans < 10000:
		return "normal"
	default:
		return "high"
	}
}

// formatTimeRange converts hours to human-readable string
func formatTimeRange(hours int) string {
	switch {
	case hours <= 24:
		return "24h"
	case hours <= 168:
		return "7d"
	case hours <= 720:
		return "30d"
	default:
		return "1y"
	}
}
