#pragma once

#include <string>
#include <vector>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * HTTP Sender for metrics transmission
 * Supports both JSON (REST) and binary protocol transmission
 * Handles TLS 1.3, mTLS, JWT authentication, compression
 */
class Sender {
public:
    /**
     * Protocol used for transmission
     */
    enum class Protocol {
        JSON = 0,      // REST with JSON + gzip (original)
        BINARY = 1,    // Custom binary protocol + zstd (optimized)
    };

    /**
     * Initialize sender with backend configuration
     * @param backendUrl Base URL of backend API (https://...)
     * @param collectorId Unique collector identifier
     * @param certFile Path to client certificate (mTLS)
     * @param keyFile Path to client key (mTLS)
     * @param tlsVerify Whether to verify TLS certificates (false for demo)
     * @param protocol Protocol to use (default: JSON for backward compatibility)
     */
    Sender(
        const std::string& backendUrl,
        const std::string& collectorId,
        const std::string& certFile,
        const std::string& keyFile,
        bool tlsVerify = true,
        Protocol protocol = Protocol::JSON
    );

    /**
     * Set protocol for transmission
     * @param protocol Protocol to use
     */
    void setProtocol(Protocol protocol);

    /**
     * Get current protocol
     */
    Protocol getProtocol() const;

    /**
     * Push metrics to backend
     * Uses configured protocol (JSON or binary)
     * @param metrics JSON object containing metrics
     * @return true if successful, false otherwise
     */
    bool pushMetrics(const json& metrics);

    /**
     * Push metrics using binary protocol (optimized)
     * @param metrics JSON object containing metrics
     * @return true if successful, false otherwise
     */
    bool pushMetricsBinary(const json& metrics);

    /**
     * Get JWT token for authentication
     * @return JWT token string
     */
    std::string getAuthToken();

    /**
     * Refresh JWT token
     */
    void refreshAuthToken();

    /**
     * Set JWT token (for testing or external token management)
     * @param token JWT token string
     * @param expiresAt Unix timestamp when token expires
     */
    void setAuthToken(const std::string& token, long expiresAt);

    /**
     * Check if current token is still valid
     */
    bool isTokenValid() const;

    /**
     * Pull configuration from backend
     * @param collectorId Collector ID
     * @param configToml Output: TOML configuration content
     * @param version Output: Configuration version number
     * @return true if successful, false otherwise
     */
    bool pullConfig(
        const std::string& collectorId,
        std::string& configToml,
        int& version
    );

    /**
     * Register collector with backend
     * @param registrationSecret Secret token for registration
     * @param collectorName Human-readable name for the collector
     * @param authToken Output: JWT token received from backend
     * @param collectorId Output: Collector UUID assigned by backend
     * @return true if successful, false otherwise
     */
    bool registerCollector(
        const std::string& registrationSecret,
        const std::string& collectorName,
        std::string& authToken,
        std::string& collectorId
    );

private:
    std::string backendUrl_;
    std::string collectorId_;
    std::string certFile_;
    std::string keyFile_;
    bool tlsVerify_;
    std::string authToken_;
    long tokenExpiresAt_;
    Protocol protocol_;

    // Helper methods
    static size_t writeCallback(void* contents, size_t size, size_t nmemb, std::string* userp);
    std::string compressJson(const std::string& input);
    std::string generateJwt();
    bool setupCurl(void* curl);

    // Binary protocol helpers
    bool sendBinaryMessage(const std::vector<uint8_t>& message, const std::string& endpoint);
    std::vector<uint8_t> createBinaryMetricsMessage(
        const json& metrics,
        const std::string& version
    );

    // Compression helper
    std::vector<uint8_t> compressWithZstd(const std::vector<uint8_t>& data);
};
