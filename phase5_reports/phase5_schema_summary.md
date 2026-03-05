# Phase 5 Schema Features

## Anomaly Detection Tables
- `query_baselines`: Statistical baselines for query metrics
  - Stores mean, stddev, min, max, percentiles
  - Calculates rolling 7-day (168-hour) window
  - Enables Z-score analysis

- `query_anomalies`: Detected anomalies
  - Z-score, deviation percentage tracking
  - Severity levels: low, medium, high, critical
  - Status management (active, resolved)

## Alert Rules & Notifications
- `alert_rules`: Rule definitions
  - Multiple types: threshold, change, anomaly, composite
  - Flexible JSON conditions
  - Notification channel assignments

- `fired_alerts`: Alert instances
  - Status tracking (firing, alerting, resolved, acknowledged)
  - Fingerprinting for deduplication
  - Context capture

- `notification_channels`: Multi-channel delivery
  - Email, Slack, Teams, PagerDuty, webhooks
  - Rate limiting and batching
  - Delivery tracking

## Enterprise Auth (Phase 3)
- OAuth, SAML, LDAP support
- MFA/2FA implementation
- JWT token management
- Session management

## Data Encryption (Phase 3)
- Column-level encryption
- Key rotation management
- Encrypted field tracking
- Audit logging

## Audit Logging (Phase 3)
- All admin operations tracked
- User action history
- Database change tracking
- Compliance reporting

## Phase 4 Optimizations
- TimescaleDB hypertables for metrics
- Advanced caching with TTL management
- Circuit breaker pattern for external services
- Rate limiting with token bucket algorithm
