# Getting Started with pgAnalytics v3.0

## Welcome! 👋

This document will help you get started with pgAnalytics v3.0 development. The project structure is complete and ready for Phase 2 implementation.

## Prerequisites

Before starting, ensure you have:

### Required
- **Docker** 20.10+ ([install](https://docs.docker.com/engine/install/))
- **Docker Compose** 2.0+ ([install](https://docs.docker.com/compose/install/))
- **Git** ([install](https://git-scm.com/))

### For Backend Development (Go)
- **Go** 1.22+ ([install](https://golang.org/dl/))
- **Make** (usually pre-installed on macOS/Linux)

### For Collector Development (C/C++)
- **CMake** 3.25+ ([install](https://cmake.org/download/))
- **C++ compiler** (GCC/Clang with C++17 support)
- **Dependencies**: libcurl, openssl 3.0+, postgresql-dev

## Project Layout

```
pganalytics-v3/
├── backend/          # Go backend API
├── collector/        # C/C++ distributed collector
├── grafana/          # Dashboards and visualization
├── docs/             # Documentation
└── docker-compose.yml # Full environment
```

## Quick Start

### 1. Clone or Navigate to Repository

```bash
cd /Users/glauco.torres/git/pganalytics-v3
```

### 2. Review Key Documentation

Start with understanding the architecture:

```bash
# Architecture and design decisions
cat docs/ARCHITECTURE.md

# Project overview
cat README.md

# Phase 1 summary
cat PHASE_1_SUMMARY.md
```

### 3. Verify Setup

Check that all files are in place:

```bash
# List key files
make version  # Show version info
make help     # Show available commands

# Verify docker-compose configuration
docker-compose config
```

### 4. Check Git Status

```bash
git log --oneline
git status
```

## Development Workflow

### Building Locally (Backend)

```bash
# Navigate to backend
cd backend

# Install dependencies
go mod download

# Build the backend binary
go build -o pganalytics-api ./cmd/pganalytics-api

# Run tests
go test ./...

# Format code
go fmt ./...
```

### Building Locally (Collector)

```bash
cd collector

# Create build directory
mkdir -p build && cd build

# Configure with CMake
cmake ..

# Build
make

# Run tests
make test
```

### Using Make (Recommended)

```bash
# Show all available commands
make help

# Build everything
make build

# Run tests
make test

# Start Docker environment (Phase 2+)
make docker-up

# Stop Docker environment
make docker-down

# Clean build artifacts
make clean
```

## Phase 1 Status

✅ **COMPLETE**

- Project structure (monorepo)
- Database schemas (PostgreSQL + TimescaleDB)
- Backend skeleton (Go + Gin)
- Collector architecture (C++17)
- Docker Compose configuration
- Documentation and guides
- Git repository initialized

## Phase 2 Preview (Backend Core)

The next phase will focus on implementing the core backend API:

### Key Files to Modify

```
backend/
├── internal/
│   ├── api/
│   │   ├── handlers.go      ← HTTP handlers
│   │   ├── routes.go        ← API routes
│   │   └── middleware.go    ← Auth middleware
│   ├── auth/
│   │   ├── jwt.go           ← Token management
│   │   └── mtls.go          ← Certificate handling
│   ├── collector/
│   │   ├── service.go       ← Business logic
│   │   └── store.go         ← Database access
│   ├── metrics/
│   │   ├── receiver.go      ← Parse metrics
│   │   └── storage.go       ← Insert to DB
│   └── storage/
│       ├── postgres.go      ← PostgreSQL layer
│       └── queries/         ← SQL queries
└── tests/
    └── integration/         ← E2E tests
```

### Key Tasks for Phase 2

1. **API Endpoints**
   - `POST /api/v1/collectors/register`
   - `GET /api/v1/collectors`
   - `POST /api/v1/metrics/push`
   - `GET /api/v1/config/{id}`
   - `GET /api/v1/health`

2. **Authentication**
   - JWT token generation
   - Token validation
   - mTLS certificate verification

3. **Metrics Ingestion**
   - JSON schema validation
   - Data insertion to TimescaleDB
   - Compression/decompression

4. **Testing**
   - Unit tests for each package
   - Integration tests with databases
   - Load tests

## Code Style & Standards

### Go
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- 70%+ test coverage target

### C/C++
- C++17 standard
- Follow [Google C++ Style Guide](https://google.github.io/styleguide/cppguide.html)
- Use clang-format for formatting
- 60%+ test coverage target

### SQL
- Use SQLC for type-safe queries
- Include comments for complex logic
- Test migrations thoroughly

## Database Development

### PostgreSQL Migrations

Migrations are in `backend/migrations/`:

```bash
# View migration status
docker-compose exec postgres psql -U postgres -d pganalytics -c "\d"

# Run a specific migration
docker-compose exec postgres psql -U postgres -d pganalytics < migrations/001_init.sql

# Connect to database
make psql
```

### SQLC for Type-Safe Queries

We use SQLC to generate type-safe database code. Add queries to `backend/queries/` directory and run:

```bash
cd backend
go generate ./...
```

## Testing Strategy

### Unit Tests (Backend)

```bash
cd backend
go test -v ./internal/auth
go test -v ./internal/metrics
go test -v ./internal/collector
```

### Integration Tests

```bash
# Start services
docker-compose up -d

# Run integration tests
go test -v -tags=integration ./tests/integration/...
```

### Unit Tests (Collector)

```bash
cd collector/build
ctest --verbose
```

## Debugging

### Backend

```bash
# Run with verbose logging
ENVIRONMENT=development LOG_LEVEL=debug ./pganalytics-api

# Attach to running container
docker-compose exec backend ./pganalytics-api
```

### Collector

```bash
# Build with debug symbols
cd collector/build
cmake -DCMAKE_BUILD_TYPE=Debug ..
make

# Run with logging
./src/pganalytics
```

### Database

```bash
# Connect to PostgreSQL
psql postgres://postgres:pganalytics@localhost:5432/pganalytics

# Connect to TimescaleDB
psql postgres://postgres:pganalytics@localhost:5433/metrics

# View logs
docker-compose logs -f postgres
docker-compose logs -f timescale
```

## Git Workflow

```bash
# Create feature branch
git checkout -b feature/collector-registration

# Make changes and commit
git add .
git commit -m "feat: implement collector registration endpoint"

# Push to origin
git push origin feature/collector-registration

# Create pull request on GitHub
```

## Documentation

Always keep documentation updated:

- **Code comments** for complex logic
- **README files** in each package
- **API docs** via Swagger/OpenAPI
- **Architecture decisions** in ARCHITECTURE.md

## Common Tasks

### Run the Demo Environment

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f backend

# Stop all services
docker-compose down
```

### Generate API Documentation

```bash
cd backend
swag init -g cmd/pganalytics-api/main.go
```

### Update Dependencies

```bash
# Go
cd backend
go mod tidy
go mod download

# C++ (using vcpkg)
cd collector
vcpkg install
```

### Run Security Checks

```bash
# Go security
gosec ./...

# C++ security (future)
# Add clang-tidy configuration
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Docker Issues

```bash
# Clean up everything
docker-compose down -v
docker system prune

# Rebuild images
docker-compose build --no-cache
```

### Database Connection Issues

```bash
# Check database is running
docker-compose ps

# Test connection
docker-compose exec postgres pg_isready

# View logs
docker-compose logs postgres
```

## Performance Tips

### Go Development
- Use `go run` for quick testing
- Use `go build -race` to detect race conditions
- Profile with `pprof`

### Compiler Development
- Use incremental builds with CMake
- Cache dependencies with ccache

## Resources

### Documentation
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - Design overview
- [README.md](README.md) - Project introduction
- [PHASE_1_SUMMARY.md](PHASE_1_SUMMARY.md) - What we've built

### External
- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://github.com/gin-gonic/gin)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)
- [TimescaleDB Docs](https://docs.timescale.com/)

## Getting Help

1. Check existing documentation in `/docs/`
2. Look at ARCHITECTURE.md for design decisions
3. Review comments in source code
4. Check git history for context: `git log --oneline`

## Next Steps

1. **Review** the ARCHITECTURE.md document
2. **Explore** the codebase structure
3. **Read** the README.md and PHASE_1_SUMMARY.md
4. **Setup** your local development environment
5. **Start** Phase 2 implementation!

---

**Questions?** Check the docs or review the architecture document. Everything is documented!

**Ready to code?** Pick a Phase 2 task and start implementing!
