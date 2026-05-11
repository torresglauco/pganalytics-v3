# Stack Research: PostgreSQL Performance Optimization

**Domain:** Performance optimization for PostgreSQL monitoring platform
**Researched:** 2026-05-11
**Confidence:** HIGH

## Recommended Stack Additions

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **pgx v5** | v5.9.2 | PostgreSQL driver | 2-3x faster than lib/pq, native connection pooling (pgxpool), better prepared statement handling, PostgreSQL-specific optimizations. Current lib/pq is in maintenance mode. |
| **go-redis v9** | v9.19.0 | Distributed caching | Already in docker-compose.yml. Enables cache sharing across API instances, pub/sub for cache invalidation, atomic operations for rate limiting. Complements existing in-memory cache. |
| **pprof** | net/http/pprof (stdlib) | Performance profiling | Built into Go runtime. Enables on-demand CPU/memory profiling in production without restart. Critical for identifying bottlenecks in dashboard queries. |
| **Prometheus client_golang** | v1.23.2 (already installed) | Performance metrics | Already integrated. Add histogram metrics for API response times, query durations, cache hit rates. Use existing metrics infrastructure. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| **miniredis v2** | v2.37.0 | Redis mock for testing | Unit/integration tests without Redis dependency. Faster than testcontainers for cache tests. |
| **go-redis/redis_rate v10** | v10.0.1 | Rate limiting with Redis | When implementing distributed rate limiting per instance/user. Uses Redis for atomic token bucket. |
| **freecache v1.2.7** | v1.2.7 | Alternative in-memory cache | If existing cache shows memory pressure. Pre-allocates memory, zero GC overhead. Consider if profiling shows cache GC impact. |

## What NOT to Add

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **lib/pq** | In maintenance mode, slower, no native pooling | pgx v5 - actively maintained, faster, pgxpool included |
| **GORM/sqlx** | Adds abstraction overhead, less control | Native pgx queries with prepared statements |
| **Third-party cache libraries** | Already have functional in-memory cache | Enhance existing cache or add Redis for distributed scenarios |
| **New monitoring tools** | Already have Prometheus infrastructure | Extend existing metrics with performance histograms |

## Installation

```bash
# Core - PostgreSQL driver migration
go get github.com/jackc/pgx/v5@v5.9.2

# Core - Redis client
go get github.com/redis/go-redis/v9@v9.19.0

# Development/Testing - Redis mock
go get github.com/alicebob/miniredis/v2@v2.37.0

# Optional - Rate limiting with Redis
go get github.com/go-redis/redis_rate/v10@v10.0.1

# Profiling (stdlib, no install needed)
# Just import: import _ "net/http/pprof"
```

## Integration Strategy

### 1. pgx Migration (HIGH PRIORITY)

**Current State:** Using `lib/pq` v1.10.9 with standard `database/sql` interface.

**Migration Path:**
1. Add pgx v5 as new driver alongside existing lib/pq
2. Migrate storage layer to use pgxpool for connection pooling
3. Update connection string format (minimal changes)
4. Leverage pgx-specific features: prepared statements, batch queries, COPY

**Benefits:**
- Native connection pooling (no external pooler needed)
- Binary protocol (faster than lib/pq text protocol)
- Better prepared statement caching
- PostgreSQL-specific types (pgtype) without overhead

**Code Pattern:**
```go
// Before (lib/pq)
db, err := sql.Open("postgres", connString)

// After (pgxpool)
pool, err := pgxpool.New(ctx, connString)
```

### 2. Redis Integration (MEDIUM PRIORITY)

**Current State:** Redis container in docker-compose.yml (optional profile), not used in code.

**Integration Points:**
- Dashboard aggregation cache (5-minute TTL)
- Query execution plan cache (30-minute TTL)
- Index recommendation cache (15-minute TTL)
- Session cache for frequently accessed data

