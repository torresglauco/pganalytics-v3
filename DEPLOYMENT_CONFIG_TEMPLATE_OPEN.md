# pgAnalytics v3.2.0 - Open Configuration Template

**Purpose:** Define all infrastructure parameters for ANY deployment environment

**Status:** Template - Works with any infrastructure, any scale, any configuration

**Key Principle:** No assumptions about your infrastructure - YOU define everything

---

## Configuration File Structure

Create a file: `~/.env.pganalytics`

```bash
# ~/.env.pganalytics - Your deployment configuration
# Copy this entire section and fill with YOUR values
# This file works with ANY infrastructure (AWS, on-prem, K8s, Docker, hybrid, etc.)

################################################################################
# INFRASTRUCTURE PARAMETERS
# Fill in YOUR hosts, ports, and infrastructure details
# Works with ANY infrastructure - cloud, on-premises, Kubernetes, Docker, etc.
################################################################################

# Database Configuration
# - Can be AWS RDS, on-premises PostgreSQL, Kubernetes service, Docker container, etc.
# - Just provide the hostname/IP and port you use to connect
export DB_HOST="your-database-host"              # Hostname or IP of your PostgreSQL
export DB_PORT="5432"                            # Port your PostgreSQL listens on
export DB_NAME="pganalytics"                     # Database name
export DB_USER="pganalytics"                     # Database user
export DB_PASSWORD="your-secure-password"        # Database password (generate: openssl rand -base64 32)
export DB_ADMIN_USER="postgres"                  # Admin user for initial setup
export DB_ADMIN_PASSWORD="your-admin-password"   # Admin password

# API Server Configuration
# - Can be on EC2, physical machines, Kubernetes, Docker, etc.
# - Just provide how you connect to your API servers
export API_HOST_1="your-api-server-1"            # First API server hostname or IP
export API_HOST_2="your-api-server-2"            # Second API server hostname or IP (for HA)
export API_PORT="8080"                           # Port your API listens on
export API_USER="pganalytics"                    # User to run API service
export API_INSTALL_PATH="/opt/pganalytics"       # Where to install API binary

# Collector Configuration
# - Can be anywhere: EC2, physical machines, Kubernetes pods, Docker containers, etc.
# - List all your collector hosts/IPs
# - Can be 1, 5, 100, or 10000+ collectors - doesn't matter
export COLLECTOR_HOSTS=(
  "your-collector-1"                             # First collector
  "your-collector-2"                             # Second collector
  "your-collector-3"                             # Third collector
  # Add as many as you have: 5, 100, 1000, doesn't matter
)
export COLLECTOR_USER="pganalytics"              # User to run collectors
export COLLECTOR_INSTALL_PATH="/opt/pganalytics" # Where to install collector

# Monitoring Configuration
# - Grafana: Any server running Grafana (AWS, on-prem, K8s, Docker, etc.)
# - Prometheus: Any server running Prometheus
# - Can be same machine, different machines, anywhere
export GRAFANA_HOST="your-grafana-host"          # Hostname or IP of Grafana server
export GRAFANA_PORT="3000"                       # Port Grafana listens on
export GRAFANA_ADMIN_PASSWORD="your-grafana-password"

export PROMETHEUS_HOST="your-prometheus-host"    # Hostname or IP of Prometheus server
export PROMETHEUS_PORT="9090"                    # Port Prometheus listens on

################################################################################
# SECURITY PARAMETERS
# Generate random secrets for your deployment
# Command: openssl rand -base64 32
################################################################################

# JWT Secrets (Generate with: openssl rand -base64 32)
export JWT_SECRET_KEY="your-random-jwt-secret"              # 32-byte random secret
export REGISTRATION_SECRET="your-random-registration-secret" # 32-byte random secret
export BACKUP_KEY="your-random-backup-key"                  # 32-byte random secret

# TLS/SSL Configuration
# - Enable/disable TLS based on your needs
# - Provide paths to your certificates
export TLS_ENABLED="true"                        # true or false
export TLS_CERT_PATH="/etc/pganalytics/cert.pem" # Path to your certificate
export TLS_KEY_PATH="/etc/pganalytics/key.pem"   # Path to your private key
export TLS_CA_PATH="/etc/pganalytics/ca.pem"     # Path to CA cert (if using CA-signed)

# Mutual TLS for Collectors
export MTLS_ENABLED="true"                       # true or false
export MTLS_CLIENT_CERT="/etc/pganalytics/client.crt" # Client certificate path
export MTLS_CLIENT_KEY="/etc/pganalytics/client.key"  # Client key path

################################################################################
# CLOUD PROVIDER (Optional)
# Fill only if using cloud services (AWS, GCP, Azure, etc.)
# Leave blank if using on-premises only
################################################################################

# AWS (if using AWS services)
export AWS_REGION="your-aws-region"              # Region (us-east-1, eu-west-1, etc.)
export AWS_ACCOUNT_ID="your-aws-account-id"      # Your account ID
export AWS_RDS_INSTANCE_ID="your-rds-instance"   # RDS instance name (if using RDS)
export AWS_RDS_ENGINE="postgres"                 # Engine type
export AWS_RDS_VERSION="16.12"                   # Engine version
export AWS_SECRETS_MANAGER_ENABLED="true"        # Use AWS Secrets Manager (true/false)
export AWS_PROFILE="default"                     # AWS CLI profile

# Self-Hosted (if NOT using AWS)
export SELF_HOSTED="false"                       # Set to true if on-premises only

################################################################################
# DEPLOYMENT PARAMETERS
# Define how YOUR deployment should behave
################################################################################

# Environment
export ENVIRONMENT="production"                  # production, staging, or development
export DEPLOYMENT_NAME="pganalytics-v3.2.0"      # Your deployment name

# Deployment Mode
export DEPLOYMENT_MODE="distributed"             # Options: distributed, single-machine, kubernetes
export CONTAINER_RUNTIME="none"                  # Options: none, docker, kubernetes

# Logging
export LOG_LEVEL="info"                          # debug, info, warn, error
export LOG_PATH="/var/log/pganalytics"           # Where to write logs
export LOG_RETENTION_DAYS="30"                   # How long to keep logs

# Backup Configuration
export BACKUP_ENABLED="true"                     # true or false
export BACKUP_RETENTION_DAYS="30"                # How long to keep backups
export BACKUP_SCHEDULE="0 3 * * *"               # Cron schedule (3 AM daily)

################################################################################
# PERFORMANCE PARAMETERS (Optional)
# Tune for your workload - defaults are reasonable
################################################################################

export API_MAX_CONNECTIONS="100"                 # Max concurrent API connections
export DB_CONNECTION_POOL_SIZE="20"              # Database connection pool size
export METRICS_BATCH_SIZE="100"                  # Metrics per batch
export METRICS_FLUSH_INTERVAL="60"               # Flush interval (seconds)
export COLLECTOR_TIMEOUT="30"                    # Collector timeout (seconds)
export RATE_LIMIT_REQUESTS="1000"                # Rate limit (requests/minute)
export RATE_LIMIT_WINDOW="60"                    # Rate limit window (seconds)

################################################################################
# NOTIFICATION PARAMETERS (Optional)
# Configure alerting - can be empty if not needed
################################################################################

# Slack
export SLACK_WEBHOOK_URL=""                      # Your Slack webhook (leave empty to disable)
export SLACK_CHANNEL="#alerts"                   # Slack channel

# Email
export EMAIL_ENABLED="false"                     # true or false
export EMAIL_FROM="pganalytics@example.com"      # From email
export EMAIL_TO="ops@example.com"                # To email
export SMTP_HOST="smtp.example.com"              # SMTP server
export SMTP_PORT="587"                           # SMTP port
export SMTP_USER="username"                      # SMTP username
export SMTP_PASSWORD="password"                  # SMTP password

# PagerDuty
export PAGERDUTY_ENABLED="false"                 # true or false
export PAGERDUTY_KEY=""                          # PagerDuty key (leave empty to disable)

################################################################################
# NETWORK PARAMETERS (Optional)
# Define network access - open by default
################################################################################

# Allowed IP Ranges (CIDR notation)
export ALLOWED_API_CLIENTS="0.0.0.0/0"           # IPs allowed to API
export ALLOWED_COLLECTOR_CLIENTS="0.0.0.0/0"     # IPs allowed from collectors
export ALLOWED_GRAFANA_CLIENTS="0.0.0.0/0"       # IPs allowed to Grafana

# Timeouts
export NETWORK_TIMEOUT="30"                      # Network timeout (seconds)
export CONNECTION_TIMEOUT="10"                   # Connection timeout (seconds)

################################################################################
# POSTGRESQL SPECIFIC (Optional)
# PostgreSQL tuning - defaults are good
################################################################################

export PG_STAT_STATEMENTS_ENABLED="true"         # Enable query stats
export PG_LOG_STATEMENT="all"                    # Log all statements
export PG_LOG_DURATION="on"                      # Log duration
export PG_LOG_MIN_DURATION="1000"                # Min duration to log (ms)
export PG_SHARED_BUFFERS="256MB"                 # Shared buffer pool
export PG_EFFECTIVE_CACHE_SIZE="1GB"             # Cache size
export PG_WORK_MEM="4MB"                         # Work memory

################################################################################
# KUBERNETES SPECIFIC (Optional)
# Only fill if using Kubernetes
################################################################################

export K8S_NAMESPACE="pganalytics"               # Your K8s namespace
export K8S_CONTEXT="default"                     # Your K8s context
export K8S_STORAGE_CLASS="standard"              # Storage class for PVCs
export K8S_STORAGE_SIZE="100Gi"                  # Storage size

################################################################################
# DOCKER SPECIFIC (Optional)
# Only fill if using Docker
################################################################################

export DOCKER_REGISTRY="docker.io"               # Docker registry
export DOCKER_IMAGE_TAG="v3.2.0"                 # Docker image tag
export DOCKER_NETWORK="pganalytics-network"      # Docker network

################################################################################
# INTERNAL VARIABLES (Set automatically - do not change)
################################################################################

export DEPLOYMENT_ID="pganalytics-$(date +%s)"   # Unique deployment ID
export DEPLOYMENT_DATE="$(date +%Y-%m-%d)"       # Deployment date
export DEPLOYMENT_TIME="$(date +%H:%M:%S)"       # Deployment time
```

