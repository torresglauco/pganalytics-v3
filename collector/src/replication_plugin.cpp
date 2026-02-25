#include "../include/replication_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>
#include <algorithm>
#include <cmath>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>
#endif

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
 * Constructor
 */
PgReplicationCollector::PgReplicationCollector(
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
      enabled_(true),
      postgres_version_major_(0),
      postgres_version_minor_(0),
      version_detected_(false) {
    hostname_ = hostname;
    collectorId_ = collectorId;
}

/**
 * Destructor
 */
PgReplicationCollector::~PgReplicationCollector() = default;

/**
 * Detect PostgreSQL version
 */
int PgReplicationCollector::detectPostgresVersion() {
#ifdef HAVE_LIBPQ
    if (version_detected_) {
        return postgres_version_major_;
    }

    PQconn* conn = connectToDatabase("postgres");
    if (!conn) {
        std::cerr << "Failed to detect PostgreSQL version" << std::endl;
        return 0;
    }

    PGresult* res = PQexec(conn, "SELECT current_setting('server_version_num')::int");
    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Failed to get server version" << std::endl;
        PQclear(res);
        PQfinish(conn);
        return 0;
    }

    if (PQntuples(res) > 0) {
        int version = std::stoi(PQgetvalue(res, 0, 0));
        postgres_version_major_ = version / 10000;
        postgres_version_minor_ = (version % 10000) / 100;
        version_detected_ = true;
        std::cerr << "Detected PostgreSQL version: " << postgres_version_major_ << "." << postgres_version_minor_ << std::endl;
    }

    PQclear(res);
    PQfinish(conn);
    return postgres_version_major_;

#else
    return 0;
#endif
}

/**
 * Connect to PostgreSQL database
 */
PQconn* PgReplicationCollector::connectToDatabase(const std::string& dbname) {
#ifdef HAVE_LIBPQ
    std::string connstr = "host=" + postgresHost_ +
                         " port=" + std::to_string(postgresPort_) +
                         " dbname=" + dbname +
                         " user=" + postgresUser_;

    if (!postgresPassword_.empty()) {
        connstr += " password=" + postgresPassword_;
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

#else
    return nullptr;
#endif
}

/**
 * Parse LSN string to bytes
 * LSN format: "X/XXXXXXXX" where each part is hex
 */
uint64_t PgReplicationCollector::parseLsn(const std::string& lsn) {
    try {
        size_t slash_pos = lsn.find('/');
        if (slash_pos == std::string::npos) {
            return 0;
        }

        std::string high_str = lsn.substr(0, slash_pos);
        std::string low_str = lsn.substr(slash_pos + 1);

        uint32_t high = std::stoul(high_str, nullptr, 16);
        uint32_t low = std::stoul(low_str, nullptr, 16);

        return ((uint64_t)high << 32) | low;
    } catch (const std::exception& e) {
        std::cerr << "Error parsing LSN: " << e.what() << std::endl;
        return 0;
    }
}

/**
 * Calculate bytes behind based on LSN values
 */
int64_t PgReplicationCollector::calculateBytesBehind(const std::string& write_lsn, const std::string& replay_lsn) {
    uint64_t write_bytes = parseLsn(write_lsn);
    uint64_t replay_bytes = parseLsn(replay_lsn);

    if (write_bytes >= replay_bytes) {
        return write_bytes - replay_bytes;
    }
    return 0;
}

/**
 * Collect replication slots information
 */
std::vector<PgReplicationCollector::ReplicationSlot> PgReplicationCollector::collectReplicationSlots() {
#ifdef HAVE_LIBPQ
    std::vector<ReplicationSlot> slots;

    PQconn* conn = connectToDatabase("postgres");
    if (!conn) {
        std::cerr << "Failed to connect for replication slots collection" << std::endl;
        return slots;
    }

    // Query replication slots view
    const char* query = R"(
        SELECT
            slot_name,
            slot_type,
            active,
            restart_lsn,
            confirmed_flush_lsn,
            COALESCE(ROUND(EXTRACT(EPOCH FROM (NOW() - pg_postmaster_start_time())) * 1024 * 1024), 0) as wal_retained_mb,
            plugin_active,
            COALESCE(backend_pid, 0) as backend_pid,
            NULL as database,
            COALESCE(OCTET_LENGTH(restart_lsn::text), 0) as bytes_retained
        FROM pg_replication_slots
        ORDER BY slot_name
    )";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying replication slots: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return slots;
    }

    int nrows = PQntuples(res);
    for (int i = 0; i < nrows; ++i) {
        try {
            ReplicationSlot slot;
            slot.slot_name = PQgetvalue(res, i, 0);
            slot.slot_type = PQgetvalue(res, i, 1);
            slot.active = (std::string(PQgetvalue(res, i, 2)) == "t");
            slot.restart_lsn = PQgetvalue(res, i, 3);
            slot.confirmed_flush_lsn = PQgetvalue(res, i, 4);
            slot.wal_retained_mb = std::stoll(PQgetvalue(res, i, 5));
            slot.plugin_active = (std::string(PQgetvalue(res, i, 6)) == "t");
            slot.backend_pid = std::stoll(PQgetvalue(res, i, 7));
            slot.database = PQgetvalue(res, i, 8);
            slot.bytes_retained = std::stoll(PQgetvalue(res, i, 9));

            slots.push_back(slot);
        } catch (const std::exception& e) {
            std::cerr << "Error parsing replication slot row " << i << ": " << e.what() << std::endl;
            continue;
        }
    }

    PQclear(res);
    PQfinish(conn);
    return slots;

