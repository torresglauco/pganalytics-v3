#include "../include/host_inventory_plugin.h"
#include <iostream>
#include <fstream>
#include <sstream>
#include <ctime>
#include <iomanip>
#include <cstring>
#include <unistd.h>
#include <sys/utsname.h>
#include <sys/statvfs.h>

/**
 * Get current timestamp in ISO8601 format
 */
static std::string getCurrentTimestamp() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");
    return oss.str();
}

/**
 * Trim whitespace from string
 */
static std::string trim(const std::string& str) {
    size_t first = str.find_first_not_of(" \t\n\r\"");
    if (first == std::string::npos) return "";
    size_t last = str.find_last_not_of(" \t\n\r\"");
    return str.substr(first, last - first + 1);
}

/**
 * Parse a simple key=value file (like /etc/os-release)
 */
static std::map<std::string, std::string> parseKeyValueFile(const std::string& path) {
    std::map<std::string, std::string> result;
    std::ifstream file(path);
    if (!file.is_open()) return result;

    std::string line;
    while (std::getline(file, line)) {
        size_t pos = line.find('=');
        if (pos != std::string::npos) {
            std::string key = trim(line.substr(0, pos));
            std::string value = trim(line.substr(pos + 1));
            result[key] = value;
        }
    }
    file.close();
    return result;
}

/**
 * Constructor
 */
HostInventoryCollector::HostInventoryCollector(
    const std::string& hostname,
    const std::string& collectorId,
    const std::string& postgresHost,
    int postgresPort,
    const std::string& postgresUser,
    const std::string& postgresPassword
)
    : postgresHost_(postgresHost),
      postgresPort_(postgresPort),
      postgresUser_(postgresUser),
      postgresPassword_(postgresPassword),
      enabled_(true) {
    hostname_ = hostname;
    collectorId_ = collectorId;
}

/**
 * Destructor
 */
HostInventoryCollector::~HostInventoryCollector() = default;

/**
 * Execute host inventory collection
 */
json HostInventoryCollector::execute() {
    json result = {
        {"type", "host_inventory"},
        {"timestamp", getCurrentTimestamp()},
        {"hostname", hostname_}
    };

    // Collect OS information
    auto osInfo = collectOsInfo();
    if (!osInfo.is_null()) {
        result["os"] = osInfo;
    }

    // Collect CPU information
    auto cpuInfo = collectCpuInfo();
    if (!cpuInfo.is_null()) {
        result["cpu"] = cpuInfo;
    }

    // Collect memory information
    auto memInfo = collectMemoryInfo();
    if (!memInfo.is_null()) {
        result["memory"] = memInfo;
    }

    // Collect disk information
    auto diskInfo = collectDiskInfo();
    if (!diskInfo.is_null()) {
        result["disk"] = diskInfo;
    }

    // Collect PostgreSQL configuration (requires connection)
    auto pgConfig = collectPostgresConfig();
    if (!pgConfig.is_null()) {
        result["postgres"] = pgConfig;
    }

    return result;
}

/**
 * Collect OS information from /etc/os-release and uname
 */
json HostInventoryCollector::collectOsInfo() {
    json result = json::object();

    // Parse /etc/os-release for distribution info
    auto osRelease = parseKeyValueFile("/etc/os-release");

    if (!osRelease.empty()) {
        if (osRelease.count("NAME")) {
            result["os_name"] = osRelease["NAME"];
        }
        if (osRelease.count("VERSION")) {
            result["os_version"] = osRelease["VERSION"];
        } else if (osRelease.count("VERSION_ID")) {
            result["os_version"] = osRelease["VERSION_ID"];
        }
    }

    // Get kernel version from uname
    struct utsname buf;
    if (uname(&buf) == 0) {
        result["os_kernel"] = std::string(buf.release);
        result["os_arch"] = std::string(buf.machine);
    }

    return result;
}

/**
 * Collect CPU information from /proc/cpuinfo
 */
