#!/bin/bash

##############################################################################
# pgAnalytics Backup Script
# Creates comprehensive backups of database and configurations
##############################################################################

set -e

# Configuration
BACKUP_DIR="${1:-.}/backups"
RETENTION_DAYS=30
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_FILE="$BACKUP_DIR/pganalytics-backup-$TIMESTAMP"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Create backup directory
mkdir -p "$BACKUP_DIR"
log_info "Backup directory: $BACKUP_DIR"

# Backup PostgreSQL database
log_info "Backing up PostgreSQL database..."
docker-compose exec -T postgres pg_dump -U postgres pganalytics | gzip > "$BACKUP_FILE.sql.gz"
log_success "Database backup: $BACKUP_FILE.sql.gz"

# Backup configuration files
log_info "Backing up configuration..."
tar -czf "$BACKUP_FILE-config.tar.gz" .env docker-compose.yml 2>/dev/null || true
log_success "Config backup: $BACKUP_FILE-config.tar.gz"

# List backups
log_info "Existing backups:"
ls -lh "$BACKUP_DIR"/ | tail -10

# Clean old backups
log_info "Cleaning backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -name "pganalytics-backup-*" -mtime +$RETENTION_DAYS -delete
log_success "Cleanup complete"

log_success "Backup completed successfully"
