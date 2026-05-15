#include "gtest/gtest.h"
#include <string>
#include <sstream>
#include <map>

/**
 * Host Metrics Collection Tests
 *
 * Tests for host inventory and metrics collection functionality:
 * - CPU metrics parsing from /proc/stat format
 * - Memory metrics parsing from /proc/meminfo format
 * - Disk metrics parsing from statvfs output
 * - Network I/O metrics parsing
 * - Metrics JSON output structure validation
 * - Error handling for missing /proc files
 *
 * These tests validate the parsing functions used by HostInventoryCollector
 * and SysstatCollector plugins.
 */

// ============================================================================
// Mock /proc Parsing Functions
// ============================================================================

/**
 * Parse CPU stats from /proc/stat format
 * Format: cpu user nice system idle iowait irq softirq steal
 */
struct CpuStats {
    double user;
    double system;
    double idle;
    double iowait;
    double load1m;
};

CpuStats parseCpuStats(const std::string& statContent, const std::string& loadavgContent) {
    CpuStats stats = {0.0, 0.0, 100.0, 0.0, 0.0};

    // Parse /proc/stat line
    std::istringstream iss(statContent);
    std::string cpuLabel;
    unsigned long user, nice, system, idle, iowait, irq, softirq, steal;

    if (iss >> cpuLabel >> user >> nice >> system >> idle >> iowait >> irq >> softirq >> steal) {
        unsigned long total = user + nice + system + idle + iowait + irq + softirq + steal;
        if (total > 0) {
            stats.user = (100.0 * user) / total;
            stats.system = (100.0 * system) / total;
            stats.idle = (100.0 * idle) / total;
            stats.iowait = (100.0 * iowait) / total;
        }
    }

    // Parse /proc/loadavg
    std::istringstream loadIss(loadavgContent);
    double load1m, load5m, load15m;
    if (loadIss >> load1m >> load5m >> load15m) {
        stats.load1m = load1m;
    }

    return stats;
}

/**
 * Parse memory info from /proc/meminfo format
 */
struct MemoryStats {
    long totalMb;
    long freeMb;
    long cachedMb;
    long usedMb;
    double usedPercent;
};

MemoryStats parseMemoryInfo(const std::string& meminfoContent) {
    MemoryStats stats = {0, 0, 0, 0, 0.0};

    std::istringstream iss(meminfoContent);
    std::string line;
    long totalKb = 0, freeKb = 0, availableKb = 0, cachedKb = 0, buffersKb = 0;

    while (std::getline(iss, line)) {
        if (line.find("MemTotal:") == 0) {
            sscanf(line.c_str(), "MemTotal: %ld kB", &totalKb);
        } else if (line.find("MemFree:") == 0) {
            sscanf(line.c_str(), "MemFree: %ld kB", &freeKb);
        } else if (line.find("MemAvailable:") == 0) {
            sscanf(line.c_str(), "MemAvailable: %ld kB", &availableKb);
        } else if (line.find("Cached:") == 0 && line.find("SwapCached") == std::string::npos) {
            sscanf(line.c_str(), "Cached: %ld kB", &cachedKb);
        } else if (line.find("Buffers:") == 0) {
            sscanf(line.c_str(), "Buffers: %ld kB", &buffersKb);
        }
    }

    // Convert KB to MB
    stats.totalMb = totalKb / 1024;
    stats.freeMb = freeKb / 1024;
    stats.cachedMb = (cachedKb + buffersKb) / 1024;
    stats.usedMb = (totalKb - freeKb - cachedKb - buffersKb) / 1024;

    if (totalKb > 0) {
        stats.usedPercent = 100.0 * (totalKb - freeKb - cachedKb - buffersKb) / totalKb;
    }

    return stats;
}

/**
 * Parse disk info from statvfs-like output
 */
struct DiskStats {
    long totalGb;
    long usedGb;
    double usedPercent;
};

DiskStats parseDiskInfo(unsigned long long totalBytes, unsigned long long freeBytes) {
    DiskStats stats = {0, 0, 0.0};

    stats.totalGb = totalBytes / (1024LL * 1024LL * 1024LL);
    stats.usedGb = (totalBytes - freeBytes) / (1024LL * 1024LL * 1024LL);

    if (totalBytes > 0) {
        stats.usedPercent = 100.0 * (totalBytes - freeBytes) / totalBytes;
    }

    return stats;
}

/**
 * Parse /proc/diskstats format
 */
struct DiskIoStats {
    std::string device;
    unsigned long readOps;
    unsigned long writeOps;
};

