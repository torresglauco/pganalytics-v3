#include "../include/metrics_buffer.h"
#include <zlib.h>
#include <cstring>
#include <iostream>
#include <sstream>

MetricsBuffer::MetricsBuffer(size_t maxSizeBytes)
    : maxSizeBytes_(maxSizeBytes),
      currentSizeBytes_(0),
      lastCompressedSize_(0) {
}

MetricsBuffer::~MetricsBuffer() {
    clear();
}

bool MetricsBuffer::append(const json& metrics) {
    size_t jsonSize = calculateJsonSize(metrics);

    // Check if adding this metric would exceed capacity
    if (currentSizeBytes_ + jsonSize > maxSizeBytes_) {
        return false;
    }

    metrics_.push_back(metrics);
    currentSizeBytes_ += jsonSize;
    return true;
}

bool MetricsBuffer::getUncompressed(json& metrics) {
    if (metrics_.empty()) {
        return false;
    }

    // Serialize all metrics to JSON array
    metrics = json::array();
    for (const auto& metric : metrics_) {
        metrics.push_back(metric);
    }

    return true;
}

bool MetricsBuffer::getCompressed(std::string& compressed) {
    if (metrics_.empty()) {
        compressed.clear();
        lastCompressedSize_ = 0;
        return true;
    }

    // Serialize all metrics to JSON array
    json metricsArray = json::array();
    for (const auto& metric : metrics_) {
        metricsArray.push_back(metric);
    }

    std::string uncompressed = metricsArray.dump();

    // Compress using gzip
    if (!compressData(uncompressed, compressed)) {
        return false;
    }

    lastCompressedSize_ = compressed.size();
    return true;
}

bool MetricsBuffer::compressData(const std::string& input, std::string& output) {
    // Allocate output buffer (estimate: input size * 0.9 for compression, plus header)
    size_t compressedSize = compressBound(input.size());
    output.resize(compressedSize);

    // Compress using zlib with gzip format
    int result = compress2(
        reinterpret_cast<unsigned char*>(output.data()),
        &compressedSize,
        reinterpret_cast<const unsigned char*>(input.c_str()),
        input.size(),
        6  // Compression level (1-9, 6 is default)
    );

    if (result != Z_OK) {
        output.clear();
        return false;
    }

    output.resize(compressedSize);
    return true;
}

size_t MetricsBuffer::calculateJsonSize(const json& obj) {
    // Serialize and get string size
    return obj.dump().size();
}

size_t MetricsBuffer::getUncompressedSize() const {
    return currentSizeBytes_;
}

size_t MetricsBuffer::getEstimatedCompressedSize() const {
    return lastCompressedSize_;
}

double MetricsBuffer::getCompressionRatio() const {
    if (currentSizeBytes_ == 0) {
        return 0.0;
    }
    return (static_cast<double>(lastCompressedSize_) / static_cast<double>(currentSizeBytes_)) * 100.0;
}

bool MetricsBuffer::isEmpty() const {
    return metrics_.empty();
}

bool MetricsBuffer::isFull() const {
    return currentSizeBytes_ >= maxSizeBytes_;
}

void MetricsBuffer::clear() {
    metrics_.clear();
    currentSizeBytes_ = 0;
    lastCompressedSize_ = 0;
}

size_t MetricsBuffer::getMetricCount() const {
    return metrics_.size();
}

json MetricsBuffer::getStats() const {
    return json{
        {"metric_count", getMetricCount()},
        {"uncompressed_size_bytes", getUncompressedSize()},
        {"compressed_size_bytes", getEstimatedCompressedSize()},
        {"max_size_bytes", maxSizeBytes_},
        {"compression_ratio_percent", getCompressionRatio()},
        {"is_empty", isEmpty()},
        {"is_full", isFull()}
    };
}
