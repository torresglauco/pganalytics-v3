# Architecture Diagrams - pgAnalytics v3 Backend

## Overall System Architecture

```
┌────────────────────────────────────────────────────────────────────────┐
│                         CLIENT APPLICATIONS                            │
├────────────────────────────────────────────────────────────────────────┤
│                   ┌─────────────┐         ┌──────────────┐             │
│                   │  Web Browser│         │  Collector   │             │
│                   │  (Grafana)  │         │  (C/C++)     │             │
│                   └──────┬──────┘         └──────┬───────┘             │
└────────────────────────┬───────────────────────┬───────────────────────┘
                         │                       │
                    HTTPS │                       │ HTTPS+mTLS
                         │                       │
┌────────────────────────▼───────────────────────▼───────────────────────┐
│                      API SERVER (Go + Gin)                             │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  ┌─ MIDDLEWARE LAYER ──────────────────────────────────────────────┐ │
│  │  • LoggingMiddleware                                           │ │
│  │  • CORSMiddleware                                              │ │
│  │  • AuthMiddleware (JWT validation)                             │ │
│  │  • CollectorAuthMiddleware (collector JWT validation)          │ │
│  │  • MTLSMiddleware (certificate validation)                     │ │
│  │  • ErrorResponseMiddleware                                     │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                        │
│  ┌─ HTTP HANDLERS LAYER ────────────────────────────────────────────┐ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐            │ │
│  │  │  Auth       │  │  Collectors │  │  Metrics     │            │ │
│  │  │  • login    │  │  • register │  │  • push      │            │ │
│  │  │  • refresh  │  │  • list     │  │  • push      │            │ │
│  │  │  • logout   │  │  • get      │  └──────────────┘            │ │
│  │  │             │  │  • delete   │                              │ │
│  │  └─────────────┘  └─────────────┘  ┌──────────────┐            │ │
│  │                                     │  Servers     │            │ │
│  │  ┌─────────────┐  ┌─────────────┐  │  • list      │            │ │
│  │  │  Config     │  │  Alerts     │  │  • get       │            │ │
│  │  │  • get      │  │  • list     │  │  • metrics   │            │ │
│  │  │  • update   │  │  • get      │  └──────────────┘            │ │
│  │  └─────────────┘  │  • ack      │                              │ │
│  │                   └─────────────┘                              │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                        │
│  ┌─ SERVICE LAYER ──────────────────────────────────────────────────┐ │
│  │  ┌─────────────────────────────────────────────────────────┐   │ │
│  │  │               AuthService                              │   │ │
│  │  │  • LoginUser(username, password)                       │   │ │
│  │  │  • RefreshUserToken(refreshToken)                      │   │ │
│  │  │  • RegisterCollector(name, hostname)                   │   │ │
│  │  │  • ValidateUserToken(token)                            │   │ │
│  │  │  • ValidateCollectorToken(token)                       │   │ │
│  │  └─────────────────────────────────────────────────────────┘   │ │
│  │                                                                   │ │
│  │  ┌──────────────────┐  ┌──────────────────┐                    │ │
│  │  │  JWTManager      │  │ PasswordManager  │                    │ │
│  │  │  • Generate token│  │ • HashPassword   │                    │ │
│  │  │  • Validate token│  │ • VerifyPassword │                    │ │
│  │  └──────────────────┘  └──────────────────┘                    │ │
│  │                                                                   │ │
│  │  ┌──────────────────────────────────────┐                      │ │
│  │  │    CertificateManager                │                      │ │
│  │  │    • GenerateCollectorCertificate()  │                      │ │
│  │  │    • ValidateCertificate()           │                      │ │
│  │  └──────────────────────────────────────┘                      │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                        │
│  ┌─ DATA ACCESS LAYER ──────────────────────────────────────────────┐ │
│  │  ┌──────────────────┐  ┌──────────────────┐                    │ │
│  │  │  PostgresDB      │  │  TimescaleDB     │                    │ │
│  │  │  • Users         │  │  • Metrics       │                    │ │
│  │  │  • Collectors    │  │  • Aggregates    │                    │ │
│  │  │  • Servers       │  │  • Time-series   │                    │ │
│  │  │  • Tokens        │  └──────────────────┘                    │ │
│  │  │  • Audit logs    │                                          │ │
│  │  └──────────────────┘                                          │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                        │
└────────────────────────────┬───────────────────┬──────────────────────┘
                             │                   │
                    PostgreSQL│              TimescaleDB
                             │                   │
┌────────────────────────────▼───────────────────▼──────────────────────┐
│                         DATABASES                                      │
├────────────────────────────────────────────────────────────────────────┤
│                                                                        │
│  ┌─ PostgreSQL ──────────────────┐  ┌─ TimescaleDB ───────────────┐  │
│  │ • users                        │  │ • metrics_pg_stats          │  │
│  │ • collectors                   │  │ • metrics_sysstat           │  │
│  │ • collector_config             │  │ • metrics_disk_usage        │  │
│  │ • api_tokens                   │  │ • metrics_pg_log            │  │
│  │ • servers                      │  │ • metrics_replication       │  │
│  │ • databases                    │  │                             │  │
│  │ • alerts                       │  │ (All configured as          │  │
│  │ • alert_rules                  │  │  hypertables with           │  │
│  │ • audit_logs                   │  │  retention policies)        │  │
│  │ • secrets                      │  └─────────────────────────────┘  │
│  └────────────────────────────────┘                                   │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

---

## Authentication Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ POST /auth/login
       │ {username, password}
       ▼
┌──────────────────┐
│ AuthMiddleware   │
│ (not required)   │
└──────┬───────────┘
       │
       ▼
┌──────────────────┐
│ handleLogin()    │
└──────┬───────────┘
       │
       ▼
┌──────────────────────────────┐
│ authService.LoginUser()      │
│ ┌────────────────────────┐   │
│ │ 1. Get user from DB    │   │
│ │ 2. Verify password     │   │
│ │ 3. Check if active     │   │
│ └────────────────────────┘   │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│ jwtManager.GenerateUserToken()
│ ┌────────────────────────┐   │
│ │ 1. Create claims       │   │
│ │ 2. Sign with secret    │   │
│ │ 3. Return JWT string   │   │
│ └────────────────────────┘   │
└──────┬───────────────────────┘
       │
       ▼
┌──────────────────┐
│  Response 200    │
│  {              │
│   token,        │
│   refresh_token,│
│   expires_at,   │
│   user          │
│  }              │
└──────────────────┘
```

