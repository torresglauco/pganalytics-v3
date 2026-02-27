# 🚀 pgAnalytics v3.3.0 - Quick Access Guide

## ⚡ Start in 30 Seconds

```bash
# 1. Start services
docker-compose up -d

# 2. Wait for services to initialize
sleep 30

# 3. Open browser
open http://localhost:4000
```

## 🔑 Login Credentials

```
Username: demo
Password: Demo@12345
```

## 🎯 What to Test

### Test 1: View Collectors
- Go to "Active Collectors" tab
- See 3 registered collectors:
  - Production-DB-01
  - Staging-DB-01
  - Development-DB-01

### Test 2: Test Connection (FIXED! 🎉)
- Go to "Managed Instances" tab
- Click lightning bolt icon (⚡) on pganalytics-postgres-instance
- See: "✓ Connection successful"

### Test 3: Delete Collector
- Go to "Active Collectors"
- Click delete button on any collector
- See: Collector removed from list (no errors)

### Test 4: View Admin Features
- Go to "Registration Secrets" tab
- See 2 active secrets

## 📊 System Status

| Component | Status | URL |
|-----------|--------|-----|
| Frontend | ✅ Ready | http://localhost:4000 |
| Backend API | ✅ Ready | http://localhost:8080/api/v1 |
| PostgreSQL | ✅ Connected | localhost:5432 |
| Grafana | ✅ Ready | http://localhost:3000 |

## 🔧 Useful Commands

### Check Services
```bash
docker-compose ps
```

### View Logs
```bash
# Backend logs
docker-compose logs backend -f

# Frontend logs
docker-compose logs frontend -f

# PostgreSQL logs
docker-compose logs postgres -f
```

### Run Full Test Suite
```bash
bash /tmp/test_all_features.sh
```

### Setup Fresh Demo Data
```bash
bash /tmp/setup_complete_demo.sh
```

## 🐛 Common Issues & Fixes

### "Connection test failed"
- ✅ **FIXED!** Backend now uses stored encrypted credentials
- If still seeing error, refresh page and try again

### "Collectors not showing"
- Refresh page (Ctrl+F5)
- Clear browser cache
- Check backend logs: `docker-compose logs backend`

### "Frontend not loading"
```bash
docker-compose restart frontend
sleep 10
# Then refresh browser
```

### "Cannot login"
```bash
# Check backend health
curl http://localhost:8080/api/v1/health

# Check PostgreSQL
docker exec pganalytics-postgres pg_isready -U postgres
```

## 📝 Documentation

- **Full Setup**: See `READY_FOR_TESTING.md`
- **Complete Validation**: See `VALIDATION_COMPLETE.md`
- **Session Details**: See `SESSION_COMPLETE.md`
- **Implementation Details**: See `FINAL_STATUS_REPORT.md`

## ✅ What's Been Fixed

1. ✅ Delete Collector endpoint (working)
2. ✅ Registration Secrets loading (working)
3. ✅ Connection test authentication (FIXED - uses stored credentials)
4. ✅ Real PostgreSQL database (connected)

## 🎓 API Endpoints (Testing)

### Get All Collectors
```bash
curl http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Test Connection
```bash
curl -X POST http://localhost:8080/api/v1/managed-instances/3/test-connection \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
# Returns: {"success":true}
```

### Get Secrets (Admin)
```bash
curl http://localhost:8080/api/v1/registration-secrets \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🎉 Summary

**Everything is working and ready!**

- ✅ Real collectors registered
- ✅ Real PostgreSQL database connected
- ✅ All CRUD operations functional
- ✅ Connection testing works
- ✅ Admin features accessible
- ✅ All tests passing

**Access**: http://localhost:4000
**Credentials**: demo / Demo@12345

---

*Last Updated: 2026-02-27*
*Status: ✅ Production Ready*
