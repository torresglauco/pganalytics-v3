# pgAnalytics v3 - Action Items para Melhorar Testes e Validação

**Last Updated:** 14 de abril de 2026
**Referência:** [TEST_AND_VALIDATION_ANALYSIS.md](TEST_AND_VALIDATION_ANALYSIS.md)

---

## 🔴 PRIORITY 0 - HOJE (< 1 hora cada)

### Action 0.1: Fix Playwright Installation
**Status:** ❌ BLOCKER
**Effort:** 5 minutos

```bash
# 1. Verificar package.json
cd /Users/glauco.torres/git/pganalytics-v3/frontend
grep "@playwright/test" package.json
# ✅ Deve estar em devDependencies: "@playwright/test": "^1.59.1"

# 2. Instalar
npm install @playwright/test --save-dev

# 3. Verificar instalação
npm run test:e2e -- --list
# Deve listar os testes sem erro

# 4. Rodar testes
npm run test:e2e
```

**Verification:**
```bash
npm run test:e2e -- --reporter=list | head -20
# Deve mostrar testes sendo executados, não "Failed to resolve import"
```

**Owner:** DevOps/Frontend Lead
**Dependency:** None

---

### Action 0.2: Remove Silent Error Catching in E2E Tests
**Status:** ❌ TEST QUALITY ISSUE
**Effort:** 15 minutos

**File:** `/Users/glauco.torres/git/pganalytics-v3/frontend/e2e/tests/05-user-management.spec.ts`

**Current (Bad):**
```typescript
test('should validate email format', async ({ page }) => {
  // ...
  try {
    await expect(error.or(form)).toBeVisible({ timeout: 3000 });
  } catch {
    console.log('Email validation verified');  // ❌ PROBLEM
  }
});
```

**Fixed (Good):**
```typescript
test('should validate email format', async ({ page }) => {
  // ...
  // ✅ Let the test fail naturally
  await expect(error.or(form)).toBeVisible({ timeout: 3000 });
});
```

**Changes Needed:**
- [ ] Line 61-64: Remove try/catch
- [ ] Line 130-131: Remove try/catch
- [ ] Line 157-159: Remove try/catch
- [ ] Line 241-247: Remove try/catch
- [ ] Line 257-261: Remove try/catch

**Verification:**
```bash
npm run test:e2e -- 05-user-management.spec.ts
# Testes devem falhar claramente se há problema, não ser silenciados
```

**Owner:** QA/Frontend
**Dependency:** Action 0.1

---

### Action 0.3: Fix Backend Compilation Errors
**Status:** ❌ BUILD FAILURE
**Effort:** 20 minutos

#### Error 0.3.1: load-test-runner redundant newline
**File:** `/Users/glauco.torres/git/pganalytics-v3/backend/cmd/load-test-runner/main.go:27`

```go
// ❌ CURRENT
fmt.Println("Starting load test...")

// ✅ FIX
fmt.Print("Starting load test...\n")
// OR
fmt.Println("Starting load test")  // sem extra newline
```

**Verification:**
```bash
cd /Users/glauco.torres/git/pganalytics-v3/backend
go build ./cmd/load-test-runner
# Deve compilar sem erro
```

#### Error 0.3.2: Undefined index_advisor reference
**File:** `/Users/glauco.torres/git/pganalytics-v3/backend/tests/integration/full_system_integration_test.go:140`

```go
// ❌ CURRENT (linha 140)
adviser := index_advisor.NewIndexAdvisor(...)

// ✅ VERIFICAR
// 1. Se index_advisor package existe
find /Users/glauco.torres/git/pganalytics-v3/backend -type f -name "*.go" | xargs grep "package index_advisor"

// 2. Se package está importado
grep "import.*index_advisor" /Users/glauco.torres/git/pganalytics-v3/backend/tests/integration/full_system_integration_test.go

// 3. Se função existe
grep "func NewIndexAdvisor" /Users/glauco.torres/git/pganalytics-v3/backend/internal/services/index_advisor/*.go
```

**Fix Options:**
1. Add missing import
2. Check if package/function is renamed
3. Create NewIndexAdvisor if missing

