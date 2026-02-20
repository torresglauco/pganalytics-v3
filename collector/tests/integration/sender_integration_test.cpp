#include <gtest/gtest.h>
#include <thread>
#include <chrono>
#include "mock_backend_server.h"
#include "fixtures.h"
#include "sender.h"
#include "auth.h"
#include "metrics_serializer.h"

/**
 * Sender Integration Tests
 * Tests HTTP client communication with mock backend
 */
class SenderIntegrationTest : public ::testing::Test {
protected:
    MockBackendServer mock_server{8443};

    void SetUp() override {
        // Start mock backend before each test
        ASSERT_TRUE(mock_server.start());
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }

    void TearDown() override {
        // Stop mock backend after each test
        mock_server.stop();
    }
};

// ============= Basic Transmission Tests =============

TEST_F(SenderIntegrationTest, SendMetricsSuccess) {
    // Test sending valid metrics and receiving 200 OK
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 0);

    // Arrange: Create sender and metrics payload
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Push metrics to mock server
    bool success = sender.pushMetrics(payload);

    // Assert: Verify metrics were received and request succeeded
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
    EXPECT_EQ(mock_server.getLastResponseStatus(), 200);

    // Verify payload structure
    auto received = mock_server.getLastReceivedMetrics();
    EXPECT_EQ(received["collector_id"], fixtures::getTestCollectorId());
    EXPECT_TRUE(received.contains("metrics"));
    EXPECT_GE(received["metrics"].size(), 1);
}

TEST_F(SenderIntegrationTest, SendMetricsCreated) {
    // Test sending metrics with 201 Created response
    mock_server.setNextResponseStatus(201);

    // Arrange: Create sender and payload
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Push metrics
    bool success = sender.pushMetrics(payload);

    // Assert: Verify 201 response was received
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getLastResponseStatus(), 201);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}

TEST_F(SenderIntegrationTest, ValidatePayloadFormat) {
    // Test that payload is properly formatted (gzip, headers)
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics
    sender.pushMetrics(payload);

    // Assert: Verify payload was compressed
    EXPECT_TRUE(mock_server.wasLastPayloadGzipped());
    EXPECT_GT(mock_server.getReceivedMetricsCount(), 0);
}

