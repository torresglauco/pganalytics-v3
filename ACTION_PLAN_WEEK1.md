# 🚀 WEEK 1 ACTION PLAN - Critical Fixes

**Objective:** Fix 1 critical + 8 high severity security/testing issues
**Timeline:** 40-50 hours (4-5 working days)
**Team:** 1-2 senior engineers

---

## 📋 TASK BREAKDOWN

### TASK 1: Fix MD5 UUID Generation (30 min)

**File:** `backend/internal/auth/service.go`

**Current Code (Line 151-152):**
```go
hostHash := md5.Sum([]byte(req.Hostname))
collectorID := uuid.NewSHA1(uuid.Nil, hostHash[:])
```

**Fixed Code:**
```go
// Use random UUID v4 instead of deterministic hash
collectorID := uuid.New()  // Returns uuid.UUID

// Store mapping separately if needed for tracking
log.Printf("Registered collector %s (hostname: %s)", collectorID, req.Hostname)
```

**Verification:**
```bash
# Run tests to ensure no regression
cd backend
go test ./internal/auth -v

# Check that UUID is different each time
go run main.go register-collector --test
```

**Risk:** Low (UUID v4 is standard, no compatibility issues)

---

### TASK 2: Fix CORS Configuration (20 min)

**File:** `backend/internal/api/middleware.go`

**Current Code (Line 260-261):**
```go
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
```

**Fix Step 1 - Update middleware:**
```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        // Get allowed origins from config
        allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
        if allowedOrigins == "" {
            allowedOrigins = "http://localhost:3000,http://localhost:5173"
        }

        if isOriginAllowed(origin, allowedOrigins) {
            c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
            c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
            c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        }

        c.Next()
    }
}

func isOriginAllowed(origin, allowedOrigins string) bool {
    for _, allowed := range strings.Split(allowedOrigins, ",") {
        if strings.TrimSpace(allowed) == origin {
            return true
        }
    }
    return false
}
```

**Fix Step 2 - Update docker-compose.yml:**
```yaml
backend:
  environment:
    # For development - add your frontend URLs
    ALLOWED_ORIGINS: "http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000"
```

**Fix Step 3 - Production environment:**
```bash
# In .env.production
ALLOWED_ORIGINS=https://app.example.com,https://monitoring.example.com
```

**Verification:**
```bash
# Test CORS headers
curl -H "Origin: http://localhost:3000" http://localhost:8080/api/v1/health -v

# Should see:
# Access-Control-Allow-Origin: http://localhost:3000
```

---

### TASK 3: Enable Database SSL (20 min)

**Files:**
- `docker-compose.yml`
- `docker-compose.production.yml`
- `.env.example`

**Step 1 - Update docker-compose.yml for development:**
```yaml
backend:
  environment:
    # Add sslmode to connections
    DATABASE_URL: "postgres://postgres:pganalytics@postgres:5432/pganalytics?sslmode=disable"
    TIMESCALE_URL: "postgres://postgres:pganalytics@timescale:5433/metrics?sslmode=disable"
    # Note: disable is OK in docker-compose for local dev

postgres:
  environment:
    POSTGRES_DB: pganalytics

timescale:
  environment:
    POSTGRES_DB: metrics
```

**Step 2 - Create .env.production:**
```env
# Production database connections with SSL
DATABASE_URL=postgres://pganalytics:CHANGE_ME@db.production.com:5432/pganalytics?sslmode=verify-full
TIMESCALE_URL=postgres://timescale:CHANGE_ME@timescale.production.com:5433/metrics?sslmode=verify-full

# Or if using internal network without proper certs
DATABASE_URL=postgres://pganalytics:CHANGE_ME@db.production.internal:5432/pganalytics?sslmode=require
```

