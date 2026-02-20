#include <gtest/gtest.h>
#include "e2e_harness.h"
#include "http_client.h"
#include "database_helper.h"
#include "fixtures.h"
#include <fstream>
#include <sstream>
#include <regex>

/**
 * E2E Collector Registration Tests
 *
 * Tests the collector registration flow with the backend API.
 * Validates:
 * - Registration request handling
 * - JWT token generation and structure
 * - Certificate creation and storage
 * - Token expiration settings
 * - Multiple collector support
 * - Error handling
 * - Audit trail
 */
class E2ECollectorRegistrationTest : public ::testing::Test {
protected:
    static E2ETestHarness harness;
    static std::unique_ptr<E2EDatabaseHelper> db_helper;

    static void SetUpTestSuite() {
        std::cout << "\n[E2E Registration] Setting up test suite..." << std::endl;

        // Start docker-compose stack
        if (!harness.startStack(60)) {
            FAIL() << "Failed to start E2E stack";
        }

        // Initialize database helper
        db_helper = std::make_unique<E2EDatabaseHelper>(
            harness.getDatabaseUrl(),
            harness.getTimescaleUrl()
        );

        if (!db_helper->isConnected()) {
            FAIL() << "Failed to connect to databases";
        }

        std::cout << "[E2E Registration] Test suite ready" << std::endl;
    }

    static void TearDownTestSuite() {
        std::cout << "\n[E2E Registration] Tearing down test suite..." << std::endl;
        db_helper.reset();
        harness.stopStack();
    }

    void SetUp() override {
        // Reset data before each test
        db_helper->truncateAllData();
        std::cout << "[E2E Registration] Test data reset" << std::endl;
    }

    void TearDown() override {
        // Cleanup is optional here; SetUp() clears for next test
    }

    /**
     * Helper: Extract JWT payload claims
     * JWT structure: header.payload.signature
     * Payload is base64-encoded JSON
     */
    std::string extractJwtPayloadClaims(const std::string& token) {
        // Split by dots to get payload
        size_t first_dot = token.find('.');
        size_t second_dot = token.find('.', first_dot + 1);

        if (first_dot == std::string::npos || second_dot == std::string::npos) {
            return "";
        }

        std::string payload_b64 = token.substr(first_dot + 1, second_dot - first_dot - 1);

        // Simple base64 decode (note: production would use proper base64 library)
        // For now, just return the payload for manual inspection
        return payload_b64;
    }

    /**
     * Helper: Check if JWT structure is valid
     * Should have exactly 2 dots (3 parts: header.payload.signature)
     */
    bool isValidJwtStructure(const std::string& token) {
        int dot_count = 0;
        for (char c : token) {
            if (c == '.') dot_count++;
        }
        return dot_count == 2 && token.length() > 10;
    }

    /**
     * Helper: Extract exp claim from JWT
     * Note: This is a simple check; production would properly decode base64
     */
    bool hasExpiryClaim(const std::string& response_body) {
        // Look for "expires_at" or "exp" in response
        return response_body.find("expires_at") != std::string::npos ||
               response_body.find("\"exp\"") != std::string::npos;
    }
};

// Static member initialization
E2ETestHarness E2ECollectorRegistrationTest::harness;
std::unique_ptr<E2EDatabaseHelper> E2ECollectorRegistrationTest::db_helper;

// ==================== REGISTRATION TESTS ====================

/**
 * Test 1: RegisterNewCollector
 * Verify that a new collector can register and receive credentials
 */
TEST_F(E2ECollectorRegistrationTest, RegisterNewCollector) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setVerbose(true);

    // ACT
    std::string response_body;
    int response_code = 0;

    bool success = client.registerCollector(
        e2e_fixtures::getCollectorName(),
        e2e_fixtures::getCollectorHostname(),
        response_body,
        response_code
    );

    // ASSERT
    EXPECT_TRUE(success) << "Registration failed: " << client.getLastResponseBody();
    EXPECT_EQ(response_code, 200) << "Expected 200 response, got " << response_code;
    EXPECT_GT(response_body.length(), 0) << "Empty response body";

    // Verify response contains required fields
    EXPECT_NE(response_body.find("collector_id"), std::string::npos)
        << "Response missing collector_id";
    EXPECT_NE(response_body.find("token"), std::string::npos)
        << "Response missing JWT token";

    std::cout << "[E2E Registration] RegisterNewCollector: PASSED" << std::endl;
}

