# pgAnalytics Collector Registration UI

**Purpose:** User-friendly interface for registering and managing collectors on central RDS backend
**Status:** Design & Requirements
**Date:** February 26, 2026
**Version:** 1.0

---

## Overview

The collector registration interface enables operators to:
- Register new database collectors
- Manage collector credentials and configuration
- Monitor collector health and status
- View metrics collection statistics
- Bulk operations (register multiple collectors)
- Collector grouping and organization

---

## Architecture Decision

### Option 1: React-Based Web UI (Recommended)
**Pros:**
- Modern, responsive interface
- Real-time updates with WebSocket/SSE
- Rich dashboards and visualizations
- Better UX for complex workflows

**Cons:**
- Requires Node.js build toolchain
- Additional frontend deployment
- Larger attack surface

### Option 2: Server-Side Rendered (Simpler)
**Pros:**
- Single binary deployment
- No JavaScript build process
- Traditional server-side sessions

**Cons:**
- Less responsive
- Limited real-time capabilities
- Less modern UX

### Recommendation
**React-based UI** - Better aligns with modern development practices and provides superior UX for fleet management.

---

## Project Structure

```
pganalytics-v3/
├── backend/                          # Go backend (API)
│   ├── internal/api/handlers.go     # REST endpoints
│   ├── internal/auth/               # Authentication
│   └── ...
│
├── frontend/                         # React frontend (NEW)
│   ├── src/
│   │   ├── components/
│   │   │   ├── CollectorRegistration/
│   │   │   │   ├── index.tsx
│   │   │   │   ├── RegistrationForm.tsx
│   │   │   │   ├── BulkUpload.tsx
│   │   │   │   ├── styles.module.css
│   │   │   │   └── types.ts
│   │   │   ├── CollectorList/
│   │   │   │   ├── index.tsx
│   │   │   ├── CollectorDetail/
│   │   │   │   ├── index.tsx
│   │   │   ├── CollectorGroups/
│   │   │   │   ├── index.tsx
│   │   │   └── Common/
│   │   │       ├── Header.tsx
│   │   │       ├── Sidebar.tsx
│   │   │       └── ...
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Collectors.tsx
│   │   │   ├── Settings.tsx
│   │   │   └── ...
│   │   ├── services/
│   │   │   ├── api.ts
│   │   │   ├── collectors.ts
│   │   │   └── auth.ts
│   │   ├── hooks/
│   │   │   ├── useCollectors.ts
│   │   │   ├── useAuth.ts
│   │   │   └── ...
│   │   ├── store/                   # Redux or Context API
│   │   │   ├── actions/
│   │   │   ├── reducers/
│   │   │   └── ...
│   │   ├── App.tsx
│   │   ├── index.tsx
│   │   └── ...
│   ├── public/
│   │   ├── index.html
│   │   ├── favicon.ico
│   │   └── ...
│   ├── package.json
│   ├── tsconfig.json
│   ├── webpack.config.js (or vite.config.js)
│   └── .env.example
│
├── docker-compose.yml                # Updated with frontend service
├── Dockerfile.frontend              # Multi-stage build
└── ...
```

---

## UI Mockups & Components

### 1. Collector Registration Form

**Location:** `/collectors/register`

