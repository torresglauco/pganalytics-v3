# Security Policy

**pgAnalytics-v3**

---

## 1. Security Architecture Overview

pgAnalytics-v3 implements a multi-layered security model to protect sensitive database metrics and maintain system integrity:

### Trust Boundaries

```
┌─────────────────────────────────────────────────────────────┐
│                     Public Internet                         │
└──────────────────┬──────────────────────────────────────────┘
                   │
          ┌────────▼────────┐
          │   TLS/mTLS      │  (Encryption in Transit)
          │   (Phase 2)      │
          └────────┬────────┘
                   │
    ┌──────────────┴──────────────┐
    │    API Gateway Layer         │
    │  - Rate Limiting             │
    │  - Authentication (JWT)      │
    │  - CORS Policy               │
    └──────────────┬──────────────┘
                   │
    ┌──────────────┴──────────────┐
    │  Authorization Layer (RBAC)  │
    │  - Role Validation           │
    │  - Endpoint ACLs             │
    └──────────────┬──────────────┘
                   │
    ┌──────────────┴──────────────┐
    │   API Handler Layer          │
    │  - Input Validation          │
    │  - SQL Injection Prevention   │
    │  - Sensitive Data Masking     │
    └──────────────┬──────────────┘
                   │
    ┌──────────────┴──────────────┐
    │   Database Layer             │
    │  - Parameterized Queries     │
    │  - Row-Level Security (RLS)  │
    │  - Encryption at Rest        │
    └──────────────────────────────┘
```

### Security Principles

1. **Defense in Depth:** Multiple independent security layers
2. **Fail Secure:** Default to deny, whitelist allowed operations
3. **Least Privilege:** Users and services get minimum required permissions
4. **Zero Trust:** Every request is authenticated and authorized
5. **Audit Trail:** All sensitive operations are logged

---

## 2. Authentication Mechanisms

### 2.1 User Authentication (JWT)

Users authenticate via username/password and receive JWT tokens for subsequent requests.

**Flow:**

```
1. POST /api/v1/auth/login
   {"username": "user@example.com", "password": "secret"}

2. Response:
   {
     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "refresh_token": "...",
     "expires_at": "2026-02-24T22:15:00Z",
     "user": {"id": 1, "username": "user@example.com", "role": "user"}
   }

3. Subsequent requests:
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Implementation:**

- **Algorithm:** HS256 (HMAC-SHA256)
- **Secret:** `JWT_SECRET` environment variable (required in production)
- **Expiration:** Configurable (default: 15 minutes)
- **Refresh Token:** Long-lived token for obtaining new access tokens (default: 24 hours)

**Security Requirements:**

```yaml
Requirements:
  - JWT_SECRET:
      description: "Minimum 64 characters, cryptographically random"
      production: required
      environment: required

  - Password Hash:
      algorithm: "bcrypt"
      cost: 12

  - Token Storage:
      location: "HTTP-Only Secure Cookie or localStorage"
      never_expose: "In URLs, HTTP headers beyond Bearer"
```

**Validation Process:**

```go
// All protected endpoints validate:
1. Token signature (HS256)
2. Token expiration (iat + exp claims)
3. User exists and is active
4. User role is valid (matches endpoint requirements)
```

### 2.2 Collector Authentication (mTLS + JWT)

Collectors authenticate using mutual TLS certificates and receive JWT tokens.

**Registration Flow:**

```
1. POST /api/v1/collectors/register
   Header: X-Registration-Secret: ${REGISTRATION_SECRET}
   {
     "name": "prod-db-001",
     "hostname": "db.example.com",
     "address": "192.168.1.10"
   }

2. Response:
   {
     "collector_id": "550e8400-e29b-41d4-a716-446655440000",
     "token": "eyJhbGc...",
     "certificate": "-----BEGIN CERTIFICATE-----...",
     "private_key": "-----BEGIN PRIVATE KEY-----...",
     "expires_at": "2027-02-24T00:00:00Z"
   }
