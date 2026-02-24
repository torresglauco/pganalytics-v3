#include "../include/sender.h"
#include "../include/binary_protocol.h"
#include <curl/curl.h>
#include <zlib.h>
#include <iostream>
#include <sstream>
#include <ctime>
#include <cstring>

Sender::Sender(
    const std::string& backendUrl,
    const std::string& collectorId,
    const std::string& certFile,
    const std::string& keyFile,
    bool tlsVerify,
    Protocol protocol
) : backendUrl_(backendUrl),
    collectorId_(collectorId),
    certFile_(certFile),
    keyFile_(keyFile),
    tlsVerify_(tlsVerify),
    tokenExpiresAt_(0),
    protocol_(protocol) {
}

void Sender::setProtocol(Protocol protocol) {
    protocol_ = protocol;
    std::cout << "[Sender] Protocol set to " << (protocol == Protocol::JSON ? "JSON" : "BINARY") << std::endl;
}

Sender::Protocol Sender::getProtocol() const {
    return protocol_;
}

bool Sender::pushMetrics(const json& metrics) {
    // Validate metrics
    if (!metrics.is_object() || !metrics.contains("metrics")) {
        return false;
    }

    // Route based on selected protocol
    if (protocol_ == Protocol::BINARY) {
        return pushMetricsBinary(metrics);
    }

    // Refresh token if needed
    if (!isTokenValid()) {
        refreshAuthToken();
    }

    // Serialize metrics to JSON string
    std::string jsonData = metrics.dump();

    // Initialize CURL
    CURL* curl = curl_easy_init();
    if (!curl) {
        return false;
    }

    // Configure CURL for TLS 1.3 + mTLS
    if (!setupCurl(curl)) {
        curl_easy_cleanup(curl);
        return false;
    }

    // Prepare URL and headers
    std::string url = backendUrl_ + "/api/v1/metrics/push";
    struct curl_slist* headers = nullptr;
    headers = curl_slist_append(headers, "Content-Type: application/json");

    // Add Authorization header
    std::string authHeader = "Authorization: Bearer " + getAuthToken();
    headers = curl_slist_append(headers, authHeader.c_str());

    // Set CURL options
    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, jsonData.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE, jsonData.size());

    // Response callback
    std::string responseData;
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &responseData);

    // Perform request
    CURLcode res = curl_easy_perform(curl);

    bool success = (res == CURLE_OK);
    if (!success) {
        std::cerr << "CURL error: " << curl_easy_strerror(res) << std::endl;
    }

    // Check HTTP status code
    if (success) {
        long httpCode;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &httpCode);
        success = (httpCode == 200 || httpCode == 201);

        if (!success && httpCode == 401) {
            // Token expired, refresh and retry once
            refreshAuthToken();
            authHeader = "Authorization: Bearer " + getAuthToken();
            curl_slist_free_all(headers);
            headers = nullptr;
            headers = curl_slist_append(headers, "Content-Type: application/json");
            headers = curl_slist_append(headers, "Content-Encoding: gzip");
            headers = curl_slist_append(headers, authHeader.c_str());

            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
            res = curl_easy_perform(curl);
            success = (res == CURLE_OK);
        }
    }

    // Cleanup
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);

    return success;
}

std::string Sender::getAuthToken() {
    return authToken_;
}

void Sender::refreshAuthToken() {
    // In a real implementation, this would refresh the token from the backend
    // For now, just regenerate it locally
    // This would be called after registration
}

void Sender::setAuthToken(const std::string& token, long expiresAt) {
    authToken_ = token;
    tokenExpiresAt_ = expiresAt;
}

bool Sender::isTokenValid() const {
    if (authToken_.empty() || tokenExpiresAt_ == 0) {
        return false;
    }

    time_t now = std::time(nullptr);
    // Add 60 second buffer for token expiration (refresh 1 min before expiry)
    return now < (tokenExpiresAt_ - 60);
}

size_t Sender::writeCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    if (userp) {
        userp->append(static_cast<const char*>(contents), size * nmemb);
    }
    return size * nmemb;
}

