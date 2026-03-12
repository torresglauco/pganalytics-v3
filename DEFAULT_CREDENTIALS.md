# Default Credentials - pgAnalytics v3.3.0 Staging Deployment

**IMPORTANT:** These are default staging credentials. For production, change all passwords immediately after first login.

---

## Admin User (Created on First Deployment)

### Web Login (Frontend)
```
URL:              http://localhost:3000
Username:         admin
Email:            admin@pganalytics.local
Initial Password: PgAnalytics2026
```

**First Login Flow:**
1. Access http://localhost:3000
2. Login with username: `admin` and password: `PgAnalytics2026`
3. **REQUIRED:** Change password on first login
4. After password change, you will have full access to the dashboard

### API Login (Backend)
```
URL:              http://localhost:8080/api/v1/auth/login
Method:           POST
Content-Type:     application/json

Request:
{
  "username": "admin",
  "password": "PgAnalytics2026"
}

Response:
{
  "token": "JWT_TOKEN_HERE",
  "access_token": "...",
  "refresh_token": "...",
  "expires_at": "...",
  "user": {
    "id": 3,
    "username": "admin",
    "email": "admin@pganalytics.local",
    "role": "admin",
    "password_changed": false
  }
}
```

---

## Database Credentials

### PostgreSQL (Main Database)
```
Host:     postgres-staging (or localhost from host)
Port:     5432
Database: pganalytics_staging
Username: postgres
Password: staging_password
SSL Mode: disable (staging only)
```

**Connection String:**
```
postgres://postgres:staging_password@localhost:5432/pganalytics_staging?sslmode=disable
```

### TimescaleDB (Metrics Database)
```
Host:     timescale-staging (or localhost from host)
Port:     5433
Database: metrics_staging
Username: postgres
Password: staging_password
SSL Mode: disable (staging only)
```

**Connection String:**
```
postgres://postgres:staging_password@localhost:5433/metrics_staging?sslmode=disable
```

---

## Monitoring Tools

### Grafana
```
URL:      http://localhost:3001
Username: admin
Password: staging_admin
```

### Prometheus
```
URL: http://localhost:9090
(No authentication required)
```

---

## Application Configuration (Docker)

These are set in `docker-compose.staging.yml`:

```yaml
Backend Environment Variables:
  DATABASE_URL: postgres://postgres:staging_password@postgres-staging:5432/pganalytics_staging?sslmode=disable
  TIMESCALE_URL: postgres://postgres:staging_password@timescale-staging:5432/metrics_staging?sslmode=disable
  JWT_SECRET: staging-jwt-secret-change-in-production
  JWT_EXPIRATION: 3600
  ENCRYPTION_KEY: WkSMJvo2wKQ1FuceaE2yW2lEyxKIcJ1wfbrcNUOGUkE=
  REGISTRATION_SECRET: staging-registration-secret
  SETUP_ENDPOINT_ENABLED: true
  TLS_CERT: /etc/pganalytics/tls/server.crt
  TLS_KEY: /etc/pganalytics/tls/server.key
  PORT: 8080
  LOG_LEVEL: info
  ENVIRONMENT: staging
  MAX_CONNECTIONS: 100
  REQUEST_TIMEOUT: 30s
  RATE_LIMIT: 100/min

Frontend Environment Variables:
  REACT_APP_API_URL: https://localhost:8080
  REACT_APP_ENVIRONMENT: staging
  VITE_API_BACKEND_HOST: backend-staging
  VITE_API_BACKEND_PORT: 8080
  VITE_API_BACKEND_PROTOCOL: http
  PORT: 3000
```

---

## Quick Start Commands

### 1. Login via API and Get JWT Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "PgAnalytics2026"
  }'
