#include "../include/connection_pool.h"
#include <iostream>
#include <sstream>
#include <algorithm>

#ifdef HAVE_LIBPQ
// Connection pool implementation only available with libpq

// ============================================================================
// PooledConnection Implementation
// ============================================================================

PooledConnection::PooledConnection(PGconn* conn, const std::string& pool_id)
    : conn_(conn),
      pool_id_(pool_id),
      idle_(true),
      created_at_(std::chrono::steady_clock::now()),
      last_activity_(std::chrono::steady_clock::now()) {
}

PooledConnection::~PooledConnection() {
    if (conn_) {
        PQfinish(conn_);
    }
}

bool PooledConnection::isHealthy() const {
    if (!conn_) return false;
    return PQstatus(conn_) == CONNECTION_OK;
}

std::chrono::seconds PooledConnection::getIdleTime() const {
    auto now = std::chrono::steady_clock::now();
    return std::chrono::duration_cast<std::chrono::seconds>(now - last_activity_);
}

// ============================================================================
// ConnectionPool Implementation
// ============================================================================

ConnectionPool::ConnectionPool(
    const std::string& host,
    int port,
    const std::string& user,
    const std::string& password,
    const std::string& dbname,
    size_t min_size,
    size_t max_size
) : host_(host),
    port_(port),
    user_(user),
    password_(password),
    dbname_(dbname),
    min_size_(min_size),
    max_size_(max_size),
    failed_connections_(0),
    created_at_(std::chrono::steady_clock::now()) {

    // Generate pool ID
    pool_id_ = host + ":" + std::to_string(port) + "/" + dbname;

    // Initialize pool
    initialize();
}

ConnectionPool::~ConnectionPool() {
    close();
}

std::string ConnectionPool::buildConnectionString() const {
    std::ostringstream oss;
    oss << "host=" << host_
        << " port=" << port_
        << " user=" << user_
        << " password=" << password_
        << " dbname=" << dbname_
        << " connect_timeout=10"
        << " application_name=pganalytics-collector";
    return oss.str();
}

std::shared_ptr<PooledConnection> ConnectionPool::createConnection() {
    std::string conn_str = buildConnectionString();

    // Create connection
    PGconn* conn = PQconnectdb(conn_str.c_str());

    if (!conn) {
        std::cerr << "[ConnectionPool] Failed to allocate connection" << std::endl;
        failed_connections_++;
        return nullptr;
    }

    if (PQstatus(conn) != CONNECTION_OK) {
        std::cerr << "[ConnectionPool] Connection failed: " << PQerrorMessage(conn) << std::endl;
        PQfinish(conn);
        failed_connections_++;
        return nullptr;
    }

    // Set statement timeout to 5 seconds
    const char* set_timeout_sql = "SET statement_timeout = '5s'";
    PGresult* res = PQexec(conn, set_timeout_sql);

    if (PQresultStatus(res) != PGRES_COMMAND_OK) {
        std::cerr << "[ConnectionPool] Failed to set statement timeout: "
                  << PQerrorMessage(conn) << std::endl;
        PQclear(res);
        PQfinish(conn);
        return nullptr;
    }

    PQclear(res);

    std::cout << "[ConnectionPool] Created new connection to " << pool_id_ << std::endl;
    return std::make_shared<PooledConnection>(conn, pool_id_);
}

void ConnectionPool::initialize() {
    std::unique_lock<std::mutex> lock(mutex_);

    for (size_t i = 0; i < min_size_; i++) {
        auto conn = createConnection();
        if (conn) {
            connections_.push_back(conn);
            available_.push(conn);
        }
    }

    std::cout << "[ConnectionPool] Pool initialized with " << available_.size()
              << " connections (target: " << min_size_ << ")" << std::endl;
}

