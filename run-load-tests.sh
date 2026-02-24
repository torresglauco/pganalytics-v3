#!/bin/bash

################################################################################
# pgAnalytics Load Test Runner
#
# Executes comprehensive load tests with 10, 50, 100, 500+ collectors
# Compares JSON vs Binary protocol performance
#
# Usage: ./run-load-tests.sh [options]
#
# Options:
#   --quick    Run quick tests (10 collectors only)
#   --full     Run full test suite (10, 50, 100, 500 collectors)
#   --json     Test JSON protocol only
#   --binary   Test Binary protocol only
#   --backend  Backend URL (default: https://localhost:8080)
#
################################################################################

set -eo pipefail
set +u  # Allow unbound variables for now, we'll check them explicitly

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Defaults
BACKEND_URL="https://localhost:8080"
TEST_MODE="full"
PROTOCOLS=("json" "binary")
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOAD_TEST_SCRIPT="${SCRIPT_DIR}/tools/load-test/load_test.py"
RESULTS_DIR="${SCRIPT_DIR}/load-test-results"

# Test scenarios
declare -A SCENARIOS=(
    [quick]="10"
    [standard]="10 50 100"
    [full]="10 50 100 500"
    [extreme]="10 50 100 500"
)

# Test configuration
DURATION_SECONDS=900  # 15 minutes
COLLECTION_INTERVAL=60
METRICS_PER_COLLECTOR=50

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

check_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 not found"
        exit 1
    fi
    log_success "Python 3 found"

    if ! python3 -c "import requests" 2>/dev/null; then
        log_warning "requests library not installed, installing..."
        pip3 install requests --quiet
    fi
    log_success "requests library available"

    if [ ! -f "$LOAD_TEST_SCRIPT" ]; then
        log_error "Load test script not found: $LOAD_TEST_SCRIPT"
        exit 1
    fi
    log_success "Load test script found"

    # Create results directory
    mkdir -p "$RESULTS_DIR"
}

verify_backend() {
    log_info "Verifying backend connectivity..."

    if curl -s -k -f "${BACKEND_URL}/api/v1/health" > /dev/null 2>&1; then
        log_success "Backend is reachable"
    else
        log_warning "Backend may not be ready. Proceeding anyway..."
    fi
}

run_test() {
    local collectors=$1
    local protocol=$2
    local duration=$3
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local test_name="test_${collectors}c_${protocol}_${timestamp}"
    local output_file="${RESULTS_DIR}/${test_name}.log"

    log_info "Running test: $collectors collectors, $protocol protocol"
    log_info "Duration: ${duration}s, Output: $output_file"

    python3 "$LOAD_TEST_SCRIPT" \
        --collectors "$collectors" \
        --duration "$duration" \
        --interval "$COLLECTION_INTERVAL" \
        --metrics "$METRICS_PER_COLLECTOR" \
        --backend "$BACKEND_URL" \
        --protocol "$protocol" \
        2>&1 | tee "$output_file"

    echo "Test completed: $test_name" >> "${RESULTS_DIR}/summary.txt"
}

compare_protocols() {
    local collectors=$1

    log_info ""
    log_info "Protocol Comparison: $collectors collectors"
    log_info "════════════════════════════════════════"

    local json_duration=$((DURATION_SECONDS / 2))
    local binary_duration=$((DURATION_SECONDS / 2))

    log_info "Phase 1: Testing JSON protocol (${json_duration}s)..."
    run_test "$collectors" "json" "$json_duration"

    sleep 10

    log_info "Phase 2: Testing Binary protocol (${binary_duration}s)..."
    run_test "$collectors" "binary" "$binary_duration"
}

run_full_suite() {
    log_info "Starting full load test suite"
    log_info "════════════════════════════════════════"
    log_info "Scenarios: ${SCENARIOS[full]}"
    log_info "Backend: $BACKEND_URL"
    log_info "Results: $RESULTS_DIR"
    log_info ""

    > "${RESULTS_DIR}/summary.txt"

    for collectors in ${SCENARIOS[full]}; do
        log_info ""
        log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        log_info "Test Suite: $collectors Collectors"
        log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

        for protocol in "${PROTOCOLS[@]}"; do
            compare_protocols "$collectors"

            # Cool down between test suites
            if [ "$collectors" != "${SCENARIOS[full]##* }" ]; then
                log_info "Cooling down for 30 seconds..."
                sleep 30
            fi
        done
    done
}

run_quick_suite() {
    log_info "Starting quick load test (10 collectors only)"
    log_info "════════════════════════════════════════"

    > "${RESULTS_DIR}/summary.txt"

    for protocol in "${PROTOCOLS[@]}"; do
        log_info ""
        log_info "Testing $protocol protocol..."
        run_test 10 "$protocol" "$((DURATION_SECONDS / 2))"
    done
}

show_help() {
    cat << 'EOF'
pgAnalytics Load Test Runner

USAGE:
  ./run-load-tests.sh [options]

OPTIONS:
  --quick       Run quick tests (10 collectors, both protocols)
  --full        Run full test suite (10, 50, 100, 500 collectors)
  --json        Test JSON protocol only
  --binary      Test Binary protocol only
  --backend URL Backend URL (default: https://localhost:8080)
  --help        Show this help message

EXAMPLES:
  # Quick test
  ./run-load-tests.sh --quick

  # Full test suite
  ./run-load-tests.sh --full

  # Test only JSON protocol
  ./run-load-tests.sh --json

  # Custom backend
  ./run-load-tests.sh --full --backend https://api.example.com:8080

TEST SCENARIOS:
  Quick:    10 collectors
  Standard: 10, 50, 100 collectors
  Full:     10, 50, 100, 500 collectors

RESULTS:
  Results are saved to: load-test-results/
  Summary: load-test-results/summary.txt

EOF
}

################################################################################
# Main
################################################################################

main() {
    log_info "pgAnalytics Load Test Runner"
    log_info "════════════════════════════════════════"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --quick)
                TEST_MODE="quick"
                shift
                ;;
            --full)
                TEST_MODE="full"
                shift
                ;;
            --json)
                PROTOCOLS=("json")
                shift
                ;;
            --binary)
                PROTOCOLS=("binary")
                shift
                ;;
            --backend)
                BACKEND_URL="$2"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Check prerequisites
    check_prerequisites
    verify_backend

    # Run tests
    if [ "$TEST_MODE" = "quick" ]; then
        run_quick_suite
    else
        run_full_suite
    fi

    # Print summary
    log_success ""
    log_success "Load test completed!"
    log_success "Results saved to: $RESULTS_DIR"
    log_success ""
    log_info "To view results:"
    log_info "  cat $RESULTS_DIR/summary.txt"
    log_info ""
}

main "$@"