**Components:**
```
┌────────────────────────────────────────────────────┐
│ Register New Collector                          [X] │
├────────────────────────────────────────────────────┤
│                                                    │
│  Collector Information                             │
│  ┌──────────────────────────────────────────────┐ │
│  │ Collector Name *              [Input field] │ │
│  │                                              │ │
│  │ Collector Type *              [Dropdown ▼]  │ │
│  │  ├─ PostgreSQL C++                           │ │
│  │  ├─ MySQL                                    │ │
│  │  └─ MongoDB                                  │ │
│  │                                              │ │
│  │ Environment *                 [Dropdown ▼]  │ │
│  │  ├─ Development                              │ │
│  │  ├─ Staging                                  │ │
│  │  └─ Production                               │ │
│  │                                              │ │
│  │ Collector Group               [Input field]  │ │
│  │  (e.g., "AWS", "On-Prem", "RDS")             │ │
│  │                                              │ │
│  │ Description                   [Textarea]     │ │
│  │  [................................................] │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
│  Database Connection                               │
│  ┌──────────────────────────────────────────────┐ │
│  │ Host *                        [Input field]  │ │
│  │  pganalytics-db.region.rds.amazonaws.com     │ │
│  │                                              │ │
│  │ Port *                        [Input: 5432]  │ │
│  │                                              │ │
│  │ Database Name *               [Input field]  │ │
│  │  pganalytics                                 │ │
│  │                                              │ │
│  │ Username *                    [Input field]  │ │
│  │  postgres                                    │ │
│  │                                              │ │
│  │ Password *                    [Input field]  │ │
│  │  [••••••••••] [Show]                        │ │
│  │                                              │ │
│  │ SSL Mode                      [Dropdown ▼]  │ │
│  │  ├─ Disable                                  │ │
│  │  ├─ Allow                                    │ │
│  │  ├─ Prefer                                   │ │
│  │  ├─ Require                                  │ │
│  │  ├─ Require (CA cert)                        │ │
│  │  └─ Require (Full verification)              │ │
│  │                                              │ │
│  │ SSL Certificate               [File upload]  │ │
│  │  [Choose file...]                            │ │
│  │                                              │ │
│  │ [ ] Test Connection          [Button: Test]  │ │
│  │  Connection successful! ✓                    │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
│  Advanced Options                                  │
│  [+] Advanced Configuration                       │
│      ├─ Collection Interval: 60 seconds          │
│      ├─ Query Limit per DB: 100                  │
│      ├─ Enable TLS: [Toggle]                     │
│      └─ Custom Tags: [key1=value1, ...]          │
│                                                    │
│  ┌──────────────────────────────────────────────┐ │
│  │ [ Cancel ]              [ Register Collector] │ │
│  └──────────────────────────────────────────────┘ │
│                                                    │
└────────────────────────────────────────────────────┘
```

### 2. Collector List & Management

**Location:** `/collectors`

**Components:**
```
┌────────────────────────────────────────────────────┐
│ Database Collectors                             [+] │
├────────────────────────────────────────────────────┤
│                                                    │
│ [Search collectors...]  [Filter ▼] [Sort ▼]      │
│                                                    │
│ ┌─────────────────────────────────────────────┐  │
│ │ # │ Name       │ Database   │ Env  │ Status │  │
│ ├─────────────────────────────────────────────┤  │
│ │[•]│ prod-rds-1 │ pganalytics│ Prod │ ✓ OK   │  │
│ │[ ]│ staging-db │ pganalytics│ Stg  │ ⚠ Slow │  │
│ │[ ]│ dev-local  │ pganalytics│ Dev  │ ✗ Down │  │
│ │[ ]│ prod-rds-2 │ pganalytics│ Prod │ ✓ OK   │  │
│ └─────────────────────────────────────────────┘  │
│                                                    │
│ Actions: [Edit] [Copy Config] [Restart] [Delete]  │
│                                                    │
│ Metrics Collected: 1,234,567                      │
│ Last Collection: 2 seconds ago                     │
│ Collection Success Rate: 99.8%                     │
│                                                    │
└────────────────────────────────────────────────────┘
```

### 3. Collector Details & Status

**Location:** `/collectors/{collectorId}`

**Components:**
```
┌────────────────────────────────────────────────────┐
│ Collector: prod-rds-1                    [Edit][•••] │
├────────────────────────────────────────────────────┤
│                                                    │
│ Status: ✓ Active                                   │
│ Last Heartbeat: 2 seconds ago                      │
│ Registered: 2024-01-15 10:30 AM                    │
│                                                    │
│ Connection Details                                 │
│  Host: pganalytics-db.region.rds.amazonaws.com    │
│  Port: 5432                                        │
│  Database: pganalytics                             │
│  Version: PostgreSQL 15.2                          │
│  Uptime: 99.8% (last 30 days)                      │
│                                                    │
│ Collection Metrics                                 │
│  ┌─────────────────────────────────────────────┐  │
│  │ Metrics Collected: 5,234,123                │  │
│  │ Collection Frequency: Every 60 seconds      │  │
│  │ Avg Collection Time: 234 ms                 │  │
│  │ Success Rate: 99.98%                        │  │
│  │ Last Error: (none)                          │  │
│  └─────────────────────────────────────────────┘  │
│                                                    │
│ Assigned Dashboards                                │
│  [✓] Overview Dashboard                            │
│  [✓] Performance Analysis                          │
│  [✓] Query Stats                                   │
│  [ ] Anomaly Detection (available)                 │
│                                                    │
│ Configuration                                      │
│  Query Limit: 100/database                         │
│  Collection Interval: 60 seconds                   │
│  Auto-restart on failure: Enabled                  │
│  Tags: [prod, aws, rds]                            │
│                                                    │
│ ┌─────────────────────────────────────────────┐  │
│ │ [Test Connection] [Restart] [Delete]        │  │
│ └─────────────────────────────────────────────┘  │
│                                                    │
│ Recent Activity                                    │
│  2024-01-20 10:30:45 - Metrics collected (1234)   │
│  2024-01-20 10:29:45 - Metrics collected (1198)   │
│  2024-01-20 10:28:45 - Metrics collected (1256)   │
│  [View all activity logs...]                      │
│                                                    │
└────────────────────────────────────────────────────┘
```

