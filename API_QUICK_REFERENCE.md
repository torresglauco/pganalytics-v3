# pgAnalytics v3 API Quick Reference

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All endpoints except `/auth/login` and `/collectors/register` require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

---

## Authentication Endpoints

### 1. User Login
```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password123"
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-02-20T11:00:00Z",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "full_name": "Administrator",
    "role": "admin",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-02-20T10:30:00Z"
  }
}
```

**Errors:**
- `401 Unauthorized` - Invalid credentials or user not found
- `400 Bad Request` - Missing or malformed request body

---

### 2. Refresh Token
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-02-20T11:15:00Z",
  "user": {
    "id": 1,
    "username": "admin",
    ...
  }
}
```

**Errors:**
- `401 Unauthorized` - Invalid or expired refresh token
- `400 Bad Request` - Missing refresh token

---

### 3. User Logout
```http
POST /auth/logout
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

---

## Collector Management

### 1. Register Collector
**No authentication required**
```http
POST /collectors/register
Content-Type: application/json

{
  "name": "main-db-collector",
  "hostname": "db-server-01.example.com",
  "address": "192.168.1.100",
  "version": "3.0.0"
}
```

**Response (200 OK):**
```json
{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "certificate": "-----BEGIN CERTIFICATE-----\nMIIDXTC...\n-----END CERTIFICATE-----",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQI...\n-----END PRIVATE KEY-----",
  "expires_at": "2025-02-20T10:30:00Z"
}
```

**Notes:**
- Save the certificate and private key for mTLS authentication
- The token is used in the Authorization header for subsequent requests
- Token expires in 30 minutes by default

---

### 2. List Collectors
```http
GET /collectors?page=1&page_size=20
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "main-db-collector",
      "hostname": "db-server-01.example.com",
      "address": "192.168.1.100",
      "version": "3.0.0",
      "status": "active",
      "last_seen": "2024-02-20T10:25:00Z",
      "certificate_expires_at": "2025-02-20T10:30:00Z",
      "metrics_count_total": 150000,
      "metrics_count_24h": 1440,
      "created_at": "2024-02-15T08:00:00Z",
      "updated_at": "2024-02-20T10:25:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

---

### 3. Get Collector Details
```http
GET /collectors/{collector_id}
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "main-db-collector",
  "hostname": "db-server-01.example.com",
  ...
}
```

---

### 4. Delete Collector
```http
DELETE /collectors/{collector_id}
Authorization: Bearer <token>
```

**Response (204 No Content)**

---

## Metrics Management

### 1. Push Metrics
```http
POST /metrics/push
Authorization: Bearer <collector_token>
Content-Type: application/json

{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "hostname": "db-server-01.example.com",
  "timestamp": "2024-02-20T10:30:00Z",
  "version": "3.0.0",
  "metrics_count": 1250,
  "metrics": [
    {
      "type": "pg_stats",
      "database": "postgres",
      "timestamp": "2024-02-20T10:30:00Z",
      "tables": [
        {
          "schema": "public",
          "name": "users",
          "rows": 1000,
          "size_bytes": 65536,
          "last_vacuum": "2024-02-20T10:00:00Z"
        }
      ]
    },
    {
      "type": "sysstat",
      "timestamp": "2024-02-20T10:30:00Z",
      "cpu": {
        "user": 10.5,
        "system": 3.2,
        "idle": 86.3
      },
      "memory": {
        "total_mb": 16384,
        "used_mb": 8192
      }
    }
  ]
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "metrics_inserted": 1250,
  "bytes_received": 458000,
  "processing_time_ms": 145,
  "next_config_version": 1,
  "next_check_in_seconds": 300
}
```

**Errors:**
- `401 Unauthorized` - Invalid or missing token, token mismatch
- `400 Bad Request` - Invalid metrics format

---

## Server Management

### 1. List Servers
```http
GET /servers?page=1&page_size=20
Authorization: Bearer <token>
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Production DB 01",
      "hostname": "db-server-01.example.com",
      "address": "192.168.1.100",
      "environment": "production",
      "collector_id": "550e8400-e29b-41d4-a716-446655440000",
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-02-20T10:25:00Z"
    }
  ],
  "total": 3,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

---

### 2. Get Server Details
```http
GET /servers/{server_id}
Authorization: Bearer <token>
```

---

### 3. Get Server Metrics
```http
GET /servers/{server_id}/metrics?metric_type=pg_stats&start_time=2024-02-20T00:00:00Z&end_time=2024-02-20T23:59:59Z
Authorization: Bearer <token>
```

**Query Parameters:**
- `metric_type` - Type of metric (pg_stats, sysstat, pg_log, disk_usage) - default: all
- `start_time` - ISO 8601 timestamp for start of range
- `end_time` - ISO 8601 timestamp for end of range

