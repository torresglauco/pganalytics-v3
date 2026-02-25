# pgAnalytics-v3 Collector Enhancement Plan - Advanced PostgreSQL Monitoring

**Date:** February 25, 2026
**Status:** Planning Phase - Ready for Implementation
**Scope:** Add critical PostgreSQL replication, health, and resource monitoring
**Estimated Effort:** 3-4 weeks (Phase 1: Replication Metrics)

---

## 📋 Executive Summary

A estrutura atual do coletor C++ possui **capacidades sólidas** para coletar métricas de query e sistema operacional. Este plano adiciona:

### ✅ Sim, é possível coletar:
1. **CPU e Memória** - Já existem no `SysstatCollector` (via `/proc/stat` e `/proc/meminfo`)
2. **Status do PostgreSQL** - Pode ser expandido com health checks avançados

### 🔴 Precisa adicionar:
1. **Replication Lag Metrics** (replication delay, replay, diff)
2. **Replication Slot Status** (disabled slots, WAL retention risk)
3. **VACUUM Prevention Risk** (wraparound threshold monitoring)
4. **GraphQL Integration** (query metrics via GraphQL API)
5. **AI-Driven Alerting** (anomaly detection no backend)

---

## 🏗️ ARQUITETURA ATUAL DO COLETOR

### Collectors Existentes

| Collector | Propósito | Métrica Count | Status |
|-----------|-----------|---------------|--------|
| **pg_stats** | Database/Table/Index stats | ~50 | ✅ Ativo |
| **pg_query_stats** | Query execution stats | 25 por query | ✅ Ativo |
| **sysstat** | CPU/Memory/Disk I/O | 10+ | ✅ Ativo |
| **disk_usage** | Filesystem usage | 4 per filesystem | ✅ Ativo |
| **pg_log** | Log parsing | N/A | ✅ Ativo |

### Capacidades Atuais

**CPU & Memória:**
```cpp
// Já implementado em sysstat_plugin.cpp
- CPU: user%, system%, idle%, iowait%, load_1m/5m/15m
- Memory: total_mb, free_mb, cached_mb, used_mb
- Disk I/O: read_ops, write_ops, read_sectors, write_sectors
- Filesystems: used_gb, free_gb, percent_used
```

**PostgreSQL Service Health:**
```cpp
// Métodos atuais:
1. Connection timeout: 5 segundos (postgres_plugin.cpp:62)
2. Query timeout: 30 segundos (query_stats_plugin.cpp:78)
3. Extension validation: pg_stat_statements check
4. Token expiration: JWT refresh com 60s buffer
5. Configuration pull: a cada 300s
6. Buffer health: monitora fullness
```

---

## 🎯 PHASE 1: REPLICATION METRICS (3-4 semanas)

### 1.1 Replication Lag Collector

**Novo arquivo:** `collector/src/replication_plugin.cpp`

