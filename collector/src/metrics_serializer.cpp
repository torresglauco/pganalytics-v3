#include "../include/metrics_serializer.h"
#include <ctime>
#include <iomanip>
#include <sstream>
#include <iostream>

std::string MetricsSerializer::lastValidationError_;

json MetricsSerializer::createPayload(
    const std::string& collectorId,
    const std::string& hostname,
    const std::string& version,
    const std::vector<json>& metrics
) {
    // Get current ISO 8601 timestamp
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");

    json payload = {
        {"collector_id", collectorId},
        {"hostname", hostname},
        {"timestamp", oss.str()},
        {"version", version},
        {"metrics", json::array()}
    };

    // Add all metrics to the payload
    for (const auto& metric : metrics) {
        payload["metrics"].push_back(metric);
    }

    return payload;
}

bool MetricsSerializer::validatePayload(const json& payload) {
    try {
        // Check top-level fields
        if (!validateField(payload, "collector_id", {"string"})) {
            lastValidationError_ = "Missing or invalid collector_id (must be string)";
            return false;
        }

        if (!validateField(payload, "hostname", {"string"})) {
            lastValidationError_ = "Missing or invalid hostname (must be string)";
            return false;
        }

        if (!validateField(payload, "timestamp", {"string"})) {
            lastValidationError_ = "Missing or invalid timestamp (must be ISO 8601 string)";
            return false;
        }

        if (!validateField(payload, "version", {"string"})) {
            lastValidationError_ = "Missing or invalid version (must be string)";
            return false;
        }

        // Check metrics array exists
        if (!payload.contains("metrics") || !payload["metrics"].is_array()) {
            lastValidationError_ = "Missing metrics array or not an array";
            return false;
        }

        // Validate each metric
        for (const auto& metric : payload["metrics"]) {
            if (!validateMetric(metric)) {
                // lastValidationError_ is already set by validateMetric
                return false;
            }
        }

        return true;
    } catch (const std::exception& e) {
        lastValidationError_ = std::string("Exception during validation: ") + e.what();
        return false;
    }
}

bool MetricsSerializer::validateMetric(const json& metric) {
    try {
        if (!metric.is_object()) {
            lastValidationError_ = "Metric is not a JSON object";
            return false;
        }

        if (!validateField(metric, "type", {"string"})) {
            lastValidationError_ = "Metric missing or invalid type field";
            return false;
        }

        if (!validateField(metric, "timestamp", {"string"})) {
            lastValidationError_ = "Metric missing or invalid timestamp field";
            return false;
        }

        std::string type = metric["type"].get<std::string>();

        if (type == "pg_stats") {
            return validatePgStatsMetric(metric);
        } else if (type == "pg_query_stats") {
            return validatePgQueryStatsMetric(metric);
        } else if (type == "pg_log") {
            return validatePgLogMetric(metric);
        } else if (type == "sysstat") {
            return validateSysstatMetric(metric);
        } else if (type == "disk_usage") {
            return validateDiskUsageMetric(metric);
        } else {
            lastValidationError_ = "Unknown metric type: " + type;
            return false;
        }
    } catch (const std::exception& e) {
        lastValidationError_ = std::string("Exception validating metric: ") + e.what();
        return false;
    }
}

bool MetricsSerializer::validatePgStatsMetric(const json& metric) {
    if (!validateField(metric, "database", {"string"})) {
        lastValidationError_ = "pg_stats metric missing or invalid database field";
        return false;
    }

    // tables, indexes, databases are optional but if present must be arrays
    if (metric.contains("tables") && !metric["tables"].is_array()) {
        lastValidationError_ = "pg_stats metric: tables must be an array";
        return false;
    }

    if (metric.contains("indexes") && !metric["indexes"].is_array()) {
        lastValidationError_ = "pg_stats metric: indexes must be an array";
        return false;
    }

    if (metric.contains("databases") && !metric["databases"].is_array()) {
        lastValidationError_ = "pg_stats metric: databases must be an array";
        return false;
    }

    // Validate table objects if present
    if (metric.contains("tables") && metric["tables"].is_array()) {
        for (const auto& table : metric["tables"]) {
            if (!validateField(table, "schema", {"string"})) {
                lastValidationError_ = "pg_stats table missing or invalid schema field";
                return false;
            }
            if (!validateField(table, "name", {"string"})) {
                lastValidationError_ = "pg_stats table missing or invalid name field";
                return false;
            }
        }
    }

    return true;
}

