# PostgreSQL Version Support Matrix

## Supported PostgreSQL Versions

pgAnalytics v3 supports the following PostgreSQL versions with full feature compatibility:

| Version | Release Date | End of Life | Status | Support Level |
|---------|--------------|-------------|--------|---------------|
| 14      | 2021-10-14   | 2026-10-14  | ✅     | Full Support  |
| 15      | 2022-10-13   | 2027-10-13  | ✅     | Full Support  |
| 16      | 2023-10-12   | 2028-10-12  | ✅     | Full Support  |
| 17      | 2024-10-10   | 2029-10-10  | ✅     | Full Support  |
| 18      | 2025-10-09   | 2030-10-09  | ✅     | Full Support  |

**Note**: End of Life dates follow PostgreSQL's official release cycle. We recommend upgrading before the EOL date.

## Version-Specific Features

All pgAnalytics features are compatible across supported versions. The following table shows availability:

| Feature | PG14 | PG15 | PG16 | PG17 | PG18 |
|---------|------|------|------|------|------|
| JSONB Operations | ✅ | ✅ | ✅ | ✅ | ✅ |
| Logical Replication | ✅ | ✅ | ✅ | ✅ | ✅ |
| Full Text Search | ✅ | ✅ | ✅ | ✅ | ✅ |
| Table Partitioning | ✅ | ✅ | ✅ | ✅ | ✅ |
| Window Functions | ✅ | ✅ | ✅ | ✅ | ✅ |
| Common Table Expressions (CTE) | ✅ | ✅ | ✅ | ✅ | ✅ |
| Materialized Views | ✅ | ✅ | ✅ | ✅ | ✅ |
| JSON Path (jsonpath) | ✅ | ✅ | ✅ | ✅ | ✅ |
| Generated Columns | ✅ | ✅ | ✅ | ✅ | ✅ |
| Range Types | ✅ | ✅ | ✅ | ✅ | ✅ |
| pg_stat_statements | ✅ | ✅ | ✅ | ✅ | ✅ |
| pgvector Extension | ✅ | ✅ | ✅ | ✅ | ✅ |

## PostgreSQL 14 (Current EOL: Oct 2026)

### Specifications
- **Latest Minor**: 14.13+
- **Compatibility**: Full backward compatibility with pgAnalytics v3
- **Performance**: Production-ready

### Installation

**macOS**:
```bash
brew install postgresql@14
brew services start postgresql@14
```

**Ubuntu/Debian**:
```bash
sudo apt update
sudo apt install postgresql-14 postgresql-contrib-14
sudo systemctl start postgresql
```

**CentOS/RHEL**:
```bash
sudo dnf install postgresql14-server postgresql14-contrib
sudo systemctl start postgresql
```

**Docker**:
```bash
docker run --name pganalytics-db \
  -e POSTGRES_DB=pganalytics \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:14
```

### Notable Features
- Improved query planner
- Logical replication enhancements
- Better PARTITION handling

## PostgreSQL 15 (Current EOL: Oct 2027)

### Specifications
- **Latest Minor**: 15.8+
- **Compatibility**: Full backward compatibility with pgAnalytics v3
- **Performance**: Production-ready, recommended

### Installation

**macOS**:
```bash
brew install postgresql@15
brew services start postgresql@15
```

**Ubuntu/Debian**:
```bash
sudo apt update
sudo apt install postgresql-15 postgresql-contrib-15
sudo systemctl start postgresql
```

**Docker**:
```bash
docker run --name pganalytics-db \
  -e POSTGRES_DB=pganalytics \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:15
```

### Notable Features
- Performance improvements for COPY
- Better EXPLAIN output
- Enhanced security features

### Migration from PostgreSQL 14

```bash
# Backup existing database
pg_dump -U postgres olddb > backup_14.sql

# Upgrade using pg_upgrade
pg_upgrade \
  -d /var/lib/postgresql/14/main \
  -D /var/lib/postgresql/15/main \
  -b /usr/lib/postgresql/14/bin \
  -B /usr/lib/postgresql/15/bin

# Restart PostgreSQL
sudo systemctl restart postgresql
```

## PostgreSQL 16 (Current EOL: Oct 2028)

### Specifications
- **Latest Minor**: 16.3+
- **Compatibility**: Full backward compatibility with pgAnalytics v3
- **Performance**: Production-ready, highly recommended

### Installation

**macOS**:
```bash
brew install postgresql@16
brew services start postgresql@16
```

