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

// ptrTime returns a pointer to a time.Time value
func ptrTime(t time.Time) *time.Time {
	return &t
}

// TestReplicationAPI_StoreAndGetReplicationMetrics tests storing and retrieving replication metrics
func TestReplicationAPI_StoreAndGetReplicationMetrics(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store replication metrics
	err := db.StoreReplicationMetrics(ctx, []*models.ReplicationStatus{
		{
			CollectorID:     collectorID,
			ServerPID:       12345,
			Usename:         "replicator",
			ApplicationName: "walreceiver",
			State:           "streaming",
			SyncState:       "sync",
			WriteLsn:        "0/6000000",
			FlushLsn:        "0/6000000",
			ReplayLsn:       "0/5F00000",
			WriteLagMs:      5,
			FlushLagMs:      10,
			ReplayLagMs:     50,
			BehindByMb:      0,
			ClientAddr:      "192.168.1.101",
			BackendStart:    ptrTime(time.Now().Add(-30 * time.Minute)),
		},
	})
	require.NoError(t, err, "Should store replication metrics without error")

	// Retrieve replication metrics
	metrics, err := db.GetReplicationMetrics(ctx, collectorID, 100, 0)
	require.NoError(t, err, "Should retrieve replication metrics without error")

	assert.NotNil(t, metrics, "Metrics response should not be nil")
	assert.NotEmpty(t, metrics.ReplicationStatus, "Should have replication status entries")
	assert.Len(t, metrics.ReplicationStatus, 1, "Should have exactly one replication status")

	status := metrics.ReplicationStatus[0]
	assert.Equal(t, collectorID, status.CollectorID)
	assert.Equal(t, 12345, status.ServerPID)
	assert.Equal(t, "replicator", status.Usename)
	assert.Equal(t, "walreceiver", status.ApplicationName)
	assert.Equal(t, "streaming", status.State)
	assert.Equal(t, "sync", status.SyncState)
}

// TestReplicationAPI_GetReplicationTopology tests building replication topology
func TestReplicationAPI_GetReplicationTopology(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store replication status to simulate a primary with standbys
	err := db.StoreReplicationMetrics(ctx, []*models.ReplicationStatus{
		{
			CollectorID:     collectorID,
			ServerPID:       12345,
			Usename:         "replicator",
			ApplicationName: "standby1",
			State:           "streaming",
			SyncState:       "sync",
			WriteLsn:        "0/5000000",
			FlushLsn:        "0/5000000",
			ReplayLsn:       "0/4F00000",
			WriteLagMs:      10,
			FlushLagMs:      15,
			ReplayLagMs:     100,
			BehindByMb:      1,
			ClientAddr:      "192.168.1.100",
			BackendStart:    ptrTime(time.Now().Add(-1 * time.Hour)),
		},
		{
			CollectorID:     collectorID,
			ServerPID:       12346,
			Usename:         "replicator",
			ApplicationName: "standby2",
			State:           "streaming",
			SyncState:       "async",
			WriteLsn:        "0/5000000",
			FlushLsn:        "0/5000000",
			ReplayLsn:       "0/4E00000",
			WriteLagMs:      10,
			FlushLagMs:      15,
			ReplayLagMs:     200,
			BehindByMb:      2,
			ClientAddr:      "192.168.1.101",
			BackendStart:    ptrTime(time.Now().Add(-2 * time.Hour)),
		},
	})
	require.NoError(t, err, "Should store replication metrics")

	// Get topology
	topology, err := db.GetReplicationTopology(ctx, collectorID)
	require.NoError(t, err, "Should build topology without error")

	assert.NotNil(t, topology, "Topology should not be nil")
	assert.Equal(t, collectorID, topology.CollectorID)
	assert.Equal(t, "primary", topology.NodeRole, "Should be identified as primary")
	assert.Equal(t, 2, topology.DownstreamCount, "Should have 2 downstream nodes")
	assert.Len(t, topology.DownstreamNodes, 2, "Should have 2 downstream nodes in list")
}

// TestReplicationAPI_GetReplicationTopologyStandalone tests topology for standalone instance
func TestReplicationAPI_GetReplicationTopologyStandalone(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Get topology for a collector with no replication data
	topology, err := db.GetReplicationTopology(ctx, collectorID)
	require.NoError(t, err, "Should return topology even for standalone instance")

	assert.NotNil(t, topology)
	assert.Equal(t, collectorID, topology.CollectorID)
	assert.Equal(t, "primary", topology.NodeRole, "Standalone should default to primary role")
	assert.Equal(t, 0, topology.DownstreamCount, "Should have no downstream nodes")
	assert.Empty(t, topology.DownstreamNodes)
}

