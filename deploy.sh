#!/bin/bash

################################################################################
# pgAnalytics-v3 Collector Deployment Script
#
# Deploys the pganalytics-v3 C/C++ collector to test/production environments
# Supports Docker Compose, Kubernetes, and standalone binary installations
#
# Usage:
#   ./deploy.sh [environment] [options]
#
# Environments:
#   docker-compose  - Deploy via Docker Compose (default)
#   e2e             - Deploy E2E test environment
#   standalone      - Deploy standalone binary
#   kubernetes      - Deploy to Kubernetes
#
# Example:
#   ./deploy.sh docker-compose
#   ./deploy.sh e2e
#   ./deploy.sh standalone --host postgres.example.com
#
################################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COLLECTOR_BUILD="${PROJECT_ROOT}/collector/build"
COLLECTOR_BIN="${COLLECTOR_BUILD}/src/pganalytics"

# Deployment configuration
ENVIRONMENT="${1:-docker-compose}"
VERBOSE="${VERBOSE:-false}"

################################################################################
# Utility Functions
################################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $*"
}

log_error() {
    echo -e "${RED}[✗]${NC} $*"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $*"
}

check_binary() {
    if [ ! -f "$COLLECTOR_BIN" ]; then
        log_error "Collector binary not found: $COLLECTOR_BIN"
        log_info "Building collector first..."
        build_collector
    else
        local size=$(ls -lh "$COLLECTOR_BIN" | awk '{print $5}')
        log_success "Collector binary found: $size"
    fi
}

build_collector() {
    log_info "Building collector binary..."
    cd "$PROJECT_ROOT/collector"

    if [ ! -d "build" ]; then
        mkdir -p build
    fi

    cd build
    cmake .. -DCMAKE_BUILD_TYPE=Release
    make -j$(nproc)
    cd "$PROJECT_ROOT"

    log_success "Collector built successfully"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker not installed"
        exit 1
    fi
    log_success "Docker found: $(docker --version)"
}

check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose not installed"
        exit 1
    fi
    log_success "Docker Compose found: $(docker-compose --version)"
}

check_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not installed"
        exit 1
    fi
    log_success "kubectl found: $(kubectl version --client --short 2>/dev/null || echo '(version check skipped)')"
}

################################################################################
# Docker Compose Deployment
################################################################################

deploy_docker_compose() {
    log_info "Deploying with Docker Compose..."

    check_docker
    check_docker_compose
    check_binary

    cd "$PROJECT_ROOT"

    # Build images
    log_info "Building Docker images..."
    docker-compose build 2>&1 | grep -E "(Building|Successfully|error)" || true

    # Start services
    log_info "Starting services..."
    docker-compose up -d

    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 5

    # Check status
    log_info "Service status:"
    docker-compose ps

    # Verify services
    log_info "Verifying backend health..."
    if docker-compose exec -T backend curl -s http://localhost:8080/api/v1/health &>/dev/null; then
        log_success "Backend is healthy"
    else
        log_warning "Backend health check delayed, may still be starting"
    fi

    log_success "Docker Compose deployment complete!"
    log_info ""
    log_info "Access services:"
    log_info "  Grafana:    http://localhost:3000 (admin/admin)"
    log_info "  Backend:    https://localhost:8080"
    log_info "  PostgreSQL: localhost:5432 (postgres/pganalytics)"
    log_info "  TimescaleDB: localhost:5433 (postgres/pganalytics)"
    log_info ""
    log_info "View logs:"
    log_info "  docker-compose logs -f pganalytics-collector-demo"
}

################################################################################
# E2E Test Deployment
################################################################################

deploy_e2e() {
    log_info "Deploying E2E test environment..."

    check_docker
    check_docker_compose
    check_binary

    cd "$PROJECT_ROOT"

    # Build images
    log_info "Building Docker images for E2E..."
    docker-compose -f collector/tests/e2e/docker-compose.e2e.yml build 2>&1 | grep -E "(Building|Successfully|error)" || true

    # Start services
    log_info "Starting E2E services..."
    docker-compose -f collector/tests/e2e/docker-compose.e2e.yml up -d

    # Wait for services
    log_info "Waiting for services to be healthy..."
    sleep 10

    # Check status
    log_info "E2E service status:"
    docker-compose -f collector/tests/e2e/docker-compose.e2e.yml ps

    log_success "E2E deployment complete!"
    log_info ""
    log_info "E2E collector uses 10-second intervals for faster testing"
    log_info "View logs:"
    log_info "  docker-compose -f collector/tests/e2e/docker-compose.e2e.yml logs -f e2e-collector"
}

