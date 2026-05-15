package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// TestHostMonitoringAPI_StoreAndGetHostMetrics tests storing and retrieving host metrics
func TestHostMonitoringAPI_StoreAndGetHostMetrics(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForHostMonitoring(t, db)

	// Store host metrics
	err := db.StoreHostMetrics(ctx, []*models.HostMetrics{
		{
			Time:              time.Now(),
			CollectorID:       collectorID,
			CpuUser:           25.5,
			CpuSystem:         10.2,
			CpuIdle:           60.0,
			CpuIowait:         4.3,
			CpuLoad1m:         2.5,
			CpuLoad5m:         2.0,
			CpuLoad15m:        1.8,
			MemoryTotalMb:     16384,
			MemoryFreeMb:      4096,
			MemoryUsedMb:      12288,
			MemoryCachedMb:    2048,
			MemoryUsedPercent: 75.0,
			DiskTotalGb:       500,
			DiskUsedGb:        250,
			DiskFreeGb:        250,
			DiskUsedPercent:   50.0,
			DiskIoReadOps:     1500,
			DiskIoWriteOps:    800,
			NetworkRxBytes:    1073741824, // 1 GB
			NetworkTxBytes:    536870912,  // 512 MB
		},
	})
	require.NoError(t, err, "Should store host metrics without error")

	// Retrieve host metrics
	metrics, err := db.GetHostMetrics(ctx, collectorID, "24h", 100)
	require.NoError(t, err, "Should retrieve host metrics without error")

	assert.NotEmpty(t, metrics, "Should have host metrics")

	metric := metrics[0]
	assert.Equal(t, collectorID, metric.CollectorID)
	assert.Equal(t, 25.5, metric.CpuUser)
	assert.Equal(t, 10.2, metric.CpuSystem)
	assert.Equal(t, 75.0, metric.MemoryUsedPercent)
	assert.Equal(t, 50.0, metric.DiskUsedPercent)
}

// TestHostMonitoringAPI_GetHostStatus_Up tests host status when collector is active
func TestHostMonitoringAPI_GetHostStatus_Up(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForHostMonitoring(t, db)

	// Get host status with default threshold
	status, err := db.GetHostStatus(ctx, collectorID, 300)
	require.NoError(t, err, "Should get host status without error")

	assert.NotNil(t, status)
	assert.Equal(t, collectorID, status.CollectorID)
	assert.Equal(t, "up", status.Status, "Should be up when last_seen is recent")
	assert.True(t, status.IsHealthy)
	assert.Equal(t, 0, status.UnresponsiveForSeconds)
}

// TestHostMonitoringAPI_GetHostStatus_Down tests host status when collector is inactive
func TestHostMonitoringAPI_GetHostStatus_Down(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createInactiveTestCollector(t, db)

	// Get host status with threshold
	status, err := db.GetHostStatus(ctx, collectorID, 300)
	require.NoError(t, err, "Should get host status without error")

	assert.NotNil(t, status)
	assert.Equal(t, "down", status.Status, "Should be down when last_seen is old")
	assert.False(t, status.IsHealthy)
	assert.Greater(t, status.UnresponsiveForSeconds, int64(0))
}

// TestHostMonitoringAPI_GetAllHostStatuses tests retrieving all hosts status
func TestHostMonitoringAPI_GetAllHostStatuses(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Create multiple collectors
	collector1 := createTestCollectorForHostMonitoring(t, db)
	collector2 := createTestCollectorForHostMonitoring(t, db)
	collector3 := createInactiveTestCollector(t, db)

	// Get all host statuses
	statuses, err := db.GetAllHostStatuses(ctx, 300)
	require.NoError(t, err, "Should get all host statuses without error")

	assert.NotEmpty(t, statuses, "Should have host statuses")

	// Verify our test collectors are in the list
	collectorIDs := make(map[uuid.UUID]bool)
	for _, s := range statuses {
		collectorIDs[s.CollectorID] = true
	}

	assert.True(t, collectorIDs[collector1], "Should have collector1 in statuses")
	assert.True(t, collectorIDs[collector2], "Should have collector2 in statuses")
	assert.True(t, collectorIDs[collector3], "Should have collector3 in statuses")
}

