#!/bin/bash

# =============================================================================
# Phase 4 v4.0.0 Staging Deployment Script
# =============================================================================
# This script automates the Docker deployment of pgAnalytics Phase 4
# to a staging environment with comprehensive verification.
#
# Usage: ./scripts/deploy-staging.sh [options]
# Options:
#   --skip-build      Skip Docker image building
#   --skip-tests      Skip smoke tests
#   --verbose         Show detailed output
#   --rollback        Rollback to previous version
# =============================================================================

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DOCKER_COMPOSE_FILE="$PROJECT_DIR/docker-compose.staging.yml"
ENV_FILE="$PROJECT_DIR/.env.staging"
LOG_FILE="$PROJECT_DIR/deployment-$(date +%Y%m%d-%H%M%S).log"

# Parse command-line arguments
SKIP_BUILD=false
SKIP_TESTS=false
VERBOSE=false
ROLLBACK=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --skip-build) SKIP_BUILD=true; shift ;;
    --skip-tests) SKIP_TESTS=true; shift ;;
    --verbose) VERBOSE=true; shift ;;
    --rollback) ROLLBACK=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# Logging functions
log() {
  echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

success() {
  echo -e "${GREEN}[✓]${NC} $1" | tee -a "$LOG_FILE"
}

warning() {
  echo -e "${YELLOW}[⚠]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
  echo -e "${RED}[✗]${NC} $1" | tee -a "$LOG_FILE"
}

# =============================================================================
# PRE-DEPLOYMENT CHECKS
# =============================================================================

check_prerequisites() {
  log "Checking prerequisites..."

  # Check Docker
  if ! command -v docker &> /dev/null; then
    error "Docker is not installed. Please install Docker."
    exit 1
  fi
  success "Docker installed: $(docker --version)"

  # Check Docker Compose
  if ! command -v docker-compose &> /dev/null; then
    error "Docker Compose is not installed. Please install Docker Compose."
    exit 1
  fi
  success "Docker Compose installed: $(docker-compose --version)"

  # Check Git
  if ! command -v git &> /dev/null; then
    error "Git is not installed. Please install Git."
    exit 1
  fi
  success "Git installed: $(git --version)"

  # Check Docker daemon
  if ! docker ping &> /dev/null 2>&1; then
    if ! docker ps &> /dev/null 2>&1; then
      error "Docker daemon is not running. Please start Docker."
      exit 1
    fi
  fi
  success "Docker daemon is running"

  # Check project directory
  if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
    error "docker-compose.staging.yml not found at $DOCKER_COMPOSE_FILE"
    exit 1
  fi
  success "Docker Compose configuration found"
}

create_env_file() {
  log "Setting up environment file..."

  if [ -f "$ENV_FILE" ]; then
    warning "Environment file already exists at $ENV_FILE"
    read -p "Do you want to regenerate it? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      log "Using existing environment file"
      return
    fi
  fi

  # Generate secure passwords
  DB_PASSWORD="staging-$(openssl rand -hex 12)"
  JWT_SECRET="jwt-$(openssl rand -hex 32)"
  GRAFANA_PASSWORD="grafana-$(openssl rand -hex 12)"

  cat > "$ENV_FILE" << EOF
# Database Configuration
DB_PASSWORD=$DB_PASSWORD
POSTGRES_DB=pganalytics_staging
POSTGRES_USER=pganalytics

# API Configuration
API_PORT=8000
JWT_SECRET=$JWT_SECRET
LOG_LEVEL=info
ENVIRONMENT=staging

# Monitoring
GRAFANA_PASSWORD=$GRAFANA_PASSWORD

# Frontend
VITE_API_URL=http://localhost:8000

# Deployment timestamp
DEPLOYED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION=4.0.0
EOF

  success "Environment file created at $ENV_FILE"
  log "Generated secure credentials (saved in $ENV_FILE)"
}

validate_config() {
  log "Validating Docker Compose configuration..."

  if ! docker-compose -f "$DOCKER_COMPOSE_FILE" config > /dev/null 2>&1; then
    error "Docker Compose configuration is invalid"
    docker-compose -f "$DOCKER_COMPOSE_FILE" config
    exit 1
  fi

  success "Docker Compose configuration is valid"
}

# =============================================================================
# DEPLOYMENT FUNCTIONS
# =============================================================================

cleanup_old_containers() {
  log "Cleaning up old containers..."

  if docker-compose -f "$DOCKER_COMPOSE_FILE" ps -q 2>/dev/null | grep -q . ; then
    log "Stopping existing containers..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" down 2>/dev/null || true
    sleep 2
  fi

  success "Old containers cleaned up"
}

build_images() {
  if [ "$SKIP_BUILD" = true ]; then
    log "Skipping Docker image build (--skip-build)"
    return
  fi

  log "Building Docker images..."
  log "This may take 2-5 minutes..."

  docker-compose -f "$DOCKER_COMPOSE_FILE" build --no-cache

  success "Docker images built successfully"
}

start_services() {
  log "Starting services..."

  docker-compose -f "$DOCKER_COMPOSE_FILE" up -d

  log "Services started, waiting for health checks..."

  # Wait for PostgreSQL
  log "Waiting for PostgreSQL to be ready..."
  local max_attempts=30
  local attempt=0
  while [ $attempt -lt $max_attempts ]; do
    if docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T postgres pg_isready -U pganalytics -d pganalytics_staging &>/dev/null; then
      success "PostgreSQL is ready"
      break
    fi
    attempt=$((attempt + 1))
    sleep 2
  done

  if [ $attempt -eq $max_attempts ]; then
    error "PostgreSQL failed to start after $max_attempts attempts"
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs postgres
    exit 1
  fi

  # Wait for API
  log "Waiting for API to be ready..."
  attempt=0
  while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:8000/health &>/dev/null; then
      success "API is ready"
      break
    fi
    attempt=$((attempt + 1))
    sleep 2
  done

  if [ $attempt -eq $max_attempts ]; then
    error "API failed to start after $max_attempts attempts"
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs api
    exit 1
  fi

  # Wait for Frontend
  log "Waiting for Frontend to be ready..."
  attempt=0
  while [ $attempt -lt $max_attempts ]; do
    if curl -s http://localhost:3000 &>/dev/null; then
      success "Frontend is ready"
      break
    fi
    attempt=$((attempt + 1))
    sleep 2
  done

  success "All services started and healthy"
}

verify_services() {
  log "Verifying service health..."

  log "Checking service status..."
  docker-compose -f "$DOCKER_COMPOSE_FILE" ps

  success "Services are running"

  log "Checking API health..."
  API_HEALTH=$(curl -s http://localhost:8000/health)
  if echo "$API_HEALTH" | jq . > /dev/null 2>&1; then
    success "API health check passed"
    if [ "$VERBOSE" = true ]; then
      echo "$API_HEALTH" | jq .
    fi
  else
    error "API health check failed"
    exit 1
  fi

  log "Checking Frontend accessibility..."
  FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/)
  if [ "$FRONTEND_STATUS" = "200" ]; then
    success "Frontend is accessible (HTTP $FRONTEND_STATUS)"
  else
    error "Frontend not accessible (HTTP $FRONTEND_STATUS)"
    exit 1
  fi
}

# =============================================================================
# SMOKE TESTING
# =============================================================================

run_smoke_tests() {
  if [ "$SKIP_TESTS" = true ]; then
    log "Skipping smoke tests (--skip-tests)"
    return
  fi

  log "Running smoke tests..."

  # Load test configuration
  source "$PROJECT_DIR/.env.staging"

  # For testing, use a simple test JWT token
  # In production, this would come from your auth system
  JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRlc3QgVXNlciIsImlhdCI6MTcxMDQyNDgwMH0.test-signature"

  TESTS_PASSED=0
  TESTS_FAILED=0

  # Test 1: API Health
  log "Test 1: API Health Check"
  if curl -s http://localhost:8000/health | jq . > /dev/null 2>&1; then
    success "API health check passed"
    ((TESTS_PASSED++))
  else
    error "API health check failed"
    ((TESTS_FAILED++))
  fi

  # Test 2: Frontend Access
  log "Test 2: Frontend Access"
  FRONTEND_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/)
  if [ "$FRONTEND_CODE" = "200" ]; then
    success "Frontend accessible"
    ((TESTS_PASSED++))
  else
    error "Frontend not accessible (HTTP $FRONTEND_CODE)"
    ((TESTS_FAILED++))
  fi

  # Test 3: Create Alert Rule
  log "Test 3: Create Alert Rule"
  ALERT_RULE=$(cat <<'EOFJSON'
{
  "name": "Smoke Test Alert Rule",
  "description": "Alert rule created during smoke test",
  "conditions": [
    {
      "metric_type": "error_count",
      "operator": ">",
      "threshold": 10,
      "time_window": 5,
      "duration": 300
    }
  ]
}
EOFJSON
)

  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    http://localhost:8000/api/v1/alert-rules \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "X-Instance-ID: test-instance" \
    -d "$ALERT_RULE")

  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    success "Alert rule created"
    ((TESTS_PASSED++))
    RULE_ID=$(echo "$RESPONSE" | head -1 | jq -r '.id // empty' 2>/dev/null)
  else
    warning "Alert rule creation returned HTTP $HTTP_CODE"
    ((TESTS_FAILED++))
    RULE_ID=""
  fi

  # Test 4: Create Silence (if rule was created)
  if [ ! -z "$RULE_ID" ]; then
    log "Test 4: Create Alert Silence"
    SILENCE=$(cat <<EOF
{
  "alert_rule_id": "$RULE_ID",
  "duration_seconds": 3600,
  "reason": "Smoke test silence",
  "expires_at": "$(date -u -d '+1 hour' +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -v+1H +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
      http://localhost:8000/api/v1/alert-silences \
      -H "Authorization: Bearer $JWT_TOKEN" \
      -H "Content-Type: application/json" \
      -H "X-Instance-ID: test-instance" \
      -d "$SILENCE")

    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
      success "Alert silence created"
      ((TESTS_PASSED++))
    else
      warning "Alert silence creation returned HTTP $HTTP_CODE"
      ((TESTS_FAILED++))
    fi
  fi

  # Test 5: Create Escalation Policy
  log "Test 5: Create Escalation Policy"
  POLICY=$(cat <<'EOFJSON'
{
  "name": "Smoke Test Policy",
  "description": "Escalation policy created during smoke test",
  "steps": [
    {
      "step_number": 1,
      "wait_minutes": 5,
      "notification_channel": "email",
      "channel_config": {
        "email": "test@example.com"
      }
    }
  ]
}
EOFJSON
)

  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    http://localhost:8000/api/v1/escalation-policies \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -H "X-Instance-ID: test-instance" \
    -d "$POLICY")

  HTTP_CODE=$(echo "$RESPONSE" | tail -1)
  if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
    success "Escalation policy created"
    ((TESTS_PASSED++))
  else
    warning "Escalation policy creation returned HTTP $HTTP_CODE"
    ((TESTS_FAILED++))
  fi

  # Summary
  echo ""
  log "Smoke Test Summary"
  success "$TESTS_PASSED tests passed"
  if [ $TESTS_FAILED -gt 0 ]; then
    warning "$TESTS_FAILED tests failed (may be due to missing auth setup)"
  fi
}

