# pgAnalytics v3.2.0 - Deployment Configuration Template

**Purpose:** Define all infrastructure parameters for any deployment environment (AWS, on-premises, physical machines, etc.)

**Status:** Template - Copy and fill with your environment values

---

## Environment Variables Configuration

Create a file named `.env.deployment` with your specific environment values. This file is used by all deployment scripts.

```bash
# .env.deployment - Your deployment configuration
# Copy this section and fill with YOUR values, then: source .env.deployment

################################################################################
# INFRASTRUCTURE PARAMETERS (Required)
################################################################################

# Database Configuration
export DB_HOST="your-postgres-host"              # IP or hostname of PostgreSQL server
export DB_PORT="5432"                             # PostgreSQL port (default: 5432)
export DB_NAME="pganalytics"                      # Database name
export DB_USER="pganalytics"                      # Database user
export DB_PASSWORD="generate-secure-password"     # 32-byte random password (use: openssl rand -base64 32)
export DB_ADMIN_USER="admin"                      # Admin user for initial setup
export DB_ADMIN_PASSWORD="admin-secure-password"  # Admin password

# API Server Configuration
export API_HOST_1="your-api-server-1-ip"          # IP/hostname of first API server
export API_HOST_2="your-api-server-2-ip"          # IP/hostname of second API server
export API_PORT="8080"                            # API port (default: 8080)
export API_USER="pganalytics"                     # User to run API service
export API_INSTALL_PATH="/opt/pganalytics"        # Installation directory

# Collector Configuration
export COLLECTOR_HOSTS=(                          # Array of collector hosts
  "collector-1-ip"
  "collector-2-ip"
  "collector-3-ip"
  "collector-4-ip"
  "collector-5-ip"
)
export COLLECTOR_USER="pganalytics"               # User to run collectors
export COLLECTOR_INSTALL_PATH="/opt/pganalytics"  # Installation directory

# Grafana Configuration
export GRAFANA_HOST="your-grafana-host-ip"        # IP/hostname of Grafana server
export GRAFANA_PORT="3000"                        # Grafana port (default: 3000)
export GRAFANA_ADMIN_PASSWORD="grafana-password"  # Admin password for Grafana

# Monitoring Configuration (Prometheus)
export PROMETHEUS_HOST="your-prometheus-host-ip"  # IP/hostname of Prometheus server
export PROMETHEUS_PORT="9090"                     # Prometheus port (default: 9090)

################################################################################
# SECURITY PARAMETERS (Required)
################################################################################

# JWT Secrets (Generate with: openssl rand -base64 32)
export JWT_SECRET_KEY="generate-with-openssl"              # 32-byte random secret
export REGISTRATION_SECRET="generate-with-openssl"         # 32-byte random secret
export BACKUP_KEY="generate-with-openssl"                  # 32-byte random secret

# TLS/SSL Configuration
export TLS_ENABLED="true"                         # Enable TLS (true/false)
export TLS_CERT_PATH="/etc/pganalytics/cert.pem"  # Path to TLS certificate
export TLS_KEY_PATH="/etc/pganalytics/key.pem"    # Path to TLS private key
export TLS_CA_PATH="/etc/pganalytics/ca.pem"      # Path to CA certificate (if using CA-signed)

# mTLS Configuration for Collectors
export MTLS_ENABLED="true"                        # Enable mutual TLS (true/false)
export MTLS_CLIENT_CERT="/etc/pganalytics/client.crt"  # Client certificate
export MTLS_CLIENT_KEY="/etc/pganalytics/client.key"   # Client private key

################################################################################
# CLOUD PROVIDER PARAMETERS (Optional - only if using cloud services)
################################################################################

# AWS Configuration (if using AWS)
export AWS_REGION="us-east-1"                     # AWS region (change to your region)
export AWS_ACCOUNT_ID="123456789012"              # Your AWS account ID
export AWS_RDS_INSTANCE_ID="pganalytics-prod"    # RDS instance identifier (if using RDS)
export AWS_RDS_ENGINE="postgres"                  # RDS engine (postgres)
export AWS_RDS_VERSION="16.12"                    # RDS engine version
export AWS_SECRETS_MANAGER_ENABLED="true"         # Use AWS Secrets Manager (true/false)
export AWS_PROFILE="default"                      # AWS CLI profile to use

# Alternative: Self-hosted or on-premises (fill these instead)
export SELF_HOSTED="false"                        # Set to true if not using AWS

################################################################################
# DEPLOYMENT PARAMETERS (Required)
################################################################################

# Environment
export ENVIRONMENT="production"                   # Environment name (production/staging/dev)
export DEPLOYMENT_NAME="pganalytics-v3.2.0"       # Deployment identifier

# Deployment Mode
export DEPLOYMENT_MODE="distributed"              # Options: distributed, single-machine, kubernetes
export CONTAINER_RUNTIME="none"                   # Options: none, docker, kubernetes

# Logging
export LOG_LEVEL="info"                           # Log level (debug/info/warn/error)
export LOG_PATH="/var/log/pganalytics"            # Log directory
export LOG_RETENTION_DAYS="30"                    # Log retention

# Backup Configuration
export BACKUP_ENABLED="true"                      # Enable automated backups
export BACKUP_RETENTION_DAYS="30"                 # Backup retention period
export BACKUP_SCHEDULE="0 3 * * *"                # Backup schedule (cron format: 3 AM daily)

################################################################################
# PERFORMANCE PARAMETERS (Optional - defaults are good for most cases)
################################################################################

export API_MAX_CONNECTIONS="100"                  # Max concurrent API connections
export DB_CONNECTION_POOL_SIZE="20"               # Database connection pool size
export METRICS_BATCH_SIZE="100"                   # Batch size for metric collection
export METRICS_FLUSH_INTERVAL="60"                # Metrics flush interval (seconds)
export COLLECTOR_TIMEOUT="30"                     # Collector timeout (seconds)
export RATE_LIMIT_REQUESTS="1000"                 # Rate limit: requests per minute
export RATE_LIMIT_WINDOW="60"                     # Rate limit window (seconds)

################################################################################
# NOTIFICATION PARAMETERS (Optional - for alerting)
################################################################################

# Slack Integration
export SLACK_WEBHOOK_URL=""                       # Slack webhook URL (leave empty to disable)
export SLACK_CHANNEL="#alerts"                    # Slack channel for alerts

# Email Integration
export EMAIL_ENABLED="false"                      # Enable email notifications
export EMAIL_FROM="pganalytics@example.com"       # From email address
export EMAIL_TO="ops-team@example.com"            # To email address
export SMTP_HOST="smtp.example.com"               # SMTP server
export SMTP_PORT="587"                            # SMTP port
export SMTP_USER="smtp-user"                      # SMTP username
export SMTP_PASSWORD="smtp-password"              # SMTP password

# PagerDuty Integration
export PAGERDUTY_ENABLED="false"                  # Enable PagerDuty integration
export PAGERDUTY_KEY=""                           # PagerDuty integration key

################################################################################
# NETWORK PARAMETERS (Optional - for firewalls/security groups)
################################################################################

# Allowed IP Ranges (CIDR notation)
export ALLOWED_API_CLIENTS="0.0.0.0/0"            # IPs allowed to call API
export ALLOWED_COLLECTOR_CLIENTS="0.0.0.0/0"      # IPs allowed from collectors
export ALLOWED_GRAFANA_CLIENTS="0.0.0.0/0"        # IPs allowed to access Grafana

# Network Timeouts
export NETWORK_TIMEOUT="30"                       # Network timeout (seconds)
export CONNECTION_TIMEOUT="10"                    # Connection timeout (seconds)

################################################################################
# POSTGRESQL SPECIFIC PARAMETERS (Optional)
################################################################################

export PG_STAT_STATEMENTS_ENABLED="true"          # Enable pg_stat_statements
export PG_LOG_STATEMENT="all"                     # Log all statements
export PG_LOG_DURATION="on"                       # Log query duration
export PG_LOG_MIN_DURATION="1000"                 # Minimum duration to log (ms)
export PG_SHARED_BUFFERS="256MB"                  # PostgreSQL shared_buffers
export PG_EFFECTIVE_CACHE_SIZE="1GB"              # PostgreSQL effective_cache_size
export PG_WORK_MEM="4MB"                          # PostgreSQL work_mem

################################################################################
# KUBERNETES SPECIFIC (Optional - only if DEPLOYMENT_MODE=kubernetes)
################################################################################

export K8S_NAMESPACE="pganalytics"                # Kubernetes namespace
export K8S_CONTEXT="default"                      # Kubernetes context
export K8S_STORAGE_CLASS="standard"               # Storage class for PVCs
export K8S_STORAGE_SIZE="100Gi"                   # Storage size for database

################################################################################
# DOCKER SPECIFIC (Optional - only if CONTAINER_RUNTIME=docker)
################################################################################

export DOCKER_REGISTRY="docker.io"                # Docker registry
export DOCKER_IMAGE_TAG="v3.2.0"                  # Docker image tag
export DOCKER_NETWORK="pganalytics-network"       # Docker network name

################################################################################
# INTERNAL VARIABLES (Set automatically - do not change)
################################################################################

export DEPLOYMENT_ID="pganalytics-$(date +%s)"    # Unique deployment ID
export DEPLOYMENT_DATE="$(date +%Y-%m-%d)"        # Deployment date
export DEPLOYMENT_TIME="$(date +%H:%M:%S)"        # Deployment time
```

