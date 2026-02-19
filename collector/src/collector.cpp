#include "../include/collector.h"
#include <iostream>
#include <ctime>

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
    // TODO: Execute `df -B1` and parse output
    // Or use statfs() system call
    // Schema:
    // {
    //   "mount": "/",
    //   "device": "/dev/sda1",
    //   "total_gb": 100,
    //   "used_gb": 45,
    //   "free_gb": 55,
    //   "percent_used": 45
    // }

    return json::array();
}

// CollectorManager implementation
CollectorManager::CollectorManager(const std::string& hostname, const std::string& collectorId)
    : hostname_(hostname), collectorId_(collectorId) {
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

void CollectorManager::configure(const json& config) {
    // TODO: Implement configuration from JSON
}