#### Error 0.3.3: Duplicate mock definitions
**File:** `/Users/glauco.torres/git/pganalytics-v3/backend/tests/integration/testhelpers.go`

```go
// ❌ PROBLEM: MockExplainOutput defined multiple times
var MockExplainOutput = ...
var MockExplainOutput = ...  // Redeclared!

// ✅ FIX: Keep only one definition or move to separate package
```

**Verification:**
```bash
cd /Users/glauco.torres/git/pganalytics-v3/backend
go test ./tests/integration -v
# Deve compilar sem erro "redeclared"
```

**Owner:** Backend Lead
**Dependency:** None

---

### Action 0.4: Fix E2E Test User Creation
**Status:** ⚠️ TEST CORRECTNESS
**Effort:** 10 minutos

**File:** `/Users/glauco.torres/git/pganalytics-v3/frontend/e2e/tests/05-user-management.spec.ts`

**Issue:** Test creates user but doesn't verify in API

**Current (Line 67-92):**
```typescript
test('should create user successfully', async ({ page }) => {
  // Creates user via UI
  await usersPage.saveUser();
  await usersPage.expectSuccessMessage();
  await usersPage.expectUserInList(testEmail);
  // ✅ Good, but missing API verification
});
```

**Improved:**
```typescript
test('should create user successfully', async ({ page, request }) => {
  const testEmail = `test-user-${Date.now()}@example.com`;

  // 1. Create via UI
  await usersPage.clickCreateUser();
  await usersPage.fillUserForm({
    email: testEmail,
    password: 'SecurePassword123!',
    name: 'Test User',
    role: 'viewer',
  });
  await usersPage.saveUser();

  // 2. Verify UI
  await usersPage.expectSuccessMessage();
  await usersPage.expectUserInList(testEmail);

  // 3. Verify API (NEW)
  const response = await request.get('/api/v1/users', {
    headers: { 'Authorization': `Bearer ${page.context().authToken}` }
  });
  const data = await response.json();
  const createdUser = data.data.find(u => u.email === testEmail);

  expect(createdUser).toBeDefined();
  expect(createdUser.role).toBe('viewer');
});
```

**Owner:** QA
**Dependency:** Action 0.1, 0.2

---

## 🟡 PRIORITY 1 - THIS WEEK (< 4 hours total)

### Action 1.1: Implement Zod Validation Schemas
**Status:** 🟡 PARTIAL IMPLEMENTATION
**Effort:** 2 hours

**Files to Create:**

#### 1. `/frontend/src/schemas/auth.ts`
```typescript
import { z } from 'zod';

// Existing in package.json: "zod": "^3.22.4"

export const LoginSchema = z.object({
  username: z.string()
    .min(3, 'Username must be at least 3 characters')
    .max(50, 'Username must be less than 50 characters'),
  password: z.string()
    .min(8, 'Password must be at least 8 characters'),
});

export const SignupSchema = LoginSchema.extend({
  email: z.string()
    .email('Invalid email address'),
  full_name: z.string()
    .min(1, 'Full name required')
    .max(100),
  password: z.string()
    .min(12, 'Password must be at least 12 characters')
    .regex(/[A-Z]/, 'Password must contain uppercase letter')
    .regex(/[0-9]/, 'Password must contain number'),
  password_confirm: z.string(),
}).refine((data) => data.password === data.password_confirm, {
  message: 'Passwords do not match',
  path: ['password_confirm'],
});

export type LoginInput = z.infer<typeof LoginSchema>;
export type SignupInput = z.infer<typeof SignupSchema>;
```

#### 2. `/frontend/src/schemas/user.ts`
```typescript
import { z } from 'zod';

export const CreateUserSchema = z.object({
  email: z.string()
    .email('Invalid email address'),
  password: z.string()
    .min(12, 'Password must be at least 12 characters')
    .regex(/[A-Z]/, 'Must contain uppercase')
    .regex(/[0-9]/, 'Must contain number'),
  name: z.string()
    .min(1, 'Name required')
    .max(100),
  role: z.enum(['admin', 'viewer'], {
    errorMap: () => ({ message: 'Invalid role' })
  }),
});

export const EditUserSchema = CreateUserSchema.partial().merge(
  z.object({
    id: z.number().positive(),
  })
);

export type CreateUserInput = z.infer<typeof CreateUserSchema>;
export type EditUserInput = z.infer<typeof EditUserSchema>;
```