// TestReplicationAPI_StoreAndGetReplicationSlots tests replication slot operations
func TestReplicationAPI_StoreAndGetReplicationSlots(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store replication slots
	err := db.StoreReplicationSlots(ctx, []*models.ReplicationSlot{
		{
			CollectorID:       collectorID,
			DatabaseName:      "postgres",
			SlotName:          "standby_slot",
			SlotType:          "physical",
			Active:            true,
			RestartLsn:        "0/5000000",
			ConfirmedFlushLsn: "0/5000000",
			WalRetainedMb:     100,
			BackendPid:        54321,
			BytesRetained:     104857600,
		},
		{
			CollectorID:       collectorID,
			DatabaseName:      "appdb",
			SlotName:          "logical_slot",
			SlotType:          "logical",
			Active:            false,
			RestartLsn:        "0/3000000",
			ConfirmedFlushLsn: "0/3000000",
			WalRetainedMb:     500,
			BackendPid:        0,
			BytesRetained:     524288000,
		},
	})
	require.NoError(t, err, "Should store replication slots without error")

	// Retrieve replication slots
	slots, err := db.GetReplicationSlots(ctx, collectorID, 100, 0)
	require.NoError(t, err, "Should retrieve replication slots without error")

	assert.NotEmpty(t, slots, "Should have replication slots")
	assert.Len(t, slots, 2, "Should have 2 slots")

	// Verify first slot (physical)
	physicalSlot := slots[0]
	assert.Equal(t, "postgres", physicalSlot.DatabaseName)
	assert.Equal(t, "standby_slot", physicalSlot.SlotName)
	assert.Equal(t, "physical", physicalSlot.SlotType)
	assert.True(t, physicalSlot.Active)
}

// TestReplicationAPI_StoreAndGetLogicalSubscriptions tests logical subscription operations
func TestReplicationAPI_StoreAndGetLogicalSubscriptions(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store logical subscriptions
	err := db.StoreLogicalSubscriptions(ctx, []*models.LogicalSubscription{
		{
			CollectorID:           collectorID,
			DatabaseName:          "appdb",
			SubName:               "sub_to_remote",
			SubState:              "enabled",
			SubRecvLsn:            "0/3000000",
			SubLatestEndLsn:       "0/3000000",
			SubLastMsgReceiptTime: ptrTime(time.Now()),
			SubWorkerPid:          12345,
		},
	})
	require.NoError(t, err, "Should store logical subscriptions without error")

	// Retrieve logical subscriptions
	subs, err := db.GetLogicalSubscriptions(ctx, collectorID, nil, 100, 0)
	require.NoError(t, err, "Should retrieve logical subscriptions without error")

	assert.NotEmpty(t, subs, "Should have subscriptions")
	assert.Len(t, subs, 1, "Should have 1 subscription")

	sub := subs[0]
	assert.Equal(t, "appdb", sub.DatabaseName)
	assert.Equal(t, "sub_to_remote", sub.SubName)
	assert.Equal(t, "enabled", sub.SubState)
}

// TestReplicationAPI_StoreAndGetPublications tests publication operations
func TestReplicationAPI_StoreAndGetPublications(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store publications
	err := db.StorePublications(ctx, []*models.Publication{
		{
			CollectorID:  collectorID,
			DatabaseName: "appdb",
			PubName:      "pub_for_subscribers",
			PubOwner:     "postgres",
			PubAllTables: false,
			PubInsert:    true,
			PubUpdate:    true,
			PubDelete:    true,
			PubTruncate:  false,
		},
	})
	require.NoError(t, err, "Should store publications without error")

	// Retrieve publications
	pubs, err := db.GetPublications(ctx, collectorID, nil, 100, 0)
	require.NoError(t, err, "Should retrieve publications without error")

	assert.NotEmpty(t, pubs, "Should have publications")
	assert.Len(t, pubs, 1, "Should have 1 publication")

	pub := pubs[0]
	assert.Equal(t, "appdb", pub.DatabaseName)
	assert.Equal(t, "pub_for_subscribers", pub.PubName)
	assert.Equal(t, "postgres", pub.PubOwner)
	assert.False(t, pub.PubAllTables)
	assert.True(t, pub.PubInsert)
	assert.True(t, pub.PubUpdate)
	assert.True(t, pub.PubDelete)
	assert.False(t, pub.PubTruncate)
}

// TestReplicationAPI_WALReceivers tests WAL receiver operations for standby detection
func TestReplicationAPI_WALReceivers(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store WAL receiver data (indicates this is a standby)
	err := db.StoreWalReceivers(ctx, []*models.WalReceiver{
		{
			CollectorID:  collectorID,
			Status:       "streaming",
			SenderHost:   "primary.example.com",
			SenderPort:   5432,
			ReceivedLsn:  "0/5000000",
			LatestEndLsn: "0/5000000",
			SlotName:     "standby_slot",
			ConnInfo:     "user=replicator",
		},
	})
	require.NoError(t, err, "Should store WAL receivers without error")

	// Retrieve WAL receivers
	receivers, err := db.GetWalReceivers(ctx, collectorID)
	require.NoError(t, err, "Should retrieve WAL receivers without error")

	assert.NotEmpty(t, receivers, "Should have WAL receivers")
	assert.Len(t, receivers, 1, "Should have 1 WAL receiver")

	receiver := receivers[0]
	assert.Equal(t, "streaming", receiver.Status)
	assert.Equal(t, "primary.example.com", receiver.SenderHost)
	assert.Equal(t, 5432, receiver.SenderPort)
}

