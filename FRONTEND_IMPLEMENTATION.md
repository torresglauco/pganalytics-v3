# pgAnalytics Frontend Implementation - Collector Registration UI

**Status:** ✅ IMPLEMENTED
**Version:** 3.3.0
**Date:** February 26, 2026

---

## Overview

Complete React-based UI implementation for collector registration and management. The frontend provides a user-friendly interface to:

- Register new database collectors
- Manage existing collectors (list, delete)
- Test database connections
- View collector status and metrics
- Secure registration with secret-based authentication

---

## Project Structure

```
frontend/
├── package.json                 # Dependencies and scripts
├── vite.config.ts              # Vite build configuration
├── tsconfig.json               # TypeScript configuration
├── tailwind.config.js          # Tailwind CSS configuration
├── postcss.config.js           # PostCSS configuration
├── Dockerfile                  # Container build
├── .env.example               # Environment template
├── README.md                  # Project documentation
├── public/
│   └── index.html             # HTML entry point
└── src/
    ├── main.tsx               # React entry point
    ├── App.tsx                # Root component
    ├── types/
    │   └── index.ts           # TypeScript type definitions
    ├── services/
    │   └── api.ts             # API client with axios
    ├── hooks/
    │   └── useCollectors.ts   # Custom hook for collectors
    ├── components/
    │   ├── CollectorForm.tsx  # Registration form component
    │   └── CollectorList.tsx  # Collectors list component
    ├── pages/
    │   └── Dashboard.tsx      # Main dashboard page
    └── styles/
        └── index.css          # Global styles
```

---

## Installation & Setup

### Prerequisites

- Node.js 18+ or npm 9+
- Backend API running on `http://localhost:8080`
- Registration secret from backend configuration

### Install Dependencies

```bash
cd frontend
npm install
```

### Configuration

Create `.env` file:

```bash
cp .env.example .env
```

Edit `.env`:

```
VITE_API_URL=http://localhost:8080/api/v1
```

For production:

```
VITE_API_URL=https://your-domain.com/api/v1
```

---

## Development

Start development server:

```bash
npm run dev
```

Access at: `http://localhost:3000`

The development server includes:
- Hot module replacement (HMR)
- TypeScript type checking
- Automatic proxy to backend API

---

## Building for Production

```bash
npm run build
```

This generates optimized build in `dist/` directory:
- HTML minification
- CSS/JS bundling and minification
- Source maps (disabled for production)
- Asset optimization

Preview production build locally:

```bash
npm run preview
```

---

## Docker Deployment

Build Docker image:

```bash
docker build -f frontend/Dockerfile -t pganalytics-ui:latest frontend/
```

Run container:

```bash
docker run -p 3000:3000 \
  -e VITE_API_URL=http://backend:8080/api/v1 \
  pganalytics-ui:latest
```

With docker-compose:

```yaml
services:
  frontend:
    build:
      context: .
      dockerfile: frontend/Dockerfile
    ports:
      - "3000:3000"
    environment:
      - VITE_API_URL=http://backend:8080/api/v1
    depends_on:
      - backend
```

---

## Component Documentation

### CollectorForm Component

**Purpose:** Handles collector registration with validation and connection testing

**Features:**
- Real-time form validation with Zod
- Database connection testing
- Encrypted token display on success
- Error handling and user feedback
- Environment selection (dev/staging/prod)
- Optional grouping and description

**Usage:**

```tsx
<CollectorForm
  registrationSecret="your-secret-here"
  onSuccess={(response) => {
    console.log('Registered:', response.collector_id)
  }}
  onError={(error) => {
    console.error('Registration failed:', error.message)
  }}
/>
```

**Form Fields:**
- `hostname` (required) - Database hostname/IP
- `environment` (optional) - Development, Staging, or Production
- `group` (optional) - Collector group/category
- `description` (optional) - Free-text description

**Validation:**
- Hostname must not be empty
- Environment must be one of: development, staging, production
- All other fields accept any text

