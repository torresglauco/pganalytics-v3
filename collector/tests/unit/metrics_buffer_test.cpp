#include <gtest/gtest.h>
#include <nlohmann/json.hpp>
#include "metrics_buffer.h"

using json = nlohmann::json;

class MetricsBufferTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Create buffer with 1MB capacity for testing
        buffer = std::make_unique<MetricsBuffer>(1024 * 1024);
    }

    void TearDown() override {
        buffer.reset();
    }

    std::unique_ptr<MetricsBuffer> buffer;

    json createTestMetric(const std::string& type) {
        json metric;
        metric["type"] = type;
        metric["timestamp"] = "2024-02-20T10:30:00Z";

        if (type == "pg_stats") {
            metric["database"] = "postgres";
            metric["tables"] = json::array();
        } else if (type == "sysstat") {
            metric["cpu"] = json::object();
        } else if (type == "pg_log") {
            metric["database"] = "postgres";
            metric["entries"] = json::array();
        } else if (type == "disk_usage") {
            metric["filesystems"] = json::array();
        }

        return metric;
    }
};

// Test 1: Create buffer instance
TEST_F(MetricsBufferTest, CreateInstance) {
    EXPECT_NE(buffer, nullptr);
}

// Test 2: Buffer starts empty
TEST_F(MetricsBufferTest, BufferStartsEmpty) {
    EXPECT_TRUE(buffer->isEmpty());
    EXPECT_FALSE(buffer->isFull());
}

// Test 3: Append metric to buffer
TEST_F(MetricsBufferTest, AppendMetric) {
    json metric = createTestMetric("pg_stats");

    bool success = buffer->append(metric);

    EXPECT_TRUE(success);
    EXPECT_FALSE(buffer->isEmpty());
}

// Test 4: Get metric count
TEST_F(MetricsBufferTest, GetMetricCount) {
    json metric1 = createTestMetric("pg_stats");
    json metric2 = createTestMetric("sysstat");

    buffer->append(metric1);
    buffer->append(metric2);

    EXPECT_EQ(buffer->getMetricCount(), 2);
}

// Test 5: Get uncompressed size
TEST_F(MetricsBufferTest, GetUncompressedSize) {
    json metric = createTestMetric("pg_stats");

    buffer->append(metric);

    EXPECT_GT(buffer->getUncompressedSize(), 0);
}

// Test 6: Get compressed data
TEST_F(MetricsBufferTest, GetCompressedData) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    std::string compressed;
    bool success = buffer->getCompressed(compressed);

    EXPECT_TRUE(success);
    EXPECT_FALSE(compressed.empty());
}

// Test 7: Compression ratio
TEST_F(MetricsBufferTest, CompressionRatio) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    std::string compressed;
    buffer->getCompressed(compressed);

    double ratio = buffer->getCompressionRatio();

    // Compression ratio should be between 0-100
    EXPECT_GE(ratio, 0.0);
    EXPECT_LE(ratio, 100.0);

    // For JSON, we expect some compression
    EXPECT_LT(ratio, 100.0);
}

// Test 8: Clear buffer
TEST_F(MetricsBufferTest, ClearBuffer) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    EXPECT_FALSE(buffer->isEmpty());

    buffer->clear();

    EXPECT_TRUE(buffer->isEmpty());
    EXPECT_EQ(buffer->getMetricCount(), 0);
}

// Test 9: Multiple metrics compression
TEST_F(MetricsBufferTest, MultipleMetricsCompression) {
    for (int i = 0; i < 10; i++) {
        json metric = createTestMetric("pg_stats");
        buffer->append(metric);
    }

    std::string compressed;
    buffer->getCompressed(compressed);

    EXPECT_FALSE(compressed.empty());
    EXPECT_EQ(buffer->getMetricCount(), 10);
}

// Test 10: Large metric
TEST_F(MetricsBufferTest, LargeMetric) {
    json metric = createTestMetric("pg_stats");

    // Add a large table array
    json tables = json::array();
    for (int i = 0; i < 100; i++) {
        json table;
        table["schema"] = "public";
        table["name"] = std::string("table_") + std::to_string(i);
        table["rows"] = 1000000;
        tables.push_back(table);
    }
    metric["tables"] = tables;

    bool success = buffer->append(metric);

    EXPECT_TRUE(success);
    EXPECT_FALSE(buffer->isEmpty());
}

// Test 11: Get statistics
TEST_F(MetricsBufferTest, GetStats) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    json stats = buffer->getStats();

    EXPECT_TRUE(stats.contains("metric_count"));
    EXPECT_TRUE(stats.contains("uncompressed_size_bytes"));
    EXPECT_TRUE(stats.contains("compressed_size_bytes"));
    EXPECT_TRUE(stats.contains("max_size_bytes"));
    EXPECT_TRUE(stats.contains("compression_ratio_percent"));
    EXPECT_TRUE(stats.contains("is_empty"));
    EXPECT_TRUE(stats.contains("is_full"));
}

