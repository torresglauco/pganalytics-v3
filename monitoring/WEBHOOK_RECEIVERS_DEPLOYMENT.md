# Webhook Receivers Deployment Guide - pgAnalytics v3 Phase 5

**Date**: March 3, 2026
**Phase**: 5 - Alerting & Automation (Week 2)
**Status**: Ready for Deployment

---

## Overview

This guide covers deployment of two Flask webhook receiver applications:

1. **Incident Tracking Receiver** - Receives Grafana alerts and creates incidents
2. **JIRA Auto-Ticket Receiver** - Receives Grafana alerts and creates JIRA tickets

Both applications are production-ready with error handling, logging, and configuration management.

---

## Prerequisites

### System Requirements
- Python 3.8+
- pip (Python package manager)
- curl (for testing)
- systemd (for service management) or Docker

### Python Dependencies
```bash
pip install flask requests
```

### Environment Secrets
- **Incident Tracking**: API URL and Bearer token
- **JIRA**: Instance URL, user email, and API token
- **Slack**: (for test notifications, configured in Grafana)

---

## Deployment Method 1: Systemd Services (Linux/macOS)

### Step 1: Install Python Dependencies

```bash
# Create virtual environment
python3 -m venv /opt/pganalytics/venv
source /opt/pganalytics/venv/bin/activate

# Install dependencies
pip install flask requests
```

### Step 2: Setup Application Files

```bash
# Create application directory
sudo mkdir -p /opt/pganalytics/webhook-receivers

# Copy Python scripts
sudo cp monitoring/webhook_incident_receiver.py /opt/pganalytics/webhook-receivers/
sudo cp monitoring/webhook_jira_receiver.py /opt/pganalytics/webhook-receivers/

# Set permissions
sudo chmod +x /opt/pganalytics/webhook-receivers/*.py
sudo chown -R pganalytics:pganalytics /opt/pganalytics/webhook-receivers
```

### Step 3: Create Environment File

Create `/etc/pganalytics/webhook-receivers.env`:

```bash
# Incident Tracking Configuration
FLASK_HOST=127.0.0.1
FLASK_PORT=5000
INCIDENT_TRACKING_URL=https://incident-api.internal/api/incidents
INCIDENT_TRACKING_TOKEN=Bearer_token_here
LOG_LEVEL=INFO

# JIRA Configuration (in separate environment for JIRA receiver)
FLASK_HOST=127.0.0.1
FLASK_PORT=5001
JIRA_URL=https://jira.company.com
JIRA_PROJECT=DB
JIRA_USER=bot-email@company.com
JIRA_API_TOKEN=jira_api_token_here
JIRA_ISSUE_TYPE=Task
LOG_LEVEL=INFO
```

### Step 4: Create Systemd Services

**File: `/etc/systemd/system/pganalytics-incident-webhook.service`**

```ini
[Unit]
Description=pgAnalytics Incident Webhook Receiver
After=network.target
Requires=network.target

[Service]
Type=simple
User=pganalytics
Group=pganalytics
WorkingDirectory=/opt/pganalytics/webhook-receivers
Environment="PATH=/opt/pganalytics/venv/bin"
EnvironmentFile=/etc/pganalytics/webhook-receivers.env
ExecStart=/opt/pganalytics/venv/bin/python webhook_incident_receiver.py
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**File: `/etc/systemd/system/pganalytics-jira-webhook.service`**

```ini
[Unit]
Description=pgAnalytics JIRA Webhook Receiver
After=network.target
Requires=network.target

[Service]
Type=simple
User=pganalytics
Group=pganalytics
WorkingDirectory=/opt/pganalytics/webhook-receivers
Environment="PATH=/opt/pganalytics/venv/bin"
EnvironmentFile=/etc/pganalytics/webhook-receivers.env
ExecStart=/opt/pganalytics/venv/bin/python webhook_jira_receiver.py
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

### Step 5: Enable and Start Services

