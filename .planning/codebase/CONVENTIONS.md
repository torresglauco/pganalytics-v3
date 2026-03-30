# Coding Conventions

**Analysis Date:** 2026-03-30

## Naming Patterns

**Files:**
- React components: PascalCase with `.tsx` extension (e.g., `CollectorForm.tsx`, `Dashboard.tsx`)
- Services/utilities: camelCase with `.ts` extension (e.g., `api.ts`, `realtime.ts`, `authStore.ts`)
- Test files: match source file name with `.test.ts` or `.test.tsx` suffix (e.g., `App.test.tsx`, `realtimeStore.test.ts`)
- Go files: snake_case with `.go` extension (e.g., `config_cache.go`, `key_manager.go`)
- Go test files: suffix `_test.go` (e.g., `circuit_breaker_test.go`, `mfa_test.go`)

**Functions:**
- Frontend: camelCase for all functions (e.g., `registerCollector()`, `handleConnected()`, `createMockApiClient()`)
- Go: PascalCase for exported functions, camelCase for unexported (e.g., `NewConfigCache()`, `Set()`, `hashData()`)
- Event handlers in React: prefix with `handle` (e.g., `handleConnected()`, `handleDisconnected()`, `handleError()`)

**Variables:**
- Frontend TypeScript: camelCase (e.g., `mockRegistrationSecret`, `connectAttempts`, `oldTimestamp`)
- Go: camelCase for package scope, PascalCase for exported variables
- Constants: UPPER_SNAKE_CASE (e.g., `COLORS`, `DEBUG`, `ERROR`, `FATAL`)

**Types:**
- Interfaces: PascalCase with optional `Props`, `State`, or `Response` suffix (e.g., `LogDistributionChartProps`, `AuthState`, `CollectorRegisterResponse`)
- Type aliases: PascalCase (e.g., `AuthMethod`, `MFAMethod`, `OAuthProvider`)
- Go structs: PascalCase (e.g., `ConfigCache`, `CachedConfig`, `ConfigChangeEvent`, `CircuitBreaker`)

## Code Style

**Formatting:**
- No explicit Prettier configuration found - code uses implicit defaults
- Indentation: 2 spaces in TypeScript/JavaScript, idiomatic Go indentation
- Line length: no hard limit enforced, but components stay within 100-120 columns
- Trailing commas: used in multi-line structures (arrays, objects, function parameters)

**Linting:**
- ESLint installed but no `.eslintrc` config file detected
- Package suggests usage: `npm run lint` executes `eslint src --ext ts,tsx`
- TypeScript strict mode enabled in `frontend/tsconfig.json`: `"strict": true`
- Compiler flags enforced:
  - `noUnusedLocals: true` - warns on unused local variables
  - `noUnusedParameters: true` - warns on unused function parameters
  - `noFallthroughCasesInSwitch: true` - prevents accidental fallthrough

## Import Organization

**Order:**
1. External libraries (React, axios, libraries from node_modules)
2. Type imports from external packages (e.g., `import type { ... } from 'axios'`)
3. Internal modules and services (from `src/`)
4. Type imports from internal modules (e.g., `import type { ... } from '../types'`)
5. Styles (e.g., `import './styles/index.css'`)

**Path Aliases:**
- No path aliases configured (no `jsconfig.paths` or `tsconfig.paths` configured)
- Use relative imports throughout: `../services/api`, `../stores/authStore`

**Pattern Example from `App.tsx`:**
```typescript
import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Dashboard } from './components/dashboard/Dashboard'
import { useAuthStore } from './stores/authStore'
import { realtimeClient } from './services/realtime'
import { apiClient } from './services/api'
import './styles/index.css'
```

## Error Handling

**Patterns:**
- Frontend API errors: wrapped in `try/catch` blocks with custom `handleError()` method
- Axios interceptor pattern: response interceptor checks for 401 status and redirects to login
- Grace degradation: UI renders fallback content when data unavailable
  - Example in `LogDistributionChart.tsx`: renders loading state, then "No data available" message
- Go errors: explicit error checks with `if err != nil { return err }` pattern
- Circuit breaker pattern for external service calls (see `backend/internal/ml/circuit_breaker.go`)

**Pattern Example from `api.ts`:**
```typescript
async registerCollector(
  data: CollectorRegisterRequest,
  registrationSecret: string
): Promise<CollectorRegisterResponse> {
  try {
    const response = await this.client.post<CollectorRegisterResponse>(...)
    return response.data
  } catch (error) {
    throw this.handleError(error)
  }
}
```

## Logging

**Framework:** console object in frontend (no dedicated logger), `go.uber.org/zap` in Go backend

**Patterns:**
- Frontend: `console.error()` for errors (e.g., `console.error('Auth check failed:', err)`)
- Go: structured logging with zap
  - Example: `logger.Debug("Config cached", zap.String("key", key), zap.Int("version", newVersion))`
  - Severity levels: `Debug()`, `Info()`, `Warn()`, `Error()`, `Fatal()`
- No logging statements in test files (use mocks instead)

## Comments

**When to Comment:**
- JSDoc-style comments for exported functions and types (not yet enforced but visible in some files)
- Inline comments for non-obvious logic (e.g., explaining why a workaround is needed)
- Test descriptions via `describe()` and `it()` function names are primary documentation

**JSDoc/TSDoc:**
- Not consistently used across codebase
- Package.json includes `@types/` packages for type definitions
- Go uses comment blocks above exported types (e.g., `// ConfigCache caches collector and query configurations with versioning`)

## Function Design

**Size:**
- Small focused functions preferred (50-100 lines typical)
- Zustand store actions are single-responsibility (e.g., `setConnected()`, `setError()`, `setLastUpdate()` are separate)

**Parameters:**
- Typed parameters required in TypeScript (no implicit `any`)
- Go uses receiver methods on structs: `func (cc *ConfigCache) Set(key string, data json.RawMessage) error`
- Optional parameters documented in function signatures

**Return Values:**
- Explicit return types required for TypeScript
- Go pattern: error as final return value (e.g., `func (...) error`)
- Promises for async operations in frontend
- Union types for conditional returns (e.g., returning `null` or error states)

## Module Design

**Exports:**
- Frontend: default exports for React components (e.g., `export const LogDistributionChart: React.FC<LogDistributionChartProps> = ...`)
- Zustand stores: default export of store hook (e.g., `export const useAuthStore = create<AuthState>(...)`)
- Services: export singleton instances (e.g., `export const apiClient = new ApiClient()`)
- Go: capitalized names for public API (e.g., `NewConfigCache()`, `Set()`, `Get()`)

**Barrel Files:**
- Used in `frontend/src/types/` for re-exporting multiple type definitions
- Example: `frontend/src/types/index.ts` exports from `./auth.ts`, `./api.ts`, etc.

---

*Convention analysis: 2026-03-30*