### 4. Bulk Registration

**Location:** `/collectors/bulk-import`

**Components:**
```
┌────────────────────────────────────────────────────┐
│ Bulk Collector Import                          [X] │
├────────────────────────────────────────────────────┤
│                                                    │
│ Upload CSV File                                    │
│ ┌──────────────────────────────────────────────┐  │
│ │ Drag file here or click to browse            │  │
│ │ [Supported: CSV, JSON]                       │  │
│ │ [Choose file...]                             │  │
│ └──────────────────────────────────────────────┘  │
│                                                    │
│ CSV Format:                                        │
│  name,host,port,database,username,password,env   │
│  prod-rds-1,pganalytics-db.region.rds.aws,5432... │
│  staging-db,staging-db.region.rds.aws,5432...    │
│                                                    │
│ Or paste JSON:                                     │
│ ┌──────────────────────────────────────────────┐  │
│ │ [                                            │  │
│ │   {                                          │  │
│ │     "name": "prod-rds-1",                   │  │
│ │     "host": "pganalytics-db.region.rds...",│  │
│ │     "port": 5432,                           │  │
│ │     "database": "pganalytics",              │  │
│ │     "environment": "production"              │  │
│ │   }                                          │  │
│ │ ]                                            │  │
│ └──────────────────────────────────────────────┘  │
│                                                    │
│ Import Results:                                    │
│  Total Rows: 5                                     │
│  Valid: 4 ✓                                        │
│  Invalid: 1 ✗                                      │
│  Row 3: Missing password                          │
│                                                    │
│ ┌──────────────────────────────────────────────┐  │
│ │ [ Cancel ]              [ Import 4 Collectors] │ │
│ └──────────────────────────────────────────────┘  │
│                                                    │
└────────────────────────────────────────────────────┘
```

---

## API Endpoints (Backend)

### Collector Management

```
POST   /api/v1/collectors/register
GET    /api/v1/collectors
GET    /api/v1/collectors/{collectorId}
PUT    /api/v1/collectors/{collectorId}
DELETE /api/v1/collectors/{collectorId}
POST   /api/v1/collectors/{collectorId}/test-connection
POST   /api/v1/collectors/{collectorId}/restart
GET    /api/v1/collectors/{collectorId}/metrics
GET    /api/v1/collectors/{collectorId}/status
POST   /api/v1/collectors/bulk-import
POST   /api/v1/collectors/{collectorId}/assign-dashboards
```

### Request Examples

**Register Collector**
```bash
POST /api/v1/collectors/register
Content-Type: application/json

{
  "name": "prod-rds-1",
  "type": "postgresql",
  "environment": "production",
  "group": "aws",
  "description": "Production RDS Database",
  "host": "pganalytics-db.region.rds.amazonaws.com",
  "port": 5432,
  "database": "pganalytics",
  "username": "postgres",
  "password": "encrypted_password_here",
  "ssl_mode": "require",
  "ssl_certificate": "base64_encoded_cert",
  "collection_interval": 60,
  "query_limit": 100,
  "tags": ["prod", "aws", "rds"]
}

Response:
{
  "id": "col_12345",
  "jwt_token": "eyJhbGc...",
  "status": "active",
  "created_at": "2024-01-20T10:30:00Z"
}
```

**Bulk Import**
```bash
POST /api/v1/collectors/bulk-import
Content-Type: application/json

{
  "collectors": [
    {
      "name": "prod-rds-1",
      "host": "pganalytics-db.region.rds.amazonaws.com",
      "port": 5432,
      "database": "pganalytics",
      "username": "postgres",
      "password": "encrypted_password_here",
      "environment": "production"
    },
    {
      "name": "staging-db",
      "host": "staging-db.region.rds.amazonaws.com",
      "port": 5432,
      "database": "pganalytics",
      "username": "postgres",
      "password": "encrypted_password_here",
      "environment": "staging"
    }
  ]
}

Response:
{
  "total": 2,
  "successful": 2,
  "failed": 0,
  "results": [
    {
      "name": "prod-rds-1",
      "collector_id": "col_12345",
      "status": "success"
    },
    {
      "name": "staging-db",
      "collector_id": "col_12346",
      "status": "success"
    }
  ]
}
```

