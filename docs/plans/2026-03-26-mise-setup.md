# Mise Setup Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Centralizar o setup local do projeto com mise, permitindo que um dev novo clone e rode `mise run setup && mise run dev` para ter tudo funcionando.

**Architecture:** mise.toml na raiz define runtimes (Node 18, Go 1.24) e tasks (setup, dev, down, logs, reset, lint, test). Um script `scripts/setup.sh` orquestrado pelo mise faz o bootstrap idempotente. `.env.example` na raiz documenta todas as variáveis. `package-lock.json` passa a ser commitado.

**Tech Stack:** mise, Docker Compose, npm, Go, shell scripts

---

### Task 1: Criar `.env.example` na raiz

**Files:**
- Create: `.env.example`

**Step 1: Criar o arquivo com todas as variáveis do docker-compose.yml**

```env
# pgAnalytics Local Development Environment
# Copy this file to .env: cp .env.example .env

# Registration secret for collector auto-registration
REGISTRATION_SECRET=demo-registration-secret-change-in-production
```

**Step 2: Commit**

```bash
git add .env.example
git commit -m "chore: add root .env.example for local development"
```

---

### Task 2: Remover `package-lock.json` do `.gitignore` e gerar lockfile

**Files:**
- Modify: `.gitignore:54-55` (remover `package-lock.json` e `yarn.lock`)

**Step 1: Editar `.gitignore`**

Remover as linhas:
```
package-lock.json
yarn.lock
```

**Step 2: Gerar o lockfile**

```bash
cd frontend && npm install
```

**Step 3: Commit**

```bash
git add .gitignore frontend/package-lock.json
git commit -m "chore: track package-lock.json for reproducible builds"
```

---

### Task 3: Criar `mise.toml` com runtimes e tasks

**Files:**
- Create: `mise.toml`

**Step 1: Criar o arquivo**

```toml
[tools]
node = "18"
go = "1.24"

[tasks.setup]
description = "Bootstrap the project for local development"
run = "bash scripts/setup.sh"

[tasks.dev]
description = "Start all services for development"
run = """
docker compose up --build -d
echo ""
echo "Waiting for services to be healthy..."
echo ""
timeout=120
elapsed=0
while [ $elapsed -lt $timeout ]; do
  healthy=$(docker compose ps --format json | grep -c '"healthy"' || true)
  total=$(docker compose ps --format json | grep -c '"running"\\|"healthy"' || true)
  if [ "$healthy" -ge 4 ]; then
    break
  fi
  sleep 2
  elapsed=$((elapsed + 2))
done
echo "==================================="
echo " pgAnalytics is running!"
echo "==================================="
echo ""
echo " Frontend:  http://localhost:3000"
echo " API:       http://localhost:8080"
echo " Grafana:   http://localhost:3001"
echo ""
echo " Use 'mise run logs' to follow logs"
echo " Use 'mise run down' to stop"
echo "==================================="
"""

[tasks.dev-frontend]
description = "Start frontend in dev mode (hot reload)"
dir = "frontend"
run = "npm run dev"

[tasks.down]
description = "Stop all services"
run = "docker compose down"

[tasks.logs]
description = "Follow logs from all services"
run = "docker compose logs -f"

[tasks.reset]
description = "Stop services, remove volumes, and start fresh"
run = """
echo "This will delete all local data (databases, grafana, etc.)"
read -p "Are you sure? [y/N] " confirm
if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
  docker compose down -v
  echo "Volumes removed. Run 'mise run setup && mise run dev' to start fresh."
else
  echo "Cancelled."
fi
"""

[tasks.lint]
description = "Run linters for frontend"
dir = "frontend"
run = "npm run lint"

[tasks.test]
description = "Run tests for frontend"
dir = "frontend"
run = "npm run test"

[tasks.typecheck]
description = "Run TypeScript type checking"
dir = "frontend"
run = "npm run type-check"

[tasks.ps]
description = "Show status of all services"
run = "docker compose ps"
```

**Step 2: Commit**

```bash
git add mise.toml
git commit -m "feat: add mise.toml with runtimes and development tasks"
```

---

### Task 4: Criar script de setup idempotente

**Files:**
- Create: `scripts/setup.sh`

**Step 1: Criar o script**

