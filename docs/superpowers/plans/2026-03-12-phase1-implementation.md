# Phase 1: Email Alerts, Log Analysis & Extended Metrics Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement three critical features to achieve feature parity with pganalytics_community: Email Alert System, PostgreSQL Log Analysis, and Extended PostgreSQL Metrics.

**Timeline:** 2 weeks (14 days)

**Architecture:**
- Email system: SMTP integration with templated notifications
- Log analysis: PostgreSQL log collector with parsing and dashboard
- Extended metrics: New collectors for WAL, archival, background writer, postmaster stats

**Tech Stack:** Go, PostgreSQL, TimescaleDB, React, Grafana, SMTP

---

## File Structure & Decomposition

### New Files to Create

#### Backend (Go)
```
backend/internal/notifications/
  ├── email_service.go          # SMTP integration, email sending
  ├── email_templates.go        # Email template rendering
  ├── notification_sender.go    # Notification dispatch
  └── notification_models.go    # Email/notification data structures

backend/internal/logs/
  ├── log_collector.go          # PostgreSQL log collection
  ├── log_parser.go             # Log event parsing and analysis
  ├── log_store.go              # Log storage operations
  └── log_models.go             # Log data structures

backend/internal/storage/
  ├── metrics_extended_store.go # WAL, archival, postmaster metrics
  └── log_store.go              # (new file for logs)

backend/internal/api/
  ├── handlers_alerts.go        # Alert notification handlers (new)
  ├── handlers_logs.go          # Log query handlers (new)
  └── handlers_extended_metrics.go # Extended metrics endpoints (new)
```

#### Database Migrations
```
backend/migrations/
  ├── 020_email_alerts_system.sql       # Email configuration and history
  ├── 021_postgresql_logs.sql           # Log storage schema
  └── 022_extended_metrics.sql          # WAL, archival, postmaster metrics
```

#### Frontend (React)
```
frontend/src/pages/
  ├── AlertNotificationsPage.tsx        # Email alert configuration
  ├── LogsAnalyzerPage.tsx              # Log analysis and search

frontend/src/components/
  ├── EmailAlertConfig.tsx              # Email settings form
  ├── LogEventTable.tsx                 # Log display with filtering
  └── LogAnalysisCharts.tsx             # Log trend visualizations

frontend/src/api/
  ├── emailAlertsApi.ts                 # Email alert endpoints
  └── logsApi.ts                        # Log query endpoints
```

#### Grafana Dashboards
```
grafana/dashboards/
  ├── wal-archival-monitoring.json      # WAL and archival stats
  ├── background-writer-stats.json      # Background writer metrics
  ├── postmaster-metrics.json           # Postmaster and startup metrics
  └── postgresql-logs-analysis.json     # Log visualization dashboard
```

#### Documentation
```
docs/
  ├── features/
  │   ├── EMAIL_ALERTS_SETUP.md         # Email configuration guide
  │   ├── LOG_ANALYSIS.md               # Log analysis features
  │   └── EXTENDED_METRICS.md           # New metrics documentation
  └── superpowers/
      └── plans/
          └── 2026-03-12-phase1-implementation.md (this file)
```

---

## Implementation Tasks

### CHUNK 1: Database Schema Extensions (2 days)

#### Task 1: Create Email Alerts Schema

**Files:**
- Create: `backend/migrations/020_email_alerts_system.sql`
- Modify: `backend/internal/storage/postgres.go` (add email alert queries)
- Test: Unit test for migration execution

- [ ] **Step 1: Write email alerts schema SQL**

Create `backend/migrations/020_email_alerts_system.sql`:

