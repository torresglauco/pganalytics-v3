# pgAnalytics v3 - Enterprise Installation Guide

**Version:** 1.0
**Date:** February 25, 2026
**Status:** Production Ready
**Audience:** Enterprise Architects, DevOps Engineers, Database Administrators

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [PostgreSQL Installation](#postgresql-installation)
3. [Backend Installation](#backend-installation)
4. [Collector Installation](#collector-installation)
5. [Grafana Installation](#grafana-installation)
6. [Networking & Security](#networking--security)
7. [Multi-Region/Multi-Collector Setup](#multi-regionmulti-collector-setup)
8. [Troubleshooting](#troubleshooting)

---

## Architecture Overview

### Deployment Models

#### Model 1: Single Machine (Development/Small Deployments)
```
┌─────────────────────────────────────┐
│      Single Server (Ubuntu/CentOS)   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │   PostgreSQL 16+ (Port 5432) │   │
│  │   - pganalytics database     │   │
│  │   - timescale extension      │   │
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  Backend API (Port 8080)     │   │
│  │  - Go binary                 │   │
│  │  - JWT auth                  │   │
│  │  - Collector registration    │   │
│  └─────────────────────────────┘   │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  Grafana (Port 3000)         │   │
│  │  - Dashboards                │   │
│  │  - Alerting                  │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

#### Model 2: Distributed (Production)
```
┌─────────────────────────────────────────────────────────────────┐
│                      Load Balancer / Reverse Proxy               │
│                      (nginx / HAProxy)                           │
└────────────┬──────────────────────────┬─────────────────────────┘
             │                          │
    ┌────────▼──────────┐      ┌────────▼──────────┐
    │  Backend Server 1  │      │  Backend Server 2  │
    │  (Port 8080)       │      │  (Port 8080)       │
    │  - API             │      │  - API             │
    │  - Go binary       │      │  - Go binary       │
    │  - Collector mgmt  │      │  - Collector mgmt  │
    └────────┬───────────┘      └────────┬───────────┘
             │                           │
             └────────┬──────────────────┘
                      │
        ┌─────────────▼──────────────┐
        │  PostgreSQL Server (HA)    │
        │  - Primary instance        │
        │  - Replication standby     │
        │  - Port 5432              │
        │  - Backups enabled         │
        └─────────────┬──────────────┘
                      │
    ┌─────────────────┴─────────────────┐
    │ Collector 1          Collector 2    │ ... Collector N
    │ (EC2/Bare Metal)     (EC2/Bare Metal)│   (EC2/Bare Metal)
    │ - C++ binary         - C++ binary    │   - C++ binary
    │ - JWT auth           - JWT auth      │   - JWT auth
    │ - TLS 1.3            - TLS 1.3       │   - TLS 1.3
    └──────────────────────────────────────┘

    ┌─────────────────────────────┐
    │  Grafana Server             │
    │  (Separate Instance)        │
    │  - Dashboards              │
    │  - Alert notifications     │
    │  - RBAC users              │
    └─────────────────────────────┘
```

### Component Communication Flow

```
Collector Process:
1. Read PostgreSQL statistics (pg_stat_statements, pg_stat_replication, etc.)
2. Collect system metrics (CPU, memory, disk I/O)
3. Build metrics payload (JSON)
4. HTTP POST to Backend /api/v1/metrics/push with JWT token
5. Backend validates JWT, stores in TimescaleDB
6. Grafana queries TimescaleDB via PostgreSQL datasource
7. Dashboards display metrics in real-time

Security:
- Collector ←→ Backend: TLS 1.3 + JWT tokens (1-year expiration)
- Backend ←→ Database: TLS connection (sslmode=require)
- Grafana ←→ Database: TLS + authenticated connection
- User ←→ Backend: HTTPS + JWT tokens (short-lived)
```

---

## PostgreSQL Installation

### Prerequisites Checklist

- [ ] SSH access to target server
- [ ] Root or sudo privileges
- [ ] Network access to port 5432
- [ ] Disk space: minimum 100GB for metrics
- [ ] 8+ CPU cores recommended
- [ ] 16GB+ RAM recommended

### 1. Ubuntu/Debian Installation

#### 1a. Install PostgreSQL 16

```bash
# Update package lists
sudo apt-get update
sudo apt-get upgrade -y

# Install PostgreSQL repository
sudo apt-get install -y gnupg2 wget ca-certificates
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo apt-key add -

# Install PostgreSQL 16
sudo apt-get update
sudo apt-get install -y postgresql-16 postgresql-contrib-16 postgresql-16-timescaledb

# Start and enable PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Verify installation
sudo -u postgres psql --version
```

#### 1b. Install PostgreSQL 15, 14, or Older

```bash
# Replace "16" with your version (15, 14, 13, 12)
sudo apt-get install -y postgresql-15 postgresql-contrib-15 postgresql-15-timescaledb

# For compatibility with PostgreSQL 9.4-11 (legacy):
# Note: These versions are EOL. Consider upgrading if possible.
sudo apt-get install -y postgresql-11 postgresql-contrib-11
```

#### 1c. Configure PostgreSQL

```bash
# Switch to postgres user
sudo -u postgres psql

# Create pganalytics role with monitoring privileges
CREATE ROLE pganalytics WITH LOGIN NOINHERIT;
ALTER ROLE pganalytics WITH PASSWORD 'YOUR-SECURE-PASSWORD';
GRANT pg_monitor TO pganalytics;

# Create pganalytics database
CREATE DATABASE pganalytics OWNER pganalytics;

# Create timescale database
CREATE DATABASE timescale OWNER pganalytics;

# Exit psql
\q
```

#### 1d. Enable TimescaleDB Extension

```bash
# Connect as pganalytics user
psql -U pganalytics -d pganalytics

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- Enable pgstattuple for index stats
CREATE EXTENSION IF NOT EXISTS pgstattuple;

-- Enable pg_stat_statements for query analysis
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Exit
\q
```

#### 1e. Configure TLS/SSL

```bash
# Generate self-signed certificate (development) or use CA-signed (production)
sudo mkdir -p /etc/postgresql/16/main/certs
sudo openssl req -new -x509 -days 365 -nodes \
  -out /etc/postgresql/16/main/certs/server.crt \
  -keyout /etc/postgresql/16/main/certs/server.key

# Set permissions
sudo chown postgres:postgres /etc/postgresql/16/main/certs/*
sudo chmod 600 /etc/postgresql/16/main/certs/server.key

# Edit postgresql.conf
sudo nano /etc/postgresql/16/main/postgresql.conf

# Find and uncomment these lines:
# ssl = on
# ssl_cert_file = '/etc/postgresql/16/main/certs/server.crt'
# ssl_key_file = '/etc/postgresql/16/main/certs/server.key'
# ssl_protocols = 'TLSv1.2,TLSv1.3'

# Restart PostgreSQL
sudo systemctl restart postgresql
```

#### 1f. Configure pg_hba.conf for Network Access

```bash
# Edit pg_hba.conf
sudo nano /etc/postgresql/16/main/pg_hba.conf

# Add these lines (adjust IP ranges as needed):
# TYPE  DATABASE        USER            ADDRESS                 METHOD
# Local connections (unix socket)
local   all             all                                     trust

# IPv4 connections - local network (TLS required)
hostssl all             pganalytics     192.168.1.0/24          password

# IPv4 connections - specific API servers (TLS required)
hostssl pganalytics     pganalytics     203.0.113.10/32         password
hostssl pganalytics     pganalytics     203.0.113.11/32         password

# Reload PostgreSQL
sudo systemctl reload postgresql
```

#### 1g. Configure PostgreSQL Performance

```bash
# Edit postgresql.conf
sudo nano /etc/postgresql/16/main/postgresql.conf

# Adjust for your hardware (16GB RAM example):
shared_buffers = 4GB                    # 25% of RAM
effective_cache_size = 12GB             # 75% of RAM
maintenance_work_mem = 1GB              # shared_buffers / 4
checkpoint_completion_target = 0.9
wal_buffers = 16MB
max_wal_size = 2GB
min_wal_size = 1GB
work_mem = 16MB                         # shared_buffers / max_connections
max_parallel_workers_per_gather = 4
max_parallel_workers = 8

# Enable query logging for pganalytics role
log_min_duration_statement = 1000       # Log queries > 1 second
log_connections = on
log_disconnections = on
log_statement = 'all'

# Restart PostgreSQL
sudo systemctl restart postgresql
```

#### 1h. Enable Replication (HA Setup)

```bash
# On PRIMARY server
sudo nano /etc/postgresql/16/main/postgresql.conf

# Configure:
wal_level = replica
max_wal_senders = 3
max_replication_slots = 3
hot_standby = on

# Edit pg_hba.conf to allow replication:
# TYPE  DATABASE        USER            ADDRESS                 METHOD
hostssl replication     pganalytics     STANDBY_IP/32           password

sudo systemctl restart postgresql

# Create replication slot
sudo -u postgres psql -c "SELECT * FROM pg_create_physical_replication_slot('pganalytics_slot', false);"

# On STANDBY server
# Create base backup from primary
sudo -u postgres pg_basebackup \
  -h PRIMARY_IP \
  -U pganalytics \
  -D /var/lib/postgresql/16/main \
  -v -P --wal-method=stream \
  -W -R -S pganalytics_slot

# Create standby.signal file
sudo touch /var/lib/postgresql/16/main/standby.signal

# Adjust permissions
sudo chown -R postgres:postgres /var/lib/postgresql/16/main

# Start standby
sudo systemctl start postgresql
```

### 2. RHEL/CentOS Installation

#### 2a. Install PostgreSQL 16

```bash
# Enable PostgreSQL repository
sudo dnf install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-8-x86_64/pgdg-redhat-repo-latest.noarch.rpm

# Install PostgreSQL
sudo dnf install -y postgresql16-server postgresql16-contrib

# Initialize database
sudo /usr/pgsql-16/bin/postgresql-16-setup initdb

# Enable and start PostgreSQL
sudo systemctl enable postgresql-16
sudo systemctl start postgresql-16

# Verify
psql --version
```

#### 2b. Configure PostgreSQL (same as Ubuntu above)

Reference section 1c-1h for configuration steps.

### 3. Cloud-Managed PostgreSQL Options

#### 3a. AWS RDS PostgreSQL 16

```bash
# Create RDS instance via AWS Console or CLI
aws rds create-db-instance \
  --db-instance-identifier pganalytics-prod \
  --db-instance-class db.r6i.2xlarge \
  --engine postgres \
  --engine-version 16.2 \
  --allocated-storage 100 \
  --storage-type gp3 \
  --master-username pganalytics \
  --master-user-password 'YOUR-SECURE-PASSWORD' \
  --db-name pganalytics \
  --db-subnet-group-name default \
  --publicly-accessible false \
  --enable-iam-database-authentication \
  --enable-cloudwatch-logs-exports postgresql \
  --backup-retention-period 30 \
  --multi-az \
  --storage-encrypted \
  --kms-key-id arn:aws:kms:us-east-1:ACCOUNT:key/KEY-ID

# Wait for instance to be available
aws rds describe-db-instances --db-instance-identifier pganalytics-prod

# Connect and configure
psql -h pganalytics-prod.XXXX.us-east-1.rds.amazonaws.com \
  -U pganalytics -d pganalytics

# Create pganalytics role and database (same as above)
```

**Advantages:**
- Automatic backups
- High availability with Multi-AZ
- Automated patching
- Encryption at rest and in transit
- Parameter Groups for tuning

**Considerations:**
- Limited direct OS access
- Additional costs
- Network latency for collectors

#### 3b. Google Cloud SQL PostgreSQL 16

```bash
# Create Cloud SQL instance
gcloud sql instances create pganalytics-prod \
  --database-version=POSTGRES_16 \
  --tier=db-custom-4-16384 \
  --region=us-central1 \
  --database-flags \
    shared_preload_libraries=pg_stat_statements,pgaudit \
  --backup-start-time=03:00 \
  --enable-bin-log \
  --storage-size=100GB \
  --storage-auto-increase \
  --availability-type=REGIONAL

# Create database
gcloud sql databases create pganalytics \
  --instance=pganalytics-prod

# Create user
gcloud sql users create pganalytics \
  --instance=pganalytics-prod \
  --password='YOUR-SECURE-PASSWORD'

# Get connection string
gcloud sql instances describe pganalytics-prod \
  --format="value(connectionNames[0])"
```

#### 3c. Azure Database for PostgreSQL

```bash
# Create Azure Database instance
az postgres flexible-server create \
  --resource-group pganalytics-rg \
  --name pganalytics-prod \
  --location eastus \
  --admin-user pganalytics \
  --admin-password 'YOUR-SECURE-PASSWORD' \
  --sku-name Standard_B4ms \
  --storage-size 102400 \
  --version 16 \
  --high-availability Enabled \
  --backup-retention 30

# Create database
az postgres flexible-server db create \
  --resource-group pganalytics-rg \
  --server-name pganalytics-prod \
  --database-name pganalytics

# Configure firewall
az postgres flexible-server firewall-rule create \
  --resource-group pganalytics-rg \
  --name pganalytics-prod \
  --server-name pganalytics-prod \
  --start-ip-address 203.0.113.10 \
  --end-ip-address 203.0.113.11
```

### 4. Verify PostgreSQL Installation

```bash
# Connect to database
psql -U pganalytics -d pganalytics -h localhost

-- Check version
SELECT version();

-- Check extensions
\dx

-- Check pg_monitor role
SELECT * FROM pg_roles WHERE rolname = 'pganalytics';

-- Verify TimescaleDB
CREATE TABLE test_hypertable (
  time BIGINT NOT NULL,
  metric_name TEXT NOT NULL,
  value FLOAT8 NOT NULL
);

SELECT create_hypertable('test_hypertable', 'time', if_not_exists => TRUE);

-- Drop test table
DROP TABLE test_hypertable;

-- Exit
\q
```

---

## Backend Installation

### Prerequisites Checklist

- [ ] PostgreSQL 16+ running and accessible
- [ ] Go 1.21+ installed
- [ ] libpq development library installed
- [ ] Dedicated server (physical or VM)
- [ ] 4+ CPU cores
- [ ] 8GB+ RAM
- [ ] 50GB+ disk space for binary and logs

### 1. Ubuntu/Debian Installation

#### 1a. Install Dependencies

```bash
# Update packages
sudo apt-get update
sudo apt-get upgrade -y

# Install Go
wget https://go.dev/dl/go1.22.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
source /etc/profile

# Verify Go installation
go version

# Install PostgreSQL client and development libraries
sudo apt-get install -y postgresql-client libpq-dev

# Install build tools
sudo apt-get install -y git build-essential
```

#### 1b. Build Backend Binary

```bash
# Clone repository or use your release archive
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3

# Download dependencies
cd backend
go mod download

# Build binary
go build -o pganalytics-api ./cmd/pganalytics-api

# Verify binary
./pganalytics-api --version

# Copy to installation directory
sudo mkdir -p /opt/pganalytics
sudo cp pganalytics-api /opt/pganalytics/
sudo chmod +x /opt/pganalytics/pganalytics-api
```

#### 1c. Configure Environment Variables

```bash
# Create .env file
sudo nano /opt/pganalytics/.env

# Add (adjust values for your environment):
export DATABASE_URL="postgresql://pganalytics:PASSWORD@localhost:5432/pganalytics?sslmode=require"
export TIMESCALE_URL="postgresql://pganalytics:PASSWORD@localhost:5432/pganalytics?sslmode=require"
export JWT_SECRET="your-64-character-random-secret-key-here"
export REGISTRATION_SECRET="your-32-character-registration-secret"
export PORT=8080
export ENVIRONMENT="production"
export TLS_ENABLED="true"
export TLS_CERT="/etc/pganalytics/backend.crt"
export TLS_KEY="/etc/pganalytics/backend.key"
export LOG_LEVEL="info"
export CORS_ORIGINS="https://grafana.example.com,https://api.example.com"
```

#### 1d. Create systemd Service

```bash
# Create service file
sudo nano /etc/systemd/system/pganalytics-api.service

# Add:
[Unit]
Description=pgAnalytics Backend API
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=pganalytics
WorkingDirectory=/opt/pganalytics
EnvironmentFile=/opt/pganalytics/.env
ExecStart=/opt/pganalytics/pganalytics-api
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=pganalytics-api

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/opt/pganalytics/logs

[Install]
WantedBy=multi-user.target

# Create pganalytics user
sudo useradd -r -s /bin/false pganalytics

# Set directory permissions
sudo mkdir -p /opt/pganalytics/logs
sudo chown -R pganalytics:pganalytics /opt/pganalytics

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable pganalytics-api
sudo systemctl start pganalytics-api

# Check status
sudo systemctl status pganalytics-api
```

#### 1e. Configure TLS Certificates

```bash
# Generate self-signed certificate (development)
sudo openssl req -new -x509 -days 365 -nodes \
  -out /etc/pganalytics/backend.crt \
  -keyout /etc/pganalytics/backend.key \
  -subj "/CN=api.example.com"

# For production, use CA-signed certificate
# Copy your certificate and key:
sudo cp your-cert.crt /etc/pganalytics/backend.crt
sudo cp your-key.key /etc/pganalytics/backend.key

# Set permissions
sudo chown root:pganalytics /etc/pganalytics/backend.*
sudo chmod 640 /etc/pganalytics/backend.*

# Update environment variable
sudo systemctl restart pganalytics-api
```

#### 1f. Configure Reverse Proxy (nginx)

```bash
# Install nginx
sudo apt-get install -y nginx

# Create nginx config
sudo nano /etc/nginx/sites-available/pganalytics-api

# Add:
upstream pganalytics_backend {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.example.com;

    ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;
    limit_req zone=api_limit burst=200 nodelay;

    # Proxy to backend
    location / {
        proxy_pass http://pganalytics_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
        proxy_send_timeout 60s;
    }
}

server {
    listen 80;
    listen [::]:80;
    server_name api.example.com;
    return 301 https://$server_name$request_uri;
}

# Enable site
sudo ln -s /etc/nginx/sites-available/pganalytics-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

#### 1g. Verify Backend Installation

```bash
# Check service status
sudo systemctl status pganalytics-api

# Check logs
sudo journalctl -u pganalytics-api -f

# Test API endpoint
curl http://localhost:8080/api/v1/health

# Test through reverse proxy
curl https://api.example.com/api/v1/health -k
```

### 2. RHEL/CentOS Installation

#### 2a. Install Dependencies

```bash
# Update packages
sudo dnf update -y
sudo dnf groupinstall -y "Development Tools"

# Install Go
wget https://go.dev/dl/go1.22.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
source /etc/profile
go version

# Install PostgreSQL client
sudo dnf install -y postgresql16-devel postgresql16

# Clone and build (same as Ubuntu 1b)
```

#### 2b. Create systemd Service (same as Ubuntu 1d)

### 3. Docker Deployment

#### 3a. Build Docker Image

```bash
# Dockerfile (included in repository)
cd backend
docker build -t pganalytics-api:3.2.0 .

# Run container
docker run -d \
  --name pganalytics-api \
  --env DATABASE_URL="postgresql://pganalytics:password@postgres:5432/pganalytics?sslmode=require" \
  --env JWT_SECRET="your-64-character-secret" \
  --env REGISTRATION_SECRET="your-32-character-secret" \
  --env PORT=8080 \
  --env ENVIRONMENT=production \
  -p 8080:8080 \
  pganalytics-api:3.2.0

# Check logs
docker logs -f pganalytics-api
```

#### 3b. Docker Compose (Production)

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: pganalytics-db
    environment:
      POSTGRES_USER: pganalytics
      POSTGRES_PASSWORD: secure-password
      POSTGRES_DB: pganalytics
      POSTGRES_INITDB_ARGS: "-c shared_preload_libraries=timescaledb"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./backend/migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    networks:
      - pganalytics

  api:
    build: ./backend
    container_name: pganalytics-api
    environment:
      DATABASE_URL: "postgresql://pganalytics:secure-password@postgres:5432/pganalytics?sslmode=require"
      JWT_SECRET: "your-64-character-secret"
      REGISTRATION_SECRET: "your-32-character-secret"
      PORT: 8080
      ENVIRONMENT: production
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    networks:
      - pganalytics
    restart: always

  grafana:
    image: grafana/grafana:10.0.0
    container_name: pganalytics-grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_SECURITY_ADMIN_USER: admin
    volumes:
      - grafana-storage:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    ports:
      - "3000:3000"
    depends_on:
      - postgres
    networks:
      - pganalytics
    restart: always

volumes:
  pgdata:
  grafana-storage:

networks:
  pganalytics:
```

### 4. Kubernetes Deployment

#### 4a. Create ConfigMap and Secrets

```bash
# Create namespace
kubectl create namespace pganalytics

# Create secrets
kubectl create secret generic pganalytics-secrets \
  --from-literal=database-url="postgresql://pganalytics:password@postgres:5432/pganalytics?sslmode=require" \
  --from-literal=jwt-secret="your-64-character-secret" \
  --from-literal=registration-secret="your-32-character-secret" \
  -n pganalytics

# Create TLS certificate secret
kubectl create secret tls pganalytics-tls \
  --cert=/path/to/backend.crt \
  --key=/path/to/backend.key \
  -n pganalytics
```

#### 4b. Deploy Backend

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pganalytics-api
  namespace: pganalytics
spec:
  replicas: 2
  selector:
    matchLabels:
      app: pganalytics-api
  template:
    metadata:
      labels:
        app: pganalytics-api
    spec:
      containers:
      - name: api
        image: pganalytics-api:3.2.0
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: pganalytics-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: pganalytics-secrets
              key: jwt-secret
        - name: REGISTRATION_SECRET
          valueFrom:
            secretKeyRef:
              name: pganalytics-secrets
              key: registration-secret
        - name: PORT
          value: "8080"
        - name: ENVIRONMENT
          value: "production"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 2000m
            memory: 2Gi

---
apiVersion: v1
kind: Service
metadata:
  name: pganalytics-api-service
  namespace: pganalytics
spec:
  selector:
    app: pganalytics-api
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  type: LoadBalancer

# Deploy
kubectl apply -f deployment.yaml
```

---

## Collector Installation

### Prerequisites Checklist

- [ ] C++ compiler (g++7+ or clang5+)
- [ ] libpq development library
- [ ] CMake 3.15+
- [ ] Network access to Backend API
- [ ] PostgreSQL instance to monitor

### 1. Build from Source (Ubuntu/Debian)

#### 1a. Install Dependencies

```bash
# Install build tools
sudo apt-get update
sudo apt-get install -y \
  build-essential \
  cmake \
  git \
  libpq-dev \
  pkg-config \
  curl \
  openssl \
  libssl-dev \
  zlib1g-dev

# Clone repository
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3/collector
```

#### 1b. Build Collector

```bash
# Create build directory
mkdir build && cd build

# Configure build
cmake ..

# Build (with optimizations)
cmake --build . --config Release -j$(nproc)

# Install binary
sudo install -D pganalytics /usr/local/bin/pganalytics
sudo chmod +x /usr/local/bin/pganalytics

# Verify
pganalytics --version
```

#### 1c. Configure Collector

```bash
# Create configuration directory
sudo mkdir -p /etc/pganalytics

# Copy example configuration
sudo cp config.toml.example /etc/pganalytics/collector.toml

# Edit configuration
sudo nano /etc/pganalytics/collector.toml

# Minimal configuration example:
[collector]
id = "collector-001"
hostname = "db-server-01"
interval = 60

[backend]
url = "https://api.example.com"
tls_cert = "/etc/pganalytics/collector.crt"
tls_key = "/etc/pganalytics/collector.key"
ca_cert = "/etc/pganalytics/ca.crt"

[postgres]
host = "localhost"
port = 5432
user = "monitoring_user"
password = "secure-password"
databases = ["postgres"]

[metrics]
enable_pg_stat_statements = true
enable_replication = true
enable_wal_level = true

[buffer]
max_size = 10000
flush_interval = 60

[security]
tls_enabled = true
verify_ssl = true
jwt_token = "will-be-set-by-registration"
```

#### 1d. Create Monitoring User on Target PostgreSQL

```bash
# Connect to PostgreSQL instance
psql -U postgres

-- Create monitoring user
CREATE ROLE pganalytics_monitoring WITH LOGIN NOINHERIT PASSWORD 'secure-password';

-- Grant monitoring privileges
GRANT pg_monitor TO pganalytics_monitoring;

-- Grant table access
GRANT USAGE ON SCHEMA public TO pganalytics_monitoring;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO pganalytics_monitoring;

-- Exit
\q
```

#### 1e. Register Collector with Backend

```bash
# Generate JWT token from backend
curl -X POST https://api.example.com/api/v1/collectors/register \
  -H "X-Registration-Secret: $REGISTRATION_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "collector-001",
    "hostname": "db-server-01",
    "address": "192.168.1.100"
  }' \
  -k  # Ignore certificate verification in development

# Response:
# {
#   "collector_id": "550e8400-e29b-41d4-a716-446655440000",
#   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "certificate": "-----BEGIN CERTIFICATE-----...",
#   "private_key": "-----BEGIN PRIVATE KEY-----...",
#   "expires_at": "2027-02-25T00:00:00Z"
# }

# Save certificate and key
echo "CERTIFICATE_HERE" | sudo tee /etc/pganalytics/collector.crt
echo "PRIVATE_KEY_HERE" | sudo tee /etc/pganalytics/collector.key
sudo chmod 600 /etc/pganalytics/collector.key

# Add JWT token to collector.toml
sudo nano /etc/pganalytics/collector.toml
# Update: jwt_token = "TOKEN_HERE"
```

#### 1f. Create systemd Service

```bash
# Create service file
sudo nano /etc/systemd/system/pganalytics-collector.service

# Add:
[Unit]
Description=pgAnalytics Distributed Collector
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=pganalytics
WorkingDirectory=/opt/pganalytics
ExecStart=/usr/local/bin/pganalytics --config /etc/pganalytics/collector.toml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=pganalytics-collector

[Install]
WantedBy=multi-user.target

# Create pganalytics user
sudo useradd -r -s /bin/false pganalytics 2>/dev/null || true

# Set permissions
sudo mkdir -p /opt/pganalytics
sudo chown pganalytics:pganalytics /opt/pganalytics
sudo chown root:pganalytics /etc/pganalytics/*.toml
sudo chmod 640 /etc/pganalytics/*.toml

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable pganalytics-collector
sudo systemctl start pganalytics-collector

# Verify
sudo systemctl status pganalytics-collector
sudo journalctl -u pganalytics-collector -f
```

#### 1g. Test Collector Connection

```bash
# Test dry-run mode
pganalytics --config /etc/pganalytics/collector.toml --dry-run

# Check logs
sudo journalctl -u pganalytics-collector -n 50

# Verify metrics in database
psql -U pganalytics -d pganalytics -h api.example.com

SELECT COUNT(*) FROM metrics WHERE collector_id = 'collector-001';

SELECT * FROM metrics
WHERE collector_id = 'collector-001'
ORDER BY timestamp DESC
LIMIT 10;
```

### 2. RHEL/CentOS Installation

#### 2a. Install Dependencies

```bash
sudo dnf groupinstall -y "Development Tools"
sudo dnf install -y \
  cmake \
  postgresql16-devel \
  openssl-devel \
  zlib-devel \
  curl

# Build from source (same as Ubuntu 1b)
```

### 3. Docker Deployment

```bash
# Build collector image
cd collector
docker build -t pganalytics-collector:3.2.0 .

# Run container
docker run -d \
  --name pganalytics-collector \
  -v /etc/pganalytics:/etc/pganalytics:ro \
  -e BACKEND_URL="https://api.example.com" \
  -e JWT_TOKEN="your-token" \
  pganalytics-collector:3.2.0

# Check logs
docker logs -f pganalytics-collector
```

### 4. AWS EC2 Deployment

#### 4a. Launch EC2 Instance

```bash
# Launch t3.medium instance with:
# - Ubuntu 22.04 LTS AMI
# - Security group allowing port 443 outbound
# - IAM role with CloudWatch Logs access (optional)

aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.medium \
  --key-name your-key \
  --security-group-ids sg-xxxxxxxx \
  --iam-instance-profile Name=pganalytics-role
```

#### 4b. User Data Script

```bash
#!/bin/bash
# Save as user-data.sh

set -e

# Update system
apt-get update
apt-get upgrade -y

# Install dependencies
apt-get install -y \
  build-essential \
  cmake \
  git \
  libpq-dev \
  curl \
  openssl

# Clone and build collector
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3/collector
mkdir build && cd build
cmake ..
cmake --build . --config Release -j$(nproc)
sudo install -D pganalytics /usr/local/bin/pganalytics

# Create configuration (replace placeholders)
mkdir -p /etc/pganalytics
cat > /etc/pganalytics/collector.toml << 'EOF'
[collector]
id = "aws-ec2-001"
hostname = "$(hostname)"
interval = 60

[backend]
url = "https://api.example.com"
tls_cert = "/etc/pganalytics/collector.crt"
tls_key = "/etc/pganalytics/collector.key"
ca_cert = "/etc/pganalytics/ca.crt"

[postgres]
host = "localhost"
port = 5432
user = "pganalytics_monitoring"
password = "REPLACE_WITH_PASSWORD"
databases = ["postgres"]

[metrics]
enable_pg_stat_statements = true
enable_replication = true
EOF

# Register collector (manual step - needs API call)
# curl -X POST https://api.example.com/api/v1/collectors/register...

# Create systemd service
cat > /etc/systemd/system/pganalytics-collector.service << 'EOF'
[Unit]
Description=pgAnalytics Collector
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/pganalytics --config /etc/pganalytics/collector.toml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

systemctl enable pganalytics-collector
systemctl start pganalytics-collector
```

---

## Grafana Installation

### Prerequisites Checklist

- [ ] Standalone server (separate from Backend/DB)
- [ ] 2+ CPU cores
- [ ] 4GB+ RAM
- [ ] Network access to PostgreSQL database
- [ ] Port 3000 accessible from users

### 1. Ubuntu/Debian Installation

#### 1a. Install Grafana

```bash
# Add Grafana repository
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -

# Install Grafana
sudo apt-get update
sudo apt-get install -y grafana-server

# Start and enable service
sudo systemctl start grafana-server
sudo systemctl enable grafana-server

# Verify
sudo systemctl status grafana-server
```

#### 1b. Configure Grafana

```bash
# Edit configuration
sudo nano /etc/grafana/grafana.ini

# Update settings:
[server]
http_addr = 127.0.0.1
http_port = 3000
root_url = https://grafana.example.com/

[security]
admin_user = admin
admin_password = your-secure-password
secret_key = your-secret-key
cookie_secure = true
cookie_samesite = Strict
strict_transport_security = true

[database]
type = postgres
host = postgres.example.com:5432
name = grafana
user = grafana_user
password = grafana_password
ssl_mode = require

[users]
allow_sign_up = false
auto_assign_org_role = Viewer

[auth]
oauth_auto_login = false

# Restart Grafana
sudo systemctl restart grafana-server
```

#### 1c. Create PostgreSQL User for Grafana

```bash
# Connect to PostgreSQL
psql -U pganalytics -d pganalytics

-- Create grafana user
CREATE ROLE grafana_user WITH LOGIN NOINHERIT PASSWORD 'secure-password';

-- Grant read access to metrics
GRANT USAGE ON SCHEMA public TO grafana_user;
GRANT SELECT ON metrics TO grafana_user;

-- Exit
\q
```

#### 1d. Configure PostgreSQL Data Source

```bash
# Access Grafana UI: http://localhost:3000
# Login with admin/admin

# Configuration -> Data Sources -> Add data source

Name: pganalytics-metrics
Type: PostgreSQL
Host: postgres.example.com:5432
Database: pganalytics
User: grafana_user
Password: (secure password)
TLS/SSL Mode: require
Save & Test
```

#### 1e. Import Dashboards

```bash
# Copy dashboard files from repository
cp grafana/dashboards/*.json /var/lib/grafana/dashboards/

# Or import via UI:
# Dashboards -> Import JSON -> Select file -> Load

# Available dashboards:
# 1. Overview.json - System overview
# 2. Performance.json - Query performance
# 3. Replication.json - Replication metrics
# 4. Health.json - Database health
# 5. Connections.json - Connection pooling
# 6. WAL.json - Write-ahead log
# 7. Cache.json - Cache hit ratios
# 8. Transactions.json - Transaction stats
# 9. Storage.json - Storage usage
```

#### 1f. Configure Alerts

```bash
# Create alert channel: Alerting -> Notification channels

Type: Email
Email address: ops@example.com

Type: Slack
Webhook URL: https://hooks.slack.com/...

Type: PagerDuty
Integration Key: xxxxx

# Create alert rules on dashboards
# Click panel -> Alert -> Create Alert
# Condition: If metric > threshold for 5 minutes
# Then: Notify via channel
```

#### 1g. Configure Reverse Proxy (nginx)

```bash
# Create nginx config
sudo nano /etc/nginx/sites-available/grafana

# Add:
upstream grafana {
    server 127.0.0.1:3000;
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name grafana.example.com;

    ssl_certificate /etc/letsencrypt/live/grafana.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/grafana.example.com/privkey.pem;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    location / {
        proxy_pass http://grafana;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
    }
}

server {
    listen 80;
    listen [::]:80;
    server_name grafana.example.com;
    return 301 https://$server_name$request_uri;
}

# Enable
sudo ln -s /etc/nginx/sites-available/grafana /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 2. Docker Installation

```yaml
version: '3.8'

services:
  grafana:
    image: grafana/grafana:10.0.0
    container_name: pganalytics-grafana
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_USER: admin
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_INSTALL_PLUGINS: grafana-piechart-panel
      GF_SERVER_ROOT_URL: https://grafana.example.com/
      GF_USERS_ALLOW_SIGN_UP: "false"
    volumes:
      - grafana-storage:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
      - ./grafana/grafana.ini:/etc/grafana/grafana.ini:ro
    networks:
      - pganalytics
    restart: always

volumes:
  grafana-storage:

networks:
  pganalytics:
```

### 3. Kubernetes Installation

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: pganalytics
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:10.0.0
        ports:
        - containerPort: 3000
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: grafana-secrets
              key: admin-password
        - name: GF_DATABASE_HOST
          value: "postgres.example.com"
        - name: GF_DATABASE_NAME
          value: "pganalytics"
        - name: GF_DATABASE_USER
          value: "grafana_user"
        - name: GF_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: grafana-secrets
              key: db-password
        volumeMounts:
        - name: dashboards
          mountPath: /etc/grafana/provisioning/dashboards
        - name: datasources
          mountPath: /etc/grafana/provisioning/datasources
      volumes:
      - name: dashboards
        configMap:
          name: grafana-dashboards
      - name: datasources
        configMap:
          name: grafana-datasources

---
apiVersion: v1
kind: Service
metadata:
  name: grafana-service
  namespace: pganalytics
spec:
  selector:
    app: grafana
  ports:
  - protocol: TCP
    port: 3000
    targetPort: 3000
  type: LoadBalancer
```

---

## Networking & Security

### Firewall Rules

#### Load Balancer / Reverse Proxy

```bash
# Inbound Rules:
Port 80/tcp    From: 0.0.0.0/0          (HTTP redirect)
Port 443/tcp   From: 0.0.0.0/0          (HTTPS users)

# Outbound Rules:
Port 8080/tcp  To: API Servers          (Internal API)
Port 3000/tcp  To: Grafana Servers      (Internal Grafana)
```

#### API Servers

```bash
# Inbound Rules:
Port 8080/tcp  From: Load Balancer      (API traffic)
Port 22/tcp    From: Admin Networks     (SSH)
Port 5432/tcp  From: Database Network   (PostgreSQL connection - optional if DB is separate)

# Outbound Rules:
Port 5432/tcp  To: Database Servers     (PostgreSQL)
Port 443/tcp   To: 0.0.0.0/0            (TLS connections to collectors)
Port 123/udp   To: 0.0.0.0/0            (NTP)
Port 53/udp    To: 0.0.0.0/0            (DNS)
```

#### Collector Servers

```bash
# Inbound Rules:
Port 22/tcp    From: Admin Networks     (SSH)
Port 5432/tcp  From: Local             (PostgreSQL monitoring)

# Outbound Rules:
Port 443/tcp   To: API Servers          (Metrics push)
Port 5432/tcp  To: PostgreSQL Instance  (Query monitoring)
Port 123/udp   To: 0.0.0.0/0            (NTP)
Port 53/udp    To: 0.0.0.0/0            (DNS)
```

#### Database Servers

```bash
# Inbound Rules:
Port 5432/tcp  From: API Servers        (Metrics storage)
Port 5432/tcp  From: Grafana Servers    (Dashboard queries)
Port 5432/tcp  From: Collectors         (Config sync)
Port 22/tcp    From: Admin Networks     (SSH)

# Outbound Rules:
Port 123/udp   To: 0.0.0.0/0            (NTP)
Port 53/udp    To: 0.0.0.0/0            (DNS)
```

#### Grafana Servers

```bash
# Inbound Rules:
Port 443/tcp   From: Load Balancer      (HTTPS)
Port 80/tcp    From: Load Balancer      (HTTP redirect)
Port 22/tcp    From: Admin Networks     (SSH)

# Outbound Rules:
Port 5432/tcp  To: Database Servers     (PostgreSQL datasource)
Port 123/udp   To: 0.0.0.0/0            (NTP)
Port 53/udp    To: 0.0.0.0/0            (DNS)
```

### VPC/Subnet Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        VPC: 10.0.0.0/16                       │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Public Subnet (10.0.1.0/24)                         │    │
│  │ - Load Balancer / Reverse Proxy                     │    │
│  │ - Internet Gateway                                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Private Subnet A (10.0.10.0/24)                     │    │
│  │ - API Server 1                                      │    │
│  │ - Grafana Server                                    │    │
│  │ - NAT Gateway                                       │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Private Subnet B (10.0.20.0/24)                     │    │
│  │ - API Server 2                                      │    │
│  │ - PostgreSQL (Multi-AZ)                             │    │
│  │ - NAT Gateway                                       │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Private Subnet C (10.0.30.0/24)                     │    │
│  │ - Collectors                                        │    │
│  │ - Database replicas                                 │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                                │
└──────────────────────────────────────────────────────────────┘

Network Flow:
1. User → Load Balancer (0.0.0.0:443)
2. Load Balancer → API Servers (10.0.x.x:8080)
3. API Servers → PostgreSQL (10.0.x.x:5432)
4. Collectors → API Servers (10.0.x.x:8080)
5. Grafana → PostgreSQL (10.0.x.x:5432)
```

### TLS/mTLS Configuration

#### Backend to PostgreSQL

```bash
# Generate PostgreSQL certificates (if not using cloud provider)
openssl req -new -x509 -days 365 -nodes \
  -out postgres.crt -keyout postgres.key -subj "/CN=postgres.example.com"

# Backend connection string
DATABASE_URL="postgresql://user:pass@db.example.com:5432/pganalytics?sslmode=require&sslcert=/path/to/client.crt&sslkey=/path/to/client.key&sslrootcert=/path/to/ca.crt"
```

#### Collector to Backend (mTLS)

```bash
# Collector certificate (generated by backend during registration)
# Stored in: /etc/pganalytics/collector.crt
# Stored in: /etc/pganalytics/collector.key

# Backend configuration
[backend]
url = "https://api.example.com"
tls_cert = "/etc/pganalytics/collector.crt"
tls_key = "/etc/pganalytics/collector.key"
ca_cert = "/etc/pganalytics/ca.crt"
verify_ssl = true
```

#### User to Backend (HTTPS)

```bash
# Reverse proxy handles TLS termination
# Certificate obtained from Let's Encrypt or CA

# Nginx configuration
ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers HIGH:!aNULL:!MD5;
```

---

## Multi-Region/Multi-Collector Setup

### Collector Registration & Authentication

```bash
# Pre-requisites
REGISTRATION_SECRET="your-32-character-secret"
API_URL="https://api.example.com"

# Step 1: Register each collector
for region in us-east us-west eu-west; do
  for i in 1 2; do
    curl -X POST $API_URL/api/v1/collectors/register \
      -H "X-Registration-Secret: $REGISTRATION_SECRET" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "collector-'$region'-'$i'",
        "hostname": "db-'$region'-'$i'.example.com",
        "address": "203.0.113.'$((10+i))'",
        "region": "'$region'"
      }' | jq . > collector-$region-$i.json
  done
done

# Step 2: Extract and deploy credentials
for file in collector-*.json; do
  region=$(echo $file | cut -d- -f2)

  token=$(jq -r '.token' $file)
  cert=$(jq -r '.certificate' $file)
  key=$(jq -r '.private_key' $file)

  echo "Deploying to $region..."
  # Copy to collector servers and configure
done
```

### Token Rotation (1-year validity)

```bash
# Collectors receive 1-year JWT tokens
# Implement automated rotation before expiration

# Cron job on collector to refresh token
0 2 * * * curl -X POST https://api.example.com/api/v1/collectors/refresh-token \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"collector_id":"'$COLLECTOR_ID'"}' > /tmp/new-token.json && \
  update_config_with_new_token.sh
```

### Regional Collector Coordination

```
Region: US-EAST (Primary)
├── Collector 1: db-us-east-1 → API (primary) → DB (primary)
├── Collector 2: db-us-east-2 → API (primary)

Region: US-WEST
├── Collector 3: db-us-west-1 → API (regional) → DB (replica)
├── Collector 4: db-us-west-2 → API (regional)

Region: EU-WEST
├── Collector 5: db-eu-west-1 → API (regional) → DB (replica)
├── Collector 6: db-eu-west-2 → API (regional)

All regions → Primary Grafana (us-east)
```

### Collector Failover Setup

```bash
# Monitoring collector health
curl https://api.example.com/api/v1/collectors | jq '.collectors[] | select(.status != "active")'

# Automated failover script
check_collector_health() {
  collector_id=$1
  last_seen=$(curl -s https://api.example.com/api/v1/collectors/$collector_id | jq '.last_seen')

  if [ $(date +%s) - $(date -d "$last_seen" +%s) -gt 300 ]; then
    # Collector hasn't reported in 5 minutes
    trigger_failover_alert
  fi
}

# Run as cron job
*/5 * * * * for col in collector-*; do check_collector_health $col; done
```

---

## Troubleshooting

### Backend API Issues

#### "connection refused" to PostgreSQL

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql -U pganalytics -d pganalytics -h localhost

# Check if TLS is required
grep "sslmode" /opt/pganalytics/.env

# Verify pg_hba.conf
sudo -u postgres grep "pganalytics" /etc/postgresql/16/main/pg_hba.conf
```

#### API not responding on port 8080

```bash
# Check service status
sudo systemctl status pganalytics-api

# Check if port is in use
sudo netstat -tlnp | grep 8080

# Check logs
sudo journalctl -u pganalytics-api -n 50 -e

# Try starting manually
cd /opt/pganalytics
source .env
./pganalytics-api
```

#### "Invalid JWT token" errors

```bash
# Verify JWT_SECRET is set
echo $JWT_SECRET

# Check secret length (minimum 64 characters)
echo $JWT_SECRET | wc -c

# If empty, set it:
export JWT_SECRET=$(openssl rand -base64 64)
echo $JWT_SECRET | sudo tee /opt/pganalytics/.env

# Restart API
sudo systemctl restart pganalytics-api
```

### Collector Issues

#### Collector fails to start

```bash
# Check config file
pganalytics --config /etc/pganalytics/collector.toml --validate

# Test PostgreSQL connection
psql -U pganalytics_monitoring -d postgres \
  -h localhost -p 5432

# Check certificate validity
openssl x509 -in /etc/pganalytics/collector.crt -text -noout | grep -A2 "Validity"

# Run in foreground to see errors
pganalytics --config /etc/pganalytics/collector.toml --verbose
```

#### Metrics not pushing to backend

```bash
# Check collector logs
sudo journalctl -u pganalytics-collector -n 100 -e

# Verify backend connectivity
curl -v https://api.example.com/api/v1/health

# Test metrics push manually
curl -X POST https://api.example.com/api/v1/metrics/push \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"collector_id":"test","metrics":[]}' -k

# Check network firewall
sudo iptables -L -n | grep 443
```

#### "Permission denied" querying PostgreSQL

```bash
# Verify monitoring user permissions
psql -U postgres -d postgres

SELECT * FROM pg_user WHERE usename = 'pganalytics_monitoring';

GRANT SELECT ON pg_stat_statements TO pganalytics_monitoring;
GRANT SELECT ON pg_stat_replication TO pganalytics_monitoring;

-- For PostgreSQL 10+
GRANT pg_monitor TO pganalytics_monitoring;
```

### Grafana Issues

#### Dashboard showing "No data"

```bash
# Check datasource configuration
# Grafana UI → Configuration → Data Sources → pganalytics-metrics

# Test datasource query
SELECT * FROM metrics LIMIT 1;

# Verify PostgreSQL user has access
psql -U grafana_user -d pganalytics

\dt metrics

SELECT COUNT(*) FROM metrics;
```

#### "Cannot reach PostgreSQL"

```bash
# Check datasource host/port
# Verify from Grafana container/host:
telnet postgres.example.com 5432

psql -h postgres.example.com -U grafana_user -d pganalytics

# Check SSL mode in datasource config
# Should match PostgreSQL SSL configuration
```

### Database Performance Issues

#### Slow queries on metrics table

```bash
-- Check table size
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
FROM pg_tables
WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Check index usage
SELECT * FROM pg_stat_user_indexes;

-- Vacuum and analyze
VACUUM ANALYZE metrics;

-- Check hypertable stats (if using TimescaleDB)
SELECT * FROM timescaledb_information.hypertable;
```

#### Disk space filling up

```bash
-- Check data growth rate
SELECT
  DATE_TRUNC('hour', time) as hour,
  COUNT(*) as record_count,
  pg_size_pretty(SUM(pg_column_size(*))::bigint) as size
FROM metrics
GROUP BY hour
ORDER BY hour DESC
LIMIT 24;

-- Drop old data
SELECT drop_chunks('metrics', INTERVAL '30 days');

-- Check retention policy
SELECT * FROM _timescaledb_config.timescaledb_information.hypertable;
```

---

## Summary

This guide provides production-ready installation procedures for all pgAnalytics v3 components. Follow the architecture choice appropriate for your environment:

- **Single Machine**: Development, testing, small deployments
- **Distributed**: Production, high availability, multi-collector monitoring

All components support multiple deployment methods (bare metal, Docker, Kubernetes) and are compatible with PostgreSQL versions 12-18 (with guidance for older versions).

For additional support, refer to:
- `/docs/ARCHITECTURE.md` - Technical design
- `/docs/REPLICATION_COLLECTOR_GUIDE.md` - Collector details
- `/docs/API_SECURITY_REFERENCE.md` - Security implementation
- `/SECURITY.md` - Security guidelines
- `/DEPLOYMENT_PLAN_v3.2.0.md` - Deployment procedures

---

**Version:** 1.0
**Last Updated:** February 25, 2026
**Status:** Production Ready
