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
 * PostgreSQL Extension Collector
 *
 * Collects database extensions and modules information:
 * - Installed extensions list
 * - Extension version tracking
 * - Extension owners and schemas
 * - Extension configuration
 *
 * Requirements:
 * - PostgreSQL 9.1+ (pg_extension available)
 * - No special permissions required
 *
 * Metrics collected:
 * - Extension names and versions
 * - Extension schemas
 * - Extension owners
 * - Extension availability
 */
class PgExtensionCollector : public Collector {
public:
    explicit PgExtensionCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_extensions"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    json collectExtensionInfo(const std::string& dbname);
};
