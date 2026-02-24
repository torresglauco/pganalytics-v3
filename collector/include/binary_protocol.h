#ifndef BINARY_PROTOCOL_H_
#define BINARY_PROTOCOL_H_

#include <cstdint>
#include <vector>
#include <string>
#include <cstring>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * Binary Protocol for pgAnalytics Collector
 *
 * Provides efficient serialization of metrics for transmission to backend.
 * Supports compression (zstd) and optional encryption (TLS overlay).
 *
 * Design Goals:
 * - 60% bandwidth reduction vs JSON+gzip
 * - <1ms serialization/deserialization
 * - Support 100,000+ collectors
 * - Backwards compatible with REST API
 */

// Magic number for protocol validation
constexpr uint32_t PROTOCOL_MAGIC = 0xDEADBEEF;

// Protocol version for future compatibility
constexpr uint32_t PROTOCOL_VERSION = 1;

// Message types
enum class MessageType : uint32_t {
    MetricsBatch = 1,
    ConfigRequest = 2,
    ConfigResponse = 3,
    RegistrationRequest = 4,
    RegistrationResponse = 5,
    HealthCheck = 6,
    HealthCheckResponse = 7,
};

// Compression algorithms
enum class CompressionType : uint8_t {
    None = 0,
    Zstd = 1,
    Snappy = 2,
};

/**
 * Message header for binary protocol
 * Total size: 32 bytes (cache-line aligned)
 *
 * Layout:
 * [0-3]:   Magic number (0xDEADBEEF)
 * [4-7]:   Version (1)
 * [8-11]:  Message type (MetricsBatch, ConfigRequest, etc.)
 * [12-15]: Payload length
 * [16-19]: CRC32 checksum of payload
 * [20]:    Compression type (0=none, 1=zstd, 2=snappy)
 * [21]:    Encryption flag
 * [22-31]: Reserved/padding
 */
struct MessageHeader {
    uint32_t magic;              // 0xDEADBEEF
    uint32_t version;            // Protocol version
    uint32_t message_type;       // MessageType enum value
    uint32_t payload_len;        // Length of payload in bytes
    uint32_t checksum_crc32;     // CRC32 of entire payload (for integrity check)
    uint8_t  compression;        // CompressionType enum value
    uint8_t  encrypted;          // 1 if encrypted, 0 otherwise
    uint8_t  reserved1;
    uint8_t  reserved2;
    uint32_t reserved3;
    uint32_t reserved4;          // Padding to 32 bytes

    /**
     * Initialize header with defaults
     */
    MessageHeader()
        : magic(PROTOCOL_MAGIC),
          version(PROTOCOL_VERSION),
          message_type(static_cast<uint32_t>(MessageType::MetricsBatch)),
          payload_len(0),
          checksum_crc32(0),
          compression(static_cast<uint8_t>(CompressionType::None)),
          encrypted(0),
          reserved1(0),
          reserved2(0),
          reserved3(0),
          reserved4(0) {}

    /**
     * Validate header integrity
     */
    bool validate() const {
        return magic == PROTOCOL_MAGIC && version == PROTOCOL_VERSION;
    }

    /**
     * Serialize header to binary buffer (32 bytes)
     */
    std::vector<uint8_t> serialize() const {
        std::vector<uint8_t> buffer(32);
        uint8_t* ptr = buffer.data();

        // Write as little-endian
        std::memcpy(ptr, &magic, 4); ptr += 4;
        std::memcpy(ptr, &version, 4); ptr += 4;
        std::memcpy(ptr, &message_type, 4); ptr += 4;
        std::memcpy(ptr, &payload_len, 4); ptr += 4;
        std::memcpy(ptr, &checksum_crc32, 4); ptr += 4;
        *ptr++ = compression;
        *ptr++ = encrypted;
        *ptr++ = reserved1;
        *ptr++ = reserved2;
        std::memcpy(ptr, &reserved3, 4); ptr += 4;
        std::memcpy(ptr, &reserved4, 4); ptr += 4;

        return buffer;
    }

