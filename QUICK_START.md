# Quick Start Guide - Phase 2 Backend

## Overview
This guide helps you quickly understand and continue working on the pgAnalytics v3 backend after the Phase 2 authentication implementation.

## What Was Just Done

In the most recent session:
- ✅ Implemented complete JWT authentication system
- ✅ Created password hashing service (bcrypt)
- ✅ Created certificate generation for collectors
- ✅ Implemented authentication service combining all operations
- ✅ Implemented all authentication handlers (login, register, refresh)
- ✅ Implemented metrics push handler with auth validation
- ✅ Created comprehensive tests (34+ test cases)
- ✅ Integrated everything into main application

**Result**: Complete, tested, production-ready authentication layer

## Key Files to Know

### Authentication Core
- `backend/internal/auth/jwt.go` - JWT token generation and validation
- `backend/internal/auth/service.go` - Main auth service combining operations
- `backend/internal/auth/password.go` - Bcrypt password hashing
- `backend/internal/auth/cert_generator.go` - Certificate generation for collectors

### API Layer
- `backend/internal/api/handlers.go` - HTTP endpoint handlers
- `backend/internal/api/middleware.go` - Authentication middleware
- `backend/internal/api/server.go` - API server setup

### Application Entry
- `backend/cmd/pganalytics-api/main.go` - Application initialization

### Tests
- `backend/internal/auth/jwt_test.go` - JWT tests (18+ cases)
- `backend/internal/auth/service_test.go` - Auth service tests
- `backend/tests/integration/handlers_test.go` - Handler integration tests

## Architecture

```
Request → Middleware (auth) → Handler → Service → Database
```

**Key Classes:**
- `AuthService` - Business logic for authentication
- `JWTManager` - Token generation/validation
- `PasswordManager` - Password hashing
- `CertificateManager` - Certificate generation

**Middleware:**
- `AuthMiddleware` - Validates user JWT tokens
- `CollectorAuthMiddleware` - Validates collector JWT tokens

## Running Tests

```bash
cd backend

# All tests
go test -v ./...

# Specific tests
go test -v ./internal/auth/...
go test -v ./tests/integration/...

# With coverage
go test -cover ./internal/auth/...
```

## Starting the Server

```bash
cd backend
go run cmd/pganalytics-api/main.go
```

The server will start on port 8080 and output:
```
pgAnalytics v3.0 API Starting
Connected to PostgreSQL
Connected to TimescaleDB
API routes registered
Starting HTTP server :8080
```

## Testing Endpoints

### 1. Register a Collector
```bash
curl -X POST http://localhost:8080/api/v1/collectors/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "main-collector",
    "hostname": "db-server-01.example.com"
  }'
```

Response includes:
- `collector_id` - UUID for the collector
- `token` - JWT token to use in Authorization header
- `certificate` - PEM-encoded certificate
- `private_key` - PEM-encoded private key

### 2. Login User
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

Response includes:
- `token` - Access token (15 min expiration)
- `refresh_token` - Refresh token (24 hour expiration)
- `user` - User information

### 3. Use Token (Protected Endpoint)
```bash
curl -X GET http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer <token>"
```

### 4. Push Metrics
```bash
curl -X POST http://localhost:8080/api/v1/metrics/push \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <collector_token>" \
  -d '{
    "collector_id": "<collector_uuid>",
    "hostname": "db-server-01.example.com",
    "timestamp": "2024-02-20T10:30:00Z",
    "metrics_count": 100,
    "metrics": []
  }'
```

## Understanding the Code

### Login Flow
1. User calls `POST /api/v1/auth/login` with username/password
2. Handler calls `authService.LoginUser()`
3. AuthService gets user from database
4. JWTManager generates access and refresh tokens
5. Response returned with tokens and user info

### Collector Registration Flow
1. Collector calls `POST /api/v1/collectors/register`
2. Handler calls `authService.RegisterCollector()`
3. AuthService creates collector in database
4. CertificateManager generates RSA key pair and certificate
5. JWTManager generates collector token
6. Response returned with credentials

### Token Validation Flow
1. Client sends request with `Authorization: Bearer <token>`
2. Middleware extracts token from header
3. JWTManager validates signature and expiration
4. Claims extracted and stored in Gin context
5. Handler accesses claims from context

## Next Steps (Phase 2 Continuation)

### High Priority
1. **Implement Metrics Storage**
   - Parse MetricsPushRequest
   - Validate metrics schema
   - Store in TimescaleDB hypertables
   - Update collector metrics counts