#### 3. `/frontend/src/schemas/alerts.ts`
```typescript
import { z } from 'zod';

export const AlertRuleSchema = z.object({
  name: z.string()
    .min(1, 'Alert name required')
    .max(100),
  description: z.string().optional(),
  metric: z.string()
    .min(1, 'Metric required'),
  condition: z.enum(['>', '<', '>=', '<=', '==', '!='], {
    errorMap: () => ({ message: 'Invalid condition' })
  }),
  threshold: z.number()
    .min(0, 'Threshold must be positive'),
  duration: z.number()
    .min(1, 'Duration must be at least 1')
    .max(86400, 'Duration cannot exceed 1 day'),
  enabled: z.boolean().default(true),
});

export type AlertRuleInput = z.infer<typeof AlertRuleSchema>;
```

#### 4. Update LoginForm.tsx
**File:** `/frontend/src/components/LoginForm.tsx`

```typescript
import { useState } from 'react';
import { LoginSchema, type LoginInput } from '../schemas/auth';

export function LoginForm() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});

    // Validate with Zod
    const result = LoginSchema.safeParse({ username, password });

    if (!result.success) {
      // Convert Zod errors to field errors
      const newErrors: Record<string, string> = {};
      result.error.issues.forEach(issue => {
        const path = issue.path[0] as string;
        newErrors[path] = issue.message;
      });
      setErrors(newErrors);
      return;
    }

    // Proceed with API call
    setIsLoading(true);
    try {
      const response = await apiClient.login(
        result.data.username,
        result.data.password
      );
      // ... handle success
    } catch (error) {
      setErrors({ submit: error.message });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        placeholder="Username"
        aria-invalid={!!errors.username}
      />
      {errors.username && (
        <span className="error">{errors.username}</span>
      )}

      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
        aria-invalid={!!errors.password}
      />
      {errors.password && (
        <span className="error">{errors.password}</span>
      )}

      {errors.submit && (
        <div className="error">{errors.submit}</div>
      )}

      <button type="submit" disabled={isLoading}>
        {isLoading ? 'Logging in...' : 'Login'}
      </button>
    </form>
  );
}
```

#### 5. Add Schema Tests
**File:** `/frontend/src/schemas/auth.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { LoginSchema, SignupSchema } from './auth';

describe('Auth Schemas', () => {
  describe('LoginSchema', () => {
    it('should accept valid credentials', () => {
      const result = LoginSchema.safeParse({
        username: 'admin',
        password: 'SecurePass123'
      });
      expect(result.success).toBe(true);
    });

    it('should reject username < 3 chars', () => {
      const result = LoginSchema.safeParse({
        username: 'ab',
        password: 'SecurePass123'
      });
      expect(result.success).toBe(false);
      expect(result.error?.issues[0].path).toContain('username');
    });

    it('should reject password < 8 chars', () => {
      const result = LoginSchema.safeParse({
        username: 'admin',
        password: 'short'
      });
      expect(result.success).toBe(false);
      expect(result.error?.issues[0].path).toContain('password');
    });
  });

  describe('SignupSchema', () => {
    it('should accept valid signup data', () => {
      const result = SignupSchema.safeParse({
        username: 'newuser',
        email: 'user@example.com',
        full_name: 'New User',
        password: 'SecurePass123!',
        password_confirm: 'SecurePass123!'
      });
      expect(result.success).toBe(true);
    });

    it('should reject mismatched passwords', () => {
      const result = SignupSchema.safeParse({
        username: 'newuser',
        email: 'user@example.com',
        full_name: 'New User',
        password: 'SecurePass123!',
        password_confirm: 'DifferentPass123!'
      });
      expect(result.success).toBe(false);
      expect(result.error?.issues[0].code).toBe('custom');
    });

    it('should reject weak password (no uppercase)', () => {
      const result = SignupSchema.safeParse({
        username: 'newuser',
        email: 'user@example.com',
        full_name: 'New User',
        password: 'securepass123',
        password_confirm: 'securepass123'
      });
      expect(result.success).toBe(false);
    });

    it('should reject invalid email', () => {
      const result = SignupSchema.safeParse({
        username: 'newuser',
        email: 'not-an-email',
        full_name: 'New User',
        password: 'SecurePass123!',
        password_confirm: 'SecurePass123!'
      });
      expect(result.success).toBe(false);
      expect(result.error?.issues[0].path).toContain('email');
    });
  });
});
```

