#pragma once

#include "collector.h"
#include <string>
#include <vector>
#include <cstdint>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>
#else
// Forward declaration when libpq not available
typedef struct pg_conn PGconn;
typedef struct pg_result PGresult;
#endif

/**
 * PostgreSQL Logical Replication Collector
 *
 * Gathers logical replication metrics including:
 * - Logical subscriptions (pg_stat_subscription)
 * - Publications (pg_publication)
 * - WAL receiver status (pg_stat_wal_receiver) for topology detection
 *
 * Requirements:
 * - PostgreSQL 10+ (logical replication introduced)
 * - PostgreSQL 9.6+ (pg_stat_wal_receiver)
 * - Permissions: Monitoring user must have pg_monitor role or SUPERUSER
 *
 * Used for:
 * - Monitoring logical replication subscription health
 * - Tracking publication definitions
 * - Determining replication topology (primary/standby/cascading)
 */
class LogicalReplicationCollector : public Collector {
public:
    /**
     * Data structure for logical subscription
     */
    struct LogicalSubscription {
        std::string sub_name;
        std::string sub_state;            // "ready", "syncing", "error", "disabled"
        std::string received_lsn;
        std::string latest_end_lsn;
        std::string last_msg_receipt_time;
        std::string last_msg_send_time;
        int64_t worker_pid;
        int worker_count;
        std::string database;
    };

    /**
     * Data structure for publication
     */
    struct Publication {
        std::string pub_name;
        std::string pub_owner;
        bool pub_all_tables;
        bool pub_insert;
        bool pub_update;
        bool pub_delete;
        bool pub_truncate;
        std::string database;
    };

    /**
     * Data structure for WAL receiver (standby detection)
     */
    struct WalReceiver {
        std::string status;               // "streaming", "catching up", etc.
        std::string sender_host;
        int sender_port;
        std::string received_lsn;
        std::string latest_end_lsn;
        std::string slot_name;
        std::string conn_info;
    };

    /**
     * Constructor
     * @param hostname Collector hostname
     * @param collectorId Unique collector identifier
     * @param postgresHost PostgreSQL server host
     * @param postgresPort PostgreSQL server port
     * @param postgresUser Database user for connection
     * @param postgresPassword Database password
     * @param databases List of databases to monitor (empty = all)
     */
    LogicalReplicationCollector(
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
    ~LogicalReplicationCollector();

    /**
     * Execute logical replication metrics collection
     * @return JSON object with all logical replication metrics
     */
    json execute() override;

    /**
     * Get collector type
     */
    std::string getType() const override { return "pg_logical_replication"; }

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
     * @return Major version (9, 10, 11, 12, 13, 14, 15, 16, 17)
     */
    int detectPostgresVersion();

    /**
     * Collect logical subscriptions from pg_stat_subscription (PG 10+)
     * @param dbname Database name to query
     * @return Vector of LogicalSubscription structures
     */
    std::vector<LogicalSubscription> collectLogicalSubscriptions(const std::string& dbname);

    /**
     * Collect publications from pg_publication
     * @param dbname Database name to query
     * @return Vector of Publication structures
     */
    std::vector<Publication> collectPublications(const std::string& dbname);

    /**
     * Collect WAL receiver status from pg_stat_wal_receiver (PG 9.6+)
     * Used to determine if this node is a standby
     * @return WalReceiver structure (single row)
     */
    WalReceiver collectWalReceiver();

    /**
     * Connect to a PostgreSQL database
     * @param dbname Database name to connect to
     * @return PGconn pointer or nullptr on failure
     */
    PGconn* connectToDatabase(const std::string& dbname);

    /**
     * Get list of databases to monitor
     * @return Vector of database names
     */
    std::vector<std::string> getDatabases();
};