################################################################################
# Standalone Binary Deployment
################################################################################

deploy_standalone() {
    log_warning "Standalone deployment requires manual SSH/file transfer"
    log_info ""
    log_info "Steps to deploy standalone binary:"
    log_info ""
    log_info "1. Copy binary to target host:"
    log_info "   scp ${COLLECTOR_BIN} postgres@<target-host>:/tmp/pganalytics"
    log_info ""
    log_info "2. SSH to target host:"
    log_info "   ssh postgres@<target-host>"
    log_info ""
    log_info "3. Install binary:"
    log_info "   sudo cp /tmp/pganalytics /usr/local/bin/pganalytics-collector"
    log_info "   sudo chmod +x /usr/local/bin/pganalytics-collector"
    log_info ""
    log_info "4. Create configuration:"
    log_info "   sudo mkdir -p /etc/pganalytics /var/lib/pganalytics"
    log_info "   sudo cp ${PROJECT_ROOT}/collector/config.toml.sample /etc/pganalytics/collector.toml"
    log_info "   sudo vim /etc/pganalytics/collector.toml"
    log_info ""
    log_info "5. Copy TLS certificates (if using mTLS):"
    log_info "   sudo cp /path/to/client.crt /etc/pganalytics/"
    log_info "   sudo cp /path/to/client.key /etc/pganalytics/"
    log_info "   sudo chmod 600 /etc/pganalytics/client.key"
    log_info ""
    log_info "6. Test installation:"
    log_info "   /usr/local/bin/pganalytics-collector cron"
    log_info ""
    log_info "7. Setup systemd service (optional):"
    log_info "   See DEPLOYMENT_GUIDE.md for systemd configuration"
}

################################################################################
# Kubernetes Deployment
################################################################################

deploy_kubernetes() {
    log_warning "Kubernetes deployment requires manual setup"
    log_info ""
    log_info "Steps for Kubernetes deployment:"
    log_info ""
    log_info "1. Build Docker image:"
    log_info "   docker build -f collector/Dockerfile -t pganalytics/collector:1.0.0 ."
    log_info "   docker tag pganalytics/collector:1.0.0 your-registry/pganalytics/collector:1.0.0"
    log_info "   docker push your-registry/pganalytics/collector:1.0.0"
    log_info ""
    log_info "2. Create Kubernetes manifest (see DEPLOYMENT_GUIDE.md)"
    log_info ""
    log_info "3. Deploy to cluster:"
    log_info "   kubectl apply -f pganalytics-collector-k8s.yaml"
    log_info ""
    log_info "4. Verify deployment:"
    log_info "   kubectl get ds -n monitoring"
    log_info "   kubectl logs -n monitoring -l app=pganalytics-collector -f"
}

################################################################################
# Verification Functions
################################################################################

verify_deployment() {
    log_info "Verifying deployment..."

    case "$ENVIRONMENT" in
        docker-compose)
            verify_docker_compose
            ;;
        e2e)
            verify_e2e
            ;;
        *)
            log_warning "Verification not available for $ENVIRONMENT"
            ;;
    esac
}

verify_docker_compose() {
    log_info "Checking service status..."

    # Count running services
    local running=$(docker-compose ps --services --filter "status=running" | wc -l)
    local total=$(docker-compose ps --services | wc -l)

    if [ "$running" -eq "$total" ]; then
        log_success "All services running ($running/$total)"
    else
        log_warning "Some services not running ($running/$total)"
        docker-compose ps
    fi

    # Test backend
    log_info "Testing backend connectivity..."
    if docker-compose exec -T backend curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/health | grep -q "200"; then
        log_success "Backend is responding (HTTP 200)"
    else
        log_warning "Backend connection test inconclusive"
    fi
}

verify_e2e() {
    log_info "Checking E2E environment..."

    local running=$(docker-compose -f collector/tests/e2e/docker-compose.e2e.yml ps --services --filter "status=running" | wc -l)
    local total=$(docker-compose -f collector/tests/e2e/docker-compose.e2e.yml ps --services | wc -l)

    if [ "$running" -eq "$total" ]; then
        log_success "All E2E services running ($running/$total)"
    else
        log_warning "Some E2E services not running ($running/$total)"
    fi
}

################################################################################
# Status and Logs
################################################################################

show_status() {
    log_info "Deployment Status"
    log_info "================="

    case "$ENVIRONMENT" in
        docker-compose)
            docker-compose ps
            ;;
        e2e)
            docker-compose -f collector/tests/e2e/docker-compose.e2e.yml ps
            ;;
        *)
            log_warning "Status not available for $ENVIRONMENT"
            ;;
    esac
}