**Verification:**
```bash
cd frontend
npm test -- schemas/auth.test.ts
# Todos os testes devem passar
```

**Owner:** Frontend Lead
**Dependency:** Action 0.1, 0.2

---

### Action 1.2: Run E2E Tests Successfully
**Status:** ⚠️ BLOCKED
**Effort:** 30 minutos

```bash
# 1. Garante Playwright instalado
cd /Users/glauco.torres/git/pganalytics-v3/frontend
npm install @playwright/test --save-dev

# 2. Inicia backend (assumindo Docker Compose)
# Em outro terminal:
docker compose up -d api postgres timescaledb

# 3. Aguarda backend estar pronto
sleep 5
curl http://localhost:8080/api/v1/health

# 4. Roda E2E tests
npm run test:e2e -- --reporter=html

# 5. Visualiza relatório
# Abre: playwright-report/index.html
```

**Expected Results:**
```
✅ 01-login-logout.spec.ts         (2-3 tests)
✅ 02-collector-registration.spec.ts
✅ 05-user-management.spec.ts      (agora sem silent failures)
⚠️ Pode haver outras falhas por setup (ignorar por enquanto)
```

**Owner:** DevOps/QA
**Dependency:** Action 0.1, 0.2

---

### Action 1.3: Fix Collector Integration Tests
**Status:** 🔴 16/19 FAILING
**Effort:** 90 minutos

**Files Affected:**
- `/collector/tests/integration/mock_backend_server.cpp`
- `/collector/tests/integration/sender_integration_test.cpp`
- `/collector/tests/integration/auth_integration_test.cpp`

**Issue 1: Mock Server Token Refresh**

```cpp
// ❌ CURRENT - mock_backend_server.cpp
class MockBackendServer {
  // Token validation sem refresh logic

  bool ValidateToken(const std::string& token) {
    // Simple validation, não implementa refresh
    return token != "";
  }
};

// ✅ FIX
class MockBackendServer {
  std::string refresh_token_secret_;

  RefreshTokenResponse RefreshToken(const std::string& old_token) {
    // 1. Validate old token hasn't expired
    if (!IsTokenValid(old_token)) {
      throw std::runtime_error("Token expired");
    }

    // 2. Generate new token
    std::string new_token = GenerateJWT();

    // 3. Return with new expiration
    return {new_token, 3600};  // 1 hour expiration
  }
};
```

**Issue 2: Metrics Payload Validation**

```cpp
// ❌ CURRENT
TEST_F(SenderIntegrationTest, SendMetricsSuccess) {
  // Sends metrics but doesn't validate payload format
  sender.SendMetrics(metrics);
}

// ✅ FIX
TEST_F(SenderIntegrationTest, SendMetricsSuccess) {
  // 1. Prepare test metrics
  std::vector<Metric> metrics = {
    Metric{.timestamp = 1000, .value = 45.5},
    Metric{.timestamp = 2000, .value = 46.0},
  };

  // 2. Send and capture request
  auto captured_request = server.CaptureLastRequest();
  sender.SendMetrics(metrics);

  // 3. Validate payload structure
  auto payload = json::parse(captured_request.body);
  ASSERT_TRUE(payload.contains("metrics"));
  ASSERT_TRUE(payload.contains("timestamp"));
  ASSERT_EQ(payload["metrics"].size(), 2);

  // 4. Validate each metric
  for (const auto& m : payload["metrics"]) {
    ASSERT_TRUE(m.contains("timestamp"));
    ASSERT_TRUE(m.contains("value"));
  }
}
```

