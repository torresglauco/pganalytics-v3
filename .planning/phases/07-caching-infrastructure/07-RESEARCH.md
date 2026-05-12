# Phase 07: Caching Infrastructure - Research

**Gathered:** 2026-05-12
**Status:** Complete (derived from milestone research)

---

## Phase Boundary

This phase adds response caching for faster API responses:
- Implement cache layer for dashboard and slow query APIs
- Add cache hit/miss metrics via Prometheus
- Implement cache invalidation strategies

**Depends on:** Phase 06 (Query Optimization Foundation - pgx v5, pool metrics)

**Enables:** Phase 08 (Dashboard Optimization), Phase 09 (Index Intelligence)

**Requirements:** API-01, MON-04

---

## Technical Approach

### 1. API Response Caching (API-01)

**Current State:**
- Existing `internal/cache/` package with LRU cache implementation
- `Cache[K, V]` interface with `Get`, `Set`, `Delete` methods
- TTL support already implemented
- NOT currently used for API responses

**Target State:**
- Cache middleware for API responses
- Cache key generation from request path/query params
- Cache warming for frequently accessed endpoints
- Graceful cache miss handling

**Implementation:**
1. Create `internal/middleware/cache_middleware.go`
2. Implement cache key hashing (path + query params)
3. Add response serialization for cache storage
4. Configure TTL per endpoint type

**Key Files:**
- `backend/internal/cache/cache.go` - existing cache implementation
- `backend/internal/middleware/cache_middleware.go` - new middleware
- `backend/internal/api/server.go` - middleware registration

**Cache Strategy by Endpoint:**
| Endpoint | TTL | Invalidation Trigger |
|----------|-----|---------------------|
| GET /api/v1/databases/:id/slow-queries | 5 min | New slow query detected |
| GET /api/v1/queries/:hash/timeline | 10 min | New timeline data |
| GET /api/v1/databases/:id/index-stats | 10 min | Index changes |
| GET /api/v1/system/pool-metrics | 30 sec | None (always fresh) |

### 2. Cache Metrics (MON-04)

**Approach:**
- Prometheus gauges for cache hits/misses
- Histogram for cache latency
- Per-endpoint metrics

**Metrics to Add:**
```
pganalytics_cache_hits_total{endpoint, method}
pganalytics_cache_misses_total{endpoint, method}
pganalytics_cache_size_bytes{cache_name}
pganalytics_cache_entries{cache_name}
pganalytics_cache_latency_seconds{operation}
```

**Implementation:**
1. Extend `internal/metrics/prometheus.go` with cache metrics
2. Add metrics recording in cache middleware
3. Create `/api/v1/system/cache-metrics` endpoint

**Key Files:**
- `backend/internal/metrics/prometheus.go` - add cache gauges
- `backend/internal/middleware/cache_middleware.go` - record metrics
- `backend/internal/api/handlers_metrics.go` - cache metrics endpoint

### 3. Cache Invalidation Strategies

**Automatic Invalidation:**
- Time-based TTL (already implemented)
- Size-based eviction (LRU already implemented)

**Manual Invalidation:**
- DELETE /api/v1/system/cache - clear all caches
- DELETE /api/v1/system/cache/:key - clear specific key
- Cache tag invalidation for grouped entries

**Event-Based Invalidation:**
- Invalidate slow queries cache when new slow query detected
- Invalidate index stats cache when index recommendations change

---

## Validation Architecture

### Dimension 1: Cache Effectiveness

**Test Strategy:**
- Unit tests for cache middleware
- Integration tests for cached endpoints
- Benchmark before/after with cache enabled

**Acceptance Criteria:**
- Cache hit rate > 80% for frequently accessed endpoints
- Cache miss penalty < 5ms (key generation + miss recording)
- No stale data returned

### Dimension 2: Metrics Accuracy

**Test Strategy:**
- Verify Prometheus gauges match actual cache operations
- Test counter increments on hit/miss
- Verify histogram percentiles

**Acceptance Criteria:**
- Metrics accurate to within 1 operation
- No negative gauge values
- Latency histogram covers 1ms-10s range

### Dimension 3: Invalidation Correctness

**Test Strategy:**
- Test TTL expiration
- Test manual invalidation endpoints
- Test event-based invalidation

**Acceptance Criteria:**
- Expired entries not returned
- Manual invalidation clears correct keys
- Event invalidation clears dependent caches

---

## Integration Points

### Backend
- `internal/cache/cache.go` - existing cache implementation
- `internal/middleware/cache_middleware.go` - new middleware
- `internal/metrics/prometheus.go` - cache metrics
- `internal/api/handlers_metrics.go` - metrics endpoint
- `internal/api/server.go` - middleware registration

### Existing Infrastructure
- LRU cache with TTL (already implemented)
- Prometheus metrics (Phase 06)
- Middleware pattern (Phase 06)

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Cache returns stale data | High | Short TTLs, invalidation on data change |
| Memory exhaustion | Medium | Size limits, LRU eviction, monitoring |
| Cache stampede | Medium | Single-flight pattern, cache warming |
| Cache key collision | Low | Include request hash in key |

---

## Dependencies

**Existing (extend):**
- `internal/cache/` - LRU cache with TTL
- `internal/metrics/prometheus.go` - Prometheus client

**No new external dependencies required.**

---

*Research completed: 2026-05-12*
*Source: Milestone research (.planning/research/), existing codebase analysis*