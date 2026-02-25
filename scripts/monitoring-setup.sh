#!/bin/bash

##############################################################################
# pgAnalytics v3.2.0 - Monitoring & Alerting Setup Script
# Purpose: Configure Prometheus scraping, AlertManager rules, and dashboards
# Usage: ./monitoring-setup.sh [--prometheus-host <host>] [--slack-webhook <url>]
##############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROMETHEUS_HOST="${PROMETHEUS_HOST:-localhost}"
PROMETHEUS_PORT="${PROMETHEUS_PORT:-9090}"
ALERTMANAGER_HOST="${ALERTMANAGER_HOST:-localhost}"
ALERTMANAGER_PORT="${ALERTMANAGER_PORT:-9093}"
API_SERVER="${API_SERVER:-http://localhost:8080}"
SLACK_WEBHOOK="${SLACK_WEBHOOK:-}"
PAGERDUTY_KEY="${PAGERDUTY_KEY:-}"
PROMETHEUS_CONFIG_DIR="${PROMETHEUS_CONFIG_DIR:-/etc/prometheus}"
ALERTMANAGER_CONFIG_DIR="${ALERTMANAGER_CONFIG_DIR:-/etc/alertmanager}"

##############################################################################
# Helper Functions
##############################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "Required command not found: $1"
    fi
}

##############################################################################
# Prometheus Configuration
##############################################################################

create_prometheus_scrape_config() {
    log_info "Creating Prometheus scrape configuration..."

    cat > /tmp/prometheus_scrape.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'pganalytics'
    environment: 'production'

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - localhost:9093

# Load alert rules
rule_files:
  - /etc/prometheus/pganalytics_rules.yml
  - /etc/prometheus/alert_rules.yml

scrape_configs:
  # pgAnalytics API Server
  - job_name: 'pganalytics-api'
    metrics_path: '/metrics'
    static_configs:
      - targets: ['PROMETHEUS_PLACEHOLDER:PROMETHEUS_PORT_PLACEHOLDER']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance

  # PostgreSQL via postgres_exporter
  - job_name: 'postgresql'
    static_configs:
      - targets: ['localhost:9187']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance

  # Node Exporter (system metrics)
  - job_name: 'node'
    static_configs:
      - targets: ['localhost:9100']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance

  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF

    # Replace placeholders
    sed -i "s|PROMETHEUS_PLACEHOLDER|${API_SERVER#http://}|g" /tmp/prometheus_scrape.yml
    sed -i "s|PROMETHEUS_PORT_PLACEHOLDER|${PROMETHEUS_PORT}|g" /tmp/prometheus_scrape.yml

    sudo cp /tmp/prometheus_scrape.yml "$PROMETHEUS_CONFIG_DIR/pganalytics_scrape.yml" 2>/dev/null || \
        log_warning "Cannot write to $PROMETHEUS_CONFIG_DIR (may require manual setup)"

    rm -f /tmp/prometheus_scrape.yml
    log_success "Prometheus scrape configuration created"
}

##############################################################################
# Alert Rules Configuration
##############################################################################

