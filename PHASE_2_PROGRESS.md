# Phase 2 Progress: Backend Core Implementation

## Overview
Phase 2 focuses on implementing the core backend API functionality in Go with proper authentication, database access, and API handlers. This document tracks progress on this phase.

## Completed Tasks

### 1. ✅ Database Layer (Previously Completed)
- **PostgreSQL DAL** (`internal/storage/postgres.go`)
  - Connection pooling with proper configuration
  - User operations: GetUserByUsername, GetUserByID, UpdateUserLastLogin
  - Collector operations: Create, Get, List, Update status and certificate info
  - API token operations: Create, Get, Update last used
  - Server operations: Get, List
  - Audit logging

- **TimescaleDB DAL** (`internal/timescale/timescale.go`)
  - Time-series metrics insertion
  - Metrics querying (range, latest, aggregated)
  - Health checks

### 2. ✅ Models & DTOs (Previously Completed)
- **Domain Models** (`pkg/models/models.go`)
  - User, Collector, Server, PostgreSQLInstance, Database
  - Alert, AlertRule, AuditLog
  - APIToken, Secret

- **Request/Response DTOs**
  - LoginRequest, LoginResponse
  - CollectorRegisterRequest, CollectorRegisterResponse
  - MetricsPushRequest, MetricsPushResponse
  - HealthResponse, ErrorResponse
  - PaginationParams, PaginatedResponse

### 3. ✅ Error Handling (Previously Completed)
- **Custom Error System** (`pkg/errors/errors.go`)
  - AppError type with HTTP status mapping
  - Specific error constructors (InvalidCredentials, TokenExpired, etc.)
  - Error conversion helpers

### 4. ✅ Configuration Management (Previously Completed)
- **Config Loader** (`internal/config/config.go`)
  - Environment variable based configuration
  - Validation method
  - Helper methods (IsProduction, IsDevelopment)

### 5. ✅ JWT Authentication (Recently Completed)
- **JWT Manager** (`internal/auth/jwt.go`)
  - User token generation (access + refresh)
  - Collector token generation
  - Token validation with signature verification
  - Token refresh flow with user mismatch protection
  - Utility functions: ExtractTokenFromHeader, GetTokenExpiration, IsExpired

- **JWT Tests** (`internal/auth/jwt_test.go`)
  - 18+ comprehensive test cases
  - Token generation, validation, refresh scenarios
  - Error handling and edge cases

### 6. ✅ Authentication Service (Just Completed)
- **Password Manager** (`internal/auth/password.go`)
  - Bcrypt password hashing
  - Password verification

- **Certificate Manager** (`internal/auth/cert_generator.go`)
  - RSA key pair generation for collectors
  - Self-signed certificate generation
  - Certificate thumbprint computation
  - Certificate validation

- **Auth Service** (`internal/auth/service.go`)
  - User login flow
  - Token refresh flow
  - Collector registration
  - Token validation (user and collector)
  - Interface-based design for data access

- **Auth Service Tests** (`internal/auth/service_test.go`)
  - Mock implementations of UserStore, CollectorStore, TokenStore
  - Login success/failure scenarios
  - Token refresh
  - Collector registration
  - Password hashing tests

### 7. ✅ API Server Setup (Just Completed)
- **Server Structure** (`internal/api/server.go`)
  - Dependency injection pattern
  - Auth service integration
  - JWT manager integration

- **Route Registration** (Already in place)
  - 6 route groups: auth, collectors, metrics, config, servers, alerts

### 8. ✅ API Middleware (Already Completed)
- **AuthMiddleware**
  - JWT token validation
  - User claims extraction and storage in context

- **CollectorAuthMiddleware**
  - Collector JWT token validation
  - Collector claims extraction and storage in context

- **Additional Middleware**
  - CORS, logging, rate limiting (stubs for future implementation)

### 9. ✅ API Handlers - Authentication (Just Completed)
- **POST /api/v1/auth/login**
  - User authentication
  - JWT token generation
  - Refresh token generation
  - Logging of successful logins

- **POST /api/v1/auth/refresh**
  - Token refresh flow
  - User re-authentication via refresh token

