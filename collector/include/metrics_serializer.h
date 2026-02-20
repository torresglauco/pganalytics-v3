#pragma once

#include <string>
#include <vector>
#include <memory>
#include <nlohmann/json.hpp>
#include "collector.h"

using json = nlohmann::json;

/**
 * Metrics Serializer
 * Converts collector output to JSON format expected by backend API
 * Schema: https://pganalytics-backend/api/v1/metrics/push (documentation)
 */
class MetricsSerializer {
public:
    /**
     * Create a metrics payload from collector output
     * @param collectorId Unique collector identifier
     * @param hostname Name of the host running collector
     * @param version Collector version (e.g., "3.0.0")
     * @param metrics Array of JSON objects from individual collectors
     * @return Complete metrics payload ready for transmission
     */
    static json createPayload(
        const std::string& collectorId,
        const std::string& hostname,
        const std::string& version,
        const std::vector<json>& metrics
    );

    /**
     * Validate metrics payload against expected schema
     * @param payload JSON object to validate
     * @return true if valid, false otherwise
     */
    static bool validatePayload(const json& payload);

    /**
     * Validate individual metric against schema for its type
     * @param metric JSON metric object
     * @return true if valid, false otherwise
     */
    static bool validateMetric(const json& metric);

    /**
     * Get schema error details after validation failure
     * @return String describing what validation failed
     */
    static std::string getLastValidationError();

    /**
     * Get the current schema version
     * @return Schema version string
     */
    static std::string getSchemaVersion();

private:
    static std::string lastValidationError_;

    /**
     * Validate pg_stats metric schema
     */
    static bool validatePgStatsMetric(const json& metric);

    /**
     * Validate pg_log metric schema
     */
    static bool validatePgLogMetric(const json& metric);

    /**
     * Validate sysstat metric schema
     */
    static bool validateSysstatMetric(const json& metric);

    /**
     * Validate disk_usage metric schema
     */
    static bool validateDiskUsageMetric(const json& metric);

    /**
     * Validate that a field exists and is of expected type
     */
    static bool validateField(
        const json& obj,
        const std::string& fieldName,
        const std::vector<std::string>& expectedTypes
    );
};
