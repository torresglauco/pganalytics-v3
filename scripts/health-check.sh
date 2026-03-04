#!/bin/bash

##############################################################################
# pgAnalytics Health Check Script
# Verifies all services are running and healthy
##############################################################################

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

FAILED=0

check_service() {
    local name=$1
    local url=$2
    
    echo -n "Checking $name... "
    local status=$(curl -s -o /dev/null -w "%{http_code}" "$url")
    
    if [ "$status" == "200" ]; then
        echo -e "${GREEN}✓ OK${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

check_docker_service() {
    local name=$1
    local container=$2
    
    echo -n "Checking $name container... "
    
    if docker ps | grep -q "$container"; then
        echo -e "${GREEN}✓ RUNNING${NC}"
        return 0
    else
        echo -e "${RED}✗ STOPPED${NC}"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "=== pgAnalytics Health Check ==="
echo ""

# Check Docker daemon
echo "Checking Docker daemon..."
if docker ps > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Docker is running${NC}\n"
else
    echo -e "${RED}✗ Docker is not running${NC}\n"
    exit 1
fi

# Check containers
check_docker_service "PostgreSQL" "postgres"
check_docker_service "API" "api"
check_docker_service "Frontend" "frontend"

echo ""

# Check services
check_service "Backend API" "http://localhost:8080/api/v1/health"
check_service "Frontend" "http://localhost:3000"

echo ""

# Summary
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}=== All Health Checks Passed ===${NC}"
    exit 0
else
    echo -e "${RED}=== $FAILED Health Check(s) Failed ===${NC}"
    exit 1
fi
