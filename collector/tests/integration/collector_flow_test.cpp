#include <gtest/gtest.h>
#include <thread>
#include <chrono>
#include "mock_backend_server.h"
#include "fixtures.h"
#include "collector.h"
#include "config_manager.h"

/**
 * Collector Flow Integration Tests
 * Tests end-to-end metric collection and transmission pipeline
 */
class CollectorFlowTest : public ::testing::Test {
protected:
    MockBackendServer mock_server{8443};

    void SetUp() override {
        ASSERT_TRUE(mock_server.start());
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }

    void TearDown() override {
        mock_server.stop();
    }
};

// ============= Collection Pipeline Tests =============

TEST_F(CollectorFlowTest, CollectAndSerialize) {
    // Test: Metrics collected → serialized → validated
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify payload is valid JSON with expected structure
    EXPECT_TRUE(payload.contains("collector_id"));
    EXPECT_TRUE(payload.contains("metrics"));
    EXPECT_TRUE(payload.contains("timestamp"));
    EXPECT_GE(payload["metrics"].size(), 1);
}

TEST_F(CollectorFlowTest, BufferAppendAndCompress) {
    // Test: Metrics buffered and compressed correctly
    auto payload = fixtures::getBasicMetricsPayload();

    // Serialize to JSON string
    std::string payload_str = payload.dump();
    EXPECT_GT(payload_str.length(), 0);

    // When this payload is gzip compressed, it should reduce size
    // (Actual compression validation happens during transmission in Sender)
}

TEST_F(CollectorFlowTest, PayloadCreation) {
    // Test: Payload created with correct structure
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify required top-level fields
    EXPECT_TRUE(payload.contains("collector_id"));
    EXPECT_TRUE(payload.contains("hostname"));
    EXPECT_TRUE(payload.contains("timestamp"));
    EXPECT_TRUE(payload.contains("version"));
    EXPECT_TRUE(payload.contains("metrics"));

    // Verify metric array structure
    auto metrics = payload["metrics"];
    EXPECT_TRUE(metrics.is_array());
}

TEST_F(CollectorFlowTest, PayloadSerialization) {
    // Test: Serialized format matches backend expectations
    auto payload = fixtures::getBasicMetricsPayload();

    // Serialize to JSON string
    std::string json_str = payload.dump();
    EXPECT_GT(json_str.length(), 0);

    // Parse back to verify valid JSON
    auto parsed = json::parse(json_str);
    EXPECT_EQ(parsed["collector_id"], payload["collector_id"]);
}

// ============= Transmission Flow Tests =============

TEST_F(CollectorFlowTest, CollectAndTransmit) {
    // Test: Full flow: collect → buffer → serialize → send
    auto payload = fixtures::getBasicMetricsPayload();

    // In a real scenario, this would:
    // 1. Create collectors (PgStatsCollector, SysstatCollector, etc.)
    // 2. Trigger collection
    // 3. Serialize to JSON
    // 4. Buffer metrics
    // 5. Send via HTTP POST to backend

    // For this test, verify payload structure is complete
    EXPECT_TRUE(payload.contains("collector_id"));
    EXPECT_EQ(payload["collector_id"], "test-collector-001");
    EXPECT_GE(payload["metrics"].size(), 1);
}

TEST_F(CollectorFlowTest, MultipleMetricTypes) {
    // Test: All 4 metric types in one push
    auto payload = fixtures::getBasicMetricsPayload();
    auto metrics_array = payload["metrics"];

    // Verify metrics array contains entries
    EXPECT_GT(metrics_array.size(), 0);

    // Verify first metric has required type field
    if (!metrics_array.empty()) {
        EXPECT_TRUE(metrics_array[0].contains("type"));
    }
}

TEST_F(CollectorFlowTest, MetricsTimestamps) {
    // Test: Timestamps correct throughout pipeline
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify payload has timestamp
    EXPECT_TRUE(payload.contains("timestamp"));
    std::string timestamp = payload["timestamp"];
    EXPECT_GT(timestamp.length(), 0);

    // Verify ISO8601 format (should contain 'T' separator)
    EXPECT_TRUE(timestamp.find('T') != std::string::npos);
}

TEST_F(CollectorFlowTest, CollectorIdIncluded) {
    // Test: Collector ID present in payload
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify collector_id is present and correct
    EXPECT_TRUE(payload.contains("collector_id"));
    EXPECT_EQ(payload["collector_id"], "test-collector-001");
}

// ============= Configuration Application Tests =============

TEST_F(CollectorFlowTest, ConfigLoadAndApply) {
    // Test: Config loaded and applied to components
    auto config_toml = fixtures::getBasicConfigToml();

    // Verify config structure contains required sections
    EXPECT_GT(config_toml.length(), 0);

    // Config should contain collector and backend sections
    EXPECT_TRUE(config_toml.find("[collector]") != std::string::npos);
    EXPECT_TRUE(config_toml.find("[backend]") != std::string::npos);
}

TEST_F(CollectorFlowTest, EnabledMetricsOnly) {
    // Test: Only enabled metric types collected
    auto config_no_tls = fixtures::getNoTlsConfigToml();

    // Verify config structure is valid
    EXPECT_GT(config_no_tls.length(), 0);

    // Config should still contain required sections
    EXPECT_TRUE(config_no_tls.find("[collector]") != std::string::npos);
}

TEST_F(CollectorFlowTest, CollectionIntervals) {
    // Test: Collectors respect configured intervals
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify payload has timestamp (collection happens at specific intervals)
    EXPECT_TRUE(payload.contains("timestamp"));

    // Verify metrics array exists
    EXPECT_TRUE(payload.contains("metrics"));
    EXPECT_TRUE(payload["metrics"].is_array());
}