**Step 3 - Add validation to backend config:**
```go
// backend/internal/config/config.go
func LoadConfig() *Config {
    config := &Config{}

    // Validate SSL mode in production
    if os.Getenv("ENVIRONMENT") == "production" {
        dbURL := os.Getenv("DATABASE_URL")
        if !strings.Contains(dbURL, "sslmode=require") &&
           !strings.Contains(dbURL, "sslmode=verify-full") {
            log.Fatal("CRITICAL: DATABASE_URL must use sslmode=require or verify-full in production")
        }
    }

    return config
}
```

**Verification:**
```bash
# Test connection with SSL
export DATABASE_URL="postgres://user:pass@localhost:5432/db?sslmode=require"
./pganalytics-api
# Should connect successfully (or fail with SSL error, which is expected if certs not configured)
```

---

### TASK 4: Remove Hardcoded Credentials (30 min)

**File:** `docker-compose.yml`

**Step 1 - Create .env.example:**
```env
# PostgreSQL Credentials (CHANGE IN PRODUCTION)
POSTGRES_USER=postgres
POSTGRES_PASSWORD=CHANGE_ME_IN_PRODUCTION
POSTGRES_DB=pganalytics

# Backend
JWT_SECRET=CHANGE_ME_IN_PRODUCTION_USE_openssl_rand_-base64_32
REGISTRATION_SECRET=CHANGE_ME_IN_PRODUCTION_USE_openssl_rand_-base64_32

# Grafana (CHANGE IN PRODUCTION)
GF_SECURITY_ADMIN_USER=admin
GF_SECURITY_ADMIN_PASSWORD=CHANGE_ME_IN_PRODUCTION

# TLS (only in production)
TLS_ENABLED=false
TLS_CERT_PATH=/etc/pganalytics/cert.pem
TLS_KEY_PATH=/etc/pganalytics/key.pem
```

**Step 2 - Update docker-compose.yml:**
```yaml
services:
  postgres:
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-postgres}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-pganalytics}
      POSTGRES_DB: ${POSTGRES_DB:-pganalytics}

  backend:
    environment:
      JWT_SECRET: ${JWT_SECRET:-demo-secret-change-in-production}
      REGISTRATION_SECRET: ${REGISTRATION_SECRET:-demo-registration-secret}

  grafana:
    environment:
      GF_SECURITY_ADMIN_USER: ${GF_SECURITY_ADMIN_USER:-admin}
      GF_SECURITY_ADMIN_PASSWORD: ${GF_SECURITY_ADMIN_PASSWORD:-admin}
```

**Step 3 - Create setup script:**
```bash
#!/bin/bash
# scripts/generate-secrets.sh

echo "🔐 Generating production secrets..."

JWT_SECRET=$(openssl rand -base64 32)
REGISTRATION_SECRET=$(openssl rand -base64 32)
POSTGRES_PASSWORD=$(openssl rand -base64 24)
GF_PASSWORD=$(openssl rand -base64 24)

cat > .env.production << EOF
# Auto-generated secrets - $(date)
POSTGRES_PASSWORD=$POSTGRES_PASSWORD
JWT_SECRET=$JWT_SECRET
REGISTRATION_SECRET=$REGISTRATION_SECRET
GF_SECURITY_ADMIN_PASSWORD=$GF_PASSWORD
ENVIRONMENT=production
TLS_ENABLED=true
EOF

echo "✅ Secrets generated in .env.production"
echo "⚠️  IMPORTANT: Keep this file secure!"
```

**Verification:**
```bash
# Verify no secrets are hardcoded
grep -r "pganalytics\|demo-secret\|admin123" docker-compose.yml
# Should return nothing (no hardcoded values)

# Test with .env file
cp .env.example .env.test
./scripts/generate-secrets.sh
docker-compose --env-file .env.test up
```

---

### TASK 5: Fix Setup Endpoint (10 min)

**File:** `docker-compose.yml`

**Current Code:**
```yaml
backend:
  environment:
    SETUP_ENDPOINT_ENABLED: "true"
```

