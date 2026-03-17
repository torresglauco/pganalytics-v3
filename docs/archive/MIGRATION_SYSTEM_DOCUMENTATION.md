# Migration System Documentation

## Overview

pgAnalytics v3 uses an **automatic database migration system** that ensures schema consistency and prevents duplicate execution of migrations.

## How It Works

### 1. Automatic Execution on Startup

When the backend starts (`./cmd/pganalytics-api/main.go`):

1. PostgreSQL connection is established
2. `NewPostgresDB()` is called, which internally calls `runMigrations()`
3. `MigrationRunner` loads all `.sql` files from the `/migrations/` directory
4. **Only `.sql` files are executed** - `.disabled` and `.backup` files are skipped
5. Each migration is executed in alphabetical order (000_, 001_, etc.)
6. Migration execution is tracked in `pganalytics.schema_versions` table
7. **Already-executed migrations are never run again** (idempotent)

### 2. Idempotent Execution Guarantee

The system ensures migrations only run once:

```sql
-- schema_versions table tracks all executed migrations
SELECT * FROM pganalytics.schema_versions;

version              | executed_at              | execution_time_ms
--------------------|--------------------------|------------------
000_complete_schema  | 2026-03-11 21:25:00+00  | 1250
```

Before executing a migration:
```go
// Check if migration has already been executed
var executed bool
err := db.QueryRow(
    `SELECT EXISTS(SELECT 1 FROM pganalytics.schema_versions WHERE version = $1)`,
    migrationName,
).Scan(&executed)

if executed {
    skip migration  // Already run, don't run again
}
```

### 3. Transaction Safety

Each migration is executed in a PostgreSQL transaction:
- All statements execute together
- If any statement fails, the entire migration is rolled back
- Migration is only marked as executed if it fully succeeds
- This prevents partial schema updates

## Migration Files

### Naming Convention

```
NNN_description.sql    (active migration)
NNN_description.sql.disabled    (disabled - will not execute)
NNN_description.sql.backup      (backup - will not execute)
```

### File Naming Rules

1. **Leading zeros**: Use `000_`, `001_`, `002_` (3-digit prefix)
2. **Alphabetical order**: Files execute in alphanumeric order
3. **Descriptive names**: Use underscores for spaces
4. **Extensions**: Only `.sql` files execute (not `.disabled` or `.backup`)

### Example Migration Files

| Filename | Status | Will Execute? |
|----------|--------|---------------|
| `000_complete_schema.sql` | Active | ✅ Yes |
| `001_feature_x.sql.disabled` | Disabled | ❌ No |
| `002_fix_bug.sql.backup` | Backup | ❌ No |
| `003_update.sql` | Active | ✅ Yes |

## Current Schema Structure

### Complete Unified Schema (000_complete_schema.sql)

The `000_complete_schema.sql` migration creates the **entire database schema** in one migration:

#### Tables Created

**Authentication & Users:**
- `users` - User accounts with password change tracking
- `api_tokens` - API authentication tokens
- `registration_secrets` - Secrets for collector self-registration
- `registration_secret_audit` - Audit trail of secret usage

**Collectors (Monitoring Agents):**
- `collectors` - Collector registration and status
- `collector_tokens` - Collector authentication tokens
- `collector_config` - Collector configuration versions

**Managed Instances (RDS/Aurora):**
- `managed_instances` - Managed PostgreSQL database instances
- `managed_instance_databases` - Databases within managed instances

**Infrastructure:**
- `servers` - Physical/virtual servers
- `postgresql_instances` - PostgreSQL instances on servers
- `databases` - Databases within instances

**Monitoring:**
- `metric_types` - Types of metrics collected
- `alert_rules` - Alert rule definitions
- `alerts` - Alert instances
- `audit_log` - Activity audit log

**Secrets:**
- `secrets` - Encrypted credential storage

#### Key Features

✅ **Complete** - All necessary tables created in one migration
✅ **Comprehensive** - Includes all relationships and constraints
✅ **Safe** - Uses `CREATE TABLE IF NOT EXISTS` for idempotency
✅ **Indexed** - All critical columns have indexes
✅ **Secure** - RBAC setup, encrypted fields
✅ **Auditable** - Audit log and tracking tables

## Important Fields

### Users Table

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),

    -- NEW: Password change tracking
    password_changed BOOLEAN DEFAULT false,
    last_password_change TIMESTAMP,

    -- Security
    is_active BOOLEAN DEFAULT true,
    role VARCHAR(50) CHECK (role IN ('admin', 'user', 'viewer')),
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,

    -- Audit
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Collectors Table

