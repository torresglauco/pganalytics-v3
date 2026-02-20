#pragma once

#include <string>
#include <memory>
#include <map>

/**
 * E2E HTTP Client
 *
 * Makes HTTPS requests to the real pgAnalytics backend API.
 * Handles:
 * - TLS 1.3 with mTLS certificates
 * - JWT token injection in headers
 * - Request/response logging
 * - Error handling
 */
class E2EHttpClient {
public:
    /**
     * Constructor
     * @param backend_url Backend API URL (e.g., https://localhost:8080)
     * @param cert_file Path to client certificate
     * @param key_file Path to client private key
     * @param verify_ssl Whether to verify server SSL (false for self-signed)
     */
    E2EHttpClient(
        const std::string& backend_url,
        const std::string& cert_file = "",
        const std::string& key_file = "",
        bool verify_ssl = false
    );

    ~E2EHttpClient();

    // Authentication
    void setJwtToken(const std::string& token);
    void clearJwtToken();

    // HTTP Methods
    bool postJson(
        const std::string& endpoint,
        const std::string& json_body,
        std::string& response_body,
        int& response_code
    );

    bool getJson(
        const std::string& endpoint,
        std::string& response_body,
        int& response_code
    );

    bool postGzipJson(
        const std::string& endpoint,
        const std::string& json_body,
        std::string& response_body,
        int& response_code
    );

    // Collector registration
    bool registerCollector(
        const std::string& collector_name,
        const std::string& hostname,
        std::string& response_body,
        int& response_code
    );

    // Metrics submission
    bool submitMetrics(
        const std::string& metrics_json,
        bool compress = true,
        std::string& response_body,
        int& response_code
    );

    // Configuration retrieval
    bool getConfig(
        const std::string& collector_id,
        std::string& config_toml,
        int& response_code
    );

    // Response helpers
    std::string getLastResponseStatus() const;
    std::string getLastResponseBody() const;
    int getLastResponseCode() const;
    std::map<std::string, std::string> getLastResponseHeaders() const;

    // Logging
    void setVerbose(bool verbose);
    void setLogFile(const std::string& filepath);

private:
    // Internal HTTPS request
    bool performRequest(
        const std::string& method,
        const std::string& endpoint,
        const std::string& body,
        const std::map<std::string, std::string>& headers,
        std::string& response_body,
        int& response_code,
        bool gzip_body = false
    );

    // Header management
    void addAuthHeaders(std::map<std::string, std::string>& headers);
    void addJsonHeaders(std::map<std::string, std::string>& headers);
    void addGzipHeaders(std::map<std::string, std::string>& headers);

    // Member variables
    std::string m_backend_url;
    std::string m_cert_file;
    std::string m_key_file;
    bool m_verify_ssl;
    std::string m_jwt_token;
    std::string m_last_response_body;
    int m_last_response_code;
    std::map<std::string, std::string> m_last_response_headers;
    bool m_verbose;
    std::string m_log_file;

    // CURL handle
    void* m_curl_handle;
};

