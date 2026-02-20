#pragma once

#include <string>
#include <vector>
#include <memory>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * Metrics Buffer
 * Implements circular buffer with gzip compression for metrics before transmission
 * Provides buffering to handle temporary network outages and reduce bandwidth
 */
class MetricsBuffer {
public:
    /**
     * Create a metrics buffer with specified capacity
     * @param maxSizeBytes Maximum buffer size in bytes (default 10MB)
     */
    explicit MetricsBuffer(size_t maxSizeBytes = 10 * 1024 * 1024);

    /**
     * Destroy buffer and cleanup resources
     */
    ~MetricsBuffer();

    /**
     * Add metrics JSON object to buffer
     * @param metrics JSON object to buffer
     * @return true if added successfully, false if buffer is full
     */
    bool append(const json& metrics);

    /**
     * Compress and get all buffered metrics
     * @param compressed Output string with compressed data (gzip)
     * @return true if successful, false on error
     */
    bool getCompressed(std::string& compressed);

    /**
     * Get uncompressed size of buffered metrics
     * @return Size in bytes
     */
    size_t getUncompressedSize() const;

    /**
     * Get estimated compressed size (after gzip)
     * @return Size in bytes
     */
    size_t getEstimatedCompressedSize() const;

    /**
     * Get compression ratio
     * @return Percentage (0-100) of original size
     */
    double getCompressionRatio() const;

    /**
     * Check if buffer has any data
     */
    bool isEmpty() const;

    /**
     * Check if buffer is at capacity
     */
    bool isFull() const;

    /**
     * Clear all buffered data
     */
    void clear();

    /**
     * Get count of buffered metric objects
     */
    size_t getMetricCount() const;

    /**
     * Get detailed statistics about buffer state
     */
    json getStats() const;

private:
    size_t maxSizeBytes_;
    std::vector<json> metrics_;
    size_t currentSizeBytes_;
    size_t lastCompressedSize_;

    /**
     * Perform gzip compression on JSON data
     * @param input Uncompressed data
     * @param output Compressed data buffer
     * @return true if successful
     */
    bool compressData(const std::string& input, std::string& output);

    /**
     * Calculate size of a JSON object when serialized
     */
    static size_t calculateJsonSize(const json& obj);
};
