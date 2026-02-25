# Phase 1 Execution Checklist - pgAnalytics v3.2.0

**Objective:** Complete all pre-deployment setup for any infrastructure (AWS, on-premises, Kubernetes, Docker, physical machines)

**Timeline:** 8 hours

**Key Principle:** This deployment is environment-agnostic. You define your infrastructure, we handle the rest.

---

## SECTION 1: Prepare Your Deployment Configuration (30 minutes)

### Step 1.1: Copy the Configuration Template

```bash
# Copy the deployment configuration template
cp DEPLOYMENT_CONFIG_TEMPLATE.md ~/.env.pganalytics

# Make it executable
chmod 600 ~/.env.pganalytics
```

### Step 1.2: Choose Your Infrastructure Model

Choose ONE of the examples below and use as your starting point:

**Option A: AWS EC2 Distributed**
- PostgreSQL on RDS
- API on 2× EC2 t3.large instances
- Collectors on 5× EC2 c5.large instances
- Grafana on 1× EC2 t3.medium instance
- See: DEPLOYMENT_CONFIG_TEMPLATE.md → "Example 1"

**Option B: On-Premises / Physical Machines**
- PostgreSQL on dedicated physical server
- API on 2× dedicated physical servers
- Collectors on 5× dedicated physical servers
- Grafana on 1× dedicated physical server
- See: DEPLOYMENT_CONFIG_TEMPLATE.md → "Example 2"

**Option C: Kubernetes**
- PostgreSQL as K8s StatefulSet
- API as K8s Deployment with 2 replicas
- Collectors as K8s DaemonSet or StatefulSet
- Grafana as K8s Deployment
- See: DEPLOYMENT_CONFIG_TEMPLATE.md → "Example 4"

**Option D: Single Machine (Development/Testing)**
- Everything on one machine or using Docker Compose
- PostgreSQL + API + Collectors + Grafana
- See: DEPLOYMENT_CONFIG_TEMPLATE.md → "Example 3"

**Option E: Multi-Region**
- Distributed across multiple regions
- Primary and replica databases
- API servers in multiple regions
- Collectors in all regions
- See: DEPLOYMENT_CONFIG_TEMPLATE.md → "Example 5"

**Your chosen infrastructure model: _______________**

---

### Step 1.3: Fill in Your Configuration

Edit `~/.env.pganalytics` and fill in YOUR actual infrastructure values:

```bash
# Open the configuration file
nano ~/.env.pganalytics
```

**Minimum required values (everything else has defaults):**

```bash
# Database - Where is your PostgreSQL?
export DB_HOST="your-postgres-host"           # IP or hostname
export DB_PORT="5432"                         # PostgreSQL port
export DB_NAME="pganalytics"                  # Database name
export DB_USER="pganalytics"                  # Database user
export DB_PASSWORD="$(openssl rand -base64 32)"  # Generate random password
export DB_ADMIN_USER="postgres"               # Admin user
export DB_ADMIN_PASSWORD="admin-password"     # Admin password

# API Servers - Where will your API run?
export API_HOST_1="your-api-server-1"         # IP or hostname
export API_HOST_2="your-api-server-2"         # IP or hostname
export API_PORT="8080"                        # API port

# Collectors - Where are your collectors?
export COLLECTOR_HOSTS=(
  "collector-1-ip"
  "collector-2-ip"
  "collector-3-ip"
  "collector-4-ip"
  "collector-5-ip"
)

# Monitoring - Where will Grafana and Prometheus run?
export GRAFANA_HOST="your-grafana-host"       # IP or hostname
export PROMETHEUS_HOST="your-prometheus-host" # IP or hostname

# Secrets - Generate random ones
export JWT_SECRET_KEY="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_KEY="$(openssl rand -base64 32)"

# Your infrastructure
export DEPLOYMENT_MODE="distributed"  # Options: distributed, single-machine, kubernetes
export ENVIRONMENT="production"       # Options: production, staging, development
```

**Examples of real values:**

AWS environment:
```bash
export DB_HOST="pganalytics-prod.abc123def456.us-east-1.rds.amazonaws.com"
export API_HOST_1="10.0.1.50"
export API_HOST_2="10.0.1.51"
export COLLECTOR_HOSTS=("10.0.2.10" "10.0.2.11" "10.0.2.12" "10.0.2.13" "10.0.2.14")
export GRAFANA_HOST="10.0.3.100"
export AWS_REGION="us-east-1"
export DEPLOYMENT_MODE="distributed"
```