---

## How to Use This Template

### Step 1: Copy the Template
```bash
# Copy template to your environment
cp DEPLOYMENT_CONFIG_TEMPLATE.md /etc/pganalytics/deployment.env

# Or create a new file with only the values you need
cat > ~/.env.pganalytics << 'EOF'
# Paste the configuration above and fill with YOUR values
EOF
```

### Step 2: Fill in Your Values

Open the file and replace all placeholder values with your actual infrastructure:

```bash
# Example: AWS environment
export DB_HOST="pganalytics-prod.abc123.us-east-1.rds.amazonaws.com"
export DB_PASSWORD="$(openssl rand -base64 32)"
export API_HOST_1="10.0.1.50"
export API_HOST_2="10.0.1.51"
export COLLECTOR_HOSTS=("10.0.2.10" "10.0.2.11" "10.0.2.12" "10.0.2.13" "10.0.2.14")
export GRAFANA_HOST="10.0.3.100"
export AWS_REGION="us-east-1"
export JWT_SECRET_KEY="$(openssl rand -base64 32)"
```

**Or for on-premises/physical machines:**

```bash
# Example: On-premises environment
export DB_HOST="192.168.1.50"
export DB_PASSWORD="$(openssl rand -base64 32)"
export API_HOST_1="192.168.1.51"
export API_HOST_2="192.168.1.52"
export COLLECTOR_HOSTS=("192.168.1.60" "192.168.1.61" "192.168.1.62" "192.168.1.63" "192.168.1.64")
export GRAFANA_HOST="192.168.1.70"
export DEPLOYMENT_MODE="distributed"
export CONTAINER_RUNTIME="none"
export JWT_SECRET_KEY="$(openssl rand -base64 32)"
```

