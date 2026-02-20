#include <gtest/gtest.h>
#include <nlohmann/json.hpp>
#include "metrics_serializer.h"

using json = nlohmann::json;

class MetricsSerializerTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Reset any state before each test
    }

    void TearDown() override {
        // Cleanup after each test
    }
};

// Test 1: Create basic payload
TEST_F(MetricsSerializerTest, CreateBasicPayload) {
    std::vector<json> metrics;

    json payload = MetricsSerializer::createPayload(
        "col-test-001",
        "test-host",
        "3.0.0",
        metrics
    );

    ASSERT_TRUE(payload.contains("collector_id"));
    ASSERT_TRUE(payload.contains("hostname"));
    ASSERT_TRUE(payload.contains("timestamp"));
    ASSERT_TRUE(payload.contains("version"));
    ASSERT_TRUE(payload.contains("metrics"));

    EXPECT_EQ(payload["collector_id"], "col-test-001");
    EXPECT_EQ(payload["hostname"], "test-host");
    EXPECT_EQ(payload["version"], "3.0.0");
    EXPECT_TRUE(payload["metrics"].is_array());
}

// Test 2: Payload with metrics
TEST_F(MetricsSerializerTest, PayloadWithMetrics) {
    std::vector<json> metrics;

    json metric1;
    metric1["type"] = "pg_stats";
    metric1["timestamp"] = "2024-02-20T10:30:00Z";
    metric1["database"] = "postgres";
    metrics.push_back(metric1);

    json payload = MetricsSerializer::createPayload(
        "col-001",
        "host-01",
        "3.0.0",
        metrics
    );

    EXPECT_EQ(payload["metrics"].size(), 1);
    EXPECT_EQ(payload["metrics"][0]["type"], "pg_stats");
}

// Test 3: Validate valid payload
TEST_F(MetricsSerializerTest, ValidateValidPayload) {
    json payload;
    payload["collector_id"] = "col-001";
    payload["hostname"] = "host-01";
    payload["timestamp"] = "2024-02-20T10:30:00Z";
    payload["version"] = "3.0.0";
    payload["metrics"] = json::array();

    EXPECT_TRUE(MetricsSerializer::validatePayload(payload));
}

// Test 4: Validate missing collector_id
TEST_F(MetricsSerializerTest, ValidateMissingCollectorId) {
    json payload;
    payload["hostname"] = "host-01";
    payload["timestamp"] = "2024-02-20T10:30:00Z";
    payload["version"] = "3.0.0";
    payload["metrics"] = json::array();

    EXPECT_FALSE(MetricsSerializer::validatePayload(payload));
    EXPECT_FALSE(MetricsSerializer::getLastValidationError().empty());
}

// Test 5: Validate missing metrics array
TEST_F(MetricsSerializerTest, ValidateMissingMetricsArray) {
    json payload;
    payload["collector_id"] = "col-001";
    payload["hostname"] = "host-01";
    payload["timestamp"] = "2024-02-20T10:30:00Z";
    payload["version"] = "3.0.0";

    EXPECT_FALSE(MetricsSerializer::validatePayload(payload));
}

// Test 6: Validate pg_stats metric
TEST_F(MetricsSerializerTest, ValidatePgStatsMetric) {
    json metric;
    metric["type"] = "pg_stats";
    metric["timestamp"] = "2024-02-20T10:30:00Z";
    metric["database"] = "postgres";
    metric["tables"] = json::array();

    EXPECT_TRUE(MetricsSerializer::validateMetric(metric));
}

// Test 7: Validate pg_stats without database field
TEST_F(MetricsSerializerTest, ValidatePgStatsWithoutDatabase) {
    json metric;
    metric["type"] = "pg_stats";
    metric["timestamp"] = "2024-02-20T10:30:00Z";

    EXPECT_FALSE(MetricsSerializer::validateMetric(metric));
}

// Test 8: Validate pg_log metric
TEST_F(MetricsSerializerTest, ValidatePgLogMetric) {
    json metric;
    metric["type"] = "pg_log";
    metric["timestamp"] = "2024-02-20T10:30:00Z";
    metric["database"] = "postgres";
    metric["entries"] = json::array();

    EXPECT_TRUE(MetricsSerializer::validateMetric(metric));
}

// Test 9: Validate sysstat metric
TEST_F(MetricsSerializerTest, ValidateSysstatMetric) {
    json metric;
    metric["type"] = "sysstat";
    metric["timestamp"] = "2024-02-20T10:30:00Z";

    json cpu;
    cpu["user"] = 10.5;
    cpu["system"] = 3.2;
    cpu["idle"] = 86.3;
    metric["cpu"] = cpu;

    EXPECT_TRUE(MetricsSerializer::validateMetric(metric));
}

// Test 10: Validate disk_usage metric
TEST_F(MetricsSerializerTest, ValidateDiskUsageMetric) {
    json metric;
    metric["type"] = "disk_usage";
    metric["timestamp"] = "2024-02-20T10:30:00Z";
    metric["filesystems"] = json::array();

    EXPECT_TRUE(MetricsSerializer::validateMetric(metric));
}