```sql
-- Email configuration table
CREATE TABLE pganalytics.email_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    smtp_host VARCHAR(255) NOT NULL,
    smtp_port INTEGER NOT NULL DEFAULT 587,
    smtp_username VARCHAR(255) NOT NULL,
    smtp_password TEXT NOT NULL,  -- Encrypted in application
    smtp_from_address VARCHAR(255) NOT NULL,
    smtp_from_name VARCHAR(255),
    use_tls BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Email alert recipients
CREATE TABLE pganalytics.email_alert_recipients (
    id SERIAL PRIMARY KEY,
    alert_rule_id INTEGER NOT NULL REFERENCES pganalytics.alert_rules(id) ON DELETE CASCADE,
    email_address VARCHAR(255) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,  -- critical, warning, info
    notification_mode VARCHAR(50) DEFAULT 'immediate',  -- immediate, daily, weekly
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Email alert history (for audit trail)
CREATE TABLE pganalytics.email_alert_history (
    id BIGSERIAL PRIMARY KEY,
    alert_rule_id INTEGER REFERENCES pganalytics.alert_rules(id) ON DELETE SET NULL,
    recipient_email VARCHAR(255) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    sent_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL,  -- sent, failed, bounced
    error_message TEXT,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP
);

-- Create index for history queries
CREATE INDEX idx_email_alert_history_rule_id ON pganalytics.email_alert_history(alert_rule_id);
CREATE INDEX idx_email_alert_history_sent_at ON pganalytics.email_alert_history(sent_at);

-- Notification channels extension
ALTER TABLE pganalytics.notification_channels ADD COLUMN IF NOT EXISTS
    smtp_config_id INTEGER REFERENCES pganalytics.email_configs(id) ON DELETE SET NULL;
```

- [ ] **Step 2: Write migration execution test**

Create test in `backend/internal/storage/migration_test.go`:

```go
func TestEmailAlertsMigration(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Run migration
    err := db.Exec(`
        CREATE TABLE IF NOT EXISTS pganalytics.email_configs (
            id SERIAL PRIMARY KEY,
            smtp_host VARCHAR(255) NOT NULL
        );
    `).Error

    if err != nil {
        t.Fatalf("migration failed: %v", err)
    }

    // Verify table exists
    var tableName string
    err = db.Raw("SELECT table_name FROM information_schema.tables WHERE table_name = 'email_configs'").Scan(&tableName).Error
    if err != nil || tableName == "" {
        t.Fatal("email_configs table not created")
    }
}
```

- [ ] **Step 3: Add migration to migrations.go**

Modify `backend/internal/storage/migrations.go` to include new migration file

- [ ] **Step 4: Run migration and verify**

```bash
make migrate-up
psql -U postgres -d pganalytics -c "\dt pganalytics.email*"
```

Expected: Tables `email_configs`, `email_alert_recipients`, `email_alert_history` visible

- [ ] **Step 5: Commit migration**

```bash
git add backend/migrations/020_email_alerts_system.sql
git add backend/internal/storage/migrations.go
git commit -m "feat: add email alerts schema migration"
```

---

#### Task 2: Create PostgreSQL Logs Schema

**Files:**
- Create: `backend/migrations/021_postgresql_logs.sql`
- Test: Verify schema creation

- [ ] **Step 1: Write PostgreSQL logs schema**

Create `backend/migrations/021_postgresql_logs.sql`:

```sql
-- PostgreSQL logs table with hypertable support
CREATE TABLE IF NOT EXISTS pganalytics.postgresql_logs (
    time TIMESTAMP NOT NULL,
    server_id INTEGER REFERENCES pganalytics.servers(id) ON DELETE SET NULL,
    database_name VARCHAR(255),
    user_name VARCHAR(255),
    application_name VARCHAR(255),
    process_id INTEGER,
    connection_from VARCHAR(255),
    session_id VARCHAR(255),
    log_level VARCHAR(50),  -- LOG, WARNING, ERROR, FATAL, PANIC
    message TEXT NOT NULL,
    detail TEXT,
    hint TEXT,
    context TEXT,
    query TEXT,
    query_pos INTEGER,
    location VARCHAR(255),
    function_name VARCHAR(255),
    error_code VARCHAR(10),  -- SQLSTATE error code
    parsed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common log queries
CREATE INDEX IF NOT EXISTS idx_postgresql_logs_time ON pganalytics.postgresql_logs(time DESC);
CREATE INDEX IF NOT EXISTS idx_postgresql_logs_server_id ON pganalytics.postgresql_logs(server_id);
CREATE INDEX IF NOT EXISTS idx_postgresql_logs_log_level ON pganalytics.postgresql_logs(log_level);
CREATE INDEX IF NOT EXISTS idx_postgresql_logs_error_code ON pganalytics.postgresql_logs(error_code);

-- Log aggregation table (hourly summaries)
CREATE TABLE IF NOT EXISTS pganalytics.log_events_hourly (
    hour TIMESTAMP NOT NULL,
    server_id INTEGER REFERENCES pganalytics.servers(id) ON DELETE SET NULL,
    log_level VARCHAR(50),
    error_code VARCHAR(10),
    event_count INTEGER DEFAULT 1,
    PRIMARY KEY (hour, server_id, log_level, error_code)
);

-- Log statistics view
CREATE OR REPLACE VIEW pganalytics.log_stats_hourly AS
SELECT
    DATE_TRUNC('hour', time) as hour,
    server_id,
    log_level,
    error_code,
    COUNT(*) as event_count,
    COUNT(DISTINCT database_name) as databases_affected,
    COUNT(DISTINCT user_name) as users_affected
FROM pganalytics.postgresql_logs
GROUP BY DATE_TRUNC('hour', time), server_id, log_level, error_code;
```

