package query_performance

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"go.uber.org/zap"
)

// mockQueryPerformanceStore is a mock implementation for testing
type mockQueryPerformanceStore struct {
	slowQueries []storage.SlowQuery
	timeline    []storage.QueryTimelinePoint
	indexStats  []storage.IndexUsageStats
	err         error
}

func (m *mockQueryPerformanceStore) GetSlowQueries(ctx context.Context, databaseID int, limit int) ([]storage.SlowQuery, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.slowQueries, nil
}

func (m *mockQueryPerformanceStore) GetQueryTimeline(ctx context.Context, queryHash string, since time.Time) ([]storage.QueryTimelinePoint, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.timeline, nil
}

func (m *mockQueryPerformanceStore) GetIndexStats(ctx context.Context, databaseID int) ([]storage.IndexUsageStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.indexStats, nil
}

func TestService_GetSlowQueries(t *testing.T) {
	t.Run("returns formatted response with pagination metadata", func(t *testing.T) {
		mockStore := &mockQueryPerformanceStore{
			slowQueries: []storage.SlowQuery{
				{QueryID: 1, QueryHash: "abc123", QueryText: "SELECT * FROM users", Calls: 100, MeanTime: 50.0},
				{QueryID: 2, QueryHash: "def456", QueryText: "SELECT * FROM orders", Calls: 50, MeanTime: 40.0},
			},
		}

		logger := zap.NewNop()
		service := NewServiceWithStore(mockStore, logger)

		ctx := context.Background()
		response, err := service.GetSlowQueries(ctx, 1, 20)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Queries, 2)
		assert.Equal(t, 2, response.Total)
		assert.Equal(t, 20, response.Limit)
		assert.Equal(t, "mean_time", response.SortedBy)
		assert.Equal(t, 1, response.DatabaseID)
	})

	t.Run("handles empty results gracefully", func(t *testing.T) {
		mockStore := &mockQueryPerformanceStore{
			slowQueries: []storage.SlowQuery{},
		}

		logger := zap.NewNop()
		service := NewServiceWithStore(mockStore, logger)

		ctx := context.Background()
		response, err := service.GetSlowQueries(ctx, 1, 20)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.Queries)
		assert.Equal(t, 0, response.Total)
	})
}

func TestService_GetQueryTimeline(t *testing.T) {
	t.Run("aggregates data with statistics", func(t *testing.T) {
		now := time.Now()
		mockStore := &mockQueryPerformanceStore{
			timeline: []storage.QueryTimelinePoint{
				{Timestamp: now.Add(-2 * time.Hour), AvgDuration: 45.5, MaxDuration: 120.0, Executions: 50},
				{Timestamp: now.Add(-1 * time.Hour), AvgDuration: 42.0, MaxDuration: 115.0, Executions: 48},
			},
		}

		logger := zap.NewNop()
		service := NewServiceWithStore(mockStore, logger)

		ctx := context.Background()
		response, err := service.GetQueryTimeline(ctx, "abc123", 24)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "abc123", response.QueryHash)
		assert.Equal(t, "24h", response.TimeRange)
		assert.Len(t, response.DataPoints, 2)
		assert.Equal(t, 43.75, response.Statistics.AvgDuration)
		assert.Equal(t, 120.0, response.Statistics.MaxDuration)
		assert.Equal(t, 115.0, response.Statistics.MinDuration)
		assert.Equal(t, int64(98), response.Statistics.TotalCalls)
	})

	t.Run("handles empty timeline gracefully", func(t *testing.T) {
		mockStore := &mockQueryPerformanceStore{
			timeline: []storage.QueryTimelinePoint{},
		}

		logger := zap.NewNop()
		service := NewServiceWithStore(mockStore, logger)

		ctx := context.Background()
		response, err := service.GetQueryTimeline(ctx, "abc123", 24)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.DataPoints)
		assert.Equal(t, 0.0, response.Statistics.AvgDuration)
	})
}

func TestService_GetIndexStats(t *testing.T) {
	t.Run("categorizes indexes correctly", func(t *testing.T) {
		mockStore := &mockQueryPerformanceStore{
			indexStats: []storage.IndexUsageStats{
				{SchemaName: "public", TableName: "users", IndexName: "users_pkey", IdxScan: 10000, IsUnused: false},
				{SchemaName: "public", TableName: "users", IndexName: "idx_users_email", IdxScan: 50, IsUnused: false},
				{SchemaName: "public", TableName: "orders", IndexName: "idx_orders_created", IdxScan: 0, IsUnused: true},
			},
		}

		logger := zap.NewNop()
		service := NewServiceWithStore(mockStore, logger)

		ctx := context.Background()
		response, err := service.GetIndexStats(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Indexes, 3)
		assert.Equal(t, 3, response.TotalIndexes)
		assert.Equal(t, 1, response.UnusedCount)

		// Verify categories
		assert.Equal(t, "high", response.Indexes[0].Category)   // 10000 scans
		assert.Equal(t, "low", response.Indexes[1].Category)    // 50 scans
		assert.Equal(t, "unused", response.Indexes[2].Category) // 0 scans
	})

	t.Run("handles empty results gracefully", func(t *testing.T) {
		mockStore := &mockQueryPerformanceStore{
			indexStats: []storage.IndexUsageStats{},
		}

		logger := zap.NewNop()
		service := NewServiceWithStore(mockStore, logger)

		ctx := context.Background()
		response, err := service.GetIndexStats(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.Indexes)
		assert.Equal(t, 0, response.TotalIndexes)
		assert.Equal(t, 0, response.UnusedCount)
	})
}
