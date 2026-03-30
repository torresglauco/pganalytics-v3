# External Integrations

**Analysis Date:** 2026-03-30

## APIs & External Services

**ML Service Integration:**
- Service: Internal ML service (Python Flask)
- What it's used for: Query performance prediction, workload pattern detection, model training
- SDK/Client: `backend/internal/ml/client.go` - Custom HTTP client with circuit breaker
- Base URL: Environment-configured, default `http://ml-service:8081`
- Key Operations:
  - `TrainPerformanceModel` - Async model training for query prediction
  - `PredictQueryExecution` - Predict query execution time with confidence ranges
  - `ValidatePrediction` - Validate prediction accuracy against actual results
  - `DetectWorkloadPatterns` - Identify database workload patterns
  - `GetTrainingStatus` - Poll async training job status
- Health Check: `/api/health` endpoint with circuit breaker

**Authentication Providers:**
- OAuth 2.0/OIDC (Google, GitHub, Azure AD, Custom)
  - SDK: `golang.org/x/oauth2` v0.35.0
  - Implementation: `backend/internal/auth/oauth.go`
  - Supported: Google OAuth, GitHub OAuth, Azure AD, custom providers

- SAML 2.0 Single Sign-On
  - SDK: `crewjam/saml` v0.5.1
  - Implementation: `backend/internal/auth/saml.go`
  - Features: IdP metadata parsing, assertion validation, ACS endpoint

- LDAP/Active Directory
  - Implementation: `backend/internal/auth/ldap.go`
  - Use case: Enterprise directory integration

- Multi-Factor Authentication (MFA)
  - TOTP/HOTP: `pquerna/otp` v1.5.0 (`backend/internal/auth/mfa.go`)
  - SMS: Configurable via MFA service
  - Email: Code delivery via email channel

## Data Storage

**Databases:**

Primary Metadata Database:
- Type: PostgreSQL 16
- Connection: `DATABASE_URL` env var (default: `postgres://postgres:pganalytics@postgres:5432/pganalytics`)
- Client: `lib/pq` v1.10.9 (pure Go driver)
- Purpose: User accounts, collectors, configurations, audit logs
- Location: `backend/internal/storage/postgres.go`
- Encryption: Field-level encryption for sensitive data (passwords, secrets)

Metrics Time-Series Database:
- Type: PostgreSQL 16 (TimescaleDB extension-ready)
- Connection: `TIMESCALE_URL` env var (default: `postgres://postgres:pganalytics@timescale:5432/metrics`)
- Client: `lib/pq` v1.10.9
- Purpose: High-volume time-series metrics from collectors
- Location: `backend/internal/timescale/timescale.go`
- Features: Hypertables for efficient time-series storage
- Schema: Stored procedures for metric aggregation and retention

**Caching:**
- Type: Redis 7-alpine (optional, profile: optional)
- Connection: In-memory cache manager (optional)
- Location: `backend/internal/cache/manager.go`
- Purpose: Feature cache and prediction cache TTL
- Keys: Config cache (TTL: configurable), Prediction results cache
- Implementation: Custom memory cache with optional Redis backend

**File Storage:**
- Type: Local filesystem only
- Paths: `collector_data/` volume in Docker, `/var/lib/pganalytics` in collector container
- Purpose: Collector state, cached metrics between sync cycles
- Git Ignore: `collector_data/` excluded from version control

## Authentication & Identity

**Auth Provider:**
- Custom multi-provider implementation
- Location: `backend/internal/auth/` directory
- Components:
  - JWT Manager: `jwt.go` - Token generation, validation, refresh tokens
  - OAuth: `oauth.go` - OAuth 2.0/OIDC connector
  - SAML: `saml.go` - SAML 2.0 assertion processing
  - LDAP: `ldap.go` - Active Directory/LDAP binding
  - MFA: `mfa.go` - TOTP/HOTP/SMS/Email
  - Password: `password.go` - Bcrypt hashing, validation
  - Certificates: `cert_generator.go` - Client certificate generation for collectors

**Session Management:**
- Location: `backend/internal/session/session.go`
- Type: Stateless JWT tokens (can be backed by Redis)
- Token Types:
  - Access token: 15 minutes (user authentication)
  - Refresh token: 24 hours (token renewal)
  - Collector token: 30 minutes (collector heartbeat/metrics)

**Secret Encryption:**
- Algorithm: AES-256 GCM
- Key: `ENCRYPTION_KEY` env var (base64-encoded 32 bytes)
- Location: `backend/internal/crypto/`
- Usage: Encrypt PostgreSQL connection passwords, API keys, MFA secrets
- Implementation: Column-level encryption in database

## Monitoring & Observability

**Metrics:**
- Type: Prometheus
- Exporter: `backend` exposes `/metrics` endpoint
- SDK: `prometheus/client_golang` v1.23.2
- Metrics: HTTP request latency, database query time, cache hit rates, ML service latency
- Location: Prometheus scrape targets defined in `monitoring/prometheus.staging.yml`

**Dashboards:**
- Type: Grafana 11.0.0
- Purpose: Real-time database performance monitoring, alert visualization
- Provisioned Dashboards:
  - `dashboards/system-overview.json` - Overall cluster health
  - `dashboards/cache-performance.json` - Cache hit/miss rates
  - `dashboards/connection-tracking.json` - Active connections
  - `dashboards/lock-monitoring.json` - Lock contention
  - `dashboards/bloat-analysis.json` - Table/index bloat
  - `dashboards/extensions-config.json` - PostgreSQL extensions
  - `dashboards/schema-overview.json` - Schema statistics