// TestHostMonitoringAPI_HostMetricsTimeRange tests time range filtering for host metrics
func TestHostMonitoringAPI_HostMetricsTimeRange(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForHostMonitoring(t, db)

	// Store metrics with different timestamps
	for i := 0; i < 3; i++ {
		err := db.StoreHostMetrics(ctx, []*models.HostMetrics{
			{
				Time:              time.Now().Add(-time.Duration(i) * time.Hour),
				CollectorID:       collectorID,
				CpuUser:           float64(20 + i*5),
				CpuSystem:         10.0,
				CpuIdle:           60.0,
				CpuIowait:         5.0,
				CpuLoad1m:         2.0,
				CpuLoad5m:         2.0,
				CpuLoad15m:        2.0,
				MemoryTotalMb:     16384,
				MemoryFreeMb:      4096,
				MemoryUsedMb:      12288,
				MemoryCachedMb:    2048,
				MemoryUsedPercent: 75.0,
				DiskTotalGb:       500,
				DiskUsedGb:        250,
				DiskFreeGb:        250,
				DiskUsedPercent:   50.0,
				DiskIoReadOps:     1000,
				DiskIoWriteOps:    500,
				NetworkRxBytes:    1073741824,
				NetworkTxBytes:    536870912,
			},
		})
		require.NoError(t, err)
	}

	// Test different time ranges
	testCases := []struct {
		name      string
		timeRange string
		expectMin int
	}{
		{"1h", "1h", 1},
		{"24h", "24h", 1},
		{"7d", "7d", 1},
		{"30d", "30d", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics, err := db.GetHostMetrics(ctx, collectorID, tc.timeRange, 100)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(metrics), tc.expectMin, "Should have at least %d metrics for %s", tc.expectMin, tc.timeRange)
		})
	}
}

// TestHostMonitoringAPI_StoreAndGetHostInventory tests host inventory operations
func TestHostMonitoringAPI_StoreAndGetHostInventory(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForHostMonitoring(t, db)

	// Store host inventory
	err := db.StoreHostInventory(ctx, []*models.HostInventory{
		{
			Time:                    time.Now(),
			CollectorID:             collectorID,
			OsName:                  "Ubuntu",
			OsVersion:               "22.04 LTS",
			OsKernel:                "5.15.0-91-generic",
			CpuCores:                8,
			CpuModel:                "Intel(R) Xeon(R) CPU E5-2670 v3 @ 2.30GHz",
			CpuMHz:                  2300,
			MemoryTotalMb:           32768,
			DiskTotalGb:             1000,
			PostgresVersion:         "16.1",
			PostgresEdition:         "Community",
			PostgresPort:            5432,
			PostgresDataDir:         "/var/lib/postgresql/16/main",
			PostgresMaxConnections:  200,
			PostgresSharedBuffersMb: 8192,
			PostgresWorkMemMb:       64,
		},
	})
	require.NoError(t, err, "Should store host inventory without error")

	// Retrieve host inventory
	inventory, err := db.GetHostInventory(ctx, collectorID)
	require.NoError(t, err, "Should retrieve host inventory without error")

	assert.NotNil(t, inventory)
	assert.Equal(t, collectorID, inventory.CollectorID)
	assert.Equal(t, "Ubuntu", inventory.OsName)
	assert.Equal(t, "22.04 LTS", inventory.OsVersion)
	assert.Equal(t, 8, inventory.CpuCores)
	assert.Equal(t, "16.1", inventory.PostgresVersion)
	assert.Equal(t, 5432, inventory.PostgresPort)
}