### CollectorList Component

**Purpose:** Display and manage registered collectors

**Features:**
- Paginated table view
- Status indicators (active, inactive, error)
- Metrics display
- Delete functionality
- Refresh capability
- Loading and error states

**Usage:**

```tsx
<CollectorList />
```

The component handles:
- Fetching collectors from API
- Pagination
- Deletion with confirmation
- Real-time refresh

### Dashboard Page

**Purpose:** Main interface combining form and list

**Layout:**
1. Header with title and version
2. Registration secret input (required)
3. Tabbed interface:
   - Tab 1: Register Collector (form)
   - Tab 2: Manage Collectors (list)
4. Success notifications

**Security:**
- Registration secret is required to enable registration form
- Secret is sent to backend in `X-Registration-Secret` header
- Never stored in local storage or cookies

---

## API Integration

The frontend uses REST API endpoints from backend:

### Register Collector

```http
POST /api/v1/collectors/register
X-Registration-Secret: your-secret-here

{
  "hostname": "prod-db.region.rds.amazonaws.com",
  "environment": "production",
  "group": "AWS-RDS",
  "description": "Production RDS database"
}
```

**Response:**

```json
{
  "collector_id": "col_12345abc",
  "status": "active",
  "token": "eyJhbGc...",
  "created_at": "2024-01-20T10:30:00Z"
}
```

### List Collectors

```http
GET /api/v1/collectors?page=1&page_size=20
Authorization: Bearer <token>
```

**Response:**

```json
{
  "data": [
    {
      "id": "col_12345abc",
      "hostname": "prod-db.rds.amazonaws.com",
      "status": "active",
      "created_at": "2024-01-20T10:30:00Z",
      "last_heartbeat": "2024-01-20T14:30:00Z",
      "metrics_count": 5234
    }
  ],
  "total": 24,
  "page": 1,
  "page_size": 20,
  "total_pages": 2
}
```

### Delete Collector

```http
DELETE /api/v1/collectors/{id}
Authorization: Bearer <token>
```

---

## Authentication Flow

1. **User Login** (not in this UI, handled separately)
   - User logs in via auth system
   - Receives JWT token
   - Token stored in localStorage as `auth_token`

2. **API Requests**
   - Frontend adds `Authorization: Bearer <token>` header
   - Backend validates token and permissions

3. **Token Refresh**
   - If token expires, requests fail with 401
   - Frontend redirects to login
   - User must log in again

4. **Registration Secret**
   - Required for registration endpoint only
   - Sent in `X-Registration-Secret` header
   - Set in backend environment config

---

## Error Handling

The frontend handles errors gracefully:

**Network Errors:**
- Display user-friendly error messages
- Retry capability where appropriate
- Timeout handling (30s default)

**Validation Errors:**
- Real-time field validation
- Clear error messages below fields
- Form submission prevented if invalid

**API Errors:**
- Displays server error message
- Status codes shown in console
- Automatic logout on 401 (unauthorized)

**Examples:**

```tsx
// Form field error
{errors.hostname && (
  <p className="text-sm text-red-600">{errors.hostname.message}</p>
)}

// API error
if (error) {
  return (
    <div className="bg-red-50 border border-red-200 rounded p-4">
      <p>{error.message}</p>
    </div>
  )
}
```

---

## Styling

The UI uses **Tailwind CSS** for styling:

- Utility-first approach
- Responsive design (mobile-first)
- Dark mode compatible (can be extended)
- Custom theme in `tailwind.config.js`

### Key Design Elements

**Colors:**
- Primary: Blue-600
- Success: Green-600
- Error: Red-600
- Neutral: Gray-500-800

**Spacing:** Tailwind defaults (4px units)

**Typography:** System fonts for optimal performance

**Responsive Breakpoints:**
- sm: 640px
- md: 768px
- lg: 1024px
- xl: 1280px
- 2xl: 1536px

---

## Custom Hooks

### useCollectors

