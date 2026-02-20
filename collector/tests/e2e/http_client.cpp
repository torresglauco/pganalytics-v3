#include "http_client.h"
#include <curl/curl.h>
#include <iostream>
#include <fstream>
#include <sstream>

// CURL callback for response body
static size_t writeCallback(void* contents, size_t size, size_t nmemb, std::string* userp) {
    userp->append((char*)contents, size * nmemb);
    return size * nmemb;
}

// CURL callback for response headers
static size_t headerCallback(char* buffer, size_t size, size_t nmemb, std::map<std::string, std::string>* userp) {
    std::string header(buffer, size * nmemb);

    // Parse header: "Key: Value\r\n"
    size_t colon_pos = header.find(':');
    if (colon_pos != std::string::npos) {
        std::string key = header.substr(0, colon_pos);
        std::string value = header.substr(colon_pos + 2);

        // Trim trailing whitespace and newlines
        value.erase(value.find_last_not_of(" \r\n") + 1);

        userp->insert({key, value});
    }

    return size * nmemb;
}

E2EHttpClient::E2EHttpClient(
    const std::string& backend_url,
    const std::string& cert_file,
    const std::string& key_file,
    bool verify_ssl
)
    : m_backend_url(backend_url),
      m_cert_file(cert_file),
      m_key_file(key_file),
      m_verify_ssl(verify_ssl),
      m_last_response_code(0),
      m_verbose(false) {
    m_curl_handle = curl_easy_init();
}

E2EHttpClient::~E2EHttpClient() {
    if (m_curl_handle) {
        curl_easy_cleanup((CURL*)m_curl_handle);
    }
}

void E2EHttpClient::setJwtToken(const std::string& token) {
    m_jwt_token = token;
}

void E2EHttpClient::clearJwtToken() {
    m_jwt_token.clear();
}

bool E2EHttpClient::postJson(
    const std::string& endpoint,
    const std::string& json_body,
    std::string& response_body,
    int& response_code
) {
    std::map<std::string, std::string> headers;
    addJsonHeaders(headers);
    addAuthHeaders(headers);

    return performRequest("POST", endpoint, json_body, headers, response_body, response_code, false);
}

bool E2EHttpClient::getJson(
    const std::string& endpoint,
    std::string& response_body,
    int& response_code
) {
    std::map<std::string, std::string> headers;
    addJsonHeaders(headers);
    addAuthHeaders(headers);

    return performRequest("GET", endpoint, "", headers, response_body, response_code, false);
}

bool E2EHttpClient::postGzipJson(
    const std::string& endpoint,
    const std::string& json_body,
    std::string& response_body,
    int& response_code
) {
    std::map<std::string, std::string> headers;
    addJsonHeaders(headers);
    addGzipHeaders(headers);
    addAuthHeaders(headers);

    return performRequest("POST", endpoint, json_body, headers, response_body, response_code, true);
}

bool E2EHttpClient::registerCollector(
    const std::string& collector_name,
    const std::string& hostname,
    std::string& response_body,
    int& response_code
) {
    // Construct JSON request
    std::string json_body = R"({"name":")" + collector_name + R"(","hostname":")" + hostname + R"("})";

    return postJson("/api/v1/collectors/register", json_body, response_body, response_code);
}

bool E2EHttpClient::submitMetrics(
    const std::string& metrics_json,
    bool compress,
    std::string& response_body,
    int& response_code
) {
    if (compress) {
        return postGzipJson("/api/v1/metrics/push", metrics_json, response_body, response_code);
    } else {
        return postJson("/api/v1/metrics/push", metrics_json, response_body, response_code);
    }
}

bool E2EHttpClient::getConfig(
    const std::string& collector_id,
    std::string& config_toml,
    int& response_code
) {
    std::string endpoint = "/api/v1/config/" + collector_id;
    return getJson(endpoint, config_toml, response_code);
}

std::string E2EHttpClient::getLastResponseStatus() const {
    return std::to_string(m_last_response_code);
}

