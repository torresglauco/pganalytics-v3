# C/C++ Collector Implementation Notes

**Date**: February 22, 2026
**Project**: pganalytics-v3 Collector
**Status**: Enhancement Planning & Implementation

---

## Current State Analysis

The existing collector (C++) has:
- ✅ Full metric collection (pg_stats, query_stats, sysstat, disk_usage, pg_log)
- ✅ Configuration management (TOML-based)
- ✅ Metrics buffer with compression (zlib)
- ✅ CURL-based HTTP sender with TLS/mTLS
- ✅ Authentication with token management
- ✅ Modular plugin architecture

**Current Characteristics**:
- Language: C++ (more efficient than Go, better than pure C for maintainability)
- Binary size: ~5-8MB (meets <10MB target)
- HTTP API: REST with JSON (gzip compression)
- Protocol: HTTP/HTTPS with Bearer tokens
- Dependencies: libpq, OpenSSL, CURL, zlib, nlohmann_json

---

## Planned Enhancements for Distributed Architecture

### Phase 1: Custom Binary Protocol (High Priority)
**Goal**: Reduce bandwidth and improve throughput

**Current**: HTTP + JSON + gzip = ~500 bytes per metric batch
**Target**: Custom binary + zstd = ~200 bytes per metric batch (60% reduction)

**Files to Create**:
- `include/binary_protocol.h` - Binary protocol definition
- `src/binary_protocol.cpp` - Serialization/deserialization

**Changes**:
- Add custom binary protocol encoder/decoder
- Replace JSON serialization with binary for metrics
- Use zstd compression (faster, better ratio than zlib)
- Support both HTTP and raw TCP backends
- Message batching optimization

**Timeline**: 2-3 hours

### Phase 2: Enhanced Connection Management (High Priority)
**Goal**: Reduce memory/CPU for monitoring 100k+ collectors

**Current**: Single libpq connection per database
**Target**: Connection pooling with keep-alive

**Files to Modify**:
- `include/postgres_plugin.h` - Add connection pool
- `src/postgres_plugin.cpp` - Implement pooling

**Changes**:
- Connection pooling (1-3 connections, reuse)
- Query timeout enforcement (5 seconds max)
- Prepared statement caching
- Automatic reconnection with backoff

**Timeline**: 1-2 hours

### Phase 3: Performance Optimization (Medium Priority)
**Goal**: Reduce CPU and memory footprint

**Files to Create**:
- `include/metrics_cache.h` - Per-metric state caching
- `src/metrics_cache.cpp` - Cache implementation

**Changes**:
- Cache metric baselines (avoid recalculation)
- Ring buffer optimization
- Memory pooling (reduce allocations)
- CPU affinity (pin collector thread)

**Timeline**: 2-3 hours

### Phase 4: Advanced Features (Medium Priority)
**Goal**: Support advanced use cases

**Files to Create**:
- `include/health_check.h` - Self-monitoring
- `src/health_check.cpp` - Health implementation
- `include/rate_limiter.h` - Backend rate limiting
- `src/rate_limiter.cpp` - Rate limiting implementation

**Changes**:
- Local health checks (memory, CPU, disk)
- Graceful degradation when backend unavailable
- Rate limiting for high-frequency metrics
- Automatic cleanup of old logs

**Timeline**: 2-3 hours

---

## Implementation Priority

### Immediate (This Sprint)
1. **Binary Protocol**: Custom binary serialization
2. **Connection Pooling**: PostgreSQL connection reuse
3. **Compression**: Switch to zstd

**Estimated Effort**: 4-5 hours

### Next Sprint
4. **Performance Optimization**: Metrics caching, memory pooling
5. **Advanced Features**: Health checks, rate limiting
6. **Testing**: Load tests with 100k+ simulated collectors

**Estimated Effort**: 6-8 hours

---