```bash
# Reload systemd daemon
sudo systemctl daemon-reload

# Enable services to start on boot
sudo systemctl enable pganalytics-incident-webhook.service
sudo systemctl enable pganalytics-jira-webhook.service

# Start services
sudo systemctl start pganalytics-incident-webhook.service
sudo systemctl start pganalytics-jira-webhook.service

# Verify services are running
sudo systemctl status pganalytics-incident-webhook.service
sudo systemctl status pganalytics-jira-webhook.service

# View logs
sudo journalctl -u pganalytics-incident-webhook.service -f
sudo journalctl -u pganalytics-jira-webhook.service -f
```

---

## Deployment Method 2: Docker Containers

### Step 1: Create Dockerfile

**File: `monitoring/Dockerfile.webhooks`**

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install dependencies
RUN pip install --no-cache-dir flask requests

# Copy receivers
COPY webhook_incident_receiver.py .
COPY webhook_jira_receiver.py .

# Create non-root user
RUN useradd -m pganalytics

USER pganalytics

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import requests; requests.get('http://localhost:${FLASK_PORT:-5000}/webhook/health')"

# Default to incident receiver, override with jira receiver if needed
ENTRYPOINT ["python"]
CMD ["webhook_incident_receiver.py"]
```

### Step 2: Build and Run Containers

```bash
# Build image
docker build -f monitoring/Dockerfile.webhooks -t pganalytics-webhooks:latest .

# Run incident receiver
docker run -d \
  --name pganalytics-incident-webhook \
  -p 5000:5000 \
  -e INCIDENT_TRACKING_URL=https://incident-api.internal/api/incidents \
  -e INCIDENT_TRACKING_TOKEN=bearer_token_here \
  -e LOG_LEVEL=INFO \
  pganalytics-webhooks:latest \
  python webhook_incident_receiver.py

# Run JIRA receiver
docker run -d \
  --name pganalytics-jira-webhook \
  -p 5001:5001 \
  -e JIRA_URL=https://jira.company.com \
  -e JIRA_PROJECT=DB \
  -e JIRA_USER=bot-email@company.com \
  -e JIRA_API_TOKEN=token_here \
  -e LOG_LEVEL=INFO \
  pganalytics-webhooks:latest \
  python webhook_jira_receiver.py

# Verify containers are running
docker ps | grep pganalytics
```

### Step 3: Docker Compose (Optional)

**File: `docker-compose.webhook-receivers.yml`**

```yaml
version: '3.8'

services:
  incident-webhook:
    build:
      context: .
      dockerfile: monitoring/Dockerfile.webhooks
    container_name: pganalytics-incident-webhook
    ports:
      - "5000:5000"
    environment:
      FLASK_HOST: 0.0.0.0
      FLASK_PORT: 5000
      INCIDENT_TRACKING_URL: ${INCIDENT_TRACKING_URL}
      INCIDENT_TRACKING_TOKEN: ${INCIDENT_TRACKING_TOKEN}
      LOG_LEVEL: INFO
    command: python webhook_incident_receiver.py
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5000/webhook/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  jira-webhook:
    build:
      context: .
      dockerfile: monitoring/Dockerfile.webhooks
    container_name: pganalytics-jira-webhook
    ports:
      - "5001:5001"
    environment:
      FLASK_HOST: 0.0.0.0
      FLASK_PORT: 5001
      JIRA_URL: ${JIRA_URL}
      JIRA_PROJECT: ${JIRA_PROJECT}
      JIRA_USER: ${JIRA_USER}
      JIRA_API_TOKEN: ${JIRA_API_TOKEN}
      LOG_LEVEL: INFO
    command: python webhook_jira_receiver.py
    restart: always
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5001/webhook/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

Deploy with Docker Compose:

```bash
# Create .env file with secrets
cat > .env << EOF
INCIDENT_TRACKING_URL=https://incident-api.internal/api/incidents
INCIDENT_TRACKING_TOKEN=bearer_token_here
JIRA_URL=https://jira.company.com
JIRA_PROJECT=DB
JIRA_USER=bot-email@company.com
JIRA_API_TOKEN=token_here
EOF

# Start services
docker-compose -f docker-compose.webhook-receivers.yml up -d

# View logs
docker-compose -f docker-compose.webhook-receivers.yml logs -f
```

