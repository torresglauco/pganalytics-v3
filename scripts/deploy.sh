#!/bin/bash
set -e

##############################################################################
# pgAnalytics v3 Deployment Script
# Automates deployment with health checks, rollback, and verification
##############################################################################

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
ENVIRONMENT=${1:-development}
DOCKER_COMPOSE_FILE="docker-compose.yml"
BACKUP_DIR="./backups"
LOG_FILE="deployment-$(date +%Y%m%d-%H%M%S).log"
HEALTH_CHECK_TIMEOUT=60
HEALTH_CHECK_INTERVAL=5

##############################################################################
# Helper Functions
##############################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

##############################################################################
# Pre-Deployment Checks
##############################################################################

pre_deployment_checks() {
    log_info "Running pre-deployment checks..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    log_success "Docker is installed"

    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    log_success "Docker Compose is installed"

    # Check if docker-compose file exists
    if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
        log_error "docker-compose.yml not found"
        exit 1
    fi
    log_success "docker-compose.yml found"

    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    log_success "Backup directory ready"
}

##############################################################################
# Backup Database
##############################################################################

backup_database() {
    log_info "Backing up database..."

    local backup_file="$BACKUP_DIR/pganalytics-$(date +%Y%m%d-%H%M%S).sql.gz"

    docker-compose exec -T postgres pg_dump -U postgres pganalytics | gzip > "$backup_file"

    if [ $? -eq 0 ]; then
        log_success "Database backed up to $backup_file"
        echo "$backup_file"
    else
        log_error "Database backup failed"
        exit 1
    fi
}

##############################################################################
# Stop Services
##############################################################################

stop_services() {
    log_info "Stopping services..."

    docker-compose down

    if [ $? -eq 0 ]; then
        log_success "Services stopped"
    else
        log_error "Failed to stop services"
        exit 1
    fi
}

##############################################################################
# Start Services
##############################################################################

start_services() {
    log_info "Starting services..."

    docker-compose up -d

    if [ $? -eq 0 ]; then
        log_success "Services started"
    else
        log_error "Failed to start services"
        exit 1
    fi
}

##############################################################################
# Health Checks
##############################################################################

health_check() {
    log_info "Performing health checks (timeout: ${HEALTH_CHECK_TIMEOUT}s)..."

    local elapsed=0
    local success=false

    while [ $elapsed -lt $HEALTH_CHECK_TIMEOUT ]; do
        # Check backend health
        if curl -sf http://localhost:8080/api/v1/health > /dev/null 2>&1; then
            log_success "Backend is healthy"

            # Check database
            if docker-compose exec -T postgres psql -U postgres -d pganalytics -c "SELECT 1" > /dev/null 2>&1; then
                log_success "Database is accessible"
                success=true
                break
            fi
        fi

        elapsed=$((elapsed + HEALTH_CHECK_INTERVAL))
        log_warn "Waiting for services... ($elapsed/${HEALTH_CHECK_TIMEOUT}s)"
        sleep $HEALTH_CHECK_INTERVAL
    done

    if [ "$success" = false ]; then
        log_error "Health checks failed after ${HEALTH_CHECK_TIMEOUT}s"
        return 1
    fi

    return 0
}

##############################################################################
# Rollback
##############################################################################

rollback() {
    local backup_file=$1

    log_warn "Rolling back deployment..."

    # Stop services
    docker-compose down

    # Restore database
    log_info "Restoring database from backup..."
    gunzip < "$backup_file" | docker-compose exec -T postgres psql -U postgres

    # Start services
    docker-compose up -d

    # Health check
    if health_check; then
        log_success "Rollback completed successfully"
        return 0
    else
        log_error "Rollback failed - manual intervention required"
        return 1
    fi
}

##############################################################################
# Main Deployment Flow
##############################################################################

main() {
    log_info "=== pgAnalytics v3 Deployment ==="
    log_info "Environment: $ENVIRONMENT"
    log_info "Log file: $LOG_FILE"
    log_info ""

    # Pre-deployment checks
    pre_deployment_checks

    # Backup database
    local backup_file
    backup_file=$(backup_database)

    # Stop services
    stop_services

    # Start services
    start_services

    # Health checks
    if ! health_check; then
        log_error "Health checks failed - rolling back..."
        if rollback "$backup_file"; then
            log_error "Deployment failed - rolled back to previous version"
            exit 1
        else
            log_error "Deployment failed and rollback failed - manual intervention required"
            exit 1
        fi
    fi

    log_success ""
    log_success "=== Deployment Successful ==="
    log_success "Backup saved: $backup_file"
    log_success "Frontend: http://localhost:3000"
    log_success "Backend: http://localhost:8080"
    log_success ""
}

# Run main function
main
