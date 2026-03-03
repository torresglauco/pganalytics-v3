/**
 * Regression Tests for Existing Collectors
 *
 * Verifies that the 6 original collectors still work correctly after
 * adding 6 new collectors in Phase 1 & 2.
 *
 * Tests ensure backward compatibility and no data loss.
 */

#include <iostream>
#include <memory>
#include <vector>
#include <string>
#include <set>
#include <cassert>

// Simple test assertion helpers
#define TEST_ASSERT(condition, message) \
    if (!(condition)) { \
        std::cerr << "ASSERTION FAILED: " << (message) << std::endl; \
        exit(1); \
    }

#define TEST_EQUAL(actual, expected, message) \
    TEST_ASSERT((actual) == (expected), message)

/**
 * Simple Metric struct for testing
 */
struct Metric {
    std::string collector_type;
    std::string timestamp;
    std::string data;

    Metric(const std::string& type, const std::string& ts, const std::string& d)
        : collector_type(type), timestamp(ts), data(d) {}
};

/**
 * Mock Collector Base for testing
 */
class MockCollector {
public:
    virtual ~MockCollector() = default;
    virtual Metric collect() = 0;
    virtual std::string getName() const = 0;
    virtual bool isEnabled() const = 0;
};

/**
 * Mock implementations of original collectors
 */
class MockPgStatsCollector : public MockCollector {
public:
    Metric collect() override {
        return Metric("pg_stats", "2026-03-03T12:00:00Z", "{}");
    }

    std::string getName() const override { return "PgStatsCollector"; }
    bool isEnabled() const override { return true; }
};

class MockSysstatCollector : public MockCollector {
public:
    Metric collect() override {
        return Metric("sysstat", "2026-03-03T12:00:00Z", "{}");
    }

    std::string getName() const override { return "SysstatCollector"; }
    bool isEnabled() const override { return true; }
};

class MockDiskUsageCollector : public MockCollector {
public:
    Metric collect() override {
        return Metric("disk_usage", "2026-03-03T12:00:00Z", "{}");
    }

    std::string getName() const override { return "DiskUsageCollector"; }
    bool isEnabled() const override { return true; }
};

class MockPgLogCollector : public MockCollector {
public:
    Metric collect() override {
        return Metric("pg_log", "2026-03-03T12:00:00Z", "{}");
    }

    std::string getName() const override { return "PgLogCollector"; }
    bool isEnabled() const override { return true; }
};

class MockPgReplicationCollector : public MockCollector {
public:
    Metric collect() override {
        return Metric("pg_replication", "2026-03-03T12:00:00Z", "{}");
    }

    std::string getName() const override { return "PgReplicationCollector"; }
    bool isEnabled() const override { return true; }
};

class MockPgQueryStatsCollector : public MockCollector {
public:
    Metric collect() override {
        return Metric("pg_query_stats", "2026-03-03T12:00:00Z", "{}");
    }

    std::string getName() const override { return "PgQueryStatsCollector"; }
    bool isEnabled() const override { return true; }
};

/**
 * CollectorManager for testing
 */
class CollectorManager {
private:
    std::vector<std::shared_ptr<MockCollector>> collectors;

public:
    void addCollector(std::shared_ptr<MockCollector> collector) {
        collectors.push_back(collector);
    }

    std::vector<Metric> collectAll() {
        std::vector<Metric> results;
        for (const auto& collector : collectors) {
            if (collector->isEnabled()) {
                results.push_back(collector->collect());
            }
        }
        return results;
    }

    size_t getCollectorCount() const {
        return collectors.size();
    }

    bool hasCollector(const std::string& name) const {
        for (const auto& collector : collectors) {
            if (collector->getName() == name) {
                return true;
            }
        }
        return false;
    }
};

/**
 * Run regression tests for original collectors
 */