// Test 11: Validate unknown metric type
TEST_F(MetricsSerializerTest, ValidateUnknownMetricType) {
    json metric;
    metric["type"] = "unknown_metric";
    metric["timestamp"] = "2024-02-20T10:30:00Z";

    EXPECT_FALSE(MetricsSerializer::validateMetric(metric));
    std::string error = MetricsSerializer::getLastValidationError();
    EXPECT_TRUE(error.find("Unknown metric type") != std::string::npos);
}

// Test 12: Validate pg_stats with table entries
TEST_F(MetricsSerializerTest, ValidatePgStatsWithTables) {
    json metric;
    metric["type"] = "pg_stats";
    metric["timestamp"] = "2024-02-20T10:30:00Z";
    metric["database"] = "postgres";

    json table;
    table["schema"] = "public";
    table["name"] = "users";
    table["rows"] = 1000000;

    json tables = json::array();
    tables.push_back(table);
    metric["tables"] = tables;

    EXPECT_TRUE(MetricsSerializer::validateMetric(metric));
}

// Test 13: Schema version
TEST_F(MetricsSerializerTest, GetSchemaVersion) {
    std::string version = MetricsSerializer::getSchemaVersion();
    EXPECT_EQ(version, "1.0.0");
}

// Test 14: Validate with invalid metric object type
TEST_F(MetricsSerializerTest, ValidateInvalidMetricObject) {
    json notAnObject = "this is a string";
    EXPECT_FALSE(MetricsSerializer::validateMetric(notAnObject));
}

// Test 15: Validate pg_log with entries
TEST_F(MetricsSerializerTest, ValidatePgLogWithEntries) {
    json metric;
    metric["type"] = "pg_log";
    metric["timestamp"] = "2024-02-20T10:30:00Z";
    metric["database"] = "postgres";

    json entry;
    entry["timestamp"] = "2024-02-20T10:29:55Z";
    entry["level"] = "LOG";
    entry["message"] = "checkpoint complete";

    json entries = json::array();
    entries.push_back(entry);
    metric["entries"] = entries;

    EXPECT_TRUE(MetricsSerializer::validateMetric(metric));
}

// Test 16: Validate pg_log without message field
TEST_F(MetricsSerializerTest, ValidatePgLogEntryWithoutMessage) {
    json metric;
    metric["type"] = "pg_log";
    metric["timestamp"] = "2024-02-20T10:30:00Z";
    metric["database"] = "postgres";

    json entry;
    entry["timestamp"] = "2024-02-20T10:29:55Z";
    entry["level"] = "LOG";
    // Missing "message" field

    json entries = json::array();
    entries.push_back(entry);
    metric["entries"] = entries;

    EXPECT_FALSE(MetricsSerializer::validateMetric(metric));
}

// Test 17: Multiple metrics in payload
TEST_F(MetricsSerializerTest, ValidateMultipleMetrics) {
    json payload;
    payload["collector_id"] = "col-001";
    payload["hostname"] = "host-01";
    payload["timestamp"] = "2024-02-20T10:30:00Z";
    payload["version"] = "3.0.0";

    json metric1;
    metric1["type"] = "pg_stats";
    metric1["timestamp"] = "2024-02-20T10:30:00Z";
    metric1["database"] = "postgres";

    json metric2;
    metric2["type"] = "sysstat";
    metric2["timestamp"] = "2024-02-20T10:30:00Z";

    json metrics = json::array();
    metrics.push_back(metric1);
    metrics.push_back(metric2);
    payload["metrics"] = metrics;

    EXPECT_TRUE(MetricsSerializer::validatePayload(payload));
}

// Test 18: Payload field type validation
TEST_F(MetricsSerializerTest, PayloadFieldTypes) {
    json payload;
    payload["collector_id"] = 123;  // Should be string
    payload["hostname"] = "host-01";
    payload["timestamp"] = "2024-02-20T10:30:00Z";
    payload["version"] = "3.0.0";
    payload["metrics"] = json::array();

    EXPECT_FALSE(MetricsSerializer::validatePayload(payload));
}

// Test 19: Empty metrics array is valid
TEST_F(MetricsSerializerTest, EmptyMetricsArray) {
    json payload;
    payload["collector_id"] = "col-001";
    payload["hostname"] = "host-01";
    payload["timestamp"] = "2024-02-20T10:30:00Z";
    payload["version"] = "3.0.0";
    payload["metrics"] = json::array();

    EXPECT_TRUE(MetricsSerializer::validatePayload(payload));
}

// Test 20: Sysstat with all optional fields
TEST_F(MetricsSerializerTest, SysstatWithAllFields) {
    json metric;
    metric["type"] = "sysstat";
    metric["timestamp"] = "2024-02-20T10:30:00Z";

    json cpu;
    cpu["user"] = 10.5;
    cpu["system"] = 3.2;
    cpu["idle"] = 86.3;
    metric["cpu"] = cpu;

    json memory;
    memory["total_mb"] = 16384;
    memory["used_mb"] = 8192;
    metric["memory"] = memory;

    json io = json::array();
    metric["disk_io"] = io;

    EXPECT_TRUE(MetricsSerializer::validateMetric(metric));
}
