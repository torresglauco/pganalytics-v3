# API Security Reference Guide

**Date:** February 25, 2026
**Version:** 1.0
**Status:** Production-Ready

---

## Endpoint Security Matrix

### Overview

This guide documents security requirements for every API endpoint in pgAnalytics-v3, including authentication, authorization, and rate limiting specifications.

### Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Implemented and enforced |
| ⚠️ | Partial implementation or TODO |
| ❌ | Not implemented |

---

## Authentication Endpoints

### POST /api/v1/auth/login

**Description:** User login with username and password

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ❌ Public | No auth required for login |
| Authorization | ✅ RBAC | All roles can login |
| Rate Limit | ✅ 100/min | Per IP address |
| HTTPS | ✅ Required | TLS 1.2+ |

**Request Validation:**
```json
{
  "username": "string (required, 3-255 chars)",
  "password": "string (required, 8-255 chars)"
}
```

**Response:**
```json
{
  "access_token": "JWT token (15 min expiration)",
  "refresh_token": "JWT token (7 day expiration)",
  "expires_at": "RFC3339 timestamp",
  "user": {
    "id": "integer",
    "username": "string",
    "email": "string",
    "role": "admin|user|viewer"
  }
}
```

**Security Checks:**
1. Username exists in database
2. Password matches bcrypt hash (timing-safe comparison)
3. User account is active
4. Token claims populated correctly
5. Rate limited per IP

---

### POST /api/v1/auth/logout

**Description:** User logout and token invalidation

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT token in Authorization header |
| Authorization | ✅ RBAC | All authenticated users |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

**Header Required:**
```
Authorization: Bearer <jwt-token>
```

**Response:**
```json
{
  "message": "Logged out successfully"
}
```

**Note:** Token blacklist not yet implemented. Token remains valid until expiration (15 min).

---

### POST /api/v1/auth/refresh

**Description:** Refresh access token using refresh token

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | Refresh token in Authorization header |
| Authorization | ✅ RBAC | All authenticated users |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

**Request:**
```json
{
  "refresh_token": "JWT token (optional, can use header)"
}
```

**Response:**
```json
{
  "access_token": "New JWT token (15 min expiration)",
  "expires_at": "RFC3339 timestamp"
}
```

---

## Collector Endpoints

### POST /api/v1/collectors/register

**Description:** Register new collector and receive credentials

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ⚠️ Secret | X-Registration-Secret header required |
| Authorization | ✅ RBAC | Admin/User can register |
| Rate Limit | ✅ 10/min | Per IP address |
| HTTPS | ✅ Required | TLS 1.2+ |

**Headers Required:**
```
X-Registration-Secret: <shared-secret-from-config>
```

**Request:**
```json
{
  "name": "string (required, 1-255 chars)",
  "hostname": "string (required, valid hostname)"
}
```

**Response:**
```json
{
  "collector_id": "UUID",
  "token": "JWT token (365 day expiration)",
  "certificate": "X.509 PEM format",
  "private_key": "RSA 2048-bit PEM format",
  "expires_at": "RFC3339 timestamp (1 year)"
}
```

**Security Checks:**
1. Registration secret matches configuration
2. Collector name is unique
3. Hostname is valid (DNS check optional)
4. Certificate generated with RSA 2048-bit
5. JWT token includes collector_id and hostname
6. Private key returned ONLY on registration (not stored)

**⚠️ Important:** Save certificate and private key immediately. They are not retrievable later.

---

### GET /api/v1/collectors

**Description:** List all collectors (admin) or own collectors (user)

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user (own only), admin (all) |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

**Header Required:**
```
Authorization: Bearer <user-jwt-token>
```

**Response:**
```json
{
  "collectors": [
    {
      "id": "UUID",
      "name": "production-db-01",
      "hostname": "db.example.com",
      "status": "active|inactive",
      "last_seen": "RFC3339 timestamp",
      "created_at": "RFC3339 timestamp"
    }
  ]
}
```

**RBAC Rules:**
- **admin:** Can see all collectors
- **user:** Can see only collectors they own
- **viewer:** Cannot access

---

### GET /api/v1/collectors/{id}

**Description:** Get specific collector details

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user (owner only), admin (any) |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

---

### DELETE /api/v1/collectors/{id}

**Description:** Delete collector and revoke access

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | admin only |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

