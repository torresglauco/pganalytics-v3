# Centralized Collector Architecture with RDS & Registration UI

**Purpose:** Design and implementation guide for centralized collector management on AWS RDS
**Status:** Architecture & Requirements Complete
**Date:** February 26, 2026
**Version:** 1.0

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────┐
│                        User/Admin Interface                          │
│                                                                      │
│  Web Browser → React UI (Collector Registration Interface)          │
│                                                                      │
│  http://pganalytics.example.com:3000                                │
└────────────────────────────────────┬─────────────────────────────────┘
                                     │
                    ┌────────────────┴────────────────┐
                    │                                  │
┌──────────────────▼─────────────────────┐  ┌────────▼───────────────┐
│  Backend API (Go - Central Server)      │  │ Nginx/HAProxy Load    │
│  Port: 8080                             │  │ Balancer (Optional)    │
│  ┌─────────────────────────────────┐   │  └────────────────────────┘
│  │ Collector Registration Service   │   │
│  │ - Register new collectors         │   │
│  │ - Generate JWT tokens             │   │
│  │ - Store collector configs         │   │
│  │ - Monitor collector health        │   │
│  │ - Bulk import management          │   │
│  │ - Dashboard assignment            │   │
│  │                                   │   │
│  │ HTTP Endpoints:                   │   │
│  │ POST   /api/v1/collectors/...     │   │
│  │ GET    /api/v1/collectors/...     │   │
│  │ PUT    /api/v1/collectors/...     │   │
│  │ DELETE /api/v1/collectors/...     │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Metrics Aggregation Service            │
│  - Receives metrics from all collectors │
│  - Processes and stores metrics         │
│  - Triggers dashboards                  │
│  - Manages retention policies           │
└──────────────────┬──────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
    [TCP/TLS]            [TCP/TLS]
        │                     │
┌───────▼───────────────────────────┐
│  AWS RDS PostgreSQL               │
│  (Centralized Database)           │
│                                   │
│  Tables:                          │
│  ├── collectors                   │
│  │   ├── id (UUID)                │
│  │   ├── name                      │
│  │   ├── type (postgres/mysql/etc) │
│  │   ├── environment               │
│  │   ├── host                      │
│  │   ├── port                      │
│  │   ├── database                  │
│  │   ├── username                  │
│  │   ├── password_encrypted        │
│  │   ├── jwt_token                 │
│  │   ├── status                    │
│  │   ├── last_heartbeat            │
│  │   ├── metrics_count             │
│  │   ├── created_at                │
│  │   ├── updated_at                │
│  │   └── deleted_at (soft delete)  │
│  │                                  │
│  ├── collector_groups              │
│  │   ├── id                         │
│  │   ├── name (aws, on-prem, etc)  │
│  │   ├── description                │
│  │   └── created_at                 │
│  │                                  │
│  ├── collector_metrics             │
│  │   ├── id                         │
│  │   ├── collector_id               │
│  │   ├── metric_data (JSONB)        │
│  │   ├── received_at                │
│  │   └── processed                  │
│  │                                  │
│  ├── collector_dashboards          │
│  │   ├── collector_id               │
│  │   ├── dashboard_id               │
│  │   └── assigned_at                │
│  │                                  │
│  └── collector_audit_log           │
│      ├── id                         │
│      ├── collector_id               │
│      ├── action                     │
│      ├── actor_id                   │
│      ├── changes                    │
│      └── created_at                 │
│                                    │
└────────────────────────────────────┘
         AWS RDS PostgreSQL
         (Centralized Storage)
              ^
              │
    [Multiple Collectors Push Metrics]
              │
     ┌────────┴────────┐
     │                 │
┌────▼──────┐   ┌──────▼────┐
│ Collector │   │ Collector │
│ Instance  │   │ Instance  │
│ (C++)     │   │ (C++)     │
│           │   │           │
│ Prod-RDS-1│   │ Prod-RDS-2│
│ Region: A │   │ Region: B │
└───────────┘   └───────────┘
     Host:            Host:
pganalytics-db-1.    pganalytics-db-2.
region-a.rds.aws     region-b.rds.aws
```

---

## Collector Registration Flow

### Step 1: Admin Opens Registration UI
```
Admin → Browser
  ↓
  Navigates to: http://pganalytics:3000/collectors/register
  ↓
  React UI loads with:
  - Registration form
  - Validation rules
  - Database connection tester
