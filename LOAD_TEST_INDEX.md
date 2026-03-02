# pgAnalytics v3 - Full Regression Test Suite
## Complete Index & Documentation

---

## 📋 Quick Navigation

### Getting Started
1. **First Time?** → Read: `QUICK_START_LOAD_TEST.md` (4 min read)
2. **Need Details?** → Read: `REGRESSION_TEST_README.md` (20 min read)
3. **Visual Guide?** → Read: `TEST_FLOW_GUIDE.txt` (10 min read)

### Execution Scripts
- **Phase 1**: `./cleanup-and-start-load-test.sh` - Infrastructure setup
- **Phase 2**: `./test-setup-managed-instances.sh` - Managed instances registration
- **Phase 3**: `./verify-regression-tests.sh` - Comprehensive validation

### Infrastructure
- **Composition**: `docker-compose-load-test.yml` - 84 services definition

---

## 📁 Files Overview

### Documentation Files

#### 1. `QUICK_START_LOAD_TEST.md` (4.3 KB)
**Best for**: Quick reference, fast setup
- TL;DR commands
- What's being tested summary
- Success indicators
- Quick troubleshooting
- Access points

#### 2. `REGRESSION_TEST_README.md` (15 KB)
**Best for**: Comprehensive understanding, detailed execution
- Complete overview and context
- Network architecture diagram
- Step-by-step execution guide
- Success criteria checklist
- Monitoring commands
- Extended testing procedures
- Full troubleshooting guide
- Performance expectations
- Docker resource requirements

#### 3. `TEST_FLOW_GUIDE.txt` (8.7 KB)
**Best for**: Visual understanding of test flow, expected outputs
- Timeline breakdown
- What happens in each phase
- Expected output examples
- Log patterns to look for
- Warning signs
- Success indicators
- Next steps

#### 4. `LOAD_TEST_IMPLEMENTATION_SUMMARY.md` (5.9 KB)
**Best for**: Implementation status, file inventory
- Completion status
- Deliverables overview
- File summary table
- Validation checklist
- Execution timeline
- Pre-execution checklist

---

### Execution Scripts

#### 1. `cleanup-and-start-load-test.sh` (5.8 KB, executable)
**Purpose**: Phase 1 - Infrastructure Setup
**Duration**: 10-15 minutes

**What it does**:
- Stops existing containers (docker-compose.yml & load test)
- Removes all pganalytics volumes
- Cleans Docker images
- Starts fresh infrastructure with docker-compose-load-test.yml
- Waits for core services to be healthy
- Displays access URLs

**Usage**:
```bash
./cleanup-and-start-load-test.sh
```

**Expected Output**:
```
✓ Infrastructure started
✓ PostgreSQL OK
✓ TimescaleDB OK
✓ Backend OK
✓ Frontend OK
```

---

#### 2. `test-setup-managed-instances.sh` (7.3 KB, executable)
**Purpose**: Phase 2 - Managed Instances Registration
**Duration**: 1-2 minutes

**What it does**:
- Authenticates with admin credentials
- Creates 20 managed instance entries via API
- Validates each registration
- Generates setup report

**Usage**:
```bash
./test-setup-managed-instances.sh
```

**Expected Output**:
```
✓ Authenticated successfully
Registering managed instance 1/20... OK (ID: 1)
Registering managed instance 2/20... OK (ID: 2)
...
✓ Report generated: regression-test-setup-report.txt
```

---

#### 3. `verify-regression-tests.sh` (11 KB, executable)
**Purpose**: Phase 3 - Comprehensive Validation
**Duration**: 2-3 minutes

**What it validates**:
1. Authentication
2. Collector count (40)
3. Managed instances (20)
4. Collector status
5. Metrics collection
6. ID persistence
7. Registration secrets
8. Frontend accessibility

**Usage**:
```bash
./verify-regression-tests.sh
```

**Expected Output**:
```
Phase 1: Authentication
  ✓ Authenticated successfully
Phase 2: Collector Registration Validation
  ✓ Collector count equals 40
  ✓ No duplicate collector UUIDs
...
=== ALL TESTS PASSED ===
```

---

### Infrastructure File

#### `docker-compose-load-test.yml` (58 KB)
**Purpose**: Complete infrastructure definition for 40+40 setup

**Services**:
- 1 PostgreSQL metadata DB (172.20.0.10:5432)
- 1 TimescaleDB metrics DB (172.20.0.11:5433)
- 1 Backend API (172.20.0.20:8080)
- 1 Frontend UI (172.20.0.60:4000)
- 40 Target PostgreSQL instances (172.20.1.101-140)
- 40 Collectors (172.20.1.201-240)
- 40 unique collector data volumes

**Features**:
- Fixed IP addresses for reproducibility
- Auto-registration enabled for all collectors
- Health checks configured
- Proper dependency ordering
- Persistent storage for collector IDs

---

## 🚀 Quick Start

### Minimal Commands

```bash
# Phase 1: Setup infrastructure (10-15 min)
./cleanup-and-start-load-test.sh

# Phase 2: Register managed instances (1-2 min)
./test-setup-managed-instances.sh

# Phase 3: Run verification (2-3 min)
./verify-regression-tests.sh
```

**Total time: ~30-45 minutes**

---

## ✅ What Gets Tested

| Component | Count | Test |
|-----------|-------|------|
| Collectors | 40 | Auto-registration, ID persistence, no duplicates |
| Managed Instances | 20 | API registration, connection testing |
| Target Databases | 40 | Connectivity, metrics collection |
| Core Services | 4 | Health, accessibility, API functionality |
| Metrics | 40+ per collector | Collection, storage, monotonic growth |