---

## Deployment Method 3: Kubernetes

### Step 1: Create ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pganalytics-webhook-config
  namespace: pganalytics
data:
  LOG_LEVEL: INFO
  FLASK_HOST: 0.0.0.0
```

### Step 2: Create Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pganalytics-webhook-secrets
  namespace: pganalytics
type: Opaque
stringData:
  INCIDENT_TRACKING_TOKEN: "bearer_token_here"
  INCIDENT_TRACKING_URL: "https://incident-api.internal/api/incidents"
  JIRA_API_TOKEN: "jira_token_here"
  JIRA_USER: "bot-email@company.com"
```

### Step 3: Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pganalytics-incident-webhook
  namespace: pganalytics
spec:
  replicas: 2
  selector:
    matchLabels:
      app: pganalytics-incident-webhook
  template:
    metadata:
      labels:
        app: pganalytics-incident-webhook
    spec:
      containers:
      - name: incident-webhook
        image: pganalytics-webhooks:latest
        ports:
        - containerPort: 5000
        envFrom:
        - configMapRef:
            name: pganalytics-webhook-config
        env:
        - name: INCIDENT_TRACKING_TOKEN
          valueFrom:
            secretKeyRef:
              name: pganalytics-webhook-secrets
              key: INCIDENT_TRACKING_TOKEN
        - name: INCIDENT_TRACKING_URL
          valueFrom:
            secretKeyRef:
              name: pganalytics-webhook-secrets
              key: INCIDENT_TRACKING_URL
        livenessProbe:
          httpGet:
            path: /webhook/health
            port: 5000
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /webhook/health
            port: 5000
          initialDelaySeconds: 5
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: pganalytics-incident-webhook
  namespace: pganalytics
spec:
  selector:
    app: pganalytics-incident-webhook
  ports:
  - protocol: TCP
    port: 5000
    targetPort: 5000
  type: ClusterIP
```

---

## Verification & Testing

### Health Check Endpoints

Both receivers provide health check endpoints:

```bash
# Incident Webhook Health
curl -s http://localhost:5000/webhook/health | jq .

# JIRA Webhook Health
curl -s http://localhost:5001/webhook/health | jq .
```

Expected responses:

```json
{
  "status": "healthy",
  "service": "pgAnalytics Webhook Receiver",
  "timestamp": "2026-03-03T12:00:00Z",
  "incidents_cached": 5
}
```

### Configuration Endpoints (Debugging)

```bash
# Check Incident Receiver configuration
curl -s http://localhost:5000/webhook/config | jq .

# Check JIRA Receiver configuration
curl -s http://localhost:5001/webhook/config | jq .

# View cached incidents
curl -s http://localhost:5000/webhook/metrics | jq .
```

### Test Alert Webhooks

Use the provided test script:

```bash
# Set environment variables for testing
export INCIDENT_TRACKING_URL=http://localhost:5000/webhook/incident
export JIRA_WEBHOOK_URL=http://localhost:5001/webhook/jira

# Run tests
chmod +x monitoring/test_notification_channels.sh
./monitoring/test_notification_channels.sh incident
./monitoring/test_notification_channels.sh jira
```

---

## Network Configuration

### Firewall Rules (Linux)

Allow incoming connections to webhook receivers:

```bash
# UFW (Ubuntu)
sudo ufw allow 5000/tcp  # Incident webhook
sudo ufw allow 5001/tcp  # JIRA webhook

# firewalld (RHEL/CentOS)
sudo firewall-cmd --permanent --add-port=5000/tcp
sudo firewall-cmd --permanent --add-port=5001/tcp
sudo firewall-cmd --reload
```

### Reverse Proxy Configuration (Nginx)

```nginx
upstream incident_webhook {
    server 127.0.0.1:5000;
}

upstream jira_webhook {
    server 127.0.0.1:5001;
}

