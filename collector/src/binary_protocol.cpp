#include "../include/binary_protocol.h"
#include <zstd.h>
#include <cstring>
#include <iostream>
#include <algorithm>

// Simple CRC32 implementation (polynomial 0xEDB88320)
static uint32_t crc32_table[256];
static bool crc32_initialized = false;

static void crc32_init() {
    if (crc32_initialized) return;
    for (int i = 0; i < 256; i++) {
        uint32_t crc = i;
        for (int j = 0; j < 8; j++) {
            if (crc & 1) {
                crc = (crc >> 1) ^ 0xEDB88320;
            } else {
                crc >>= 1;
            }
        }
        crc32_table[i] = crc;
    }
    crc32_initialized = true;
}

// ============================================================================
// MetricEncoder Implementation
// ============================================================================

std::vector<uint8_t> MetricEncoder::encodeVarint(uint64_t value) {
    std::vector<uint8_t> result;

    while (value >= 0x80) {
        result.push_back((uint8_t)((value & 0x7F) | 0x80));
        value >>= 7;
    }
    result.push_back((uint8_t)value);

    return result;
}

uint64_t MetricEncoder::decodeVarint(const uint8_t* data, size_t& offset) {
    uint64_t result = 0;
    int shift = 0;

    while (offset < 1024) {  // Safety limit
        uint8_t byte = data[offset++];
        result |= ((uint64_t)(byte & 0x7F)) << shift;

        if ((byte & 0x80) == 0) {
            break;
        }
        shift += 7;
    }

    return result;
}

std::vector<uint8_t> MetricEncoder::encodeString(const std::string& str) {
    auto length_bytes = encodeVarint(str.length());
    std::vector<uint8_t> result = length_bytes;
    result.insert(result.end(), str.begin(), str.end());
    return result;
}

std::string MetricEncoder::decodeString(const uint8_t* data, size_t& offset) {
    uint64_t length = decodeVarint(data, offset);
    std::string result((const char*)&data[offset], length);
    offset += length;
    return result;
}

void MetricEncoder::append(std::vector<uint8_t>& buffer, const uint8_t* data, size_t len) {
    buffer.insert(buffer.end(), data, data + len);
}

std::vector<uint8_t> MetricEncoder::encodeValue(const json& value) {
    std::vector<uint8_t> result;

    if (value.is_null()) {
        result.push_back(0x00);
    } else if (value.is_boolean()) {
        result.push_back(0x01);
        result.push_back(value.get<bool>() ? 1 : 0);
    } else if (value.is_number_integer()) {
        int64_t val = value.get<int64_t>();

        // Use smallest size that fits
        if (val >= INT32_MIN && val <= INT32_MAX) {
            result.push_back(0x02);  // Int32
            int32_t v32 = (int32_t)val;
            result.resize(5);
            std::memcpy(result.data() + 1, &v32, 4);
        } else {
            result.push_back(0x03);  // Int64
            result.resize(9);
            std::memcpy(result.data() + 1, &val, 8);
        }
    } else if (value.is_number_float()) {
        result.push_back(0x05);  // Float64
        double val = value.get<double>();
        result.resize(9);
        std::memcpy(result.data() + 1, &val, 8);
    } else if (value.is_string()) {
        result.push_back(0x06);  // String
        auto str_bytes = encodeString(value.get<std::string>());
        result.insert(result.end(), str_bytes.begin(), str_bytes.end());
    } else if (value.is_array()) {
        result.push_back(0x07);  // Array
        auto length_bytes = encodeVarint(value.size());
        result.insert(result.end(), length_bytes.begin(), length_bytes.end());

        for (const auto& elem : value) {
            auto elem_bytes = encodeValue(elem);
            result.insert(result.end(), elem_bytes.begin(), elem_bytes.end());
        }
    } else if (value.is_object()) {
        result.push_back(0x08);  // Object
        auto length_bytes = encodeVarint(value.size());
        result.insert(result.end(), length_bytes.begin(), length_bytes.end());

        for (auto it = value.begin(); it != value.end(); ++it) {
            auto key_bytes = encodeString(it.key());
            auto val_bytes = encodeValue(it.value());
            result.insert(result.end(), key_bytes.begin(), key_bytes.end());
            result.insert(result.end(), val_bytes.begin(), val_bytes.end());
        }
    }

    return result;
}

json MetricEncoder::decodeValue(const uint8_t* data, size_t& offset) {
    uint8_t type = data[offset++];

    switch (type) {
        case 0x00:  // Null
            return json(nullptr);

        case 0x01: {  // Boolean
            bool val = data[offset++] != 0;
            return json(val);
        }

        case 0x02: {  // Int32
            int32_t val;
            std::memcpy(&val, &data[offset], 4);
            offset += 4;
            return json(val);
        }

        case 0x03: {  // Int64
            int64_t val;
            std::memcpy(&val, &data[offset], 8);
            offset += 8;
            return json(val);
        }

        case 0x05: {  // Float64
            double val;
            std::memcpy(&val, &data[offset], 8);
            offset += 8;
            return json(val);
        }

        case 0x06:  // String
            return json(decodeString(data, offset));

        case 0x07: {  // Array
            size_t length = decodeVarint(data, offset);
            json arr = json::array();
            for (size_t i = 0; i < length; i++) {
                arr.push_back(decodeValue(data, offset));
            }
            return arr;
        }

        case 0x08: {  // Object
            size_t length = decodeVarint(data, offset);
            json obj = json::object();
            for (size_t i = 0; i < length; i++) {
                std::string key = decodeString(data, offset);
                json val = decodeValue(data, offset);
                obj[key] = val;
            }
            return obj;
        }

        default:
            return json(nullptr);
    }
}

std::vector<uint8_t> MetricEncoder::encodeMetrics(const json& metrics) {
    return encodeValue(metrics);
}

