#!/bin/bash

# Test Runner Script for pgAnalytics Collector
# Usage: ./run_tests.sh [options]
# Options:
#   --all       Run all tests with verbose output
#   --quick     Run tests with minimal output
#   --coverage  Generate code coverage report
#   --filter    Run specific tests (pass pattern after flag)
#   --repeat    Repeat tests N times (pass number after flag)
#   --help      Show this help message

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_DIR/build"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
VERBOSE=0
QUICK=0
COVERAGE=0
TEST_FILTER=""
REPEAT_COUNT=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --all)
            VERBOSE=1
            shift
            ;;
        --quick)
            QUICK=1
            shift
            ;;
        --coverage)
            COVERAGE=1
            shift
            ;;
        --filter)
            TEST_FILTER="$2"
            shift 2
            ;;
        --repeat)
            REPEAT_COUNT="$2"
            shift 2
            ;;
        --help)
            grep "^#" "$0" | tail -n +2
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}====== pgAnalytics Collector Test Suite ======${NC}"
echo ""

# Check if build directory exists
if [ ! -d "$BUILD_DIR" ]; then
    echo -e "${RED}Error: Build directory not found at $BUILD_DIR${NC}"
    echo "Please run: cd $PROJECT_DIR && mkdir build && cd build && cmake .. && make"
    exit 1
fi

# Check if test executable exists
if [ ! -f "$BUILD_DIR/tests/pganalytics-tests" ]; then
    echo -e "${RED}Error: Test executable not found at $BUILD_DIR/tests/pganalytics-tests${NC}"
    echo "Please rebuild tests: cd $BUILD_DIR && make"
    exit 1
fi

# Build test command
TEST_CMD="$BUILD_DIR/tests/pganalytics-tests"

# Add verbosity flags
if [ $VERBOSE -eq 1 ]; then
    TEST_CMD="$TEST_CMD --gtest_print_time=1"
fi

# Add filter if specified
if [ -n "$TEST_FILTER" ]; then
    echo -e "${YELLOW}Running tests matching: $TEST_FILTER${NC}"
    TEST_CMD="$TEST_CMD --gtest_filter='$TEST_FILTER'"
fi

# Add repeat count if specified
if [ -n "$REPEAT_COUNT" ]; then
    echo -e "${YELLOW}Repeating tests $REPEAT_COUNT times${NC}"
    TEST_CMD="$TEST_CMD --gtest_repeat=$REPEAT_COUNT"
fi

# Save results to XML if running full suite
if [ -z "$TEST_FILTER" ] && [ $QUICK -eq 0 ]; then
    TEST_CMD="$TEST_CMD --gtest_output='xml:test-results.xml'"
fi

echo ""
echo -e "${YELLOW}Building test executable...${NC}"
cd "$BUILD_DIR"
make -j$(nproc) pganalytics-tests

echo ""
echo -e "${YELLOW}Running tests...${NC}"
echo ""

# Run tests
if eval "$TEST_CMD"; then
    echo ""
    echo -e "${GREEN}✓ All tests passed!${NC}"

    # Show test results summary if XML file was created
    if [ -f "test-results.xml" ]; then
        echo ""
        echo -e "${YELLOW}Test Results Summary:${NC}"
        grep "tests=" test-results.xml | head -1
    fi

    # Generate coverage if requested
    if [ $COVERAGE -eq 1 ]; then
        echo ""
        echo -e "${YELLOW}Generating code coverage report...${NC}"

        # Check if coverage tools are available
        if command -v lcov &> /dev/null; then
            lcov --directory . --capture --output-file coverage.info
            lcov --remove coverage.info '/usr/*' '*/tests/*' --output-file coverage.info
            genhtml coverage.info --output-directory coverage-report

            echo -e "${GREEN}✓ Coverage report generated in: $BUILD_DIR/coverage-report/index.html${NC}"
        else
            echo -e "${YELLOW}⚠ lcov not found. Install with: brew install lcov (macOS) or apt-get install lcov (Ubuntu)${NC}"
        fi
    fi

    exit 0
else
    echo ""
    echo -e "${RED}✗ Some tests failed!${NC}"
    echo ""
    echo -e "${YELLOW}Tips for debugging:${NC}"
    echo "  1. Run with verbose output: ./run_tests.sh --all"
    echo "  2. Run specific test: ./run_tests.sh --filter 'TestName'"
    echo "  3. Check /tmp for temporary test files"
    echo "  4. Review test output above for details"
    exit 1
fi
