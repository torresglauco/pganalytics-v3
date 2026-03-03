#include "../include/lock_plugin.h"
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
PgLockCollector::PgLockCollector(
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
 * Collect active locks from pg_locks
 */
json PgLockCollector::collectActiveLocks(const std::string& dbname) {
    json locks = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return locks;

    // Query for active locks
    const char* query = R"(
        SELECT
            l.pid,
            l.usesysid,
            l.database,
            l.relation,
            l.page,
            l.tuple,
            l.virtualxid,
            l.transactionid,
            l.classid,
            l.objid,
            l.objsubid,
            l.locktype,
            l.mode,
            l.granted,
            EXTRACT(EPOCH FROM (NOW() - a.query_start)) as lock_age_seconds,
            a.usename,
            a.state,
            a.query
        FROM pg_locks l
        LEFT JOIN pg_stat_activity a ON l.pid = a.pid
        ORDER BY l.pid
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json lock = json::object();
            lock["pid"] = std::stoi(PQgetvalue(result, i, 0));

            if (!PQgetisnull(result, i, 1)) {
                lock["usesysid"] = std::stoi(PQgetvalue(result, i, 1));
            }
            if (!PQgetisnull(result, i, 2)) {
                lock["database"] = std::stoi(PQgetvalue(result, i, 2));
            }
            if (!PQgetisnull(result, i, 3)) {
                lock["relation"] = std::stoi(PQgetvalue(result, i, 3));
            }
            if (!PQgetisnull(result, i, 4)) {
                lock["page"] = std::stoi(PQgetvalue(result, i, 4));
            }
            if (!PQgetisnull(result, i, 5)) {
                lock["tuple"] = std::stoi(PQgetvalue(result, i, 5));
            }
            if (!PQgetisnull(result, i, 6)) {
                lock["virtualxid"] = PQgetvalue(result, i, 6);
            }
            if (!PQgetisnull(result, i, 7)) {
                lock["transactionid"] = std::stoll(PQgetvalue(result, i, 7));
            }
            if (!PQgetisnull(result, i, 8)) {
                lock["classid"] = std::stoi(PQgetvalue(result, i, 8));
            }
            if (!PQgetisnull(result, i, 9)) {
                lock["objid"] = std::stoi(PQgetvalue(result, i, 9));
            }
            if (!PQgetisnull(result, i, 10)) {
                lock["objsubid"] = std::stoi(PQgetvalue(result, i, 10));
            }

            lock["locktype"] = PQgetvalue(result, i, 11);
            lock["mode"] = PQgetvalue(result, i, 12);
            lock["granted"] = std::string(PQgetvalue(result, i, 13)) == "t";

            if (!PQgetisnull(result, i, 14)) {
                lock["lock_age_seconds"] = std::stod(PQgetvalue(result, i, 14));
            }
            if (!PQgetisnull(result, i, 15)) {
                lock["username"] = PQgetvalue(result, i, 15);
            }
            if (!PQgetisnull(result, i, 16)) {
                lock["state"] = PQgetvalue(result, i, 16);
            }
            if (!PQgetisnull(result, i, 17)) {
                lock["query"] = PQgetvalue(result, i, 17);
            }

            locks.push_back(lock);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return locks;
}

/**
 * Collect lock wait chains and blocking detection
 */
json PgLockCollector::collectLockWaitChains(const std::string& dbname) {
    json waits = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return waits;

    // Query for lock wait chains
    const char* query = R"(
        SELECT
            blocked_locks.pid AS blocked_pid,
            blocked_activity.usename AS blocked_user,
            blocking_locks.pid AS blocking_pid,
            blocking_activity.usename AS blocking_user,
            blocked_activity.query AS blocked_statement,
            blocking_activity.query AS blocking_statement,
            blocked_activity.application_name AS blocked_application,
            blocking_activity.application_name AS blocking_application,
            EXTRACT(EPOCH FROM (NOW() - blocked_activity.query_start)) as wait_time_seconds
        FROM pg_catalog.pg_locks blocked_locks
        JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
        JOIN pg_catalog.pg_locks blocking_locks
            ON blocking_locks.locktype = blocked_locks.locktype
            AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
            AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
            AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
            AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
            AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
            AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
            AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
            AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
            AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
            AND blocking_locks.pid != blocked_locks.pid
        JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
        WHERE NOT blocked_locks.granted
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json wait = json::object();
            wait["blocked_pid"] = std::stoi(PQgetvalue(result, i, 0));
            wait["blocked_user"] = PQgetvalue(result, i, 1);
            wait["blocking_pid"] = std::stoi(PQgetvalue(result, i, 2));
            wait["blocking_user"] = PQgetvalue(result, i, 3);
            wait["blocked_query"] = PQgetvalue(result, i, 4);
            wait["blocking_query"] = PQgetvalue(result, i, 5);
            wait["blocked_application"] = PQgetvalue(result, i, 6);
            wait["blocking_application"] = PQgetvalue(result, i, 7);

            if (!PQgetisnull(result, i, 8)) {
                wait["wait_time_seconds"] = std::stod(PQgetvalue(result, i, 8));
            }

            waits.push_back(wait);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return waits;
}

/**
 * Collect blocking queries information
 */
json PgLockCollector::collectBlockingQueries(const std::string& dbname) {
    json blocking = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return blocking;

    // Query for blocking queries
    const char* query = R"(
        SELECT
            pid,
            usename,
            datname,
            state,
            query,
            query_start,
            state_change,
            EXTRACT(EPOCH FROM (NOW() - query_start)) as duration_seconds,
            application_name,
            client_addr::text,
            backend_start
        FROM pg_stat_activity
        WHERE state != 'idle'
            AND datname = %L
            AND query NOT ILIKE '%pg_sleep%'
        ORDER BY query_start
    )";

    // Build query with database name parameter
    std::string safe_query = "SELECT \
            pid, \
            usename, \
            datname, \
            state, \
            query, \
            query_start, \
            state_change, \
            EXTRACT(EPOCH FROM (NOW() - query_start)) as duration_seconds, \
            application_name, \
            client_addr::text, \
            backend_start \
        FROM pg_stat_activity \
        WHERE state != 'idle' \
            AND datname = '" + dbname + "' \
        ORDER BY query_start";

    PGresult* result = executeQuery(conn, safe_query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json block = json::object();
            block["pid"] = std::stoi(PQgetvalue(result, i, 0));
            block["username"] = PQgetvalue(result, i, 1);
            block["database"] = PQgetvalue(result, i, 2);
            block["state"] = PQgetvalue(result, i, 3);
            block["query"] = PQgetvalue(result, i, 4);
            block["query_start"] = PQgetvalue(result, i, 5);
            block["state_change"] = PQgetvalue(result, i, 6);

            if (!PQgetisnull(result, i, 7)) {
                block["duration_seconds"] = std::stod(PQgetvalue(result, i, 7));
            }

            block["application_name"] = PQgetvalue(result, i, 8);

            if (!PQgetisnull(result, i, 9)) {
                block["client_address"] = PQgetvalue(result, i, 9);
            }

            block["backend_start"] = PQgetvalue(result, i, 10);

            blocking.push_back(block);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return blocking;
}

/**
 * Execute collector
 */
json PgLockCollector::execute() {
    auto timestamp = getCurrentTimestamp();

    json result = {
        {"type", "pg_locks"},
        {"timestamp", timestamp},
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"databases", json::object()}
    };

    if (!enabled_) {
        result["error"] = "Lock collector is disabled";
        return result;
    }

#ifdef HAVE_LIBPQ
    for (const auto& dbname : databases_) {
        json dbLocks = json::object();

        dbLocks["active_locks"] = collectActiveLocks(dbname);
        dbLocks["lock_wait_chains"] = collectLockWaitChains(dbname);
        dbLocks["blocking_queries"] = collectBlockingQueries(dbname);

        result["databases"][dbname] = dbLocks;
    }
#else
    result["error"] = "libpq not available";
#endif

    return result;
}

#else
// Stub implementations when libpq is not available
PgLockCollector::PgLockCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword,
    const std::vector<std::string>& databases
) : enabled_(false) {
    std::cerr << "WARNING: Lock collector requires libpq" << std::endl;
}

json PgLockCollector::execute() {
    return json::object({
        {"type", "pg_locks"},
        {"error", "libpq not available"}
    });
}

json PgLockCollector::collectActiveLocks(const std::string& dbname) { return json::array(); }
json PgLockCollector::collectLockWaitChains(const std::string& dbname) { return json::array(); }
json PgLockCollector::collectBlockingQueries(const std::string& dbname) { return json::array(); }

#endif
