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
 * PostgreSQL Bloat Collector
 *
 * Analyzes and collects bloat metrics for tables and indexes:
 * - Table bloat percentage and dead tuples
 * - Index bloat percentage
 * - Space reclamable metrics
 * - Bloat effectiveness tracking
 *
 * Requirements:
 * - PostgreSQL 8.2+ (pg_stat_user_tables available)
 * - No special permissions required
 *
 * Metrics collected:
 * - Table bloat ratio
 * - Dead tuple count and percentage
 * - Index bloat analysis
 * - Space reclamation potential
 */
class PgBloatCollector : public Collector {
public:
    explicit PgBloatCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_bloat"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    json collectTableBloat(const std::string& dbname);
    json collectIndexBloat(const std::string& dbname);
};