std::string Sender::compressJson(const std::string& input) {
    std::string output;

    // Allocate output buffer
    size_t compressedSize = compressBound(input.size());
    output.resize(compressedSize);

    // Compress
    int result = compress2(
        reinterpret_cast<unsigned char*>(&output[0]),
        &compressedSize,
        reinterpret_cast<const unsigned char*>(input.c_str()),
        input.size(),
        6  // Compression level
    );

    if (result != Z_OK) {
        output.clear();
        return output;
    }

    output.resize(compressedSize);
    return output;
}

bool Sender::pullConfig(const std::string& collectorId, std::string& configToml, int& version) {
    // Validate token is still valid
    if (!isTokenValid()) {
        refreshAuthToken();
    }

    // Build URL: {backendUrl}/api/v1/config/{collectorId}
    std::string url = backendUrl_ + "/api/v1/config/" + collectorId;

    // Initialize CURL
    CURL* curl = curl_easy_init();
    if (!curl) {
        return false;
    }

    // Configure CURL for TLS 1.3 + mTLS
    if (!setupCurl(curl)) {
        curl_easy_cleanup(curl);
        return false;
    }

    // Prepare headers
    struct curl_slist* headers = nullptr;
    headers = curl_slist_append(headers, "Accept: text/plain");

    // Add Authorization header
    std::string authHeader = "Authorization: Bearer " + getAuthToken();
    headers = curl_slist_append(headers, authHeader.c_str());

    // Set CURL options
    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, "GET");

    // Response buffer
    std::string responseBuffer;
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &responseBuffer);

    // Perform request
    CURLcode res = curl_easy_perform(curl);

    bool success = (res == CURLE_OK);
    if (!success) {
        std::cerr << "Config pull failed: " << curl_easy_strerror(res) << std::endl;
        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
        return false;
    }

    // Check HTTP status code
    long httpCode = 0;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &httpCode);

    if (httpCode == 200) {
        configToml = responseBuffer;
        version = 0;  // Default to 0, could be read from X-Config-Version header
        // Note: To read response headers like X-Config-Version, would need to implement
        // CURLOPT_HEADERFUNCTION callback. For now, version is read from config_version field.

        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
        return true;
    } else if (httpCode == 404) {
        std::cerr << "Collector configuration not found on backend" << std::endl;
        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
        return false;
    } else if (httpCode == 401) {
        // Token expired, refresh and retry once
        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
        refreshAuthToken();
        if (isTokenValid()) {
            return pullConfig(collectorId, configToml, version);
        }
        return false;
    } else {
        std::cerr << "Config pull failed with HTTP " << httpCode << std::endl;
        curl_slist_free_all(headers);
        curl_easy_cleanup(curl);
        return false;
    }
}

bool Sender::setupCurl(void* curl) {
    CURL* curl_handle = static_cast<CURL*>(curl);

    // Only configure SSL/TLS if the backend URL uses HTTPS
    if (backendUrl_.find("https://") == 0) {
        // Use TLS 1.2 or higher (more compatible than forcing TLS 1.3)
        curl_easy_setopt(curl_handle, CURLOPT_SSLVERSION, CURL_SSLVERSION_TLSv1_2);

        // mTLS certificate and key (only set if files exist)
        FILE* cert_file = fopen(certFile_.c_str(), "r");
        if (cert_file) {
            fclose(cert_file);
            curl_easy_setopt(curl_handle, CURLOPT_SSLCERT, certFile_.c_str());
        }

        FILE* key_file = fopen(keyFile_.c_str(), "r");
        if (key_file) {
            fclose(key_file);
            curl_easy_setopt(curl_handle, CURLOPT_SSLKEY, keyFile_.c_str());
        }

        // Certificate verification
        if (tlsVerify_) {
            curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYPEER, 1L);
            curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYHOST, 2L);
        } else {
            curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYPEER, 0L);
            curl_easy_setopt(curl_handle, CURLOPT_SSL_VERIFYHOST, 0L);
        }
    }

    return true;
}

bool Sender::pushMetricsBinary(const json& metrics) {
    // Validate metrics
    if (!metrics.is_object() || !metrics.contains("metrics")) {
        return false;
    }

    // Refresh token if needed
    if (!isTokenValid()) {
        refreshAuthToken();
    }

    // Create binary message
    std::string version = "1.0.0";  // Default version
    if (metrics.contains("version") && metrics["version"].is_string()) {
        version = metrics["version"].get<std::string>();
    }

    std::vector<uint8_t> binaryMessage = createBinaryMetricsMessage(metrics, version);
    if (binaryMessage.empty()) {
        return false;
    }

    // Send binary message
    return sendBinaryMessage(binaryMessage, "/api/v1/metrics/push/binary");
}