create_alert_rules() {
    log_info "Creating Prometheus alert rules..."

    cat > /tmp/pganalytics_rules.yml << 'EOF'
groups:
  - name: pganalytics_alerts
    interval: 30s
    rules:
      # Authentication Alerts
      - alert: HighFailedAuthAttempts
        expr: rate(pganalytics_auth_failures_total[5m]) > 5
        for: 5m
        labels:
          severity: warning
          service: pganalytics
        annotations:
          summary: "High failed authentication attempts"
          description: "Failed auth attempts > 5 per minute (current: {{ $value | humanize }})"

      # Rate Limiting Alerts
      - alert: HighRate429Responses
        expr: rate(pganalytics_http_429_total[5m]) > 100
        for: 5m
        labels:
          severity: warning
          service: pganalytics
        annotations:
          summary: "High rate limit responses"
          description: "HTTP 429 responses > 100 per minute (current: {{ $value | humanize }})"

      # Collector Alerts
      - alert: CollectorPushFailures
        expr: rate(pganalytics_collector_push_failures_total[10m]) / rate(pganalytics_collector_pushes_total[10m]) > 0.01
        for: 10m
        labels:
          severity: critical
          service: pganalytics
        annotations:
          summary: "High collector push failure rate"
          description: "Collector push failures > 1% (current: {{ $value | humanizePercentage }})"

      # Database Connection Alerts
      - alert: DatabaseConnectionErrors
        expr: rate(pganalytics_db_connection_errors_total[5m]) > 5
        for: 5m
        labels:
          severity: critical
          service: pganalytics
        annotations:
          summary: "High database connection errors"
          description: "DB connection errors > 5 per minute (current: {{ $value | humanize }})"

      # API Response Time Alerts
      - alert: SlowAPIResponse
        expr: histogram_quantile(0.95, rate(pganalytics_http_request_duration_seconds_bucket[5m])) > 1
        for: 10m
        labels:
          severity: warning
          service: pganalytics
        annotations:
          summary: "Slow API response times"
          description: "p95 response time > 1s (current: {{ $value | humanizeDuration }})"

      # Memory Alerts
      - alert: HighMemoryUsage
        expr: process_resident_memory_bytes / (1024 * 1024 * 1024) > 0.5
        for: 15m
        labels:
          severity: warning
          service: pganalytics
        annotations:
          summary: "High memory usage"
          description: "Memory usage > 500MB (current: {{ $value | humanize }} GB)"

      # Disk Space Alerts
      - alert: LowDiskSpace
        expr: node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"} < 0.1
        for: 15m
        labels:
          severity: critical
          service: pganalytics
        annotations:
          summary: "Low disk space"
          description: "Free disk space < 10% (current: {{ $value | humanizePercentage }})"

      # Service Availability
      - alert: ServiceDown
        expr: up{job="pganalytics-api"} == 0
        for: 5m
        labels:
          severity: critical
          service: pganalytics
        annotations:
          summary: "pgAnalytics API service is down"
          description: "pgAnalytics API is not responding"

      # PostgreSQL Alerts
      - alert: PostgreSQLDown
        expr: up{job="postgresql"} == 0
        for: 5m
        labels:
          severity: critical
          service: postgresql
        annotations:
          summary: "PostgreSQL is down"
          description: "PostgreSQL server is not responding"

      # Cache Hit Rate
      - alert: LowCacheHitRate
        expr: (rate(pganalytics_cache_hits_total[5m])) / (rate(pganalytics_cache_hits_total[5m]) + rate(pganalytics_cache_misses_total[5m])) < 0.7
        for: 15m
        labels:
          severity: warning
          service: pganalytics
        annotations:
          summary: "Low cache hit rate"
          description: "Cache hit rate < 70% (current: {{ $value | humanizePercentage }})"
EOF

    sudo cp /tmp/pganalytics_rules.yml "$PROMETHEUS_CONFIG_DIR/pganalytics_rules.yml" 2>/dev/null || \
        log_warning "Cannot write to $PROMETHEUS_CONFIG_DIR"

    rm -f /tmp/pganalytics_rules.yml
    log_success "Prometheus alert rules created"
}

##############################################################################
# AlertManager Configuration
##############################################################################

create_alertmanager_config() {
    log_info "Creating AlertManager configuration..."

    # Build routes based on available integrations
    local routes=""
    local receivers="- name: 'null'"

    if [ -n "$SLACK_WEBHOOK" ]; then
        routes="$routes
      - match:
          severity: critical
        receiver: 'slack-critical'
        continue: true"

        receivers="$receivers
  - name: 'slack-critical'
    slack_configs:
      - api_url: '$SLACK_WEBHOOK'
        channel: '#pganalytics-critical'
        title: '[{{ .GroupLabels.severity | toUpper }}] {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true"
    fi

    if [ -n "$PAGERDUTY_KEY" ]; then
        routes="$routes
      - match:
          severity: critical
        receiver: 'pagerduty'
        continue: true"

        receivers="$receivers
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: '$PAGERDUTY_KEY'
        description: '{{ .GroupLabels.alertname }}: {{ (index .Alerts 0).Annotations.summary }}'
        client_url: 'https://alertmanager.example.com'"
    fi

    cat > /tmp/alertmanager.yml << EOF
global:
  resolve_timeout: 5m

route:
  receiver: 'null'
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  routes:
    - match:
        severity: critical
      receiver: 'critical-notifications'
      continue: true
    $routes

receivers:
  $receivers
  - name: 'critical-notifications'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK}'
        channel: '#pganalytics-alerts'
        title: '{{ .GroupLabels.severity | toUpper }}: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}\n{{ .Annotations.description }}\n{{ end }}'
        send_resolved: true
inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF

    sudo cp /tmp/alertmanager.yml "$ALERTMANAGER_CONFIG_DIR/config.yml" 2>/dev/null || \
        log_warning "Cannot write to $ALERTMANAGER_CONFIG_DIR"

    rm -f /tmp/alertmanager.yml
    log_success "AlertManager configuration created"
}

##############################################################################
# Integration Setup
##############################################################################

verify_slack_webhook() {
    log_info "Verifying Slack webhook..."

    if [ -z "$SLACK_WEBHOOK" ]; then
        log_warning "Slack webhook not configured"
        return
    fi

    # Test webhook with curl
    local test_payload='{
      "text": "pgAnalytics Monitoring Setup - Test Message",
      "attachments": [{
        "color": "good",
        "title": "Monitoring Configuration",
        "text": "Slack integration is working correctly",
        "ts": '$(date +%s)'
      }]
    }'

    if curl -X POST -H 'Content-type: application/json' \
        --data "$test_payload" \
        "$SLACK_WEBHOOK" > /dev/null 2>&1; then
        log_success "Slack webhook verified"
    else
        log_warning "Slack webhook test failed"
    fi
}

##############################################################################
# Service Restart & Verification
##############################################################################

restart_prometheus() {
    log_info "Restarting Prometheus..."

    if systemctl is-active --quiet prometheus; then
        systemctl restart prometheus
        sleep 2
        if systemctl is-active --quiet prometheus; then
            log_success "Prometheus restarted successfully"
        else
            log_error "Prometheus failed to start"
        fi
    else
        log_warning "Prometheus service not found"
    fi
}

restart_alertmanager() {
    log_info "Restarting AlertManager..."

    if systemctl is-active --quiet alertmanager; then
        systemctl restart alertmanager
        sleep 2
        if systemctl is-active --quiet alertmanager; then
            log_success "AlertManager restarted successfully"
        else
            log_error "AlertManager failed to start"
        fi
    else
        log_warning "AlertManager service not found"
    fi
}

verify_prometheus_targets() {
    log_info "Verifying Prometheus targets..."

    # Wait for Prometheus to be ready
    sleep 3

    local prometheus_url="http://$PROMETHEUS_HOST:$PROMETHEUS_PORT"
    local targets_json

    targets_json=$(curl -s "$prometheus_url/api/v1/targets" 2>/dev/null || echo "{}")

    if echo "$targets_json" | grep -q "pganalytics-api"; then
        log_success "Prometheus targets configured"
    else
        log_warning "Could not verify Prometheus targets"
    fi
}

##############################################################################
# Dashboard Setup
##############################################################################

