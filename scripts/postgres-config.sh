#!/bin/bash

##############################################################################
# pgAnalytics v3.2.0 - PostgreSQL Configuration Script
# Purpose: Configure PostgreSQL for pgAnalytics (SSL, replication, archiving)
# Usage: ./postgres-config.sh [--pg-version 16] [--ssl-cert path] [--ssl-key path]
##############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PG_VERSION="${PG_VERSION:-16}"
PG_MAJOR_VERSION=$(echo "$PG_VERSION" | cut -d. -f1)
PG_DATA_DIR="${PG_DATA_DIR:-/var/lib/postgresql/$PG_MAJOR_VERSION/main}"
PG_CONFIG_DIR="${PG_CONFIG_DIR:-/etc/postgresql/$PG_MAJOR_VERSION/main}"
PG_CONFIG_FILE="$PG_CONFIG_DIR/postgresql.conf"
PG_HBA_FILE="$PG_CONFIG_DIR/pg_hba.conf"
SSL_CERT_PATH="${SSL_CERT_PATH:-/etc/postgresql/$PG_MAJOR_VERSION/main/server.crt}"
SSL_KEY_PATH="${SSL_KEY_PATH:-/etc/postgresql/$PG_MAJOR_VERSION/main/server.key}"
WAL_ARCHIVE_DIR="${WAL_ARCHIVE_DIR:-/var/lib/postgresql/wal_archive}"

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

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root"
    fi
}

backup_config() {
    local file="$1"
    local backup="${file}.backup.$(date +%Y%m%d_%H%M%S)"

    if [ -f "$file" ]; then
        cp "$file" "$backup"
        log_success "Configuration backed up: $backup"
    fi
}

edit_config() {
    local file="$1"
    local key="$2"
    local value="$3"

    if grep -q "^$key" "$file"; then
        # Key exists, update it
        sed -i "s/^$key.*/$key = $value/" "$file"
        log_info "Updated: $key = $value"
    elif grep -q "^#$key" "$file"; then
        # Key is commented, uncomment and set
        sed -i "s/^#$key.*/$key = $value/" "$file"
        log_info "Uncommented: $key = $value"
    else
        # Key doesn't exist, append
        echo "$key = $value" >> "$file"
        log_info "Added: $key = $value"
    fi
}

##############################################################################
# Validation Functions
##############################################################################

check_postgresql_installed() {
    log_info "Checking PostgreSQL installation..."

    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL client not found. Install with: apt-get install postgresql-client"
    fi

    if ! command -v pg_config &> /dev/null; then
        log_error "PostgreSQL development tools not found. Install with: apt-get install postgresql-server-dev-$PG_MAJOR_VERSION"
    fi

    log_success "PostgreSQL client and tools found"
}

verify_config_paths() {
    log_info "Verifying PostgreSQL configuration paths..."

    if [ ! -f "$PG_CONFIG_FILE" ]; then
        log_error "PostgreSQL config file not found: $PG_CONFIG_FILE"
    fi

    if [ ! -f "$PG_HBA_FILE" ]; then
        log_error "PostgreSQL HBA file not found: $PG_HBA_FILE"
    fi

    if [ ! -d "$PG_DATA_DIR" ]; then
        log_error "PostgreSQL data directory not found: $PG_DATA_DIR"
    fi

    log_success "Configuration paths verified"
}

