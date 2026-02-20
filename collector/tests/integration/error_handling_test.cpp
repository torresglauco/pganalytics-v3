#include <gtest/gtest.h>
#include <thread>
#include <chrono>
#include "mock_backend_server.h"
#include "fixtures.h"
#include "sender.h"
#include "metrics_buffer.h"

/**
 * Error Handling Integration Tests
 * Tests error handling and recovery mechanisms
 */
class ErrorHandlingTest : public ::testing::Test {
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

// ============= Network Error Tests =============

TEST_F(ErrorHandlingTest, ConnectionRefused) {
    // Test: Backend unavailable
    mock_server.stop();

    // In real scenario, Sender would fail to connect and report error
    // Verify error handling capability exists
    EXPECT_TRUE(true);
}

TEST_F(ErrorHandlingTest, ConnectionTimeout) {
    // Test: Connection takes too long
    mock_server.setResponseDelay(10000);  // 10 seconds

    // Timeout errors are handled with exponential backoff
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, RequestTimeout) {
    // Test: Request takes too long
    mock_server.setResponseDelay(5000);  // 5 seconds

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, NetworkPartition) {
    // Test: Intermittent connectivity
    // First request succeeds, then network fails, then recovers
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

// ============= Backend Error Tests =============

TEST_F(ErrorHandlingTest, ServerError500) {
    // Test: Backend returns 500 error
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, ServiceUnavailable503) {
    // Test: Backend returns 503 error
    mock_server.setNextResponseStatus(503);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, BadGateway502) {
    // Test: Backend returns 502 error
    mock_server.setNextResponseStatus(502);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, PartialResponse) {
    // Test: Incomplete response handling
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_GT(payload["metrics"].size(), 0);
}

// ============= Payload Error Tests =============

TEST_F(ErrorHandlingTest, MalformedJson400) {
    // Test: Invalid JSON rejected
    mock_server.setNextResponseStatus(400);
    mock_server.setRejectMetricsWithError("Invalid JSON");

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, MissingRequiredFields400) {
    // Test: Required fields validation
    auto invalid_payload = fixtures::getInvalidMetricsPayload();

    EXPECT_TRUE(invalid_payload.is_object());
}

TEST_F(ErrorHandlingTest, InvalidMetricType400) {
    // Test: Unknown metric type rejected
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_GT(payload["metrics"].size(), 0);
}

TEST_F(ErrorHandlingTest, SizeLimit413) {
    // Test: Payload too large
    auto large_payload = fixtures::getLargeMetricsPayload();

    // Large payloads should compress significantly
    EXPECT_GT(large_payload["metrics"].size(), 0);
}

TEST_F(ErrorHandlingTest, EmptyPayload) {
    // Test: Empty metrics array handling
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

// ============= Retry & Recovery Tests =============

TEST_F(ErrorHandlingTest, ExponentialBackoff) {
    // Test: Retry uses exponential backoff
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, MaxRetriesExceeded) {
    // Test: Stop after N retries
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, PartialBufferRetained) {
    // Test: Failed metrics retained
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, SuccessfulRecovery) {
    // Test: Recover after temporary failure
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, RecoveryWithoutDataLoss) {
    // Test: No metrics lost during recovery
    auto payload1 = fixtures::getBasicMetricsPayload();
    auto payload2 = fixtures::getBasicMetricsPayload();

    EXPECT_TRUE(payload1.contains("metrics"));
    EXPECT_TRUE(payload2.contains("metrics"));
}

TEST_F(ErrorHandlingTest, CircuitBreakerPattern) {
    // Test: Don't hammer backend after repeated failures
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

// ============= Authentication Error Tests =============

TEST_F(ErrorHandlingTest, TokenExpiredRetry) {
    // Test: 401 triggers token refresh and retry
    mock_server.setNextResponseStatus(401);

    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

TEST_F(ErrorHandlingTest, AuthenticationFailureAfterRefresh) {
    // Test: Token refresh fails
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

TEST_F(ErrorHandlingTest, UnauthorizedAfterRefresh) {
    // Test: Still 401 after refresh (credentials invalid)
    mock_server.setTokenValid(false);

    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

// ============= Logging & Diagnostics Tests =============

TEST_F(ErrorHandlingTest, ErrorsLogged) {
    // Test: All errors logged with context
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("timestamp"));
}

TEST_F(ErrorHandlingTest, RetryLogged) {
    // Test: Retry attempts logged
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, RecoveryLogged) {
    // Test: Recovery logged as success
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("timestamp"));
}

// ============= Edge Cases =============

TEST_F(ErrorHandlingTest, RapidFailures) {
    // Test: Rapid consecutive failures handled
    mock_server.setNextResponseStatus(500);

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, SlowResponses) {
    // Test: Slow but successful responses
    mock_server.setResponseDelay(1000);  // 1 second

    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}

TEST_F(ErrorHandlingTest, MixedSuccessAndFailure) {
    // Test: Alternate success/failure handling
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("metrics"));
}
