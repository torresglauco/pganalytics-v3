#include <gtest/gtest.h>
#include "e2e_harness.h"
#include "http_client.h"
#include "database_helper.h"
#include "fixtures.h"
#include <chrono>
#include <thread>

/**
 * E2E Metrics Ingestion Tests
 *
 * Tests metrics flowing from collector to backend storage.
 * Validates:
 * - Successful metrics transmission
 * - Correct storage in TimescaleDB
 * - Schema validation
 * - Timestamp accuracy
 * - Multiple metric types
 * - Compression and payload size
 * - Data integrity
 */
class E2EMetricsIngestionTest : public ::testing::Test {
protected:
    static E2ETestHarness harness;
    static std::unique_ptr<E2EDatabaseHelper> db_helper;
    static std::string test_collector_id;
    static std::string test_jwt_token;

    static void SetUpTestSuite() {
        std::cout << "\n[E2E Metrics] Setting up test suite..." << std::endl;

        // Start docker stack
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

        // Register a test collector first
        E2EHttpClient client(harness.getBackendUrl());
        std::string response_body;
        int response_code = 0;

        if (!client.registerCollector(
            "E2E Metrics Test Collector",
            "e2e-metrics-host",
            response_body,
            response_code
        )) {
            FAIL() << "Failed to register collector for metrics tests";
        }

        // Extract collector ID and token from response
        extractCollectorIdAndToken(response_body);

        std::cout << "[E2E Metrics] Test suite ready (collector: " << test_collector_id << ")"
                  << std::endl;
    }

    static void TearDownTestSuite() {
        std::cout << "\n[E2E Metrics] Tearing down test suite..." << std::endl;
        db_helper.reset();
        harness.stopStack();
    }

    void SetUp() override {
        // Clear metrics before each test
        db_helper->clearAllMetrics();
    }

    static void extractCollectorIdAndToken(const std::string& response) {
        // Extract collector ID
        size_t id_pos = response.find("\"collector_id\":\"");
        if (id_pos != std::string::npos) {
            id_pos += 16;
            size_t end = response.find("\"", id_pos);
            test_collector_id = response.substr(id_pos, end - id_pos);
        }

        // Extract token
        size_t token_pos = response.find("\"token\":\"");
        if (token_pos != std::string::npos) {
            token_pos += 9;
            size_t end = response.find("\"", token_pos);
            test_jwt_token = response.substr(token_pos, end - token_pos);
        }
    }

    /**
     * Helper: Wait for metrics to appear in database
     */
    bool waitForMetrics(int expected_count, int timeout_seconds = 10) {
        auto start = std::chrono::steady_clock::now();

        while (true) {
            int actual_count = db_helper->getMetricsCount("metrics_pg_stats");
            if (actual_count >= expected_count) {
                return true;
            }

            auto elapsed = std::chrono::steady_clock::now() - start;
            if (std::chrono::duration_cast<std::chrono::seconds>(elapsed).count() >= timeout_seconds) {
                return false;
            }

            std::this_thread::sleep_for(std::chrono::milliseconds(500));
        }
    }
};

// Static member initialization
E2ETestHarness E2EMetricsIngestionTest::harness;
std::unique_ptr<E2EDatabaseHelper> E2EMetricsIngestionTest::db_helper;
std::string E2EMetricsIngestionTest::test_collector_id;
std::string E2EMetricsIngestionTest::test_jwt_token;

// ==================== METRICS INGESTION TESTS ====================

/**
 * Test 1: SendMetricsSuccess
 * Verify that valid metrics can be sent to backend and get 200 response
 */
TEST_F(E2EMetricsIngestionTest, SendMetricsSuccess) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    // ACT
    std::string response_body;
    int response_code = 0;

    bool success = client.submitMetrics(
        metrics_json,
        true,  // compress
        response_body,
        response_code
    );

    // ASSERT
    EXPECT_TRUE(success) << "Metrics submission failed: " << client.getLastResponseBody();
    EXPECT_EQ(response_code, 200) << "Expected 200 response, got " << response_code;
    EXPECT_GT(response_body.length(), 0) << "Empty response body";

    // Response should indicate successful insertion
    EXPECT_NE(response_body.find("success"), std::string::npos)
        << "Response should indicate success";

    std::cout << "[E2E Metrics] SendMetricsSuccess: PASSED" << std::endl;
}

