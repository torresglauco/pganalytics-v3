#!/bin/bash
#
# Test Notification Channels - pgAnalytics v3 Phase 5 Week 2
#
# This script tests all notification channels: Slack, PagerDuty, Email, and Webhooks
# Requires environment variables to be set for each channel being tested
#
# Usage:
#   ./test_notification_channels.sh [all|slack|pagerduty|email|webhooks]
#
# Environment Variables Required:
#   SLACK_WEBHOOK_CRITICAL: Slack webhook for #critical-alerts
#   SLACK_WEBHOOK_WARNING: Slack webhook for #database-alerts
#   SLACK_WEBHOOK_INFO: Slack webhook for #database-info
#   PAGERDUTY_INTEGRATION_KEY: PagerDuty integration key
#   PAGERDUTY_ROUTING_KEY: PagerDuty routing key (v3)
#   INCIDENT_TRACKING_URL: Incident tracking webhook URL
#   INCIDENT_TRACKING_TOKEN: Bearer token for incident tracking
#   JIRA_WEBHOOK_URL: JIRA webhook URL (or use receiver endpoint)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test type
TEST_TYPE="${1:-all}"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
DATABASE="production"

# Counter for tests
TESTS_PASSED=0
TESTS_FAILED=0

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Test Slack Critical Channel
test_slack_critical() {
    log_info "Testing Slack Critical Channel (#critical-alerts)..."

    if [ -z "$SLACK_WEBHOOK_CRITICAL" ]; then
        log_warning "SLACK_WEBHOOK_CRITICAL not set, skipping"
        return 1
    fi

    local payload=$(cat <<'EOF'
{
  "title": "Test: Lock Contention Alert",
  "description": "This is a test critical alert for lock contention",
  "severity": "critical",
  "database": "production",
  "alert_name": "lock_contention_critical",
  "value": "15",
  "threshold": "10",
  "timestamp": "2026-03-03T12:00:00Z",
  "dashboard_url": "https://grafana.internal/d/lock-monitoring",
  "runbook_url": "https://docs.internal/runbooks/lock-contention.md"
}
EOF
)

    local response=$(curl -s -w "\n%{http_code}" -X POST "$SLACK_WEBHOOK_CRITICAL" \
        -H 'Content-Type: application/json' \
        -d "$payload")

    local http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ]; then
        log_success "Slack Critical channel test passed (HTTP $http_code)"
    else
        log_error "Slack Critical channel test failed (HTTP $http_code)"
        return 1
    fi
}

# Test Slack Warning Channel
test_slack_warning() {
    log_info "Testing Slack Warning Channel (#database-alerts)..."

    if [ -z "$SLACK_WEBHOOK_WARNING" ]; then
        log_warning "SLACK_WEBHOOK_WARNING not set, skipping"
        return 1
    fi

    local payload=$(cat <<'EOF'
{
  "title": "Test: High Table Bloat Warning",
  "description": "This is a test warning alert for table bloat",
  "severity": "warning",
  "database": "production",
  "alert_name": "high_table_bloat_warning",
  "value": "65",
  "threshold": "50",
  "timestamp": "2026-03-03T12:00:00Z",
  "dashboard_url": "https://grafana.internal/d/bloat-analysis",
  "runbook_url": "https://docs.internal/runbooks/high-bloat.md"
}
EOF
)

    local response=$(curl -s -w "\n%{http_code}" -X POST "$SLACK_WEBHOOK_WARNING" \
        -H 'Content-Type: application/json' \
        -d "$payload")

    local http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ]; then
        log_success "Slack Warning channel test passed (HTTP $http_code)"
    else
        log_error "Slack Warning channel test failed (HTTP $http_code)"
        return 1
    fi
}

# Test Slack Info Channel
test_slack_info() {
    log_info "Testing Slack Info Channel (#database-info)..."

    if [ -z "$SLACK_WEBHOOK_INFO" ]; then
        log_warning "SLACK_WEBHOOK_INFO not set, skipping"
        return 1
    fi

    local payload=$(cat <<'EOF'
{
  "title": "Test: Schema Change Detected",
  "description": "This is a test info alert for schema changes",
  "severity": "info",
  "database": "production",
  "alert_name": "schema_growth_info",
  "value": "1543",
  "timestamp": "2026-03-03T12:00:00Z",
  "dashboard_url": "https://grafana.internal/d/schema-tracking"
}
EOF
)

    local response=$(curl -s -w "\n%{http_code}" -X POST "$SLACK_WEBHOOK_INFO" \
        -H 'Content-Type: application/json' \
        -d "$payload")

    local http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ]; then
        log_success "Slack Info channel test passed (HTTP $http_code)"
    else
        log_error "Slack Info channel test failed (HTTP $http_code)"
        return 1
    fi
}

