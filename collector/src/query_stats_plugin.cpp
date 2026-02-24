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

    connstr += " connect_timeout=5";

    PGconn* conn = PQconnectdb(connstr.c_str());

    if (PQstatus(conn) != CONNECTION_OK) {
        std::cerr << "Connection to " << dbname << " failed: " << PQerrorMessage(conn) << std::endl;
        PQfinish(conn);
        return nullptr;
    }

    // Set statement timeout to 30 seconds
    PGresult* res = PQexec(conn, "SET statement_timeout = '30s'");
    if (PQresultStatus(res) != PGRES_COMMAND_OK) {
        std::cerr << "Failed to set statement timeout: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return nullptr;
    }
    PQclear(res);

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
    // Get top 100 queries by execution time (or calls)
    // Note: query text is already normalized by PostgreSQL
    // PostgreSQL 16+ renamed total_time/mean_time/etc to *_exec_time/*_plan_time
    const char* query =
        "SELECT "
        "  queryid, "
        "  query, "
        "  calls, "
        "  COALESCE(mean_exec_time, 0) as total_time, "
        "  COALESCE(mean_exec_time, 0), "
        "  COALESCE(min_exec_time, 0), "
        "  COALESCE(max_exec_time, 0), "
        "  COALESCE((max_exec_time - min_exec_time), 0), "
        "  COALESCE(rows, 0), "
        "  COALESCE(shared_blks_hit, 0), "
        "  COALESCE(shared_blks_read, 0), "
        "  COALESCE(shared_blks_dirtied, 0), "
        "  COALESCE(shared_blks_written, 0), "
        "  COALESCE(local_blks_hit, 0), "
        "  COALESCE(local_blks_read, 0), "
        "  COALESCE(local_blks_dirtied, 0), "
        "  COALESCE(local_blks_written, 0), "
        "  0 as temp_blks_read, "
        "  0 as temp_blks_written, "
        "  COALESCE(blk_read_time, 0), "
        "  COALESCE(blk_write_time, 0), "
        "  0 as wal_records, "
        "  0 as wal_fpi, "
        "  0 as wal_bytes, "
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

/**
 * Execute EXPLAIN ANALYZE (JSON) on a query to capture execution plan
 * Phase 4.4.2: EXPLAIN PLAN Integration
 */
json PgQueryStatsCollector::executeExplainPlan(
    const std::string& dbname,
    int64_t queryHash,
    const std::string& queryText) {

#ifdef HAVE_LIBPQ
    try {
        // Build connection string
        std::ostringstream conn_str;
        conn_str << "host=" << postgresHost_
                 << " port=" << postgresPort_
                 << " user=" << postgresUser_
                 << " password=" << postgresPassword_
                 << " dbname=" << dbname
                 << " connect_timeout=5";

        PGconn* conn = PQconnectdb(conn_str.str().c_str());

        if (PQstatus(conn) != CONNECTION_OK) {
            std::cerr << "Failed to connect to " << dbname << " for EXPLAIN: " << PQerrorMessage(conn) << std::endl;
            PQfinish(conn);
            return json::object();
        }

        // Execute EXPLAIN ANALYZE (FORMAT JSON)
        // Build query: EXPLAIN (ANALYZE, FORMAT JSON) <original query>
        std::ostringstream explain_query;
        explain_query << "EXPLAIN (ANALYZE, FORMAT JSON, BUFFERS) " << queryText;

        // Set statement timeout to 30 seconds
        PQexec(conn, "SET statement_timeout = '30s'");

        PGresult* res = PQexec(conn, explain_query.str().c_str());

        if (PQresultStatus(res) != PGRES_TUPLES_OK) {
            std::cerr << "EXPLAIN failed for query " << queryHash << ": " << PQerrorMessage(conn) << std::endl;
            PQclear(res);
            PQfinish(conn);
            return json::object();
        }

        // Extract JSON plan from result
        json plan_json = json::object();
        plan_json["query_hash"] = queryHash;
        plan_json["database"] = dbname;
        plan_json["collected_at"] = getCurrentTimestamp();
        plan_json["query_text"] = queryText;

        if (PQntuples(res) > 0) {
            try {
                // PostgreSQL returns EXPLAIN JSON as a single text field
                std::string explain_result = PQgetvalue(res, 0, 0);

                // Parse the JSON result
                auto plan_data = json::parse(explain_result);

                if (plan_data.is_array() && plan_data.size() > 0) {
                    auto plan_obj = plan_data[0];

                    // Extract key metrics from plan
                    plan_json["plan"] = plan_obj["Plan"];
                    plan_json["planning_time_ms"] = plan_obj.value("Planning Time", 0.0);
                    plan_json["execution_time_ms"] = plan_obj.value("Execution Time", 0.0);

                    // Extract row counts if present
                    if (plan_obj.contains("Plan")) {
                        auto& plan = plan_obj["Plan"];
                        if (plan.contains("Actual Rows")) {
                            plan_json["rows_actual"] = plan["Actual Rows"];
                        }
                        if (plan.contains("Rows")) {
                            plan_json["rows_expected"] = plan["Rows"];
                        }

                        // Check for scan types
                        plan_json["has_seq_scan"] = plan_json.dump().find("Seq Scan") != std::string::npos;
                        plan_json["has_index_scan"] = plan_json.dump().find("Index") != std::string::npos;
                        plan_json["has_bitmap_scan"] = plan_json.dump().find("Bitmap") != std::string::npos;
                        plan_json["has_nested_loop"] = plan_json.dump().find("Nested Loop") != std::string::npos;

                        // Extract buffer statistics
                        if (plan.contains("Shared Hit Blocks")) {
                            plan_json["shared_blocks_hit"] = plan["Shared Hit Blocks"];
                        }
                        if (plan.contains("Shared Read Blocks")) {
                            plan_json["shared_blocks_read"] = plan["Shared Read Blocks"];
                        }
                    }
                }
            } catch (const std::exception& e) {
                std::cerr << "Error parsing EXPLAIN JSON for query " << queryHash << ": " << e.what() << std::endl;
                plan_json["parse_error"] = e.what();
            }
        }

        PQclear(res);
        PQfinish(conn);

        return plan_json;

    } catch (const std::exception& e) {
        std::cerr << "Exception executing EXPLAIN for query " << queryHash << ": " << e.what() << std::endl;
        return json::object();
    }

#else
    // libpq not available
    std::cerr << "libpq not available for EXPLAIN execution" << std::endl;
    return json::object();
#endif
}
