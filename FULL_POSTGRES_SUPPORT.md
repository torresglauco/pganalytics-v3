# Full PostgreSQL 14-18 Support

## Overview
pgAnalytics v3 provides COMPLETE support for monitoring PostgreSQL 14, 15, 16, 17, and 18.

This means:
1. Backend can run on any of these PostgreSQL versions
2. Collector can connect to and monitor any of these PostgreSQL versions
3. All features work identically across all versions
4. Zero version-specific limitations

## Supported Versions

| Version | Release | EOL | Recommendation | Status |
|---------|---------|-----|-----------------|--------|
| 14 | 2021-10 | 2026-10 | Baseline support | ✅ Full |
| 15 | 2022-10 | 2027-10 | Upgrade path | ✅ Full |
| 16 | 2023-10 | 2028-10 | **RECOMMENDED** | ✅ Full |
| 17 | 2024-10 | 2029-10 | Latest stable | ✅ Full |
| 18 | 2025-10 | TBD | Future-proof | ✅ Full |

## What is Supported

### Collector Features (Works on all versions)

#### Query Monitoring
- pg_stat_statements integration
- Query extraction and analysis
- Execution time tracking
- Execution count tracking
- Query fingerprinting for aggregation
- EXPLAIN ANALYZE plan capture
- Performance trending

#### Log Collection
- PostgreSQL server logs
- Slow query detection (configured via log_min_duration_statement)
- Error extraction
- Warning detection
- Application-level events

#### Metrics Collection
- Table metrics (size, row count, dead rows)
- Index metrics (size, usage, bloat)
- Replication metrics (if configured)
- Connection metrics
- Transaction metrics
- Cache metrics

#### Analysis & Recommendations
- Query performance analysis
- Index recommendations
- Missing index detection
- Unused index identification
- Table bloat detection
- VACUUM recommendations
- Query optimization suggestions

### Backend Features (Analyzes data from all versions)

#### Query Performance
- Latency analysis
- Bottleneck detection
- Performance trending
- Comparative analysis

#### Anomaly Detection
- Statistical anomaly detection
- Threshold-based alerts
- Pattern-based detection
- Baseline comparison

#### Index Optimization
- Missing index suggestions
- Index efficiency analysis
- Index maintenance recommendations

#### Maintenance Advisor
- VACUUM scheduling
- ANALYZE timing
- Bloat management
- Resource optimization

## Compatibility Details

### Extensions Required
All required extensions work on PostgreSQL 14-18:
- uuid-ossp: ✅ All versions
- pgcrypto: ✅ All versions
- pg_stat_statements: ✅ All versions
- btree_gin: ✅ All versions

### Query Compatibility
All queries used by Collector work on all versions:
- pg_stat_statements queries: ✅
- information_schema queries: ✅
- pg_catalog queries: ✅
- System function calls: ✅

### Wire Protocol
PostgreSQL wire protocol is backwards compatible:
- PG 14 wire protocol: ✅ Supported
- PG 15 protocol: ✅ Supported
- PG 16 protocol: ✅ Supported
- PG 17 protocol: ✅ Supported
- PG 18 protocol: ✅ Supported

## Testing

All features have been tested against:
- ✅ PostgreSQL 14.x
- ✅ PostgreSQL 15.x
- ✅ PostgreSQL 16.x
- ✅ PostgreSQL 17.x
- ✅ PostgreSQL 18.x

## Known Limitations

None identified. Full support across all features.

## Performance Notes

Performance is consistent across versions. No version-specific tuning required.

## Security

All security features are available on all versions:
- SSL/TLS connections: ✅ All versions
- Authentication: ✅ All versions
- Authorization: ✅ All versions

## Support & Troubleshooting

See POSTGRES_VERSIONS.md for version-specific deployment notes.

## Version Selection

**Recommended**: PostgreSQL 16
**Why**: Good balance of stability and features

**For New Deployments**: PostgreSQL 17 or later
**Why**: Latest features and performance improvements

**For Legacy Systems**: PostgreSQL 14
**Why**: Works perfectly with older installations

All versions receive full support.
