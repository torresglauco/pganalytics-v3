#pragma once

#include "collector.h"
#include <string>
#include <vector>
#include <cstdint>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

// Forward declaration
class PQconn;

/**
 * PostgreSQL Replication Collector
 *
 * Gathers comprehensive replication metrics including:
 * - Streaming replication status (write/flush/replay lag)
 * - Physical and logical replication slots
 * - WAL segment status and retention
 * - Transaction ID (XID) wraparound risk assessment
 * - Replication slot resource usage
 *
 * Requirements:
 * - PostgreSQL 9.4+ (basic replication)
 * - PostgreSQL 10+ (logical replication)
 * - PostgreSQL 13+ (enhanced metrics like write_lag, flush_lag)
 * - Permissions: Monitoring user must have SUPERUSER or pg_monitor role
 * - Configuration: wal_level = 'replica' or 'logical' (for replication)
 *
 * Metrics collected:
 * - Replication lag (write, flush, replay in milliseconds)
 * - Replication slot status (physical, logical, inactive)
 * - WAL segment size and growth rate
 * - XID wraparound risk (percentage remaining)
 * - Tables requiring vacuum (age analysis)
 * - Logical replication subscriptions status
 */
class PgReplicationCollector : public Collector {
public:
    /**
     * Data structure for replication slot information
     */
    struct ReplicationSlot {
        std::string slot_name;
        std::string slot_type;           // "physical" or "logical"
        bool active;
        std::string restart_lsn;         // For physical slots
        std::string confirmed_flush_lsn; // For logical slots
        int64_t wal_retained_mb;
        bool plugin_active;              // For logical slots
        int64_t backend_pid;             // 0 if not active
        std::string database;            // NULL for physical, database name for logical
        int64_t bytes_retained;          // WAL bytes retained
    };

    /**
     * Data structure for streaming replication status
     */
    struct ReplicationStatus {
        int64_t server_pid;
        std::string usename;
        std::string application_name;
        std::string state;               // "streaming", "catchup", "backup", "backup_canceling"
        std::string sync_state;          // "sync" or "async"
        std::string write_lsn;           // PG13+
        std::string flush_lsn;           // PG13+
        std::string replay_lsn;          // PG13+
        int64_t write_lag_ms;            // PG13+ lag in milliseconds
        int64_t flush_lag_ms;            // PG13+ lag in milliseconds
        int64_t replay_lag_ms;           // PG13+ lag in milliseconds
        int64_t behind_by_mb;            // Estimated bytes behind (calculated)
        std::string client_addr;
        std::string backend_start;
    };

    /**
     * Data structure for XID wraparound risk assessment
     */
    struct VacuumWrapAroundRisk {
        std::string database;
        int64_t relfrozenxid;
        int64_t current_xid;
        int64_t xid_until_wraparound;
        int64_t percent_until_wraparound;
        bool at_risk;                    // true if < 20% remaining
        int64_t tables_needing_vacuum;
        int64_t oldest_table_age;        // Age of oldest table in transaction ids
    };

    /**
     * Data structure for WAL segment status
     */
    struct WalSegmentStatus {
        int64_t total_segments;
        int64_t current_wal_size_mb;
        int64_t wal_directory_size_mb;
        std::string last_wal_segment;
        std::string oldest_wal_segment;
        int64_t segments_since_checkpoint;
        double growth_rate_mb_per_hour;  // Estimated based on recent data
        std::string pg_wal_space_mb;     // Output of pg_wal_space() if available
    };

    /**
     * Constructor
     * @param hostname Collector hostname
     * @param collectorId Unique collector identifier
     * @param postgresHost PostgreSQL server host
     * @param postgresPort PostgreSQL server port
     * @param postgresUser Database user for connection (must be superuser/pg_monitor)
     * @param postgresPassword Database password
     * @param databases List of databases to monitor (empty = all databases)
     */
    PgReplicationCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases = {}
    );

    /**
     * Destructor
     */
    ~PgReplicationCollector();

    /**
     * Execute replication metrics collection
     * @return JSON object with all replication metrics
     */
    json execute() override;

    /**
     * Get collector type
     */
    std::string getType() const override { return "pg_replication"; }

    /**
     * Check if this collector is enabled
     */
    bool isEnabled() const override { return enabled_; }

private:
    // Configuration
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    // PostgreSQL version cache
    int postgres_version_major_;
    int postgres_version_minor_;
    bool version_detected_;

    /**
     * Detect PostgreSQL version
     * @return Major version (9, 10, 11, 12, 13, 14, 15, 16)
     */
    int detectPostgresVersion();

    /**
     * Collect replication slots from pg_replication_slots view
     * @return Vector of ReplicationSlot structures
     */
    std::vector<ReplicationSlot> collectReplicationSlots();

    /**
     * Collect streaming replication status from pg_stat_replication view
     * @return Vector of ReplicationStatus structures
     */
    std::vector<ReplicationStatus> collectReplicationStatus();

    /**
     * Collect WAL segment status information
     * @return WalSegmentStatus structure
     */
    WalSegmentStatus collectWalSegmentStatus();

    /**
     * Collect vacuum wraparound risk for all databases
     * @return Vector of VacuumWrapAroundRisk structures
     */
    std::vector<VacuumWrapAroundRisk> collectVacuumWrapAroundRisk();

    /**
     * Collect logical replication subscription status
     * @return JSON array of subscription status (PG10+ feature)
     */
    json collectLogicalSubscriptions();

    /**
     * Connect to a PostgreSQL database
     * @param dbname Database name to connect to (or "postgres" for cluster-wide queries)
     * @return PQconn pointer or nullptr on failure
     */
    PQconn* connectToDatabase(const std::string& dbname);

    /**
     * Execute a SELECT query and return results as JSON array
     * @param conn Database connection
     * @param query SQL query to execute
     * @return JSON array with query results
     */
    json executeQuery(PQconn* conn, const std::string& query);

    /**
     * Parse LSN string to bytes for comparison
     * @param lsn LSN string in format "X/XXXXXXXX"
     * @return LSN as uint64_t
     */
    uint64_t parseLsn(const std::string& lsn);

    /**
     * Calculate bytes behind based on LSN values
     * @param write_lsn LSN being written
     * @param replay_lsn LSN being replayed
     * @return Approximate bytes behind
     */
    int64_t calculateBytesBehind(const std::string& write_lsn, const std::string& replay_lsn);

    /**
     * Check if table requires vacuum due to XID age
     * @param age Transaction ID age
     * @param autovacuum_freeze_max_age Config value
     * @return true if vacuum needed
     */
    bool tableNeedsVacuum(int64_t age, int64_t autovacuum_freeze_max_age);
};