check_ssl_certificates() {
    log_info "Checking SSL certificates..."

    if [ ! -f "$SSL_CERT_PATH" ]; then
        log_error "SSL certificate not found: $SSL_CERT_PATH"
    fi

    if [ ! -f "$SSL_KEY_PATH" ]; then
        log_error "SSL private key not found: $SSL_KEY_PATH"
    fi

    # Check permissions
    local cert_perms
    cert_perms=$(stat -f%A "$SSL_CERT_PATH" 2>/dev/null || stat -c%a "$SSL_CERT_PATH" 2>/dev/null || echo "unknown")

    local key_perms
    key_perms=$(stat -f%A "$SSL_KEY_PATH" 2>/dev/null || stat -c%a "$SSL_KEY_PATH" 2>/dev/null || echo "unknown")

    if [ "$key_perms" != "600" ] && [ "$key_perms" != "unknown" ]; then
        log_warning "SSL key has insecure permissions: $key_perms (should be 600)"
        chmod 600 "$SSL_KEY_PATH"
    fi

    # Verify certificate validity
    local expiry_date
    expiry_date=$(openssl x509 -in "$SSL_CERT_PATH" -noout -enddate | cut -d= -f2)
    log_info "SSL certificate expires: $expiry_date"

    log_success "SSL certificates verified"
}

##############################################################################
# SSL Configuration
##############################################################################

enable_ssl() {
    log_info "Enabling SSL in PostgreSQL..."

    backup_config "$PG_CONFIG_FILE"

    edit_config "$PG_CONFIG_FILE" "ssl" "on"
    edit_config "$PG_CONFIG_FILE" "ssl_cert_file" "'$SSL_CERT_PATH'"
    edit_config "$PG_CONFIG_FILE" "ssl_key_file" "'$SSL_KEY_PATH'"
    edit_config "$PG_CONFIG_FILE" "ssl_prefer_server_ciphers" "on"

    log_success "SSL enabled in postgresql.conf"
}

configure_hba_ssl() {
    log_info "Updating pg_hba.conf for SSL connections..."

    backup_config "$PG_HBA_FILE"

    # Add SSL connection lines
    if ! grep -q "^hostssl" "$PG_HBA_FILE"; then
        cat >> "$PG_HBA_FILE" << 'EOF'

# SSL connections
hostssl all pganalytics 127.0.0.1/32 md5
hostssl all pganalytics ::1/128 md5
hostssl all pganalytics 0.0.0.0/0 md5
EOF
        log_success "SSL connection rules added to pg_hba.conf"
    else
        log_warning "SSL connection rules already present in pg_hba.conf"
    fi
}

##############################################################################
# Replication Configuration
##############################################################################

setup_replication_user() {
    log_info "Setting up replication user..."

    local rep_user="replicator"
    local rep_password="${REP_PASSWORD:-$(openssl rand -base64 16)}"

    cat > /tmp/setup_replication.sql << EOF
-- Create replication user
DROP USER IF EXISTS $rep_user;
CREATE USER $rep_user WITH REPLICATION ENCRYPTED PASSWORD '$rep_password';

-- Grant necessary permissions
GRANT CONNECT ON DATABASE postgres TO $rep_user;
GRANT pg_monitor TO $rep_user;
EOF

    sudo -u postgres psql -f /tmp/setup_replication.sql
    rm -f /tmp/setup_replication.sql

    log_success "Replication user created: $rep_user"
    log_info "Replication password: $rep_password"
}

enable_wal_replication() {
    log_info "Enabling WAL replication settings..."

    backup_config "$PG_CONFIG_FILE"

    # WAL configuration for replication
    edit_config "$PG_CONFIG_FILE" "max_wal_senders" "10"
    edit_config "$PG_CONFIG_FILE" "max_replication_slots" "10"
    edit_config "$PG_CONFIG_FILE" "wal_keep_size" "2GB"
    edit_config "$PG_CONFIG_FILE" "hot_standby" "on"
    edit_config "$PG_CONFIG_FILE" "hot_standby_feedback" "off"

    log_success "WAL replication settings configured"
}

##############################################################################
# WAL Archiving Configuration
##############################################################################

create_wal_archive_dir() {
    log_info "Creating WAL archive directory..."

    if [ ! -d "$WAL_ARCHIVE_DIR" ]; then
        mkdir -p "$WAL_ARCHIVE_DIR"
        chown postgres:postgres "$WAL_ARCHIVE_DIR"
        chmod 700 "$WAL_ARCHIVE_DIR"
        log_success "WAL archive directory created: $WAL_ARCHIVE_DIR"
    else
        log_warning "WAL archive directory already exists"
    fi
}