// Test 12: Buffer overflow handling
TEST_F(MetricsBufferTest, BufferOverflow) {
    // Create a small buffer
    auto small_buffer = std::make_unique<MetricsBuffer>(100);

    json metric = createTestMetric("pg_stats");

    // Add a metric
    bool success1 = small_buffer->append(metric);
    EXPECT_TRUE(success1);

    // Try to add a very large metric
    json large_metric = createTestMetric("pg_stats");
    json huge_tables = json::array();
    for (int i = 0; i < 1000; i++) {
        json table;
        table["schema"] = "public";
        table["name"] = std::string("huge_table_") + std::to_string(i);
        table["rows"] = 999999999;
        huge_tables.push_back(table);
    }
    large_metric["tables"] = huge_tables;

    bool success2 = small_buffer->append(large_metric);
    EXPECT_FALSE(success2);  // Should fail due to buffer overflow
}

// Test 13: Empty buffer compression
TEST_F(MetricsBufferTest, EmptyBufferCompression) {
    std::string compressed;
    bool success = buffer->getCompressed(compressed);

    EXPECT_TRUE(success);
    // Empty buffer should compress to empty or minimal data
    EXPECT_LE(compressed.size(), 10);  // gzip header is small
}

// Test 14: Estimated compressed size
TEST_F(MetricsBufferTest, EstimatedCompressedSize) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    std::string compressed;
    buffer->getCompressed(compressed);

    size_t estimated = buffer->getEstimatedCompressedSize();
    size_t actual = compressed.size();

    // Estimated should match actual
    EXPECT_EQ(estimated, actual);
}

// Test 15: Size calculation consistency
TEST_F(MetricsBufferTest, SizeCalculationConsistency) {
    json metric1 = createTestMetric("pg_stats");
    json metric2 = createTestMetric("sysstat");

    buffer->append(metric1);
    size_t size_after_one = buffer->getUncompressedSize();

    buffer->append(metric2);
    size_t size_after_two = buffer->getUncompressedSize();

    // Size should increase after adding another metric
    EXPECT_GT(size_after_two, size_after_one);
}

// Test 16: Different metric types
TEST_F(MetricsBufferTest, DifferentMetricTypes) {
    buffer->append(createTestMetric("pg_stats"));
    buffer->append(createTestMetric("sysstat"));
    buffer->append(createTestMetric("pg_log"));
    buffer->append(createTestMetric("disk_usage"));

    EXPECT_EQ(buffer->getMetricCount(), 4);

    std::string compressed;
    buffer->getCompressed(compressed);
    EXPECT_FALSE(compressed.empty());
}

// Test 17: Clear after compression
TEST_F(MetricsBufferTest, ClearAfterCompression) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    std::string compressed;
    buffer->getCompressed(compressed);

    EXPECT_GT(buffer->getMetricCount(), 0);

    buffer->clear();

    EXPECT_EQ(buffer->getMetricCount(), 0);
    EXPECT_TRUE(buffer->isEmpty());
}

// Test 18: Repeated compress
TEST_F(MetricsBufferTest, RepeatedCompress) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    std::string compressed1;
    buffer->getCompressed(compressed1);

    // Compress again without clearing
    std::string compressed2;
    buffer->getCompressed(compressed2);

    // Should be identical
    EXPECT_EQ(compressed1, compressed2);
}

// Test 19: Buffer stats after clear
TEST_F(MetricsBufferTest, BufferStatsAfterClear) {
    json metric = createTestMetric("pg_stats");
    buffer->append(metric);

    buffer->clear();

    json stats = buffer->getStats();

    EXPECT_EQ(stats["metric_count"], 0);
    EXPECT_EQ(stats["uncompressed_size_bytes"], 0);
    EXPECT_TRUE(stats["is_empty"]);
    EXPECT_FALSE(stats["is_full"]);
}

// Test 20: Compression efficiency
TEST_F(MetricsBufferTest, CompressionEfficiency) {
    // Add multiple similar metrics to test compression efficiency
    for (int i = 0; i < 50; i++) {
        json metric = createTestMetric("pg_stats");
        metric["database"] = "postgres_db_" + std::to_string(i % 5);
        buffer->append(metric);
    }

    size_t uncompressed = buffer->getUncompressedSize();
    std::string compressed;
    buffer->getCompressed(compressed);
    size_t actual_compressed = compressed.size();

    // Compression should reduce size significantly
    EXPECT_LT(actual_compressed, uncompressed);

    // For similar JSON, we should get at least 40% compression
    double ratio = (static_cast<double>(actual_compressed) / uncompressed) * 100.0;
    EXPECT_LT(ratio, 70.0);  // At least 30% compression
}
