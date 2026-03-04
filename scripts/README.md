# Deployment & Operations Scripts

This directory contains automated scripts for deploying, monitoring, and maintaining pgAnalytics v3.

## Available Scripts

### 1. `deploy.sh` - Full Deployment Automation

Automates the complete deployment process with health checks and rollback capability.

**Usage**:
```bash
./deploy.sh [environment]
```

**What it does**:
- Runs pre-deployment checks (Docker, Docker Compose)
- Backs up the current database
- Stops all services
- Starts services with Docker Compose
- Performs health checks (API, database, frontend)
- Automatically rolls back on failure
- Logs all actions to timestamped log file

**Features**:
- Automatic rollback on failure
- Comprehensive health checks
- Database backup before deployment
- Detailed logging
- Color-coded output

**Example**:
```bash
./deploy.sh production
# Logs to: deployment-20260324-120000.log
# Backup created if needed
```

### 2. `health-check.sh` - Service Health Verification

Verifies that all services are running and responding correctly.

**Usage**:
```bash
./health-check.sh
```

**Checks**:
- Docker daemon is running
- PostgreSQL container running
- Backend API container running
- Frontend container running
- Backend API health endpoint (HTTP 200)
- Frontend accessibility (HTTP 200)
- Database connectivity

**Exit Codes**:
- `0`: All checks passed
- `1`: One or more checks failed

**Example**:
```bash
./health-check.sh
# === pgAnalytics Health Check ===
# 
# Checking Docker daemon...
# ✓ Docker is running
# 
# Checking PostgreSQL container... ✓ RUNNING
# Checking API container... ✓ RUNNING
# Checking Frontend container... ✓ RUNNING
# 
# Checking Backend API... ✓ OK
# Checking Frontend... ✓ OK
# 
# === All Health Checks Passed ===
```

### 3. `backup.sh` - Database & Configuration Backup

Creates compressed backups of database and configuration files.

**Usage**:
```bash
./backup.sh [backup_directory]
```

**What it backs up**:
- PostgreSQL database (pganalytics)
- Configuration files (.env, docker-compose.yml)

**Features**:
- Automatic retention policy (keeps 30 days)
- Timestamped backups
- Compressed format (gzip)
- Automatic cleanup of old backups
- Clear listing of existing backups

**Example**:
```bash
./backup.sh ./backups
# [INFO] Backup directory: ./backups
# [INFO] Backing up PostgreSQL database...
# [✓] Database backup: ./backups/pganalytics-backup-20260324-120000.sql.gz
# [✓] Config backup: ./backups/pganalytics-backup-20260324-120000-config.tar.gz
# [INFO] Cleaning backups older than 30 days...
# [✓] Cleanup complete
# [✓] Backup completed successfully
```

### 4. `restore.sh` - Restore from Backup

Restores database from a backup file created by backup.sh.

**Usage**:
```bash
./restore.sh <backup_file>
```

**What it restores**:
- PostgreSQL database from backup

**Features**:
- Validates backup file exists
- Confirms before proceeding
- Automatically starts PostgreSQL if needed
- Handles both gzipped and plain SQL files
- Clear confirmation prompts

**Example**:
```bash
./restore.sh ./backups/pganalytics-backup-20260324-120000.sql.gz
# [INFO] Restoring from backup: ./backups/pganalytics-backup-20260324-120000.sql.gz
# [WARN] This will overwrite the current database!
# Continue? (yes/no): yes
# [INFO] Restoring database...
# [✓] Database restored successfully
# [✓] Restore completed
```

---

## Common Workflows

### Daily Backup

```bash
# Create daily backup
0 2 * * * /path/to/pganalytics/scripts/backup.sh /path/to/pganalytics/backups
```

Add to crontab for automatic daily backups at 2 AM.

### Monitoring

```bash
# Check health every 5 minutes
*/5 * * * * /path/to/pganalytics/scripts/health-check.sh || \
    mail -s "pgAnalytics Health Check Failed" admin@example.com
```

Add to crontab for automated monitoring with email alerts.

### Deployment

```bash
# Manual deployment
cd /path/to/pganalytics
./scripts/deploy.sh production

# Check deployment
./scripts/health-check.sh

# Verify data
docker-compose exec postgres psql -U postgres -d pganalytics \
  -c "SELECT COUNT(*) FROM collectors"
```

### Emergency Restore

```bash
# If deployment fails, restore from backup
./scripts/restore.sh ./backups/pganalytics-backup-20260324-120000.sql.gz

# Verify restoration
./scripts/health-check.sh
```

---

## Script Details

### Environment Variables

All scripts respect standard environment variables:

```bash
# Custom docker-compose file
export DOCKER_COMPOSE_FILE=docker-compose.prod.yml

# Custom backup location
export BACKUP_DIR=/mnt/backups/pganalytics

# Custom timeouts
export HEALTH_CHECK_TIMEOUT=120
```

### Logging

All scripts log to stdout and (where applicable) to timestamped log files:

```bash
deployment-20260324-120000.log
```

### Error Handling

Scripts use strict error handling:

```bash
set -e  # Exit on error
```

If any command fails, the script stops immediately.

---

## Best Practices

### 1. Regular Backups

```bash
# Setup daily backups
0 2 * * * cd /path/to/pganalytics && ./scripts/backup.sh ./backups
```

### 2. Pre-Deployment Verification

```bash
# Before deploying, always backup
./scripts/backup.sh ./backups

# Then deploy
./scripts/deploy.sh production

# Verify
./scripts/health-check.sh
```

### 3. Regular Health Checks

```bash
# Monitor continuously
watch -n 5 './scripts/health-check.sh'

# Or via cron
*/5 * * * * /path/to/pganalytics/scripts/health-check.sh
```

### 4. Disaster Recovery Testing

```bash
# Test restore procedure monthly
./scripts/restore.sh ./backups/latest-backup.sql.gz

# Verify data integrity
docker-compose exec postgres psql -U postgres -d pganalytics -c \
  "SELECT COUNT(*) FROM collectors, metrics"

# If successful, re-deploy latest version
./scripts/deploy.sh production
```

### 5. Keep Logs

```bash
# Archive deployment logs
tar -czf logs-archive-$(date +%Y%m).tar.gz deployment-*.log

# Store with backups
mv logs-archive-*.tar.gz ./backups/
```

---

## Troubleshooting

### Script won't execute

```bash
# Make sure scripts are executable
chmod +x ./deploy.sh
chmod +x ./health-check.sh
chmod +x ./backup.sh
chmod +x ./restore.sh
```

### Docker command not found

```bash
# Add docker to PATH
export PATH="/usr/bin:$PATH"

# Or use full path
/usr/bin/docker ps
```

### Backup file is corrupted

```bash
# Test backup file
gunzip -t ./backups/pganalytics-backup-*.sql.gz

# If it fails, restore from earlier backup
ls -lt ./backups/ | head -5
```

### Restore fails with permission denied

```bash
# Check docker-compose is running as correct user
whoami
id

# May need to use sudo
sudo ./scripts/restore.sh ./backups/backup.sql.gz
```

---

## Support

For issues with deployment scripts:

1. Check script output and logs
2. Review DEPLOYMENT_QUICK_REFERENCE.md
3. See FAQ_AND_TROUBLESHOOTING.md
4. Open GitHub issue with logs attached

---

**Remember**: Always backup before deploying! 🚀