```

**Metrics Push Authentication:**

```
POST /api/v1/metrics/push
Authorization: Bearer ${COLLECTOR_TOKEN}
Content-Type: application/json

{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "metrics": [...]
}

Validation:
1. Bearer token present
2. Token signature valid (HS256)
3. Token not expired
4. collector_id claim matches request.collector_id
5. Collector status is 'active'
```

**Security Requirements:**

```yaml
Requirements:
  - REGISTRATION_SECRET:
      description: "Unique pre-shared secret for collector registration"
      production: required
      minimum_length: 32

  - Collector Certificate:
      type: "X.509"
      validity: "1 year (configurable)"
      renewal: "Before expiration (manual/automated)"

  - Token Rotation:
      frequency: "Every 90 days (recommended)"
      process: "Re-register collector with new credentials"
```

### 2.3 API Key Authentication (Future)

Reserved for service-to-service communication (not yet implemented).

---

## 3. Authorization Model (RBAC)

Role-Based Access Control enforces least-privilege access to API endpoints.

### 3.1 Roles

```
Role Hierarchy:
┌─────────────┐
│    admin    │  (Level 3) - Full system access
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    user     │  (Level 2) - Normal API access
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   viewer    │  (Level 1) - Read-only access
└─────────────┘
```

### 3.2 Role Capabilities Matrix

| Feature | Admin | User | Viewer | Collector |
|---------|-------|------|--------|-----------|
| View metrics | ✅ | ✅ | ✅ | - |
| View collectors | ✅ | ✅ | ✅ | - |
| Register collector | ✅ | ✅ | ❌ | ✅ |
| Update config | ✅ | ❌ | ❌ | ❌ |
| Delete collector | ✅ | ❌ | ❌ | ❌ |
| Manage users | ✅ | ❌ | ❌ | ❌ |
| Push metrics | - | - | - | ✅ |
| View audit logs | ✅ | ❌ | ❌ | ❌ |

### 3.3 Endpoint Authorization

Protected endpoints enforce role requirements via middleware:

```go
// Example: Admin-only endpoint
router.PUT("/api/v1/config/:collector_id",
  s.AuthMiddleware(),
  s.RoleMiddleware("admin"),  // Requires admin role
  s.handleUpdateConfig,
)
```

---

## 4. Known Vulnerabilities & Mitigations

### 4.1 Critical Issues (Must Fix Before Production)

#### Issue: Metrics Push Authentication Disabled
- **Status:** FIXED (v3.1.0)
- **Details:** Previously, `/api/v1/metrics/push` accepted unauthenticated requests
- **Fix:** Require valid collector JWT token and validate collector_id matches
- **Test:** `curl -X POST ... /api/v1/metrics/push` without token must return 401

#### Issue: Collector Registration Unauthenticated
- **Status:** FIXED (v3.1.0)
- **Details:** Any entity could register as collector and obtain JWT
- **Fix:** Require `X-Registration-Secret` header matching server config
- **Test:** Omitting header must return 401

#### Issue: Password Verification Broken
- **Status:** FIXED (v3.1.0)
- **Details:** Login accepted any non-empty password string
- **Fix:** Use `bcrypt.CompareHashAndPassword()` for actual verification
- **Test:** Wrong password must return 401

#### Issue: RBAC Not Implemented
- **Status:** FIXED (v3.1.0)
- **Details:** RoleMiddleware was empty stub
- **Fix:** Check user role against role hierarchy
- **Test:** Admin-only endpoints must reject viewer/user roles

### 4.2 High Priority Issues

#### Issue: Rate Limiting Missing
- **Status:** FIXED (v3.1.0)
- **Details:** No protection against brute force or DDoS
- **Fix:** Implement token bucket rate limiter (100 req/min per user, 1000 per collector)
- **Test:** 101st request in 60s must return 429

#### Issue: mTLS Not Implemented
- **Status:** PLANNED (Phase 2)
- **Details:** Collector certificates not validated
- **Fix:** Implement full mTLS handshake with certificate pinning
- **Target:** v3.2.0

#### Issue: Security Headers Missing
- **Status:** FIXED (v3.1.0)
- **Details:** Missing XSS, clickjacking, and content-type protections
- **Fix:** Add middleware for security headers (X-Frame-Options, CSP, etc.)
- **Test:** All responses must include headers

### 4.3 Medium Priority Issues

#### Issue: SQL Injection
- **Status:** PROTECTED
- **Details:** Using parameterized queries throughout (✅ No vulnerability)
- **Review:** `backend/internal/storage/postgres.go` - All queries use `$1, $2` placeholders
- **Test:** SQL injection in query parameters must be properly escaped

#### Issue: Cross-Site Scripting (XSS)
- **Status:** PROTECTED
- **Details:** API returns JSON only (no HTML templates)
- **Review:** All error messages are JSON-encoded
- **Test:** Special characters in error responses must be escaped

#### Issue: Cross-Site Request Forgery (CSRF)
- **Status:** PROTECTED
- **Details:** Uses JWT tokens (not cookies), no state-changing GET requests
- **Review:** All mutations use POST/PUT/DELETE
- **Test:** CSRF tokens not needed for API

---

## 5. Security Deployment Checklist

### Pre-Production Security Checklist

Use this checklist before deploying to production:

```yaml
Environment Variables:
  ✅ JWT_SECRET: Set to 64+ character random string
  ✅ REGISTRATION_SECRET: Set to unique pre-shared secret
  ✅ DATABASE_URL: Using TLS connection (sslmode=require)
  ✅ ENVIRONMENT: Set to "production"
  ✅ TLS_CERT: Valid certificate path
  ✅ TLS_KEY: Valid private key path
  ✅ TLS_ENABLED: true