**Architecture Decision:**
- Use L1 (in-memory) + L2 (Redis) cache pattern
- In-memory: Hot data, sub-millisecond access
- Redis: Warm data, cross-instance sharing

**When Redis is Worth It:**
- Multiple API instances running
- Cache invalidation events (schema changes, vacuum)
- Rate limiting per instance

**When to Skip Redis:**
- Single instance deployment
- Adding complexity without benefit
- Development/staging environments

### 3. Performance Profiling (LOW PRIORITY - Quick Win)

**Current State:** No pprof integration detected.

**Implementation:**
```go
import _ "net/http/pprof"

// Already have HTTP server, pprof registers at /debug/pprof/
// No code changes needed beyond import
```

**Use Cases:**
- Profile slow dashboard loads in staging
- Identify memory leaks from cache
- CPU hotspots in query analysis

## Stack Patterns by Deployment Scenario

**Single Instance (Development/Staging):**
- Use pgxpool with appropriate MaxConns (default 4-8)
- Keep in-memory cache only (no Redis)
- Enable pprof for profiling
- Skip redis_rate

**Multi-Instance (Production):**
- Use pgxpool with MaxConns based on DB capacity (20-50 per instance)
- Enable Redis for L2 cache + cache invalidation
- Use redis_rate for distributed rate limiting
- Keep pprof behind auth middleware

## Version Compatibility

| Package | Compatible With | Notes |
|---------|-----------------|-------|
| pgx v5.9.2 | Go 1.21+ | Current project uses Go 1.26.1 - compatible |
| go-redis v9.19.0 | Go 1.21+ | Compatible with current Go version |
| miniredis v2.37.0 | go-redis v9 | Uses same API, drop-in mock |
| pgxpool (pgx sub-package) | pgx v5 | Included in pgx v5, no separate install |
| Prometheus v1.23.2 | Go 1.21+ | Already compatible, using v1.23.2 |

## Performance Optimization Features Mapping

| Performance Feature | Primary Tool | Supporting Tools |
|---------------------|--------------|------------------|
| Query optimization | pgx prepared statements, EXPLAIN caching | In-memory cache (existing), Redis (optional) |
| Connection pooling | pgxpool (native) | Remove external pooler complexity |
| API response caching | In-memory cache (existing) | Redis L2 for distributed deployments |
| Performance monitoring | Prometheus histograms | pprof for deep dive profiling |
| Index analysis | Existing index_advisor service | Cache recommendations in Redis |
| Dashboard aggregations | pgx batch queries | TimescaleDB optimizations |

## Migration Risk Assessment

| Migration | Risk | Mitigation |
|-----------|------|------------|
| lib/pq to pgx | MEDIUM - interface changes | Incremental migration, maintain lib/pq fallback initially |
| Add Redis | LOW - additive change | Feature flag, Redis optional, fallback to in-memory |
| Enable pprof | LOW - stdlib import | Add auth middleware in production |

## Existing Infrastructure Leveraged

| Existing | How to Extend |
|----------|---------------|
| Prometheus client | Add histogram metrics for query_duration_seconds, api_response_seconds, cache_hit_rate |
| Zap logger | Add structured logging for slow queries (>100ms), cache misses |
| Cache metrics | Extend with Redis metrics, L1/L2 hit rates |
| testcontainers | Add Redis testcontainer for integration tests (or use miniredis for unit tests) |
| Docker Compose | Redis already defined, just needs profile removal and Go integration |

## Sources

- Go module version queries (go list -m -versions) - HIGH confidence
- Existing codebase analysis (go.mod, cache implementations) - HIGH confidence
- Docker compose configuration - HIGH confidence
- pgx v5 documentation (github.com/jackc/pgx) - MEDIUM confidence (WebFetch failed, using module list)
- go-redis documentation (github.com/redis/go-redis) - MEDIUM confidence (WebFetch failed, using module list)

---
*Stack research for: PostgreSQL Performance Optimization*
*Researched: 2026-05-11*