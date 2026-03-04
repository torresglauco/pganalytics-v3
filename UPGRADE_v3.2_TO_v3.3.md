# Upgrade Guide: pgAnalytics v3.2.0 → v3.3.0
**Date**: March 4, 2026
**Current Version**: v3.2.0 (Production)
**Target Version**: v3.3.0 (April 30, 2026)
**Estimated Time**: 2-4 hours
**Difficulty**: Medium

---

## 📋 Pre-Upgrade Checklist

Before starting the upgrade, ensure you have:

- [ ] **Backup**: Complete backup of PostgreSQL database
- [ ] **Staging Environment**: Test upgrade in staging first
- [ ] **Downtime Plan**: Scheduled upgrade window
- [ ] **Communication**: Notified users of maintenance
- [ ] **Rollback Plan**: Know how to rollback if needed
- [ ] **Monitoring**: Have monitoring dashboard ready
- [ ] **Support Team**: Team available for issues

---

## ⚠️ Breaking Changes in v3.3.0

### 1. **Database Schema Changes**

Two new tables will be created:

#### `audit_logs` (NEW)
For compliance and security auditing
```sql
CREATE TABLE audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  action VARCHAR(255) NOT NULL,
  resource_type VARCHAR(100),
  resource_id VARCHAR(255),
  changes JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ip_address INET,
  user_agent TEXT
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

#### `auth_logs` (NEW)
For authentication event tracking
```sql
CREATE TABLE auth_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id INTEGER,
  event VARCHAR(100),
  success BOOLEAN,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_auth_logs_user_id ON auth_logs(user_id);
```

#### `collectors` (MODIFIED)
One new column added:
```sql
ALTER TABLE collectors ADD COLUMN IF NOT EXISTS
  ldap_sync BOOLEAN DEFAULT FALSE;
```

### 2. **Environment Variable Changes**

New required variables for v3.3.0:

```bash
# NEW - LDAP Configuration
LDAP_ENABLED=false
LDAP_URL=ldap://ldap.example.com
LDAP_BIND_DN=cn=admin,dc=example,dc=com
LDAP_BIND_PASSWORD=password
LDAP_BASE_DN=dc=example,dc=com
LDAP_USER_FILTER=(uid=%s)
LDAP_GROUP_FILTER=(memberUid=%s)

# NEW - Encryption at Rest
ENCRYPTION_ENABLED=true
ENCRYPTION_KEY_FILE=/etc/pganalytics/encryption.key

# NEW - Audit Logging
AUDIT_LOG_ENABLED=true
AUDIT_LOG_RETENTION_DAYS=90

# NEW - Backup Configuration
BACKUP_ENABLED=true
BACKUP_SCHEDULE="0 2 * * *"  # Daily at 2 AM UTC
BACKUP_RETENTION_DAYS=30
BACKUP_LOCATION=s3://pganalytics-backups
```

### 3. **API Changes**

#### New Endpoints (v3.3.0)
```
POST   /api/v1/auth/ldap            # LDAP login
POST   /api/v1/auth/saml            # SAML login
POST   /api/v1/auth/oauth           # OAuth login
GET    /api/v1/audit-logs          # View audit logs
GET    /api/v1/backups              # List backups
POST   /api/v1/backups              # Trigger backup
```

#### Modified Endpoints
No breaking changes to existing endpoints. New fields are optional.

#### Deprecated (v3.3.0)
No endpoints deprecated in v3.3.

### 4. **Configuration Changes**

#### Before (v3.2.0)
```yaml
# docker-compose.yml
services:
  api:
    environment:
      JWT_SECRET: your-secret
      DATABASE_URL: postgres://...
      TLS_CERT: /path/to/cert
      TLS_KEY: /path/to/key
```

#### After (v3.3.0)
```yaml
# docker-compose.yml
services:
  api:
    environment:
      JWT_SECRET: your-secret
      DATABASE_URL: postgres://...
      TLS_CERT: /path/to/cert
      TLS_KEY: /path/to/key
      # NEW - Optional authentication methods
      LDAP_ENABLED: "false"
      # NEW - Encryption
      ENCRYPTION_ENABLED: "true"
      ENCRYPTION_KEY_FILE: /etc/pganalytics/encryption.key
```

---

## 📋 Step-by-Step Upgrade Procedure

### Phase 1: Preparation (30 minutes)

#### Step 1: Backup Database
```bash
# Full backup
pg_dump -h localhost -U postgres -d pganalytics > backup_v3.2.0_$(date +%Y%m%d_%H%M%S).sql

# TimescaleDB backup (if using)
pg_dump -h localhost -U postgres -d pganalytics --format=custom > backup_timescale.dump

# Verify backup
ls -lah backup_*.sql
```

#### Step 2: Backup Configuration Files
```bash
# Backup current config
cp .env .env.v3.2.0.backup
cp docker-compose.yml docker-compose.yml.v3.2.0.backup