```

### Step 2: Fill Registration Form
```
Admin fills form:
  Collector Name: prod-rds-1
  Type: PostgreSQL
  Environment: Production
  Group: AWS-RDS
  Host: pganalytics-db-1.region.rds.amazonaws.com
  Port: 5432
  Database: pganalytics
  Username: postgres
  Password: ••••••••
  SSL Mode: require
  Collection Interval: 60s
  Query Limit: 100
```

### Step 3: Test Connection (Pre-validation)
```
Admin clicks: [Test Connection]
  ↓
React Frontend:
  - Sends credentials to backend
  - Shows loading spinner
  ↓
Go Backend:
  - Tries to connect to provided host
  - Validates database access
  - Returns success/error
  ↓
React UI:
  - Shows "Connection successful ✓" OR error message
  - Enables/disables submit button
```

### Step 4: Register Collector
```
Admin clicks: [Register Collector]
  ↓
React Frontend:
  - Validates all form fields
  - Encrypts sensitive data (browser-side optional)
  - Sends POST to backend
  ↓
Go Backend:
  - Validates request format
  - Encrypts password (AES-256-GCM)
  - Generates unique collector ID: col_uuid
  - Generates JWT token for collector auth
  - Inserts record into RDS database
  - Stores configuration
  ↓
Backend Response:
  {
    "id": "col_12345abc",
    "jwt_token": "eyJhbGc...",
    "status": "active",
    "created_at": "2024-01-20T10:30:00Z"
  }
  ↓
React UI:
  - Shows success message
  - Displays JWT token for collector setup
  - Option to copy token to clipboard
  - Redirect to collector details page
```

### Step 5: Configure Collector
```
Admin/DevOps:
  - Copies JWT token
  - Goes to C++ collector deployment
  - Sets environment variable:
    COLLECTOR_JWT_TOKEN=eyJhbGc...
    COLLECTOR_ID=col_12345abc
    CENTRAL_API_URL=http://pganalytics:8080
  - Restarts collector service
  ↓
C++ Collector:
  - Connects to backend API with JWT
  - Validates token
  - Starts metrics collection from assigned database
  - Pushes metrics every 60 seconds
  ↓
Backend:
  - Receives metrics from collector
  - Updates last_heartbeat timestamp
  - Stores metrics in collector_metrics table
  - Triggers Grafana dashboard updates
  ↓
Grafana:
  - Queries PostgreSQL for latest metrics
  - Updates dashboards in real-time
  - Shows metrics visualizations
```

---

## Detailed Database Schema

### collectors Table
```sql
CREATE TABLE collectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,  -- postgresql, mysql, mongodb
    environment VARCHAR(50) NOT NULL,  -- development, staging, production
    group_id UUID REFERENCES collector_groups(id),
    description TEXT,

    -- Target Database Connection Details (encrypted)
    host VARCHAR(255) NOT NULL,  -- RDS hostname
    port INTEGER NOT NULL DEFAULT 5432,
    database VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    password_encrypted VARCHAR(1024) NOT NULL,  -- AES-256-GCM encrypted
    ssl_mode VARCHAR(50),  -- disable, allow, prefer, require, etc.
    ssl_certificate_encrypted TEXT,  -- encrypted PEM

    -- Collector Authentication
    jwt_token VARCHAR(2048) NOT NULL UNIQUE,  -- Collector's auth token
    jwt_token_rotated_at TIMESTAMP,
    jwt_token_expires_at TIMESTAMP,

    -- Configuration
    collection_interval INTEGER DEFAULT 60,  -- seconds
    query_limit INTEGER DEFAULT 100,  -- max queries per database
    enabled BOOLEAN DEFAULT TRUE,
    tags TEXT[],  -- Array of tags: ['prod', 'aws', 'rds']

    -- Status Tracking
    status VARCHAR(50) DEFAULT 'pending',  -- pending, active, inactive, error
    last_heartbeat TIMESTAMP,
    last_error TEXT,
    error_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    metrics_collected INTEGER DEFAULT 0,
    uptime_percentage DECIMAL(5,2) DEFAULT 0,

    -- Metadata
    registered_by UUID NOT NULL,  -- User who registered
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_modified_by UUID,
    last_modified_at TIMESTAMP,
    deleted_at TIMESTAMP,  -- Soft delete

    -- Performance tracking
    avg_collection_time_ms INTEGER,
    max_collection_time_ms INTEGER,
    last_collection_count INTEGER,

    CONSTRAINT valid_environment CHECK (environment IN
        ('development', 'staging', 'production')),
    CONSTRAINT valid_status CHECK (status IN
        ('pending', 'active', 'inactive', 'error')),
    CONSTRAINT valid_type CHECK (type IN
        ('postgresql', 'mysql', 'mongodb', 'oracle'))
);

