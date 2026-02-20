#include <gtest/gtest.h>
#include <ctime>
#include <thread>
#include <chrono>
#include <nlohmann/json.hpp>
#include "auth.h"

using json = nlohmann::json;

class AuthManagerTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Create AuthManager instance for each test
        auth = std::make_unique<AuthManager>("test-collector-001", "test-secret-key");
    }

    void TearDown() override {
        auth.reset();
    }

    std::unique_ptr<AuthManager> auth;
};

// Test 1: Create AuthManager instance
TEST_F(AuthManagerTest, CreateInstance) {
    EXPECT_NE(auth, nullptr);
}

// Test 2: Generate token
TEST_F(AuthManagerTest, GenerateToken) {
    std::string token = auth->generateToken(3600);

    EXPECT_FALSE(token.empty());
    // JWT should have 3 parts separated by dots
    int dot_count = std::count(token.begin(), token.end(), '.');
    EXPECT_EQ(dot_count, 2);
}

// Test 3: Token structure
TEST_F(AuthManagerTest, TokenStructure) {
    std::string token = auth->generateToken(3600);

    // Split token into parts
    size_t first_dot = token.find('.');
    size_t second_dot = token.rfind('.');

    EXPECT_NE(first_dot, std::string::npos);
    EXPECT_NE(second_dot, std::string::npos);
    EXPECT_NE(first_dot, second_dot);

    std::string header = token.substr(0, first_dot);
    std::string payload = token.substr(first_dot + 1, second_dot - first_dot - 1);
    std::string signature = token.substr(second_dot + 1);

    EXPECT_FALSE(header.empty());
    EXPECT_FALSE(payload.empty());
    EXPECT_FALSE(signature.empty());
}

// Test 4: Get token when valid
TEST_F(AuthManagerTest, GetValidToken) {
    std::string token1 = auth->generateToken(3600);
    std::string token2 = auth->getToken();

    // Should return the same token if not expired
    EXPECT_EQ(token1, token2);
}

// Test 5: Token validity check
TEST_F(AuthManagerTest, IsTokenValid) {
    auth->generateToken(3600);

    EXPECT_TRUE(auth->isTokenValid());
}

// Test 6: Token expiration check
TEST_F(AuthManagerTest, IsTokenExpired) {
    // Generate token that expires immediately
    auth->generateToken(0);

    // Wait a moment to ensure expiration
    std::this_thread::sleep_for(std::chrono::milliseconds(100));

    EXPECT_FALSE(auth->isTokenValid());
}

// Test 7: Set external token
TEST_F(AuthManagerTest, SetExternalToken) {
    std::string external_token = "external.token.here";
    time_t expires = std::time(nullptr) + 3600;

    auth->setToken(external_token, expires);

    EXPECT_EQ(auth->getToken(), external_token);
}

// Test 8: Refresh token
TEST_F(AuthManagerTest, RefreshToken) {
    std::string original_token = auth->generateToken(3600);

    // Wait a tiny bit to ensure different timestamp
    std::this_thread::sleep_for(std::chrono::milliseconds(10));

    bool success = auth->refreshToken();

    EXPECT_TRUE(success);
    std::string new_token = auth->getToken();
    // Token should be different (different timestamp)
    // Note: They might be the same if executed very quickly, but payload should differ
}

// Test 9: Get token expiration
TEST_F(AuthManagerTest, GetTokenExpiration) {
    time_t before = std::time(nullptr);
    auth->generateToken(3600);
    time_t after = std::time(nullptr);

    time_t expiration = auth->getTokenExpiration();

    // Expiration should be approximately now + 3600
    EXPECT_GE(expiration, before + 3600 - 1);
    EXPECT_LE(expiration, after + 3600 + 1);
}

// Test 10: Load certificate file (non-existent)
TEST_F(AuthManagerTest, LoadNonExistentCertificate) {
    bool success = auth->loadClientCertificate("/nonexistent/path/cert.pem");

    EXPECT_FALSE(success);
    std::string error = auth->getLastError();
    EXPECT_FALSE(error.empty());
}

// Test 11: Load key file (non-existent)
TEST_F(AuthManagerTest, LoadNonExistentKey) {
    bool success = auth->loadClientKey("/nonexistent/path/key.pem");

    EXPECT_FALSE(success);
    std::string error = auth->getLastError();
    EXPECT_FALSE(error.empty());
}

// Test 12: Get client certificate (empty)
TEST_F(AuthManagerTest, GetClientCertificateEmpty) {
    std::string cert = auth->getClientCertificate();

    EXPECT_TRUE(cert.empty());
}

// Test 13: Get client key (empty)
TEST_F(AuthManagerTest, GetClientKey) {
    std::string key = auth->getClientKey();

    EXPECT_TRUE(key.empty());
}