**Ubuntu/Debian**:
```bash
sudo add-apt-repository "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main"
sudo apt update
sudo apt install postgresql-16 postgresql-contrib-16
sudo systemctl start postgresql
```

**Docker**:
```bash
docker run --name pganalytics-db \
  -e POSTGRES_DB=pganalytics \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:16
```

### Notable Features
- Significant performance improvements
- Better concurrency handling
- SQL/JSON path expressions
- Enhanced monitoring with pg_stat_io

### Migration from PostgreSQL 15

```bash
# Using pg_upgrade (recommended)
pg_upgrade \
  -d /var/lib/postgresql/15/main \
  -D /var/lib/postgresql/16/main \
  -b /usr/lib/postgresql/15/bin \
  -B /usr/lib/postgresql/16/bin \
  -P 5432 \
  -p 5433

# Run analyze on new cluster
/usr/lib/postgresql/16/bin/vacuumdb -U postgres --all --analyze
```

## PostgreSQL 17 (Current EOL: Oct 2029)

### Specifications
- **Latest Minor**: 17.1+
- **Compatibility**: Full backward compatibility with pgAnalytics v3
- **Performance**: Production-ready, cutting-edge

### Installation

**macOS**:
```bash
brew install postgresql@17
brew services start postgresql@17
```

**Ubuntu/Debian**:
```bash
sudo add-apt-repository "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main"
sudo apt update
sudo apt install postgresql-17 postgresql-contrib-17
sudo systemctl start postgresql
```

**Docker**:
```bash
docker run --name pganalytics-db \
  -e POSTGRES_DB=pganalytics \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:17
```

### Notable Features
- Further performance improvements
- Enhanced EXPLAIN planning
- Better connection pooling
- Logical replication improvements

## PostgreSQL 18 (Current EOL: Oct 2030)

### Specifications
- **Latest Minor**: 18.0+
- **Compatibility**: Full forward compatibility with pgAnalytics v3
- **Performance**: Production-ready, latest release

### Installation

**macOS**:
```bash
brew install postgresql@18
brew services start postgresql@18
```

**Ubuntu/Debian**:
```bash
sudo add-apt-repository "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main"
sudo apt update
sudo apt install postgresql-18 postgresql-contrib-18
sudo systemctl start postgresql
```

**Docker**:
```bash
docker run --name pganalytics-db \
  -e POSTGRES_DB=pganalytics \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  postgres:18
```

### Notable Features
- Latest security patches
- Performance optimizations
- New SQL features
- Enhanced monitoring capabilities

## Testing Against Multiple Versions

### Automated Testing

pgAnalytics includes tests against all supported versions:

```bash
# Test against all supported versions
mise run test:postgres:all

# Test specific version
POSTGRES_VERSION=16 mise run test

# Test compatibility
./scripts/test-postgres-versions.sh
```

### Docker Compose Multi-Version Testing

```yaml
services:
  postgres-14:
    image: postgres:14
    environment:
      POSTGRES_DB: pganalytics

  postgres-15:
    image: postgres:15
    environment:
      POSTGRES_DB: pganalytics

  postgres-16:
    image: postgres:16
    environment:
      POSTGRES_DB: pganalytics
```

### Manual Testing

```bash
# Start specific PostgreSQL version
docker run -d --name pg16-test -e POSTGRES_PASSWORD=test postgres:16

# Connect and test
psql -h localhost -U postgres -c "SELECT version();"

# Clean up
docker stop pg16-test && docker rm pg16-test
```

## Upgrading PostgreSQL

### In-Place Upgrade (Minor Versions)

Minor version upgrades (14.x to 14.y) are simple:

```bash
# Ubuntu
sudo apt update
sudo apt install postgresql-14
sudo systemctl restart postgresql

# Verify
psql --version
```

### Major Version Upgrade

Major version upgrades (14 to 15) require migration:

#### Method 1: pg_upgrade (Fastest)

```bash
# Stop PostgreSQL
sudo systemctl stop postgresql

# Run pg_upgrade
sudo -u postgres pg_upgrade \
  -d /var/lib/postgresql/14/main \
  -D /var/lib/postgresql/15/main \
  -b /usr/lib/postgresql/14/bin \
  -B /usr/lib/postgresql/15/bin

# Verify
sudo -u postgres /usr/lib/postgresql/15/bin/vacuumdb --all --analyze

# Remove old cluster
sudo -u postgres rm -rf /var/lib/postgresql/14/main

# Start PostgreSQL
sudo systemctl start postgresql

# Verify
psql --version
```