On-premises environment:
```bash
export DB_HOST="192.168.1.50"
export API_HOST_1="192.168.1.51"
export API_HOST_2="192.168.1.52"
export COLLECTOR_HOSTS=("192.168.1.60" "192.168.1.61" "192.168.1.62" "192.168.1.63" "192.168.1.64")
export GRAFANA_HOST="192.168.1.70"
export DEPLOYMENT_MODE="distributed"
export CONTAINER_RUNTIME="none"
```

---

### Step 1.4: Verify Configuration

```bash
# Source your configuration
source ~/.env.pganalytics

# Verify all values are set
echo "Database: $DB_HOST:$DB_PORT"
echo "API Servers: $API_HOST_1, $API_HOST_2"
echo "Collectors: ${COLLECTOR_HOSTS[@]}"
echo "Grafana: $GRAFANA_HOST"
echo "Deployment Mode: $DEPLOYMENT_MODE"
echo "Environment: $ENVIRONMENT"
```

**Configuration Status:**
- [ ] All required values filled in
- [ ] Configuration loaded successfully: `source ~/.env.pganalytics`
- [ ] No errors when echoing values

---

## SECTION 2: Prepare Your Infrastructure (Varies by Environment)

### AWS EC2 Approach

**If using AWS EC2:**

```bash
# Set your AWS configuration
export AWS_REGION="us-east-1"  # Change to your region
export AWS_PROFILE="default"   # Change if using different profile

# Verify AWS credentials
aws sts get-caller-identity

# Create RDS instance (if needed)
aws rds create-db-instance \
  --db-instance-identifier pganalytics-prod \
  --engine postgres \
  --engine-version 16.12 \
  --db-instance-class db.t3.large \
  --allocated-storage 100 \
  --master-username admin \
  --master-user-password "$(openssl rand -base64 32)" \
  --region $AWS_REGION

# Get RDS endpoint after it's created
aws rds describe-db-instances \
  --db-instance-identifier pganalytics-prod \
  --query 'DBInstances[0].Endpoint.Address'
```

### On-Premises / Physical Machines Approach

**If using physical machines:**

```bash
# 1. Verify PostgreSQL is installed and running
ssh postgres-admin@$DB_HOST "psql --version"

# 2. Verify API server is accessible
ssh api-admin@$API_HOST_1 "hostname"
ssh api-admin@$API_HOST_2 "hostname"

# 3. Verify collectors are accessible
for collector in "${COLLECTOR_HOSTS[@]}"; do
  ssh collector-admin@$collector "hostname"
done

# 4. Verify Grafana server is accessible
ssh grafana-admin@$GRAFANA_HOST "which grafana-server"

# 5. Verify Prometheus server is accessible
ssh prometheus-admin@$PROMETHEUS_HOST "which prometheus"
```

### Kubernetes Approach

**If using Kubernetes:**

```bash
# 1. Verify cluster access
kubectl config current-context

# 2. Create namespace
kubectl create namespace pganalytics

# 3. Verify cluster is ready
kubectl cluster-info
kubectl get nodes
```

### Docker Approach

**If using Docker:**

```bash
# 1. Verify Docker is running
docker ps

# 2. Create network
docker network create pganalytics-network

# 3. Start PostgreSQL
docker run -d \
  --name pganalytics-postgres \
  --network pganalytics-network \
  -e POSTGRES_PASSWORD=$(openssl rand -base64 32) \
  postgres:16.12
```

**Your Infrastructure Status:**
- [ ] Primary infrastructure verified
- [ ] Database server is accessible
- [ ] API servers are accessible
- [ ] Collectors are accessible
- [ ] Monitoring infrastructure is ready

---

## SECTION 3: Run Phase 1 Automation Script (45 minutes)

### Step 3.1: Make Script Executable

```bash
chmod +x /tmp/phase1_automated_setup_v2.sh
```

### Step 3.2: Run the Script

```bash
# Load your configuration
source ~/.env.pganalytics

# Run Phase 1 automation
bash /tmp/phase1_automated_setup_v2.sh
```

