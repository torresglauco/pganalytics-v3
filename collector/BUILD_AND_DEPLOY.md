# Collector Build and Deployment Guide

**Date**: February 22, 2026
**Project**: pganalytics-v3 C/C++ Collector
**Version**: 3.0.0

---

## Prerequisites

### Required Packages (Ubuntu/Debian)
```bash
sudo apt-get update
sudo apt-get install -y \
    build-essential \
    cmake (>= 3.15) \
    libpq-dev \
    libssl-dev (OpenSSL 3.0+) \
    libcurl4-openssl-dev \
    zlib1g-dev \
    libzstd-dev \
    nlohmann-json3-dev
```

### Required Packages (CentOS/RHEL)
```bash
sudo yum groupinstall -y "Development Tools"
sudo yum install -y \
    cmake \
    postgresql-devel \
    openssl-devel \
    libcurl-devel \
    zlib-devel \
    libzstd-devel \
    nlohmann_json-devel
```

### Required Packages (macOS)
```bash
brew install cmake postgresql openssl curl zlib zstd nlohmann-json
```

---

## Building from Source

### 1. Development Build (Fast, with Debugging)
```bash
cd /path/to/pganalytics-v3/collector

# Create build directory
mkdir build-dev
cd build-dev

# Configure CMake
cmake -DCMAKE_BUILD_TYPE=Debug \
       -DBUILD_TESTS=ON \
       ..

# Build
make -j$(nproc)

# Binary location
./src/pganalytics --version
```

**Characteristics**:
- Compiler optimizations: -g (debug symbols)
- No LTO (Link Time Optimization)
- Faster compile, larger binary (~10-15MB)
- Good for development and testing

### 2. Production Build (Optimized)
```bash
cd /path/to/pganalytics-v3/collector

# Create build directory
mkdir build-prod
cd build-prod

# Configure CMake with optimizations
cmake -DCMAKE_BUILD_TYPE=Release \
       -DCMAKE_CXX_FLAGS_RELEASE="-O3 -march=native -flto" \
       -DBUILD_TESTS=OFF \
       ..

# Build
make -j$(nproc)

# Verify binary size
ls -lh ./src/pganalytics
# Should be 4-6MB

# Check for debugging symbols
strip --strip-all ./src/pganalytics
# Final size should be <5MB
```

**Characteristics**:
- Compiler optimizations: -O3, LTO
- No debug symbols (stripped)
- Optimized for CPU (native architecture)
- Smaller binary (~4-5MB)
- Better performance

### 3. Static Build (Portable)
```bash
cd /path/to/pganalytics-v3/collector
mkdir build-static
cd build-static

# Configure for static linking
cmake -DCMAKE_BUILD_TYPE=Release \
       -DBUILD_SHARED_LIBS=OFF \
       -DCMAKE_FIND_LIBRARY_SUFFIXES=".a" \
       ..

make -j$(nproc)

# Results in completely static binary (no library dependencies)
ldd ./src/pganalytics
# Should show "not a dynamic executable" or similar
```

**Characteristics**:
- All dependencies statically linked
- Single portable binary
- ~8-10MB size
- Can run anywhere (no dependency hell)
- Good for containerized deployments

---

## Testing the Build

### Run Unit Tests
```bash
cd build-dev
ctest --output-on-failure
```

### Manual Testing
```bash
# Show help
./src/pganalytics --help

# Run in test mode (cron)
./src/pganalytics cron

# Check for memory leaks (requires valgrind)
valgrind --leak-check=full --show-leak-kinds=all \
         --log-file=valgrind-report.txt \
         ./src/pganalytics cron
```

### Binary Verification
```bash
# Check dependencies
ldd ./src/pganalytics
# Should show libpq, libssl, libcurl, libz, libzstd, etc.

# Check binary size
du -h ./src/pganalytics

# Get file info
file ./src/pganalytics
# Should show "ELF 64-bit LSB executable"

# Check for RELRO/PIE/Canary (security)
checksec --file ./src/pganalytics
```

---

## Installation

### Option 1: System-wide Binary Installation
```bash
# Copy binary to system path
sudo cp build-prod/src/pganalytics /usr/local/bin/

# Verify
which pganalytics
pganalytics --version
```

### Option 2: DEB Package Creation
```bash
cd build-prod

# Create package structure
mkdir -p pganalytics_3.0.0/usr/local/bin
mkdir -p pganalytics_3.0.0/etc/pganalytics
mkdir -p pganalytics_3.0.0/lib/systemd/system

# Copy files
cp src/pganalytics pganalytics_3.0.0/usr/local/bin/
cp ../config.toml.sample pganalytics_3.0.0/etc/pganalytics/collector.conf
cp ../systemd/pganalytics-collector.service pganalytics_3.0.0/lib/systemd/system/

# Create DEBIAN control file
mkdir pganalytics_3.0.0/DEBIAN
cat > pganalytics_3.0.0/DEBIAN/control << 'EOF'
Package: pganalytics-collector
Version: 3.0.0
Architecture: amd64
Maintainer: pgAnalytics Team <team@pganalytics.local>
Description: Lightweight PostgreSQL collector for pgAnalytics
 Distributed PostgreSQL metrics collector for the pgAnalytics monitoring platform.
 Minimal resource footprint (<50MB memory, <1% CPU).
Depends: libpq5, libssl3, libcurl4, zlib1g, libzstd1
EOF

# Build DEB
dpkg-deb --build pganalytics_3.0.0
# Result: pganalytics_3.0.0_amd64.deb

# Install
sudo dpkg -i pganalytics_3.0.0_amd64.deb
```

