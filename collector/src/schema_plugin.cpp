#include "../include/schema_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>

/**
 * Constructor
 */
PgSchemaCollector::PgSchemaCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword,
    const std::vector<std::string>& databases
)
    : postgresHost_(postgresHost),
      postgresPort_(postgresPort),
      postgresUser_(postgresUser),
      postgresPassword_(postgresPassword),
      databases_(databases),
      enabled_(true) {
    hostname_ = hostname;
    collectorId_ = collectorId;
}

/**
 * Get current timestamp in ISO8601 format
 */
static std::string getCurrentTimestamp() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");
    return oss.str();
}

/**
 * Connect to a PostgreSQL database
 */
static PGconn* connectToDatabase(
    const std::string& host,
    int port,
    const std::string& user,
    const std::string& password,
    const std::string& database) {

    std::string connStr = "host=" + host + " port=" + std::to_string(port) +
                         " user=" + user + " password=" + password +
                         " dbname=" + database;

    PGconn* conn = PQconnectdb(connStr.c_str());
    if (PQstatus(conn) != CONNECTION_OK) {
        std::cerr << "ERROR: Connection failed: " << PQerrorMessage(conn) << std::endl;
        PQfinish(conn);
        return nullptr;
    }
    return conn;
}

/**
 * Execute a query and return result
 */
static PGresult* executeQuery(PGconn* conn, const std::string& query) {
    PGresult* result = PQexec(conn, query.c_str());
    if (PQresultStatus(result) != PGRES_TUPLES_OK) {
        std::cerr << "ERROR: Query failed: " << PQerrorMessage(conn) << std::endl;
        std::cerr << "Query: " << query << std::endl;
        PQclear(result);
        return nullptr;
    }
    return result;
}

/**
 * Collect column information from information_schema
 */
json PgSchemaCollector::collectColumnInfo(const std::string& dbname) {
    json columns = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return columns;

    // Query for column information
    const char* query = R"(
        SELECT
            table_schema,
            table_name,
            column_name,
            data_type,
            is_nullable,
            column_default,
            ordinal_position,
            character_maximum_length,
            numeric_precision,
            numeric_scale
        FROM information_schema.columns
        WHERE table_schema NOT IN ('pg_catalog', 'information_schema', 'pganalytics')
        ORDER BY table_schema, table_name, ordinal_position
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json column = json::object();
            column["schema"] = PQgetvalue(result, i, 0);
            column["table"] = PQgetvalue(result, i, 1);
            column["name"] = PQgetvalue(result, i, 2);
            column["data_type"] = PQgetvalue(result, i, 3);
            column["is_nullable"] = std::string(PQgetvalue(result, i, 4)) == "YES";
            column["default"] = PQgetvalue(result, i, 5);
            column["position"] = std::stoi(PQgetvalue(result, i, 6));

            // Optional fields (may be NULL)
            if (!PQgetisnull(result, i, 7)) {
                column["max_length"] = std::stoi(PQgetvalue(result, i, 7));
            }
            if (!PQgetisnull(result, i, 8)) {
                column["numeric_precision"] = std::stoi(PQgetvalue(result, i, 8));
            }
            if (!PQgetisnull(result, i, 9)) {
                column["numeric_scale"] = std::stoi(PQgetvalue(result, i, 9));
            }

            columns.push_back(column);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return columns;
}

/**
 * Collect table constraint information
 */
json PgSchemaCollector::collectTableConstraints(const std::string& dbname) {
    json constraints = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return constraints;

    // Query for constraint information
    const char* query = R"(
        SELECT
            tc.table_schema,
            tc.table_name,
            tc.constraint_name,
            tc.constraint_type,
            string_agg(kcu.column_name, ',' ORDER BY kcu.ordinal_position) as columns
        FROM information_schema.table_constraints tc
        LEFT JOIN information_schema.key_column_usage kcu
            ON tc.constraint_name = kcu.constraint_name
            AND tc.table_schema = kcu.table_schema
        WHERE tc.table_schema NOT IN ('pg_catalog', 'information_schema', 'pganalytics')
        GROUP BY tc.table_schema, tc.table_name, tc.constraint_name, tc.constraint_type
        ORDER BY tc.table_schema, tc.table_name
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json constraint = json::object();
            constraint["schema"] = PQgetvalue(result, i, 0);
            constraint["table"] = PQgetvalue(result, i, 1);
            constraint["name"] = PQgetvalue(result, i, 2);
            constraint["type"] = PQgetvalue(result, i, 3);
            constraint["columns"] = PQgetvalue(result, i, 4);

            constraints.push_back(constraint);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return constraints;
}

/**
 * Collect foreign key information
 */