**Security Checks:**
1. Requesting user is admin
2. Collector exists
3. Soft delete (mark inactive, don't remove)
4. Revoke associated JWT tokens
5. Audit log the deletion

---

## Metrics Endpoints

### POST /api/v1/metrics/push

**Description:** High-volume metrics ingestion endpoint for collectors

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT collector token |
| Authorization | ✅ RBAC | Collector role |
| Rate Limit | ✅ 1000/min | Per collector (10x user limit) |
| HTTPS | ✅ Required | TLS 1.2+ |

**Header Required:**
```
Authorization: Bearer <collector-jwt-token>
```

**Request:**
```json
{
  "collector_id": "UUID (must match token)",
  "timestamp": "RFC3339 timestamp",
  "metrics": [
    {
      "type": "pg_stats|system|replication",
      "data": "JSON object"
    }
  ]
}
```

**Response:**
```json
{
  "processed": 42,
  "errors": 0,
  "timestamp": "RFC3339 timestamp"
}
```

**Security Checks:**
1. Collector JWT token valid and not expired
2. collector_id claim matches request body
3. Collector status is 'active'
4. Metrics array not empty
5. Timestamp is recent (within 24 hours)
6. Rate limited to 1000 requests/minute per collector
7. Request size < 10MB

**⚠️ High-Volume:** This endpoint receives most traffic. Rate limiting is critical.

---

### GET /api/v1/metrics

**Description:** Query metrics by collector, type, and time range

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user (own collectors), admin (all) |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

**Query Parameters:**
```
?collector_id=UUID (required)
&type=pg_stats|system|replication (optional)
&start_time=RFC3339 (optional)
&end_time=RFC3339 (optional)
&limit=100 (max 10000)
&offset=0
```

**Response:**
```json
{
  "metrics": [
    {
      "id": "UUID",
      "collector_id": "UUID",
      "type": "string",
      "timestamp": "RFC3339",
      "data": "JSON object"
    }
  ],
  "total": 1234,
  "limit": 100,
  "offset": 0
}
```

**Security Checks:**
1. User JWT token valid
2. User authorized to view collector metrics
3. Time range is valid (start < end)
4. Limit ≤ 10000 (pagination protection)
5. Results filtered by user's access level

---

### GET /api/v1/metrics/cache

**Description:** Get cached metrics for real-time dashboard display

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user, admin |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |
| Cache | ✅ 30s TTL | Redis cached |

---

## Configuration Endpoints

### GET /api/v1/config/{collector_id}

**Description:** Get collector configuration

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token or mTLS cert |
| Authorization | ✅ RBAC | Collector (own), admin (any) |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

---

### PUT /api/v1/config/{collector_id}

**Description:** Update collector configuration

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | admin only |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

**Security Checks:**
1. User is admin
2. Collector exists
3. Configuration is valid
4. No security-related configs exposed in response
5. Change logged in audit trail

---

## Server Endpoints

### GET /api/v1/servers

**Description:** List servers and their metrics

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user, admin |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

---

### GET /api/v1/servers/{id}/metrics

**Description:** Get metrics for specific server

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user, admin |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

---

## Alerts Endpoints

### GET /api/v1/alerts

**Description:** List alerts for user's collectors

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user, admin |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

---

### POST /api/v1/alerts

**Description:** Create new alert rule

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ✅ Required | JWT user token |
| Authorization | ✅ RBAC | user, admin |
| Rate Limit | ✅ 100/min | Per user |
| HTTPS | ✅ Required | TLS 1.2+ |

**Request:**
```json
{
  "name": "High CPU Alert",
  "collector_id": "UUID",
  "condition": "cpu_percent > 80",
  "threshold": 80,
  "duration": 300,
  "actions": [
    {"type": "email", "recipient": "admin@example.com"},
    {"type": "webhook", "url": "https://example.com/alert"}
  ]
}
```

**Security Checks:**
1. User authenticated
2. User owns collector
3. Condition syntax is valid
4. Actions are safe (no RCE vectors)
5. Webhook URL is whitelisted (optional)

---

## Health & Version Endpoints

### GET /api/v1/health

**Description:** API health check

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ❌ Public | No auth required |
| Authorization | ✅ RBAC | All users |
| Rate Limit | ✅ 100/min | Per IP address |
| HTTPS | ❌ Optional | HTTP OK for health |

**Response:**
```json
{
  "status": "healthy|degraded|unhealthy",
  "uptime": 3600,
  "version": "3.2.0",
  "checks": {
    "database": "healthy",
    "cache": "healthy",
    "ml_service": "healthy"
  }
}
```

---

### GET /version

**Description:** API version information

**Security:**
| Aspect | Status | Details |
|--------|--------|---------|
| Authentication | ❌ Public | No auth required |
| Authorization | ✅ RBAC | All users |
| Rate Limit | ✅ 100/min | Per IP address |
| HTTPS | ❌ Optional | HTTP OK for version |

**Response:**
```json
{
  "version": "3.2.0",
  "build": "abc123def456",
  "commit": "1a2b3c4d",
  "timestamp": "2026-02-25T10:30:00Z"
}
```

---

## Request/Response Security Patterns

### JWT Token Format

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

**Header:** Algorithm and token type
**Payload:** Claims (user_id, role, exp, etc.)
**Signature:** HMAC-SHA256(secret)

### Error Responses

All errors use consistent format (no sensitive information):

```json
{
  "error": "Human-readable error message",
  "code": 400,
  "timestamp": "RFC3339 timestamp"
}
```

**Status Codes:**
- **400 Bad Request:** Invalid input
- **401 Unauthorized:** Missing/invalid authentication
- **403 Forbidden:** Insufficient permissions
- **404 Not Found:** Resource not found
- **429 Too Many Requests:** Rate limit exceeded
- **500 Internal Server Error:** Server-side issue

### Response Headers

All responses include security headers:

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; ...
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 85
X-RateLimit-Reset: 1645793400
```

---

## Rate Limiting Details

### Token Bucket Algorithm

**Parameters:**
- Capacity: Configurable per client type
- Refill rate: Tokens distributed evenly over 60 seconds
- Client ID: Derived from user_id, collector_id, or IP address

**Limits:**
- **Users:** 100 requests/minute
- **Collectors:** 1000 requests/minute
- **Public endpoints:** 10-100 requests/minute

**Tracking:**
- Per-user ID (from JWT claims)
- Per-collector ID (from JWT claims)
- Per-IP address (fallback for public endpoints)

**Response when exceeded:**
```
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1645793460
Retry-After: 60

{"error": "Rate limit exceeded", "code": 429}
```

---

## RBAC Access Matrix

| Endpoint | Admin | User | Viewer | Collector |
|----------|-------|------|--------|-----------|
| POST /auth/login | ✅ | ✅ | ✅ | - |
| POST /collectors/register | ✅ | ✅ | ❌ | ✅ |
| GET /collectors | ✅ | ✅ Own | ❌ | - |
| DELETE /collectors/{id} | ✅ | ❌ | ❌ | - |
| POST /metrics/push | - | - | - | ✅ |
| GET /metrics | ✅ All | ✅ Own | ✅ | - |
| GET /config | ✅ | ❌ | ❌ | ✅ Own |
| PUT /config | ✅ | ❌ | ❌ | ❌ |
| GET /servers | ✅ | ✅ | ✅ | - |
| GET /alerts | ✅ | ✅ Own | ❌ | - |
| POST /alerts | ✅ | ✅ | ❌ | - |

---

## Testing Checklist

Before deploying, verify all endpoints:

```bash
# 1. Test authentication
curl -X POST https://api.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 2. Test authorization (admin-only endpoint)
TOKEN="eyJ..."
curl -X PUT https://api.example.com/api/v1/config/123 \
  -H "Authorization: Bearer $TOKEN" # Should 403 if not admin

# 3. Test rate limiting
for i in {1..150}; do
  curl -s https://api.example.com/api/v1/health
done | grep -c "429"  # Should see ~50 429 responses

# 4. Test metrics push (requires collector token)
COLLECTOR_TOKEN="collector-jwt-token"
curl -X POST https://api.example.com/api/v1/metrics/push \
  -H "Authorization: Bearer $COLLECTOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"collector_id":"uuid","metrics":[]}'

# 5. Test security headers
curl -I https://api.example.com/api/v1/health
# Verify: Strict-Transport-Security, X-Frame-Options, etc.
```

---

**Version:** 1.0
**Last Updated:** February 25, 2026
**Status:** Production-Ready

