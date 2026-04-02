# PostgreSQL Version Support Matrix

## Supported Versions

pgAnalytics v3 is fully compatible with PostgreSQL versions **14, 15, 16, 17, and 18**.

| PostgreSQL Version | Status | Notes |
|-------------------|--------|-------|
| 14 | ✅ **Full Support** | All features available, optimized baseline |
| 15 | ✅ **Full Support** | Enhanced performance, ICU collations available |
| 16 | ✅ **Full Support** | Production recommended, improved JSON operators |
| 17 | ✅ **Full Support** | Latest stable features, enhanced window functions |
| 18 | ✅ **Full Support** | Latest release, full feature parity |

## Feature Compatibility Matrix

This matrix documents feature availability across all supported PostgreSQL versions. All features are available in **all supported versions** unless noted otherwise.

| Feature | PG 14 | PG 15 | PG 16 | PG 17 | PG 18 | Notes |
|---------|-------|-------|-------|-------|-------|-------|
| **Core Extensions** |
| uuid-ossp | ✅ | ✅ | ✅ | ✅ | ✅ | UUID generation |
| pgcrypto | ✅ | ✅ | ✅ | ✅ | ✅ | Cryptographic functions |
| pg_trgm | ✅ | ✅ | ✅ | ✅ | ✅ | Trigram text search |
| btree_gin | ✅ | ✅ | ✅ | ✅ | ✅ | GIN index support |
| **Data Types** |
| BIGSERIAL | ✅ | ✅ | ✅ | ✅ | ✅ | 64-bit auto-increment |
| UUID | ✅ | ✅ | ✅ | ✅ | ✅ | Universally unique identifiers |
| JSONB | ✅ | ✅ | ✅ | ✅ | ✅ | Binary JSON storage |
| TIMESTAMP WITH TIME ZONE | ✅ | ✅ | ✅ | ✅ | ✅ | TZ-aware timestamps |
| DECIMAL/NUMERIC | ✅ | ✅ | ✅ | ✅ | ✅ | Precise numeric values |
| INTERVAL | ✅ | ✅ | ✅ | ✅ | ✅ | Time intervals |
| BYTEA | ✅ | ✅ | ✅ | ✅ | ✅ | Binary data |
| TEXT[] | ✅ | ✅ | ✅ | ✅ | ✅ | Text arrays |
| **Schema Features** |
| Schemas (CREATE SCHEMA) | ✅ | ✅ | ✅ | ✅ | ✅ | Namespace support |
| Constraints (PRIMARY KEY, UNIQUE, FOREIGN KEY) | ✅ | ✅ | ✅ | ✅ | ✅ | Data integrity |
| CHECK constraints | ✅ | ✅ | ✅ | ✅ | ✅ | Column validation |
| PARTIAL INDEXES (WHERE clause) | ✅ | ✅ | ✅ | ✅ | ✅ | Conditional indexes |
| **Functions & Procedures** |
| PL/pgSQL | ✅ | ✅ | ✅ | ✅ | ✅ | Procedural language |
| CREATE OR REPLACE FUNCTION | ✅ | ✅ | ✅ | ✅ | ✅ | Function management |
| TRIGGER functions | ✅ | ✅ | ✅ | ✅ | ✅ | Event triggers |
| BEFORE/AFTER triggers | ✅ | ✅ | ✅ | ✅ | ✅ | Trigger timing |
| FOR EACH ROW triggers | ✅ | ✅ | ✅ | ✅ | ✅ | Row-level triggers |
| **Transactions & Locking** |
| Transactions (BEGIN/COMMIT) | ✅ | ✅ | ✅ | ✅ | ✅ | ACID compliance |
| MVCC | ✅ | ✅ | ✅ | ✅ | ✅ | Multi-version concurrency |
| Row-level locking | ✅ | ✅ | ✅ | ✅ | ✅ | Pessimistic locking |
| **Views** |
| CREATE VIEW | ✅ | ✅ | ✅ | ✅ | ✅ | Standard views |
| CREATE OR REPLACE VIEW | ✅ | ✅ | ✅ | ✅ | ✅ | View updates |
| Materialized views | ✅ | ✅ | ✅ | ✅ | ✅ | Cached views |
| **Query Features** |
| Aggregate functions (SUM, COUNT, AVG, etc.) | ✅ | ✅ | ✅ | ✅ | ✅ | Analytics |
| Window functions (OVER, PARTITION BY) | ✅ | ✅ | ✅ | ✅ | ✅ | Advanced analytics |
| DISTINCT ON | ✅ | ✅ | ✅ | ✅ | ✅ | PostgreSQL extension |
| CASE expressions | ✅ | ✅ | ✅ | ✅ | ✅ | Conditional logic |
| JOINs (INNER, LEFT, RIGHT, FULL) | ✅ | ✅ | ✅ | ✅ | ✅ | Join operations |
| CTE (WITH clause) | ✅ | ✅ | ✅ | ✅ | ✅ | Common table expressions |

## Per-Version Release Notes

### PostgreSQL 14 (Initial Support)
- **Release Date**: October 2021
- **End of Support**: October 2026
- **Baseline for pgAnalytics v3 compatibility**
- All core features fully tested and working
- Recommended minimum version for production

### PostgreSQL 15 (Enhanced Support)
- **Release Date**: October 2022
- **End of Support**: October 2027
- **Key Improvements**:
  - Enhanced JSON operators with improved performance
  - Improved sorting algorithms
  - Better handling of DISTINCT and window functions
- **New Features**:
  - ICU collations for better international text handling
  - Improved JSONB performance
  - All features 100% compatible with pgAnalytics

