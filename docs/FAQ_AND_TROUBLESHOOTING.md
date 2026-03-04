# FAQ & Troubleshooting Guide

## Frequently Asked Questions

### Installation & Setup

#### Q: What are the system requirements?

**A:**
- **Docker**: 20.10+ with Docker Compose 2.0+
- **Go**: 1.22+ (for local backend development)
- **Node.js**: 18+ with npm 9+ (for frontend)
- **C++**: C++17 compiler (for collector compilation)
- **PostgreSQL**: 16+ (can use Docker)
- **RAM**: 4GB minimum, 8GB recommended
- **Disk**: 10GB+ for development environment

#### Q: How do I quickly get started?

**A:**
```bash
# Clone repository
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3

# Start with Docker Compose
docker-compose up -d

# Access services
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# Grafana: http://localhost:3000 (admin/admin)
# PostgreSQL: localhost:5432
```

#### Q: Can I run without Docker?

**A:** Yes, but more complex:
1. Install PostgreSQL 16+
2. Create databases manually
3. Run backend: `go run ./cmd/pganalytics-api/main.go`
4. Run frontend: `npm run dev` (in frontend/)
5. Install collector separately

Recommended: Use Docker Compose for local development

#### Q: How do I configure the backend?

**A:** Environment variables in `.env`:
```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/pganalytics
TIMESCALE_URL=postgres://user:pass@localhost:5433/pganalytics_metrics

# JWT
JWT_SECRET=your-secret-key-min-32-chars

# API
API_PORT=8080
API_HOST=0.0.0.0

# Logging
LOG_LEVEL=info

# CORS
CORS_ORIGINS=http://localhost:3000
```

### Development

#### Q: How do I run tests locally?

**A:**
```bash
# Frontend tests
cd frontend
npm test              # Run all tests
npm test -- --watch   # Watch mode
npm run type-check    # Type checking
npm run lint          # ESLint

# Backend tests
cd backend
go test ./...                        # Run all tests
go test -v ./...                     # Verbose
go test -race ./...                  # With race detector
go test -coverprofile=c.out ./...    # With coverage

# E2E tests
cd frontend
npx playwright test                  # All browsers
npx playwright test --project=chromium  # Specific browser
npx playwright test --debug          # Debug mode
```

#### Q: How do I debug the backend?

**A:** Using Delve debugger:
```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug ./cmd/pganalytics-api/main.go

# In debugger:
(dlv) break main.main  # Set breakpoint
(dlv) continue         # Continue execution
(dlv) print var_name   # Print variable
(dlv) quit             # Exit debugger
```

Or use VS Code:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Connect to Delve",
      "type": "go",
      "mode": "remote",
      "remotePath": "${workspaceFolder}",
      "port": 2345,
      "host": "localhost",
      "showLog": true,
      "trace": "verbose"
    }
  ]
}
```

#### Q: How do I debug the frontend?

**A:**
```bash
# Chrome DevTools
npm run dev
# Open http://localhost:3000
# Press F12 for DevTools

