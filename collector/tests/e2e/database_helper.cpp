#include "database_helper.h"
#include <iostream>
#include <cstdlib>
#include <sstream>

E2EDatabaseHelper::E2EDatabaseHelper(
    const std::string& main_db_url,
    const std::string& metrics_db_url
)
    : m_main_db_url(main_db_url),
      m_metrics_db_url(metrics_db_url),
      m_connected(false) {
    m_connected = testConnection();
}

int E2EDatabaseHelper::getMetricsCount(const std::string& table) {
    std::string query = "SELECT COUNT(*) FROM " + table + ";";
    std::string result = executeQuery(query, true);

    try {
        return std::stoi(result);
    } catch (...) {
        return 0;
    }
}

int E2EDatabaseHelper::getMetricsCountForCollector(
    const std::string& table,
    const std::string& collector_id
) {
    std::string query = "SELECT COUNT(*) FROM " + table +
                        " WHERE collector_id = '" + collector_id + "';";
    std::string result = executeQuery(query, true);

    try {
        return std::stoi(result);
    } catch (...) {
        return 0;
    }
}

bool E2EDatabaseHelper::metricsExist(const std::string& collector_id) {
    std::string query = "SELECT COUNT(*) FROM metrics_pg_stats WHERE collector_id = '" +
                        collector_id + "';";
    std::string result = executeQuery(query, true);

    try {
        int count = std::stoi(result);
        return count > 0;
    } catch (...) {
        return false;
    }
}

std::string E2EDatabaseHelper::getLatestMetricTimestamp(const std::string& table) {
    std::string query = "SELECT MAX(time) FROM " + table + ";";
    return executeQuery(query, true);
}

bool E2EDatabaseHelper::tableExists(const std::string& table) {
    std::string query = "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = '" +
                        table + "');";
    std::string result = executeQuery(query, true);
    return result.find("t") != std::string::npos;
}

bool E2EDatabaseHelper::columnExists(const std::string& table, const std::string& column) {
    std::string query = "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = '" +
                        table + "' AND column_name = '" + column + "');";
    std::string result = executeQuery(query, true);
    return result.find("t") != std::string::npos;
}

std::vector<std::string> E2EDatabaseHelper::getTableColumns(const std::string& table) {
    std::string query = "SELECT column_name FROM information_schema.columns WHERE table_name = '" +
                        table + "' ORDER BY ordinal_position;";
    std::string result = executeQuery(query, true);

    std::vector<std::string> columns;
    std::istringstream iss(result);
    std::string line;

    while (std::getline(iss, line)) {
        if (!line.empty()) {
            columns.push_back(line);
        }
    }

    return columns;
}

bool E2EDatabaseHelper::collectorExists(const std::string& collector_id) {
    std::string query = "SELECT EXISTS(SELECT 1 FROM pganalytics.collector_registry WHERE collector_id = '" +
                        collector_id + "');";
    std::string result = executeQuery(query, false);
    return result.find("t") != std::string::npos;
}

std::string E2EDatabaseHelper::getCollectorStatus(const std::string& collector_id) {
    std::string query = "SELECT status FROM pganalytics.collector_registry WHERE collector_id = '" +
                        collector_id + "';";
    return executeQuery(query, false);
}

bool E2EDatabaseHelper::configurationExists(const std::string& collector_id) {
    // Check if collector has configuration stored
    std::string query = "SELECT EXISTS(SELECT 1 FROM pganalytics.collector_config WHERE collector_id = '" +
                        collector_id + "');";
    std::string result = executeQuery(query, false);
    return result.find("t") != std::string::npos;
}

void E2EDatabaseHelper::clearAllMetrics() {
    std::cout << "[E2E DB] Clearing all metrics..." << std::endl;

    executeUpdate("TRUNCATE TABLE IF EXISTS metrics_pg_stats CASCADE;", true);
    executeUpdate("TRUNCATE TABLE IF EXISTS metrics_pg_log CASCADE;", true);
    executeUpdate("TRUNCATE TABLE IF EXISTS metrics_sysstat CASCADE;", true);
    executeUpdate("TRUNCATE TABLE IF EXISTS metrics_disk_usage CASCADE;", true);
}

