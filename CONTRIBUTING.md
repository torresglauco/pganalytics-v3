# Contributing to pgAnalytics v3

Thank you for your interest in contributing to pgAnalytics! This document provides guidelines for submitting contributions, including code, documentation, and bug reports.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Environment](#development-environment)
4. [Git Workflow](#git-workflow)
5. [Code Standards](#code-standards)
6. [Testing Requirements](#testing-requirements)
7. [Commit Guidelines](#commit-guidelines)
8. [Pull Request Process](#pull-request-process)
9. [Documentation](#documentation)
10. [Security Issues](#security-issues)

## Code of Conduct

Be respectful and constructive in all interactions. We are committed to providing a welcoming and inclusive environment for all contributors.

## Getting Started

### Prerequisites

- **Go 1.22+** - Backend API development
- **Node.js 18+** - Frontend development
- **Docker 20.10+** - Running services locally
- **C++17 compiler** - Collector modifications
- **PostgreSQL 16+** - Database testing

### Clone the Repository

```bash
git clone https://github.com/torresglauco/pganalytics-v3.git
cd pganalytics-v3
```

## Development Environment

### Backend Setup

```bash
# Install Go dependencies
cd backend
go mod download

# Run tests
go test ./...

# Start API locally
go run cmd/pganalytics-api/main.go

# Access Swagger UI
# http://localhost:8080/swagger
```

### Frontend Setup

```bash
# Install dependencies
cd frontend
npm install

# Start dev server
npm run dev

# Run tests
npm test

# Build for production
npm run build
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## Git Workflow

### Branch Naming

Follow this naming convention for branches:

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions
- `chore/description` - Maintenance tasks

Examples:
```
feature/collector-management
fix/memory-leak-metrics
docs/upgrade-guide
test/e2e-dashboard
```

### Creating a Branch

```bash
# Update main
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/your-feature-name
```

## Code Standards

### Go Backend

#### Formatting & Style

- **Use `gofmt`** for code formatting
- **Follow `golangci-lint`** rules (configuration in `.golangci.yml`)
- **Line length**: Maximum 100 characters where practical
- **Package naming**: lowercase, no underscores

```bash
# Format code
gofmt -w .

# Run linter
golangci-lint run ./...

# Run tests with coverage
go test -cover ./...
```

#### Code Organization

```
backend/
├── cmd/pganalytics-api/          # Entry point
├── internal/
│   ├── api/                      # HTTP handlers
│   ├── auth/                     # Authentication
│   ├── storage/                  # Database layer
│   ├── metrics/                  # Metrics processing
│   ├── cache/                    # Caching logic
│   └── config/                   # Configuration
└── tests/                        # Tests
```

#### Naming Conventions

- **Exported functions**: `PascalCase`
- **Unexported functions**: `camelCase`
- **Variables**: `camelCase`
- **Constants**: `UPPER_CASE` (exported), `upper_case` (unexported)
- **Interfaces**: End with `-er` (e.g., `Reader`, `Writer`)

#### Error Handling

```go
// Good: Wrapping errors with context
if err != nil {
    return fmt.Errorf("failed to fetch metrics: %w", err)
}

// Avoid: Ignoring errors
_ = someFunction()

// Avoid: Generic error messages
return errors.New("error")
```

#### Comments

- Exported functions must have a comment starting with the function name
- Comments should explain "why", not "what"

```go
// FetchMetrics retrieves metrics from TimescaleDB for the specified time range.
func FetchMetrics(ctx context.Context, collectorID string, start, end time.Time) ([]Metric, error) {
    // Use prepared statements to prevent SQL injection
    stmt, err := db.Prepare("SELECT * FROM metrics WHERE collector_id = ? AND time BETWEEN ? AND ?")
    if err != nil {
        return nil, fmt.Errorf("failed to prepare statement: %w", err)
    }
    // ...
}
```

### TypeScript/React Frontend

#### Formatting & Style

- **Use ESLint** for linting
- **Use Prettier** for formatting (`.prettierrc` included)
- **Line length**: 100 characters max

```bash
# Lint and fix
npm run lint -- --fix

# Type checking
npm run type-check
```

#### File Organization

```
frontend/src/
├── components/          # React components
│   ├── common/         # Reusable components
│   ├── cards/          # Card components
│   ├── forms/          # Form components
│   └── charts/         # Chart components
├── pages/              # Page components
├── hooks/              # Custom hooks
├── services/           # API services
├── utils/              # Utility functions
├── types/              # TypeScript types
└── styles/             # Global styles
```

#### Naming Conventions

- **Files**: `PascalCase` for components, `camelCase` for utilities
- **Components**: `PascalCase`
- **Functions**: `camelCase`
- **Constants**: `UPPER_CASE`
- **Types/Interfaces**: `PascalCase`

#### Component Structure

```tsx
// Good: Clear component structure with proper typing
interface CollectorFormProps {
  onSubmit: (data: CollectorData) => Promise<void>;
  initialData?: Partial<CollectorData>;
  isLoading?: boolean;
}

export const CollectorForm: React.FC<CollectorFormProps> = ({
  onSubmit,
  initialData,
  isLoading = false,
}) => {
  // ... component logic

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      {/* form fields */}
    </form>
  );
};
```

#### Testing

- Test files: `ComponentName.test.tsx` (same directory as component)
- Use React Testing Library for testing
- Write meaningful test descriptions

```tsx
describe('CollectorForm', () => {
  it('should validate required hostname field', async () => {
    const { getByRole } = render(<CollectorForm onSubmit={vi.fn()} />);
    const submitBtn = getByRole('button', { name: /register/i });

    await userEvent.click(submitBtn);

    expect(getByText(/hostname is required/i)).toBeInTheDocument();
  });

  it('should call onSubmit with form data when valid', async () => {
    const onSubmit = vi.fn();
    const { getByLabelText, getByRole } = render(
      <CollectorForm onSubmit={onSubmit} />
    );

    await userEvent.type(getByLabelText(/hostname/i), 'localhost');
    await userEvent.click(getByRole('button', { name: /register/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
        hostname: 'localhost',
      }));
    });
  });
});
```

### Collector (C++)

#### Style Guide

- **Clang-format**: Use `.clang-format` configuration
- **Naming**: `camelCase` for functions/variables, `PascalCase` for classes
- **Comments**: Document public APIs and complex logic

```bash
# Format code
clang-format -i src/*.cpp include/*.h

# Build
cmake -B build
cmake --build build

# Run tests
ctest --test-dir build
```

## Testing Requirements

### Backend Tests

**Unit Tests**: Test individual functions in isolation

```go
func TestFetchMetrics_ValidInput(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer db.Close()

    // Act
    metrics, err := FetchMetrics(context.Background(), "collector-1", now, now.Add(1*time.Hour))

    // Assert
    require.NoError(t, err)
    require.NotEmpty(t, metrics)
}

func TestFetchMetrics_InvalidCollectorID(t *testing.T) {
    // Arrange
    db := setupTestDB(t)
    defer db.Close()

    // Act
    _, err := FetchMetrics(context.Background(), "", now, now.Add(1*time.Hour))

    // Assert
    require.Error(t, err)
}
```

**Integration Tests**: Test multiple components working together

- Located in `backend/tests/integration/`
- Use test database fixtures
- Clean up resources after tests

**Running Tests**

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestFetchMetrics ./...

# Coverage report
go test -cover ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Frontend Tests

**Unit Tests**: Test components in isolation

```tsx
import { render, screen } from '@testing-library/react';
import { CollectorList } from './CollectorList';

describe('CollectorList', () => {
  it('should render empty state when no collectors', () => {
    render(<CollectorList collectors={[]} />);
    expect(screen.getByText(/no collectors/i)).toBeInTheDocument();
  });

  it('should render collector items', () => {
    const collectors = [
      { id: '1', hostname: 'db1.example.com', port: 5432 },
    ];
    render(<CollectorList collectors={collectors} />);
    expect(screen.getByText('db1.example.com')).toBeInTheDocument();
  });
});
```

**Running Tests**

```bash
# Run all tests
npm test

# Run specific test file
npm test CollectorForm

# Watch mode
npm test -- --watch

# Coverage report
npm run test:coverage
```

### End-to-End Tests

Located in `frontend/e2e/` using Playwright:

```bash
# Run all E2E tests
npx playwright test

# Run specific test file
npx playwright test 01-login-logout.spec.ts

# Debug mode
npx playwright test --debug

# Run on specific browser
npx playwright test --project=chromium
```

**E2E Test Standards**

- Use Page Object Model pattern for reusability
- Meaningful test descriptions
- Proper error handling with timeouts
- Clean up state between tests

## Commit Guidelines

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting (no code changes)
- `refactor`: Code refactoring
- `test`: Test additions/modifications
- `chore`: Build, dependencies, etc.
- `perf`: Performance improvements
- `security`: Security fixes

### Examples

```
feat(auth): implement JWT token refresh mechanism

- Add refresh token endpoint
- Update auth middleware to handle token expiry
- Add refresh token rotation tests

Fixes #123
```

```
fix(metrics): prevent duplicate data insertion

Ensure metrics are deduplicated using collector_id + timestamp
combination as unique key.

Fixes #456
```

```
docs(contributing): clarify code style requirements

Added examples for Go package organization and TypeScript
naming conventions.
```

### Best Practices

- Keep commits focused and atomic
- One feature/fix per commit
- Write descriptive messages (50 char limit for subject)
- Reference issues in footer: `Fixes #123`, `Relates to #456`

## Pull Request Process

### Before Submitting

1. **Update from main**
   ```bash
   git fetch origin
   git rebase origin/main
   ```

2. **Run all tests**
   ```bash
   # Backend
   cd backend && go test ./... && cd ..

   # Frontend
   cd frontend && npm test && npm run lint && npm run type-check && cd ..
   ```

3. **Check for linting issues**
   ```bash
   # Backend
   cd backend && golangci-lint run ./... && cd ..

   # Frontend
   cd frontend && npm run lint && cd ..
   ```

### PR Title & Description

**Title Format**: Same as commit messages

```
feat(collector): add health check endpoint
```

**Description Template**

```markdown
## Summary
Briefly describe what this PR does.

## Changes
- Change 1
- Change 2
- Change 3

## Testing
How was this tested?
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
- [ ] Tests pass locally
```

### Review Process

1. **Initial Review**: Maintainers check for completeness
2. **Code Review**: Quality and standards verification
3. **Testing**: Automated tests run (CI/CD)
4. **Approval**: At least one maintainer approval
5. **Merge**: Maintainer merges to main

### Common Issues

| Issue | Solution |
|-------|----------|
| "Conflicts with main" | Rebase: `git rebase origin/main` |
| "Tests failing" | Check test output, fix issues locally |
| "Lint errors" | Run formatter: `gofmt -w .` or `npm run lint -- --fix` |
| "Missing tests" | Add tests for new functionality |

## Documentation

### Code Documentation

**Go**: Godoc comments for exported items

```go
// MetricsStore handles all metrics storage operations.
type MetricsStore interface {
    // InsertMetrics stores metrics in TimescaleDB.
    InsertMetrics(ctx context.Context, metrics []Metric) error
}
```

**TypeScript**: JSDoc for complex functions

```typescript
/**
 * Fetches collectors from the API and filters by status.
 * @param status - Filter by collector status (active, inactive, error)
 * @returns Promise resolving to filtered collectors
 */
export async function fetchCollectors(status?: string): Promise<Collector[]> {
  // ...
}
```

### Adding Documentation Files

When adding features, create/update:

- **Feature Documentation**: `docs/FEATURE_NAME.md`
- **API Documentation**: `docs/api/ENDPOINT_DESCRIPTION.md`
- **Deployment Guides**: `docs/DEPLOYMENT_*.md`
- **Runbooks**: `docs/RUNBOOK_*.md`

### Documentation Standards

- Clear headings and subheadings
- Code examples where applicable
- Links to related documentation
- Troubleshooting sections
- Updated table of contents

## Security Issues

### Reporting Vulnerabilities

**Do not open public issues for security vulnerabilities.**

Email security concerns to: security@pganalytics.local

Include:
- Description of vulnerability
- Affected versions
- Proof of concept (if possible)
- Suggested fix

### Security Review Checklist

Before submitting code with security implications:

- [ ] Input validation on all user-facing endpoints
- [ ] Parameterized queries for database access
- [ ] TLS/SSL for all network communication
- [ ] Proper error handling (no sensitive data in errors)
- [ ] Authentication/authorization checks
- [ ] No hardcoded secrets or credentials
- [ ] Dependencies updated and scanned for vulnerabilities

## Helpful Resources

- [SETUP.md](SETUP.md) - Development environment setup
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture
- [docs/API_SECURITY_REFERENCE.md](docs/API_SECURITY_REFERENCE.md) - API guidelines
- [GitHub Issues](https://github.com/torresglauco/pganalytics-v3/issues) - Bug reports and features

## Questions?

- Open a discussion in GitHub Discussions
- Check existing issues and pull requests
- Review documentation in `docs/` directory

Thank you for contributing to pgAnalytics! 🚀