# VS Code Debugger
# Add to .vscode/launch.json:
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "chrome",
      "request": "attach",
      "name": "Attach Chrome",
      "port": 9222,
      "pathMapping": {
        "/": "${workspaceRoot}/",
        "/src": "${workspaceRoot}/src"
      }
    }
  ]
}
```

#### Q: How do I add a new API endpoint?

**A:**
1. Define handler in `backend/internal/api/handlers.go`:
```go
// GetStatus returns system status
// @Summary Get system status
// @ID get-status
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /status [get]
func (s *Server) GetStatus(c *gin.Context) {
    status := StatusResponse{
        Healthy: true,
        Version: version,
    }
    c.JSON(http.StatusOK, status)
}
```

2. Register route in `main()`:
```go
router.GET("/api/v1/status", server.GetStatus)
```

3. Add tests in `backend/tests/integration/handlers_test.go`:
```go
func TestGetStatus(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/v1/status", nil)
    w := httptest.NewRecorder()

    server.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Database

#### Q: How do I migrate the database?

**A:**
```bash
# Migrations are in backend/migrations/
# They run automatically on backend startup

# Or manually with psql:
psql -h localhost -U postgres -d pganalytics \
     -f backend/migrations/001_init.sql

# Check migration status:
SELECT version, dirty FROM schema_migrations;
```

#### Q: How do I reset the database?

**A:**
```bash
# Stop containers
docker-compose down

# Remove volumes (WARNING: data loss)
docker volume rm pganalytics-v3_postgres_data

# Restart
docker-compose up -d

# Database will be initialized automatically
```

#### Q: How do I backup the database?

**A:**
```bash
# Full backup
pg_dump -h localhost -U postgres pganalytics > backup.sql

# Compressed backup
pg_dump -h localhost -U postgres pganalytics | gzip > backup.sql.gz

# Restore from backup
psql -h localhost -U postgres pganalytics < backup.sql
```

### Collectors

#### Q: How do I register a collector?

**A:**
1. In UI: Go to Collectors → Register Collector
2. Fill in details:
   - Hostname/IP of PostgreSQL server
   - Port (default: 5432)
   - Database name
   - Username
   - Password (optional, uses .pgpass)

3. Or via API:
```bash
curl -X POST http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "db.example.com",
    "port": 5432,
    "database": "mydb",
    "username": "monitoring",
    "encrypted_password": "..."
  }'
```

#### Q: How do I verify collector connectivity?

**A:**
```bash
# From UI: Click "Test Connection" button

# From CLI:
psql -h db.example.com -U monitoring -d mydb -c "SELECT 1"

# Check collector logs:
docker-compose logs collector-1
```

#### Q: How do I remove a collector?

**A:**
```bash
# From UI: Select collector → Delete → Confirm

# Via API:
curl -X DELETE http://localhost:8080/api/v1/collectors/collector-id \
  -H "Authorization: Bearer $TOKEN"
```

### Metrics

#### Q: How long are metrics retained?

**A:**
- 7-day retention: Table metrics (stats, queries, indexes)
- 30-day retention: System metrics (disk usage)
- Configurable in database schema

#### Q: How do I query metrics directly?

**A:**
```bash
# Connect to TimescaleDB
psql -h localhost -U postgres -d pganalytics_metrics

# Query latest metrics
SELECT * FROM metrics_pg_stats_table
  WHERE time > NOW() - INTERVAL '1 day'
  ORDER BY time DESC
  LIMIT 10;

# Aggregate metrics
SELECT
  DATE_TRUNC('hour', time) as hour,
  AVG(cache_hit_ratio) as avg_cache_hits
FROM metrics_pg_stats_database
WHERE time > NOW() - INTERVAL '7 days'
GROUP BY hour
ORDER BY hour DESC;
```

#### Q: How do I export metrics?

**A:**
```bash
# Export to CSV
COPY (
  SELECT time, collector_id, metric_name, metric_value
  FROM metrics_data
  WHERE time > NOW() - INTERVAL '1 day'
) TO STDOUT CSV HEADER;

# Or use tools like pgAdmin
# Or REST API: /api/v1/metrics/export?format=csv
```

### Deployment

#### Q: How do I deploy to production?

**A:** Follow deployment guide:
```bash
1. Review DEPLOYMENT_START_HERE.md
2. Use DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md for setup
3. Follow 4-phase deployment in DEPLOYMENT_PLAN_v3.2.0.md
4. Verify with PHASE1_EXECUTION_CHECKLIST_V2.md
5. Monitor with docs/OPERATIONS_HA_DR.md procedures
```

#### Q: How do I update from v3.2 to v3.3?

**A:** Follow upgrade guide:
```bash
1. Read UPGRADE_v3.2_TO_v3.3.md completely
2. Review breaking changes section
3. Backup database: pg_dump > backup.sql
4. Follow 4-phase upgrade procedure
5. Have rollback plan ready
6. Test thoroughly before production
```

#### Q: How do I set up high availability?

**A:**
Follow `docs/OPERATIONS_HA_DR.md`:
- 3-server setup with load balancer
- Database streaming replication
- Automated daily backups
- Failover procedures
- Monitoring with Prometheus

### Performance

#### Q: How can I improve query performance?

**A:**
1. Check slow queries in dashboards
2. Add indexes: `CREATE INDEX idx_name ON table(column);`
3. Analyze queries: `EXPLAIN ANALYZE SELECT ...;`
4. Use ML optimization: Check AI Suggestions tab
5. Archive old metrics

#### Q: How do I monitor system performance?

**A:**
- Grafana dashboards: http://localhost:3000
- Backend metrics: /api/v1/health
- Database metrics: Check pgAdmin
- Frontend performance: DevTools → Performance tab

#### Q: How do I configure performance thresholds?

**A:**
Edit backend/internal/config/config.go:
```go
const (
    DefaultQueryWarningMS = 1000      // Warn if > 1s
    DefaultLockWarningMS  = 500       // Warn if > 500ms
    DefaultCacheMissRatio = 0.1       // Warn if < 90% hit ratio
)
```

### Security

#### Q: How do I reset a user password?

**A:**
```bash
# Via UI: Users → Select user → Change Password

# Via SQL:
UPDATE users SET password_hash = crypt('newpass', gen_salt('bf'))
WHERE email = 'user@example.com';

# Via API:
curl -X POST http://localhost:8080/api/v1/users/user-id/password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"new_password": "newpass123"}'
```

#### Q: How do I manage API tokens?

**A:**
```bash
# Create token
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "monitoring-token"}'

# List tokens
curl http://localhost:8080/api/v1/tokens \
  -H "Authorization: Bearer $TOKEN"

# Revoke token
curl -X DELETE http://localhost:8080/api/v1/tokens/token-id \
  -H "Authorization: Bearer $TOKEN"
```

#### Q: How do I secure collector connections?

**A:**
- Use mTLS for collector communication
- Store passwords in secure vault
- Use .pgpass for PostgreSQL auth
- Enable encryption in transit
- See docs/API_SECURITY_REFERENCE.md

### Troubleshooting

#### Q: Services won't start with docker-compose

**A:**
```bash
# Check logs
docker-compose logs

# Check specific service
docker-compose logs api
docker-compose logs postgres

# Rebuild containers
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

#### Q: Frontend won't connect to backend

**A:**
```bash
# Check backend is running
curl http://localhost:8080/api/v1/health

# Check CORS configuration
# See API_SECURITY_REFERENCE.md → CORS

# Check firewall
netstat -tulnp | grep 8080

# Update CORS in backend config if needed
```

#### Q: Database connection fails

**A:**
```bash
# Check database is running
docker-compose ps

# Check connection string
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL -c "SELECT 1"

# Check postgres logs
docker-compose logs postgres
```

#### Q: Tests fail randomly (flaky)

**A:**
```bash
# Increase timeout
export TEST_TIMEOUT=30000

# Run specific test multiple times
npm test -- --repeat=5 test-name.test.ts

# Check test logs for timing issues
```

#### Q: Build fails with "out of memory"

**A:**
```bash
# Increase Node.js memory
export NODE_OPTIONS="--max-old-space-size=4096"

# Or increase system memory
# Docker Desktop: Settings → Resources

# Clean build cache
npm run build:clean
npm run build
```

### Support

#### Q: Where can I get help?

**A:**
- **Documentation**: See `docs/` directory
- **Issues**: GitHub Issues tab
- **Discussions**: GitHub Discussions
- **Security**: Email security@pganalytics.local
- **Community**: Check contributing guide

#### Q: How do I report a bug?

**A:**
1. Check if issue exists
2. Create detailed issue with:
   - Version (v3.2.0, etc.)
   - Environment (Linux, macOS, etc.)
   - Reproduction steps
   - Expected vs actual behavior
   - Logs and screenshots
3. Use issue template if available

#### Q: How do I contribute?

**A:**
See CONTRIBUTING.md:
1. Fork repository
2. Create feature branch
3. Follow code standards
4. Add tests
5. Submit pull request
6. Get review and merge

---

## Quick Reference

### Common Commands

```bash
# Development
npm run dev              # Frontend dev server
go run ./cmd/.../main.go # Backend dev server
docker-compose up -d    # Start services

# Testing
npm test                # Frontend tests
go test ./...           # Backend tests
npx playwright test     # E2E tests

# Building
npm run build          # Frontend production build
go build ./cmd/...     # Backend binary build

# Database
docker-compose exec postgres psql -U postgres
pg_dump > backup.sql
psql < backup.sql

# Debugging
npm run debug          # Frontend with debugging
dlv debug ./cmd/...    # Backend with Delve

# Cleanup
docker-compose down    # Stop containers
docker volume prune    # Remove unused volumes
npm ci                 # Clean install npm
```

### Useful Links

- **Repository**: https://github.com/torresglauco/pganalytics-v3
- **Issues**: https://github.com/torresglauco/pganalytics-v3/issues
- **Discussions**: https://github.com/torresglauco/pganalytics-v3/discussions
- **Docker Hub**: https://hub.docker.com/pganalytics
- **API Docs**: http://localhost:8080/swagger

---

## Still Need Help?

1. Check this FAQ first
2. Search existing GitHub issues
3. Check documentation in `docs/`
4. Ask in discussions
5. Email support for security issues

We're here to help! 🚀
