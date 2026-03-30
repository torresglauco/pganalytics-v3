# Codebase Structure

**Analysis Date:** 2026-03-30

## Directory Layout

```
pganalytics-v3/
├── backend/                    # Go REST API server
│   ├── cmd/pganalytics-api/    # Entry point: main.go
│   ├── internal/               # Core packages (not exported)
│   │   ├── api/                # HTTP handlers and routes
│   │   ├── auth/               # Authentication services
│   │   ├── cache/              # Caching layer
│   │   ├── config/             # Configuration loading
│   │   ├── crypto/             # Encryption and secrets
│   │   ├── jobs/               # Background job schedulers
│   │   ├── ml/                 # ML service client
│   │   ├── metrics/            # Metrics utilities
│   │   ├── notifications/      # Notification channels
│   │   ├── storage/            # Database layer
│   │   ├── timescale/          # TimescaleDB client
│   │   ├── audit/              # Audit logging
│   │   ├── session/            # Session management
│   │   └── collector/          # Collector management
│   ├── migrations/             # Database schema migrations
│   ├── tests/                  # Integration and load tests
│   │   ├── integration/        # API integration tests
│   │   ├── unit/               # Unit tests
│   │   ├── load/               # Load testing
│   │   ├── benchmarks/         # Performance benchmarks
│   │   ├── security/           # Security tests
│   │   └── mocks/              # Test mocks
│   ├── pkg/                    # Exported utility packages
│   │   ├── models/             # Data models
│   │   ├── services/           # Reusable services
│   │   ├── handlers/           # Domain-specific handlers
│   │   └── errors/             # Error types
│   ├── bin/                    # Compiled binaries
│   ├── logs/                   # Application logs directory
│   ├── go.mod                  # Go module definition
│   └── go.sum                  # Go dependencies lock
│
├── frontend/                   # React + TypeScript UI
│   ├── src/                    # Source code
│   │   ├── components/         # Reusable React components
│   │   ├── pages/              # Page-level components
│   │   ├── stores/             # Zustand state management
│   │   ├── contexts/           # React contexts (Auth, Theme, Toast)
│   │   ├── hooks/              # Custom React hooks
│   │   ├── services/           # API client, real-time client
│   │   ├── api/                # API endpoint definitions
│   │   ├── types/              # TypeScript type definitions
│   │   ├── styles/             # TailwindCSS + custom styles
│   │   ├── utils/              # Utility functions
│   │   ├── test/               # Test utilities and setup
│   │   ├── App.tsx             # Root component
│   │   └── main.tsx            # Entry point (React bootstrap)
│   ├── e2e/                    # End-to-end tests (Playwright/similar)
│   ├── public/                 # Static assets (favicon, etc.)
│   ├── dist/                   # Built production bundle (generated)
│   ├── coverage/               # Test coverage reports (generated)
│   ├── package.json            # npm dependencies
│   ├── vite.config.ts          # Vite build configuration
│   └── tsconfig.json           # TypeScript configuration
│
├── collector/                  # C/C++ distributed metrics collector
│   ├── src/                    # Source implementation
│   │   ├── main.cpp            # Collector entry point
│   │   ├── collector.cpp       # Main collector loop
│   │   ├── auth.cpp            # mTLS and JWT authentication
│   │   ├── *_plugin.cpp        # Metrics collection plugins
│   │   │   ├── postgres_plugin.cpp       # Core PostgreSQL stats
│   │   │   ├── connection_plugin.cpp     # Connection metrics
│   │   │   ├── replication_plugin.cpp    # Replication lag, status
│   │   │   ├── query_stats_plugin.cpp    # Query performance
│   │   │   ├── schema_plugin.cpp         # Table/index metrics
│   │   │   ├── lock_plugin.cpp           # Lock contention
│   │   │   ├── cache_hit_plugin.cpp      # Cache effectiveness
│   │   │   ├── bloat_plugin.cpp          # Table/index bloat
│   │   │   ├── sysstat_plugin.cpp        # System CPU/memory/disk
│   │   │   ├── extension_plugin.cpp      # Extension info
│   │   │   └── log_plugin.cpp            # PostgreSQL logs
│   │   ├── sender.cpp          # HTTPS metric transmission
│   │   ├── connection_pool.cpp  # PostgreSQL connection pooling
│   │   ├── config_manager.cpp   # Configuration pull/apply
│   │   ├── metrics_buffer.cpp   # Local metric buffering
│   │   ├── metrics_serializer.cpp # Binary protocol serialization
│   │   ├── binary_protocol.cpp  # Wire protocol implementation
│   │   ├── thread_pool.cpp      # Async execution pool
│   │   └── {plugins}/           # Additional plugin modules
│   ├── include/                # Header files
│   │   ├── *.h                 # Plugin interfaces, utilities
│   │   ├── binary_protocol.h
│   │   ├── sender.h
│   │   ├── metrics_serializer.h
│   │   ├── *_plugin.h
│   │   └── auth.h
│   ├── sql/                    # SQL query templates
│   ├── config/                 # Sample configurations
│   ├── tests/                  # Unit and integration tests
│   ├── build/                  # CMake build output (generated)
│   ├── CMakeLists.txt          # CMake build configuration
│   └── systemd/                # systemd service definition
│
├── ml-service/                 # Python ML service (optional)
│   ├── api/                    # Flask API endpoints
│   ├── models/                 # ML models and training
│   ├── utils/                  # Utilities
│   ├── tests/                  # Tests
│   └── app.py                  # Flask entry point
│
├── grafana/                    # Grafana dashboards & provisioning
│   ├── dashboards/             # JSON dashboard definitions
│   ├── datasources/            # Data source configurations
│   └── provisioning/           # Grafana provisioning config
│
├── config/                     # Configuration templates
├── scripts/                    # Setup and utility scripts
│   └── docker/                 # Docker-related scripts
├── helm/                       # Kubernetes Helm charts
├── tls/                        # TLS certificate templates
├── tools/                      # Development tools
│   ├── load-test/              # Load testing tools
│   └── mock-backend/           # Mock API for testing
├── dashboards/                 # Additional dashboard definitions
├── docs/                       # Documentation
│   ├── api/                    # API documentation
│   ├── phases/                 # Phase-specific docs
│   └── superpowers/            # Feature documentation
├── docker-compose.yml          # Local development services
├── go.mod                      # Go module file
├── go.sum                      # Go dependencies lock
├── Makefile                    # Build targets
├── README.md                   # Project overview
├── CONTRIBUTING.md             # Developer guidelines
└── .planning/                  # GSD planning directory
    └── codebase/               # Architecture documentation
```