```sql
CREATE TABLE collectors (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'registered',  -- registered, active, offline, error

    -- Tracking
    last_seen TIMESTAMP,
    metrics_count_24h BIGINT,

    -- Configuration
    config_version INTEGER DEFAULT 0,
    health_check_interval INTEGER DEFAULT 300,

    -- Audit
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Registration Secrets Table

```sql
CREATE TABLE registration_secrets (
    id UUID PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    secret_value VARCHAR(255) UNIQUE NOT NULL,

    -- Lifecycle
    active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP,

    -- Usage tracking
    total_registrations INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    max_registrations INTEGER,  -- NULL = unlimited

    -- Audit
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### Managed Instances Table

```sql
CREATE TABLE managed_instances (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    aws_region VARCHAR(50) NOT NULL,
    rds_endpoint VARCHAR(255) UNIQUE NOT NULL,

    -- Configuration
    port INTEGER DEFAULT 5432,
    engine_version VARCHAR(50),
    db_instance_class VARCHAR(50),
    ssl_enabled BOOLEAN DEFAULT true,
    ssl_mode VARCHAR(20) DEFAULT 'require',

    -- Status
    is_active BOOLEAN DEFAULT true,
    last_connection_status VARCHAR(50) DEFAULT 'unknown',
    last_heartbeat TIMESTAMP,
    last_error_message TEXT,

    -- AWS metadata
    environment VARCHAR(50) DEFAULT 'production',
    multi_az BOOLEAN DEFAULT false,
    backup_retention_days INTEGER,

    -- Audit
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## Migration Execution Flow

```
┌─────────────────────────────────────────────────────────┐
│         Backend Startup (main.go)                       │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│   NewPostgresDB(connectionString)                       │
│   - Connects to PostgreSQL                              │
│   - Sets search_path to pganalytics schema              │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│   runMigrations() [in postgres.go]                      │
│   - Creates MigrationRunner                             │
│   - Calls runner.Run(ctx)                               │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│   MigrationRunner.Run() [in migrations.go]              │
│   1. Create schema_versions table (if needed)           │
│   2. Load executed migrations from table                │
│   3. Find .sql files in /migrations/ directory          │
│   4. Sort files alphabetically                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│   For Each Migration File (in order)                    │
│   1. Check if already in schema_versions                │
│   2. If yes: Skip it                                    │
│   3. If no:                                             │
│      - Begin transaction                                │
│      - Execute all SQL statements                       │
│      - Record in schema_versions                        │
│      - Commit transaction                               │
│      - Log success with execution time                  │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│   Database Ready for API Operations                     │
└─────────────────────────────────────────────────────────┘
```

## Disabling Migrations

To disable a migration (prevent it from running):

```bash
# Rename the file to add .disabled extension
mv backend/migrations/001_init.sql backend/migrations/001_init.sql.disabled

# File will no longer be executed
# (Migration runner skips .disabled and .backup files)
```

This is useful for:
- Deactivating broken migrations
- Preventing duplicate executions of old migrations
- Testing with different schema versions

## Verification

### Check Executed Migrations

```sql
SELECT * FROM pganalytics.schema_versions ORDER BY executed_at;
```

### Check Migration Logs

The migration runner logs to the application logs:

```
INFO: Starting migration runner
INFO: Found migration files count=1
INFO: Executing migration name=000_complete_schema
INFO: Migration executed successfully name=000_complete_schema execution_time_ms=1250
INFO: Migrations completed
```

### Verify Table Existence

```sql
-- Check if all critical tables exist
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'pganalytics'
ORDER BY table_name;

-- Check specific table
\d pganalytics.collectors
\d pganalytics.registration_secrets
\d pganalytics.managed_instances
```

## Troubleshooting

### Problem: Migrations Don't Execute

**Cause 1: Migrations directory not found**
```
Solution: Ensure /migrations/ directory is mounted in Docker
or set MIGRATIONS_PATH environment variable
```

**Cause 2: File extension is wrong**
```
Solution: Check that files end in .sql (not .disabled or .backup)
```

**Cause 3: Already executed**
```
Solution: This is normal - migrations only run once
Check schema_versions table to see execution history
```

### Problem: Migration Fails During Execution

**Solution: Check logs for SQL errors**
- Error message shows which statement failed
- Check if database already has tables with same names
- Ensure PostgreSQL has required extensions (uuid-ossp, pgcrypto, etc.)

### Problem: Need to Re-Run a Migration

**For development only:**
```sql
-- Delete migration from tracking table
DELETE FROM pganalytics.schema_versions
WHERE version = '000_complete_schema';

-- Restart the application - migration will run again
```

⚠️ **WARNING**: Never do this in production without a backup!

## Best Practices

### 1. New Features

Create new migration with next number:
```bash
# If latest is 000_complete_schema.sql, create:
# 001_new_feature.sql
```

### 2. Hotfixes

For urgent fixes in production:
1. Create a new migration (not modifying existing ones)
2. Test thoroughly in staging
3. Deploy with version bump

### 3. Schema Changes

For adding/modifying columns:
```sql
-- Use ALTER TABLE with IF NOT EXISTS / IF EXISTS clauses
ALTER TABLE users ADD COLUMN IF NOT EXISTS new_column VARCHAR(255);
ALTER TABLE users DROP COLUMN IF EXISTS old_column;
```

### 4. Data Migrations

For data transformations:
```sql
-- Use transactions to ensure atomicity
BEGIN;
-- Multiple statements
COMMIT;
```

## Docker Configuration

Migrations are automatically copied to the container:

```dockerfile
# In backend/Dockerfile
COPY backend/migrations ./migrations

# At runtime, migrations are executed automatically
# when NewPostgresDB() is called in main.go
```

The `docker-compose.staging.yml` mounts migrations:
```yaml
backend-staging:
  volumes:
    - ./backend/migrations:/migrations:ro
```

## Environment Variables

**MIGRATIONS_PATH** (optional)
- Override the default migration directory
- Example: `MIGRATIONS_PATH=/custom/path`
- If not set, the runner searches common locations

## Summary

✅ **Automatic**: Migrations run automatically on startup
✅ **Safe**: Idempotent - never runs the same migration twice
✅ **Transactional**: All-or-nothing execution
✅ **Tracked**: Every executed migration is recorded
✅ **Flexible**: Easy to disable or skip migrations
✅ **Complete**: Single comprehensive schema migration
✅ **Production-Ready**: Used in staging and production deployments
