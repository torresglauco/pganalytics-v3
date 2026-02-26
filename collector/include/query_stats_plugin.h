#pragma once

#include <string>
#include <vector>
#include <memory>
#include <nlohmann/json.hpp>
#include "connection_pool.h"

using json = nlohmann::json;

// Forward declarations
class PQconn;

/**
 * PostgreSQL Query Statistics Collector
 * Gathers pg_stat_statements data for query-level performance analysis
 *
 * Requirements:
 * - PostgreSQL extension: pg_stat_statements must be installed
 * - Configuration: shared_preload_libraries = 'pg_stat_statements'
 * - Permissions: Monitoring user must have SELECT on pg_stat_statements
 *
 * Metrics collected:
 * - Query hash and normalized text
 * - Execution count and timing (total, mean, min, max, stddev)
 * - Rows processed
 * - Buffer cache statistics (shared, local, temp blocks)
 * - I/O timing (block read/write time)
 * - WAL statistics (PG13+ optional)
 * - Query planning/execution time (PG13+ optional)
 */
class PgQueryStatsCollector {
public:
    /**
     * Constructor
     * @param hostname Collector hostname
     * @param collectorId Unique collector identifier
     * @param postgresHost PostgreSQL server host
     * @param postgresPort PostgreSQL server port
     * @param postgresUser Database user for connection
     * @param postgresPassword Database password
     * @param databases List of databases to monitor
     */
    PgQueryStatsCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    /**
     * Execute query stats collection
     * @return JSON object with collected query statistics
     */
    json execute();

    /**
     * Get collector type
     */
    std::string getType() const { return "pg_query_stats"; }

    /**
     * Check if this collector is enabled
     */
    bool isEnabled() const { return enabled_; }

private:
    // Configuration
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    // Connection pooling (configured via config file)
    // Pool maintains min_size=2 to max_size=10 persistent connections
    // Reduces 200-400ms connection overhead to 5-10ms per collection
    std::unique_ptr<ConnectionPool> pool_;

    // Pool monitoring metrics
    struct PoolMetrics {
        size_t acquisitions_;     // Total pool acquisitions
        size_t reuses_;           // Times connection was reused
        size_t new_connections_;  // Times new connection created
        double avg_acquire_ms_;   // Average acquisition time
    } pool_metrics_;

    /**
     * Initialize connection pool (called in constructor)
     */
    void initializeConnectionPool();

    /**
     * Collect query statistics from a single database
     * @param dbname Database name
     * @return JSON object with queries from pg_stat_statements
     */
    json collectQueryStats(const std::string& dbname);

    /**
     * Execute EXPLAIN ANALYZE on a query and capture the plan
     * Phase 4.4.2: EXPLAIN PLAN Integration
     * @param dbname Database name
     * @param queryHash Query hash identifier
     * @param queryText Query text to explain
     * @return JSON object with EXPLAIN plan or null if execution fails
     */
    json executeExplainPlan(const std::string& dbname, int64_t queryHash, const std::string& queryText);

    /**
     * Check if query should be explained (execution time > threshold)
     * @param meanTime Average execution time in milliseconds
     * @return true if query exceeds EXPLAIN threshold (1000ms)
     */
    bool shouldExplainQuery(float meanTime) const { return meanTime > 1000.0f; }
};
