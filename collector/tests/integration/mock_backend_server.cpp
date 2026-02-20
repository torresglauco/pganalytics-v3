#include "mock_backend_server.h"
#include <iostream>
#include <sstream>
#include <algorithm>
#include <zlib.h>
#include <cstring>
#include <chrono>
#include <thread>

// Simple HTTP server implementation using sockets
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <fcntl.h>

MockBackendServer::MockBackendServer(int port, bool useTls)
    : port_(port), use_tls_(useTls) {
}

MockBackendServer::~MockBackendServer() {
    if (is_running_) {
        stop();
    }
}

bool MockBackendServer::start() {
    if (is_running_) {
        return false;
    }

    is_running_ = true;
    server_thread_ = std::make_unique<std::thread>([this]() { serverLoop(); });

    // Give server time to start
    std::this_thread::sleep_for(std::chrono::milliseconds(100));

    return true;
}

bool MockBackendServer::stop() {
    if (!is_running_) {
        return false;
    }

    is_running_ = false;

    if (server_thread_ && server_thread_->joinable()) {
        server_thread_->join();
    }

    return true;
}

void MockBackendServer::serverLoop() {
    // Create socket
    int server_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (server_socket < 0) {
        std::cerr << "Failed to create socket" << std::endl;
        return;
    }

    // Allow socket reuse
    int opt = 1;
    if (setsockopt(server_socket, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt)) < 0) {
        std::cerr << "setsockopt failed" << std::endl;
        close(server_socket);
        return;
    }

    // Bind socket
    struct sockaddr_in server_addr;
    server_addr.sin_family = AF_INET;
    server_addr.sin_addr.s_addr = inet_addr("127.0.0.1");
    server_addr.sin_port = htons(port_);

    if (bind(server_socket, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        std::cerr << "Bind failed on port " << port_ << std::endl;
        close(server_socket);
        return;
    }

    // Listen for connections
    if (listen(server_socket, 5) < 0) {
        std::cerr << "Listen failed" << std::endl;
        close(server_socket);
        return;
    }

    // Set socket to non-blocking for timeout
    fcntl(server_socket, F_SETFL, O_NONBLOCK);

    while (is_running_) {
        struct sockaddr_in client_addr;
        socklen_t client_addr_len = sizeof(client_addr);

        int client_socket = accept(server_socket, (struct sockaddr*)&client_addr, &client_addr_len);

        if (client_socket < 0) {
            // No connection available, sleep briefly
            std::this_thread::sleep_for(std::chrono::milliseconds(10));
            continue;
        }

        // Read HTTP request
        char buffer[65536] = {0};
        ssize_t bytes_read = read(client_socket, buffer, sizeof(buffer) - 1);

        if (bytes_read > 0) {
            buffer[bytes_read] = '\0';
            std::string request(buffer);

            // Parse HTTP request
            std::istringstream iss(request);
            std::string method, path, http_version;
            iss >> method >> path >> http_version;

            // Track request
            {
                std::lock_guard<std::mutex> lock(metrics_mutex_);
                request_count_++;
            }

            // Extract headers and body
            std::string auth_header;
            size_t headers_end = request.find("\r\n\r\n");
            std::string headers_section = request.substr(0, headers_end);
            std::string body = request.substr(headers_end + 4);

            // Parse Authorization header
            size_t auth_pos = headers_section.find("Authorization: ");
            if (auth_pos != std::string::npos) {
                size_t auth_start = auth_pos + 15;
                size_t auth_end = headers_section.find("\r\n", auth_start);
                auth_header = headers_section.substr(auth_start, auth_end - auth_start);
            }

            // Track authorization header
            {
                std::lock_guard<std::mutex> lock(metrics_mutex_);
                last_authorization_header_ = auth_header;
                if (!auth_header.empty()) {
                    received_tokens_.push_back(auth_header);
                }
            }

            // Handle metrics push endpoint
            json response;
            int status = 200;

            if (path == "/api/v1/metrics/push" && method == "POST") {
                // Track endpoint access
                {
                    std::lock_guard<std::mutex> lock(metrics_mutex_);
                    endpoint_access_count_[path]++;
                }

                // Apply response delay if configured
                if (response_delay_ms_ > 0) {
                    std::this_thread::sleep_for(std::chrono::milliseconds(response_delay_ms_));
                }

                // Check token validity
                if (!token_valid_ && !auth_header.empty()) {
                    status = 401;
                    response["error"] = "Unauthorized";
                } else if (!reject_with_error_.empty()) {
                    status = 400;
                    response["error"] = reject_with_error_;
                } else if (next_response_status_ != 200) {
                    status = next_response_status_;
                    response["error"] = "Server error";
                    next_response_status_ = 200;  // Reset after use
                } else {
                    // Try to decompress and parse metrics
                    std::string decompressed = decompressGzip(body);

                    if (decompressed.empty() && !body.empty()) {
                        // Decompression failed
                        status = 400;
                        response["error"] = "Failed to decompress gzip payload";
                    } else {
                        // Parse JSON
                        try {
                            json metrics_json;
                            if (!decompressed.empty()) {
                                metrics_json = json::parse(decompressed);
                            } else if (!body.empty()) {
                                metrics_json = json::parse(body);
                            } else {
                                throw std::exception();
                            }

                            // Store received metrics
                            {
                                std::lock_guard<std::mutex> lock(metrics_mutex_);
                                received_metrics_.push_back(metrics_json);
                                last_payload_gzipped_ = !decompressed.empty();
                            }

                            status = 200;
                            response["status"] = "success";
                            response["metrics_inserted"] = 100;
                            response["collector_id"] = metrics_json.value("collector_id", "unknown");
                        } catch (const std::exception& e) {
                            status = 400;
                            response["error"] = "Invalid JSON in metrics payload";
                        }
                    }
                }
            } else if (path == "/api/v1/collectors/register" && method == "POST") {
                // Registration endpoint
                {
                    std::lock_guard<std::mutex> lock(metrics_mutex_);
                    endpoint_access_count_[path]++;
                }

                status = 200;
                response["collector_id"] = "test-collector-001";
                response["token"] = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test";
                response["certificate"] = "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----";
                response["private_key"] = "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----";
            } else if (path.find("/api/v1/config/") == 0 && method == "GET") {
                // Config pull endpoint
                {
                    std::lock_guard<std::mutex> lock(metrics_mutex_);
                    endpoint_access_count_[path]++;
                }

                status = 200;
                response["collector_id"] = "test-collector-001";
                response["backend_url"] = "https://localhost:8443";
                response["push_interval"] = 60;
                response["config_version"] = 1;
            } else {
                status = 404;
                response["error"] = "Endpoint not found";
            }

            // Update last response status
            {
                std::lock_guard<std::mutex> lock(metrics_mutex_);
                last_response_status_ = status;
            }

            // Send HTTP response
            std::string response_body = response.dump();
            std::ostringstream response_stream;
            response_stream << "HTTP/1.1 " << status << " OK\r\n";
            response_stream << "Content-Type: application/json\r\n";
            response_stream << "Content-Length: " << response_body.length() << "\r\n";
            response_stream << "Connection: close\r\n";
            response_stream << "\r\n";
            response_stream << response_body;

            std::string http_response = response_stream.str();
            write(client_socket, http_response.c_str(), http_response.length());
        }

        close(client_socket);
    }

    close(server_socket);
}

