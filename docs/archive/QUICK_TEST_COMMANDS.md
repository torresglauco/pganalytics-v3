# Phase 3.1 Authentication - Quick Test Commands

**Last Updated**: March 5, 2026
**Status**: All tests ready to execute

---

## One-Liner Test Execution

```bash
# Run all Phase 3.1 tests
cd /Users/glauco.torres/git/pganalytics-v3 && go test ./backend/internal/{auth,session} -v

# Run with benchmarks
cd /Users/glauco.torres/git/pganalytics-v3 && go test -bench=. ./backend/internal/{auth,session} -benchmem
```

---

## Test Packages

### All Tests
```bash
go test ./backend/internal/auth ./backend/internal/session -v
```

### Authentication Tests Only
```bash
go test ./backend/internal/auth -v
```

### Session Tests Only
```bash
go test ./backend/internal/session -v
```

---

## Specific Test Suites

### LDAP Tests
```bash
go test ./backend/internal/auth -run TestLDAP -v
go test ./backend/internal/auth -run TestNewLDAPConnector -v
go test ./backend/internal/auth -run TestResolveRole -v
go test ./backend/internal/auth -run TestLDAPConnectorFields -v
go test ./backend/internal/auth -run TestLDAPClose -v
```

### OAuth Tests
```bash
go test ./backend/internal/auth -run TestOAuth -v
go test ./backend/internal/auth -run TestNewOAuthConnector -v
go test ./backend/internal/auth -run TestGetAuthCodeURL -v
go test ./backend/internal/auth -run TestProviderConfiguration -v
```

### MFA Tests
```bash
go test ./backend/internal/auth -run TestMFA -v
go test ./backend/internal/auth -run TestNewMFAManager -v
go test ./backend/internal/auth -run TestGenerateTOTPSecret -v
go test ./backend/internal/auth -run TestVerifyTOTP -v
go test ./backend/internal/auth -run TestGenerateBackupCodes -v
go test ./backend/internal/auth -run TestGenerateSecureCode -v
go test ./backend/internal/auth -run TestGenerateRandomCode -v
go test ./backend/internal/auth -run TestHashCode -v
go test ./backend/internal/auth -run TestValidateTOTPSecret -v
go test ./backend/internal/auth -run TestMFATypeValues -v
```

### Session Tests
```bash
go test ./backend/internal/session -v
go test ./backend/internal/session -run TestSessionStructure -v
go test ./backend/internal/session -run TestGenerateSecureToken -v
go test ./backend/internal/session -run TestGenerateSessionID -v
go test ./backend/internal/session -run TestGenerateSecureRandomString -v
go test ./backend/internal/session -run TestParseInt -v
go test ./backend/internal/session -run TestParseInt64 -v
go test ./backend/internal/session -run TestSessionCreation -v
go test ./backend/internal/session -run TestSessionExpiry -v
go test ./backend/internal/session -run TestIPAddressParsing -v
```

---

## Benchmark Tests

### All Benchmarks
```bash
go test -bench=. ./backend/internal/auth ./backend/internal/session -benchmem
```

### LDAP Benchmarks
```bash
go test -bench=BenchmarkResolveRole ./backend/internal/auth -benchmem
```

### OAuth Benchmarks
```bash
go test -bench=BenchmarkGetAuthCodeURL ./backend/internal/auth -benchmem
```

### MFA Benchmarks
```bash
go test -bench=Benchmark ./backend/internal/auth -benchmem | grep -E "MFA|Code|Hash"
go test -bench=BenchmarkGenerateSecureCode ./backend/internal/auth -benchmem
go test -bench=BenchmarkGenerateRandomCode ./backend/internal/auth -benchmem
go test -bench=BenchmarkHashCode ./backend/internal/auth -benchmem
```

### Session Benchmarks
```bash
go test -bench=Benchmark ./backend/internal/session -benchmem
go test -bench=BenchmarkGenerateSecureToken ./backend/internal/session -benchmem
go test -bench=BenchmarkGenerateSessionID ./backend/internal/session -benchmem
go test -bench=BenchmarkSessionCreation ./backend/internal/session -benchmem
```

---

## Coverage Analysis

### Generate Coverage Report
```bash
go test -coverprofile=coverage.out ./backend/internal/{auth,session}
go tool cover -html=coverage.out
```

### View Coverage Summary
```bash
go test -cover ./backend/internal/{auth,session}
```

### Detailed Coverage by Package
```bash
go test -coverprofile=coverage.out ./backend/internal/auth
go tool cover -func=coverage.out | head -20

go test -coverprofile=coverage.out ./backend/internal/session
go tool cover -func=coverage.out | head -20
```

---

## Verbose & Detailed Output

### With Maximum Verbosity
```bash
go test -v -race ./backend/internal/{auth,session}
```

### With Race Detector (detects concurrent access issues)
```bash
go test -race ./backend/internal/{auth,session}
```

### With Timeout (prevent hanging tests)
```bash
go test -timeout 30s ./backend/internal/{auth,session} -v
```

### With Parallel Execution Control
```bash
go test -parallel 1 ./backend/internal/{auth,session} -v  # Sequential
go test -parallel 4 ./backend/internal/{auth,session} -v  # 4 parallel
```

---

## Long-Running Tests

