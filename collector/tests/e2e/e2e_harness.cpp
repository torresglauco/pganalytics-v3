#include "e2e_harness.h"
#include <iostream>
#include <thread>
#include <chrono>
#include <cstdlib>
#include <sstream>
#include <fstream>

E2ETestHarness::E2ETestHarness()
    : m_compose_dir("collector/tests/e2e"),
      m_backend_url("https://localhost:8080"),
      m_database_url("postgresql://postgres:pganalytics@localhost:5432/pganalytics"),
      m_grafana_url("http://localhost:3000"),
      m_timescale_url("postgresql://postgres:pganalytics@localhost:5433/metrics"),
      m_collector_id("e2e_col_001"),
      m_stack_running(false),
      m_log_level("info") {
}

E2ETestHarness::~E2ETestHarness() {
    if (m_stack_running) {
        stopStack();
    }
}

bool E2ETestHarness::startStack(int timeout_seconds) {
    std::cout << "\n[E2E] Starting docker-compose stack..." << std::endl;

    // Change to compose directory
    std::string original_dir = getenv("PWD") ? getenv("PWD") : "";

    // Run docker-compose up
    std::string command = "cd " + m_compose_dir + " && docker-compose -f docker-compose.e2e.yml up -d 2>&1";
    std::string output;

    if (!executeCommand(command, output)) {
        std::cerr << "[E2E] Failed to start docker-compose: " << output << std::endl;
        return false;
    }

    std::cout << "[E2E] Docker-compose started" << std::endl;
    m_stack_running = true;

    // Wait for services to be ready
    if (!waitForCondition(
        [this]() {
            return isBackendReady(5) &&
                   isDatabaseReady(5) &&
                   isTimescaleReady(5);
        },
        timeout_seconds
    )) {
        std::cerr << "[E2E] Services failed to become ready" << std::endl;
        printStackStatus();
        return false;
    }

    std::cout << "[E2E] All services ready" << std::endl;
    return true;
}

bool E2ETestHarness::stopStack() {
    std::cout << "\n[E2E] Stopping docker-compose stack..." << std::endl;

    std::string command = "cd " + m_compose_dir + " && docker-compose -f docker-compose.e2e.yml down 2>&1";
    std::string output;

    if (!executeCommand(command, output)) {
        std::cerr << "[E2E] Warning: docker-compose down failed: " << output << std::endl;
        // Don't return false - try to continue cleanup
    }

    m_stack_running = false;
    std::cout << "[E2E] Docker-compose stopped" << std::endl;
    return true;
}

bool E2ETestHarness::resetData() {
    std::cout << "\n[E2E] Resetting test data..." << std::endl;

    // Clear metrics tables in TimescaleDB
    std::string psql_cmd =
        "PGPASSWORD=pganalytics psql -h localhost -p 5433 -U postgres -d metrics "
        "-c 'TRUNCATE TABLE IF EXISTS metrics_pg_stats CASCADE;' 2>&1";

    std::string output;
    executeCommand(psql_cmd, output);

    // Clear collector registry
    psql_cmd =
        "PGPASSWORD=pganalytics psql -h localhost -p 5432 -U postgres -d pganalytics "
        "-c 'TRUNCATE TABLE IF EXISTS pganalytics.collector_registry CASCADE;' 2>&1";

    executeCommand(psql_cmd, output);

    std::cout << "[E2E] Test data reset" << std::endl;
    return true;
}

bool E2ETestHarness::isStackRunning() {
    return m_stack_running;
}

bool E2ETestHarness::isBackendReady(int timeout_seconds) {
    return waitForCondition(
        [this]() { return testHttpConnection(); },
        timeout_seconds
    );
}

bool E2ETestHarness::isDatabaseReady(int timeout_seconds) {
    return waitForCondition(
        [this]() { return testDatabaseConnection(); },
        timeout_seconds
    );
}

bool E2ETestHarness::isTimescaleReady(int timeout_seconds) {
    // Test TimescaleDB on port 5433
    std::string cmd = "PGPASSWORD=pganalytics psql -h localhost -p 5433 -U postgres -d metrics "
                      "-c 'SELECT version();' > /dev/null 2>&1";

    return waitForCondition(
        [cmd]() {
            int ret = system(cmd.c_str());
            return ret == 0;
        },
        timeout_seconds
    );
}

