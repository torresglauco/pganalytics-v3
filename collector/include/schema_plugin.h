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
 * PostgreSQL Schema Collector
 *
 * Gathers comprehensive schema information including:
 * - Table definitions (columns, data types, constraints)
 * - Column metadata (nullability, defaults, constraints)
 * - Table constraints (PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK)
 * - Foreign key relationships and their definitions
 * - Index definitions and properties
 * - Trigger information
 *
 * Requirements:
 * - PostgreSQL 8.0+ (information_schema available)
 * - No special permissions required (information_schema is public)
 *
 * Metrics collected:
 * - Column definitions (table, column name, data type, constraints)
 * - Table constraints (type, definition, column involvement)
 * - Foreign key relationships (source and target columns)
 * - Trigger definitions (timing, events, function reference)
 * - Index definitions (name, columns, uniqueness, partial predicates)
 */
class PgSchemaCollector : public Collector {
public:
    explicit PgSchemaCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost,
        int postgresPort,
        const std::string& postgresUser,
        const std::string& postgresPassword,
        const std::vector<std::string>& databases
    );

    json execute() override;
    std::string getType() const override { return "pg_schema"; }
    bool isEnabled() const override { return enabled_; }

private:
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    std::vector<std::string> databases_;
    bool enabled_;

    // Collection methods for different schema components
    json collectTableSchema(const std::string& dbname);
    json collectColumnInfo(const std::string& dbname);
    json collectTableConstraints(const std::string& dbname);
    json collectForeignKeys(const std::string& dbname);
    json collectIndexInfo(const std::string& dbname);
    json collectTriggerInfo(const std::string& dbname);
};