json HostInventoryCollector::collectCpuInfo() {
    json result = json::object();

    // Get CPU cores using sysconf
    long cores = sysconf(_SC_NPROCESSORS_ONLN);
    if (cores > 0) {
        result["cpu_cores"] = static_cast<int>(cores);
    }

    // Parse /proc/cpuinfo for model name and frequency
    std::ifstream cpuinfo("/proc/cpuinfo");
    if (cpuinfo.is_open()) {
        std::string line;
        std::string model_name;
        double cpu_mhz = 0.0;
        int processor_count = 0;

        while (std::getline(cpuinfo, line)) {
            if (line.find("model name") == 0) {
                size_t pos = line.find(':');
                if (pos != std::string::npos && model_name.empty()) {
                    model_name = trim(line.substr(pos + 1));
                }
            } else if (line.find("cpu MHz") == 0) {
                size_t pos = line.find(':');
                if (pos != std::string::npos && cpu_mhz == 0.0) {
                    std::string mhz_str = trim(line.substr(pos + 1));
                    try {
                        cpu_mhz = std::stod(mhz_str);
                    } catch (...) {
                        // Ignore parse errors
                    }
                }
            } else if (line.find("processor") == 0) {
                processor_count++;
            }
        }
        cpuinfo.close();

        if (!model_name.empty()) {
            result["cpu_model"] = model_name;
        }
        if (cpu_mhz > 0.0) {
            result["cpu_mhz"] = cpu_mhz;
        }
        // Fallback for cores if sysconf didn't work
        if (result["cpu_cores"].is_null() && processor_count > 0) {
            result["cpu_cores"] = processor_count;
        }
    }

    return result;
}

/**
 * Collect memory information from /proc/meminfo
 */
json HostInventoryCollector::collectMemoryInfo() {
    json result = json::object();

    std::ifstream meminfo("/proc/meminfo");
    if (meminfo.is_open()) {
        std::string line;
        long total_kb = 0;

        while (std::getline(meminfo, line)) {
            if (line.find("MemTotal:") == 0) {
                sscanf(line.c_str(), "MemTotal: %ld kB", &total_kb);
                break;
            }
        }
        meminfo.close();

        if (total_kb > 0) {
            result["memory_total_mb"] = total_kb / 1024;
        }
    }

    return result;
}

/**
 * Collect disk information using statvfs
 */
json HostInventoryCollector::collectDiskInfo() {
    json result = json::object();

    // Get root filesystem stats
    struct statvfs stat;
    if (statvfs("/", &stat) == 0) {
        // Calculate total disk space in GB
        unsigned long long total_bytes = stat.f_blocks * stat.f_frsize;
        long total_gb = total_bytes / (1024LL * 1024LL * 1024LL);

        result["disk_total_gb"] = total_gb;
    }

    return result;
}

/**
 * Collect PostgreSQL configuration from pg_settings
 */
