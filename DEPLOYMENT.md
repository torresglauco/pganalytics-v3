# pgAnalytics v3 Deployment Guide

## Prerequisites

### System Requirements

- **OS**: Linux (Ubuntu 20.04+, CentOS 8+), macOS 12+, or Windows with WSL2
- **PostgreSQL**: 14, 15, 16, 17, or 18
- **Go**: 1.26 or higher
- **Node.js**: 20.0 or higher (LTS)
- **C++ Compiler**: GCC 9+ or Clang 10+
- **CMake**: 3.15+

### Required Ports

- **8080**: Backend API
- **3000**: Frontend (development) / 80, 443 (production)
- **5432**: PostgreSQL database
- **5000**: ML Service
- **9090**: Prometheus metrics (optional)

### Disk Space

- **Minimum**: 10GB
- **Recommended**: 50GB+ for production databases
- **With historical data**: 100GB+

## Deployment Options

### Option 1: Docker Compose (Recommended for Development)

#### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+

#### Quick Start
```bash
cd /path/to/pganalytics-v3

# Copy example environment file
cp .env.example .env

# Edit configuration if needed
nano .env

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Verify services are running
docker-compose ps
```

#### Docker Compose Services

```yaml
Services included:
- backend: pgAnalytics API server (port 8080)
- frontend: React web UI (port 3000)
- postgres: PostgreSQL database (port 5432)
- ml-service: FastAPI ML service (port 5000)
- collector: Optional C++ collector (internal)
```

#### Configuration

Edit `docker-compose.yml` for:
- Resource limits
- Volume mounts
- Environment variables
- Network settings

```yaml
backend:
  environment:
    DATABASE_URL: postgres://user:pass@postgres:5432/pganalytics
    LISTEN_PORT: 8080
    LOG_LEVEL: info
  resources:
    limits:
      cpus: '2'
      memory: 2G
```

#### Health Checks

```bash
# Check backend health
curl http://localhost:8080/health

# Check frontend
curl http://localhost:3000

# Check ML service
curl http://localhost:5000/health
```

#### Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (deletes data)
docker-compose down -v
```

### Option 2: Kubernetes (Production)

#### Prerequisites
- Kubernetes cluster 1.24+
- kubectl configured
- Helm 3.0+
- Storage class for persistent volumes

#### Installation

```bash
# Add Helm repository
helm repo add pganalytics https://helm.pganalytics.io
helm repo update

# Create namespace
kubectl create namespace pganalytics

# Install chart with custom values
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values.yaml
```

#### values.yaml Example

```yaml
replicaCount: 3

backend:
  image: pganalytics/backend:latest
  replicas: 3
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "2Gi"
      cpu: "1000m"

frontend:
  image: pganalytics/frontend:latest
  replicas: 2

postgres:
  enabled: true
  storage: 100Gi
  version: "16"

mlService:
  image: pganalytics/ml-service:latest
  replicas: 2

ingress:
  enabled: true
  className: nginx
  hosts:
    - host: pganalytics.example.com
      paths:
        - path: /
          pathType: Prefix

tls:
  enabled: true
  certManager: true
  issuer: letsencrypt-prod
```

#### Deployment Verification

```bash
# Check pods
kubectl get pods -n pganalytics

# Check services
kubectl get svc -n pganalytics

# View logs
kubectl logs -n pganalytics deployment/backend

# Access web UI
kubectl port-forward -n pganalytics svc/frontend 3000:80
# Then visit http://localhost:3000
```

#### Scaling

```bash
# Scale backend replicas
kubectl scale deployment backend --replicas=5 -n pganalytics

# Auto-scaling setup
kubectl autoscale deployment backend \
  --min=2 --max=10 \
  --cpu-percent=70 \
  -n pganalytics
```

### Option 3: On-Premise (Manual Installation)

#### Backend Installation

```bash
# Create application directory
sudo mkdir -p /opt/pganalytics
sudo chown $USER:$USER /opt/pganalytics

# Clone repository
cd /opt/pganalytics
git clone https://github.com/pganalytics/pganalytics-v3.git
cd pganalytics-v3/backend

# Build binary
go build -o pganalytics-api ./cmd/server

# Create systemd service
sudo tee /etc/systemd/system/pganalytics-api.service > /dev/null <<EOF
[Unit]
Description=pgAnalytics API Server
After=network.target postgresql.service