**Fixed Code:**
```yaml
backend:
  environment:
    # Disable setup endpoint after initial setup
    SETUP_ENDPOINT_ENABLED: "${SETUP_ENDPOINT_ENABLED:-false}"
```

**Update backend to verify:**
```go
// backend/internal/api/handlers.go
func SetupHandler(c *gin.Context) {
    setupEnabled := os.Getenv("SETUP_ENDPOINT_ENABLED") == "true"

    if !setupEnabled {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "Setup endpoint is disabled. Use SETUP_ENDPOINT_ENABLED=true in development only.",
        })
        return
    }

    // ... rest of setup logic
}
```

**Verification:**
```bash
# Verify endpoint is disabled by default
curl -X POST http://localhost:8080/api/v1/auth/setup
# Should return 403 Forbidden

# Enable for development
SETUP_ENDPOINT_ENABLED=true ./pganalytics-api
```

---

### TASK 6: Implement Token Blacklist (3-4 hours)

**Component:** Token revocation via Redis

**Step 1 - Add Redis configuration:**
```go
// backend/internal/config/config.go
type Config struct {
    RedisURL    string
    RedisEnabled bool
    // ... other fields
}

func LoadConfig() *Config {
    return &Config{
        RedisURL: os.Getenv("REDIS_URL"),
        RedisEnabled: os.Getenv("REDIS_ENABLED") == "true",
    }
}
```

**Step 2 - Create token blacklist service:**
```go
// backend/internal/auth/blacklist.go
package auth

import (
    "context"
    "github.com/redis/go-redis/v9"
    "time"
)

type TokenBlacklist struct {
    client *redis.Client
}

func NewTokenBlacklist(redisURL string) (*TokenBlacklist, error) {
    opts, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, err
    }

    client := redis.NewClient(opts)
    return &TokenBlacklist{client: client}, nil
}

// RevokeToken adds token to blacklist until expiration
func (tb *TokenBlacklist) RevokeToken(ctx context.Context, token string, expiresAt time.Time) error {
    ttl := time.Until(expiresAt)
    if ttl <= 0 {
        return nil // Token already expired
    }

    return tb.client.Set(ctx, "blacklist:"+token, "true", ttl).Err()
}

// IsBlacklisted checks if token is revoked
func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
    val, err := tb.client.Get(ctx, "blacklist:"+token).Result()
    if err == redis.Nil {
        return false, nil // Not blacklisted
    }
    return val == "true", err
}
```

**Step 3 - Add to middleware:**
```go
// backend/internal/api/middleware.go
func AuthMiddleware(blacklist *auth.TokenBlacklist) gin.HandlerFunc {
    return func(c *gin.Context) {
        token, err := extractToken(c)
        if err != nil {
            c.JSON(401, gin.H{"error": "missing token"})
            c.Abort()
            return
        }

        // Check if token is blacklisted
        isBlacklisted, _ := blacklist.IsBlacklisted(c.Request.Context(), token)
        if isBlacklisted {
            c.JSON(401, gin.H{"error": "token revoked"})
            c.Abort()
            return
        }

        // ... rest of auth logic
        c.Next()
    }
}
```

**Step 4 - Add logout endpoint:**
```go
// backend/internal/api/handlers.go
func LogoutHandler(blacklist *auth.TokenBlacklist) gin.HandlerFunc {
    return func(c *gin.Context) {
        token, _ := extractToken(c)
        claims, _ := parseToken(token)

        // Revoke the token
        blacklist.RevokeToken(c.Request.Context(), token, claims.ExpiresAt)

        c.JSON(200, gin.H{"message": "logged out successfully"})
    }
}
```

**Step 5 - Update docker-compose:**
```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data

backend:
  environment:
    REDIS_ENABLED: "true"
    REDIS_URL: "redis://redis:6379/0"
  depends_on:
    - redis

volumes:
  redis_data:
```