```

### 2. Use JWT Token in API Calls
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X GET http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Access Frontend
```bash
# Open in browser
http://localhost:3000
```

### 4. Access Grafana Dashboards
```bash
# Open in browser
http://localhost:3001
```

### 5. Query PostgreSQL Directly
```bash
psql -h localhost -U postgres -d pganalytics_staging -c "SELECT * FROM pganalytics.users;"
```

---

## Security Notes

### ⚠️ Staging Passwords
These credentials are for **staging/development only**. They are:
- ✅ Documented for team use
- ✅ Set in version control as staging defaults
- ✅ Easily changed for new deployments
- ❌ NOT suitable for production

### 🔒 Production Checklist
Before deploying to production:

- [ ] Change all default passwords
- [ ] Use strong passwords (minimum 16 characters, mixed case, numbers, symbols)
- [ ] Generate new JWT_SECRET (use `openssl rand -hex 32`)
- [ ] Generate new ENCRYPTION_KEY (use `openssl rand -base64 32`)
- [ ] Generate new REGISTRATION_SECRET
- [ ] Enable TLS with valid certificates (not self-signed)
- [ ] Set ENVIRONMENT=production
- [ ] Set LOG_LEVEL=warn
- [ ] Disable SETUP_ENDPOINT_ENABLED
- [ ] Use strong database passwords (not "staging_password")
- [ ] Enable SSL Mode in database connections
- [ ] Configure HTTPS for all endpoints
- [ ] Set up rate limiting appropriately
- [ ] Configure backup and recovery procedures
- [ ] Enable audit logging
- [ ] Review all CORS settings
- [ ] Configure firewall rules

---

## Database Schema

All tables are created in the `pganalytics` schema:

### User & Authentication Tables
- `pganalytics.users` - User accounts
- `pganalytics.api_tokens` - API authentication tokens

### Monitoring & Infrastructure
- `pganalytics.collectors` - Monitoring agents
- `pganalytics.servers` - Physical/virtual servers
- `pganalytics.postgresql_instances` - PostgreSQL servers
- `pganalytics.databases` - PostgreSQL databases

### Managed Instances (AWS)
- `pganalytics.managed_instances` - RDS/Aurora instances
- `pganalytics.managed_instance_databases` - RDS databases

### Alerting
- `pganalytics.alert_rules` - Alert rule definitions
- `pganalytics.alerts` - Active alerts

### Configuration & Secrets
- `pganalytics.secrets` - Encrypted credentials
- `pganalytics.collector_config` - Collector configurations
- `pganalytics.registration_secrets` - Collector registration tokens
- `pganalytics.collector_tokens` - Collector authentication

### Auditing & Metrics
- `pganalytics.audit_log` - User action audit trail
- `pganalytics.metric_types` - Metric type definitions
- `pganalytics.schema_versions` - Migration tracking

---

## Troubleshooting

### "Invalid credentials" when logging in
- Verify you're using the correct username: `admin`
- Verify you're using the correct password: `PgAnalytics2026`
- Check that the user was created: `SELECT * FROM pganalytics.users;`
- Check backend logs: `docker logs pganalytics-staging-backend`

### "Setup already completed" when creating new admin
- The setup endpoint can only be used once
- To create additional users, use the admin account
- To reset, delete from `pganalytics.users` and restart backend

### "Connection refused" to database
- Verify PostgreSQL is running: `docker ps`
- Verify port mapping: `docker port pganalytics-staging-postgres`
- Check database logs: `docker logs pganalytics-staging-postgres`

### "password_changed is false" - Password change required
- This is by design - admins must change password on first login
- Change password via frontend or use `/api/v1/auth/change-password` endpoint

---

## Documentation References

- **Database Migration System:** See `MIGRATION_SYSTEM_DOCUMENTATION.md`
- **Deployment Rules:** See `MIGRATION_AND_DEPLOYMENT_RULES.md`
- **Migration Fixes:** See `MIGRATION_FIXES_SUMMARY.md`
- **Fresh Deployment Validation:** See `FRESH_DEPLOYMENT_VALIDATION.md`
- **API Documentation:** See `docs/API.md` (if available)

---

**Created:** 2026-03-12
**Status:** ✅ Tested and Validated
**Environment:** Staging
**Deployment:** Docker Compose
