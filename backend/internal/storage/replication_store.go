package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
)

// ============================================================================
// STREAMING REPLICATION OPERATIONS
// ============================================================================

// StoreReplicationMetrics inserts replication status metrics into the database
func (p *PostgresDB) StoreReplicationMetrics(ctx context.Context, status []*models.ReplicationStatus) error {
	if len(status) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_replication_status (time, collector_id, server_pid, usename, application_name, state, sync_state, write_lsn, flush_lsn, replay_lsn, write_lag_ms, flush_lag_ms, replay_lag_ms, behind_by_mb, client_addr, backend_start)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare replication status insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, s := range status {
		if _, err := stmt.ExecContext(ctx, time.Now(), s.CollectorID, s.ServerPID, s.Usename, s.ApplicationName, s.State, s.SyncState, s.WriteLsn, s.FlushLsn, s.ReplayLsn, s.WriteLagMs, s.FlushLagMs, s.ReplayLagMs, s.BehindByMb, s.ClientAddr, s.BackendStart); err != nil {
			return apperrors.DatabaseError("insert replication status", err.Error())
		}
	}

	return tx.Commit()
}

// GetReplicationMetrics retrieves replication status metrics for a collector
func (p *PostgresDB) GetReplicationMetrics(ctx context.Context, collectorID uuid.UUID, limit int, offset int) (*models.ReplicationMetricsResponse, error) {
	resp := &models.ReplicationMetricsResponse{}

	query := `
		SELECT server_pid, usename, application_name, state, sync_state, write_lsn, flush_lsn, replay_lsn, write_lag_ms, flush_lag_ms, replay_lag_ms, behind_by_mb, client_addr, backend_start
		FROM metrics_replication_status
		WHERE collector_id = $1
		ORDER BY time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.db.QueryContext(ctx, query, collectorID, limit, offset)
	if err != nil {
		return nil, apperrors.DatabaseError("query replication status", err.Error())
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		s := &models.ReplicationStatus{CollectorID: collectorID}
		if err := rows.Scan(&s.ServerPID, &s.Usename, &s.ApplicationName, &s.State, &s.SyncState, &s.WriteLsn, &s.FlushLsn, &s.ReplayLsn, &s.WriteLagMs, &s.FlushLagMs, &s.ReplayLagMs, &s.BehindByMb, &s.ClientAddr, &s.BackendStart); err != nil {
			return nil, apperrors.DatabaseError("scan replication status", err.Error())
		}
		resp.ReplicationStatus = append(resp.ReplicationStatus, s)
	}

	return resp, nil
}

// StoreReplicationSlots inserts replication slots into the database
func (p *PostgresDB) StoreReplicationSlots(ctx context.Context, slots []*models.ReplicationSlot) error {
	if len(slots) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_replication_slots (time, collector_id, database_name, slot_name, slot_type, active, restart_lsn, confirmed_flush_lsn, wal_retained_mb, backend_pid, bytes_retained)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare replication slots insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, s := range slots {
		if _, err := stmt.ExecContext(ctx, time.Now(), s.CollectorID, s.DatabaseName, s.SlotName, s.SlotType, s.Active, s.RestartLsn, s.ConfirmedFlushLsn, s.WalRetainedMb, s.BackendPid, s.BytesRetained); err != nil {
			return apperrors.DatabaseError("insert replication slot", err.Error())
		}
	}

	return tx.Commit()
}

// GetReplicationSlots retrieves replication slots for a collector
func (p *PostgresDB) GetReplicationSlots(ctx context.Context, collectorID uuid.UUID, limit int, offset int) ([]*models.ReplicationSlot, error) {
	query := `
		SELECT database_name, slot_name, slot_type, active, restart_lsn, confirmed_flush_lsn, wal_retained_mb, backend_pid, bytes_retained
		FROM metrics_replication_slots
		WHERE collector_id = $1
		ORDER BY time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := p.db.QueryContext(ctx, query, collectorID, limit, offset)
	if err != nil {
		return nil, apperrors.DatabaseError("query replication slots", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var slots []*models.ReplicationSlot
	for rows.Next() {
		s := &models.ReplicationSlot{CollectorID: collectorID}
		if err := rows.Scan(&s.DatabaseName, &s.SlotName, &s.SlotType, &s.Active, &s.RestartLsn, &s.ConfirmedFlushLsn, &s.WalRetainedMb, &s.BackendPid, &s.BytesRetained); err != nil {
			return nil, apperrors.DatabaseError("scan replication slot", err.Error())
		}
		slots = append(slots, s)
	}

	return slots, nil
}

// ============================================================================
// LOGICAL REPLICATION OPERATIONS
// ============================================================================

// StoreLogicalSubscriptions inserts logical subscriptions into the database
func (p *PostgresDB) StoreLogicalSubscriptions(ctx context.Context, subs []*models.LogicalSubscription) error {
	if len(subs) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_logical_subscriptions (time, collector_id, database_name, sub_name, sub_state, sub_recv_lsn, sub_latest_end_lsn, sub_last_msg_receipt_time, sub_worker_pid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare logical subscriptions insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, s := range subs {
		if _, err := stmt.ExecContext(ctx, time.Now(), s.CollectorID, s.DatabaseName, s.SubName, s.SubState, s.SubRecvLsn, s.SubLatestEndLsn, s.SubLastMsgReceiptTime, s.SubWorkerPid); err != nil {
			return apperrors.DatabaseError("insert logical subscription", err.Error())
		}
	}

	return tx.Commit()
}

// GetLogicalSubscriptions retrieves logical subscriptions for a collector
func (p *PostgresDB) GetLogicalSubscriptions(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) ([]*models.LogicalSubscription, error) {
	query := `SELECT database_name, sub_name, sub_state, sub_recv_lsn, sub_latest_end_lsn, sub_last_msg_receipt_time, sub_worker_pid FROM metrics_logical_subscriptions WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query logical subscriptions", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var subs []*models.LogicalSubscription
	for rows.Next() {
		s := &models.LogicalSubscription{CollectorID: collectorID}
		if err := rows.Scan(&s.DatabaseName, &s.SubName, &s.SubState, &s.SubRecvLsn, &s.SubLatestEndLsn, &s.SubLastMsgReceiptTime, &s.SubWorkerPid); err != nil {
			return nil, apperrors.DatabaseError("scan logical subscription", err.Error())
		}
		subs = append(subs, s)
	}

	return subs, nil
}

// StorePublications inserts publications into the database
func (p *PostgresDB) StorePublications(ctx context.Context, pubs []*models.Publication) error {
	if len(pubs) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_publications (time, collector_id, database_name, pub_name, pub_owner, pub_all_tables, pub_insert, pub_update, pub_delete, pub_truncate)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare publications insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, pub := range pubs {
		if _, err := stmt.ExecContext(ctx, time.Now(), pub.CollectorID, pub.DatabaseName, pub.PubName, pub.PubOwner, pub.PubAllTables, pub.PubInsert, pub.PubUpdate, pub.PubDelete, pub.PubTruncate); err != nil {
			return apperrors.DatabaseError("insert publication", err.Error())
		}
	}

	return tx.Commit()
}

// GetPublications retrieves publications for a collector
func (p *PostgresDB) GetPublications(ctx context.Context, collectorID uuid.UUID, database *string, limit int, offset int) ([]*models.Publication, error) {
	query := `SELECT database_name, pub_name, pub_owner, pub_all_tables, pub_insert, pub_update, pub_delete, pub_truncate FROM metrics_publications WHERE collector_id = $1`
	args := []interface{}{collectorID}

	if database != nil {
		query += ` AND database_name = $2`
		args = append(args, *database)
	}

	query += ` ORDER BY time DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.DatabaseError("query publications", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var pubs []*models.Publication
	for rows.Next() {
		p := &models.Publication{CollectorID: collectorID}
		if err := rows.Scan(&p.DatabaseName, &p.PubName, &p.PubOwner, &p.PubAllTables, &p.PubInsert, &p.PubUpdate, &p.PubDelete, &p.PubTruncate); err != nil {
			return nil, apperrors.DatabaseError("scan publication", err.Error())
		}
		pubs = append(pubs, p)
	}

	return pubs, nil
}

// StoreWalReceivers inserts WAL receiver data into the database
func (p *PostgresDB) StoreWalReceivers(ctx context.Context, receivers []*models.WalReceiver) error {
	if len(receivers) == 0 {
		return nil
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.DatabaseError("begin transaction", err.Error())
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics_wal_receivers (time, collector_id, status, sender_host, sender_port, received_lsn, latest_end_lsn, slot_name, conn_info)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return apperrors.DatabaseError("prepare wal receivers insert", err.Error())
	}
	defer func() { _ = stmt.Close() }()

	for _, r := range receivers {
		if _, err := stmt.ExecContext(ctx, time.Now(), r.CollectorID, r.Status, r.SenderHost, r.SenderPort, r.ReceivedLsn, r.LatestEndLsn, r.SlotName, r.ConnInfo); err != nil {
			return apperrors.DatabaseError("insert wal receiver", err.Error())
		}
	}

	return tx.Commit()
}

// GetWalReceivers retrieves WAL receiver data for a collector
func (p *PostgresDB) GetWalReceivers(ctx context.Context, collectorID uuid.UUID) ([]*models.WalReceiver, error) {
	query := `
		SELECT status, sender_host, sender_port, received_lsn, latest_end_lsn, slot_name, conn_info
		FROM metrics_wal_receivers
		WHERE collector_id = $1
		ORDER BY time DESC
		LIMIT 1
	`

	rows, err := p.db.QueryContext(ctx, query, collectorID)
	if err != nil {
		return nil, apperrors.DatabaseError("query wal receivers", err.Error())
	}
	defer func() { _ = rows.Close() }()

	var receivers []*models.WalReceiver
	for rows.Next() {
		r := &models.WalReceiver{CollectorID: collectorID}
		if err := rows.Scan(&r.Status, &r.SenderHost, &r.SenderPort, &r.ReceivedLsn, &r.LatestEndLsn, &r.SlotName, &r.ConnInfo); err != nil {
			return nil, apperrors.DatabaseError("scan wal receiver", err.Error())
		}
		receivers = append(receivers, r)
	}

	return receivers, nil
}

// GetReplicationTopology builds the replication topology for a collector
func (p *PostgresDB) GetReplicationTopology(ctx context.Context, collectorID uuid.UUID) (*models.ReplicationTopology, error) {
	topology := &models.ReplicationTopology{
		CollectorID: collectorID,
		NodeRole:    "primary", // Default to primary
	}

	// Check if this node has a WAL receiver (meaning it's a standby)
	walReceivers, err := p.GetWalReceivers(ctx, collectorID)
	if err != nil {
		return nil, err
	}

	if len(walReceivers) > 0 {
		// This is a standby
		topology.NodeRole = "standby"
		topology.UpstreamHost = walReceivers[0].SenderHost
		topology.UpstreamPort = walReceivers[0].SenderPort
	}

	// Get downstream nodes (replicas connected to this node)
	query := `
		SELECT application_name, client_addr, state, sync_state, replay_lag_ms
		FROM metrics_replication_status
		WHERE collector_id = $1
		ORDER BY time DESC
		LIMIT 10
	`

	rows, err := p.db.QueryContext(ctx, query, collectorID)
	if err != nil {
		return nil, apperrors.DatabaseError("query downstream nodes", err.Error())
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		node := models.TopologyNode{CollectorID: collectorID}
		if err := rows.Scan(&node.ApplicationName, &node.ClientAddr, &node.State, &node.SyncState, &node.ReplayLagMs); err != nil {
			return nil, apperrors.DatabaseError("scan downstream node", err.Error())
		}
		topology.DownstreamNodes = append(topology.DownstreamNodes, node)
	}

	topology.DownstreamCount = len(topology.DownstreamNodes)

	// If this is a standby with downstream nodes, it's a cascading standby
	if len(walReceivers) > 0 && len(topology.DownstreamNodes) > 0 {
		topology.NodeRole = "cascading_standby"
	}

	return topology, nil
}