---

## How to Fill This Template

### For ANY Infrastructure

**Step 1: Copy Template**
```bash
cp DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md ~/.env.pganalytics
```

**Step 2: Edit and Fill YOUR Values**
```bash
nano ~/.env.pganalytics
```

**Step 3: Replace Placeholders**

```bash
# Example: AWS EC2
export DB_HOST="my-rds.abc123.us-east-1.rds.amazonaws.com"
export API_HOST_1="10.0.1.50"
export API_HOST_2="10.0.1.51"
export COLLECTOR_HOSTS=("10.0.2.10" "10.0.2.11" "10.0.2.12")
```

OR

```bash
# Example: On-premises
export DB_HOST="192.168.1.50"
export API_HOST_1="192.168.1.51"
export API_HOST_2="192.168.1.52"
export COLLECTOR_HOSTS=("192.168.1.60" "192.168.1.61")
```

OR

```bash
# Example: Kubernetes
export DB_HOST="postgres-service"
export API_HOST_1="api-service"
export COLLECTOR_HOSTS=("collector-1" "collector-2")
export DEPLOYMENT_MODE="kubernetes"
```

OR

```bash
# Example: Docker
export DB_HOST="postgres-container"
export API_HOST_1="localhost"
export COLLECTOR_HOSTS=("collector-1")
export CONTAINER_RUNTIME="docker"
```