json PgSchemaCollector::collectForeignKeys(const std::string& dbname) {
    json fkeys = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return fkeys;

    // Query for foreign key information
    const char* query = R"(
        SELECT
            tc.table_schema,
            tc.table_name,
            kcu.column_name,
            ccu.table_schema as referenced_schema,
            ccu.table_name as referenced_table,
            ccu.column_name as referenced_column,
            rc.update_rule,
            rc.delete_rule
        FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu
            ON tc.constraint_name = kcu.constraint_name
            AND tc.table_schema = kcu.table_schema
        JOIN information_schema.constraint_column_usage ccu
            ON ccu.constraint_name = tc.constraint_name
            AND ccu.table_schema = tc.table_schema
        JOIN information_schema.referential_constraints rc
            ON rc.constraint_name = tc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY'
            AND tc.table_schema NOT IN ('pg_catalog', 'information_schema', 'pganalytics')
        ORDER BY tc.table_schema, tc.table_name
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json fkey = json::object();
            fkey["schema"] = PQgetvalue(result, i, 0);
            fkey["table"] = PQgetvalue(result, i, 1);
            fkey["column"] = PQgetvalue(result, i, 2);
            fkey["referenced_schema"] = PQgetvalue(result, i, 3);
            fkey["referenced_table"] = PQgetvalue(result, i, 4);
            fkey["referenced_column"] = PQgetvalue(result, i, 5);
            fkey["update_rule"] = PQgetvalue(result, i, 6);
            fkey["delete_rule"] = PQgetvalue(result, i, 7);

            fkeys.push_back(fkey);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return fkeys;
}

/**
 * Collect index information with sizes and usage stats (INV-03)
 */
json PgSchemaCollector::collectIndexInfo(const std::string& dbname) {
    json indexes = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return indexes;

    // Query for index information with size in MB and OID (INV-03)
    const char* query = R"(
        SELECT
            schemaname,
            relname,
            indexrelname,
            indexdef,
            idx_scan,
            idx_tup_read,
            idx_tup_fetch,
            pg_relation_size(indexrelid) / 1024 / 1024 as index_size_mb,
            pg_size_pretty(pg_relation_size(indexrelid)) as index_size_pretty,
            CASE WHEN idx_scan = 0 THEN 'UNUSED'
                 WHEN idx_scan < 100 THEN 'RARELY_USED'
                 ELSE 'ACTIVE' END as usage_status,
            indisprimary as is_primary,
            indisunique as is_unique,
            indexrelid as index_oid
        FROM pg_stat_user_indexes
        JOIN pg_index ON pg_index.indexrelid = pg_stat_user_indexes.indexrelid
        ORDER BY schemaname, relname, indexrelname
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json idx = json::object();
            idx["schema"] = PQgetvalue(result, i, 0);
            idx["table"] = PQgetvalue(result, i, 1);
            idx["name"] = PQgetvalue(result, i, 2);
            idx["definition"] = PQgetvalue(result, i, 3);
            idx["scans"] = std::stoll(PQgetvalue(result, i, 4));
            idx["tuples_read"] = std::stoll(PQgetvalue(result, i, 5));
            idx["tuples_fetched"] = std::stoll(PQgetvalue(result, i, 6));
            // INV-03: Add index size in MB for inventory
            idx["index_size_mb"] = std::stoll(PQgetvalue(result, i, 7));
            idx["size"] = PQgetvalue(result, i, 8);  // pretty size for display
            idx["usage_status"] = PQgetvalue(result, i, 9);
            // INV-03: Add is_primary, is_unique, and index_oid
            idx["is_primary"] = std::string(PQgetvalue(result, i, 10)) == "t";
            idx["is_unique"] = std::string(PQgetvalue(result, i, 11)) == "t";
            idx["index_oid"] = std::stoull(PQgetvalue(result, i, 12));

            indexes.push_back(idx);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return indexes;
}

/**
 * Collect trigger information
 */
json PgSchemaCollector::collectTriggerInfo(const std::string& dbname) {
    json triggers = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return triggers;

    // Query for trigger information
    const char* query = R"(
        SELECT
            trigger_schema,
            trigger_name,
            event_manipulation,
            event_object_schema,
            event_object_table,
            action_timing,
            action_orientation,
            action_statement
        FROM information_schema.triggers
        WHERE trigger_schema NOT IN ('pg_catalog', 'information_schema', 'pganalytics')
        ORDER BY trigger_schema, trigger_name
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json trigger = json::object();
            trigger["schema"] = PQgetvalue(result, i, 0);
            trigger["name"] = PQgetvalue(result, i, 1);
            trigger["event"] = PQgetvalue(result, i, 2);
            trigger["table_schema"] = PQgetvalue(result, i, 3);
            trigger["table"] = PQgetvalue(result, i, 4);
            trigger["timing"] = PQgetvalue(result, i, 5);
            trigger["orientation"] = PQgetvalue(result, i, 6);
            trigger["statement"] = PQgetvalue(result, i, 7);

            triggers.push_back(trigger);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return triggers;
}