std::string E2EHttpClient::getLastResponseBody() const {
    return m_last_response_body;
}

int E2EHttpClient::getLastResponseCode() const {
    return m_last_response_code;
}

std::map<std::string, std::string> E2EHttpClient::getLastResponseHeaders() const {
    return m_last_response_headers;
}

void E2EHttpClient::setVerbose(bool verbose) {
    m_verbose = verbose;
}

void E2EHttpClient::setLogFile(const std::string& filepath) {
    m_log_file = filepath;
}

bool E2EHttpClient::performRequest(
    const std::string& method,
    const std::string& endpoint,
    const std::string& body,
    const std::map<std::string, std::string>& headers,
    std::string& response_body,
    int& response_code,
    bool gzip_body
) {
    CURL* curl = (CURL*)m_curl_handle;
    if (!curl) {
        return false;
    }

    std::string full_url = m_backend_url + endpoint;

    // Log request
    if (m_verbose) {
        std::cout << "[E2E HTTP] " << method << " " << full_url << std::endl;
    }

    // Set URL
    curl_easy_setopt(curl, CURLOPT_URL, full_url.c_str());

    // Set SSL options
    if (!m_verify_ssl) {
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0L);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, 0L);
    }

    // Set TLS version to 1.3
    curl_easy_setopt(curl, CURLOPT_SSLVERSION, CURL_SSLVERSION_TLSv1_3);

    // Set client certificate if provided
    if (!m_cert_file.empty()) {
        curl_easy_setopt(curl, CURLOPT_SSLCERT, m_cert_file.c_str());
    }

    if (!m_key_file.empty()) {
        curl_easy_setopt(curl, CURLOPT_SSLKEY, m_key_file.c_str());
    }

    // Set HTTP method
    if (method == "POST") {
        curl_easy_setopt(curl, CURLOPT_POST, 1L);
    } else if (method == "GET") {
        curl_easy_setopt(curl, CURLOPT_HTTPGET, 1L);
    } else {
        curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, method.c_str());
    }

    // Set request body
    if (!body.empty()) {
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, body.c_str());
    }

    // Build header list
    struct curl_slist* header_list = nullptr;
    for (const auto& [key, value] : headers) {
        std::string header_str = key + ": " + value;
        header_list = curl_slist_append(header_list, header_str.c_str());
    }

    if (header_list) {
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, header_list);
    }

    // Set response callbacks
    response_body.clear();
    m_last_response_headers.clear();

    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeCallback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response_body);

    curl_easy_setopt(curl, CURLOPT_HEADERFUNCTION, headerCallback);
    curl_easy_setopt(curl, CURLOPT_HEADERDATA, &m_last_response_headers);

    // Set timeout
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 30L);

    // Perform request
    CURLcode res = curl_easy_perform(curl);

    // Get response code
    long http_code = 0;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
    m_last_response_code = http_code;
    response_code = http_code;
    m_last_response_body = response_body;

    // Clean up headers
    if (header_list) {
        curl_slist_free_all(header_list);
    }

    if (m_verbose) {
        std::cout << "[E2E HTTP] Response: " << http_code << std::endl;
        if (!response_body.empty()) {
            std::cout << "[E2E HTTP] Body: " << response_body.substr(0, 200) << "..." << std::endl;
        }
    }

    return res == CURLE_OK && http_code >= 200 && http_code < 300;
}

void E2EHttpClient::addAuthHeaders(std::map<std::string, std::string>& headers) {
    if (!m_jwt_token.empty()) {
        headers["Authorization"] = "Bearer " + m_jwt_token;
    }
}

void E2EHttpClient::addJsonHeaders(std::map<std::string, std::string>& headers) {
    headers["Content-Type"] = "application/json";
    headers["Accept"] = "application/json";
}

void E2EHttpClient::addGzipHeaders(std::map<std::string, std::string>& headers) {
    addJsonHeaders(headers);
    headers["Content-Encoding"] = "gzip";
}

