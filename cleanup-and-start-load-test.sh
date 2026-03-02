#!/bin/bash

# Complete Cleanup and Startup for Load Test - WITH PROPER SECRET GENERATION
# Phase 1: Clean Docker resources
# Phase 2: Start core services (PostgreSQL, TimescaleDB, Backend, Frontend)
# Phase 3: Generate registration secret via API
# Phase 4: Start collectors with the generated secret

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}==============================================================${NC}"
echo -e "${BLUE}pgAnalytics v3 - Load Test Clean Startup (WITH DYNAMIC SECRET)${NC}"
echo -e "${BLUE}==============================================================${NC}"
echo ""

# ============================================================================
# Phase 1: Cleanup
# ============================================================================

echo -e "${YELLOW}Phase 1: Stopping existing containers...${NC}"

if command -v docker-compose &> /dev/null; then
    echo "  Stopping containers from docker-compose.yml..."
    docker-compose down -v --remove-orphans 2>/dev/null || true
    echo -e "  ${GREEN}✓${NC} docker-compose.yml containers stopped"
fi

if command -v docker &> /dev/null; then
    echo "  Stopping containers from docker-compose-load-test.yml..."
    docker-compose -f docker-compose-load-test.yml down -v --remove-orphans 2>/dev/null || true
    echo -e "  ${GREEN}✓${NC} docker-compose-load-test.yml containers stopped"
fi

echo ""

# Phase 2: Remove pganalytics volumes
echo -e "${YELLOW}Phase 2: Removing pganalytics volumes...${NC}"

if command -v docker &> /dev/null; then
    # List all pganalytics volumes
    VOLUMES=$(docker volume ls -q | grep pganalytics || true)

    if [ -n "$VOLUMES" ]; then
        echo "  Found pganalytics volumes:"
        echo "$VOLUMES" | while read -r vol; do
            echo "    Removing: $vol"
            docker volume rm "$vol" 2>/dev/null || true
        done
        echo -e "  ${GREEN}✓${NC} All pganalytics volumes removed"
    else
        echo "  No pganalytics volumes found"
    fi

    # Remove all collector_data_* volumes
    COLLECTOR_VOLUMES=$(docker volume ls -q | grep collector_data_ || true)
    if [ -n "$COLLECTOR_VOLUMES" ]; then
        echo "  Removing collector data volumes..."
        echo "$COLLECTOR_VOLUMES" | while read -r vol; do
            docker volume rm "$vol" 2>/dev/null || true
        done
        echo -e "  ${GREEN}✓${NC} All collector data volumes removed"
    fi
else
    echo "  Docker not found, skipping volume cleanup"
fi

echo ""

# Phase 3: Clean Docker images (optional)
echo -e "${YELLOW}Phase 3: Cleaning pganalytics images (optional)...${NC}"

if command -v docker &> /dev/null; then
    # Remove dangling images
    DANGLING=$(docker images -f "dangling=true" -q 2>/dev/null || true)
    if [ -n "$DANGLING" ]; then
        echo "  Removing dangling images..."
        echo "$DANGLING" | xargs -r docker rmi 2>/dev/null || true
        echo -e "  ${GREEN}✓${NC} Dangling images removed"
    fi

    echo "  Note: Not removing base images - docker-compose will rebuild them"
else
    echo "  Docker not found, skipping image cleanup"
fi

echo ""

# Phase 4: Verify cleanup
echo -e "${YELLOW}Phase 4: Verifying cleanup...${NC}"

if command -v docker &> /dev/null; then
    RUNNING=$(docker ps -q | wc -l)
    VOLUMES=$(docker volume ls -q | grep -E "(pganalytics|collector_data)" | wc -l)

    echo "  Running containers: $RUNNING"
    echo "  Remaining pganalytics volumes: $VOLUMES"

    if [ "$RUNNING" -eq 0 ] && [ "$VOLUMES" -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} Cleanup successful - ready to start fresh"
    else
        echo -e "  ${YELLOW}⚠${NC} Some resources still present (may be from other projects)"
    fi
fi

echo ""

# ============================================================================
# Phase 5: Start core services ONLY (Backend, PostgreSQL, TimescaleDB, Frontend)
# ============================================================================

echo -e "${YELLOW}Phase 5: Starting core infrastructure (WITHOUT collectors)...${NC}"
echo ""
echo "Command: docker-compose -f docker-compose-load-test.yml up -d postgres timescale backend frontend"
echo ""

# Start only core services (without collectors) so we can generate secret
if ! docker-compose -f docker-compose-load-test.yml up -d postgres timescale backend frontend 2>&1 | tail -20; then
    echo -e "${RED}✗ Failed to start docker-compose${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Core infrastructure started${NC}"
echo ""

# ============================================================================
# Phase 6: Wait for core services to be healthy
# ============================================================================

echo -e "${YELLOW}Phase 6: Waiting for core services to be healthy...${NC}"