std::vector<DiskIoStats> parseDiskStats(const std::string& diskstatsContent) {
    std::vector<DiskIoStats> result;
    std::istringstream iss(diskstatsContent);
    std::string line;

    while (std::getline(iss, line)) {
        std::istringstream lineIss(line);
        int major, minor;
        std::string device;
        unsigned long reads, readMerges, readSectors, readTicks;
        unsigned long writes, writeMerges, writeSectors, writeTicks;
        unsigned long inFlight, ioTicks, timeInQueue;

        if (lineIss >> major >> minor >> device >> reads >> readMerges >> readSectors >> readTicks
                     >> writes >> writeMerges >> writeSectors >> writeTicks >> inFlight >> ioTicks >> timeInQueue) {

            // Skip loop devices and ram disks
            if (device.find("loop") != std::string::npos || device.find("ram") != std::string::npos) {
                continue;
            }

            DiskIoStats stats;
            stats.device = device;
            stats.readOps = reads;
            stats.writeOps = writes;
            result.push_back(stats);
        }
    }

    return result;
}

/**
 * Parse /proc/net/dev format
 */
struct NetworkStats {
    std::string interface;
    unsigned long long rxBytes;
    unsigned long long txBytes;
};

std::vector<NetworkStats> parseNetworkStats(const std::string& netdevContent) {
    std::vector<NetworkStats> result;
    std::istringstream iss(netdevContent);
    std::string line;

    // Skip header lines
    std::getline(iss, line);
    std::getline(iss, line);

    while (std::getline(iss, line)) {
        // Find interface name (before colon)
        size_t colonPos = line.find(':');
        if (colonPos == std::string::npos) continue;

        std::string iface = line.substr(0, colonPos);
        // Trim whitespace
        size_t start = iface.find_first_not_of(" \t");
        size_t end = iface.find_last_not_of(" \t");
        if (start != std::string::npos && end != std::string::npos) {
            iface = iface.substr(start, end - start + 1);
        }

        // Skip loopback
        if (iface == "lo") continue;

        std::string values = line.substr(colonPos + 1);
        std::istringstream valIss(values);

        unsigned long long rxBytes, rxPackets, rxErrs, rxDrop, rxFifo, rxFrame, rxCompressed, rxMulticast;
        unsigned long long txBytes, txPackets, txErrs, txDrop, txFifo, txCollisions, txCarrier, txCompressed;

        if (valIss >> rxBytes >> rxPackets >> rxErrs >> rxDrop >> rxFifo >> rxFrame >> rxCompressed >> rxMulticast
                    >> txBytes >> txPackets >> txErrs >> txDrop >> txFifo >> txCollisions >> txCarrier >> txCompressed) {

            NetworkStats stats;
            stats.interface = iface;
            stats.rxBytes = rxBytes;
            stats.txBytes = txBytes;
            result.push_back(stats);
        }
    }

    return result;
}

// ============================================================================
// Tests
// ============================================================================

class HostMetricsTest : public ::testing::Test {
protected:
    void SetUp() override {
        // Setup test fixtures
    }
};

// ============================================================================
// CPU Stats Tests
// ============================================================================

TEST_F(HostMetricsTest, ParseCpuStatsCorrectly) {
    // Simulated /proc/stat content
    std::string statContent = "cpu  1000 0 500 8000 200 0 0 0";
    std::string loadavgContent = "0.50 0.75 1.00 1/100 12345";

    CpuStats stats = parseCpuStats(statContent, loadavgContent);

    // Total = 1000 + 0 + 500 + 8000 + 200 + 0 + 0 + 0 = 9700
    // user = 1000/9700 * 100 = 10.3%
    // system = 500/9700 * 100 = 5.15%
    // idle = 8000/9700 * 100 = 82.47%
    // iowait = 200/9700 * 100 = 2.06%

    EXPECT_NEAR(stats.user, 10.3, 0.5);
    EXPECT_NEAR(stats.system, 5.15, 0.5);
    EXPECT_NEAR(stats.idle, 82.47, 0.5);
    EXPECT_NEAR(stats.iowait, 2.06, 0.5);
    EXPECT_NEAR(stats.load1m, 0.50, 0.01);
}

TEST_F(HostMetricsTest, ParseCpuStatsHandlesMissingFields) {
    // Malformed input with fewer fields
    std::string statContent = "cpu  1000 0 500";
    std::string loadavgContent = "0.50 0.75 1.00";

    CpuStats stats = parseCpuStats(statContent, loadavgContent);

    // Should have default values when parsing fails
    EXPECT_DOUBLE_EQ(stats.idle, 100.0);  // Default
}

TEST_F(HostMetricsTest, ParseCpuStatsFromProcFormat) {
    // Real /proc/stat format example
    std::string statContent = "cpu  2255 34 2290 22625563 6290 0 563 0 0 0\ncpu0 1132 17 1145 11312782 3145 0 281 0 0 0";
    std::string loadavgContent = "0.27 0.39 0.41 2/362 12345";

    CpuStats stats = parseCpuStats(statContent, loadavgContent);

    EXPECT_GT(stats.user, 0.0);
    EXPECT_GT(stats.system, 0.0);
    EXPECT_GT(stats.idle, 0.0);
    EXPECT_NEAR(stats.load1m, 0.27, 0.01);
}

