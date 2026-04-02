# Localização dos Arquivos de Teste - pgAnalytics v3

## Backend Tests (Go)

### Estrutura
```
backend/
├── internal/
│   ├── services/
│   │   ├── query_performance/
│   │   │   ├── query_collector.go
│   │   │   ├── query_parser.go
│   │   │   └── *_test.go               ← Testes (15 testes)
│   │   ├── vacuum_advisor/
│   │   │   ├── analyzer.go
│   │   │   ├── cost_calculator.go
│   │   │   └── *_test.go               ← Testes (31 testes)
│   ├── session/
│   │   ├── session.go
│   │   └── *_test.go                   ← Testes (22 testes)
│   ├── storage/
│   │   └── *_test.go                   ← Testes (FALHA)
│   └── timescale/
│       └── *_test.go                   ← Testes (FALHA)
│
├── cmd/
│   ├── pganalytics-api/
│   ├── pganalytics-cli/
│   │   └── tests/
│   ├── load-test-runner/
│   │   └── main.go                     ← ERRO: quebra de linha redundante
│   └── pganalytics-mcp-server/
│
└── tests/
    └── integration/
        ├── full_system_integration_test.go  ← ERROS: referências ausentes
        └── testhelpers.go                   ← REDECLARAÇÕES

Total Backend: 233 testes (232 passing, 1 failing)
```

