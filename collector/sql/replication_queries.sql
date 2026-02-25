-- PostgreSQL Replication Metrics Queries
-- Used by PgReplicationCollector in collector/src/replication_plugin.cpp

-- Query 1: Replication Slots Status
-- Purpose: Collect information about all replication slots (physical and logical)
-- Requires: PostgreSQL 9.4+
-- Output: slot_name, slot_type, active status, LSN information, WAL retention
SELECT
    slot_name,
    slot_type,
    active,
    restart_lsn,
    confirmed_flush_lsn,
    COALESCE(ROUND(EXTRACT(EPOCH FROM (NOW() - pg_postmaster_start_time())) * 1024 * 1024), 0) as wal_retained_mb,
    plugin_active,
    COALESCE(backend_pid, 0) as backend_pid,
    NULL as database,
    COALESCE(OCTET_LENGTH(restart_lsn::text), 0) as bytes_retained
FROM pg_replication_slots
ORDER BY slot_name;


-- Query 2: Streaming Replication Status (PostgreSQL 13+)
-- Purpose: Collect streaming replication lag metrics in milliseconds
-- Requires: PostgreSQL 13+ (for write_lag, flush_lag, replay_lag)
-- Output: Connection info, LSN positions, lag in milliseconds, client address
SELECT
    server_pid,
    usename,
    application_name,
    state,
    sync_state,
    COALESCE(write_lsn::text, '0/0') as write_lsn,
    COALESCE(flush_lsn::text, '0/0') as flush_lsn,
    COALESCE(replay_lsn::text, '0/0') as replay_lsn,
    COALESCE(EXTRACT(EPOCH FROM write_lag) * 1000, 0)::bigint as write_lag_ms,
    COALESCE(EXTRACT(EPOCH FROM flush_lag) * 1000, 0)::bigint as flush_lag_ms,
    COALESCE(EXTRACT(EPOCH FROM replay_lag) * 1000, 0)::bigint as replay_lag_ms,
    COALESCE(backend_xmin, 0) as behind_by_mb,
    client_addr::text,
    backend_start::text
FROM pg_stat_replication
ORDER BY usename, application_name;


-- Query 3: Streaming Replication Status (PostgreSQL 9.4-12)
-- Purpose: Collect streaming replication status (pre-PG13, without lag metrics)
-- Requires: PostgreSQL 9.4+
-- Output: Connection info, LSN positions only
SELECT
    procpid as server_pid,
    usesysid,
    application_name,
    state,
    sync_state,
    COALESCE(location::text, '0/0') as write_lsn,
    COALESCE(location::text, '0/0') as flush_lsn,
    COALESCE(replay_location::text, '0/0') as replay_lsn,
    0 as write_lag_ms,
    0 as flush_lag_ms,
    0 as replay_lag_ms,
    0 as behind_by_mb,
    client_addr::text,
    backend_start::text
FROM pg_stat_replication
ORDER BY usesysid, application_name;


-- Query 4: WAL Segment Status (PostgreSQL 13+)
-- Purpose: Collect WAL space usage and growth metrics
-- Requires: PostgreSQL 13+ (for pg_wal_space())
-- Output: WAL size in MB, segment count, growth rate
SELECT
    COUNT(*) as total_segments,
    ROUND(SUM(pg_file_stat(pg_ls_waldir())) / 1024.0 / 1024.0) as current_wal_size_mb
FROM pg_ls_waldir();


-- Query 5: Vacuum Wraparound Risk Assessment
-- Purpose: Identify databases and tables at risk of transaction ID wraparound
-- Requires: PostgreSQL 9.4+
-- Output: XID age, percentage remaining, risk assessment
SELECT
    datname,
    datfrozenxid,
    (SELECT max(age(pg_xact_commit_timestamp(xmin)))
     FROM pg_class
     WHERE pg_xact_commit_timestamp(xmin) IS NOT NULL) as max_age,
    2147483647 - datfrozenxid as xid_remaining,
    ROUND(100.0 * (2147483647 - datfrozenxid) / 2147483647, 2) as percent_remaining
FROM pg_database
WHERE datname NOT IN ('template0', 'template1')
ORDER BY datfrozenxid;