#### Method 2: Dump and Restore (Safer)

```bash
# Create backup
sudo -u postgres pg_dump -U postgres pganalytics > backup.sql

# Install new version
sudo apt install postgresql-15

# Create empty database
sudo -u postgres psql -U postgres << EOF
DROP DATABASE IF EXISTS pganalytics;
CREATE DATABASE pganalytics;
EOF

# Restore data
sudo -u postgres psql -U postgres pganalytics < backup.sql

# Verify
psql -U postgres pganalytics -c "SELECT COUNT(*) FROM information_schema.tables;"
```

#### Method 3: Logical Replication (Zero Downtime)

```bash
# 1. Set up logical replication on source (PG14)
ALTER SYSTEM SET wal_level = logical;
sudo systemctl restart postgresql

# 2. Create publication
psql -U postgres pganalytics << EOF
CREATE PUBLICATION all_tables FOR ALL TABLES;
EOF

# 3. Set up new PostgreSQL 15 instance
# 4. Create subscription
psql -U postgres -h new-host pganalytics << EOF
CREATE SUBSCRIPTION pganalytics_sub CONNECTION 'host=old-host dbname=pganalytics' PUBLICATION all_tables;
EOF

# 5. Monitor replication
SELECT * FROM pg_stat_replication;

# 6. Switch applications to new host
# 7. Drop subscription on new host
DROP SUBSCRIPTION pganalytics_sub;
```

## Migration Best Practices

### Before Upgrade
- [ ] Backup all data
- [ ] Test in staging environment
- [ ] Document current configuration
- [ ] Review release notes for breaking changes
- [ ] Disable automatic vacuuming

### During Upgrade
- [ ] Schedule during maintenance window
- [ ] Keep network connectivity stable
- [ ] Monitor system resources
- [ ] Have rollback plan ready

### After Upgrade
- [ ] Verify data integrity
- [ ] Run ANALYZE on all tables
- [ ] Update database statistics
- [ ] Test all applications
- [ ] Monitor performance

## Compatibility Notes

### Schema Compatibility

All pgAnalytics schemas are compatible across versions:

```bash
# Check schema compatibility
./scripts/check-schema-compatibility.sh --from 14 --to 16
```

### Extension Compatibility

Required extensions work on all versions:

- `pgvector`: Supports PG14+
- `uuid-ossp`: Supports all versions
- `pg_stat_statements`: Supports all versions
- `hstore`: Supports all versions

### Connection String Format

The connection string format is identical across versions:

```
postgres://username:password@host:port/database?sslmode=require
```

## Troubleshooting Version Issues

### Incompatible Data Type

```sql
-- Check for incompatible types
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_schema = 'public';
```

### Encoding Issues

```bash
# Verify encoding
psql -U postgres -c "SELECT datname, pg_encoding_to_char(encoding) FROM pg_database;"

# Convert if needed
ALTER DATABASE pganalytics SET client_encoding = 'UTF8';
```

### Performance Regression After Upgrade

```bash
# Regenerate statistics
ANALYZE;

# Reindex if needed
REINDEX DATABASE pganalytics;

# Check execution plans
EXPLAIN ANALYZE SELECT * FROM your_table;
```

## Version Recommendation

For new deployments, we recommend:

- **Development**: PostgreSQL 17 or 18 (latest features)
- **Staging**: Same as production
- **Production**: PostgreSQL 16 (proven stability, modern features)

## Support Timeline

- **PostgreSQL 14**: Support until October 2026
- **PostgreSQL 15**: Support until October 2027
- **PostgreSQL 16**: Support until October 2028
- **PostgreSQL 17**: Support until October 2029
- **PostgreSQL 18**: Support until October 2030

Plan upgrades to avoid EOL without supported path forward.

## Additional Resources

- [PostgreSQL Official Downloads](https://www.postgresql.org/download/)
- [PostgreSQL Release Notes](https://www.postgresql.org/docs/release/)
- [PostgreSQL Upgrade Documentation](https://www.postgresql.org/docs/current/upgrading.html)
- [pg_upgrade Documentation](https://www.postgresql.org/docs/current/pgupgrade.html)
- [pgAnalytics Compatibility Report](https://github.com/pganalytics/pganalytics-v3/wiki/Postgres-Compatibility)