## Directory Purposes

**backend/cmd/pganalytics-api/:**
- Purpose: API server entry point
- Contains: main.go (initialization, configuration loading, service setup, graceful shutdown)
- Key files: `main.go` (221 lines)

**backend/internal/api/:**
- Purpose: HTTP request handling and route orchestration
- Contains: Handler functions for all endpoints, middleware, rate limiting, WebSocket support
- Key files:
  - `server.go` (516 lines) - Server struct, route registration
  - `handlers.go` (1,916 lines) - Core CRUD handlers (users, collectors, servers, metrics)
  - `handlers_auth.go` (752 lines) - Authentication endpoints
  - `handlers_ml.go` (668 lines) - ML service integration
  - `handlers_advanced.go` (880 lines) - Advanced analysis endpoints
  - `middleware.go` - Request/response middleware
  - `ratelimit.go`, `ratelimit_enhanced.go` - Rate limiting

**backend/internal/storage/:**
- Purpose: Database abstraction and persistence
- Contains: PostgreSQL operations, store classes, encryption, migrations
- Key files:
  - `postgres.go` (2,498 lines) - Main PostgreSQL client and queries
  - `metrics_store.go` (586 lines) - Time-series metrics operations
  - `user_store.go`, `collector_store.go`, `token_store.go` - Entity stores
  - `encrypted_fields.go` - Column-level encryption utilities
  - `migrations.go` - Schema management

**backend/internal/auth/:**
- Purpose: Authentication and authorization services
- Contains: JWT management, password hashing, certificate handling, MFA
- Key files: jwt_manager.go, password_manager.go, certificate_manager.go, mfa_manager.go

**backend/internal/jobs/:**
- Purpose: Background job execution
- Contains: Scheduled tasks for alert evaluation, anomaly detection, health checks
- Key files:
  - `alert_rule_engine.go` (828 lines) - Rule evaluation and alert generation
  - `anomaly_detector.go` (707 lines) - Anomaly detection logic
  - `health_check_scheduler.go` - Managed instance health checks