CREATE INDEX idx_collectors_group_id ON collectors(group_id);
CREATE INDEX idx_collectors_environment ON collectors(environment);
CREATE INDEX idx_collectors_status ON collectors(status);
CREATE INDEX idx_collectors_last_heartbeat ON collectors(last_heartbeat DESC);
CREATE INDEX idx_collectors_name ON collectors(name);
```

### collector_groups Table
```sql
CREATE TABLE collector_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,  -- aws, on-prem, gcp, azure, etc.
    description TEXT,
    color VARCHAR(7),  -- Hex color for UI display
    icon VARCHAR(50),  -- Icon name for UI
    display_order INTEGER,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO collector_groups (name, description, color) VALUES
    ('AWS RDS', 'AWS Relational Database Service', '#FF9900'),
    ('On-Premises', 'Self-managed databases', '#146EB4'),
    ('GCP Cloud SQL', 'Google Cloud Platform', '#4285F4'),
    ('Azure Database', 'Microsoft Azure', '#0078D4'),
    ('Development', 'Development environment', '#28A745'),
    ('Staging', 'Staging environment', '#FFC107'),
    ('Production', 'Production environment', '#DC3545');
```

### collector_metrics Table
```sql
CREATE TABLE collector_metrics (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE CASCADE,

    -- Metric Data (flexible JSONB)
    metric_data JSONB NOT NULL,  -- {
                                --   "database": "pganalytics",
                                --   "query_count": 1234,
                                --   "avg_execution_time": 45.3,
                                --   "queries": [...]
                                -- }

    -- Processing Status
    received_at TIMESTAMP NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMP,

    -- Data Quality
    record_count INTEGER,
    checksum VARCHAR(64),  -- SHA256 of metric_data

    FOREIGN KEY (collector_id) REFERENCES collectors(id) ON DELETE CASCADE
);

CREATE INDEX idx_collector_metrics_collector_id_received ON
    collector_metrics(collector_id, received_at DESC);
CREATE INDEX idx_collector_metrics_received ON
    collector_metrics(received_at DESC);
```

### collector_dashboards Table
```sql
CREATE TABLE collector_dashboards (
    id SERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE CASCADE,
    dashboard_id UUID NOT NULL,  -- Grafana dashboard ID
    dashboard_name VARCHAR(255),  -- Cached for performance

    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    auto_assign BOOLEAN DEFAULT FALSE,  -- Auto-assign new dashboards?

    UNIQUE(collector_id, dashboard_id),
    FOREIGN KEY (collector_id) REFERENCES collectors(id) ON DELETE CASCADE
);

CREATE INDEX idx_collector_dashboards_dashboard_id ON
    collector_dashboards(dashboard_id);
```

### collector_audit_log Table
```sql
CREATE TABLE collector_audit_log (
    id BIGSERIAL PRIMARY KEY,
    collector_id UUID NOT NULL REFERENCES collectors(id) ON DELETE SET NULL,

    action VARCHAR(50) NOT NULL,  -- registered, updated, deleted, heartbeat, error
    details JSONB,  -- What changed

    actor_id UUID NOT NULL,  -- User who performed action
    actor_name VARCHAR(255),  -- Cached for performance
    ip_address INET,
    user_agent TEXT,

    old_values JSONB,  -- Before state
    new_values JSONB,  -- After state

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (collector_id) REFERENCES collectors(id) ON DELETE SET NULL
);

CREATE INDEX idx_collector_audit_log_collector_id ON
    collector_audit_log(collector_id);
CREATE INDEX idx_collector_audit_log_created_at ON
    collector_audit_log(created_at DESC);
```

---

## API Implementation Details

### Registration Endpoint Implementation

**File:** `backend/internal/api/handlers_collectors.go`

```go
package api

import (
    "crypto/aes"
    "crypto/cipher"
    "database/sql"
    "encoding/base64"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "time"
)