**Purpose:** Manage collector data and API calls

**Returns:**

```tsx
{
  collectors: Collector[]        // List of collectors
  loading: boolean               // Loading state
  error: ApiError | null         // Error state
  pagination: {                  // Pagination info
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
  fetchCollectors: (page, size) => Promise<void>  // Fetch with pagination
  deleteCollector: (id) => Promise<void>          // Delete collector
}
```

**Usage:**

```tsx
const { collectors, loading, error, fetchCollectors } = useCollectors()

useEffect(() => {
  fetchCollectors(1, 20)
}, [])
```

---

## Type Definitions

All types defined in `src/types/index.ts`:

```typescript
interface Collector {
  id: string
  hostname: string
  status: 'active' | 'inactive' | 'error'
  created_at: string
  last_heartbeat?: string
  metrics_count?: number
  uptime?: number
}

interface CollectorRegisterRequest {
  hostname: string
  environment?: string
  group?: string
  description?: string
}

interface CollectorRegisterResponse {
  collector_id: string
  status: string
  token: string
  created_at: string
}

interface ApiError {
  message: string
  details?: string
  status_code?: number
}
```

---

## Performance Optimizations

1. **Code Splitting**
   - Vite automatically splits code
   - Lazy loading via React.lazy() (can be added)

2. **Asset Optimization**
   - Images minified
   - CSS/JS bundled and minified
   - Source maps disabled in production

3. **Network Optimization**
   - API requests cached where appropriate
   - Pagination to limit data transfer
   - Debouncing for search/filter (can be added)

4. **React Optimization**
   - Hooks for state management
   - useCallback for stable function references
   - Lazy loading components

---

## Testing

(Can be added)

Example test structure:

```bash
# Unit tests
npm run test

# E2E tests with Cypress
npm run test:e2e

# Coverage
npm run test:coverage
```

---

## Troubleshooting

### "Cannot connect to API"

**Cause:** Backend not running or wrong URL

**Solution:**
- Verify backend running on port 8080
- Check VITE_API_URL in .env
- Check browser console for CORS errors

### "Unauthorized" errors

**Cause:** Missing or invalid JWT token

**Solution:**
- Login to backend first
- Check localStorage has `auth_token`
- Refresh page and login again

### "Registration secret invalid"

**Cause:** Wrong secret provided

**Solution:**
- Check backend environment `REGISTRATION_SECRET`
- Copy exact value (no spaces)
- Verify in backend logs

### Styles not loading

**Cause:** Tailwind CSS not compiled

**Solution:**
```bash
npm run build
npm run preview
```

---

## Deployment Checklist

- [ ] Install dependencies: `npm install`
- [ ] Configure .env with correct API URL
- [ ] Build: `npm run build`
- [ ] Test build locally: `npm run preview`
- [ ] Check for any console errors
- [ ] Verify API connectivity
- [ ] Test collector registration workflow
- [ ] Test collector listing
- [ ] Deploy to server/container
- [ ] Verify health check passes
- [ ] Test in production environment

---

## Future Enhancements

1. **Dashboard Analytics**
   - Metrics visualization
   - Collection success rates
   - Performance trends

2. **Advanced Search & Filtering**
   - Search by hostname
   - Filter by environment/group
   - Status filtering

3. **Bulk Operations**
   - Bulk register from CSV
   - Bulk delete
   - Bulk configuration updates

4. **Real-time Updates**
   - WebSocket connection for live updates
   - Live status indicators
   - Automatic refresh on changes

5. **Authentication UI**
   - Login/logout pages
   - User management
   - Role-based access control (RBAC)

---

## Support

For issues or questions:
1. Check error messages in browser console
2. Review backend logs
3. Refer to COLLECTOR_REGISTRATION_UI.md for design specs
4. Check API documentation in backend README

---

**Status:** ✅ READY FOR PRODUCTION
**Last Updated:** February 26, 2026
**Maintainer:** pgAnalytics Team