---

## Protected Endpoint Access Flow

```
┌─────────────┐
│   Client    │
│ with Token  │
└──────┬──────┘
       │
       │ GET /api/v1/collectors
       │ Authorization: Bearer <token>
       │
       ▼
┌──────────────────────────────────┐
│ AuthMiddleware                    │
│ ┌──────────────────────────────┐ │
│ │ 1. Get header                │ │
│ │ 2. Extract token             │ │
│ │ 3. Validate signature        │ │
│ │ 4. Check expiration          │ │
│ │ 5. Store claims in context   │ │
│ └──────────────────────────────┘ │
└──────┬───────────────────────────┘
       │
       │ Token valid?
       ├─ No ──→ Return 401 Unauthorized
       │
       │ Yes
       ▼
┌──────────────────────┐
│ handleListCollectors │
│ ┌────────────────┐   │
│ │ 1. Parse query │   │
│ │ 2. Call service│   │
│ │ 3. Return JSON │   │
│ └────────────────┘   │
└──────┬───────────────┘
       │
       ▼
┌────────────────────┐
│ Response 200 {data}│
└────────────────────┘
```

---

## Collector Registration Flow

```
┌──────────────┐
│  Collector   │
│  (C/C++)     │
└──────┬───────┘
       │
       │ POST /api/v1/collectors/register
       │ (No Auth Required)
       │ {name, hostname}
       │
       ▼
┌──────────────────────┐
│ handleCollectorReg() │
└──────┬───────────────┘
       │
       ▼
┌───────────────────────────────────┐
│ authService.RegisterCollector()   │
│ ┌─────────────────────────────┐   │
│ │ 1. Validate input           │   │
│ │ 2. Create Collector record  │   │
│ │ 3. Save to PostgreSQL       │   │
│ └─────────────────────────────┘   │
└──────┬────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ certManager.GenerateCollector()      │
│ ┌──────────────────────────────────┐ │
│ │ 1. Generate RSA key pair (2048)  │ │
│ │ 2. Create cert template          │ │
│ │ 3. Self-sign certificate         │ │
│ │ 4. Encode to PEM format          │ │
│ │ 5. Compute thumbprint            │ │
│ └──────────────────────────────────┘ │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────┐
│ jwtManager.GenerateCollectorToken()  │
│ ┌──────────────────────────────────┐ │
│ │ 1. Create claims with ID         │ │
│ │ 2. Set 30 min expiration         │ │
│ │ 3. Sign with secret              │ │
│ │ 4. Return JWT string             │ │
│ └──────────────────────────────────┘ │
└──────┬───────────────────────────────┘
       │
       ▼
┌──────────────────────────────┐
│      Response 200 {          │
│   collector_id: UUID,        │
│   token: JWT,                │
│   certificate: PEM,          │
│   private_key: PEM,          │
│   expires_at: timestamp      │
│      }                        │
└──────────────────────────────┘
```

---

## Metrics Push Flow

