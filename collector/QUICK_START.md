# Collector Development Quick Start

**Fast-track guide for building and testing the collector**

---

## 5-Minute Setup (Linux/macOS)

```bash
# 1. Install dependencies
# Ubuntu/Debian
sudo apt-get install -y cmake build-essential libpq-dev libssl-dev libcurl4-openssl-dev zlib1g-dev libzstd-dev

# macOS
brew install cmake postgresql openssl curl zlib zstd

# 2. Build
cd /path/to/pganalytics-v3/collector
mkdir build && cd build
cmake ..
make -j4

# 3. Verify
./src/pganalytics --help
```

---

## Common Build Commands

```bash
# Development build (with debug symbols)
cmake -DCMAKE_BUILD_TYPE=Debug ..

# Production build (optimized)
cmake -DCMAKE_BUILD_TYPE=Release ..

# With tests
cmake -DBUILD_TESTS=ON ..

# Without tests
cmake -DBUILD_TESTS=OFF ..

# Full rebuild
make clean && make -j4

# Install
sudo make install  # Installs to /usr/local/bin
```

---

## Testing

```bash
# Run unit tests
ctest --output-on-failure

# Check binary size
ls -lh ./src/pganalytics

# Check dependencies
ldd ./src/pganalytics

# Run collector (test mode)
./src/pganalytics cron
```

---

## New Components

### Binary Protocol
- **Header**: `include/binary_protocol.h`
- **Source**: `src/binary_protocol.cpp`
- **Purpose**: Efficient metrics serialization
- **Usage**: See COLLECTOR_IMPLEMENTATION_SUMMARY.md

### Connection Pool
- **Header**: `include/connection_pool.h`
- **Source**: `src/connection_pool.cpp`
- **Purpose**: Reusable PostgreSQL connections
- **Usage**: See COLLECTOR_IMPLEMENTATION_SUMMARY.md

---

## Integration Checklist

When integrating new components into existing collector:

- [ ] Add `#include "binary_protocol.h"` to sender.cpp
- [ ] Add `#include "connection_pool.h"` to postgres_plugin.cpp
- [ ] Replace JSON serialization with `MessageBuilder::createMetricsBatch()`
- [ ] Replace `PQconnectdb()` with connection pool
- [ ] Update CMakeLists.txt (already done)
- [ ] Compile and verify no errors
- [ ] Run unit tests
- [ ] Performance test

---

## Documentation Files

- **COLLECTOR_IMPLEMENTATION_NOTES.md** - Design decisions and implementation strategy
- **COLLECTOR_IMPLEMENTATION_SUMMARY.md** - What was implemented and how
- **BUILD_AND_DEPLOY.md** - Detailed build, test, and deployment guide
- **QUICK_START.md** - This file

---

## File Structure

```
collector/
├── include/
│   ├── binary_protocol.h          ← NEW: Binary protocol definition
│   ├── connection_pool.h          ← NEW: Connection pooling
│   ├── collector.h
│   ├── postgres_plugin.h
│   ├── sender.h
│   └── ... (other headers)
├── src/
│   ├── binary_protocol.cpp        ← NEW: Binary protocol impl
│   ├── connection_pool.cpp        ← NEW: Connection pool impl
│   ├── main.cpp
│   ├── postgres_plugin.cpp
│   ├── sender.cpp
│   └── ... (other sources)
├── tests/
│   ├── unit/                      ← Unit tests (to be added)
│   ├── integration/               ← Integration tests (to be added)
│   └── ... (existing tests)
├── CMakeLists.txt                 ← UPDATED with new sources
├── config.toml.sample
└── BUILD_AND_DEPLOY.md            ← NEW: Build guide
```

---

## Performance Targets

After full integration, you should see:

| Metric | Target |
|--------|--------|
| Binary size | <5MB |
| Memory (idle) | <50MB |
| CPU (idle) | <1% |
| Metrics latency | <50ms |
| Network bandwidth | <200B/metric |

---

## Troubleshooting

**CMake errors**: Install missing packages (see "5-Minute Setup")

**Compilation errors**:
- Check PostgreSQL: `pg_config --version`
- Check OpenSSL: `openssl version`
- Try: `cmake --debug-output ..`

**Runtime errors**:
- Check PostgreSQL is running: `psql -c "SELECT 1"`
- Check config: `cat /etc/pganalytics/collector.conf`
- Check logs: `tail -f /var/log/pganalytics/collector.log`

---

## Next Steps

1. Build collector: `make -j4`
2. Run tests: `ctest`
3. Check performance: `ls -lh ./src/pganalytics`
4. Integrate components (see COLLECTOR_IMPLEMENTATION_SUMMARY.md)
5. Performance test

---

For detailed information, see the comprehensive documentation files listed above.