- [ ] **Step 2: Add log schema migration to migrations.go**

- [ ] **Step 3: Run migration and verify**

```bash
make migrate-up
psql -U postgres -d pganalytics -c "\dt pganalytics.postgresql_logs"
```

- [ ] **Step 4: Commit**

```bash
git add backend/migrations/021_postgresql_logs.sql
git commit -m "feat: add postgresql logs schema and indexes"
```

---

#### Task 3: Create Extended Metrics Schema

**Files:**
- Create: `backend/migrations/022_extended_metrics.sql`

- [ ] **Step 1: Write extended metrics schema**

Create `backend/migrations/022_extended_metrics.sql`:

```sql
-- WAL (Write-Ahead Log) statistics table
CREATE TABLE IF NOT EXISTS pganalytics.wal_stats (
    time TIMESTAMP NOT NULL,
    server_id INTEGER REFERENCES pganalytics.servers(id) ON DELETE SET NULL,
    wal_generated_bytes BIGINT,
    wal_records INTEGER,
    wal_fpi INTEGER,  -- Full page images
    wal_buffers_full INTEGER,
    wal_write_time_ms BIGINT,
    wal_sync_time_ms BIGINT
);

-- Archive statistics table
CREATE TABLE IF NOT EXISTS pganalytics.archive_stats (
    time TIMESTAMP NOT NULL,
    server_id INTEGER REFERENCES pganalytics.servers(id) ON DELETE SET NULL,
    archive_count INTEGER,
    archive_success_count INTEGER,
    archive_failed_count INTEGER,
    last_archived_wal VARCHAR(255),
    last_archive_time TIMESTAMP,
    archive_status VARCHAR(50)  -- active, paused, failed
);

-- Background writer statistics
CREATE TABLE IF NOT EXISTS pganalytics.bgwriter_stats (
    time TIMESTAMP NOT NULL,
    server_id INTEGER REFERENCES pganalytics.servers(id) ON DELETE SET NULL,
    checkpoints_timed INTEGER,
    checkpoints_req INTEGER,
    checkpoint_write_time_ms BIGINT,
    checkpoint_sync_time_ms BIGINT,
    buffers_checkpoint INTEGER,
    buffers_clean INTEGER,
    buffers_backend INTEGER,
    buffers_backend_fsync INTEGER,
    buffers_alloc INTEGER,
    maxwritten_clean INTEGER
);

-- Postmaster metrics (startup, uptime, restart count)
CREATE TABLE IF NOT EXISTS pganalytics.postmaster_stats (
    time TIMESTAMP NOT NULL,
    server_id INTEGER REFERENCES pganalytics.servers(id) ON DELETE SET NULL,
    postmaster_start_time TIMESTAMP,
    uptime_seconds BIGINT,
    restart_count INTEGER,
    pg_version VARCHAR(50),
    pg_version_num INTEGER
);

-- Create indexes for time-based queries
CREATE INDEX IF NOT EXISTS idx_wal_stats_time ON pganalytics.wal_stats(time DESC);
CREATE INDEX IF NOT EXISTS idx_wal_stats_server ON pganalytics.wal_stats(server_id);
CREATE INDEX IF NOT EXISTS idx_archive_stats_time ON pganalytics.archive_stats(time DESC);
CREATE INDEX IF NOT EXISTS idx_archive_stats_server ON pganalytics.archive_stats(server_id);
CREATE INDEX IF NOT EXISTS idx_bgwriter_stats_time ON pganalytics.bgwriter_stats(time DESC);
CREATE INDEX IF NOT EXISTS idx_bgwriter_stats_server ON pganalytics.bgwriter_stats(server_id);
CREATE INDEX IF NOT EXISTS idx_postmaster_stats_time ON pganalytics.postmaster_stats(time DESC);
```

