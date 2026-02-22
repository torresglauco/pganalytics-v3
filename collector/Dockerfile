# Multi-stage build for pgAnalytics Collector

# Stage 1: Build environment
FROM ubuntu:22.04 AS builder

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    pkg-config \
    libssl-dev \
    libcurl4-openssl-dev \
    libpq-dev \
    zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

# Copy source code
COPY collector /build

# Build
RUN mkdir -p build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTS=OFF && \
    make -j$(nproc)

# Stage 2: Runtime
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    libssl3 \
    libcurl4 \
    libpq5 \
    zlib1g \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/build/src/pganalytics .

# Copy config sample
COPY collector/config.toml.sample .

# Create non-root user
RUN groupadd -r pganalytics && useradd -r -g pganalytics pganalytics && \
    mkdir -p /etc/pganalytics /var/lib/pganalytics && \
    chown -R pganalytics:pganalytics /etc/pganalytics /var/lib/pganalytics /app

USER pganalytics

# Expose metrics port (future Prometheus integration)
EXPOSE 9090

# Run collector
CMD ["./pganalytics", "cron"]
