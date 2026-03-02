# Quick Start - pgAnalytics Load Test (40+40)

## TL;DR - Run These Commands

```bash
# 1. Clean everything and start fresh (takes ~15 min)
./cleanup-and-start-load-test.sh

# 2. Wait 1-2 minutes for collectors to register

# 3. Create managed instances (takes ~2 min)
./test-setup-managed-instances.sh

# 4. Wait 2-3 minutes for metrics to be collected

# 5. Verify everything is working (takes ~3 min)
./verify-regression-tests.sh
```

**Total time: ~30-40 minutes**

---

## What Gets Tested

| Component | Count | Status |
|-----------|-------|--------|
| PostgreSQL Metadata DB | 1 | Core |
| TimescaleDB (Metrics) | 1 | Core |
| Backend API | 1 | Core |
| Frontend UI | 1 | Core |
| **Target PostgreSQL** | **40** | Monitored |
| **Collectors** | **40** | Monitoring |
| **Managed Instances** | **20** | API-registered |

---

## Success Looks Like This

```
=== ALL TESTS PASSED ===
✓ 40 collectors registered (0 duplicates)
✓ 20 managed instances created
✓ All collectors sending heartbeats
✓ Collectors collecting metrics
✓ Frontend accessible
```

---

## Monitoring During Test

Open three terminals:

```bash
# Terminal 1: Watch backend logs
docker-compose -f docker-compose-load-test.yml logs -f backend

# Terminal 2: Watch a specific collector
docker-compose -f docker-compose-load-test.yml logs -f collector-001

# Terminal 3: Test API endpoints
watch 'curl -s http://localhost:8080/api/v1/collectors \
  -H "Authorization: Bearer YOUR_TOKEN" | jq ". | length"'
```

---

## Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| API | http://localhost:8080 | — |
| Frontend | http://localhost:4000 | — |
| API Docs | http://localhost:8080/swagger/index.html | — |
| PostgreSQL | localhost:5432 | postgres:pganalytics |
| TimescaleDB | localhost:5433 | postgres:pganalytics |

---

## Quick Checks

```bash
# Check how many containers are running
docker ps | grep pganalytics | wc -l
# Expected: 84 (4 core + 40 targets + 40 collectors)

# Check if backend is healthy
curl http://localhost:8080/api/v1/health

# Check if frontend is accessible
curl http://localhost:4000

# Get list of collectors (requires auth)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/collectors | jq '. | length'
# Expected: 40
```

---

## Troubleshooting

**Containers not starting?**
```bash
docker-compose -f docker-compose-load-test.yml logs backend
```

**Collectors not registering?**
```bash
docker-compose -f docker-compose-load-test.yml logs collector-001 | tail -20
```

**Metrics not appearing?**
```bash
docker-compose -f docker-compose-load-test.yml logs backend | grep -i metric
```

**Need to restart?**
```bash
docker-compose -f docker-compose-load-test.yml restart
```

**Need to clean everything?**
```bash
docker-compose -f docker-compose-load-test.yml down -v
```

---

## Log Patterns (What To Look For)

| Event | Log Pattern | Appears |
|-------|-------------|---------|
| Collector registers | "Auto-registering collector" | Once per collector |
| Collector running | "Collector already registered" | Every 60 sec |
| Metrics collected | "Pushing X metrics" | Every 60 sec |
| Error | "error", "failed", "exception" | Should be minimal |

---

## Files Created

```
docker-compose-load-test.yml              40+40 infrastructure
cleanup-and-start-load-test.sh            Phase 1: Setup
test-setup-managed-instances.sh           Phase 2: Register MIs
verify-regression-tests.sh                Phase 3: Validate
regression-test-setup-report.txt          Phase 2 results
regression-test-report.txt                Phase 3 results
```

---

## Expected Metrics

After 5 minutes, each collector should have:
- ✓ Unique UUID
- ✓ Status: "registered"
- ✓ Last heartbeat: recent
- ✓ Metrics count: 50+

---

## Stop Everything

```bash
# Keep data
docker-compose -f docker-compose-load-test.yml stop

# Full cleanup
docker-compose -f docker-compose-load-test.yml down -v

# Switch back to standard compose
docker-compose up -d
```

---

## Next Steps

After success:
1. Read full docs: `REGRESSION_TEST_README.md`
2. Run extended tests: Leave running for 1+ hour
3. Test collector restart: See extended testing section
4. Review metrics growth in TimescaleDB
5. Test managed instance connection failover

---

See `REGRESSION_TEST_README.md` for detailed documentation.