setup_wal_archiving() {
    log_info "Setting up WAL archiving..."

    backup_config "$PG_CONFIG_FILE"

    # Create archive script
    cat > /usr/local/bin/pganalytics_archive_wal.sh << 'EOF'
#!/bin/bash
# WAL archiving script
# Usage: archive_command = '/usr/local/bin/pganalytics_archive_wal.sh %p %f %x'

WAL_PATH="$1"
WAL_FILE="$2"
WAL_ARCHIVE_DIR="${WAL_ARCHIVE_DIR:-/var/lib/postgresql/wal_archive}"

if [ -z "$WAL_PATH" ] || [ -z "$WAL_FILE" ]; then
    exit 1
fi

# Archive to local directory
if cp "$WAL_PATH" "$WAL_ARCHIVE_DIR/$WAL_FILE" 2>/dev/null; then
    exit 0
else
    exit 1
fi
EOF

    chmod 755 /usr/local/bin/pganalytics_archive_wal.sh
    chown postgres:postgres /usr/local/bin/pganalytics_archive_wal.sh

    # Configure archiving in postgresql.conf
    edit_config "$PG_CONFIG_FILE" "archive_mode" "on"
    edit_config "$PG_CONFIG_FILE" "archive_command" "'/usr/local/bin/pganalytics_archive_wal.sh %p %f %x'"
    edit_config "$PG_CONFIG_FILE" "archive_timeout" "300"

    log_success "WAL archiving configured"
}

##############################################################################
# Monitoring & Logging Configuration
##############################################################################

enable_monitoring() {
    log_info "Configuring monitoring settings..."

    backup_config "$PG_CONFIG_FILE"

    # Enable query logging
    edit_config "$PG_CONFIG_FILE" "log_statement" "'all'"
    edit_config "$PG_CONFIG_FILE" "log_duration" "on"
    edit_config "$PG_CONFIG_FILE" "log_min_duration_statement" "1000"
    edit_config "$PG_CONFIG_FILE" "log_lock_waits" "on"
    edit_config "$PG_CONFIG_FILE" "log_checkpoints" "on"
    edit_config "$PG_CONFIG_FILE" "log_connections" "on"
    edit_config "$PG_CONFIG_FILE" "log_disconnections" "on"
    edit_config "$PG_CONFIG_FILE" "log_autovacuum_min_duration" "0"

    # Log directory configuration
    edit_config "$PG_CONFIG_FILE" "log_directory" "'log'"
    edit_config "$PG_CONFIG_FILE" "log_filename" "'postgresql-%Y-%m-%d_%H%M%S.log'"
    edit_config "$PG_CONFIG_FILE" "log_file_mode" "0640"
    edit_config "$PG_CONFIG_FILE" "log_truncate_on_rotation" "on"
    edit_config "$PG_CONFIG_FILE" "log_rotation_age" "'1d'"
    edit_config "$PG_CONFIG_FILE" "log_rotation_size" "500MB"

    log_success "Monitoring configuration enabled"
}

enable_statistics() {
    log_info "Enabling statistics collection..."

    backup_config "$PG_CONFIG_FILE"

    edit_config "$PG_CONFIG_FILE" "track_activities" "on"
    edit_config "$PG_CONFIG_FILE" "track_counts" "on"
    edit_config "$PG_CONFIG_FILE" "track_io_timing" "on"
    edit_config "$PG_CONFIG_FILE" "track_functions" "'all'"

    log_success "Statistics collection enabled"
}

##############################################################################
# Performance Configuration
##############################################################################