### Option 3: Docker Container
```dockerfile
# Dockerfile for pganalytics-collector
FROM ubuntu:22.04 AS builder

RUN apt-get update && apt-get install -y \
    build-essential cmake libpq-dev libssl-dev libcurl4-openssl-dev \
    zlib1g-dev libzstd-dev nlohmann-json3-dev

COPY . /src
WORKDIR /src/collector
RUN mkdir build && cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTS=OFF .. && \
    make -j4 && \
    strip src/pganalytics

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    libpq5 libssl3 libcurl4 zlib1g libzstd1 ca-certificates

COPY --from=builder /src/collector/build/src/pganalytics /usr/local/bin/
COPY --from=builder /src/collector/config.toml.sample /etc/pganalytics/collector.conf

ENTRYPOINT ["/usr/local/bin/pganalytics"]
CMD ["cron"]
```

Build and push:
```bash
docker build -t pganalytics-collector:3.0.0 .
docker push your-registry/pganalytics-collector:3.0.0

# Test locally
docker run -e PG_HOST=localhost pganalytics-collector:3.0.0 cron
```

### Option 4: Kubernetes DaemonSet
```bash
# Create namespace
kubectl create namespace pganalytics

# Create ConfigMap for configuration
kubectl create configmap pganalytics-config \
  --from-file=collector.conf=/path/to/config.toml.sample \
  -n pganalytics

# Create DaemonSet
kubectl apply -f daemonset.yaml

# Verify deployment
kubectl get daemonset -n pganalytics
kubectl logs -n pganalytics daemonset/pganalytics-collector
```

---

## Configuration

### Configuration File
**Location**: `/etc/pganalytics/collector.conf`

See `config.toml.sample` for full configuration options.

### Environment Variables
```bash
# Override config file location
export PGANALYTICS_CONFIG=/custom/path/collector.conf

# Override PostgreSQL connection
export PG_HOST=localhost
export PG_PORT=5432
export PG_USER=postgres
export PG_PASSWORD=secret

# Override backend connection
export BACKEND_URL=https://analytics.example.com:9090
export BACKEND_API_KEY=your-api-key

# Set logging level
export LOG_LEVEL=info  # or debug, warn, error

# Run collector
pganalytics cron
```

---

## Deployment Checklist

- [ ] Binary built and size verified (<5MB)
- [ ] All dependencies checked (ldd output)
- [ ] Unit tests pass
- [ ] Memory leak tests pass (valgrind clean)
- [ ] Configuration created
- [ ] Backend API key obtained
- [ ] TLS certificates in place (if using mTLS)
- [ ] Systemd service enabled (on Linux)
- [ ] Log rotation configured
- [ ] Monitoring/alerting configured
- [ ] Backups tested

---

## Troubleshooting

### Build Fails: Missing Dependencies
```bash
# Check which dependencies are missing
cmake --debug-output ..

# Install missing packages
# On Ubuntu:
sudo apt-get install lib<missing>-dev
```

### Runtime: Connection Refused
```bash
# Check PostgreSQL is running
psql -h localhost -U postgres -c "SELECT 1"

# Check collector configuration
cat /etc/pganalytics/collector.conf

# Test manually
PGPASSWORD=password psql -h localhost -U postgres -c "SELECT * FROM pg_stat_database LIMIT 1"
```

### Runtime: High Memory Usage
```bash
# Check if buffer is growing
tail -f /var/log/pganalytics/collector.log | grep "buffer"

# Reduce buffer size in config
# Change: buffer_size = 1000 to buffer_size = 100

# Restart collector
sudo systemctl restart pganalytics-collector
```

### Runtime: High CPU Usage
```bash
# Check if collection interval is too short
grep "collection_interval" /etc/pganalytics/collector.conf

# Increase interval if needed
# Change: collection_interval = 5 to collection_interval = 10

# Restart collector
sudo systemctl restart pganalytics-collector
```

---

## Performance Targets

| Metric | Target | How to Measure |
|--------|--------|---|
| Binary size | <5MB | `ls -lh ./src/pganalytics` |
| Memory (idle) | <50MB | `ps aux \| grep pganalytics` |
| Memory (peak) | <100MB | Monitor during heavy load |
| CPU (idle) | <1% | `top` while idle |
| CPU (collecting) | <2% | `top` during collection |
| Startup time | <1s | `time ./pganalytics cron` |
| Metrics latency | <100ms | Check logs for timing |
| Network (compressed) | <200B/metric | Monitor network traffic |

---

## Next Steps

1. **Deploy to test environment** - DEB package on single server
2. **Run under load** - Simulate 100+ metrics per collection
3. **Monitor resource usage** - Ensure memory/CPU within targets
4. **Test failover** - Backend outage, network interruption
5. **Production deployment** - Systemd service, DaemonSet, or containers

---

## Support

For issues or questions:
- Check logs: `/var/log/pganalytics/collector.log`
- Review configuration: `/etc/pganalytics/collector.conf`
- Run in debug mode: `PGANALYTICS_LOG_LEVEL=debug pganalytics cron`
