#include "../include/auth.h"
#include <fstream>
#include <sstream>
#include <openssl/hmac.h>
#include <openssl/sha.h>
#include <openssl/bio.h>
#include <openssl/buffer.h>
#include <cstring>
#include <ctime>

AuthManager::AuthManager(const std::string& collectorId, const std::string& collectorSecret)
    : collectorId_(collectorId),
      collectorSecret_(collectorSecret),
      tokenExpiresAt_(0) {
}

std::string AuthManager::generateToken(int expiresInSeconds) {
    time_t now = std::time(nullptr);
    time_t expiresAt = now + expiresInSeconds;

    // Create JWT payload
    json payload = createTokenPayload(expiresAt);
    std::string payloadStr = payload.dump();

    // Base64 encode header and payload
    std::string headerStr = R"({"alg":"HS256","typ":"JWT"})";
    std::string header = base64Encode(headerStr);
    std::string encodedPayload = base64Encode(payloadStr);

    // Create signature
    std::string signatureInput = header + "." + encodedPayload;
    std::string signature = hmacSha256(signatureInput, collectorSecret_);

    // Combine all parts
    std::string token = signatureInput + "." + signature;

    currentToken_ = token;
    tokenExpiresAt_ = expiresAt;

    return token;
}

std::string AuthManager::getToken() {
    // Check if token is still valid
    if (!isTokenValid()) {
        // Generate new token
        return generateToken();
    }
    return currentToken_;
}

bool AuthManager::refreshToken() {
    if (collectorSecret_.empty()) {
        lastError_ = "Cannot refresh token: collector secret not set";
        return false;
    }
    generateToken();
    return true;
}

void AuthManager::setToken(const std::string& token, time_t expiresAt) {
    currentToken_ = token;
    tokenExpiresAt_ = expiresAt;
}

bool AuthManager::isTokenValid() const {
    if (currentToken_.empty() || tokenExpiresAt_ == 0) {
        return false;
    }

    time_t now = std::time(nullptr);
    // Add 60 second buffer for token expiration (refresh 1 min before expiry)
    return now < (tokenExpiresAt_ - 60);
}

time_t AuthManager::getTokenExpiration() const {
    return tokenExpiresAt_;
}

bool AuthManager::validateTokenSignature(const std::string& token) const {
    std::string header, payload, signature;
    if (!parseJwt(token, header, payload, signature)) {
        lastError_ = "Invalid JWT format";
        return false;
    }

    // Recalculate signature
    std::string signatureInput = header + "." + payload;
    std::string expectedSignature = hmacSha256(signatureInput, collectorSecret_);

    // Compare signatures (timing-safe comparison would be better for production)
    return signature == expectedSignature;
}

bool AuthManager::loadClientCertificate(const std::string& certFilePath) {
    std::ifstream file(certFilePath);
    if (!file.is_open()) {
        lastError_ = "Cannot open certificate file: " + certFilePath;
        return false;
    }

    std::stringstream buffer;
    buffer << file.rdbuf();
    clientCertificate_ = buffer.str();
    file.close();

    return true;
}

bool AuthManager::loadClientKey(const std::string& keyFilePath) {
    std::ifstream file(keyFilePath);
    if (!file.is_open()) {
        lastError_ = "Cannot open key file: " + keyFilePath;
        return false;
    }

    std::stringstream buffer;
    buffer << file.rdbuf();
    clientKey_ = buffer.str();
    file.close();

    return true;
}

std::string AuthManager::getClientCertificate() const {
    return clientCertificate_;
}

std::string AuthManager::getClientKey() const {
    return clientKey_;
}

std::string AuthManager::getLastError() const {
    return lastError_;
}

std::string AuthManager::hmacSha256(const std::string& data, const std::string& secret) {
    unsigned char result[EVP_MAX_MD_SIZE];
    unsigned int resultLen = 0;

    HMAC(
        EVP_sha256(),
        secret.c_str(),
        static_cast<int>(secret.length()),
        reinterpret_cast<const unsigned char*>(data.c_str()),
        data.length(),
        result,
        &resultLen
    );

    // Base64 encode the result
    return base64Encode(std::string(reinterpret_cast<const char*>(result), resultLen));
}

std::string AuthManager::base64Encode(const std::string& input) {
    BIO* bio = BIO_new(BIO_s_mem());
    BIO* b64 = BIO_new(BIO_f_base64());
    BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
    bio = BIO_push(b64, bio);

    BIO_write(bio, input.c_str(), static_cast<int>(input.length()));
    BIO_flush(bio);

    BUF_MEM* bufferPtr;
    BIO_get_mem_ptr(bio, &bufferPtr);

    std::string output(bufferPtr->data, bufferPtr->length);
    BIO_free_all(bio);

    return output;
}

std::string AuthManager::base64Decode(const std::string& input) {
    BIO* bio = BIO_new_mem_buf(input.c_str(), static_cast<int>(input.length()));
    BIO* b64 = BIO_new(BIO_f_base64());
    BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
    bio = BIO_push(b64, bio);

    std::string output(input.length(), '\0');
    int decodedLen = BIO_read(bio, &output[0], static_cast<int>(input.length()));
    BIO_free_all(bio);

    if (decodedLen < 0) {
        return "";
    }

    output.resize(decodedLen);
    return output;
}

json AuthManager::createTokenPayload(time_t expiresAt) const {
    time_t now = std::time(nullptr);

    json payload;
    payload["iss"] = "pganalytics-collector";
    payload["sub"] = collectorId_;
    payload["iat"] = static_cast<long long>(now);
    payload["exp"] = static_cast<long long>(expiresAt);
    payload["collector_id"] = collectorId_;

    return payload;
}

bool AuthManager::parseJwt(
    const std::string& token,
    std::string& header,
    std::string& payload,
    std::string& signature
) {
    // Split by '.'
    size_t firstDot = token.find('.');
    size_t secondDot = token.rfind('.');

    if (firstDot == std::string::npos || secondDot == std::string::npos || firstDot == secondDot) {
        return false;
    }

    header = token.substr(0, firstDot);
    payload = token.substr(firstDot + 1, secondDot - firstDot - 1);
    signature = token.substr(secondDot + 1);

    return true;
}