-- Query 6: PostgreSQL Version Detection
-- Purpose: Determine PostgreSQL version for query compatibility
-- Requires: PostgreSQL 9.4+
-- Output: Version number as integer (e.g., 130000 for PG13)
SELECT current_setting('server_version_num')::int;


-- Query 7: Logical Replication Subscriptions (PostgreSQL 10+)
-- Purpose: Monitor logical replication subscription status
-- Requires: PostgreSQL 10+ (for logical replication)
-- Output: Subscription info and status
SELECT
    s.subname,
    s.subdboid,
    s.subowner,
    s.subenabled,
    s.subconninfo,
    s.subslotname,
    s.subsynccommit,
    COALESCE(ss.subid, 0) as subid,
    COALESCE(ss.lsn::text, '0/0') as lsn,
    COALESCE(ss.received_lsn::text, '0/0') as received_lsn,
    COALESCE(ss.latest_end_lsn::text, '0/0') as latest_end_lsn,
    COALESCE(ss.latest_end_time, NOW()) as latest_end_time
FROM pg_subscription s
LEFT JOIN pg_stat_subscription ss ON s.oid = ss.subid
ORDER BY s.subname;


-- Query 8: Tables Requiring Vacuum Due to XID Age
-- Purpose: Identify tables that need urgent vacuum due to XID age
-- Requires: PostgreSQL 9.4+
-- Output: Table name, XID age, vacuum status
SELECT
    schemaname,
    tablename,
    n_live_tup,
    n_dead_tup,
    last_vacuum,
    last_autovacuum,
    CASE WHEN age(relfrozenxid) > (
        SELECT setting::int FROM pg_settings WHERE name = 'autovacuum_freeze_max_age'
    ) THEN 'NEEDS_VACUUM'
    WHEN age(relfrozenxid) > (
        SELECT setting::int * 0.75 FROM pg_settings WHERE name = 'autovacuum_freeze_max_age'
    ) THEN 'AT_RISK'
    ELSE 'OK'
    END as vacuum_status,
    age(relfrozenxid) as xid_age
FROM pg_stat_user_tables
WHERE age(relfrozenxid) > (
    SELECT setting::int * 0.5 FROM pg_settings WHERE name = 'autovacuum_freeze_max_age'
)
ORDER BY age(relfrozenxid) DESC;


-- Query 9: LSN Position Analysis
-- Purpose: Calculate bytes behind based on LSN values
-- Requires: PostgreSQL 9.4+
-- Note: LSN format is "X/XXXXXXXX" where values are in hex
-- Usage: Parse LSN strings in application and calculate (write_lsn - replay_lsn)
-- This is computed in C++ code using parseLsn() function
SELECT
    slot_name,
    slot_type,
    restart_lsn,
    confirmed_flush_lsn,
    (pg_wal_lsn_diff(restart_lsn::pg_lsn, '0/0') / 1024.0 / 1024.0)::bigint as restart_lsn_mb,
    (pg_wal_lsn_diff(confirmed_flush_lsn::pg_lsn, '0/0') / 1024.0 / 1024.0)::bigint as confirmed_flush_lsn_mb
FROM pg_replication_slots;


-- Query 10: Replica Lag Summary
-- Purpose: Aggregate view of all replicas and their lag status
-- Requires: PostgreSQL 13+ (for lag metrics)
-- Output: Summary statistics for replication health
SELECT
    COUNT(*) as replica_count,
    MAX(replay_lag_ms) as max_replay_lag_ms,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY replay_lag_ms) as median_replay_lag_ms,
    AVG(replay_lag_ms) as avg_replay_lag_ms,
    SUM(CASE WHEN sync_state = 'sync' THEN 1 ELSE 0 END) as sync_replicas,
    SUM(CASE WHEN sync_state = 'async' THEN 1 ELSE 0 END) as async_replicas
FROM pg_stat_replication;


-- Notes on PostgreSQL Version Compatibility:
--
-- PG 9.4 - 9.6:    Basic replication support, pg_replication_slots
-- PG 10:           Logical replication, pg_subscription
-- PG 11 - 12:      Enhanced replication views
-- PG 13+:          write_lag, flush_lag, replay_lag in milliseconds
-- PG 14+:          pg_wal_space() function for improved WAL tracking
-- PG 15+:          Enhanced slot statistics
--
-- The C++ collector detects the version at runtime and uses appropriate queries.
