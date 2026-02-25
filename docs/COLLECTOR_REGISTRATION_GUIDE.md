# Collector Registration & Authentication Guide

**Version:** 1.0
**Date:** February 25, 2026
**Status:** Production Ready
**Audience:** DevOps Engineers, Database Administrators, SREs

---

## Table of Contents

1. [Overview](#overview)
2. [Registration Flow](#registration-flow)
3. [Collector Configuration](#collector-configuration)
4. [Multi-Collector Scenarios](#multi-collector-scenarios)
5. [Security Best Practices](#security-best-practices)
6. [Token Management](#token-management)
7. [Troubleshooting](#troubleshooting)

---

## Overview

### What is Collector Registration?

Collector registration is the process by which distributed collectors securely authenticate with the pgAnalytics backend API. During registration:

1. Collector sends registration request with pre-shared secret
2. Backend validates secret and generates JWT token
3. Backend creates X.509 certificate for mTLS
4. Collector receives token + certificate (valid 1 year)
5. Collector uses token for all subsequent metrics pushes

### Security Model

```
┌──────────────────────────────────────────────────────────┐
│                 Authentication Layer                      │
├──────────────────────────────────────────────────────────┤
│                                                            │
│  Pre-Registration (Bootstrap)                             │
│  ├── Environment: REGISTRATION_SECRET (pre-shared)        │
│  ├── Method: HTTP POST with secret header                 │
│  ├── Frequency: Once during deployment                    │
│  └── Trust: Out-of-band secret distribution               │
│                                                            │
│  Post-Registration (Operational)                          │
│  ├── Environment: JWT_TOKEN (issued by backend)           │
│  ├── Method: HTTP Bearer token (JWT)                      │
│  ├── Frequency: Every metrics push                        │
│  ├── Validity: 1 year (or custom duration)                │
│  └── Trust: Cryptographic signature (HS256)               │
│                                                            │
│  Transport Security (Always)                              │
│  ├── Protocol: TLS 1.3 minimum                            │
│  ├── Certificates: mTLS + client cert verification        │
│  ├── Root CA: Self-signed or CA-issued                    │
│  └── Cipher Suites: HIGH:!aNULL:!MD5                      │
│                                                            │
└──────────────────────────────────────────────────────────┘
```

---

## Registration Flow

### Prerequisites

Before registration, ensure:

1. **Backend API is running**
   ```bash
   curl https://api.example.com/api/v1/health
   # Response: {"status":"healthy","version":"3.2.0"}
   ```

2. **REGISTRATION_SECRET is known**
   - Securely distributed by infrastructure team
   - 32+ characters, cryptographically random
   - Never commit to version control

3. **Collector binary is available**
   - Built or downloaded (see ENTERPRISE_INSTALLATION.md)
   - Ready to configure and deploy

4. **Network access to backend**
   - Port 443 (HTTPS) accessible
   - DNS resolution working
   - TLS certificate trusted

### Step-by-Step Registration

#### Step 1: Prepare Registration Request

```bash
# Define variables
BACKEND_URL="https://api.example.com"
REGISTRATION_SECRET="your-32-character-secret"
COLLECTOR_NAME="collector-001"
COLLECTOR_HOSTNAME="db-prod-01.example.com"
COLLECTOR_IP="203.0.113.10"

# Create request payload
cat > registration_request.json << 'EOF'
{
  "name": "COLLECTOR_NAME",
  "hostname": "COLLECTOR_HOSTNAME",
  "address": "COLLECTOR_IP",
  "region": "us-east-1",
  "tags": {
    "environment": "production",
    "tier": "primary"
  }
}
EOF

# Replace placeholders
sed -i "s/COLLECTOR_NAME/$COLLECTOR_NAME/g" registration_request.json
sed -i "s/COLLECTOR_HOSTNAME/$COLLECTOR_HOSTNAME/g" registration_request.json
sed -i "s/COLLECTOR_IP/$COLLECTOR_IP/g" registration_request.json
```

#### Step 2: Submit Registration Request

```bash
# Register collector
curl -X POST $BACKEND_URL/api/v1/collectors/register \
  -H "X-Registration-Secret: $REGISTRATION_SECRET" \
  -H "Content-Type: application/json" \
  -d @registration_request.json \
  -k > collector_credentials.json  # -k ignores self-signed certs (dev only)

# Verify response
cat collector_credentials.json | jq .
```

#### Step 3: Parse Registration Response

```bash
# Extract credentials
COLLECTOR_ID=$(jq -r '.collector_id' collector_credentials.json)
JWT_TOKEN=$(jq -r '.token' collector_credentials.json)
CERT_PEM=$(jq -r '.certificate' collector_credentials.json)
KEY_PEM=$(jq -r '.private_key' collector_credentials.json)
TOKEN_EXPIRES=$(jq -r '.expires_at' collector_credentials.json)

echo "Collector ID: $COLLECTOR_ID"
echo "Token expires: $TOKEN_EXPIRES"

# Save full response for reference
cp collector_credentials.json collector-$COLLECTOR_NAME-credentials.json
```

#### Step 4: Save Credentials Securely

```bash
# Create credentials directory
sudo mkdir -p /etc/pganalytics/credentials
sudo chmod 700 /etc/pganalytics/credentials

# Save JWT token
echo "$JWT_TOKEN" | sudo tee /etc/pganalytics/credentials/jwt_token.txt
sudo chmod 600 /etc/pganalytics/credentials/jwt_token.txt

# Save certificate
echo "$CERT_PEM" | sudo tee /etc/pganalytics/collector.crt
sudo chmod 644 /etc/pganalytics/collector.crt

# Save private key
echo "$KEY_PEM" | sudo tee /etc/pganalytics/collector.key
sudo chmod 600 /etc/pganalytics/collector.key

# Store collector ID
echo "$COLLECTOR_ID" | sudo tee /etc/pganalytics/collector_id.txt
sudo chmod 644 /etc/pganalytics/collector_id.txt
```

#### Step 5: Verify Credentials

```bash
# Check certificate validity
openssl x509 -in /etc/pganalytics/collector.crt -text -noout | grep -A2 "Validity"

# Check private key format
openssl pkey -in /etc/pganalytics/collector.key -text -noout | head -5

# Verify JWT token structure
echo "$JWT_TOKEN" | cut -d'.' -f1 | base64 -d | jq .
echo "$JWT_TOKEN" | cut -d'.' -f2 | base64 -d | jq .

# Expected JWT payload:
# {
#   "sub": "COLLECTOR_ID",
#   "collector_id": "COLLECTOR_ID",
#   "iat": 1709000000,
#   "exp": 1740536000,  (1 year later)
#   "iss": "pganalytics-api"
# }
```

#### Step 6: Store Credentials in Configuration

```bash
# Edit collector configuration
sudo nano /etc/pganalytics/collector.toml

# Minimal example:
[collector]
id = "COLLECTOR_ID"
hostname = "db-prod-01.example.com"
interval = 60

[backend]
url = "https://api.example.com"
jwt_token = "JWT_TOKEN_HERE"
tls_cert = "/etc/pganalytics/collector.crt"
tls_key = "/etc/pganalytics/collector.key"
ca_cert = "/etc/pganalytics/ca.crt"
verify_ssl = true

[postgres]
host = "localhost"
port = 5432
user = "pganalytics_monitoring"
password = "secure-password"
databases = ["postgres"]

[metrics]
enable_pg_stat_statements = true
enable_replication = true
enable_wal_level = true

[security]
tls_enabled = true
verify_certificate = true
```

#### Step 7: Validate Configuration

```bash
# Validate collector configuration
pganalytics --config /etc/pganalytics/collector.toml --validate

# Test connection without pushing metrics
pganalytics --config /etc/pganalytics/collector.toml --dry-run

# Check for errors
pganalytics --config /etc/pganalytics/collector.toml --verbose --dry-run 2>&1 | head -50
```

#### Step 8: Start Collector Service

```bash
# Start systemd service
sudo systemctl start pganalytics-collector

# Verify it's running
sudo systemctl status pganalytics-collector

# Watch logs in real-time
sudo journalctl -u pganalytics-collector -f

# Expected output:
# INFO: Collector started with ID: collector-001
# INFO: PostgreSQL connection established
# INFO: Successfully registered with backend
# INFO: Pushing metrics... (1000 metrics in 250ms)
```

#### Step 9: Verify Collector Registration

```bash
# Check if collector is registered in backend
curl -H "Authorization: Bearer $USER_TOKEN" \
  https://api.example.com/api/v1/collectors | jq '.[] | select(.name == "collector-001")'

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "name": "collector-001",
#   "hostname": "db-prod-01.example.com",
#   "status": "active",
#   "last_seen": "2026-02-25T10:15:30Z",
#   "metrics_count": 1450,
#   "created_at": "2026-02-25T10:10:00Z"
# }

# Check metrics in database
psql -U pganalytics -d pganalytics

SELECT COUNT(*) FROM metrics WHERE collector_id = 'collector-001';
# Should show > 0 after first collection cycle (~60 seconds)
```

### Registration API Reference

#### Endpoint: POST /api/v1/collectors/register

**Request:**
```bash
curl -X POST https://api.example.com/api/v1/collectors/register \
  -H "X-Registration-Secret: $REGISTRATION_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "collector-001",
    "hostname": "db.example.com",
    "address": "192.168.1.100",
    "region": "us-east-1",
    "tags": {
      "environment": "production",
      "tier": "primary",
      "cluster": "postgres-cluster-1"
    }
  }'
```

**Response (201 Created):**
```json
{
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "certificate": "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
  "ca_certificate": "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
  "expires_at": "2027-02-25T00:00:00Z",
  "fingerprint": "SHA256:abcd1234..."
}
```

**Error Responses:**
```json
// 401 Unauthorized - Missing or invalid REGISTRATION_SECRET
{
  "error": "invalid_registration_secret",
  "message": "Registration secret not provided or invalid"
}

// 400 Bad Request - Invalid payload
{
  "error": "invalid_request",
  "message": "Missing required field: name"
}

// 409 Conflict - Collector already registered
{
  "error": "collector_exists",
  "message": "Collector with name 'collector-001' already registered"
}
```

---

## Collector Configuration

### Configuration File (collector.toml)

```toml
# pgAnalytics Collector Configuration
# Version: 3.2.0
# Generated: 2026-02-25

[collector]
# Unique identifier (auto-generated by backend during registration)
id = "550e8400-e29b-41d4-a716-446655440000"

# Human-readable name
name = "collector-001"

# Hostname or IP of PostgreSQL instance being monitored
hostname = "db-prod-01.example.com"

# Collection interval in seconds
interval = 60

# Enable verbose logging
verbose = false

# Log level: trace, debug, info, warn, error
log_level = "info"

# Log file path (empty = stdout)
log_file = "/var/log/pganalytics/collector.log"

# Max log file size in MB
log_max_size = 100

# Keep log backups
log_backup_count = 5

[backend]
# Backend API URL
url = "https://api.example.com"

# Path to client certificate (for mTLS)
tls_cert = "/etc/pganalytics/collector.crt"

# Path to client private key (for mTLS)
tls_key = "/etc/pganalytics/collector.key"

# Path to CA certificate for verification
ca_cert = "/etc/pganalytics/ca.crt"

# Verify server certificate
verify_ssl = true

# JWT token from registration (or refresh endpoint)
jwt_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Token refresh endpoint (optional)
token_refresh_url = "https://api.example.com/api/v1/collectors/token/refresh"

# Refresh token before expiration (days)
token_refresh_days = 30

# Metrics push timeout in seconds
push_timeout = 30

# Retry failed pushes
max_retries = 3

# Retry backoff: exponential or linear
retry_backoff = "exponential"

[postgres]
# PostgreSQL instance details
host = "localhost"
port = 5432
user = "pganalytics_monitoring"
password = "secure-password"

# Connection timeout
connect_timeout = 10

# Query timeout in seconds
query_timeout = 30

# Connection pool settings
pool_min_connections = 2
pool_max_connections = 10

# SSL mode: disable, allow, prefer, require
ssl_mode = "require"

# SSL certificate verification
ssl_verify_server = true

# PostgreSQL version (auto-detected if empty)
# Supports: 9.4, 9.5, 9.6, 10, 11, 12, 13, 14, 15, 16, 17, 18
version = ""

# Databases to monitor (empty = all)
databases = ["postgres", "app_db"]

# Exclude system databases
exclude_system_databases = false

[metrics]
# PostgreSQL statistics to collect
enable_pg_stat_statements = true
enable_pg_stat_database = true
enable_pg_stat_table = true
enable_pg_stat_index = true
enable_pg_stat_user_functions = true
enable_pg_stat_io = true

# Replication metrics
enable_replication = true
enable_wal_level = true

# System metrics
enable_system_metrics = true
enable_disk_metrics = true
enable_memory_metrics = true

# Log file analysis
enable_log_analysis = false
log_path = "/var/log/postgresql/postgresql.log"

# Metrics to exclude (regex patterns)
exclude_metrics = []

# Sample rate for high-cardinality metrics (0-100)
sample_rate = 100

[buffer]
# In-memory buffer for metrics before sending
buffer_enabled = true

# Max metrics to buffer
buffer_max_size = 10000

# Flush buffered metrics interval (seconds)
buffer_flush_interval = 60

# Compress buffered metrics
compress_buffer = true

[security]
# Enable TLS for metrics push
tls_enabled = true

# Verify API certificate
verify_certificate = true

# API timeout
api_timeout = 30

# Rate limiting: max requests per minute
rate_limit_rps = 100

[health_check]
# Enable health check endpoint (for monitoring)
enabled = true

# Health check port
port = 9090

# Health check path
path = "/health"

# Check intervals (seconds)
interval = 30

# Write health status to file
status_file = "/var/run/pganalytics/collector.status"

[telemetry]
# Enable Prometheus metrics export
prometheus_enabled = false
prometheus_port = 9091

# Send telemetry to backend
telemetry_enabled = true
```

### Configuration Examples

#### Minimal Configuration (Quick Start)

```toml
[collector]
id = "COLLECTOR_ID_FROM_REGISTRATION"
hostname = "localhost"
interval = 60

[backend]
url = "https://api.example.com"
jwt_token = "JWT_TOKEN_FROM_REGISTRATION"

[postgres]
host = "localhost"
port = 5432
user = "pganalytics_monitoring"
password = "secure-password"
```

#### High-Security Configuration (Production)

```toml
[collector]
id = "550e8400-e29b-41d4-a716-446655440000"
verbose = false
log_level = "warn"
log_file = "/var/log/pganalytics/collector.log"
log_max_size = 100

[backend]
url = "https://api.example.com"
tls_cert = "/etc/pganalytics/collector.crt"
tls_key = "/etc/pganalytics/collector.key"
ca_cert = "/etc/pganalytics/ca.crt"
verify_ssl = true
jwt_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
token_refresh_url = "https://api.example.com/api/v1/collectors/token/refresh"
token_refresh_days = 30
max_retries = 5

[postgres]
host = "db-prod-01.example.com"
port = 5432
user = "pganalytics_monitoring"
password = "$(cat /run/secrets/pg_password)"
connect_timeout = 10
query_timeout = 30
ssl_mode = "require"
ssl_verify_server = true
version = "16"
databases = ["production_db"]

[metrics]
enable_pg_stat_statements = true
enable_replication = true
enable_system_metrics = true
sample_rate = 100

[buffer]
buffer_enabled = true
buffer_max_size = 10000
buffer_flush_interval = 60
compress_buffer = true

[security]
tls_enabled = true
verify_certificate = true
api_timeout = 30

[health_check]
enabled = true
status_file = "/var/run/pganalytics/collector.status"
```

#### Multi-Database Configuration

```toml
[collector]
id = "collector-multi-001"
hostname = "multi-db-server.example.com"
interval = 60

[backend]
url = "https://api.example.com"
jwt_token = "..."

[postgres]
host = "localhost"
port = 5432
user = "pganalytics_monitoring"
password = "secure-password"

# Monitor multiple databases
databases = [
  "postgres",
  "application_db",
  "reporting_db",
  "analytics_db"
]

# Or monitor all databases except these
exclude_databases = [
  "template0",
  "template1",
  "test_db"
]

[metrics]
enable_pg_stat_statements = true
enable_pg_stat_database = true
enable_pg_stat_table = true
enable_replication = true
enable_system_metrics = true
```

#### Environment Variable Configuration

```bash
# Load secrets from environment (for Docker/K8s)
export PG_PASSWORD=$(cat /run/secrets/pg_password)
export JWT_TOKEN=$(cat /run/secrets/jwt_token)
export BACKEND_URL=$(cat /run/secrets/backend_url)

# Generate configuration from template
envsubst < collector.toml.template > collector.toml

# Start collector
pganalytics --config collector.toml
```

---

## Multi-Collector Scenarios

### Scenario 1: 5 Collectors Monitoring Different PostgreSQL Instances

```bash
#!/bin/bash
# register-collectors.sh - Register 5 collectors

BACKEND_URL="https://api.example.com"
REGISTRATION_SECRET="your-secret"

# Define collectors
COLLECTORS=(
  "primary-db:db-prod-01.example.com:203.0.113.10"
  "replica-1:db-prod-02.example.com:203.0.113.11"
  "replica-2:db-prod-03.example.com:203.0.113.12"
  "analytics-db:db-analytics.example.com:203.0.113.13"
  "backup-db:db-backup.example.com:203.0.113.14"
)

# Register each collector
for collector in "${COLLECTORS[@]}"; do
  IFS=':' read -r name hostname ip <<< "$collector"

  echo "Registering $name..."

  response=$(curl -s -X POST $BACKEND_URL/api/v1/collectors/register \
    -H "X-Registration-Secret: $REGISTRATION_SECRET" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "'$name'",
      "hostname": "'$hostname'",
      "address": "'$ip'"
    }' -k)

  # Extract credentials
  collector_id=$(echo $response | jq -r '.collector_id')
  token=$(echo $response | jq -r '.token')
  cert=$(echo $response | jq -r '.certificate')
  key=$(echo $response | jq -r '.private_key')

  # Save credentials
  mkdir -p credentials/$name
  echo "$collector_id" > credentials/$name/id
  echo "$token" > credentials/$name/token
  echo "$cert" > credentials/$name/cert.pem
  echo "$key" > credentials/$name/key.pem

  echo "✓ Registered $name with ID: $collector_id"
done
```

### Scenario 2: Regional Collectors (US, EU, APAC)

```bash
#!/bin/bash
# register-regional-collectors.sh

REGIONS=(
  "us-east:US-EAST-1:203.0.113.20:203.0.113.21"
  "us-west:US-WEST-1:203.0.113.22:203.0.113.23"
  "eu-west:EU-WEST-1:203.0.113.24:203.0.113.25"
  "apac:APAC-1:203.0.113.26:203.0.113.27"
)

for region in "${REGIONS[@]}"; do
  IFS=':' read -r region_code region_name ip1 ip2 <<< "$region"

  # Register 2 collectors per region
  for i in 1 2; do
    collector_name="$region_code-$i"
    hostname="db-$region_code-$i.example.com"
    ip_addr=$([ $i -eq 1 ] && echo $ip1 || echo $ip2)

    echo "Registering $collector_name ($region_name)..."

    curl -s -X POST https://api.example.com/api/v1/collectors/register \
      -H "X-Registration-Secret: $SECRET" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "'$collector_name'",
        "hostname": "'$hostname'",
        "address": "'$ip_addr'",
        "region": "'$region_name'"
      }' -k | jq . > credentials/$collector_name.json

    echo "✓ Registered $collector_name"
  done
done
```

### Scenario 3: Failover Collector Setup

```toml
# Primary collector configuration
[collector]
id = "primary-collector-001"
hostname = "db-prod-primary.example.com"

[backend]
url = "https://api.example.com"
jwt_token = "primary-token"

# Secondary/standby collector (same PostgreSQL instance)
[collector]
id = "standby-collector-001"
hostname = "db-prod-standby.example.com"

[backend]
url = "https://api.example.com"
jwt_token = "standby-token"

# Both collectors query the same PostgreSQL instance
[postgres]
host = "db-prod-primary.example.com"
port = 5432
user = "pganalytics_monitoring"
password = "secure-password"

# Failover logic: if primary unreachable, try secondary
[failover]
primary_host = "db-prod-primary.example.com"
fallback_host = "db-prod-standby.example.com"
health_check_interval = 30
```

### Scenario 4: Load-Balanced Collector Deployment

```bash
#!/bin/bash
# deploy-collectors-elb.sh - Deploy 10 collectors behind AWS ELB

for i in {1..10}; do
  # Launch EC2 instance
  instance_id=$(aws ec2 run-instances \
    --image-id ami-0c55b159cbfafe1f0 \
    --instance-type t3.medium \
    --user-data file://collector-userdata.sh \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=pganalytics-collector-'$i'},{Key=Fleet,Value=pganalytics}]' \
    --query 'Instances[0].InstanceId' \
    --output text)

  echo "Launched instance: $instance_id"

  # Wait for instance to be ready
  aws ec2 wait instance-running --instance-ids $instance_id

  # Register collector
  ip=$(aws ec2 describe-instances \
    --instance-ids $instance_id \
    --query 'Reservations[0].Instances[0].PrivateIpAddress' \
    --output text)

  curl -X POST https://api.example.com/api/v1/collectors/register \
    -H "X-Registration-Secret: $SECRET" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "collector-'$i'",
      "hostname": "db-prod-'$i'.example.com",
      "address": "'$ip'"
    }' -k

  echo "✓ Registered collector-$i"

  # Add to load balancer
  aws elbv2 register-targets \
    --target-group-arn arn:aws:elasticloadbalancing:... \
    --targets Id=$instance_id

done
```

---

## Security Best Practices

### 1. REGISTRATION_SECRET Management

```bash
# Generate strong registration secret
REGISTRATION_SECRET=$(openssl rand -base64 32)
echo "REGISTRATION_SECRET=$REGISTRATION_SECRET"

# Store in secure location
# Option A: AWS Secrets Manager
aws secretsmanager create-secret \
  --name pganalytics/registration-secret \
  --secret-string "$REGISTRATION_SECRET"

# Option B: HashiCorp Vault
vault kv put secret/pganalytics/registration \
  secret="$REGISTRATION_SECRET"

# Option C: Kubernetes Secrets
kubectl create secret generic pganalytics-registration \
  --from-literal=secret="$REGISTRATION_SECRET" \
  -n pganalytics

# Retrieve when needed (never hardcode)
REGISTRATION_SECRET=$(aws secretsmanager get-secret-value \
  --secret-id pganalytics/registration-secret \
  --query SecretString --output text)
```

### 2. JWT Token Security

```bash
# Tokens are valid for 1 year
# Implement rotation before expiration

# Monitor token expiration
curl -H "Authorization: Bearer $JWT_TOKEN" \
  https://api.example.com/api/v1/collectors/token/info | jq '.expires_at'

# Refresh token 30 days before expiration
curl -X POST https://api.example.com/api/v1/collectors/token/refresh \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"collector_id":"'$COLLECTOR_ID'"}' | jq .

# Update configuration with new token
new_token=$(curl ... | jq -r '.token')
sed -i "s/jwt_token = .*/jwt_token = \"$new_token\"/" /etc/pganalytics/collector.toml

# Restart collector
sudo systemctl restart pganalytics-collector
```

### 3. Certificate Management

```bash
# Certificates are valid for 1 year
# Monitor expiration

# Check certificate validity
openssl x509 -in /etc/pganalytics/collector.crt -text -noout | \
  grep -A2 "Validity"

# Alert if expiration < 30 days
cert_expiry=$(openssl x509 -in /etc/pganalytics/collector.crt -noout -dates | grep notAfter | cut -d= -f2)
cert_expiry_epoch=$(date -d "$cert_expiry" +%s)
current_epoch=$(date +%s)
days_until_expiry=$(( ($cert_expiry_epoch - $current_epoch) / 86400 ))

if [ $days_until_expiry -lt 30 ]; then
  echo "WARNING: Certificate expires in $days_until_expiry days"
  # Trigger rotation (re-register collector)
fi
```

### 4. Credential Rotation

```bash
# Implement quarterly credential rotation

# Cron job for quarterly rotation
0 2 1 */3 * /opt/pganalytics/rotate-credentials.sh

#!/bin/bash
# rotate-credentials.sh

BACKEND_URL="https://api.example.com"
REGISTRATION_SECRET=$(aws secretsmanager get-secret-value ...)

# Get current collector ID
COLLECTOR_ID=$(cat /etc/pganalytics/collector_id.txt)

# Register new credentials (backend marks old ones for rotation)
curl -X POST $BACKEND_URL/api/v1/collectors/rotate \
  -H "Authorization: Bearer $(cat /etc/pganalytics/credentials/jwt_token.txt)" \
  -H "X-Registration-Secret: $REGISTRATION_SECRET" \
  -H "Content-Type: application/json" \
  -d '{"collector_id":"'$COLLECTOR_ID'"}' | jq . > /tmp/new_credentials.json

# Update files
jq -r '.token' /tmp/new_credentials.json > /etc/pganalytics/credentials/jwt_token.txt
jq -r '.certificate' /tmp/new_credentials.json > /etc/pganalytics/collector.crt
jq -r '.private_key' /tmp/new_credentials.json > /etc/pganalytics/collector.key

# Fix permissions
chmod 600 /etc/pganalytics/credentials/jwt_token.txt
chmod 600 /etc/pganalytics/collector.key

# Restart collector
systemctl restart pganalytics-collector

# Verify
systemctl status pganalytics-collector
```

### 5. Network Security

```bash
# Firewall rules for collector
sudo iptables -A OUTPUT -p tcp --dport 443 -d 203.0.113.0/24 -j ACCEPT
sudo iptables -A OUTPUT -p tcp --dport 5432 -d 10.0.0.0/8 -j ACCEPT
sudo iptables -A OUTPUT -p udp --dport 53 -j ACCEPT  # DNS
sudo iptables -A OUTPUT -p udp --dport 123 -j ACCEPT # NTP
sudo iptables -A OUTPUT -j DROP  # Default deny

# AWS Security Group
aws ec2 authorize-security-group-egress \
  --group-id sg-xxxxx \
  --protocol tcp --port 443 --cidr 203.0.113.0/24

# Certificate pinning (optional)
# Collector verifies backend certificate fingerprint
[backend]
pin_certificate_fingerprint = "SHA256:abcd1234ef5678..."
```

### 6. Secret Management Integration

#### AWS Secrets Manager
```bash
# Store credentials in Secrets Manager
aws secretsmanager create-secret \
  --name pganalytics/collector-001 \
  --secret-string '{
    "jwt_token": "...",
    "collector_id": "...",
    "pg_password": "..."
  }'

# Load in systemd service
sudo nano /etc/systemd/system/pganalytics-collector.service

[Service]
ExecStartPre=/usr/local/bin/load-secrets.sh
Environment=JWT_TOKEN=${JWT_TOKEN_FROM_SECRETS}
```

#### HashiCorp Vault
```bash
# Store in Vault
vault kv put secret/pganalytics/collectors/collector-001 \
  jwt_token="..." \
  collector_id="..." \
  pg_password="..."

# Load at runtime
curl -s http://vault.example.com:8200/v1/secret/data/pganalytics/collectors/collector-001 \
  --header "X-Vault-Token: $VAULT_TOKEN" | jq '.data.data'
```

#### Kubernetes Secrets
```bash
# Create secret
kubectl create secret generic collector-001-credentials \
  --from-literal=jwt_token="..." \
  --from-literal=collector_id="..." \
  --from-literal=pg_password="..." \
  -n pganalytics

# Use in pod
spec:
  containers:
  - name: collector
    env:
    - name: JWT_TOKEN
      valueFrom:
        secretKeyRef:
          name: collector-001-credentials
          key: jwt_token
```

---

## Token Management

### JWT Token Structure

```json
Header:
{
  "alg": "HS256",
  "typ": "JWT"
}

Payload:
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "collector_id": "550e8400-e29b-41d4-a716-446655440000",
  "iat": 1709000000,
  "exp": 1740536000,
  "iss": "pganalytics-api",
  "aud": "pganalytics-collectors"
}

Signature:
HMACSHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload), secret)
```

### Token Validation

```bash
# Decode JWT token
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Header
echo $JWT_TOKEN | cut -d'.' -f1 | base64 -d | jq .

# Payload
echo $JWT_TOKEN | cut -d'.' -f2 | base64 -d | jq .

# Verify expiration
expiry=$(echo $JWT_TOKEN | cut -d'.' -f2 | base64 -d | jq '.exp')
current=$(date +%s)
if [ $expiry -lt $current ]; then
  echo "Token has expired"
else
  echo "Token valid for $((expiry - current)) seconds"
fi
```

### Token Refresh Process

```bash
# Automatic token refresh (cron job)
0 */6 * * * /usr/local/bin/refresh-jwt-token.sh

#!/bin/bash
# refresh-jwt-token.sh

COLLECTOR_ID=$(cat /etc/pganalytics/collector_id.txt)
OLD_TOKEN=$(cat /etc/pganalytics/credentials/jwt_token.txt)

# Refresh token
response=$(curl -s -X POST https://api.example.com/api/v1/collectors/token/refresh \
  -H "Authorization: Bearer $OLD_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"collector_id":"'$COLLECTOR_ID'"}')

# Extract new token
NEW_TOKEN=$(echo $response | jq -r '.token')

if [ "$NEW_TOKEN" != "null" ] && [ ! -z "$NEW_TOKEN" ]; then
  # Save new token
  echo "$NEW_TOKEN" > /etc/pganalytics/credentials/jwt_token.txt.new
  chmod 600 /etc/pganalytics/credentials/jwt_token.txt.new

  # Atomic replace
  mv /etc/pganalytics/credentials/jwt_token.txt.new /etc/pganalytics/credentials/jwt_token.txt

  # Update collector config
  sed -i 's/jwt_token = .*/jwt_token = "'"$NEW_TOKEN"'"/' /etc/pganalytics/collector.toml

  # Restart collector
  systemctl restart pganalytics-collector

  echo "Token refreshed successfully"
else
  echo "Token refresh failed: $response"
  exit 1
fi
```

---

## Troubleshooting

### Common Registration Issues

#### "Invalid REGISTRATION_SECRET"

```bash
# Verify secret is set
echo $REGISTRATION_SECRET | wc -c  # Should be 32+ characters

# Check if header is correct
curl -X POST https://api.example.com/api/v1/collectors/register \
  -H "X-Registration-Secret: $REGISTRATION_SECRET" \
  -v  # Check request headers

# Verify backend logs
sudo journalctl -u pganalytics-api -n 50 | grep "registration"
```

#### "Collector Already Exists"

```bash
# Check existing collectors
curl https://api.example.com/api/v1/collectors | jq '.[] | .name' | grep collector-001

# Use different name or delete existing
curl -X DELETE https://api.example.com/api/v1/collectors/collector-001 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

#### Certificate Validation Failures

```bash
# Verify certificate chain
openssl verify -CAfile /etc/pganalytics/ca.crt /etc/pganalytics/collector.crt

# Check certificate CN/SAN
openssl x509 -in /etc/pganalytics/collector.crt -text -noout | grep -A2 "Subject:"

# For self-signed certs in development, disable verification
[backend]
verify_ssl = false
```

### Token Issues

#### "Invalid JWT Token"

```bash
# Check token format
echo $JWT_TOKEN | wc -c  # Should be > 100 characters
echo $JWT_TOKEN | grep -E '^\w+\.\w+\.\w+$'  # Should have 3 parts

# Check expiration
echo $JWT_TOKEN | cut -d'.' -f2 | base64 -d | jq '.exp'
date +%s  # Compare with current time

# If expired, refresh or re-register
```

#### "Token Expired"

```bash
# Refresh token immediately
curl -X POST https://api.example.com/api/v1/collectors/token/refresh \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"collector_id":"'$COLLECTOR_ID'"}'

# Or re-register if refresh fails
curl -X POST https://api.example.com/api/v1/collectors/register \
  -H "X-Registration-Secret: $REGISTRATION_SECRET" \
  -d '{...}'
```

### Configuration Issues

#### Collector Fails to Start

```bash
# Validate TOML syntax
pganalytics --config /etc/pganalytics/collector.toml --validate

# Check file permissions
ls -la /etc/pganalytics/collector.*
# Should show: -rw------- or -rw-r-----

# Test with verbose logging
pganalytics --config /etc/pganalytics/collector.toml --verbose 2>&1 | head -100

# Check systemd service logs
sudo journalctl -u pganalytics-collector -n 100 -e
```

#### "Cannot Connect to PostgreSQL"

```bash
# Test PostgreSQL connection directly
psql -h localhost -U pganalytics_monitoring -d postgres

# Check pg_hba.conf
sudo -u postgres grep "pganalytics" /etc/postgresql/16/main/pg_hba.conf

# Verify host/port in config
grep -A5 "\[postgres\]" /etc/pganalytics/collector.toml

# Check firewall
sudo iptables -L -n | grep 5432
```

#### "Cannot Connect to Backend API"

```bash
# Test API connectivity
curl -v https://api.example.com/api/v1/health

# Check DNS resolution
nslookup api.example.com

# Check firewall/security group
telnet api.example.com 443

# Verify certificate trust
curl https://api.example.com --cacert /etc/pganalytics/ca.crt
```

---

## Summary

This guide provides complete procedures for:

1. **Registering collectors** with pre-shared secret
2. **Managing JWT tokens** with 1-year validity
3. **Configuring collectors** with security best practices
4. **Handling multi-collector** scenarios
5. **Rotating credentials** quarterly
6. **Troubleshooting** common issues

For additional information:
- `ENTERPRISE_INSTALLATION.md` - Full installation procedures
- `SECURITY.md` - Security guidelines
- `docs/ARCHITECTURE.md` - Technical architecture
- `docs/API_SECURITY_REFERENCE.md` - API security details

---

**Version:** 1.0
**Last Updated:** February 25, 2026
**Status:** Production Ready
