#include "../include/collector.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>
#endif

PgStatsCollector::PgStatsCollector(
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
 * Helper function to get current ISO8601 timestamp
 */
static std::string getCurrentTimestamp() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");
    return oss.str();
}

/**
 * Helper function to connect to PostgreSQL database
 */
#ifdef HAVE_LIBPQ
static PGconn* connectToDatabase(
    const std::string& host,
    int port,
    const std::string& dbname,
    const std::string& user,
    const std::string& password
) {
    std::string connstr = "host=" + host +
                         " port=" + std::to_string(port) +
                         " dbname=" + dbname +
                         " user=" + user;

    if (!password.empty()) {
        connstr += " password=" + password;
    }

    connstr += " connect_timeout=5";

    PGconn* conn = PQconnectdb(connstr.c_str());

    if (PQstatus(conn) != CONNECTION_OK) {
        std::cerr << "Connection failed: " << PQerrorMessage(conn) << std::endl;
        PQfinish(conn);
        return nullptr;
    }

    return conn;
}
#endif

json PgStatsCollector::execute() {
    json result = {
        {"type", "pg_stats"},
        {"timestamp", getCurrentTimestamp()},
        {"databases", json::array()}
    };

    // Collect stats for each configured database
    for (const auto& dbname : databases_) {
        json db_stats = json::object();
        db_stats["database"] = dbname;
        db_stats["timestamp"] = getCurrentTimestamp();

        // Collect database-level stats
        auto db_info = collectDatabaseStats(dbname);
        if (!db_info.empty()) {
            db_stats.update(db_info);
        }

        // Collect table stats
        auto tables = collectTableStats(dbname);
        if (tables.is_array()) {
            db_stats["tables"] = tables;
        } else {
            db_stats["tables"] = json::array();
        }

        // Collect index stats
        auto indexes = collectIndexStats(dbname);
        if (indexes.is_array()) {
            db_stats["indexes"] = indexes;
        } else {
            db_stats["indexes"] = json::array();
        }

        result["databases"].push_back(db_stats);
    }

    return result;
}

json PgStatsCollector::collectDatabaseStats(const std::string& dbname) {
    json result = json::object();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, dbname,
                                      postgresUser_, postgresPassword_);
    if (!conn) {
        // Return default values if connection fails
        result["size_bytes"] = 0;
        result["transactions_committed"] = 0;
        result["transactions_rolledback"] = 0;
        result["tuples_returned"] = 0;
        result["tuples_fetched"] = 0;
        result["tuples_inserted"] = 0;
        result["tuples_updated"] = 0;
        result["tuples_deleted"] = 0;
        return result;
    }

    // Query: Get database statistics from pg_stat_database
    const char* query =
        "SELECT pg_database_size(datname) as size_bytes, "
        "       xact_commit, xact_rollback, "
        "       tup_returned, tup_fetched, "
        "       tup_inserted, tup_updated, tup_deleted "
        "FROM pg_stat_database "
        "WHERE datname = $1";

    const char* paramValues[] = {dbname.c_str()};
    int paramLengths[] = {(int)dbname.length()};
    int paramFormats[] = {0};

    PGresult* res = PQexecParams(conn, query, 1, nullptr,
                                  paramValues, paramLengths, paramFormats, 0);

    if (PQresultStatus(res) == PGRES_TUPLES_OK && PQntuples(res) > 0) {
        // Extract values from result
        result["size_bytes"] = std::stoll(PQgetvalue(res, 0, 0));
        result["transactions_committed"] = std::stoll(PQgetvalue(res, 0, 1));
        result["transactions_rolledback"] = std::stoll(PQgetvalue(res, 0, 2));
        result["tuples_returned"] = std::stoll(PQgetvalue(res, 0, 3));
        result["tuples_fetched"] = std::stoll(PQgetvalue(res, 0, 4));
        result["tuples_inserted"] = std::stoll(PQgetvalue(res, 0, 5));
        result["tuples_updated"] = std::stoll(PQgetvalue(res, 0, 6));
        result["tuples_deleted"] = std::stoll(PQgetvalue(res, 0, 7));
    } else {
        std::cerr << "Database stats query failed: " << PQerrorMessage(conn) << std::endl;
        // Set default values
        result["size_bytes"] = 0;
        result["transactions_committed"] = 0;
        result["transactions_rolledback"] = 0;
        result["tuples_returned"] = 0;
        result["tuples_fetched"] = 0;
        result["tuples_inserted"] = 0;
        result["tuples_updated"] = 0;
        result["tuples_deleted"] = 0;
    }

    PQclear(res);
    PQfinish(conn);
#else
    // libpq not available - return default values
    result["size_bytes"] = 0;
    result["transactions_committed"] = 0;
    result["transactions_rolledback"] = 0;
    result["tuples_returned"] = 0;
    result["tuples_fetched"] = 0;
    result["tuples_inserted"] = 0;
    result["tuples_updated"] = 0;
    result["tuples_deleted"] = 0;
#endif

    return result;
}