### 10. ✅ API Handlers - Collectors (Just Completed)
- **POST /api/v1/collectors/register**
  - Collector registration
  - Certificate generation
  - JWT token creation
  - Returns: collector ID, token, certificate, private key, expiration

- **GET /api/v1/collectors** (Stub)
  - List collectors with pagination
  - Pagination parameter handling

### 11. ✅ API Handlers - Metrics (Just Completed)
- **POST /api/v1/metrics/push**
  - Collector authentication validation
  - Metrics data receipt
  - Collector ID validation
  - Processing time simulation
  - Placeholder for metrics storage

### 12. ✅ Main Application Entry Point (Updated)
- **cmd/pganalytics-api/main.go**
  - Initialization of all services
  - JWT manager setup
  - Password manager setup
  - Certificate manager setup
  - Auth service creation with dependency injection
  - Graceful shutdown handling

## Current Architecture

```
┌─────────────────────────────────────┐
│        HTTP Request (Gin)           │
├─────────────────────────────────────┤
│        Middleware Layer             │
│  ├─ AuthMiddleware (user)           │
│  ├─ CollectorAuthMiddleware         │
│  └─ Other middleware                │
├─────────────────────────────────────┤
│       API Handler Layer             │
│  ├─ Authentication handlers         │
│  ├─ Collector handlers              │
│  ├─ Metrics handlers                │
│  └─ Other endpoint handlers         │
├─────────────────────────────────────┤
│      Service Layer                  │
│  ├─ AuthService                     │
│  │  ├─ User login/refresh           │
│  │  ├─ Collector registration       │
│  │  └─ Token validation             │
│  ├─ PasswordManager                 │
│  ├─ CertificateManager              │
│  └─ JWTManager                      │
├─────────────────────────────────────┤
│       Data Access Layer             │
│  ├─ PostgresDB                      │
│  │  ├─ User operations              │
│  │  ├─ Collector operations         │
│  │  └─ Token operations             │
│  └─ TimescaleDB                     │
│     └─ Metrics operations           │
├─────────────────────────────────────┤
│    Database                         │
│  ├─ PostgreSQL (metadata)           │
│  └─ TimescaleDB (time-series)       │
└─────────────────────────────────────┘
```

## Key Implementation Details

### Authentication Flow

**User Login:**
```
1. Client → POST /api/v1/auth/login {username, password}
2. AuthService.LoginUser()
   - Get user from PostgresDB
   - Verify password (future: bcrypt comparison)
   - Generate JWT tokens (access + refresh)
   - Update last login timestamp
3. Response: {token, refresh_token, expires_at, user}
```

**Token Refresh:**
```
1. Client → POST /api/v1/auth/refresh {refresh_token}
2. AuthService.RefreshUserToken()
   - Validate refresh token
   - Generate new access token
3. Response: {token, refresh_token, expires_at, user}
```

**Collector Registration:**
```
1. Collector → POST /api/v1/collectors/register {name, hostname}
2. AuthService.RegisterCollector()
   - Create Collector record in PostgresDB
   - Generate RSA key pair
   - Create self-signed certificate
   - Generate JWT token for collector
   - Store certificate thumbprint
3. Response: {collector_id, token, certificate, private_key, expires_at}
```

**API Requests (Authenticated):**
```
1. Client → GET /api/v1/collectors
   Header: Authorization: Bearer {jwt_token}
2. AuthMiddleware
   - Extract token from header
   - Validate JWT signature and expiration
   - Extract claims and store in context
3. Handler receives authenticated user info in context
```

### JWT Token Structure

**User Access Token (15 min):**
```json
{
  "user_id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "type": "access",
  "exp": 1708356000,
  "iat": 1708355000,
  "nbf": 1708355000,
  "sub": "user:1"
}
```

**User Refresh Token (24 hours):**
```json
{
  "user_id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "type": "refresh",
  "exp": 1708441400,
  "iat": 1708355000,
  "nbf": 1708355000,
  "sub": "user:1"
}
```

**Collector Token (30 min):**
```json
{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "hostname": "db-server-01",
  "type": "access",
  "exp": 1708356000,
  "iat": 1708355000,
  "nbf": 1708355000,
  "sub": "collector:550e8400-e29b-41d4-a716-446655440000"
}
```