**Validation Steps:**
```bash
cd /Users/glauco.torres/git/pganalytics-v3/collector

# 1. Rebuild
cmake -B build -DCMAKE_BUILD_TYPE=Debug
cmake --build build --target pganalytics-tests

# 2. Run integration tests
build/tests/pganalytics-tests --gtest_filter="SenderIntegrationTest*"

# 3. Check results
# Deve passar: SendMetricsSuccess, TokenExpiredRetry, etc
```

**Owner:** Collector Lead
**Dependency:** None

---

### Action 1.4: Increase Session Package Coverage
**Status:** 🟡 26.1% → TARGET 80%
**Effort:** 45 minutos

**File:** `/backend/internal/session/`

**Coverage Analysis:**
```bash
cd /backend
go test -coverprofile=coverage.out ./internal/session/...
go tool cover -html=coverage.out -o /tmp/session-coverage.html
# Abre /tmp/session-coverage.html para ver linhas não cobertas
```

**Add Tests For:**
```go
// ❌ Current gaps
func TestSessionExpiry(t *testing.T)           // ❌ Missing
func TestConcurrentSessionAccess(t *testing.T) // ❌ Missing
func TestSessionCleanup(t *testing.T)          // ❌ Missing
func TestInvalidSessionToken(t *testing.T)     // ❌ Missing

// Example test structure
func TestSessionExpiry(t *testing.T) {
  tests := []struct {
    name        string
    expiryTime  time.Duration
    elapsed     time.Duration
    expectValid bool
  }{
    {
      name:        "not expired",
      expiryTime:  1 * time.Hour,
      elapsed:     30 * time.Minute,
      expectValid: true,
    },
    {
      name:        "expired",
      expiryTime:  1 * time.Hour,
      elapsed:     2 * time.Hour,
      expectValid: false,
    },
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      session := NewSession(uuid.New().String())
      session.ExpiresAt = time.Now().Add(tt.expiryTime)

      // Simulate time passage
      session.ExpiresAt = session.ExpiresAt.Add(-tt.elapsed)

      isValid := !time.Now().After(session.ExpiresAt)
      if isValid != tt.expectValid {
        t.Errorf("expected %v, got %v", tt.expectValid, isValid)
      }
    })
  }
}
```

**Verification:**
```bash
cd /backend
go test -v ./internal/session/...
go test -cover ./internal/session/...
# Deve mostrar coverage > 80%
```

**Owner:** Backend Lead
**Dependency:** None

---

## 🟢 PRIORITY 2 - NEXT SPRINT (< 8 hours total)

### Action 2.1: Add Boundary/Edge Case Tests
**Status:** 🔴 10% COVERAGE
**Effort:** 4 hours

**Create:** `/frontend/src/components/UserList.boundary.test.tsx`