// TestHostMonitoringAPI_MultipleHostMetrics tests handling multiple metrics for same collector
func TestHostMonitoringAPI_MultipleHostMetrics(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForHostMonitoring(t, db)

	// Store multiple metrics entries
	for i := 0; i < 10; i++ {
		err := db.StoreHostMetrics(ctx, []*models.HostMetrics{
			{
				Time:              time.Now().Add(-time.Duration(i) * time.Minute),
				CollectorID:       collectorID,
				CpuUser:           float64(20 + i),
				CpuSystem:         10.0,
				CpuIdle:           60.0,
				CpuIowait:         5.0,
				CpuLoad1m:         float64(i),
				CpuLoad5m:         float64(i),
				CpuLoad15m:        float64(i),
				MemoryTotalMb:     16384,
				MemoryFreeMb:      4096,
				MemoryUsedMb:      12288,
				MemoryCachedMb:    2048,
				MemoryUsedPercent: 75.0,
				DiskTotalGb:       500,
				DiskUsedGb:        250,
				DiskFreeGb:        250,
				DiskUsedPercent:   50.0,
				DiskIoReadOps:     int64(1000 + i*100),
				DiskIoWriteOps:    int64(500 + i*50),
				NetworkRxBytes:    int64(1073741824 + i*1048576),
				NetworkTxBytes:    int64(536870912 + i*524288),
			},
		})
		require.NoError(t, err)
	}

	// Retrieve with limit
	metrics, err := db.GetHostMetrics(ctx, collectorID, "24h", 5)
	require.NoError(t, err)

	// Should respect limit
	assert.LessOrEqual(t, len(metrics), 5)
}

// TestHostMonitoringAPI_HostStatusThreshold tests different threshold values
func TestHostMonitoringAPI_HostStatusThreshold(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForHostMonitoring(t, db)

	testCases := []struct {
		name      string
		threshold int
		expectUp  bool
	}{
		{"threshold_300_up", 300, true},
		{"threshold_60_up", 60, true},
		{"threshold_0_up", 0, true}, // Default threshold
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, err := db.GetHostStatus(ctx, collectorID, tc.threshold)
			require.NoError(t, err)

			if tc.expectUp {
				assert.Equal(t, "up", status.Status)
				assert.True(t, status.IsHealthy)
			} else {
				assert.Equal(t, "down", status.Status)
				assert.False(t, status.IsHealthy)
			}
		})
	}
}

// TestHostMonitoringAPI_HostInventoryWithNullFields tests inventory with optional fields
func TestHostMonitoringAPI_HostInventoryWithNullFields(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForHostMonitoring(t, db)

	// Store minimal host inventory
	err := db.StoreHostInventory(ctx, []*models.HostInventory{
		{
			Time:          time.Now(),
			CollectorID:   collectorID,
			OsName:        "CentOS",
			OsVersion:     "7",
			CpuCores:      4,
			MemoryTotalMb: 8192,
			DiskTotalGb:   250,
			// PostgreSQL fields minimal
			PostgresVersion: "14.0",
			PostgresPort:    5432,
		},
	})
	require.NoError(t, err, "Should store minimal inventory without error")

	// Retrieve and verify
	inventory, err := db.GetHostInventory(ctx, collectorID)
	require.NoError(t, err)

	assert.Equal(t, "CentOS", inventory.OsName)
	assert.Equal(t, 4, inventory.CpuCores)
	assert.Equal(t, "14.0", inventory.PostgresVersion)
}

// Helper functions

func createTestCollectorForHostMonitoring(t *testing.T, db *storage.PostgresDB) uuid.UUID {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collectorID := uuid.New()
	hostname := "test-host-" + collectorID.String()[:8]

	query := `
		INSERT INTO collectors (id, hostname, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET hostname = $2, last_seen = $3
	`

	now := time.Now()
	_, err := db.GetDB().ExecContext(ctx, query, collectorID, hostname, now, now, now)
	require.NoError(t, err, "Failed to create test collector for host monitoring")

	return collectorID
}

func createInactiveTestCollector(t *testing.T, db *storage.PostgresDB) uuid.UUID {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collectorID := uuid.New()
	hostname := "test-inactive-host-" + collectorID.String()[:8]

	// Set last_seen to 10 minutes ago (beyond 5-minute threshold)
	oldTime := time.Now().Add(-10 * time.Minute)

	query := `
		INSERT INTO collectors (id, hostname, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET hostname = $2, last_seen = $3
	`

	_, err := db.GetDB().ExecContext(ctx, query, collectorID, hostname, oldTime, oldTime, oldTime)
	require.NoError(t, err, "Failed to create inactive test collector")

	return collectorID
}
