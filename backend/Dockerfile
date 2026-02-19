# Multi-stage build for pgAnalytics Backend

# Stage 1: Build
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend ./backend

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o pganalytics-api \
    ./backend/cmd/pganalytics-api

# Stage 2: Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl postgresql-client bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/pganalytics-api .

# Copy migrations
COPY backend/migrations ./migrations

# Create non-root user
RUN addgroup -g 1000 pganalytics && \
    adduser -D -u 1000 -G pganalytics pganalytics && \
    chown -R pganalytics:pganalytics /app

USER pganalytics

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=3 --start-period=10s \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Expose port
EXPOSE 8080

# Run the application
CMD ["./pganalytics-api"]
