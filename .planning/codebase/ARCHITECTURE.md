# Architecture

**Analysis Date:** 2026-03-30

## Pattern Overview

**Overall:** Multi-layered distributed system with plugin-based collectors and REST API backend

**Key Characteristics:**
- **Distributed Architecture**: Lightweight C/C++ collectors on monitored servers push metrics to centralized Go backend
- **Plugin-Based Collectors**: Modular metrics collection via pluggable PostgreSQL query execution and system stat plugins
- **Layered Backend**: Clear separation between API handlers, services, and data stores
- **Real-Time Communication**: WebSocket support for live updates to React frontend
- **Time-Series Optimized**: Dual database setup (PostgreSQL for metadata, TimescaleDB for metrics)
- **ML Integration**: Optional circuit-breaker protected ML service for predictions and anomaly detection

## Layers

**API Layer (`internal/api/`):**
- Purpose: HTTP request handling, route management, middleware orchestration
- Location: `/backend/internal/api/`
- Contains: Request handlers (handlers.go, handlers_auth.go, handlers_ml.go, etc.), middleware, rate limiting, WebSocket
- Depends on: Auth service, storage layer, ML client, cache manager, notification services
- Used by: Frontend (React), collectors (metric push), external webhooks

**Service Layer (`internal/` packages):**
- Purpose: Core business logic, orchestration, and specialized concerns
- Location: `/backend/internal/{auth,notifications,jobs,ml,cache,timescale,crypto,audit}/`
- Contains:
  - Authentication & authorization (JWT, mTLS, password, certificates)
  - Job schedulers (alert rules, anomaly detection, health checks)
  - ML client and feature extraction with caching
  - Notification channels (email, Slack, webhooks)
  - Encryption and secret management
  - Audit logging
- Depends on: Storage layer, external services (ML, mail, Slack)
- Used by: API layer

**Storage Layer (`internal/storage/`):**
- Purpose: Database abstraction and persistence
- Location: `/backend/internal/storage/`
- Contains:
  - PostgreSQL connection and query execution (postgres.go - 2,498 lines)
  - Data stores: UserStore, CollectorStore, TokenStore, MetricsStore, etc.
  - Encrypted field handling (column-level encryption)
  - Migrations management
- Depends on: PostgreSQL driver (lib/pq), encryption (crypto package)
- Used by: Service layer

**Data Layer:**
- Purpose: Raw data persistence and time-series metrics
- Location: PostgreSQL (metadata), TimescaleDB (metrics storage)
- Schema: Tables for users, collectors, metrics, alerts, configurations, audit logs
- Migrations: `/backend/migrations/` (000_complete_schema.sql as main schema)

**Collector Layer (`collector/src/`):**
- Purpose: Lightweight distributed metrics collection from PostgreSQL and system
- Location: `/collector/src/`
- Contains: Main collector loop, plugin system (postgres_plugin, connection_plugin, replication_plugin, schema_plugin, lock_plugin, cache_hit_plugin, bloat_plugin, sysstat_plugin), serialization, sender, authentication
- Sends to: Backend metrics push endpoint (/api/v1/metrics/push)
- Runs on: Each monitored database server

**Frontend Layer (`frontend/src/`):**
- Purpose: User interface and visualization
- Location: `/frontend/src/`
- Contains: React components, pages, stores (Zustand), hooks, API client, real-time client
- Depends on: Backend API, WebSocket for real-time updates
- Served from: Vite development server or dist build

## Data Flow

**Metrics Collection Flow:**
1. Collector plugins execute SQL queries against PostgreSQL
2. Serializer converts metrics to binary protocol format
3. Sender transmits via HTTPS POST to `/api/v1/metrics/push`
4. API handler (handleMetricsPush) validates mTLS certificate and JWT token
5. Metrics inserted into TimescaleDB time-series tables
6. Frontend queries `/api/v1/metrics` or subscribes via WebSocket for live updates

**Alert Detection Flow:**
1. AlertRuleEngine (internal/jobs/alert_rule_engine.go) runs periodically
2. Queries metrics from TimescaleDB based on rule conditions
3. Evaluates conditions and detects state changes
4. Creates Alert records in PostgreSQL
5. NotificationService sends notifications via configured channels
6. Frontend receives alert via WebSocket or polls /api/v1/alerts
7. User can silence, acknowledge, or escalate alert

**Authentication Flow:**
1. User login: POST /api/v1/auth/login → JWT token issued
2. Frontend stores token, includes in Authorization header
3. AuthMiddleware validates JWT signature and expiration
4. Collector registration: POST /api/v1/collectors/register (requires registration secret)
5. Collector receives JWT token, uses for metrics push and config pull
6. mTLS certificates optional for collector→backend connection security

**ML Integration Flow:**
1. FeatureExtractor queries historical metrics from TimescaleDB
2. Extracts query features (execution count, average duration, standard deviation)
3. Sends TrainingRequest to ML service (if enabled)
4. ML service returns predictions and optimization suggestions
5. Results cached (if cache enabled) to avoid repeated requests
6. Frontend displays predictions in performance analysis dashboard