bool E2ETestHarness::isGrafanaReady(int timeout_seconds) {
    return waitForCondition(
        [this]() { return testGrafanaConnection(); },
        timeout_seconds
    );
}

bool E2ETestHarness::waitForCondition(
    const std::function<bool()>& condition,
    int timeout_seconds
) {
    auto start = std::chrono::steady_clock::now();

    while (true) {
        if (condition()) {
            return true;
        }

        auto elapsed = std::chrono::steady_clock::now() - start;
        if (std::chrono::duration_cast<std::chrono::seconds>(elapsed).count() >= timeout_seconds) {
            return false;
        }

        std::this_thread::sleep_for(std::chrono::milliseconds(500));
    }
}

bool E2ETestHarness::waitForMetrics(int expected_count, int timeout_seconds) {
    return waitForCondition(
        [this, expected_count]() {
            std::string cmd =
                "PGPASSWORD=pganalytics psql -h localhost -p 5433 -U postgres -d metrics "
                "-tc 'SELECT COUNT(*) FROM metrics_pg_stats;' 2>/dev/null | tr -d ' '";

            std::string output;
            if (!executeCommand(cmd, output)) {
                return false;
            }

            try {
                int count = std::stoi(output);
                return count >= expected_count;
            } catch (...) {
                return false;
            }
        },
        timeout_seconds
    );
}

std::string E2ETestHarness::getBackendUrl() const {
    return m_backend_url;
}

std::string E2ETestHarness::getBackendHost() const {
    return "localhost";
}

int E2ETestHarness::getBackendPort() const {
    return 8080;
}

std::string E2ETestHarness::getDatabaseUrl() const {
    return m_database_url;
}

std::string E2ETestHarness::getGrafanaUrl() const {
    return m_grafana_url;
}

std::string E2ETestHarness::getTimescaleUrl() const {
    return m_timescale_url;
}

std::string E2ETestHarness::getComposeDir() const {
    return m_compose_dir;
}

void E2ETestHarness::setCollectorId(const std::string& id) {
    m_collector_id = id;
}

void E2ETestHarness::setBackendUrl(const std::string& url) {
    m_backend_url = url;
}

void E2ETestHarness::setTestMode(bool enabled) {
    // For future use: configure services in test mode
}

void E2ETestHarness::setLogLevel(const std::string& level) {
    m_log_level = level;
}

void E2ETestHarness::printStackStatus() {
    std::cout << "\n[E2E] Docker Compose Stack Status:" << std::endl;

    std::string cmd = "cd " + m_compose_dir + " && docker-compose -f docker-compose.e2e.yml ps";
    std::string output;
    executeCommand(cmd, output);
    std::cout << output << std::endl;
}

bool E2ETestHarness::executeCommand(const std::string& command, std::string& output) {
    FILE* pipe = popen(command.c_str(), "r");
    if (!pipe) {
        return false;
    }

    char buffer[256];
    output.clear();

    while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
        output += buffer;
    }

    int status = pclose(pipe);
    return status == 0;
}

bool E2ETestHarness::getServiceStatus(const std::string& service, std::string& status) {
    std::string cmd = "cd " + m_compose_dir + " && docker-compose -f docker-compose.e2e.yml ps " +
                      service + " 2>&1";
    return executeCommand(cmd, status);
}

bool E2ETestHarness::checkServiceHealth(const std::string& service) {
    std::string status;
    if (!getServiceStatus(service, status)) {
        return false;
    }

    // Check if status contains "healthy" or "running"
    return status.find("healthy") != std::string::npos ||
           status.find("running") != std::string::npos;
}

bool E2ETestHarness::testHttpConnection() {
    // Test backend API health endpoint
    std::string cmd = "curl -s -k -f https://localhost:8080/api/v1/health > /dev/null 2>&1";
    int ret = system(cmd.c_str());
    return ret == 0;
}

bool E2ETestHarness::testDatabaseConnection() {
    // Test PostgreSQL connection
    std::string cmd = "PGPASSWORD=pganalytics psql -h localhost -U postgres -d pganalytics "
                      "-c 'SELECT 1;' > /dev/null 2>&1";
    int ret = system(cmd.c_str());
    return ret == 0;
}

bool E2ETestHarness::testGrafanaConnection() {
    // Test Grafana API
    std::string cmd = "curl -s -f http://localhost:3000/api/health > /dev/null 2>&1";
    int ret = system(cmd.c_str());
    return ret == 0;
}