# Test PagerDuty Integration (v3)
test_pagerduty() {
    log_info "Testing PagerDuty Integration..."

    if [ -z "$PAGERDUTY_ROUTING_KEY" ]; then
        log_warning "PAGERDUTY_ROUTING_KEY not set, skipping"
        return 1
    fi

    local payload=$(cat <<EOF
{
  "routing_key": "$PAGERDUTY_ROUTING_KEY",
  "event_action": "trigger",
  "payload": {
    "summary": "Test: Critical - Lock Contention Alert",
    "severity": "critical",
    "source": "pgAnalytics",
    "custom_details": {
      "database": "production",
      "alert_name": "lock_contention_critical",
      "value": "15",
      "threshold": "10",
      "timestamp": "$TIMESTAMP",
      "dashboard_url": "https://grafana.internal/d/lock-monitoring",
      "runbook_url": "https://docs.internal/runbooks/lock-contention.md"
    }
  }
}
EOF
)

    local response=$(curl -s -w "\n%{http_code}" -X POST "https://events.pagerduty.com/v2/enqueue" \
        -H 'Content-Type: application/json' \
        -d "$payload")

    local http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "202" ] || [ "$http_code" = "200" ]; then
        log_success "PagerDuty integration test passed (HTTP $http_code)"
    else
        log_error "PagerDuty integration test failed (HTTP $http_code)"
        echo "$response"
        return 1
    fi
}

# Test Incident Tracking Webhook
test_incident_webhook() {
    log_info "Testing Incident Tracking Webhook..."

    if [ -z "$INCIDENT_TRACKING_URL" ]; then
        log_warning "INCIDENT_TRACKING_URL not set, skipping"
        return 1
    fi

    local payload=$(cat <<'EOF'
{
  "title": "Test: Blocking Transaction Alert",
  "description": "This is a test alert for blocking transactions",
  "severity": "critical",
  "database": "production",
  "alert_name": "blocking_transaction_critical",
  "value": "450",
  "threshold": "300",
  "timestamp": "2026-03-03T12:00:00Z",
  "dashboard_url": "https://grafana.internal/d/lock-monitoring",
  "runbook_url": "https://docs.internal/runbooks/lock-contention.md"
}
EOF
)

    local auth_header=""
    if [ -n "$INCIDENT_TRACKING_TOKEN" ]; then
        auth_header="-H 'Authorization: Bearer $INCIDENT_TRACKING_TOKEN'"
    fi

    local response=$(curl -s -w "\n%{http_code}" -X POST "$INCIDENT_TRACKING_URL" \
        -H 'Content-Type: application/json' \
        $auth_header \
        -d "$payload")

    local http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        log_success "Incident tracking webhook test passed (HTTP $http_code)"
    else
        log_error "Incident tracking webhook test failed (HTTP $http_code)"
        return 1
    fi
}

# Test JIRA Webhook
test_jira_webhook() {
    log_info "Testing JIRA Auto-Ticket Webhook..."

    if [ -z "$JIRA_WEBHOOK_URL" ]; then
        log_warning "JIRA_WEBHOOK_URL not set, skipping"
        return 1
    fi

    local payload=$(cat <<'EOF'
{
  "title": "Test: Low Cache Hit Ratio Warning",
  "description": "This is a test alert for low cache hit ratio",
  "severity": "warning",
  "database": "production",
  "alert_name": "low_cache_hit_ratio_warning",
  "value": "72",
  "threshold": "80",
  "timestamp": "2026-03-03T12:00:00Z",
  "dashboard_url": "https://grafana.internal/d/cache-performance",
  "runbook_url": "https://docs.internal/runbooks/cache-hit-ratio.md"
}
EOF
)

    local response=$(curl -s -w "\n%{http_code}" -X POST "$JIRA_WEBHOOK_URL" \
        -H 'Content-Type: application/json' \
        -d "$payload")

    local http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        log_success "JIRA webhook test passed (HTTP $http_code)"
    else
        log_error "JIRA webhook test failed (HTTP $http_code)"
        return 1
    fi
}

# Main test logic
main() {
    echo "=========================================="
    echo "pgAnalytics Notification Channel Tests"
    echo "=========================================="
    echo ""

    case "$TEST_TYPE" in
        all)
            test_slack_critical
            test_slack_warning
            test_slack_info
            test_pagerduty
            test_incident_webhook
            test_jira_webhook
            ;;
        slack)
            test_slack_critical
            test_slack_warning
            test_slack_info
            ;;
        pagerduty)
            test_pagerduty
            ;;
        incident)
            test_incident_webhook
            ;;
        jira)
            test_jira_webhook
            ;;
        *)
            log_error "Unknown test type: $TEST_TYPE"
            echo "Usage: $0 [all|slack|pagerduty|incident|jira]"
            exit 1
            ;;
    esac

    echo ""
    echo "=========================================="
    echo "Test Summary"
    echo "=========================================="
    log_success "Passed: $TESTS_PASSED"
    if [ $TESTS_FAILED -gt 0 ]; then
        log_error "Failed: $TESTS_FAILED"
    fi
    echo ""

    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All tests passed!"
        exit 0
    else
        log_error "Some tests failed"
        exit 1
    fi
}

main
