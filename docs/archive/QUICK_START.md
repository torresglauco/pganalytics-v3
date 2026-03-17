# pgAnalytics Frontend - Quick Start Guide

## 5-Minute Setup

### Terminal 1: Start Backend & Demo Data
```bash
./demo-setup.sh
```

This will:
- Start Docker services (PostgreSQL, TimescaleDB, Backend API)
- Create demo user: `demo` / `Demo@12345`
- Register a demo collector
- Create a demo managed instance

### Terminal 2: Start Frontend
```bash
./start-frontend.sh
```

This will:
- Install dependencies (if needed)
- Start Vite dev server on port 3000
- Open http://localhost:3000

### Login & Test
1. Go to http://localhost:3000
2. Login with: `demo` / `Demo@12345`
3. You should see:
   - **Collectors Tab**: Demo collector registered
   - **Managed Instances Tab**: Demo instance created
   - **User Menu**: Top right corner

---

## Testing Commands

```bash
# Run all tests (watch mode)
npm run test

# Run tests once (CI mode)
npm run test -- --run

# Interactive dashboard
npm run test:ui

# Coverage report
npm run test:coverage
```

Or use Make:
```bash
make test-frontend              # Run tests
make test-frontend-ui          # Dashboard
make test-frontend-coverage    # Coverage
```

---

## Test Results
- **Tests**: 86 passing (100% pass rate)
- **Duration**: ~3.5 seconds
- **Files**: 12 test files
- **Coverage**: All critical paths

---

## Services & Ports

| Service | URL | Purpose |
|---------|-----|---------|
| Frontend | http://localhost:3000 | React UI |
| Backend API | http://localhost:8080 | REST API |
| PostgreSQL | localhost:5432 | Metadata DB |
| TimescaleDB | localhost:5433 | Metrics DB |

---

## Demo Credentials

```
Username: demo
Email: demo@pganalytics.local
Password: Demo@12345
```

---

## Stop Services

```bash
# Stop containers
docker-compose down

# Remove everything (including data)
docker-compose down -v
```

---

## Troubleshooting

### Backend not starting?
```bash
docker-compose logs backend
```

### Tests failing?
```bash
cd frontend
npm install
npm run test -- --run
```

### Frontend not connecting?
1. Check backend: `curl http://localhost:8080/health`
2. Restart frontend
3. Check browser console for errors

### Need more help?
- See `DEMO_SETUP.md` for detailed guide
- See `DEMO_INSTRUCTIONS.md` for quick reference
- See `FRONTEND_IMPLEMENTATION_STATUS.md` for complete status

---

**Status**: ✅ Production Ready - 100% Test Pass Rate