### Step 3: Source the Configuration

Before running any deployment scripts:

```bash
# Load your configuration
source ~/.env.pganalytics

# Verify configuration is loaded
echo "Database: $DB_HOST:$DB_PORT"
echo "API Servers: $API_HOST_1, $API_HOST_2"
echo "Collectors: ${COLLECTOR_HOSTS[@]}"
```

### Step 4: Use in Deployment Scripts

All deployment scripts will read from these environment variables:

```bash
# Example deployment script
#!/bin/bash
source ~/.env.pganalytics

# Now use the variables
psql -h "$DB_HOST" -U "$DB_ADMIN_USER" -d "$DB_NAME" << EOF
  CREATE ROLE "$DB_USER" WITH LOGIN PASSWORD '$DB_PASSWORD';
EOF

# SSH to API servers and deploy
for api_host in "$API_HOST_1" "$API_HOST_2"; do
  ssh -i ~/.ssh/id_rsa "$API_USER@$api_host" << 'SCRIPT'
    sudo cp /tmp/pganalytics-api "${{ API_INSTALL_PATH }}/"
  SCRIPT
done
```

---

## Configuration Examples for Different Environments

### Example 1: AWS EC2 Distributed Deployment

```bash
# AWS environment - Multiple EC2 instances
export DB_HOST="pganalytics-prod.abc123.us-east-1.rds.amazonaws.com"
export DB_PORT="5432"
export DB_NAME="pganalytics"
export DB_USER="pganalytics"
export DB_PASSWORD="very-secure-32-byte-password"
export DB_ADMIN_USER="admin"
export DB_ADMIN_PASSWORD="admin-secure-password"

export API_HOST_1="10.0.1.50"
export API_HOST_2="10.0.1.51"
export API_PORT="8080"
export API_USER="ubuntu"
export API_INSTALL_PATH="/opt/pganalytics"

export COLLECTOR_HOSTS=(
  "10.0.2.10"   # us-east-1a
  "10.0.2.11"   # us-east-1b
  "10.0.2.12"   # us-west-1
  "10.0.2.13"   # eu-west-1
  "10.0.2.14"   # ap-southeast-1
)

export GRAFANA_HOST="10.0.3.100"
export PROMETHEUS_HOST="10.0.3.101"

export AWS_REGION="us-east-1"
export AWS_ACCOUNT_ID="123456789012"
export AWS_SECRETS_MANAGER_ENABLED="true"

export DEPLOYMENT_MODE="distributed"
export ENVIRONMENT="production"

export JWT_SECRET_KEY="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_KEY="$(openssl rand -base64 32)"

export TLS_ENABLED="true"
export MTLS_ENABLED="true"
```