json MetricEncoder::decodeMetrics(const std::vector<uint8_t>& data) {
    size_t offset = 0;
    return decodeValue(data.data(), offset);
}

// ============================================================================
// MessageBuilder Implementation
// ============================================================================

std::vector<uint8_t> MessageBuilder::buildMessage(
    MessageType type,
    const std::vector<uint8_t>& payload,
    CompressionType compression
) {
    // Compress payload if needed
    std::vector<uint8_t> compressed_payload = payload;
    if (compression != CompressionType::None) {
        compressed_payload = CompressionUtil::compress(payload, compression);
    }

    // Create header
    MessageHeader header;
    header.message_type = static_cast<uint32_t>(type);
    header.payload_len = compressed_payload.size();
    header.compression = static_cast<uint8_t>(compression);
    header.checksum_crc32 = Checksum::crc32(
        compressed_payload.data(),
        compressed_payload.size()
    );

    // Build complete message
    auto header_bytes = header.serialize();
    std::vector<uint8_t> message = header_bytes;
    message.insert(message.end(), compressed_payload.begin(), compressed_payload.end());

    return message;
}

std::vector<uint8_t> MessageBuilder::createMetricsBatch(
    const std::string& collector_id,
    const std::string& hostname,
    const std::string& version,
    const std::vector<json>& metrics,
    CompressionType compression
) {
    // Build payload
    json payload_obj;
    payload_obj["collector_id"] = collector_id;
    payload_obj["hostname"] = hostname;
    payload_obj["version"] = version;
    payload_obj["timestamp"] = std::time(nullptr);
    payload_obj["metrics"] = metrics;

    // Encode to binary
    auto encoded = MetricEncoder::encodeMetrics(payload_obj);

    // Build message
    return buildMessage(MessageType::MetricsBatch, encoded, compression);
}

std::vector<uint8_t> MessageBuilder::createHealthCheck(
    const std::string& collector_id,
    uint32_t memory_mb,
    uint32_t cpu_percent
) {
    json payload;
    payload["collector_id"] = collector_id;
    payload["timestamp"] = std::time(nullptr);
    payload["memory_mb"] = memory_mb;
    payload["cpu_percent"] = cpu_percent;

    auto encoded = MetricEncoder::encodeMetrics(payload);
    return buildMessage(MessageType::HealthCheck, encoded, CompressionType::None);
}

std::vector<uint8_t> MessageBuilder::createRegistrationRequest(
    const std::string& hostname,
    const std::string& api_key
) {
    json payload;
    payload["hostname"] = hostname;
    payload["api_key"] = api_key;
    payload["timestamp"] = std::time(nullptr);
    payload["protocol_version"] = PROTOCOL_VERSION;

    auto encoded = MetricEncoder::encodeMetrics(payload);
    return buildMessage(MessageType::RegistrationRequest, encoded, CompressionType::None);
}

// ============================================================================
// Checksum Implementation
// ============================================================================

uint32_t Checksum::crc32(const uint8_t* data, size_t len) {
    crc32_init();
    uint32_t crc = 0xFFFFFFFF;

    for (size_t i = 0; i < len; i++) {
        crc = (crc >> 8) ^ crc32_table[(crc ^ data[i]) & 0xFF];
    }

    return crc ^ 0xFFFFFFFF;
}

bool Checksum::verifyCrc32(const uint8_t* data, size_t len, uint32_t expected_crc) {
    return crc32(data, len) == expected_crc;
}

// ============================================================================
// CompressionUtil Implementation
// ============================================================================

std::vector<uint8_t> CompressionUtil::compress(
    const std::vector<uint8_t>& data,
    CompressionType type
) {
    switch (type) {
        case CompressionType::None:
            return data;

        case CompressionType::Zstd: {
            // Zstandard compression
            size_t bound = ZSTD_compressBound(data.size());
            std::vector<uint8_t> compressed(bound);

            size_t compressed_size = ZSTD_compress(
                compressed.data(),
                bound,
                data.data(),
                data.size(),
                3  // Compression level (1-22, 3 is default)
            );

            if (ZSTD_isError(compressed_size)) {
                std::cerr << "ZSTD compression error" << std::endl;
                return data;  // Fallback to uncompressed
            }

            compressed.resize(compressed_size);
            return compressed;
        }

        case CompressionType::Snappy: {
            // Snappy compression (if available)
            // For now, fallback to uncompressed
            std::cerr << "Snappy compression not yet implemented" << std::endl;
            return data;
        }

        default:
            return data;
    }
}

std::vector<uint8_t> CompressionUtil::decompress(
    const std::vector<uint8_t>& data,
    CompressionType type
) {
    switch (type) {
        case CompressionType::None:
            return data;

        case CompressionType::Zstd: {
            // Get original size
            unsigned long long original_size = ZSTD_getFrameContentSize(data.data(), data.size());

            if (original_size == ZSTD_CONTENTSIZE_ERROR) {
                std::cerr << "Invalid ZSTD frame" << std::endl;
                return {};
            }

            std::vector<uint8_t> decompressed(original_size);

            size_t result = ZSTD_decompress(
                decompressed.data(),
                original_size,
                data.data(),
                data.size()
            );

            if (ZSTD_isError(result)) {
                std::cerr << "ZSTD decompression error" << std::endl;
                return {};
            }

            return decompressed;
        }

        case CompressionType::Snappy: {
            // Snappy decompression (if available)
            std::cerr << "Snappy decompression not yet implemented" << std::endl;
            return {};
        }

        default:
            return {};
    }
}

int CompressionUtil::getCompressionRatio(size_t original_size, size_t compressed_size) {
    if (original_size == 0) return 0;
    return (int)((100 * compressed_size) / original_size);
}
