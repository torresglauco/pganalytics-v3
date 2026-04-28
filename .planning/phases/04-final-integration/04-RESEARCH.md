# Phase 4: Frontend Integration Testing & Quality - Research

**Researched:** 2026-04-28
**Domain:** Frontend testing (Vitest, Testing Library, Playwright), TypeScript linting (ESLint), React component integration
**Confidence:** HIGH

## Summary

Phase 4 focuses on validating frontend integration with the backend API and establishing code quality tooling for TypeScript. The project has substantial existing test infrastructure (30+ unit test files, 11 E2E test specs), but ESLint configuration is missing entirely. The authentication system uses httpOnly cookies, requiring specific testing approaches for session persistence. Form components use react-hook-form with Zod validation, and state management uses Zustand stores.

**Primary recommendation:** Create ESLint flat config first to unblock QUAL-02, then enhance existing unit tests for API integration scenarios and add authentication persistence tests.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Vitest | 1.6.1 | Unit testing framework | Fast, Vite-native, Jest-compatible API |
| @testing-library/react | 14.1.2 | React component testing | User-centric testing, recommended by React team |
| @playwright/test | 1.59.1 | E2E testing | Cross-browser, built-in fixtures, API testing support |
| eslint | 8.57.1 | Linting (needs config) | Industry standard, highly configurable |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @vitest/coverage-v8 | 1.0.0 | Coverage reporting | Generate coverage reports |
| @testing-library/user-event | 14.5.1 | User interaction simulation | Form testing, click events |
| @testing-library/jest-dom | 6.1.5 | Custom DOM matchers | Better assertions (toBeInTheDocument, etc.) |
| jsdom | 23.0.1 | DOM environment for tests | Already configured in vite.config.ts |
| msw | N/A | API mocking | NOT installed - consider for API mocking |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Vitest | Jest | Jest requires more config, slower. Vitest is native to Vite |
| Playwright | Cypress | Playwright faster, better API testing, already configured |
| No API mocking | MSW (Mock Service Worker) | MSW provides realistic API mocking for integration tests |

**Installation:**
```bash
# ESLint flat config packages needed
npm install -D @typescript-eslint/parser @typescript-eslint/eslint-plugin eslint-plugin-react eslint-plugin-react-hooks eslint-config-prettier
```

**Version verification:**
```bash
npm view vitest version          # 4.1.5 (project uses 1.6.1)
npm view @testing-library/react version  # 16.3.2 (project uses 14.1.2)
npm view @playwright/test version        # 1.59.1 (current)
npm view eslint version          # 10.2.1 (project uses 8.57.1 - flat config requires 9+)
npm view typescript version      # 6.0.3 (project uses 5.3.3)
```

## Architecture Patterns

### Recommended Project Structure
```
frontend/
├── src/
│   ├── components/           # React components (many have .test.tsx files)
│   ├── hooks/                # Custom hooks (some have .test.ts files)
│   ├── services/             # API services (some have .test.ts files)
│   ├── stores/               # Zustand stores (some have .test.ts files)
│   ├── test/                 # Test utilities and setup
│   │   ├── setup.ts          # Vitest setup (jsdom, localStorage mock)
│   │   └── utils.ts          # Test helpers, mock factories
│   └── __tests__/            # Integration tests
│       └── integration/
├── e2e/
│   ├── fixtures/             # Playwright fixtures (auth.ts)
│   ├── pages/                # Page Object Models
│   └── tests/                # E2E test specs
├── vite.config.ts            # Vitest config embedded
└── playwright.config.ts      # Playwright config
```

### Pattern 1: Component Testing with Mocked API
**What:** Test components in isolation with mocked API responses
**When to use:** Unit tests for components that fetch data
**Example:**
```typescript
// Source: Existing pattern in frontend/src/hooks/useCollectors.test.ts
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { renderHook, waitFor } from '@testing-library/react'
import { useCollectors } from './useCollectors'
import { apiClient } from '../services/api'

vi.mock('../services/api')

describe('useCollectors', () => {
  const mockCollectors = [
    { id: 'collector-1', hostname: 'localhost', status: 'active' },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should fetch collectors on mount', async () => {
    vi.mocked(apiClient.listCollectors).mockResolvedValue({
      data: mockCollectors,
      total: 1,
      page: 1,
      page_size: 20,
      total_pages: 1,
    })

    const { result } = renderHook(() => useCollectors())

    await waitFor(() => {
      expect(result.current.loading).toBe(false)
    })

    expect(result.current.collectors).toEqual(mockCollectors)
  })
})
```