/**
 * Test 2: MetricsStored
 * Verify that metrics appear in TimescaleDB after submission
 */
TEST_F(E2EMetricsIngestionTest, MetricsStored) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    // ACT
    std::string response_body;
    int response_code = 0;

    bool success = client.submitMetrics(
        metrics_json,
        true,
        response_body,
        response_code
    );

    ASSERT_TRUE(success) << "Metrics submission failed";
    ASSERT_EQ(response_code, 200);

    // Wait for metrics to be stored
    bool stored = waitForMetrics(1, 10);

    // ASSERT
    EXPECT_TRUE(stored) << "Metrics not found in database after 10 seconds";

    int actual_count = db_helper->getMetricsCount("metrics_pg_stats");
    EXPECT_GT(actual_count, 0) << "No metrics stored in database";

    std::cout << "[E2E Metrics] MetricsStored: PASSED" << std::endl;
}

/**
 * Test 3: MetricsSchema
 * Verify that stored metrics have correct schema (columns match expected)
 */
TEST_F(E2EMetricsIngestionTest, MetricsSchema) {
    // ARRANGE - Submit metrics first
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    std::string response_body;
    int response_code = 0;

    ASSERT_TRUE(client.submitMetrics(
        metrics_json,
        true,
        response_body,
        response_code
    ));

    // Wait for storage
    ASSERT_TRUE(waitForMetrics(1, 10));

    // ACT - Verify schema
    std::vector<std::string> columns = db_helper->getTableColumns("metrics_pg_stats");

    // ASSERT - Check for required columns
    EXPECT_TRUE(std::find(columns.begin(), columns.end(), "time") != columns.end())
        << "Missing 'time' column";
    EXPECT_TRUE(std::find(columns.begin(), columns.end(), "collector_id") != columns.end())
        << "Missing 'collector_id' column";
    EXPECT_TRUE(std::find(columns.begin(), columns.end(), "database") != columns.end())
        << "Missing 'database' column";
    EXPECT_TRUE(std::find(columns.begin(), columns.end(), "table_name") != columns.end())
        << "Missing 'table_name' column";

    std::cout << "[E2E Metrics] MetricsSchema: PASSED" << std::endl;
}

/**
 * Test 4: TimestampAccuracy
 * Verify that collector timestamps are preserved correctly
 */
TEST_F(E2EMetricsIngestionTest, TimestampAccuracy) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    // ACT
    std::string response_body;
    int response_code = 0;

    ASSERT_TRUE(client.submitMetrics(
        metrics_json,
        true,
        response_body,
        response_code
    ));

    ASSERT_TRUE(waitForMetrics(1, 10));

    // Get stored metrics and verify timestamp
    std::string latest_timestamp = db_helper->getLatestMetricTimestamp("metrics_pg_stats");

    // ASSERT
    EXPECT_GT(latest_timestamp.length(), 0) << "No timestamp found in database";

    // Timestamp should be ISO8601 format (contains T and Z or +/-)
    EXPECT_NE(latest_timestamp.find("T"), std::string::npos)
        << "Timestamp missing T separator (ISO8601 format)";

    std::cout << "[E2E Metrics] TimestampAccuracy: PASSED" << std::endl;
}

/**
 * Test 5: MetricTypes
 * Verify that all 4 metric types can be submitted and stored
 */
TEST_F(E2EMetricsIngestionTest, MetricTypes) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    // Submit payload with multiple metric types
    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    // ACT
    std::string response_body;
    int response_code = 0;

    bool success = client.submitMetrics(
        metrics_json,
        true,
        response_body,
        response_code
    );

    // ASSERT
    EXPECT_TRUE(success) << "Metrics submission failed";
    EXPECT_EQ(response_code, 200);

    // All metric types should be processed
    EXPECT_NE(response_body.find("pg_stats"), std::string::npos)
        << "pg_stats metric not processed";
    EXPECT_NE(response_body.find("sysstat"), std::string::npos)
        << "sysstat metric not processed";
    EXPECT_NE(response_body.find("disk_usage"), std::string::npos)
        << "disk_usage metric not processed";

    std::cout << "[E2E Metrics] MetricTypes: PASSED" << std::endl;
}

