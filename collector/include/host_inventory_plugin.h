#pragma once

#include "collector.h"
#include <string>
#include <cstdint>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

#ifdef HAVE_LIBPQ
#include <libpq-fe.h>
#else
// Forward declaration when libpq not available
typedef struct pg_conn PGconn;
typedef struct pg_result PGresult;
#endif

/**
 * Host Inventory Collector
 *
 * Gathers comprehensive host information including:
 * - Operating system version and kernel
 * - CPU specifications (cores, model, frequency)
 * - Memory and disk capacity
 * - PostgreSQL version and configuration
 *
 * Requirements:
 * - Linux operating system (uses /proc filesystem)
 * - PostgreSQL 9.0+ for pg_settings queries
 * - Permissions: Read access to /proc, PostgreSQL connection for pg_settings
 *
 * Metrics collected:
 * - OS name, version, kernel (from /etc/os-release and uname)
 * - CPU cores, model name, frequency (from /proc/cpuinfo)
 * - Total memory in MB (from /proc/meminfo)
 * - Total disk capacity in GB (from statvfs)
 * - PostgreSQL version, port, data directory, and key settings
 */
class HostInventoryCollector : public Collector {
public:
    /**
     * Constructor
     * @param hostname Collector hostname
     * @param collectorId Unique collector identifier
     * @param postgresHost PostgreSQL server host (optional for OS-only collection)
     * @param postgresPort PostgreSQL server port
     * @param postgresUser Database user for connection
     * @param postgresPassword Database password
     */
    HostInventoryCollector(
        const std::string& hostname,
        const std::string& collectorId,
        const std::string& postgresHost = "",
        int postgresPort = 5432,
        const std::string& postgresUser = "",
        const std::string& postgresPassword = ""
    );

    /**
     * Destructor
     */
    ~HostInventoryCollector();

    /**
     * Execute host inventory collection
     * @return JSON object with all host inventory data
     */
    json execute() override;

    /**
     * Get collector type
     */
    std::string getType() const override { return "host_inventory"; }

    /**
     * Check if this collector is enabled
     */
    bool isEnabled() const override { return enabled_; }

private:
    // Configuration
    std::string postgresHost_;
    int postgresPort_;
    std::string postgresUser_;
    std::string postgresPassword_;
    bool enabled_;

    /**
     * Collect OS information from /etc/os-release and uname
     * @return JSON object with os_name, os_version, os_kernel
     */
    json collectOsInfo();

    /**
     * Collect CPU information from /proc/cpuinfo
     * @return JSON object with cpu_cores, cpu_model, cpu_mhz
     */
    json collectCpuInfo();

    /**
     * Collect memory information from /proc/meminfo
     * @return JSON object with memory_total_mb
     */
    json collectMemoryInfo();

    /**
     * Collect disk information using statvfs
     * @return JSON object with disk_total_gb
     */
    json collectDiskInfo();

    /**
     * Collect PostgreSQL configuration from pg_settings
     * Requires PostgreSQL connection
     * @return JSON object with postgres_* fields
     */
    json collectPostgresConfig();

#ifdef HAVE_LIBPQ
    /**
     * Connect to PostgreSQL database
     * @param dbname Database name to connect to
     * @return PGconn pointer or nullptr on failure
     */
    PGconn* connectToDatabase(const std::string& dbname);

    /**
     * Get PostgreSQL version string
     * @param conn Database connection
     * @return Version string (e.g., "16.2")
     */
    std::string getPostgresVersion(PGconn* conn);

    /**
     * Get a pg_settings value by name
     * @param conn Database connection
     * @param setting_name Name of the setting
     * @return Setting value as string
     */
    std::string getPostgresSetting(PGconn* conn, const std::string& setting_name);
#endif
};