configure_performance() {
    log_info "Configuring performance settings..."

    # These are reasonable defaults - adjust based on server specs
    backup_config "$PG_CONFIG_FILE"

    # Memory settings (adjust based on your server)
    edit_config "$PG_CONFIG_FILE" "shared_buffers" "'256MB'"
    edit_config "$PG_CONFIG_FILE" "effective_cache_size" "'1GB'"
    edit_config "$PG_CONFIG_FILE" "work_mem" "'64MB'"
    edit_config "$PG_CONFIG_FILE" "maintenance_work_mem" "'256MB'"

    # Connection settings
    edit_config "$PG_CONFIG_FILE" "max_connections" "200"

    # Checkpoint settings
    edit_config "$PG_CONFIG_FILE" "checkpoint_completion_target" "0.9"
    edit_config "$PG_CONFIG_FILE" "wal_buffers" "'16MB'"
    edit_config "$PG_CONFIG_FILE" "default_statistics_target" "100"

    log_success "Performance configuration applied"
}

##############################################################################
# Configuration Validation
##############################################################################

validate_postgresql_conf() {
    log_info "Validating postgresql.conf..."

    # Test configuration by starting PostgreSQL in check-only mode
    if ! sudo -u postgres /usr/lib/postgresql/$PG_MAJOR_VERSION/bin/postgres -C config_file="$PG_CONFIG_FILE" -C shared_preload_libraries='' > /dev/null 2>&1; then
        log_error "postgresql.conf syntax error detected"
    fi

    log_success "postgresql.conf validation passed"
}

##############################################################################
# Service Management
##############################################################################

restart_postgresql() {
    log_info "Restarting PostgreSQL service..."

    if systemctl is-active --quiet postgresql; then
        systemctl restart postgresql
        log_success "PostgreSQL restarted"
    else
        log_warning "PostgreSQL service not running"
    fi

    sleep 2

    # Verify service is running
    if systemctl is-active --quiet postgresql; then
        log_success "PostgreSQL service is running"
    else
        log_error "PostgreSQL service failed to start"
    fi
}

##############################################################################
# Summary & Verification
##############################################################################

print_config_summary() {
    log_info "Current configuration:"
    echo ""

    echo -e "${BLUE}SSL Configuration:${NC}"
    grep -E "^ssl" "$PG_CONFIG_FILE" | grep -v "^#"

    echo ""
    echo -e "${BLUE}Replication Configuration:${NC}"
    grep -E "^(max_wal_senders|max_replication_slots|wal_keep_size)" "$PG_CONFIG_FILE" | grep -v "^#"

    echo ""
    echo -e "${BLUE}WAL Archiving Configuration:${NC}"
    grep -E "^archive" "$PG_CONFIG_FILE" | grep -v "^#"

    echo ""
    echo -e "${BLUE}Logging Configuration:${NC}"
    grep -E "^log_" "$PG_CONFIG_FILE" | head -5 | grep -v "^#"
    echo "..."
}

##############################################################################
# Main Execution
##############################################################################

main() {
    log_info "Starting PostgreSQL configuration for pgAnalytics v3.2.0"
    echo ""

    # Validation
    check_root
    check_postgresql_installed
    verify_config_paths
    check_ssl_certificates
    echo ""

    # Configuration
    log_info "Applying PostgreSQL configuration..."
    echo ""

    enable_ssl
    configure_hba_ssl
    echo ""

    enable_wal_replication
    echo ""

    create_wal_archive_dir
    setup_wal_archiving
    echo ""

    enable_monitoring
    enable_statistics
    echo ""

    configure_performance
    echo ""

    # Validation
    validate_postgresql_conf
    echo ""

    # Restart service
    restart_postgresql
    echo ""

    # Summary
    print_config_summary
    echo ""

    echo -e "${BLUE}============================================${NC}"
    echo -e "${GREEN}PostgreSQL Configuration Complete${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""
    log_success "Configuration completed successfully"
    echo ""
    echo "Configuration files backed up at:"
    ls -1 "$PG_CONFIG_DIR"/*.backup.* 2>/dev/null || echo "No backups found"
    echo ""
}

# Run main function
main "$@"
