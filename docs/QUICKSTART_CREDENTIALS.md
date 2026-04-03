# pgAnalytics v3.1.0 - Quick Start Credentials

**Last Updated:** April 3, 2026
**Version:** v3.1.0
**Status:** ✅ Production Ready

---

## 🚀 Quick Access

After running `docker-compose up -d` or `mise run dev`, use these credentials to access the system:

### Web Frontend (Dashboard & Administration)
```
URL:      http://localhost:3000
Username: admin
Password: admin
Path:     /admin (for administration panel)
```

### Grafana (Monitoring Dashboards)
```
URL:      http://localhost:3001
Username: admin
Password: Th101327!!!
```

### PostgreSQL Database
```
Host:     localhost
Port:     5432
Database: pganalytics
Username: postgres
Password: pganalytics
Schema:   pganalytics
```

### Backend API
```
URL:  http://localhost:8080/api/v1
Auth: Bearer Token (obtained via /auth/login)
```

---

## 🔐 Login Flow

### Step 1: Get JWT Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@pganalytics.local",
    "role": "admin",
    "is_active": true
  }
}
```

### Step 2: Use Token in API Calls
```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer eyJhbGci..."
```

---

## 📊 Key Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/auth/login` | POST | ❌ | User authentication |
| `/auth/me` | GET | ✅ | Get current user |
| `/users` | GET | ✅ | List all users (admin only) |
| `/users/{id}` | GET | ✅ | Get user details |
| `/users/{id}` | PUT | ✅ | Update user (admin only) |
| `/users/{id}` | DELETE | ✅ | Delete user (admin only) |
| `/health` | GET | ❌ | System health check |

---

## 👥 User Roles

### Admin
- **Default Username:** admin
- **Default Password:** admin
- **Permissions:** Full system access
- **Can:** Manage users, create collectors, configure alerts, view all metrics

### User
- **Permissions:** Standard user access
- **Can:** View metrics, create alerts (if assigned), run queries

### Viewer
- **Permissions:** Read-only access
- **Can:** View dashboards and metrics only

---

## 🔧 First-Time Setup

1. **Login to Frontend**
   ```
   http://localhost:3000
   Username: admin
   Password: admin
   ```

2. **Navigate to Administration Panel**
   ```
   Click: Settings → Administration → Users
   ```

3. **Create Additional Users** (if needed)
   - Click "+ New User"
   - Set username, email, password, and role
   - Click "Create"

4. **Register a Collector**
   ```
   Go to: Collectors → Register Collector
   Copy the Registration Secret from: Settings → API Credentials
   ```

5. **Configure Monitoring**
   ```
   Navigate to: Collectors → Select Collector → Configure
   ```

---

## 🔐 Security Recommendations

### Before Production Deployment

1. **Change Default Passwords**
   - Change admin password in Frontend → Settings → Account
   - Change PostgreSQL password in environment
   - Change Grafana password

2. **Enable TLS/SSL**
   ```bash
   # Update docker-compose.yml with TLS certificates
   TLS_CERT: "/path/to/cert.pem"
   TLS_KEY: "/path/to/key.pem"
   ```

3. **Set Strong JWT Secret**
   ```bash
   # Update docker-compose.yml
   JWT_SECRET: "your-very-long-secure-random-string"
   ```

4. **Configure Encryption Key**
   ```bash
   # Generate new key
   openssl rand -base64 32

   # Update docker-compose.yml
   ENCRYPTION_KEY: "base64-encoded-32-byte-key"
   ```

5. **Disable Debug Mode**
   ```bash
   # Update docker-compose.yml
   LOG_LEVEL: "info"  # Change from "debug"
   ```

---

## 🚨 Troubleshooting

### "Unauthorized" when accessing /admin
- Verify login was successful
- Check that JWT token is valid (not expired)
- Token expires after 15 minutes by default
- Use /auth/refresh to get new token

### "User not found" on login
- Ensure database migrations ran successfully
- Check: `docker-compose logs postgres`
- Verify admin user exists in database:
  ```bash
  docker-compose exec -T postgres psql -U postgres -d pganalytics \
    -c "SELECT * FROM pganalytics.users WHERE username='admin';"
  ```

### Frontend can't reach Backend API
- Verify backend is running: `docker-compose ps`
- Check backend health: `curl http://localhost:8080/api/v1/health`
- Verify API_URL in frontend is correct
- Check browser console for CORS errors

### Grafana can't connect to PostgreSQL
- Verify postgres service is healthy
- Check Grafana logs: `docker-compose logs grafana`
- Verify connection string in Grafana data source

---

## 📝 Default User Database Entry

```sql
-- Default admin user (created during initial setup)
INSERT INTO pganalytics.users (
  username,
  email,
  password_hash,
  full_name,
  role,
  is_active
) VALUES (
  'admin',
  'admin@pganalytics.local',
  crypt('admin', gen_salt('bf')),
  'Administrator',
  'admin',
  true
);
```

---

## 🔗 Related Documentation

- **[SECURITY.md](./SECURITY.md)** - Security best practices
- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Production deployment guide
- **[docs/API_SECURITY_REFERENCE.md](./docs/API_SECURITY_REFERENCE.md)** - API security details
- **[docs/TEAM_TRAINING_GUIDE.md](./docs/TEAM_TRAINING_GUIDE.md)** - User onboarding

---

## ✅ Verification Checklist

After deployment, verify:

- [ ] Frontend login works (admin/admin)
- [ ] Backend API returns 200 on /health
- [ ] Users list displays in admin panel
- [ ] Grafana dashboards load
- [ ] PostgreSQL connection is healthy
- [ ] Can create new users
- [ ] Can register collectors

---

**Generated:** April 3, 2026
**pgAnalytics Version:** 3.1.0
