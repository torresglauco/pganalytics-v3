# Troubleshooting: Collectors Not Showing in UI

**Data:** 2026-02-27
**Status:** ISSUE IDENTIFIED - Backend Authentication Bug

## Problem
The "Manage Collectors" page in the UI shows a blank page instead of listing collectors.

## Root Cause
**Backend authentication handler has a nil pointer dereference bug:**
- File: `backend/internal/auth/service.go:66`
- Endpoint: `POST /api/v1/auth/login`
- Error: `runtime error: invalid memory address or nil pointer dereference`

## What's Working
✅ Frontend UI (http://localhost:4000) - Loads correctly
✅ Database - Has collectors table with data
✅ Backend API - Running and responding to health checks
✅ Database migrations - Schema created successfully

## What's NOT Working
❌ `/api/v1/auth/login` endpoint - Crashes with nil pointer panic
❌ User authentication - Cannot get auth token
❌ `/api/v1/collectors` endpoint - Requires valid auth token

## Database Status
```
Collectors table: pganalytics.collectors
- Contains 1 collector: (28bfde91-1f7e-481b-a748-dcd816dfd915)
- Schema: pganalytics (not public!)
- Status: Accessible via psql
```

## Why UI Is Blank
1. UI loads successfully
2. UI tries to call `GET /api/v1/collectors`
3. Backend requires Authorization header (Bearer token)
4. UI has no token (no login happened)
5. GET request returns 401 Unauthorized
6. UI shows blank page (error not displayed)

## Quick Workaround (Browser Console)

In the browser Developer Console (F12), paste this:
```javascript
localStorage.setItem('auth_token', 'temp-token');
window.location.reload();
```

This will add a fake token, allowing the UI to make API calls. (Will get 401, but shows the real error.)

## Proper Fix Needed

Fix the backend authentication handler:

**File:** `backend/internal/auth/service.go` line 66
**Issue:** Nil pointer in login handler
**Solution:** Debug and fix the nil pointer in the authentication service

**File:** `backend/internal/api/handlers.go` line 86
**Related:** Check the `handleLogin` function for nil context

## Temporary Workaround (Backend)

Make the `/api/v1/collectors` endpoint public (no auth required):

**File:** `backend/internal/api/server.go`

Change:
```go
api.GET("/collectors", s.AuthMiddleware(), s.handleListCollectors)
```

To:
```go
api.GET("/collectors", s.handleListCollectors) // Remove auth middleware
```

Then rebuild:
```bash
docker-compose build backend --no-cache
docker-compose up -d backend
```

## Database Schema Issue

Tables are in `pganalytics` schema, but backend might be looking in `public` schema.

**Verify:** Check `backend/internal/config/config.go` for schema configuration.

**Fix:** Either:
1. Set `search_path = pganalytics, public` in PostgreSQL configuration
2. Or update backend queries to explicitly use `pganalytics.` prefix

## Files to Check

1. `backend/internal/auth/service.go` - Line 66
2. `backend/internal/api/handlers.go` - Line 86 (login handler)
3. `backend/internal/api/server.go` - Line 121 (collectors route)
4. `backend/internal/config/config.go` - Schema configuration

## Next Steps

1. **Fix the authentication bug** in the backend
2. **Test login endpoint:** `curl -X POST http://localhost:8080/api/v1/auth/login`
3. **Get valid JWT token** from login response
4. **Call collectors endpoint with token:** `curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/collectors`
5. **Verify collectors appear in UI**

## Collector Data Status

Collector exists in database:
```
ID: 28bfde91-1f7e-481b-a748-dcd816dfd915
Name: (empty - needs update)
Hostname: localhost
Status: active
```

The collector IS registered and active, but unreachable via API due to auth bug.

---

**Issue Type:** Backend code bug
**Severity:** High (blocks UI functionality)
**Component:** Authentication service
**Impact:** Cannot view collectors in web UI
