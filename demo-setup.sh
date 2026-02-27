#!/bin/bash

set -e

echo "🚀 pgAnalytics Frontend Demo Setup"
echo "=================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
BACKEND_URL="http://localhost:8080/api/v1"
FRONTEND_PORT=3000
BACKEND_PORT=8080
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="pganalytics"
POSTGRES_DB="pganalytics"

# Step 1: Start Docker services
echo -e "${BLUE}[1/6]${NC} Starting Docker services..."
docker-compose up -d
echo -e "${GREEN}✓ Docker services started${NC}"
echo ""

# Step 2: Wait for backend to be ready
echo -e "${BLUE}[2/6]${NC} Waiting for backend to be ready..."
for i in {1..30}; do
  if curl -s http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Backend is ready${NC}"
    break
  fi
  echo "  Waiting... ($i/30)"
  sleep 2
done
echo ""

# Step 3: Create test user
echo -e "${BLUE}[3/6]${NC} Creating test user..."
USER_RESPONSE=$(curl -s -X POST $BACKEND_URL/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo",
    "email": "demo@pganalytics.local",
    "password": "Demo@12345",
    "full_name": "Demo User"
  }')

AUTH_TOKEN=$(echo $USER_RESPONSE | jq -r '.token // empty')

if [ -z "$AUTH_TOKEN" ] || [ "$AUTH_TOKEN" = "null" ]; then
  echo -e "${YELLOW}⚠ User already exists, trying to login...${NC}"
  LOGIN_RESPONSE=$(curl -s -X POST $BACKEND_URL/auth/login \
    -H "Content-Type: application/json" \
    -d '{
      "username": "demo",
      "password": "Demo@12345"
    }')
  AUTH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
fi

echo -e "${GREEN}✓ User ready (token: ${AUTH_TOKEN:0:20}...)${NC}"
echo ""

# Step 4: Create registration secret
echo -e "${BLUE}[4/6]${NC} Creating registration secret..."
SECRET=$(openssl rand -hex 16)
echo -e "${GREEN}✓ Registration secret: ${SECRET}${NC}"
echo ""

# Step 5: Register a collector
echo -e "${BLUE}[5/6]${NC} Registering demo collector..."
COLLECTOR_RESPONSE=$(curl -s -X POST $BACKEND_URL/collectors/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "X-Registration-Secret: $SECRET" \
  -d '{
    "hostname": "demo-collector.pganalytics.local",
    "environment": "demo",
    "group": "showcase",
    "description": "Demo collector for testing"
  }')

COLLECTOR_ID=$(echo $COLLECTOR_RESPONSE | jq -r '.collector_id // empty')

if [ -z "$COLLECTOR_ID" ] || [ "$COLLECTOR_ID" = "null" ]; then
  echo -e "${RED}✗ Failed to register collector${NC}"
  echo "Response: $COLLECTOR_RESPONSE"
else
  echo -e "${GREEN}✓ Collector registered (ID: $COLLECTOR_ID)${NC}"
fi
echo ""

# Step 6: Create a managed instance
echo -e "${BLUE}[6/6]${NC} Creating demo managed instance..."
INSTANCE_RESPONSE=$(curl -s -X POST $BACKEND_URL/managed-instances \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -d '{
    "name": "demo-postgres",
    "hostname": "demo-db.pganalytics.local",
    "port": 5432,
    "database": "postgres",
    "description": "Demo PostgreSQL managed instance"
  }')

INSTANCE_ID=$(echo $INSTANCE_RESPONSE | jq -r '.id // empty')

if [ -z "$INSTANCE_ID" ] || [ "$INSTANCE_ID" = "null" ]; then
  echo -e "${YELLOW}⚠ Note: Managed instance creation may need admin setup${NC}"
  echo "Response: $INSTANCE_RESPONSE"
else
  echo -e "${GREEN}✓ Managed instance created (ID: $INSTANCE_ID)${NC}"
fi
echo ""

# Summary
echo -e "${GREEN}=================================="
echo "Demo Setup Complete! ✓"
echo "==================================${NC}"
echo ""
echo -e "📋 ${BLUE}Demo Credentials:${NC}"
echo "  Username: demo"
echo "  Email: demo@pganalytics.local"
echo "  Password: Demo@12345"
echo ""
echo -e "🔌 ${BLUE}Collector Details:${NC}"
echo "  ID: $COLLECTOR_ID"
echo "  Hostname: demo-collector.pganalytics.local"
echo "  Environment: demo"
echo ""
echo -e "🗄️  ${BLUE}Managed Instance Details:${NC}"
echo "  Hostname: demo-db.pganalytics.local"
echo "  Port: 5432"
echo ""
echo -e "🌐 ${BLUE}Services:${NC}"
echo "  Frontend: http://localhost:$FRONTEND_PORT"
echo "  Backend: http://localhost:$BACKEND_PORT"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Run: npm run dev (from frontend directory)"
echo "  2. Open browser to http://localhost:$FRONTEND_PORT"
echo "  3. Login with demo credentials"
echo "  4. View registered collector and managed instance"
echo ""
