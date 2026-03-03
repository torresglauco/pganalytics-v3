#include "../include/extension_plugin.h"
#include <iostream>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <cstring>

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>

PgExtensionCollector::PgExtensionCollector(
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

json PgExtensionCollector::collectExtensionInfo(const std::string& dbname) {
    json extensions = json::array();

#ifdef HAVE_LIBPQ
    PGconn* conn = connectToDatabase(postgresHost_, postgresPort_, postgresUser_, postgresPassword_, dbname);
    if (!conn) return extensions;

    const char* query = R"(
        SELECT
            extname,
            extversion,
            extowner::regrole::text,
            extnamespace::regnamespace::text,
            extrelocatable,
            obj_description(oid, 'pg_extension') as description
        FROM pg_extension
        ORDER BY extname
    )";

    PGresult* result = executeQuery(conn, query);
    if (result) {
        int rows = PQntuples(result);
        for (int i = 0; i < rows; i++) {
            json ext = json::object();
            ext["name"] = PQgetvalue(result, i, 0);
            ext["version"] = PQgetvalue(result, i, 1);
            ext["owner"] = PQgetvalue(result, i, 2);
            ext["schema"] = PQgetvalue(result, i, 3);
            ext["relocatable"] = std::string(PQgetvalue(result, i, 4)) == "t";

            if (!PQgetisnull(result, i, 5)) {
                ext["description"] = PQgetvalue(result, i, 5);
            }

            extensions.push_back(ext);
        }
        PQclear(result);
    }

    PQfinish(conn);
#endif

    return extensions;
}

json PgExtensionCollector::execute() {
    auto timestamp = getCurrentTimestamp();

    json result = {
        {"type", "pg_extensions"},
        {"timestamp", timestamp},
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"databases", json::object()}
    };

    if (!enabled_) {
        result["error"] = "Extension collector is disabled";
        return result;
    }

#ifdef HAVE_LIBPQ
    for (const auto& dbname : databases_) {
        json dbExtensions = json::object();

        dbExtensions["extensions"] = collectExtensionInfo(dbname);

        result["databases"][dbname] = dbExtensions;
    }
#else
    result["error"] = "libpq not available";
#endif

    return result;
}

#else
PgExtensionCollector::PgExtensionCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword,
    const std::vector<std::string>& databases
) : enabled_(false) {
    std::cerr << "WARNING: Extension collector requires libpq" << std::endl;
}

json PgExtensionCollector::execute() {
    return json::object({
        {"type", "pg_extensions"},
        {"error", "libpq not available"}
    });
}

json PgExtensionCollector::collectExtensionInfo(const std::string& dbname) { return json::array(); }

#endif