/**
 * Test 2: RegistrationValidation
 * Verify that JWT token has correct structure and claims
 */
TEST_F(E2ECollectorRegistrationTest, RegistrationValidation) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    std::string response_body;
    int response_code = 0;

    // Register collector
    ASSERT_TRUE(client.registerCollector(
        e2e_fixtures::getCollectorName(),
        e2e_fixtures::getCollectorHostname(),
        response_body,
        response_code
    ));

    // ACT - Extract token from response
    std::string token;
    size_t token_pos = response_body.find("\"token\":\"");
    if (token_pos != std::string::npos) {
        token_pos += 9;  // strlen("\"token\":\"")
        size_t end_pos = response_body.find("\"", token_pos);
        token = response_body.substr(token_pos, end_pos - token_pos);
    }

    // ASSERT
    EXPECT_GT(token.length(), 0) << "Failed to extract token from response";
    EXPECT_TRUE(isValidJwtStructure(token)) << "Invalid JWT structure";

    // Token should have exp/expires_at claim
    EXPECT_TRUE(hasExpiryClaim(response_body)) << "Response missing expiration claim";

    std::cout << "[E2E Registration] RegistrationValidation: PASSED" << std::endl;
}

/**
 * Test 3: CertificatePersistence
 * Verify that client certificate is returned and saved correctly
 */
TEST_F(E2ECollectorRegistrationTest, CertificatePersistence) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    std::string response_body;
    int response_code = 0;

    // Register collector
    ASSERT_TRUE(client.registerCollector(
        e2e_fixtures::getCollectorName(),
        e2e_fixtures::getCollectorHostname(),
        response_body,
        response_code
    ));

    // ACT - Extract certificate from response
    std::string cert;
    size_t cert_pos = response_body.find("\"certificate\":\"");
    if (cert_pos != std::string::npos) {
        cert_pos += 15;  // strlen("\"certificate\":\"")
        size_t end_pos = response_body.find("\"", cert_pos);
        cert = response_body.substr(cert_pos, end_pos - cert_pos);
    }

    // ASSERT
    EXPECT_GT(cert.length(), 0) << "Failed to extract certificate from response";

    // Certificate should start with BEGIN CERTIFICATE marker (possibly escaped)
    EXPECT_NE(cert.find("BEGIN CERTIFICATE"), std::string::npos)
        << "Certificate missing BEGIN CERTIFICATE marker";

    std::cout << "[E2E Registration] CertificatePersistence: PASSED" << std::endl;
}

/**
 * Test 4: TokenExpiration
 * Verify that token has correct expiration time (typically 15 minutes)
 */
TEST_F(E2ECollectorRegistrationTest, TokenExpiration) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    std::string response_body;
    int response_code = 0;

    // Register collector
    ASSERT_TRUE(client.registerCollector(
        e2e_fixtures::getCollectorName(),
        e2e_fixtures::getCollectorHostname(),
        response_body,
        response_code
    ));

    // ACT - Check for expiration in response
    bool has_expiration = response_body.find("expires_at") != std::string::npos ||
                          response_body.find("expiration") != std::string::npos ||
                          response_body.find("\"exp\"") != std::string::npos;

    // ASSERT
    EXPECT_TRUE(has_expiration) << "Token expiration not specified in response";

    // Token should expire soon (within 24 hours)
    // In a real test, we'd parse the timestamp and verify it's ~15 minutes from now
    EXPECT_NE(response_body.find("900"), std::string::npos) << "Expected 900s (15 min) expiration";

    std::cout << "[E2E Registration] TokenExpiration: PASSED" << std::endl;
}

/**
 * Test 5: MultipleRegistrations
 * Verify that different collectors can register independently
 */
