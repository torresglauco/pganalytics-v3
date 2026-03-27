#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
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

# Container runtime (Docker or Podman)
if command -v docker &> /dev/null; then
  runtime="docker"
  echo -e "  Container runtime: $CHECK ($(docker --version | head -1 | cut -d' ' -f3 | tr -d ','))"
elif command -v podman &> /dev/null; then
  runtime="podman"
  echo -e "  Container runtime: $CHECK (podman $(podman --version | cut -d' ' -f3))"
else
  echo -e "  Container runtime: $FAIL - Install Docker or Podman"
  errors=$((errors + 1))
fi

# Check if container runtime is running
if [ "${runtime:-}" = "docker" ]; then
  if docker info &> /dev/null 2>&1; then
    echo -e "  Daemon running:    $CHECK"
  else
    echo -e "  Daemon running:    $FAIL - Start Docker Desktop or the Docker daemon"
    errors=$((errors + 1))
  fi
elif [ "${runtime:-}" = "podman" ]; then
  if podman info &> /dev/null 2>&1; then
    echo -e "  Daemon running:    $CHECK"
  else
    echo -e "  Daemon running:    $FAIL - Start Podman machine: podman machine start"
    errors=$((errors + 1))
  fi
fi

# Compose
if command -v docker &> /dev/null && docker compose version &> /dev/null 2>&1; then
  echo -e "  Compose:           $CHECK (docker compose)"
elif command -v podman-compose &> /dev/null; then
  echo -e "  Compose:           $CHECK (podman-compose)"
elif command -v docker-compose &> /dev/null; then
  echo -e "  Compose:           $CHECK (docker-compose)"
else
  echo -e "  Compose:           $FAIL - Install docker compose or podman-compose"
  errors=$((errors + 1))
fi

if command -v node &> /dev/null; then
  echo -e "  Node.js:           $CHECK ($(node --version))"
else
  echo -e "  Node.js:           $FAIL - Run 'mise install' to install Node.js"
  errors=$((errors + 1))
fi

if command -v go &> /dev/null; then
  echo -e "  Go:                $CHECK ($(go version | cut -d' ' -f3))"
else
  echo -e "  Go:                $FAIL - Run 'mise install' to install Go"
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
  echo -e "  .env:              ${GREEN}Created from .env.example${NC}"
else
  echo -e "  .env:              $CHECK (already exists, skipping)"
fi

# 3. TLS certificates
echo ""
echo "Setting up TLS certificates..."

if [ -f "$PROJECT_ROOT/tls/server.crt" ] && [ -f "$PROJECT_ROOT/tls/server.key" ]; then
  echo -e "  TLS certs:         $CHECK (already exist, skipping)"
else
  mkdir -p "$PROJECT_ROOT/tls"
  openssl req -x509 -newkey rsa:2048 -keyout "$PROJECT_ROOT/tls/server.key" \
    -out "$PROJECT_ROOT/tls/server.crt" -days 365 -nodes \
    -subj "/CN=localhost/O=pgAnalytics Dev" 2>/dev/null
  echo -e "  TLS certs:         ${GREEN}Generated self-signed certificates${NC}"
fi

# 4. Frontend dependencies
echo ""
echo "Installing frontend dependencies..."

cd "$PROJECT_ROOT/frontend" && npm install --no-audit --no-fund
echo -e "  npm install:       $CHECK"

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