```typescript
describe('UserList - Boundary Cases', () => {
  it('should handle empty user list', async () => {
    mockApiClient.listUsers.mockResolvedValue({
      data: [],
      page: 1,
      page_size: 10,
      total: 0,
      total_pages: 0
    });

    render(<UserList />);

    expect(screen.getByText(/no users|empty/i)).toBeInTheDocument();
  });

  it('should handle very large user list (1000+)', async () => {
    const largeList = Array.from({ length: 1000 }, (_, i) => ({
      id: i,
      email: `user${i}@example.com`,
      name: `User ${i}`,
      role: 'viewer' as const,
    }));

    mockApiClient.listUsers.mockResolvedValue({
      data: largeList.slice(0, 10),  // First page
      page: 1,
      page_size: 10,
      total: 1000,
      total_pages: 100
    });

    render(<UserList />);

    // Should show pagination
    expect(screen.getByText(/1 - 10 of 1000/i)).toBeInTheDocument();
  });

  it('should handle very long usernames (500 chars)', async () => {
    const longName = 'a'.repeat(500);

    mockApiClient.listUsers.mockResolvedValue({
      data: [{
        id: 1,
        email: 'user@example.com',
        name: longName,
        role: 'viewer'
      }],
      page: 1,
      page_size: 10,
      total: 1,
      total_pages: 1
    });

    render(<UserList />);

    // Should truncate or handle gracefully
    const cells = screen.getAllByText((content) =>
      content.includes('a'.repeat(50))
    );
    expect(cells.length).toBeGreaterThan(0);
  });

  it('should handle unicode and special characters', async () => {
    const names = [
      'José María',
      'François Müller',
      '李明',
      '🎉 Party User',
      'User™®©'
    ];

    mockApiClient.listUsers.mockResolvedValue({
      data: names.map((name, i) => ({
        id: i,
        email: `user${i}@example.com`,
        name,
        role: 'viewer'
      })),
      page: 1,
      page_size: 10,
      total: names.length,
      total_pages: 1
    });

    render(<UserList />);

    names.forEach(name => {
      expect(screen.getByText(name)).toBeInTheDocument();
    });
  });

  it('should handle null/undefined values gracefully', async () => {
    mockApiClient.listUsers.mockResolvedValue({
      data: [{
        id: 1,
        email: 'user@example.com',
        name: null,
        role: null
      }],
      page: 1,
      page_size: 10,
      total: 1,
      total_pages: 1
    });

    render(<UserList />);

    // Should not crash
    expect(screen.getByRole('table')).toBeInTheDocument();
  });

  it('should handle pagination edge cases', async () => {
    mockApiClient.listUsers.mockResolvedValue({
      data: Array.from({ length: 5 }, (_, i) => ({
        id: i,
        email: `user${i}@example.com`,
        name: `User ${i}`,
        role: 'viewer'
      })),
      page: 10,  // Last page with fewer items
      page_size: 10,
      total: 95,
      total_pages: 10
    });

    render(<UserList />);

    expect(screen.getByText(/page 10 of 10/i)).toBeInTheDocument();
    expect(screen.queryByText(/next/i)).not.toBeInTheDocument();
  });
});
```

**Owner:** QA
**Dependency:** Action 1.1

---

### Action 2.2: Setup Coverage Enforcement in CI/CD
**Status:** ⚠️ WARNING ONLY, NO ENFORCEMENT
**Effort:** 90 minutos

#### Update: `.github/workflows/frontend-quality.yml`

```yaml
  coverage:
    name: Coverage Check
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'

      - name: Install dependencies
        working-directory: frontend
        run: npm install

      - name: Generate coverage report
        working-directory: frontend
        run: npm run test:coverage -- --reporter=json

      - name: Check coverage threshold
        working-directory: frontend
        run: |
          # Parse coverage JSON
          COVERAGE=$(cat coverage/coverage-final.json | jq '.total.lines.pct')
          echo "Total coverage: ${COVERAGE}%"

          MIN_COVERAGE=80
          if (( $(echo "$COVERAGE < $MIN_COVERAGE" | bc -l) )); then
            echo "❌ Coverage is below ${MIN_COVERAGE}% threshold (currently ${COVERAGE}%)"
            exit 1
          else
            echo "✅ Coverage meets ${MIN_COVERAGE}% threshold"
          fi

      - name: Comment PR with coverage
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const coverage = JSON.parse(fs.readFileSync('frontend/coverage/coverage-final.json', 'utf8'));

            const comment = `## Coverage Report

            - Lines: ${coverage.total.lines.pct}%
            - Functions: ${coverage.total.functions.pct}%
            - Branches: ${coverage.total.branches.pct}%
            - Statements: ${coverage.total.statements.pct}%
            `;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: comment
            });
```

#### Update: `.github/workflows/backend-tests.yml`

```yaml
  coverage-check:
    name: Coverage Threshold
    runs-on: ubuntu-latest
    needs: test

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Download coverage
        uses: actions/download-artifact@v4
        with:
          name: coverage.out

      - name: Check coverage threshold
        run: |
          # Expect 80% minimum
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Total coverage: ${COVERAGE}%"

          MIN_COVERAGE=80
          if (( $(echo "$COVERAGE < $MIN_COVERAGE" | bc -l) )); then
            echo "❌ Coverage below ${MIN_COVERAGE}% (current: ${COVERAGE}%)"
            exit 1
          fi
```