---

## 📊 Success Criteria

All of the following must be true:

- ✅ 40 collectors registered (0 duplicates)
- ✅ Each collector has unique UUID
- ✅ 20 managed instances created
- ✅ All collectors in "registered" status
- ✅ All collectors sending heartbeats
- ✅ Metrics being collected from each collector
- ✅ Collector IDs persisted to volumes
- ✅ Registration secret properly tracked
- ✅ Frontend accessible and responsive
- ✅ No errors in logs

---

## 🌐 Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| Backend API | http://localhost:8080 | — |
| Frontend UI | http://localhost:4000 | — |
| API Documentation | http://localhost:8080/swagger | — |
| PostgreSQL Metadata | localhost:5432 | postgres:pganalytics |
| TimescaleDB Metrics | localhost:5433 | postgres:pganalytics |
| API Authentication | — | admin:admin |

---

## 📝 Generated Reports

After running the test scripts, these files are created:

| File | Created By | Contents |
|------|-----------|----------|
| `regression-test-setup-report.txt` | Phase 2 script | Managed instance registration results |
| `regression-test-report.txt` | Phase 3 script | Comprehensive test validation results |

---

## 🔍 Monitoring During Test

```bash
# Watch all logs
docker-compose -f docker-compose-load-test.yml logs -f

# Watch backend
docker-compose -f docker-compose-load-test.yml logs -f backend

# Watch specific collector
docker-compose -f docker-compose-load-test.yml logs -f collector-001

# Check running containers
docker ps | grep pganalytics | wc -l
# Expected: 84
```

---

## 🐛 Troubleshooting

### Containers not starting?
```bash
docker-compose -f docker-compose-load-test.yml logs backend
```

### Collectors not registering?
```bash
docker-compose -f docker-compose-load-test.yml logs collector-001 | tail -20
```

### Metrics not appearing?
```bash
docker-compose -f docker-compose-load-test.yml logs backend | grep -i metric
```

### Need to restart everything?
```bash
docker-compose -f docker-compose-load-test.yml down -v
./cleanup-and-start-load-test.sh
```

---

## 📈 Extended Testing (Optional)

After verification passes, you can:

1. **Monitor for 1+ hour** - Verify sustained metrics collection
2. **Test collector restart** - Verify ID persistence
3. **Test service failures** - Verify resilience
4. **Frontend testing** - Verify UI functionality

See `REGRESSION_TEST_README.md` for detailed extended testing procedures.

---

## 📚 Documentation Map

```
Getting Started:
  ├── QUICK_START_LOAD_TEST.md ...................... 4 min read
  ├── TEST_FLOW_GUIDE.txt ........................... 10 min read
  └── LOAD_TEST_IMPLEMENTATION_SUMMARY.md ........... 5 min read

Detailed Reference:
  └── REGRESSION_TEST_README.md ..................... 20 min read
      ├── Network architecture
      ├── Detailed step-by-step guide
      ├── Success criteria checklist
      ├── Monitoring commands
      ├── Extended testing procedures
      ├── Troubleshooting guide
      └── Performance expectations

Execution:
  ├── cleanup-and-start-load-test.sh ............... Phase 1
  ├── test-setup-managed-instances.sh ............. Phase 2
  └── verify-regression-tests.sh ................... Phase 3

Infrastructure:
  └── docker-compose-load-test.yml ................. 84 services
```

---

## ⏱️ Timeline

| Phase | Duration | Task |
|-------|----------|------|
| Setup | 10-15 min | Start infrastructure, wait for services |
| Collect | 1-2 min | Collectors auto-register |
| Register | 1-2 min | Create managed instances |
| Collect | 2-3 min | Collectors send metrics |
| Verify | 2-3 min | Run validation tests |
| **Total** | **~30-45 min** | Complete regression test |

---

## 🎯 Next Steps

1. **Quick overview**: Read `QUICK_START_LOAD_TEST.md`
2. **Visual flow**: Read `TEST_FLOW_GUIDE.txt`
3. **Execute Phase 1**: `./cleanup-and-start-load-test.sh`
4. **Execute Phase 2**: `./test-setup-managed-instances.sh`
5. **Execute Phase 3**: `./verify-regression-tests.sh`
6. **Review reports**: Check generated report files
7. **Extended testing**: Optional - see `REGRESSION_TEST_README.md`

---

## 📞 Support

For detailed information, refer to:
- **Quick questions?** → `QUICK_START_LOAD_TEST.md`
- **Specific issue?** → Search `REGRESSION_TEST_README.md` or `TEST_FLOW_GUIDE.txt`
- **Implementation questions?** → `LOAD_TEST_IMPLEMENTATION_SUMMARY.md`
- **Log analysis?** → Run phase script and check output/logs

---

## ✨ Summary

This regression test suite comprehensively validates pgAnalytics v3 with:
- **84 Docker containers** (4 core + 40 targets + 40 collectors)
- **40 collectors** registering automatically with no duplicates
- **20 managed instances** registered via API
- **Persistent storage** for collector IDs
- **Metrics collection** from each collector
- **Complete validation** suite with 8+ test categories

**All features from recent fixes are validated**:
- ✅ Collector ID persistence
- ✅ Auto-registration only on first startup
- ✅ No duplicate registrations
- ✅ Registration secret tracking
- ✅ Token refresh functionality
- ✅ Metrics collection and storage

---

**Ready to begin? Start with:** `./cleanup-and-start-load-test.sh`
