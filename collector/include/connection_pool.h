#ifndef CONNECTION_POOL_H_
#define CONNECTION_POOL_H_

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>
#else
// Forward declare PGconn if libpq is not available
typedef void PGconn;
#endif
#include <vector>
#include <queue>
#include <mutex>
#include <condition_variable>
#include <chrono>
#include <memory>
#include <string>

/**
 * PostgreSQL Connection Pool
 *
 * Manages a pool of reusable PostgreSQL connections to reduce:
 * - Connection establishment overhead
 * - Memory fragmentation
 * - CPU usage from repeated authentication
 *
 * Features:
 * - Configurable min/max connections
 * - Automatic health checks
 * - Exponential backoff on connection failures
 * - Thread-safe concurrent access
 * - Connection timeout enforcement
 */

class PooledConnection {
public:
    PooledConnection(PGconn* conn, const std::string& pool_id);
    ~PooledConnection();

    PGconn* getConn() const { return conn_; }
    bool isHealthy() const;
    bool isIdle() const { return idle_; }
    void markActive() { idle_ = false; last_activity_ = std::chrono::steady_clock::now(); }
    void markIdle() { idle_ = true; last_activity_ = std::chrono::steady_clock::now(); }
    std::chrono::seconds getIdleTime() const;

private:
    PGconn* conn_;
    std::string pool_id_;
    bool idle_;
    std::chrono::steady_clock::time_point created_at_;
    std::chrono::steady_clock::time_point last_activity_;
};

class ConnectionPool {
public:
    /**
     * Create a new connection pool
     * @param host PostgreSQL hostname
     * @param port PostgreSQL port
     * @param user Database user
     * @param password Database password
     * @param dbname Database name
     * @param min_size Minimum pool size
     * @param max_size Maximum pool size
     */
    ConnectionPool(
        const std::string& host,
        int port,
        const std::string& user,
        const std::string& password,
        const std::string& dbname,
        size_t min_size = 1,
        size_t max_size = 3
    );

    ~ConnectionPool();

    /**
     * Get a connection from the pool
     * Blocks if no connections available and pool is at max capacity
     * Times out after 5 seconds
     */
    std::shared_ptr<PooledConnection> acquire(int timeout_sec = 5);

    /**
     * Return a connection to the pool
     */
    void release(std::shared_ptr<PooledConnection> conn);

    /**
     * Get number of active connections
     */
    size_t getActiveCount() const;

    /**
     * Get total pool size
     */
    size_t getPoolSize() const;

    /**
     * Perform health check on all idle connections
     * Removes unhealthy connections
     */
    void healthCheck();

    /**
     * Reconnect all connections
     * Useful after network failure
     */
    void reconnectAll();

    /**
     * Close all connections and reset pool
     */
    void close();

    /**
     * Get pool statistics for monitoring
     */
    struct PoolStats {
        size_t total_size;
        size_t active_count;
        size_t idle_count;
        size_t failed_attempts;
        std::chrono::seconds uptime;
    };

    PoolStats getStats() const;

private:
    std::string host_;
    int port_;
    std::string user_;
    std::string password_;
    std::string dbname_;
    std::string pool_id_;

    size_t min_size_;
    size_t max_size_;

    std::vector<std::shared_ptr<PooledConnection>> connections_;
    std::queue<std::shared_ptr<PooledConnection>> available_;

    mutable std::mutex mutex_;
    std::condition_variable cv_;

    size_t failed_connections_;
    std::chrono::steady_clock::time_point created_at_;

    /**
     * Create a new connection
     */
    std::shared_ptr<PooledConnection> createConnection();

    /**
     * Build connection string
     */
    std::string buildConnectionString() const;

    /**
     * Initialize pool to min_size
     */
    void initialize();
};

#endif  // CONNECTION_POOL_H_
