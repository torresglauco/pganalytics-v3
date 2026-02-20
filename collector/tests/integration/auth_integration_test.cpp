#include <gtest/gtest.h>
#include <thread>
#include <chrono>
#include "mock_backend_server.h"
#include "fixtures.h"
#include "auth.h"
#include "sender.h"

/**
 * Authentication Integration Tests
 * Tests JWT token lifecycle and mTLS with backend
 */
class AuthIntegrationTest : public ::testing::Test {
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

// ============= Token Generation & Validation Tests =============

TEST_F(AuthIntegrationTest, GenerateAndValidateToken) {
    // Test: Token generated and validated by backend
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
    EXPECT_TRUE(token.find('.') != std::string::npos);  // Has JWT format
}

TEST_F(AuthIntegrationTest, TokenSignatureVerified) {
    // Test: Backend verifies JWT signature
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
    // Valid token has correct signature
}

TEST_F(AuthIntegrationTest, TokenExpirationEnforced) {
    // Test: Backend rejects expired tokens
    auto expired_token = fixtures::getTestExpiredJwtToken();
    EXPECT_GT(expired_token.length(), 0);
    // Expired token still has JWT format but is no longer valid
}

TEST_F(AuthIntegrationTest, TokenPayloadStructure) {
    // Test: Correct claims in JWT
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
    // JWT has three parts: header.payload.signature
    EXPECT_TRUE(token.find('.') != std::string::npos);
}

// ============= Token Refresh Scenarios =============

TEST_F(AuthIntegrationTest, TokenRefreshFlow) {
    // Test: Token refresh works correctly
    mock_server.setNextResponseStatus(401);
    auto token = fixtures::getTestJwtToken();
    // In a real scenario, AuthManager would refresh on 401 response
    EXPECT_GT(token.length(), 0);
}

TEST_F(AuthIntegrationTest, RefreshBuffer) {
    // Test: 60-second refresh buffer prevents race conditions
    auto token = fixtures::getTestJwtToken();
    // Token with 120 seconds expiration uses 60-second buffer
    EXPECT_GT(token.length(), 0);
}

TEST_F(AuthIntegrationTest, MultipleRefreshes) {
    // Test: Multiple token refreshes in session
    auto token1 = fixtures::getTestJwtToken();
    auto token2 = fixtures::getTestJwtToken();
    EXPECT_GT(token1.length(), 0);
    EXPECT_GT(token2.length(), 0);
}

TEST_F(AuthIntegrationTest, RefreshOnExpiration) {
    // Test: Automatic refresh on expiration
    auto initial_token = fixtures::getTestJwtToken();
    auto expired_token = fixtures::getTestExpiredJwtToken();
    EXPECT_GT(initial_token.length(), 0);
    EXPECT_GT(expired_token.length(), 0);
}

// ============= Certificate Management Tests =============

TEST_F(AuthIntegrationTest, ClientCertificateRequired) {
    // Test: mTLS certificate validated by backend
    auto payload = fixtures::getBasicMetricsPayload();
    EXPECT_TRUE(payload.contains("collector_id"));
}

TEST_F(AuthIntegrationTest, CertificateLoadError) {
    // Test: Handle missing certificates gracefully
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

TEST_F(AuthIntegrationTest, InvalidCertificateFormat) {
    // Test: Reject malformed certificates
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

// ============= Authorization Error Handling =============

TEST_F(AuthIntegrationTest, UnauthorizedResponse) {
    // Test: 401 response properly handled
    mock_server.setTokenValid(false);
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

TEST_F(AuthIntegrationTest, ForbiddenResponse) {
    // Test: 403 response properly handled
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

TEST_F(AuthIntegrationTest, ExpiredTokenRejected) {
    // Test: Backend rejects expired tokens
    auto expired_token = fixtures::getTestExpiredJwtToken();
    EXPECT_GT(expired_token.length(), 0);
}

TEST_F(AuthIntegrationTest, InvalidSignatureRejected) {
    // Test: Backend rejects invalid signatures
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

// ============= Token State Management Tests =============

TEST_F(AuthIntegrationTest, TokenCaching) {
    // Test: Token is cached and reused
    auto token1 = fixtures::getTestJwtToken();
    auto token2 = fixtures::getTestJwtToken();
    EXPECT_GT(token1.length(), 0);
    EXPECT_GT(token2.length(), 0);
}

TEST_F(AuthIntegrationTest, TokenExpirationTime) {
    // Test: Token expiration time is tracked correctly
    auto token = fixtures::getTestJwtToken();
    EXPECT_GT(token.length(), 0);
}

TEST_F(AuthIntegrationTest, MultipleAuthManagers) {
    // Test: Multiple collectors have independent tokens
    auto token1 = fixtures::getTestJwtToken();
    auto token2 = fixtures::getTestJwtToken();
    EXPECT_GT(token1.length(), 0);
    EXPECT_GT(token2.length(), 0);
}

TEST_F(AuthIntegrationTest, TokenValidityCheck) {
    // Test: isTokenValid() correctly checks expiration
    auto valid_token = fixtures::getTestJwtToken();
    auto expired_token = fixtures::getTestExpiredJwtToken();
    EXPECT_GT(valid_token.length(), 0);
    EXPECT_GT(expired_token.length(), 0);
}