/**
 * Collect table schema information with sizes and row counts (INV-01)
 */
json PgSchemaCollector::collectTableSchema(const std::string& dbname) {
    json tables = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return tables;

    // Query for table information with sizes from pg_stat_user_tables and pg_class
    // INV-01: Table inventory with row counts and sizes
    const char* query = R"(
        SELECT
            n.nspname as schema_name,
            c.relname as table_name,
            CASE c.relkind WHEN 'r' THEN 'BASE TABLE' WHEN 'v' THEN 'VIEW' END as table_type,
            COALESCE(s.n_live_tup, 0) as row_count,
            pg_total_relation_size(c.oid) / 1024 / 1024 as total_size_mb,
            pg_relation_size(c.oid) / 1024 / 1024 as table_size_mb,
            COALESCE((SELECT sum(pg_relation_size(indexrelid)) / 1024 / 1024
                     FROM pg_index WHERE indrelid = c.oid), 0) as index_size_mb,
            COALESCE(pg_relation_size(c.reltoastrelid) / 1024 / 1024, 0) as toast_size_mb,
            c.reloptions IS NOT NULL AND array_to_string(c.reloptions, ',') LIKE '%oids=on%' as has_oids,
            c.oid as table_oid
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        LEFT JOIN pg_stat_user_tables s ON s.relid = c.oid
        WHERE c.relkind IN ('r', 'v')
            AND n.nspname NOT IN ('pg_catalog', 'information_schema', 'pganalytics')
        ORDER BY n.nspname, c.relname
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json table = json::object();
            table["schema"] = PQgetvalue(result, i, 0);
            table["name"] = PQgetvalue(result, i, 1);
            table["type"] = PQgetvalue(result, i, 2);
            // INV-01: Add inventory fields
            table["row_count"] = std::stoll(PQgetvalue(result, i, 3));
            table["total_size_mb"] = std::stoll(PQgetvalue(result, i, 4));
            table["table_size_mb"] = std::stoll(PQgetvalue(result, i, 5));
            table["index_size_mb"] = std::stoll(PQgetvalue(result, i, 6));
            table["toast_size_mb"] = std::stoll(PQgetvalue(result, i, 7));
            table["has_oids"] = std::string(PQgetvalue(result, i, 8)) == "t";
            table["table_oid"] = std::stoull(PQgetvalue(result, i, 9));

            tables.push_back(table);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return tables;
}

/**
 * Execute collector
 */
json PgSchemaCollector::execute() {
    auto timestamp = getCurrentTimestamp();

    json result = {
        {"type", "pg_schema"},
        {"timestamp", timestamp},
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"databases", json::object()}
    };

    if (!enabled_) {
        result["error"] = "Schema collector is disabled";
        return result;
    }

#ifdef HAVE_LIBPQ
    for (const auto& dbname : databases_) {
        json dbSchema = json::object();

        dbSchema["tables"] = collectTableSchema(dbname);
        dbSchema["columns"] = collectColumnInfo(dbname);
        dbSchema["constraints"] = collectTableConstraints(dbname);
        dbSchema["foreign_keys"] = collectForeignKeys(dbname);
        dbSchema["indexes"] = collectIndexInfo(dbname);
        dbSchema["triggers"] = collectTriggerInfo(dbname);

        result["databases"][dbname] = dbSchema;
    }
#else
    result["error"] = "libpq not available";
#endif

    return result;
}

#else
// Stub implementations when libpq is not available
PgSchemaCollector::PgSchemaCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword,
    const std::vector<std::string>& databases
) : enabled_(false) {
    std::cerr << "WARNING: Schema collector requires libpq" << std::endl;
}

json PgSchemaCollector::execute() {
    return json::object({
        {"type", "pg_schema"},
        {"error", "libpq not available"}
    });
}

json PgSchemaCollector::collectTableSchema(const std::string& dbname) { return json::array(); }
json PgSchemaCollector::collectColumnInfo(const std::string& dbname) { return json::array(); }
json PgSchemaCollector::collectTableConstraints(const std::string& dbname) { return json::array(); }
json PgSchemaCollector::collectForeignKeys(const std::string& dbname) { return json::array(); }
json PgSchemaCollector::collectIndexInfo(const std::string& dbname) { return json::array(); }
json PgSchemaCollector::collectTriggerInfo(const std::string& dbname) { return json::array(); }

#endif
