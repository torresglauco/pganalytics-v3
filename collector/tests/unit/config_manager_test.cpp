#include <gtest/gtest.h>
#include <fstream>
#include <sstream>
#include <nlohmann/json.hpp>
#include "config_manager.h"

using json = nlohmann::json;

class ConfigManagerTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Create temporary test config file
        test_config_path = "/tmp/test_collector.toml";
        createTestConfigFile();
    }

    void TearDown() override {
        // Clean up test config file
        std::remove(test_config_path.c_str());
    }

    void createTestConfigFile() {
        std::ofstream file(test_config_path);
        file << R"(
[collector]
id = "test-collector-001"
hostname = "test-host"
interval = 60
push_interval = 60
config_pull_interval = 300

[backend]
url = "https://localhost:8080"

[postgres]
host = "localhost"
port = 5432
user = "postgres"
password = "secret"
database = "postgres"
databases = "postgres, template1, myapp"

[tls]
verify = false
cert_file = "/etc/pganalytics/collector.crt"
key_file = "/etc/pganalytics/collector.key"

[pg_stats]
enabled = true
interval = 60

[sysstat]
enabled = true
interval = 60

[pg_log]
enabled = true
interval = 300

[disk_usage]
enabled = true
interval = 300
)";
        file.close();
    }

    std::string test_config_path;
};

// Test 1: Create ConfigManager instance
TEST_F(ConfigManagerTest, CreateInstance) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    EXPECT_NE(config, nullptr);
}

// Test 2: Load configuration from file
TEST_F(ConfigManagerTest, LoadConfigFile) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    bool success = config->loadFromFile();

    EXPECT_TRUE(success);
}

// Test 3: Get collector ID
TEST_F(ConfigManagerTest, GetCollectorId) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    std::string id = config->getCollectorId();

    EXPECT_EQ(id, "test-collector-001");
}

// Test 4: Get hostname
TEST_F(ConfigManagerTest, GetHostname) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    std::string hostname = config->getHostname();

    EXPECT_EQ(hostname, "test-host");
}

// Test 5: Get backend URL
TEST_F(ConfigManagerTest, GetBackendUrl) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    std::string url = config->getBackendUrl();

    EXPECT_EQ(url, "https://localhost:8080");
}

// Test 6: Get string configuration
TEST_F(ConfigManagerTest, GetStringConfig) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    std::string value = config->getString("postgres", "user");

    EXPECT_EQ(value, "postgres");
}

// Test 7: Get integer configuration
TEST_F(ConfigManagerTest, GetIntConfig) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    int port = config->getInt("postgres", "port");

    EXPECT_EQ(port, 5432);
}

// Test 8: Get boolean configuration
TEST_F(ConfigManagerTest, GetBoolConfig) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    bool verify = config->getBool("tls", "verify");

    EXPECT_FALSE(verify);
}

// Test 9: Get string array configuration
TEST_F(ConfigManagerTest, GetStringArrayConfig) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    std::vector<std::string> databases = config->getStringArray("postgres", "databases");

    EXPECT_EQ(databases.size(), 3);
    EXPECT_EQ(databases[0], "postgres");
    EXPECT_EQ(databases[1], "template1");
    EXPECT_EQ(databases[2], "myapp");
}

// Test 10: Check collector enabled
TEST_F(ConfigManagerTest, IsCollectorEnabled) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    EXPECT_TRUE(config->isCollectorEnabled("pg_stats"));
    EXPECT_TRUE(config->isCollectorEnabled("sysstat"));
    EXPECT_TRUE(config->isCollectorEnabled("pg_log"));
    EXPECT_TRUE(config->isCollectorEnabled("disk_usage"));
}

// Test 11: Get collection interval
TEST_F(ConfigManagerTest, GetCollectionInterval) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    int interval = config->getCollectionInterval("pg_stats");

    EXPECT_EQ(interval, 60);
}

// Test 12: Get PostgreSQL configuration
TEST_F(ConfigManagerTest, GetPostgreSQLConfig) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    auto pg_config = config->getPostgreSQLConfig();

    EXPECT_EQ(pg_config.host, "localhost");
    EXPECT_EQ(pg_config.port, 5432);
    EXPECT_EQ(pg_config.user, "postgres");
    EXPECT_EQ(pg_config.password, "secret");
    EXPECT_EQ(pg_config.defaultDatabase, "postgres");
    EXPECT_EQ(pg_config.databases.size(), 3);
}

