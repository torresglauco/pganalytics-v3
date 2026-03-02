#!/bin/bash
set -e

# Check if collector has been previously registered and has a persisted ID
PERSISTED_COLLECTOR_ID=""
if [ -f /var/lib/pganalytics/collector.id ]; then
    PERSISTED_COLLECTOR_ID=$(cat /var/lib/pganalytics/collector.id)
    echo "Found persisted collector ID: $PERSISTED_COLLECTOR_ID"
fi

# Use persisted ID if available, otherwise use the configured ID
COLLECTOR_ID_TO_USE="${PERSISTED_COLLECTOR_ID:-${COLLECTOR_ID:-collector-001}}"

# Generate configuration file from environment variables
cat > /etc/pganalytics/collector.toml << EOF
# pgAnalytics Collector v3.0 Configuration
# Generated from environment variables

[collector]
id = "$COLLECTOR_ID_TO_USE"
hostname = "${COLLECTOR_NAME:-localhost}"
interval = ${COLLECTION_INTERVAL:-60}
push_interval = ${PUSH_INTERVAL:-60}
config_pull_interval = ${CONFIG_PULL_INTERVAL:-300}

[backend]
url = "${BACKEND_URL:-http://backend:8080}"

[postgres]
host = "${POSTGRES_HOST:-postgres}"
port = ${POSTGRES_PORT:-5432}
user = "${POSTGRES_USER:-postgres}"
password = "${POSTGRES_PASSWORD:-}"
database = "${POSTGRES_DB:-postgres}"
databases = "${POSTGRES_DATABASES:-postgres}"

[tls]
verify = ${BACKEND_TLS_VERIFY:-false}
cert_file = "${TLS_CERT_FILE:-/etc/pganalytics/collector.crt}"
key_file = "${TLS_KEY_FILE:-/etc/pganalytics/collector.key}"

[pg_stats]
enabled = false
interval = ${COLLECTOR_INTERVAL_PG_STATS:-60}

[sysstat]
enabled = true
interval = ${COLLECTOR_INTERVAL_SYSSTAT:-60}

[pg_log]
enabled = true
interval = ${COLLECTOR_INTERVAL_PG_LOG:-300}

[disk_usage]
enabled = true
interval = ${COLLECTOR_INTERVAL_DISK_USAGE:-300}

[pg_query_stats]
enabled = true
interval = ${COLLECTOR_INTERVAL_PG_QUERY_STATS:-60}

[registration]
auto_register = ${AUTO_REGISTER:-false}
secret = "${REGISTRATION_SECRET:-}"
EOF

# Ensure pganalytics user can read the config
chown pganalytics:pganalytics /etc/pganalytics/collector.toml
chmod 640 /etc/pganalytics/collector.toml

# Auto-register only if:
# 1. AUTO_REGISTER is enabled
# 2. REGISTRATION_SECRET is provided
# 3. Collector hasn't been registered yet (no persisted ID)
if [ "${AUTO_REGISTER}" = "true" ] && [ -n "${REGISTRATION_SECRET}" ] && [ -z "${PERSISTED_COLLECTOR_ID}" ]; then
    echo "Auto-registering collector for the first time..."
    /usr/local/bin/pganalytics-collector register || {
        echo "Warning: Auto-registration failed, continuing anyway..."
    }
elif [ -n "${PERSISTED_COLLECTOR_ID}" ]; then
    echo "Collector already registered with ID: $PERSISTED_COLLECTOR_ID (skipping registration)"
fi

# Execute the collector
# If only the binary path is provided (from CMD), run in normal collection mode
# Otherwise, pass arguments to the binary
if [ $# -eq 1 ] && [ "$1" = "/usr/local/bin/pganalytics-collector" ]; then
    # Standard mode - just run the collector
    exec "$1"
elif [ $# -eq 0 ]; then
    # No arguments provided
    exec /usr/local/bin/pganalytics-collector
else
    # Arguments provided - pass them through
    exec "$@"
fi
