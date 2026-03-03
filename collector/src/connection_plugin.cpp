#include "../include/connection_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>

PgConnectionCollector::PgConnectionCollector(
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

json PgConnectionCollector::collectConnectionStats(const std::string& dbname) {
    json stats = json::object();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return stats;

    const char* query = R"(
        SELECT
            datname,
            state,
            COUNT(*) as count,
            MAX(EXTRACT(EPOCH FROM (NOW() - backend_start))) as max_age_seconds,
            MIN(EXTRACT(EPOCH FROM (NOW() - backend_start))) as min_age_seconds
        FROM pg_stat_activity
        WHERE datname IS NOT NULL
        GROUP BY datname, state
        ORDER BY datname, state
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        json states = json::array();
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json state = json::object();
            state["database"] = PQgetvalue(result, i, 0);
            state["state"] = PQgetvalue(result, i, 1);
            state["count"] = std::stoi(PQgetvalue(result, i, 2));

            if (!PQgetisnull(result, i, 3)) {
                state["max_age_seconds"] = std::stod(PQgetvalue(result, i, 3));
            }
            if (!PQgetisnull(result, i, 4)) {
                state["min_age_seconds"] = std::stod(PQgetvalue(result, i, 4));
            }

            states.push_back(state);
        }
        stats["by_state"] = states;
        PQclear(result);
    }

    // Get total connection count
    const char* total_query = "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = $1";
    std::string safe_query = "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '" + dbname + "'";
    PGresult* total_result = executeQuery(conn, safe_query);
    if (total_result) {
        stats["total_connections"] = std::stoi(PQgetvalue(total_result, 0, 0));
        PQclear(total_result);
    }

    PQfinish(conn);
#endif

    return stats;
}

json PgConnectionCollector::collectLongRunningTransactions(const std::string& dbname) {
    json transactions = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return transactions;

    const char* query = R"(
        SELECT
            pid,
            usename,
            datname,
            state,
            query,
            query_start,
            EXTRACT(EPOCH FROM (NOW() - query_start)) as duration_seconds,
            application_name,
            client_addr::text,
            backend_start
        FROM pg_stat_activity
        WHERE state NOT IN ('idle', 'idle in transaction')
            AND datname = %L
            AND query_start < NOW() - INTERVAL '5 minutes'
        ORDER BY query_start
    )";

    // Build safe query
    std::string safe_query = "SELECT \
            pid, \
            usename, \
            datname, \
            state, \
            query, \
            query_start, \
            EXTRACT(EPOCH FROM (NOW() - query_start)) as duration_seconds, \
            application_name, \
            client_addr::text, \
            backend_start \
        FROM pg_stat_activity \
        WHERE state NOT IN ('idle', 'idle in transaction') \
            AND datname = '" + dbname + "' \
            AND query_start < NOW() - INTERVAL '5 minutes' \
        ORDER BY query_start";

    PGresult* result = executeQuery(conn, safe_query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json tx = json::object();
            tx["pid"] = std::stoi(PQgetvalue(result, i, 0));
            tx["username"] = PQgetvalue(result, i, 1);
            tx["database"] = PQgetvalue(result, i, 2);
            tx["state"] = PQgetvalue(result, i, 3);
            tx["query"] = PQgetvalue(result, i, 4);
            tx["query_start"] = PQgetvalue(result, i, 5);
            tx["duration_seconds"] = std::stod(PQgetvalue(result, i, 6));
            tx["application_name"] = PQgetvalue(result, i, 7);

            if (!PQgetisnull(result, i, 8)) {
                tx["client_address"] = PQgetvalue(result, i, 8);
            }

            tx["backend_start"] = PQgetvalue(result, i, 9);

            transactions.push_back(tx);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return transactions;
}

json PgConnectionCollector::collectIdleTransactions(const std::string& dbname) {
    json transactions = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return transactions;

    const char* query = R"(
        SELECT
            pid,
            usename,
            datname,
            state,
            query_start,
            state_change,
            EXTRACT(EPOCH FROM (NOW() - state_change)) as idle_time_seconds,
            application_name,
            client_addr::text
        FROM pg_stat_activity
        WHERE state = 'idle in transaction'
            AND datname = %L
            AND state_change < NOW() - INTERVAL '1 minute'
        ORDER BY state_change
    )";

    // Build safe query
    std::string safe_query = "SELECT \
            pid, \
            usename, \
            datname, \
            state, \
            query_start, \
            state_change, \
            EXTRACT(EPOCH FROM (NOW() - state_change)) as idle_time_seconds, \
            application_name, \
            client_addr::text \
        FROM pg_stat_activity \
        WHERE state = 'idle in transaction' \
            AND datname = '" + dbname + "' \
            AND state_change < NOW() - INTERVAL '1 minute' \
        ORDER BY state_change";

    PGresult* result = executeQuery(conn, safe_query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json idle = json::object();
            idle["pid"] = std::stoi(PQgetvalue(result, i, 0));
            idle["username"] = PQgetvalue(result, i, 1);
            idle["database"] = PQgetvalue(result, i, 2);
            idle["state"] = PQgetvalue(result, i, 3);
            idle["query_start"] = PQgetvalue(result, i, 4);
            idle["state_change"] = PQgetvalue(result, i, 5);
            idle["idle_time_seconds"] = std::stod(PQgetvalue(result, i, 6));
            idle["application_name"] = PQgetvalue(result, i, 7);

            if (!PQgetisnull(result, i, 8)) {
                idle["client_address"] = PQgetvalue(result, i, 8);
            }

            transactions.push_back(idle);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return transactions;
}

json PgConnectionCollector::execute() {
    auto timestamp = getCurrentTimestamp();

    json result = {
        {"type", "pg_connections"},
        {"timestamp", timestamp},
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"databases", json::object()}
    };

    if (!enabled_) {
        result["error"] = "Connection collector is disabled";
        return result;
    }

#ifdef HAVE_LIBPQ
    for (const auto& dbname : databases_) {
        json dbConnections = json::object();

        dbConnections["connection_stats"] = collectConnectionStats(dbname);
        dbConnections["long_running_transactions"] = collectLongRunningTransactions(dbname);
        dbConnections["idle_transactions"] = collectIdleTransactions(dbname);

        result["databases"][dbname] = dbConnections;
    }
#else
    result["error"] = "libpq not available";
#endif

    return result;
}

#else
PgConnectionCollector::PgConnectionCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword,
    const std::vector<std::string>& databases
) : enabled_(false) {
    std::cerr << "WARNING: Connection collector requires libpq" << std::endl;
}

json PgConnectionCollector::execute() {
    return json::object({
        {"type", "pg_connections"},
        {"error", "libpq not available"}
    });
}

json PgConnectionCollector::collectConnectionStats(const std::string& dbname) { return json::object(); }
json PgConnectionCollector::collectLongRunningTransactions(const std::string& dbname) { return json::array(); }
json PgConnectionCollector::collectIdleTransactions(const std::string& dbname) { return json::array(); }

#endif
