# pgAnalytics v3.3.0 - Frontend Demo Setup

This guide will help you set up a complete demo environment with a registered collector and managed instance to test the pgAnalytics frontend.

## Prerequisites

- Docker and Docker Compose installed
- Node.js 18+ installed
- npm or yarn
- `jq` for JSON parsing (for demo script)

## Quick Start (5 minutes)

### Step 1: Run the Demo Setup

```bash
./demo-setup.sh
```

This script will:
- ✓ Start all Docker services (PostgreSQL, TimescaleDB, Backend)
- ✓ Wait for the backend to be ready
- ✓ Create a test user account
- ✓ Generate registration secrets
- ✓ Register a demo collector
- ✓ Create a demo managed instance
- ✓ Display all credentials and details

### Step 2: Start the Frontend

In a new terminal window:

```bash
./start-frontend.sh
```

This will:
- ✓ Install dependencies if needed
- ✓ Start the Vite development server
- ✓ Open the frontend on http://localhost:3000

### Step 3: Login and Explore

1. Open your browser to: **http://localhost:3000**
2. Login with:
   - **Username**: `demo`
   - **Password**: `Demo@12345`

3. You should see:
   - ✓ **Collectors Tab**: Shows the registered demo collector
     - Hostname: `demo-collector.pganalytics.local`
     - Status: Active
     - Environment: demo

   - ✓ **Managed Instances Tab**: Shows the demo managed instance
     - Hostname: `demo-db.pganalytics.local`
     - Port: 5432
     - Database: postgres

## Services Overview

| Service | URL | Purpose |
|---------|-----|---------|
| Frontend | http://localhost:3000 | React UI for collector management |
| Backend API | http://localhost:8080 | REST API server |
| PostgreSQL | localhost:5432 | Metadata and monitoring data |
| TimescaleDB | localhost:5433 | Time-series metrics storage |

## Demo Credentials

```
Username: demo
Email: demo@pganalytics.local
Password: Demo@12345
```

## What to Test

### 1. Dashboard Navigation
- [ ] Switch between Collectors and Managed Instances tabs
- [ ] View list of registered collectors
- [ ] View list of managed instances
- [ ] Check status indicators

### 2. Collector Registration
- [ ] View collector details (hostname, status, environment)
- [ ] See collector ID and registration info
- [ ] Check heartbeat and uptime metrics
- [ ] Test connection button (if available)

### 3. Managed Instance Management
- [ ] View instance details
- [ ] See connection status
- [ ] Check monitoring metrics

### 4. User Menu
- [ ] Click on user menu (top right)
- [ ] View account information
- [ ] Access profile or settings
- [ ] Logout and verify redirect to login

### 5. Form Validations
- [ ] Try invalid inputs in forms
- [ ] Verify error messages appear
- [ ] Test form submissions

## Cleaning Up

To stop and remove all containers:

```bash
docker-compose down -v
```

To stop services but keep data:

```bash
docker-compose down
```

To view logs:

```bash
docker-compose logs -f backend
docker-compose logs -f postgres
```

## Troubleshooting

### Backend not responding

```bash
# Check if containers are running
docker-compose ps

# Check backend logs
docker-compose logs backend

# Restart backend
docker-compose restart backend
```

### Frontend not connecting to backend

The frontend automatically proxies requests to `http://localhost:8080/api/v1`

If you get CORS errors:
1. Verify backend is running: `curl http://localhost:8080/health`
2. Check frontend proxy in `vite.config.ts`
3. Restart frontend server

### Database issues

```bash
# Check PostgreSQL
docker-compose exec postgres psql -U postgres -d pganalytics -c "\dt"

# Check TimescaleDB
docker-compose exec timescale psql -U postgres -d metrics -c "\dt"
```

## API Testing

You can also test the backend directly:

```bash
# Get auth token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"Demo@12345"}'

# List collectors
curl -X GET http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# List managed instances
curl -X GET http://localhost:8080/api/v1/managed-instances \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## Running Tests

From the frontend directory:

```bash
# Run all tests
npm run test

# Run with coverage
npm run test:coverage

# Run with interactive UI
npm run test:ui
```

## Next Steps

After verifying the demo:

1. **Try Admin Features** (if user has admin role):
   - User management
   - System settings
   - Registration secrets management

2. **Explore API** (use curl or Postman):
   - Check available endpoints
   - Test different operations
   - Review error handling

3. **Review Code**:
   - Check frontend components in `frontend/src/components/`
   - Review API integration in `frontend/src/services/api.ts`
   - Examine types in `frontend/src/types/`

## Support

For issues or questions:
1. Check the logs: `docker-compose logs`
2. Verify all services are running: `docker-compose ps`
3. Check backend health: `curl http://localhost:8080/health`
4. Review test output: `npm run test`

---

**Enjoy testing pgAnalytics! 🚀**