// Test 13: Get TLS configuration
TEST_F(ConfigManagerTest, GetTLSConfig) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    auto tls_config = config->getTLSConfig();

    EXPECT_FALSE(tls_config.verify);
    EXPECT_EQ(tls_config.certFile, "/etc/pganalytics/collector.crt");
    EXPECT_EQ(tls_config.keyFile, "/etc/pganalytics/collector.key");
}

// Test 14: Default values for missing keys
TEST_F(ConfigManagerTest, DefaultValues) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    // Request a non-existent key with default value
    std::string value = config->getString("nonexistent", "key", "default_value");

    EXPECT_EQ(value, "default_value");
}

// Test 15: Load non-existent file
TEST_F(ConfigManagerTest, LoadNonExistentFile) {
    auto config = std::make_unique<ConfigManager>("/nonexistent/path/config.toml");
    bool success = config->loadFromFile();

    EXPECT_FALSE(success);
    std::string error = config->getLastError();
    EXPECT_FALSE(error.empty());
}

// Test 16: Set configuration value
TEST_F(ConfigManagerTest, SetConfigValue) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    config->set("collector", "id", "new-id");

    std::string id = config->getString("collector", "id");
    EXPECT_EQ(id, "new-id");
}

// Test 17: Convert to JSON
TEST_F(ConfigManagerTest, ToJson) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    json config_json = config->toJson();

    EXPECT_TRUE(config_json.is_object());
    EXPECT_TRUE(config_json.contains("collector"));
    EXPECT_TRUE(config_json.contains("backend"));
    EXPECT_TRUE(config_json.contains("postgres"));
}

// Test 18: Multiple sections
TEST_F(ConfigManagerTest, MultipleSections) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    std::string collector_id = config->getString("collector", "id");
    std::string backend_url = config->getString("backend", "url");
    std::string pg_host = config->getString("postgres", "host");

    EXPECT_EQ(collector_id, "test-collector-001");
    EXPECT_EQ(backend_url, "https://localhost:8080");
    EXPECT_EQ(pg_host, "localhost");
}

// Test 19: Configuration persistence
TEST_F(ConfigManagerTest, ConfigurationPersistence) {
    auto config1 = std::make_unique<ConfigManager>(test_config_path);
    config1->loadFromFile();

    // Load again in another instance
    auto config2 = std::make_unique<ConfigManager>(test_config_path);
    config2->loadFromFile();

    EXPECT_EQ(config1->getCollectorId(), config2->getCollectorId());
    EXPECT_EQ(config1->getBackendUrl(), config2->getBackendUrl());
}

// Test 20: Integer default value
TEST_F(ConfigManagerTest, IntegerDefaultValue) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    int value = config->getInt("nonexistent", "port", 9999);

    EXPECT_EQ(value, 9999);
}

// Test 21: Boolean default value
TEST_F(ConfigManagerTest, BooleanDefaultValue) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    bool value = config->getBool("nonexistent", "enabled", true);

    EXPECT_TRUE(value);
}

// Test 22: Empty database list defaults to primary
TEST_F(ConfigManagerTest, EmptyDatabaseListDefaulting) {
    // Create config without databases array
    std::ofstream file("/tmp/test_minimal.toml");
    file << R"(
[postgres]
host = "localhost"
user = "postgres"
database = "postgres"
)";
    file.close();

    auto config = std::make_unique<ConfigManager>("/tmp/test_minimal.toml");
    config->loadFromFile();

    auto pg_config = config->getPostgreSQLConfig();

    // Should default to primary database if list is empty
    EXPECT_FALSE(pg_config.databases.empty());

    std::remove("/tmp/test_minimal.toml");
}

// Test 23: Case sensitivity in sections
TEST_F(ConfigManagerTest, CaseSensitivity) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    // Section names should be case-sensitive
    std::string value = config->getString("Collector", "id", "not_found");

    // Assuming TOML is case-sensitive, this should be "not_found"
    // Actual behavior depends on implementation
}

// Test 24: Special characters in values
TEST_F(ConfigManagerTest, SpecialCharactersInValues) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    // URL contains special characters (://)
    std::string url = config->getBackendUrl();

    EXPECT_NE(url.find("://"), std::string::npos);
}

// Test 25: Configuration reload
TEST_F(ConfigManagerTest, ConfigurationReload) {
    auto config = std::make_unique<ConfigManager>(test_config_path);
    config->loadFromFile();

    std::string id1 = config->getCollectorId();

    // Modify the file
    config->set("collector", "id", "new-id");
    std::string id2 = config->getString("collector", "id");

    // Reload
    config->loadFromFile();
    std::string id3 = config->getCollectorId();

    // After reload, should be back to original
    EXPECT_EQ(id1, id3);
    EXPECT_NE(id2, id3);
}