---

## Technology Stack

### Frontend
- **Framework:** React 18+
- **Language:** TypeScript
- **State Management:** Redux Toolkit or Zustand
- **HTTP Client:** Axios or React Query
- **UI Components:** Material-UI (MUI) or Chakra UI
- **Forms:** React Hook Form + Zod validation
- **Styling:** CSS Modules or Tailwind CSS
- **Build Tool:** Webpack or Vite
- **Testing:** Jest + React Testing Library
- **E2E Testing:** Cypress or Playwright

### Backend (Existing)
- **Language:** Go
- **Framework:** Gin or Chi
- **Database:** PostgreSQL
- **Authentication:** JWT
- **ORM:** sqlc or GORM (optional)

---

## Development Phases

### Phase 1: Backend Collector API (10-15 hours)
- [ ] Create collector registration endpoint
- [ ] Implement collector CRUD operations
- [ ] Add database connection testing
- [ ] Create collector status monitoring
- [ ] Implement bulk import endpoint
- [ ] Add authentication/authorization

### Phase 2: Frontend Setup & Components (20-25 hours)
- [ ] Create React project structure
- [ ] Set up routing
- [ ] Build layout components (Header, Sidebar)
- [ ] Create registration form component
- [ ] Build collector list component
- [ ] Implement form validation

### Phase 3: Integration & Features (15-20 hours)
- [ ] Connect API endpoints
- [ ] Implement state management
- [ ] Add real-time status updates (WebSocket)
- [ ] Build bulk import UI
- [ ] Implement error handling
- [ ] Add loading states and pagination

### Phase 4: Testing & Polish (10-15 hours)
- [ ] Unit tests for components
- [ ] Integration tests
- [ ] E2E tests
- [ ] Performance optimization
- [ ] Documentation
- [ ] Accessibility review

**Total Estimated Time:** 55-75 hours

---

## Deployment Architecture

### Development
```
localhost:3000 → React Dev Server
              ↓
localhost:8080 → Go Backend API
              ↓
localhost:5432 → PostgreSQL
```

### Production
```
pganalytics.example.com → Nginx (reverse proxy)
                        ↓
                   Frontend (React SPA)
                   Backend API (Go)
                        ↓
                   RDS PostgreSQL
```

### Docker Setup

```dockerfile
# Dockerfile.frontend
FROM node:18-alpine AS builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]
```

### docker-compose.yml Update
```yaml
services:
  frontend:
    build:
      context: .
      dockerfile: Dockerfile.frontend
    ports:
      - "3000:3000"
    environment:
      - REACT_APP_API_URL=http://localhost:8080
    depends_on:
      - backend

  backend:
    build:
      context: .
      dockerfile: Dockerfile.backend
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/pganalytics
      - JWT_SECRET=your_secret_key
    depends_on:
      - postgres

  postgres:
    image: postgres:17
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=pganalytics
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

---

## Security Considerations

1. **Password Storage**
   - Encrypt passwords before storing in database
   - Use AES-256-GCM encryption (from Week 3)
   - Never return plaintext passwords to frontend

2. **Database Connection Testing**
   - Test connection before registering collector
   - Don't expose connection errors to users
   - Implement rate limiting on test-connection endpoint

3. **SSL Certificates**
   - Support certificate upload for RDS/cloud databases
   - Validate certificate before use
   - Store securely with encryption

4. **API Authentication**
   - All collector registration endpoints require authentication
   - Use JWT tokens for API access
   - Implement RBAC (admin can register, users can view)

5. **Audit Logging**
   - Log all collector registrations
   - Track configuration changes
   - Monitor access patterns

---

## Success Criteria

✅ UI matches mockups
✅ Collector registration completes in <2 minutes
✅ Bulk import supports 100+ collectors
✅ Real-time status updates (WebSocket)
✅ Mobile responsive design
✅ <100ms response times on listing
✅ All forms validate before submission
✅ Clear error messages for failures
✅ Unit test coverage >80%
✅ Accessibility WCAG AA compliant

---

## File References & Implementation Guide

### Backend Implementation (Go)

**File:** `backend/internal/api/handlers_collectors.go`
```go
// RegisterCollectorRequest represents collector registration
type RegisterCollectorRequest struct {
    Name                 string `json:"name" binding:"required,max=255"`
    Type                 string `json:"type" binding:"required,oneof=postgresql mysql mongodb"`
    Environment          string `json:"environment" binding:"required,oneof=development staging production"`
    Group                string `json:"group" binding:"max=100"`
    Description          string `json:"description"`
    Host                 string `json:"host" binding:"required"`
    Port                 int    `json:"port" binding:"required,min=1,max=65535"`
    Database             string `json:"database" binding:"required"`
    Username             string `json:"username" binding:"required"`
    Password             string `json:"password" binding:"required"`
    SSLMode              string `json:"ssl_mode" binding:"oneof=disable allow prefer require requireCA requireFull"`
    SSLCertificate       string `json:"ssl_certificate"` // base64 encoded
    CollectionInterval   int    `json:"collection_interval" binding:"min=10,max=3600"`
    QueryLimit           int    `json:"query_limit" binding:"min=1,max=10000"`
    Tags                 []string `json:"tags"`
}

