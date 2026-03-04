# Plano de Ações Imediatas - pgAnalytics v3
**Data**: Março 4, 2026
**Horizonte**: Próximas 4 semanas
**Objetivo**: Preparar para v3.3 e melhorar qualidade

---

## 🚨 AÇÕES CRÍTICAS (Esta Semana)

### 1️⃣ IMPLEMENTAR SECURITY TESTING ⏰ 20-24 HORAS

**Por que?**
- Projeto está em produção (v3.2.0)
- Zero testes de segurança implementados
- Risco: SQL injection, XSS, auth bypass
- **Impacto**: CRÍTICO

**O que fazer**:

#### A. OWASP Top 10 Scanning (8h)
```bash
# Instalação
go get -u github.com/securego/gosec/v2/cmd/gosec
go install github.com/aquasecurity/trivy/cmd/trivy@latest

# Executar
cd backend
gosec ./...
trivy scan ./

# Revisar findings
# Prioridade: CRITICAL → HIGH → MEDIUM
```

**Checklist**:
- [ ] Instalar gosec
- [ ] Executar contra backend
- [ ] Revisar findings críticos
- [ ] Abrir issues para cada vulnerability

#### B. Testes de Injection (8h)
```go
// backend/tests/security/sql_injection_test.go
package security

import "testing"

func TestSQLInjectionProtection(t *testing.T) {
    // Test prepared statements
    // Test query parameterization
    // Test ORM protection
}

func TestAuthBypass(t *testing.T) {
    // Test JWT validation
    // Test token expiry
    // Test refresh token flow
}
```

**Checklist**:
- [ ] Criar test file
- [ ] Testar SQL injection vectors
- [ ] Testar XSS vectors
- [ ] Documentar proteções existentes

#### C. Dependency Vulnerability Check (4h)
```bash
# Backend
cd backend
go list -json -m all | nancy sleuth

# Frontend
cd frontend
npm audit
# ou yarn audit
```

**Checklist**:
- [ ] Executar nancy/audit
- [ ] Revisar vulnerabilities
- [ ] Atualizar packages críticos
- [ ] Documentar accepted risks

#### D. Security Test CI/CD Integration (4h)
```yaml
# .github/workflows/security.yml
name: Security Testing
on: [push, pull_request]
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run gosec
        run: gosec ./backend/...
      - name: Run trivy
        run: trivy scan ./
      - name: npm audit
        run: cd frontend && npm audit
```

**Checklist**:
- [ ] Criar workflow file
- [ ] Testar localmente
- [ ] Commit e push
- [ ] Revisar resultados

---

### 2️⃣ CRIAR UPGRADE GUIDE v3.2 → v3.3 ⏰ 6-8 HORAS

**Por que?**
- Usuários em v3.2 production
- v3.3 terá breaking changes potenciais
- Sem guia = usuarios bloqueados

**Arquivo**: `UPGRADE_v3.2_TO_v3.3.md`

**Estrutura**:
```markdown
# Upgrade Guide: v3.2.0 → v3.3.0

## Pre-Upgrade Checklist
- [ ] Backup completo do banco
- [ ] Teste em staging
- [ ] Planejar downtime
- [ ] Notificar users

## Breaking Changes
- Authentication: Agora suporta LDAP/SAML
  - JWT tokens ainda funcionam
  - Migração é opcional

- Database schema
  - 2 novas tabelas de audit (v3.3)
  - 1 coluna nova em collectors
  - Migrações automáticas rodam ao iniciar

- API endpoints
  - Novo endpoint: POST /api/v1/auth/ldap
  - Deprecado: POST /api/v1/collectors (use new format)

## Step-by-Step Upgrade
1. Backup
2. Pull código v3.3
3. Run migrations: `docker-compose exec api migrate-up`
4. Test endpoints
5. Restart collectors
6. Validate monitoring

## Rollback Procedure
1. Restore backup
2. Pull código v3.2
3. Restart services

## Testing Checklist
- [ ] API health: GET /health
- [ ] Login funciona
- [ ] Collectors registram
- [ ] Métricas fluem
- [ ] Dashboards renderizam

## Troubleshooting
- Issue: Collectors fail to register
  Solution: Check JWT secret in config

- Issue: Audit logs not showing
  Solution: Run migrations manually
```

**Checklist**:
- [ ] Listar breaking changes
- [ ] Escrever upgrade steps
- [ ] Escrever rollback procedure
- [ ] Incluir test checklist
- [ ] Adicionar troubleshooting

---

## 📋 AÇÕES IMPORTANTES (Próximas 2 Semanas)

### 3️⃣ CONTRIBUTING GUIDE ⏰ 6-8 HORAS

**Arquivo**: `CONTRIBUTING.md`

