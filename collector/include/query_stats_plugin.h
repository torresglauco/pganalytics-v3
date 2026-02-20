#pragma once

#include <string>
#include <vector>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

// Forward declaration
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

    /**
     * Collect query statistics from a single database
     * @param dbname Database name
     * @return JSON object with queries from pg_stat_statements
     */
    json collectQueryStats(const std::string& dbname);
};