**Verification:**
```bash
# Test locally
npm run test:coverage

# Check format
cat coverage/coverage-final.json | jq '.total.lines.pct'
```

**Owner:** DevOps
**Dependency:** Action 1.2

---

### Action 2.3: Document Flaky Test Procedure
**Status:** ⚠️ NO PROCEDURE
**Effort:** 30 minutos

**Create:** `/docs/FLAKY_TESTS.md`

```markdown
# Flaky Test Reporting and Fixing

## How to Identify Flaky Tests

1. Run tests multiple times locally:
\`\`\`bash
for i in {1..10}; do npm test -- --reporter=json >> test-results.jsonl; done
\`\`\`

2. Check for inconsistent results:
\`\`\`bash
cat test-results.jsonl | grep '"success"' | sort | uniq -c
# If different results, test is flaky
\`\`\`

## Known Flaky Tests

| Test | Location | Issue | Fix |
|------|----------|-------|-----|
| (none reported yet) | - | - | - |

## Common Flaky Patterns

1. **Timeout-dependent tests**
   ```typescript
   // ❌ BAD
   await waitFor(() => expect(el).toBeVisible(), { timeout: 1000 })

   // ✅ GOOD
   await page.waitForNavigation();
   await expect(el).toBeVisible();
   ```

2. **Order-dependent tests**
   ```typescript
   // ❌ BAD - Depends on test execution order
   beforeEach(() => {
     state = globalState;  // May be modified by other tests
   });

   // ✅ GOOD
   beforeEach(() => {
     state = createFreshState();
   });
   ```

3. **External dependency races**
   ```typescript
   // ❌ BAD
   setTimeout(() => expect(state).toBe(value), 100);

   // ✅ GOOD
   await waitFor(() => expect(state).toBe(value), {
     timeout: 5000,  // Generous timeout
   });
   ```

## How to Report

If you find a flaky test:
1. Run it 10 times to confirm flakiness
2. Record failure rate and pattern
3. Create issue with tag `flaky-test`
4. Add to table above
5. DO NOT skip test - must be fixed instead

## CI/CD Retry Policy

- Frontend E2E: 2 retries on CI (see e2e-tests.yml)
- Backend: No retries (if flaky, fix immediately)
- Collector: No retries (if flaky, fix immediately)
```

**Owner:** QA Lead
**Dependency:** None

---

## 🔵 PRIORITY 3 - FUTURE QUARTERS (< 16 hours total)

### Action 3.1: Add API Contract Validation Tests
**Status:** ✅ TEST DEFINED, ❌ NOT EXECUTING
**Effort:** 2 hours (already written, just needs to run)

The test file exists at `/frontend/e2e/tests/10-api-contracts.spec.ts` with comprehensive API validation.

**Once E2E tests run (Action 1.2), these tests will execute automatically.**

**Verification:**
```bash
npm run test:e2e -- 10-api-contracts.spec.ts
```

---

### Action 3.2: Service-to-Service Integration Tests
**Status:** 🔴 NOT IMPLEMENTED
**Effort:** 8 hours

**Create:** `/backend/tests/integration/collector_backend_integration_test.go`

```go
package integration_test

import (
  "testing"
  "context"
  "pganalytics/backend/internal/collector"
  "pganalytics/backend/internal/database"
)

func TestCollectorMetricsFlow(t *testing.T) {
  // 1. Setup: Start backend API
  api := startTestBackend(t)
  defer api.Stop()

  // 2. Setup: Register collector
  collectorToken, err := api.RegisterCollector("test-collector")
  if err != nil {
    t.Fatalf("failed to register collector: %v", err)
  }

  // 3. Act: Collector sends metrics
  metrics := []collector.Metric{
    {Timestamp: 1000, Value: 45.5},
    {Timestamp: 2000, Value: 46.0},
  }

  err = sendMetricsToBackend(api.URL(), collectorToken, metrics)
  if err != nil {
    t.Fatalf("failed to send metrics: %v", err)
  }

  // 4. Assert: Verify metrics in database
  stored, err := api.DB().GetMetrics(context.Background(), "test-collector")
  if err != nil {
    t.Fatalf("failed to get metrics: %v", err)
  }

  if len(stored) != 2 {
    t.Errorf("expected 2 metrics, got %d", len(stored))
  }
}

func TestCollectorAuthExpiry(t *testing.T) {
  // Test token expiry and refresh flow
}

func TestCollectorErrorHandling(t *testing.T) {
  // Test invalid token, network errors, etc
}
```