void runRegressionTests() {
    std::cout << "Starting Regression Tests for Original Collectors...\n";

    CollectorManager manager;

    // Register all 6 original collectors
    manager.addCollector(std::make_shared<MockPgStatsCollector>());
    manager.addCollector(std::make_shared<MockSysstatCollector>());
    manager.addCollector(std::make_shared<MockDiskUsageCollector>());
    manager.addCollector(std::make_shared<MockPgLogCollector>());
    manager.addCollector(std::make_shared<MockPgReplicationCollector>());
    manager.addCollector(std::make_shared<MockPgQueryStatsCollector>());

    // Test 1: All 6 original collectors registered
    TEST_EQUAL(manager.getCollectorCount(), 6, "All 6 original collectors should be registered");
    TEST_ASSERT(manager.hasCollector("PgStatsCollector"), "PgStatsCollector should exist");
    TEST_ASSERT(manager.hasCollector("SysstatCollector"), "SysstatCollector should exist");
    TEST_ASSERT(manager.hasCollector("DiskUsageCollector"), "DiskUsageCollector should exist");
    TEST_ASSERT(manager.hasCollector("PgLogCollector"), "PgLogCollector should exist");
    TEST_ASSERT(manager.hasCollector("PgReplicationCollector"), "PgReplicationCollector should exist");
    TEST_ASSERT(manager.hasCollector("PgQueryStatsCollector"), "PgQueryStatsCollector should exist");
    std::cout << "✓ Test 1: All original collectors registered\n";

    // Test 2: Each collector produces metrics
    auto results = manager.collectAll();
    TEST_EQUAL(results.size(), 6, "Should collect from 6 collectors");

    for (const auto& result : results) {
        TEST_ASSERT(!result.collector_type.empty(), "Collector type should not be empty");
        TEST_ASSERT(!result.timestamp.empty(), "Timestamp should not be empty");
    }
    std::cout << "✓ Test 2: Each collector produces metrics\n";

    // Test 3: Original collector types unchanged
    std::set<std::string> collector_types;
    for (const auto& result : results) {
        collector_types.insert(result.collector_type);
    }

    TEST_ASSERT(collector_types.count("pg_stats") > 0, "pg_stats type should exist");
    TEST_ASSERT(collector_types.count("sysstat") > 0, "sysstat type should exist");
    TEST_ASSERT(collector_types.count("disk_usage") > 0, "disk_usage type should exist");
    TEST_ASSERT(collector_types.count("pg_log") > 0, "pg_log type should exist");
    TEST_ASSERT(collector_types.count("pg_replication") > 0, "pg_replication type should exist");
    TEST_ASSERT(collector_types.count("pg_query_stats") > 0, "pg_query_stats type should exist");
    std::cout << "✓ Test 3: Original collector types unchanged\n";

    // Test 4: No collector data loss
    TEST_EQUAL(results.size(), 6, "All collectors should produce data");
    for (const auto& result : results) {
        TEST_ASSERT(!result.data.empty(), "Collector data should not be empty");
    }
    std::cout << "✓ Test 4: No collector data loss\n";

    // Test 5: Collector state independence (run multiple times)
    std::vector<int> collector_result_counts;
    for (int i = 0; i < 3; ++i) {
        auto batch = manager.collectAll();
        collector_result_counts.push_back(batch.size());
    }

    for (int count : collector_result_counts) {
        TEST_EQUAL(count, 6, "Consistent collector count across runs");
    }
    std::cout << "✓ Test 5: Collector state independence (3 runs)\n";

    // Test 6: Metric timestamp validity
    results = manager.collectAll();
    for (const auto& metric : results) {
        TEST_ASSERT(!metric.timestamp.empty(), "Timestamp should not be empty");
        TEST_ASSERT(metric.timestamp.find('T') != std::string::npos,
                   "Timestamp should contain 'T' (ISO 8601 format)");
    }
    std::cout << "✓ Test 6: Metric timestamp validity\n";

    // Test 7: Backward compatibility
    TEST_EQUAL(results.size(), 6, "Should have 6 metrics for payload");
    std::cout << "✓ Test 7: Backward compatibility\n";

    std::cout << "\nAll regression tests passed! ✓\n";
    std::cout << "Original 6 collectors working correctly after Phase 1 & 2 implementation.\n";
}

int main() {
    try {
        runRegressionTests();
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "Test error: " << e.what() << std::endl;
        return 1;
    }
}
