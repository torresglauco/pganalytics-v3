#!/bin/bash

##############################################################################
# pgAnalytics Restore Script
# Restores database from backup file
##############################################################################

set -e

# Configuration
BACKUP_FILE=$1

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Validate backup file
if [ -z "$BACKUP_FILE" ]; then
    log_error "Usage: $0 <backup-file>"
    echo "Example: $0 ./backups/pganalytics-backup-20260324-120000.sql.gz"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    log_error "Backup file not found: $BACKUP_FILE"
    exit 1
fi

log_info "Restoring from backup: $BACKUP_FILE"
log_warn "This will overwrite the current database!"
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    log_info "Restore cancelled"
    exit 0
fi

# Check if containers are running
if ! docker-compose ps | grep -q postgres; then
    log_warn "PostgreSQL container not running, starting services..."
    docker-compose up -d postgres
    sleep 5
fi

# Restore database
log_info "Restoring database..."
if [[ "$BACKUP_FILE" == *.gz ]]; then
    gunzip < "$BACKUP_FILE" | docker-compose exec -T postgres psql -U postgres
else
    cat "$BACKUP_FILE" | docker-compose exec -T postgres psql -U postgres
fi

if [ $? -eq 0 ]; then
    log_success "Database restored successfully"
else
    log_error "Database restore failed"
    exit 1
fi

log_success "Restore completed"