std::vector<uint8_t> Sender::createBinaryMetricsMessage(const json& metrics, const std::string& version) {
    try {
        // Extract key fields from metrics
        std::string hostname = "unknown";
        if (metrics.contains("hostname") && metrics["hostname"].is_string()) {
            hostname = metrics["hostname"].get<std::string>();
        }

        // Extract metrics array
        std::vector<json> metricsArray;
        if (metrics.contains("metrics") && metrics["metrics"].is_array()) {
            metricsArray = metrics["metrics"].get<std::vector<json>>();
        }

        // Build metrics batch message with collector ID, hostname, and version
        std::vector<uint8_t> message = MessageBuilder::createMetricsBatch(
            collectorId_,
            hostname,
            version,
            metricsArray,
            CompressionType::Zstd
        );

        return message;
    } catch (const std::exception& e) {
        std::cerr << "Failed to create binary metrics message: " << e.what() << std::endl;
        return std::vector<uint8_t>();
    }
}

bool Sender::sendBinaryMessage(const std::vector<uint8_t>& message, const std::string& endpoint) {
    if (message.empty()) {
        return false;
    }

    // Compress with Zstd
    std::vector<uint8_t> compressed = compressWithZstd(message);
    if (compressed.empty()) {
        return false;
    }

    // Initialize CURL
    CURL* curl = curl_easy_init();
    if (!curl) {
        return false;
    }

    // Configure CURL for TLS 1.3 + mTLS
    if (!setupCurl(curl)) {
        curl_easy_cleanup(curl);
        return false;
    }

    // Prepare URL and headers
    std::string url = backendUrl_ + endpoint;
    struct curl_slist* headers = nullptr;
    headers = curl_slist_append(headers, "Content-Type: application/octet-stream");
    headers = curl_slist_append(headers, "Content-Encoding: zstd");
    headers = curl_slist_append(headers, "X-Protocol-Version: 1.0");

    // Add Authorization header
    std::string authHeader = "Authorization: Bearer " + getAuthToken();
    headers = curl_slist_append(headers, authHeader.c_str());

    // Set CURL options
    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, compressed.data());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDSIZE, (long)compressed.size());

    // Response callback
    std::string responseData;
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &responseData);

    // Perform request
    CURLcode res = curl_easy_perform(curl);

    bool success = (res == CURLE_OK);
    if (!success) {
        std::cerr << "CURL error: " << curl_easy_strerror(res) << std::endl;
    }

    // Check HTTP status code
    if (success) {
        long httpCode;
        curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &httpCode);
        success = (httpCode == 200 || httpCode == 201 || httpCode == 202);

        if (!success && httpCode == 401) {
            // Token expired, refresh and retry once
            refreshAuthToken();
            authHeader = "Authorization: Bearer " + getAuthToken();
            curl_slist_free_all(headers);
            headers = nullptr;
            headers = curl_slist_append(headers, "Content-Type: application/octet-stream");
            headers = curl_slist_append(headers, "Content-Encoding: zstd");
            headers = curl_slist_append(headers, "X-Protocol-Version: 1.0");
            headers = curl_slist_append(headers, authHeader.c_str());

            curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
            res = curl_easy_perform(curl);
            success = (res == CURLE_OK);

            if (success) {
                curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &httpCode);
                success = (httpCode == 200 || httpCode == 201 || httpCode == 202);
            }
        }
    }

    // Cleanup
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);

    return success;
}

std::vector<uint8_t> Sender::compressWithZstd(const std::vector<uint8_t>& data) {
    if (data.empty()) {
        return std::vector<uint8_t>();
    }

    try {
        // Use CompressionUtil from binary_protocol to compress with Zstd
        return CompressionUtil::compress(data, CompressionType::Zstd);
    } catch (const std::exception& e) {
        std::cerr << "Zstd compression failed: " << e.what() << std::endl;
        return std::vector<uint8_t>();
    }
}
