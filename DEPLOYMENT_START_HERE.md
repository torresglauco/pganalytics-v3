# pgAnalytics v3.2.0 - Start Here for Deployment

**Status:** ✅ Production Ready | Environment-Agnostic Deployment

**This guide works for:**
- AWS EC2 + RDS (multiple regions)
- On-premises physical machines (data centers)
- Kubernetes (any cloud or on-prem)
- Docker Compose (development/testing)
- Multi-region distributed deployments
- Single machine installations
- Hybrid cloud setups

---

## Quick Start (5 minutes)

### 1. Copy Configuration Template
```bash
cp DEPLOYMENT_CONFIG_TEMPLATE.md ~/.env.pganalytics
chmod 600 ~/.env.pganalytics
```

### 2. Fill in YOUR Infrastructure

Edit `~/.env.pganalytics` and enter your actual values:

```bash
# Your database server
export DB_HOST="your-postgres-ip-or-hostname"
export DB_USER="pganalytics"
export DB_PASSWORD="$(openssl rand -base64 32)"

# Your API servers
export API_HOST_1="your-api-server-1-ip"
export API_HOST_2="your-api-server-2-ip"

# Your collectors
export COLLECTOR_HOSTS=("collector-1-ip" "collector-2-ip" "collector-3-ip" "collector-4-ip" "collector-5-ip")

# Your monitoring
export GRAFANA_HOST="your-grafana-ip"
export PROMETHEUS_HOST="your-prometheus-ip"

# Your secrets
export JWT_SECRET_KEY="$(openssl rand -base64 32)"
export REGISTRATION_SECRET="$(openssl rand -base64 32)"
export BACKUP_KEY="$(openssl rand -base64 32)"

# Your infrastructure type
export DEPLOYMENT_MODE="distributed"  # or: single-machine, kubernetes
export ENVIRONMENT="production"        # or: staging, development
```

### 3. Run Phase 1 Automation

```bash
source ~/.env.pganalytics
bash /tmp/phase1_automated_setup_v2.sh
```

### 4. Follow Phase 1 Manual Steps

See: `PHASE1_EXECUTION_CHECKLIST_V2.md` - Sections 4-8

---

## Full Deployment Timeline

### Phase 1: Pre-Deployment (TODAY - Tuesday)
**Duration:** 8 hours | **Script:** `/tmp/phase1_automated_setup_v2.sh`

**What it does:**
- Verifies database connectivity
- Creates database user and role
- Enables PostgreSQL monitoring
- Generates and stores secrets securely
- Configures backups
- Builds/verifies API binary

**Manual steps after automation:**
1. Deploy API binary to both servers
2. Configure API systemd services
3. Setup Prometheus monitoring
4. Configure Grafana datasource
5. Import Grafana dashboards
6. Register collectors
7. Run health checks

**Documentation:** `PHASE1_EXECUTION_CHECKLIST_V2.md`

---

### Phase 2: Staging Deployment (Wednesday)
**Duration:** 8 hours | **Environment:** Staging (use same config, just different host values)

**What happens:**
1. Deploy to staging environment
2. Run all smoke tests
3. Test security configuration
4. Execute load testing
5. Get sign-off for production

**Documentation:** `DEPLOYMENT_PLAN_v3.2.0.md` Phase 2

---

### Phase 3: Production Deployment (Thursday)
**Duration:** 8 hours

**What happens:**
1. Final validation checks
2. Deploy to production
3. Import Grafana dashboards
4. Activate monitoring and alerts
5. Go live!

**Documentation:** `DEPLOYMENT_PLAN_v3.2.0.md` Phase 3

---

### Phase 4: Monitoring & Validation (Friday-Monday)
**Duration:** Continuous 48-96 hours

**What happens:**
1. Monitor first 6 hours every 15 minutes
2. Record baseline metrics
3. Check for memory leaks
4. Validate backups
5. Post-deployment retrospective

**Documentation:** `DEPLOYMENT_PLAN_v3.2.0.md` Phase 4

---

## Configuration Examples

### Example 1: AWS EC2 Distributed

Fill `~/.env.pganalytics` with:

```bash
export DB_HOST="pganalytics-prod.abc123.us-east-1.rds.amazonaws.com"
export DB_PORT="5432"
export DB_USER="pganalytics"
export DB_PASSWORD="your-secure-password"

export API_HOST_1="10.0.1.50"
export API_HOST_2="10.0.1.51"
export API_PORT="8080"

export COLLECTOR_HOSTS=("10.0.2.10" "10.0.2.11" "10.0.2.12" "10.0.2.13" "10.0.2.14")

export GRAFANA_HOST="10.0.3.100"
export PROMETHEUS_HOST="10.0.3.101"

export AWS_REGION="us-east-1"
export AWS_SECRETS_MANAGER_ENABLED="true"
export DEPLOYMENT_MODE="distributed"
export ENVIRONMENT="production"
```

