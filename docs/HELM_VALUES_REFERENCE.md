# pgAnalytics Helm Chart - Values Reference

**Version**: 3.3.0
**Last Updated**: February 26, 2026

Complete reference for all configurable values in the pgAnalytics Helm chart.

---

## Table of Contents

1. [Global Configuration](#global-configuration)
2. [Backend Configuration](#backend-configuration)
3. [Collector Configuration](#collector-configuration)
4. [PostgreSQL Configuration](#postgresql-configuration)
5. [Redis Configuration](#redis-configuration)
6. [Grafana Configuration](#grafana-configuration)
7. [Networking](#networking)
8. [Security & RBAC](#security--rbac)
9. [Features & Advanced Options](#features--advanced-options)
10. [Secrets Management](#secrets-management)
11. [Examples](#examples)

---

## Global Configuration

### `global.projectName`
**Type**: `string`
**Default**: `pganalytics`
**Description**: Project name used for Kubernetes labels and naming

**Example**:
```yaml
global:
  projectName: my-pganalytics
```

---

### `global.environment`
**Type**: `string` (`development|staging|production`)
**Default**: `development`
**Description**: Deployment environment

**Example**:
```yaml
global:
  environment: production
```

---

### `global.domain`
**Type**: `string`
**Default**: `pganalytics.local`
**Description**: Base domain for all ingress hosts

**Example**:
```yaml
global:
  domain: pganalytics.example.com
```

---

### `global.imageRegistry`
**Type**: `string`
**Default**: `docker.io`
**Description**: Docker image registry URL

**Example**:
```yaml
global:
  imageRegistry: gcr.io/my-project
```

---

### `global.imagePullPolicy`
**Type**: `string` (`Always|IfNotPresent|Never`)
**Default**: `IfNotPresent`
**Description**: Image pull policy for all containers

**Example**:
```yaml
global:
  imagePullPolicy: Always
```

---

### `global.tls.enabled`
**Type**: `boolean`
**Default**: `false`
**Description**: Enable TLS/HTTPS for ingress

**Example**:
```yaml
global:
  tls:
    enabled: true
    issuer: "letsencrypt-prod"
```

---

## Backend Configuration

### `backend.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Deploy pgAnalytics backend API

---

### `backend.replicaCount`
**Type**: `integer`
**Default**: `3`
**Valid Range**: `1-100`
**Description**: Number of backend replicas (pods)

**Example**:
```yaml
backend:
  replicaCount: 5
```

---

### `backend.image.repository`
**Type**: `string`
**Default**: `pganalytics/api`
**Description**: Docker image repository for backend

---

### `backend.image.tag`
**Type**: `string`
**Default**: `3.3.0`
**Description**: Docker image tag/version

**Example**:
```yaml
backend:
  image:
    tag: "3.3.0"
```

---

### `backend.service.type`
**Type**: `string` (`ClusterIP|NodePort|LoadBalancer`)
**Default**: `ClusterIP`
**Description**: Kubernetes service type for backend

**Example**:
```yaml
backend:
  service:
    type: LoadBalancer
```

---

### `backend.service.port`
**Type**: `integer`
**Default**: `8080`
**Valid Range**: `1-65535`
**Description**: Backend service port

---

### `backend.resources.requests.cpu`
**Type**: `string`
**Default**: `250m`
**Description**: CPU request per backend pod

**Valid Values**: `100m`, `250m`, `500m`, `1000m`, etc.

**Example**:
```yaml
backend:
  resources:
    requests:
      cpu: 500m
      memory: 256Mi
```

---

### `backend.resources.requests.memory`
**Type**: `string`
**Default**: `256Mi`
**Description**: Memory request per backend pod

**Valid Values**: `128Mi`, `256Mi`, `512Mi`, `1Gi`, etc.

---

### `backend.resources.limits.cpu`
**Type**: `string`
**Default**: `1000m`
**Description**: CPU limit per backend pod

---

### `backend.resources.limits.memory`
**Type**: `string`
**Default**: `512Mi`
**Description**: Memory limit per backend pod

---

### `backend.autoscaling.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable Horizontal Pod Autoscaler (HPA)

**Example**:
```yaml
backend:
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70
```

---

### `backend.autoscaling.minReplicas`
**Type**: `integer`
**Default**: `3`
**Description**: Minimum number of replicas for autoscaling

---

### `backend.autoscaling.maxReplicas`
**Type**: `integer`
**Default**: `10`
**Description**: Maximum number of replicas for autoscaling

---

### `backend.autoscaling.targetCPUUtilizationPercentage`
**Type**: `integer`
**Default**: `70`
**Valid Range**: `1-100`
**Description**: Target CPU utilization percentage for scaling

---

### `backend.ingress.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable ingress for backend API

---

### `backend.ingress.hosts[].host`
**Type**: `string`
**Default**: `api.pganalytics.local`
**Description**: Hostname for ingress

**Example**:
```yaml
backend:
  ingress:
    hosts:
      - host: api.pganalytics.example.com
        paths:
          - path: /
            pathType: Prefix
```

---

## Collector Configuration

### `collector.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Deploy pgAnalytics collectors

---

### `collector.mode`
**Type**: `string` (`daemonset|deployment`)
**Default**: `daemonset`
**Description**: Deployment mode for collectors

| Mode | Use Case |
|------|----------|
| `daemonset` | Run one collector per node (recommended) |
| `deployment` | Run fixed number of replicas |

---

### `collector.image.tag`
**Type**: `string`
**Default**: `3.3.0`
**Description**: Collector Docker image tag

---

### `collector.resources.requests.cpu`
**Type**: `string`
**Default**: `100m`
**Description**: CPU request per collector pod

---

### `collector.resources.requests.memory`
**Type**: `string`
**Default**: `128Mi`
**Description**: Memory request per collector pod

---

### `collector.daemonset.updateStrategy.type`
**Type**: `string` (`RollingUpdate|OnDelete`)
**Default**: `RollingUpdate`
**Description**: DaemonSet update strategy

---

## PostgreSQL Configuration

### `postgresql.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Deploy PostgreSQL database

**Note**: Set to `false` if using external PostgreSQL

---

### `postgresql.image.tag`
**Type**: `string`
**Default**: `17-alpine`
**Description**: PostgreSQL image tag

**Available Versions**: `13-alpine`, `14-alpine`, `15-alpine`, `16-alpine`, `17-alpine`

---

### `postgresql.persistence.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable persistent volume for PostgreSQL

---

### `postgresql.persistence.size`
**Type**: `string`
**Default**: `50Gi`
**Description**: Persistent volume size for PostgreSQL

**Recommendations**:
- Development: `10Gi`
- Production: `100Gi`
- Enterprise: `500Gi+`

---

### `postgresql.persistence.storageClass`
**Type**: `string`
**Default**: `standard`
**Description**: Kubernetes storage class for PostgreSQL

**Examples**:
- AWS: `gp2`, `gp3`, `io1`
- GCP: `standard`, `premium-rwo`
- Azure: `managed-standard`, `managed-premium`

---

### `postgresql.resources.requests.memory`
**Type**: `string`
**Default**: `1Gi`
**Description**: Memory request for PostgreSQL

---

### `postgresql.resources.limits.memory`
**Type**: `string`
**Default**: `2Gi`
**Description**: Memory limit for PostgreSQL

---

## Redis Configuration

### `redis.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Deploy Redis for session storage

---

### `redis.image.tag`
**Type**: `string`
**Default**: `7-alpine`
**Description**: Redis image tag

---

### `redis.persistence.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable persistent storage for Redis

---

### `redis.persistence.size`
**Type**: `string`
**Default**: `5Gi`
**Description**: Persistent volume size for Redis

**Recommendations**:
- Development: `2Gi`
- Production: `10Gi`
- Enterprise: `20Gi+`

---

### `redis.resources.requests.memory`
**Type**: `string`
**Default**: `128Mi`
**Description**: Memory request for Redis

---

### `redis.resources.limits.memory`
**Type**: `string`
**Default**: `256Mi`
**Description**: Memory limit for Redis

---

## Grafana Configuration

### `grafana.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Deploy Grafana for dashboards

---

### `grafana.image.tag`
**Type**: `string`
**Default**: `11.0.0`
**Description**: Grafana image tag

---

### `grafana.service.port`
**Type**: `integer`
**Default**: `3000`
**Description**: Grafana service port

---

### `grafana.persistence.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable persistent storage for Grafana dashboards

---

### `grafana.persistence.size`
**Type**: `string`
**Default**: `10Gi`
**Description**: Persistent volume size for Grafana

---

### `grafana.ingress.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable ingress for Grafana

---

### `grafana.ingress.hosts[].host`
**Type**: `string`
**Default**: `grafana.pganalytics.local`
**Description**: Hostname for Grafana ingress

---

## Networking

### `networkPolicy.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable Kubernetes NetworkPolicy

**Example** (restrict traffic):
```yaml
networkPolicy:
  enabled: true

  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: pganalytics
```

---

## Security & RBAC

### `rbac.create`
**Type**: `boolean`
**Default**: `true`
**Description**: Create RBAC resources (ServiceAccount, Role, RoleBinding)

---

### `rbac.serviceAccount.create`
**Type**: `boolean`
**Default**: `true`
**Description**: Create service account

---

### `rbac.serviceAccount.name`
**Type**: `string`
**Default**: (auto-generated)
**Description**: Service account name

**Example**:
```yaml
rbac:
  serviceAccount:
    create: true
    name: pganalytics-sa
```

---

### `podDisruptionBudget.enabled`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable Pod Disruption Budget for HA

---

### `podDisruptionBudget.minAvailable`
**Type**: `integer`
**Default**: `1`
**Description**: Minimum available pods during disruption

---

## Features & Advanced Options

### `features.anomalyDetection`
**Type**: `boolean`
**Default**: `false`
**Description**: Enable ML-based anomaly detection

---

### `features.tokenBlacklist`
**Type**: `boolean`
**Default**: `false`
**Description**: Enable token blacklisting for logout

---

### `features.corsWhitelisting`
**Type**: `boolean`
**Default**: `true`
**Description**: Enable CORS origin whitelisting

---

### `features.mlService`
**Type**: `boolean`
**Default**: `false`
**Description**: Enable ML service for analytics

---

### `features.advancedAuth`
**Type**: `boolean`
**Default**: `false`
**Description**: Enable LDAP/SAML/OAuth authentication

---

## Secrets Management

### `secrets.create`
**Type**: `boolean`
**Default**: `true`
**Description**: Create Kubernetes secrets

---

### `secrets.jwtSecret`
**Type**: `string`
**Default**: `change-me-in-production`
**Description**: JWT signing secret

**Security**: Change this value before production deployment!

**Generate secure secret**:
```bash
openssl rand -base64 32
```

---

### `secrets.dbPassword`
**Type**: `string`
**Default**: `change-me-in-production`
**Description**: PostgreSQL database password

---

### `secrets.redisPassword`
**Type**: `string`
**Default**: `change-me-in-production`
**Description**: Redis password

---

### `secrets.postgresRootPassword`
**Type**: `string`
**Default**: `change-me-in-production`
**Description**: PostgreSQL root password

---

### `secrets.grafanaPassword`
**Type**: `string`
**Default**: `change-me-in-production`
**Description**: Grafana admin password

---

### `secrets.collectorApiKey`
**Type**: `string`
**Default**: `change-me-in-production`
**Description**: API key for collectors to authenticate

---

## Examples

### Minimal Development Deployment

```yaml
global:
  environment: development
  domain: pganalytics.local

backend:
  replicaCount: 1
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 256Mi
  autoscaling:
    enabled: false

collector:
  resources:
    requests:
      cpu: 50m
      memory: 64Mi

postgresql:
  persistence:
    size: 10Gi

redis:
  persistence:
    size: 2Gi

features:
  anomalyDetection: false
  advancedAuth: false
```

### Production Deployment

```yaml
global:
  environment: production
  domain: pganalytics.example.com
  tls:
    enabled: true

backend:
  replicaCount: 3
  resources:
    requests:
      cpu: 1000m
      memory: 512Mi
    limits:
      cpu: 2000m
      memory: 1Gi
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 70

postgresql:
  persistence:
    size: 100Gi
    storageClass: fast-ssd

redis:
  persistence:
    size: 20Gi
    storageClass: fast-ssd

networkPolicy:
  enabled: true

podDisruptionBudget:
  enabled: true
  minAvailable: 2

features:
  advancedAuth: true
  tokenBlacklist: true
```

### Enterprise Deployment

```yaml
global:
  environment: production
  domain: pganalytics.enterprise.com
  tls:
    enabled: true

backend:
  replicaCount: 5
  autoscaling:
    enabled: true
    minReplicas: 5
    maxReplicas: 20

collector:
  mode: daemonset

postgresql:
  enabled: false  # Use managed database

redis:
  persistence:
    size: 50Gi
    storageClass: fast-ssd

networkPolicy:
  enabled: true

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

features:
  anomalyDetection: true
  tokenBlacklist: true
  advancedAuth: true
  mlService: true

secrets:
  create: false  # Use external secret management
```

---

## Installation Commands

```bash
# Development
helm install pganalytics pganalytics/pganalytics \
  -f values-dev.yaml

# Production
helm install pganalytics pganalytics/pganalytics \
  -f values-prod.yaml \
  --set secrets.jwtSecret="$(openssl rand -base64 32)"

# Enterprise
helm install pganalytics pganalytics/pganalytics \
  -f values-enterprise.yaml \
  --set global.domain="pganalytics.example.com"

# Custom values
helm install pganalytics pganalytics/pganalytics \
  --set backend.replicaCount=5 \
  --set postgresql.persistence.size=100Gi \
  --set features.advancedAuth=true
```

---

**Last Updated**: February 26, 2026
**Version**: 3.3.0
**Status**: Production Ready
