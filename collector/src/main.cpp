#include <iostream>
#include <string>
#include <vector>
#include <chrono>
#include <thread>
#include <signal.h>
#include "../include/collector.h"
#include "../include/config_manager.h"
#include "../include/auth.h"
#include "../include/sender.h"
#include "../include/metrics_serializer.h"
#include "../include/metrics_buffer.h"

// Global configuration and state
std::shared_ptr<ConfigManager> gConfig = nullptr;
volatile sig_atomic_t shouldExit = 0;

// Signal handler for graceful shutdown
void signalHandler(int signum) {
    std::cout << "\nReceived signal " << signum << ", shutting down..." << std::endl;
    shouldExit = 1;
}

int runCronMode() {
    std::cout << "Starting collector in cron mode..." << std::endl;

    // Load configuration
    if (!gConfig->loadFromFile()) {
        std::cerr << "Failed to load configuration: " << gConfig->getLastError() << std::endl;
        return 1;
    }

    std::cout << "Configuration loaded successfully" << std::endl;
    std::cout << "Collector ID: " << gConfig->getCollectorId() << std::endl;
    std::cout << "Backend URL: " << gConfig->getBackendUrl() << std::endl;

    // Initialize authentication
    AuthManager authMgr(gConfig->getCollectorId());

    // Load TLS configuration
    auto tlsConfig = gConfig->getTLSConfig();
    if (!authMgr.loadClientCertificate(tlsConfig.certFile)) {
        std::cerr << "Failed to load client certificate: " << authMgr.getLastError() << std::endl;
        // This might be okay for first run - could be not registered yet
    }

    // Initialize collector manager
    CollectorManager collectorMgr(gConfig->getHostname(), gConfig->getCollectorId());

    // Add collectors
    auto pgConfig = gConfig->getPostgreSQLConfig();
    if (gConfig->isCollectorEnabled("pg_stats")) {
        auto pgStatsCollector = std::make_shared<PgStatsCollector>(
            gConfig->getHostname(),
            gConfig->getCollectorId(),
            pgConfig.host,
            pgConfig.port,
            pgConfig.user,
            pgConfig.password,
            pgConfig.databases
        );
        collectorMgr.addCollector(pgStatsCollector);
        std::cout << "Added PgStatsCollector" << std::endl;
    }

    if (gConfig->isCollectorEnabled("sysstat")) {
        auto sysstatCollector = std::make_shared<SysstatCollector>(
            gConfig->getHostname(),
            gConfig->getCollectorId()
        );
        collectorMgr.addCollector(sysstatCollector);
        std::cout << "Added SysstatCollector" << std::endl;
    }

    if (gConfig->isCollectorEnabled("disk_usage")) {
        auto diskCollector = std::make_shared<DiskUsageCollector>(
            gConfig->getHostname(),
            gConfig->getCollectorId()
        );
        collectorMgr.addCollector(diskCollector);
        std::cout << "Added DiskUsageCollector" << std::endl;
    }

    if (gConfig->isCollectorEnabled("pg_log")) {
        auto logCollector = std::make_shared<PgLogCollector>(
            gConfig->getHostname(),
            gConfig->getCollectorId(),
            pgConfig.host,
            pgConfig.port,
            pgConfig.user,
            pgConfig.password
        );
        collectorMgr.addCollector(logCollector);
        std::cout << "Added PgLogCollector" << std::endl;
    }

    // Initialize metrics buffer and sender
    MetricsBuffer buffer;
    Sender sender(
        gConfig->getBackendUrl(),
        gConfig->getCollectorId(),
        tlsConfig.certFile,
        tlsConfig.keyFile,
        tlsConfig.verify
    );

    // Set initial auth token (would be fetched during registration)
    // For now, generate one
    sender.setAuthToken(authMgr.generateToken());

    // Main collection loop
    int collectionInterval = gConfig->getCollectionInterval("collector", 60);
    int pushInterval = gConfig->getInt("collector", "push_interval", 60);
    int configPullInterval = gConfig->getInt("collector", "config_pull_interval", 300);

    std::cout << "Starting collection loop (collect every " << collectionInterval
              << "s, push every " << pushInterval << "s)" << std::endl;

    auto lastPushTime = std::chrono::steady_clock::now();
    auto lastConfigPullTime = std::chrono::steady_clock::now();

    while (!shouldExit) {
        // Collect metrics
        std::cout << "Collecting metrics..." << std::endl;
        json collectedMetrics = collectorMgr.collectAll();

        // Validate collected metrics
        if (collectedMetrics.contains("metrics") && collectedMetrics["metrics"].is_array()) {
            for (const auto& metric : collectedMetrics["metrics"]) {
                if (MetricsSerializer::validateMetric(metric)) {
                    if (!buffer.append(metric)) {
                        std::cerr << "Failed to append metric to buffer (buffer full)" << std::endl;
                    }
                } else {
                    std::cerr << "Invalid metric: " << MetricsSerializer::getLastValidationError() << std::endl;
                }
            }
        }

        // Check if it's time to push metrics
        auto now = std::chrono::steady_clock::now();
        auto secsSincePush = std::chrono::duration_cast<std::chrono::seconds>(now - lastPushTime).count();

        if (secsSincePush >= pushInterval && !buffer.isEmpty()) {
            std::cout << "Pushing " << buffer.getMetricCount() << " metrics to backend..." << std::endl;

            // Create payload
            json payload = MetricsSerializer::createPayload(
                gConfig->getCollectorId(),
                gConfig->getHostname(),
                "3.0.0",
                std::vector<json>()  // Metrics will be sent compressed
            );

            // Get compressed data
            std::string compressed;
            if (buffer.getCompressed(compressed)) {
                std::cout << "Compressed " << buffer.getUncompressedSize() << " bytes to "
                          << buffer.getEstimatedCompressedSize() << " bytes ("
                          << buffer.getCompressionRatio() << "%)" << std::endl;

                if (sender.pushMetrics(payload)) {
                    std::cout << "Metrics pushed successfully" << std::endl;
                    buffer.clear();
                } else {
                    std::cerr << "Failed to push metrics" << std::endl;
                }
            }

            lastPushTime = now;
        }

        // Sleep before next collection
        std::this_thread::sleep_for(std::chrono::seconds(collectionInterval));
    }

    std::cout << "Collector stopped" << std::endl;
    return 0;
}

int runRegister() {
    std::cout << "Collector registration mode" << std::endl;
    std::cout << "TODO: Implement registration flow" << std::endl;
    return 0;
}

int main(int argc, char* argv[]) {
    std::cout << "pgAnalytics Collector v3.0.0" << std::endl;

    // Setup signal handlers
    signal(SIGTERM, signalHandler);
    signal(SIGINT, signalHandler);

    // Initialize global config
    gConfig = std::make_shared<ConfigManager>("/etc/pganalytics/collector.toml");

    std::string action = "cron";
    if (argc > 1) {
        action = argv[1];
    }

    std::cout << "Action: " << action << std::endl;

    if (action == "cron") {
        return runCronMode();
    } else if (action == "register") {
        return runRegister();
    } else if (action == "help") {
        std::cout << "Usage: pganalytics [action]" << std::endl;
        std::cout << "Actions:" << std::endl;
        std::cout << "  cron       - Run continuous collection (default)" << std::endl;
        std::cout << "  register   - Register with backend and get credentials" << std::endl;
        std::cout << "  help       - Show this help message" << std::endl;
        return 0;
    } else {
        std::cerr << "Unknown action: " << action << std::endl;
        return 1;
    }

    return 0;
}
