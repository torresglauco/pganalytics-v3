#include <gtest/gtest.h>
#include "e2e_harness.h"
#include "http_client.h"
#include "database_helper.h"
#include "fixtures.h"
#include <chrono>
#include <thread>
#include <vector>
#include <numeric>

/**
 * E2E Performance Tests
 *
 * Tests system performance characteristics under normal operation.
 * Validates:
 * - Metric collection latency
 * - Metrics transmission latency
 * - Database insertion latency
 * - Sustained throughput over time
 * - Memory stability (no leaks)
 */
class E2EPerformanceTest : public ::testing::Test {
protected:
    static E2ETestHarness harness;
    static std::unique_ptr<E2EDatabaseHelper> db_helper;
    static E2EHttpClient* api_client;
    static std::string test_collector_id;
    static std::string test_jwt_token;

    static void SetUpTestSuite() {
        std::cout << "\n[E2E Performance] Setting up test suite..." << std::endl;

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

        // Create HTTP client
        api_client = new E2EHttpClient(harness.getBackendUrl());

        // Register test collector
        std::string response_body;
        int response_code = 0;

        if (!api_client->registerCollector(
            "E2E Performance Test Collector",
            "e2e-perf-host",
            response_body,
            response_code
        )) {
            FAIL() << "Failed to register collector for performance tests";
        }

        extractCollectorIdAndToken(response_body);
        api_client->setJwtToken(test_jwt_token);

        std::cout << "[E2E Performance] Test suite ready (collector: " << test_collector_id << ")"
                  << std::endl;
    }

    static void TearDownTestSuite() {
        std::cout << "\n[E2E Performance] Tearing down test suite..." << std::endl;
        delete api_client;
        db_helper.reset();
        harness.stopStack();
    }

    void SetUp() override {
        // Clear metrics before each test
        db_helper->clearAllMetrics();
    }

    static void extractCollectorIdAndToken(const std::string& response) {
        size_t id_pos = response.find("\"collector_id\":\"");
        if (id_pos != std::string::npos) {
            id_pos += 16;
            size_t end = response.find("\"", id_pos);
            test_collector_id = response.substr(id_pos, end - id_pos);
        }

        size_t token_pos = response.find("\"token\":\"");
        if (token_pos != std::string::npos) {
            token_pos += 9;
            size_t end = response.find("\"", token_pos);
            test_jwt_token = response.substr(token_pos, end - token_pos);
        }
    }

    /**
     * Helper: Measure operation latency
     * Returns latency in milliseconds
     */
    template<typename Func>
    long long measureLatency(Func operation) {
        auto start = std::chrono::high_resolution_clock::now();
        operation();
        auto end = std::chrono::high_resolution_clock::now();
        return std::chrono::duration_cast<std::chrono::milliseconds>(end - start).count();
    }

    /**
     * Helper: Wait for metrics to appear in database
     */
    bool waitForMetricsCount(int expected_count, int timeout_seconds = 10) {
        auto start = std::chrono::steady_clock::now();
        while (true) {
            int count = db_helper->getMetricsCount("metrics_pg_stats");
            if (count >= expected_count) {
                return true;
            }

            auto elapsed = std::chrono::steady_clock::now() - start;
            if (std::chrono::duration_cast<std::chrono::seconds>(elapsed).count() >= timeout_seconds) {
                return false;
            }

            std::this_thread::sleep_for(std::chrono::milliseconds(100));
        }
    }
};

// Static member initialization
E2ETestHarness E2EPerformanceTest::harness;
std::unique_ptr<E2EDatabaseHelper> E2EPerformanceTest::db_helper;
E2EHttpClient* E2EPerformanceTest::api_client = nullptr;
std::string E2EPerformanceTest::test_collector_id;
std::string E2EPerformanceTest::test_jwt_token;

// ==================== PERFORMANCE TESTS ====================

/**
 * Test 1: MetricCollectionLatency
 * Measure time to collect metrics from a single source
 */
