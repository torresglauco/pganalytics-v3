#include "../include/cache_hit_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>

PgCacheHitCollector::PgCacheHitCollector(
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

json PgCacheHitCollector::collectTableCacheHit(const std::string& dbname) {
    json cache = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return cache;

    const char* query = R"(
        SELECT
            schemaname,
            tablename,
            heap_blks_hit,
            heap_blks_read,
            heap_blks_hit + heap_blks_read as total_heap_blks,
            ROUND(100.0 * heap_blks_hit / NULLIF(heap_blks_hit + heap_blks_read, 0), 2) as cache_hit_ratio,
            idx_blks_hit,
            idx_blks_read,
            idx_blks_hit + idx_blks_read as total_idx_blks,
            ROUND(100.0 * idx_blks_hit / NULLIF(idx_blks_hit + idx_blks_read, 0), 2) as idx_cache_hit_ratio,
            toast_blks_hit,
            toast_blks_read,
            tidx_blks_hit,
            tidx_blks_read
        FROM pg_statio_user_tables
        WHERE (heap_blks_hit + heap_blks_read) > 0
        ORDER BY schemaname, tablename
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json item = json::object();
            item["schema"] = PQgetvalue(result, i, 0);
            item["table"] = PQgetvalue(result, i, 1);
            item["heap_blks_hit"] = std::stoll(PQgetvalue(result, i, 2));
            item["heap_blks_read"] = std::stoll(PQgetvalue(result, i, 3));
            item["total_heap_blks"] = std::stoll(PQgetvalue(result, i, 4));
            item["heap_cache_hit_ratio"] = std::stod(PQgetvalue(result, i, 5));
            item["idx_blks_hit"] = std::stoll(PQgetvalue(result, i, 6));
            item["idx_blks_read"] = std::stoll(PQgetvalue(result, i, 7));
            item["total_idx_blks"] = std::stoll(PQgetvalue(result, i, 8));
            item["idx_cache_hit_ratio"] = std::stod(PQgetvalue(result, i, 9));
            item["toast_blks_hit"] = std::stoll(PQgetvalue(result, i, 10));
            item["toast_blks_read"] = std::stoll(PQgetvalue(result, i, 11));
            item["tidx_blks_hit"] = std::stoll(PQgetvalue(result, i, 12));
            item["tidx_blks_read"] = std::stoll(PQgetvalue(result, i, 13));

            cache.push_back(item);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return cache;
}

json PgCacheHitCollector::collectIndexCacheHit(const std::string& dbname) {
    json cache = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return cache;

    const char* query = R"(
        SELECT
            schemaname,
            tablename,
            indexrelname,
            idx_blks_hit,
            idx_blks_read,
            idx_blks_hit + idx_blks_read as total_blks,
            ROUND(100.0 * idx_blks_hit / NULLIF(idx_blks_hit + idx_blks_read, 0), 2) as cache_hit_ratio
        FROM pg_statio_user_indexes
        WHERE (idx_blks_hit + idx_blks_read) > 0
        ORDER BY schemaname, tablename, indexrelname
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json item = json::object();
            item["schema"] = PQgetvalue(result, i, 0);
            item["table"] = PQgetvalue(result, i, 1);
            item["index"] = PQgetvalue(result, i, 2);
            item["blks_hit"] = std::stoll(PQgetvalue(result, i, 3));
            item["blks_read"] = std::stoll(PQgetvalue(result, i, 4));
            item["total_blks"] = std::stoll(PQgetvalue(result, i, 5));
            item["cache_hit_ratio"] = std::stod(PQgetvalue(result, i, 6));

            cache.push_back(item);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return cache;
}

json PgCacheHitCollector::execute() {
    auto timestamp = getCurrentTimestamp();

    json result = {
        {"type", "pg_cache"},
        {"timestamp", timestamp},
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"databases", json::object()}
    };

    if (!enabled_) {
        result["error"] = "Cache hit collector is disabled";
        return result;
    }

#ifdef HAVE_LIBPQ
    for (const auto& dbname : databases_) {
        json dbCache = json::object();

        dbCache["table_cache_hit"] = collectTableCacheHit(dbname);
        dbCache["index_cache_hit"] = collectIndexCacheHit(dbname);

        result["databases"][dbname] = dbCache;
    }
#else
    result["error"] = "libpq not available";
#endif

    return result;
}

#else
PgCacheHitCollector::PgCacheHitCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword,
    const std::vector<std::string>& databases
) : enabled_(false) {
    std::cerr << "WARNING: Cache hit collector requires libpq" << std::endl;
}

json PgCacheHitCollector::execute() {
    return json::object({
        {"type", "pg_cache"},
        {"error", "libpq not available"}
    });
}

json PgCacheHitCollector::collectTableCacheHit(const std::string& dbname) { return json::array(); }
json PgCacheHitCollector::collectIndexCacheHit(const std::string& dbname) { return json::array(); }

#endif
