#!/bin/bash

##############################################################################
# pgAnalytics v3.2.0 - Secrets Generation Script
# Purpose: Generate cryptographically secure secrets for production deployment
# Usage: ./secrets-generation.sh [--vault-type <type>] [--output <file>]
# Supported Vault Types: aws-secrets, hashicorp-vault, k8s-secrets, azure-keyvault
##############################################################################

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
VAULT_TYPE="${1:-hashicorp-vault}"
OUTPUT_FILE="${2:-./secrets.env}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "Required command not found: $1"
    fi
}

##############################################################################
# Validation Functions
##############################################################################

validate_secret_length() {
    local secret="$1"
    local min_length="${2:-32}"

    if [ ${#secret} -lt "$min_length" ]; then
        log_error "Secret too short. Minimum length: $min_length bytes, got: ${#secret}"
    fi
}

validate_base64() {
    local secret="$1"

    if ! echo "$secret" | base64 -d >/dev/null 2>&1; then
        log_error "Secret is not valid base64"
    fi
}

##############################################################################
# Secret Generation Functions
##############################################################################

generate_jwt_secret() {
    log_info "Generating JWT_SECRET_KEY (256-bit = 32 bytes)..."

    local jwt_secret
    jwt_secret=$(openssl rand -base64 32)

    validate_secret_length "$jwt_secret" 32

    echo "$jwt_secret"
    log_success "JWT_SECRET_KEY generated (base64 encoded, ${#jwt_secret} bytes)"
}

generate_registration_secret() {
    log_info "Generating REGISTRATION_SECRET (256-bit = 32 bytes)..."

    local reg_secret
    reg_secret=$(openssl rand -base64 32)

    validate_secret_length "$reg_secret" 32

    echo "$reg_secret"
    log_success "REGISTRATION_SECRET generated (base64 encoded, ${#reg_secret} bytes)"
}

generate_database_password() {
    log_info "Generating secure database password (32 characters)..."

    # Generate password with mixed case, numbers, and special chars
    local db_password
    db_password=$(openssl rand -base64 24 | tr -d "=+/" | cut -c1-24)

    echo "$db_password"
    log_success "Database password generated (${#db_password} characters)"
}

generate_backup_encryption_key() {
    log_info "Generating backup encryption key (256-bit = 32 bytes)..."

    local backup_key
    backup_key=$(openssl rand -hex 32)

    validate_secret_length "$backup_key" 64

    echo "$backup_key"
    log_success "Backup encryption key generated (hex encoded, ${#backup_key} bytes)"
}

##############################################################################
# Vault Storage Functions
##############################################################################

store_in_aws_secrets() {
    log_info "Storing secrets in AWS Secrets Manager..."

    check_command "aws"

    local jwt_secret="$1"
    local reg_secret="$2"
    local db_password="$3"
    local backup_key="$4"

    # Check AWS credentials
    if ! aws sts get-caller-identity >/dev/null 2>&1; then
        log_error "AWS credentials not configured. Run: aws configure"
    fi

    # Store JWT secret
    log_info "Storing JWT_SECRET_KEY..."
    aws secretsmanager create-secret \
        --name "pganalytics/jwt-secret" \
        --description "JWT signing secret for pgAnalytics API" \
        --secret-string "$jwt_secret" \
        --tags Key=Environment,Value=production Key=Application,Value=pganalytics \
        2>/dev/null || log_warning "Secret already exists, updating..."
    aws secretsmanager update-secret \
        --secret-id "pganalytics/jwt-secret" \
        --secret-string "$jwt_secret" 2>/dev/null
    log_success "JWT_SECRET_KEY stored in AWS Secrets Manager"

    # Store registration secret
    log_info "Storing REGISTRATION_SECRET..."
    aws secretsmanager create-secret \
        --name "pganalytics/registration-secret" \
        --description "Collector registration secret" \
        --secret-string "$reg_secret" \
        --tags Key=Environment,Value=production Key=Application,Value=pganalytics \
        2>/dev/null || log_warning "Secret already exists, updating..."
    aws secretsmanager update-secret \
        --secret-id "pganalytics/registration-secret" \
        --secret-string "$reg_secret" 2>/dev/null
    log_success "REGISTRATION_SECRET stored in AWS Secrets Manager"

    # Store database password
    log_info "Storing database password..."
    aws secretsmanager create-secret \
        --name "pganalytics/database-password" \
        --description "PostgreSQL pganalytics user password" \
        --secret-string "$db_password" \
        --tags Key=Environment,Value=production Key=Application,Value=pganalytics \
        2>/dev/null || log_warning "Secret already exists, updating..."
    aws secretsmanager update-secret \
        --secret-id "pganalytics/database-password" \
        --secret-string "$db_password" 2>/dev/null
    log_success "Database password stored in AWS Secrets Manager"

    # Store backup encryption key
    log_info "Storing backup encryption key..."
    aws secretsmanager create-secret \
        --name "pganalytics/backup-encryption-key" \
        --description "Backup encryption key" \
        --secret-string "$backup_key" \
        --tags Key=Environment,Value=production Key=Application,Value=pganalytics \
        2>/dev/null || log_warning "Secret already exists, updating..."
    aws secretsmanager update-secret \
        --secret-id "pganalytics/backup-encryption-key" \
        --secret-string "$backup_key" 2>/dev/null
    log_success "Backup encryption key stored in AWS Secrets Manager"
}

store_in_hashicorp_vault() {
    log_info "Storing secrets in HashiCorp Vault..."

    check_command "vault"

    local jwt_secret="$1"
    local reg_secret="$2"
    local db_password="$3"
    local backup_key="$4"

    # Check Vault connection
    if ! vault status >/dev/null 2>&1; then
        log_error "Cannot connect to Vault. Check VAULT_ADDR and VAULT_TOKEN"
    fi

    log_info "Storing secrets at secret/pganalytics/..."

    vault kv put secret/pganalytics/api \
        jwt_secret="$jwt_secret" \
        registration_secret="$reg_secret"
    log_success "API secrets stored"

    vault kv put secret/pganalytics/database \
        password="$db_password"
    log_success "Database password stored"

    vault kv put secret/pganalytics/backup \
        encryption_key="$backup_key"
    log_success "Backup encryption key stored"
}

store_in_k8s_secrets() {
    log_info "Storing secrets in Kubernetes..."

    check_command "kubectl"

    local jwt_secret="$1"
    local reg_secret="$2"
    local db_password="$3"
    local backup_key="$4"
    local namespace="${5:-pganalytics}"

    # Create namespace if it doesn't exist
    if ! kubectl get namespace "$namespace" >/dev/null 2>&1; then
        log_info "Creating namespace: $namespace"
        kubectl create namespace "$namespace"
    fi

    # Create secret
    kubectl create secret generic pganalytics-secrets \
        --from-literal=jwt-secret="$jwt_secret" \
        --from-literal=registration-secret="$reg_secret" \
        --from-literal=database-password="$db_password" \
        --from-literal=backup-encryption-key="$backup_key" \
        --namespace="$namespace" \
        -o yaml --dry-run=client | kubectl apply -f -

    log_success "Secrets stored in Kubernetes namespace: $namespace"
}

store_in_azure_keyvault() {
    log_info "Storing secrets in Azure Key Vault..."

    check_command "az"

    local jwt_secret="$1"
    local reg_secret="$2"
    local db_password="$3"
    local backup_key="$4"
    local vault_name="${5:-pganalytics-vault}"

    # Check Azure CLI authentication
    if ! az account show >/dev/null 2>&1; then
        log_error "Not logged into Azure. Run: az login"
    fi

    log_info "Storing secrets in vault: $vault_name"

    az keyvault secret set --vault-name "$vault_name" \
        --name "jwt-secret" --value "$jwt_secret"
    log_success "JWT_SECRET_KEY stored"

    az keyvault secret set --vault-name "$vault_name" \
        --name "registration-secret" --value "$reg_secret"
    log_success "REGISTRATION_SECRET stored"

    az keyvault secret set --vault-name "$vault_name" \
        --name "database-password" --value "$db_password"
    log_success "Database password stored"

    az keyvault secret set --vault-name "$vault_name" \
        --name "backup-encryption-key" --value "$backup_key"
    log_success "Backup encryption key stored"
}

##############################################################################
# Output Functions
##############################################################################

write_env_file() {
    local jwt_secret="$1"
    local reg_secret="$2"
    local db_password="$3"
    local backup_key="$4"

    cat > "$OUTPUT_FILE" << 'EOF'
#!/bin/bash
# pgAnalytics v3.2.0 Secrets
# Generated: $(date)
# WARNING: This file contains sensitive information
# Keep secure and do NOT commit to version control
# Use .gitignore to prevent accidental commits

export JWT_SECRET_KEY="JWT_PLACEHOLDER"
export REGISTRATION_SECRET="REG_PLACEHOLDER"
export DATABASE_PASSWORD="DB_PLACEHOLDER"
export BACKUP_ENCRYPTION_KEY="BACKUP_PLACEHOLDER"

# Usage:
# source ./secrets.env
# docker-compose up -d
EOF

    # Replace placeholders
    sed -i.bak "s|JWT_PLACEHOLDER|$jwt_secret|g" "$OUTPUT_FILE"
    sed -i.bak "s|REG_PLACEHOLDER|$reg_secret|g" "$OUTPUT_FILE"
    sed -i.bak "s|DB_PLACEHOLDER|$db_password|g" "$OUTPUT_FILE"
    sed -i.bak "s|BACKUP_PLACEHOLDER|$backup_key|g" "$OUTPUT_FILE"

    # Remove backup file
    rm -f "${OUTPUT_FILE}.bak"

    # Secure file permissions
    chmod 600 "$OUTPUT_FILE"

    log_success "Secrets written to: $OUTPUT_FILE"
    log_warning "IMPORTANT: Keep this file secure and do NOT commit to git"
}

print_summary() {
    local jwt_secret="$1"
    local reg_secret="$2"
    local db_password="$3"
    local backup_key="$4"

    echo ""
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}Secrets Generation Complete${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo ""
    echo -e "${GREEN}✓ JWT_SECRET_KEY${NC} (${#jwt_secret} bytes)"
    echo "   First 16 chars: ${jwt_secret:0:16}..."
    echo ""
    echo -e "${GREEN}✓ REGISTRATION_SECRET${NC} (${#reg_secret} bytes)"
    echo "   First 16 chars: ${reg_secret:0:16}..."
    echo ""
    echo -e "${GREEN}✓ DATABASE_PASSWORD${NC} (${#db_password} chars)"
    echo "   First 8 chars: ${db_password:0:8}..."
    echo ""
    echo -e "${GREEN}✓ BACKUP_ENCRYPTION_KEY${NC} (${#backup_key} bytes)"
    echo "   First 16 chars: ${backup_key:0:16}..."
    echo ""
    echo -e "${BLUE}Storage Type: ${VAULT_TYPE}${NC}"
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo "1. Verify secrets are stored in vault"
    echo "2. Update deployment configurations"
    echo "3. Test secret retrieval from vault"
    echo "4. Proceed with Phase 2 staging deployment"
    echo ""
}

##############################################################################
# Main Execution
##############################################################################

main() {
    log_info "Starting secrets generation for pgAnalytics v3.2.0"
    echo ""

    # Generate all secrets
    JWT_SECRET=$(generate_jwt_secret)
    echo ""

    REG_SECRET=$(generate_registration_secret)
    echo ""

    DB_PASSWORD=$(generate_database_password)
    echo ""

    BACKUP_KEY=$(generate_backup_encryption_key)
    echo ""

    # Store secrets based on vault type
    case "$VAULT_TYPE" in
        aws-secrets)
            store_in_aws_secrets "$JWT_SECRET" "$REG_SECRET" "$DB_PASSWORD" "$BACKUP_KEY"
            ;;
        hashicorp-vault)
            store_in_hashicorp_vault "$JWT_SECRET" "$REG_SECRET" "$DB_PASSWORD" "$BACKUP_KEY"
            ;;
        k8s-secrets)
            store_in_k8s_secrets "$JWT_SECRET" "$REG_SECRET" "$DB_PASSWORD" "$BACKUP_KEY"
            ;;
        azure-keyvault)
            store_in_azure_keyvault "$JWT_SECRET" "$REG_SECRET" "$DB_PASSWORD" "$BACKUP_KEY"
            ;;
        *)
            log_warning "Unknown vault type: $VAULT_TYPE"
            log_info "Writing to environment file instead: $OUTPUT_FILE"
            write_env_file "$JWT_SECRET" "$REG_SECRET" "$DB_PASSWORD" "$BACKUP_KEY"
            ;;
    esac

    echo ""

    # Print summary
    print_summary "$JWT_SECRET" "$REG_SECRET" "$DB_PASSWORD" "$BACKUP_KEY"

    log_success "Secrets generation completed successfully"
}

# Run main function
main "$@"
