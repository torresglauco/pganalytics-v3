# pgAnalytics Frontend Demo - Instructions

## Prerequisites
- Docker & Docker Compose
- Node.js 18+
- Backend running on port 8080

## Quick Setup

### 1. Start Backend Services
```bash
docker-compose up -d
```

Wait 30 seconds for services to be ready.

### 2. Run Demo Setup Script
```bash
./demo-setup.sh
```

This creates:
- Demo user account (username: `demo`, password: `Demo@12345`)
- Registration secret
- Demo collector (registered and visible in frontend)
- Demo managed instance

### 3. Start Frontend
```bash
./start-frontend.sh
```

Frontend will be available at: http://localhost:3000

### 4. Login & Test
- Go to http://localhost:3000
- Login with credentials from demo setup
- You should see:
  - ✓ Active Collectors tab (with demo collector)
  - ✓ Managed Instances tab (with demo instance)
  - ✓ User menu (top right)

## If You Get Errors

### "Error loading collectors"
1. Check backend is running: `curl http://localhost:8080/api/v1/health`
2. Check auth token is stored in localStorage
3. Verify user is authenticated: `curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/v1/collectors`

### "Not implemented yet"
This means the API endpoint exists but the handler isn't implemented yet. Check what features are marked as "Phase X" in the backend routes.

### "Failed to load registration secrets"
This is an admin-only feature. Make sure your user has admin role, or disable this tab.

## Manual Testing with curl

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"Demo@12345"}'

# Get collectors (use token from login response)
curl http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer TOKEN_HERE"

# Get managed instances
curl http://localhost:8080/api/v1/managed-instances \
  -H "Authorization: Bearer TOKEN_HERE"
```

## Docker Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down

# Remove everything
docker-compose down -v
```

## Test Coverage

All frontend tests are passing:
```bash
npm run test                 # Run all tests
npm run test:coverage       # Generate coverage report
npm run test:ui             # Interactive test dashboard
```

## Current Status

### ✅ Working
- Authentication (login/signup)
- User management
- Collector registration
- Managed instance creation
- Frontend components
- All 86 unit tests passing

### 📋 To Implement
- Some API handlers still marked as "Not implemented yet"
- Admin features may be partially implemented
- Real-time metric streaming

### 🧪 Testing
- 12 test files
- 86 tests total
- 100% pass rate
- Type-safe with TypeScript