json PgStatsCollector::collectTableStats(const std::string& dbname) {
    json tables = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, dbname,
                                      postgresUser_, postgresPassword_);
    if (!conn) {
        return tables;  // Return empty array
    }

    // Query: Get top tables by row count
    const char* query =
        "SELECT schemaname, relname, "
        "       n_live_tup, n_dead_tup, "
        "       n_mod_since_analyze, "
        "       pg_total_relation_size(schemaname||'.'||relname) as size_bytes, "
        "       last_vacuum, last_autovacuum, "
        "       last_analyze, last_autoanalyze, "
        "       vacuum_count, autovacuum_count "
        "FROM pg_stat_user_tables "
        "ORDER BY n_live_tup DESC "
        "LIMIT 100";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) == PGRES_TUPLES_OK) {
        for (int i = 0; i < PQntuples(res); i++) {
            json table = {
                {"schema", PQgetvalue(res, i, 0)},
                {"name", PQgetvalue(res, i, 1)},
                {"live_tuples", std::stoll(PQgetvalue(res, i, 2))},
                {"dead_tuples", std::stoll(PQgetvalue(res, i, 3))},
                {"modified_since_analyze", std::stoll(PQgetvalue(res, i, 4))},
                {"size_bytes", std::stoll(PQgetvalue(res, i, 5))},
                {"last_vacuum", PQgetvalue(res, i, 6) ? std::string(PQgetvalue(res, i, 6)) : "never"},
                {"last_autovacuum", PQgetvalue(res, i, 7) ? std::string(PQgetvalue(res, i, 7)) : "never"},
                {"last_analyze", PQgetvalue(res, i, 8) ? std::string(PQgetvalue(res, i, 8)) : "never"},
                {"last_autoanalyze", PQgetvalue(res, i, 9) ? std::string(PQgetvalue(res, i, 9)) : "never"},
                {"vacuum_count", std::stoll(PQgetvalue(res, i, 10))},
                {"autovacuum_count", std::stoll(PQgetvalue(res, i, 11))}
            };
            tables.push_back(table);
        }
    } else {
        std::cerr << "Table stats query failed: " << PQerrorMessage(conn) << std::endl;
    }

    PQclear(res);
    PQfinish(conn);
#endif

    return tables;
}

json PgStatsCollector::collectIndexStats(const std::string& dbname) {
    json indexes = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, dbname,
                                      postgresUser_, postgresPassword_);
    if (!conn) {
        return indexes;  // Return empty array
    }

    // Query: Get index statistics - both unused and frequently used
    const char* query =
        "SELECT schemaname, indexrelname, relname, "
        "       idx_scan, idx_tup_read, idx_tup_fetch, "
        "       pg_relation_size(indexrelid) as size_bytes, "
        "       CASE WHEN idx_scan = 0 THEN 'UNUSED' ELSE 'USED' END as status "
        "FROM pg_stat_user_indexes "
        "ORDER BY pg_relation_size(indexrelid) DESC "
        "LIMIT 100";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) == PGRES_TUPLES_OK) {
        for (int i = 0; i < PQntuples(res); i++) {
            json index = {
                {"schema", PQgetvalue(res, i, 0)},
                {"name", PQgetvalue(res, i, 1)},
                {"table", PQgetvalue(res, i, 2)},
                {"scans", std::stoll(PQgetvalue(res, i, 3))},
                {"tuples_read", std::stoll(PQgetvalue(res, i, 4))},
                {"tuples_returned", std::stoll(PQgetvalue(res, i, 5))},
                {"size_bytes", std::stoll(PQgetvalue(res, i, 6))},
                {"status", PQgetvalue(res, i, 7)}
            };
            indexes.push_back(index);
        }
    } else {
        std::cerr << "Index stats query failed: " << PQerrorMessage(conn) << std::endl;
    }

    PQclear(res);
    PQfinish(conn);
#endif

    return indexes;
}

json PgStatsCollector::collectDatabaseGlobalStats() {
    json result = json::object();

#ifdef HAVE_LIBPQ
    // Connect to any available database to get global stats
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, "postgres",
                                      postgresUser_, postgresPassword_);
    if (!conn) {
        return result;
    }

    // Query: Get summary of all databases
    const char* query =
        "SELECT datname, pg_database_size(datname) as size_bytes, "
        "       numbackends, xact_commit, xact_rollback "
        "FROM pg_stat_database "
        "ORDER BY pg_database_size(datname) DESC";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) == PGRES_TUPLES_OK) {
        json databases = json::array();
        for (int i = 0; i < PQntuples(res); i++) {
            json db = {
                {"name", PQgetvalue(res, i, 0)},
                {"size_bytes", std::stoll(PQgetvalue(res, i, 1))},
                {"backends", std::stoll(PQgetvalue(res, i, 2))},
                {"transactions_committed", std::stoll(PQgetvalue(res, i, 3))},
                {"transactions_rolledback", std::stoll(PQgetvalue(res, i, 4))}
            };
            databases.push_back(db);
        }
        result["databases"] = databases;
    } else {
        std::cerr << "Global stats query failed: " << PQerrorMessage(conn) << std::endl;
    }

    PQclear(res);
    PQfinish(conn);
#endif

    return result;
}