OR

```bash
# Example: Hybrid/Mixed
export DB_HOST="my-postgres-vm.company.com"
export API_HOST_1="api-server-1.aws.company.com"
export API_HOST_2="api-server-2.onprem.company.com"
export COLLECTOR_HOSTS=("col-aws-1" "col-onprem-1" "col-k8s-1")
```

**Step 4: Source Configuration**
```bash
source ~/.env.pganalytics
```

**Step 5: Verify**
```bash
echo "Database: $DB_HOST:$DB_PORT"
echo "API: $API_HOST_1, $API_HOST_2"
echo "Collectors: ${COLLECTOR_HOSTS[@]}"
echo "Mode: $DEPLOYMENT_MODE"
```

**Step 6: Deploy**
```bash
bash scripts/phase1_automated_setup.sh
```

---

## What's Required vs Optional

| Parameter | Required | Purpose |
|-----------|----------|---------|
| **DB_HOST, DB_PORT, DB_USER, DB_PASSWORD** | ✅ YES | Connect to your database |
| **API_HOST_1, API_HOST_2** | ✅ YES | Where to deploy API |
| **COLLECTOR_HOSTS** | ✅ YES | Where collectors run |
| **GRAFANA_HOST, PROMETHEUS_HOST** | ✅ YES | Monitoring setup |
| **JWT_SECRET_KEY, REGISTRATION_SECRET, BACKUP_KEY** | ✅ YES | Security |
| **ENVIRONMENT, DEPLOYMENT_MODE** | ✅ YES | Deployment config |
| Everything else | 🔧 OPTIONAL | Tuning, alerting, advanced config |

---

## Key Points

✅ **NO assumptions** about your infrastructure
✅ **Works with ANYTHING** - AWS, on-premises, Kubernetes, Docker, hybrid, etc.
✅ **Scale doesn't matter** - 1 collector, 1000 collectors, 10000 collectors
✅ **Cloud provider doesn't matter** - AWS, GCP, Azure, or none
✅ **YOU define everything** - We just use your values

---

## Examples: Fill for Different Scenarios

