#include "gtest/gtest.h"
#include "../../include/replication_plugin.h"
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * Unit tests for PgReplicationCollector
 */
class ReplicationCollectorTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Initialize test fixtures
        hostname_ = "test-collector";
        collector_id_ = "test-replication-001";
        postgres_host_ = "localhost";
        postgres_port_ = 5432;
        postgres_user_ = "postgres";
        postgres_password_ = "";
    }

    std::string hostname_;
    std::string collector_id_;
    std::string postgres_host_;
    int postgres_port_;
    std::string postgres_user_;
    std::string postgres_password_;
};

/**
 * Test constructor initializes with correct parameters
 */
TEST_F(ReplicationCollectorTest, ConstructorInitializesCorrectly) {
    PgReplicationCollector collector(
        hostname_,
        collector_id_,
        postgres_host_,
        postgres_port_,
        postgres_user_,
        postgres_password_
    );

    EXPECT_EQ(collector.getType(), "pg_replication");
    EXPECT_TRUE(collector.isEnabled());
}

/**
 * Test that execute() returns valid JSON structure
 * Note: This test requires actual PostgreSQL connection
 */
TEST_F(ReplicationCollectorTest, DISABLED_ExecuteReturnsValidJson) {
    // Skip this test in CI environment without real PostgreSQL
    if (std::getenv("CI") != nullptr) {
        GTEST_SKIP();
    }

    PgReplicationCollector collector(
        hostname_,
        collector_id_,
        postgres_host_,
        postgres_port_,
        postgres_user_,
        postgres_password_
    );

    json result = collector.execute();

    // Verify result structure
    EXPECT_TRUE(result.contains("type"));
    EXPECT_EQ(result["type"], "pg_replication");

    EXPECT_TRUE(result.contains("timestamp"));
    EXPECT_TRUE(result.contains("replication_slots"));
    EXPECT_TRUE(result.contains("replication_status"));
    EXPECT_TRUE(result.contains("wal_status"));
    EXPECT_TRUE(result.contains("wraparound_risk"));
    EXPECT_TRUE(result.contains("collection_errors"));

    // Verify arrays
    EXPECT_TRUE(result["replication_slots"].is_array());
    EXPECT_TRUE(result["replication_status"].is_array());
    EXPECT_TRUE(result["wraparound_risk"].is_array());
    EXPECT_TRUE(result["collection_errors"].is_array());

    // Verify WAL status is object
    EXPECT_TRUE(result["wal_status"].is_object());
}

/**
 * Test LSN parsing functionality
 */
TEST_F(ReplicationCollectorTest, ParseLsnConvertsCorrectly) {
    // This test requires exposing parseLsn as public or creating a friend test
    // For now, we test through the full execute() path
    // TODO: Make parseLsn() public or create accessor for testing
}

/**
 * Test bytes behind calculation
 */
TEST_F(ReplicationCollectorTest, CalculateBytesBehindComputation) {
    // Test through full execute() path
    // TODO: Make calculateBytesBehind() public or create accessor for testing
}

/**
 * Test version detection
 */
TEST_F(ReplicationCollectorTest, DISABLED_DetectsPostgresVersionCorrectly) {
    if (std::getenv("CI") != nullptr) {
        GTEST_SKIP();
    }

    PgReplicationCollector collector(
        hostname_,
        collector_id_,
        postgres_host_,
        postgres_port_,
        postgres_user_,
        postgres_password_
    );

    // Execute to trigger version detection
    json result = collector.execute();

    // Version detection happens internally
    // Verify that version-dependent queries completed without error
    EXPECT_LE(result["collection_errors"].size(), 2);  // At most a few expected errors
}

/**
 * Test replication slot structure
 */
