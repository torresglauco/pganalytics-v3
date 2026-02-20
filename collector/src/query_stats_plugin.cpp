#include "../include/query_stats_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>
#include <algorithm>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>
#endif

/**
 * Constructor
 */
PgQueryStatsCollector::PgQueryStatsCollector(
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
    // Set base collector properties
    // Note: hostname_ and collectorId_ are inherited from Collector base class
    // but since PgQueryStatsCollector doesn't inherit from Collector,
    // we store them locally if needed for the JSON output
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

    connstr += " connect_timeout=5 statement_timeout=30000"; // 30 second timeout

    PGconn* conn = PQconnectdb(connstr.c_str());

    if (PQstatus(conn) != CONNECTION_OK) {
        std::cerr << "Connection to " << dbname << " failed: " << PQerrorMessage(conn) << std::endl;
        PQfinish(conn);
        return nullptr;
    }

    return conn;
}
#endif

/**
 * Execute query stats collection for all configured databases
 */
json PgQueryStatsCollector::execute() {
    json result = {
        {"type", "pg_query_stats"},
        {"timestamp", getCurrentTimestamp()},
        {"databases", json::array()}
    };

    if (databases_.empty()) {
        std::cerr << "No databases configured for query stats collection" << std::endl;
        return result;
    }

    // Collect stats for each configured database
    for (const auto& dbname : databases_) {
        auto db_stats = collectQueryStats(dbname);
        if (!db_stats.is_null() && db_stats.contains("queries")) {
            result["databases"].push_back(db_stats);
        }
    }

    return result;
}

/**
 * Collect query statistics from pg_stat_statements for a single database
 */
json PgQueryStatsCollector::collectQueryStats(const std::string& dbname) {
#ifdef HAVE_LIBPQ
    // Connect to database
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, dbname, postgresUser_, postgresPassword_);
    if (!conn) {
        return json::object();
    }

    // Initialize result object
    json db_stats = {
        {"database", dbname},
        {"timestamp", getCurrentTimestamp()},
        {"queries", json::array()}
    };

    // Check if pg_stat_statements extension is installed
    PGresult* check_res = PQexec(conn, "SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements'");
    bool has_extension = PQntuples(check_res) > 0;
    PQclear(check_res);

    if (!has_extension) {
        std::cerr << "pg_stat_statements extension not installed on database: " << dbname << std::endl;
        PQfinish(conn);
        return db_stats;  // Return empty queries array
    }

    // Query pg_stat_statements
    // Get top 100 queries by total_time (or calls)
    // Note: query text is already normalized by PostgreSQL
    const char* query =
        "SELECT "
        "  queryid, "
        "  query, "
        "  calls, "
        "  total_time, "
        "  mean_time, "
        "  min_time, "
        "  max_time, "
        "  stddev_time, "
        "  rows, "
        "  shared_blks_hit, "
        "  shared_blks_read, "
        "  shared_blks_dirtied, "
        "  shared_blks_written, "
        "  local_blks_hit, "
        "  local_blks_read, "
        "  local_blks_dirtied, "
        "  local_blks_written, "
        "  temp_blks_read, "
        "  temp_blks_written, "
        "  blk_read_time, "
        "  blk_write_time, "
        "  COALESCE(wal_records, 0) as wal_records, "
        "  COALESCE(wal_fpi, 0) as wal_fpi, "
        "  COALESCE(wal_bytes, 0) as wal_bytes, "
        "  COALESCE(query_time, 0) as query_time, "
        "  COALESCE(exec_time, 0) as exec_time "
        "FROM pg_stat_statements "
        "WHERE userid != (SELECT usesysid FROM pg_user WHERE usesuper LIMIT 1) "
        "ORDER BY total_time DESC "
        "LIMIT 100";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Query execution failed on " << dbname << ": " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return db_stats;
    }

    int nrows = PQntuples(res);
    int nfields = PQnfields(res);

    // Parse each row from pg_stat_statements
    for (int i = 0; i < nrows; ++i) {
        try {
            json query_entry = {
                {"hash", std::stoll(PQgetvalue(res, i, 0))},           // queryid
                {"text", PQgetvalue(res, i, 1)},                       // query
                {"calls", std::stoll(PQgetvalue(res, i, 2))},          // calls
                {"total_time", std::stod(PQgetvalue(res, i, 3))},      // total_time
                {"mean_time", std::stod(PQgetvalue(res, i, 4))},       // mean_time
                {"min_time", std::stod(PQgetvalue(res, i, 5))},        // min_time
                {"max_time", std::stod(PQgetvalue(res, i, 6))},        // max_time
                {"stddev_time", std::stod(PQgetvalue(res, i, 7))},     // stddev_time
                {"rows", std::stoll(PQgetvalue(res, i, 8))},           // rows
                {"shared_blks_hit", std::stoll(PQgetvalue(res, i, 9))},
                {"shared_blks_read", std::stoll(PQgetvalue(res, i, 10))},
                {"shared_blks_dirtied", std::stoll(PQgetvalue(res, i, 11))},
                {"shared_blks_written", std::stoll(PQgetvalue(res, i, 12))},
                {"local_blks_hit", std::stoll(PQgetvalue(res, i, 13))},
                {"local_blks_read", std::stoll(PQgetvalue(res, i, 14))},
                {"local_blks_dirtied", std::stoll(PQgetvalue(res, i, 15))},
                {"local_blks_written", std::stoll(PQgetvalue(res, i, 16))},
                {"temp_blks_read", std::stoll(PQgetvalue(res, i, 17))},
                {"temp_blks_written", std::stoll(PQgetvalue(res, i, 18))},
                {"blk_read_time", std::stod(PQgetvalue(res, i, 19))},
                {"blk_write_time", std::stod(PQgetvalue(res, i, 20))}
            };

            // Optional fields (PG13+) - WAL and timing statistics
            if (nfields >= 25) {
                const char* wal_records_val = PQgetvalue(res, i, 21);
                const char* wal_fpi_val = PQgetvalue(res, i, 22);
                const char* wal_bytes_val = PQgetvalue(res, i, 23);

                if (wal_records_val && std::string(wal_records_val) != "0") {
                    query_entry["wal_records"] = std::stoll(wal_records_val);
                }
                if (wal_fpi_val && std::string(wal_fpi_val) != "0") {
                    query_entry["wal_fpi"] = std::stoll(wal_fpi_val);
                }
                if (wal_bytes_val && std::string(wal_bytes_val) != "0") {
                    query_entry["wal_bytes"] = std::stoll(wal_bytes_val);
                }
            }

            if (nfields >= 27) {
                const char* query_time_val = PQgetvalue(res, i, 24);
                const char* exec_time_val = PQgetvalue(res, i, 25);

                if (query_time_val && std::string(query_time_val) != "0") {
                    query_entry["query_plan_time"] = std::stod(query_time_val);
                }
                if (exec_time_val && std::string(exec_time_val) != "0") {
                    query_entry["query_exec_time"] = std::stod(exec_time_val);
                }
            }

            db_stats["queries"].push_back(query_entry);
        } catch (const std::exception& e) {
            std::cerr << "Error parsing query stats row " << i << " from " << dbname << ": " << e.what() << std::endl;
            continue;
        }
    }

    PQclear(res);
    PQfinish(conn);

    return db_stats;

#else
    // libpq not available
    std::cerr << "libpq not available at compile time" << std::endl;
    return json::object();
#endif
}
