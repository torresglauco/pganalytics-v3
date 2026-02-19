#pragma once

#include <string>
#include <map>
#include <vector>
#include <nlohmann/json.hpp>
#include <memory>

using json = nlohmann::json;

/**
 * Configuration Manager
 * Handles loading local TOML configuration and pulling dynamic config from backend API
 */
class ConfigManager {
public:
    /**
     * Create configuration manager
     * @param configFilePath Path to local TOML configuration file
     */
    explicit ConfigManager(const std::string& configFilePath);

    /**
     * Load configuration from file
     * @return true if successful, false otherwise
     */
    bool loadFromFile();

    /**
     * Get a string configuration value
     * @param section Section name (e.g., "backend", "postgres")
     * @param key Configuration key
     * @param defaultValue Value to return if key not found
     * @return Configuration value
     */
    std::string getString(const std::string& section, const std::string& key, const std::string& defaultValue = "");

    /**
     * Get an integer configuration value
     */
    int getInt(const std::string& section, const std::string& key, int defaultValue = 0);

    /**
     * Get a boolean configuration value
     */
    bool getBool(const std::string& section, const std::string& key, bool defaultValue = false);

    /**
     * Get a string array configuration value
     */
    std::vector<std::string> getStringArray(const std::string& section, const std::string& key);

    /**
     * Set a configuration value
     */
    void set(const std::string& section, const std::string& key, const std::string& value);

    /**
     * Get the entire configuration as JSON
     */
    json toJson() const;

    /**
     * Get the collector ID from config
     */
    std::string getCollectorId() const;

    /**
     * Get the backend URL from config
     */
    std::string getBackendUrl() const;

    /**
     * Get the hostname from config
     */
    std::string getHostname() const;

    /**
     * Get which collectors are enabled
     */
    bool isCollectorEnabled(const std::string& collectorType) const;

    /**
     * Get the collection interval in seconds
     */
    int getCollectionInterval(const std::string& collectorType, int defaultSeconds = 60) const;

    /**
     * Get PostgreSQL connection parameters
     */
    struct PostgreSQLConfig {
        std::string host;
        int port;
        std::string user;
        std::string password;
        std::string defaultDatabase;
        std::vector<std::string> databases;
    };

    PostgreSQLConfig getPostgreSQLConfig() const;

    /**
     * Get TLS configuration
     */
    struct TLSConfig {
        bool verify;
        std::string certFile;
        std::string keyFile;
        std::string caFile;
    };

    TLSConfig getTLSConfig() const;

    /**
     * Get last error message
     */
    std::string getLastError() const;

private:
    std::string configFilePath_;
    std::map<std::string, std::map<std::string, std::string>> config_;
    std::string lastError_;

    /**
     * Parse TOML file (simplified TOML parser)
     */
    bool parseTOML(const std::string& content);

    /**
     * Trim whitespace from string
     */
    static std::string trim(const std::string& str);

    /**
     * Split string by delimiter
     */
    static std::vector<std::string> split(const std::string& str, char delimiter);
};

// Forward declare for global config instance
extern std::shared_ptr<ConfigManager> gConfig;
