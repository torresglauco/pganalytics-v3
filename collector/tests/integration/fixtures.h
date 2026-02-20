#pragma once

#include <string>
#include <nlohmann/json.hpp>
#include <ctime>
#include <sstream>
#include <iomanip>

using json = nlohmann::json;

/**
 * Test Data Fixtures for Integration Testing
 * Provides reusable test data, configuration files, and metric payloads
 */
namespace fixtures {

// ============= Configuration Fixtures =============

inline std::string getBasicConfigToml() {
    return R"(
[collector]
id = "test-collector-001"
hostname = "test-host"
enabled = true
collection_interval = 60
push_interval = 60
config_pull_interval = 300

[backend]
url = "https://127.0.0.1:8443"

[postgres]
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"
databases = "postgres,template1"

[tls]
verify = false
cert_file = "/tmp/test_client.crt"
key_file = "/tmp/test_client.key"

[pg_stats]
enabled = true
interval = 60

[sysstat]
enabled = true
interval = 60

[pg_log]
enabled = true
interval = 60

[disk_usage]
enabled = true
interval = 60
)";
}

inline std::string getFullConfigToml() {
    return R"(
[collector]
id = "test-collector-full"
hostname = "test-host-full"
enabled = true
collection_interval = 30
push_interval = 60
config_pull_interval = 300
log_level = "debug"

[backend]
url = "https://127.0.0.1:8443"
timeout = 30
retry_count = 3
retry_backoff = "exponential"

[postgres]
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"
databases = "postgres,template1,myapp"
connection_timeout = 10

[tls]
verify = false
cert_file = "/tmp/test_client.crt"
key_file = "/tmp/test_client.key"
ca_file = "/tmp/test_ca.crt"

[pg_stats]
enabled = true
interval = 60
include_replication = true

[sysstat]
enabled = true
interval = 60
include_network = true

[pg_log]
enabled = true
interval = 60
min_level = "WARNING"

[disk_usage]
enabled = true
interval = 300
include_iops = true
)";
}

inline std::string getNoTlsConfigToml() {
    return R"(
[collector]
id = "test-collector-no-tls"
hostname = "test-host-no-tls"
enabled = true

[backend]
url = "http://127.0.0.1:8080"

[postgres]
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"

[tls]
verify = false
cert_file = ""
key_file = ""
)";
}

inline std::string getInvalidConfigToml() {
    return R"(
[collector
id = "malformed"
# Missing closing bracket and other errors
[backend
)";
}

// ============= Metric Payload Fixtures =============

inline json getPgStatsMetric() {
    return json{
        {"type", "pg_stats"},
        {"database", "postgres"},
        {"timestamp", "2024-02-20T10:30:00Z"},
        {"tables", json::array({
            json{
                {"schema", "public"},
                {"name", "users"},
                {"rows", 1000},
                {"size_bytes", 65536},
                {"last_vacuum", "2024-02-20T10:00:00Z"},
                {"last_analyze", "2024-02-20T10:15:00Z"}
            },
            json{
                {"schema", "public"},
                {"name", "posts"},
                {"rows", 50000},
                {"size_bytes", 5242880},
                {"last_vacuum", "2024-02-20T09:50:00Z"},
                {"last_analyze", "2024-02-20T09:55:00Z"}
            }
        })},
        {"indexes", json::array()},
        {"databases", json::array({
            json{
                {"name", "postgres"},
                {"size_bytes", 10485760}
            }
        })}
    };
}

inline json getSysstatMetric() {
    return json{
        {"type", "sysstat"},
        {"timestamp", "2024-02-20T10:30:00Z"},
        {"cpu", json{
            {"user", 15.5},
            {"system", 3.2},
            {"idle", 81.3},
            {"load_1m", 1.2},
            {"load_5m", 1.4},
            {"load_15m", 1.3}
        }},
        {"memory", json{
            {"total_mb", 16384},
            {"used_mb", 8192},
            {"cached_mb", 4096},
            {"free_mb", 4096}
        }},
        {"disk_io", json::array({
            json{
                {"device", "sda"},
                {"read_iops", 150},
                {"write_iops", 320},
                {"read_mb_s", 45.5},
                {"write_mb_s", 120.3}
            }
        })}
    };
}

inline json getPgLogMetric() {
    return json{
        {"type", "pg_log"},
        {"database", "postgres"},
        {"timestamp", "2024-02-20T10:30:00Z"},
        {"entries", json::array({
            json{
                {"timestamp", "2024-02-20T10:29:55Z"},
                {"level", "LOG"},
                {"message", "checkpoint complete"},
                {"duration_ms", 1234}
            },
            json{
                {"timestamp", "2024-02-20T10:29:30Z"},
                {"level", "WARNING"},
                {"message", "unused index"},
                {"detail", "index_name"}
            }
        })}
    };
}

inline json getDiskUsageMetric() {
    return json{
        {"type", "disk_usage"},
        {"timestamp", "2024-02-20T10:30:00Z"},
        {"filesystems", json::array({
            json{
                {"mount", "/"},
                {"device", "/dev/sda1"},
                {"total_gb", 100},
                {"used_gb", 45},
                {"free_gb", 55},
                {"percent_used", 45}
            },
            json{
                {"mount", "/var/lib/postgresql"},
                {"device", "/dev/sdb1"},
                {"total_gb", 500},
                {"used_gb", 250},
                {"free_gb", 250},
                {"percent_used", 50}
            }
        })}
    };
}

inline json getBasicMetricsPayload(const std::string& collector_id = "test-collector-001") {
    json payload;
    payload["collector_id"] = collector_id;
    payload["hostname"] = "test-host";
    payload["timestamp"] = "2024-02-20T10:30:00Z";
    payload["version"] = "3.0.0";
    payload["metrics"] = json::array({
        getPgStatsMetric(),
        getSysstatMetric(),
        getPgLogMetric(),
        getDiskUsageMetric()
    });
    return payload;
}

inline json getLargeMetricsPayload() {
    // Create a large payload by duplicating metrics
    json payload = getBasicMetricsPayload();
    json metrics_array = json::array();

    // Add 100 copies of each metric type
    for (int i = 0; i < 100; i++) {
        metrics_array.push_back(getPgStatsMetric());
        metrics_array.push_back(getSysstatMetric());
        metrics_array.push_back(getPgLogMetric());
        metrics_array.push_back(getDiskUsageMetric());
    }

    payload["metrics"] = metrics_array;
    return payload;
}

inline json getInvalidMetricsPayload() {
    // Missing required fields
    return json{
        {"type", "pg_stats"},
        // Missing: database, timestamp, tables, indexes, databases
    };
}

inline json getMultipleMetricsPayload() {
    json payload = getBasicMetricsPayload();
    payload["metrics"].push_back(getPgStatsMetric());
    payload["metrics"].push_back(getSysstatMetric());
    return payload;
}

// ============= Test Data Fixtures =============

inline std::string getTestCollectorId() {
    return "test-collector-001";
}

inline std::string getTestHostname() {
    return "test-host";
}

inline std::string getTestJwtToken() {
    return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0b3JfaWQiOiJ0ZXN0LWNvbGxlY3Rvci0wMDEiLCJleHAiOjk5OTk5OTk5OTksImlhdCI6MTcxNTc3NzAwMH0.test";
}

inline std::string getTestExpiredJwtToken() {
    return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0b3JfaWQiOiJ0ZXN0LWNvbGxlY3Rvci0wMDEiLCJleHAiOjEsImlhdCI6MH0.expired";
}

inline std::string getCurrentTimestamp() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");
    return oss.str();
}

}  // namespace fixtures
