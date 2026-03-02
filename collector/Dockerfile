# Build stage for pgAnalytics Collector
FROM ubuntu:22.04 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    pkg-config \
    libssl-dev \
    libcurl4-openssl-dev \
    libpq-dev \
    zlib1g-dev \
    nlohmann-json3-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /build

# Copy collector source code
COPY collector /build

# Create build directory and compile
RUN rm -rf build && mkdir -p build && cd build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTS=OFF && \
    make -j$(nproc)

# Runtime stage
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    libssl3 \
    libcurl4 \
    libpq5 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create pganalytics user and group
RUN groupadd -g 1000 pganalytics && \
    useradd -u 1000 -g pganalytics -s /bin/bash -m pganalytics && \
    mkdir -p /var/lib/pganalytics /etc/pganalytics && \
    chown -R pganalytics:pganalytics /var/lib/pganalytics /etc/pganalytics

# Copy compiled binary from builder
COPY --from=builder /build/build/src/pganalytics /usr/local/bin/pganalytics-collector
RUN chmod +x /usr/local/bin/pganalytics-collector

# Copy entrypoint script
COPY collector/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh && \
    chown pganalytics:pganalytics /app/entrypoint.sh

# Copy config files
COPY collector/config.toml.sample /etc/pganalytics/config.toml.sample
RUN chown pganalytics:pganalytics /etc/pganalytics/config.toml.sample

# Set working directory
WORKDIR /var/lib/pganalytics

# Switch to non-root user
USER pganalytics

# Set environment variables
ENV LOG_LEVEL=info
ENV BACKEND_URL=http://backend:8080
ENV BACKEND_TLS_VERIFY=false
ENV REGISTRATION_SECRET=""
ENV COLLECTOR_ID=""
ENV COLLECTOR_NAME="pgAnalytics Collector"
ENV POSTGRES_HOST=localhost
ENV POSTGRES_PORT=5432
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DATABASES=postgres
ENV COLLECTION_INTERVAL=30

# Expose metrics port
EXPOSE 9090

# Start the collector
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/usr/local/bin/pganalytics-collector"]
