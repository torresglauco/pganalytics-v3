# pgAnalytics v3 Integration Guide

## Architecture Overview

pgAnalytics v3 is a comprehensive PostgreSQL monitoring and analytics platform with the following components:

- **Backend**: Go REST API and WebSocket server
- **Frontend**: Node.js/React web application
- **Collector**: C++ agent for collecting database metrics
- **CLI**: Go command-line interface for automation
- **MCP Server**: Go Model Context Protocol server for AI integration
- **ML Service**: Python/FastAPI service for anomaly detection and predictions

## Component Integration

### Backend ↔ Frontend
The frontend communicates with the backend through multiple channels:

- **REST APIs**: Standard HTTP endpoints for CRUD operations
- **WebSocket**: Real-time metric streaming and notifications
- **Authentication/Authorization**: JWT token-based security

**Key Endpoints**:
- `/api/v1/metrics/*` - Metric data
- `/api/v1/queries/*` - Query analytics
- `/api/v1/agents/*` - Agent management
- `/ws/*` - WebSocket connections

### Backend ↔ Collector
The collector agents register with the backend and push collected metrics:

- **HTTP/gRPC registration**: Initial agent registration and keepalive
- **Data ingestion**: Batch metrics transmission
- **Configuration sync**: Remote configuration updates

**Protocol**:
- POST `/api/v1/agents/register` - Agent registration
- POST `/api/v1/metrics/batch` - Bulk metric ingestion
- GET `/api/v1/agents/{id}/config` - Config retrieval

### Backend ↔ MCP Server
The MCP server enables AI model integration through stdio JSON-RPC:

- **Process management**: Lifecycle management via Go process
- **stdio JSON-RPC**: Standard MCP communication protocol
- **Tool registration**: Dynamic tool exposure to Claude

**Available Tools**:
- `query_database` - Execute SQL queries
- `get_metrics` - Retrieve metric data
- `analyze_anomalies` - ML-powered anomaly detection
- `get_agent_status` - Agent health status
- `list_databases` - Database enumeration

### Backend ↔ CLI
The CLI provides command-line automation and integration:

- **HTTP API calls**: Direct backend communication
- **Config file management**: ~/.pganalytics/config.yaml
- **Authentication**: Token-based or system authentication

**Main Commands**:
- `pganalytics-cli config set` - Configuration management
- `pganalytics-cli metric get` - Metric retrieval
- `pganalytics-cli query run` - Query execution
- `pganalytics-cli agent list` - Agent management

### ML Service Integration
The FastAPI service provides machine learning capabilities:

- **Service Port**: 5000
- **Endpoints**:
  - `/predict` - Predictive analytics (port 5000)
  - `/detect-anomaly` - Anomaly detection
  - `/health` - Service health check

**Communication**: Backend calls ML service via HTTP POST with metric data

## Setup Instructions

### 1. Backend Setup
```bash
cd backend
go mod download
mise run build:backend
mise run test:backend
```

### 2. Frontend Setup
```bash
cd frontend
npm install
npm run build
npm run test
```

### 3. Collector Setup
```bash
cd collector
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
mise run build:collector
```

### 4. CLI Setup
```bash
cd backend
go build -o pganalytics-cli ./cmd/pganalytics-cli
```

### 5. MCP Server Setup
```bash
cd backend
go build -o pganalytics-mcp ./cmd/mcp-server
```

### 6. ML Service Setup
```bash
cd ml-service
pip install -r requirements.txt
python -m uvicorn main:app --port 5000
```

### 7. Integration Testing
```bash
mise run test:integration
mise run test:e2e
```

## Testing Integration

### Unit Tests per Component
- **Backend**: `go test ./...` (90%+ coverage)
- **Frontend**: `npm run test` (85%+ coverage)
- **Collector**: C++ unit tests
- **CLI**: `go test ./cmd/pganalytics-cli/...`
- **MCP**: `go test ./internal/mcp/...`

### Integration Tests
Tests that verify communication between components:

```bash
mise run test:integration
```

These tests validate:
- Agent registration and metrics ingestion
- Frontend API communication
- WebSocket real-time updates
- CLI operations
- MCP tool invocation

### E2E Tests
Full workflow tests using Playwright:

```bash
mise run test:frontend:e2e
```

Scenarios:
- Complete monitoring setup workflow
- Metric collection and visualization
- Anomaly detection and alerting
- Query execution and analysis

## Deployment

### Docker Compose (Recommended for Development)
```bash
docker-compose up -d
```

This starts:
- PostgreSQL database
- Backend API server
- Frontend web UI
- ML Service
- Collector (optional)

### Kubernetes (Production)
```bash
# Install Helm chart
helm install pganalytics ./helm-charts/pganalytics
```

Features:
- Horizontal scaling
- High availability
- Auto-recovery
- Rolling updates

### On-Premise (Manual Installation)
1. Install PostgreSQL 14+
2. Build all components
3. Configure systemd services
4. Set up TLS certificates
5. Configure firewall rules

## Security Considerations

### Component Communication
- **Backend ↔ Collector**: mTLS with certificate pinning
- **Backend ↔ ML Service**: Internal network or mTLS
- **Frontend ↔ Backend**: HTTPS with CORS validation
- **CLI ↔ Backend**: API token authentication

### Data Protection
- All sensitive data encrypted at rest
- Database connection pooling with limits
- Input validation on all endpoints
- Rate limiting and DDoS protection

### Access Control
- Role-based access control (RBAC)
- Agent identity verification
- API key management
- Audit logging

## Troubleshooting

### Agent Connection Issues
```bash
# Check agent status
pganalytics-cli agent list

# Test connectivity
curl -X GET http://backend:8080/api/v1/agents/{id}/health
```

### Metric Collection
```bash
# Check collector logs
tail -f /var/log/pganalytics-collector.log

# Verify metric ingestion
pganalytics-cli metric get --agent-id {id} --limit 10
```

### WebSocket Connection
```bash
# Test WebSocket connectivity
wscat -c ws://localhost:8080/ws/metrics
```

### ML Service
```bash
# Check service status
curl http://localhost:5000/health

# Test prediction endpoint
curl -X POST http://localhost:5000/predict \
  -H "Content-Type: application/json" \
  -d '{"metrics": [...]}'
```

## Performance Optimization

### Caching Strategy
- Frontend caches API responses (5 minute TTL)
- Backend caches query plans
- ML service caches models

### Database Optimization
- Indexing on frequently queried columns
- Partition tables by time range
- Regular VACUUM and ANALYZE

### Connection Pooling
- Backend: PgBouncer with 100 connections
- ML Service: Connection pool of 20
- Collector: Batch requests to minimize connections

## Monitoring Integration Points

- **Prometheus**: Export metrics endpoint at `/metrics`
- **Grafana**: Pre-built dashboards
- **ELK Stack**: Centralized logging
- **Jaeger**: Distributed tracing (optional)

## API Versioning

All APIs follow semantic versioning:
- Current version: v1
- Backward compatible changes within v1
- Breaking changes in v2+
- Deprecation notice period: 6 months