#else
    std::cerr << "libpq not available" << std::endl;
    std::vector<ReplicationSlot> empty;
    return empty;
#endif
}

/**
 * Collect streaming replication status
 */
std::vector<PgReplicationCollector::ReplicationStatus> PgReplicationCollector::collectReplicationStatus() {
#ifdef HAVE_LIBPQ
    std::vector<ReplicationStatus> replicas;

    PQconn* conn = connectToDatabase("postgres");
    if (!conn) {
        std::cerr << "Failed to connect for replication status collection" << std::endl;
        return replicas;
    }

    // Build query based on PostgreSQL version
    // PG13+ has write_lag, flush_lag, replay_lag in milliseconds
    std::string query;

    if (postgres_version_major_ >= 13) {
        query = R"(
            SELECT
                server_pid,
                usename,
                application_name,
                state,
                sync_state,
                COALESCE(write_lsn::text, '0/0') as write_lsn,
                COALESCE(flush_lsn::text, '0/0') as flush_lsn,
                COALESCE(replay_lsn::text, '0/0') as replay_lsn,
                COALESCE(EXTRACT(EPOCH FROM write_lag) * 1000, 0)::bigint as write_lag_ms,
                COALESCE(EXTRACT(EPOCH FROM flush_lag) * 1000, 0)::bigint as flush_lag_ms,
                COALESCE(EXTRACT(EPOCH FROM replay_lag) * 1000, 0)::bigint as replay_lag_ms,
                COALESCE(backend_xmin, 0) as behind_by_mb,
                client_addr::text,
                backend_start::text
            FROM pg_stat_replication
            ORDER BY usename, application_name
        )";
    } else {
        // PG9.4 - PG12: calculate lags from LSN values
        query = R"(
            SELECT
                procpid as server_pid,
                usesysid,
                application_name,
                state,
                sync_state,
                COALESCE(location::text, '0/0') as write_lsn,
                COALESCE(location::text, '0/0') as flush_lsn,
                COALESCE(replay_location::text, '0/0') as replay_lsn,
                0 as write_lag_ms,
                0 as flush_lag_ms,
                0 as replay_lag_ms,
                0 as behind_by_mb,
                client_addr::text,
                backend_start::text
            FROM pg_stat_replication
            ORDER BY usesysid, application_name
        )";
    }

    PGresult* res = PQexec(conn, query.c_str());

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying replication status: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return replicas;
    }

    int nrows = PQntuples(res);
    for (int i = 0; i < nrows; ++i) {
        try {
            ReplicationStatus status;
            status.server_pid = std::stoll(PQgetvalue(res, i, 0));
            status.usename = PQgetvalue(res, i, 1);
            status.application_name = PQgetvalue(res, i, 2);
            status.state = PQgetvalue(res, i, 3);
            status.sync_state = PQgetvalue(res, i, 4);
            status.write_lsn = PQgetvalue(res, i, 5);
            status.flush_lsn = PQgetvalue(res, i, 6);
            status.replay_lsn = PQgetvalue(res, i, 7);
            status.write_lag_ms = std::stoll(PQgetvalue(res, i, 8));
            status.flush_lag_ms = std::stoll(PQgetvalue(res, i, 9));
            status.replay_lag_ms = std::stoll(PQgetvalue(res, i, 10));
            status.behind_by_mb = calculateBytesBehind(status.write_lsn, status.replay_lsn) / (1024 * 1024);
            status.client_addr = PQgetvalue(res, i, 12);
            status.backend_start = PQgetvalue(res, i, 13);

            replicas.push_back(status);
        } catch (const std::exception& e) {
            std::cerr << "Error parsing replication status row " << i << ": " << e.what() << std::endl;
            continue;
        }
    }

    PQclear(res);
    PQfinish(conn);
    return replicas;