### Pattern 2: Form Validation Testing
**What:** Test form components with react-hook-form and Zod validation
**When to use:** Testing form submission, validation errors, success states
**Example:**
```typescript
// Source: Pattern from frontend/src/components/CollectorForm.test.tsx
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { screen } from '@testing-library/react'
import { CollectorForm } from './CollectorForm'
import { apiClient } from '../services/api'
import { render } from '../test/utils'

vi.mock('../services/api')

describe('CollectorForm', () => {
  const mockOnSuccess = vi.fn()
  const mockOnError = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiClient.testConnection).mockResolvedValue(true)
    vi.mocked(apiClient.registerCollector).mockResolvedValue({
      collector_id: 'new-collector',
      status: 'registered',
      token: 'collector-token',
      created_at: '2024-01-01T00:00:00Z',
    })
  })

  it('should render form with hostname field', () => {
    render(
      <CollectorForm
        registrationSecret="test-secret"
        onSuccess={mockOnSuccess}
        onError={mockOnError}
      />
    )
    expect(document.body).toBeInTheDocument()
  })
})
```

### Pattern 3: E2E Authentication Testing
**What:** Test authentication flow with Playwright using API fixtures
**When to use:** Testing login, logout, session persistence
**Example:**
```typescript
// Source: frontend/e2e/tests/01-login-logout.spec.ts
import { test, expect } from '@playwright/test'
import { LoginPage } from '../pages/LoginPage'
import { DashboardPage } from '../pages/DashboardPage'

test.describe('Login/Logout Flow', () => {
  test('should login with valid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page)
    const dashboardPage = new DashboardPage(page)

    await loginPage.goto()
    await loginPage.login('admin', 'admin')
    await loginPage.expectLoggedIn()
    await dashboardPage.expectLoaded()
  })

  test('should maintain session after page reload', async ({ page }) => {
    const loginPage = new LoginPage(page)
    await loginPage.goto()
    await loginPage.login('admin', 'admin')
    await loginPage.expectLoggedIn()

    await page.reload()
    await expect(page).toHaveURL(/dashboard/)
  })
})
```

### Pattern 4: ESLint Flat Config (REQUIRED - Missing)
**What:** Modern ESLint configuration using flat config format
**When to use:** Required for ESLint 9+, recommended for all new configs
**Example:**
```javascript
// eslint.config.mjs (TO BE CREATED)
import js from '@eslint/js'
import typescript from '@typescript-eslint/eslint-plugin'
import typescriptParser from '@typescript-eslint/parser'
import react from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'

export default [
  js.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      parser: typescriptParser,
      parserOptions: {
        ecmaVersion: 2020,
        sourceType: 'module',
        ecmaFeatures: { jsx: true },
      },
    },
    plugins: {
      '@typescript-eslint': typescript,
      'react': react,
      'react-hooks': reactHooks,
    },
    rules: {
      // TypeScript strict rules
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',

      // React rules
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      'react/jsx-uses-react': 'off',
      'react/react-in-jsx-scope': 'off',
    },
    settings: {
      react: { version: 'detect' },
    },
  },
  {
    ignores: ['dist/', 'node_modules/', 'coverage/', 'e2e/'],
  },
]
```

### Anti-Patterns to Avoid
- **Testing implementation details:** Use `getByRole`, `getByText` not `querySelector` or test IDs
- **Mocking child components excessively:** Prefer integration tests over heavy mocking
- **Testing localStorage for auth:** The app uses httpOnly cookies - test via API calls, not localStorage
- **Skipping ESLint config:** Must create config before other quality improvements

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| API mocking | Custom fetch interceptors | MSW (Mock Service Worker) | Service worker isolation, realistic network behavior |
| Form validation tests | Manual state assertions | react-hook-form's `handleSubmit` mock | Form library handles validation complexity |
| Auth state in tests | Manual Zustand state setup | Use `useAuthStore.getState().setToken()` | Store state management is complex |
| Test fixtures | Manual data creation | Use `src/test/utils.ts` mock factories | Existing factories provide consistent test data |

**Key insight:** The test utilities file (`src/test/utils.ts`) already provides `mockUser`, `mockAdminUser`, `mockCollector`, `mockAuthResponse` factories. Use these instead of creating new mock data.

## Common Pitfalls

### Pitfall 1: Testing httpOnly Cookie Authentication Incorrectly
**What goes wrong:** Tests try to read `localStorage.getItem('auth_token')` but token is in httpOnly cookie
**Why it happens:** Code was recently changed from localStorage to httpOnly cookies for security
**How to avoid:**
- Test authentication via API calls (`page.request.post('/api/v1/auth/login')`)
- Use Playwright's `storageState` for session persistence tests
- Check `isAuthenticated` state in Zustand store, not localStorage
**Warning signs:** Tests checking `localStorage.getItem('auth_token')` will fail