```bash
#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'
CHECK="${GREEN}OK${NC}"
FAIL="${RED}FAIL${NC}"

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo ""
echo "=============================="
echo " pgAnalytics - Project Setup"
echo "=============================="
echo ""

errors=0

# 1. Check prerequisites
echo "Checking prerequisites..."

if command -v docker &> /dev/null; then
  echo -e "  Docker:          $CHECK ($(docker --version | cut -d' ' -f3 | tr -d ','))"
else
  echo -e "  Docker:          $FAIL - Install from https://docs.docker.com/get-docker/"
  errors=$((errors + 1))
fi

if docker info &> /dev/null 2>&1; then
  echo -e "  Docker running:  $CHECK"
else
  echo -e "  Docker running:  $FAIL - Start Docker Desktop or the Docker daemon"
  errors=$((errors + 1))
fi

if command -v node &> /dev/null; then
  echo -e "  Node.js:         $CHECK ($(node --version))"
else
  echo -e "  Node.js:         $FAIL - Run 'mise install' to install Node.js"
  errors=$((errors + 1))
fi

if command -v go &> /dev/null; then
  echo -e "  Go:              $CHECK ($(go version | cut -d' ' -f3))"
else
  echo -e "  Go:              $FAIL - Run 'mise install' to install Go"
  errors=$((errors + 1))
fi

if [ $errors -gt 0 ]; then
  echo ""
  echo -e "${RED}Setup cannot continue. Fix the issues above and try again.${NC}"
  exit 1
fi

echo ""

# 2. Environment file
echo "Setting up environment..."

if [ ! -f "$PROJECT_ROOT/.env" ]; then
  cp "$PROJECT_ROOT/.env.example" "$PROJECT_ROOT/.env"
  echo -e "  .env:            ${GREEN}Created from .env.example${NC}"
else
  echo -e "  .env:            $CHECK (already exists, skipping)"
fi

# 3. TLS certificates
echo ""
echo "Setting up TLS certificates..."

if [ -f "$PROJECT_ROOT/tls/server.crt" ] && [ -f "$PROJECT_ROOT/tls/server.key" ]; then
  echo -e "  TLS certs:       $CHECK (already exist, skipping)"
else
  mkdir -p "$PROJECT_ROOT/tls"
  openssl req -x509 -newkey rsa:2048 -keyout "$PROJECT_ROOT/tls/server.key" \
    -out "$PROJECT_ROOT/tls/server.crt" -days 365 -nodes \
    -subj "/CN=localhost/O=pgAnalytics Dev" 2>/dev/null
  echo -e "  TLS certs:       ${GREEN}Generated self-signed certificates${NC}"
fi

# 4. Frontend dependencies
echo ""
echo "Installing frontend dependencies..."

if [ -d "$PROJECT_ROOT/frontend/node_modules" ]; then
  echo -e "  node_modules:    $CHECK (already installed)"
  echo "  Running npm install to sync with lockfile..."
fi
cd "$PROJECT_ROOT/frontend" && npm install --no-audit --no-fund
echo -e "  npm install:     $CHECK"

# 5. Summary
echo ""
echo "=============================="
echo -e " ${GREEN}Setup complete!${NC}"
echo "=============================="
echo ""
echo " Next steps:"
echo "   mise run dev          # Start all services"
echo "   mise run dev-frontend # Start frontend with hot reload"
echo "   mise run logs         # Follow service logs"
echo ""
```

**Step 2: Tornar executável**

```bash
chmod +x scripts/setup.sh
```

**Step 3: Commit**

```bash
git add scripts/setup.sh
git commit -m "feat: add idempotent setup script orchestrated by mise"
```

---

### Task 5: Limpar scripts duplicados da raiz

**Files:**
- Delete: `start-frontend.sh` (substituído por `mise run dev-frontend`)
- Delete: `setup-registration-secret.sh` (coberto pelo .env.example)
- Delete: `cleanup-docs.sh` (utilitário pontual, não precisa ficar na raiz)
- Move: `timescale-init.sh` → já referenciado pelo docker-compose, manter

**Step 1: Remover scripts obsoletos**

```bash
git rm start-frontend.sh setup-registration-secret.sh cleanup-docs.sh
```

**Step 2: Commit**

```bash
git commit -m "chore: remove setup scripts replaced by mise tasks"
```

---

### Task 6: Atualizar README com instruções de setup via mise

**Files:**
- Modify: `README.md` (seção de Quick Start)

**Step 1: Adicionar seção de Quick Start no README**

Adicionar após o título principal, uma seção como:

```markdown
## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [mise](https://mise.jdx.dev/getting-started.html)

### Setup

```bash
# Install runtimes (Node.js, Go)
mise install

# Bootstrap the project (idempotent, safe to re-run)
mise run setup

# Start all services
mise run dev
```

### Available Commands

| Command | Description |
|---------|-------------|
| `mise run setup` | Bootstrap project (install deps, generate .env, TLS certs) |
| `mise run dev` | Start all services via Docker Compose |
| `mise run dev-frontend` | Start frontend with hot reload (Vite) |
| `mise run down` | Stop all services |
| `mise run logs` | Follow logs from all services |
| `mise run reset` | Remove all data and start fresh |
| `mise run test` | Run frontend tests |
| `mise run lint` | Run frontend linters |
| `mise run ps` | Show service status |
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: update README with mise-based quick start guide"
```

---

### Task 7: Testar o fluxo completo

**Step 1: Verificar que mise reconhece as tasks**

```bash
mise tasks
```

Expected: lista todas as tasks definidas no mise.toml

**Step 2: Rodar setup**

```bash
mise run setup
```

Expected: checklist verde, sem erros

**Step 3: Verificar que dev sobe**

```bash
mise run dev
```

Expected: containers sobem, URLs impressas no terminal
