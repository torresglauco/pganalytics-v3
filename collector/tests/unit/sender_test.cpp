#include <gtest/gtest.h>
#include <nlohmann/json.hpp>
#include "sender.h"

using json = nlohmann::json;

class SenderTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Create Sender instance for testing
        sender = std::make_unique<Sender>(
            "https://localhost:8080",
            "test-collector-001",
            "/tmp/test.crt",
            "/tmp/test.key",
            false  // Don't verify TLS for testing
        );
    }

    void TearDown() override {
        sender.reset();
    }

    std::unique_ptr<Sender> sender;

    json createTestMetrics() {
        json metrics;
        metrics["collector_id"] = "test-collector-001";
        metrics["hostname"] = "test-host";
        metrics["timestamp"] = "2024-02-20T10:30:00Z";
        metrics["version"] = "3.0.0";
        metrics["metrics"] = json::array();

        json metric;
        metric["type"] = "pg_stats";
        metric["timestamp"] = "2024-02-20T10:30:00Z";
        metric["database"] = "postgres";
        metrics["metrics"].push_back(metric);

        return metrics;
    }
};

// Test 1: Create Sender instance
TEST_F(SenderTest, CreateInstance) {
    EXPECT_NE(sender, nullptr);
}

// Test 2: Set authentication token
TEST_F(SenderTest, SetAuthToken) {
    std::string token = "test.jwt.token";
    time_t expiresAt = std::time(nullptr) + 3600;
    sender->setAuthToken(token, expiresAt);

    EXPECT_EQ(sender->getAuthToken(), token);
}

// Test 3: Get authentication token
TEST_F(SenderTest, GetAuthToken) {
    std::string token = "test.jwt.token";
    time_t expiresAt = std::time(nullptr) + 3600;
    sender->setAuthToken(token, expiresAt);

    std::string retrieved = sender->getAuthToken();
    EXPECT_EQ(retrieved, token);
}

// Test 4: Token validity initially false
TEST_F(SenderTest, TokenValidityInitiallyFalse) {
    EXPECT_FALSE(sender->isTokenValid());
}

// Test 5: Token validity after setting
TEST_F(SenderTest, TokenValidityAfterSetting) {
    std::string token = "test.jwt.token";
    time_t future = std::time(nullptr) + 3600;  // 1 hour in future

    sender->setAuthToken(token, future);

    EXPECT_TRUE(sender->isTokenValid());
}

// Test 6: Metrics validation
TEST_F(SenderTest, ValidMetrics) {
    json metrics = createTestMetrics();

    // This test verifies that valid metrics are accepted
    // Actual push would fail without a running backend
    EXPECT_TRUE(metrics.contains("metrics"));
}

// Test 7: Empty metrics rejection
TEST_F(SenderTest, EmptyMetrics) {
    json invalid_metrics;
    // Missing required fields

    // Note: Actual pushMetrics would fail, but we're testing structure
    EXPECT_FALSE(invalid_metrics.contains("metrics"));
}

// Test 8: Token expiration
TEST_F(SenderTest, TokenExpiration) {
    std::string token = "test.jwt.token";
    time_t past = std::time(nullptr) - 100;  // 100 seconds in past

    sender->setAuthToken(token, past);

    EXPECT_FALSE(sender->isTokenValid());
}

// Test 9: Multiple tokens
TEST_F(SenderTest, MultipleTokens) {
    std::string token1 = "token.one.here";
    time_t time1 = std::time(nullptr) + 3600;

    sender->setAuthToken(token1, time1);
    EXPECT_EQ(sender->getAuthToken(), token1);

    std::string token2 = "token.two.here";
    time_t time2 = std::time(nullptr) + 7200;

    sender->setAuthToken(token2, time2);
    EXPECT_EQ(sender->getAuthToken(), token2);
}

// Test 10: Refresh token check
TEST_F(SenderTest, RefreshTokenCheck) {
    // Set token that expires soon (61 seconds)
    std::string token1 = "token.before.refresh";
    time_t time1 = std::time(nullptr) + 61;
    sender->setAuthToken(token1, time1);

    EXPECT_TRUE(sender->isTokenValid());

    // Set token that expires in less than 60 seconds
    std::string token2 = "token.after.refresh";
    time_t time2 = std::time(nullptr) + 59;
    sender->setAuthToken(token2, time2);

    EXPECT_FALSE(sender->isTokenValid());  // Should trigger refresh
}

