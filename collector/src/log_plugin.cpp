#include "../include/collector.h"
#include <iostream>
#include <fstream>
#include <ctime>
#include <iomanip>
#include <sstream>

PgLogCollector::PgLogCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword
)
    : postgresHost_(postgresHost),
      postgresPort_(postgresPort),
      postgresUser_(postgresUser),
      postgresPassword_(postgresPassword),
      enabled_(true) {
    hostname_ = hostname;
    collectorId_ = collectorId;
}

json PgLogCollector::execute() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");

    json result = {
        {"type", "pg_log"},
        {"timestamp", oss.str()},
        {"database", "postgres"},
        {"entries", json::array()}
    };

    // Collect log entries
    json entries = collectLogs();
    if (entries.is_array()) {
        result["entries"] = entries;
    }

    std::cout << "PgLogCollector::execute() - gathering PostgreSQL logs from " << postgresHost_ << std::endl;

    return result;
}

json PgLogCollector::collectLogs() {
    json entries = json::array();

    // Try common PostgreSQL log paths
    std::vector<std::string> log_paths = {
        "/var/log/postgresql/postgresql.log",
        "/var/log/postgresql-12.log",
        "/var/log/postgresql-13.log",
        "/var/log/postgresql-14.log",
        "/var/log/postgresql-15.log",
        "/var/log/postgresql-16.log",
        "/var/lib/postgresql/data/log/postgresql.log",
    };

    for (const auto& log_path : log_paths) {
        std::ifstream log_file(log_path);
        if (!log_file.is_open()) {
            continue;
        }

        std::string line;
        // Read last 100 lines (to avoid reading entire huge log)
        std::vector<std::string> lines;
        while (std::getline(log_file, line)) {
            lines.push_back(line);
            if (lines.size() > 100) {
                lines.erase(lines.begin());
            }
        }
        log_file.close();

        // Parse collected lines
        for (const auto& log_line : lines) {
            if (log_line.empty()) continue;

            json entry = json::object();
            entry["timestamp"] = "2024-02-20T10:29:55Z";  // Would parse from log line

            // Parse log level from line
            std::string level = "LOG";
            if (log_line.find("ERROR") != std::string::npos) {
                level = "ERROR";
            } else if (log_line.find("WARNING") != std::string::npos) {
                level = "WARNING";
            } else if (log_line.find("FATAL") != std::string::npos) {
                level = "FATAL";
            } else if (log_line.find("INFO") != std::string::npos) {
                level = "INFO";
            } else if (log_line.find("DEBUG") != std::string::npos) {
                level = "DEBUG";
            }

            entry["level"] = level;
            entry["message"] = log_line.substr(0, 255);  // Truncate to 255 chars

            entries.push_back(entry);
        }

        // Only process first available log file
        break;
    }

    return entries;
}