**backend/internal/ml/:**
- Purpose: ML service integration with resilience
- Contains: ML client, feature extraction, circuit breaker, caching
- Key files:
  - `client.go` (466 lines) - ML API client
  - `features.go` - Query feature extraction
  - `features_cache.go` - Feature caching wrapper
  - `circuit_breaker.go` - Fault tolerance

**backend/internal/notifications/:**
- Purpose: Alert notification delivery
- Contains: Channel implementations (Email, Slack, Webhooks), notification service
- Key files: `channels.go` (674 lines), `notification_service.go` (574 lines)

**backend/migrations/:**
- Purpose: Database schema versioning
- Contains: SQL migration files (executed in order)
- Key files:
  - `000_complete_schema.sql` - Main schema definition
  - `001_triggers.sql` - Database triggers
  - Individual migration files (mostly disabled, merged into 000)

**backend/tests/:**
- Purpose: Test coverage across all layers
- Contains: Integration tests, unit tests, load tests, benchmarks, security tests
- Key files:
  - `integration/handlers_test.go` - API handler tests
  - `load/load_test.go` - Performance and load testing
  - `security/sql_injection_test.go` - Security validation
  - `benchmarks/` - Performance benchmarks

**frontend/src/components/:**
- Purpose: Reusable UI components
- Contains: Page layout, forms, tables, charts, modals
- Key files: Dashboard.tsx, AlertRuleForm.tsx, ManagedInstancesTable.tsx, etc.

**frontend/src/pages/:**
- Purpose: Page-level components (route endpoints)
- Contains: AlertsPage, SettingsAdmin, CollectorsManagement, MetricsPage, LogsPage
- Key files:
  - `SettingsAdmin.tsx` (1,124 lines) - Admin panel
  - `CollectorsManagement.tsx` (583 lines) - Collector CRUD and status

**frontend/src/services/:**
- Purpose: API communication and real-time updates
- Contains: API client, real-time WebSocket client
- Key files:
  - `api.ts` (501 lines) - REST API client with axios
  - `realtime.ts` - WebSocket connection for live updates

**frontend/src/stores/:**
- Purpose: Global state management
- Contains: Zustand stores for auth, real-time updates, UI state
- Key files: authStore.ts, realtimeStore.ts

**collector/src/:**
- Purpose: Distributed metrics collection
- Contains: Plugin system, metric serialization, secure transmission
- Key files:
  - `main.cpp` - Entry point and initialization
  - `collector.cpp` - Main collection loop
  - Plugin files (1,000+ lines each) - Query execution
  - `sender.cpp` - HTTPS transmission with retry
  - `auth.cpp` - mTLS and JWT handling

**collector/include/:**
- Purpose: Header definitions for collector modules
- Contains: Plugin interfaces, data structures, utility declarations

**grafana/:**
- Purpose: Pre-built monitoring dashboards
- Contains: JSON dashboard definitions, data source configs
- Deployed alongside backend for visualization

## Key File Locations

**Entry Points:**
- `backend/cmd/pganalytics-api/main.go` - Backend API server startup
- `frontend/src/main.tsx` - React bootstrap and app initialization
- `collector/src/main.cpp` - Collector startup and configuration

**Configuration:**
- `backend/internal/config/config.go` - Configuration loader
- `collector/config/collector.toml.example` - Collector config template
- `docker-compose.yml` - Development environment setup

**Core Logic:**
- `backend/internal/api/server.go` - Route registration and middleware setup
- `backend/internal/storage/postgres.go` - All database operations (2,498 lines)
- `backend/internal/jobs/alert_rule_engine.go` - Alert detection logic
- `collector/src/collector.cpp` - Main collection loop

**Authentication & Security:**
- `backend/internal/auth/jwt_manager.go` - JWT token generation/validation
- `backend/internal/crypto/key_manager.go` - Encryption key management
- `collector/src/auth.cpp` - mTLS certificate handling

**Testing:**
- `backend/tests/integration/handlers_test.go` - API tests
- `backend/tests/load/load_test.go` - Load and performance tests
- `backend/tests/security/sql_injection_test.go` - SQL injection prevention
- `frontend/src/**/*.test.tsx` - React component tests

## Naming Conventions

**Files:**
- Go: `snake_case.go` (lowercase, underscores)
  - Example: `alert_rule_engine.go`, `circuit_breaker.go`