#else
    std::cerr << "libpq not available" << std::endl;
    std::vector<ReplicationStatus> empty;
    return empty;
#endif
}

/**
 * Collect WAL segment status
 */
PgReplicationCollector::WalSegmentStatus PgReplicationCollector::collectWalSegmentStatus() {
#ifdef HAVE_LIBPQ
    WalSegmentStatus wal_status;
    wal_status.total_segments = 0;
    wal_status.current_wal_size_mb = 0;
    wal_status.wal_directory_size_mb = 0;
    wal_status.segments_since_checkpoint = 0;
    wal_status.growth_rate_mb_per_hour = 0.0;

    PQconn* conn = connectToDatabase("postgres");
    if (!conn) {
        std::cerr << "Failed to connect for WAL status collection" << std::endl;
        return wal_status;
    }

    // Get pg_wal_space() if available (PG13+)
    std::string query;
    if (postgres_version_major_ >= 13) {
        query = "SELECT round((pg_wal_space()).name::numeric), round((pg_wal_space()).bytes / 1024 / 1024)";
    } else {
        query = R"(
            SELECT
                COUNT(*),
                ROUND(SUM(pg_file_stat(pg_ls_waldir)) / 1024.0 / 1024.0)
            FROM pg_ls_waldir()
        )";
    }

    PGresult* res = PQexec(conn, query.c_str());

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying WAL status: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return wal_status;
    }

    if (PQntuples(res) > 0) {
        try {
            wal_status.total_segments = std::stoll(PQgetvalue(res, 0, 0));
            wal_status.current_wal_size_mb = std::stoll(PQgetvalue(res, 0, 1));
        } catch (const std::exception& e) {
            std::cerr << "Error parsing WAL status: " << e.what() << std::endl;
        }
    }

    PQclear(res);
    PQfinish(conn);
    return wal_status;

#else
    return WalSegmentStatus();
#endif
}

/**
 * Collect vacuum wraparound risk
 */
std::vector<PgReplicationCollector::VacuumWrapAroundRisk> PgReplicationCollector::collectVacuumWrapAroundRisk() {
#ifdef HAVE_LIBPQ
    std::vector<VacuumWrapAroundRisk> risks;

    PQconn* conn = connectToDatabase("postgres");
    if (!conn) {
        std::cerr << "Failed to connect for wraparound risk collection" << std::endl;
        return risks;
    }

    // Query XID status for each database
    const char* query = R"(
        SELECT
            datname,
            datfrozenxid,
            (SELECT max(age(pg_xact_commit_timestamp(xmin))) FROM pg_class WHERE pg_xact_commit_timestamp(xmin) IS NOT NULL) as max_age,
            2147483647 - datfrozenxid as xid_remaining,
            ROUND(100.0 * (2147483647 - datfrozenxid) / 2147483647, 2) as percent_remaining
        FROM pg_database
        WHERE datname NOT IN ('template0', 'template1')
        ORDER BY datfrozenxid
    )";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying wraparound risk: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return risks;
    }

    int nrows = PQntuples(res);
    for (int i = 0; i < nrows; ++i) {
        try {
            VacuumWrapAroundRisk risk;
            risk.database = PQgetvalue(res, i, 0);
            risk.relfrozenxid = std::stoll(PQgetvalue(res, i, 1));
            risk.current_xid = risk.relfrozenxid + std::stoll(PQgetvalue(res, i, 2));
            risk.xid_until_wraparound = std::stoll(PQgetvalue(res, i, 3));
            risk.percent_until_wraparound = std::stoll(PQgetvalue(res, i, 4));
            risk.at_risk = risk.percent_until_wraparound < 20;
            risk.tables_needing_vacuum = 0;  // Would need more complex query
            risk.oldest_table_age = std::stoll(PQgetvalue(res, i, 2));

            risks.push_back(risk);
        } catch (const std::exception& e) {
            std::cerr << "Error parsing wraparound risk row " << i << ": " << e.what() << std::endl;
            continue;
        }
    }

    PQclear(res);
    PQfinish(conn);
    return risks;