**What the script does automatically:**
1. ✓ Loads your configuration from `~/.env.pganalytics`
2. ✓ Verifies database connectivity
3. ✓ Verifies API server connectivity
4. ✓ Verifies collector connectivity
5. ✓ Generates cryptographic secrets (32-byte)
6. ✓ Stores secrets securely (AWS Secrets Manager or file-based)
7. ✓ Creates database user and role
8. ✓ Enables PostgreSQL monitoring extensions
9. ✓ Configures query logging
10. ✓ Configures automated backups

### Step 3.3: Review Output

The script will generate a summary file:

```bash
# Find and view the summary
ls -la /tmp/pganalytics_deployment_summary_*.txt
cat /tmp/pganalytics_deployment_summary_*.txt
```

**Expected output:**
- All 10 steps completed (green ✓ marks)
- No critical errors
- Secrets securely stored
- Database configured

---

## SECTION 4: Phase 1 Manual Steps (3-4 hours)

After running the automation script, complete these manual steps:

### Step 4.1: Deploy API Binary

**For each API server ($API_HOST_1 and $API_HOST_2):**

```bash
# Source configuration
source ~/.env.pganalytics

# On your local machine, build/get the API binary
# Option A: Build from source
cd pganalytics-v3/backend
go build -o pganalytics-api ./cmd/pganalytics-api
cd -

# Option B: Download pre-built binary
# wget https://github.com/torresglauco/pganalytics-v3/releases/download/v3.2.0/pganalytics-api

# Deploy to API server 1
scp -i ~/.ssh/your-key.pem pganalytics-v3/backend/pganalytics-api ubuntu@$API_HOST_1:/tmp/

ssh -i ~/.ssh/your-key.pem ubuntu@$API_HOST_1 << 'SCRIPT'
  # Create directories
  sudo mkdir -p /opt/pganalytics /var/log/pganalytics /etc/pganalytics
  sudo useradd -r -s /bin/bash pganalytics 2>/dev/null || true

  # Install binary
  sudo cp /tmp/pganalytics-api /opt/pganalytics/
  sudo chmod +x /opt/pganalytics/pganalytics-api
  sudo chown -R pganalytics:pganalytics /opt/pganalytics /var/log/pganalytics /etc/pganalytics

  echo "API binary deployed to $API_HOST_1"
SCRIPT

# Deploy to API server 2
scp -i ~/.ssh/your-key.pem pganalytics-v3/backend/pganalytics-api ubuntu@$API_HOST_2:/tmp/

ssh -i ~/.ssh/your-key.pem ubuntu@$API_HOST_2 << 'SCRIPT'
  sudo mkdir -p /opt/pganalytics /var/log/pganalytics /etc/pganalytics
  sudo useradd -r -s /bin/bash pganalytics 2>/dev/null || true
  sudo cp /tmp/pganalytics-api /opt/pganalytics/
  sudo chmod +x /opt/pganalytics/pganalytics-api
  sudo chown -R pganalytics:pganalytics /opt/pganalytics /var/log/pganalytics /etc/pganalytics
  echo "API binary deployed to $API_HOST_2"
SCRIPT
```

**Verification:**
- [ ] API binary exists on $API_HOST_1
- [ ] API binary exists on $API_HOST_2
- [ ] Permissions are correct (executable)

### Step 4.2: Configure API Services

**Create systemd service file on each API server:**

```bash
source ~/.env.pganalytics

# For each API server
for api_host in "$API_HOST_1" "$API_HOST_2"; do
  ssh -i ~/.ssh/your-key.pem ubuntu@$api_host << 'SCRIPT'
    # Create systemd service file
    sudo tee /etc/systemd/system/pganalytics-api.service > /dev/null << 'EOF'
[Unit]
Description=pgAnalytics API Server
After=network.target postgresql.service

[Service]
Type=simple
User=pganalytics
Group=pganalytics
ExecStart=/opt/pganalytics/pganalytics-api
WorkingDirectory=/opt/pganalytics
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment="DATABASE_URL=postgres://pganalytics:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
Environment="JWT_SECRET=${JWT_SECRET_KEY}"
Environment="REGISTRATION_SECRET=${REGISTRATION_SECRET}"
Environment="PORT=${API_PORT}"
Environment="LOG_LEVEL=info"

[Install]
WantedBy=multi-user.target
EOF

    # Enable and start service
    sudo systemctl daemon-reload
    sudo systemctl enable pganalytics-api
    sudo systemctl start pganalytics-api
    sudo systemctl status pganalytics-api

    echo "API service configured on $(hostname)"
  SCRIPT
done
```

