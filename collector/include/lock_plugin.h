#pragma once

#include "collector.h"
#include <string>
#include <vector>
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
 * PostgreSQL Lock Collector
 *
 * Gathers comprehensive lock monitoring information including:
 * - Active locks from pg_locks
 * - Lock wait chains and blocking detection
 * - Blocking queries and session information
 * - Lock mode and transaction details
 * - Lock contention metrics
 *
 * Requirements:
 * - PostgreSQL 8.1+ (pg_locks available)
 * - Permissions: Superuser or pg_monitor role for full information
 *
 * Metrics collected:
 * - Active locks (mode, type, granted status)
 * - Blocking relationships (who is blocking whom)
 * - Waiting sessions (query, transaction age)
 * - Lock modes (AccessShare, RowShare, RowExclusive, etc.)
 * - Lock types (relation, extend, page, tuple, etc.)
 */
class PgLockCollector : public Collector {
public:
    explicit PgLockCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_locks"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    // Collection methods
    json collectActiveLocks(const std::string& dbname);
    json collectLockWaitChains(const std::string& dbname);
    json collectBlockingQueries(const std::string& dbname);
};
