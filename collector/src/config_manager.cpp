#include "../include/config_manager.h"
#include <fstream>
#include <sstream>
#include <algorithm>
#include <cctype>

std::shared_ptr<ConfigManager> gConfig = nullptr;

ConfigManager::ConfigManager(const std::string& configFilePath)
    : configFilePath_(configFilePath) {
}

bool ConfigManager::loadFromFile() {
    std::ifstream file(configFilePath_);
    if (!file.is_open()) {
        lastError_ = "Cannot open config file: " + configFilePath_;
        return false;
    }

    std::stringstream buffer;
    buffer << file.rdbuf();
    file.close();

    return parseTOML(buffer.str());
}

bool ConfigManager::loadFromString(const std::string& tomlContent) {
    return parseTOML(tomlContent);
}

bool ConfigManager::parseTOML(const std::string& content) {
    std::istringstream stream(content);
    std::string line;
    std::string currentSection;

    while (std::getline(stream, line)) {
        // Trim line
        line = trim(line);

        // Skip empty lines and comments
        if (line.empty() || line[0] == '#' || line[0] == ';') {
            continue;
        }

        // Check for section header [section]
        if (line[0] == '[' && line[line.length() - 1] == ']') {
            currentSection = line.substr(1, line.length() - 2);
            currentSection = trim(currentSection);
            continue;
        }

        // Parse key=value
        size_t eqPos = line.find('=');
        if (eqPos == std::string::npos) {
            continue;
        }

        std::string key = trim(line.substr(0, eqPos));
        std::string value = trim(line.substr(eqPos + 1));

        // Remove quotes if present
        if (!value.empty() && value[0] == '"' && value[value.length() - 1] == '"') {
            value = value.substr(1, value.length() - 2);
        } else if (!value.empty() && value[0] == '\'' && value[value.length() - 1] == '\'') {
            value = value.substr(1, value.length() - 2);
        }

        config_[currentSection][key] = value;
    }

    return true;
}

std::string ConfigManager::getString(const std::string& section, const std::string& key, const std::string& defaultValue) const {
    auto secIt = config_.find(section);
    if (secIt == config_.end()) {
        return defaultValue;
    }

    auto keyIt = secIt->second.find(key);
    if (keyIt == secIt->second.end()) {
        return defaultValue;
    }

    return keyIt->second;
}

int ConfigManager::getInt(const std::string& section, const std::string& key, int defaultValue) const {
    std::string value = getString(section, key);
    if (value.empty()) {
        return defaultValue;
    }

    try {
        return std::stoi(value);
    } catch (...) {
        return defaultValue;
    }
}

bool ConfigManager::getBool(const std::string& section, const std::string& key, bool defaultValue) const {
    std::string value = getString(section, key);
    if (value.empty()) {
        return defaultValue;
    }

    // Convert to lowercase for comparison
    std::transform(value.begin(), value.end(), value.begin(), ::tolower);
    return value == "true" || value == "yes" || value == "1";
}

std::vector<std::string> ConfigManager::getStringArray(const std::string& section, const std::string& key) const {
    std::string value = getString(section, key);
    if (value.empty()) {
        return {};
    }

    // Remove brackets if present (for arrays like [db1, db2, db3])
    if (value[0] == '[' && value[value.length() - 1] == ']') {
        value = value.substr(1, value.length() - 2);
    }

    return split(value, ',');
}

void ConfigManager::set(const std::string& section, const std::string& key, const std::string& value) {
    config_[section][key] = value;
}

json ConfigManager::toJson() const {
    json result = json::object();
    for (const auto& section : config_) {
        json sectionObj = json::object();
        for (const auto& kv : section.second) {
            sectionObj[kv.first] = kv.second;
        }
        result[section.first] = sectionObj;
    }
    return result;
}

std::string ConfigManager::getCollectorId() const {
    return getString("collector", "id", "collector-001");
}

std::string ConfigManager::getBackendUrl() const {
    return getString("backend", "url", "https://localhost:8080");
}

std::string ConfigManager::getHostname() const {
    return getString("collector", "hostname", "localhost");
}

bool ConfigManager::isCollectorEnabled(const std::string& collectorType) const {
    return getBool(collectorType, "enabled", true);
}

int ConfigManager::getCollectionInterval(const std::string& collectorType, int defaultSeconds) const {
    return getInt(collectorType, "interval", defaultSeconds);
}

ConfigManager::PostgreSQLConfig ConfigManager::getPostgreSQLConfig() const {
    PostgreSQLConfig cfg;
    cfg.host = getString("postgres", "host", "localhost");
    cfg.port = getInt("postgres", "port", 5432);
    cfg.user = getString("postgres", "user", "postgres");
    cfg.password = getString("postgres", "password", "");
    cfg.defaultDatabase = getString("postgres", "database", "postgres");
    cfg.databases = getStringArray("postgres", "databases");
    if (cfg.databases.empty()) {
        cfg.databases.push_back(cfg.defaultDatabase);
    }
    return cfg;
}

ConfigManager::TLSConfig ConfigManager::getTLSConfig() const {
    TLSConfig cfg;
    cfg.verify = getBool("tls", "verify", false);  // Default to false for self-signed certs
    cfg.certFile = getString("tls", "cert_file", "/etc/pganalytics/collector.crt");
    cfg.keyFile = getString("tls", "key_file", "/etc/pganalytics/collector.key");
    cfg.caFile = getString("tls", "ca_file", "");
    return cfg;
}

std::string ConfigManager::getLastError() const {
    return lastError_;
}

std::string ConfigManager::trim(const std::string& str) {
    size_t start = 0;
    while (start < str.length() && std::isspace(str[start])) {
        start++;
    }

    size_t end = str.length();
    while (end > start && std::isspace(str[end - 1])) {
        end--;
    }

    return str.substr(start, end - start);
}

std::vector<std::string> ConfigManager::split(const std::string& str, char delimiter) {
    std::vector<std::string> tokens;
    std::stringstream ss(str);
    std::string token;

    while (std::getline(ss, token, delimiter)) {
        token = trim(token);
        if (!token.empty()) {
            tokens.push_back(token);
        }
    }

    return tokens;
}