**Verification:**
- [ ] systemd services created
- [ ] Services enabled
- [ ] Services started
- [ ] Health check: `curl http://localhost:$API_PORT/api/v1/health`

### Step 4.3: Setup Prometheus Monitoring

```bash
# SSH to Prometheus server (or install locally)
ssh -i ~/.ssh/your-key.pem ubuntu@$PROMETHEUS_HOST << 'SCRIPT'
  # Create Prometheus configuration
  sudo mkdir -p /etc/prometheus
  sudo tee /etc/prometheus/prometheus.yml > /dev/null << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'pganalytics-api'
    static_configs:
      - targets: ['${API_HOST_1}:${API_PORT}', '${API_HOST_2}:${API_PORT}']

  - job_name: 'collectors'
    static_configs:
      - targets:
EOF

  # Add collector targets
  for collector in "${COLLECTOR_HOSTS[@]}"; do
    echo "        - $collector:9090" >> /etc/prometheus/prometheus.yml
  done

  # Start Prometheus
  docker run -d \
    --name prometheus \
    -p 9090:9090 \
    -v /etc/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml \
    prom/prometheus:latest

  echo "Prometheus started on $PROMETHEUS_HOST:9090"
SCRIPT
```

**Verification:**
- [ ] Prometheus running on $PROMETHEUS_HOST:9090
- [ ] Prometheus scraping metrics from API servers
- [ ] Prometheus scraping metrics from collectors

### Step 4.4: Configure Grafana Datasource

```bash
source ~/.env.pganalytics

# SSH to Grafana server
ssh -i ~/.ssh/your-key.pem ubuntu@$GRAFANA_HOST << 'SCRIPT'
  # Add PostgreSQL datasource
  curl -X POST http://localhost:3000/api/datasources \
    -H "Content-Type: application/json" \
    -u admin:admin \
    -d '{
      "name": "PostgreSQL",
      "type": "postgres",
      "url": "postgres://'${DB_HOST}':'${DB_PORT}'/'${DB_NAME}'",
      "user": "'${DB_USER}'",
      "password": "'${DB_PASSWORD}'",
      "database": "'${DB_NAME}'",
      "isDefault": true
    }'

  # Add Prometheus datasource
  curl -X POST http://localhost:3000/api/datasources \
    -H "Content-Type: application/json" \
    -u admin:admin \
    -d '{
      "name": "Prometheus",
      "type": "prometheus",
      "url": "http://'${PROMETHEUS_HOST}':9090",
      "isDefault": false
    }'

  echo "Grafana datasources configured"
SCRIPT
```

**Verification:**
- [ ] Grafana accessible at $GRAFANA_HOST:3000
- [ ] PostgreSQL datasource configured
- [ ] Prometheus datasource configured
- [ ] Test query returns data

### Step 4.5: Import Grafana Dashboards

```bash
source ~/.env.pganalytics

# Copy dashboards to Grafana server
scp -i ~/.ssh/your-key.pem grafana/dashboards/*.json ubuntu@$GRAFANA_HOST:/tmp/

# Import dashboards
ssh -i ~/.ssh/your-key.pem ubuntu@$GRAFANA_HOST << 'SCRIPT'
  for dashboard in /tmp/*.json; do
    curl -X POST http://localhost:3000/api/dashboards/db \
      -H "Content-Type: application/json" \
      -u admin:admin \
      -d @"$dashboard"
  done
  echo "All dashboards imported"
SCRIPT
```

**Verification:**
- [ ] All dashboards imported
- [ ] Dashboards showing data
- [ ] No query errors

---

## SECTION 5: Register Collectors

After API servers are running:

```bash
source ~/.env.pganalytics

# For each collector
for i in "${!COLLECTOR_HOSTS[@]}"; do
  collector_host="${COLLECTOR_HOSTS[$i]}"
  collector_id="collector-prod-$((i+1))"

  # Register with API
  curl -X POST "https://$API_HOST_1:$API_PORT/api/v1/collectors/register" \
    -H "X-Registration-Secret: $REGISTRATION_SECRET" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "'$collector_id'",
      "hostname": "'$collector_host'",
      "type": "postgresql"
    }'

  echo "Registered $collector_id"
done
```

