#include <iostream>
#include <string>
#include <vector>
#include <chrono>
#include <thread>
#include <signal.h>
#include <fstream>
#include <sys/stat.h>
#include <ctime>
#include "../include/collector.h"
#include "../include/config_manager.h"
#include "../include/auth.h"
#include "../include/sender.h"
#include "../include/metrics_serializer.h"
#include "../include/metrics_buffer.h"
#include "../include/query_stats_plugin.h"
#include "../include/replication_plugin.h"

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
    std::cout << "DEBUG: PostgreSQL Config - host=" << pgConfig.host << ", port=" << pgConfig.port
              << ", databases count=" << pgConfig.databases.size() << std::endl;
    for (const auto& db : pgConfig.databases) {
        std::cout << "  - " << db << std::endl;
    }

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

    if (gConfig->isCollectorEnabled("pg_replication")) {
        auto replicationCollector = std::make_shared<PgReplicationCollector>(
            gConfig->getHostname(),
            gConfig->getCollectorId(),
            pgConfig.host,
            pgConfig.port,
            pgConfig.user,
            pgConfig.password,
            pgConfig.databases
        );
        collectorMgr.addCollector(replicationCollector);
        std::cout << "Added PgReplicationCollector" << std::endl;
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

    // Initialize metrics buffer with larger capacity (50MB to handle query stats with 100 queries × 2 databases)
    MetricsBuffer buffer(50 * 1024 * 1024);
    Sender sender(
        gConfig->getBackendUrl(),
        gConfig->getCollectorId(),
        tlsConfig.certFile,
        tlsConfig.keyFile,
        tlsConfig.verify
    );

    // Try to load auth token from file (saved during registration)
    std::string authToken;
    std::string registeredCollectorId;
    std::string tokenFilePath = "/etc/pganalytics/collector.token";
    std::string collectorIdFilePath = "/etc/pganalytics/collector.id";
    std::ifstream tokenFile(tokenFilePath);
    std::ifstream collectorIdFile(collectorIdFilePath);

    bool tokenLoaded = false;
    if (tokenFile.is_open()) {
        std::getline(tokenFile, authToken);
        tokenFile.close();

        if (!authToken.empty()) {
            std::cout << "Loaded auth token from file" << std::endl;
            tokenLoaded = true;
            // Set token with 24-hour expiration
            time_t now = std::time(nullptr);
            sender.setAuthToken(authToken, now + 86400);
        } else {
            std::cerr << "Warning: Token file exists but is empty" << std::endl;
        }
    } else {
        std::cerr << "Warning: Auth token file not found at " << tokenFilePath << std::endl;
    }

    if (collectorIdFile.is_open()) {
        std::getline(collectorIdFile, registeredCollectorId);
        collectorIdFile.close();

        if (!registeredCollectorId.empty()) {
            std::cout << "Loaded collector ID from file: " << registeredCollectorId << std::endl;
        }
    } else {
        std::cerr << "Warning: Collector ID file not found at " << collectorIdFilePath << std::endl;
    }

    if (!tokenLoaded) {
        std::cerr << "Falling back to local token generation (collector may not be registered)" << std::endl;
        // Fall back to local token generation
        authMgr.generateToken(3600);  // 1 hour
        sender.setAuthToken(authMgr.getToken(), authMgr.getTokenExpiration());
    }

    // Main collection loop
    int collectionInterval = gConfig->getCollectionInterval("collector", 60);
    int pushInterval = gConfig->getInt("collector", "push_interval", 60);
    int configPullInterval = gConfig->getInt("collector", "config_pull_interval", 300);

    std::cout << "Starting collection loop (collect every " << collectionInterval
              << "s, push every " << pushInterval << "s, config pull every " << configPullInterval << "s)" << std::endl;

    auto lastPushTime = std::chrono::steady_clock::now();
    auto lastConfigPullTime = std::chrono::steady_clock::now();

    while (!shouldExit) {
        // Collect metrics (Phase 1.1: now using parallel execution via thread pool)
        std::cout << "Collecting metrics..." << std::endl;
        json collectedMetrics = collectorMgr.collectAllParallel();

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
                // Flatten: create individual metrics for each database with type and timestamp
                std::string timestamp = queryStats.contains("timestamp") ? queryStats["timestamp"].get<std::string>() : "";
                for (const auto& db_metric : queryStats["databases"]) {
                    // Create a metric object that matches the pg_query_stats validation schema:
                    // It should have: type, timestamp, database, queries
                    json metric = {
                        {"type", "pg_query_stats"},
                        {"timestamp", timestamp},
                        {"database", db_metric["database"]},
                        {"queries", db_metric["queries"]}
                    };

                    if (MetricsSerializer::validateMetric(metric)) {
                        if (!buffer.append(metric)) {
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

        std::cout << "DEBUG: Push check - secsSincePush=" << secsSincePush << ", pushInterval=" << pushInterval
                  << ", bufferEmpty=" << buffer.isEmpty() << ", bufferCount=" << buffer.getMetricCount() << std::endl;

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

            // Use registered collector ID if available, otherwise use config ID
            std::string collectorIdForPayload = registeredCollectorId.empty() ?
                gConfig->getCollectorId() : registeredCollectorId;

            json payload = MetricsSerializer::createPayload(
                collectorIdForPayload,
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

    // Load configuration
    if (!gConfig->loadFromFile()) {
        std::cerr << "Failed to load configuration: " << gConfig->getLastError() << std::endl;
        return 1;
    }

    std::cout << "Configuration loaded successfully" << std::endl;
    std::cout << "Collector ID: " << gConfig->getCollectorId() << std::endl;
    std::cout << "Backend URL: " << gConfig->getBackendUrl() << std::endl;

    // Get registration secret from environment variable
    const char* registrationSecret = std::getenv("REGISTRATION_SECRET");
    if (!registrationSecret || std::string(registrationSecret).empty()) {
        std::cerr << "Error: REGISTRATION_SECRET environment variable not set" << std::endl;
        std::cerr << "Please set it before running registration" << std::endl;
        return 1;
    }

    // Load TLS configuration
    auto tlsConfig = gConfig->getTLSConfig();

    // Initialize sender for registration
    Sender sender(
        gConfig->getBackendUrl(),
        gConfig->getCollectorId(),
        tlsConfig.certFile,
        tlsConfig.keyFile,
        tlsConfig.verify
    );

    // Attempt registration
    std::string authToken;
    std::string registeredCollectorId;
    std::string collectorName = gConfig->getHostname();
    std::cout << "Registering with backend as '" << collectorName << "'..." << std::endl;

    if (!sender.registerCollector(registrationSecret, collectorName, authToken, registeredCollectorId)) {
        std::cerr << "Registration failed" << std::endl;
        return 1;
    }

    std::cout << "Registration successful!" << std::endl;
    std::cout << "Auth Token: " << authToken.substr(0, 20) << "..." << std::endl;
    std::cout << "Collector ID: " << gConfig->getCollectorId() << std::endl;

    // Save auth token to file for later use
    std::string tokenFilePath = "/etc/pganalytics/collector.token";
    std::ofstream tokenFile(tokenFilePath);
    if (tokenFile.is_open()) {
        tokenFile << authToken;
        tokenFile.close();
        std::cout << "Auth token saved to " << tokenFilePath << std::endl;
        // Make it readable only by pganalytics user
        chmod(tokenFilePath.c_str(), 0600);
    } else {
        std::cerr << "Warning: Could not save auth token to file" << std::endl;
    }

    // Save registered collector ID to file for later use in metrics push
    if (!registeredCollectorId.empty()) {
        std::string collectorIdFilePath = "/etc/pganalytics/collector.id";
        std::ofstream collectorIdFile(collectorIdFilePath);
        if (collectorIdFile.is_open()) {
            collectorIdFile << registeredCollectorId;
            collectorIdFile.close();
            std::cout << "Collector ID saved to " << collectorIdFilePath << std::endl;
            // Make it readable only by pganalytics user
            chmod(collectorIdFilePath.c_str(), 0600);
        } else {
            std::cerr << "Warning: Could not save collector ID to file" << std::endl;
        }
    } else {
        std::cerr << "Warning: No collector ID received from backend" << std::endl;
    }

    std::cout << "You can now run the collector in normal mode" << std::endl;

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
