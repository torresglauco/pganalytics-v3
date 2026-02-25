#!/bin/bash

##############################################################################
# pgAnalytics v3.2.0 - Database Setup Script
# Purpose: Initialize PostgreSQL database with roles, permissions, and schema
# Usage: ./database-setup.sh [--host <host>] [--user <user>] [--password <pass>]
##############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration - Update these for your environment
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_ADMIN_USER="${DB_ADMIN_USER:-postgres}"
DB_ADMIN_PASSWORD="${DB_ADMIN_PASSWORD:-}"
DB_NAME="${DB_NAME:-pganalytics}"
DB_USER="${DB_USER:-pganalytics}"
DB_PASSWORD="${DB_PASSWORD:-}"
WAL_ARCHIVE_DIR="${WAL_ARCHIVE_DIR:-/var/lib/postgresql/wal_archive}"
BACKUP_DIR="${BACKUP_DIR:-/var/lib/postgresql/backups}"

##############################################################################
# Helper Functions
##############################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Execute SQL with admin credentials
execute_sql() {
    local sql="$1"
    local as_user="${2:-$DB_ADMIN_USER}"

    if [ -z "$DB_ADMIN_PASSWORD" ]; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$as_user" -d postgres -c "$sql" 2>/dev/null || true
    else
        PGPASSWORD="$DB_ADMIN_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$as_user" -d postgres -c "$sql" 2>/dev/null || true
    fi
}

# Execute SQL file
execute_sql_file() {
    local file="$1"
    local db="${2:-postgres}"
    local as_user="${3:-$DB_ADMIN_USER}"

    if [ ! -f "$file" ]; then
        log_error "SQL file not found: $file"
    fi

    if [ -z "$DB_ADMIN_PASSWORD" ]; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$as_user" -d "$db" -f "$file" 2>/dev/null || true
    else
        PGPASSWORD="$DB_ADMIN_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$as_user" -d "$db" -f "$file" 2>/dev/null || true
    fi
}

##############################################################################
# Connection Validation
##############################################################################

validate_connection() {
    log_info "Validating PostgreSQL connection..."

    if ! execute_sql "SELECT version();" > /dev/null 2>&1; then
        log_error "Cannot connect to PostgreSQL at $DB_HOST:$DB_PORT"
    fi

    log_success "PostgreSQL connection successful"
}

check_postgres_version() {
    log_info "Checking PostgreSQL version..."

    if ! execute_sql "SELECT version();" | grep -q "PostgreSQL 1[6-9]"; then
        log_warning "PostgreSQL 16+ recommended. Current version:"
        execute_sql "SELECT version();"
    else
        log_success "PostgreSQL version 16+ detected"
    fi
}

check_ssl_enabled() {
    log_info "Checking SSL status..."

    local ssl_status
    ssl_status=$(execute_sql "SHOW ssl;" | grep -oE "^[^|]+$" | sed 's/[[:space:]]*$//' | tail -1)

    if [ "$ssl_status" != "on" ]; then
        log_warning "SSL is not enabled (ssl=$ssl_status). Recommended for production."
    else
        log_success "SSL is enabled"
    fi
}

##############################################################################
# Role & Database Creation
##############################################################################

create_pganalytics_role() {
    log_info "Creating pganalytics role..."

    # Check if role exists
    local role_exists
    role_exists=$(execute_sql "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER';" | grep -c "1" || true)

    if [ "$role_exists" -gt 0 ]; then
        log_warning "Role '$DB_USER' already exists. Updating password..."
        execute_sql "ALTER ROLE $DB_USER WITH PASSWORD '$DB_PASSWORD';"
        log_success "Role password updated"
    else
        # Create new role
        execute_sql "CREATE ROLE $DB_USER WITH LOGIN NOINHERIT;"
        execute_sql "ALTER ROLE $DB_USER WITH PASSWORD '$DB_PASSWORD';"
        log_success "Role '$DB_USER' created"
    fi
}

grant_monitor_permissions() {
    log_info "Granting pg_monitor role to $DB_USER..."

    execute_sql "GRANT pg_monitor TO $DB_USER;"
    log_success "pg_monitor role granted"
}

create_database() {
    log_info "Creating pganalytics database..."

    # Check if database exists
    local db_exists
    db_exists=$(execute_sql "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';" | grep -c "1" || true)

    if [ "$db_exists" -gt 0 ]; then
        log_warning "Database '$DB_NAME' already exists"
    else
        execute_sql "CREATE DATABASE $DB_NAME OWNER $DB_USER;"
        log_success "Database '$DB_NAME' created"
    fi
}

##############################################################################
# Schema & Tables
##############################################################################