show_logs() {
    log_info "Recent logs:"

    case "$ENVIRONMENT" in
        docker-compose)
            docker-compose logs --tail=50 pganalytics-collector-demo
            ;;
        e2e)
            docker-compose -f collector/tests/e2e/docker-compose.e2e.yml logs --tail=50 e2e-collector
            ;;
        *)
            log_warning "Logs not available for $ENVIRONMENT"
            ;;
    esac
}

################################################################################
# Help and Usage
################################################################################

show_help() {
    cat << 'EOF'
pgAnalytics-v3 Deployment Script

USAGE:
  ./deploy.sh [environment] [command] [options]

ENVIRONMENTS:
  docker-compose  Deploy via Docker Compose (default, recommended for testing)
  e2e            Deploy E2E test environment with fast collection intervals
  standalone     Deploy standalone binary to remote host
  kubernetes     Deploy to Kubernetes cluster

COMMANDS:
  deploy         Deploy the collector (default)
  status         Show deployment status
  logs           Show collector logs
  verify         Verify deployment health
  stop           Stop services
  restart        Restart services
  clean          Remove containers/volumes
  help           Show this help message

EXAMPLES:
  # Quick start with Docker Compose
  ./deploy.sh docker-compose

  # Deploy E2E test environment
  ./deploy.sh e2e

  # Check deployment status
  ./deploy.sh docker-compose status

  # View collector logs
  ./deploy.sh docker-compose logs

  # Stop and clean up
  ./deploy.sh docker-compose stop
  ./deploy.sh docker-compose clean

ENVIRONMENT VARIABLES:
  VERBOSE         Enable verbose logging (true/false)
  PROJECT_ROOT    Override project root directory

DEPLOYMENT FEATURES:
  ✓ Binary protocol support (60% bandwidth reduction)
  ✓ TLS 1.3 + mTLS authentication
  ✓ Connection pooling for PostgreSQL
  ✓ Automated testing and verification
  ✓ Support for 100,000+ collectors

For detailed information, see:
  DEPLOYMENT_GUIDE.md
  BINARY_PROTOCOL_USAGE_GUIDE.md

EOF
}

################################################################################
# Main Script
################################################################################

main() {
    log_info "pgAnalytics-v3 Collector Deployment"
    log_info "===================================="
    log_info "Environment: $ENVIRONMENT"
    log_info "Project: $PROJECT_ROOT"
    log_info ""

    # Get command
    local command="${2:-deploy}"

    case "$command" in
        deploy)
            case "$ENVIRONMENT" in
                docker-compose)
                    deploy_docker_compose
                    verify_deployment
                    ;;
                e2e)
                    deploy_e2e
                    verify_deployment
                    ;;
                standalone)
                    deploy_standalone
                    ;;
                kubernetes)
                    deploy_kubernetes
                    ;;
                *)
                    log_error "Unknown environment: $ENVIRONMENT"
                    show_help
                    exit 1
                    ;;
            esac
            ;;

        status)
            show_status
            ;;

        logs)
            show_logs
            ;;

        verify)
            verify_deployment
            ;;

        stop)
            log_info "Stopping services..."
            case "$ENVIRONMENT" in
                docker-compose)
                    docker-compose stop
                    log_success "Services stopped"
                    ;;
                e2e)
                    docker-compose -f collector/tests/e2e/docker-compose.e2e.yml stop
                    log_success "E2E services stopped"
                    ;;
                *)
                    log_warning "Stop not supported for $ENVIRONMENT"
                    ;;
            esac
            ;;

        restart)
            log_info "Restarting services..."
            case "$ENVIRONMENT" in
                docker-compose)
                    docker-compose restart
                    log_success "Services restarted"
                    ;;
                e2e)
                    docker-compose -f collector/tests/e2e/docker-compose.e2e.yml restart
                    log_success "E2E services restarted"
                    ;;
                *)
                    log_warning "Restart not supported for $ENVIRONMENT"
                    ;;
            esac
            ;;

        clean)
            log_warning "This will remove containers, networks, and volumes!"
            read -p "Continue? (y/N) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                log_info "Cleaning up..."
                case "$ENVIRONMENT" in
                    docker-compose)
                        docker-compose down -v
                        log_success "Cleaned up Docker Compose environment"
                        ;;
                    e2e)
                        docker-compose -f collector/tests/e2e/docker-compose.e2e.yml down -v
                        log_success "Cleaned up E2E environment"
                        ;;
                    *)
                        log_warning "Cleanup not supported for $ENVIRONMENT"
                        ;;
                esac
            else
                log_info "Cleanup cancelled"
            fi
            ;;

        help)
            show_help
            ;;

        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
