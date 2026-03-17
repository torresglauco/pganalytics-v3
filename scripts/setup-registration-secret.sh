#!/bin/bash

# Setup Registration Secret for Load Test
# This script creates a new registration secret via the backend API
# and outputs the secret value to be used by collectors

set -e

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin}"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==============================================================${NC}"
echo -e "${BLUE}pgAnalytics v3 - Setup Registration Secret${NC}"
echo -e "${BLUE}==============================================================${NC}"
echo ""

# Phase 1: Wait for backend to be healthy
echo -e "${YELLOW}Phase 1: Waiting for backend to be healthy...${NC}"

wait_for_backend() {
    local max_attempts=30
    local attempt=0

    echo -n "  Waiting for Backend... "
    while [ $attempt -lt $max_attempts ]; do
        if curl -sf "$API_BASE_URL/api/v1/health" > /dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
            return 0
        fi
        echo -n "."
        attempt=$((attempt + 1))
        sleep 2
    done
    echo -e "${RED}TIMEOUT${NC}"
    return 1
}

if ! wait_for_backend; then
    echo -e "${RED}Backend health check failed after 60 seconds. Exiting.${NC}"
    exit 1
fi

echo ""

# Phase 2: Authenticate with backend
echo -e "${YELLOW}Phase 2: Authenticating with backend...${NC}"

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

# Phase 3: Generate registration secret
echo -e "${YELLOW}Phase 3: Generating new registration secret...${NC}"

# Generate a random secret (32 bytes, base64 encoded)
RANDOM_SECRET=$(openssl rand -base64 32 | tr -d '\n')
SECRET_NAME="regression-test-$(date +%s)"
SECRET_DESCRIPTION="Regression test secret - auto-generated for load test"

echo "  Secret name: $SECRET_NAME"
echo "  Secret value: ${RANDOM_SECRET:0:20}..."
echo ""

# Create the secret via API
echo -n "  Creating secret... "
SECRET_RESPONSE=$(curl -sf -X POST \
    "$API_BASE_URL/api/v1/registration-secrets" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
        \"name\": \"$SECRET_NAME\",
        \"secret_value\": \"$RANDOM_SECRET\",
        \"description\": \"$SECRET_DESCRIPTION\",
        \"active\": true
    }" 2>/dev/null)

SECRET_ID=$(echo "$SECRET_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -n "$SECRET_ID" ]; then
    echo -e "${GREEN}OK${NC}"
    echo "  Secret ID: $SECRET_ID"
else
    echo -e "${RED}FAILED${NC}"
    echo "  Response: $SECRET_RESPONSE"
    exit 1
fi

echo ""

# Phase 4: Output the secret for use by collectors
echo -e "${YELLOW}Phase 4: Registration Secret Generated${NC}"
echo ""
echo -e "${GREEN}✅ SUCCESS!${NC}"
echo ""
echo -e "${BLUE}Export this environment variable before starting collectors:${NC}"
echo ""
echo "export REGISTRATION_SECRET=\"$RANDOM_SECRET\""
echo ""
echo -e "${BLUE}Or pass it directly to docker-compose:${NC}"
echo ""
echo "REGISTRATION_SECRET=\"$RANDOM_SECRET\" docker-compose -f docker-compose-load-test.yml up -d"
echo ""
echo -e "${BLUE}Secret Details:${NC}"
echo "  Name: $SECRET_NAME"
echo "  ID: $SECRET_ID"
echo "  Active: true"
echo "  Created: $(date)"
echo ""

# Save to file for easy reference
SECRET_FILE=".registration-secret"
echo "$RANDOM_SECRET" > "$SECRET_FILE"
echo -e "${GREEN}✓ Secret saved to: $SECRET_FILE${NC}"
echo ""

# Output just the secret for shell scripts to capture
echo "$RANDOM_SECRET"