### Example 2: On-Premises Distributed (Physical Machines)

```bash
# On-premises with physical machines in your data center
export DB_HOST="postgres-server.internal.company.com"
export DB_PORT="5432"
export DB_NAME="pganalytics"
export DB_USER="pganalytics"
export DB_PASSWORD="$(openssl rand -base64 32)"
export DB_ADMIN_USER="postgres"
export DB_ADMIN_PASSWORD="your-postgres-password"

export API_HOST_1="api-01.internal.company.com"
export API_HOST_2="api-02.internal.company.com"
export API_PORT="8080"
export API_USER="pganalytics"
export API_INSTALL_PATH="/opt/pganalytics"

export COLLECTOR_HOSTS=(
  "collector-01.internal.company.com"
  "collector-02.internal.company.com"
  "collector-03.internal.company.com"
  "collector-04.internal.company.com"
  "collector-05.internal.company.com"
)

export GRAFANA_HOST="monitoring.internal.company.com"
export PROMETHEUS_HOST="monitoring.internal.company.com"

export DEPLOYMENT_MODE="distributed"
export CONTAINER_RUNTIME="none"
export ENVIRONMENT="production"

export JWT_SECRET_KEY="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_KEY="$(openssl rand -base64 32)"

export TLS_ENABLED="true"
export TLS_CERT_PATH="/etc/pganalytics/certs/server.crt"
export TLS_KEY_PATH="/etc/pganalytics/certs/server.key"
export MTLS_ENABLED="true"
```

### Example 3: Single Machine (Development/Testing)

```bash
# Single machine - everything on localhost or one server
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="pganalytics"
export DB_USER="pganalytics"
export DB_PASSWORD="dev-password"
export DB_ADMIN_USER="postgres"
export DB_ADMIN_PASSWORD="postgres"

export API_HOST_1="localhost"
export API_HOST_2=""  # Not used in single-machine mode
export API_PORT="8080"
export API_USER="$(whoami)"
export API_INSTALL_PATH="./backend"

export COLLECTOR_HOSTS=(
  "localhost"
)

export GRAFANA_HOST="localhost"
export PROMETHEUS_HOST="localhost"

export DEPLOYMENT_MODE="single-machine"
export CONTAINER_RUNTIME="docker"
export ENVIRONMENT="development"

export TLS_ENABLED="false"
export MTLS_ENABLED="false"
```

### Example 4: Kubernetes Deployment