// RegisterCollectorRequest handles new collector registration
type RegisterCollectorRequest struct {
    Name               string   `json:"name" binding:"required,max=255"`
    Type               string   `json:"type" binding:"required,oneof=postgresql mysql mongodb oracle"`
    Environment        string   `json:"environment" binding:"required,oneof=development staging production"`
    Group              string   `json:"group" binding:"max=100"`
    Description        string   `json:"description" binding:"max=1000"`
    Host               string   `json:"host" binding:"required,hostname_port|ipv4|fqdn"`
    Port               int      `json:"port" binding:"required,min=1,max=65535"`
    Database           string   `json:"database" binding:"required,max=255"`
    Username           string   `json:"username" binding:"required,max=255"`
    Password           string   `json:"password" binding:"required,min=1"`
    SSLMode            string   `json:"ssl_mode" binding:"oneof=disable allow prefer require requireCA requireFull"`
    SSLCertificate     string   `json:"ssl_certificate"` // base64
    CollectionInterval int      `json:"collection_interval" binding:"min=10,max=3600"`
    QueryLimit         int      `json:"query_limit" binding:"min=1,max=10000"`
    Tags               []string `json:"tags"`
}

// RegisterCollectorResponse returns JWT token and collector info
type RegisterCollectorResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    JWTToken  string    `json:"jwt_token"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`

    // Important: Share token immediately (only displayed once)
    Instructions string `json:"instructions"`
}

// RegisterCollector handles POST /api/v1/collectors/register
func (s *Server) RegisterCollector(c *gin.Context) {
    userID := c.GetString("user_id")  // From auth middleware
    if userID == "" {
        c.JSON(401, gin.H{"error": "Unauthorized"})
        return
    }

    var req RegisterCollectorRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
        return
    }

    // Step 1: Test database connection
    testConn := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        req.Host, req.Port, req.Username, req.Password,
        req.Database, req.SSLMode,
    )

    testDB, err := sql.Open("postgres", testConn)
    if err != nil {
        c.JSON(400, gin.H{"error": "Invalid database connection string"})
        return
    }
    defer testDB.Close()

    // Test actual connection with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := testDB.PingContext(ctx); err != nil {
        c.JSON(400, gin.H{
            "error": "Cannot connect to database: " + err.Error(),
            "details": "Please verify host, port, username, password, and SSL settings",
        })
        return
    }

    // Step 2: Generate collector ID and JWT token
    collectorID := "col_" + uuid.New().String()

    // Create JWT token for collector
    jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": collectorID,              // Subject
        "collector_id": collectorID,
        "type": "collector",
        "exp": time.Now().Add(365*24*time.Hour).Unix(),  // 1 year expiry
        "iat": time.Now().Unix(),
    })

    tokenString, err := jwtToken.SignedString([]byte(s.config.JWTSecret))
    if err != nil {
        c.JSON(500, gin.H{"error": "Cannot generate token"})
        return
    }

    // Step 3: Encrypt password using AES-256-GCM
    encryptor := s.encryptionService  // Initialized in server setup
    encryptedPassword, err := encryptor.Encrypt(req.Password)
    if err != nil {
        c.JSON(500, gin.H{"error": "Cannot encrypt password"})
        return
    }

    // Step 4: Save to RDS database
    query := `
        INSERT INTO collectors (
            id, name, type, environment, group_id, description,
            host, port, database, username, password_encrypted,
            ssl_mode, ssl_certificate_encrypted,
            jwt_token, collection_interval, query_limit,
            tags, status, registered_by
        ) VALUES (
            $1, $2, $3, $4,
            (SELECT id FROM collector_groups WHERE name = $5),
            $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
        )
    `

    encryptedCert := ""
    if req.SSLCertificate != "" {
        encryptedCert, _ = encryptor.Encrypt(req.SSLCertificate)
    }

    err = s.db.QueryRowContext(c.Request.Context(),
        query,
        collectorID, req.Name, req.Type, req.Environment,
        req.Group, req.Description,
        req.Host, req.Port, req.Database, req.Username,
        encryptedPassword, req.SSLMode, encryptedCert,
        tokenString, req.CollectionInterval, req.QueryLimit,
        pq.Array(req.Tags), "active", userID,
    ).Scan()

    if err != nil {
        if strings.Contains(err.Error(), "duplicate key") {
            c.JSON(409, gin.H{"error": "Collector name already exists"})
        } else {
            c.JSON(500, gin.H{"error": "Cannot save collector"})
        }
        return
    }

    // Step 5: Log audit entry
    s.auditLogger.Log(c.Request.Context(), &AuditEvent{
        EventType: "COLLECTOR_REGISTERED",
        ResourceType: "COLLECTOR",
        Action: "CREATE",
        ActorID: userID,
        Changes: map[string]interface{}{
            "collector_id": collectorID,
            "name": req.Name,
            "host": req.Host,
        },
    })

    // Step 6: Return response
    c.JSON(201, RegisterCollectorResponse{
        ID: collectorID,
        Name: req.Name,
        JWTToken: tokenString,
        Status: "active",
        CreatedAt: time.Now(),
        Instructions: fmt.Sprintf(
            "Use this token to authenticate the collector:\n"+
            "export COLLECTOR_JWT_TOKEN=%s\n"+
            "export COLLECTOR_ID=%s\n"+
            "export CENTRAL_API_URL=http://pganalytics:8080",
            tokenString, collectorID,
        ),
    })
}