**Configuration Management Flow:**
1. Collector pulls config: GET /api/v1/config/{collector_id}
2. Backend returns TOML-formatted configuration
3. Collector applies settings: interval, database list, plugin enablement
4. Admin updates config: PUT /api/v1/config/{collector_id}
5. Config cached in backend; collectors fetch on next interval

**State Management:**
- Frontend: Zustand stores (authStore, realtimeStore) for global state
- Backend: In-memory caches for frequently accessed data (configuration, feature cache)
- Databases: PostgreSQL for transactional consistency, TimescaleDB for time-series

## Key Abstractions

**Plugin System:**
- Purpose: Modular metric collection without monolithic code
- Examples: `collector/include/postgres_plugin.h`, `collector/include/replication_plugin.h`
- Pattern: Each plugin implements a common interface, returns structured metrics
- Used by: Collector main loop to invoke specialized queries

**Store Pattern:**
- Purpose: Encapsulate database operations per entity
- Examples: `storage/user_store.go`, `storage/collector_store.go`, `storage/metrics_store.go`
- Pattern: Each store wraps PostgresDB connection, provides typed methods (GetUser, CreateUser, etc.)

**Middleware Chain:**
- Purpose: Cross-cutting concerns (auth, rate limiting, CORS, audit logging)
- Examples: AuthMiddleware, CollectorAuthMiddleware, MTLSMiddleware, RateLimitMiddleware
- Pattern: Gin middleware functions that wrap handlers, modify context, or reject requests

**Feature Extractor (with Optional Caching):**
- Purpose: Extract ML features from raw metrics
- Examples: `ml/features.go` (base), `ml/features_cache.go` (cached wrapper)
- Pattern: Interface-based design allows swapping cached vs. non-cached implementation
- Used by: ML client to prepare training data

**Circuit Breaker:**
- Purpose: Protect against cascading failures in ML service calls
- Examples: `ml/circuit_breaker.go`
- Pattern: State machine (Closed → Open → Half-Open) for fault tolerance
- Used by: ML client for resilient external service calls

**Handler Pattern:**
- Purpose: Process specific domain concerns (alerts, silences, escalations)
- Examples: `pkg/handlers/conditions.go`, `pkg/handlers/silences.go`, `pkg/handlers/escalations.go`
- Pattern: Specialized handler types for each concern, used by API endpoints

## Entry Points

**Backend API Server:**
- Location: `/backend/cmd/pganalytics-api/main.go`
- Triggers: Docker container startup or direct binary execution
- Responsibilities:
  - Load configuration from environment variables
  - Initialize databases (PostgreSQL, TimescaleDB)
  - Set up authentication, caching, secret management
  - Register all API routes
  - Start health check scheduler
  - Accept interrupt signals for graceful shutdown

**Collector:**
- Location: `/collector/src/main.cpp`
- Triggers: Systemd service or container startup
- Responsibilities:
  - Load TOML configuration
  - Establish mTLS connection to backend
  - Authenticate with JWT token
  - Execute plugin queries at configured intervals
  - Serialize and send metrics
  - Handle push failures with local buffer retry

**Frontend:**
- Location: `/frontend/src/main.tsx`
- Triggers: Vite dev server or served dist/ files in production
- Responsibilities:
  - Check authentication status
  - Initialize API client and real-time connection
  - Render routes based on authentication
  - Manage global state (auth, real-time updates)

## Error Handling

**Strategy:** Multi-layered with fallbacks

**Patterns:**
- **API Layer**: Gin error handlers return JSON error responses with status codes
- **Database Errors**: Storage layer returns wrapped errors with context
- **Service Layer**: Custom error types (pkg/errors/errors.go) for semantic error handling
- **External Services**: Circuit breaker protects against ML service failures
- **Collector**: Local metrics buffer on network failures, retries on next interval
- **Frontend**: Toast notifications for user-facing errors, fallback to default states

## Cross-Cutting Concerns

**Logging:**
- Backend: Structured logging with `go.uber.org/zap` in JSON format
- Production mode: JSON output for log aggregation
- Development mode: Human-readable output with colors

**Validation:**
- API requests: Struct tag validation via Gin binding
- Database constraints: PostgreSQL constraints + application-level checks
- Configuration: Config.Validate() method validates required fields

**Authentication:**
- JWT tokens for users and collectors (15 min expiry for users, 30 min for collectors)
- Refresh tokens for token rotation (24 hour expiry)
- mTLS certificates for secure collector←→backend communication
- Password hashing with bcrypt for user credentials

**Rate Limiting:**
- Per-user limit: 100 req/min (standard)
- Per-collector limit: 1000 req/min (high volume metric pushes)
- Implemented in `api/ratelimit.go` and `api/ratelimit_enhanced.go`

**Encryption:**
- Column-level encryption for sensitive fields (passwords stored in encrypted form)
- Secret key from environment: ENCRYPTION_KEY (base64-encoded AES-256)
- TLS 1.3 for all network connections

---

*Architecture analysis: 2026-03-30*