    /**
     * Deserialize header from binary buffer
     */
    static MessageHeader deserialize(const std::vector<uint8_t>& buffer) {
        MessageHeader header;
        if (buffer.size() < 32) {
            header.magic = 0; // Invalid
            return header;
        }

        const uint8_t* ptr = buffer.data();
        std::memcpy(&header.magic, ptr, 4); ptr += 4;
        std::memcpy(&header.version, ptr, 4); ptr += 4;
        std::memcpy(&header.message_type, ptr, 4); ptr += 4;
        std::memcpy(&header.payload_len, ptr, 4); ptr += 4;
        std::memcpy(&header.checksum_crc32, ptr, 4); ptr += 4;
        header.compression = *ptr++;
        header.encrypted = *ptr++;
        header.reserved1 = *ptr++;
        header.reserved2 = *ptr++;
        std::memcpy(&header.reserved3, ptr, 4); ptr += 4;
        std::memcpy(&header.reserved4, ptr, 4);

        return header;
    }
} __attribute__((packed));

static_assert(sizeof(MessageHeader) == 32, "MessageHeader must be exactly 32 bytes");

/**
 * Binary encoding for metric value
 * Efficiently encodes common metric types
 *
 * Type byte:
 * 0x00: Null
 * 0x01: Boolean (1 byte data)
 * 0x02: Int32 (4 bytes)
 * 0x03: Int64 (8 bytes)
 * 0x04: Float32 (4 bytes)
 * 0x05: Float64 (8 bytes)
 * 0x06: String (varint length + data)
 */
class MetricEncoder {
public:
    /**
     * Encode metric snapshot to binary format
     * Returns serialized bytes
     */
    static std::vector<uint8_t> encodeMetrics(const json& metrics);

    /**
     * Decode binary format back to JSON
     */
    static json decodeMetrics(const std::vector<uint8_t>& data);

    /**
     * Encode single value with type information
     */
    static std::vector<uint8_t> encodeValue(const json& value);

    /**
     * Decode single value
     */
    static json decodeValue(const uint8_t* data, size_t& offset);

private:
    /**
     * Encode variable-length integer (varint)
     * Used for string lengths, array sizes
     */
    static std::vector<uint8_t> encodeVarint(uint64_t value);

    /**
     * Decode variable-length integer
     */
    static uint64_t decodeVarint(const uint8_t* data, size_t& offset);

    /**
     * Encode string with length prefix
     */
    static std::vector<uint8_t> encodeString(const std::string& str);

    /**
     * Decode string with length prefix
     */
    static std::string decodeString(const uint8_t* data, size_t& offset);

    /**
     * Helper to append bytes to buffer
     */
    static void append(std::vector<uint8_t>& buffer, const uint8_t* data, size_t len);
};

/**
 * Message builder for creating protocol messages
 */
class MessageBuilder {
public:
    /**
     * Create metrics batch message
     */
    static std::vector<uint8_t> createMetricsBatch(
        const std::string& collector_id,
        const std::string& hostname,
        const std::string& version,
        const std::vector<json>& metrics,
        CompressionType compression = CompressionType::Zstd
    );

    /**
     * Create health check message
     */
    static std::vector<uint8_t> createHealthCheck(
        const std::string& collector_id,
        uint32_t memory_mb,
        uint32_t cpu_percent
    );

    /**
     * Create registration request
     */
    static std::vector<uint8_t> createRegistrationRequest(
        const std::string& hostname,
        const std::string& api_key
    );

private:
    /**
     * Build complete message with header + payload
     */
    static std::vector<uint8_t> buildMessage(
        MessageType type,
        const std::vector<uint8_t>& payload,
        CompressionType compression
    );
};

/**
 * CRC32 checksum for data integrity
 */
class Checksum {
public:
    /**
     * Calculate CRC32 checksum
     */
    static uint32_t crc32(const uint8_t* data, size_t len);

    /**
     * Verify CRC32 checksum
     */
    static bool verifyCrc32(const uint8_t* data, size_t len, uint32_t expected_crc);
};

/**
 * Compression utilities
 */
class CompressionUtil {
public:
    /**
     * Compress data using specified algorithm
     */
    static std::vector<uint8_t> compress(
        const std::vector<uint8_t>& data,
        CompressionType type
    );

    /**
     * Decompress data
     */
    static std::vector<uint8_t> decompress(
        const std::vector<uint8_t>& data,
        CompressionType type
    );

    /**
     * Get compression ratio (0-100)
     */
    static int getCompressionRatio(
        size_t original_size,
        size_t compressed_size
    );
};

#endif  // BINARY_PROTOCOL_H_