func (s *Server) RegisterCollector(c *gin.Context) {
    var req RegisterCollectorRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Validate database connection
    db, err := sql.Open("postgres", fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        req.Host, req.Port, req.Username, req.Password, req.Database,
        req.SSLMode,
    ))
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid database connection"})
        return
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        c.JSON(400, gin.H{"error": "Cannot connect to database"})
        return
    }

    // Create collector record
    collectorID := "col_" + uuid.New().String()
    jwtToken := generateCollectorToken(collectorID)

    // Save collector to database with encrypted password
    // ... implementation

    c.JSON(201, gin.H{
        "id":        collectorID,
        "jwt_token": jwtToken,
        "status":    "active",
        "created_at": time.Now(),
    })
}

func (s *Server) BulkImportCollectors(c *gin.Context) {
    // ... bulk import implementation
}
```

### Frontend Implementation (React/TypeScript)

**File:** `frontend/src/components/CollectorRegistration/index.tsx`
```typescript
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { collectorApi } from '../../services/api';

const collectorSchema = z.object({
    name: z.string().min(1).max(255),
    type: z.enum(['postgresql', 'mysql', 'mongodb']),
    environment: z.enum(['development', 'staging', 'production']),
    group: z.string().optional(),
    host: z.string().min(1),
    port: z.number().min(1).max(65535),
    database: z.string().min(1),
    username: z.string().min(1),
    password: z.string().min(1),
    sslMode: z.enum(['disable', 'allow', 'prefer', 'require', 'requireCA', 'requireFull']),
});

type CollectorFormData = z.infer<typeof collectorSchema>;

export const CollectorRegistration: React.FC = () => {
    const [isLoading, setIsLoading] = useState(false);
    const {
        register,
        handleSubmit,
        formState: { errors },
        reset,
    } = useForm<CollectorFormData>({
        resolver: zodResolver(collectorSchema),
    });

    const onSubmit = async (data: CollectorFormData) => {
        setIsLoading(true);
        try {
            const response = await collectorApi.registerCollector(data);
            // Show success message
            // Redirect to collector list
            reset();
        } catch (error) {
            // Show error message
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <form onSubmit={handleSubmit(onSubmit)}>
            <input {...register('name')} placeholder="Collector Name" />
            {errors.name && <span>{errors.name.message}</span>}

            <select {...register('type')}>
                <option value="postgresql">PostgreSQL</option>
                <option value="mysql">MySQL</option>
                <option value="mongodb">MongoDB</option>
            </select>

            {/* More form fields... */}

            <button type="submit" disabled={isLoading}>
                {isLoading ? 'Registering...' : 'Register Collector'}
            </button>
        </form>
    );
};
```

---

## Next Steps

1. **Approve Design** - Review mockups and make any adjustments
2. **Backend API Implementation** - Create REST endpoints (10-15h)
3. **Frontend Project Setup** - Initialize React project (5h)
4. **Component Development** - Build UI components (20-25h)
5. **Integration Testing** - Connect frontend to backend (15-20h)
6. **Deployment** - Docker setup and testing (10h)

---

**Status:** Ready for approval and implementation
**Estimated Total Time:** 55-75 hours
**Priority:** High (Needed for collector fleet management)