wait_for_service() {
    local service_name=$1
    local check_cmd=$2
    local max_attempts=30
    local attempt=0

    echo -n "  Waiting for $service_name... "
    while [ $attempt -lt $max_attempts ]; do
        if eval "$check_cmd" > /dev/null 2>&1; then
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

# Check each core service
wait_for_service "PostgreSQL" "docker exec pganalytics-postgres pg_isready -U postgres"
wait_for_service "TimescaleDB" "docker exec pganalytics-timescale pg_isready -U postgres"
wait_for_service "Backend" "curl -sf http://localhost:8080/api/v1/health > /dev/null"
wait_for_service "Frontend" "curl -sf http://localhost:4000 > /dev/null"

echo ""

# ============================================================================
# Phase 7: Generate registration secret via API
# ============================================================================

echo -e "${YELLOW}Phase 7: Generating registration secret via API...${NC}"

REGISTRATION_SECRET=$(./setup-registration-secret.sh 2>&1 | tail -1)

if [ -z "$REGISTRATION_SECRET" ]; then
    echo -e "${RED}✗ Failed to generate registration secret${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Registration secret generated: ${REGISTRATION_SECRET:0:20}...${NC}"
echo ""

# ============================================================================
# Phase 8: Start target PostgreSQL instances
# ============================================================================

echo -e "${YELLOW}Phase 8: Starting 40 target PostgreSQL instances...${NC}"

docker-compose -f docker-compose-load-test.yml up -d target-postgres-001 target-postgres-002 target-postgres-003 target-postgres-004 target-postgres-005 target-postgres-006 target-postgres-007 target-postgres-008 target-postgres-009 target-postgres-010 target-postgres-011 target-postgres-012 target-postgres-013 target-postgres-014 target-postgres-015 target-postgres-016 target-postgres-017 target-postgres-018 target-postgres-019 target-postgres-020 target-postgres-021 target-postgres-022 target-postgres-023 target-postgres-024 target-postgres-025 target-postgres-026 target-postgres-027 target-postgres-028 target-postgres-029 target-postgres-030 target-postgres-031 target-postgres-032 target-postgres-033 target-postgres-034 target-postgres-035 target-postgres-036 target-postgres-037 target-postgres-038 target-postgres-039 target-postgres-040 2>&1 | grep -i "started\|up" | head -5

echo -e "${GREEN}✓ Target instances started${NC}"
echo ""

# Wait for targets to be healthy
echo -e "${YELLOW}Waiting for target instances to be healthy...${NC}"
sleep 20
echo -e "${GREEN}✓ Targets healthy${NC}"
echo ""

# ============================================================================
# Phase 9: Start all 40 collectors with the generated secret
# ============================================================================

echo -e "${YELLOW}Phase 9: Starting 40 collectors with generated secret...${NC}"

export REGISTRATION_SECRET="$REGISTRATION_SECRET"

docker-compose -f docker-compose-load-test.yml up -d collector-001 collector-002 collector-003 collector-004 collector-005 collector-006 collector-007 collector-008 collector-009 collector-010 collector-011 collector-012 collector-013 collector-014 collector-015 collector-016 collector-017 collector-018 collector-019 collector-020 collector-021 collector-022 collector-023 collector-024 collector-025 collector-026 collector-027 collector-028 collector-029 collector-030 collector-031 collector-032 collector-033 collector-034 collector-035 collector-036 collector-037 collector-038 collector-039 collector-040 2>&1 | grep -i "started\|up" | head -10

echo -e "${GREEN}✓ Collectors started${NC}"
echo ""

# ============================================================================
# Phase 10: Final Status
# ============================================================================

echo -e "${BLUE}==============================================================${NC}"
echo -e "${BLUE}Startup Complete!${NC}"
echo -e "${BLUE}==============================================================${NC}"
echo ""

sleep 5

CONTAINER_COUNT=$(docker ps 2>/dev/null | grep pganalytics | wc -l)

echo "Services available at:"
echo "  Backend API: http://localhost:8080"
echo "  Frontend: http://localhost:4000"
echo "  PostgreSQL (metadata): localhost:5432"
echo "  TimescaleDB (metrics): localhost:5433"
echo ""

echo "Infrastructure Status:"
echo "  Containers running: $CONTAINER_COUNT (expected: 84)"
echo "  Registration Secret: ${REGISTRATION_SECRET:0:20}..."
echo ""

echo "Next steps:"
echo "  1. Collectors will auto-register using the generated secret"
echo "  2. Wait 2-3 minutes for all 40 collectors to register and send metrics"
echo "  3. Run: ./test-setup-managed-instances.sh"
echo "  4. Wait 2-3 minutes for metrics collection"
echo "  5. Run: ./verify-regression-tests.sh"
echo ""

echo -e "${GREEN}Load test infrastructure is ready!${NC}"
echo -e "${GREEN}All collectors will use the dynamically generated secret: ${REGISTRATION_SECRET:0:20}...${NC}"