Infrastructure:
  ✅ API behind HTTPS reverse proxy (nginx/haproxy)
  ✅ Database behind private network (no public access)
  ✅ Collectors behind VPN or mTLS
  ✅ Secrets stored in vault (not environment files)
  ✅ Audit logging enabled (PostgreSQL logging)
  ✅ Monitoring configured for failed auth attempts
  ✅ Backup encryption enabled
  ✅ Disaster recovery tested

Code:
  ✅ All passwords hashed with bcrypt (cost 12)
  ✅ All SQL queries parameterized
  ✅ No sensitive data in logs
  ✅ No credentials in code/config files
  ✅ Error responses don't expose stack traces
  ✅ Rate limiting configured
  ✅ RBAC enforced on protected endpoints

Operations:
  ✅ Incident response plan documented
  ✅ Security team trained on procedures
  ✅ Automated security scanning configured
  ✅ Key rotation schedule established
  ✅ Certificate renewal automated
  ✅ Penetration testing completed
  ✅ Security review completed
```

### Post-Deployment Monitoring

```yaml
Alerts to Configure:
  - Failed login attempts > 5 in 5 minutes
  - Rate limit 429 responses > 100/min per user
  - Authentication token validation failures
  - SQL error rate spike
  - Unusual query patterns
  - Unauthorized endpoint access attempts
  - Certificate expiration < 30 days
  - Backup integrity checks

Regular Reviews:
  - Weekly: Review authentication failures
  - Monthly: Audit access control policies
  - Quarterly: Review and rotate secrets
  - Annually: Security assessment