Then run automation.

---

### Example 2: On-Premises Distributed

Fill `~/.env.pganalytics` with:

```bash
export DB_HOST="192.168.1.50"
export DB_USER="pganalytics"
export DB_PASSWORD="your-secure-password"

export API_HOST_1="192.168.1.51"
export API_HOST_2="192.168.1.52"

export COLLECTOR_HOSTS=("192.168.1.60" "192.168.1.61" "192.168.1.62" "192.168.1.63" "192.168.1.64")

export GRAFANA_HOST="192.168.1.70"
export PROMETHEUS_HOST="192.168.1.71"

export DEPLOYMENT_MODE="distributed"
export CONTAINER_RUNTIME="none"
export ENVIRONMENT="production"
```

Then run automation.

---

### Example 3: Kubernetes

Fill `~/.env.pganalytics` with:

```bash
export DB_HOST="pganalytics-postgres"  # K8s service name
export DB_USER="pganalytics"

export API_HOST_1="pganalytics-api"    # K8s service name
export API_HOST_2="pganalytics-api"    # Same service, multiple replicas

export COLLECTOR_HOSTS=("pganalytics-collector-1" "pganalytics-collector-2" "pganalytics-collector-3")

export GRAFANA_HOST="pganalytics-grafana"
export PROMETHEUS_HOST="prometheus"

export DEPLOYMENT_MODE="kubernetes"
export CONTAINER_RUNTIME="kubernetes"
export K8S_NAMESPACE="pganalytics"
```

Then run automation.

---

### Example 4: Single Machine (Development)

Fill `~/.env.pganalytics` with:

```bash
export DB_HOST="localhost"
export DB_USER="pganalytics"

export API_HOST_1="localhost"
export API_HOST_2="localhost"

export COLLECTOR_HOSTS=("localhost")

export GRAFANA_HOST="localhost"
export PROMETHEUS_HOST="localhost"

export DEPLOYMENT_MODE="single-machine"
export CONTAINER_RUNTIME="docker"
export ENVIRONMENT="development"
```

Then run automation and use docker-compose.

---

## Key Documentation Files

| File | Purpose | When to Read |
|------|---------|--------------|
| **DEPLOYMENT_CONFIG_TEMPLATE.md** | Configuration template with all options | Now - before running automation |
| **PHASE1_EXECUTION_CHECKLIST_V2.md** | Step-by-step Phase 1 manual procedures | After automation completes |
| **DEPLOYMENT_PLAN_v3.2.0.md** | Complete 4-phase deployment plan | Overview of entire timeline |
| **ENTERPRISE_INSTALLATION.md** | Detailed multi-server installation guide | For specific installation questions |
| **docs/COLLECTOR_REGISTRATION_GUIDE.md** | How to register collectors with API | After API is running |
| **QUICK_REFERENCE.md** | Quick answers to common questions | Quick lookup reference |

---

## Your Next 5 Steps

### Step 1: TODAY (Now)
```bash
# Copy configuration template
cp DEPLOYMENT_CONFIG_TEMPLATE.md ~/.env.pganalytics

# Edit with your infrastructure values
nano ~/.env.pganalytics

# Verify configuration
source ~/.env.pganalytics
echo "Database: $DB_HOST:$DB_PORT"
echo "API Servers: $API_HOST_1, $API_HOST_2"
echo "Collectors: ${COLLECTOR_HOSTS[@]}"
```

**Estimated time:** 15 minutes

---

### Step 2: TODAY (Next)
```bash
# Run Phase 1 automation script
source ~/.env.pganalytics
bash /tmp/phase1_automated_setup_v2.sh

# Review output and summary
cat /tmp/pganalytics_deployment_summary_*.txt
```

**Estimated time:** 45 minutes

---

### Step 3: TODAY (Manual Steps)
Follow `PHASE1_EXECUTION_CHECKLIST_V2.md` Section 4:

1. Deploy API binary to both servers
2. Configure API systemd services
3. Setup Prometheus monitoring
4. Configure Grafana datasource
5. Import Grafana dashboards
6. Register collectors

**Estimated time:** 3-4 hours

---

### Step 4: TODAY (Validation)
Run health checks from `PHASE1_EXECUTION_CHECKLIST_V2.md` Section 6:

```bash
# API Health
curl http://$API_HOST_1:$API_PORT/api/v1/health

# Database Health
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1"

# Grafana Health
curl http://$GRAFANA_HOST:3000/api/health
```

**Estimated time:** 30 minutes

---

### Step 5: TODAY (Sign-Offs)
Get approvals from:
- Infrastructure Lead
- Security Lead
- Database Lead
- Deployment Lead