### PostgreSQL 16 (Recommended Version)
- **Release Date**: October 2023
- **End of Support**: October 2028
- **Current Production Standard for pgAnalytics**
- **Key Improvements**:
  - Performance improvements in JSON operations
  - Better TOAST compression
  - Enhanced query planner
  - Improved index efficiency
- **New Features**:
  - Enhanced JSON/JSONB operators
  - Improved window function performance
  - Better INTERVAL type handling
- **All features tested and verified working**

### PostgreSQL 17 (Latest Stable)
- **Release Date**: October 2024
- **End of Support**: October 2029
- **Key Improvements**:
  - Performance improvements in aggregate functions
  - Enhanced window function optimization
  - Better memory management for large datasets
- **Fully compatible with pgAnalytics v3**
- Excellent choice for new deployments

### PostgreSQL 18 (Future Ready)
- **Release Date**: October 2025
- **End of Support**: October 2030
- **Status**: Full support and testing completed
- **Future-proof deployments**
- All pgAnalytics v3 features fully operational

## Migration Compatibility Analysis

All SQL migrations (000-027) are fully compatible with PostgreSQL 14-18:

### Migration Categories

**Core Schema (Migration 000)**
- Uses standard SQL DDL statements
- No version-specific features
- Compatible since PostgreSQL 12+

**Triggers & Functions (Migration 001)**
- Standard PL/pgSQL syntax
- No version-specific trigger features
- BEFORE UPDATE triggers work identically across versions

**PostgreSQL Logs Schema (Migration 021)**
- Standard table and index creation
- BIGSERIAL primary keys (PG 14+)
- INTERVAL columns supported in all versions
- WHERE clauses in indexes (PG 12+)

**Realtime Tables (Migration 022)**
- JSONB columns fully supported in all versions
- GIN indexes for JSONB (PG 12+)
- No version-specific features

**Phase 4 Tables (Migration 023)**
- Standard table structures
- Unique constraints (PG 14+)
- No breaking changes across versions

**Query Performance Schema (Migration 024)**
- BIGSERIAL primary keys
- JSONB plan_json column
- Compatible across all versions

**Log Analysis Schema (Migration 025)**
- TEXT[] array types (PG 14+)
- Standard table creation syntax
- Fully compatible

**Index Advisor (Migration 026)**
- TEXT[] for column_names arrays
- Standard table operations
- No version-specific features

**Vacuum Advisor (Migration 027)**
- DECIMAL type with precision
- INTERVAL type for autovacuum_naptime
- Fully compatible across all versions

## Performance Notes by Version

| Version | Strengths | Considerations |
|---------|-----------|-----------------|
| PG 14 | Stable baseline, well-tested | Baseline performance |
| PG 15 | Improved JSON handling, ICU support | Slightly improved analytics performance |
| PG 16 | Better query planning, improved indexes | Recommended for new deployments |
| PG 17 | Optimized aggregates, better window functions | Best for complex queries |
| PG 18 | Latest optimizations, best performance | Production-ready, future-proof |

## Known Limitations

**None identified.** All migrations are fully compatible with PostgreSQL 14-18.

## Testing Status

✅ **All migrations tested and verified** against:
- PostgreSQL 14.x
- PostgreSQL 15.x
- PostgreSQL 16.x
- PostgreSQL 17.x
- PostgreSQL 18.x

## Deployment Recommendations

### For New Deployments
- **Recommended**: PostgreSQL 16 or 17
- **Rationale**: Optimal balance of stability, performance, and long-term support
- **LTS**: All versions have 5+ years of support

### For Existing Deployments
- **PG 14 Users**: Can upgrade to PG 15, 16, 17, or 18 without issues
- **PG 15 Users**: Can upgrade to PG 16, 17, or 18 without issues
- **No downtime required** for minor version upgrades within the same major version
- **Planned outage required** for major version upgrades (use PostgreSQL pg_upgrade utility)

### Upgrade Path Examples

```
PostgreSQL 14 -> 15 -> 16 -> 17 -> 18  ✅ Fully supported
PostgreSQL 14 -> 17 (skipping versions)  ✅ Compatible
PostgreSQL 14 -> 18 (major jump)  ✅ Compatible
```

## Environment Variables & Configuration

All standard PostgreSQL connection parameters work identically across versions:

```bash
DATABASE_URL="postgres://user:password@host:5432/database?sslmode=require"
TEST_DATABASE_URL="postgres://user:password@localhost:5432/test_database"
```

No version-specific environment variables required.

## Docker Compatibility

All Docker image tags are supported:

```dockerfile
# All equivalent and fully supported
FROM postgres:14-bullseye
FROM postgres:15-bullseye
FROM postgres:16-bullseye
FROM postgres:17-bullseye
FROM postgres:18-bullseye
```

## Support & Documentation

For PostgreSQL version-specific documentation, see:
- [PostgreSQL 14 Docs](https://www.postgresql.org/docs/14/)
- [PostgreSQL 15 Docs](https://www.postgresql.org/docs/15/)
- [PostgreSQL 16 Docs](https://www.postgresql.org/docs/16/)
- [PostgreSQL 17 Docs](https://www.postgresql.org/docs/17/)
- [PostgreSQL 18 Docs](https://www.postgresql.org/docs/18/)

## Summary

**pgAnalytics v3 maintains 100% compatibility** with PostgreSQL versions 14 through 18. All migrations are version-agnostic, using standard SQL that works identically across all supported versions. Choose your PostgreSQL version based on operational preferences and support timelines rather than pgAnalytics compatibility concerns.

---

**Last Updated**: 2026-04-02
**Status**: ✅ All versions tested and verified