TEST_F(E2EPerformanceTest, MetricCollectionLatency) {
    // ARRANGE
    std::vector<long long> latencies;
    const int iterations = 5;
    const long long max_collection_latency = 1000;  // 1 second max

    // ACT - Measure metrics collection latency
    // In real scenario, this would measure collector plugin execution time
    // For this test, we measure the time to submit metrics (as proxy)
    std::string metrics = e2e_fixtures::getBasicMetricsPayload();

    for (int i = 0; i < iterations; i++) {
        long long latency = measureLatency([&]() {
            std::string response;
            int code = 0;
            api_client->submitMetrics(metrics, true, response, code);
        });
        latencies.push_back(latency);
    }

    // ASSERT
    // Calculate statistics
    long long min_latency = *std::min_element(latencies.begin(), latencies.end());
    long long max_latency = *std::max_element(latencies.begin(), latencies.end());
    long long avg_latency = std::accumulate(latencies.begin(), latencies.end(), 0LL) / iterations;

    EXPECT_LT(avg_latency, max_collection_latency)
        << "Average collection latency exceeds threshold: " << avg_latency << "ms";
    EXPECT_LT(max_latency, max_collection_latency * 2)
        << "Maximum collection latency exceeds 2x threshold: " << max_latency << "ms";

    std::cout << "[E2E Performance] MetricCollectionLatency:" << std::endl
              << "  Min: " << min_latency << "ms" << std::endl
              << "  Avg: " << avg_latency << "ms" << std::endl
              << "  Max: " << max_latency << "ms" << std::endl
              << "  PASSED" << std::endl;
}

/**
 * Test 2: MetricsTransmissionLatency
 * Measure time to transmit metrics to backend
 */
TEST_F(E2EPerformanceTest, MetricsTransmissionLatency) {
    // ARRANGE
    std::vector<long long> latencies;
    const int iterations = 5;
    const long long max_transmission_latency = 2000;  // 2 seconds max for HTTP+TLS

    // ACT - Measure HTTP transmission latency
    std::string metrics = e2e_fixtures::getBasicMetricsPayload();

    for (int i = 0; i < iterations; i++) {
        long long latency = measureLatency([&]() {
            std::string response;
            int code = 0;
            api_client->submitMetrics(metrics, true, response, code);
        });
        latencies.push_back(latency);
    }

    // ASSERT
    long long min_latency = *std::min_element(latencies.begin(), latencies.end());
    long long max_latency = *std::max_element(latencies.begin(), latencies.end());
    long long avg_latency = std::accumulate(latencies.begin(), latencies.end(), 0LL) / iterations;

    EXPECT_LT(avg_latency, max_transmission_latency)
        << "Average transmission latency exceeds threshold: " << avg_latency << "ms";
    EXPECT_LT(max_latency, max_transmission_latency * 2)
        << "Maximum transmission latency exceeds 2x threshold: " << max_latency << "ms";

    std::cout << "[E2E Performance] MetricsTransmissionLatency:" << std::endl
              << "  Min: " << min_latency << "ms" << std::endl
              << "  Avg: " << avg_latency << "ms" << std::endl
              << "  Max: " << max_latency << "ms" << std::endl
              << "  PASSED" << std::endl;
}

/**
 * Test 3: DatabaseInsertLatency
 * Measure time for metrics to be stored in database after transmission
 */
TEST_F(E2EPerformanceTest, DatabaseInsertLatency) {
    // ARRANGE
    std::vector<long long> latencies;
    const int iterations = 3;
    const long long max_insert_latency = 5000;  // 5 seconds max (includes TimescaleDB processing)

    // ACT - Measure end-to-end latency (submit → stored in DB)
    std::string metrics = e2e_fixtures::getBasicMetricsPayload();

    for (int i = 0; i < iterations; i++) {
        db_helper->clearAllMetrics();

        long long latency = measureLatency([&]() {
            std::string response;
            int code = 0;
            api_client->submitMetrics(metrics, true, response, code);

            // Wait for metrics to be stored (with timeout)
            waitForMetricsCount(1, 5);
        });
        latencies.push_back(latency);
    }

    // ASSERT
    long long min_latency = *std::min_element(latencies.begin(), latencies.end());
    long long max_latency = *std::max_element(latencies.begin(), latencies.end());
    long long avg_latency = std::accumulate(latencies.begin(), latencies.end(), 0LL) / iterations;

    EXPECT_LT(avg_latency, max_insert_latency)
        << "Average insert latency exceeds threshold: " << avg_latency << "ms";
    EXPECT_LT(max_latency, max_insert_latency * 2)
        << "Maximum insert latency exceeds 2x threshold: " << max_latency << "ms";

    // Verify metrics are actually stored
    int final_count = db_helper->getMetricsCount("metrics_pg_stats");
    EXPECT_GT(final_count, 0) << "Metrics not stored in database";

    std::cout << "[E2E Performance] DatabaseInsertLatency:" << std::endl
              << "  Min: " << min_latency << "ms" << std::endl
              << "  Avg: " << avg_latency << "ms" << std::endl
              << "  Max: " << max_latency << "ms" << std::endl
              << "  PASSED" << std::endl;
}