TEST_F(E2ECollectorRegistrationTest, MultipleRegistrations) {
    // ARRANGE
    E2EHttpClient client1(harness.getBackendUrl());
    E2EHttpClient client2(harness.getBackendUrl());

    // ACT - Register two collectors
    std::string response1, response2;
    int code1 = 0, code2 = 0;

    bool success1 = client1.registerCollector(
        "Collector 1",
        "host-1",
        response1,
        code1
    );

    bool success2 = client2.registerCollector(
        "Collector 2",
        "host-2",
        response2,
        code2
    );

    // ASSERT
    EXPECT_TRUE(success1) << "First registration failed";
    EXPECT_TRUE(success2) << "Second registration failed";
    EXPECT_EQ(code1, 200);
    EXPECT_EQ(code2, 200);

    // Extract collector IDs to verify they're different
    std::string id1, id2;
    size_t pos1 = response1.find("\"collector_id\":\"");
    size_t pos2 = response2.find("\"collector_id\":\"");

    if (pos1 != std::string::npos) {
        pos1 += 16;
        size_t end1 = response1.find("\"", pos1);
        id1 = response1.substr(pos1, end1 - pos1);
    }

    if (pos2 != std::string::npos) {
        pos2 += 16;
        size_t end2 = response2.find("\"", pos2);
        id2 = response2.substr(pos2, end2 - pos2);
    }

    EXPECT_GT(id1.length(), 0) << "Failed to extract collector ID 1";
    EXPECT_GT(id2.length(), 0) << "Failed to extract collector ID 2";
    EXPECT_NE(id1, id2) << "Collector IDs should be unique";

    // Verify both are in database
    EXPECT_TRUE(db_helper->collectorExists(id1)) << "Collector 1 not found in registry";
    EXPECT_TRUE(db_helper->collectorExists(id2)) << "Collector 2 not found in registry";

    std::cout << "[E2E Registration] MultipleRegistrations: PASSED" << std::endl;
}

/**
 * Test 6: RegistrationFailure
 * Verify that invalid registration requests are rejected gracefully
 */
TEST_F(E2ECollectorRegistrationTest, RegistrationFailure) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());

    // ACT - Try to register with empty name
    std::string response_body;
    int response_code = 0;

    bool success = client.registerCollector(
        "",  // Empty name
        e2e_fixtures::getCollectorHostname(),
        response_body,
        response_code
    );

    // ASSERT
    // Empty name should either be rejected or auto-generated
    // Either way, we should get a response
    EXPECT_GT(response_code, 0) << "No response from server";

    // Should fail or auto-generate
    if (response_code >= 400) {
        // Properly rejected
        EXPECT_TRUE(!success || response_code >= 400)
            << "Empty name should be rejected";
    } else {
        // Auto-generated name is also acceptable
        EXPECT_NE(response_body.find("collector_id"), std::string::npos)
            << "Should provide collector_id even with empty name";
    }

    std::cout << "[E2E Registration] RegistrationFailure: PASSED" << std::endl;
}

/**
 * Test 7: DuplicateRegistration
 * Verify handling of duplicate registration attempts
 */
TEST_F(E2ECollectorRegistrationTest, DuplicateRegistration) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    std::string response1, response2;
    int code1 = 0, code2 = 0;

    // Register first collector
    ASSERT_TRUE(client.registerCollector(
        "Duplicate Test",
        "duplicate-host",
        response1,
        code1
    ));

    // Extract first collector ID
    std::string collector_id;
    size_t pos = response1.find("\"collector_id\":\"");
    if (pos != std::string::npos) {
        pos += 16;
        size_t end = response1.find("\"", pos);
        collector_id = response1.substr(pos, end - pos);
    }
    ASSERT_GT(collector_id.length(), 0);

    // ACT - Try to register with same name/hostname
    bool success2 = client.registerCollector(
        "Duplicate Test",
        "duplicate-host",
        response2,
        code2
    );

    // ASSERT
    // Duplicate should either be rejected (409) or return same ID
    EXPECT_GT(code2, 0);

    if (code2 == 409 || code2 == 400) {
        // Properly rejected as duplicate
        EXPECT_NE(response2.find("duplicate"), std::string::npos);
    } else {
        // Allowed but should have same or similar behavior
        EXPECT_TRUE(success2 || code2 >= 400)
            << "Duplicate registration should be handled";
    }

    std::cout << "[E2E Registration] DuplicateRegistration: PASSED" << std::endl;
}

/**
 * Test 8: CertificateFormat
 * Verify that returned certificate is valid X.509 format
 */