- TypeScript/React: `PascalCase.tsx` (components), `camelCase.ts` (utilities/services)
  - Example: `Dashboard.tsx`, `api.ts`, `authStore.ts`
- C/C++: `snake_case.cpp`, `snake_case.h`
  - Example: `postgres_plugin.cpp`, `binary_protocol.h`

**Directories:**
- Feature-based: `internal/auth/`, `internal/notifications/`
- Lowercase with hyphens: `registration-secrets/`, `managed-instances/`
- Descriptive: `cmd/`, `pkg/`, `migrations/`, `includes/`

**Functions/Methods:**
- Go: `camelCase`, handlers prefixed with `handle` or `Handle`
  - Example: `handleCreateUser`, `GetMetrics`, `NewServer`
- TypeScript: `camelCase` for functions, `PascalCase` for components/classes
  - Example: `useAuthStore`, `Dashboard`, `apiClient.login()`
- C/C++: `snake_case` for functions, `PascalCase` for classes
  - Example: `collect_metrics()`, `PostgresPlugin`

**Variables:**
- Go: `camelCase` (public), `lowercase` (private)
  - Example: `authService`, `postgresDB`, `jwtToken`
- TypeScript: `camelCase`, constants `UPPER_SNAKE_CASE`
  - Example: `isAuthenticated`, `API_BASE_URL`

**Types:**
- Go: `PascalCase` structs and interfaces
  - Example: `Server`, `AuthService`, `MetricsStore`
- TypeScript: `PascalCase` for types, interfaces, and classes
  - Example: `User`, `AlertRule`, `MetricsResponse`

## Where to Add New Code

**New Feature (e.g., new alert type):**
- Backend logic: `backend/internal/{services,api,storage}/` (add handler in handlers_*.go, service logic, database queries)
- Frontend UI: `frontend/src/pages/` (new page component), `frontend/src/components/` (reusable UI parts)
- Test: `backend/tests/integration/` (test the endpoint), `frontend/src/__tests__/` (test React components)

**New Metric Collection Plugin (in collector):**
- Implementation: `collector/src/{metric_name}_plugin.cpp`
- Header: `collector/include/{metric_name}_plugin.h`
- Registration: Add to main.cpp plugin initialization loop
- SQL templates: `collector/sql/` (if complex queries)

**New Component/Module (backend):**
- Go package: `backend/internal/{name}/` or `backend/pkg/{name}/`
- Interface definition: Typically in a `types.go` or `{name}.go` file
- Implementation: In separate files for clarity
- Tests: `backend/tests/{type}/{name}_test.go`

**Utilities:**
- Shared backend utilities: `backend/pkg/{category}/`
  - Example: `pkg/errors/`, `pkg/models/`, `pkg/services/`
- Frontend hooks: `frontend/src/hooks/{hookName}.ts`
- Frontend utilities: `frontend/src/utils/{utilityName}.ts`
- Shared models: `frontend/src/types/` for TypeScript interfaces

## Special Directories

**backend/migrations/:**
- Purpose: Database schema versioning and evolution
- Generated: No (manually authored)
- Committed: Yes
- Usage: Executed in order during initialization; 000_complete_schema.sql contains full schema

**frontend/dist/:**
- Purpose: Production-ready bundle
- Generated: Yes (via `npm run build`)
- Committed: No (excluded in .gitignore)
- Usage: Served by web server in production

**collector/build/:**
- Purpose: CMake build artifacts
- Generated: Yes (via `mkdir build && cd build && cmake .. && make`)
- Committed: No (excluded in .gitignore)
- Usage: Contains compiled collector binary and object files

**backend/logs/:**
- Purpose: Application runtime logs
- Generated: Yes (during execution)
- Committed: No (excluded in .gitignore)
- Usage: Persistent storage of structured logs

**frontend/coverage/:**
- Purpose: Test coverage reports
- Generated: Yes (via `npm run test:coverage`)
- Committed: No (excluded in .gitignore)
- Usage: Coverage analysis and metrics

**tls/:**
- Purpose: TLS certificate templates and generation scripts
- Generated: Partially (certs generated during setup)
- Committed: Templates yes, generated certs no
- Usage: mTLS between collector and backend, HTTPS for API

---

*Structure analysis: 2026-03-30*