std::shared_ptr<PooledConnection> ConnectionPool::acquire(int timeout_sec) {
    std::unique_lock<std::mutex> lock(mutex_);

    // Try to get an available connection
    while (available_.empty()) {
        // Can we create a new connection?
        if (connections_.size() < max_size_) {
            auto conn = createConnection();
            if (conn) {
                conn->markActive();
                connections_.push_back(conn);
                return conn;
            }
        }

        // Wait for a connection to be returned
        if (cv_.wait_for(lock, std::chrono::seconds(timeout_sec)) == std::cv_status::timeout) {
            std::cerr << "[ConnectionPool] Timeout waiting for connection" << std::endl;
            return nullptr;
        }
    }

    // Get the first available connection
    auto conn = available_.front();
    available_.pop();

    if (!conn || !conn->isHealthy()) {
        // Connection is dead, try another
        if (conn) {
            connections_.erase(
                std::remove(connections_.begin(), connections_.end(), conn),
                connections_.end()
            );
        }
        return acquire(timeout_sec);  // Retry
    }

    conn->markActive();
    return conn;
}

void ConnectionPool::release(std::shared_ptr<PooledConnection> conn) {
    if (!conn) return;

    std::unique_lock<std::mutex> lock(mutex_);

    conn->markIdle();

    // Only return healthy connections to pool
    if (conn->isHealthy()) {
        available_.push(conn);
        cv_.notify_one();
    } else {
        // Remove unhealthy connection
        connections_.erase(
            std::remove(connections_.begin(), connections_.end(), conn),
            connections_.end()
        );
    }
}

size_t ConnectionPool::getActiveCount() const {
    std::unique_lock<std::mutex> lock(mutex_);
    return connections_.size() - available_.size();
}

size_t ConnectionPool::getPoolSize() const {
    std::unique_lock<std::mutex> lock(mutex_);
    return connections_.size();
}

void ConnectionPool::healthCheck() {
    std::unique_lock<std::mutex> lock(mutex_);

    // Check each idle connection
    std::vector<std::shared_ptr<PooledConnection>> healthy;

    for (auto& conn : connections_) {
        if (conn->isIdle() && conn->isHealthy()) {
            healthy.push_back(conn);
        }
    }

    // Rebuild pool with healthy connections
    connections_ = healthy;

    // Refill to min_size
    while (connections_.size() < min_size_) {
        auto conn = createConnection();
        if (conn) {
            connections_.push_back(conn);
            available_.push(conn);
        } else {
            break;
        }
    }

    std::cout << "[ConnectionPool] Health check complete: " << connections_.size()
              << "/" << max_size_ << " connections healthy" << std::endl;
}

void ConnectionPool::reconnectAll() {
    std::unique_lock<std::mutex> lock(mutex_);

    // Clear all connections
    connections_.clear();
    while (!available_.empty()) {
        available_.pop();
    }

    // Recreate pool
    for (size_t i = 0; i < min_size_; i++) {
        auto conn = createConnection();
        if (conn) {
            connections_.push_back(conn);
            available_.push(conn);
        }
    }

    std::cout << "[ConnectionPool] Reconnected " << available_.size()
              << " connections" << std::endl;
}

void ConnectionPool::close() {
    std::unique_lock<std::mutex> lock(mutex_);

    connections_.clear();
    while (!available_.empty()) {
        available_.pop();
    }

    std::cout << "[ConnectionPool] Closed pool " << pool_id_ << std::endl;
}

ConnectionPool::PoolStats ConnectionPool::getStats() const {
    std::unique_lock<std::mutex> lock(mutex_);

    auto now = std::chrono::steady_clock::now();
    auto uptime = std::chrono::duration_cast<std::chrono::seconds>(now - created_at_);

    return {
        .total_size = connections_.size(),
        .active_count = connections_.size() - available_.size(),
        .idle_count = available_.size(),
        .failed_attempts = failed_connections_,
        .uptime = uptime,
    };
}

#endif  // HAVE_LIBPQ