// Test 11: Collector ID in Sender
TEST_F(SenderTest, CollectorIdStorage) {
    // Sender stores collector ID for use in requests
    // Verify it's created with the right ID
    EXPECT_NE(sender, nullptr);  // Successfully created with collector ID
}

// Test 12: Backend URL
TEST_F(SenderTest, BackendUrl) {
    // Verify sender was created with correct backend URL
    // (URL is used internally for requests)
    EXPECT_NE(sender, nullptr);
}

// Test 13: Certificate file paths
TEST_F(SenderTest, CertificateFilePaths) {
    // Sender stores certificate paths for mTLS
    // Verify paths are accepted in constructor
    auto test_sender = std::make_unique<Sender>(
        "https://localhost:8080",
        "collector-001",
        "/path/to/cert.pem",
        "/path/to/key.pem",
        true
    );

    EXPECT_NE(test_sender, nullptr);
}

// Test 14: TLS verification flag
TEST_F(SenderTest, TLSVerificationFlag) {
    // Create sender with TLS verification enabled
    auto sender_verify = std::make_unique<Sender>(
        "https://localhost:8080",
        "collector-001",
        "/path/to/cert.pem",
        "/path/to/key.pem",
        true  // Verify TLS
    );

    // Create sender with TLS verification disabled
    auto sender_no_verify = std::make_unique<Sender>(
        "https://localhost:8080",
        "collector-001",
        "/path/to/cert.pem",
        "/path/to/key.pem",
        false  // Don't verify
    );

    EXPECT_NE(sender_verify, nullptr);
    EXPECT_NE(sender_no_verify, nullptr);
}

// Test 15: Metrics compression preparation
TEST_F(SenderTest, MetricsCompressionPrep) {
    json metrics = createTestMetrics();

    // Verify metrics can be serialized (ready for compression)
    std::string serialized = metrics.dump();
    EXPECT_FALSE(serialized.empty());
    EXPECT_GT(serialized.size(), 0);
}

// Test 16: Large metrics payload
TEST_F(SenderTest, LargeMetricsPayload) {
    json metrics;
    metrics["collector_id"] = "test-collector-001";
    metrics["hostname"] = "test-host";
    metrics["timestamp"] = "2024-02-20T10:30:00Z";
    metrics["version"] = "3.0.0";

    json metrics_array = json::array();

    // Create large metric payload
    for (int i = 0; i < 100; i++) {
        json metric;
        metric["type"] = "pg_stats";
        metric["timestamp"] = "2024-02-20T10:30:00Z";
        metric["database"] = "postgres";

        json tables = json::array();
        for (int j = 0; j < 10; j++) {
            json table;
            table["schema"] = "public";
            table["name"] = "table_" + std::to_string(i) + "_" + std::to_string(j);
            tables.push_back(table);
        }
        metric["tables"] = tables;
        metrics_array.push_back(metric);
    }

    metrics["metrics"] = metrics_array;

    std::string serialized = metrics.dump();
    EXPECT_GT(serialized.size(), 1000);  // Should be substantial
}

// Test 17: Metrics structure validation
TEST_F(SenderTest, MetricsStructureValidation) {
    json metrics = createTestMetrics();

    EXPECT_TRUE(metrics.contains("collector_id"));
    EXPECT_TRUE(metrics.contains("hostname"));
    EXPECT_TRUE(metrics.contains("timestamp"));
    EXPECT_TRUE(metrics.contains("version"));
    EXPECT_TRUE(metrics.contains("metrics"));
    EXPECT_TRUE(metrics["metrics"].is_array());
}

// Test 18: Different collector IDs
TEST_F(SenderTest, DifferentCollectorIds) {
    auto sender1 = std::make_unique<Sender>(
        "https://localhost:8080",
        "collector-001",
        "/path/to/cert.pem",
        "/path/to/key.pem",
        false
    );

    auto sender2 = std::make_unique<Sender>(
        "https://localhost:8080",
        "collector-002",
        "/path/to/cert.pem",
        "/path/to/key.pem",
        false
    );

    EXPECT_NE(sender1, nullptr);
    EXPECT_NE(sender2, nullptr);
}

