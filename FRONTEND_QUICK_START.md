# Frontend Quick Start Guide

## How to Run the Collector Registration UI

### Step 1: Verify Backend is Running

Make sure the Go backend API is running:

```bash
# From project root
cd backend
go run ./cmd/pganalytics-api/main.go

# Expected output:
# INFO: Server listening on :8080
# INFO: PostgreSQL connected
```

Get the **Registration Secret** from backend environment:

```bash
echo $REGISTRATION_SECRET
# If not set, set it:
export REGISTRATION_SECRET="your-secret-123"
```

### Step 2: Install Frontend Dependencies

```bash
cd frontend
npm install
```

### Step 3: Configure Environment

```bash
# Copy example config
cp .env.example .env

# Edit .env if needed (default localhost:8080 is fine for dev)
# VITE_API_URL=http://localhost:8080/api/v1
```

### Step 4: Start Development Server

```bash
npm run dev
```

**Output:**
```
  VITE v5.0.8  dev server running at:

  ➜  Local:   http://localhost:3000/
  ➜  press h + enter to show help
```

### Step 5: Open Browser

Visit: **http://localhost:3000**

---

## Using the UI

### Register a Collector

1. **Enter Registration Secret**
   - Top of page: "Registration Secret Required"
   - Paste the secret you got from backend
   - Click "Show" to verify it's correct

2. **Go to "Register Collector" Tab**
   - Enter hostname (e.g., `localhost:5432`)
   - Click "Test" to verify connectivity
   - Optionally fill: Environment, Group, Description
   - Click "Register Collector"

3. **Copy Your Token**
   - Success message shows `collector_id` and `jwt_token`
   - Copy these for your collector configuration
   - Set environment variables:
     ```bash
     export COLLECTOR_ID=col_xxx
     export COLLECTOR_JWT_TOKEN=eyJhbGc...
     export CENTRAL_API_URL=http://localhost:8080
     ```

### View Registered Collectors

1. Go to "Manage Collectors" Tab
2. See list of all registered collectors
3. Check status (active/inactive/error)
4. View last heartbeat and metrics count
5. Click delete (trash icon) to remove

---

## Troubleshooting

### "Cannot POST /api/v1/collectors/register"

**Problem:** Backend API not accessible

**Solution:**
```bash
# Check backend is running
curl http://localhost:8080/api/v1/health

# Should return:
# {"status":"ok","version":"3.0.0-alpha",...}
```

### "Invalid or missing registration secret"

**Problem:** Wrong secret provided

**Solution:**
```bash
# Check backend environment
echo $REGISTRATION_SECRET

# Should output your secret
# If empty, set it:
export REGISTRATION_SECRET="your-secret"

# Restart backend:
# Kill current process (Ctrl+C)
# Run again: go run ./cmd/pganalytics-api/main.go
```

### "Connection failed"

**Problem:** Cannot connect to database

**Solution:**
- Verify hostname is correct (e.g., `localhost:5432`)
- Check PostgreSQL is running
- Check credentials if authentication required
- See backend logs for details

### "Cannot GET /api/v1/collectors"

**Problem:** Not authenticated (missing token)

**Solution:**
- First register a collector via the form
- Token is shown in success message
- It's automatically used for subsequent requests

### Page shows blank/white screen

**Problem:** Frontend build issue

**Solution:**
```bash
# Clear cache and reinstall
rm -rf node_modules dist
npm install
npm run dev
```

---

## Development Workflow

### Make Changes

Edit files in `src/`:
- Components in `src/components/`
- Pages in `src/pages/`
- Styles in `src/styles/`
- API calls in `src/services/`

### Hot Reload

Changes are automatically reflected in browser (HMR enabled)

### Type Checking

```bash
npm run type-check
```

### Linting

```bash
npm run lint
```

---

## Building for Production

```bash
# Create optimized build
npm run build

# Build output in dist/ folder
ls -la dist/

# Test production build locally
npm run preview
```

---

## Docker Deployment

```bash
# Build image
docker build -f Dockerfile -t pganalytics-ui:latest .

# Run container
docker run -p 3000:3000 \
  -e VITE_API_URL=http://backend:8080/api/v1 \
  pganalytics-ui:latest

# With docker-compose
docker-compose up frontend
```

---

## API Endpoints Used

The UI communicates with these backend endpoints:

```
POST   /api/v1/collectors/register        → Register collector
GET    /api/v1/collectors                 → List collectors
DELETE /api/v1/collectors/{id}            → Delete collector
POST   /api/v1/collectors/test-connection → Test connection
```

All endpoints require:
- `Authorization: Bearer <token>` header (except registration)
- `X-Registration-Secret: <secret>` header (for registration)

---

## Key Files

| File | Purpose |
|------|---------|
| `src/App.tsx` | Root component |
| `src/pages/Dashboard.tsx` | Main interface |
| `src/components/CollectorForm.tsx` | Registration form |
| `src/components/CollectorList.tsx` | Collectors table |
| `src/services/api.ts` | API client |
| `src/hooks/useCollectors.ts` | Data fetching |
| `src/types/index.ts` | Type definitions |
| `vite.config.ts` | Build configuration |
| `tailwind.config.js` | Styling config |

---

## Next Steps

1. ✅ Register a test collector
2. ✅ Verify it shows in the list
3. ✅ Integrate with your database
4. ✅ Configure your C++ collector with the token
5. ✅ Monitor metrics in Grafana

---

## See Also

- `FRONTEND_IMPLEMENTATION.md` - Detailed documentation
- `backend/README.md` - Backend API documentation
- `COLLECTOR_REGISTRATION_UI.md` - UI design specs

---

**Happy Collecting! 🚀**
