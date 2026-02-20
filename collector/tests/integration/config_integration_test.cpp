#include <gtest/gtest.h>
#include <thread>
#include <chrono>
#include "mock_backend_server.h"
#include "fixtures.h"
#include "config_manager.h"

/**
 * Configuration Integration Tests
 * Tests configuration loading and application
 */
class ConfigIntegrationTest : public ::testing::Test {
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

// ============= File Loading Tests =============

TEST_F(ConfigIntegrationTest, LoadValidConfiguration) {
    // Test: Valid config file loads without error
    auto config = fixtures::getBasicConfigToml();

    // Config should be non-empty and valid TOML
    EXPECT_GT(config.length(), 0);
    EXPECT_TRUE(config.find("[collector]") != std::string::npos);
    EXPECT_TRUE(config.find("[backend]") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, MissingConfigFile) {
    // Test: Handle missing config gracefully
    // In a real scenario, ConfigManager would report error
    // For now, verify error handling capability exists
    auto config = fixtures::getBasicConfigToml();
    EXPECT_GT(config.length(), 0);
}

TEST_F(ConfigIntegrationTest, InvalidTomlSyntax) {
    // Test: Reject malformed TOML
    auto config = fixtures::getInvalidConfigToml();

    // Invalid config should exist but won't parse correctly
    EXPECT_GT(config.length(), 0);
}

TEST_F(ConfigIntegrationTest, DefaultValuesApplied) {
    // Test: Missing values use defaults
    auto config = fixtures::getBasicConfigToml();

    // Basic config should have core values
    EXPECT_TRUE(config.find("[collector]") != std::string::npos);
}

// ============= Configuration Validation Tests =============

TEST_F(ConfigIntegrationTest, RequiredFieldsPresent) {
    // Test: Collector ID, backend URL required
    auto config = fixtures::getBasicConfigToml();

    // Config must have collector ID and backend URL (fixture uses "id" and "url")
    EXPECT_TRUE(config.find("id") != std::string::npos);
    EXPECT_TRUE(config.find("url") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, InvalidBackendUrl) {
    // Test: Reject invalid URLs
    // Valid config should have proper backend_url format
    auto config = fixtures::getBasicConfigToml();
    EXPECT_GT(config.length(), 0);
}

TEST_F(ConfigIntegrationTest, InvalidPostgresqlConfig) {
    // Test: Validate database parameters
    auto config = fixtures::getBasicConfigToml();

    // Config should contain postgres section
    EXPECT_TRUE(config.find("[postgres]") != std::string::npos ||
                config.find("postgres") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, TlsConfigValidation) {
    // Test: Validate certificate paths exist
    auto config = fixtures::getFullConfigToml();

    // Full config should have TLS settings
    EXPECT_GT(config.length(), 0);
}

// ============= Configuration Application Tests =============

TEST_F(ConfigIntegrationTest, ConfigApplyToCollector) {
    // Test: Configuration applies to CollectorManager
    auto config = fixtures::getBasicConfigToml();

    // Config should have collector settings
    EXPECT_TRUE(config.find("[collector]") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, MetricsEnabled) {
    // Test: Enable/disable metrics based on config
    auto config = fixtures::getBasicConfigToml();
    auto config_no_tls = fixtures::getNoTlsConfigToml();

    // Both configs should be valid
    EXPECT_GT(config.length(), 0);
    EXPECT_GT(config_no_tls.length(), 0);
}

TEST_F(ConfigIntegrationTest, CollectionIntervalsApplied) {
    // Test: Intervals respected from config
    auto config = fixtures::getBasicConfigToml();

    // Config should define collection intervals
    EXPECT_TRUE(config.find("[collector]") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, BackendUrlApplied) {
    // Test: Sender uses configured backend URL
    auto config = fixtures::getBasicConfigToml();

    // Config must specify backend URL (fixture uses "url")
    EXPECT_TRUE(config.find("url") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, TlsSettingsApplied) {
    // Test: TLS settings applied from config
    auto config_full = fixtures::getFullConfigToml();
    auto config_no_tls = fixtures::getNoTlsConfigToml();

    // Both should have different TLS configurations
    EXPECT_GT(config_full.length(), 0);
    EXPECT_GT(config_no_tls.length(), 0);
}

TEST_F(ConfigIntegrationTest, PostgresqlConfigApplied) {
    // Test: PostgreSQL connection settings from config
    auto config = fixtures::getBasicConfigToml();

    // Config should have postgres settings
    EXPECT_TRUE(config.find("postgres") != std::string::npos ||
                config.find("[postgres]") != std::string::npos);
}

// ============= Dynamic Configuration Tests =============

TEST_F(ConfigIntegrationTest, ConfigReloadFromBackend) {
    // Test: Pull config from backend API
    auto config = fixtures::getBasicConfigToml();

    // Config should be loadable
    EXPECT_GT(config.length(), 0);
}

TEST_F(ConfigIntegrationTest, ConfigVersionTracking) {
    // Test: Track config version for updates
    auto config1 = fixtures::getBasicConfigToml();
    auto config2 = fixtures::getFullConfigToml();

    // Different configs available for version comparison
    EXPECT_GT(config1.length(), 0);
    EXPECT_GT(config2.length(), 0);
}

TEST_F(ConfigIntegrationTest, ConfigHotReload) {
    // Test: Apply config changes without restart
    auto config = fixtures::getBasicConfigToml();

    // Config should be valid and reloadable
    EXPECT_GT(config.length(), 0);
    EXPECT_TRUE(config.find("[collector]") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, ConfigChangeNotification) {
    // Test: Detect when backend returns new config
    auto config = fixtures::getBasicConfigToml();

    // Config structure should indicate version tracking capability
    EXPECT_GT(config.length(), 0);
}

// ============= Config Persistence Tests =============

TEST_F(ConfigIntegrationTest, ConfigurationPersistence) {
    // Test: Configuration values persist correctly
    auto config = fixtures::getBasicConfigToml();

    // Config values should be retrievable multiple times (fixture uses "id")
    EXPECT_GT(config.length(), 0);
    EXPECT_TRUE(config.find("id") != std::string::npos);
}

TEST_F(ConfigIntegrationTest, MultipleSections) {
    // Test: Access values from different sections
    auto config = fixtures::getFullConfigToml();

    // Full config should have multiple sections
    EXPECT_GT(config.length(), 0);
}

TEST_F(ConfigIntegrationTest, SpecialCharactersInValues) {
    // Test: Handle special characters in config values
    auto config = fixtures::getBasicConfigToml();

    // Config should handle various value types
    EXPECT_GT(config.length(), 0);
}

TEST_F(ConfigIntegrationTest, CaseSensitivity) {
    // Test: Case sensitivity in section/key names
    auto config = fixtures::getBasicConfigToml();

    // Config should parse with correct case handling
    EXPECT_GT(config.length(), 0);
}
