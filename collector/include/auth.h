#pragma once

#include <string>
#include <ctime>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

/**
 * Authentication Manager
 * Handles JWT token generation, validation, and refresh
 * Manages mTLS certificates for secure communication
 */
class AuthManager {
public:
    /**
     * Create authentication manager
     * @param collectorId Unique collector identifier
     * @param collectorSecret Shared secret for signing JWTs
     */
    explicit AuthManager(const std::string& collectorId, const std::string& collectorSecret = "");

    /**
     * Generate a new JWT token for collector authentication
     * @param expiresInSeconds Token expiration time (default: 3600 seconds / 1 hour)
     * @return JWT token string
     */
    std::string generateToken(int expiresInSeconds = 3600);

    /**
     * Get the current valid token (refresh if needed)
     * @return Valid JWT token
     */
    std::string getToken();

    /**
     * Refresh the JWT token
     * @return true if successful
     */
    bool refreshToken();

    /**
     * Set an external JWT token (useful for testing or external auth)
     */
    void setToken(const std::string& token, time_t expiresAt);

    /**
     * Check if current token is still valid
     * @return true if token exists and not expired
     */
    bool isTokenValid() const;

    /**
     * Get token expiration time
     * @return Unix timestamp of expiration
     */
    time_t getTokenExpiration() const;

    /**
     * Validate a JWT token signature
     * @param token Token to validate
     * @return true if signature is valid
     */
    bool validateTokenSignature(const std::string& token) const;

    /**
     * Load mTLS certificate from file
     * @param certFilePath Path to certificate file
     * @return true if successful
     */
    bool loadClientCertificate(const std::string& certFilePath);

    /**
     * Load mTLS private key from file
     * @param keyFilePath Path to private key file
     * @return true if successful
     */
    bool loadClientKey(const std::string& keyFilePath);

    /**
     * Get the loaded client certificate
     */
    std::string getClientCertificate() const;

    /**
     * Get the loaded client key
     */
    std::string getClientKey() const;

    /**
     * Get last error message
     */
    std::string getLastError() const;

private:
    std::string collectorId_;
    std::string collectorSecret_;
    std::string currentToken_;
    time_t tokenExpiresAt_;
    std::string clientCertificate_;
    std::string clientKey_;
    mutable std::string lastError_;

    /**
     * Encode data using HMAC-SHA256
     * @param data Data to sign
     * @param secret Secret key
     * @return Base64-encoded signature
     */
    static std::string hmacSha256(const std::string& data, const std::string& secret);

    /**
     * Base64 encode a string
     */
    static std::string base64Encode(const std::string& input);

    /**
     * Base64 decode a string
     */
    static std::string base64Decode(const std::string& input);

    /**
     * Create JWT payload
     */
    json createTokenPayload(time_t expiresAt) const;

    /**
     * Parse JWT token into header, payload, signature
     */
    static bool parseJwt(
        const std::string& token,
        std::string& header,
        std::string& payload,
        std::string& signature
    );
};