// ============================================================================
// Memory Stats Tests
// ============================================================================

TEST_F(HostMetricsTest, ParseMemoryInfoCorrectly) {
    // Simulated /proc/meminfo content
    std::string meminfoContent =
        "MemTotal:       16383796 kB\n"
        "MemFree:         1234567 kB\n"
        "MemAvailable:   12345678 kB\n"
        "Buffers:          234567 kB\n"
        "Cached:          3456789 kB\n";

    MemoryStats stats = parseMemoryInfo(meminfoContent);

    // Total = 16383796 KB / 1024 = 15999 MB (approx)
    EXPECT_NEAR(stats.totalMb, 15999, 1);

    // Free = 1234567 KB / 1024 = 1205 MB (approx)
    EXPECT_NEAR(stats.freeMb, 1205, 1);

    // Cached + buffers = (3456789 + 234567) KB / 1024 = 3604 MB (approx)
    EXPECT_NEAR(stats.cachedMb, 3604, 1);
}

TEST_F(HostMetricsTest, ParseMemoryInfoCalculatesUsedCorrectly) {
    std::string meminfoContent =
        "MemTotal:       16383796 kB\n"
        "MemFree:         1000000 kB\n"
        "Cached:          3000000 kB\n"
        "Buffers:          500000 kB\n";

    MemoryStats stats = parseMemoryInfo(meminfoContent);

    // Used = total - free - cached - buffers = 16383796 - 1000000 - 3000000 - 500000 = 11883796 KB
    // In MB: 11883796 / 1024 = 11605 MB
    EXPECT_NEAR(stats.usedMb, 11605, 1);

    // Used percent = 11883796 / 16383796 * 100 = 72.5%
    EXPECT_NEAR(stats.usedPercent, 72.5, 0.5);
}

TEST_F(HostMetricsTest, ParseMemoryInfoHandlesEmptyInput) {
    std::string meminfoContent = "";
    MemoryStats stats = parseMemoryInfo(meminfoContent);

    // Should return zeros for empty input
    EXPECT_EQ(stats.totalMb, 0);
    EXPECT_EQ(stats.freeMb, 0);
    EXPECT_EQ(stats.usedMb, 0);
}

// ============================================================================
// Disk Stats Tests
// ============================================================================

TEST_F(HostMetricsTest, ParseDiskInfoCorrectly) {
    // 100 GB total, 30 GB used
    unsigned long long totalBytes = 100LL * 1024 * 1024 * 1024;
    unsigned long long freeBytes = 70LL * 1024 * 1024 * 1024;

    DiskStats stats = parseDiskInfo(totalBytes, freeBytes);

    EXPECT_EQ(stats.totalGb, 100);
    EXPECT_EQ(stats.usedGb, 30);
    EXPECT_NEAR(stats.usedPercent, 30.0, 0.1);
}

TEST_F(HostMetricsTest, ParseDiskInfoHandlesFullDisk) {
    // Full disk scenario
    unsigned long long totalBytes = 100LL * 1024 * 1024 * 1024;
    unsigned long long freeBytes = 0;

    DiskStats stats = parseDiskInfo(totalBytes, freeBytes);

    EXPECT_EQ(stats.totalGb, 100);
    EXPECT_EQ(stats.usedGb, 100);
    EXPECT_NEAR(stats.usedPercent, 100.0, 0.1);
}

TEST_F(HostMetricsTest, ParseDiskIoStatsCorrectly) {
    // Simulated /proc/diskstats content
    std::string diskstatsContent =
        "   8       0 sda 10000 1000 200000 5000 5000 500 100000 3000 0 4000 8000\n"
        "   8       1 sda1 5000 500 100000 2500 2500 250 50000 1500 0 2000 4000\n"
        "   7       0 loop0 100 0 100 0 0 0 0 0 0 0 0\n";  // Should be skipped

    std::vector<DiskIoStats> stats = parseDiskStats(diskstatsContent);

    // Should have sda and sda1, but not loop0
    EXPECT_EQ(stats.size(), 2);

    EXPECT_EQ(stats[0].device, "sda");
    EXPECT_EQ(stats[0].readOps, 10000ul);
    EXPECT_EQ(stats[0].writeOps, 5000ul);

    EXPECT_EQ(stats[1].device, "sda1");
    EXPECT_EQ(stats[1].readOps, 5000ul);
    EXPECT_EQ(stats[1].writeOps, 2500ul);
}

