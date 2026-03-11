# Phase 1 - Lições Aprendidas & Guia de Correções
## pgAnalytics v3.3.0 - Documentação de Erros e Soluções

**Data**: March 11-12, 2026
**Objetivo**: Evitar os mesmos problemas em Phase 2 (Production)

---

## 🔴 Erros Encontrados & Correções Aplicadas

### 1. **Go Version Incompatibility**

**Problema:**
```
Error: go: go.mod requires go >= 1.24.0 (running go 1.22.12; GOTOOLCHAIN=local)
```

**Causa:**
- `go.mod` especificava `go 1.24` mas `backend/Dockerfile` usava `golang:1.22-alpine`
- Não há validação no Dockerfile contra go.mod

**Solução Aplicada:**
```dockerfile
# ANTES (ERRADO):
FROM golang:1.22-alpine AS builder

# DEPOIS (CORRETO):
FROM golang:1.24-alpine AS builder
```

**Como Evitar Próxima Vez:**
- ✅ SEMPRE validar `go.mod` vs Dockerfile base image
- ✅ Criar script de validação automático:
  ```bash
  # scripts/validate-go-version.sh
  GO_VERSION=$(grep "^go " go.mod | awk '{print $2}')
  DOCKERFILE_VERSION=$(grep "golang:" backend/Dockerfile | grep -oP '\d+\.\d+')
  if [ "$GO_VERSION" != "$DOCKERFILE_VERSION" ]; then
    echo "ERROR: go.mod requires $GO_VERSION but Dockerfile has $DOCKERFILE_VERSION"
    exit 1
  fi
  ```
- ✅ Adicionar ao GitHub Actions para CI/CD

---

### 2. **Unused Imports Compilation Error**

**Problema:**
```
backend/internal/cache/config_cache.go:4:2: "context" imported and not used
backend/internal/jobs/alert_rule_engine.go:12:2: "github.com/.../models" imported and not used
```

**Causa:**
- Code refactoring removeu usage de imports mas não removeu os imports
- Nenhuma linter configurado (golangci-lint)

**Solução Aplicada:**
```go
// ANTES (ERRADO):
import (
    "context"  // ❌ Not used
    "sync"
)

// DEPOIS (CORRETO):
import (
    "sync"
)
```

**Como Evitar Próxima Vez:**
- ✅ Configurar `.golangci.yml` com `unused` linter
- ✅ Executar antes do build:
  ```bash
  golangci-lint run ./backend/...
  ```
- ✅ Integrar no GitHub Actions pre-commit check

---

### 3. **Private Field Access in External Packages**

**Problema:**
```
backend/internal/jobs/collector_cleanup.go:198:24: ccj.db.db undefined
(cannot refer to unexported field db)
```

**Causa:**
- Package `jobs` tentou acessar campo privado `db` de `PostgresDB` struct
- Não havia método público para execução de queries genéricas

**Solução Aplicada:**
```go
// ANTES (ERRADO):
// Em jobs/collector_cleanup.go
result, err := ccj.db.db.ExecContext(ctx, query, args...)

// DEPOIS (CORRETO):
// 1. Adicionar método público em storage/postgres.go:
func (p *PostgresDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    return p.db.ExecContext(ctx, query, args...)
}

// 2. Usar em jobs/collector_cleanup.go:
result, err := ccj.db.ExecContext(ctx, query, args...)
```

**Como Evitar Próxima Vez:**
- ✅ Usar `go vet` que detecta campo privado access
- ✅ Padrão: Se pacote B precisa acessar dados de pacote A, criar método público em A
- ✅ Código de revisão deve verificar:
  - Nenhum acesso a campos minúsculos (privados) de structs
  - Métodos públicos para operações comuns

---

### 4. **TimescaleDB Extension Not Available**

**Problema:**
```
FATAL: could not access file "timescaledb": No such file or directory
```

**Causa:**
- Tentei usar `-c shared_preload_libraries=timescaledb` com image `postgres:16-bullseye`
- TimescaleDB extension não está instalada na imagem padrão

**Solução Aplicada:**
```dockerfile
# CRIADO: Dockerfile.timescaledb
FROM postgres:16-bullseye

RUN apt-get update && \
    apt-get install -y gnupg wget curl build-essential && \
    echo "deb https://packagecloud.io/timescale/timescaledb/debian/ bullseye main" | \
    tee /etc/apt/sources.list.d/timescaledb.list && \
    wget --quiet -O - https://packagecloud.io/timescale/timescaledb/gpgkey | apt-key add - && \
    apt-get update && \
    apt-get install -y timescaledb-2-postgresql-16 && \
    apt-get clean

RUN echo "shared_preload_libraries = 'timescaledb'" >> /etc/postgresql/postgresql.conf.sample

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["postgres"]
```

**Como Evitar Próxima Vez:**
- ✅ SEMPRE usar imagem oficial se disponível
- ✅ Se não disponível, verificar repositórios e criar Dockerfile customizado
- ✅ Documentar a imagem base e versões na docker-compose
- ✅ Testar extensão após build:
  ```sql
  CREATE EXTENSION IF NOT EXISTS timescaledb;
  SELECT extname FROM pg_extension WHERE extname = 'timescaledb';
  ```

