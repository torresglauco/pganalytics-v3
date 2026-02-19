# Phase 2: Backend Core Implementation Plan

## Overview

**Objective**: Implement a production-ready backend API with complete authentication, metrics ingestion, and database integration.

**Duration**: ~2 weeks (Semanas 4-6)

**Deliverables**:
- ✅ Complete REST API endpoints (15+ endpoints)
- ✅ JWT token management (generation, validation, refresh)
- ✅ mTLS certificate handling
- ✅ Metrics ingestion and validation
- ✅ TimescaleDB data insertion
- ✅ Database models (SQLC)
- ✅ Unit tests (>70% coverage)
- ✅ Integration tests
- ✅ Swagger/OpenAPI documentation
- ✅ Error handling and logging

## Implementation Order (Database-First Approach)

### 1. Core Models & Database Layer

**Files to create/modify**:
- `pkg/models/models.go` - Data structures
- `internal/storage/postgres.go` - PostgreSQL layer
- `internal/timescale/timescale.go` - TimescaleDB layer
- `backend/queries/` - SQLC query files

**Tasks**:
1. Define Go structs for all tables
2. Setup database connection pooling
3. Create SQLC query files for:
   - Collector management (register, list, update status)
   - User management (login, token management)
   - Metrics queries (insert, select)
4. Generate SQLC code

### 2. Authentication & Security

**Files to create/modify**:
- `internal/auth/jwt.go` - JWT implementation
- `internal/auth/mtls.go` - mTLS verification
- `internal/auth/middleware.go` - Auth middleware
- `internal/auth/cert_generator.go` - Certificate generation

**Tasks**:
1. Implement JWT token generation
2. Implement JWT token validation
3. Implement token refresh logic
4. Implement mTLS certificate verification
5. Create auth middleware
6. Add certificate generation for collectors

### 3. API Endpoints

**Files to create/modify**:
- `internal/api/routes.go` - Route definitions
- `internal/api/handlers.go` - HTTP handlers
- `internal/api/response.go` - Response formatting
- `internal/api/errors.go` - Error handling

**Endpoints to implement**:

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh JWT token

#### Collectors
- `POST /api/v1/collectors/register` - Register new collector
- `GET /api/v1/collectors` - List collectors (admin)
- `GET /api/v1/collectors/{id}` - Get collector details
- `DELETE /api/v1/collectors/{id}` - Deregister collector

#### Metrics
- `POST /api/v1/metrics/push` - Ingest metrics (core endpoint)
- `GET /api/v1/servers/{id}/metrics` - Query historical metrics
- `GET /api/v1/servers` - List monitored servers

#### Configuration
- `GET /api/v1/config/{collector_id}` - Pull collector config
- `PUT /api/v1/config/{collector_id}` - Update config (admin)

#### Health & System
- `GET /api/v1/health` - System health check
- `GET /api/v1/version` - API version

### 4. Business Logic Layer

**Files to create/modify**:
- `internal/collector/service.go` - Collector management
- `internal/metrics/receiver.go` - Metrics validation and processing
- `internal/metrics/storage.go` - Metrics data insertion
- `internal/config/service.go` - Configuration management

**Tasks**:
1. Collector registration workflow
2. Certificate generation and storage
3. Metrics JSON schema validation
4. Metrics decompression and insertion
5. Configuration pull/update logic
6. User session management

### 5. Testing

**Files to create**:
- `internal/auth/jwt_test.go` - JWT tests
- `internal/auth/mtls_test.go` - mTLS tests
- `internal/collector/service_test.go` - Service tests
- `internal/metrics/receiver_test.go` - Validation tests
- `tests/integration/collector_test.go` - E2E tests
- `tests/integration/metrics_test.go` - Metrics flow tests

**Coverage target**: 70%+

### 6. Documentation

**Files to create/modify**:
- `backend/main.go` - Swagger annotations
- Generate Swagger UI via swag init
- Update README with API examples

## Key Implementation Details

### Database Connection

```go
// internal/storage/postgres.go
type PostgresDB struct {
    db *sql.DB
    pool *pgxpool.Pool
}

func NewPostgresDB(connString string) (*PostgresDB, error)
func (p *PostgresDB) Close() error
```

### JWT Implementation