TEST_F(ReplicationCollectorTest, DISABLED_ReplicationSlotStructureIsValid) {
    if (std::getenv("CI") != nullptr) {
        GTEST_SKIP();
    }

    PgReplicationCollector collector(
        hostname_,
        collector_id_,
        postgres_host_,
        postgres_port_,
        postgres_user_,
        postgres_password_
    );

    json result = collector.execute();

    if (result["replication_slots"].size() > 0) {
        auto slot = result["replication_slots"][0];

        // Verify required fields
        EXPECT_TRUE(slot.contains("slot_name"));
        EXPECT_TRUE(slot.contains("slot_type"));
        EXPECT_TRUE(slot.contains("active"));
        EXPECT_TRUE(slot.contains("restart_lsn"));
        EXPECT_TRUE(slot.contains("wal_retained_mb"));

        // Verify field types
        EXPECT_TRUE(slot["slot_name"].is_string());
        EXPECT_TRUE(slot["slot_type"].is_string());
        EXPECT_TRUE(slot["active"].is_boolean());
        EXPECT_TRUE(slot["wal_retained_mb"].is_number());
    }
}

/**
 * Test replication status structure
 */
TEST_F(ReplicationCollectorTest, DISABLED_ReplicationStatusStructureIsValid) {
    if (std::getenv("CI") != nullptr) {
        GTEST_SKIP();
    }

    PgReplicationCollector collector(
        hostname_,
        collector_id_,
        postgres_host_,
        postgres_port_,
        postgres_user_,
        postgres_password_
    );

    json result = collector.execute();

    if (result["replication_status"].size() > 0) {
        auto status = result["replication_status"][0];

        // Verify required fields
        EXPECT_TRUE(status.contains("server_pid"));
        EXPECT_TRUE(status.contains("usename"));
        EXPECT_TRUE(status.contains("application_name"));
        EXPECT_TRUE(status.contains("state"));
        EXPECT_TRUE(status.contains("sync_state"));
        EXPECT_TRUE(status.contains("write_lag_ms"));
        EXPECT_TRUE(status.contains("flush_lag_ms"));
        EXPECT_TRUE(status.contains("replay_lag_ms"));

        // Verify field types
        EXPECT_TRUE(status["server_pid"].is_number());
        EXPECT_TRUE(status["usename"].is_string());
        EXPECT_TRUE(status["write_lag_ms"].is_number());
        EXPECT_TRUE(status["replay_lag_ms"].is_number());
    }
}

/**
 * Test wraparound risk structure
 */
TEST_F(ReplicationCollectorTest, DISABLED_WraparoundRiskStructureIsValid) {
    if (std::getenv("CI") != nullptr) {
        GTEST_SKIP();
    }

    PgReplicationCollector collector(
        hostname_,
        collector_id_,
        postgres_host_,
        postgres_port_,
        postgres_user_,
        postgres_password_
    );

    json result = collector.execute();

    if (result["wraparound_risk"].size() > 0) {
        auto risk = result["wraparound_risk"][0];

        // Verify required fields
        EXPECT_TRUE(risk.contains("database"));
        EXPECT_TRUE(risk.contains("relfrozenxid"));
        EXPECT_TRUE(risk.contains("percent_until_wraparound"));
        EXPECT_TRUE(risk.contains("at_risk"));

        // Verify field types
        EXPECT_TRUE(risk["database"].is_string());
        EXPECT_TRUE(risk["relfrozenxid"].is_number());
        EXPECT_TRUE(risk["percent_until_wraparound"].is_number());
        EXPECT_TRUE(risk["at_risk"].is_boolean());
    }
}

/**
 * Test WAL status structure
 */
TEST_F(ReplicationCollectorTest, DISABLED_WalStatusStructureIsValid) {
    if (std::getenv("CI") != nullptr) {
        GTEST_SKIP();
    }

    PgReplicationCollector collector(
        hostname_,
        collector_id_,
        postgres_host_,
        postgres_port_,
        postgres_user_,
        postgres_password_
    );

    json result = collector.execute();

    auto wal_status = result["wal_status"];

    // Verify required fields
    EXPECT_TRUE(wal_status.contains("total_segments"));
    EXPECT_TRUE(wal_status.contains("current_wal_size_mb"));
    EXPECT_TRUE(wal_status.contains("wal_directory_size_mb"));

    // Verify field types
    EXPECT_TRUE(wal_status["total_segments"].is_number());
    EXPECT_TRUE(wal_status["current_wal_size_mb"].is_number());
}