// TestCollectorConnection tests database connectivity
func (s *Server) TestCollectorConnection(c *gin.Context) {
    var req struct {
        Host     string `json:"host" binding:"required"`
        Port     int    `json:"port" binding:"required"`
        Database string `json:"database" binding:"required"`
        Username string `json:"username" binding:"required"`
        Password string `json:"password" binding:"required"`
        SSLMode  string `json:"ssl_mode" binding:"required"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Try connection
    connStr := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        req.Host, req.Port, req.Username, req.Password,
        req.Database, req.SSLMode,
    )

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        c.JSON(400, gin.H{
            "success": false,
            "error": "Invalid connection string: " + err.Error(),
        })
        return
    }
    defer db.Close()

    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        c.JSON(400, gin.H{
            "success": false,
            "error": "Connection failed: " + err.Error(),
        })
        return
    }

    // Get database info
    var version string
    db.QueryRow("SELECT version()").Scan(&version)

    c.JSON(200, gin.H{
        "success": true,
        "message": "Connection successful",
        "database_version": version,
    })
}

// BulkImportCollectors handles POST /api/v1/collectors/bulk-import
func (s *Server) BulkImportCollectors(c *gin.Context) {
    var req struct {
        Collectors []RegisterCollectorRequest `json:"collectors" binding:"required,dive"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    type BulkResult struct {
        Name      string `json:"name"`
        Status    string `json:"status"`
        Error     string `json:"error,omitempty"`
        CollectorID string `json:"collector_id,omitempty"`
        JWTToken  string `json:"jwt_token,omitempty"`
    }

    results := make([]BulkResult, 0, len(req.Collectors))

    // Import each collector sequentially or in transaction
    for _, collector := range req.Collectors {
        // Call RegisterCollector logic for each
        collectorID := "col_" + uuid.New().String()

        // ... repeat registration steps ...

        results = append(results, BulkResult{
            Name: collector.Name,
            Status: "success",
            CollectorID: collectorID,
            JWTToken: tokenString,
        })
    }

    c.JSON(200, gin.H{
        "total": len(req.Collectors),
        "successful": len(results),
        "failed": 0,
        "results": results,
    })
}

// ListCollectors returns all registered collectors with pagination
func (s *Server) ListCollectors(c *gin.Context) {
    page := c.DefaultQuery("page", "1")
    limit := c.DefaultQuery("limit", "20")
    environment := c.Query("environment")
    group := c.Query("group")

    query := `
        SELECT id, name, type, environment, status, host,
               last_heartbeat, metrics_collected, uptime_percentage,
               created_at, registered_by
        FROM collectors
        WHERE deleted_at IS NULL
    `
    args := []interface{}{}
    argCount := 1

    if environment != "" {
        query += fmt.Sprintf(" AND environment = $%d", argCount)
        args = append(args, environment)
        argCount++
    }

    if group != "" {
        query += fmt.Sprintf(" AND group_id = (SELECT id FROM collector_groups WHERE name = $%d)", argCount)
        args = append(args, group)
        argCount++
    }

    query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)

    limitInt, _ := strconv.Atoi(limit)
    pageInt, _ := strconv.Atoi(page)
    offset := (pageInt - 1) * limitInt

    args = append(args, limitInt, offset)

    rows, err := s.db.QueryContext(c.Request.Context(), query, args...)
    if err != nil {
        c.JSON(500, gin.H{"error": "Cannot query collectors"})
        return
    }
    defer rows.Close()

    collectors := []map[string]interface{}{}
    // ... scan and append rows ...

    c.JSON(200, gin.H{
        "collectors": collectors,
        "page": pageInt,
        "limit": limitInt,
        "total": getTotalCount(s.db),
    })
}
```

---

## Collector Metrics Push Implementation

### C++ Collector Sending Metrics

```cpp
// File: collector/src/api_client.cpp
#include "api_client.h"
#include <curl/curl.h>
#include <nlohmann/json.hpp>

