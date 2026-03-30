# Technology Stack

**Analysis Date:** 2026-03-30

## Languages

**Primary:**
- Go 1.24.0 - Backend API, services, and CLI tools (located in `backend/`)
- TypeScript 5.3.3 - Frontend React UI with strict mode enabled (located in `frontend/`)
- C++ 17 - Collector daemon for PostgreSQL metrics (located in `collector/`)
- Python 3.x - ML service for predictive analytics (located in `ml-service/`)

**Secondary:**
- Bash/Shell - DevOps scripts and tooling (located in `scripts/`)
- SQL - Database schema and migrations (located in `backend/migrations/`)
- JSON - Configuration files and data schemas
- TOML - Development environment configuration

## Runtime

**Environment:**
- Go 1.24.0 (backend, collector infrastructure)
- Node.js 18 (frontend build, dev server via vite)
- Docker containers (production deployment)
- CPython 3.x (ML service)
- CMake 3.22+ (C++ collector build system)

**Package Managers:**
- Go Modules for Go dependencies (`go.mod`, `go.sum` in root)
- npm for JavaScript/TypeScript (frontend - `package-lock.json`)
- pip for Python ML service (`requirements.txt`)
- vcpkg for C++ dependencies (collector - `vcpkg.json`)

**Lockfiles:**
- `go.sum` - Go dependency lock
- `package-lock.json` - Node/npm lock
- Python `requirements.txt` - pinned versions (not lockfile)

## Frameworks

**Core Backend:**
- Gin v1.10.0 - HTTP web framework, REST API routing, middleware
- JWT (golang-jwt/jwt v5.2.0) - Token-based authentication
- PostgreSQL (lib/pq v1.10.9) - Primary database driver
- Prometheus client v1.23.2 - Metrics and observability

**Frontend:**
- React 18.2.0 - UI library and component framework
- Vite 5.0.8 - Build tool and dev server
- React Router 6.22.0 - Client-side routing
- TailwindCSS 3.4.1 - Utility-first CSS framework
- Framer Motion 12.34.5 - Animation library
- Zustand 5.0.11 - State management (lightweight alternative to Redux)
- React Hook Form 7.50.0 - Form state management
- TanStack React Table 8.21.3 - Data table/grid component
- Recharts 3.7.0 - Chart/visualization library

**ML Service:**
- Flask 2.3.2 - Web framework for ML API
- Celery 5.3.1 - Async task queue
- SQLAlchemy 2.0.19 - ORM for database models
- scikit-learn 1.2.2 - Machine learning algorithms
- pandas 2.0.3 - Data manipulation and analysis
- numpy 1.24.3 - Numerical computation

**Testing & Quality:**
- Vitest 1.0.0 (frontend unit tests)
- Go's built-in testing (backend)
- pytest 7.4.0 (ML service)
- Testing Library (React - @testing-library/react)
- Testify (Go assertions - v1.11.1)

**Build & Development:**
- Mise 1.x - Task runner and environment manager (replaces Make for dev tasks)
- Docker Compose 3.8+ - Multi-container orchestration
- CMake 3.22+ - C++ build system for collector
- Vite 5.0.8 - JavaScript bundler and dev server

## Key Dependencies

**Critical - Authentication & Security:**
- crewjam/saml v0.5.1 - SAML 2.0 single sign-on support
- golang-jwt/jwt v5.2.0 - JWT token generation/validation
- pquerna/otp v1.5.0 - One-time password/TOTP/HOTP for MFA
- golang.org/x/oauth2 v0.35.0 - OAuth 2.0/OIDC provider support
- golang.org/x/crypto v0.41.0 - Cryptographic primitives (AES-256, hashing)

**Critical - Database & Storage:**
- lib/pq v1.10.9 - PostgreSQL driver (native pure Go implementation)
- PostgreSQL 16 container - Primary metadata and configuration database
- PostgreSQL 16 (TimescaleDB) - Time-series metrics database
- Redis 7-alpine - Optional caching (profile: optional in docker-compose)

**Monitoring & Observability:**
- prometheus/client_golang v1.23.2 - Prometheus metrics exposition
- Grafana 11.0.0 - Dashboards, alerts, visualization
- go.uber.org/zap v1.27.0 - Structured logging (backend)
- python-json-logger 2.0.7 - JSON structured logging (ML service)

**Infrastructure:**
- OpenSSL 3.0 - TLS/SSL encryption (collector requirement)
- libcurl - HTTP client for collector
- zlib - Compression library
- nlohmann_json - Header-only JSON library (C++)

## Configuration

**Environment Variables:**
- `DATABASE_URL` - PostgreSQL connection string for metadata
- `TIMESCALE_URL` - PostgreSQL/TimescaleDB connection for metrics
- `JWT_SECRET` - Secret key for token signing
- `JWT_EXPIRATION` - Token expiration in seconds
- `ENCRYPTION_KEY` - AES-256 key for field-level encryption (base64 encoded)
- `TLS_CERT`, `TLS_KEY` - TLS certificate and key paths
- `LOG_LEVEL` - Logging verbosity (debug, info, warn, error)
- `PORT` - HTTP server port (default: 8080)
- `REGISTRATION_SECRET` - Collector auto-registration secret
- `SETUP_ENDPOINT_ENABLED` - Whether initial setup endpoint is available
- `BACKEND_TLS_VERIFY` - Whether to verify TLS in collector (false for dev)
- `COLLECTION_INTERVAL` - Collector metric poll interval in seconds

**Build Configuration:**
- `CMakeLists.txt` (collector) - C++ compilation flags, dependencies
- `tsconfig.json` (frontend) - TypeScript strict mode, target ES2020
- `vite.config.ts` - Frontend build, React plugin, path resolution
- `.env.example` - Template for environment variables

**Deployment Configuration:**
- `docker-compose.yml` - Multi-service orchestration with 7 containers
- `Dockerfile` files in `backend/`, `collector/`, `frontend/`, `ml-service/`
- Helm charts in `helm/` for Kubernetes deployment
- `postgresql.conf` - PostgreSQL runtime configuration

## Platform Requirements

**Development:**
- macOS/Linux with Docker and Docker Compose installed
- Go 1.24+
- Node.js 18+ with npm
- C++ compiler (gcc/clang) with C++17 support
- CMake 3.22+
- OpenSSL 3.0 development headers
- libcurl development headers
- PostgreSQL client tools (psql)

**Production:**
- Docker and Docker Compose or Kubernetes cluster
- PostgreSQL 16+ server (or managed service)
- Optional: Redis for caching
- Optional: Kubernetes with Helm for enterprise deployment
- Reverse proxy (nginx/traefik) for TLS termination

**Container Base Images:**
- `postgres:16-bullseye` - Metadata and metrics databases
- `grafana/grafana:11.0.0` - Dashboard and alerting
- `redis:7-alpine` - Optional caching layer
- Custom multi-stage builds for backend, collector, frontend, ML service

---

*Stack analysis: 2026-03-30*
