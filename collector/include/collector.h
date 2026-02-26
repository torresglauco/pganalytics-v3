#pragma once

#include <string>
#include <vector>
#include <memory>
#include <map>
#include <chrono>
#include <nlohmann/json.hpp>
#include "thread_pool.h"

using json = nlohmann::json;

// Forward declarations
class Collector;
class PgStatsCollector;
class DiskUsageCollector;
class PgLogCollector;
class SysstatCollector;
class PgReplicationCollector;

/**
 * Base Collector interface
 */
class Collector {
public:
    virtual ~Collector() = default;

    /**
     * Execute the collector to gather metrics
     * @return JSON object with collected metrics
     */
    virtual json execute() = 0;

    /**
     * Get the type of this collector (pg_stats, disk_usage, pg_log, sysstat, pg_replication)
     */
    virtual std::string getType() const = 0;

    /**
     * Check if this collector is enabled in config
     */
    virtual bool isEnabled() const = 0;

protected:
    std::string hostname_;
    std::string collectorId_;
};

/**
 * PostgreSQL Statistics Collector
 * Gathers table, index, and database-level metrics
 */
class PgStatsCollector : public Collector {
public:
    explicit PgStatsCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_stats"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    json collectDatabaseStats(const std::string& dbname);
    json collectTableStats(const std::string& dbname);
    json collectIndexStats(const std::string& dbname);
    json collectDatabaseGlobalStats();
};

/**
 * System Statistics Collector
 * Gathers CPU, memory, disk I/O metrics
 */
class SysstatCollector : public Collector {
public:
    explicit SysstatCollector(
        const std::string& hostname,
        const std::string& collectorId
    );

    json execute() override;
    std::string getType() const override { return "sysstat"; }
    bool isEnabled() const override { return enabled_; }

private:
    bool enabled_;

    json collectCpuStats();
    json collectMemoryStats();
    json collectIoStats();
    json collectLoadAverage();
};

/**
 * Disk Usage Collector
 * Gathers filesystem usage metrics
 */
class DiskUsageCollector : public Collector {
public:
    explicit DiskUsageCollector(
        const std::string& hostname,
        const std::string& collectorId
    );

    json execute() override;
    std::string getType() const override { return "disk_usage"; }
    bool isEnabled() const override { return enabled_; }

private:
    bool enabled_;

    json collectDiskUsage();
};

/**
 * PostgreSQL Log Collector
 * Gathers PostgreSQL server logs
 */
class PgLogCollector : public Collector {
public:
    explicit PgLogCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword
    );

    json execute() override;
    std::string getType() const override { return "pg_log"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    bool enabled_;

    json collectLogs();
};

// Forward declaration only - full definition in replication_plugin.h
// (Replication collector is defined separately due to complex data structures)

/**
 * Collector Manager
 * Orchestrates all collectors and combines their output
 */
class CollectorManager {
public:
    explicit CollectorManager(const std::string& hostname, const std::string& collectorId);

    void addCollector(std::shared_ptr<Collector> collector);

    /**
     * Collect metrics from all enabled collectors (SEQUENTIAL MODE)
     * Used when thread pool is disabled or not available
     * @return JSON object with collected metrics from all collectors
     */
    json collectAll();

    /**
     * Collect metrics from all enabled collectors (PARALLEL MODE - Phase 1.1)
     * Uses thread pool for parallel execution
     * Expected improvement: 75% cycle time reduction (57.7s → 14.4s at 100 collectors)
     * @return JSON object with collected metrics from all collectors
     */
    json collectAllParallel();

    void configure(const json& config);

    /**
     * Get collection cycle time in milliseconds (for monitoring)
     */
    int getLastCycleTimeMs() const {
        return last_cycle_time_ms_;
    }

private:
    std::string hostname_;
    std::string collectorId_;
    std::vector<std::shared_ptr<Collector>> collectors_;

    // Phase 1.1: Thread pool for parallel collector execution
    // Pool size configurable via config file (default: 4, min: 1, max: 16)
    std::unique_ptr<ThreadPool> thread_pool_;
    int thread_pool_size_;
    int last_cycle_time_ms_;

    /**
     * Initialize thread pool (called in constructor)
     */
    void initializeThreadPool();
};