create_grafana_dashboard_links() {
    log_info "Creating Grafana dashboard documentation..."

    cat > /tmp/grafana_dashboards.md << 'EOF'
# pgAnalytics v3.2.0 - Grafana Dashboards

## Dashboard List

### 1. Replication Metrics Overview
- Description: High-level overview of replication status
- Panels: Replica count, replication lag, sync status
- Datasource: PostgreSQL
- Refresh: 30 seconds

### 2. Streaming Replication Status
- Description: Detailed streaming replication metrics
- Panels: Connected replicas, LSN positions, replication state
- Datasource: PostgreSQL
- Refresh: 30 seconds

### 3. WAL Activity Dashboard
- Description: WAL generation and archiving status
- Panels: WAL generation rate, archive success rate, archived bytes
- Datasource: PostgreSQL
- Refresh: 60 seconds

### 4. Replication Slots Monitoring
- Description: Replication slot utilization and health
- Panels: Slot count, retention bytes, logical slot activity
- Datasource: PostgreSQL
- Refresh: 60 seconds

### 5. XID Wraparound Risk
- Description: Transaction ID wraparound risk assessment
- Panels: XID position, percent to wraparound, estimated time
- Datasource: PostgreSQL
- Refresh: 300 seconds

### 6. Collection Performance
- Description: Collector performance and metrics throughput
- Panels: Collection success rate, push latency, metrics/sec
- Datasource: Prometheus (pgAnalytics metrics)
- Refresh: 30 seconds

### 7. API Server Health
- Description: API server performance and health
- Panels: Request rate, response time p95, error rate, active connections
- Datasource: Prometheus (pgAnalytics metrics)
- Refresh: 30 seconds

### 8. Collector Status
- Description: Collector system health and connectivity
- Panels: Active collectors, last update time, error counts
- Datasource: PostgreSQL
- Refresh: 60 seconds

### 9. System Resources
- Description: System CPU, memory, disk utilization
- Panels: CPU usage, memory usage, disk I/O, network bandwidth
- Datasource: Prometheus (Node Exporter)
- Refresh: 30 seconds

## Import Instructions

1. Log into Grafana: https://grafana.example.com
2. Go to: Settings > Data sources
3. Ensure PostgreSQL datasource is configured
4. Go to: Dashboards > Import
5. Copy dashboard JSON from: /etc/pganalytics/dashboards/
6. Paste into Grafana import dialog
7. Select PostgreSQL datasource
8. Click Import

## Alert Channels

### Slack
- Channel: #pganalytics-alerts
- Severity: All alerts
- Status: Enabled

### PagerDuty (Critical Only)
- Service: pgAnalytics Production
- Severity: Critical
- Status: Enabled

## Dashboard Variables (Templating)

- datasource: Select Prometheus/PostgreSQL datasource
- interval: Scrape interval (15s, 30s, 1m, 5m)
- environment: Filter by environment (prod, staging, dev)

EOF

    sudo cp /tmp/grafana_dashboards.md /etc/pganalytics/GRAFANA_DASHBOARDS.md 2>/dev/null || true
    rm -f /tmp/grafana_dashboards.md

    log_success "Grafana dashboard documentation created"
}

##############################################################################
# Health Check
##############################################################################

health_check() {
    log_info "Performing health checks..."
    echo ""

    # Check Prometheus
    echo -n "Prometheus: "
    if curl -s "http://$PROMETHEUS_HOST:$PROMETHEUS_PORT/-/healthy" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Healthy${NC}"
    else
        echo -e "${RED}✗ Unreachable${NC}"
    fi

    # Check AlertManager
    echo -n "AlertManager: "
    if curl -s "http://$ALERTMANAGER_HOST:$ALERTMANAGER_PORT/-/healthy" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Healthy${NC}"
    else
        echo -e "${RED}✗ Unreachable${NC}"
    fi

    # Check API server metrics endpoint
    echo -n "API Metrics Endpoint: "
    if curl -s "$API_SERVER/metrics" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Accessible${NC}"
    else
        echo -e "${RED}✗ Not responding${NC}"
    fi

    echo ""
}

##############################################################################
# Main Execution
##############################################################################

main() {
    log_info "Starting pgAnalytics monitoring and alerting setup"
    echo ""

    # Check required tools
    check_command "curl"
    check_command "sed"
    echo ""

    # Create configurations
    create_prometheus_scrape_config
    create_alert_rules
    create_alertmanager_config
    echo ""

    # Setup integrations
    verify_slack_webhook
    echo ""

    # Create dashboards documentation
    create_grafana_dashboard_links
    echo ""

    # Restart services
    restart_prometheus
    restart_alertmanager
    echo ""

    # Verification
    verify_prometheus_targets
    echo ""

    # Health check
    health_check
    echo ""

    echo -e "${BLUE}============================================${NC}"
    echo -e "${GREEN}Monitoring Setup Completed${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""
    echo "Prometheus: http://$PROMETHEUS_HOST:$PROMETHEUS_PORT"
    echo "AlertManager: http://$ALERTMANAGER_HOST:$ALERTMANAGER_PORT"
    echo ""
    echo "Alert Rules: $PROMETHEUS_CONFIG_DIR/pganalytics_rules.yml"
    echo "AlertManager Config: $ALERTMANAGER_CONFIG_DIR/config.yml"
    echo ""
    log_success "Monitoring setup completed successfully"
    echo ""
    echo "Next steps:"
    echo "1. Configure Grafana datasources"
    echo "2. Import dashboards"
    echo "3. Test alert notifications"
    echo "4. Verify metrics are being collected"
}

# Run main function
main "$@"