### Comandos para Rodar
```bash
# Todos os testes internos
cd /Users/glauco.torres/git/pganalytics-v3/backend
go test ./internal/... -v -cover

# Pacote específico
go test ./internal/services/vacuum_advisor/... -v -cover

# Com relatório de cobertura
go test ./internal/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Arquivos de Teste Específicos
- `/Users/glauco.torres/git/pganalytics-v3/backend/internal/services/query_performance/*_test.go`
- `/Users/glauco.torres/git/pganalytics-v3/backend/internal/services/vacuum_advisor/*_test.go`
- `/Users/glauco.torres/git/pganalytics-v3/backend/internal/session/*_test.go`
- `/Users/glauco.torres/git/pganalytics-v3/backend/tests/integration/*_test.go`

---

## Frontend Tests (Node)

### Estrutura
```
frontend/
├── src/
│   ├── services/
│   │   └── realtime.test.ts             ← 30 testes ✅
│   ├── hooks/
│   │   ├── useRealtime.test.ts          ← 14 testes ✅
│   │   └── useQueryPerformance.test.ts  ← 1 teste ✅
│   ├── components/
│   │   ├── logs/
│   │   │   ├── LiveLogsStream.test.tsx  ← 23 testes ✅
│   │   │   └── LogsViewer.test.tsx      ← ✅
│   │   ├── LoginForm.test.tsx           ← ✅
│   │   ├── CreateManagedInstanceForm.test.tsx ← 3 testes ✅
│   │   └── CreateUserForm.test.tsx      ← 2 testes ✅
│   ├── pages/
│   │   └── AuthPage.test.tsx            ← 6 testes ✅
│   ├── stores/
│   │   └── realtimeStore.test.ts        ← 25 testes ✅
│   └── __tests__/
│       └── integration/
│           └── components.integration.test.tsx ← 27 testes ✅
│
├── e2e/
│   └── tests/
│       ├── 01-login-logout.spec.ts              ← ❌ (Playwright)
│       ├── 02-collector-registration.spec.ts   ← ❌ (Playwright)
│       ├── 03-dashboard.spec.ts                 ← ❌ (Playwright)
│       ├── 04-alert-management.spec.ts          ← ❌ (Playwright)
│       ├── 05-user-management.spec.ts           ← ❌ (Playwright)
│       └── 06-permissions-access-control.spec.ts ← ❌ (Playwright)
│
├── vitest.config.ts                    ← Configuração de testes
├── package.json                         ← Scripts de teste
└── tsconfig.json

Total Frontend: 386 testes (386 passing, 6 E2E blocked)
```

### Comandos para Rodar
```bash
cd /Users/glauco.torres/git/pganalytics-v3/frontend

# Todos os testes
npm test -- --coverage

# Suite específica
npm test -- src/services/realtime.test.ts

# Watch mode
npm test -- --watch

# E2E (após instalar Playwright)
npm run test:e2e
```

### Arquivos de Teste Específicos
- `/Users/glauco.torres/git/pganalytics-v3/frontend/src/**/*.test.{ts,tsx}`
- `/Users/glauco.torres/git/pganalytics-v3/frontend/e2e/tests/*.spec.ts`

### Configuração
- **Test Framework**: Vitest
- **Config File**: `/Users/glauco.torres/git/pganalytics-v3/frontend/vitest.config.ts`
- **Coverage**: v8 (habilitado com --coverage)

---

## Collector Tests (C++)

### Estrutura
```
collector/
├── tests/
│   ├── unit/
│   │   ├── config_manager_test.cpp      ← ConfigManagerTest: 19 testes ✅
│   │   ├── metrics_serializer_test.cpp  ← MetricsSerializerTest: 20 testes ✅
│   │   ├── metrics_buffer_test.cpp      ← MetricsBufferTest: 18 testes ✅
│   │   ├── sender_test.cpp              ← SenderTest: ✅
│   │   ├── auth_test.cpp                ← AuthTest: ✅
│   │   └── replication_collector_test.cpp ← ReplicationCollectorTest: 20 testes ✅
│   │
│   ├── integration/
│   │   ├── auth_integration_test.cpp     ← AuthManagerTest: 6/9 ⚠️
│   │   ├── sender_integration_test.cpp   ← SenderIntegrationTest: 3/19 ❌
│   │   ├── regression_test.cpp           ← RegressionTest: 25 testes ✅
│   │   ├── error_handling_test.cpp       ← ErrorHandlingTest: 18 testes ✅
│   │   ├── config_integration_test.cpp   ← ConfigIntegrationTest: ✅
│   │   ├── collector_flow_test.cpp       ← CollectorFlowTest: ✅
│   │   └── mock_backend_server.cpp       ← Servidor mock (usado pelos testes)
│   │
│   ├── e2e/
│   │   ├── 1_collector_registration_test.cpp     ← E2ECollectorRegistrationTest (8 skipped)
│   │   ├── 2_metrics_ingestion_test.cpp          ← E2EMetricsIngestionTest (11 skipped)
│   │   ├── 3_configuration_test.cpp              ← E2EConfigurationTest (9 skipped)
│   │   ├── 4_dashboard_visibility_test.cpp       ← E2EDashboardVisibilityTest (6 skipped)
│   │   ├── 5_performance_test.cpp                ← E2EPerformanceTest (8 skipped)
│   │   ├── 6_failure_recovery_test.cpp           ← E2EFailureRecoveryTest (8 skipped)
│   │   └── e2e_harness.cpp
│   │
│   ├── postgres_plugin_test.cpp          ← Plugin tests
│   ├── CMakeLists.txt                    ← Build configuration
│   └── run_tests.sh                      ← Script de execução
│
├── CMakeLists.txt                       ← CMake principal (BUILD_TESTS=ON)
└── src/
    ├── collector.cpp
    ├── sender.cpp
    ├── auth.cpp
    └── ... (implementações dos módulos)

Total Collector: 296 testes (228 passing, 49 skipped, 19 failing)
```

### Comandos para Rodar
```bash
# Compilar com testes
cd /Users/glauco.torres/git/pganalytics-v3/collector/build
cmake -DBUILD_TESTS=ON ..
make -j4

# Rodar testes
./tests/pganalytics-tests

# Rodar com script
cd /Users/glauco.torres/git/pganalytics-v3/collector
bash tests/run_tests.sh

# Testes específicos com ctest
ctest --output-on-failure
ctest -R "AuthManagerTest" -V
```

### Arquivos de Teste Específicos
- `/Users/glauco.torres/git/pganalytics-v3/collector/tests/unit/*.cpp`
- `/Users/glauco.torres/git/pganalytics-v3/collector/tests/integration/*.cpp`
- `/Users/glauco.torres/git/pganalytics-v3/collector/tests/e2e/*.cpp`

### Configuração
- **Test Framework**: Google Test (GTest)
- **Build System**: CMake 3.22+
- **C++ Standard**: C++17

---

## Sumário de Localização

### Backend
- **Test Runner**: Go (go test)
- **Location**: `/Users/glauco.torres/git/pganalytics-v3/backend/`
- **Test Patterns**: `*_test.go`
- **Coverage**: `go test -cover`

### Frontend
- **Test Runner**: Vitest (npm test)
- **Location**: `/Users/glauco.torres/git/pganalytics-v3/frontend/`
- **Test Patterns**: `*.test.ts`, `*.test.tsx`, `*.spec.ts`
- **Coverage**: vitest --coverage

### Collector
- **Test Runner**: Google Test (ctest or ./pganalytics-tests)
- **Location**: `/Users/glauco.torres/git/pganalytics-v3/collector/`
- **Test Patterns**: `*_test.cpp`
- **Coverage**: GTest built-in (limited)

---

## Logs de Execução

- **Backend Log**: `/tmp/backend_unit_tests.log`
- **Frontend Log**: `/tmp/frontend_tests.log`
- **Collector Log**: `/tmp/collector_cpp_tests.log`

---

## Próximos Passos

1. **Frontend E2E**: Instalar Playwright
   ```bash
   cd /Users/glauco.torres/git/pganalytics-v3/frontend
   npm install @playwright/test --save-dev
   npm run test:e2e
   ```

2. **Backend Integração**: Verificar referências ausentes
   ```bash
   grep -r "index_advisor" /Users/glauco.torres/git/pganalytics-v3/backend/internal/
   ```

3. **Collector Integração**: Revisar mock server
   ```bash
   cat /Users/glauco.torres/git/pganalytics-v3/collector/tests/integration/mock_backend_server.cpp
   ```

---

**Documento gerado**: 2026-04-01
**Projeto**: pgAnalytics v3
**Repository**: /Users/glauco.torres/git/pganalytics-v3