```bash
# Kubernetes cluster (any cloud or on-premises)
export DB_HOST="pganalytics-postgres"                    # K8s service name
export DB_PORT="5432"
export DB_NAME="pganalytics"
export DB_USER="pganalytics"
export DB_PASSWORD="$(openssl rand -base64 32)"

export API_HOST_1="pganalytics-api"                      # K8s service name
export API_HOST_2="pganalytics-api"                      # Same service, multiple replicas
export API_PORT="8080"

export COLLECTOR_HOSTS=(
  "pganalytics-collector-1"
  "pganalytics-collector-2"
  "pganalytics-collector-3"
  "pganalytics-collector-4"
  "pganalytics-collector-5"
)

export GRAFANA_HOST="pganalytics-grafana"                # K8s service name
export PROMETHEUS_HOST="prometheus"                      # K8s service name

export DEPLOYMENT_MODE="kubernetes"
export CONTAINER_RUNTIME="kubernetes"
export ENVIRONMENT="production"

export K8S_NAMESPACE="pganalytics"
export K8S_CONTEXT="my-cluster"
export K8S_STORAGE_CLASS="fast-ssd"
export K8S_STORAGE_SIZE="100Gi"

export JWT_SECRET_KEY="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_KEY="$(openssl rand -base64 32)"

export TLS_ENABLED="true"
export MTLS_ENABLED="true"
```

### Example 5: Multi-Region Deployment

```bash
# Distributed across multiple regions/locations
export DB_HOST="pganalytics-primary.region1.com"         # Primary database
export DB_REPLICA_HOST="pganalytics-replica.region2.com" # Replica for HA

export API_HOST_1="api-region1.example.com"
export API_HOST_2="api-region2.example.com"

export COLLECTOR_HOSTS=(
  "collector-region1-az1.example.com"
  "collector-region1-az2.example.com"
  "collector-region2-az1.example.com"
  "collector-region2-az2.example.com"
  "collector-region3-az1.example.com"
)

export GRAFANA_HOST="grafana-central.example.com"
export PROMETHEUS_HOST="prometheus-central.example.com"

export DEPLOYMENT_MODE="distributed"
export ENVIRONMENT="production"

export JWT_SECRET_KEY="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_KEY="$(openssl rand -base64 32)"

# Regional configuration
export BACKUP_ENABLED="true"
export BACKUP_RETENTION_DAYS="30"
export BACKUP_SCHEDULE="0 3 * * *"  # 3 AM daily in primary region
```

---

## Validation

After creating your configuration, validate it:

```bash
#!/bin/bash
# validate-deployment-config.sh

source ~/.env.pganalytics

echo "=== Deployment Configuration Validation ==="
echo ""
echo "Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""
echo "API Servers:"
echo "  Server 1: $API_HOST_1:$API_PORT"
echo "  Server 2: $API_HOST_2:$API_PORT"
echo ""
echo "Collectors:"
for i in "${!COLLECTOR_HOSTS[@]}"; do
  echo "  Collector $((i+1)): ${COLLECTOR_HOSTS[$i]}"
done
echo ""
echo "Monitoring:"
echo "  Grafana: $GRAFANA_HOST:$GRAFANA_PORT"
echo "  Prometheus: $PROMETHEUS_HOST:$PROMETHEUS_PORT"
echo ""
echo "Deployment:"
echo "  Mode: $DEPLOYMENT_MODE"
echo "  Environment: $ENVIRONMENT"
echo "  TLS Enabled: $TLS_ENABLED"
echo "  mTLS Enabled: $MTLS_ENABLED"
echo ""
echo "=== Configuration Validation Complete ==="
```

---

## Next Steps

1. **Copy the template:** Create your `.env.pganalytics` file
2. **Fill in values:** Replace all placeholders with your infrastructure
3. **Choose an example:** Use one of the examples above as a starting point
4. **Validate:** Run the validation script to verify configuration
5. **Deploy:** All deployment scripts will use these variables automatically

---

**Notes:**
- Never commit `.env.pganalytics` to version control (add to `.gitignore`)
- Keep the file secure - it contains credentials
- Use strong passwords - generate with: `openssl rand -base64 32`
- For production, use a secrets management system (AWS Secrets Manager, HashiCorp Vault, etc.)
- Test the configuration in a staging environment first
