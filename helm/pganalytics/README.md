# pgAnalytics Helm Chart

[![Kubernetes Version](https://img.shields.io/badge/Kubernetes-1.24%2B-blue)](https://kubernetes.io)
[![Helm Version](https://img.shields.io/badge/Helm-3.0%2B-blue)](https://helm.sh)
[![Chart Version](https://img.shields.io/badge/Chart%20Version-3.3.0-green)](https://github.com/pganalytics/pganalytics-v3)
[![Application Version](https://img.shields.io/badge/Application%20Version-3.3.0-green)](https://github.com/pganalytics/pganalytics-v3)

Production-ready Helm chart for deploying pgAnalytics on Kubernetes.

## Overview

pgAnalytics is an enterprise-grade PostgreSQL monitoring and performance analysis platform. This Helm chart provides a complete, production-ready deployment of pgAnalytics on Kubernetes clusters.

### Features

- ✅ **Multi-environment support**: Development, Staging, Production, Enterprise
- ✅ **High Availability**: StatefulSets with replicas, pod disruption budgets, anti-affinity
- ✅ **Auto-scaling**: Horizontal Pod Autoscaler (HPA) for backend
- ✅ **Persistent Storage**: PostgreSQL, Redis, and Grafana data persistence
- ✅ **Networking**: Ingress support, NetworkPolicy, service discovery
- ✅ **Security**: RBAC, secrets management, pod security contexts
- ✅ **Cloud-native**: Works on AWS EKS, GCP GKE, Azure AKS, and on-premises
- ✅ **Monitoring**: Prometheus metrics, health checks, logging
- ✅ **Enterprise Features**: LDAP/SAML/OAuth, encryption, audit logging

## Quick Start

### 1. Prerequisites

```bash
# Check Kubernetes version
kubectl version --short

# Check Helm version
helm version

# Required: kubectl 1.24+, Helm 3.0+
```

### 2. Add Helm Repository

```bash
helm repo add pganalytics https://charts.pganalytics.io
helm repo update
```

### 3. Install Chart

```bash
# Create namespace
kubectl create namespace pganalytics

# Install with default values (development)
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics

# Or install with production values
helm install pganalytics pganalytics/pganalytics \
  --namespace pganalytics \
  -f values-prod.yaml
```

### 4. Verify Installation

```bash
# Watch pods coming up
kubectl get pods -n pganalytics -w

# Check all resources
kubectl get all -n pganalytics

# Get ingress details
kubectl get ingress -n pganalytics
```

### 5. Access pgAnalytics

```bash
# Backend API
kubectl port-forward -n pganalytics svc/pganalytics-backend 8080:8080

# Grafana (http://localhost:3000)
kubectl port-forward -n pganalytics svc/pganalytics-grafana 3000:3000
```

## Configuration

### Using Different Value Files

```bash
# Development (minimal resources)
helm install pganalytics pganalytics/pganalytics \
  -f values-dev.yaml

# Production (3 replicas, 512MB memory)
helm install pganalytics pganalytics/pganalytics \
  -f values-prod.yaml

# Enterprise (5+ replicas, 1GB+ memory, advanced features)
helm install pganalytics pganalytics/pganalytics \
  -f values-enterprise.yaml
```

### Override Values

```bash
# Override specific values
helm install pganalytics pganalytics/pganalytics \
  --set backend.replicaCount=5 \
  --set postgresql.persistence.size=100Gi \
  --set features.advancedAuth=true \
  --set global.domain="pganalytics.example.com"
```

## Chart Architecture

### Components

| Component | Type | Replicas | Purpose |
|-----------|------|----------|---------|
| Backend API | StatefulSet | 3 | Core API service |
| Collector | DaemonSet | 1 per node | PostgreSQL metrics collection |
| PostgreSQL | StatefulSet | 1 | Time-series data storage |
| Redis | Deployment | 1 | Session storage |
| Grafana | Deployment | 1 | Dashboard visualization |

### Networking

```
┌─────────────────────────────────────────────┐
│         Kubernetes Cluster                  │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │         Ingress (nginx)               │  │
│  │  api.pganalytics.local:443            │  │
│  └──────────────┬───────────────────────┘  │
│                 │                          │
│  ┌──────────────▼───────────────────────┐  │
│  │   pgAnalytics Backend                │  │
│  │   Service (ClusterIP:8080)           │  │
│  │   ┌────────────────────────────────┐ │  │
│  │   │ Backend Pod 0                  │ │  │
│  │   │ Backend Pod 1                  │ │  │
│  │   │ Backend Pod 2                  │ │  │
│  │   └────────────────────────────────┘ │  │
│  └──────────────┬──────────────┬────────┘  │
│                 │              │           │
│     ┌───────────▼──┐  ┌────────▼──────┐   │
│     │ PostgreSQL   │  │ Redis         │   │
│     │ StatefulSet  │  │ Deployment    │   │
│     └──────────────┘  └───────────────┘   │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │  Collector DaemonSet                  │  │
│  │  ┌──────────┬──────────┬──────────┐   │  │
│  │  │ Collector│ Collector│ Collector│   │  │
│  │  │ Node 1   │ Node 2   │ Node 3   │   │  │
│  │  └──────────┴──────────┴──────────┘   │  │
│  └──────────────────────────────────────┘  │
│                                             │
└─────────────────────────────────────────────┘
```

## Storage Requirements

| Component | Default Size | Min | Max |
|-----------|------|-----|-----|
| PostgreSQL | 50Gi | 10Gi | 1Ti |
| Redis | 5Gi | 2Gi | 100Gi |
| Grafana | 10Gi | 5Gi | 50Gi |

## Resource Requirements

### Minimum (Development)
- **CPU**: 1.5 CPU cores
- **Memory**: 2 GB
- **Storage**: 65 GB

### Recommended (Production)
- **CPU**: 4-8 CPU cores
- **Memory**: 8-16 GB
- **Storage**: 300+ GB

### Enterprise
- **CPU**: 8-16+ CPU cores
- **Memory**: 16-32+ GB
- **Storage**: 500+ GB

## Cloud Provider Guides

### AWS EKS

```bash
# Create EKS cluster
aws eks create-cluster --name pganalytics-prod \
  --version 1.27 \
  --role-arn arn:aws:iam::ACCOUNT:role/EKSServiceRole

# Get credentials
aws eks update-kubeconfig --name pganalytics-prod

# Deploy with AWS-specific storage
helm install pganalytics pganalytics/pganalytics \
  -f values-prod.yaml \
  --set postgresql.persistence.storageClass=gp3 \
  --set redis.persistence.storageClass=gp3
```

### GCP GKE

```bash
# Create GKE cluster
gcloud container clusters create pganalytics-prod \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n1-standard-2

# Deploy with GCP-specific storage
helm install pganalytics pganalytics/pganalytics \
  -f values-prod.yaml \
  --set postgresql.persistence.storageClass=standard
```

### Azure AKS

```bash
# Create AKS cluster
az aks create --resource-group myResourceGroup \
  --name pganalytics-prod \
  --node-count 3

# Deploy with Azure-specific storage
helm install pganalytics pganalytics/pganalytics \
  -f values-prod.yaml \
  --set postgresql.persistence.storageClass=managed-premium
```

## Features

### Core Features

| Feature | Development | Production | Enterprise |
|---------|-------------|-----------|------------|
| Backend API | ✅ | ✅ | ✅ |
| Collectors | ✅ | ✅ | ✅ |
| PostgreSQL | ✅ | ✅ | ✅ or External |
| Grafana Dashboards | ✅ | ✅ | ✅ |
| Redis Sessions | ✅ | ✅ | ✅ |
| Ingress/TLS | ❌ | ✅ | ✅ |
| Auto-scaling | ❌ | ✅ | ✅ |
| High Availability | ❌ | ✅ | ✅ |
| LDAP/SAML/OAuth | ❌ | ❌ | ✅ |
| Token Blacklist | ❌ | ❌ | ✅ |
| Anomaly Detection | ❌ | ❌ | ✅ |
| Encryption at Rest | ❌ | ❌ | ✅ |
| Audit Logging | ❌ | ❌ | ✅ |
| Backup/DR | ❌ | ❌ | ✅ |

### Feature Flags

Enable/disable features in `values.yaml`:

```yaml
features:
  anomalyDetection: true      # ML-based anomaly detection
  tokenBlacklist: true        # Token revocation support
  corsWhitelisting: true      # CORS origin control
  mlService: true             # Machine learning service
  advancedAuth: true          # LDAP/SAML/OAuth
```

## Helm Chart Structure

```
helm/pganalytics/
├── Chart.yaml                      # Chart metadata
├── values.yaml                     # Default values
├── values-dev.yaml                # Development environment
├── values-prod.yaml               # Production environment
├── values-enterprise.yaml         # Enterprise environment
├── .helmignore                    # Files to ignore
├── templates/
│   ├── _helpers.tpl              # Template helper functions
│   ├── NOTES.txt                 # Post-install instructions
│   ├── namespace.yaml            # Kubernetes namespace
│   ├── backend-statefulset.yaml  # Backend deployment
│   ├── backend-service.yaml      # Services (backend, pg, redis)
│   ├── ingress.yaml              # Ingress resources
│   ├── collector-daemonset.yaml  # Metrics collector
│   ├── postgresql-statefulset.yaml # PostgreSQL database
│   ├── redis-deployment.yaml     # Redis cache
│   ├── grafana-deployment.yaml   # Grafana dashboards
│   ├── configmap.yaml            # Configuration management
│   ├── secrets.yaml              # Kubernetes secrets
│   ├── rbac.yaml                 # RBAC configuration
│   └── namespace.yaml            # Pod disruption budget
└── README.md                       # This file
```

## Common Helm Commands

```bash
# Install chart
helm install pganalytics pganalytics/pganalytics -n pganalytics

# Upgrade chart
helm upgrade pganalytics pganalytics/pganalytics -n pganalytics

# Check values
helm values pganalytics -n pganalytics

# List releases
helm list -n pganalytics

# Get release history
helm history pganalytics -n pganalytics

# Rollback to previous version
helm rollback pganalytics 1 -n pganalytics

# Uninstall chart
helm uninstall pganalytics -n pganalytics

# Validate chart
helm lint helm/pganalytics

# Generate manifest
helm template pganalytics pganalytics/pganalytics -f values-prod.yaml
```

## Troubleshooting

### Pods stuck in Pending

```bash
# Check pod events
kubectl describe pod <pod-name> -n pganalytics

# Check PVC status
kubectl get pvc -n pganalytics

# Check storage class availability
kubectl get storageclass
```

### Database connection errors

```bash
# Check PostgreSQL logs
kubectl logs postgresql-0 -n pganalytics

# Test connection
kubectl exec -it postgresql-0 -n pganalytics -- \
  psql -U postgres -d pganalytics -c "SELECT 1"
```

### Collector not connecting

```bash
# Check collector logs
kubectl logs -l app.kubernetes.io/component=collector -n pganalytics

# Test endpoint connectivity
kubectl exec <collector-pod> -n pganalytics -- \
  curl http://pganalytics-backend:8080/api/v1/health
```

### High memory usage

```bash
# Check resource usage
kubectl top pods -n pganalytics

# Check PostgreSQL stats
kubectl exec postgresql-0 -n pganalytics -- \
  psql -U postgres -d pganalytics -c \
  "SELECT * FROM pg_stat_statements LIMIT 10"
```

## Security Best Practices

1. **Change default secrets** before production deployment
   ```bash
   helm install pganalytics pganalytics/pganalytics \
     --set secrets.jwtSecret="$(openssl rand -base64 32)" \
     --set secrets.dbPassword="$(openssl rand -base64 32)"
   ```

2. **Use external secret management** (Vault, AWS Secrets Manager)
   ```yaml
   secrets:
     create: false  # Use externalSecrets instead
   ```

3. **Enable RBAC** for access control
   ```yaml
   rbac:
     create: true
   ```

4. **Enable NetworkPolicy** for network isolation
   ```yaml
   networkPolicy:
     enabled: true
   ```

5. **Use TLS/HTTPS** for encrypted communication
   ```yaml
   global:
     tls:
       enabled: true
       issuer: "letsencrypt-prod"
   ```

## Support

- **Documentation**: https://pganalytics.io/docs
- **GitHub**: https://github.com/pganalytics/pganalytics-v3
- **Issues**: https://github.com/pganalytics/pganalytics-v3/issues
- **Community Chat**: https://slack.pganalytics.io
- **Enterprise Support**: support@pganalytics.io

## License

Apache License 2.0 - See LICENSE file

## Changelog

### v3.3.0 (February 26, 2026)
- Initial release
- Kubernetes 1.24+ support
- Helm 3.0+ support
- Multi-environment values (dev/prod/enterprise)
- High availability configuration
- Auto-scaling support
- PostgreSQL 17 support
- Redis session storage
- Grafana dashboards
- Complete documentation

---

**Maintained by**: pgAnalytics Team
**Repository**: https://github.com/pganalytics/pganalytics-v3
**Last Updated**: February 26, 2026