```

---

## 6. Incident Response Procedures

### 6.1 Authentication Incident

**Scenario:** Unauthorized access or compromised credentials

**Response:**

1. **Immediate (0-5 min):**
   - Isolate affected user/collector account
   - Disable JWT tokens (add to blacklist)
   - Enable enhanced logging

2. **Short-term (5-30 min):**
   - Rotate JWT_SECRET
   - Reset user passwords
   - Re-generate collector certificates
   - Review access logs for data exfiltration

3. **Follow-up (1-7 days):**
   - Audit all data accessed by compromised account
   - Force password reset for all users
   - Update security documentation
   - Implement additional monitoring

### 6.2 Data Breach

**Scenario:** Unauthorized query metrics or database access

**Response:**

1. **Immediate (0-1 hour):**
   - Enable all audit logging
   - Snapshot database for forensics
   - Disconnect all non-essential services
   - Notify security team

2. **Investigation (1-24 hours):**
   - Analyze access logs
   - Identify what data was accessed
   - Determine breach scope
   - Preserve evidence

3. **Remediation (1-7 days):**
   - Patch vulnerability
   - Rotate all credentials
   - Audit role assignments
   - Implement additional controls

### 6.3 Denial of Service

**Scenario:** Rate limiting bypassed or connection pool exhausted

**Response:**

1. **Immediate:**
   - Enable stricter rate limits
   - Scale up database connections
   - Implement IP-based blocking if needed
   - Notify infrastructure team

2. **Investigation:**
   - Identify attack source
   - Analyze traffic patterns
   - Review rate limiter effectiveness
   - Check for legitimate traffic issues

3. **Prevention:**
   - Increase rate limits based on workload
   - Add DDoS mitigation (WAF/CDN)
   - Implement request queuing
   - Add circuit breaker for database

---

## 7. Security Testing

### 7.1 Automated Testing

Run security tests in CI/CD pipeline:

```bash
# SQL Injection testing
./tests/security/sql_injection_test.go

# Authentication testing
./tests/security/auth_test.go

# Authorization testing
./tests/security/rbac_test.go

# Rate limiting testing
./tests/security/ratelimit_test.go
```

### 7.2 Manual Testing

Perform these manual tests before each production deployment:

```bash
# Test metrics push without auth (should fail)
curl -X POST http://localhost:8080/api/v1/metrics/push \
  -H "Content-Type: application/json" \
  -d '{}' # Should return 401

# Test registration without secret (should fail)
curl -X POST http://localhost:8080/api/v1/collectors/register \
  -H "Content-Type: application/json" \
  -d '{"name":"test"}' # Should return 401

# Test wrong password (should fail)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"wrong"}' # Should return 401

# Test rate limiting (should get 429)
for i in {1..150}; do
  curl -s http://localhost:8080/api/v1/health
done | grep -c 429 # Should see ~50 429 responses
```

### 7.3 Penetration Testing

Before production, engage third-party security firm for:

- API endpoint testing
- SQL injection assessment
- Authentication bypass attempts
- Authorization boundary testing
- Cryptographic implementation review

---

## 8. Security References

### External Standards

- **OWASP Top 10:** https://owasp.org/www-project-top-ten/
- **CWE Top 25:** https://cwe.mitre.org/top25/
- **NIST Cybersecurity Framework:** https://www.nist.gov/cyberframework

### Implementation References

- **JWT Best Practices:** https://tools.ietf.org/html/rfc7519
- **OWASP Authentication:** https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- **OWASP Authorization:** https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
- **PostgreSQL Security:** https://www.postgresql.org/docs/current/sql-syntax.html

---

## 9. Responsible Disclosure

If you discover a security vulnerability, please follow responsible disclosure:

1. **Do NOT** create a public GitHub issue
2. **Email:** security@pganalytics.dev with:
   - Vulnerability description
   - Steps to reproduce
   - Impact assessment
   - Suggested remediation

3. **Timeline:**
   - We will acknowledge receipt within 24 hours
   - We will provide status updates every 48 hours
   - We will work on a fix and target 30-day patch release
   - We will credit you in the release notes (if desired)

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-24 | Initial security documentation |
| 1.1 | 2026-02-24 | Added security fixes for v3.1.0 |

---

**Last Updated:** February 24, 2026
**Status:** ACTIVE
**Classification:** INTERNAL - Security Guidelines
