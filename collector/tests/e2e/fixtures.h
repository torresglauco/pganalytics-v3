#pragma once

#include <string>
#include <vector>

namespace e2e_fixtures {

/**
 * E2E Test Fixtures
 *
 * Provides test data and configuration for E2E tests.
 */

// ==================== Collector Data ====================

inline std::string getCollectorId() {
    return "e2e_col_001";
}

inline std::string getCollectorName() {
    return "E2E Test Collector";
}

inline std::string getCollectorHostname() {
    return "e2e-test-host";
}

// ==================== Configuration ====================

inline std::string getBasicCollectorConfig() {
    return R"(
[collector]
id = "e2e_col_001"
hostname = "e2e-test-host"
log_level = "debug"

[backend]
url = "https://backend:8080"
tls_verify = false

[postgresql]
host = "postgres"
port = 5432
user = "postgres"
password = "pganalytics"
databases = ["postgres", "pganalytics"]

[collection]
interval = 60
enabled_metrics = ["pg_stats", "pg_log", "sysstat", "disk_usage"]
)";
}

// ==================== Metrics Data ====================

inline std::string getBasicMetricsPayload() {
    return R"({
  "collector_id": "e2e_col_001",
  "hostname": "e2e-test-host",
  "timestamp": "2026-02-19T12:00:00Z",
  "version": "3.0.0",
  "metrics": [
    {
      "type": "pg_stats",
      "database": "postgres",
      "timestamp": "2026-02-19T12:00:00Z",
      "tables": [
        {
          "schema": "public",
          "name": "test_table",
          "rows": 1000,
          "size_bytes": 65536,
          "last_vacuum": "2026-02-19T11:50:00Z",
          "last_analyze": "2026-02-19T11:50:00Z"
        }
      ],
      "indexes": [],
      "databases": [
        {
          "name": "postgres",
          "size_bytes": 5242880,
          "connections": 3
        }
      ]
    },
    {
      "type": "sysstat",
      "timestamp": "2026-02-19T12:00:00Z",
      "cpu": {
        "user": 15.5,
        "system": 5.2,
        "idle": 79.3,
        "load_1m": 0.8,
        "load_5m": 1.0,
        "load_15m": 0.9
      },
      "memory": {
        "total_mb": 8192,
        "used_mb": 4096,
        "cached_mb": 2048,
        "free_mb": 2048
      },
      "disk_io": [
        {
          "device": "sda",
          "read_iops": 50,
          "write_iops": 30,
          "read_mb_s": 10,
          "write_mb_s": 5
        }
      ]
    },
    {
      "type": "disk_usage",
      "timestamp": "2026-02-19T12:00:00Z",
      "filesystems": [
        {
          "mount": "/",
          "device": "/dev/sda1",
          "total_gb": 100,
          "used_gb": 45,
          "free_gb": 55,
          "percent_used": 45
        }
      ]
    }
  ]
})";
}

inline std::string getLargeMetricsPayload(int metric_count = 10) {
    std::string payload = R"({
  "collector_id": "e2e_col_001",
  "hostname": "e2e-test-host",
  "timestamp": "2026-02-19T12:00:00Z",
  "version": "3.0.0",
  "metrics": [)";

    for (int i = 0; i < metric_count; i++) {
        if (i > 0) payload += ",";
        payload += R"({
      "type": "pg_stats",
      "database": "postgres",
      "timestamp": "2026-02-19T12:00:00Z",
      "tables": [
        {
          "schema": "public",
          "name": "table_)" + std::to_string(i) + R"(",
          "rows": )" + std::to_string(1000 * (i + 1)) + R"(,
          "size_bytes": )" + std::to_string(65536 * (i + 1)) + R"(,
          "last_vacuum": "2026-02-19T11:50:00Z",
          "last_analyze": "2026-02-19T11:50:00Z"
        }
      ],
      "indexes": [],
      "databases": []
    })";
    }

    payload += "]}\n";
    return payload;
}

// ==================== Registration ====================

inline std::string getRegistrationRequest() {
    return R"({
  "name": "E2E Test Collector",
  "hostname": "e2e-test-host"
})";
}

// ==================== Expected Responses ====================

inline std::string getExpectedRegistrationResponse() {
    return R"({
  "status": "success",
  "collector_id": "e2e_col_001",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "certificate": "-----BEGIN CERTIFICATE-----\n...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n..."
})";
}

// ==================== Test Parameters ====================

inline int getDefaultMetricsCount() {
    return 50;
}

inline int getDefaultTimeoutSeconds() {
    return 30;
}

inline int getDefaultCollectionIntervalSeconds() {
    return 60;
}

// ==================== Database Test Data ====================

inline std::string getTestCollectorId() {
    return "e2e_col_001";
}

inline std::string getTestDatabaseName() {
    return "pganalytics";
}

inline std::string getTestTableName() {
    return "metrics_pg_stats";
}

// ==================== Error Scenarios ====================

inline std::string getInvalidMetricsPayload() {
    return R"({
  "collector_id": "e2e_col_001",
  "metrics": [
    {
      "type": "invalid_type",
      "data": "missing required fields"
    }
  ]
})";
}

inline std::string getEmptyMetricsPayload() {
    return R"({
  "collector_id": "e2e_col_001",
  "hostname": "e2e-test-host",
  "timestamp": "2026-02-19T12:00:00Z",
  "version": "3.0.0",
  "metrics": []
})";
}

}  // namespace e2e_fixtures

