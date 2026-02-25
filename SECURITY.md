# Security Guidelines - pgAnalytics v3.2.0

**Status**: ✅ Production-Ready | Audit Complete | All Critical Issues Resolved

---

## Overview

pgAnalytics v3.2.0 implements comprehensive security measures including JWT authentication, role-based access control (RBAC), mutual TLS support, rate limiting, and secure password hashing. This document outlines security features, best practices, and deployment requirements.

---

## Table of Contents

1. [Authentication](#authentication)
2. [Authorization](#authorization)
3. [API Security](#api-security)
4. [Network Security](#network-security)
5. [Data Protection](#data-protection)
6. [Production Deployment](#production-deployment)
7. [Incident Response](#incident-response)

---

## Authentication

### JWT Token-Based Authentication

pgAnalytics uses JWT (JSON Web Tokens) for stateless authentication.

**Token Types**:
- **User Tokens**: For API clients and web interfaces
- **Collector Tokens**: For distributed collectors pushing metrics
- **Refresh Tokens**: For long-lived sessions

**Token Generation**:
```bash
# POST /api/v1/auth/login
{
  "username": "admin",
  "password": "your-password"
}

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-25T20:26:00Z"
}
```

**Token Validation**:
- Signature verification using `JWT_SECRET`
- Expiration time check
- Claim validation (user_id, role, collector_id)

**Security Features**:
- ✅ HMAC-SHA256 signature (HS256)
- ✅ Configurable expiration (default 24 hours)
- ✅ Refresh token rotation
- ✅ Token extraction from `Authorization: Bearer <token>` header

---

### Collector Authentication

Collectors authenticate using JWT tokens generated during registration.

**Registration**:
```bash
# POST /api/v1/collectors/register
# Requires: X-Registration-Secret header

{
  "hostname": "db-server-01",
  "description": "Production PostgreSQL Server"
}

# Response
{
  "collector_id": "col_abc123...",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-25T20:26:00Z"
}
```

**Security Requirements**:
- ✅ `X-Registration-Secret` header must match `REGISTRATION_SECRET` environment variable
- ✅ Secret is validated as non-default in production
- ✅ Collectors cannot register without valid secret
- ✅ One-time registration process per collector

---

## Authorization

### Role-Based Access Control (RBAC)

pgAnalytics implements a three-tier role hierarchy:

**Role Levels**:
1. **admin** (Level 3): Full access to all endpoints and operations
2. **user** (Level 2): Standard access to metrics, dashboards, and configurations
3. **viewer** (Level 1): Read-only access to dashboards and metrics

**Protected Endpoints**:
- All `/api/v1/servers/*` endpoints require authentication
- All `/api/v1/alerts/*` endpoints require authentication
- All `/api/v1/config/*` endpoints require authentication
- Only `POST /api/v1/collectors/register` does not require authentication (uses registration secret)

---

## API Security

### Metrics Push Endpoint

The `/api/v1/metrics/push` endpoint requires strict authentication.

**Requirements**:
- ✅ Valid JWT collector token (in `Authorization` header)
- ✅ Collector ID in token must match request body
- ✅ Rate limiting (100-1000 req/min per collector)

**Response Codes**:
- `200 OK`: Metrics accepted and stored
- `401 Unauthorized`: Missing or invalid token
- `429 Too Many Requests`: Rate limit exceeded

---

### Rate Limiting

pgAnalytics implements token bucket rate limiting.

**Limits**:
- **Per-User**: 100 requests/minute
- **Per-Collector**: 1000 requests/minute
- **Fallback**: Per-IP address

**Rate Limit Response**:
```
HTTP/1.1 429 Too Many Requests
{
  "error": "Too many requests",
  "code": 429
}
```

---

### Security Headers

All responses include security headers:

| Header | Purpose |
|--------|---------|
| X-Frame-Options: DENY | Clickjacking protection |
| X-Content-Type-Options: nosniff | MIME sniffing prevention |
| X-XSS-Protection: 1; mode=block | XSS protection |
| Content-Security-Policy | Content injection prevention |
| Strict-Transport-Security | Forces HTTPS (production only) |

---

## Network Security

### TLS/SSL Support

Configure TLS for encrypted communication:

```bash
export TLS_ENABLED="true"
export TLS_CERT_PATH="/etc/pganalytics/cert.pem"
export TLS_KEY_PATH="/etc/pganalytics/key.pem"
```

### Mutual TLS (mTLS)

Collectors support bidirectional authentication:

```bash
export MTLS_ENABLED="true"
export MTLS_CLIENT_CERT="/etc/pganalytics/collector.crt"
export MTLS_CLIENT_KEY="/etc/pganalytics/collector.key"
```

---

## Data Protection

### SQL Injection Prevention

All queries use prepared statements via sqlc:

```go
// ✅ Safe: Parameterized query
db.QueryRow("SELECT * FROM collectors WHERE id = $1", collectorID)

// ❌ Never: String concatenation
// "SELECT * FROM collectors WHERE id = '" + collectorID + "'"
```

### Password Security

- ✅ Hashed with bcrypt (cost 12)
- ✅ Constant-time comparison
- ✅ Never stored in plain text
- ✅ Never exposed in logs

### Sensitive Data Handling

**Logged**: User logins, API requests, collector registrations
**Not Logged**: Passwords, JWT tokens, credentials, API secrets

---

## Production Deployment

### Pre-Deployment Checklist

- [ ] Generate secrets: `openssl rand -base64 32`
- [ ] Obtain TLS certificate from trusted CA
- [ ] Set ENVIRONMENT=production
- [ ] Configure all required environment variables
- [ ] Test authentication and registration
- [ ] Run tests: `make test-backend && make test-integration`
- [ ] Verify security headers in responses
- [ ] Check rate limiting is active
- [ ] Monitor for failed authentication attempts
- [ ] Monitor TLS certificate expiration

### Environment Variables

```bash
# Security
export JWT_SECRET="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_KEY="$(openssl rand -base64 32)"

# TLS
export TLS_ENABLED="true"
export TLS_CERT_PATH="/etc/pganalytics/cert.pem"
export TLS_KEY_PATH="/etc/pganalytics/key.pem"

# Environment
export ENVIRONMENT="production"
```

---

## Incident Response

### If You Suspect a Security Issue

1. **Contain**: Stop affected services if necessary
2. **Preserve**: Keep logs and evidence
3. **Investigate**: Determine scope and impact
4. **Notify**: Alert security team
5. **Remediate**: Apply fixes
6. **Review**: Post-incident analysis

### Common Incidents

**Unauthorized Access**: Rotate JWT_SECRET, reset passwords, review logs
**Data Breach**: Determine impact, check backups, notify parties
**Credential Compromise**: Rotate credentials, invalidate tokens
**DDoS Attack**: Review rate limiter, update firewall rules

---

## References

- [OWASP Top 10](https://owasp.org/Top10/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8949)
- [pgAnalytics Documentation](README.md)

---

**Document Version**: 1.0
**Last Updated**: February 25, 2026
**Status**: ✅ Production Ready