- Admin Credentials: Set via `GF_SECURITY_ADMIN_USER/PASSWORD` in docker-compose

**Logs:**
- Backend: Structured logging with `go.uber.org/zap` v1.27.0
- ML Service: JSON logging with `python-json-logger` 2.0.7
- Log Level: Configurable via `LOG_LEVEL` env var (debug, info, warn, error)
- Aggregation: Available via Docker compose logs

**Alerting:**
- Type: Grafana Alert Rules
- Configuration: `monitoring/grafana-alerts.json`
- Notification Channels: `monitoring/notification-channels.json`

## CI/CD & Deployment

**Hosting:**
- Primary: Docker Compose (development and demo)
- Enterprise: Kubernetes (Helm charts in `helm/`)
- Cloud: AWS, GCP, Azure (via Dockerfile multi-stage builds)

**CI Pipeline:**
- No external CI/CD service detected in codebase
- Manual testing via Makefile targets
- Load testing: k6 framework (`tests/load/scenario.js`)

**Container Registry:**
- No external registry configured (local Docker build)

**Local Development Orchestration:**
- Tool: Mise 1.x task runner (`mise.toml`)
- Task: `mise run dev` - Starts all services
- Task: `mise run logs` - Follows all container logs
- Task: `mise run down` - Stops all services
- Task: `mise run reset` - Cleans volumes for fresh start

## Webhooks & Callbacks

**Incoming Webhooks:**
- Type: Notification channels (inbound alert delivery)
- Location: `backend/internal/notifications/channels.go`
- Supported Channels:
  - Slack: Incoming webhooks with rich message formatting
  - Email: SMTP delivery (configured but details not exposed)
  - PagerDuty: Incident integration
  - Generic webhook: Arbitrary HTTP POST endpoints

**Slack Integration:**
- Type: Incoming Webhook
- Configuration: `SlackConfig` with webhook_url, channel, username
- Implementation: `backend/internal/notifications/channels.go` - SlackChannel struct
- Message Format: Rich attachments with severity color coding (critical: red, high: orange, medium: blue, low: green)
- Fields Sent: Severity, database, metric name, alert rule, action links

**Outgoing Webhooks:**
- Collector Registration: Collector sends metrics to backend via HTTP
  - Endpoint: `/api/v1/collectors/[id]/metrics` (POST)
  - Auth: JWT token or client certificate
  - Frequency: `COLLECTION_INTERVAL` (default 60s)

- Backend to ML Service: Query performance predictions
  - Endpoint: `/api/train/performance-model`, `/api/predict/query-execution`
  - Circuit breaker: Fallback on ML service unavailability

## Collector Integration

**Collector Protocol:**
- Type: REST API with JSON over HTTP
- Client: C++ collector daemon
- Registration:
  - Endpoint: `POST /api/v1/collectors/register`
  - Auth: `REGISTRATION_SECRET` env var validation
  - Auto-register: Enabled when `AUTO_REGISTER=true`
  - Frequency: At startup and periodically

**Collector Heartbeat:**
- Endpoint: `POST /api/v1/collectors/[id]/heartbeat`
- Token Type: Collector JWT token (30 min expiry)
- Includes: System health, connection status, last metric timestamp
- Error Handling: Backend verifies collector exists before accepting metrics

**Collector Metrics Push:**
- Endpoint: `POST /api/v1/collectors/[id]/metrics`
- Payload: Batch of PostgreSQL metrics (query stats, locks, connections, bloat, cache)
- Format: JSON array with timestamp, metric type, value
- Storage: Written to TimescaleDB for time-series analysis

## Environment Configuration

**Required Environment Variables:**
- `DATABASE_URL` - PostgreSQL metadata database (required)
- `TIMESCALE_URL` - TimescaleDB for metrics (required)
- `JWT_SECRET` - Token signing key (required, min 32 bytes)
- `ENCRYPTION_KEY` - AES-256 key base64-encoded (required)
- `REGISTRATION_SECRET` - Collector registration validation (required for security)
- `TLS_CERT` - Path to server certificate
- `TLS_KEY` - Path to server private key
- `LOG_LEVEL` - Logging verbosity (debug, info, warn, error)
- `PORT` - HTTP server port
- `BACKEND_URL` - For collectors to reach backend (default: `http://backend:8080`)

**Optional Environment Variables:**
- `CACHE_ENABLED` - Enable memory caching (default: true)
- `CACHE_MAX_SIZE` - Max cache entries
- `FEATURE_CACHE_TTL` - Feature cache time-to-live
- `PREDICTION_CACHE_TTL` - Prediction cache time-to-live
- `ML_SERVICE_URL` - ML service base URL (default: `http://ml-service:8081`)
- `SETUP_ENDPOINT_ENABLED` - Allow initial setup via API
- `GF_SECURITY_ADMIN_PASSWORD` - Grafana admin password (demo only)

**Secrets Location:**
- `.env` file (local development only, git-ignored)
- Docker secrets for production deployments
- Environment variables for Kubernetes pods
- Vault/SecretManager for enterprise deployments

---

*Integration audit: 2026-03-30*