// TestReplicationAPI_TopologyIdentifiesStandby tests that topology correctly identifies standby role
func TestReplicationAPI_TopologyIdentifiesStandby(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store WAL receiver (indicates standby)
	err := db.StoreWalReceivers(ctx, []*models.WalReceiver{
		{
			CollectorID:  collectorID,
			Status:       "streaming",
			SenderHost:   "primary.example.com",
			SenderPort:   5432,
			ReceivedLsn:  "0/5000000",
			LatestEndLsn: "0/5000000",
			SlotName:     "",
			ConnInfo:     "user=replicator",
		},
	})
	require.NoError(t, err)

	// Get topology - should identify as standby
	topology, err := db.GetReplicationTopology(ctx, collectorID)
	require.NoError(t, err)

	assert.Equal(t, "standby", topology.NodeRole, "Should be identified as standby when has WAL receiver")
	assert.Equal(t, "primary.example.com", topology.UpstreamHost)
	assert.Equal(t, 5432, topology.UpstreamPort)
}

// TestReplicationAPI_TopologyIdentifiesCascadingStandby tests cascading standby detection
func TestReplicationAPI_TopologyIdentifiesCascadingStandby(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store WAL receiver (indicates standby receiving from upstream)
	err := db.StoreWalReceivers(ctx, []*models.WalReceiver{
		{
			CollectorID:  collectorID,
			Status:       "streaming",
			SenderHost:   "primary.example.com",
			SenderPort:   5432,
			ReceivedLsn:  "0/5000000",
			LatestEndLsn: "0/5000000",
			SlotName:     "",
			ConnInfo:     "user=replicator",
		},
	})
	require.NoError(t, err)

	// Also store downstream replication status (other standbys connected to this one)
	err = db.StoreReplicationMetrics(ctx, []*models.ReplicationStatus{
		{
			CollectorID:     collectorID,
			ServerPID:       54321,
			Usename:         "replicator",
			ApplicationName: "cascading_standby",
			State:           "streaming",
			SyncState:       "async",
			WriteLsn:        "0/4F00000",
			FlushLsn:        "0/4F00000",
			ReplayLsn:       "0/4E00000",
			WriteLagMs:      5,
			FlushLagMs:      10,
			ReplayLagMs:     50,
			BehindByMb:      1,
			ClientAddr:      "192.168.1.200",
			BackendStart:    ptrTime(time.Now().Add(-30 * time.Minute)),
		},
	})
	require.NoError(t, err)

	// Get topology - should identify as cascading standby
	topology, err := db.GetReplicationTopology(ctx, collectorID)
	require.NoError(t, err)

	assert.Equal(t, "cascading_standby", topology.NodeRole, "Should be identified as cascading standby")
	assert.Equal(t, 1, topology.DownstreamCount)
}

// TestReplicationAPI_MultipleReplicationStatuses tests handling multiple replication entries
func TestReplicationAPI_MultipleReplicationStatuses(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	collectorID := createTestCollectorForReplication(t, db)

	// Store multiple replication status entries
	for i := 0; i < 5; i++ {
		err := db.StoreReplicationMetrics(ctx, []*models.ReplicationStatus{
			{
				CollectorID:     collectorID,
				ServerPID:       int64(12345 + i),
				Usename:         "replicator",
				ApplicationName: "standby" + string(rune('1'+i)),
				State:           "streaming",
				SyncState:       "async",
				WriteLsn:        "0/5000000",
				FlushLsn:        "0/5000000",
				ReplayLsn:       "0/4F00000",
				WriteLagMs:      10,
				FlushLagMs:      15,
				ReplayLagMs:     int64(100 + i*10),
				BehindByMb:      int64(i + 1),
				ClientAddr:      "192.168.1.10" + string(rune('0'+i)),
				BackendStart:    ptrTime(time.Now().Add(-time.Duration(i+1) * time.Hour)),
			},
		})
		require.NoError(t, err)
	}

	// Retrieve with pagination
	metrics, err := db.GetReplicationMetrics(ctx, collectorID, 3, 0)
	require.NoError(t, err)

	// Should return up to limit
	assert.LessOrEqual(t, len(metrics.ReplicationStatus), 3)
}

// Helper function for creating test collectors for replication tests
func createTestCollectorForReplication(t *testing.T, db *storage.PostgresDB) uuid.UUID {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collectorID := uuid.New()
	hostname := "test-replication-host-" + collectorID.String()[:8]

	query := `
		INSERT INTO collectors (id, hostname, last_seen, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET hostname = $2
	`

	now := time.Now()
	_, err := db.GetDB().ExecContext(ctx, query, collectorID, hostname, now, now, now)
	require.NoError(t, err, "Failed to create test collector for replication")

	return collectorID
}