```go
// internal/auth/jwt.go
type TokenClaims struct {
    UserID    int
    Username  string
    Role      string
    ExpiresAt time.Time
    jwt.RegisteredClaims
}

func GenerateToken(user *User) (string, error)
func ValidateToken(tokenString string) (*TokenClaims, error)
func RefreshToken(tokenString string) (string, error)
```

### Metrics Ingestion Flow

```
1. Collector sends POST /api/v1/metrics/push
   ├─ TLS handshake (mTLS)
   ├─ Authorization header (JWT)
   └─ Gzip compressed JSON body

2. Backend validates
   ├─ mTLS certificate
   ├─ JWT token
   └─ JSON schema

3. Decompress metrics
   ├─ gzip decompression
   └─ JSON parsing

4. Process metrics
   ├─ Validate data types
   ├─ Extract collector info
   └─ Insert to TimescaleDB

5. Return response
   ├─ 200 OK with metrics count
   ├─ Updated config version
   └─ Next check-in interval
```

### Error Handling

```go
type APIError struct {
    Code    int
    Message string
    Details string
}

Errors:
- 400 Bad Request (validation failed)
- 401 Unauthorized (invalid auth)
- 403 Forbidden (insufficient permissions)
- 404 Not Found
- 500 Internal Server Error
```

## Dependencies to Add

```
require (
    github.com/golang-jwt/jwt/v5
    github.com/google/uuid
    github.com/sqlc-dev/sqlc
    github.com/jackc/pgx/v5
    github.com/jackc/pgx/v5/pgxpool
    github.com/swaggo/swag
    github.com/swaggo/gin-swagger
    golang.org/x/crypto
)
```

## Testing Strategy

### Unit Tests
- JWT token generation/validation
- mTLS certificate verification
- Metrics schema validation
- Business logic (service layer)

### Integration Tests
- API endpoint testing with real database
- Collector registration workflow
- Metrics ingestion end-to-end
- Authentication flows

### Load Tests (Optional for Phase 2)
- Metrics ingestion throughput
- Concurrent collector handling
- Database performance

## Documentation

### Code Examples to Include

```bash
# Collector Registration
curl -X POST https://api.pganalytics.local/api/v1/collectors/register \
  -H "Content-Type: application/json" \
  -d '{"name": "prod-db-01", "hostname": "db.example.com"}'

# Push Metrics
curl -X POST https://api.pganalytics.local/api/v1/metrics/push \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Encoding: gzip" \
  --cert collector.crt --key collector.key \
  --data-binary @metrics.json.gz

# Query Metrics
curl -X GET https://api.pganalytics.local/api/v1/servers/1/metrics \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

## Checklist

### Week 1 - Database & Models
- [ ] Create SQLC query files
- [ ] Implement PostgreSQL layer
- [ ] Implement TimescaleDB layer
- [ ] Create model structs
- [ ] Add connection pooling
- [ ] Write database tests

### Week 2 - Authentication & API
- [ ] Implement JWT (generation, validation, refresh)
- [ ] Implement mTLS verification
- [ ] Implement certificate generation
- [ ] Create API routes
- [ ] Implement all handlers
- [ ] Add middleware
- [ ] Write integration tests

### Week 3 (overlapping) - Polish & Documentation
- [ ] Swagger/OpenAPI generation
- [ ] API documentation
- [ ] Error handling refinement
- [ ] Logging implementation
- [ ] Performance optimization
- [ ] Code coverage to 70%+

## Success Criteria

✅ All 15+ endpoints implemented and tested
✅ JWT token system working correctly
✅ mTLS authentication functioning
✅ Metrics ingestion end-to-end working
✅ Database persistence verified
✅ 70%+ test coverage
✅ Swagger docs generated and accessible
✅ Error handling comprehensive
✅ Logging structured and useful
✅ No SQL injection vulnerabilities
✅ No hardcoded secrets
✅ Docker build passes

## Git Strategy

Create feature branches for each major component:

```bash
git checkout -b feat/database-layer
git checkout -b feat/jwt-auth
git checkout -b feat/mtls-support
git checkout -b feat/api-endpoints
git checkout -b feat/metrics-ingestion
git checkout -b feat/testing-integration
```

Merge back to main with clean commit history.

---

**Status**: Ready to implement
**Start Date**: 2026-02-19
**Target Completion**: ~2 weeks