### Scenario 1: Single Server (Development)
```bash
export DB_HOST="localhost"
export API_HOST_1="localhost"
export API_HOST_2="localhost"
export COLLECTOR_HOSTS=("localhost")
export GRAFANA_HOST="localhost"
export PROMETHEUS_HOST="localhost"
export ENVIRONMENT="development"
export DEPLOYMENT_MODE="single-machine"
export CONTAINER_RUNTIME="docker"
```

### Scenario 2: AWS EC2 + RDS
```bash
export DB_HOST="pganalytics.c9akciq32.us-east-1.rds.amazonaws.com"
export API_HOST_1="10.0.1.100"
export API_HOST_2="10.0.1.101"
export COLLECTOR_HOSTS=("10.0.2.1" "10.0.2.2" "10.0.2.3")
export GRAFANA_HOST="10.0.3.50"
export PROMETHEUS_HOST="10.0.3.51"
export AWS_REGION="us-east-1"
export ENVIRONMENT="production"
export DEPLOYMENT_MODE="distributed"
```

### Scenario 3: On-Premises Physical Machines
```bash
export DB_HOST="postgres-main.company.internal"
export API_HOST_1="api-01.company.internal"
export API_HOST_2="api-02.company.internal"
export COLLECTOR_HOSTS=("col-01.company.internal" "col-02.company.internal" "col-03.company.internal")
export GRAFANA_HOST="monitoring.company.internal"
export PROMETHEUS_HOST="monitoring.company.internal"
export SELF_HOSTED="true"
export ENVIRONMENT="production"
export DEPLOYMENT_MODE="distributed"
```

### Scenario 4: Kubernetes
```bash
export DB_HOST="postgres-service"
export API_HOST_1="api-service"
export API_HOST_2="api-service"
export COLLECTOR_HOSTS=("collector-pod-1" "collector-pod-2" "collector-pod-3")
export GRAFANA_HOST="grafana-service"
export PROMETHEUS_HOST="prometheus-service"
export ENVIRONMENT="production"
export DEPLOYMENT_MODE="kubernetes"
export CONTAINER_RUNTIME="kubernetes"
export K8S_NAMESPACE="pganalytics"
```

### Scenario 5: Hybrid (Multiple Infrastructures)
```bash
export DB_HOST="aws-rds.us-east-1.rds.amazonaws.com"      # AWS RDS
export API_HOST_1="api-aws.company.internal"              # On-prem server
export API_HOST_2="api-onprem.company.internal"           # On-prem server
export COLLECTOR_HOSTS=("col-aws-1" "col-onprem-1" "col-k8s-1")  # Mixed
export GRAFANA_HOST="grafana-central.company.internal"
export PROMETHEUS_HOST="prometheus-central.company.internal"
export ENVIRONMENT="production"
export DEPLOYMENT_MODE="distributed"
```

---

## For Multiple Regions/Deployments

If you have multiple regions or multiple deployments, create separate files:

```bash
# US-East-1
~/.env.pganalytics.us-east-1
source ~/.env.pganalytics.us-east-1
bash scripts/phase1_automated_setup.sh

# US-West-1
~/.env.pganalytics.us-west-1
source ~/.env.pganalytics.us-west-1
bash scripts/phase1_automated_setup.sh

# Europe
~/.env.pganalytics.eu
source ~/.env.pganalytics.eu
bash scripts/phase1_automated_setup.sh
```

---

## Usage

```bash
# 1. Copy template
cp DEPLOYMENT_CONFIG_TEMPLATE_OPEN.md ~/.env.pganalytics

# 2. Edit with YOUR infrastructure
nano ~/.env.pganalytics

# 3. Source it
source ~/.env.pganalytics

# 4. Deploy
bash scripts/phase1_automated_setup.sh
```

**That's it!** No configuration guides, no infrastructure questionnaires, no prescriptive requirements.

YOU define YOUR infrastructure. We just use your values.

---

## Important Notes

- 🔒 Never commit `~/.env.pganalytics` to version control
- 🔐 Keep it secure - contains credentials
- 🔑 Generate secrets: `openssl rand -base64 32`
- 📋 For production - use secrets management (AWS Secrets Manager, Vault, etc.)
- 🧪 Test in staging first

---

**This template works with:**
- ✅ AWS (EC2, RDS, any region)
- ✅ On-premises (physical machines)
- ✅ Kubernetes (any cluster)
- ✅ Docker (local or remote)
- ✅ Hybrid (mix of everything)
- ✅ Any scale (1 to 10000+ collectors)
- ✅ Any cloud provider (or no cloud)
- ✅ Any organization (enterprise to startup)

**No questions. No assumptions. Just YOUR infrastructure.**