```cpp
class ReplicationCollector : public Collector {
    // Coleta métricas de replicação via pg_stat_replication

    struct ReplicationSlot {
        std::string slot_name;
        std::string slot_type;              // "physical" ou "logical"
        bool active;
        std::string restart_lsn;            // LSN da primeira alteração não reproduzida
        std::string confirmed_flush_lsn;    // Logical only
        int64_t wal_retained_mb;            // WAL retido para replicação
        bool plugin_active;                 // Logical only
        int64_t backend_pid;                // PID do processo de replicação
    };

    struct ReplicationStatus {
        int64_t server_pid;                 // PID do servidor
        std::string usename;                // Username do cliente
        std::string application_name;       // Nome da aplicação (ex: "walreceiver")
        std::string client_hostname;        // Hostname do cliente
        int64_t client_port;                // Porta do cliente
        std::string state;                  // "streaming", "catchup", "backup"
        std::string sync_state;             // "sync" ou "async"
        std::string sync_priority;          // Priority para sync replication

        // LSN Information
        std::string write_lsn;              // LSN escrito no cliente
        std::string flush_lsn;              // LSN no disco do cliente
        std::string replay_lsn;             // LSN reprocessado no cliente

        // Lag Metrics
        int64_t write_lag_ms;               // Tempo para write_lsn (ms)
        int64_t flush_lag_ms;               // Tempo para flush_lsn (ms)
        int64_t replay_lag_ms;              // Tempo para replay_lsn (ms)

        // Replica Position vs Master
        std::string replay_vs_write_diff;   // LSN difference (bytes)
        std::string flush_vs_write_diff;    // LSN difference (bytes)
        int64_t behind_by_mb;               // MB atrás em termos de WAL
    };

    struct VacuumWrapAroundRisk {
        std::string database;
        int64_t relfrozenxid;               // Oldest unfrozen transaction ID
        int64_t current_xid;                // Current transaction ID
        int64_t xid_until_wraparound;       // XIDs remaining until wraparound
        int64_t percent_until_wraparound;   // Percentage (0-100%)
        bool at_risk;                       // true se < 10% remaining
        int64_t tables_needing_vacuum;      // Count of tables with high age
    };

    // Métodos
    json collectReplicationSlots();
    json collectReplicationStatus();
    json collectVacuumWrapAroundRisk();
    json collectWalSegmentStatus();
};
```

### 1.2 Queries SQL para Replication Metrics

```sql
-- Query 1: Replication Slot Status
SELECT
    slot_name,
    slot_type,
    active,
    restart_lsn::text,
    confirmed_flush_lsn::text,  -- NULL para physical
    (SELECT sum(size) FROM pg_ls_waldir()
     WHERE name >= split_part(restart_lsn::text, '/', 1) || '/' ||
                    lpad(split_part(restart_lsn::text, '/', 2), 8, '0'))::bigint / 1024 / 1024 as wal_retained_mb,
    plugin,
    decoded_xmin::text
FROM pg_replication_slots
WHERE slot_type IN ('physical', 'logical');

-- Query 2: Streaming Replication Status
SELECT
    p.pid as server_pid,
    u.usename,
    a.application_name,
    a.client_hostname,
    a.client_port,
    state,
    sync_state,
    sync_priority,
    write_lsn::text,
    flush_lsn::text,
    replay_lsn::text,
    EXTRACT(EPOCH FROM (now() - backend_start)) as uptime_seconds,
    (EXTRACT(EPOCH FROM (now() - write_lag))::int * 1000)::int as write_lag_ms,
    (EXTRACT(EPOCH FROM (now() - flush_lag))::int * 1000)::int as flush_lag_ms,
    (EXTRACT(EPOCH FROM (now() - replay_lag))::int * 1000)::int as replay_lag_ms
FROM pg_stat_replication r
JOIN pg_authid u ON r.usesysid = u.usesysid
JOIN pg_stat_activity a ON r.pid = a.pid;

-- Query 3: WAL Segment Status (PG13+)
SELECT
    count(*) as total_wal_segments,
    pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0') / (1024*1024) as current_wal_mb,
    sum(size) / (1024*1024) as wal_dir_size_mb
FROM pg_ls_waldir();

-- Query 4: Vacuum Wraparound Risk
SELECT
    datname,
    relfrozenxid,
    (SELECT max(nextxid) FROM pg_database WHERE datname = d.datname) as current_xid,
    (2147483647 - (SELECT max(nextxid) FROM pg_database WHERE datname = d.datname)) as xid_remaining,
    round(100.0 * (2147483647 - (SELECT max(nextxid) FROM pg_database WHERE datname = d.datname)) / 2147483647, 2) as percent_remaining,
    CASE
        WHEN (2147483647 - (SELECT max(nextxid) FROM pg_database WHERE datname = d.datname)) < 214748364 THEN true
        ELSE false
    END as at_risk
FROM pg_database d;

-- Query 5: Tables at Risk (age > autovacuum_freeze_max_age default 200M)
SELECT
    schemaname,
    tablename,
    age(relfrozenxid) as age_xids,
    (2147483647 - age(relfrozenxid)) as xids_until_wraparound
FROM pg_stat_user_tables
WHERE age(relfrozenxid) > 150000000  -- 75% of default threshold
ORDER BY age_xids DESC
LIMIT 20;
```

