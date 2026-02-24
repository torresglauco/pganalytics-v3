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
#include "../include/query_stats_plugin.h"

// Global state (gConfig is defined in config_manager.cpp)
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

    // Create query stats collector (will be collected separately if enabled)
    std::unique_ptr<PgQueryStatsCollector> queryStatsCollector = nullptr;
    if (gConfig->isCollectorEnabled("pg_query_stats")) {
        queryStatsCollector = std::make_unique<PgQueryStatsCollector>(
            gConfig->getHostname(),
            gConfig->getCollectorId(),
            pgConfig.host,
            pgConfig.port,
            pgConfig.user,
            pgConfig.password,
            pgConfig.databases
        );
        std::cout << "Added PgQueryStatsCollector" << std::endl;
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
    // For now, generate one with 1 hour expiration
    authMgr.generateToken(3600);  // 1 hour
    sender.setAuthToken(authMgr.getToken(), authMgr.getTokenExpiration());

    // Main collection loop
    int collectionInterval = gConfig->getCollectionInterval("collector", 60);
    int pushInterval = gConfig->getInt("collector", "push_interval", 60);
    int configPullInterval = gConfig->getInt("collector", "config_pull_interval", 300);

    std::cout << "Starting collection loop (collect every " << collectionInterval
              << "s, push every " << pushInterval << "s, config pull every " << configPullInterval << "s)" << std::endl;

    auto lastPushTime = std::chrono::steady_clock::now();
    auto lastConfigPullTime = std::chrono::steady_clock::now();

    while (!shouldExit) {
        // Collect metrics
        std::cout << "Collecting metrics..." << std::endl;
        json collectedMetrics = collectorMgr.collectAll();

        // Validate collected metrics
        if (collectedMetrics.contains("metrics") && collectedMetrics["metrics"].is_array()) {
            for (const auto& metric : collectedMetrics["metrics"]) {
                // Handle metrics with databases array (pg_stats, pg_query_stats)
                if (metric.contains("type") && (metric["type"] == "pg_stats" || metric["type"] == "pg_query_stats") &&
                    metric.contains("databases") && metric["databases"].is_array()) {
                    // Flatten: add each database as a separate metric
                    for (const auto& db_metric : metric["databases"]) {
                        if (MetricsSerializer::validateMetric(db_metric)) {
                            if (!buffer.append(db_metric)) {
                                std::cerr << "Failed to append metric to buffer (buffer full)" << std::endl;
                            }
                        } else {
                            std::cerr << "Invalid metric: " << MetricsSerializer::getLastValidationError() << std::endl;
                        }
                    }
                } else if (MetricsSerializer::validateMetric(metric)) {
                    if (!buffer.append(metric)) {
                        std::cerr << "Failed to append metric to buffer (buffer full)" << std::endl;
                    }
                } else {
                    std::cerr << "Invalid metric: " << MetricsSerializer::getLastValidationError() << std::endl;
                }
            }
        }

        // Collect query statistics if enabled
        if (queryStatsCollector && queryStatsCollector->isEnabled()) {
            std::cout << "Collecting query statistics..." << std::endl;
            json queryStats = queryStatsCollector->execute();
            if (!queryStats.is_null() && queryStats.contains("databases") && queryStats["databases"].is_array()) {
                // Flatten: add each database as a separate metric
                for (const auto& db_metric : queryStats["databases"]) {
                    if (MetricsSerializer::validateMetric(db_metric)) {
                        if (!buffer.append(db_metric)) {
                            std::cerr << "Failed to append query stats to buffer (buffer full)" << std::endl;
                        }
                    } else {
                        std::cerr << "Invalid query stats: " << MetricsSerializer::getLastValidationError() << std::endl;
                    }
                }
            }
        }

        // Check if it's time to push metrics
        auto now = std::chrono::steady_clock::now();
        auto secsSincePush = std::chrono::duration_cast<std::chrono::seconds>(now - lastPushTime).count();

        if (secsSincePush >= pushInterval && !buffer.isEmpty()) {
            std::cout << "Pushing " << buffer.getMetricCount() << " metrics to backend..." << std::endl;

            // Create payload with uncompressed metrics from buffer
            int metricCount = buffer.getMetricCount();
            json metricsArray;
            std::vector<json> metricsVector;

            if (buffer.getUncompressed(metricsArray) && metricsArray.is_array()) {
                // Convert JSON array to vector for createPayload
                for (const auto& metric : metricsArray) {
                    metricsVector.push_back(metric);
                }
            }

            json payload = MetricsSerializer::createPayload(
                gConfig->getCollectorId(),
                gConfig->getHostname(),
                "3.0.0",
                metricsVector
            );
            payload["metrics_count"] = metricCount;

            if (sender.pushMetrics(payload)) {
                std::cout << "Metrics pushed successfully" << std::endl;
                buffer.clear();
            } else {
                std::cerr << "Failed to push metrics" << std::endl;
            }

            lastPushTime = now;
        }

        // Check if it's time to pull configuration
        auto secsSinceConfigPull = std::chrono::duration_cast<std::chrono::seconds>(now - lastConfigPullTime).count();

        if (secsSinceConfigPull >= configPullInterval) {
            std::cout << "Pulling configuration from backend..." << std::endl;

            std::string newConfigToml;
            int newConfigVersion = 0;

            if (sender.pullConfig(gConfig->getCollectorId(), newConfigToml, newConfigVersion)) {
                // Check if we got a new configuration
                if (!newConfigToml.empty()) {
                    std::cout << "Applying new configuration (version " << newConfigVersion << ")..." << std::endl;

                    // Try to load the new configuration
                    if (gConfig->loadFromString(newConfigToml)) {
                        // Reconfigure collectors with new settings
                        collectorMgr.configure(gConfig->toJson());

                        std::cout << "Configuration updated successfully (version " << newConfigVersion << ")" << std::endl;
                    } else {
                        std::cerr << "Failed to parse new configuration: " << gConfig->getLastError() << std::endl;
                    }
                } else {
                    std::cout << "No configuration update available" << std::endl;
                }
            } else {
                std::cerr << "Failed to pull configuration from backend (will retry next interval)" << std::endl;
                // Continue with current config - graceful degradation
            }

            lastConfigPullTime = now;
        }

        // Sleep a bit to avoid busy-waiting
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
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
