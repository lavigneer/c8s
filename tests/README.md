# C8S Testing Infrastructure

Complete testing suite for C8S continuous integration system.

## Overview

C8S uses a comprehensive multi-layered testing approach:

- **Unit Tests**: Fast, isolated tests for individual functions and packages
- **Integration Tests**: Tests with Kubernetes envtest (fake API server)
- **E2E Tests**: Full end-to-end tests with Playwright (browser automation)
- **Contract Tests**: API contract validation (future)
- **Load Tests**: Performance and scalability tests (future)
- **Helm Tests**: Helm chart validation

## Test Statistics

| Test Type | Count | Runtime | Coverage |
|-----------|-------|---------|----------|
| Unit Tests | 50+ tests | ~2 min | Go packages |
| Integration Tests | 30+ tests | ~5 min | Controller, API |
| E2E Tests | 120+ tests | ~15 min | Full system |
| **Total** | **200+ tests** | **~22 min** | **Comprehensive** |

## Directory Structure

```
tests/
├── README.md                  # This file
├── unit/                      # Unit tests (Go)
│   ├── parser/               # Pipeline YAML parser tests
│   ├── scheduler/            # DAG scheduler tests
│   └── ...                   # Other package unit tests
├── integration/              # Integration tests (envtest)
│   ├── controller/           # Controller reconciliation tests
│   ├── webhook/              # Webhook integration tests
│   └── api/                  # API integration tests
├── e2e/                      # End-to-end tests (Playwright)
│   ├── specs/                # Test suites by feature
│   │   ├── authentication.spec.ts
│   │   ├── pipeline-creation.spec.ts
│   │   ├── log-viewing.spec.ts
│   │   ├── artifact-management.spec.ts
│   │   ├── cross-browser.spec.ts
│   │   ├── responsive.spec.ts
│   │   ├── performance.spec.ts
│   │   └── accessibility/    # WCAG 2.1 AA accessibility tests
│   ├── pages/                # Page Object Models
│   ├── fixtures/             # Test utilities and data
│   └── utils/                # Helper functions
├── helm/                     # Helm chart tests
│   └── lint-test.sh         # Helm lint validation
├── load/                     # Load testing (k6 or similar)
│   └── README.md            # Load testing guide
└── testutil/                 # Shared test utilities
    ├── mocks/               # Mock implementations
    ├── fixtures/            # Shared test data
    └── helpers/             # Test helper functions
```

## Running Tests

### Quick Commands

```bash
# Run all tests
make test-all

# Run specific test types
make test                # Unit + integration
make test-unit           # Unit tests only
make test-integration    # Integration tests only
make test-e2e            # E2E tests (all browsers)

# With coverage
make coverage            # Generate HTML coverage report
```

### Detailed Commands

#### Unit Tests
```bash
# Run all unit tests
go test ./tests/unit/... -v

# Run specific package
go test ./tests/unit/parser/... -v

# With coverage
go test ./tests/unit/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

#### Integration Tests
```bash
# Run all integration tests
make test-integration

# Run specific integration test
go test ./tests/integration/controller/... -v

# With race detection
go test ./tests/integration/... -race -v
```

#### E2E Tests
```bash
# Run all E2E tests
npm run test:e2e

# Run in UI mode (interactive)
npm run test:e2e:ui

# Run with debugger
npm run test:e2e:debug

# View test report
npm run test:e2e:report

