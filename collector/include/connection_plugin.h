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
 * PostgreSQL Connection Collector
 *
 * Tracks connection and session metrics:
 * - Active and idle connections
 * - Connection state breakdown
 * - Long-running transaction detection
 * - Connection duration analysis
 * - Idle transaction sessions
 *
 * Requirements:
 * - PostgreSQL 9.0+ (pg_stat_activity views)
 * - No special permissions required (basic info)
 * - Superuser for full details (query text)
 *
 * Metrics collected:
 * - Connection count by database
 * - Connection state breakdown
 * - Long-running transactions
 * - Idle transactions
 * - Session duration metrics
 */
class PgConnectionCollector : public Collector {
public:
    explicit PgConnectionCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_connections"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    json collectConnectionStats(const std::string& dbname);
    json collectLongRunningTransactions(const std::string& dbname);
    json collectIdleTransactions(const std::string& dbname);
};