---

## Configuration Management

### 1. Get Collector Config
```http
GET /config/{collector_id}
Authorization: Bearer <collector_token>
```

**Response (200 OK):**
```json
{
  "version": 1,
  "config": {
    "collection_interval": 60,
    "metric_types": ["pg_stats", "sysstat", "pg_log"],
    "databases": ["postgres", "myapp_db"],
    "log_level": "info"
  }
}
```

---

### 2. Update Collector Config
```http
PUT /config/{collector_id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "version": 1,
  "config": {
    "collection_interval": 120,
    "metric_types": ["pg_stats", "sysstat"],
    "databases": ["postgres"]
  }
}
```

---

## Alert Management

### 1. List Alerts
```http
GET /alerts?page=1&page_size=20&severity=warning
Authorization: Bearer <token>
```

**Query Parameters:**
- `severity` - Filter by severity (info, warning, critical)
- `page` - Page number
- `page_size` - Results per page

---

### 2. Get Alert Details
```http
GET /alerts/{alert_id}
Authorization: Bearer <token>
```

---

### 3. Acknowledge Alert
```http
POST /alerts/{alert_id}/acknowledge
Authorization: Bearer <token>
Content-Type: application/json

{
  "note": "Investigating this issue"
}
```

---

## System Endpoints

### 1. Health Check
**No authentication required**
```http
GET /health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "version": "3.0.0-alpha",
  "timestamp": "2024-02-20T10:30:00Z",
  "database_ok": true,
  "timescale_ok": true,
  "uptime": 3600
}
```

**Response (503 Service Unavailable):**
```json
{
  "status": "degraded",
  "version": "3.0.0-alpha",
  "timestamp": "2024-02-20T10:30:00Z",
  "database_ok": false,
  "timescale_ok": true
}
```

---

### 2. Get Version
**No authentication required**
```http
GET /version
```

**Response (200 OK):**
```json
{
  "version": "3.0.0-alpha",
  "api": "1.0.0"
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error code/type",
  "message": "Human-readable error message",
  "code": 400,
  "details": "Optional additional details"
}
```

### Common HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | OK - Request successful |
| 204 | No Content - Successful deletion |
| 400 | Bad Request - Invalid input |
| 401 | Unauthorized - Missing or invalid authentication |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource does not exist |
| 409 | Conflict - Resource already exists |
| 500 | Internal Server Error |
| 503 | Service Unavailable - Database or health check failed |

---

## Token Information

### JWT Token Claims

**User Token (Access):**
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

**Collector Token:**
```json
{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "hostname": "db-server-01.example.com",
  "type": "access",
  "exp": 1708356000,
  "iat": 1708355000,
  "sub": "collector:550e8400-e29b-41d4-a716-446655440000"
}
```

### Token Expiration Times

- **User Access Token**: 15 minutes
- **User Refresh Token**: 24 hours
- **Collector Token**: 30 minutes

---

## Example Workflows

### Complete User Workflow

1. **Login**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"password"}'
   ```
   Save the returned `token` value

2. **List Collectors**
   ```bash
   curl -X GET http://localhost:8080/api/v1/collectors \
     -H "Authorization: Bearer $TOKEN"
   ```

3. **Get Server Metrics**
   ```bash
   curl -X GET "http://localhost:8080/api/v1/servers/1/metrics?metric_type=pg_stats" \
     -H "Authorization: Bearer $TOKEN"
   ```

4. **Refresh Token** (before 15 minutes expire)
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/refresh \
     -H "Content-Type: application/json" \
     -d '{"refresh_token":"$REFRESH_TOKEN"}'
   ```

### Complete Collector Workflow

1. **Register Collector** (no auth required)
   ```bash
   curl -X POST http://localhost:8080/api/v1/collectors/register \
     -H "Content-Type: application/json" \
     -d '{
       "name":"main-collector",
       "hostname":"db-01.example.com"
     }'
   ```
   Save the returned `token`, `certificate`, and `private_key`

2. **Periodically Push Metrics** (every 60 seconds)
   ```bash
   curl -X POST http://localhost:8080/api/v1/metrics/push \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $COLLECTOR_TOKEN" \
     -d '{
       "collector_id":"$COLLECTOR_ID",
       "hostname":"db-01.example.com",
       "timestamp":"2024-02-20T10:30:00Z",
       "metrics_count":1250,
       "metrics":[...]
     }'
   ```

3. **Pull Configuration** (every 5 minutes)
   ```bash
   curl -X GET http://localhost:8080/api/v1/config/$COLLECTOR_ID \
     -H "Authorization: Bearer $COLLECTOR_TOKEN"
   ```