### 1.3 Data Structure JSON Output

```json
{
  "type": "replication",
  "timestamp": "2026-02-25T10:30:00Z",
  "database": "production_db",

  "replication_slots": [
    {
      "slot_name": "replica_1_slot",
      "slot_type": "physical",
      "active": true,
      "restart_lsn": "0/12345678",
      "confirmed_flush_lsn": null,
      "wal_retained_mb": 256,
      "plugin": null,
      "status": "healthy"
    },
    {
      "slot_name": "logical_replica_slot",
      "slot_type": "logical",
      "active": true,
      "restart_lsn": "0/87654321",
      "confirmed_flush_lsn": "0/87654300",
      "wal_retained_mb": 512,
      "plugin": "test_decoding",
      "decoded_xmin": "123456789",
      "status": "healthy"
    }
  ],

  "replication_status": [
    {
      "server_pid": 12345,
      "usename": "repuser",
      "application_name": "walreceiver",
      "client_hostname": "replica-01.internal",
      "client_port": 5432,
      "state": "streaming",
      "sync_state": "async",
      "sync_priority": 1,
      "uptime_seconds": 86400,
      "lsn_info": {
        "write_lsn": "0/12345600",
        "flush_lsn": "0/12345500",
        "replay_lsn": "0/12345400"
      },
      "lag_metrics": {
        "write_lag_ms": 45,
        "flush_lag_ms": 120,
        "replay_lag_ms": 250,
        "behind_by_mb": 5
      },
      "diff_info": {
        "replay_vs_write_mb": 2.5,
        "flush_vs_write_mb": 0.8
      }
    }
  ],

  "vacuum_wraparound_risk": [
    {
      "database": "production_db",
      "relfrozenxid": 1900000000,
      "current_xid": 1950000000,
      "xid_until_wraparound": 197483647,
      "percent_until_wraparound": 92,
      "at_risk": false,
      "tables_needing_vacuum": 0,
      "status": "healthy"
    }
  ],

  "wal_status": {
    "total_wal_segments": 256,
    "current_wal_mb": 8192,
    "wal_dir_size_mb": 16384,
    "estimated_wal_per_hour_mb": 512,
    "time_until_wal_full_hours": 32
  },

  "replication_issues": []
}
```

### 1.4 Alerting Rules (Backend GraphQL)

```graphql
# Estrutura de alertas que serão criados no backend

type ReplicationAlert {
  id: ID!
  severity: AlertSeverity!  # CRITICAL, WARNING, INFO
  title: String!
  description: String!
  database: String!
  collector_id: String!
  timestamp: DateTime!
  resolved: Boolean!

  detectedIssues: [DetectedIssue!]!
}

type DetectedIssue {
  type: IssueType!  # LAG_CRITICAL, SLOT_DISABLED, WRAPAROUND_RISK, WAL_PRESSURE
  metric: String!
  currentValue: String!
  threshold: String!
  recommendation: String!
}

# Queries e Mutations
extend type Query {
  replicationAlerts(
    database: String
    severity: AlertSeverity
    resolved: Boolean
    limit: Int = 50
  ): [ReplicationAlert!]!

  replicationHealth(database: String!): ReplicationHealthSummary!

  vacuumStatus(database: String!): [VacuumStatusDetail!]!
}

extend type Mutation {
  acknowledgeReplicationAlert(alertId: ID!): ReplicationAlert!

  snoozeAlert(alertId: ID!, minutes: Int!): ReplicationAlert!

  triggerManualVacuum(database: String!, table: String): Job!
}
```

---

## 🤖 PHASE 2: AI-DRIVEN ANOMALY DETECTION (2-3 semanas)

### 2.1 Backend ML Model (Python/FastAPI)