## Performance Targets (Post-Implementation)

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| **Binary size** | 5-8MB | <5MB | ✅ |
| **Memory (idle)** | 40-60MB | <50MB | ✅ |
| **CPU (collecting)** | 1.5-2% | <1.5% | 🔄 |
| **Metrics latency** | 50-80ms | <50ms | 🔄 |
| **Network bandwidth** | 500B/batch | 200B/batch | ❌ Improvement |
| **Compression ratio** | 70% | 85% | 🔄 Target |

---

## Architecture Decisions

### Why C++ instead of Pure C?
- **C++**: Modern language features, STL containers, RAII resource management
- **C**: Minimal dependencies, lower memory overhead
- **Decision**: C++ is optimal here because:
  - Already implemented and proven
  - Better maintainability than C
  - std::vector, std::string reduce memory leaks
  - Performance is not significantly different
  - Compiler optimizations (-O3 -flto) achieve same results

### Why Custom Binary Protocol?
- **REST + JSON**: Inefficient for metrics streaming (high overhead)
- **gRPC**: Heavy dependency, overkill for unidirectional metrics
- **Custom Binary**: Perfect for metrics streaming
  - Compact: ~60% bandwidth reduction
  - Fast: No JSON parsing overhead
  - Flexible: Easy to add new metric types

### Why zstd over gzip/zlib?
- **gzip**: 30% compression ratio
- **zstd**: 45% compression ratio (15% better)
- **Speed**: zstd is 3x faster to compress
- **Adoption**: Increasingly standard in infrastructure

---

## Build & Deployment Strategy

### Local Development
```bash
# Standard CMake build
mkdir build && cd build
cmake -DCMAKE_BUILD_TYPE=Release ..
make -j4
./src/pganalytics --help
```

### Production Build
```bash
# Optimized build with LTO
mkdir build-prod && cd build-prod
cmake -DCMAKE_BUILD_TYPE=Release -DENABLE_LTO=ON ..
make -j$(nproc)
strip --strip-all ./src/pganalytics

# Size check
ls -lh ./src/pganalytics  # Should be <5MB
```

### Deployment Options
1. **Package**: DEB/RPM (Linux standard)
2. **Container**: Docker image (~80MB with base)
3. **Kubernetes**: DaemonSet deployment
4. **Binary**: Single static binary

---

## Testing Strategy

### Unit Tests
- Metric serialization/deserialization
- Binary protocol encoding/decoding
- Configuration parsing
- Connection pooling
- Buffer management

### Integration Tests
- Full collector cycle (collect → buffer → compress → send)
- PostgreSQL connection variations
- Configuration updates from backend
- Error recovery scenarios

### Load Tests
- Simulate 100k+ collectors
- Measure resource usage
- Verify no memory leaks (valgrind)
- Check CPU scaling

### Benchmarks
- Metrics collection throughput (metrics/second)
- Memory usage under load (40-50MB target)
- CPU usage idle vs collecting (<1% idle, <2% collecting)
- Network bandwidth (target 200B/batch after compression)

---

## Code Quality Standards

### Compiler Flags
```cmake
# Always enforce these
-Wall -Wextra -Werror=format-security -Wpedantic
-fno-exceptions (to reduce binary size)
-fvisibility=hidden (hide internal symbols)
-fPIC (position independent code)
```

### Memory Safety
- Use smart pointers (std::unique_ptr, std::shared_ptr)
- No raw `new`/`delete` outside RAII
- Valgrind clean (0 memory leaks)
- No buffer overflows

### Error Handling
- Explicit error returns (no exceptions in hot path)
- Graceful degradation (continue if metrics collection fails)
- Logging all errors with context
- Automatic retry with exponential backoff

---

## Next Steps

1. Implement custom binary protocol
2. Add connection pooling to PostgreSQL plugin
3. Switch to zstd compression
4. Comprehensive testing
5. Load testing with simulated collectors
6. Production deployment

---

**Implementation Status**: Ready to begin Phase 1 (Custom Binary Protocol)