- [ ] **Step 2: Add migration and verify**

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/022_extended_metrics.sql
git commit -m "feat: add WAL, archive, background writer, and postmaster metrics schema"
```

---

### CHUNK 2: Backend Services Implementation (4 days)

#### Task 4: Email Service Implementation

**Files:**
- Create: `backend/internal/notifications/email_service.go`
- Create: `backend/internal/notifications/email_templates.go`
- Create: `backend/internal/notifications/notification_models.go`
- Modify: `backend/internal/storage/postgres.go` (add email config queries)

- [ ] **Step 1: Create email models**

Create `backend/internal/notifications/notification_models.go`:

```go
package notifications

import "time"

// EmailConfig represents SMTP configuration
type EmailConfig struct {
    ID            int       `json:"id"`
    Name          string    `json:"name"`
    SMTPHost      string    `json:"smtp_host"`
    SMTPPort      int       `json:"smtp_port"`
    SMTPUsername  string    `json:"smtp_username"`
    SMTPPassword  string    `json:"-"` // Never expose password
    FromAddress   string    `json:"from_address"`
    FromName      string    `json:"from_name"`
    UseTLS        bool      `json:"use_tls"`
    IsActive      bool      `json:"is_active"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

// AlertEmailNotification represents an email to send
type AlertEmailNotification struct {
    To           string
    Subject      string
    TemplateName string
    Data         map[string]interface{}
    HTMLBody     string
    PlainBody    string
    SentAt       time.Time
    Status       string // sent, failed
    Error        string
}

// EmailAlertHistory tracks sent emails
type EmailAlertHistory struct {
    ID              int64
    AlertRuleID     int
    RecipientEmail  string
    Subject         string
    SentAt          time.Time
    Status          string
    ErrorMessage    string
}
```

- [ ] **Step 2: Implement email service**

Create `backend/internal/notifications/email_service.go`:

```go
package notifications

import (
    "crypto/tls"
    "fmt"
    "net/smtp"
    "strings"
    "time"

    "github.com/pganalytics/pganalytics-v3/backend/internal/storage"
)

// EmailService handles SMTP operations
type EmailService struct {
    config *EmailConfig
    store  *storage.PostgresDB
}

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig, store *storage.PostgresDB) *EmailService {
    return &EmailService{
        config: config,
        store:  store,
    }
}

// SendEmail sends an email via SMTP
func (es *EmailService) SendEmail(notification *AlertEmailNotification) error {
    if !es.config.IsActive {
        return fmt.Errorf("email config is inactive")
    }

    // Build SMTP address
    smtpAddr := fmt.Sprintf("%s:%d", es.config.SMTPHost, es.config.SMTPPort)

    // Create SMTP client
    var conn *smtp.Client
    var err error

    if es.config.UseTLS {
        tlsConfig := &tls.Config{
            ServerName: es.config.SMTPHost,
        }
        conn, err = tls.Dial("tcp", smtpAddr, tlsConfig)
        if err != nil {
            return fmt.Errorf("failed to establish TLS connection: %w", err)
        }

        client, err := smtp.NewClient(conn, es.config.SMTPHost)
        if err != nil {
            return fmt.Errorf("failed to create SMTP client: %w", err)
        }
        defer client.Close()

        // Authenticate
        auth := smtp.PlainAuth("", es.config.SMTPUsername, es.config.SMTPPassword, es.config.SMTPHost)
        if err := client.Auth(auth); err != nil {
            return fmt.Errorf("authentication failed: %w", err)
        }

        // Send email
        if err := client.Mail(es.config.FromAddress); err != nil {
            return fmt.Errorf("failed to set from address: %w", err)
        }

        if err := client.Rcpt(notification.To); err != nil {
            return fmt.Errorf("failed to set recipient: %w", err)
        }

        w, err := client.Data()
        if err != nil {
            return fmt.Errorf("failed to get data writer: %w", err)
        }
        defer w.Close()

        // Write email headers and body
        msg := buildMessage(es.config, notification)
        if _, err := w.Write([]byte(msg)); err != nil {
            return fmt.Errorf("failed to write message: %w", err)
        }
    } else {
        // Non-TLS connection
        client, err := smtp.Dial(smtpAddr)
        if err != nil {
            return fmt.Errorf("failed to connect to SMTP: %w", err)
        }
        defer client.Close()

        // Authenticate if required
        if es.config.SMTPUsername != "" {
            auth := smtp.PlainAuth("", es.config.SMTPUsername, es.config.SMTPPassword, es.config.SMTPHost)
            if err := client.Auth(auth); err != nil {
                return fmt.Errorf("authentication failed: %w", err)
            }
        }

        // Send email
        if err := client.Mail(es.config.FromAddress); err != nil {
            return fmt.Errorf("failed to set from address: %w", err)
        }

        if err := client.Rcpt(notification.To); err != nil {
            return fmt.Errorf("failed to set recipient: %w", err)
        }

        w, err := client.Data()
        if err != nil {
            return fmt.Errorf("failed to get data writer: %w", err)
        }
        defer w.Close()

        msg := buildMessage(es.config, notification)
        if _, err := w.Write([]byte(msg)); err != nil {
            return fmt.Errorf("failed to write message: %w", err)
        }
    }

    // Log success
    notification.Status = "sent"
    notification.SentAt = time.Now()

    return nil
}

// buildMessage constructs the email message with headers
func buildMessage(config *EmailConfig, notification *AlertEmailNotification) string {
    var sb strings.Builder

    sb.WriteString(fmt.Sprintf("From: %s <%s>\r\n", config.FromName, config.FromAddress))
    sb.WriteString(fmt.Sprintf("To: %s\r\n", notification.To))
    sb.WriteString(fmt.Sprintf("Subject: %s\r\n", notification.Subject))
    sb.WriteString("MIME-Version: 1.0\r\n")
    sb.WriteString("Content-Type: multipart/alternative; boundary=\"boundary\"\r\n")
    sb.WriteString("\r\n")

    // Plain text part
    sb.WriteString("--boundary\r\n")
    sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
    sb.WriteString("\r\n")
    sb.WriteString(notification.PlainBody)
    sb.WriteString("\r\n")

    // HTML part
    sb.WriteString("--boundary\r\n")
    sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
    sb.WriteString("\r\n")
    sb.WriteString(notification.HTMLBody)
    sb.WriteString("\r\n")

    sb.WriteString("--boundary--\r\n")

    return sb.String()
}
```

- [ ] **Step 3: Implement email templates**

Create `backend/internal/notifications/email_templates.go`:

```go
package notifications

import (
    "bytes"
    "html/template"
    "time"
)

// EmailTemplate represents an email template
type EmailTemplate struct {
    Name         string
    Subject      string
    PlainBody    string
    HTMLTemplate string
}

// GetTemplates returns all email templates
func GetTemplates() map[string]*EmailTemplate {
    return map[string]*EmailTemplate{
        "alert": {
            Name:    "alert",
            Subject: "[ALERT] PostgreSQL Health Warning on {{.InstanceName}}",
            PlainBody: `Alert: {{.AlertName}}
Instance: {{.InstanceName}}
Severity: {{.Severity}}
Time: {{.Timestamp}}

Issue: {{.Message}}

Details:
{{.Details}}

Please investigate and resolve this issue.
`,
            HTMLTemplate: `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .alert { border-left: 4px solid #ff6b6b; padding: 10px; background-color: #f5f5f5; }
        .severity { font-weight: bold; color: {{.SeverityColor}}; }
        .timestamp { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <h2>PostgreSQL Alert Notification</h2>
    <div class="alert">
        <p><strong>Alert:</strong> {{.AlertName}}</p>
        <p><strong>Instance:</strong> {{.InstanceName}}</p>
        <p><strong>Severity:</strong> <span class="severity">{{.Severity}}</span></p>
        <p class="timestamp">{{.Timestamp}}</p>
        <p>{{.Message}}</p>
        <h3>Details:</h3>
        <pre>{{.Details}}</pre>
    </div>
    <p>Please log in to pgAnalytics to view more details.</p>
</body>
</html>`,
        },
    }
}

// RenderTemplate renders an email template
func RenderTemplate(templateName string, data map[string]interface{}) (string, string, error) {
    templates := GetTemplates()
    emailTpl, ok := templates[templateName]
    if !ok {
        return "", "", ErrTemplateNotFound
    }

    // Render plain text
    plainTpl, err := template.New("plain").Parse(emailTpl.PlainBody)
    if err != nil {
        return "", "", err
    }

    var plainBuf bytes.Buffer
    if err := plainTpl.Execute(&plainBuf, data); err != nil {
        return "", "", err
    }

    // Render HTML
    htmlTpl, err := template.New("html").Parse(emailTpl.HTMLTemplate)
    if err != nil {
        return "", "", err
    }

    var htmlBuf bytes.Buffer
    if err := htmlTpl.Execute(&htmlBuf, data); err != nil {
        return "", "", err
    }

    return plainBuf.String(), htmlBuf.String(), nil
}
```

- [ ] **Step 4: Add email config queries to postgres.go**

Modify `backend/internal/storage/postgres.go` to add:

```go
// GetEmailConfig retrieves email configuration
func (p *PostgresDB) GetEmailConfig(ctx context.Context, id int) (*EmailConfig, error) {
    var config EmailConfig
    err := p.db.WithContext(ctx).
        Table("pganalytics.email_configs").
        Where("id = ?", id).
        Scan(&config).Error

    if err != nil {
        return nil, apperrors.DatabaseError("get email config", err.Error())
    }

    return &config, nil
}

// SaveEmailAlertHistory saves sent email record
func (p *PostgresDB) SaveEmailAlertHistory(ctx context.Context, history *EmailAlertHistory) error {
    return p.db.WithContext(ctx).
        Table("pganalytics.email_alert_history").
        Create(history).Error
}
```

- [ ] **Step 5: Write unit tests**

Create `backend/internal/notifications/email_service_test.go`:

```go
package notifications

import (
    "testing"
)

func TestEmailService_SendEmail(t *testing.T) {
    config := &EmailConfig{
        ID:           1,
        SMTPHost:     "localhost",
        SMTPPort:     587,
        FromAddress:  "pganalytics@example.com",
        FromName:     "pgAnalytics",
        UseTLS:       false,
        IsActive:     true,
    }

    service := NewEmailService(config, nil)

    notification := &AlertEmailNotification{
        To:           "admin@example.com",
        Subject:      "Test Alert",
        TemplateName: "alert",
    }

    // This will fail without real SMTP, but tests the structure
    err := service.SendEmail(notification)
    if err == nil {
        t.Fatal("expected error for non-existent SMTP server")
    }
}

func TestRenderTemplate(t *testing.T) {
    data := map[string]interface{}{
        "AlertName":    "High CPU Usage",
        "InstanceName": "prod-db-01",
        "Severity":     "WARNING",
        "Message":      "CPU usage exceeded 80%",
        "Timestamp":    time.Now().Format(time.RFC3339),
    }

    plain, html, err := RenderTemplate("alert", data)
    if err != nil {
        t.Fatalf("failed to render template: %v", err)
    }

    if plain == "" || html == "" {
        t.Fatal("rendered templates are empty")
    }

    if !strings.Contains(plain, "High CPU Usage") {
        t.Fatal("template variable not rendered in plain text")
    }

    if !strings.Contains(html, "High CPU Usage") {
        t.Fatal("template variable not rendered in HTML")
    }
}
```

- [ ] **Step 6: Run tests**

```bash
cd backend
go test ./internal/notifications -v
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/notifications/
git add backend/internal/storage/postgres.go
git commit -m "feat: implement email notification service with SMTP and templates"
```

---

[Due to length constraints, I'll continue with the remaining tasks in a condensed format]

#### Task 5: Log Collection & Parsing (2 days)

- Create `backend/internal/logs/log_collector.go` - Collects PostgreSQL logs from stderr/csvlog
- Create `backend/internal/logs/log_parser.go` - Parses PostgreSQL CSV logs
- Create `backend/internal/logs/log_store.go` - Stores parsed logs in database
- Add handler in `backend/internal/api/handlers_logs.go` - API endpoints for log queries
- Add Grafana dashboard `grafana/dashboards/postgresql-logs-analysis.json`

Key implementation:
- Parse PostgreSQL CSV log format
- Extract ERROR, WARNING, FATAL events
- Store in postgresql_logs table
- Query via REST API with filtering (server, time range, log_level, error_code)

#### Task 6: Extended Metrics Collectors (2 days)

- Add WAL statistics collection to collector
- Add archival statistics collection
- Add background writer stats collection
- Add postmaster metrics collection
- Create Grafana dashboards for each metric type

Key implementation:
- Query pg_stat_wal_receiver, pg_stat_archiver, pg_stat_bgwriter
- Collect postmaster_start_time from pg_control_recovery()
- Store in respective tables
- Create time-series visualizations

---

### CHUNK 3: Frontend & Grafana Implementation (2 days)

#### Task 7: Email Alert Configuration UI

- Create `frontend/src/pages/AlertNotificationsPage.tsx` - Email settings interface
- Create `frontend/src/components/EmailAlertConfig.tsx` - Email form component
- Create `frontend/src/api/emailAlertsApi.ts` - API integration
- Add routes in App.tsx

#### Task 8: Log Analysis UI

- Create `frontend/src/pages/LogsAnalyzerPage.tsx` - Log search interface
- Create `frontend/src/components/LogEventTable.tsx` - Log display with pagination
- Create `frontend/src/components/LogAnalysisCharts.tsx` - Log trend charts
- Create `frontend/src/api/logsApi.ts` - API integration

#### Task 9: Grafana Dashboards

Create 4 new dashboards:
- `wal-archival-monitoring.json` - WAL generation and archival status
- `background-writer-stats.json` - Background writer performance
- `postmaster-metrics.json` - Uptime, restarts, version info
- `postgresql-logs-analysis.json` - Log event visualization

---

### CHUNK 4: Documentation & Testing (2 days)

#### Task 10: Documentation

- Create `docs/features/EMAIL_ALERTS_SETUP.md` - Email configuration guide
- Create `docs/features/LOG_ANALYSIS.md` - Log analysis features and usage
- Create `docs/features/EXTENDED_METRICS.md` - New metrics documentation
- Update main README with new features

#### Task 11: End-to-End Testing

- Integration test for email notifications
- Integration test for log collection and querying
- Integration test for extended metrics collection
- Staging deployment verification

#### Task 12: Final Commit & Cleanup

- Ensure all tests pass
- Verify code quality (0 vulnerabilities, EXCELLENT rating)
- Final commit and push to origin/main

---

## Testing Strategy

### Unit Tests
- Email service: Template rendering, message building
- Log parser: CSV parsing, event extraction
- Metrics collectors: Data structure validation

### Integration Tests
- Email sending via mock SMTP
- Log storage and retrieval
- Metrics collection and storage

### End-to-End Tests
- Fresh deployment with all Phase 1 features
- Email alert triggered and sent
- Logs collected and visible in dashboard
- New metrics appearing in Grafana

### Deployment Tests
- Docker Compose up with Phase 1 features
- Health check endpoints
- API responses
- Frontend UI loads without errors

---

## Quality Checklist

- [ ] All code follows Go/React conventions
- [ ] 100% of new code has unit tests
- [ ] All integration tests pass
- [ ] Code quality: EXCELLENT (no issues)
- [ ] Security audit: 0 vulnerabilities
- [ ] Database migrations are safe and reversible
- [ ] API documentation complete
- [ ] Frontend UI intuitive and responsive
- [ ] Grafana dashboards properly configured
- [ ] All changes committed with clear messages
- [ ] No breaking changes to existing features
- [ ] Performance acceptable (< 100ms API responses)

---

## Deployment Checklist

- [ ] All code merged to main branch
- [ ] Migrations tested in fresh environment
- [ ] Staging deployment successful (17/17 tests pass)
- [ ] All new features working in staging
- [ ] Documentation complete and clear
- [ ] Team reviewed and approved
- [ ] Ready for production deployment

---

## Success Criteria

**Feature Parity Achieved:**
- ✅ Email Alert System: SMTP integration, templated emails, delivery tracking
- ✅ PostgreSQL Log Analysis: Log collection, parsing, dashboard visualization
- ✅ Extended Metrics: WAL, archival, background writer, postmaster stats

**Quality Maintained:**
- ✅ Code quality: EXCELLENT
- ✅ Security: 0 vulnerabilities
- ✅ Testing: 100% coverage of new code
- ✅ Documentation: Complete

**Timeline Met:**
- 2 weeks from start to production ready

---

**Plan Status:** Ready for execution
**Estimated Effort:** 80 developer-hours
**Team Size:** 1-2 developers
**Start Date:** 2026-03-12
**Target Completion:** 2026-03-26