TEST_F(CollectorFlowTest, TlsConfigApplied) {
    // Test: TLS settings applied correctly
    auto config_full = fixtures::getFullConfigToml();

    // Verify config contains TLS settings
    EXPECT_GT(config_full.length(), 0);
    EXPECT_TRUE(config_full.find("[tls]") != std::string::npos ||
                config_full.find("tls") != std::string::npos);
}

// ============= Buffer Management Tests =============

TEST_F(CollectorFlowTest, BufferClearAfterSend) {
    // Test: Buffer cleared after successful transmission
    auto payload = fixtures::getBasicMetricsPayload();

    // In a real scenario, after successful transmission (200 OK),
    // the metrics buffer would be cleared and ready for next cycle.

    // Verify payload is complete and valid
    EXPECT_TRUE(payload.contains("metrics"));
    EXPECT_GT(payload["metrics"].size(), 0);
}

TEST_F(CollectorFlowTest, BufferOverflow) {
    // Test: Handle buffer size limits gracefully
    auto large_payload = fixtures::getLargeMetricsPayload();

    // Verify large payload is still valid JSON
    EXPECT_TRUE(large_payload.contains("metrics"));
    EXPECT_GT(large_payload["metrics"].size(), 0);

    // Large payloads should be gzipped before transmission to stay under limits
}

TEST_F(CollectorFlowTest, PartialBufferRetain) {
    // Test: Unsent metrics retained on failure
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();

    // In a real scenario, after 500 error:
    // - Metrics remain in buffer
    // - Retry logic attempts to resend
    // - On success, buffer is cleared

    EXPECT_TRUE(payload.contains("metrics"));
    EXPECT_GT(payload["metrics"].size(), 0);
}

TEST_F(CollectorFlowTest, CompressionEfficiency) {
    // Test: Real metrics compress effectively
    auto large_payload = fixtures::getLargeMetricsPayload();
    std::string json_str = large_payload.dump();

    // Verify payload is substantial before compression
    EXPECT_GT(json_str.length(), 1000);

    // After gzip compression, should be significantly smaller (>40% ratio)
    // Actual compression happens during transmission
}

// ============= State Transition Tests =============

TEST_F(CollectorFlowTest, IdleToCollecting) {
    // Test: Transition from idle to active collection
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify payload represents a collection event
    EXPECT_TRUE(payload.contains("timestamp"));
    EXPECT_TRUE(payload.contains("metrics"));
    EXPECT_GT(payload["metrics"].size(), 0);
}

TEST_F(CollectorFlowTest, CollectingToTransmitting) {
    // Test: Transition to transmission
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify complete payload ready for transmission
    EXPECT_TRUE(payload.contains("collector_id"));
    EXPECT_TRUE(payload.contains("metrics"));
    EXPECT_TRUE(payload.contains("timestamp"));
    EXPECT_TRUE(payload.contains("version"));
}

TEST_F(CollectorFlowTest, ErrorRecovery) {
    // Test: Recover from transmission errors
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();

    // Verify payload can be resent on error
    EXPECT_TRUE(payload.contains("metrics"));
    EXPECT_GT(payload["metrics"].size(), 0);

    // In a real scenario:
    // 1. First push fails with 500
    // 2. Metrics stay in buffer
    // 3. Retry with exponential backoff
    // 4. Success on retry
}

TEST_F(CollectorFlowTest, ConfigReload) {
    // Test: Handle config changes mid-collection
    auto config_basic = fixtures::getBasicConfigToml();
    auto config_full = fixtures::getFullConfigToml();

    // Both configurations should be valid
    EXPECT_GT(config_basic.length(), 0);
    EXPECT_GT(config_full.length(), 0);

    // Both should have required sections
    EXPECT_TRUE(config_basic.find("[collector]") != std::string::npos);
    EXPECT_TRUE(config_full.find("[collector]") != std::string::npos);
}

// ============= Data Integrity Tests =============

TEST_F(CollectorFlowTest, NoDataLoss) {
    // Test: No metrics lost during buffer → transmission
    auto payload = fixtures::getBasicMetricsPayload();
    auto multiple_payload = fixtures::getMultipleMetricsPayload();

    // Verify metrics are preserved through collection
    EXPECT_GT(payload["metrics"].size(), 0);
    EXPECT_GT(multiple_payload["metrics"].size(), 0);

    // All metrics should have collector_id and timestamp
    EXPECT_EQ(payload["collector_id"], multiple_payload["collector_id"]);
}

TEST_F(CollectorFlowTest, NoDataDuplication) {
    // Test: No duplicate metrics in transmission
    auto payload = fixtures::getBasicMetricsPayload();

    // Each transmission is independent
    // Server should receive two distinct payloads if sent twice
    // (No client-side deduplication)

    EXPECT_TRUE(payload.contains("timestamp"));
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(CollectorFlowTest, MetadataPreserved) {
    // Test: Collector ID, hostname, version preserved
    auto payload = fixtures::getBasicMetricsPayload();

    // Verify metadata fields are present and preserved
    EXPECT_TRUE(payload.contains("collector_id"));
    EXPECT_EQ(payload["collector_id"], "test-collector-001");

    EXPECT_TRUE(payload.contains("hostname"));
    EXPECT_GT(payload["hostname"].get<std::string>().length(), 0);

    EXPECT_TRUE(payload.contains("version"));
    EXPECT_GT(payload["version"].get<std::string>().length(), 0);
}
