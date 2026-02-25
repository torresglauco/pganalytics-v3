#!/bin/bash

# cleanup-docs.sh - pgAnalytics v3 Documentation Cleanup Script
# This script removes non-essential development documentation files
# and keeps only production-critical user documentation.
#
# USAGE: ./cleanup-docs.sh [--dry-run] [--force]
# --dry-run: Show what would be deleted without actually deleting
# --force: Delete without prompting for confirmation

set -e

DRY_RUN=false
FORCE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--dry-run] [--force]"
            exit 1
            ;;
    esac
done

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Files to KEEP (essential user documentation)
KEEP_FILES=(
    "README.md"
    "SECURITY.md"
    "SETUP.md"
    "DEPLOYMENT_PLAN_v3.2.0.md"
    "docs/ARCHITECTURE.md"
    "docs/REPLICATION_COLLECTOR_GUIDE.md"
    "docs/API_SECURITY_REFERENCE.md"
    "docs/GRAFANA_REPLICATION_DASHBOARDS.md"
)

# Files to DELETE (development artifacts)
DELETE_FILES=(
    "COMPREHENSIVE_AUDIT_REPORT.md"
    "PROJECT_STATUS_SUMMARY.md"
    "RELEASE_v3.2.0.md"
    "PHASE1_COMPILATION_TEST_REPORT.md"
    "PHASE1_IMPLEMENTATION_SUMMARY.md"
    "PHASE1_QUICK_REFERENCE.md"
    "PHASE1_INTEGRATION_COMPLETE.md"
    "PHASE1_EXECUTION_GUIDE.md"
    "PHASE1_COMPLETION_CHECKLIST.md"
    "PHASE1_DEPLOYMENT_INDEX.md"
    "COLLECTOR_ENHANCEMENT_PLAN.md"
    "CLEANUP_SUMMARY.md"
    "ARCHITECTURE_DIAGRAM.md"
    "AUDIT_SUMMARY.md"
    "CODE_REVIEW_FINDINGS.md"
    "DEPLOYMENT_GUIDE.md"
    "EVERYTHING_ACCOMPLISHED.md"
    "FINAL_SUMMARY.md"
    "GITHUB_RELEASES_SUMMARY.md"
    "LOAD_TEST_REPORT_FEB_2026.md"
    "MANAGEMENT_REPORT_FEBRUARY_2026.md"
    "POSTGRESQL_17_18_IMPLEMENTATION_GUIDE.md"
    "POSTGRESQL_ANALYSIS_INDEX.md"
    "POSTGRESQL_VERSION_ANALYSIS_SUMMARY.md"
    "POSTGRESQL_VERSION_COMPATIBILITY_REPORT.md"
    "PROJECT_COMPLETION_REPORT.md"
    "QUICK_START.md"
    "RELEASE_NOTES.md"
    "ROADMAP_v3.2.0.md"
    "TEAM_SUMMARY.md"
    "docs/archived"
    "docs/guides"
    "docs/tests"
    "docs/ML_WORKFLOW_DIAGRAM.md"
    "docs/ML_AND_GRAPHQL_DOCUMENTATION_INDEX.md"
    "docs/DEPLOYMENT_READY.md"
    "docs/LOAD_TEST_PLAN.md"
    "docs/COLLECTOR_IMPLEMENTATION_SUMMARY.md"
    "docs/GRAPHQL_STATUS.md"
    "docs/GETTING_STARTED.md"
    "docs/LOAD_TEST_EXECUTION.md"
    "docs/MILESTONE_1_QUICKSTART.md"
    "docs/POSTGRESQL_QUICK_REFERENCE.md"
    "docs/GRAFANA_DASHBOARD_SETUP.md"
    "docs/INDEX.md"
    "docs/DOCUMENTATION_INDEX.md"
    "docs/api/API_QUICK_REFERENCE.md"
    "docs/api/BINARY_PROTOCOL_INTEGRATION_COMPLETE.md"
    "docs/api/LOAD_TEST_RESULTS.md"
    "docs/api/BINARY_PROTOCOL_USAGE_GUIDE.md"
    "docs/ML_FEATURES_DETAILED.md"
    "docs/DEPLOYMENT_COMPLETE.md"
    "collector/COLLECTOR_IMPLEMENTATION_NOTES.md"
    "collector/QUICK_START.md"
    "collector/BUILD_AND_DEPLOY.md"
    "collector/tests/integration/README.md"
    "collector/tests/README.md"
    "collector/tests/e2e/README.md"
)

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}pgAnalytics v3 - Documentation Cleanup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Print header
if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}[DRY RUN MODE] - No files will be deleted${NC}"
    echo ""
fi

# Count files
KEEP_COUNT=${#KEEP_FILES[@]}
DELETE_COUNT=${#DELETE_FILES[@]}

echo -e "${GREEN}Files to KEEP ($KEEP_COUNT):${NC}"
for file in "${KEEP_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "  ${GREEN}✓${NC} $file"
    else
        echo -e "  ${YELLOW}?${NC} $file (not found)"
    fi
done

echo ""
echo -e "${RED}Files to DELETE ($DELETE_COUNT):${NC}"

files_to_delete=()
for file in "${DELETE_FILES[@]}"; do
    if [ -f "$file" ] || [ -d "$file" ]; then
        echo -e "  ${RED}✗${NC} $file"
        files_to_delete+=("$file")
    fi
done

echo ""
echo -e "${YELLOW}Summary:${NC}"
echo "  Files to keep: $KEEP_COUNT"
echo "  Files to delete: ${#files_to_delete[@]}"
echo ""

# If no files to delete, exit early
if [ ${#files_to_delete[@]} -eq 0 ]; then
    echo -e "${GREEN}No files to delete. Documentation is already clean!${NC}"
    exit 0
fi

# Ask for confirmation unless --force is used
if [ "$FORCE" = false ] && [ "$DRY_RUN" = false ]; then
    echo -e "${YELLOW}WARNING: This will permanently delete ${#files_to_delete[@]} files.${NC}"
    read -p "Are you sure? (type 'yes' to confirm): " confirmation
    if [ "$confirmation" != "yes" ]; then
        echo "Cleanup cancelled."
        exit 0
    fi
fi

# Delete files
deleted_count=0
for file in "${files_to_delete[@]}"; do
    if [ "$DRY_RUN" = true ]; then
        echo -e "  [DRY RUN] Would delete: $file"
    else
        if rm -rf "$file" 2>/dev/null; then
            echo -e "  ${GREEN}✓ Deleted: $file${NC}"
            ((deleted_count++))
        else
            echo -e "  ${RED}✗ Failed to delete: $file${NC}"
        fi
    fi
done

echo ""
echo -e "${BLUE}========================================${NC}"
if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}[DRY RUN] Would delete $deleted_count files${NC}"
else
    echo -e "${GREEN}Successfully deleted $deleted_count files${NC}"
    echo -e "${GREEN}Documentation cleanup complete!${NC}"
fi
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Remaining essential documentation:${NC}"
find . -maxdepth 1 -name "*.md" -type f | sort | head -20
echo ""
find docs -maxdepth 2 -name "*.md" -type f 2>/dev/null | grep -v archived | grep -v guides | grep -v tests | sort

exit 0
