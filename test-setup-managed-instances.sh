#!/bin/bash

# Full Regression Test: Setup Managed Instances
# This script registers 20 of the 40 collectors as "Managed Instances"
# for comprehensive testing of the backend's managed instance functionality

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin}"
MANAGED_INSTANCE_COUNT=20
MANAGED_INSTANCE_START_ID=1
MANAGED_INSTANCE_END_ID=20

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==============================================================${NC}"
echo -e "${BLUE}pgAnalytics v3 - Regression Test: Managed Instances Setup${NC}"
echo -e "${BLUE}==============================================================${NC}"
echo ""

# Phase 1: Wait for all services to be healthy
echo -e "${YELLOW}Phase 1: Waiting for all services to be healthy...${NC}"

wait_for_health() {
    local service_name=$1
    local url=$2
    local max_attempts=30
    local attempt=0

    echo -n "  Waiting for $service_name... "
    while [ $attempt -lt $max_attempts ]; do
        if curl -sf "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
            return 0
        fi
        echo -n "."
        attempt=$((attempt + 1))
        sleep 2
    done
    echo -e "${RED}FAILED${NC}"
    return 1
}

# Wait for backend health endpoint
if ! wait_for_health "Backend" "$API_BASE_URL/api/v1/health"; then
    echo -e "${RED}Backend health check failed after 60 seconds. Exiting.${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Phase 2: Authenticating with backend...${NC}"

# Get JWT token from backend
AUTH_RESPONSE=$(curl -sf -X POST \
    "$API_BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}")