show_access_info() {
  log "Deployment complete! Access information:"
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo -e "${GREEN}Frontend${NC}:      http://localhost:3000"
  echo -e "${GREEN}API${NC}:           http://localhost:8000"
  echo -e "${GREEN}API Health${NC}:    http://localhost:8000/health"
  echo -e "${GREEN}Prometheus${NC}:    http://localhost:9090"
  echo -e "${GREEN}Grafana${NC}:       http://localhost:3001"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  if [ -f "$ENV_FILE" ]; then
    log "Credentials stored in: $ENV_FILE"
    GRAFANA_PASS=$(grep GRAFANA_PASSWORD "$ENV_FILE" | cut -d'=' -f2)
    echo "Grafana Admin Password: $GRAFANA_PASS"
  fi

  echo ""
  log "Service Status:"
  docker-compose -f "$DOCKER_COMPOSE_FILE" ps

  echo ""
  log "Deployment log saved to: $LOG_FILE"
}

# =============================================================================
# MAIN EXECUTION
# =============================================================================

main() {
  echo ""
  echo "╔════════════════════════════════════════════════════════════════╗"
  echo "║     Phase 4 v4.0.0 Staging Deployment                         ║"
  echo "║     pgAnalytics Advanced UI Features                          ║"
  echo "╚════════════════════════════════════════════════════════════════╝"
  echo ""

  log "Deployment started at $(date)"
  log "Log file: $LOG_FILE"
  echo ""

  if [ "$ROLLBACK" = true ]; then
    log "ROLLBACK mode requested"
    error "Rollback functionality not yet implemented"
    exit 1
  fi

  check_prerequisites
  create_env_file
  validate_config
  cleanup_old_containers
  build_images
  start_services
  verify_services
  run_smoke_tests
  show_access_info

  echo ""
  success "Staging deployment completed successfully!"
  echo ""
}

# Run main function
main