---

### 5. **Frontend Proxy Not Configured Correctly**

**Problema:**
```
curl: (56) Recv failure: Connection reset by peer
Frontend logs: "API Backend: http://backend:8080"
```

**Causa:**
- Frontend proxy.js usar defaults (`backend:8080`) em vez de variáveis de ambiente
- Docker environment variables não passadas para proxy.js
- Port mapping inconsistente (3000:80 vs 3000:3000)

**Solução Aplicada:**
```yaml
# ANTES (ERRADO):
frontend-staging:
  environment:
    REACT_APP_API_URL: "https://localhost:8080"  # ❌ Wrong for Docker network
    REACT_APP_ENVIRONMENT: "staging"
  ports:
    - "3000:80"  # ❌ Mismatch with Node.js port 3000

# DEPOIS (CORRETO):
frontend-staging:
  environment:
    REACT_APP_API_URL: "https://localhost:8080"  # Browser
    REACT_APP_ENVIRONMENT: "staging"
    VITE_API_BACKEND_HOST: "backend-staging"     # ✅ Docker network
    VITE_API_BACKEND_PORT: "8080"
    VITE_API_BACKEND_PROTOCOL: "http"
    PORT: "3000"
  ports:
    - "3000:3000"  # ✅ Correct mapping
```

**Como Evitar Próxima Vez:**
- ✅ Documentar TODAS as variáveis de ambiente esperadas
- ✅ Criar tabela de configuração:

```markdown
# Frontend Environment Variables

| Variable | Purpose | Docker Network | Browser |
|----------|---------|-----------------|---------|
| VITE_API_BACKEND_HOST | API hostname | `backend-staging` | `localhost` |
| VITE_API_BACKEND_PORT | API port | `8080` | `8080` |
| VITE_API_BACKEND_PROTOCOL | API protocol | `http` | `https` |
| PORT | Server port | `3000` | `3000` |
```

- ✅ Criar `.env.staging` template com todas as variáveis
- ✅ Validar no Dockerfile que env vars estão definidas:
  ```dockerfile
  RUN test -n "$VITE_API_BACKEND_HOST" || \
      (echo "ERROR: VITE_API_BACKEND_HOST not set" && exit 1)
  ```

---

### 6. **Missing Database Schema on Startup**

**Problema:**
```
ERROR: Failed to list managed instances: pq: relation "pganalytics.managed_instances" does not exist
```

**Causa:**
- Migrations não foram executadas automaticamente (comentei no docker-compose)
- Schema não inicializado manualmente

**Solução Aplicada:**
```bash
# Criado script de inicialização manual
docker exec pganalytics-staging-postgres psql -U postgres -d pganalytics_staging << 'EOF'
CREATE SCHEMA IF NOT EXISTS pganalytics;
CREATE TABLE IF NOT EXISTS pganalytics.managed_instances (...);
-- etc
EOF
```

**Como Evitar Próxima Vez:**
- ✅ Criar `backend/migrations/init.sql` com schema base
- ✅ Configurar auto-execution no Docker:
  ```yaml
  postgres-staging:
    volumes:
      - ./backend/migrations/init.sql:/docker-entrypoint-initdb.d/000_init.sql:ro
  ```
- ✅ Numerar migrations corretamente (001_, 002_, etc)
- ✅ Adicionar healthcheck que valida schema existe
- ✅ Script de validação:
  ```sql
  -- backend/migrations/validate.sql
  DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'pganalytics'
                   AND table_name = 'managed_instances') THEN
      RAISE EXCEPTION 'Schema not initialized';
    END IF;
  END $$;
  ```

---

## 📋 Checklist para Próximos Deployments

### Pré-Build
- [ ] Validar Go version vs go.mod
- [ ] Executar golangci-lint
- [ ] Validar todos imports estão sendo usados
- [ ] Verificar todos campos privados vs métodos públicos
- [ ] Validar variáveis de ambiente esperadas estão documentadas

### Build
- [ ] Backend compila sem warnings
- [ ] Frontend compila sem warnings
- [ ] Docker images buildadas com sucesso
- [ ] Verificar tamanho das imagens (backend < 100MB)

### Deployment
- [ ] Todos containers iniciam
- [ ] Health checks passam
- [ ] API endpoints respondem
- [ ] Frontend carrega corretamente
- [ ] Banco de dados schema inicializado
- [ ] Conectividade entre serviços verificada

### Post-Deployment
- [ ] Testar cada endpoint
- [ ] Verificar logs não contêm errors
- [ ] Validar integração frontend <-> backend
- [ ] Verificar dados fluindo corretamente

---

## 🛠️ Scripts de Validação Recomendados