# Store vault credentials if using HashiCorp Vault
cp ~/.vault-token ~/.vault-token.backup
```

#### Step 3: Disable Collectors (Optional)
```bash
# Send SIGTERM to collectors to gracefully shutdown
# They will stop sending metrics during upgrade
# This prevents inconsistent data

# On each collector:
sudo systemctl stop pganalytics-collector
```

#### Step 4: Enable Maintenance Mode
```bash
# Optional: Set backend to read-only mode
# This prevents users from making changes during upgrade

curl -X POST http://localhost:8080/api/v1/admin/maintenance-mode \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": true, "message": "System upgrade in progress"}'
```

### Phase 2: Database Migration (1 hour)

#### Step 5: Run Migrations
```bash
# Pull v3.3.0 code
git pull origin main
git checkout v3.3.0  # Or use release tag when available

# Run migrations
docker-compose exec -T postgres psql -U postgres -d pganalytics << EOF
-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  action VARCHAR(255) NOT NULL,
  resource_type VARCHAR(100),
  resource_id VARCHAR(255),
  changes JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ip_address INET,
  user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Auth logs table
CREATE TABLE IF NOT EXISTS auth_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id INTEGER,
  event VARCHAR(100),
  success BOOLEAN,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_auth_logs_user_id ON auth_logs(user_id);

-- Add new columns to collectors
ALTER TABLE collectors ADD COLUMN IF NOT EXISTS ldap_sync BOOLEAN DEFAULT FALSE;
EOF
```

#### Step 6: Verify Migrations
```bash
# Check new tables exist
docker-compose exec postgres psql -U postgres -d pganalytics -c "\dt audit_logs"
docker-compose exec postgres psql -U postgres -d pganalytics -c "\dt auth_logs"

# Should see tables listed
# Output: "Did not find any relation named..."  = NOT applied
# Output: table name listed             = SUCCESS
```

### Phase 3: Application Upgrade (30 minutes)

#### Step 7: Update Configuration
```bash
# Copy template if needed
cp .env.v3.2.0.backup .env.v3.3.0

# Add new variables (or they'll use defaults)
cat >> .env.v3.3.0 << EOF
# v3.3.0 New Configuration
LDAP_ENABLED=false
ENCRYPTION_ENABLED=true
AUDIT_LOG_ENABLED=true
BACKUP_ENABLED=true
BACKUP_SCHEDULE="0 2 * * *"
EOF

# Set as active
mv .env.v3.3.0 .env
```

#### Step 8: Update Docker Image
```bash
# Pull new version
docker pull ghcr.io/torresglauco/pganalytics-api:v3.3.0

# Verify image exists
docker images | grep pganalytics
```

#### Step 9: Restart Services
```bash
# Graceful restart with health checks
docker-compose down

# Update image in docker-compose.yml manually or use:
sed -i 's/v3.2.0/v3.3.0/g' docker-compose.yml

# Start with health monitoring
docker-compose up -d

# Wait for startup
sleep 10

# Check health
curl http://localhost:8080/api/v1/health
# Should return: {"status": "ok", "version": "v3.3.0"}
```

#### Step 10: Verify API
```bash
# Test critical endpoints
curl -s http://localhost:8080/api/v1/health | jq .version
curl -s http://localhost:8080/api/v1/collectors | jq '.[] | .id' | head -5
curl -s http://localhost:8080/api/v1/audit-logs | jq '.[] | .action' | head -5
```

### Phase 4: Post-Upgrade Validation (30 minutes)

#### Step 11: Restart Collectors
```bash
# Collectors should auto-reconnect
# Monitor logs for successful reconnections

# On each collector:
sudo systemctl start pganalytics-collector

# Check collector logs
tail -f /var/log/pganalytics/collector.log
# Should see: "Successfully registered with backend"
```

#### Step 12: Validate Data Flow
```bash
# Check new metrics are being collected
curl -s http://localhost:8080/api/v1/servers | jq '.[] | .last_update'

# Check audit logs are being written
curl -s -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8080/api/v1/audit-logs | jq length
```

#### Step 13: Disable Maintenance Mode
```bash
# Re-enable write operations
curl -X POST http://localhost:8080/api/v1/admin/maintenance-mode \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

#### Step 14: Run Test Suite
```bash
# Verify all tests pass
make test-backend
make test-integration

# Expected: All tests passing ✅
```

---

## 🔄 Rollback Procedure

If upgrade fails, rollback to v3.2.0:

### Quick Rollback (30 minutes)