**Conteúdo Mínimo**:
```markdown
# Contributing to pgAnalytics

## Code Standards
- Go: Follow `go fmt`, use golangci-lint
- TypeScript: Use ESLint config provided
- C++: Follow Google C++ style guide

## Testing Requirements
- Backend: >80% code coverage
- Frontend: >70% coverage
- Add tests for new features
- All tests must pass

## Git Workflow
1. Fork repository
2. Create feature branch: git checkout -b feature/my-feature
3. Commit with meaningful messages
4. Push and open PR
5. Address review comments
6. Rebase and squash before merge

## PR Template
- [ ] Tests added
- [ ] Docs updated
- [ ] No breaking changes (or documented)
- [ ] Passes CI/CD

## Development Setup
```bash
make deps
make build
make test
```

## Performance Considerations
- Collector changes: Test with 50+ collectors
- API changes: Verify latency <500ms
- Database: Explain plan for queries
- UI: Check Core Web Vitals

## Security Checklist
- [ ] No hardcoded secrets
- [ ] Input validation
- [ ] SQL injection protected
- [ ] XSS protected
```

**Checklist**:
- [ ] Criar CONTRIBUTING.md
- [ ] Adicionar code standards
- [ ] Adicionar PR template
- [ ] Link no README

---

### 4️⃣ HA/DR OPERATIONS DOCUMENTATION ⏰ 8-10 HORAS

**Arquivo**: `docs/OPERATIONS_HA_DR.md`

**Seções**:

#### A. High Availability Setup (será implementado em v3.3)
```markdown
## HA Architecture (v3.3.0+)

### Prerequisites
- 2+ API servers
- Load balancer (HAProxy, Nginx, cloud LB)
- Shared PostgreSQL (RDS, managed)
- Redis for session state

### Configuration
- All API servers connect to same DB
- Session state in Redis
- HAProxy distributes traffic
- Collectors connect to LB virtual IP

### Validation
```bash
# Test failover
1. Start all servers
2. Kill server 1
3. Verify requests routed to server 2
4. Verify no session loss
```

#### B. Backup Procedures (será implementado em v3.3)
```markdown
## Backup Strategy

### Automated Backups
- PostgreSQL: Daily @ 2 AM UTC
- Location: S3 bucket with encryption
- Retention: 30 days

### Recovery Time Objective (RTO)
- Target: <1 hour
- Procedure: Restore from backup + replay WAL logs
```

#### C. Disaster Recovery Plan
```markdown
## DR Runbook

### Scenario 1: Database Corruption
1. Detect via monitoring alert
2. Restore from most recent good backup
3. Verify data integrity
4. Notify users of brief downtime

### Scenario 2: Complete Data Center Loss
1. Failover to backup region
2. Restore databases from geo-redundant backup
3. Point collectors to new backend
4. Verify all systems online

### Testing
- Monthly DR drill
- Simulate full datacenter failure
- Verify RTO met
- Document any issues
```

**Checklist**:
- [ ] Escrever HA architecture (teórico para v3.3)
- [ ] Documentar backup procedures
- [ ] Escrever disaster scenarios
- [ ] Criar runbooks

---

### 5️⃣ AUMENTAR COBERTURA E2E ⏰ 16-20 HORAS

**Objetivo**: De 40% para 80% de cobertura E2E

**Implementação com Playwright**:

```bash
# Instalar
cd frontend
npm install -D @playwright/test

# Exemplo de teste E2E
# frontend/e2e/collector-registration.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Collector Registration Flow', () => {
  test('should register a new collector', async ({ page }) => {
    // 1. Navigate to registration
    await page.goto('http://localhost:3000/collectors/register');

    // 2. Fill form
    await page.fill('[name="hostname"]', 'db-prod-01');
    await page.fill('[name="port"]', '5432');

    // 3. Test connection
    await page.click('button:has-text("Test Connection")');
    await expect(page.locator('text=Connection successful')).toBeVisible();

    // 4. Register
    await page.click('button:has-text("Register")');

    // 5. Verify success
    await expect(page.locator('text=Collector registered')).toBeVisible();
  });
});
```

**Cenários a Adicionar**:
1. ✅ Login/Logout flow
2. ✅ Collector registration
3. ✅ Collector management (edit, delete, pause)
4. ✅ Dashboard visualization
5. ✅ Alert creation/management
6. ✅ User management (create, edit, delete users)
7. ✅ Permission testing (unauthorized access)

**Checklist**:
- [ ] Instalar Playwright
- [ ] Criar exemplo teste
- [ ] Adicionar 7 cenários
- [ ] Integrar no CI/CD
- [ ] Documentar como rodar

---

## 🗓️ CRONOGRAMA DE 4 SEMANAS

### Semana 1 (Mar 4-8)
```
Monday-Wednesday: Security Testing
├─ OWASP scanning (8h)
├─ Injection tests (8h)
└─ Dependency checks (4h)

Thursday: Upgrade Guide
├─ Breaking changes analysis (2h)
├─ Write upgrade steps (3h)
└─ Rollback procedure (1h)

Friday: Contributing Guide
└─ Initial draft (3h)
```

**Saídas**:
- Security test findings report
- Upgrade guide v3.2→v3.3
- Contributing guide draft

---

### Semana 2 (Mar 11-15)
```
Monday-Wednesday: E2E Testing
├─ Playwright setup (4h)
├─ Scenario 1-3 (8h)
└─ CI/CD integration (4h)

