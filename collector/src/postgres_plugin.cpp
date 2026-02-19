#include "../include/collector.h"
#include <iostream>

SysstatCollector::SysstatCollector(
    const std::string& hostname,
    const std::string& collectorId
)
    : enabled_(true) {
    hostname_ = hostname;
    collectorId_ = collectorId;
}

json SysstatCollector::execute() {
    return json::object();
}

json SysstatCollector::collectCpuStats() {
    return json::object();
}

json SysstatCollector::collectMemoryStats() {
    return json::object();
}

json SysstatCollector::collectIoStats() {
    return json::object();
}

json SysstatCollector::collectLoadAverage() {
    return json::object();
}