### 1. Pre-Build Validation
```bash
#!/bin/bash
# scripts/validate-build.sh

set -e

echo "🔍 Validating build..."

# Check Go version
GO_VERSION=$(grep "^go " go.mod | awk '{print $2}')
DOCKERFILE_VERSION=$(grep "golang:" backend/Dockerfile | grep -oP '\d+\.\d+')
if [ "$GO_VERSION" != "$DOCKERFILE_VERSION" ]; then
  echo "❌ Go version mismatch"
  exit 1
fi

# Run linter
golangci-lint run ./backend/...

# Check no unused imports
go vet ./backend/...

echo "✅ Build validation passed"
```

### 2. Health Check
```bash
#!/bin/bash
# scripts/health-check.sh

set -e

echo "🏥 Checking service health..."

# Backend
curl -s http://localhost:8080/api/v1/health | grep -q '"status":"ok"' || \
  (echo "❌ Backend health check failed" && exit 1)

# Frontend
curl -s http://localhost:3000/ | grep -q '<div id="root"></div>' || \
  (echo "❌ Frontend health check failed" && exit 1)

# Grafana
curl -s http://localhost:3001/api/health | grep -q '"database":"ok"' || \
  (echo "❌ Grafana health check failed" && exit 1)

# Prometheus
curl -s http://localhost:9090/api/v1/status/config | grep -q '"status":"success"' || \
  (echo "❌ Prometheus health check failed" && exit 1)

echo "✅ All services healthy"
```

### 3. Database Validation
```bash
#!/bin/bash
# scripts/validate-db.sh

set -e

echo "🗄️ Validating database..."

docker exec pganalytics-staging-postgres psql -U postgres -d pganalytics_staging << 'EOF'
-- Check schema exists
SELECT schema_name FROM information_schema.schemata
WHERE schema_name = 'pganalytics' OR \
  RAISE EXCEPTION 'pganalytics schema not found';

-- Check all required tables exist
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'pganalytics'
  AND table_name IN (
    'managed_instances', 'servers', 'collectors', 'users', 'alerts'
  );

SELECT COUNT(*) as required_tables FROM (VALUES
  ('managed_instances'), ('servers'), ('collectors'), ('users'), ('alerts')
) AS required_tables(name)
WHERE NOT EXISTS (
  SELECT 1 FROM information_schema.tables
  WHERE table_schema = 'pganalytics' AND table_name = name
) OR \
  RAISE EXCEPTION 'Missing required tables';
EOF

docker exec pganalytics-staging-timescale psql -U postgres -d metrics_staging << 'EOF'
-- Check TimescaleDB extension
SELECT extname FROM pg_extension WHERE extname = 'timescaledb' OR \
  RAISE EXCEPTION 'TimescaleDB extension not found';

-- Check hypertable exists
SELECT hypertable_name FROM timescaledb_information.hypertables
WHERE hypertable_name = 'metrics_time_series' OR \
  RAISE EXCEPTION 'metrics_time_series hypertable not found';
EOF

echo "✅ Database validation passed"
```

---

## 📝 Documentação a Ser Adicionada

### 1. Environment Variables Documentation
```markdown
# Environment Variables Guide

## Frontend
- VITE_API_BACKEND_HOST: Backend hostname (Docker: backend-staging, Browser: localhost)
- VITE_API_BACKEND_PORT: Backend port (8080)
- VITE_API_BACKEND_PROTOCOL: http or https
- PORT: Frontend server port (3000)

## Backend
- DATABASE_URL: PostgreSQL connection string
- TIMESCALE_URL: TimescaleDB connection string
- JWT_SECRET: JWT signing key
- PORT: API server port (8080)

## Database
- POSTGRES_USER: postgres
- POSTGRES_PASSWORD: Strong password
- POSTGRES_DB: pganalytics_staging
```

### 2. Configuration Template
```bash
# .env.staging.example
VITE_API_BACKEND_HOST=backend-staging
VITE_API_BACKEND_PORT=8080
VITE_API_BACKEND_PROTOCOL=http
PORT=3000

DATABASE_URL=postgres://postgres:password@postgres-staging:5432/pganalytics_staging
TIMESCALE_URL=postgres://postgres:password@timescale-staging:5432/metrics_staging
JWT_SECRET=change-in-production
```

---

## ✅ Checklist Aplicado em Phase 1

- [x] Dockerfile Go version mismatch - FIXED
- [x] Unused imports compilation - FIXED
- [x] Private field access - FIXED
- [x] TimescaleDB extension - FIXED
- [x] Frontend proxy configuration - FIXED
- [x] Database schema initialization - FIXED

## 🎯 Para Phase 2 (Production)

- [ ] Implement all validation scripts
- [ ] Add GitHub Actions checks
- [ ] Create comprehensive environment guide
- [ ] Setup automated testing
- [ ] Configure CI/CD pipeline
- [ ] Document all configuration parameters
- [ ] Create runbooks for troubleshooting
- [ ] Setup monitoring and alerting

---

**Status**: 🟢 **All issues documented and fixed**
**Ready for Phase 2**: ✅ **YES**
**Documentation Level**: 📚 **Complete with examples**

