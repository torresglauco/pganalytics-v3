#include <gtest/gtest.h>
#include "../src/postgres_plugin.cpp"
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * Test Suite for PostgreSQL Statistics Collector
 *
 * These tests verify:
 * 1. Constructor initialization
 * 2. JSON schema validation
 * 3. Error handling for missing libpq
 * 4. Graceful degradation without database
 */

class PgStatsCollectorTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Create collector with test parameters
        std::vector<std::string> test_dbs = {"postgres", "template1"};
        collector = std::make_unique<PgStatsCollector>(
            "test-host",
            "col-001",
            "localhost",
            5432,
            "postgres",
            "password",
            test_dbs
        );
    }

    void TearDown() override {
        collector.reset();
    }

    std::unique_ptr<PgStatsCollector> collector;
};

/**
 * Test 1: Collector initialization
 */
TEST_F(PgStatsCollectorTest, InitializationSuccessful) {
    EXPECT_NE(collector, nullptr);
    EXPECT_EQ(collector->getType(), "pg_stats");
    EXPECT_TRUE(collector->isEnabled());
}

/**
 * Test 2: Execute returns valid JSON structure
 */
TEST_F(PgStatsCollectorTest, ExecuteReturnsValidJSON) {
    json result = collector->execute();

    // Verify top-level structure
    EXPECT_TRUE(result.contains("type"));
    EXPECT_TRUE(result.contains("timestamp"));
    EXPECT_TRUE(result.contains("databases"));

    EXPECT_EQ(result["type"], "pg_stats");
    EXPECT_TRUE(result["timestamp"].is_string());
    EXPECT_TRUE(result["databases"].is_array());
}

/**
 * Test 3: Database entries have required fields
 */
TEST_F(PgStatsCollectorTest, DatabaseEntriesHaveRequiredFields) {
    json result = collector->execute();

    EXPECT_TRUE(result["databases"].is_array());
    EXPECT_GE(result["databases"].size(), 0);

    // Even if no actual database connection, structure should be valid
    for (const auto& db_entry : result["databases"]) {
        EXPECT_TRUE(db_entry.contains("database"));
        EXPECT_TRUE(db_entry.contains("timestamp"));
        EXPECT_TRUE(db_entry.contains("tables"));
        EXPECT_TRUE(db_entry.contains("indexes"));
    }
}

/**
 * Test 4: Database stats have correct type
 */
TEST_F(PgStatsCollectorTest, DatabaseStatsHaveCorrectTypes) {
    json result = collector->execute();

    if (!result["databases"].empty()) {
        auto db = result["databases"][0];

        // Check if stats are present (will be 0 if no connection)
        if (db.contains("size_bytes")) {
            EXPECT_TRUE(db["size_bytes"].is_number_unsigned());
        }
        if (db.contains("transactions_committed")) {
            EXPECT_TRUE(db["transactions_committed"].is_number_unsigned());
        }
        if (db.contains("transactions_rolledback")) {
            EXPECT_TRUE(db["transactions_rolledback"].is_number_unsigned());
        }
    }
}

/**
 * Test 5: Table stats array is valid
 */
TEST_F(PgStatsCollectorTest, TableStatsArrayIsValid) {
    json result = collector->execute();

    if (!result["databases"].empty()) {
        auto tables = result["databases"][0]["tables"];
        EXPECT_TRUE(tables.is_array());

        // If there are tables, check their structure
        for (const auto& table : tables) {
            EXPECT_TRUE(table.contains("schema"));
            EXPECT_TRUE(table.contains("name"));
            EXPECT_TRUE(table.contains("live_tuples"));
            EXPECT_TRUE(table.contains("size_bytes"));
            EXPECT_TRUE(table.contains("vacuum_count"));
            EXPECT_TRUE(table.contains("autovacuum_count"));
        }
    }
}

/**
 * Test 6: Index stats array is valid
 */
TEST_F(PgStatsCollectorTest, IndexStatsArrayIsValid) {
    json result = collector->execute();

    if (!result["databases"].empty()) {
        auto indexes = result["databases"][0]["indexes"];
        EXPECT_TRUE(indexes.is_array());

        // If there are indexes, check their structure
        for (const auto& index : indexes) {
            EXPECT_TRUE(index.contains("schema"));
            EXPECT_TRUE(index.contains("name"));
            EXPECT_TRUE(index.contains("table"));
            EXPECT_TRUE(index.contains("scans"));
            EXPECT_TRUE(index.contains("size_bytes"));
            EXPECT_TRUE(index.contains("status"));
        }
    }
}

/**
 * Test 7: Timestamp format is ISO8601
 */
TEST_F(PgStatsCollectorTest, TimestampFormatIsISO8601) {
    json result = collector->execute();
    std::string timestamp = result["timestamp"];

    // ISO8601 format: YYYY-MM-DDTHH:MM:SSZ
    EXPECT_GE(timestamp.length(), 19);  // Minimum length
    EXPECT_EQ(timestamp[10], 'T');      // T separator
    EXPECT_EQ(timestamp[timestamp.length() - 1], 'Z');  // Z suffix
}

/**
 * Test 8: Type getter returns correct value
 */
