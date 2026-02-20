#pragma once

#include <string>
#include <vector>
#include <map>
#include <memory>
#include <thread>
#include <mutex>
#include <atomic>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * Mock Backend Server for Integration Testing
 * Simulates pgAnalytics backend API endpoints with HTTP/HTTPS
 *
 * Features:
 * - POST /api/v1/metrics/push - Accept gzipped JSON metrics
 * - JWT token validation in Authorization header
 * - mTLS certificate validation
 * - Configurable response scenarios (200, 201, 400, 401, 500)
 * - Request tracking for assertions in tests
 * - Self-signed TLS support
 */
class MockBackendServer {
public:
    /**
     * Constructor
     * @param port Port to listen on (default: 8443 for HTTPS)
     * @param useTls Enable TLS/HTTPS (default: true)
     */
    explicit MockBackendServer(int port = 8443, bool useTls = true);

    /**
     * Destructor - automatically stops server if running
     */
    ~MockBackendServer();

    /**
     * Start the HTTP server in background thread
     * @return true if server started successfully
     */
    bool start();

    /**
     * Stop the HTTP server
     * @return true if server stopped successfully
     */
    bool stop();

    /**
     * Check if server is currently running
     */
    bool isRunning() const { return is_running_; }

    // ============= Configuration Methods =============

    /**
     * Set the next HTTP response status code
     * @param status HTTP status (200, 201, 400, 401, 500, etc.)
     */
    void setNextResponseStatus(int status);

    /**
     * Enable/disable JWT token validation
     * @param valid If true, accepts any JWT; if false, rejects with 401
     */
    void setTokenValid(bool valid);

    /**
     * Set a delay before responding to simulate network latency
     * @param milliseconds Delay in milliseconds
     */
    void setResponseDelay(int milliseconds);

    /**
     * Configure server to reject metrics with specific error
     * @param error Error message to return in 400 response
     */
    void setRejectMetricsWithError(const std::string& error);

    /**
     * Reset server to default state
     */
    void reset();

    // ============= Assertion Helper Methods =============

    /**
     * Get count of metrics payloads received
     */
    int getReceivedMetricsCount() const;

    /**
     * Get the last received metrics payload
     */
    json getLastReceivedMetrics() const;

    /**
     * Get all received metrics payloads
     */
    std::vector<json> getAllReceivedMetrics() const;

    /**
     * Check if token was refreshed (new token received after 401)
     */
    bool wasTokenRefreshed() const;

    /**
     * Get all tokens that were sent to server
     */
    std::vector<std::string> getAllReceivedTokens() const;

    /**
     * Get last error received by server
     */
    std::string getLastError() const;

    /**
     * Get request count
     */
    int getRequestCount() const;

    /**
     * Get last HTTP status sent to client
     */
    int getLastResponseStatus() const;

    /**
     * Check if specific endpoint was accessed
     * @param endpoint Path like "/api/v1/metrics/push"
     */
    bool wasEndpointAccessed(const std::string& endpoint) const;

    /**
     * Get the Authorization header from last request
     */
    std::string getLastAuthorizationHeader() const;

    /**
     * Verify gzip decompression worked
     */
    bool wasLastPayloadGzipped() const;

    // ============= Server Information =============

    /**
     * Get server base URL (e.g., "https://localhost:8443")
     */
    std::string getBaseUrl() const;

    /**
     * Get port number
     */
    int getPort() const { return port_; }

private:
    int port_;
    bool use_tls_;
    std::atomic<bool> is_running_{false};
    std::unique_ptr<std::thread> server_thread_;

    // Response configuration
    int next_response_status_ = 200;
    bool token_valid_ = true;
    int response_delay_ms_ = 0;
    std::string reject_with_error_;

    // Request tracking
    mutable std::mutex metrics_mutex_;
    std::vector<json> received_metrics_;
    std::vector<std::string> received_tokens_;
    std::string last_error_;
    int request_count_ = 0;
    int last_response_status_ = 200;
    std::string last_authorization_header_;
    bool last_payload_gzipped_ = false;
    std::map<std::string, int> endpoint_access_count_;
    bool token_was_refreshed_ = false;

    /**
     * Server thread function - handles HTTP requests
     */
    void serverLoop();

    /**
     * Handle POST /api/v1/metrics/push request
     * @param gzipped_payload Request body (gzipped)
     * @param auth_header Authorization header value
     * @return Response JSON
     */
    json handleMetricsPush(const std::string& gzipped_payload, const std::string& auth_header);

    /**
     * Decompress gzipped payload
     * @param compressed Gzipped data
     * @return Decompressed string (or empty on error)
     */
    static std::string decompressGzip(const std::string& compressed);

    /**
     * Validate JWT token format
     * @param token JWT token string
     * @return true if token appears valid
     */
    static bool validateJwtFormat(const std::string& token);
};