TOKEN=$(echo "$AUTH_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to authenticate with backend. Response: $AUTH_RESPONSE${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Authenticated successfully${NC}"
echo ""

# Phase 3: Register managed instances
echo -e "${YELLOW}Phase 3: Registering $MANAGED_INSTANCE_COUNT managed instances...${NC}"

REGISTERED_COUNT=0
FAILED_COUNT=0
MANAGED_INSTANCES=()

for i in $(seq -w $MANAGED_INSTANCE_START_ID $MANAGED_INSTANCE_END_ID); do
    COLLECTOR_NUM=$(printf "%03d" $i)
    INSTANCE_NAME="Target PostgreSQL $COLLECTOR_NUM"
    INSTANCE_ENDPOINT="target-postgres-$COLLECTOR_NUM"
    INSTANCE_PORT=5432

    echo -n "  Registering managed instance $i/20 ($INSTANCE_NAME)... "

    # Create managed instance via API
    RESPONSE=$(curl -sf -X POST \
        "$API_BASE_URL/api/v1/rds-instances" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"$INSTANCE_NAME\",
            \"endpoint\": \"$INSTANCE_ENDPOINT\",
            \"port\": $INSTANCE_PORT,
            \"master_username\": \"postgres\",
            \"master_password\": \"pganalytics\",
            \"ssl_enabled\": false,
            \"ssl_mode\": \"disable\",
            \"monitoring_interval\": 60,
            \"connection_timeout\": 30,
            \"aws_region\": \"us-east-1\",
            \"environment\": \"test\"
        }" 2>/dev/null)

    # Extract ID from response
    INSTANCE_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

    if [ -n "$INSTANCE_ID" ] && [ "$INSTANCE_ID" -gt 0 ]; then
        echo -e "${GREEN}OK (ID: $INSTANCE_ID)${NC}"
        MANAGED_INSTANCES+=("$INSTANCE_ID")
        REGISTERED_COUNT=$((REGISTERED_COUNT + 1))
    else
        echo -e "${RED}FAILED${NC}"
        echo "    Response: $RESPONSE"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi
done

echo ""
echo -e "${BLUE}Registration Summary:${NC}"
echo "  Total Registered: $REGISTERED_COUNT / $MANAGED_INSTANCE_COUNT"
echo "  Total Failed: $FAILED_COUNT / $MANAGED_INSTANCE_COUNT"

if [ $FAILED_COUNT -gt 0 ]; then
    echo -e "${YELLOW}Warning: Some managed instances failed to register.${NC}"
fi

echo ""

# Phase 4: Verify managed instances were created
echo -e "${YELLOW}Phase 4: Verifying managed instances...${NC}"

LIST_RESPONSE=$(curl -sf -X GET \
    "$API_BASE_URL/api/v1/rds-instances" \
    -H "Authorization: Bearer $TOKEN" \
    2>/dev/null)

TOTAL_INSTANCES=$(echo "$LIST_RESPONSE" | grep -o '"id":[0-9]*' | wc -l)

echo -e "${GREEN}✓ Total managed instances in backend: $TOTAL_INSTANCES${NC}"

# Phase 5: Get registration secret statistics
echo ""
echo -e "${YELLOW}Phase 5: Checking registration secret statistics...${NC}"

# Query the database directly for registration stats
STATS=$(curl -sf -X GET \
    "$API_BASE_URL/api/v1/collectors" \
    -H "Authorization: Bearer $TOKEN" \
    2>/dev/null)

COLLECTOR_COUNT=$(echo "$STATS" | grep -o '"id":"col_[0-9]*' | wc -l)

echo -e "${BLUE}Collector Statistics:${NC}"
echo "  Total collectors registered: $COLLECTOR_COUNT / 40"

echo ""

# Phase 6: Generate summary report
echo -e "${YELLOW}Phase 6: Generating summary report...${NC}"

REPORT_FILE="regression-test-setup-report.txt"

cat > "$REPORT_FILE" << EOF
================================================================================
pgAnalytics v3 - Regression Test: Managed Instances Setup Report
Generated: $(date)
================================================================================

TEST CONFIGURATION:
  API Base URL: $API_BASE_URL
  Managed Instances to Register: $MANAGED_INSTANCE_COUNT
  Instance ID Range: 001-020

REGISTRATION RESULTS:
  Total Attempted: $MANAGED_INSTANCE_COUNT
  Successful: $REGISTERED_COUNT
  Failed: $FAILED_COUNT

VERIFICATION RESULTS:
  Total Managed Instances in Backend: $TOTAL_INSTANCES
  Total Collectors Registered: $COLLECTOR_COUNT / 40

MANAGED INSTANCE DETAILS:
EOF

for idx in "${!MANAGED_INSTANCES[@]}"; do
    INSTANCE_ID=${MANAGED_INSTANCES[$idx]}
    INSTANCE_NUM=$((idx + 1))
    printf "  Instance %02d: ID=%d\n" $INSTANCE_NUM $INSTANCE_ID >> "$REPORT_FILE"
done

cat >> "$REPORT_FILE" << EOF

EXPECTED FINAL STATE:
  ✓ 40 collectors registered (cols 001-040)
  ✓ 20 collectors also registered as managed instances (001-020)
  ✓ 20 collectors as regular collectors only (021-040)
  ✓ Registration secret shows 40 total registrations
  ✓ All collectors collecting metrics with monotonic increase

NEXT STEPS:
  1. Wait 5 minutes for collectors to send initial metrics
  2. Run: ./verify-regression-tests.sh
  3. Check Frontend UI at http://localhost:4000
  4. Monitor logs: docker-compose -f docker-compose-load-test.yml logs -f

NOTES:
  - All managed instances are configured with ssl_mode="disable" for testing
  - Collector IDs persist in /var/lib/pganalytics/collector.id volumes
  - Metrics collection interval: 60 seconds
  - Collection should be monotonically increasing
EOF

echo -e "${GREEN}✓ Report generated: $REPORT_FILE${NC}"
echo ""

# Display summary
echo -e "${BLUE}==============================================================${NC}"
echo -e "${BLUE}Setup Complete!${NC}"
echo -e "${BLUE}==============================================================${NC}"
echo ""
echo "Summary:"
echo "  Managed Instances Created: $REGISTERED_COUNT / $MANAGED_INSTANCE_COUNT"
echo "  Total Collectors: $COLLECTOR_COUNT / 40"
echo ""
echo "Report file: $REPORT_FILE"
echo ""
echo "Next: Run './verify-regression-tests.sh' to validate the system"
echo ""
