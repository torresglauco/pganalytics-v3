#include "../include/logical_replication_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>
#include <algorithm>

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
LogicalReplicationCollector::LogicalReplicationCollector(
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
LogicalReplicationCollector::~LogicalReplicationCollector() = default;

/**
 * Detect PostgreSQL version
 */
int LogicalReplicationCollector::detectPostgresVersion() {
#ifdef HAVE_LIBPQ
    if (version_detected_) {
        return postgres_version_major_;
    }

    PGconn* conn = connectToDatabase("postgres");
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
PGconn* LogicalReplicationCollector::connectToDatabase(const std::string& dbname) {
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
 * Get list of databases to monitor
 */
std::vector<std::string> LogicalReplicationCollector::getDatabases() {
#ifdef HAVE_LIBPQ
    if (!databases_.empty()) {
        return databases_;
    }

    // Get all non-template databases
    std::vector<std::string> dbs;
    PGconn* conn = connectToDatabase("postgres");
    if (!conn) {
        return dbs;
    }

    const char* query = "SELECT datname FROM pg_database WHERE NOT datistemplate AND datallowconn ORDER BY datname";
    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying databases: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return dbs;
    }

    int nrows = PQntuples(res);
    for (int i = 0; i < nrows; ++i) {
        dbs.push_back(PQgetvalue(res, i, 0));
    }

    PQclear(res);
    PQfinish(conn);
    return dbs;

#else
    return databases_;
#endif
}

/**
 * Collect logical subscriptions from pg_stat_subscription (PG 10+)
 */
std::vector<LogicalReplicationCollector::LogicalSubscription> LogicalReplicationCollector::collectLogicalSubscriptions(const std::string& dbname) {
    std::vector<LogicalSubscription> subs;

#ifdef HAVE_LIBPQ
    // Only available in PostgreSQL 10+
    if (postgres_version_major_ < 10) {
        return subs;
    }

    PGconn* conn = connectToDatabase(dbname);
    if (!conn) {
        std::cerr << "Failed to connect for logical subscriptions collection" << std::endl;
        return subs;
    }

    // Query pg_stat_subscription
    const char* query = R"(
        SELECT
            subname,
            COALESCE(status, 'unknown'),
            COALESCE(received_lsn::text, '0/0'),
            COALESCE(latest_end_lsn::text, '0/0'),
            COALESCE(last_msg_receipt_time::text, ''),
            COALESCE(last_msg_send_time::text, ''),
            COALESCE(pid, 0)
        FROM pg_stat_subscription
        ORDER BY subname
    )";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying logical subscriptions: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return subs;
    }

    int nrows = PQntuples(res);
    for (int i = 0; i < nrows; ++i) {
        try {
            LogicalSubscription sub;
            sub.sub_name = PQgetvalue(res, i, 0);
            sub.sub_state = PQgetvalue(res, i, 1);
            sub.received_lsn = PQgetvalue(res, i, 2);
            sub.latest_end_lsn = PQgetvalue(res, i, 3);
            sub.last_msg_receipt_time = PQgetvalue(res, i, 4);
            sub.last_msg_send_time = PQgetvalue(res, i, 5);
            sub.worker_pid = std::stoll(PQgetvalue(res, i, 6));
            sub.worker_count = 1;  // Default, would need more complex query for actual count
            sub.database = dbname;

            subs.push_back(sub);
        } catch (const std::exception& e) {
            std::cerr << "Error parsing logical subscription row " << i << ": " << e.what() << std::endl;
            continue;
        }
    }

    PQclear(res);
    PQfinish(conn);

#else
    std::cerr << "libpq not available" << std::endl;
#endif

    return subs;
}

/**
 * Collect publications from pg_publication
 */
std::vector<LogicalReplicationCollector::Publication> LogicalReplicationCollector::collectPublications(const std::string& dbname) {
    std::vector<Publication> pubs;

#ifdef HAVE_LIBPQ
    // Only available in PostgreSQL 10+
    if (postgres_version_major_ < 10) {
        return pubs;
    }

    PGconn* conn = connectToDatabase(dbname);
    if (!conn) {
        std::cerr << "Failed to connect for publications collection" << std::endl;
        return pubs;
    }

    // Query pg_publication with owner name from pg_roles
    const char* query = R"(
        SELECT
            p.pubname,
            r.rolname as owner,
            p.puballtables,
            p.pubinsert,
            p.pubupdate,
            p.pubdelete,
            COALESCE(p.pubtruncate, false)
        FROM pg_publication p
        JOIN pg_roles r ON p.pubowner = r.oid
        ORDER BY p.pubname
    )";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying publications: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return pubs;
    }

    int nrows = PQntuples(res);
    for (int i = 0; i < nrows; ++i) {
        try {
            Publication pub;
            pub.pub_name = PQgetvalue(res, i, 0);
            pub.pub_owner = PQgetvalue(res, i, 1);
            pub.pub_all_tables = (std::string(PQgetvalue(res, i, 2)) == "t");
            pub.pub_insert = (std::string(PQgetvalue(res, i, 3)) == "t");
            pub.pub_update = (std::string(PQgetvalue(res, i, 4)) == "t");
            pub.pub_delete = (std::string(PQgetvalue(res, i, 5)) == "t");
            pub.pub_truncate = (std::string(PQgetvalue(res, i, 6)) == "t");
            pub.database = dbname;

            pubs.push_back(pub);
        } catch (const std::exception& e) {
            std::cerr << "Error parsing publication row " << i << ": " << e.what() << std::endl;
            continue;
        }
    }

    PQclear(res);
    PQfinish(conn);