### Pitfall 2: ESLint Legacy Config vs Flat Config
**What goes wrong:** Creating `.eslintrc.js` or `.eslintrc.json` when using ESLint 8.57.1
**Why it happens:** Documentation often shows legacy config format
**How to avoid:**
- Use `eslint.config.mjs` or `eslint.config.js` (flat config format)
- ESLint 9+ requires flat config; 8.57.1 supports both but flat config is future
- The `npm run lint` currently fails because NO config exists
**Warning signs:** `npm run lint` returns "ESLint couldn't find a configuration file"

### Pitfall 3: Incomplete Form Test Coverage
**What goes wrong:** Tests only render forms without testing validation, submission, or error states
**Why it happens:** Current test files like `CollectorForm.test.tsx` have minimal assertions
**How to avoid:**
- Test validation errors for required fields
- Test successful form submission flow
- Test API error handling in form
- Use `userEvent` for realistic interactions
**Warning signs:** Tests like `expect(document.body).toBeInTheDocument()` without form-specific assertions

### Pitfall 4: Missing Navigation State Tests
**What goes wrong:** Navigation state (filters, pagination) lost when navigating between pages
**Why it happens:** React Router state not properly managed or tested
**How to avoid:**
- Test URL parameters persist across navigation
- Test that table filters survive page navigation
- Use Playwright to verify URL state in E2E tests
**Warning signs:** No tests for navigation state persistence

## Code Examples

Verified patterns from existing test files:

### Dashboard Component with API Data (TEST-12)
```typescript
// Pattern from frontend/src/pages/Dashboard.test.tsx and hooks tests
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { Dashboard } from './Dashboard'
import { apiClient } from '../services/api'

vi.mock('../services/api')

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(apiClient.getCurrentUser).mockResolvedValue({
      id: 1,
      username: 'admin',
      email: 'admin@example.com',
      role: 'admin',
      is_active: true,
      created_at: '2024-01-01',
      updated_at: '2024-01-01',
    })
  })

  it('should load and display dashboard data', async () => {
    render(<Dashboard />)

    await waitFor(() => {
      expect(screen.getByRole('heading')).toBeInTheDocument()
    })
  })
})
```

### API Error Handling in UI (TEST-15)
```typescript
// Pattern from frontend/src/hooks/useCollectors.test.ts
it('should handle fetch error', async () => {
  const error = {
    message: 'Failed to fetch collectors',
    status_code: 500,
  }

  vi.mocked(apiClient.listCollectors).mockRejectedValue(error)

  const { result } = renderHook(() => useCollectors())

  await waitFor(() => {
    expect(result.current.loading).toBe(false)
  })

  expect(result.current.error).toEqual(error)
  expect(result.current.collectors).toEqual([])
})
```

### Authentication Persistence Test (TEST-16)
```typescript
// Pattern from frontend/e2e/tests/01-login-logout.spec.ts
test('should maintain session after page reload', async ({ page }) => {
  await loginPage.goto()
  await loginPage.login('admin', 'admin')
  await loginPage.expectLoggedIn()

  // Reload page - httpOnly cookie should persist session
  await page.reload()

  // Should still be logged in (cookie-based auth)
  await dashboardPage.expectLoaded()
  expect(page.url()).toContain('/dashboard')
})
```