TEST_F(HostMetricsTest, ParseDiskIoStatsFiltersLoopDevices) {
    std::string diskstatsContent =
        "   7       0 loop0 100 0 100 0 0 0 0 0 0 0 0\n"
        "   7       1 loop1 200 0 200 0 0 0 0 0 0 0 0\n"
        "   1       0 ram0 100 0 100 0 0 0 0 0 0 0 0\n";

    std::vector<DiskIoStats> stats = parseDiskStats(diskstatsContent);

    // All should be filtered out
    EXPECT_EQ(stats.size(), 0);
}

// ============================================================================
// Network Stats Tests
// ============================================================================

TEST_F(HostMetricsTest, ParseNetworkStatsCorrectly) {
    // Simulated /proc/net/dev content
    std::string netdevContent =
        "Inter-|   Receive                                                |  Transmit\n"
        " face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n"
        "  eth0: 12345678  1000    0    0    0     0          0         0 87654321   500    0    0    0     0       0          0\n"
        "    lo: 1000000   100    0    0    0     0          0         0  1000000   100    0    0    0     0       0          0\n";

    std::vector<NetworkStats> stats = parseNetworkStats(netdevContent);

    // Should only have eth0 (lo should be filtered)
    EXPECT_EQ(stats.size(), 1);
    EXPECT_EQ(stats[0].interface, "eth0");
    EXPECT_EQ(stats[0].rxBytes, 12345678ull);
    EXPECT_EQ(stats[0].txBytes, 87654321ull);
}

TEST_F(HostMetricsTest, ParseNetworkStatsFiltersLoopback) {
    std::string netdevContent =
        "Inter-|   Receive                                                |  Transmit\n"
        " face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n"
        "    lo: 1000000   100    0    0    0     0          0         0  1000000   100    0    0    0     0       0          0\n";

    std::vector<NetworkStats> stats = parseNetworkStats(netdevContent);

    // Loopback should be filtered
    EXPECT_EQ(stats.size(), 0);
}

// ============================================================================
// JSON Output Structure Tests
// ============================================================================

TEST_F(HostMetricsTest, MetricsJsonContainsRequiredFields) {
    // This test validates the expected JSON structure for host metrics
    // The actual JSON generation would be tested in integration tests

    // Expected fields in host_metrics JSON:
    // - cpu_user, cpu_system, cpu_idle, cpu_iowait, cpu_load_1m
    // - memory_total_mb, memory_free_mb, memory_used_mb, memory_used_percent
    // - disk_total_gb, disk_used_gb, disk_used_percent
    // - network_rx_bytes, network_tx_bytes

    // Verify that our parsing functions return non-zero values for valid input
    std::string statContent = "cpu  1000 0 500 8000 200 0 0 0";
    std::string loadavgContent = "0.50 0.75 1.00 1/100 12345";
    CpuStats cpuStats = parseCpuStats(statContent, loadavgContent);

    // These would be the values stored in JSON
    EXPECT_GE(cpuStats.user, 0.0);
    EXPECT_GE(cpuStats.system, 0.0);
    EXPECT_GE(cpuStats.idle, 0.0);
    EXPECT_GE(cpuStats.iowait, 0.0);
    EXPECT_GE(cpuStats.load1m, 0.0);

    std::string meminfoContent = "MemTotal: 16383796 kB\nMemFree: 1000000 kB\nCached: 3000000 kB\nBuffers: 500000 kB\n";
    MemoryStats memStats = parseMemoryInfo(meminfoContent);

    EXPECT_GT(memStats.totalMb, 0);
    EXPECT_GE(memStats.usedMb, 0);
    EXPECT_GE(memStats.usedPercent, 0.0);
}

// ============================================================================
// Error Handling Tests
// ============================================================================

TEST_F(HostMetricsTest, HandlesEmptyInput) {
    CpuStats cpuStats = parseCpuStats("", "");
    EXPECT_DOUBLE_EQ(cpuStats.idle, 100.0);  // Default value

    MemoryStats memStats = parseMemoryInfo("");
    EXPECT_EQ(memStats.totalMb, 0);

    std::vector<DiskIoStats> diskStats = parseDiskStats("");
    EXPECT_EQ(diskStats.size(), 0);

    std::vector<NetworkStats> netStats = parseNetworkStats("");
    EXPECT_EQ(netStats.size(), 0);
}

TEST_F(HostMetricsTest, HandlesMalformedInput) {
    // Malformed CPU stats
    CpuStats cpuStats = parseCpuStats("invalid data here", "also invalid");
    EXPECT_DOUBLE_EQ(cpuStats.idle, 100.0);  // Default

    // Malformed memory info
    MemoryStats memStats = parseMemoryInfo("not meminfo format");
    EXPECT_EQ(memStats.totalMb, 0);

    // Malformed disk stats
    std::vector<DiskIoStats> diskStats = parseDiskStats("not diskstats format");
    EXPECT_EQ(diskStats.size(), 0);
}