**Verification:**
- [ ] All collectors registered
- [ ] Collectors sending metrics
- [ ] Metrics appearing in Grafana

---

## SECTION 6: Health Checks

Run these validation checks:

```bash
source ~/.env.pganalytics

echo "=== Health Checks ==="

# API Health
for api_host in "$API_HOST_1" "$API_HOST_2"; do
  curl -s "http://$api_host:$API_PORT/api/v1/health" | jq .
done

# Database Health
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1"

# Prometheus Health
curl -s "http://$PROMETHEUS_HOST:$PROMETHEUS_PORT/-/healthy"

# Grafana Health
curl -s "http://$GRAFANA_HOST:$GRAFANA_PORT/api/health" | jq .

# Collector metrics
curl -s "http://$API_HOST_1:$API_PORT/api/v1/metrics" | jq '.[] | .count' | paste -sd+ | bc
```

**Health Check Status:**
- [ ] API servers responding 200 OK
- [ ] Database connection successful
- [ ] Prometheus healthy
- [ ] Grafana healthy
- [ ] Metrics flowing

---

## SECTION 7: Team Sign-Offs

- [ ] **Infrastructure Team** - Infrastructure verified and operational
  - Signed: _________________ | Date: _________

- [ ] **Security Team** - Security configuration approved
  - Signed: _________________ | Date: _________

- [ ] **Database Team** - Database setup verified
  - Signed: _________________ | Date: _________

- [ ] **Deployment Lead** - Phase 1 complete, ready for Phase 2
  - Signed: _________________ | Date: _________

---

## SECTION 8: Success Criteria

All of these must be true:

```
✅ Configuration file created and filled with real values
✅ Phase 1 automation script completed successfully
✅ Database user and role created
✅ API binary deployed to both servers
✅ API services running and responding to health checks
✅ Prometheus collecting metrics
✅ Grafana datasources configured
✅ Grafana dashboards imported and showing data
✅ Collectors registered with API
✅ All health checks passing
✅ Team sign-offs obtained
```

---

## SECTION 9: Troubleshooting

### Database Connection Failed
```bash
# Check database is running
psql -h $DB_HOST -U $DB_ADMIN_USER -c "SELECT 1"

# Check credentials
echo "User: $DB_USER, Password: $DB_PASSWORD, Host: $DB_HOST:$DB_PORT"

# Check firewall
telnet $DB_HOST $DB_PORT
```

### API Service Won't Start
```bash
# Check service status
sudo systemctl status pganalytics-api

# Check logs
sudo journalctl -u pganalytics-api -f

# Check binary is executable
ls -la /opt/pganalytics/pganalytics-api
```

### Prometheus Not Scraping
```bash
# Check Prometheus config
cat /etc/prometheus/prometheus.yml

# Check targets in Prometheus UI
# http://prometheus-host:9090/targets

# Check API is responding
curl http://$API_HOST_1:$API_PORT/api/v1/health
```

### Grafana Datasource Error
```bash
# Check database connectivity from Grafana server
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1"

# Test in Grafana UI: Configuration → Data Sources → Test
```

---

## SECTION 10: Next Steps

**Phase 1 Complete! ✅**

Now proceed to Phase 2 (Staging):

1. Read: `DEPLOYMENT_PLAN_v3.2.0.md` Phase 2 section
2. Deploy to staging environment (same configuration)
3. Run smoke tests against staging
4. Execute load testing
5. Get sign-off for production

**Phase 2 Timeline:** Wednesday, February 26 (8 hours)

---

## Reference Documents

- **DEPLOYMENT_CONFIG_TEMPLATE.md** - Configuration template and examples
- **DEPLOYMENT_PLAN_v3.2.0.md** - Complete 4-phase deployment plan
- **ENTERPRISE_INSTALLATION.md** - Detailed multi-server installation
- **docs/COLLECTOR_REGISTRATION_GUIDE.md** - Collector registration
- **QUICK_REFERENCE.md** - Quick reference guide

---

**Status: READY FOR PHASE 1 EXECUTION**

Your deployment is environment-agnostic and works with AWS, on-premises, Kubernetes, Docker, or any infrastructure. Fill in your configuration and run Phase 1!