```bash
# 1. Stop v3.3.0 services
docker-compose down

# 2. Restore backup configuration
cp docker-compose.yml.v3.2.0.backup docker-compose.yml
cp .env.v3.2.0.backup .env

# 3. Pull v3.2.0 image
docker pull ghcr.io/torresglauco/pganalytics-api:v3.2.0

# 4. Start v3.2.0
docker-compose up -d

# 5. Verify health
curl http://localhost:8080/api/v1/health
```

### Database Rollback (if needed)

```bash
# If migration failed, restore database backup
# WARNING: This will lose data created after upgrade started

pg_restore -h localhost -U postgres -d pganalytics backup_v3.2.0_*.dump

# Verify data
docker-compose exec postgres psql -U postgres -d pganalytics -c "SELECT COUNT(*) FROM collectors"
```

### Full Rollback (with downtime)

```bash
# If partial rollback fails, do full database restore
docker-compose down -v

# Restore all data
psql -h localhost -U postgres -d pganalytics < backup_v3.2.0_*.sql

# Verify
docker-compose up -d
curl http://localhost:8080/api/v1/health
```

---

## 🧪 Testing Checklist

Before considering upgrade complete, verify:

### API Testing
- [ ] GET /api/v1/health returns v3.3.0
- [ ] GET /api/v1/collectors returns all collectors
- [ ] GET /api/v1/servers returns all servers
- [ ] POST /api/v1/metrics/push succeeds
- [ ] POST /api/v1/auth/login works (basic auth still works)
- [ ] GET /api/v1/audit-logs returns audit entries (new in v3.3)

### Data Integrity
- [ ] All collectors still registered
- [ ] All metrics data preserved
- [ ] No data loss detected
- [ ] Dashboards still display data
- [ ] Alerts still functioning

### Performance
- [ ] API response time < 500ms (normal)
- [ ] Database query time normal
- [ ] Memory usage stable
- [ ] CPU usage normal

### Security
- [ ] TLS/mTLS still working
- [ ] JWT tokens still valid
- [ ] RBAC permissions enforced
- [ ] Audit logs recording actions

---

## 🔍 Troubleshooting

### Issue: Migration Failed

**Error**: `ERROR: relation "audit_logs" already exists`

**Solution**:
```bash
# Migration already applied, safe to continue
# Run verification step to confirm
```

---

### Issue: Collectors Can't Connect

**Error**: `SSL/TLS certificate error`

**Cause**: TLS certificates may have changed

**Solution**:
```bash
# Re-register collectors with new certificates
# Or ensure certificate paths haven't changed
curl -X POST http://localhost:8080/api/v1/collectors/register \
  -H "Content-Type: application/json" \
  -d '{
    "hostname": "db-server",
    "port": 5432
  }'
```

---

### Issue: API Not Responding

**Error**: `Connection refused` or timeout

**Solution**:
```bash
# Check if container is running
docker-compose ps

# Check logs
docker-compose logs api

# Ensure migration completed
docker-compose exec postgres psql -U postgres -d pganalytics \
  -c "SELECT COUNT(*) FROM audit_logs"

# Restart if needed
docker-compose restart api
```

---

### Issue: Performance Degradation

**Error**: Slow API responses after upgrade

**Cause**: Possible index issues or missing configuration

**Solution**:
```bash
# Rebuild indexes
docker-compose exec postgres psql -U postgres -d pganalytics << EOF
REINDEX TABLE audit_logs;
REINDEX TABLE collectors;
ANALYZE;
EOF

# Check configuration
docker-compose exec api env | grep LDAP
docker-compose exec api env | grep ENCRYPTION
```

---

## 📞 Support & Rollback Contact

If you encounter issues during upgrade:

1. **Immediate Action**: Review logs and troubleshooting section above
2. **Database Issue**: Check database backups and use rollback procedure
3. **API Issue**: Restart container and check configuration
4. **Collector Issue**: Re-register collectors with backend

---

## 📝 Upgrade Documentation

Keep these files for reference:

- ✅ This upgrade guide
- ✅ Database backup (backup_v3.2.0_*.sql)
- ✅ Configuration backup (.env.v3.2.0.backup)
- ✅ Migration statements (for rollback)
- ✅ Test results (after completion)

---

## ✅ Upgrade Completion Checklist

After completing all steps:

- [ ] All services started successfully
- [ ] API health check passing
- [ ] All collectors reconnected
- [ ] Data flowing normally
- [ ] Audit logs being recorded
- [ ] Tests passing
- [ ] Users notified of completion
- [ ] Backups stored safely
- [ ] Documentation updated

---

## Summary

**Upgrade Time**: ~2-4 hours total
**Downtime**: 10-30 minutes (during Phase 3)
**Risk Level**: Low (with backups and rollback plan)
**Support**: Available in #pganalytics-support

**Next**: See EXECUTIVE_SUMMARY_MARCH_2026.md for post-upgrade recommendations

---

**Created**: March 4, 2026
**Version**: v3.3.0 Upgrade Guide
**Status**: Ready for Use