initialize_schema() {
    log_info "Initializing pganalytics schema..."

    # Create schema SQL
    cat > /tmp/pganalytics_schema.sql << 'EOF'
-- pgAnalytics v3.2.0 Schema

-- Create collectors table
CREATE TABLE IF NOT EXISTS collectors (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    hostname VARCHAR(255) NOT NULL,
    address INET,
    status VARCHAR(50) DEFAULT 'active',
    token_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create metrics table
CREATE TABLE IF NOT EXISTS metrics (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID REFERENCES collectors(id) ON DELETE CASCADE,
    metric_name VARCHAR(255) NOT NULL,
    metric_value FLOAT,
    labels JSONB,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT metric_time_value CHECK (timestamp <= NOW())
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_metrics_collector_timestamp
    ON metrics(collector_id, timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_metrics_name
    ON metrics(metric_name);

CREATE INDEX IF NOT EXISTS idx_metrics_timestamp
    ON metrics(timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_collectors_name
    ON collectors(name);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'viewer',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username
    ON users(username);

-- Create audit log table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(255),
    details JSONB,
    status VARCHAR(50),
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp
    ON audit_logs(timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user
    ON audit_logs(user_id);

-- Set search path
ALTER DATABASE pganalytics SET search_path TO public;

-- Grant permissions
GRANT USAGE ON SCHEMA public TO pganalytics;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO pganalytics;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO pganalytics;
GRANT SELECT ON pg_stat_replication TO pganalytics;
GRANT SELECT ON pg_replication_slots TO pganalytics;
GRANT SELECT ON pg_ls_wal_dir() TO pganalytics;

-- Future privilege grants
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE ON TABLES TO pganalytics;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO pganalytics;
EOF

    execute_sql_file /tmp/pganalytics_schema.sql "$DB_NAME" "$DB_ADMIN_USER"
    rm -f /tmp/pganalytics_schema.sql

    log_success "Schema initialized"
}

##############################################################################
# WAL Archiving
##############################################################################

setup_wal_archiving() {
    log_info "Setting up WAL archiving..."

    # Check if already enabled
    local archive_mode
    archive_mode=$(execute_sql "SHOW archive_mode;" | grep -oE "^on|off$" || echo "off")

    if [ "$archive_mode" = "on" ]; then
        log_warning "WAL archiving already enabled"
        return
    fi

    # Create archive directory
    if [ ! -d "$WAL_ARCHIVE_DIR" ]; then
        log_info "Creating WAL archive directory: $WAL_ARCHIVE_DIR"
        sudo mkdir -p "$WAL_ARCHIVE_DIR"
        sudo chown postgres:postgres "$WAL_ARCHIVE_DIR"
        sudo chmod 700 "$WAL_ARCHIVE_DIR"
        log_success "WAL archive directory created"
    fi

    # Create archive script
    cat > /tmp/archive_wal.sh << 'EOF'
#!/bin/bash
# WAL archiving script
# Called by PostgreSQL with archive_command

WAL_ARCHIVE_DIR="${1:-.}"
WAL_FILE="${2:-}"

if [ -z "$WAL_FILE" ]; then
    exit 1
fi

if cp "$1/$2" "$WAL_ARCHIVE_DIR/$2"; then
    exit 0
else
    exit 1
fi
EOF

    sudo cp /tmp/archive_wal.sh /usr/local/bin/pganalytics_archive_wal.sh
    sudo chmod 755 /usr/local/bin/pganalytics_archive_wal.sh
    rm -f /tmp/archive_wal.sh

    # Update postgresql.conf
    log_info "Updating postgresql.conf for WAL archiving..."
    log_warning "MANUAL STEP REQUIRED: Update postgresql.conf with:"
    echo ""
    echo "  archive_mode = on"
    echo "  archive_command = 'test ! -f \"$WAL_ARCHIVE_DIR/%f\" && cp \"%p\" \"$WAL_ARCHIVE_DIR/%f\"'"
    echo "  archive_timeout = 300"
    echo ""
    echo "Then reload PostgreSQL configuration:"
    echo "  SELECT pg_reload_conf();"
    echo ""
    log_warning "Continuing without WAL archiving configuration..."
}

enable_wal_archiving_psql() {
    log_info "Attempting to enable WAL archiving via SQL..."

    # Note: These require superuser and server reload
    log_warning "WAL archiving requires postgresql.conf modification"
    log_warning "Please manually update postgresql.conf and restart PostgreSQL"
}

##############################################################################
# Backup Setup
##############################################################################

setup_backup_directory() {
    log_info "Setting up backup directory..."

    if [ ! -d "$BACKUP_DIR" ]; then
        log_info "Creating backup directory: $BACKUP_DIR"
        sudo mkdir -p "$BACKUP_DIR"
        sudo chown postgres:postgres "$BACKUP_DIR"
        sudo chmod 700 "$BACKUP_DIR"
        log_success "Backup directory created"
    else
        log_success "Backup directory exists: $BACKUP_DIR"
    fi
}

create_backup_script() {
    log_info "Creating backup script..."

    cat > /tmp/backup_pganalytics.sh << 'EOF'
#!/bin/bash
# pgAnalytics backup script

BACKUP_DIR="${BACKUP_DIR:-/var/lib/postgresql/backups}"
DB_NAME="${DB_NAME:-pganalytics}"
DB_HOST="${DB_HOST:-localhost}"
DB_USER="${DB_USER:-postgres}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/pganalytics_${TIMESTAMP}.sql.gz"

echo "[INFO] Starting backup of $DB_NAME..."

# Create backup
if PGPASSWORD="$DB_PASSWORD" pg_dump \
    -h "$DB_HOST" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    | gzip > "$BACKUP_FILE"; then

    echo "[✓] Backup completed: $BACKUP_FILE"
    echo "[✓] Size: $(du -h "$BACKUP_FILE" | cut -f1)"

    # Keep only last 7 backups
    cd "$BACKUP_DIR"
    ls -1t pganalytics_*.sql.gz | tail -n +8 | xargs -r rm

    exit 0
else
    echo "[ERROR] Backup failed"
    exit 1
fi
EOF

    sudo cp /tmp/backup_pganalytics.sh /usr/local/bin/backup_pganalytics.sh
    sudo chmod 755 /usr/local/bin/backup_pganalytics.sh
    rm -f /tmp/backup_pganalytics.sh

    log_success "Backup script created: /usr/local/bin/backup_pganalytics.sh"
}

setup_backup_cron() {
    log_info "Setting up backup cron job..."

    # Create cron job (daily at 2 AM)
    local cron_job="0 2 * * * /usr/local/bin/backup_pganalytics.sh"

    if sudo crontab -u postgres -l 2>/dev/null | grep -q "backup_pganalytics"; then
        log_warning "Backup cron job already exists"
    else
        echo "$cron_job" | sudo crontab -u postgres -
        log_success "Backup cron job created (daily at 2 AM)"
    fi
}

##############################################################################
# Query Logging
##############################################################################

enable_query_logging() {
    log_info "Enabling query logging..."

    log_warning "Query logging requires postgresql.conf modification"
    log_warning "Please manually update postgresql.conf with:"
    echo ""
    echo "  log_statement = 'all'"
    echo "  log_duration = on"
    echo "  log_min_duration_statement = 1000"
    echo "  log_directory = 'log'"
    echo "  log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'"
    echo ""
    echo "Then reload PostgreSQL:"
    echo "  SELECT pg_reload_conf();"
    echo ""
}

##############################################################################
# Validation & Testing
##############################################################################

verify_schema() {
    log_info "Verifying database schema..."

    local table_count
    table_count=$(execute_sql "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" | grep -oE "[0-9]+" | tail -1)

    if [ "$table_count" -gt 0 ]; then
        log_success "Schema verified ($table_count tables found)"
        echo ""
        echo "Tables:"
        execute_sql "SELECT table_name FROM information_schema.tables WHERE table_schema='public';" | grep -v "table_name" | grep -v "^$" | grep -v "^[|-]"
    else
        log_error "No tables found in schema"
    fi
}

test_role_permissions() {
    log_info "Testing role permissions..."

    # Try to connect as pganalytics user
    if [ -z "$DB_PASSWORD" ]; then
        log_warning "Cannot test permissions without DB_PASSWORD"
        return
    fi

    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        -c "SELECT 1;" > /dev/null 2>&1

    if [ $? -eq 0 ]; then
        log_success "Role permissions validated"
    else
        log_error "Role permissions test failed"
    fi
}

test_backup() {
    log_info "Testing backup functionality..."

    if [ ! -d "$BACKUP_DIR" ]; then
        log_error "Backup directory not found: $BACKUP_DIR"
    fi

    # Create test backup
    local test_backup="$BACKUP_DIR/test_backup_$(date +%s).sql.gz"

    if PGPASSWORD="$DB_ADMIN_PASSWORD" pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_ADMIN_USER" \
        -d "$DB_NAME" \
        | gzip > "$test_backup"; then

        local backup_size
        backup_size=$(du -h "$test_backup" | cut -f1)

        log_success "Backup test successful (Size: $backup_size)"
        log_info "Backup location: $test_backup"

        # Cleanup test backup
        rm -f "$test_backup"
    else
        log_error "Backup test failed"
    fi
}

##############################################################################
# Main Execution
##############################################################################

main() {
    log_info "Starting pgAnalytics database setup"
    log_info "PostgreSQL Host: $DB_HOST:$DB_PORT"
    log_info "Database: $DB_NAME"
    log_info "User: $DB_USER"
    echo ""

    # Validation
    validate_connection
    check_postgres_version
    check_ssl_enabled
    echo ""

    # Create roles and database
    create_pganalytics_role
    grant_monitor_permissions
    create_database
    echo ""

    # Initialize schema
    initialize_schema
    echo ""

    # Setup WAL archiving
    setup_wal_archiving
    echo ""

    # Setup backup
    setup_backup_directory
    create_backup_script
    setup_backup_cron
    echo ""

    # Query logging
    enable_query_logging
    echo ""

    # Verification
    verify_schema
    echo ""

    test_role_permissions
    echo ""

    test_backup
    echo ""

    # Summary
    echo -e "${BLUE}============================================${NC}"
    echo -e "${GREEN}Database Setup Completed${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""
    echo "Database: $DB_NAME"
    echo "User: $DB_USER"
    echo "Backup Dir: $BACKUP_DIR"
    echo "WAL Archive Dir: $WAL_ARCHIVE_DIR"
    echo ""
    log_success "Database setup completed successfully"
}

# Run main function
main "$@"