void E2EDatabaseHelper::clearMetricsTable(const std::string& table) {
    std::string query = "TRUNCATE TABLE IF EXISTS " + table + " CASCADE;";
    executeUpdate(query, true);
}

void E2EDatabaseHelper::clearCollectorMetrics(const std::string& collector_id) {
    std::string query = "DELETE FROM metrics_pg_stats WHERE collector_id = '" + collector_id + "';";
    executeUpdate(query, true);
}

void E2EDatabaseHelper::truncateAllData() {
    std::cout << "[E2E DB] Truncating all data..." << std::endl;

    // Clear metrics
    clearAllMetrics();

    // Clear collector registry
    executeUpdate("TRUNCATE TABLE IF EXISTS pganalytics.collector_registry CASCADE;", false);
    executeUpdate("TRUNCATE TABLE IF EXISTS pganalytics.api_tokens CASCADE;", false);
    executeUpdate("TRUNCATE TABLE IF EXISTS pganalytics.collector_config CASCADE;", false);
}

std::string E2EDatabaseHelper::executeQuery(const std::string& sql, bool use_metrics_db) {
    std::string db_url = use_metrics_db ? m_metrics_db_url : m_main_db_url;
    std::string output;
    executePsqlCommand("-tc \"" + sql + "\"", db_url, output);

    // Trim output
    output.erase(0, output.find_first_not_of(" \t\r\n"));
    output.erase(output.find_last_not_of(" \t\r\n") + 1);

    return output;
}

bool E2EDatabaseHelper::executeUpdate(const std::string& sql, bool use_metrics_db) {
    std::string db_url = use_metrics_db ? m_metrics_db_url : m_main_db_url;
    std::string output;
    return executePsqlCommand("-c \"" + sql + "\"", db_url, output);
}

bool E2EDatabaseHelper::isConnected() {
    return m_connected;
}

bool E2EDatabaseHelper::testConnection() {
    std::cout << "[E2E DB] Testing database connections..." << std::endl;

    // Test main database
    std::string output;
    if (!executePsqlCommand("-c 'SELECT 1;'", m_main_db_url, output)) {
        std::cerr << "[E2E DB] Failed to connect to main database" << std::endl;
        return false;
    }

    // Test metrics database
    if (!executePsqlCommand("-c 'SELECT 1;'", m_metrics_db_url, output)) {
        std::cerr << "[E2E DB] Failed to connect to metrics database" << std::endl;
        return false;
    }

    std::cout << "[E2E DB] Database connections OK" << std::endl;
    return true;
}

bool E2EDatabaseHelper::executePsqlCommand(
    const std::string& cmd,
    const std::string& db_url,
    std::string& output
) {
    // Parse connection string: postgresql://user:password@host:port/dbname
    std::string host = "localhost";
    std::string port = "5432";
    std::string user = "postgres";
    std::string password = "pganalytics";
    std::string dbname = "metrics";

    // Try to extract components from db_url (simple parsing)
    if (db_url.find("5433") != std::string::npos) {
        port = "5433";
    }
    if (db_url.find("pganalytics") != std::string::npos) {
        dbname = "pganalytics";
    }

    // Build psql command
    std::string full_cmd = "PGPASSWORD=" + password + " psql -h " + host + " -p " + port +
                           " -U " + user + " -d " + dbname + " " + cmd + " 2>&1";

    FILE* pipe = popen(full_cmd.c_str(), "r");
    if (!pipe) {
        return false;
    }

    char buffer[256];
    output.clear();

    while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
        output += buffer;
    }

    int status = pclose(pipe);
    return status == 0;
}

