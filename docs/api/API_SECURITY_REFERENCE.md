# API Security Reference

**pgAnalytics-v3 API**

Comprehensive security requirements and authentication flow documentation for all API endpoints.

---

## Table of Contents

1. [Authentication Flows](#authentication-flows)
2. [Rate Limiting](#rate-limiting)
3. [Endpoint Security Matrix](#endpoint-security-matrix)
4. [Error Handling](#error-handling)
5. [Security Headers](#security-headers)
6. [OWASP Mapping](#owasp-mapping)

---

## Authentication Flows

### User Authentication Flow

User login endpoint for obtaining access tokens.

```
POST /api/v1/auth/login
```

**Security Requirements:**

- No authentication required (public endpoint)
- HTTPS only (enforced by reverse proxy)
- Rate limited: 10 failed attempts = temporary lockout

**Request:**

```json
{
  "username": "user@example.com",
  "password": "password123"
}
```

**Success Response (200):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-24T22:15:00Z",
  "user": {
    "id": 1,
    "username": "user@example.com",
    "email": "user@example.com",
    "role": "user",
    "is_active": true
  }
}
```

**Error Responses:**

| Status | Condition | Response |
|--------|-----------|----------|
| 400 | Missing username or password | `{"error":"Bad Request","message":"Username and password required"}` |
| 401 | Invalid credentials | `{"error":"Unauthorized","message":"Invalid username or password"}` |
| 401 | Account inactive | `{"error":"Unauthorized","message":"User account is inactive"}` |
| 429 | Too many failed attempts | `{"error":"Too many requests","message":"Too many failed login attempts. Try again later"}` |

**Security Notes:**

- ❌ Password is never stored in response
- ❌ Errors don't reveal if username exists
- ✅ Bcrypt cost 12 password hashing
- ✅ Tokens expire after 15 minutes
- ✅ Refresh tokens expire after 24 hours

---

### Token Refresh Flow

Refresh access token using refresh token.

```
POST /api/v1/auth/refresh
```

**Security Requirements:**

- No authentication required (uses refresh token in body)
- HTTPS only
- Refresh token must be valid and not expired

**Request:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response (200):**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-24T22:30:00Z",
  "user": {"id": 1, "username": "user@example.com", "role": "user"}
}
```

**Error Responses:**

| Status | Condition |
|--------|-----------|
| 400 | Missing refresh_token |
| 401 | Refresh token expired |
| 401 | Refresh token invalid signature |
| 401 | User account deleted or inactive |

---

### Collector Registration Flow

Register new collector for metrics collection.

```
POST /api/v1/collectors/register
```

**Security Requirements:**

- ✅ Requires `X-Registration-Secret` header
- ✅ Secret must match server config: `REGISTRATION_SECRET`
- ✅ HTTPS only
- ✅ Returns mTLS certificate for future requests

**Request Header:**

```
X-Registration-Secret: ${REGISTRATION_SECRET}
Content-Type: application/json
```

**Request Body:**

```json
{
  "name": "prod-db-001",
  "hostname": "db.example.com",
  "address": "192.168.1.10",
  "version": "3.0.0"
}
```

**Success Response (200):**

```json
{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "certificate": "-----BEGIN CERTIFICATE-----\nMIIBkTCB+wIJA...\n-----END CERTIFICATE-----",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBA...\n-----END PRIVATE KEY-----",
  "expires_at": "2027-02-24T00:00:00Z"
}
```

**Error Responses:**

| Status | Condition |
|--------|-----------|
| 400 | Missing required fields (name, hostname) |
| 401 | Invalid or missing X-Registration-Secret |
| 401 | Registration secret doesn't match config |
| 409 | Collector with this hostname already registered |
| 500 | Certificate generation failed |

---

### Collector Metrics Push

Push metrics from collector to backend.

```
POST /api/v1/metrics/push
```

**Security Requirements:**

- ✅ Requires JWT bearer token from collector registration
- ✅ Token must include collector_id claim
- ✅ Claim collector_id must match request collector_id
- ✅ HTTPS only
- ✅ Rate limited: 1000 requests/minute per collector

**Request Header:**

```
Authorization: Bearer ${COLLECTOR_TOKEN}
Content-Type: application/json
```

**Request Body:**

```json
{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "metrics": [
    {
      "type": "pg_query_stats",
      "database": "postgres",
      "timestamp": "2026-02-24T20:30:00Z",
      "queries": [
        {
          "hash": 1234567890,
          "text": "SELECT * FROM users WHERE id = $1",
          "calls": 150,
          "total_time": 2500.5,
          "mean_time": 16.67,
          "min_time": 5.2,
          "max_time": 450.8,
          "stddev_time": 45.2,
          "rows": 150,
          "shared_blks_hit": 15000,
          "shared_blks_read": 500,
          "shared_blks_dirtied": 0,
          "shared_blks_written": 0,
          "local_blks_hit": 0,
          "local_blks_read": 0,
          "local_blks_dirtied": 0,
          "local_blks_written": 0,
          "temp_blks_read": 0,
          "temp_blks_written": 0,
          "blk_read_time": 12.5,
          "blk_write_time": 0,
          "wal_records": 0,
          "wal_fpi": 0,
          "wal_bytes": 0,
          "query_plan_time": 2.3,
          "query_exec_time": 14.37
        }
      ]
    }
  ]
}
```

**Success Response (200):**

```json
{
  "status": "success",
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "metrics_inserted": 150,
  "bytes_received": 14287,
  "processing_time_ms": 145,
  "next_config_version": 1,
  "next_check_in_seconds": 300
}
```

**Error Responses:**

| Status | Condition |
|--------|-----------|
| 400 | Invalid metrics JSON structure |
| 401 | Missing Authorization header |
| 401 | Bearer token invalid/expired |
| 401 | Collector ID mismatch (token vs request) |
| 401 | Collector inactive or deleted |
| 413 | Request payload too large (>100MB) |
| 429 | Rate limit exceeded |

---

## Rate Limiting

### Rate Limit Configuration

All authenticated endpoints enforce per-user rate limits.

**User Limits:**

```
100 requests / 60 seconds
Limit per IP: 1000 requests / 60 seconds
```

**Collector Limits:**

```
1000 requests / 60 seconds
Metrics push endpoint: 10 requests / 60 seconds (high-volume)
```

### Rate Limit Headers

All responses include rate limit status:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 47
X-RateLimit-Reset: 1645627800
```

### Rate Limit Exceeded Response

```
HTTP/1.1 429 Too Many Requests

{
  "error": "Too many requests",
  "message": "Rate limit exceeded. Please try again later.",
  "code": 429,
  "retry_after_seconds": 55
}
```

---

## Endpoint Security Matrix

### Authentication Requirements

| Endpoint | Method | Auth | Role | Rate Limit | Notes |
|----------|--------|------|------|-----------|-------|
| `/api/v1/health` | GET | ❌ | - | ✅ User | Health check |
| `/version` | GET | ❌ | - | ❌ | Version info |
| `/api/v1/auth/login` | POST | ❌ | - | ⚠️ Strict | 10 attempts/hour |
| `/api/v1/auth/logout` | POST | ✅ JWT | Any | ✅ User | Logout token |
| `/api/v1/auth/refresh` | POST | ❌ | - | ✅ User | Refresh token |
| `/api/v1/collectors/register` | POST | ✅ Secret | - | ⚠️ Strict | X-Registration-Secret |
| `/api/v1/collectors` | GET | ✅ JWT | user+ | ✅ User | List collectors |
| `/api/v1/collectors/:id` | GET | ✅ JWT | user+ | ✅ User | Get collector |
| `/api/v1/collectors/:id` | DELETE | ✅ JWT | admin | ✅ Admin | Delete collector |
| `/api/v1/metrics/push` | POST | ✅ JWT | collector | ✅ Collector | Push metrics |
| `/api/v1/metrics/cache` | GET | ✅ JWT | user+ | ✅ User | Cache metrics |
| `/api/v1/config/:collector_id` | GET | ✅ JWT | user+ | ✅ User | Get config |
| `/api/v1/config/:collector_id` | PUT | ✅ JWT | admin | ✅ Admin | Update config |
| `/api/v1/queries/:hash/timeline` | GET | ✅ JWT | user+ | ✅ User | Query timeline |

### Access Control

```
admin:   Can access ALL endpoints, all resources
user:    Can view metrics, dashboards, perform analysis
viewer:  Can view metrics and dashboards (read-only)
collector: Can push metrics, receive configuration
```

---

## Error Handling

### Error Response Format

All error responses follow standard format:

```json
{
  "error": "Short error name",
  "message": "User-friendly error message",
  "code": 400,
  "details": "Optional detailed explanation"
}
```

### HTTP Status Codes

| Code | Meaning | Security Implication |
|------|---------|----------------------|
| 400 | Bad Request | Invalid input validation |
| 401 | Unauthorized | Missing or invalid credentials |
| 403 | Forbidden | Authenticated but insufficient permissions |
| 404 | Not Found | Resource doesn't exist (don't leak existence) |
| 409 | Conflict | Resource already exists |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | No stack trace leaked to client |

### Security Best Practices

❌ **Never expose:**
- Stack traces
- SQL query details
- System file paths
- Version numbers of dependencies
- Database schema information
- Whether a username exists (use generic 401)

✅ **Always ensure:**
- Error messages are JSON-encoded
- Special characters are escaped
- User input is not echoed back
- Timing attacks are prevented (timing-safe comparison)

---

## Security Headers

All responses include security headers:

```
X-Frame-Options: DENY
  → Prevents clickjacking attacks

X-Content-Type-Options: nosniff
  → Prevents MIME type sniffing

X-XSS-Protection: 1; mode=block
  → Enable XSS protection in browsers

Referrer-Policy: strict-origin-when-cross-origin
  → Control referrer information leakage

Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; ...
  → Restrict resource loading origins

Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
  → Force HTTPS (production only)
```

---

## OWASP Mapping

### OWASP Top 10 Coverage

| OWASP | Risk | Status | Mitigation |
|-------|------|--------|-----------|
| A1: Broken Access Control | Unauthorized access | ✅ Fixed | RBAC enforced |
| A2: Cryptographic Failure | Data exposure | ✅ Protected | TLS + encrypted storage |
| A3: Injection | SQL injection | ✅ Protected | Parameterized queries |
| A4: Insecure Design | Security gaps | ✅ Addressed | JWT + mTLS |
| A5: Security Configuration | Misconfiguration | ✅ Checked | Config validation |
| A6: Vulnerable Components | Dependency exploits | ⚠️ In Progress | Dependency scanning |
| A7: Authentication Failure | Auth bypass | ✅ Fixed | Proper verification |
| A8: Software/Data Integrity | Compromised code | ✅ Monitored | Code review + testing |
| A9: Logging Failures | Undetected breaches | ⚠️ In Progress | Audit logging |
| A10: SSRF | Server-side request forgery | ✅ Protected | Input validation |

### CWE Top 25 Coverage

| CWE | Issue | Status |
|-----|-------|--------|
| CWE-79: XSS | Cross-site scripting | ✅ Protected (JSON-only API) |
| CWE-89: SQL Injection | Database injection | ✅ Protected (parameterized) |
| CWE-287: Auth Bypass | Broken authentication | ✅ Fixed (v3.1.0) |
| CWE-200: Information Exposure | Data leaks | ✅ Masked (no stack traces) |
| CWE-352: CSRF | Cross-site forgery | ✅ Protected (JWT tokens) |
| CWE-434: Unrestricted File Upload | File upload exploits | ✅ N/A (no file uploads) |
| CWE-502: Deserialization | Object injection | ✅ Protected (JSON only) |

---

## Testing Checklist

Use this checklist to validate security implementation:

```yaml
Authentication:
  ✅ POST /auth/login rejects invalid credentials
  ✅ POST /auth/login requires both username and password
  ✅ POST /auth/refresh rejects expired token
  ✅ POST /collectors/register requires X-Registration-Secret
  ✅ POST /metrics/push requires Authorization header

Authorization:
  ✅ Admin-only endpoints reject user role
  ✅ User endpoints reject viewer role
  ✅ GET /collectors/:id validates ownership
  ✅ DELETE /collectors/:id requires admin role

Rate Limiting:
  ✅ 101st request in 60s returns 429
  ✅ Rate limit resets after 60 seconds
  ✅ Collector rate limit is higher (1000/min)

Input Validation:
  ✅ SQL injection in query params is escaped
  ✅ XSS payload in error messages is escaped
  ✅ Large payloads (>100MB) are rejected
  ✅ Invalid JSON returns 400

Security Headers:
  ✅ X-Frame-Options present
  ✅ X-Content-Type-Options present
  ✅ Content-Security-Policy present
  ✅ Strict-Transport-Security in production
```

---

## Implementation Examples

### Authenticating as User

```bash
# 1. Login
curl -X POST https://api.pganalytics.dev/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user@example.com","password":"secret"}' \
  | jq -r '.token' > token.txt

# 2. Use token in subsequent requests
TOKEN=$(cat token.txt)
curl -X GET https://api.pganalytics.dev/api/v1/collectors \
  -H "Authorization: Bearer $TOKEN"
```

### Registering Collector

```bash
# 1. Register collector
curl -X POST https://api.pganalytics.dev/api/v1/collectors/register \
  -H "X-Registration-Secret: ${REGISTRATION_SECRET}" \
  -H "Content-Type: application/json" \
  -d '{"name":"prod-db","hostname":"db.example.com"}' \
  | jq . > collector.json

# 2. Extract collector ID and token
COLLECTOR_ID=$(jq -r '.collector_id' collector.json)
COLLECTOR_TOKEN=$(jq -r '.token' collector.json)

# 3. Push metrics
curl -X POST https://api.pganalytics.dev/api/v1/metrics/push \
  -H "Authorization: Bearer $COLLECTOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d @metrics.json
```

---

## Related Documentation

- [SECURITY.md](../../SECURITY.md) - Security policy and vulnerability disclosure
- [README.md](../../README.md) - Project overview and quick start
- [API Swagger/OpenAPI](./swagger.yaml) - Full API specification

---

**Last Updated:** February 24, 2026
**Classification:** INTERNAL
**Status:** ACTIVE