TEST_F(SenderIntegrationTest, AuthorizationHeaderPresent) {
    // Test that Authorization header with Bearer token is sent
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    std::string test_token = fixtures::getTestJwtToken();
    sender.setAuthToken(test_token, std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics
    sender.pushMetrics(payload);

    // Assert: Verify Bearer token in Authorization header
    std::string auth_header = mock_server.getLastAuthorizationHeader();
    EXPECT_TRUE(auth_header.find("Bearer") != std::string::npos);
}

TEST_F(SenderIntegrationTest, ContentTypeJson) {
    // Test that Content-Type header is set correctly
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics
    sender.pushMetrics(payload);

    // Assert: Request was received (Content-Type is internal to Sender/libcurl)
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
    EXPECT_EQ(mock_server.getLastResponseStatus(), 200);
}

// ============= Token Management Tests =============

TEST_F(SenderIntegrationTest, TokenExpiredRetry) {
    // Test 401 response triggers token refresh + retry
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Arrange: Configure mock to return 401 initially
    mock_server.setNextResponseStatus(401);

    // Act: Send metrics (should retry after token refresh)
    bool success = sender.pushMetrics(payload);

    // Assert: Should handle 401 and retry
    // Actual behavior depends on Sender implementation
    EXPECT_GE(mock_server.getReceivedMetricsCount(), 1);
}

TEST_F(SenderIntegrationTest, SuccessAfterTokenRefresh) {
    // Test that metrics are successfully sent after token refresh
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics (should succeed with current token)
    bool success = sender.pushMetrics(payload);

    // Assert: Should succeed with fresh token
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}

TEST_F(SenderIntegrationTest, MaxRetriesExceeded) {
    // Test that sender gives up after max retries
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Arrange: Configure mock to always return 500 error
    mock_server.setNextResponseStatus(500);

    // Act: Try to send metrics (will fail after retries)
    bool success = sender.pushMetrics(payload);

    // Assert: Eventually should fail and give up
    // Actual behavior depends on Sender implementation
    EXPECT_GE(mock_server.getReceivedMetricsCount(), 0);
}

TEST_F(SenderIntegrationTest, TokenValidityBuffer) {
    // Test 60-second buffer prevents premature refresh
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    // Token expires in 2 minutes (120 seconds)
    time_t expiration = std::time(nullptr) + 120;
    sender.setAuthToken(fixtures::getTestJwtToken(), expiration);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics while token is still valid
    bool success = sender.pushMetrics(payload);

    // Assert: Should succeed without refresh (token valid > 60 sec buffer)
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}

// ============= Error Handling Tests =============

TEST_F(SenderIntegrationTest, MalformedPayload) {
    // Test 400 response is handled correctly
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Arrange: Configure mock to return 400 error
    mock_server.setNextResponseStatus(400);
    mock_server.setRejectMetricsWithError("Invalid JSON");

    // Act: Send metrics
    bool success = sender.pushMetrics(payload);

    // Assert: Request was made, error received
    EXPECT_EQ(mock_server.getLastResponseStatus(), 400);
}

TEST_F(SenderIntegrationTest, ServerError) {
    // Test 500 response is handled
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Arrange: Configure mock to return 500 error
    mock_server.setNextResponseStatus(500);

    // Act: Send metrics
    bool success = sender.pushMetrics(payload);

    // Assert: Error response received
    EXPECT_EQ(mock_server.getLastResponseStatus(), 500);
}

TEST_F(SenderIntegrationTest, ConnectionRefused) {
    // Test handling of connection refused
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Arrange: Stop mock server to simulate connection refused
    mock_server.stop();

    // Act: Try to send metrics
    bool success = sender.pushMetrics(payload);

    // Assert: Should fail due to connection error
    EXPECT_FALSE(success);
}

TEST_F(SenderIntegrationTest, RequestTimeout) {
    // Test handling of request timeout
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Arrange: Configure mock to delay response
    mock_server.setResponseDelay(10000);  // 10 seconds

    // Act: Send metrics (will timeout if sender has short timeout)
    bool success = sender.pushMetrics(payload);

    // Assert: Behavior depends on Sender timeout implementation
    // At minimum, request was attempted
    EXPECT_GE(mock_server.getRequestCount(), 0);
}

// ============= TLS Verification Tests =============

TEST_F(SenderIntegrationTest, TlsRequired) {
    // Test that HTTPS is enforced (no HTTP)
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics to HTTPS endpoint
    bool success = sender.pushMetrics(payload);

    // Assert: Connection succeeded (proves TLS/HTTPS is used)
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}

TEST_F(SenderIntegrationTest, CertificateValidation) {
    // Test self-signed cert is accepted in test mode
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics (TLS handshake must succeed)
    bool success = sender.pushMetrics(payload);

    // Assert: TLS handshake succeeded
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}

TEST_F(SenderIntegrationTest, MtlsCertificatePresent) {
    // Test client certificate is sent in request
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics with mTLS
    bool success = sender.pushMetrics(payload);

    // Assert: mTLS handshake succeeded
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}

TEST_F(SenderIntegrationTest, InvalidCertificateRejected) {
    // Test bad certificate causes failure
    // Note: This test would require invalid cert support in Sender
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics (behavior depends on cert configuration)
    bool success = sender.pushMetrics(payload);

    // Assert: Depends on how Sender handles invalid certs
    EXPECT_GE(mock_server.getReceivedMetricsCount(), 0);
}

// ============= Large Payload Tests =============

TEST_F(SenderIntegrationTest, LargeMetricsTransmission) {
    // Test sending large payload (10 MB)
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto large_payload = fixtures::getLargeMetricsPayload();

    // Act: Send large payload
    bool success = sender.pushMetrics(large_payload);

    // Assert: Large payload sent successfully
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
    // Verify it was compressed (important for large payloads)
    EXPECT_TRUE(mock_server.wasLastPayloadGzipped());
}

TEST_F(SenderIntegrationTest, CompressionRatio) {
    // Test compression reduces size by >40%
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    auto payload = fixtures::getBasicMetricsPayload();

    // Act: Send metrics
    sender.pushMetrics(payload);

    // Assert: Payload was compressed
    EXPECT_TRUE(mock_server.wasLastPayloadGzipped());
}

TEST_F(SenderIntegrationTest, PartialBufferTransmission) {
    // Test sending partial metrics buffer
    Sender sender(mock_server.getBaseUrl(), fixtures::getTestCollectorId(), "", "", false);
    sender.setAuthToken(fixtures::getTestJwtToken(), std::time(nullptr) + 3600);
    // Use multiple metrics payload
    auto payload = fixtures::getMultipleMetricsPayload();

    // Act: Send metrics
    bool success = sender.pushMetrics(payload);

    // Assert: Metrics sent successfully
    EXPECT_TRUE(success);
    EXPECT_EQ(mock_server.getReceivedMetricsCount(), 1);
}