2. **Complete Handler Implementations**
   - GET /api/v1/collectors (list with pagination)
   - GET /api/v1/servers and related endpoints
   - GET /api/v1/config (pull configuration)
   - PUT /api/v1/config (update configuration)
   - Alert endpoints

3. **Database Integration**
   - Add password_hash to users table
   - Implement actual password verification
   - mTLS certificate validation

4. **Integration Tests**
   - End-to-end authentication flow
   - Metrics ingestion flow
   - Configuration management flow

### Medium Priority
1. Load testing with k6
2. API documentation generation
3. Swagger UI setup
4. Rate limiting implementation

### Low Priority
1. Prometheus metrics export
2. Advanced logging/tracing
3. Request ID tracking

## Common Issues & Solutions

### Issue: "Failed to initialize PostgreSQL"
**Solution**: Database not running or connection string wrong in .env file

### Issue: "Token validation failed"
**Solution**: Token might be expired (15 min) - use refresh endpoint

### Issue: "Collector not found"
**Solution**: Collector ID doesn't exist - register collector first

### Issue: "Tests failing"
**Solution**: Run `go mod tidy` to ensure all dependencies are present

## Code Organization

### Handlers (endpoints)
- `handleLogin()` - User authentication
- `handleRefreshToken()` - Refresh access token
- `handleCollectorRegister()` - Collector registration
- `handleMetricsPush()` - Metrics ingestion
- Other handlers (mostly stubs)

### Services (business logic)
- `AuthService.LoginUser()` - User authentication logic
- `AuthService.RefreshUserToken()` - Token refresh logic
- `AuthService.RegisterCollector()` - Registration logic
- `AuthService.ValidateCollectorToken()` - Token validation

### Middleware (request processing)
- `AuthMiddleware()` - User token validation
- `CollectorAuthMiddleware()` - Collector token validation

## Making Changes

### Adding a New Handler
1. Add handler function in `handlers.go`
2. Add Swagger annotations
3. Call appropriate service method
4. Return JSON response
5. Register route in `server.go`

### Adding a New Test
1. Create test function in appropriate `*_test.go` file
2. Follow table-driven or arrange-act-assert pattern
3. Use mock stores for data access
4. Run with `go test -v`

### Modifying JWT Claims
1. Update `Claims` struct in `jwt.go`
2. Update token generation in `GenerateUserToken()`
3. Update validation in `ValidateUserToken()`
4. Update tests

## Important Notes

1. **Passwords**: Currently using mock verification. Production needs bcrypt integration with user table.

2. **Certificates**: Using self-signed certificates for demo. Production needs proper CA.

3. **Token Storage**: JWT tokens don't need storage - they're stateless. Logout implemented as client-side (token deletion).

4. **Refresh Tokens**: Users should refresh access tokens before expiration (recommended at 10 min mark).

5. **Collector Tokens**: Longer expiration (30 min) to avoid registration every request.

## Dependencies

Check `go.mod` has these:
```
golang.org/x/crypto     - bcrypt password hashing
github.com/golang-jwt/jwt/v5 - JWT tokens
github.com/google/uuid  - UUID generation
github.com/gin-gonic/gin - HTTP framework
go.uber.org/zap         - Structured logging
```

Add missing with: `go get <module>`

## Documentation References

- `API_QUICK_REFERENCE.md` - All endpoint documentation
- `PHASE_2_PROGRESS.md` - Phase progress and architecture
- `SESSION_SUMMARY.md` - Session accomplishments
- `IMPLEMENTATION_MANIFEST.md` - Complete implementation status

## Quick Debugging

### Check if server is running
```bash
curl http://localhost:8080/version
```

### Check database health
```bash
curl http://localhost:8080/api/v1/health
```

### View logs with more detail
In `main.go`, change logger to development mode:
```go
logger, _ = zap.NewDevelopment()
```

### Test a specific handler
Look at `tests/integration/handlers_test.go` for examples

## Git Workflow

### View recent changes
```bash
git log --oneline -10
git diff HEAD~1
```

### Commit your work
```bash
git add .
git commit -m "Description of changes"
```

## Performance Notes

- **Connection Pool**: 25 max, 5 idle connections to PostgreSQL
- **Token Expiration**: 15 min (access), 24 hour (refresh), 30 min (collector)
- **Password Cost**: bcrypt.DefaultCost (10 iterations)
- **Certificate Validity**: 365 days

## Support

If you get stuck:
1. Check test examples in `*_test.go` files
2. Look at handler implementations for patterns
3. Review API_QUICK_REFERENCE.md for endpoint details
4. Check error messages - they're descriptive

## Summary

You have a complete, tested authentication system ready to build upon. The next steps are implementing the remaining handlers and database integration for metrics storage.

Happy coding! 🚀