```python
# backend/ml/anomaly_detector.py

class ReplicationAnomalyDetector:
    """
    Detecta anomalias em métricas de replicação usando:
    - Isolation Forest (detecção de outliers)
    - LSTM (séries temporais)
    - Seasonal decomposition (trends sazonais)
    """

    def __init__(self):
        self.isolation_forest = IsolationForest(contamination=0.05)
        self.scaler = StandardScaler()
        self.lstm_model = None  # Treinado com histórico

    def detect_lag_anomalies(self, lag_history: List[float]) -> AnomalyResult:
        """
        Detecta saltos anormais em lag de replicação

        Cenários:
        1. Lag crescendo constantemente (problema contínuo)
        2. Spike súbito (problema agudo)
        3. Padrão cíclico anormal (queries pesadas em horários)
        """
        pass

    def detect_wraparound_risk(self, xid_data: pd.DataFrame) -> RiskAssessment:
        """
        Prediz quando wraparound pode ocorrer baseado em:
        - Velocidade de geração de XIDs
        - Histórico de vacuum
        - Tamanho do banco
        - Carga de transações
        """
        pass

    def detect_slot_issues(self, slot_metrics: Dict) -> List[SlotIssue]:
        """
        Identifica problemas com slots de replicação:
        1. Slot inativo por muito tempo
        2. Slot não conseguindo fazer progress
        3. WAL retention muito alto
        4. Decoder lag muito grande (logical slots)
        """
        pass

class ReplicationHealthScore:
    """
    Calcula score de saúde da replicação (0-100)
    """
    def __init__(self):
        self.weights = {
            "lag": 0.25,
            "slot_health": 0.20,
            "wraparound": 0.25,
            "wal_pressure": 0.15,
            "sync_state": 0.15
        }

    def calculate(self, metrics: ReplicationMetrics) -> HealthScore:
        """
        Retorna score com breakdown:
        {
            "overall": 85,
            "lag_score": 90,
            "slot_health_score": 80,
            "wraparound_score": 85,
            "wal_pressure_score": 80,
            "sync_state_score": 85,
            "risk_level": "LOW",  # LOW, MEDIUM, HIGH, CRITICAL
            "recommendations": ["Increase connection pool on replica-01"]
        }
        """
        pass
```

### 2.2 GraphQL Queries para AI

```graphql
extend type Query {
  # Anomalias detectadas
  anomalyDetection(
    database: String
    metric: String
    timeRange: TimeRange!
    confidenceLevel: Float = 0.95
  ): [AnomalyDetected!]!

  # Score de saúde com previsões
  replicationHealthPredictor(
    database: String!
    hoursAhead: Int = 24
  ): HealthPrediction!

  # Recomendações IA
  aiRecommendations(
    database: String!
    category: RecommendationCategory  # LAG, WRAPAROUND, SLOTS, WAL
  ): [AIRecommendation!]!

  # Análise de padrões
  replicationPatternAnalysis(
    database: String!
    timeRange: TimeRange!
  ): PatternAnalysis!
}

type AnomalyDetected {
  timestamp: DateTime!
  metric: String!
  expectedValue: Float!
  actualValue: Float!
  deviation: Float!  # Percentual de desvio
  severity: AnomalySeverity!
  explanation: String!
  possibleCauses: [String!]!
}

type HealthPrediction {
  database: String!
  currentScore: Int!
  predictedScore24h: Int!
  trend: Trend!  # IMPROVING, DEGRADING, STABLE
  criticalAlertsPredicted: Int!
  recommendations: [String!]!
}

type AIRecommendation {
  priority: Int!  # 1-10
  category: String!
  action: String!
  expectedImpact: String!
  estimatedTimeToImplement: String!
  riskLevel: String!
}
```

---

## 📊 PHASE 3: GRAPHQL INTEGRATION (1-2 semanas)

### 3.1 Query Metrics via GraphQL