json HostInventoryCollector::collectPostgresConfig() {
    json result = json::object();

#ifdef HAVE_LIBPQ
    // Check if PostgreSQL connection is configured
    if (postgresHost_.empty()) {
        return result;
    }

    // Connect to PostgreSQL
    PGconn* conn = connectToDatabase("postgres");
    if (!conn) {
        std::cerr << "Warning: Could not connect to PostgreSQL for host inventory" << std::endl;
        return result;
    }

    // Get PostgreSQL version
    std::string version = getPostgresVersion(conn);
    if (!version.empty()) {
        result["postgres_version"] = version;
    }

    // Determine edition (Community vs EnterpriseDB)
    const char* server_version_str = PQparameterStatus(conn, "server_version");
    if (server_version_str) {
        // Check if this might be EnterpriseDB based on version string
        // EnterpriseDB typically has versions like "16.2" but we can check for specific settings
        result["postgres_edition"] = "Community"; // Default assumption
    }

    // Get key PostgreSQL settings
    std::string port = getPostgresSetting(conn, "port");
    if (!port.empty()) {
        result["postgres_port"] = std::stoi(port);
    }

    std::string data_dir = getPostgresSetting(conn, "data_directory");
    if (!data_dir.empty()) {
        result["postgres_data_dir"] = data_dir;
    }

    std::string max_connections = getPostgresSetting(conn, "max_connections");
    if (!max_connections.empty()) {
        result["postgres_max_connections"] = std::stoi(max_connections);
    }

    std::string shared_buffers = getPostgresSetting(conn, "shared_buffers");
    if (!shared_buffers.empty()) {
        // Parse shared_buffers (e.g., "128MB" -> 128)
        std::string value = shared_buffers;
        int multiplier = 1;

        if (value.find("GB") != std::string::npos) {
            multiplier = 1024;
            value = value.substr(0, value.find("GB"));
        } else if (value.find("MB") != std::string::npos) {
            multiplier = 1;
            value = value.substr(0, value.find("MB"));
        } else if (value.find("kB") != std::string::npos) {
            // kB -> MB is divide by 1024, skip for now
            value = "0";
        }

        try {
            result["postgres_shared_buffers_mb"] = std::stoi(trim(value)) * multiplier;
        } catch (...) {
            // Ignore parse errors
        }
    }

    std::string work_mem = getPostgresSetting(conn, "work_mem");
    if (!work_mem.empty()) {
        // Parse work_mem (e.g., "4MB" -> 4)
        std::string value = work_mem;
        int mb_value = 0;

        if (value.find("GB") != std::string::npos) {
            mb_value = std::stoi(trim(value.substr(0, value.find("GB")))) * 1024;
        } else if (value.find("MB") != std::string::npos) {
            mb_value = std::stoi(trim(value.substr(0, value.find("MB"))));
        } else if (value.find("kB") != std::string::npos) {
            mb_value = std::stoi(trim(value.substr(0, value.find("kB")))) / 1024;
        }

        result["postgres_work_mem_mb"] = mb_value;
    }

    PQfinish(conn);

#else
    // libpq not available - skip PostgreSQL config
    std::cerr << "Note: libpq not available, skipping PostgreSQL configuration collection" << std::endl;
#endif

    return result;
}

#ifdef HAVE_LIBPQ
/**
 * Connect to PostgreSQL database
 */
PGconn* HostInventoryCollector::connectToDatabase(const std::string& dbname) {
    std::string connstr = "host=" + postgresHost_ +
                         " port=" + std::to_string(postgresPort_) +
                         " dbname=" + dbname +
                         " user=" + postgresUser_;

    if (!postgresPassword_.empty()) {
        connstr += " password=" + postgresPassword_;
    }

    connstr += " connect_timeout=5";

    PGconn* conn = PQconnectdb(connstr.c_str());

    if (PQstatus(conn) != CONNECTION_OK) {
        std::cerr << "Connection to " << dbname << " failed: " << PQerrorMessage(conn) << std::endl;
        PQfinish(conn);
        return nullptr;
    }

    return conn;
}

/**
 * Get PostgreSQL version string
 */
std::string HostInventoryCollector::getPostgresVersion(PGconn* conn) {
    PGresult* res = PQexec(conn, "SHOW server_version");
    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        PQclear(res);
        return "";
    }

    std::string version;
    if (PQntuples(res) > 0) {
        version = PQgetvalue(res, 0, 0);
    }

    PQclear(res);
    return version;
}

/**
 * Get a pg_settings value by name
 */
std::string HostInventoryCollector::getPostgresSetting(PGconn* conn, const std::string& setting_name) {
    std::string query = "SELECT setting FROM pg_settings WHERE name = '" + setting_name + "'";
    PGresult* res = PQexec(conn, query.c_str());

    if (PQresultStatus(res) != PGRES_TUPLES_OK) {
        PQclear(res);
        return "";
    }

    std::string value;
    if (PQntuples(res) > 0) {
        value = PQgetvalue(res, 0, 0);
    }

    PQclear(res);
    return value;
}
#endif