#else
    return std::vector<VacuumWrapAroundRisk>();
#endif
}

/**
 * Execute replication metrics collection
 */
json PgReplicationCollector::execute() {
    json result = {
        {"type", "pg_replication"},
        {"timestamp", getCurrentTimestamp()},
        {"replication_slots", json::array()},
        {"replication_status", json::array()},
        {"wal_status", json::object()},
        {"wraparound_risk", json::array()},
        {"logical_subscriptions", json::array()},
        {"collection_errors", json::array()}
    };

    // Detect PostgreSQL version
    detectPostgresVersion();

    // Collect replication slots
    try {
        auto slots = collectReplicationSlots();
        for (const auto& slot : slots) {
            json slot_obj = {
                {"slot_name", slot.slot_name},
                {"slot_type", slot.slot_type},
                {"active", slot.active},
                {"restart_lsn", slot.restart_lsn},
                {"confirmed_flush_lsn", slot.confirmed_flush_lsn},
                {"wal_retained_mb", slot.wal_retained_mb},
                {"plugin_active", slot.plugin_active},
                {"backend_pid", slot.backend_pid},
                {"bytes_retained", slot.bytes_retained}
            };
            result["replication_slots"].push_back(slot_obj);
        }
    } catch (const std::exception& e) {
        std::cerr << "Error collecting replication slots: " << e.what() << std::endl;
        result["collection_errors"].push_back("Failed to collect replication slots");
    }

    // Collect replication status
    try {
        auto replicas = collectReplicationStatus();
        for (const auto& status : replicas) {
            json status_obj = {
                {"server_pid", status.server_pid},
                {"usename", status.usename},
                {"application_name", status.application_name},
                {"state", status.state},
                {"sync_state", status.sync_state},
                {"write_lsn", status.write_lsn},
                {"flush_lsn", status.flush_lsn},
                {"replay_lsn", status.replay_lsn},
                {"write_lag_ms", status.write_lag_ms},
                {"flush_lag_ms", status.flush_lag_ms},
                {"replay_lag_ms", status.replay_lag_ms},
                {"behind_by_mb", status.behind_by_mb},
                {"client_addr", status.client_addr},
                {"backend_start", status.backend_start}
            };
            result["replication_status"].push_back(status_obj);
        }
    } catch (const std::exception& e) {
        std::cerr << "Error collecting replication status: " << e.what() << std::endl;
        result["collection_errors"].push_back("Failed to collect replication status");
    }

    // Collect WAL status
    try {
        auto wal_status = collectWalSegmentStatus();
        result["wal_status"] = {
            {"total_segments", wal_status.total_segments},
            {"current_wal_size_mb", wal_status.current_wal_size_mb},
            {"wal_directory_size_mb", wal_status.wal_directory_size_mb},
            {"segments_since_checkpoint", wal_status.segments_since_checkpoint},
            {"growth_rate_mb_per_hour", wal_status.growth_rate_mb_per_hour}
        };
    } catch (const std::exception& e) {
        std::cerr << "Error collecting WAL status: " << e.what() << std::endl;
        result["collection_errors"].push_back("Failed to collect WAL status");
    }

    // Collect wraparound risk
    try {
        auto risks = collectVacuumWrapAroundRisk();
        for (const auto& risk : risks) {
            json risk_obj = {
                {"database", risk.database},
                {"relfrozenxid", risk.relfrozenxid},
                {"current_xid", risk.current_xid},
                {"xid_until_wraparound", risk.xid_until_wraparound},
                {"percent_until_wraparound", risk.percent_until_wraparound},
                {"at_risk", risk.at_risk},
                {"tables_needing_vacuum", risk.tables_needing_vacuum},
                {"oldest_table_age", risk.oldest_table_age}
            };
            result["wraparound_risk"].push_back(risk_obj);
        }
    } catch (const std::exception& e) {
        std::cerr << "Error collecting wraparound risk: " << e.what() << std::endl;
        result["collection_errors"].push_back("Failed to collect wraparound risk");
    }

    return result;
}