### Benchmark with Custom Duration
```bash
go test -bench=. -benchtime=10s ./backend/internal/{auth,session} -benchmem
```

### Memory Profiling
```bash
go test -memprofile=mem.prof ./backend/internal/{auth,session}
go tool pprof mem.prof
```

### CPU Profiling
```bash
go test -cpuprofile=cpu.prof -bench=. ./backend/internal/{auth,session}
go tool pprof cpu.prof
```

---

## Quick Status Check

### All Tests Pass?
```bash
go test ./backend/internal/{auth,session} && echo "✅ All tests pass" || echo "❌ Tests failed"
```

### Count Tests
```bash
go test -list . ./backend/internal/{auth,session} | wc -l
```

### Quick Benchmark Summary
```bash
go test -bench=. -benchmem ./backend/internal/{auth,session} | grep -E "^Benchmark|ns/op"
```

---

## CI/CD Integration

### GitHub Actions / Jenkins
```bash
#!/bin/bash
set -e

echo "Running Phase 3.1 Authentication Tests..."
go test ./backend/internal/{auth,session} -v

echo "Running Benchmarks..."
go test -bench=. ./backend/internal/{auth,session} -benchmem

echo "Generating Coverage..."
go test -coverprofile=coverage.out ./backend/internal/{auth,session}

echo "✅ All Phase 3.1 tests passed"
```

### With JUnit Output (for CI dashboards)
```bash
go install github.com/jstemmer/go-junit-report/v2@latest
go test -v ./backend/internal/{auth,session} 2>&1 | \
  go-junit-report -set-exit-code > report.xml
```

---

## Expected Output

### Successful Test Run (2 seconds)
```
ok      backend/internal/auth       1.234s
ok      backend/internal/session    0.890s

PASS: 29 tests completed
```

### Successful Benchmark Run (5 seconds)
```
BenchmarkGenerateSecureToken        10000    125 µs/op   32 B/op   1 allocs/op
BenchmarkGenerateSessionID          20000     67 µs/op   16 B/op   1 allocs/op
BenchmarkSessionCreation             5000    235 µs/op  256 B/op   8 allocs/op
...
```

---

## Troubleshooting

### Test Hangs?
```bash
# Add timeout
go test -timeout 30s ./backend/internal/{auth,session}
```

### Database Error?
```bash
# Skip database-dependent tests
SKIP_DB_TESTS=1 go test ./backend/internal/{auth,session}
```

### LDAP Connection Timeout?
```bash
# Skip LDAP tests
SKIP_LDAP_TESTS=1 go test ./backend/internal/auth
```

### Race Condition Detected?
```bash
# Run sequentially instead of parallel
go test -parallel 1 ./backend/internal/{auth,session}
```

---

## Build Before Testing

```bash
# Build authentication modules
go build ./backend/internal/auth
go build ./backend/internal/session

# Test after building
go test ./backend/internal/{auth,session} -v
```

---

## Continuous Testing (Watch Mode)

Using `gotestsum` (install: `go install gotest.tools/gotestsum@latest`):

```bash
# Watch for file changes and re-run tests
gotestsum --watch ./backend/internal/{auth,session}
```

Using plain Go with file watcher:

```bash
# Simple loop (requires inotify on Linux or fswatch on macOS)
while true; do
  clear
  go test ./backend/internal/{auth,session} -v
  inotifywait -e modify -r ./backend/internal/{auth,session}
done
```

---

## Quick Reference Card

| Command | Purpose | Time |
|---------|---------|------|
| `go test ./backend/internal/{auth,session} -v` | Run all tests | 2s |
| `go test -bench=. ./backend/internal/{auth,session}` | Run benchmarks | 5s |
| `go test -cover ./backend/internal/{auth,session}` | Show coverage % | 2s |
| `go test -race ./backend/internal/{auth,session}` | Detect races | 3s |
| `go test ./backend/internal/auth -run TestLDAP -v` | LDAP tests only | 1s |
| `go test ./backend/internal/auth -run TestMFA -v` | MFA tests only | 1s |
| `go test ./backend/internal/session -v` | Session tests only | 1s |

---

## Full Test Suite Command

```bash
#!/bin/bash
# Complete Phase 3.1 test execution

cd /Users/glauco.torres/git/pganalytics-v3

echo "=== Phase 3.1 Authentication Tests ==="
echo ""
echo "1. Running Unit Tests..."
go test ./backend/internal/auth ./backend/internal/session -v

echo ""
echo "2. Running Benchmarks..."
go test -bench=. ./backend/internal/auth ./backend/internal/session -benchmem

echo ""
echo "3. Coverage Report..."
go test -cover ./backend/internal/auth ./backend/internal/session

echo ""
echo "✅ Phase 3.1 testing complete!"
```

Save this as `test-phase3.sh` and run:
```bash
bash test-phase3.sh
```

---

## Notes

- All tests use mocking for external dependencies
- No real LDAP/OAuth servers required for unit tests
- Database tests can be skipped with `SKIP_DB_TESTS=1`
- Tests are deterministic and repeatable
- Benchmarks show consistent results across runs

---

**Ready to test? Start with:**
```bash
cd /Users/glauco.torres/git/pganalytics-v3
go test ./backend/internal/{auth,session} -v
```
