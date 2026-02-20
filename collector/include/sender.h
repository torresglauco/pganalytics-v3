#pragma once

#include <string>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * HTTP Sender for metrics transmission
 * Handles TLS 1.3, mTLS, JWT authentication, gzip compression
 */
class Sender {
public:
    /**
     * Initialize sender with backend configuration
     * @param backendUrl Base URL of backend API (https://...)
     * @param collectorId Unique collector identifier
     * @param certFile Path to client certificate (mTLS)
     * @param keyFile Path to client key (mTLS)
     * @param tlsVerify Whether to verify TLS certificates (false for demo)
     */
    Sender(
        const std::string& backendUrl,
        const std::string& collectorId,
        const std::string& certFile,
        const std::string& keyFile,
        bool tlsVerify = true
    );

    /**
     * Push metrics to backend
     * @param metrics JSON object containing metrics
     * @return true if successful, false otherwise
     */
    bool pushMetrics(const json& metrics);

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

private:
    std::string backendUrl_;
    std::string collectorId_;
    std::string certFile_;
    std::string keyFile_;
    bool tlsVerify_;
    std::string authToken_;
    long tokenExpiresAt_;

    static size_t writeCallback(void* contents, size_t size, size_t nmemb, std::string* userp);
    std::string compressJson(const std::string& input);
    std::string generateJwt();
    bool setupCurl(void* curl);
};
