#pragma once

#include <string>
#include <memory>
#include <vector>
#include <chrono>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * E2E Test Harness
 *
 * Manages the docker-compose environment and provides utilities for E2E testing.
 * Handles:
 * - Docker compose lifecycle (start, stop, health checks)
 * - Database operations (clear, query metrics)
 * - HTTP client initialization
 * - Wait conditions for service readiness
 */
class E2ETestHarness {
public:
    E2ETestHarness();
    ~E2ETestHarness();

    // Lifecycle management
    bool startStack(int timeout_seconds = 60);
    bool stopStack();
    bool resetData();
    bool isStackRunning();

    // Health checks
    bool isBackendReady(int timeout_seconds = 30);
    bool isCollectorRunning(int timeout_seconds = 30);
    bool isDatabaseReady(int timeout_seconds = 30);
    bool isGrafanaReady(int timeout_seconds = 30);
    bool isTimescaleReady(int timeout_seconds = 30);

    // Wait utilities
    bool waitForCondition(
        const std::function<bool()>& condition,
        int timeout_seconds = 30
    );
    bool waitForMetrics(
        int expected_count,
        int timeout_seconds = 60
    );

    // URL and connection helpers
    std::string getBackendUrl() const;
    std::string getBackendHost() const;
    int getBackendPort() const;
    std::string getDatabaseUrl() const;
    std::string getGrafanaUrl() const;
    std::string getTimescaleUrl() const;

    // Getter for docker compose directory
    std::string getComposeDir() const;

    // Environment configuration
    void setCollectorId(const std::string& id);
    void setBackendUrl(const std::string& url);
    void setTestMode(bool enabled);

    // Logging
    void setLogLevel(const std::string& level);
    void printStackStatus();

private:
    // Docker compose management
    bool executeCommand(const std::string& command, std::string& output);
    bool getServiceStatus(const std::string& service, std::string& status);
    bool checkServiceHealth(const std::string& service);

    // Connection testing
    bool testHttpConnection();
    bool testDatabaseConnection();
    bool testGrafanaConnection();

    // Member variables
    std::string m_compose_dir;
    std::string m_backend_url;
    std::string m_database_url;
    std::string m_grafana_url;
    std::string m_timescale_url;
    std::string m_collector_id;
    bool m_stack_running;
    std::string m_log_level;
};

