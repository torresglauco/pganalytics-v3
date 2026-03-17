# Migration and Deployment Rules for pgAnalytics v3

This document establishes the rules and best practices for database migrations, schema management, and deployment procedures. **These rules MUST be followed on every deployment to ensure system integrity and prevent breaking changes.**

## Table of Contents

1. [Migration System Overview](#migration-system-overview)
2. [Default Credentials](#default-credentials)
3. [Schema Rules](#schema-rules)
4. [Migration Versioning Rules](#migration-versioning-rules)
5. [Migration Development Rules](#migration-development-rules)
6. [Idempotency Guarantee](#idempotency-guarantee)
7. [Deployment Checklist](#deployment-checklist)
8. [Troubleshooting](#troubleshooting)

---

## Migration System Overview

### How It Works

pgAnalytics uses an **automatic, idempotent migration system** that executes on backend startup:

```
Backend Start
    ↓
NewPostgresDB() called in main.go
    ↓
runMigrations() called in postgres.go
    ↓
MigrationRunner created and starts migration execution
    ↓
CREATE SCHEMA pganalytics (if not exists)
CREATE TABLE schema_versions (tracks executed migrations)
    ↓
Load all *.sql files from /backend/migrations/ directory
    ↓
For each migration file (in alphabetical order):
  - Check if version exists in schema_versions table
  - If NOT executed: Execute the entire SQL file in a transaction
  - If already executed: Skip it (idempotent)
  - Record in schema_versions with execution timestamp
    ↓
Database ready for API operations
```

### Why This Approach

✅ **Automatic**: No manual migration commands needed
✅ **Idempotent**: Never executes the same migration twice
✅ **Transactional**: All-or-nothing execution (rollback on error)
✅ **Tracked**: Every executed migration is recorded
✅ **Safe**: Non-blocking if migrations fail (logs warning)

---

## Default Credentials

### CRITICAL: First Deployment Setup

On the **first deployment** to a fresh database:

**There are NO pre-created default users.**

To create the first admin user, call the setup endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/auth/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "Admin@Secure123",
    "full_name": "Administrator"
  }'
```

**Response:**
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "full_name": "Administrator",
  "role": "admin",
  "password_changed": false,
  "token": "eyJhbGciOiJIUzI1NiI...",
  "refresh_token": "eyJhbGciOiJIUzI1NiI..."
}
```

**IMPORTANT:** The `password_changed: false` flag indicates that the user MUST change their password on first login.

### First Login Password Change

After creating the admin user, they must:

1. Log in with the setup password
2. Receive JWT tokens
3. Call password change endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/auth/change-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "old_password": "Admin@Secure123",
    "new_password": "NewSecure@Password123"
  }'
```

After successful password change:
- `password_changed` is set to `true`
- User can access the dashboard
- User can only change password through frontend UI (not direct API)

### No Hardcoded Default Users

❌ **DO NOT create hardcoded default users in migrations**
❌ **DO NOT hardcode passwords**
❌ **DO NOT skip the password change flow**

**Why:**
- Security risk (passwords visible in code)
- Makes tracking difficult
- Setup endpoint is the proper way

---

## Schema Rules

### ALL Tables Must Be in `pganalytics` Schema

**RULE: Every single table created by pgAnalytics must be in the `pganalytics` schema, NOT in the `public` schema.**

✅ **CORRECT:**
```sql
CREATE TABLE pganalytics.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL
);
```

❌ **WRONG:**
```sql
CREATE TABLE public.users (  -- ❌ DO NOT USE public SCHEMA
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL
);
```

❌ **WRONG:**
```sql
CREATE TABLE users (  -- ❌ Default to public, needs explicit schema
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL
);
```

### Schema Creation Order

The schema is created BEFORE any migrations run:

1. Migration runner calls `createVersionsTable()`
2. This function executes: `CREATE SCHEMA IF NOT EXISTS pganalytics`
3. Then creates the `schema_versions` tracking table
4. Then all migration SQL files are executed
5. All migrations should reference `pganalytics.table_name` explicitly

### Search Path Configuration

The backend sets the search path to include pganalytics:

```go
// In NewPostgresDB() in postgres.go
db.ExecContext(ctx, "SET search_path TO pganalytics, public")
```

This means table references can omit the schema in Go code, but **SQL files should always be explicit** to avoid ambiguity.

### Setting Search Path in Migrations

If your migration needs to ensure the search path, add this at the top:

```sql
SET search_path TO pganalytics, public;
```

---

## Migration Versioning Rules

### Naming Convention

**RULE: Migration files MUST follow the naming pattern: `NNN_description.sql`**

Format:
- **NNN**: 3-digit zero-padded number (001, 002, 003, ..., 999)
- **description**: kebab-case description (max 50 characters)
- **extension**: `.sql` for active migrations

✅ **CORRECT:**
- `000_complete_schema.sql`
- `001_user_authentication.sql`
- `002_add_collector_tables.sql`
- `015_add_new_feature.sql`

❌ **WRONG:**
- `1_init.sql` (missing zero-padding)
- `create_users_table.sql` (missing number prefix)
- `001_schema.txt` (wrong extension)
- `001_schema.sql.bak` (backup, won't execute)

### Disabling Migrations

To disable a migration **without deleting it**, rename with `.disabled` extension:

```bash
# Disable a migration
mv backend/migrations/001_old_feature.sql backend/migrations/001_old_feature.sql.disabled

# The .disabled file will NOT be executed
# The migration runner skips all .disabled and .backup files
```

### What Gets Executed

Files that execute:
- ✅ `000_complete_schema.sql`
- ✅ `001_users_table.sql`
- ✅ `015_new_feature.sql`

Files that DO NOT execute:
- ❌ `001_init.sql.disabled`
- ❌ `002_backup.sql.backup`
- ❌ `README.md`
- ❌ `.DS_Store`

### Execution Order

Migrations execute in **strict alphabetical order**:

1. `000_complete_schema.sql`
2. `001_...sql`
3. `002_...sql`
4. ...
5. `999_...sql`

**RULE: Every migration must be independent and work regardless of previous migrations (see Idempotency below)**

---

## Migration Development Rules

### Writing New Migrations

When you need to add new tables or modify schema:

1. **Find the next number** in the sequence
2. **Use descriptive names**
3. **Create ONE migration per logical change**
4. **Use IF NOT EXISTS / IF EXISTS for safety**

Example for adding a new feature:

```bash
# Check existing migrations
ls backend/migrations/

# Find latest: 017_anomaly_detection.sql
# Next number: 018

# Create new migration
touch backend/migrations/018_new_feature.sql
```

Content example:

```sql
-- New Feature Implementation
-- Adds tables for feature tracking and metrics

SET search_path TO pganalytics, public;

-- Create feature_metrics table
CREATE TABLE IF NOT EXISTS feature_metrics (
    id SERIAL PRIMARY KEY,
    feature_name VARCHAR(255) NOT NULL,
    metric_value NUMERIC(10, 2),
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_feature_metrics_name
    ON feature_metrics(feature_name);

-- Record migration completion
INSERT INTO schema_versions (version, description)
    VALUES ('018_new_feature', 'Add feature metrics tracking')
    ON CONFLICT DO NOTHING;
```

### SQL Formatting Rules

✅ **DO:**
- Use proper indentation (2 or 4 spaces)
- Add comments explaining what the migration does
- Use multi-line formatting for complex statements
- Include all necessary indexes
- Use IF NOT EXISTS / IF EXISTS clauses
- Explicitly specify `pganalytics` schema
- Keep lines under 100 characters when possible

❌ **DON'T:**
- Use hardcoded data (except in seed migrations)
- Create tables in `public` schema
- Add complex logic (use backend code instead)
- Skip schema specification (use `pganalytics.`)
- Put multiple unrelated changes in one migration

### Special SQL Handling

The migration runner properly handles:

**Dollar-quoted strings (for function definitions):**
```sql
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

**Single-quoted strings:**
```sql
INSERT INTO users (username, role)
    VALUES ('admin', 'admin');
```

**Comments:**
```sql
-- This is a comment
/* This is a
   block comment */
```

---

## Idempotency Guarantee

### What is Idempotency?

A migration is **idempotent** if it can be executed multiple times safely without errors or side effects.

### Idempotency Rules

✅ **IDEMPOTENT (Good):**

```sql
-- Safe: CREATE TABLE IF NOT EXISTS
CREATE TABLE IF NOT EXISTS pganalytics.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE
);

-- Safe: ALTER TABLE with IF NOT EXISTS
ALTER TABLE pganalytics.users
    ADD COLUMN IF NOT EXISTS email VARCHAR(255);

-- Safe: CREATE INDEX IF NOT EXISTS
CREATE INDEX IF NOT EXISTS idx_users_email
    ON pganalytics.users(email);

-- Safe: INSERT with ON CONFLICT
INSERT INTO pganalytics.metric_types (name, description)
    VALUES ('cpu_usage', 'CPU usage metrics')
    ON CONFLICT DO NOTHING;
```

❌ **NOT IDEMPOTENT (Bad):**

```sql
-- Fails on second run: table already exists
CREATE TABLE pganalytics.users (
    id SERIAL PRIMARY KEY
);

-- Fails on second run: column already exists
ALTER TABLE pganalytics.users
    ADD COLUMN email VARCHAR(255);

-- Fails if data already exists
INSERT INTO pganalytics.users (username, email)
    VALUES ('admin', 'admin@example.com');
```

### How Idempotency Works

The migration runner tracks executed migrations:

```sql
SELECT * FROM pganalytics.schema_versions;

version                 | executed_at              | execution_time_ms
------------------------|--------------------------|-------------------
000_complete_schema     | 2026-03-12 12:30:00     | 1250
```

**Before executing any migration:**
1. Check if `version` exists in `schema_versions` table
2. If YES: Skip the migration (already ran)
3. If NO: Execute the migration, then record in table

**This ensures migrations never run twice, even if:**
- You restart the backend
- You deploy to multiple servers
- You retry after a failure

### Consequence of Breaking Idempotency

If you write a migration that's NOT idempotent:

```bash
# First deployment: WORKS
# Migration creates the users table successfully

# Second deployment: FAILS with "ERROR: relation already exists"
# Backend refuses to start (or starts with warning)
# Data loss risk if migration is partially executed
```

---

## Deployment Checklist

### Before Deploying

- [ ] All migrations use `IF NOT EXISTS / IF EXISTS`
- [ ] All tables explicitly in `pganalytics` schema
- [ ] Migration file follows naming: `NNN_description.sql`
- [ ] No hardcoded passwords or secrets
- [ ] No breaking changes to existing migrations
- [ ] Migration tested on fresh database
- [ ] Code compiles without errors
- [ ] Docker image builds successfully

### Deployment Steps

1. **Build Docker image**
   ```bash
   docker-compose -f docker-compose.staging.yml build --no-cache backend-staging
   ```

2. **Start containers**
   ```bash
   docker-compose -f docker-compose.staging.yml up -d
   ```

3. **Wait for migrations to execute**
   ```bash
   sleep 20  # Wait for backend to start and run migrations
   ```

4. **Verify database**
   ```bash
   docker exec pganalytics-staging-postgres psql -U postgres -d pganalytics_staging \
     -c "SELECT COUNT(*) as table_count FROM information_schema.tables
         WHERE table_schema = 'pganalytics';"
   ```

5. **Check migration execution**
   ```bash
   docker exec pganalytics-staging-postgres psql -U postgres -d pganalytics_staging \
     -c "SELECT version, executed_at FROM pganalytics.schema_versions
         ORDER BY executed_at;"
   ```

6. **Create first admin user**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/setup \
     -H "Content-Type: application/json" \
     -d '{
       "username": "admin",
       "email": "admin@example.com",
       "password": "Admin@Secure123",
       "full_name": "Administrator"
     }'
   ```

7. **Log in and change password**
   - Go to http://localhost:3000
   - Login with credentials from setup
   - Change password when prompted
   - Access dashboard

### Production Deployment

For production, follow the same steps with:
- Production database credentials
- Production TLS certificates
- Production backend configuration
- Proper backups before migrations

---

## Troubleshooting

### Problem: "relation does not exist"

**Cause:** Migration didn't execute or tables are in wrong schema

**Solution:**
1. Check schema is `pganalytics`, not `public`
2. Verify migrations ran: `SELECT * FROM pganalytics.schema_versions;`
3. Check backend logs for migration errors
4. Ensure all tables are created with `pganalytics.` prefix

### Problem: "column already exists"

**Cause:** Migration is not idempotent (ran twice without IF NOT EXISTS)

**Solution:**
1. Fix migration to use `IF NOT EXISTS`
2. Delete from schema_versions to re-run: `DELETE FROM pganalytics.schema_versions WHERE version = '...';`
3. Restart backend

### Problem: Backend fails to start with migration error

**Cause:** SQL syntax error or permission issue

**Solution:**
1. Check backend logs: `docker-compose logs backend-staging`
2. Look for `ERROR` or `panic` messages
3. Verify SQL syntax in migration file
4. Ensure migrations directory is mounted
5. Check PostgreSQL is running and accessible

### Problem: Migrations not executing

**Cause:** Migration files not found or wrong extension

**Solution:**
1. Check file is named correctly: `NNN_description.sql` (not `.disabled` or `.backup`)
2. Check migrations directory exists: `/app/migrations/` in container
3. Verify in Docker: `docker exec container ls /app/migrations/`
4. Check backend logs for "Found migrations" message

---

## Summary of Critical Rules

1. ✅ **Idempotency First**: Every migration must be idempotent (safe to run multiple times)
2. ✅ **Schema Always Explicit**: Use `pganalytics.` prefix for all tables
3. ✅ **Versioning Strict**: Follow `NNN_description.sql` naming exactly
4. ✅ **One Change Per Migration**: Keep migrations focused and single-purpose
5. ✅ **Use IF NOT EXISTS**: Always use IF NOT EXISTS / IF EXISTS clauses
6. ✅ **No Hardcoded Data**: Credentials, passwords, secrets go in setup/config, not migrations
7. ✅ **Test Before Deploy**: Test migrations on clean database before production
8. ✅ **Document Changes**: Add comments explaining what each migration does

**Remember:** These rules exist to prevent data loss and ensure reliable deployments across multiple environments. Follow them religiously.

---

**Document Version:** 1.0
**Last Updated:** 2026-03-12
**Applies To:** pgAnalytics v3.0+
