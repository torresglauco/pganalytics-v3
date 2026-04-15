# pgAnalytics v3 - Documentação: Issues de Referência Rápida

**Data:** 14 de Abril de 2026 | **Versão:** v3.1.0 | **Status:** Production Ready

---

## Tabela de Prioridades

| ID | Severidade | Categoria | Gap | Arquivo(s) | Horas | Status |
|----|-----------|-----------|-----|-----------|-------|--------|
| 1 | CRÍTICO | Database | Schema documentation | `/backend/migrations/*.sql` | 15-20 | NEXT |
| 2 | CRÍTICO | ML Models | Algorithm documentation | `/backend/internal/ml/`, `/backend/services/ml-service/` | 20-25 | NEXT |
| 3 | CRÍTICO | API | Error codes matrix | `/backend/pkg/errors/` | 10-15 | NEXT |
| 4 | ALTO | Operations | Health checks & observability | `/docs/` | 12-15 | Week 2 |
| 5 | ALTO | Features | Feature-specific guides | `/backend/internal/services/` | 15-20 | Week 2 |
| 6 | ALTO | Integration | MCP documentation | `/backend/cmd/pganalytics-mcp-server/` | 10-12 | Week 2 |
| 7 | ALTO | Operations | Backup & disaster recovery | `/docs/OPERATIONS_HA_DR.md` | 15-18 | Week 2 |
| 8 | MÉDIO | Frontend | TypeScript JSDoc | `/frontend/src/services/` | 12-15 | Week 3 |
| 9 | MÉDIO | Collector | C++ implementation | `/collector/include/` | 15-18 | Week 3 |
| 10 | MÉDIO | Deployment | Upgrade procedures | `/docs/guides/` | 10-12 | Week 3 |
| 11 | MÉDIO | API | Rate limiting details | `/docs/api/` | 5-8 | Week 3 |
| 12 | BAIXO | Development | Code examples | `/docs/EXAMPLES/` | 10-15 | Week 4 |

---

## Crítico - Fazer AGORA (P0)

### 1. Database Schema Documentation
**Problema:** Migrations SQL sem comentários, sem ER diagram
**Arquivos:**
- `/backend/migrations/000_complete_schema.sql` (2500+ linhas)
- `/backend/migrations/001_triggers.sql`
- 7 outros migration files

**Solução - Criar `/docs/10-DATABASE/SCHEMA.md`:**
- [ ] Documentar cada tabela principal
- [ ] Criar ER diagram
- [ ] Documentar índices críticos
- [ ] Data retention policies
- [ ] Espaço de disco esperado

---

### 2. ML Models Documentation
**Problema:** Sem algoritmos, sem accuracy metrics, sem parâmetros
**Arquivos:**
- `/backend/internal/ml/models/anomaly_detector.go`
- `/backend/services/ml-service/anomaly_model.py`
- `/backend/services/ml-service/models.py`

**Solução - Criar `/docs/30-ML-MODELS/ANOMALY_DETECTION.md`:**
- [ ] Algoritmo (Isolation Forest / LSTM / outro)
- [ ] Features de entrada
- [ ] Métricas de performance (precision, recall, F1)
- [ ] Parâmetros ajustáveis (threshold, window_size)
- [ ] Guia de tuning
- [ ] Dataset de treino

---

### 3. API Error Codes Documentation
**Problema:** Sem matriz de status codes, sem recovery instructions
**Arquivo:** `/backend/pkg/errors/errors.go` (20+ funções)

**Solução - Criar `/docs/20-API/ERROR_CODES.md`:**
- [ ] Matriz de HTTP status codes
- [ ] Quando cada erro ocorre
- [ ] Como recuperar de cada erro
- [ ] Exemplos de responses
- [ ] Retry strategy

---

## Alto - Próximas 2 semanas (P1)

### 4. Health Checks & Observability
**Arquivo a criar:** `/docs/50-OPERATIONS/OBSERVABILITY.md`
- [ ] Health check endpoints
- [ ] Prometheus metrics list
- [ ] SLOs (Service Level Objectives)
- [ ] Alertas recomendados
- [ ] Performance baselines

---

### 5. Feature-Specific Documentation
**Arquivos:** `/backend/internal/services/`
- query_performance/
- index_advisor/
- vacuum_advisor/

Criar para cada: O que faz, como usar, exemplos, interpretação

---

### 6. MCP Integration
**Arquivo a criar:** `/docs/70-INTEGRATIONS/MCP.md`
- [ ] JSON-RPC 2.0 protocol
- [ ] Available tools documentation
- [ ] Tool parameters and responses
- [ ] Error codes
- [ ] Tool call examples

---

### 7. Backup & Disaster Recovery
**Arquivo a criar:** `/docs/40-DEPLOYMENT/BACKUP_RESTORE.md`
- [ ] Backup strategy
- [ ] RTO/RPO definitions
- [ ] Backup schedule
- [ ] Restore procedures
- [ ] Emergency recovery

---

## Médio - Próximo Sprint (P2)

### 8. Frontend TypeScript JSDoc
**Localização:** `/frontend/src/services/`
- [ ] Adicionar JSDoc em funções públicas
- [ ] Documentar parâmetros e return types
- [ ] Exemplos de uso

---

### 9. C++ Collector Documentation
**Localização:** `/collector/include/`
- [ ] Documentar classe MetricsCollector
- [ ] Metrics coletadas por PostgreSQL version
- [ ] Buffering strategy
- [ ] Retry logic
- [ ] SSL/TLS troubleshooting

---

### 10. Upgrade Procedures
**Arquivo a criar:** `/docs/40-DEPLOYMENT/UPGRADE_PROCEDURES.md`
- [ ] Pre-upgrade checklist
- [ ] Breaking changes
- [ ] Step-by-step instructions
- [ ] Verification
- [ ] Rollback procedures

---

### 11. Rate Limiting Documentation
**Arquivo a criar/expandir:** `/docs/20-API/RATE_LIMITING.md`
- [ ] Rate limits per endpoint
- [ ] Algorithm description
- [ ] HTTP response headers
- [ ] Retry strategy

---

## Documentação Existente (BOM)

✅ `/CONTRIBUTING.md` - 669 linhas, excelente
✅ `/README.md` - 454 linhas, bom overview
✅ `/docs/ARCHITECTURE.md` - System design bem documentado
✅ `/docs/DEPLOYMENT.md` - Setup bem estruturado
✅ `/docs/guides/PRODUCTION_MONITORING_GUIDE.md` - Operations guide

---

## Estimativa de Esforço

```
P0 (CRÍTICO):   20-30 horas  → 1 pessoa, 1 semana
P1 (ALTO):      40-50 horas  → 1-2 pessoas, 2 semanas
P2 (MÉDIO):     50-65 horas  → 1-2 pessoas, 1-2 sprints
P3 (BAIXO):     20-30 horas  → Backlog

TOTAL:          130-175 horas (3-4 sprints, 1-2 pessoas)
```

---

## Próximos Passos

1. Revisar `/DOCUMENTATION_GAP_ANALYSIS.md` para detalhes completos
2. Criar issues no GitHub para cada item
3. Reorganizar `/docs/` conforme proposto
4. Adicionar CI/CD checks para documentação

---

**Criado em:** 2026-04-14
**Relatório Completo:** `/DOCUMENTATION_GAP_ANALYSIS.md`