```
┌──────────────────┐
│    Collector     │
│   (Auth Token)   │
└──────┬───────────┘
       │
       │ POST /api/v1/metrics/push
       │ Authorization: Bearer <collector_token>
       │ {metrics: [...]}
       │
       ▼
┌────────────────────────────────────┐
│ CollectorAuthMiddleware            │
│ ┌────────────────────────────────┐ │
│ │ 1. Extract token from header   │ │
│ │ 2. Validate signature          │ │
│ │ 3. Extract collector ID        │ │
│ │ 4. Verify collector is active  │ │
│ │ 5. Store claims in context     │ │
│ └────────────────────────────────┘ │
└──────┬───────────────────────────┘
       │
       │ Valid?
       ├─ No ──→ Return 401 Unauthorized
       │
       │ Yes
       ▼
┌────────────────────────────┐
│ handleMetricsPush()        │
│ ┌──────────────────────┐   │
│ │ 1. Parse JSON        │   │
│ │ 2. Validate ID match │   │
│ │ 3. Store metrics (*)  │   │
│ │ 4. Update counts     │   │
│ │ 5. Return status     │   │
│ └──────────────────────┘   │
└──────┬───────────────────┘
       │
       ▼ (*) Future: Insert into TimescaleDB
       │
       ▼
┌─────────────────────────────────┐
│   Response 200 {                │
│    status: "success",           │
│    metrics_inserted: N,         │
│    processing_time_ms: ms,      │
│    next_config_version: v,      │
│    next_check_in_seconds: s     │
│   }                             │
└─────────────────────────────────┘
```

---

## Service Dependencies

```
┌─────────────────────────────────────────────┐
│              AuthService                    │
│  ┌───────────────────────────────────────┐  │
│  │         Depends On:                   │  │
│  │  • JWTManager (token ops)             │  │
│  │  • PasswordManager (pwd hashing)      │  │
│  │  • CertificateManager (certs)         │  │
│  │  • UserStore (interface)              │  │
│  │  • CollectorStore (interface)         │  │
│  │  • TokenStore (interface)             │  │
│  └───────────────────────────────────────┘  │
└──────────┬──────────────┬──────────────┬─────┘
           │              │              │
    ┌──────▼──┐  ┌──────▼──┐  ┌──────▼──┐
    │ JWT     │  │Password │  │ Cert    │
    │Manager  │  │Manager  │  │Manager  │
    └─────────┘  └─────────┘  └─────────┘
           │              │              │
    ┌──────▼──────────────▼──────────────▼──┐
    │         PostgreSQL Database            │
    │  (via interface implementations)       │
    └────────────────────────────────────────┘
```

---

## Testing Architecture

```
┌─────────────────────────────────────────┐
│         JWT Tests (18+ cases)           │
│  • Token generation                    │
│  • Token validation                    │
│  • Token refresh                       │
│  • Error scenarios                     │
└─────────────────────────────────────────┘
           │
┌──────────▼───────────────────────────────┐
│    Auth Service Tests (7+ cases)         │
│  • Login flow                           │
│  • Token refresh                        │
│  • Collector registration               │
│  • Error handling                       │
└──────────┬───────────────────────────────┘
           │
┌──────────▼───────────────────────────────────────┐
│      Handler Integration Tests (7+ cases)        │
│  • HTTP request/response validation             │
│  • JSON parsing                                 │
│  • Status code verification                     │
│  • Mock data store integration                  │
└─────────────────────────────────────────────────┘
           │
    ┌──────▼──────┐
    │ All Tests   │
    │ Pass ✅     │
    └─────────────┘
```

---

## Deployment Architecture

```
┌──────────────────────────────────────────────────┐
│              Docker Container                    │
├──────────────────────────────────────────────────┤
│  pganalytics-backend:3.0.0-alpha                │
│  ┌────────────────────────────────────────────┐ │
│  │  Go Application                            │ │
│  │  ├─ Gin HTTP Server                        │ │
│  │  ├─ JWT Manager                            │ │
│  │  ├─ Auth Service                           │ │
│  │  └─ Data Access Layers                     │ │
│  └────────────────────────────────────────────┘ │
│  Listening on port :8080                       │
└──────────────────────────────────────────────────┘
           │                      │
           │ PostgreSQL driver    │ TimescaleDB driver
           │                      │
  ┌────────▼──────┐      ┌────────▼──────┐
  │  PostgreSQL   │      │ TimescaleDB   │
  │  Container    │      │ Container     │
  │  Port 5432    │      │ Port 5433     │
  └───────────────┘      └───────────────┘
```

---

## Summary

This architecture provides:
- ✅ **Security**: JWT + mTLS ready, bcrypt passwords
- ✅ **Scalability**: Connection pooling, proper data access patterns
- ✅ **Testability**: Interface-based design, dependency injection
- ✅ **Maintainability**: Clean layers, clear separation of concerns
- ✅ **Extensibility**: Easy to add new handlers and services

