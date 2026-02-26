# pgAnalytics Kubernetes Deployment Guide

**Version**: 3.3.0
**Last Updated**: February 26, 2026
**Status**: Production Ready

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Installation Verification](#installation-verification)
4. [Configuration](#configuration)
5. [Cloud Providers](#cloud-providers)
6. [Troubleshooting](#troubleshooting)
7. [Upgrading](#upgrading)
8. [Advanced Topics](#advanced-topics)

---

## Prerequisites

### Required Tools

- **kubectl** 1.24 or higher
- **Helm** 3.0 or higher
- **Docker** 20.0 or higher (for building custom images)
- **Git** 2.0 or higher

### Kubernetes Cluster Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| Kubernetes Version | 1.24 | 1.27+ |
| Nodes | 3 | 5+ |
| Node CPU per node | 2 CPU | 4 CPU+ |
| Node Memory per node | 4 GB | 8 GB+ |
| Storage Class | Any | SSD for PostgreSQL |
| Network CNI | Any | Calico, Flannel |
| Ingress Controller | Optional | Nginx |

### Installed Components

The Helm chart will deploy the following components:

- **pgAnalytics Backend**: StatefulSet with 3 replicas (configurable)
- **pgAnalytics Collector**: DaemonSet running on each node
- **PostgreSQL**: StatefulSet with persistent volume
- **Redis**: Deployment with persistent volume (session storage)
- **Grafana**: Deployment with persistent volume (dashboards)

### Storage Requirements

| Component | Size | Access Mode | StorageClass |
|-----------|------|-------------|--------------|
| PostgreSQL | 50+ GB | RWO | standard/fast-ssd |
| Redis | 5-20 GB | RWO | standard/fast-ssd |
| Grafana | 10-50 GB | RWO | standard |

---

## Quick Start

### 1. Add Helm Repository

```bash
# Add pgAnalytics Helm repository
helm repo add pganalytics https://charts.pganalytics.io
helm repo update
```

### 2. Create Namespace

```bash
# Create dedicated namespace for pgAnalytics
kubectl create namespace pganalytics
```

### 3. Install Helm Chart

```bash
# Development environment
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  --values values-dev.yaml

# Production environment
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  --values values-prod.yaml

# Enterprise environment
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  --values values-enterprise.yaml
```

### 4. Wait for Deployment

```bash
# Watch pods come up
kubectl get pods -n pganalytics -w

# Wait for all pods to be Ready
kubectl wait --for=condition=Ready pod \
  -l app.kubernetes.io/name=pganalytics \
  -n pganalytics \
  --timeout=300s
```

### 5. Verify Installation

```bash
# Check all resources
kubectl get all -n pganalytics

# Check ingress (if enabled)
kubectl get ingress -n pganalytics

# Verify database is ready
kubectl exec -it postgresql-0 -n pganalytics -- \
  psql -U postgres -d pganalytics -c "SELECT 1"

# Check backend API health
kubectl exec -it pganalytics-backend-0 -n pganalytics -- \
  curl http://localhost:8080/api/v1/health
```

---

## Installation Verification

### Step 1: Check Pod Status

```bash
# All pods should be Running and Ready
kubectl get pods -n pganalytics -o wide

# Expected output:
# NAME                             READY   STATUS    RESTARTS   AGE
# pganalytics-backend-0            1/1     Running   0          2m
# pganalytics-backend-1            1/1     Running   0          2m
# pganalytics-backend-2            1/1     Running   0          2m
# pganalytics-collector-abcd1      1/1     Running   0          1m
# pganalytics-collector-efgh2      1/1     Running   0          1m
# postgresql-0                     1/1     Running   0          3m
# redis-0                          1/1     Running   0          2m
# pganalytics-grafana-0            1/1     Running   0          2m
```

### Step 2: Verify Services

```bash
# Check service endpoints
kubectl get svc -n pganalytics

# Expected services:
# NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)
# pganalytics-backend        ClusterIP   10.0.1.100      <none>        8080/TCP
# postgresql                 ClusterIP   10.0.1.101      <none>        5432/TCP
# redis                      ClusterIP   10.0.1.102      <none>        6379/TCP
# pganalytics-grafana        ClusterIP   10.0.1.103      <none>        3000/TCP
```

### Step 3: Test Database Connectivity

```bash
# Get PostgreSQL pod
POSTGRES_POD=$(kubectl get pod -n pganalytics -l app.kubernetes.io/component=postgresql -o jsonpath='{.items[0].metadata.name}')

# Connect to database
kubectl exec -it $POSTGRES_POD -n pganalytics -- psql -U postgres -d pganalytics

# In psql:
# pganalytics=# \dt  -- List tables
# pganalytics=# SELECT COUNT(*) FROM query_stats;  -- Count metrics
# pganalytics=# \q  -- Exit
```

### Step 4: Test Backend API

```bash
# Port-forward backend service
kubectl port-forward -n pganalytics svc/pganalytics-backend 8080:8080 &

# Test health endpoint
curl -v http://localhost:8080/api/v1/health

# Test API authentication
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Test metrics endpoint
curl http://localhost:8080/api/v1/metrics?limit=10 \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

### Step 5: Check Grafana

```bash
# Port-forward Grafana service
kubectl port-forward -n pganalytics svc/pganalytics-grafana 3000:3000 &

# Access http://localhost:3000 in browser
# Default credentials: admin / (from secret)

# Get password from secret
kubectl get secret -n pganalytics pganalytics-secrets \
  -o jsonpath='{.data.grafana-password}' | base64 -d && echo
```

---

## Configuration

### Environment-Specific Values

#### Development Deployment

```bash
# Uses minimal resources (1 replica, 256Mi memory)
# Good for testing and development
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-dev.yaml
```

#### Production Deployment

```bash
# 3 replicas, auto-scaling enabled, 512Mi memory
# Suitable for staging and small production deployments
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-prod.yaml \
  --set secrets.jwtSecret="$(openssl rand -base64 32)" \
  --set secrets.dbPassword="$(openssl rand -base64 32)"
```

#### Enterprise Deployment

```bash
# 5+ replicas, 1GB+ memory, advanced features
# For large-scale enterprise deployments
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-enterprise.yaml \
  --set global.domain="pganalytics.example.com"
```

### Key Configuration Options

#### Backend Configuration

```yaml
backend:
  replicaCount: 3                    # Number of backend replicas
  image:
    tag: "3.3.0"                     # Docker image tag
  resources:
    requests:
      cpu: 250m                      # CPU request per pod
      memory: 256Mi                  # Memory request per pod
    limits:
      cpu: 1000m                     # CPU limit per pod
      memory: 512Mi                  # Memory limit per pod
  autoscaling:
    enabled: true                    # Enable HPA
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
```

#### Database Configuration

```yaml
postgresql:
  persistence:
    enabled: true
    size: 50Gi                       # Persistent volume size
    storageClass: standard           # Storage class name
  resources:
    requests:
      memory: 1Gi
    limits:
      memory: 2Gi
  env:
    - name: POSTGRES_PASSWORD
      valueFrom:
        secretKeyRef:
          name: pganalytics-secrets
          key: postgres-root-password
```

#### Collector Configuration

```yaml
collector:
  mode: daemonset                    # Run on every node
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  env:
    - name: COLLECTOR_INTERVAL
      value: "60"                    # Collect every 60 seconds
    - name: METRICS_BATCH_SIZE
      value: "100"                   # Max queries per batch
```

### Custom Values Override

```bash
# Override individual values
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  --set backend.replicaCount=5 \
  --set postgresql.persistence.size=100Gi \
  --set features.advancedAuth=true \
  --set features.anomalyDetection=true
```

---

## Cloud Providers

### AWS EKS (Elastic Kubernetes Service)

#### Prerequisites

```bash
# Install AWS CLI
aws configure

# Create EKS cluster
aws eks create-cluster \
  --name pganalytics-prod \
  --version 1.27 \
  --roleArn arn:aws:iam::ACCOUNT_ID:role/EKSServiceRole \
  --resourcesVpcConfig subnetIds=subnet-xxx,subnet-yyy,subnet-zzz

# Update kubeconfig
aws eks update-kubeconfig --name pganalytics-prod --region us-east-1

# Verify connection
kubectl cluster-info
```

#### Deploy to EKS

```bash
# Install AWS Load Balancer Controller (optional)
helm repo add eks https://aws.github.io/eks-charts
helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system

# Deploy pgAnalytics
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-prod.yaml \
  --set global.domain="pganalytics.example.com" \
  --set backend.service.type="LoadBalancer"

# Get LoadBalancer DNS
kubectl get svc -n pganalytics pganalytics-backend -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
```

#### Storage Classes

```yaml
# Use AWS EBS for persistent storage
backend:
  persistence:
    storageClass: gp2              # General Purpose SSD
    size: 100Gi

postgresql:
  persistence:
    storageClass: gp3              # Latest EBS type
    size: 200Gi
```

### GCP GKE (Google Kubernetes Engine)

#### Prerequisites

```bash
# Install gcloud CLI
curl https://sdk.cloud.google.com | bash

# Create GKE cluster
gcloud container clusters create pganalytics-prod \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n1-standard-2 \
  --enable-autorepair \
  --enable-autoupgrade

# Get credentials
gcloud container clusters get-credentials pganalytics-prod --zone us-central1-a

# Verify connection
kubectl cluster-info
```

#### Deploy to GKE

```bash
# Deploy pgAnalytics
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-prod.yaml \
  --set global.domain="pganalytics.example.com"

# Get external IP
kubectl get ingress -n pganalytics
```

#### Storage Classes

```yaml
# Use GCP Persistent Disk
postgresql:
  persistence:
    storageClass: standard         # Standard Persistent Disk
    size: 100Gi

redis:
  persistence:
    storageClass: standard-rwo
    size: 20Gi
```

### Azure AKS (Azure Kubernetes Service)

#### Prerequisites

```bash
# Install Azure CLI
curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# Create AKS cluster
az aks create \
  --resource-group myResourceGroup \
  --name pganalytics-prod \
  --node-count 3 \
  --vm-set-type VirtualMachineScaleSets \
  --load-balancer-sku standard

# Get credentials
az aks get-credentials \
  --resource-group myResourceGroup \
  --name pganalytics-prod

# Verify connection
kubectl cluster-info
```

#### Deploy to AKS

```bash
# Install Nginx Ingress Controller (optional)
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace

# Deploy pgAnalytics
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-prod.yaml
```

#### Storage Classes

```yaml
# Use Azure Managed Disk
postgresql:
  persistence:
    storageClass: managed-premium  # Premium SSD
    size: 100Gi

redis:
  persistence:
    storageClass: managed-standard
    size: 20Gi
```

---

## Troubleshooting

### Common Issues

#### 1. Pods Stuck in Pending

```bash
# Check pod events
kubectl describe pod <pod-name> -n pganalytics

# Check storage availability
kubectl get pvc -n pganalytics
kubectl describe pvc -n pganalytics

# Check node resources
kubectl top nodes
kubectl describe nodes

# Solution: May need to add nodes or adjust resource requests
kubectl scale nodes --replicas=5  # Scale node group
```

#### 2. Database Connection Errors

```bash
# Check PostgreSQL is running
kubectl get statefulset -n pganalytics postgresql
kubectl logs -n pganalytics postgresql-0

# Test connection from backend pod
kubectl exec -it pganalytics-backend-0 -n pganalytics -- \
  psql -h postgresql -U pganalytics -d pganalytics -c "SELECT 1"

# Check database password
kubectl get secret pganalytics-secrets -n pganalytics \
  -o jsonpath='{.data.db-password}' | base64 -d
```

#### 3. High Memory Usage

```bash
# Check memory usage
kubectl top pods -n pganalytics

# Check process details
kubectl exec -it pganalytics-backend-0 -n pganalytics -- \
  top -b -n 1 | head -20

# Check PostgreSQL queries
kubectl exec -it postgresql-0 -n pganalytics -- \
  psql -U postgres -d pganalytics -c "SELECT * FROM pg_stat_statements LIMIT 10"

# Solution: Increase memory limits or enable query optimization
```

#### 4. Ingress Not Working

```bash
# Check ingress status
kubectl get ingress -n pganalytics
kubectl describe ingress -n pganalytics

# Check ingress controller
kubectl get deployment -n ingress-nginx

# Check DNS resolution
nslookup api.pganalytics.local

# Check certificate
kubectl get certificate -n pganalytics
kubectl describe certificate -n pganalytics
```

#### 5. Collector Not Connecting

```bash
# Check collector logs
kubectl logs -n pganalytics -l app.kubernetes.io/component=collector -f

# Test connectivity from collector pod
kubectl exec <collector-pod> -n pganalytics -- \
  curl -v http://pganalytics-backend:8080/api/v1/health

# Check environment variables
kubectl exec <collector-pod> -n pganalytics -- env | grep API_
```

### Debug Commands

```bash
# Get all resources
kubectl get all -n pganalytics

# Describe pod for events
kubectl describe pod <pod-name> -n pganalytics

# View logs
kubectl logs -n pganalytics <pod-name>
kubectl logs -n pganalytics <pod-name> --previous  # Previous crash

# Port-forward for debugging
kubectl port-forward -n pganalytics pod/<pod-name> 8080:8080

# Execute command in pod
kubectl exec -it <pod-name> -n pganalytics -- /bin/sh

# Check resource usage
kubectl top pods -n pganalytics
kubectl top nodes

# Get secret values
kubectl get secret <secret-name> -n pganalytics -o jsonpath='{.data}'

# Check events
kubectl get events -n pganalytics --sort-by='.lastTimestamp'
```

---

## Upgrading

### Before Upgrade

```bash
# Backup database
kubectl exec -it postgresql-0 -n pganalytics -- \
  pg_dump -U postgres pganalytics > pganalytics-backup.sql

# Check current version
helm list -n pganalytics
```

### Upgrade Steps

```bash
# Update Helm repository
helm repo update

# Dry run (preview changes)
helm upgrade pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-prod.yaml \
  --dry-run

# Perform upgrade
helm upgrade pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-prod.yaml

# Watch rollout
kubectl rollout status -n pganalytics statefulset/pganalytics-backend

# Verify upgrade
kubectl get pods -n pganalytics
helm list -n pganalytics
```

### Rollback

```bash
# List release history
helm history pganalytics -n pganalytics

# Rollback to previous version
helm rollback pganalytics 1 -n pganalytics

# Verify rollback
kubectl get pods -n pganalytics
```

---

## Advanced Topics

### High Availability

```yaml
# Enable pod disruption budget
podDisruptionBudget:
  enabled: true
  minAvailable: 2

# Configure pod anti-affinity
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/name
              operator: In
              values:
                - pganalytics
        topologyKey: kubernetes.io/hostname
```

### Security

```yaml
# Enable network policies
networkPolicy:
  enabled: true

# Configure RBAC
rbac:
  create: true
  serviceAccount:
    create: true

# Set security context
securityContext:
  runAsNonRoot: true
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

### Monitoring

```bash
# Install Prometheus operator
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace

# Configure ServiceMonitor
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: pganalytics
  namespace: pganalytics
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: pganalytics
  endpoints:
    - port: metrics
      interval: 30s
EOF
```

### Multi-Cloud Deployment

```bash
# Deploy to multiple clusters
for cluster in us-east eu-west ap-southeast; do
  kubectl --context=$cluster apply -f helm/pganalytics/templates/
done
```

---

## Support

For issues and questions:

- **Documentation**: https://pganalytics.io/docs
- **GitHub Issues**: https://github.com/pganalytics/pganalytics-v3/issues
- **Community Chat**: https://slack.pganalytics.io
- **Enterprise Support**: support@pganalytics.io

---

**Last Updated**: February 26, 2026
**Version**: 3.3.0
**Status**: Production Ready