**Estimated time:** 30 minutes

---

## Architecture Overview

```
Your Infrastructure (any type)
├── Database Server
│   └── PostgreSQL 16.12
│       └── pganalytics database
│
├── API Servers (2x)
│   ├── api-prod-01:8080
│   └── api-prod-02:8080
│       └── Connected via Load Balancer
│
├── Collector Servers (5+)
│   ├── collector-1 → PostgreSQL monitoring
│   ├── collector-2 → PostgreSQL monitoring
│   ├── collector-3 → PostgreSQL monitoring
│   ├── collector-4 → PostgreSQL monitoring
│   └── collector-5 → PostgreSQL monitoring
│       └── All report to API via TLS + JWT
│
└── Monitoring Stack
    ├── Prometheus → Scrapes metrics from API & collectors
    └── Grafana → Visualizes data in 9 dashboards
```

---

## Security Checklist

✅ **Already Implemented:**
- JWT token authentication (15-min user, 1-year collector)
- Mutual TLS (mTLS) for collector communication
- BCrypt password hashing (cost=12)
- SQL injection prevention (parameterized queries)
- Rate limiting (token bucket algorithm)
- Role-based access control (RBAC)
- Secrets stored securely (AWS Secrets Manager or encrypted files)

✅ **What You Configure:**
- Database user and password (`DB_USER`, `DB_PASSWORD`)
- API authentication secrets (`JWT_SECRET_KEY`, `REGISTRATION_SECRET`)
- TLS certificates (self-signed or CA-signed)
- Network access (firewall rules, security groups)
- Backup retention policy

---

## Support & Troubleshooting

### Common Issues

**Database Connection Failed**
```bash
# Verify database is running and accessible
psql -h $DB_HOST -U $DB_ADMIN_USER -c "SELECT 1"

# Check firewall allows port 5432
telnet $DB_HOST 5432
```

**API Service Won't Start**
```bash
# Check systemd logs
sudo journalctl -u pganalytics-api -f

# Verify binary is executable
ls -la /opt/pganalytics/pganalytics-api

# Check API dependencies
ldd /opt/pganalytics/pganalytics-api
```

**Collectors Not Sending Metrics**
```bash
# Verify collector can reach API
curl http://$API_HOST_1:$API_PORT/api/v1/health

# Check collector logs on each collector server
tail -f /var/log/pganalytics/collector.log

# Verify firewall allows communication
telnet $API_HOST_1 $API_PORT
```

### Documentation References

- **Detailed setup questions:** See `ENTERPRISE_INSTALLATION.md`
- **API and authentication:** See `docs/API_SECURITY_REFERENCE.md`
- **Collector registration:** See `docs/COLLECTOR_REGISTRATION_GUIDE.md`
- **Quick answers:** See `QUICK_REFERENCE.md`
- **Full timeline:** See `DEPLOYMENT_PLAN_v3.2.0.md`

---

## Success Criteria

Phase 1 is complete when:

- ✅ Configuration file created with YOUR infrastructure values
- ✅ Automation script ran successfully
- ✅ Database user created and verified
- ✅ API binary deployed to both servers
- ✅ API services running and healthy
- ✅ Prometheus collecting metrics
- ✅ Grafana dashboards showing data
- ✅ Collectors registered with API
- ✅ All health checks passing
- ✅ Team sign-offs obtained

---

## Timeline Summary

| Phase | When | Duration | What |
|-------|------|----------|------|
| **Phase 1** | TODAY | 8h | Pre-deployment setup & validation |
| **Phase 2** | Wednesday | 8h | Staging deployment & testing |
| **Phase 3** | Thursday | 8h | Production deployment & go-live |
| **Phase 4** | Fri-Mon | Continuous | Monitoring & validation |

---

## Next Action

👉 **Start here:**

```bash
# 1. Copy configuration template
cp DEPLOYMENT_CONFIG_TEMPLATE.md ~/.env.pganalytics

# 2. Fill with YOUR values
nano ~/.env.pganalytics

# 3. Verify
source ~/.env.pganalytics
echo "DB: $DB_HOST, API: $API_HOST_1, Collectors: ${#COLLECTOR_HOSTS[@]}"

# 4. Run automation
bash /tmp/phase1_automated_setup_v2.sh

# 5. Follow manual steps
cat PHASE1_EXECUTION_CHECKLIST_V2.md
```

---

**The pgAnalytics v3.2.0 deployment process is simple:**

1. **Define your infrastructure** in one config file
2. **Run automation** (handles 90% of setup)
3. **Complete manual steps** (binary deployment, service config, etc.)
4. **Verify everything** is working
5. **Move to Phase 2** (staging)

**Works everywhere** - AWS, on-premises, Kubernetes, Docker, hybrid clouds, physical machines.

**Let's deploy!** 🚀