server {
    listen 443 ssl http2;
    server_name webhooks.internal;

    # SSL configuration
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;

    # Incident webhook
    location /webhook/incident {
        proxy_pass http://incident_webhook;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 10s;
    }

    # JIRA webhook
    location /webhook/jira {
        proxy_pass http://jira_webhook;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 10s;
    }
}
```

---

## Troubleshooting

### Service Not Starting

```bash
# Check service status
sudo systemctl status pganalytics-incident-webhook.service

# View detailed logs
sudo journalctl -u pganalytics-incident-webhook.service -n 50 -p err

# Test Python script directly
source /opt/pganalytics/venv/bin/activate
python /opt/pganalytics/webhook-receivers/webhook_incident_receiver.py
```

### Connection Refused

```bash
# Verify port is listening
sudo netstat -tlnp | grep 5000
lsof -i :5000

# Check firewall rules
sudo ufw status
sudo firewall-cmd --list-all
```

### Authentication Failures

```bash
# Verify environment variables are set
sudo systemctl show-environment pganalytics-incident-webhook.service | grep INCIDENT

# Check bearer token format
echo $INCIDENT_TRACKING_TOKEN
# Should be: Bearer actual_token_here
```

### No Incidents Created

1. Verify webhook is receiving requests
2. Check logs for parsing errors
3. Verify external API is accessible
4. Test with curl manually

```bash
curl -v -X POST http://localhost:5000/webhook/incident \
  -H 'Content-Type: application/json' \
  -d '{"title":"Test","severity":"warning","database":"test"}'
```

---

## Monitoring & Metrics

### Log Monitoring

```bash
# Real-time log monitoring
sudo journalctl -u pganalytics-incident-webhook.service -f

# Extract metrics
sudo journalctl -u pganalytics-incident-webhook.service | grep "Incident created"

# Count by severity
sudo journalctl -u pganalytics-incident-webhook.service | grep -o "severity.*" | sort | uniq -c
```

### Prometheus Metrics (Future)

Can be extended with `prometheus-client` for metrics export:

```python
from prometheus_client import Counter, Histogram

incident_created = Counter('incidents_created_total', 'Total incidents created')
webhook_latency = Histogram('webhook_latency_seconds', 'Webhook processing latency')
```

---

## Security Best Practices

1. **Secrets Management**
   - Use environment files with restricted permissions (600)
   - Consider using Kubernetes Secrets or HashiCorp Vault
   - Rotate API tokens regularly

2. **Network Security**
   - Deploy behind reverse proxy (nginx, HAProxy)
   - Use HTTPS for all connections
   - Restrict source IPs for webhook endpoints

3. **Access Control**
   - Run services as non-root user
   - Implement rate limiting
   - Add request validation

4. **Monitoring**
   - Monitor error rates
   - Alert on service failures
   - Log all incidents created

---

## Maintenance

### Updating Receivers

```bash
# Stop services
sudo systemctl stop pganalytics-incident-webhook.service
sudo systemctl stop pganalytics-jira-webhook.service

# Backup current version
sudo cp /opt/pganalytics/webhook-receivers/webhook_*.py \
   /opt/pganalytics/webhook-receivers/backup/

# Copy new version
sudo cp monitoring/webhook_*.py /opt/pganalytics/webhook-receivers/

# Restart services
sudo systemctl start pganalytics-incident-webhook.service
sudo systemctl start pganalytics-jira-webhook.service

# Verify
sudo systemctl status pganalytics-incident-webhook.service
```

### Log Rotation

Configure logrotate for systemd journal:

```bash
# Edit /etc/systemd/journald.conf
MaxRetentionSec=30day
SystemMaxUse=1G
```

---

## Summary

| Aspect | Details |
|--------|---------|
| **Deployment Methods** | Systemd, Docker, Kubernetes |
| **Python Version** | 3.8+ |
| **Dependencies** | Flask, requests |
| **Incident Port** | 5000 (default) |
| **JIRA Port** | 5001 (default) |
| **Health Check** | GET /webhook/health |
| **Logging** | Systemd journal or container logs |
| **Restart Policy** | Automatic (systemd/k8s) |

---

Generated: March 3, 2026
Status: Ready for Week 2 Notification Setup