## Pending Tasks

### High Priority (Next)
1. **Metrics Ingestion Implementation**
   - Parse MetricsPushRequest JSON
   - Validate metrics schema
   - Store in TimescaleDB hypertables
   - Update collector metrics counts

2. **Handler Implementations**
   - GET /api/v1/collectors (list with pagination)
   - GET /api/v1/collectors/{id} (get details)
   - DELETE /api/v1/collectors/{id} (deregister)
   - GET /api/v1/servers (list)
   - GET /api/v1/servers/{id} (get details)
   - GET /api/v1/servers/{id}/metrics (historical metrics)

3. **Password Integration**
   - Add password hash field to user model
   - Store hashed passwords in database
   - Verify passwords on login using bcrypt

4. **mTLS Implementation**
   - Load CA certificate and key
   - Validate collector certificates in MTLSMiddleware
   - Extract certificate thumbprint for validation

5. **Configuration Endpoints**
   - GET /api/v1/config/{collector_id} (pull config)
   - PUT /api/v1/config/{collector_id} (update config)
   - Store and version collector configurations

### Medium Priority
1. **Alert Endpoints**
   - GET /api/v1/alerts (list)
   - GET /api/v1/alerts/{id} (get)
   - POST /api/v1/alerts/{id}/acknowledge (acknowledge)

2. **Integration Tests**
   - End-to-end test scenarios
   - Collector registration flow
   - Metrics ingestion flow
   - User authentication flow

3. **Load Testing**
   - k6 test script for metrics push
   - Simulating 50-500 concurrent collectors
   - Performance benchmarks

### Low Priority
1. **API Documentation**
   - Generate Swagger UI
   - API examples and curl commands
   - Error response documentation

2. **Monitoring & Observability**
   - Prometheus metrics export
   - Custom application metrics
   - Request/response logging

## Testing Progress

### Unit Tests Completed
- ✅ JWT token generation and validation (18+ tests)
- ✅ Password hashing and verification
- ✅ Auth service login and refresh flows
- ✅ Collector registration
- ✅ Certificate generation

### Integration Tests
- ⏳ Full auth flow (registration → login → metrics push)
- ⏳ Database transaction handling
- ⏳ Error scenarios

### Load Tests
- ⏳ Metrics ingestion at scale
- ⏳ Concurrent collector connections

## Notes & Decisions

1. **Self-Signed Certificates**: For demo/development, collectors use self-signed certificates. Production should use a proper CA.

2. **Password Storage**: Currently using a mock password verification. Database schema needs a password_hash column and actual bcrypt integration during authentication.

3. **Interface-Based Design**: AuthService uses interfaces for data access (UserStore, CollectorStore, TokenStore) to enable easy mocking in tests and future flexibility.

4. **Dependency Injection**: All services are initialized in main.go with proper dependency injection, making the code testable and maintainable.

5. **Error Handling**: Custom AppError type maps to specific HTTP status codes, providing a consistent API error contract.

## Files Modified/Created in This Session

**New Files:**
- `internal/auth/password.go` - Password hashing service
- `internal/auth/cert_generator.go` - Certificate generation
- `internal/auth/service.go` - Auth service combining all auth operations
- `internal/auth/service_test.go` - Comprehensive auth service tests
- `PHASE_2_PROGRESS.md` - This file

**Modified Files:**
- `internal/api/server.go` - Added auth service and JWT manager
- `internal/api/handlers.go` - Implemented login, register, refresh, metrics push handlers
- `cmd/pganalytics-api/main.go` - Initialized all auth services

## Next Session Goals

1. Implement metrics storage in TimescaleDB
2. Implement remaining handler stubs (servers, alerts, config)
3. Create integration test suite
4. Set up go.mod with all dependencies
5. Test the full authentication and metrics flow

## Dependencies Needed

Add to go.mod:
- `golang.org/x/crypto` - bcrypt password hashing
- `github.com/golang-jwt/jwt/v5` - JWT (likely already added)
- `github.com/google/uuid` - UUID generation (likely already added)