---

### Action 3.3: Performance Regression Testing
**Status:** 🔴 NOT IMPLEMENTED
**Effort:** 4 hours

Add to `.github/workflows/backend-tests.yml`:

```yaml
  performance:
    name: Performance Benchmarks
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4

      - name: Run benchmarks
        working-directory: backend
        run: |
          go test -bench=. -benchmem ./... > bench-new.txt

          # Compare with baseline if exists
          if [ -f bench-baseline.txt ]; then
            go install golang.org/x/perf/cmd/benchstat@latest
            benchstat bench-baseline.txt bench-new.txt > bench-comparison.txt
            cat bench-comparison.txt
          fi

      - name: Store baseline
        run: cp bench-new.txt bench-baseline.txt

      - name: Upload results
        uses: actions/upload-artifact@v4
        with:
          name: benchmark-results
          path: bench-*.txt
```

---

## Summary Table

| Action | Priority | Effort | Status | Owner | Target |
|--------|----------|--------|--------|-------|--------|
| 0.1 Fix Playwright | 🔴 P0 | 5m | ❌ | DevOps | Today |
| 0.2 Remove Silent Errors | 🔴 P0 | 15m | ❌ | QA | Today |
| 0.3 Fix Backend Compilation | 🔴 P0 | 20m | ❌ | Backend | Today |
| 0.4 Fix User Creation E2E | 🔴 P0 | 10m | ❌ | QA | Today |
| 1.1 Zod Validation | 🟡 P1 | 2h | ❌ | Frontend | This Week |
| 1.2 Run E2E Tests | 🟡 P1 | 30m | ❌ | DevOps | This Week |
| 1.3 Fix Collector Tests | 🟡 P1 | 90m | ❌ | Collector | This Week |
| 1.4 Session Coverage | 🟡 P1 | 45m | ❌ | Backend | This Week |
| 2.1 Boundary Tests | 🟢 P2 | 4h | ❌ | QA | Next Sprint |
| 2.2 Coverage Enforcement | 🟢 P2 | 90m | ❌ | DevOps | Next Sprint |
| 2.3 Flaky Tests Doc | 🟢 P2 | 30m | ❌ | QA | Next Sprint |
| 3.1 API Contracts | 🔵 P3 | 2h | ✅ Written | - | Q2 2026 |
| 3.2 Service Tests | 🔵 P3 | 8h | ❌ | Backend | Q2 2026 |
| 3.3 Perf Tests | 🔵 P3 | 4h | ❌ | DevOps | Q2 2026 |

**Total Effort: ~32 hours**
**Recommended Timeline: 4 weeks at 8 hours/week**

---

## How to Track Progress

1. Create GitHub Issues for each action
2. Label with `priority:p0`, `priority:p1`, etc
3. Link to this document
4. Update status as work progresses
5. Close when verification steps pass

Example:
```markdown
# Fix Playwright Installation (Action 0.1)
**Priority:** 🔴 P0 (Blocker)
**Effort:** 5 minutes
**Assigned to:** @devops-lead

## Checklist
- [ ] npm install @playwright/test --save-dev
- [ ] npm run test:e2e -- --list (no errors)
- [ ] npm run test:e2e (tests run)
- [ ] Regression tests pass

## Notes
E2E tests are blocked without this
```

---

**Last Updated:** 14 de abril de 2026
**Next Review:** 21 de abril de 2026
**Owner:** QA/Testing Lead
