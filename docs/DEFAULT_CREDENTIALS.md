# pgAnalytics v3.1.0 - Default Credentials

## Frontend (React Application)

**URL:** `http://localhost:3000`

### Default Admin User
```
Username: admin
Password: admin
```

## Grafana Dashboard

**URL:** `http://localhost:3001`

### Default Admin Credentials
```
Username: admin
Password: Th101327!!!
```

## Backend API

**URL:** `http://localhost:8080`

### Health Check Endpoint
```bash
curl http://localhost:8080/api/v1/health
```

Expected Response:
```json
{
  "status": "ok",
  "version": "3.0.0-alpha",
  "timestamp": "2026-04-02T23:36:23.90684309Z",
  "uptime": 0,
  "database_ok": true,
  "timescale_ok": true
}
```

## Database Access

### PostgreSQL Metadata Database (pganalytics)

**Connection Details:**
```
Host: localhost
Port: 5432
Database: pganalytics
Username: postgres
Password: pganalytics
```

**Application User:**
```
Username: pganalytics
Password: (set via environment variable DB_PASSWORD)
```

### TimescaleDB Metrics Database (metrics)

**Connection Details:**
```
Host: localhost
Port: 5433
Database: metrics
Username: postgres
Password: pganalytics
```

## Security Notes

### Development-Only Credentials
⚠️ **WARNING:** These credentials are for **development/testing only**. Do NOT use in production.

### For Production Deployment
1. Change all default passwords immediately after deployment
2. Use environment variables for sensitive credentials
3. Enable TLS/SSL for all connections
4. Implement proper RBAC (Role-Based Access Control)
5. Set up audit logging for all access
6. Rotate credentials regularly

See [SECURITY.md](./SECURITY.md) for detailed security guidelines.

## Changing Default Credentials

### Frontend Admin Password
1. Login with default credentials
2. Navigate to Settings → Account
3. Click "Change Password"
4. Enter new password and confirm

### Grafana Admin Password
1. Login to Grafana at http://localhost:3001
2. Go to Configuration → Users
3. Select the admin user
4. Click "Change Password"

### Database Passwords
Update environment variables in `docker-compose.yml` or `.env`:
```bash
# PostgreSQL metadata database
POSTGRES_PASSWORD=<new-password>
DATABASE_URL="postgres://postgres:<new-password>@postgres:5432/pganalytics?sslmode=disable"

# TimescaleDB metrics database
TIMESCALE_URL="postgres://postgres:<new-password>@timescale:5432/metrics?sslmode=disable"
```

Then restart services:
```bash
docker-compose down
docker-compose up -d
```

## API Token Generation

To create a new API token for programmatic access:

```bash
curl -X POST http://localhost:8080/api/v1/tokens \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "My Integration Token",
    "expires_in": 2592000
  }'
```

## Troubleshooting

### "401 Unauthorized" when accessing API
- Verify JWT_SECRET in docker-compose.yml matches frontend configuration
- Check that Bearer token is included in Authorization header

### Frontend Login Fails
- Ensure backend service is running: `docker-compose ps`
- Check backend logs: `docker-compose logs backend`
- Verify database connectivity: `docker-compose logs postgres`

### Grafana Login Issues
- Check Grafana container is healthy: `docker-compose ps grafana`
- Verify GF_SECURITY_ADMIN_PASSWORD in docker-compose.yml

---

**Last Updated:** April 2, 2026
**Version:** pgAnalytics v3.1.0