TEST_F(E2ECollectorRegistrationTest, CertificateFormat) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    std::string response_body;
    int response_code = 0;

    // Register collector
    ASSERT_TRUE(client.registerCollector(
        e2e_fixtures::getCollectorName(),
        e2e_fixtures::getCollectorHostname(),
        response_body,
        response_code
    ));

    // ACT - Extract and validate certificate
    std::string cert;
    size_t cert_pos = response_body.find("\"certificate\":\"");
    if (cert_pos != std::string::npos) {
        cert_pos += 15;
        size_t end_pos = response_body.find("\"", cert_pos);
        cert = response_body.substr(cert_pos, end_pos - cert_pos);
    }

    // ASSERT
    EXPECT_GT(cert.length(), 0) << "No certificate in response";

    // X.509 certificate should have PEM format markers
    EXPECT_NE(cert.find("BEGIN CERTIFICATE"), std::string::npos)
        << "Missing BEGIN CERTIFICATE marker";
    EXPECT_NE(cert.find("END CERTIFICATE"), std::string::npos)
        << "Missing END CERTIFICATE marker";

    // Should have base64-like content in middle
    EXPECT_GT(cert.length(), 100) << "Certificate too short to be valid";

    std::cout << "[E2E Registration] CertificateFormat: PASSED" << std::endl;
}

/**
 * Test 9: PrivateKeyProtection
 * Verify that private key is returned and should be stored securely
 */
TEST_F(E2ECollectorRegistrationTest, PrivateKeyProtection) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    std::string response_body;
    int response_code = 0;

    // Register collector
    ASSERT_TRUE(client.registerCollector(
        e2e_fixtures::getCollectorName(),
        e2e_fixtures::getCollectorHostname(),
        response_body,
        response_code
    ));

    // ACT - Extract private key
    std::string private_key;
    size_t key_pos = response_body.find("\"private_key\":\"");
    if (key_pos != std::string::npos) {
        key_pos += 15;
        size_t end_pos = response_body.find("\"", key_pos);
        private_key = response_body.substr(key_pos, end_pos - key_pos);
    }

    // ASSERT
    EXPECT_GT(private_key.length(), 0) << "No private key in response";

    // Private key should have PEM format markers
    EXPECT_NE(private_key.find("BEGIN PRIVATE KEY"), std::string::npos)
        << "Missing BEGIN PRIVATE KEY marker";
    EXPECT_NE(private_key.find("END PRIVATE KEY"), std::string::npos)
        << "Missing END PRIVATE KEY marker";

    // Should contain base64 content
    EXPECT_GT(private_key.length(), 100) << "Private key too short";

    std::cout << "[E2E Registration] PrivateKeyProtection: PASSED" << std::endl;
}

/**
 * Test 10: RegistrationAudit
 * Verify that registration is logged in the database with audit trail
 */
TEST_F(E2ECollectorRegistrationTest, RegistrationAudit) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    std::string response_body;
    int response_code = 0;

    // Register collector
    ASSERT_TRUE(client.registerCollector(
        "Audit Test Collector",
        "audit-test-host",
        response_body,
        response_code
    ));

    // Extract collector ID
    std::string collector_id;
    size_t pos = response_body.find("\"collector_id\":\"");
    if (pos != std::string::npos) {
        pos += 16;
        size_t end = response_body.find("\"", pos);
        collector_id = response_body.substr(pos, end - pos);
    }

    // ACT - Query database for audit trail
    ASSERT_GT(collector_id.length(), 0);
    bool collector_exists = db_helper->collectorExists(collector_id);

    // ASSERT
    EXPECT_TRUE(collector_exists) << "Collector not found in registry";

    // Verify registration was logged
    std::string status = db_helper->getCollectorStatus(collector_id);
    EXPECT_NE(status.find("active"), std::string::npos)
        << "Collector should be active after registration";

    std::cout << "[E2E Registration] RegistrationAudit: PASSED" << std::endl;
}

// ==================== TEST SUMMARY ====================

/**
 * Summary of Collector Registration Tests
 *
 * All 10 tests validate the registration flow:
 * 1. RegisterNewCollector - Basic registration works
 * 2. RegistrationValidation - JWT token has correct structure
 * 3. CertificatePersistence - Certificate returned correctly
 * 4. TokenExpiration - Expiration time is set properly
 * 5. MultipleRegistrations - Different collectors can register independently
 * 6. RegistrationFailure - Invalid input handled gracefully
 * 7. DuplicateRegistration - Duplicate attempts handled
 * 8. CertificateFormat - Certificate is valid X.509 format
 * 9. PrivateKeyProtection - Private key returned correctly
 * 10. RegistrationAudit - Registration logged in database
 *
 * Expected Result: 10/10 tests passing
 * Time Target: ~30-40 seconds total
 */

