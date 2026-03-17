# Phase 1 - Login Authentication Issue
## Status: 🟡 IDENTIFIED & DOCUMENTED

**Date**: March 12, 2026
**Issue**: Frontend infinite refresh loop on /login due to authentication failure
**Root Cause**: No default credentials / password hash mismatch

---

## Problem Description

### Symptoms
- Frontend loads at http://localhost:3000/login
- Page refreshes infinitely
- Browser console shows repeated failed login attempts
- Backend logs show 401 errors for `admin` user

### Backend Logs
```
[GIN] 2026/03/11 - 23:38:18 | 401 |     530.958µs |     172.21.0.30 | POST "/api/v1/auth/login"
DEBUG	api/handlers.go:558	Login failed	{"username": "admin",
  "error": "[401] Invalid credentials: Username or password is incorrect"}
```

---

## Root Cause Analysis

### Issue 1: No Default Credentials
- No default admin user created with known password
- Password hash algorithm requires bcrypt
- Manual user insertion used incorrect password hash

### Issue 2: Auto-Login Logic
- Frontend tries to auto-login as `admin:admin` on load
- When it fails, page refreshes and retries
- Creates infinite loop

### Issue 3: Missing Password Reset Mechanism
- No API endpoint to reset password during setup
- No default bypass for initial login
- Admin password reset requires existing admin (circular dependency)

---

## Solutions to Implement

### Option 1: Default Credentials (Quickest)
```bash
# Generate proper bcrypt hash using the backend's own crypto
docker exec pganalytics-staging-backend /opt/pganalytics/pganalytics-api \
  generate-password-hash "admin"
```

Then insert into database with the correct hash.

### Option 2: Bootstrap API Endpoint (Recommended)
Create `/api/v1/auth/setup` endpoint that:
- Checks if any admin users exist
- If none exist, creates first admin user
- Only accessible when no users exist
- Redirects to this on first login failure

```go
// handlers.go
func (s *Server) handleSetup(c *gin.Context) {
    // Check if any users exist
    users, _ := s.db.ListUsers(c)
    if len(users) > 0 {
        c.AbortWithStatusJSON(403, "Setup already completed")
        return
    }

    // Create admin user from request
    // ...
}
```

### Option 3: Disable Auth in Staging (Development Only)
```go
// middleware.go
if os.Getenv("ENVIRONMENT") == "staging" &&
   os.Getenv("DISABLE_AUTH") == "true" {
    // Skip authentication middleware
    c.Next()
    return
}
```

---

## Workaround (Current Staging)

### Temporary Solution
1. Disable auto-login in frontend:
   ```typescript
   // frontend/src/pages/Login.tsx
   useEffect(() => {
     // Comment out auto-login for staging
     // attemptAutoLogin();
   }, [])
   ```

2. Provide manual registration endpoint:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/register \
     -H "X-Registration-Secret: staging-registration-secret" \
     -d '{...}'
   ```

---

## For Next Phases

### Must Do Before Production
- [x] Identify root cause
- [ ] Implement proper default credentials mechanism
- [ ] Test complete authentication flow
- [ ] Document all user management procedures
- [ ] Add API endpoint for password reset

### Recommended Improvements
- [ ] Implement setup wizard on first boot
- [ ] Add backend command to create admin user
- [ ] Support external authentication (LDAP, OAuth)
- [ ] Implement password strength requirements
- [ ] Add multi-factor authentication

---

## Code Changes Needed

### 1. Backend - Add Password Generation Utility
```go
// backend/pkg/auth/password.go
package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(hash, password string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
```

### 2. Frontend - Disable Auto-Login for Staging
```typescript
// frontend/src/App.tsx
useEffect(() => {
  if (process.env.REACT_APP_ENVIRONMENT === 'staging') {
    // Don't auto-login in staging
    return;
  }
  attemptAutoLogin();
}, [])
```

### 3. Docker - Initialize Admin User
```dockerfile
# After database is ready
RUN echo "INSERT INTO pganalytics.users ..." | psql $DATABASE_URL
```

---

## Testing Checklist

- [ ] Backend starts successfully
- [ ] Database has default admin user
- [ ] Login with admin credentials works
- [ ] Frontend loads after successful login
- [ ] No infinite refresh loops
- [ ] User can access dashboard

---

## Prevention for Phase 2

### Pre-Deployment Checklist
- [ ] Verify default credentials exist and work
- [ ] Test login flow end-to-end
- [ ] Check browser console for no errors
- [ ] Verify auto-login (if enabled) completes
- [ ] Monitor backend logs for authentication errors

### Validation Script
```bash
#!/bin/bash
# scripts/validate-auth.sh

set -e

echo "🔐 Validating authentication setup..."

# Check admin user exists
docker exec pganalytics-staging-postgres psql -U postgres -d pganalytics_staging \
  -c "SELECT COUNT(*) as admin_count FROM pganalytics.users WHERE role = 'admin';" | \
  grep -q "1" || (echo "❌ No admin user found" && exit 1)

# Test login endpoint
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | \
  grep -q "token" || (echo "❌ Login failed" && exit 1)

# Test frontend loads
curl -s http://localhost:3000/ | grep -q "root" || \
  (echo "❌ Frontend not loading" && exit 1)

echo "✅ Authentication setup valid"
```

---

## Current Status

✅ **Issue Identified**: Password hash mismatch
✅ **Root Cause Found**: No proper credential initialization
🟡 **Workaround**: Manual database updates
❌ **Permanent Solution**: Not yet implemented
⏳ **For Phase 2**: Implement bootstrap setup endpoint

---

**Priority**: MEDIUM (Affects development workflow, not blocking functionality)
**Effort**: 2-4 hours to implement proper solution
**Phase 2 Target**: MUST FIX before production