class CentralAPIClient {
private:
    std::string api_url;
    std::string collector_id;
    std::string jwt_token;
    CURL* curl;

public:
    CentralAPIClient(const std::string& url, const std::string& id,
                     const std::string& token)
        : api_url(url), collector_id(id), jwt_token(token) {
        curl = curl_easy_init();
    }

    bool PushMetrics(const nlohmann::json& metrics) {
        std::string url = api_url + "/api/v1/metrics/push";

        // Prepare headers with JWT token
        struct curl_slist* headers = nullptr;
        headers = curl_slist_append(headers, "Content-Type: application/json");
        headers = curl_slist_append(headers,
            ("Authorization: Bearer " + jwt_token).c_str());
        headers = curl_slist_append(headers,
            ("X-Collector-ID: " + collector_id).c_str());

        // Prepare payload
        std::string payload = metrics.dump();

        // Setup CURL
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, payload.c_str());
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 30L);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 1L);

        // Send request
        CURLcode res = curl_easy_perform(curl);

        curl_slist_free_all(headers);

        return (res == CURLE_OK);
    }

    bool SendHeartbeat() {
        std::string url = api_url + "/api/v1/collectors/" + collector_id + "/heartbeat";

        struct curl_slist* headers = nullptr;
        headers = curl_slist_append(headers, "Content-Type: application/json");
        headers = curl_slist_append(headers,
            ("Authorization: Bearer " + jwt_token).c_str());

        nlohmann::json heartbeat = {
            {"timestamp", getCurrentTimestamp()},
            {"status", "active"},
            {"metrics_count", getMetricsCount()}
        };

        std::string payload = heartbeat.dump();

        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_POSTFIELDS, payload.c_str());
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);

        CURLcode res = curl_easy_perform(curl);
        curl_slist_free_all(headers);

        return (res == CURLE_OK);
    }

    ~CentralAPIClient() {
        if (curl) {
            curl_easy_cleanup(curl);
        }
    }
};

