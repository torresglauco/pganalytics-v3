#include "../include/collector.h"
#include <iostream>
#include <fstream>
#include <sstream>
#include <ctime>
#include <iomanip>
#include <cstring>
#include <unistd.h>

SysstatCollector::SysstatCollector(
    const std::string& hostname,
    const std::string& collectorId
)
    : enabled_(true) {
    hostname_ = hostname;
    collectorId_ = collectorId;
}

json SysstatCollector::execute() {
    auto now = std::time(nullptr);
    auto tm = *std::gmtime(&now);
    std::ostringstream oss;
    oss << std::put_time(&tm, "%Y-%m-%dT%H:%M:%SZ");

    json result = {
        {"type", "sysstat"},
        {"timestamp", oss.str()}
    };

    // Collect system statistics
    auto cpu = collectCpuStats();
    if (!cpu.is_null()) {
        result["cpu"] = cpu;
    }

    auto memory = collectMemoryStats();
    if (!memory.is_null()) {
        result["memory"] = memory;
    }

    auto io = collectIoStats();
    if (!io.is_null()) {
        result["disk_io"] = io;
    }

    auto load = collectLoadAverage();
    if (!load.is_null()) {
        result["load"] = load;
    }

    return result;
}

json SysstatCollector::collectCpuStats() {
    json result = json::object();
    result["user"] = 0.0;
    result["system"] = 0.0;
    result["idle"] = 100.0;
    result["iowait"] = 0.0;

    // Load average from getloadavg() (preferred method)
    double loads[3];
    if (getloadavg(loads, 3) != -1) {
        result["load_1m"] = loads[0];
        result["load_5m"] = loads[1];
        result["load_15m"] = loads[2];
    } else {
        // Fallback: parse /proc/loadavg
        std::ifstream loadavg_file("/proc/loadavg");
        if (loadavg_file.is_open()) {
            double load_1m, load_5m, load_15m;
            int running, total;
            int last_pid;
            if (loadavg_file >> load_1m >> load_5m >> load_15m >> running >> total >> last_pid) {
                result["load_1m"] = load_1m;
                result["load_5m"] = load_5m;
                result["load_15m"] = load_15m;
            }
            loadavg_file.close();
        }
    }

    // Parse /proc/stat for CPU percentages
    std::ifstream stat_file("/proc/stat");
    if (stat_file.is_open()) {
        std::string line;
        if (std::getline(stat_file, line)) {
            // First line is aggregate CPU stats
            std::istringstream iss(line);
            std::string cpu_label;
            unsigned long user, nice, system, idle, iowait, irq, softirq, steal;
            if (iss >> cpu_label >> user >> nice >> system >> idle >> iowait >> irq >> softirq >> steal) {
                unsigned long total = user + nice + system + idle + iowait + irq + softirq + steal;
                if (total > 0) {
                    // Calculate percentages (approximate since we don't have previous sample)
                    // In real usage, would need delta between samples
                    result["user"] = (100.0 * user) / total;
                    result["system"] = (100.0 * system) / total;
                    result["idle"] = (100.0 * idle) / total;
                    result["iowait"] = (100.0 * iowait) / total;
                }
            }
        }
        stat_file.close();
    }

    return result;
}

json SysstatCollector::collectMemoryStats() {
    json result = json::object();
    result["total_mb"] = 0;
    result["free_mb"] = 0;
    result["cached_mb"] = 0;
    result["used_mb"] = 0;

    std::ifstream meminfo_file("/proc/meminfo");
    if (meminfo_file.is_open()) {
        std::string line;
        long total_kb = 0, free_kb = 0, available_kb = 0, cached_kb = 0, buffers_kb = 0;

        while (std::getline(meminfo_file, line)) {
            if (line.find("MemTotal:") == 0) {
                sscanf(line.c_str(), "MemTotal: %ld kB", &total_kb);
            } else if (line.find("MemFree:") == 0) {
                sscanf(line.c_str(), "MemFree: %ld kB", &free_kb);
            } else if (line.find("MemAvailable:") == 0) {
                sscanf(line.c_str(), "MemAvailable: %ld kB", &available_kb);
            } else if (line.find("Cached:") == 0 && line.find("SwapCached") == std::string::npos) {
                sscanf(line.c_str(), "Cached: %ld kB", &cached_kb);
            } else if (line.find("Buffers:") == 0) {
                sscanf(line.c_str(), "Buffers: %ld kB", &buffers_kb);
            }
        }

        meminfo_file.close();

        // Convert KB to MB
        result["total_mb"] = total_kb / 1024;
        result["free_mb"] = free_kb / 1024;
        result["cached_mb"] = (cached_kb + buffers_kb) / 1024;
        result["used_mb"] = (total_kb - free_kb - cached_kb - buffers_kb) / 1024;
    }

    return result;
}

json SysstatCollector::collectIoStats() {
    json result = json::array();

    // Parse /proc/diskstats
    // Format: major_num minor_num device_name reads read_merges read_sectors read_ticks
    //         writes write_merges write_sectors write_ticks in_flight io_ticks time_in_queue
    std::ifstream diskstats_file("/proc/diskstats");
    if (diskstats_file.is_open()) {
        std::string line;
        while (std::getline(diskstats_file, line)) {
            std::istringstream iss(line);
            int major, minor;
            std::string device;
            unsigned long reads, read_merges, read_sectors, read_ticks;
            unsigned long writes, write_merges, write_sectors, write_ticks;
            unsigned long in_flight, io_ticks, time_in_queue;

            if (iss >> major >> minor >> device >> reads >> read_merges >> read_sectors >> read_ticks
                  >> writes >> write_merges >> write_sectors >> write_ticks >> in_flight >> io_ticks >> time_in_queue) {

                // Skip loop devices and ram disks
                if (device.find("loop") != std::string::npos || device.find("ram") != std::string::npos) {
                    continue;
                }

                json io_entry = json::object();
                io_entry["device"] = device;

                // Note: These are cumulative values since boot, not instantaneous rates
                // In a real implementation, would need to store previous values and calculate delta
                io_entry["read_ops"] = static_cast<long long>(reads);
                io_entry["write_ops"] = static_cast<long long>(writes);
                io_entry["read_sectors"] = static_cast<long long>(read_sectors);
                io_entry["write_sectors"] = static_cast<long long>(write_sectors);

                result.push_back(io_entry);
            }
        }
        diskstats_file.close();
    }

    return result;
}

json SysstatCollector::collectLoadAverage() {
    // Load average is already collected in collectCpuStats()
    // This method kept for interface compatibility
    return json();
}
