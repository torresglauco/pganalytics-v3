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

    std::cerr << "DEBUG: PgQueryStatsCollector::execute() - databases count: " << databases_.size() << std::endl;

    if (databases_.empty()) {
        std::cerr << "ERROR: No databases configured for query stats collection" << std::endl;
        return result;
    }

    // Collect stats for each configured database
    for (const auto& dbname : databases_) {
        std::cerr << "DEBUG: Collecting query stats for database: " << dbname << std::endl;
        auto db_stats = collectQueryStats(dbname);

        if (db_stats.is_null()) {
            std::cerr << "DEBUG: db_stats is null for " << dbname << std::endl;
        } else if (!db_stats.contains("queries")) {
            std::cerr << "DEBUG: db_stats doesn't contain queries for " << dbname << std::endl;
        } else {
            std::cerr << "DEBUG: Successfully collected " << db_stats["queries"].size() << " queries from " << dbname << std::endl;
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
    std::cout << "DEBUG: collectQueryStats() called for " << dbname << std::endl;

    // Connect to database
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, dbname, postgresUser_, postgresPassword_);
    if (!conn) {
        std::cout << "DEBUG: Connection failed for " << dbname << std::endl;
        return json::object();
    }
    std::cout << "DEBUG: Connected to " << dbname << std::endl;

    // Initialize result object for this database
    // Note: type and timestamp are at the top level in execute(), not here
    json db_stats = {
        {"database", dbname},
        {"queries", json::array()},
        {"stats", {
            {"configured_limit", query_limit},
            {"queries_collected", 0},
            {"unique_queries_total", 0},
            {"sampling_percent", 0.0},
            {"collection_time_ms", 0}
        }}
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

    // Build query with configurable limit for pg_stat_statements
    // Default limit: 100 (backward compatible)
    // Min limit: 10, Max limit: 10000
    // This enables adaptive sampling at different scale levels:
    // - Development (< 100 QPS): limit=100 (100% collection)
    // - Small Prod (100-1K QPS): limit=500 (5% sampling)
    // - Medium Prod (1K-10K QPS): limit=1000 (1-10% sampling)
    // - Large Prod (10K+ QPS): limit=5000 (0.1-1% sampling)
    int query_limit = 100;  // Default

    // TODO: Read from global config when available (Phase 1.2)
    // For now, hardcode but mark for enhancement
    // Example when config available:
    // config->getInt("postgresql", "query_stats_limit", 100)

    std::string query_str = "SELECT queryid, query, calls, COALESCE(total_exec_time, 0), COALESCE(mean_exec_time, 0), COALESCE(min_exec_time, 0), COALESCE(max_exec_time, 0), COALESCE(stddev_exec_time, 0), COALESCE(rows, 0), COALESCE(shared_blks_hit, 0), COALESCE(shared_blks_read, 0), COALESCE(shared_blks_dirtied, 0), COALESCE(shared_blks_written, 0), COALESCE(local_blks_hit, 0), COALESCE(local_blks_read, 0), COALESCE(local_blks_dirtied, 0), COALESCE(local_blks_written, 0), COALESCE(temp_blks_read, 0), COALESCE(temp_blks_written, 0), COALESCE(blk_read_time, 0), COALESCE(blk_write_time, 0), COALESCE(wal_records, 0), COALESCE(wal_fpi, 0), COALESCE(wal_bytes, 0) FROM pg_stat_statements ORDER BY COALESCE(total_exec_time, 0) DESC LIMIT " + std::to_string(query_limit);
    const char* query = query_str.c_str();

    std::cout << "Collecting query stats from " << dbname << " (limit=" << query_limit << ")" << std::endl;

    std::cerr << "DEBUG: About to execute query on database: " << dbname << std::endl;

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "ERROR: Query execution failed on " << dbname << ". Status: " << PQresultStatus(res) << " Error: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return db_stats;
    }

    int nrows = PQntuples(res);
    int nfields = PQnfields(res);

    std::cerr << "DEBUG: Query returned " << nrows << " rows and " << nfields << " fields from " << dbname << std::endl;

    // Update metrics
    db_stats["stats"]["queries_collected"] = nrows;
    double sampling_percent = (query_limit > 0) ? ((double)nrows / query_limit) * 100.0 : 0.0;
    db_stats["stats"]["sampling_percent"] = sampling_percent;
    std::cout << "Query stats: limit=" << query_limit << ", collected=" << nrows
              << ", sampling=" << sampling_percent << "%" << std::endl;

    // Parse each row from pg_stat_statements
    // Query now returns exactly 25 fields: queryid through wal_bytes
    for (int i = 0; i < nrows; ++i) {
        try {
            json query_entry = {
                {"hash", std::stoll(PQgetvalue(res, i, 0))},           // queryid (0)
                {"text", PQgetvalue(res, i, 1)},                       // query (1)
                {"calls", std::stoll(PQgetvalue(res, i, 2))},          // calls (2)
                {"total_time", std::stod(PQgetvalue(res, i, 3))},      // total_time (3)
                {"mean_time", std::stod(PQgetvalue(res, i, 4))},       // mean_time (4)
                {"min_time", std::stod(PQgetvalue(res, i, 5))},        // min_time (5)
                {"max_time", std::stod(PQgetvalue(res, i, 6))},        // max_time (6)
                {"stddev_time", std::stod(PQgetvalue(res, i, 7))},     // stddev_time (7)
                {"rows", std::stoll(PQgetvalue(res, i, 8))},           // rows (8)
                {"shared_blks_hit", std::stoll(PQgetvalue(res, i, 9))},     // (9)
                {"shared_blks_read", std::stoll(PQgetvalue(res, i, 10))},   // (10)
                {"shared_blks_dirtied", std::stoll(PQgetvalue(res, i, 11))},// (11)
                {"shared_blks_written", std::stoll(PQgetvalue(res, i, 12))},// (12)
                {"local_blks_hit", std::stoll(PQgetvalue(res, i, 13))},     // (13)
                {"local_blks_read", std::stoll(PQgetvalue(res, i, 14))},    // (14)
                {"local_blks_dirtied", std::stoll(PQgetvalue(res, i, 15))}, // (15)
                {"local_blks_written", std::stoll(PQgetvalue(res, i, 16))}, // (16)
                {"temp_blks_read", std::stoll(PQgetvalue(res, i, 17))},     // (17)
                {"temp_blks_written", std::stoll(PQgetvalue(res, i, 18))},  // (18)
                {"blk_read_time", std::stod(PQgetvalue(res, i, 19))},       // (19)
                {"blk_write_time", std::stod(PQgetvalue(res, i, 20))},      // (20)
                {"wal_records", std::stoll(PQgetvalue(res, i, 21))},        // (21)
                {"wal_fpi", std::stoll(PQgetvalue(res, i, 22))},            // (22)
                {"wal_bytes", std::stoll(PQgetvalue(res, i, 23))}          // (23)
            };

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