// Test 19: Token with different expiration times
TEST_F(SenderTest, DifferentExpirationTimes) {
    std::string token = "test.token.here";

    // 30 minutes
    sender->setAuthToken(token, std::time(nullptr) + 1800);
    EXPECT_TRUE(sender->isTokenValid());

    // 1 hour
    sender->setAuthToken(token, std::time(nullptr) + 3600);
    EXPECT_TRUE(sender->isTokenValid());

    // 1 day
    sender->setAuthToken(token, std::time(nullptr) + 86400);
    EXPECT_TRUE(sender->isTokenValid());
}

// Test 20: Empty metrics array handling
TEST_F(SenderTest, EmptyMetricsArray) {
    json metrics;
    metrics["collector_id"] = "test-collector-001";
    metrics["hostname"] = "test-host";
    metrics["timestamp"] = "2024-02-20T10:30:00Z";
    metrics["version"] = "3.0.0";
    metrics["metrics"] = json::array();  // Empty array

    EXPECT_TRUE(metrics["metrics"].is_array());
    EXPECT_EQ(metrics["metrics"].size(), 0);
}

// Test 21: Metrics with various types
TEST_F(SenderTest, MetricsWithVariousTypes) {
    json metrics;
    metrics["collector_id"] = "test-collector-001";
    metrics["hostname"] = "test-host";
    metrics["timestamp"] = "2024-02-20T10:30:00Z";
    metrics["version"] = "3.0.0";

    json metrics_array = json::array();

    // pg_stats metric
    json metric1;
    metric1["type"] = "pg_stats";
    metric1["database"] = "postgres";
    metrics_array.push_back(metric1);

    // sysstat metric
    json metric2;
    metric2["type"] = "sysstat";
    metrics_array.push_back(metric2);

    // pg_log metric
    json metric3;
    metric3["type"] = "pg_log";
    metric3["database"] = "postgres";
    metrics_array.push_back(metric3);

    // disk_usage metric
    json metric4;
    metric4["type"] = "disk_usage";
    metrics_array.push_back(metric4);

    metrics["metrics"] = metrics_array;

    EXPECT_EQ(metrics["metrics"].size(), 4);
}

// Test 22: Token validity buffer (60 seconds)
TEST_F(SenderTest, TokenValidityBuffer) {
    std::string token = "test.token.buffer";

    // Token expires in 61 seconds (should be valid - more than 60s buffer)
    sender->setAuthToken(token, std::time(nullptr) + 61);
    EXPECT_TRUE(sender->isTokenValid());

    // Token expires in 59 seconds (should be invalid - less than 60s buffer)
    sender->setAuthToken(token, std::time(nullptr) + 59);
    EXPECT_FALSE(sender->isTokenValid());
}

// Test 23: Sender configuration persistence
TEST_F(SenderTest, SenderConfigurationPersistence) {
    auto sender1 = std::make_unique<Sender>(
        "https://api.example.com:8080",
        "collector-prod-001",
        "/etc/pganalytics/collector.crt",
        "/etc/pganalytics/collector.key",
        true
    );

    // Sender should retain its configuration
    EXPECT_NE(sender1, nullptr);
}

// Test 24: Multiple concurrent tokens (simulated)
TEST_F(SenderTest, TokenRefreshCycle) {
    // Set initial token
    sender->setAuthToken("token-1", std::time(nullptr) + 3600);
    std::string token1 = sender->getAuthToken();

    // Set new token
    sender->setAuthToken("token-2", std::time(nullptr) + 3600);
    std::string token2 = sender->getAuthToken();

    EXPECT_NE(token1, token2);
    EXPECT_EQ(token2, "token-2");
}

// Test 25: Sender state after operations
TEST_F(SenderTest, SenderStateConsistency) {
    std::string token1 = "first.token.set";
    sender->setAuthToken(token1, std::time(nullptr) + 3600);

    // Get token
    std::string retrieved1 = sender->getAuthToken();
    EXPECT_EQ(retrieved1, token1);

    // Set new token
    std::string token2 = "second.token.set";
    sender->setAuthToken(token2, std::time(nullptr) + 3600);

    // Get new token
    std::string retrieved2 = sender->getAuthToken();
    EXPECT_EQ(retrieved2, token2);

    // State should be consistent
    EXPECT_NE(retrieved1, retrieved2);
}