### Zustand Store Authentication State
```typescript
// Source: frontend/src/stores/authStore.ts
import { create } from 'zustand'

interface AuthState {
  user: User | null
  token: string | null  // Note: actual token in httpOnly cookie
  isAuthenticated: boolean
  // ...
  setToken: (token: string) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,  // Token is in httpOnly cookie, not localStorage
  isAuthenticated: false,
  setToken: (token) => set({ token: '', isAuthenticated: true }),
  logout: () => set({ user: null, token: null, isAuthenticated: false }),
}))
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| localStorage for auth tokens | httpOnly cookies | Phase 1 (Week 1) | More secure, CSRF protection, requires different testing |
| Jest for unit testing | Vitest | Project start | Faster, Vite-native |
| .eslintrc config | eslint.config.mjs (flat config) | Pending | ESLint 9+ requires flat config format |
| Manual mock data | Test factories in utils.ts | Existing | Consistent test data across test files |

**Deprecated/outdated:**
- `localStorage.getItem('auth_token')`: No longer used - httpOnly cookies
- `.eslintrc.*` files: Deprecated in ESLint 9+ - use flat config

## Open Questions

1. **Should we install MSW (Mock Service Worker)?**
   - What we know: Tests currently use `vi.mock('../services/api')` to mock entire API client
   - What's unclear: MSW would provide more realistic API mocking for integration tests
   - Recommendation: Defer MSW installation - current vi.mock approach works for unit tests. Consider MSW for Phase 5 if integration testing needs expand.

2. **Upgrade ESLint to 9.x or stay on 8.57.1?**
   - What we know: ESLint 9.x requires flat config; 8.57.1 supports both
   - What's unclear: Flat config is future-proof but may have plugin compatibility issues
   - Recommendation: Stay on ESLint 8.57.1 with flat config format (`eslint.config.mjs`) - this works with 8.x and provides migration path to 9.x

3. **How comprehensive should form validation tests be?**
   - What we know: Current form tests are minimal (just checking render)
   - What's unclear: Balance between test coverage and test maintenance
   - Recommendation: Test the critical validation paths (required fields, format validation, API errors) not every edge case

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Vitest 1.6.1 + @testing-library/react 14.1.2 |
| Config file | vite.config.ts (embedded test config) |
| Quick run command | `npm run test -- --run` |
| Full suite command | `npm run test` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TEST-12 | Dashboard components render with API data | unit | `vitest run src/pages/Dashboard.test.tsx` | YES |
| TEST-13 | Form validation and error display | unit | `vitest run src/components/CollectorForm.test.tsx` | YES (minimal) |
| TEST-14 | Navigation state persistence | e2e | `playwright test e2e/tests/07-pages-navigation.spec.ts` | YES |
| TEST-15 | API error handling in UI | unit | `vitest run src/hooks/useCollectors.test.ts` | YES |
| TEST-16 | Auth persistence with httpOnly cookies | e2e | `playwright test e2e/tests/01-login-logout.spec.ts` | YES |
| QUAL-02 | TypeScript passes ESLint strict config | lint | `npm run lint` | NO - config missing |
| QUAL-04 | Code comments explain "why" | manual | Code review | N/A |

### Sampling Rate
- **Per task commit:** `npm run test -- --run` (quick unit tests)
- **Per wave merge:** `npm run test && npm run test:e2e` (full suite)
- **Phase gate:** All tests green + `npm run lint` passes

### Wave 0 Gaps
- [ ] `frontend/eslint.config.mjs` - ESLint flat config required for QUAL-02
- [ ] `frontend/src/components/CollectorForm.test.tsx` - Enhanced validation tests for TEST-13
- [ ] Enhanced navigation state tests in `e2e/tests/07-pages-navigation.spec.ts` for TEST-14

*(If no gaps: "None - existing test infrastructure covers all phase requirements")*

**Note:** Test infrastructure exists but ESLint config is completely missing. This is the primary Wave 0 gap.

## Sources

### Primary (HIGH confidence)
- Project files analyzed: package.json, vite.config.ts, playwright.config.ts, tsconfig.json
- Existing test files: 30+ unit tests, 11 E2E specs
- src/test/utils.ts - Test utilities and mock factories
- src/stores/authStore.ts - Authentication state management (httpOnly cookies)

### Secondary (MEDIUM confidence)
- Playwright test fixtures: e2e/fixtures/auth.ts
- UI Structure documentation: frontend/UI_STRUCTURE.md

### Tertiary (LOW confidence)
- Web search for ESLint flat config best practices (API errors)
- Web search for Vitest testing patterns (API errors)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Verified from package.json and test execution
- Architecture: HIGH - Analyzed 30+ existing test files and patterns
- Pitfalls: HIGH - Identified from code analysis (httpOnly cookies, missing ESLint config)
- ESLint config: MEDIUM - Need to create flat config (standard pattern but not yet verified in project)

**Research date:** 2026-04-28
**Valid until:** 30 days - stable testing ecosystem

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TEST-12 | Dashboard components render correctly with API data | Existing Dashboard.test.tsx, useCollectors.test.ts patterns |
| TEST-13 | Form components validate input and display errors correctly | CollectorForm.test.tsx exists but needs enhancement |
| TEST-14 | Navigation between pages maintains state properly | 07-pages-navigation.spec.ts exists, may need filter/pagination tests |
| TEST-15 | API error responses handled gracefully in UI | useCollectors.test.ts has error handling tests |
| TEST-16 | Authentication state persists across page refresh with httpOnly cookies | 01-login-logout.spec.ts has session persistence test |
| QUAL-02 | TypeScript code passes ESLint with strict config | **ESLint config missing** - must create eslint.config.mjs |
| QUAL-04 | Code comments explain "why" not "what" for complex logic | Manual code review required - no automated check |