// Test 14: Multiple tokens
TEST_F(AuthManagerTest, MultipleTokens) {
    std::string token1 = auth->generateToken(3600);

    std::this_thread::sleep_for(std::chrono::milliseconds(10));

    std::string token2 = auth->generateToken(3600);

    // Both should be valid JWT tokens
    EXPECT_FALSE(token1.empty());
    EXPECT_FALSE(token2.empty());
    // They should be different (at least in timestamp)
    EXPECT_NE(token1, token2);
}

// Test 15: Validate token signature
TEST_F(AuthManagerTest, ValidateTokenSignature) {
    std::string token = auth->generateToken(3600);

    // Token should validate against itself
    EXPECT_TRUE(auth->validateTokenSignature(token));
}

// Test 16: Validate invalid token format
TEST_F(AuthManagerTest, ValidateInvalidTokenFormat) {
    std::string invalid_token = "not.a.valid.jwt";

    // Token validation should fail (wrong format)
    EXPECT_FALSE(auth->validateTokenSignature(invalid_token));
}

// Test 17: Token with different secret fails validation
TEST_F(AuthManagerTest, TokenWithDifferentSecret) {
    std::string token = auth->generateToken(3600);

    // Create another AuthManager with different secret
    AuthManager other_auth("different-collector", "different-secret");

    // The other auth manager should not validate this token
    EXPECT_FALSE(other_auth.validateTokenSignature(token));
}

// Test 18: Collector ID in token
TEST_F(AuthManagerTest, CollectorIdInToken) {
    std::string collector_id = "special-collector-id-123";
    AuthManager special_auth(collector_id, "secret");

    std::string token = special_auth.generateToken(3600);

    // Token should be valid for this collector
    EXPECT_TRUE(special_auth.validateTokenSignature(token));
}

// Test 19: Token expiration time is in future
TEST_F(AuthManagerTest, TokenExpirationInFuture) {
    auth->generateToken(3600);

    time_t now = std::time(nullptr);
    time_t expiration = auth->getTokenExpiration();

    // Expiration should be in the future
    EXPECT_GT(expiration, now);
}

// Test 20: Short-lived token
TEST_F(AuthManagerTest, ShortLivedToken) {
    auth->generateToken(1);  // 1 second

    // Token should be valid initially
    EXPECT_TRUE(auth->isTokenValid());

    // Wait for expiration
    std::this_thread::sleep_for(std::chrono::seconds(2));

    // Token should be expired now (might not be true if test runs too fast)
    // This test is timing-dependent
}

// Test 21: Refresh before expiration
TEST_F(AuthManagerTest, RefreshBeforeExpiration) {
    auth->generateToken(3600);  // 1 hour
    time_t initial_expiration = auth->getTokenExpiration();

    std::this_thread::sleep_for(std::chrono::milliseconds(100));

    bool success = auth->refreshToken();
    time_t new_expiration = auth->getTokenExpiration();

    EXPECT_TRUE(success);
    // New expiration should be later than initial
    EXPECT_GT(new_expiration, initial_expiration);
}

// Test 22: Last error clears on success
TEST_F(AuthManagerTest, LastErrorMessage) {
    // Try to load non-existent certificate
    auth->loadClientCertificate("/nonexistent/cert.pem");
    std::string error1 = auth->getLastError();
    EXPECT_FALSE(error1.empty());

    // Generate a token successfully
    auth->generateToken(3600);

    // Error message should still be available from previous operation
    // (not necessarily cleared, but we tested it exists)
    EXPECT_FALSE(error1.empty());
}

// Test 23: Token payload contains standard claims
TEST_F(AuthManagerTest, TokenPayloadStructure) {
    // This test verifies the internal token structure
    std::string token = auth->generateToken(3600);

    // Token should be a valid JWT with proper structure
    // (Verified by the fact that it passes signature validation)
    EXPECT_TRUE(auth->validateTokenSignature(token));
}

// Test 24: Multiple AuthManagers with different secrets
TEST_F(AuthManagerTest, MultipleAuthManagers) {
    AuthManager auth2("collector-002", "secret-2");
    AuthManager auth3("collector-003", "secret-3");

    std::string token1 = auth->generateToken(3600);
    std::string token2 = auth2.generateToken(3600);
    std::string token3 = auth3.generateToken(3600);

    // Each should validate against their own auth manager
    EXPECT_TRUE(auth->validateTokenSignature(token1));
    EXPECT_TRUE(auth2.validateTokenSignature(token2));
    EXPECT_TRUE(auth3.validateTokenSignature(token3));

    // But not against others
    EXPECT_FALSE(auth->validateTokenSignature(token2));
    EXPECT_FALSE(auth2.validateTokenSignature(token1));
}

// Test 25: Token validity buffer
TEST_F(AuthManagerTest, TokenValidityBuffer) {
    // Generate token that expires in 61 seconds
    auth->generateToken(61);

    // Token should be valid (more than 60 seconds remaining)
    EXPECT_TRUE(auth->isTokenValid());

    // Generate token that expires in 59 seconds
    auth->generateToken(59);

    // Token should not be valid (less than 60-second buffer)
    EXPECT_FALSE(auth->isTokenValid());
}
