#!/bin/bash
set -e

# Generate configuration file from environment variables
cat > /etc/pganalytics/collector.toml << EOF
# pgAnalytics Collector v3.0 Configuration
# Generated from environment variables

[collector]
id = "${COLLECTOR_ID:-collector-001}"
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
enabled = false
interval = ${COLLECTOR_INTERVAL_PG_QUERY_STATS:-60}
EOF

# Ensure pganalytics user can read the config
chown pganalytics:pganalytics /etc/pganalytics/collector.toml
chmod 640 /etc/pganalytics/collector.toml

# Switch to pganalytics user and execute the collector
exec su -s /bin/bash pganalytics -c "cd /app && $@"