```graphql
extend type Query {
  # Queries com maior lag
  slowestQueries(
    database: String
    limit: Int = 50
    orderBy: QueryOrderBy = MEAN_TIME_DESC
    timeRange: TimeRange
  ): [QueryMetric!]!

  # Queries causando replicação lag
  queriesAffectingReplication(
    database: String!
    replicaName: String
  ): [QueryWithReplicationImpact!]!

  # Análise de queries por hora
  queryMetricsByHour(
    database: String!
    query: String!
    timeRange: TimeRange!
  ): [HourlyQueryMetrics!]!

  # Queries com padrão anormal
  anomalousQueries(
    database: String!
    timeRange: TimeRange!
  ): [AnomalousQuery!]!
}

type QueryMetric {
  hash: String!
  text: String!
  database: String!

  executionStats: ExecutionStats!
  bufferStats: BufferStats!
  ioStats: IOStats!
  walStats: WALStats!

  trend: QueryTrend!
  anomalies: [QueryAnomaly!]!
}

type QueryTrend {
  direction: Direction!  # UP, DOWN, STABLE
  changePercent: Float!
  lastUpdated: DateTime!
}

type QueryAnomaly {
  type: String!  # EXECUTION_TIME, ROW_COUNT, BUFFER_RATIO, WAL_GENERATION
  severity: String!
  explanation: String!
}

# Mutation para triggers
extend type Mutation {
  explainQuery(
    database: String!
    hash: String!
    forceReplan: Boolean = false
  ): ExplainPlan!

  analyzeQueryPerformance(
    database: String!
    hash: String!
    samples: Int = 100
  ): PerformanceAnalysis!
}
```

### 3.2 Health Dashboard GraphQL

```graphql
type DashboardData {
  replicationHealth: ReplicationHealthDashboard!
  queryPerformance: QueryPerformanceDashboard!
  systemMetrics: SystemMetricsDashboard!
  alerts: [Alert!]!
  aiRecommendations: [AIRecommendation!]!
}

type ReplicationHealthDashboard {
  overallHealth: HealthScore!

  replicationStatus: [ReplicaStatus!]!

  lagMetrics: LagMetricsSummary!

  slotsStatus: SlotsHealthSummary!

  vacuumStatus: VacuumStatusSummary!

  walStatus: WALStatusSummary!

  criticalIssues: [Issue!]!
}

type QueryPerformanceDashboard {
  slowestQueries: [QueryMetric!]!

  highWALGenerationQueries: [QueryMetric!]!

  bufferCacheHitRatio: Float!

  averageExecutionTime: Float!

  anomalousQueries: [AnomalousQuery!]!
}

type SystemMetricsDashboard {
  cpuUsage: CPUMetrics!

  memoryUsage: MemoryMetrics!

  diskIO: DiskIOMetrics!

  diskUsage: DiskUsageMetrics!

  networkIO: NetworkIOMetrics!
}

# Super query para dashboard único
extend type Query {
  healthDashboard(database: String!): DashboardData!
}
```

---

## 🛠️ PHASE 4: COLLECTOR ENHANCEMENTS (1-2 semanas)

### 4.1 Atualizar `replication_plugin.cpp`

```cpp
// Adicionar métodos para coletar dados contínuos

class ReplicationCollector : public Collector {
public:
    json collect() override {
        json result = json::object();

        result["replication_slots"] = collectReplicationSlots();
        result["replication_status"] = collectReplicationStatus();
        result["vacuum_status"] = collectVacuumWrapAroundRisk();
        result["wal_status"] = collectWalSegmentStatus();
        result["active_subscriptions"] = collectLogicalSubscriptions();  // Para logical replication

        // Calcular métricas derivadas
        result["derived_metrics"] = calculateDerivedMetrics(result);

        // Adicionar timestamps
        result["collected_at"] = getCurrentTimestamp();
        result["is_primary"] = isPrimaryServer();

        return result;
    }

private:
    json calculateDerivedMetrics(const json& raw_metrics) {
        json derived = json::object();

        // LSN calculations
        derived["lsn_replication_diff"] = calculateLSNDiff(
            raw_metrics["replication_status"][0]["write_lsn"],
            raw_metrics["replication_status"][0]["replay_lsn"]
        );

        // Wraparound prediction
        derived["wraparound_eta_days"] = predictWraparoundDate(
            raw_metrics["vacuum_status"]
        );

        // WAL pressure estimate
        derived["wal_generation_rate_mb_per_hour"] = estimateWALRate();

        return derived;
    }

    int64_t calculateLSNDiff(const std::string& lsn1, const std::string& lsn2) {
        // Converter LSNs "0/12345678" para bytes
        uint32_t wal1_hi, wal1_lo, wal2_hi, wal2_lo;
        sscanf(lsn1.c_str(), "%X/%X", &wal1_hi, &wal1_lo);
        sscanf(lsn2.c_str(), "%X/%X", &wal2_hi, &wal2_lo);

        uint64_t pos1 = ((uint64_t)wal1_hi << 32) | wal1_lo;
        uint64_t pos2 = ((uint64_t)wal2_hi << 32) | wal2_lo;

        return (int64_t)(pos1 - pos2);
    }

    int64_t predictWraparoundDate(const json& vacuum_data) {
        // Usar histórico de XIDs para prever quando wraparound ocorrerá
        // Fórmula: (2^31 - current_xid) / (xid_generation_rate_per_second)
        pass;
    }
};
```