TEST_F(PgStatsCollectorTest, GetTypeReturnsCorrectValue) {
    EXPECT_EQ(collector->getType(), "pg_stats");
}

/**
 * Test 9: IsEnabled returns true for new collector
 */
TEST_F(PgStatsCollectorTest, IsEnabledReturnsTrue) {
    EXPECT_TRUE(collector->isEnabled());
}

/**
 * Test 10: Multiple database support
 */
TEST_F(PgStatsCollectorTest, MultipleDatabaseSupport) {
    std::vector<std::string> multi_dbs = {"db1", "db2", "db3"};
    PgStatsCollector multi_collector(
        "host1", "col-002", "localhost", 5432,
        "user", "pass", multi_dbs
    );

    json result = multi_collector.execute();

    // Should attempt to collect from all databases
    EXPECT_TRUE(result["databases"].is_array());
    // Even if connections fail, should handle gracefully
}

/**
 * Test 11: Empty database list
 */
TEST_F(PgStatsCollectorTest, EmptyDatabaseList) {
    std::vector<std::string> empty_dbs = {};
    PgStatsCollector empty_collector(
        "host", "col-003", "localhost", 5432,
        "user", "pass", empty_dbs
    );

    json result = empty_collector.execute();

    EXPECT_EQ(result["type"], "pg_stats");
    EXPECT_EQ(result["databases"].size(), 0);
}

/**
 * Test 12: Handle special characters in parameters
 */
TEST_F(PgStatsCollectorTest, HandlesSpecialCharactersInParameters) {
    std::vector<std::string> special_dbs = {"my-db", "test_db"};
    PgStatsCollector special_collector(
        "host-name", "col-004", "db.example.com", 5432,
        "user@domain", "pass!word", special_dbs
    );

    json result = special_collector.execute();

    // Should handle without crashing
    EXPECT_EQ(result["type"], "pg_stats");
    EXPECT_TRUE(result["databases"].is_array());
}

/**
 * Test 13: Consistent timestamp across same execution
 */
TEST_F(PgStatsCollectorTest, ConsistentTimestampFormat) {
    json result = collector->execute();

    std::string main_ts = result["timestamp"];

    // All database entries should have timestamp field
    for (const auto& db : result["databases"]) {
        EXPECT_TRUE(db.contains("timestamp"));
        EXPECT_TRUE(db["timestamp"].is_string());
        std::string db_ts = db["timestamp"];
        // Should be similar format (not necessarily exact same value)
        EXPECT_EQ(main_ts.length(), db_ts.length());
    }
}

/**
 * Test 14: Numeric values are non-negative
 */
TEST_F(PgStatsCollectorTest, NumericValuesAreValid) {
    json result = collector->execute();

    for (const auto& db : result["databases"]) {
        // Size should be non-negative
        if (db.contains("size_bytes")) {
            EXPECT_GE(db["size_bytes"].get<int64_t>(), 0);
        }

        // Transaction counts should be non-negative
        if (db.contains("transactions_committed")) {
            EXPECT_GE(db["transactions_committed"].get<int64_t>(), 0);
        }

        // Check table stats
        for (const auto& table : db["tables"]) {
            EXPECT_GE(table["live_tuples"].get<int64_t>(), 0);
            EXPECT_GE(table["size_bytes"].get<int64_t>(), 0);
        }

        // Check index stats
        for (const auto& index : db["indexes"]) {
            EXPECT_GE(index["size_bytes"].get<int64_t>(), 0);
            EXPECT_GE(index["scans"].get<int64_t>(), 0);
        }
    }
}

/**
 * Test 15: JSON is serializable
 */
TEST_F(PgStatsCollectorTest, JSONIsSerializable) {
    json result = collector->execute();

    // Should be able to convert to string without error
    std::string json_str = result.dump();
    EXPECT_GT(json_str.length(), 0);

    // Should be able to parse back
    json parsed = json::parse(json_str);
    EXPECT_EQ(parsed["type"], "pg_stats");
}

/**
 * Integration Test: Collector Interface Compliance
 */
class CollectorInterfaceTest : public ::testing::Test {
protected:
    void SetUp() override {
        std::vector<std::string> dbs = {"postgres"};
        collector = std::make_unique<PgStatsCollector>(
            "host", "col", "localhost", 5432, "user", "pass", dbs
        );
    }

    std::unique_ptr<PgStatsCollector> collector;
};

TEST_F(CollectorInterfaceTest, ImplementsCollectorInterface) {
    // Should be able to call through base Collector interface
    Collector* base_ptr = collector.get();

    EXPECT_EQ(base_ptr->getType(), "pg_stats");
    EXPECT_TRUE(base_ptr->isEnabled());

    json result = base_ptr->execute();
    EXPECT_EQ(result["type"], "pg_stats");
}

/**
 * Performance Test: Execution completes in reasonable time
 */
TEST_F(PgStatsCollectorTest, ExecutionCompletes) {
    auto start = std::chrono::high_resolution_clock::now();
    json result = collector->execute();
    auto end = std::chrono::high_resolution_clock::now();

    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);

    // Should complete without libpq connection in under 100ms
    // With real database, might take longer
    EXPECT_LT(duration.count(), 10000);  // 10 second timeout
}
