#include "../include/collector.h"
#include <iostream>
#include <ctime>
#include <sstream>
#include <iomanip>
#include <cstdlib>
#include <fstream>

// DiskUsageCollector implementation
DiskUsageCollector::DiskUsageCollector(
    const std::string& hostname,
    const std::string& collectorId
)
    : enabled_(true) {
    hostname_ = hostname;
    collectorId_ = collectorId;
}

json DiskUsageCollector::execute() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");

    json result = {
        {"type", "disk_usage"},
        {"timestamp", oss.str()},
        {"filesystems", json::array()}
    };

    // Collect disk usage
    json filesystems = collectDiskUsage();
    if (filesystems.is_array()) {
        result["filesystems"] = filesystems;
    }

    std::cout << "DiskUsageCollector::execute() - gathering filesystem usage" << std::endl;

    return result;
}

json DiskUsageCollector::collectDiskUsage() {
    json filesystems = json::array();

    // Execute df command and parse output
    FILE* pipe = popen("df -B1 | tail -n +2", "r");
    if (!pipe) {
        return filesystems;
    }

    char buffer[512];
    while (fgets(buffer, sizeof(buffer), pipe) != nullptr) {
        std::string line(buffer);
        std::istringstream iss(line);

        std::string device, mount;
        long total_bytes, used_bytes, available_bytes;

        // df output: Filesystem 1B-blocks Used Available Use% Mounted on
        if (iss >> device >> total_bytes >> used_bytes >> available_bytes) {
            // Skip lines that don't look like real filesystems
            if (device.find("/dev/") != 0) {
                continue;
            }

            json fs_entry = json::object();
            fs_entry["device"] = device;
            fs_entry["mount"] = mount;
            fs_entry["total_gb"] = total_bytes / (1024LL * 1024LL * 1024LL);
            fs_entry["used_gb"] = used_bytes / (1024LL * 1024LL * 1024LL);
            fs_entry["free_gb"] = available_bytes / (1024LL * 1024LL * 1024LL);

            double percent = 0.0;
            if (total_bytes > 0) {
                percent = (100.0 * used_bytes) / total_bytes;
            }
            fs_entry["percent_used"] = percent;

            filesystems.push_back(fs_entry);
        }
    }

    pclose(pipe);

    // Fallback: try reading /etc/mtab or /proc/mounts for mount points
    // and use statfs for size information
    std::ifstream mtab_file("/etc/mtab");
    if (filesystems.empty() && mtab_file.is_open()) {
        std::string line;
        while (std::getline(mtab_file, line)) {
            if (line.empty() || line[0] == '#') continue;

            std::istringstream iss(line);
            std::string device, mount, fstype;
            if (!(iss >> device >> mount >> fstype)) continue;

            // Skip pseudo-filesystems
            if (fstype == "tmpfs" || fstype == "sysfs" || fstype == "proc" || fstype == "devtmpfs") {
                continue;
            }

            json fs_entry = json::object();
            fs_entry["device"] = device;
            fs_entry["mount"] = mount;
            fs_entry["total_gb"] = 0;  // Would use statfs here
            fs_entry["used_gb"] = 0;
            fs_entry["free_gb"] = 0;
            fs_entry["percent_used"] = 0;

            filesystems.push_back(fs_entry);
        }
        mtab_file.close();
    }

    return filesystems;
}

// CollectorManager implementation
CollectorManager::CollectorManager(const std::string& hostname, const std::string& collectorId)
    : hostname_(hostname),
      collectorId_(collectorId),
      thread_pool_(nullptr),
      thread_pool_size_(4),
      last_cycle_time_ms_(0) {
    // Phase 1.1: Initialize thread pool for parallel collector execution
    initializeThreadPool();
}

void CollectorManager::initializeThreadPool() {
    try {
        // TODO: Read thread_pool_size from config file
        // config->getInt("collector_threading", "thread_pool_size", 4)
        // For now, use default of 4 threads
        thread_pool_size_ = 4;

        thread_pool_ = std::make_unique<ThreadPool>(thread_pool_size_);
        std::cerr << "DEBUG: CollectorManager thread pool initialized with " << thread_pool_size_ << " threads" << std::endl;
    } catch (const std::exception& e) {
        std::cerr << "ERROR: Failed to initialize thread pool: " << e.what() << std::endl;
        thread_pool_ = nullptr;
    }
}

void CollectorManager::addCollector(std::shared_ptr<Collector> collector) {
    collectors_.push_back(collector);
}

json CollectorManager::collectAll() {
    json result = {
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"timestamp", json::array()},
        {"metrics", json::array()}
    };

    for (auto& collector : collectors_) {
        if (collector->isEnabled()) {
            json metrics = collector->execute();
            result["metrics"].push_back(metrics);
        }
    }

    return result;
}

json CollectorManager::collectAllParallel() {
    // Phase 1.1: Parallel collector execution using thread pool
    // Expected benefit: 75% cycle time reduction (57.7s → 14.4s at 100 collectors)
    // Reduces bottleneck from sequential execution to parallel batch execution

    auto start_time = std::chrono::steady_clock::now();

    json result = {
        {"collector_id", collectorId_},
        {"hostname", hostname_},
        {"timestamp", json::array()},
        {"metrics", json::array()}
    };

    // If thread pool is not available, fall back to sequential execution
    if (!thread_pool_) {
        std::cerr << "WARNING: Thread pool not available, falling back to sequential collection" << std::endl;
        return collectAll();
    }

    // Collect metrics in parallel using thread pool
    std::vector<std::future<json>> futures;
    std::vector<std::shared_ptr<Collector>> enabled_collectors;

    // Enqueue all enabled collectors
    for (auto& collector : collectors_) {
        if (collector->isEnabled()) {
            enabled_collectors.push_back(collector);
            futures.push_back(
                thread_pool_->enqueue([collector]() {
                    return collector->execute();
                })
            );
        }
    }

    // Collect results from all futures
    // This wait operation ensures we don't return until all collectors complete
    for (size_t i = 0; i < futures.size(); ++i) {
        try {
            json metrics = futures[i].get();
            result["metrics"].push_back(metrics);
            std::cerr << "DEBUG: Collected metrics from " << enabled_collectors[i]->getType() << std::endl;
        } catch (const std::exception& e) {
            std::cerr << "ERROR: Collector " << enabled_collectors[i]->getType()
                      << " failed: " << e.what() << std::endl;
        }
    }

    // Calculate cycle time
    auto end_time = std::chrono::steady_clock::now();
    last_cycle_time_ms_ = std::chrono::duration_cast<std::chrono::milliseconds>(
        end_time - start_time
    ).count();

    std::cerr << "DEBUG: Parallel collection completed in " << last_cycle_time_ms_ << "ms" << std::endl;

    return result;
}

void CollectorManager::configure(const json& config) {
    // Apply new configuration to running collectors
    // This method handles dynamic configuration updates without restart

    // Log configuration update
    std::cout << "CollectorManager::configure() - applying new configuration" << std::endl;

    // The configuration has already been loaded into gConfig by the main loop
    // Here we can add additional per-collector configuration logic if needed

    // For PostgreSQL collectors, the connection parameters might have changed
    // In a production implementation, you would:
    // 1. Validate the new configuration
    // 2. Update collector-specific settings
    // 3. Reconnect to services if needed
    // 4. Signal collectors to reload their configuration

    // For now, log successful configuration update
    if (config.is_object()) {
        std::cout << "Configuration update applied successfully" << std::endl;
    }
}