void MockBackendServer::setNextResponseStatus(int status) {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    next_response_status_ = status;
}

void MockBackendServer::setTokenValid(bool valid) {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    token_valid_ = valid;
}

void MockBackendServer::setResponseDelay(int milliseconds) {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    response_delay_ms_ = milliseconds;
}

void MockBackendServer::setRejectMetricsWithError(const std::string& error) {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    reject_with_error_ = error;
}

void MockBackendServer::reset() {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    next_response_status_ = 200;
    token_valid_ = true;
    response_delay_ms_ = 0;
    reject_with_error_.clear();
    received_metrics_.clear();
    received_tokens_.clear();
    last_error_.clear();
    request_count_ = 0;
    last_response_status_ = 200;
    last_authorization_header_.clear();
    last_payload_gzipped_ = false;
    endpoint_access_count_.clear();
    token_was_refreshed_ = false;
}

int MockBackendServer::getReceivedMetricsCount() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return static_cast<int>(received_metrics_.size());
}

json MockBackendServer::getLastReceivedMetrics() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    if (received_metrics_.empty()) {
        return json();
    }
    return received_metrics_.back();
}

std::vector<json> MockBackendServer::getAllReceivedMetrics() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return received_metrics_;
}

bool MockBackendServer::wasTokenRefreshed() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return token_was_refreshed_;
}

std::vector<std::string> MockBackendServer::getAllReceivedTokens() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return received_tokens_;
}

std::string MockBackendServer::getLastError() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return last_error_;
}

int MockBackendServer::getRequestCount() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return request_count_;
}

int MockBackendServer::getLastResponseStatus() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return last_response_status_;
}

bool MockBackendServer::wasEndpointAccessed(const std::string& endpoint) const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return endpoint_access_count_.find(endpoint) != endpoint_access_count_.end();
}

std::string MockBackendServer::getLastAuthorizationHeader() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return last_authorization_header_;
}

bool MockBackendServer::wasLastPayloadGzipped() const {
    std::lock_guard<std::mutex> lock(metrics_mutex_);
    return last_payload_gzipped_;
}

std::string MockBackendServer::getBaseUrl() const {
    std::string protocol = use_tls_ ? "https" : "http";
    return protocol + "://127.0.0.1:" + std::to_string(port_);
}

std::string MockBackendServer::decompressGzip(const std::string& compressed) {
    if (compressed.empty()) {
        return "";
    }

    // Check for gzip magic number
    if (compressed.size() < 2 || (unsigned char)compressed[0] != 0x1f ||
        (unsigned char)compressed[1] != 0x8b) {
        // Not gzipped, return as-is
        return compressed;
    }

    z_stream stream = {};
    stream.avail_in = compressed.size();
    stream.next_in = (unsigned char*)const_cast<char*>(compressed.data());

    if (inflateInit2(&stream, 16 + MAX_WBITS) != Z_OK) {
        return "";
    }

    std::string decompressed;
    char out_buffer[4096];

    int ret = Z_OK;
    while (ret != Z_STREAM_END) {
        stream.avail_out = sizeof(out_buffer);
        stream.next_out = (unsigned char*)out_buffer;

        ret = inflate(&stream, Z_NO_FLUSH);

        if (ret != Z_OK && ret != Z_STREAM_END) {
            inflateEnd(&stream);
            return "";
        }

        size_t produced = sizeof(out_buffer) - stream.avail_out;
        decompressed.append(out_buffer, produced);
    }

    inflateEnd(&stream);
    return decompressed;
}

bool MockBackendServer::validateJwtFormat(const std::string& token) {
    // Simple JWT validation - should have 3 parts separated by dots
    if (token.empty() || token.find("Bearer ") == 0) {
        return false;
    }

    int dot_count = std::count(token.begin(), token.end(), '.');
    return dot_count == 2;
}