// Main collection loop
void CollectorMain() {
    // Get environment variables
    std::string jwt_token = getenv("COLLECTOR_JWT_TOKEN");
    std::string collector_id = getenv("COLLECTOR_ID");
    std::string api_url = getenv("CENTRAL_API_URL");

    CentralAPIClient client(api_url, collector_id, jwt_token);

    while (true) {
        try {
            // Collect metrics from target database
            nlohmann::json metrics = CollectMetrics();

            // Push to central backend
            if (client.PushMetrics(metrics)) {
                LOG_INFO("Metrics pushed successfully");
            } else {
                LOG_ERROR("Failed to push metrics");
            }

            // Send heartbeat every 60 seconds
            if (shouldSendHeartbeat()) {
                client.SendHeartbeat();
            }

            // Sleep until next collection interval
            std::this_thread::sleep_for(
                std::chrono::seconds(COLLECTION_INTERVAL)
            );
        } catch (const std::exception& e) {
            LOG_ERROR("Collection error: " + std::string(e.what()));
        }
    }
}
```

---

## Frontend Architecture

### React Component Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── CollectorRegistration/
│   │   │   ├── index.tsx          # Main component
│   │   │   ├── RegistrationForm.tsx   # Form component
│   │   │   ├── BulkUpload.tsx     # Bulk import
│   │   │   ├── TestConnection.tsx # Connection tester
│   │   │   ├── styles.module.css  # Styles
│   │   │   └── types.ts           # Type definitions
│   │   │
│   │   ├── CollectorList/
│   │   │   ├── index.tsx          # List component
│   │   │   ├── CollectorCard.tsx  # Card for each collector
│   │   │   ├── CollectorFilters.tsx # Filter UI
│   │   │   └── styles.module.css
│   │   │
│   │   ├── CollectorDetail/
│   │   │   ├── index.tsx          # Detail page
│   │   │   ├── StatusPanel.tsx    # Status display
│   │   │   ├── MetricsPanel.tsx   # Metrics display
│   │   │   ├── SettingsPanel.tsx  # Edit settings
│   │   │   └── ActivityLog.tsx    # Audit log
│   │   │
│   │   ├── CollectorGroups/
│   │   │   ├── index.tsx          # Group management
│   │   │   └── GroupForm.tsx      # Create/edit group
│   │   │
│   │   └── Common/
│   │       ├── Header.tsx         # App header
│   │       ├── Sidebar.tsx        # Navigation
│   │       ├── Toast.tsx          # Notifications
│   │       ├── Modal.tsx          # Modal dialogs
│   │       ├── LoadingSpinner.tsx # Loading state
│   │       └── ErrorBoundary.tsx  # Error handling
│   │
│   ├── pages/
│   │   ├── Dashboard.tsx          # Home page
│   │   ├── Collectors.tsx         # List page
│   │   ├── CollectorDetail.tsx    # Detail page
│   │   ├── Register.tsx           # Registration page
│   │   ├── BulkImport.tsx         # Bulk import page
│   │   ├── Groups.tsx             # Group management page
│   │   └── Settings.tsx           # Admin settings
│   │
│   ├── services/
│   │   ├── api.ts                 # Axios instance
│   │   ├── collectors.ts          # Collector API calls
│   │   ├── auth.ts                # Authentication API
│   │   └── metrics.ts             # Metrics API
│   │
│   ├── hooks/
│   │   ├── useCollectors.ts       # Custom hook for collectors
│   │   ├── useCollectorDetail.ts  # Detail fetch hook
│   │   ├── useAuth.ts             # Auth hook
│   │   ├── usePagination.ts       # Pagination hook
│   │   └── useForm.ts             # Form handling
│   │
│   ├── store/
│   │   ├── store.ts               # Redux store setup
│   │   ├── actions/
│   │   │   ├── collectors.ts
│   │   │   ├── auth.ts
│   │   │   └── ui.ts
│   │   ├── reducers/
│   │   │   ├── collectors.ts
│   │   │   ├── auth.ts
│   │   │   └── ui.ts
│   │   └── selectors/
│   │       ├── collectors.ts
│   │       └── auth.ts
│   │
│   ├── utils/
│   │   ├── validation.ts          # Form validation
│   │   ├── formatting.ts          # Data formatting
│   │   ├── constants.ts           # App constants
│   │   └── helpers.ts             # Utility functions
│   │
│   ├── App.tsx                    # Root component
│   ├── index.tsx                  # Entry point
│   └── styles.css                 # Global styles
│
├── public/
│   └── index.html                 # HTML template
│
├── package.json
├── tsconfig.json
├── webpack.config.js (or vite.config.js)
├── .env.example
└── README.md
```

---

## Deployment Checklist

### Pre-Deployment
- [ ] Database schema migrations applied to RDS
- [ ] JWT secret configured in backend
- [ ] Encryption keys generated and secured
- [ ] Collector groups created
- [ ] React build optimized
- [ ] API documentation updated
- [ ] Security audit completed
- [ ] CORS configured correctly

### Deployment Steps
- [ ] Deploy backend API
- [ ] Deploy React frontend
- [ ] Verify API endpoints accessible
- [ ] Test JWT token generation
- [ ] Test password encryption
- [ ] Test collector registration
- [ ] Test metrics push
- [ ] Verify dashboards receive metrics

### Post-Deployment
- [ ] Monitor error logs
- [ ] Test collector registration via UI
- [ ] Verify metrics flow to Grafana
- [ ] Load test with multiple collectors
- [ ] Monitor RDS performance
- [ ] Verify audit logs are captured

---

## Success Criteria

✅ Single collector registration <2 minutes
✅ Bulk import 100+ collectors <5 minutes
✅ JWT tokens valid for 1 year
✅ Password encryption using AES-256-GCM
✅ Real-time status updates every 60 seconds
✅ Metrics arrive at backend <5 seconds after collection
✅ Grafana dashboards update within 10 seconds
✅ Mobile-responsive registration form
✅ All form fields validated
✅ Clear error messages for failures
✅ Audit log tracks all changes
✅ <100ms response time for API endpoints

---

**Status:** Architecture and design complete
**Next Phase:** Implementation (55-75 hours estimated)
**Priority:** High (Core feature for multi-collector management)