Thursday-Friday: HA/DR Docs
├─ Architecture docs (4h)
├─ Backup procedures (2h)
└─ DR runbooks (2h)
```

**Saídas**:
- E2E test suite 50% completo
- HA/DR operations guide

---

### Semana 3 (Mar 18-22)
```
Monday-Wednesday: E2E Complete
├─ Scenarios 4-7 (10h)
└─ Documentation (2h)

Thursday-Friday: Contributing Guide Complete
├─ Finalize standards (3h)
├─ Add PR template (1h)
└─ Link everywhere (1h)
```

**Saídas**:
- E2E test suite 100% completo
- Contributing guide finalizado

---

### Semana 4 (Mar 25-29)
```
Monday: Integration & Validation
├─ Run all new tests (2h)
├─ Fix any issues (4h)
└─ Performance check (1h)

Tuesday-Wednesday: Documentation Cleanup
├─ Review all docs (2h)
├─ Add cross-references (2h)
└─ Update README links (1h)

Thursday-Friday: Prepare for v3.3
├─ Create v3.3 branch (1h)
├─ Verify test passes (2h)
└─ Write release notes (2h)
```

**Saídas**:
- Todos testes passando
- Documentação completa e linked
- v3.3 pronto para iniciar desenvolvimento

---

## 📊 MÉTRICAS DE SUCESSO

### Semana 1
- [ ] 0 críticas issues encontradas (ou mitigação documentada)
- [ ] Upgrade guide completo
- [ ] Contributing guide 50% pronto

### Semana 2
- [ ] E2E coverage: 50% → 70%
- [ ] HA/DR docs 100% pronto
- [ ] Contributing guide finalizado

### Semana 3
- [ ] E2E coverage: 70% → 90%
- [ ] Todos novos tests passando
- [ ] Zero novos security findings

### Semana 4
- [ ] E2E coverage: 90% → 95%+
- [ ] Documentação 100% linked
- [ ] v3.3 pronto para iniciar

---

## 👥 ALOCAÇÃO DE RECURSOS

### Recomendado para 4 Semanas

```
Desenvolvedor Backend (0.5 FTE)
├─ Semana 1: Security testing + Upgrade guide
├─ Semana 2: HA/DR review + Contributing guide
└─ Semana 3-4: Code review para E2E tests

Desenvolvedor Frontend (0.5 FTE)
├─ Semana 2-3: Playwright setup + E2E tests
└─ Semana 4: E2E maintenance + docs

QA Engineer (1.0 FTE)
├─ Semana 1: Security test analysis
├─ Semana 2-3: E2E test execution + validation
└─ Semana 4: Test consolidation + reporting

DevOps (0.25 FTE)
├─ Semana 1: CI/CD security scanning setup
├─ Semana 2: HA/DR procedures documentation
└─ Semana 3-4: CI/CD for E2E tests
```

**Total**: 2.25 FTE × 4 weeks = 36 person-days

---

## 🎯 DEFINIÇÃO DE CONCLUÍDO

### Para cada item:

#### Security Testing ✅
- [ ] Gosec rodando com 0 CRITICAL findings
- [ ] Todas HIGH findings mitigadas ou aceitas
- [ ] Injection tests implementados
- [ ] CI/CD pipeline configurado

#### Upgrade Guide ✅
- [ ] Breaking changes documentados
- [ ] Upgrade steps claro e testado
- [ ] Rollback procedure documentado
- [ ] Incluído no release v3.3

#### Contributing Guide ✅
- [ ] Code standards definidos
- [ ] PR template pronto
- [ ] Dev setup instruções
- [ ] Linked no README

#### E2E Tests ✅
- [ ] 7 principais cenários cobertos
- [ ] >95% de sucesso em execução
- [ ] Integrado no CI/CD
- [ ] Documentação de manutenção

#### HA/DR Docs ✅
- [ ] Architecture doc completo
- [ ] Backup procedures claro
- [ ] DR runbooks por cenário
- [ ] Procedures testáveis

---

## 🚀 DEPOIS: INICIAR v3.3

Uma vez concluídas estas 4 semanas:

```
v3.3 Development (4 weeks)
├─ Week 1: Kubernetes Native
├─ Week 2: HA Load Balancing
├─ Week 3: Enterprise Auth
└─ Week 4: Encryption, Audit, Backup

v3.4 Development (4 weeks)
├─ Week 1-2: Thread Pool
├─ Week 2-3: Distributed Collection
└─ Week 4: Caching

v3.5 Development (4 weeks)
├─ Week 1-2: Anomaly Detection
├─ Week 2-3: Intelligent Alerting
└─ Week 4: Workload Analysis
```

---

## 📞 SUPORTE

**Perguntas sobre este plano?**
- Ref: PROJECT_ANALYSIS_REPORT_MARCH_2026.md
- Contato: Project Lead

**Próxima revisão**: Março 29, 2026 (antes de iniciar v3.3)

---

**Plano Preparado**: Março 4, 2026
**Status**: Pronto para Implementação
**Responsável**: Engineering Team