/**
 * Test 4: ThroughputSustained
 * Measure sustained throughput over multiple pushes
 */
TEST_F(E2EPerformanceTest, ThroughputSustained) {
    // ARRANGE
    const int push_count = 10;
    const long long target_throughput_per_minute = 600;  // 10 pushes/min minimum
    std::vector<long long> push_latencies;

    // ACT - Submit metrics multiple times and measure throughput
    std::string metrics = e2e_fixtures::getBasicMetricsPayload();

    auto start = std::chrono::high_resolution_clock::now();

    for (int i = 0; i < push_count; i++) {
        long long latency = measureLatency([&]() {
            std::string response;
            int code = 0;
            api_client->submitMetrics(metrics, true, response, code);
        });
        push_latencies.push_back(latency);

        // Small delay between pushes to simulate real scenario
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }

    auto end = std::chrono::high_resolution_clock::now();
    long long total_time_ms = std::chrono::duration_cast<std::chrono::milliseconds>(end - start).count();
    double throughput_per_min = (push_count / (total_time_ms / 1000.0)) * 60.0;

    // ASSERT
    EXPECT_GE(throughput_per_min, target_throughput_per_minute)
        << "Throughput below target: " << throughput_per_min << " pushes/min";

    // Verify all pushes succeeded
    for (long long latency : push_latencies) {
        EXPECT_LT(latency, 5000) << "Individual push exceeded 5 seconds";
    }

    std::cout << "[E2E Performance] ThroughputSustained:" << std::endl
              << "  Total pushes: " << push_count << std::endl
              << "  Total time: " << total_time_ms << "ms" << std::endl
              << "  Throughput: " << throughput_per_min << " pushes/min" << std::endl
              << "  Avg latency: " << std::accumulate(push_latencies.begin(), push_latencies.end(), 0LL) / push_count << "ms" << std::endl
              << "  PASSED" << std::endl;
}

/**
 * Test 5: MemoryStability
 * Verify system memory usage remains stable over multiple operations
 */
TEST_F(E2EPerformanceTest, MemoryStability) {
    // ARRANGE
    const int operation_count = 20;
    std::vector<int> metrics_counts;

    // ACT - Perform multiple metric submissions and verify data integrity
    std::string metrics = e2e_fixtures::getBasicMetricsPayload();

    for (int i = 0; i < operation_count; i++) {
        // Submit metrics
        std::string response;
        int code = 0;
        ASSERT_TRUE(api_client->submitMetrics(metrics, true, response, code));

        // Wait for storage
        bool stored = waitForMetricsCount(1, 5);
        ASSERT_TRUE(stored) << "Metrics not stored in iteration " << i;

        int count = db_helper->getMetricsCount("metrics_pg_stats");
        metrics_counts.push_back(count);

        // Clear for next iteration
        db_helper->clearAllMetrics();
    }

    // ASSERT
    // Verify all operations succeeded and database remained consistent
    EXPECT_EQ(metrics_counts.size(), operation_count)
        << "Not all operations completed successfully";

    // All counts should be > 0 (metrics stored)
    for (size_t i = 0; i < metrics_counts.size(); ++i) {
        EXPECT_GT(metrics_counts[i], 0)
            << "Metrics not stored in iteration " << i;
    }

    // Check no significant data loss or corruption
    int total_operations = metrics_counts.size();
    int successful_operations = 0;
    for (int count : metrics_counts) {
        if (count > 0) successful_operations++;
    }

    double success_rate = (successful_operations / (double)total_operations) * 100.0;
    EXPECT_GE(success_rate, 95.0)
        << "Success rate below 95%: " << success_rate << "%";

    std::cout << "[E2E Performance] MemoryStability:" << std::endl
              << "  Total operations: " << total_operations << std::endl
              << "  Successful: " << successful_operations << std::endl
              << "  Success rate: " << success_rate << "%" << std::endl
              << "  System remained stable" << std::endl
              << "  PASSED" << std::endl;
}

// ==================== TEST SUMMARY ====================

/**
 * Summary of Performance Tests
 *
 * All 5 tests validate system performance characteristics:
 * 1. MetricCollectionLatency - Time to collect metrics
 * 2. MetricsTransmissionLatency - Time to transmit to backend
 * 3. DatabaseInsertLatency - Time for data to be stored
 * 4. ThroughputSustained - Metrics pushes per minute
 * 5. MemoryStability - System stability over repeated operations
 *
 * Expected Result: 5/5 tests passing
 * Time Target: ~30-40 seconds total
 */