---

## 📈 IMPLEMENTATION TIMELINE

### Week 1-2: Replication Metrics
- [ ] Implementar `ReplicationCollector`
- [ ] Adicionar queries SQL para slots, status, WAL
- [ ] Adicionar data structures JSON
- [ ] Testes unitários
- [ ] Integração com `MetricsSerializer`

### Week 3: GraphQL & Backend
- [ ] Schema GraphQL para replication metrics
- [ ] Backend models para replication data
- [ ] API endpoints para queries
- [ ] Dashboards no frontend

### Week 4: AI & Alerting
- [ ] Treinar modelos ML
- [ ] Implementar anomaly detection
- [ ] Criar alerting rules
- [ ] Integrar com notification system

### Week 5: Testing & Deployment
- [ ] Teste de carga com collectors
- [ ] Validação em staging
- [ ] Documentação
- [ ] Deploy em produção

---

## 📋 IMPLEMENTATION CHECKLIST

### Collector Changes
- [ ] Nova classe `ReplicationCollector`
- [ ] SQL queries para 5 tipos de dados
- [ ] JSON serialization
- [ ] Error handling (standby servers, older PG versions)
- [ ] Configuration TOML updates
- [ ] Health checks

### Backend Changes
- [ ] TypeScript models
- [ ] GraphQL schema
- [ ] MongoDB collections
- [ ] API resolvers
- [ ] Alerting logic

### Frontend Changes
- [ ] Replication health dashboard
- [ ] Alert visualization
- [ ] Query performance charts
- [ ] AI recommendations widget

### Testing
- [ ] Unit tests (collector)
- [ ] Integration tests (e2e)
- [ ] Load tests (1000+ metrics/cycle)
- [ ] Regression tests

### Documentation
- [ ] Architecture diagram
- [ ] Metrics reference
- [ ] Troubleshooting guide
- [ ] API documentation

---

## 🎯 SUCCESS CRITERIA

✅ **Phase 1:** Coletar 100% das métricas de replicação sem performance impact (<5% CPU)
✅ **Phase 2:** Detectar 90%+ das anomalias reais com <5% false positives
✅ **Phase 3:** Visualizar todas as métricas em GraphQL com latency <200ms
✅ **Phase 4:** Suportar 10,000+ metricated queries/cycle sem buffer overflow

---

## 🔧 TECHNICAL DEBT & NOTES

1. **PostgreSQL Versions:** Support PG 12+ (WAL stats, LAG functions available in 10+)
2. **Standby Servers:** Replication collector must handle read-only mode gracefully
3. **Cluster Monitoring:** Consider multi-server replication topology (cascading replicas)
4. **Performance:** Consider sampling queries at scale (top 100 by lag impact)
5. **Storage:** Replication metrics history = ~2KB per update × 60/hour × 30 days = ~88MB per database

---

**Next Step:** Começar a implementação da Phase 1 (Replication Collector)

