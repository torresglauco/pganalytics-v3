#pragma once

#include "collector.h"
#include <string>
#include <vector>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>
#else
typedef struct pg_conn PGconn;
typedef struct pg_result PGresult;
#endif

/**
 * PostgreSQL Cache Hit Ratio Collector
 *
 * Collects cache and buffer pool metrics:
 * - Table cache hit ratios
 * - Index cache hit ratios
 * - Buffer pool efficiency metrics
 * - Heap and index block statistics
 *
 * Requirements:
 * - PostgreSQL 8.1+ (pg_statio_user_tables available)
 * - No special permissions required
 *
 * Metrics collected:
 * - Cache hit ratio by table
 * - Cache hit ratio by index
 * - Block read/hit statistics
 * - Buffer efficiency analysis
 */
class PgCacheHitCollector : public Collector {
public:
    explicit PgCacheHitCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_cache"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    json collectTableCacheHit(const std::string& dbname);
    json collectIndexCacheHit(const std::string& dbname);
};
