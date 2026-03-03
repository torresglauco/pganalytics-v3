#include "../include/bloat_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>

PgBloatCollector::PgBloatCollector(
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

static std::string getCurrentTimestamp() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");
    return oss.str();
}

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

static PGresult* executeQuery(PGconn* conn, const std::string& query) {
    PGresult* result = PQexec(conn, query.c_str());
    if (PQresultStatus(result) != PGRES_TUPLES_OK) {
        std::cerr << "ERROR: Query failed: " << PQerrorMessage(conn) << std::endl;
        PQclear(result);
        return nullptr;
    }
    return result;
}

json PgBloatCollector::collectTableBloat(const std::string& dbname) {
    json bloat = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return bloat;

    const char* query = R"(
        SELECT
            schemaname,
            tablename,
            n_dead_tup,
            n_live_tup,
            ROUND(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) as dead_ratio,
            pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
            ROUND(100.0 * pg_relation_size(schemaname||'.'||tablename) * n_dead_tup /
                NULLIF(pg_relation_size(schemaname||'.'||tablename) * (n_live_tup + n_dead_tup), 0), 2) as space_wasted_percent,
            last_vacuum,
            last_autovacuum,
            vacuum_count,
            autovacuum_count
        FROM pg_stat_user_tables
        WHERE n_dead_tup > 0
        ORDER BY n_dead_tup DESC
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json item = json::object();
            item["schema"] = PQgetvalue(result, i, 0);
            item["table"] = PQgetvalue(result, i, 1);
            item["dead_tuples"] = std::stoll(PQgetvalue(result, i, 2));
            item["live_tuples"] = std::stoll(PQgetvalue(result, i, 3));
            item["dead_ratio_percent"] = std::stod(PQgetvalue(result, i, 4));
            item["table_size"] = PQgetvalue(result, i, 5);

            if (!PQgetisnull(result, i, 6)) {
                item["space_wasted_percent"] = std::stod(PQgetvalue(result, i, 6));
            }

            if (!PQgetisnull(result, i, 7)) {
                item["last_vacuum"] = PQgetvalue(result, i, 7);
            }
            if (!PQgetisnull(result, i, 8)) {
                item["last_autovacuum"] = PQgetvalue(result, i, 8);
            }

            item["vacuum_count"] = std::stoll(PQgetvalue(result, i, 9));
            item["autovacuum_count"] = std::stoll(PQgetvalue(result, i, 10));

            bloat.push_back(item);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return bloat;
}

json PgBloatCollector::collectIndexBloat(const std::string& dbname) {
    json bloat = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return bloat;

    const char* query = R"(
        SELECT
            schemaname,
            tablename,
            indexname,
            idx_scan,
            idx_tup_read,
            idx_tup_fetch,
            pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
            CASE WHEN idx_scan = 0 THEN 'UNUSED'
                 WHEN idx_scan < 100 THEN 'RARELY_USED'
                 ELSE 'ACTIVE' END as usage_status,
            CASE WHEN idx_scan = 0 THEN 'CONSIDER_DROPPING'
                 ELSE 'IN_USE' END as recommendation
        FROM pg_stat_user_indexes
        ORDER BY pg_relation_size(indexrelid) DESC
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json item = json::object();
            item["schema"] = PQgetvalue(result, i, 0);
            item["table"] = PQgetvalue(result, i, 1);
            item["index_name"] = PQgetvalue(result, i, 2);
            item["scans"] = std::stoll(PQgetvalue(result, i, 3));
            item["tuples_read"] = std::stoll(PQgetvalue(result, i, 4));
            item["tuples_fetched"] = std::stoll(PQgetvalue(result, i, 5));
            item["index_size"] = PQgetvalue(result, i, 6);
            item["usage_status"] = PQgetvalue(result, i, 7);
            item["recommendation"] = PQgetvalue(result, i, 8);

            bloat.push_back(item);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return bloat;
}

json PgBloatCollector::execute() {
    auto timestamp = getCurrentTimestamp();

    json result = {
        {"type", "pg_bloat"},
        {"timestamp", timestamp},
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"databases", json::object()}
    };

    if (!enabled_) {
        result["error"] = "Bloat collector is disabled";
        return result;
    }

#ifdef HAVE_LIBPQ
    for (const auto& dbname : databases_) {
        json dbBloat = json::object();

        dbBloat["table_bloat"] = collectTableBloat(dbname);
        dbBloat["index_bloat"] = collectIndexBloat(dbname);

        result["databases"][dbname] = dbBloat;
    }
#else
    result["error"] = "libpq not available";
#endif

    return result;
}

#else
PgBloatCollector::PgBloatCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword,
    const std::vector<std::string>& databases
) : enabled_(false) {
    std::cerr << "WARNING: Bloat collector requires libpq" << std::endl;
}

json PgBloatCollector::execute() {
    return json::object({
        {"type", "pg_bloat"},
        {"error", "libpq not available"}
    });
}

json PgBloatCollector::collectTableBloat(const std::string& dbname) { return json::array(); }
json PgBloatCollector::collectIndexBloat(const std::string& dbname) { return json::array(); }

#endif