**Verification:**
```bash
# Test token revocation
TOKEN=$(curl -s http://localhost:8080/api/v1/auth/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

# Token should work
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/user

# Logout
curl -H "Authorization: Bearer $TOKEN" \
  -X POST http://localhost:8080/api/v1/auth/logout

# Token should not work now
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/user
# Should return 401 Unauthorized
```

---

### TASK 7: Migrate JWT to httpOnly Cookies (2 hours)

**Files:**
- `frontend/src/api/authApi.ts`
- `backend/internal/api/handlers.go`

**Step 1 - Backend: Set httpOnly cookie:**
```go
// backend/internal/api/handlers.go
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    c.ShouldBindJSON(&req)

    // ... validate credentials

    token := generateToken(user)

    // Set httpOnly, Secure, SameSite cookie
    c.SetCookie(
        "auth_token",              // name
        token,                      // value
        3600,                       // max age (1 hour)
        "/",                        // path
        "localhost",                // domain (change in production)
        !isProduction(),            // secure (only HTTPS in prod)
        true,                       // httpOnly (not accessible via JS)
    )

    // Also set CSRF token for mutations
    csrfToken := generateCSRFToken()
    c.SetCookie(
        "csrf_token",
        csrfToken,
        3600,
        "/",
        "localhost",
        !isProduction(),
        false,  // Accessible to JS for CSRF headers
    )

    c.JSON(200, gin.H{
        "message": "logged in",
        "csrf_token": csrfToken, // Return to frontend for headers
    })
}
```

**Step 2 - Backend: Configure CORS for credentials:**
```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")
        c.Next()
    }
}
```

**Step 3 - Frontend: Update API client:**
```typescript
// frontend/src/api/client.ts
import axios from 'axios';

const client = axios.create({
  baseURL: 'http://localhost:8080',
  withCredentials: true,  // ✅ Send cookies with requests
});

// Add CSRF token to all requests
client.interceptors.request.use((config) => {
  const csrfToken = document.cookie
    .split('; ')
    .find(row => row.startsWith('csrf_token='))
    ?.split('=')[1];

  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken;
  }

  return config;
});

export default client;
```

**Step 4 - Frontend: Update auth API:**
```typescript
// frontend/src/api/authApi.ts
import client from './client';

export const authApi = {
  login: async (username: string, password: string) => {
    // Token is set as httpOnly cookie automatically
    const response = await client.post('/api/v1/auth/login', {
      username,
      password,
    });

    // Store CSRF token for later use
    const csrfToken = response.data.csrf_token;
    localStorage.setItem('csrf_token', csrfToken);

    return response.data;
  },

  logout: async () => {
    // Cookie is sent automatically
    await client.post('/api/v1/auth/logout');
    // Clear CSRF token
    localStorage.removeItem('csrf_token');
  },
};
```

**Step 5 - Frontend: Remove localStorage auth:**
```typescript
// ❌ DELETE THIS:
// localStorage.getItem('auth_token')
// localStorage.setItem('auth_token', token)

// ✅ USE THIS INSTEAD:
// Cookies are sent automatically with withCredentials: true
```

**Verification:**
```bash
# Test login
curl -v -c cookies.txt -b cookies.txt \
  -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Should see Set-Cookie header
# Set-Cookie: auth_token=...; HttpOnly; Secure; SameSite=Strict

# Test that cookie is sent in subsequent requests
curl -b cookies.txt http://localhost:8080/api/v1/user
# Should work with authenticated user
```

---

### TASK 8: Fix E2E Test Credentials (1-2 hours)

**File:** `frontend/e2e/tests/05-user-management.spec.ts`

**Step 1 - Fix login credentials:**
```typescript
// Before (WRONG)
await loginPage.login('demo@pganalytics.com', 'password123');

// After (CORRECT)
await loginPage.login('admin', 'admin');
```