[Service]
Type=simple
User=pganalytics
WorkingDirectory=/opt/pganalytics/pganalytics-v3/backend
ExecStart=/opt/pganalytics/pganalytics-v3/backend/pganalytics-api
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable pganalytics-api
sudo systemctl start pganalytics-api

# Check status
sudo systemctl status pganalytics-api
```

#### Frontend Installation

```bash
cd /opt/pganalytics/pganalytics-v3/frontend

# Build production bundle
npm install
npm run build

# Configure Nginx as reverse proxy
sudo tee /etc/nginx/sites-available/pganalytics > /dev/null <<EOF
server {
    listen 80;
    server_name pganalytics.example.com;

    location / {
        root /opt/pganalytics/pganalytics-v3/frontend/dist;
        try_files \$uri \$uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/pganalytics /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

#### Collector Installation

```bash
# Build collector
cd /opt/pganalytics/pganalytics-v3/collector
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
sudo cp build/pganalytics-collector /usr/local/bin/

# Create systemd service
sudo tee /etc/systemd/system/pganalytics-collector.service > /dev/null <<EOF
[Unit]
Description=pgAnalytics Collector
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/pganalytics-collector
Restart=on-failure
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable pganalytics-collector
sudo systemctl start pganalytics-collector
```

## Configuration

### Backend Configuration

Create or edit `/etc/pganalytics/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  tls:
    enabled: true
    cert_path: "/etc/pganalytics/certs/server.crt"
    key_path: "/etc/pganalytics/certs/server.key"

database:
  url: "postgres://user:password@localhost:5432/pganalytics"
  max_connections: 100
  connection_timeout: 30s
  pool_min_idle: 10
  pool_max_idle: 20

logging:
  level: "info"
  format: "json"
  output: "stdout"

metrics:
  retention_days: 30
  aggregation_interval: 5m

api:
  rate_limit: 1000
  rate_limit_window: 1m
  max_body_size: "10MB"

security:
  jwt_secret: "your-secret-key-change-this"
  cors_origins:
    - "https://pganalytics.example.com"
  require_https: true
```

### Frontend Configuration

Create `.env.production`:

```bash
VITE_API_URL=https://api.pganalytics.example.com
VITE_WS_URL=wss://api.pganalytics.example.com/ws
VITE_LOG_LEVEL=info
```

### Collector Configuration

Create `/etc/pganalytics/collector.conf`:

```yaml
collector:
  server_url: "https://api.pganalytics.example.com"
  agent_id: "agent-001"
  interval: "30s"
  batch_size: 100

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  databases:
    - "postgres"
    - "myapp"

metrics:
  collect_slow_queries: true
  slow_query_threshold: "1s"
  collect_locks: true
  collect_replication: true
```

### ML Service Configuration

Create `.env` for ML service:

```bash
MODEL_PATH=/models/anomaly_detector.pkl
BATCH_SIZE=32
DEVICE=cpu  # or cuda
LOG_LEVEL=info
```

## TLS/SSL Configuration

### Self-Signed Certificates (Development)

```bash
# Generate private key
openssl genrsa -out server.key 2048

# Generate certificate
openssl req -new -x509 -key server.key -out server.crt -days 365

# Copy to config directory
sudo cp server.crt server.key /etc/pganalytics/certs/
sudo chmod 600 /etc/pganalytics/certs/server.key
```

### Let's Encrypt (Production)

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot certonly --nginx -d pganalytics.example.com

# Configure auto-renewal
sudo systemctl enable certbot.timer
```

## Database Setup

### PostgreSQL Installation

```bash
# Ubuntu
sudo apt update
sudo apt install postgresql-16 postgresql-contrib-16

# macOS
brew install postgresql@16

# Start service
sudo systemctl start postgresql
```

### Database Initialization

```bash
# Create database and user
sudo -u postgres psql << EOF
CREATE DATABASE pganalytics;
CREATE USER pganalytics WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE pganalytics TO pganalytics;
EOF

# Run migrations
cd backend
DATABASE_URL="postgres://pganalytics:secure_password@localhost:5432/pganalytics" \
  go run cmd/migrate/main.go up
```

## Monitoring and Logging

### Prometheus Setup

```bash
# Install Prometheus
sudo apt install prometheus

# Configure scrape targets
sudo tee /etc/prometheus/prometheus.yml > /dev/null <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'pganalytics'
    static_configs:
      - targets: ['localhost:8080']
EOF

# Restart Prometheus
sudo systemctl restart prometheus
```

### Grafana Dashboards

```bash
# Install Grafana
sudo apt install grafana-server

# Access at http://localhost:3000
# Default: admin/admin

# Import pgAnalytics dashboards
# Dashboard ID: 12345 (example)
```

### Logging with ELK Stack

```bash
# Filebeat configuration for system logs
sudo tee /etc/filebeat/filebeat.yml > /dev/null <<EOF
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/pganalytics/*.log

output.elasticsearch:
  hosts: ["localhost:9200"]

setup.kibana:
  host: "localhost:5601"
EOF

sudo systemctl restart filebeat
```

## Backup and Recovery

### Database Backup

```bash
# Full backup
pg_dump -U pganalytics pganalytics > backup.sql

# Automated daily backups
0 2 * * * pg_dump -U pganalytics pganalytics | gzip > /backups/pganalytics-$(date +\%Y\%m\%d).sql.gz
```

### Configuration Backup

```bash
# Backup configuration files
tar czf config-backup.tar.gz /etc/pganalytics/

# Store in safe location
cp config-backup.tar.gz /backups/
```

### Recovery Procedure

```bash
# Restore database
psql -U pganalytics pganalytics < backup.sql

# Restore configuration
tar xzf config-backup.tar.gz -C /
```

## Security Hardening

### Network Security

```bash
# Configure firewall (UFW)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8080/tcp  # Internal only
sudo ufw enable

# Restrict backend access
sudo ufw allow from 10.0.0.0/8 to any port 8080
```

### Database Security

```bash
# Create read-only user for collectors
CREATE ROLE collector_role WITH LOGIN PASSWORD 'password';
GRANT CONNECT ON DATABASE pganalytics TO collector_role;
GRANT USAGE ON SCHEMA public TO collector_role;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO collector_role;
```

### API Security

- Enable HTTPS/TLS
- Set strong JWT secrets
- Implement rate limiting
- Add CORS restrictions
- Use API keys for service-to-service communication

## Performance Tuning

### PostgreSQL Tuning

```bash
# In postgresql.conf
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 16MB
maintenance_work_mem = 128MB
```

### Backend Optimization

```yaml
database:
  pool_max_idle: 50
  pool_min_idle: 10

metrics:
  batch_size: 1000
  flush_interval: 10s
```

### Frontend Performance

```bash
# Enable compression in Nginx
gzip on;
gzip_types text/plain text/css application/json;

# Cache static assets
location ~* \.(js|css|png|jpg)$ {
    expires 30d;
}
```

## Troubleshooting Deployment

### Backend Won't Start

```bash
# Check logs
sudo journalctl -u pganalytics-api -n 50 -f

# Verify database connectivity
pg_isready -h localhost -p 5432

# Check port availability
lsof -i :8080
```

### High Memory Usage

```bash
# Monitor process
top -p $(pgrep pganalytics-api)

# Check database connections
ps aux | grep postgres | wc -l
```

### Slow Queries

```bash
# Enable query logging
ALTER DATABASE pganalytics SET log_min_duration_statement = 1000;

# Analyze slow queries
SELECT query, calls, mean_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

## Maintenance

### Regular Tasks

- **Daily**: Check logs, verify backups
- **Weekly**: Review metrics and performance
- **Monthly**: Update dependencies, test disaster recovery
- **Quarterly**: Security audit, capacity planning

### Updating

```bash
# Backend update
cd /opt/pganalytics/pganalytics-v3
git pull origin main
cd backend && go build -o pganalytics-api ./cmd/server
sudo systemctl restart pganalytics-api

# Database migrations
DATABASE_URL="..." go run cmd/migrate/main.go up

# Frontend update
cd frontend && npm install && npm run build
# Restart Nginx
```

### Health Checks

```bash
# API health
curl https://api.pganalytics.example.com/health

# Database connectivity
nc -zv localhost 5432

# Collector status
curl http://localhost:8080/api/v1/agents
```

## Support and Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [GitHub Issues](https://github.com/pganalytics/pganalytics-v3/issues)