bool MetricsSerializer::validatePgLogMetric(const json& metric) {
    if (!validateField(metric, "database", {"string"})) {
        lastValidationError_ = "pg_log metric missing or invalid database field";
        return false;
    }

    // entries is optional but if present must be array
    if (metric.contains("entries") && !metric["entries"].is_array()) {
        lastValidationError_ = "pg_log metric: entries must be an array";
        return false;
    }

    // Validate entry objects if present
    if (metric.contains("entries") && metric["entries"].is_array()) {
        for (const auto& entry : metric["entries"]) {
            if (!validateField(entry, "timestamp", {"string"})) {
                lastValidationError_ = "pg_log entry missing or invalid timestamp field";
                return false;
            }
            if (!validateField(entry, "level", {"string"})) {
                lastValidationError_ = "pg_log entry missing or invalid level field";
                return false;
            }
            if (!validateField(entry, "message", {"string"})) {
                lastValidationError_ = "pg_log entry missing or invalid message field";
                return false;
            }
        }
    }

    return true;
}

bool MetricsSerializer::validatePgQueryStatsMetric(const json& metric) {
    if (!validateField(metric, "database", {"string"})) {
        lastValidationError_ = "pg_query_stats metric missing or invalid database field";
        return false;
    }

    // queries is optional but if present must be array
    if (metric.contains("queries") && !metric["queries"].is_array()) {
        lastValidationError_ = "pg_query_stats metric: queries must be an array";
        return false;
    }

    // Validate query objects if present
    if (metric.contains("queries") && metric["queries"].is_array()) {
        for (const auto& query : metric["queries"]) {
            if (!validateField(query, "hash", {"number"})) {
                lastValidationError_ = "pg_query_stats query missing or invalid hash field";
                return false;
            }
            if (!validateField(query, "text", {"string"})) {
                lastValidationError_ = "pg_query_stats query missing or invalid text field";
                return false;
            }
        }
    }

    return true;
}

bool MetricsSerializer::validateSysstatMetric(const json& metric) {
    // CPU stats optional but if present must be object
    if (metric.contains("cpu") && !metric["cpu"].is_object()) {
        lastValidationError_ = "sysstat metric: cpu must be an object";
        return false;
    }

    // Memory stats optional but if present must be object
    if (metric.contains("memory") && !metric["memory"].is_object()) {
        lastValidationError_ = "sysstat metric: memory must be an object";
        return false;
    }

    // Disk IO stats optional but if present must be array
    if (metric.contains("disk_io") && !metric["disk_io"].is_array()) {
        lastValidationError_ = "sysstat metric: disk_io must be an array";
        return false;
    }

    return true;
}

bool MetricsSerializer::validateDiskUsageMetric(const json& metric) {
    // filesystems must be array
    if (metric.contains("filesystems") && !metric["filesystems"].is_array()) {
        lastValidationError_ = "disk_usage metric: filesystems must be an array";
        return false;
    }

    // Validate filesystem objects
    if (metric.contains("filesystems") && metric["filesystems"].is_array()) {
        for (const auto& fs : metric["filesystems"]) {
            if (!validateField(fs, "mount", {"string"})) {
                lastValidationError_ = "disk_usage filesystem missing or invalid mount field";
                return false;
            }
            if (!validateField(fs, "device", {"string"})) {
                lastValidationError_ = "disk_usage filesystem missing or invalid device field";
                return false;
            }
        }
    }

    return true;
}

bool MetricsSerializer::validateField(
    const json& obj,
    const std::string& fieldName,
    const std::vector<std::string>& expectedTypes
) {
    if (!obj.contains(fieldName)) {
        return false;
    }

    const auto& field = obj[fieldName];

    for (const auto& expectedType : expectedTypes) {
        if (expectedType == "string" && field.is_string()) {
            return true;
        }
        if (expectedType == "number" && field.is_number()) {
            return true;
        }
        if (expectedType == "integer" && field.is_number_integer()) {
            return true;
        }
        if (expectedType == "array" && field.is_array()) {
            return true;
        }
        if (expectedType == "object" && field.is_object()) {
            return true;
        }
    }

    return false;
}

std::string MetricsSerializer::getLastValidationError() {
    return lastValidationError_;
}

std::string MetricsSerializer::getSchemaVersion() {
    return "1.0.0";
}