# Run specific test file
npx playwright test tests/e2e/specs/authentication.spec.ts
```

## Unit Tests

**Location**: `tests/unit/`

**Purpose**: Test individual functions and packages in isolation

**Technology**: Go testing framework

**Key Characteristics**:
- Fast execution (<2 minutes)
- No external dependencies
- Mocked interfaces
- High code coverage target (80%+)

**Example**:
```go
// tests/unit/parser/parser_test.go
func TestParsePipelineConfig(t *testing.T) {
    yaml := `
version: v1alpha1
name: test-pipeline
steps:
  - name: build
    image: golang:1.25
    commands:
      - go build
`
    config, err := parser.Parse([]byte(yaml))
    assert.NoError(t, err)
    assert.Equal(t, "test-pipeline", config.Name)
    assert.Len(t, config.Steps, 1)
}
```

## Integration Tests

**Location**: `tests/integration/`

**Purpose**: Test components with Kubernetes API (using envtest)

**Technology**: Go + controller-runtime envtest

**Key Characteristics**:
- Uses fake Kubernetes API server
- Tests CRD create/update/delete
- Controller reconciliation loops
- Webhook validation
- Slower than unit tests (~5 minutes)

**Setup**:
```bash
# Install envtest binaries (one-time)
make envtest
```

**Example**:
```go
// tests/integration/controller/pipelinerun_test.go
func TestPipelineRunReconciliation(t *testing.T) {
    // Create PipelineRun CRD
    pr := &v1alpha1.PipelineRun{
        ObjectMeta: metav1.ObjectMeta{Name: "test-run"},
        Spec: v1alpha1.PipelineRunSpec{...},
    }
    err := k8sClient.Create(ctx, pr)
    assert.NoError(t, err)

    // Wait for controller to reconcile
    Eventually(func() bool {
        err := k8sClient.Get(ctx, key, pr)
        return err == nil && pr.Status.Phase == "Running"
    }).Should(BeTrue())
}
```

## E2E Tests

**Location**: `tests/e2e/`

**Purpose**: Full system testing with real browser automation

**Technology**: Playwright + TypeScript

**Coverage**:
- **Functional**: 29 tests (auth, pipelines, logs, artifacts)
- **Accessibility**: 38 tests (WCAG 2.1 Level AA compliance)
- **Cross-Browser**: 14 tests (Chrome, Firefox, Safari, Edge)
- **Responsive**: 11 tests (desktop, tablet, mobile)
- **Performance**: 11 tests (load time, memory, network)

**Key Features**:
- Page Object Model for maintainability
- axe-core integration for accessibility testing
- Cross-browser testing (4 browsers)
- Responsive design testing (3 viewports)
- Screenshot comparison
- Performance metrics capture

**Configuration**: See [playwright.config.ts](../playwright.config.ts)

**Documentation**: See [CLAUDE.md E2E Testing Framework](../CLAUDE.md#e2e-testing-framework-005-create-a-robust)

## Test Organization Guidelines

### Writing Good Unit Tests

1. **Fast**: No I/O, no sleeps, no external dependencies
2. **Isolated**: Mock all dependencies
3. **Deterministic**: Same input → same output
4. **Descriptive**: Test name explains what's being tested
5. **Arrange-Act-Assert**: Clear test structure

### Writing Good Integration Tests

1. **Use envtest**: Don't require real cluster
2. **Clean up**: Delete resources after test
3. **Wait properly**: Use Eventually() for async operations
4. **Test edge cases**: Not just happy path
5. **Parallel safe**: Can run concurrently

### Writing Good E2E Tests

1. **Use Page Objects**: Don't query DOM directly in tests
2. **Wait for elements**: Use Playwright's auto-waiting
3. **Test user workflows**: Not implementation details
4. **Independent**: Each test can run standalone
5. **Readable**: Test reads like user story

## CI/CD Integration

Tests run automatically in GitHub Actions:

| Workflow | Tests | Trigger |
|----------|-------|---------|
| **ci.yaml** | Unit + Integration | Every PR/push |
| **e2e-tests.yml** | E2E (Playwright) | PR affecting frontend |
| **tilt-ci.yml** | Full integration | PR/push to main |
| **c8s-dogfood.yml** | Real pipeline | PR/push (optional) |

See [.github/workflows/README.md](../.github/workflows/README.md) for details.

## Test Coverage

View coverage reports:

```bash
# Generate coverage
make coverage

# Open HTML report
open coverage.html
```

**Coverage Targets**:
- Unit Tests: 80%+ code coverage
- Integration Tests: All critical paths
- E2E Tests: All user-facing features

## Writing New Tests

### For New Features

When adding a new feature:

1. **Start with unit tests** - Test core logic
2. **Add integration tests** - Test Kubernetes integration
3. **Add E2E tests** - Test user workflows (if UI changes)
4. **Update test matrix** - Document new test coverage

### Test File Naming

- Unit tests: `*_test.go` alongside source
- Integration tests: `tests/integration/<package>/*_test.go`
- E2E tests: `tests/e2e/specs/<feature>.spec.ts`

### Running Tests Locally Before PR

```bash
# Full test suite (what CI runs)
make test-all

# Quick check
make test lint

# E2E tests (if frontend changes)
npm run test:e2e
```

## Troubleshooting Tests

### "envtest not found"
```bash
make envtest  # Downloads envtest binaries
```

### "Playwright browsers not installed"
```bash
npx playwright install
```

### "Tests timeout"
Increase timeout in test or check for deadlocks:
```go
// Go tests
go test -timeout 30m ./...

// Playwright tests
// Edit playwright.config.ts → timeout: 60000
```

### "Flaky tests"
1. Check for race conditions
2. Add proper waits (Eventually() in Go, waitFor() in Playwright)
3. Increase timeout for slow operations
4. Check test isolation (cleanup between tests)

## Performance

Test execution times:

| Environment | Unit | Integration | E2E | Total |
|-------------|------|-------------|-----|-------|
| **Local (M1 Mac)** | 1m | 3m | 10m | 14m |
| **GitHub Actions** | 2m | 5m | 15m | 22m |

## Future Improvements

- [ ] Contract tests for API endpoints
- [ ] Load tests with k6 or similar
- [ ] Mutation testing for test quality
- [ ] Visual regression testing
- [ ] Chaos engineering tests
- [ ] Security scanning in tests

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Playwright Documentation](https://playwright.dev/)
- [Kubebuilder Testing](https://book.kubebuilder.io/reference/testing)
- [Test Coverage](https://go.dev/blog/cover)

## Questions?

- **How do I test my controller changes?** → Integration tests in `tests/integration/controller/`
- **How do I test API changes?** → Unit tests + E2E tests
- **How do I test frontend changes?** → E2E tests with Playwright
- **How do I run tests faster?** → Run specific test subset: `go test ./tests/unit/parser/...`
- **Why are E2E tests failing in CI?** → Check Playwright report artifact in GitHub Actions