#else
    std::cerr << "libpq not available" << std::endl;
#endif

    return pubs;
}

/**
 * Collect WAL receiver status from pg_stat_wal_receiver (PG 9.6+)
 */
LogicalReplicationCollector::WalReceiver LogicalReplicationCollector::collectWalReceiver() {
    WalReceiver receiver;
    receiver.status = "";
    receiver.sender_host = "";
    receiver.sender_port = 0;
    receiver.received_lsn = "";
    receiver.latest_end_lsn = "";
    receiver.slot_name = "";
    receiver.conn_info = "";

#ifdef HAVE_LIBPQ
    // Only available in PostgreSQL 9.6+
    if (postgres_version_major_ < 9 || (postgres_version_major_ == 9 && postgres_version_minor_ < 6)) {
        return receiver;
    }

    PGconn* conn = connectToDatabase("postgres");
    if (!conn) {
        std::cerr << "Failed to connect for WAL receiver collection" << std::endl;
        return receiver;
    }

    // Query pg_stat_wal_receiver (single row - one receiver per standby)
    const char* query = R"(
        SELECT
            status,
            COALESCE(sender_host, ''),
            COALESCE(sender_port, 0),
            COALESCE(received_lsn::text, '0/0'),
            COALESCE(latest_end_lsn::text, '0/0'),
            COALESCE(slot_name, ''),
            COALESCE(conninfo, '')
        FROM pg_stat_wal_receiver
        LIMIT 1
    )";

    PGresult* res = PQexec(conn, query);

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        std::cerr << "Error querying WAL receiver: " << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return receiver;
    }

    if (PQntuples(res) > 0) {
        try {
            receiver.status = PQgetvalue(res, 0, 0);
            receiver.sender_host = PQgetvalue(res, 0, 1);
            receiver.sender_port = std::stoi(PQgetvalue(res, 0, 2));
            receiver.received_lsn = PQgetvalue(res, 0, 3);
            receiver.latest_end_lsn = PQgetvalue(res, 0, 4);
            receiver.slot_name = PQgetvalue(res, 0, 5);
            receiver.conn_info = PQgetvalue(res, 0, 6);
        } catch (const std::exception& e) {
            std::cerr << "Error parsing WAL receiver row: " << e.what() << std::endl;
        }
    }

    PQclear(res);
    PQfinish(conn);

#else
    std::cerr << "libpq not available" << std::endl;
#endif

    return receiver;
}

/**
 * Execute logical replication metrics collection
 */
json LogicalReplicationCollector::execute() {
    json result = {
        {"type", "pg_logical_replication"},
        {"timestamp", getCurrentTimestamp()},
        {"logical_subscriptions", json::array()},
        {"publications", json::array()},
        {"wal_receiver", json::object()},
        {"collection_errors", json::array()}
    };

    // Detect PostgreSQL version
    detectPostgresVersion();

    // Get list of databases to monitor
    auto databases = getDatabases();

    // Collect logical subscriptions and publications for each database
    for (const auto& dbname : databases) {
        // Collect logical subscriptions
        try {
            auto subs = collectLogicalSubscriptions(dbname);
            for (const auto& sub : subs) {
                json sub_obj = {
                    {"sub_name", sub.sub_name},
                    {"sub_state", sub.sub_state},
                    {"received_lsn", sub.received_lsn},
                    {"latest_end_lsn", sub.latest_end_lsn},
                    {"last_msg_receipt_time", sub.last_msg_receipt_time},
                    {"last_msg_send_time", sub.last_msg_send_time},
                    {"worker_pid", sub.worker_pid},
                    {"worker_count", sub.worker_count},
                    {"database", sub.database}
                };
                result["logical_subscriptions"].push_back(sub_obj);
            }
        } catch (const std::exception& e) {
            std::cerr << "Error collecting logical subscriptions for " << dbname << ": " << e.what() << std::endl;
            result["collection_errors"].push_back("Failed to collect logical subscriptions for " + dbname);
        }

        // Collect publications
        try {
            auto pubs = collectPublications(dbname);
            for (const auto& pub : pubs) {
                json pub_obj = {
                    {"pub_name", pub.pub_name},
                    {"pub_owner", pub.pub_owner},
                    {"pub_all_tables", pub.pub_all_tables},
                    {"pub_insert", pub.pub_insert},
                    {"pub_update", pub.pub_update},
                    {"pub_delete", pub.pub_delete},
                    {"pub_truncate", pub.pub_truncate},
                    {"database", pub.database}
                };
                result["publications"].push_back(pub_obj);
            }
        } catch (const std::exception& e) {
            std::cerr << "Error collecting publications for " << dbname << ": " << e.what() << std::endl;
            result["collection_errors"].push_back("Failed to collect publications for " + dbname);
        }
    }

    // Collect WAL receiver status (once, at cluster level)
    try {
        auto wal_receiver = collectWalReceiver();
        if (!wal_receiver.status.empty()) {
            result["wal_receiver"] = {
                {"status", wal_receiver.status},
                {"sender_host", wal_receiver.sender_host},
                {"sender_port", wal_receiver.sender_port},
                {"received_lsn", wal_receiver.received_lsn},
                {"latest_end_lsn", wal_receiver.latest_end_lsn},
                {"slot_name", wal_receiver.slot_name},
                {"conn_info", wal_receiver.conn_info}
            };
        }
    } catch (const std::exception& e) {
        std::cerr << "Error collecting WAL receiver: " << e.what() << std::endl;
        result["collection_errors"].push_back("Failed to collect WAL receiver status");
    }

    return result;
}