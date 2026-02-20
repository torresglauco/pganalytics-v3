#pragma once

#include <string>
#include <vector>
#include <map>

/**
 * E2E Database Helper
 *
 * Provides utilities for querying PostgreSQL/TimescaleDB during E2E tests.
 * Handles:
 * - Metrics verification (count, schema)
 * - Configuration verification
 * - Data cleanup between tests
 */
class E2EDatabaseHelper {
public:
    /**
     * Constructor
     * @param main_db_url PostgreSQL URL for pganalytics metadata
     * @param metrics_db_url PostgreSQL/TimescaleDB URL for metrics
     */
    E2EDatabaseHelper(
        const std::string& main_db_url,
        const std::string& metrics_db_url
    );

    // Metrics queries
    int getMetricsCount(const std::string& table);
    int getMetricsCountForCollector(const std::string& table, const std::string& collector_id);
    bool metricsExist(const std::string& collector_id);
    std::string getLatestMetricTimestamp(const std::string& table);

    // Schema verification
    bool tableExists(const std::string& table);
    bool columnExists(const std::string& table, const std::string& column);
    std::vector<std::string> getTableColumns(const std::string& table);

    // Configuration verification
    bool collectorExists(const std::string& collector_id);
    std::string getCollectorStatus(const std::string& collector_id);
    bool configurationExists(const std::string& collector_id);

    // Data manipulation
    void clearAllMetrics();
    void clearMetricsTable(const std::string& table);
    void clearCollectorMetrics(const std::string& collector_id);
    void truncateAllData();

    // Raw query execution
    std::string executeQuery(const std::string& sql, bool use_metrics_db = true);
    bool executeUpdate(const std::string& sql, bool use_metrics_db = true);

    // Connection testing
    bool isConnected();
    bool testConnection();

private:
    std::string m_main_db_url;
    std::string m_metrics_db_url;
    bool m_connected;

    // Helper for executing psql commands
    bool executePsqlCommand(const std::string& cmd, const std::string& db_url, std::string& output);
};