**Step 2 - Remove silent error catching:**
```typescript
// Before
try {
  await usersPage.expectUserInList(testEmail);
} catch {
  console.log('Error ignored');  // ❌ Problem hidden
}

// After
await usersPage.expectUserInList(testEmail);
// ✅ Throws error if condition not met
```

**Step 3 - Add API response validation:**
```typescript
test('should fetch users with correct response format', async ({ page }) => {
  await page.goto('/');
  await loginPage.login('admin', 'admin');

  // Intercept API response
  const response = await page.waitForResponse(
    response => response.url().includes('/api/v1/users') && response.status() === 200
  );

  const data = await response.json();

  // Validate response format
  expect(data).toHaveProperty('data');
  expect(data).toHaveProperty('page');
  expect(data).toHaveProperty('page_size');
  expect(data).toHaveProperty('total');
  expect(Array.isArray(data.data)).toBe(true);
});
```

**Step 4 - Run tests:**
```bash
cd frontend
npm run test:e2e
# Should see passing tests
```

---

### TASK 9: Fix Collector Integration Tests (90 min)

**File:** `collector/tests/integration/collector_test.cpp`

**Step 1 - Debug connection issues:**
```bash
cd collector
cmake -B build
make -C build test VERBOSE=1 2>&1 | tee test.log

# Look for failure messages
grep -i "error\|fail" test.log
```

**Step 2 - Common fixes:**
```cpp
// Fix 1: Mock PostgreSQL connection
TEST(CollectorTest, ConnectToMockPostgres) {
    MockConnection conn;
    EXPECT_CALL(conn, connect()).WillOnce(Return(true));

    Collector collector(&conn);
    ASSERT_TRUE(collector.start());
}

// Fix 2: Ensure plugin initialization
TEST(CollectorTest, InitializePlugins) {
    Collector collector;
    collector.registerPlugin(make_unique<PostgresPlugin>());

    ASSERT_EQ(collector.pluginCount(), 1);
}
```

**Step 3 - Run tests again:**
```bash
make -C build test
# Should see 19/19 passing
```

---

## ✅ COMPLETION CHECKLIST

**Day 1 (Tasks 1-5):**
- [ ] Fix MD5 UUID (30 min)
- [ ] Fix CORS (20 min)
- [ ] Enable DB SSL (20 min)
- [ ] Remove hardcoded credentials (30 min)
- [ ] Fix setup endpoint (10 min)
- **Total: 2 hours**

**Day 2 (Task 6):**
- [ ] Implement token blacklist with Redis (3-4 hours)
- **Total: 4 hours**

**Day 3 (Task 7):**
- [ ] Migrate to httpOnly cookies (2 hours)
- **Total: 2 hours**

**Day 4 (Tasks 8-9):**
- [ ] Fix E2E tests (1-2 hours)
- [ ] Fix collector integration tests (90 min)
- **Total: 2.5-3.5 hours**

**TOTAL WEEK 1: 10.5-11.5 hours**

---

## 🧪 VALIDATION

After completing all tasks, run:

```bash
# Backend tests
cd backend
go test ./...
go test -race ./...  # Check for race conditions

# Frontend tests
cd ../frontend
npm run test:unit
npm run test:e2e

# Collector tests
cd ../collector
cmake -B build && make -C build test

# Integration tests
./scripts/run-integration-tests.sh

# Security checks
./scripts/security-scan.sh
```

Expected results:
- ✅ All tests passing
- ✅ Zero race conditions
- ✅ No XSS vulnerabilities (localStorage token removed)
- ✅ CORS configured properly
- ✅ Database connections encrypted
- ✅ Tokens blacklisted on logout
- ✅ No hardcoded credentials

---

## 📝 DOCUMENTATION UPDATES

After completing tasks, update:

1. `SECURITY.md` - Document new security practices
2. `DEPLOYMENT.md` - Document credential generation
3. `README.md` - Update setup instructions
4. `.env.example` - Ensure all variables documented

---

**Next:** After completing Week 1, proceed to Phase 2 (Testing & Validation)