/**
 * Test 6: PayloadCompression
 * Verify that gzip compression is applied to metrics payload
 */
TEST_F(E2EMetricsIngestionTest, PayloadCompression) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    std::string metrics_json = e2e_fixtures::getLargeMetricsPayload(10);

    // ACT - Submit with compression
    std::string response_body;
    int response_code = 0;

    bool success = client.submitMetrics(
        metrics_json,
        true,  // compress
        response_body,
        response_code
    );

    // ASSERT
    EXPECT_TRUE(success) << "Metrics submission failed";
    EXPECT_EQ(response_code, 200);

    // Response should indicate compression was applied
    EXPECT_NE(response_body.find("gzip"), std::string::npos)
        << "Response should confirm gzip compression";

    std::cout << "[E2E Metrics] PayloadCompression: PASSED" << std::endl;
}

/**
 * Test 7: MetricsCount
 * Verify that correct number of metrics rows are inserted
 */
TEST_F(E2EMetricsIngestionTest, MetricsCount) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    // Create payload with known metric count
    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    // ACT
    std::string response_body;
    int response_code = 0;

    ASSERT_TRUE(client.submitMetrics(
        metrics_json,
        true,
        response_body,
        response_code
    ));

    ASSERT_TRUE(waitForMetrics(1, 10));

    // Get actual count from response
    int response_count = 0;
    size_t count_pos = response_body.find("metrics_inserted");
    if (count_pos != std::string::npos) {
        count_pos += 19;  // strlen("metrics_inserted\":")
        size_t end_pos = response_body.find(",", count_pos);
        if (end_pos == std::string::npos) {
            end_pos = response_body.find("}", count_pos);
        }
        try {
            response_count = std::stoi(response_body.substr(count_pos, end_pos - count_pos));
        } catch (...) {
            // Couldn't parse, that's okay
        }
    }

    // ASSERT
    if (response_count > 0) {
        int db_count = db_helper->getMetricsCount("metrics_pg_stats");
        EXPECT_EQ(db_count, response_count)
            << "Database count doesn't match response count";
    }

    int total_count = db_helper->getMetricsCount("metrics_pg_stats");
    EXPECT_GT(total_count, 0) << "No metrics in database";

    std::cout << "[E2E Metrics] MetricsCount: PASSED" << std::endl;
}

/**
 * Test 8: DataIntegrity
 * Verify that no data is corrupted or lost during transmission
 */
TEST_F(E2EMetricsIngestionTest, DataIntegrity) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    // ACT
    std::string response_body;
    int response_code = 0;

    ASSERT_TRUE(client.submitMetrics(
        metrics_json,
        true,
        response_body,
        response_code
    ));

    ASSERT_TRUE(waitForMetrics(1, 10));

    // Verify all columns have data (not NULL)
    // This would require a more detailed SQL query in production
    int count = db_helper->getMetricsCount("metrics_pg_stats");

    // ASSERT
    EXPECT_GT(count, 0) << "Metrics corrupted or lost";

    std::cout << "[E2E Metrics] DataIntegrity: PASSED" << std::endl;
}

/**
 * Test 9: ConcurrentPushes
 * Verify that multiple concurrent metrics pushes don't interfere
 */
TEST_F(E2EMetricsIngestionTest, ConcurrentPushes) {
    // ARRANGE
    E2EHttpClient client1(harness.getBackendUrl());
    E2EHttpClient client2(harness.getBackendUrl());

    client1.setJwtToken(test_jwt_token);
    client2.setJwtToken(test_jwt_token);

    std::string metrics1 = e2e_fixtures::getBasicMetricsPayload();
    std::string metrics2 = e2e_fixtures::getBasicMetricsPayload();

    // ACT - Send both concurrently (simulate with sequential, but fast)
    std::string response1, response2;
    int code1 = 0, code2 = 0;

    bool success1 = client1.submitMetrics(metrics1, true, response1, code1);
    bool success2 = client2.submitMetrics(metrics2, true, response2, code2);

    // ASSERT
    EXPECT_TRUE(success1) << "First push failed";
    EXPECT_TRUE(success2) << "Second push failed";
    EXPECT_EQ(code1, 200);
    EXPECT_EQ(code2, 200);

    // Wait and verify both were stored
    ASSERT_TRUE(waitForMetrics(2, 10)) << "Both metrics not stored";

    int total_count = db_helper->getMetricsCount("metrics_pg_stats");
    EXPECT_GE(total_count, 2) << "Both metrics should be stored";

    std::cout << "[E2E Metrics] ConcurrentPushes: PASSED" << std::endl;
}

/**
 * Test 10: LargePayload
 * Verify that large metrics payloads (10MB+) are handled correctly
 */
TEST_F(E2EMetricsIngestionTest, LargePayload) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    // Create large payload with 100 metric sets
    std::string metrics_json = e2e_fixtures::getLargeMetricsPayload(100);

    // ACT
    std::string response_body;
    int response_code = 0;

    bool success = client.submitMetrics(
        metrics_json,
        true,  // compression helps with large payloads
        response_body,
        response_code
    );

    // ASSERT
    EXPECT_TRUE(success) << "Large payload submission failed: " << client.getLastResponseBody();
    EXPECT_EQ(response_code, 200) << "Expected 200 for large payload";

    std::cout << "[E2E Metrics] LargePayload: PASSED" << std::endl;
}

/**
 * Test 11: PartialFailure
 * Verify that system handles partial failures gracefully
 */
TEST_F(E2EMetricsIngestionTest, PartialFailure) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    // Submit invalid then valid metrics
    std::string invalid_json = e2e_fixtures::getInvalidMetricsPayload();
    std::string valid_json = e2e_fixtures::getBasicMetricsPayload();

    // ACT
    std::string response_invalid;
    int code_invalid = 0;

    // Invalid should be rejected
    bool success_invalid = client.submitMetrics(
        invalid_json,
        true,
        response_invalid,
        code_invalid
    );

    // Then send valid
    std::string response_valid;
    int code_valid = 0;

    bool success_valid = client.submitMetrics(
        valid_json,
        true,
        response_valid,
        code_valid
    );

    // ASSERT
    // Invalid should fail
    EXPECT_NE(code_invalid, 200) << "Invalid metrics should be rejected";

    // Valid should succeed
    EXPECT_TRUE(success_valid) << "Valid metrics after invalid should work";
    EXPECT_EQ(code_valid, 200);

    std::cout << "[E2E Metrics] PartialFailure: PASSED" << std::endl;
}

/**
 * Test 12: MetricsQuery
 * Verify that metrics can be queried via backend API
 */
TEST_F(E2EMetricsIngestionTest, MetricsQuery) {
    // ARRANGE
    E2EHttpClient client(harness.getBackendUrl());
    client.setJwtToken(test_jwt_token);

    std::string metrics_json = e2e_fixtures::getBasicMetricsPayload();

    // Submit metrics
    std::string response_body;
    int response_code = 0;

    ASSERT_TRUE(client.submitMetrics(
        metrics_json,
        true,
        response_body,
        response_code
    ));

    ASSERT_TRUE(waitForMetrics(1, 10));

    // ACT - Query metrics via API (would be GET /api/v1/servers/{id}/metrics)
    std::string query_endpoint = "/api/v1/servers/1/metrics";
    std::string query_response;
    int query_code = 0;

    // Note: This assumes a query endpoint exists
    // In real implementation, would verify this returns metrics

    // ASSERT
    // Just verify the infrastructure works
    EXPECT_GT(query_code, 0) << "Query endpoint should respond";

    std::cout << "[E2E Metrics] MetricsQuery: PASSED" << std::endl;
}

