# Contributing to C8S

Thank you for your interest in contributing to C8S! This guide explains how to contribute effectively.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Review Process](#review-process)

---

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors.

### Our Values

- **Respectful**: Treat all contributors with respect
- **Inclusive**: Welcome people of all backgrounds
- **Constructive**: Focus on ideas, not personalities
- **Professional**: Keep discussions on-topic

### Expected Behavior

- Use welcoming and inclusive language
- Be respectful of differing opinions
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

### Unacceptable Behavior

- Harassment or discrimination
- Offensive language or comments
- Trolling or deliberate disruption
- Unwelcome advances or attention
- Publishing others' private information

---

## Getting Started

### Prerequisites

**For detailed setup, see**: [Development QUICKSTART](./development/QUICKSTART.md) (5-minute setup)

**Quick Setup**:
```bash
# 1. Install Devbox
curl -fsSL https://get.jetify.com/devbox | bash

# 2. Clone repository
git clone https://github.com/lavigneer/c8s.git
cd c8s

# 3. Enter dev environment
devbox shell

# 4. Start development
make dev    # or: tilt up
```

**Alternative manual setup**: See [CLAUDE.md](../CLAUDE.md)

### Development Tools

- **Go 1.25+** - Implementation language
- **Kubernetes 1.24+** - Development cluster
- **Tilt** or **Make** - Local development
- **git** - Version control

### Project Structure

See [pkg/README.md](../pkg/README.md) for detailed package documentation.

```
c8s/
├── cmd/                    # Executable entry points
├── pkg/                    # Shared packages (see pkg/README.md)
├── tests/                  # Test files (see tests/README.md)
├── docs/                   # Documentation
│   ├── guides/            # User guides
│   ├── development/       # Developer docs
│   └── operations/        # Operations guides
├── chart/c8s/             # Helm chart
├── config/                # CRD manifests
├── tilt/                  # Tilt configuration
└── Makefile              # Build commands
```
```

---

## Development Workflow

### Creating a Branch

```bash
# Start from main
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/short-description

# Examples:
# feature/add-webhook-validation
# fix/memory-leak-in-controller
# docs/update-readme
```

### Local Development

1. **Make changes** to source code
2. **Write/update tests** for changes
3. **Run tests locally**: `make test`
4. **Build binaries**: `make build`
5. **Test manually** (see CLAUDE.md for instructions)

### Testing Locally

```bash
# Run all tests
make test

# Run specific test
go test ./tests/unit/handlers -v -run TestListProjects

# Run with coverage
make test-coverage

# Integration test
make test-integration

# E2E test
npm run test:e2e
```

---

## Code Style

### Go Code

Follow standard Go conventions:

```bash
# Format code
gofmt -s -w .
go vet ./...

# Lint
golangci-lint run
```

**Style Guidelines**:

1. **Names**: Use clear, descriptive names
   ```go
   // Good
   func (c *Controller) processNextWorkItem() error {}

   // Avoid
   func (c *Controller) process() error {}
   ```

2. **Error Handling**: Always handle errors
   ```go
   // Good
   if err := operation(); err != nil {
       log.Printf("Operation failed: %v", err)
       return err
   }

   // Avoid
   operation()  // Error ignored
   ```

3. **Comments**: Document exported items
   ```go
   // Good
   // ListProjects returns all projects for the authenticated user
   func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {}

   // Avoid
   func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {}
   ```

4. **Constants**: Use UPPER_SNAKE_CASE
   ```go
   const (
       DefaultTimeout  = 30 * time.Second
       MaxRetries      = 3
   )
   ```

### JavaScript/TypeScript

For dashboard and tests:

```bash
# Format
prettier --write .

# Lint
npm run lint
```

### YAML/Kubernetes Manifests

```yaml
# Indentation: 2 spaces
# Comments: Explain complex configurations
apiVersion: v1
kind: ConfigMap
metadata:
  name: example        # Lowercase with hyphens
  namespace: default
data:
  key: value
```

---

## Testing

### Unit Tests

Write unit tests for all new code:

```go
func TestListProjectsHandlerSuccess(t *testing.T) {
    // Arrange
    mockK8sClient := &MockK8sClient{...}

    // Act
    w := httptest.NewRecorder()
    handler := ListProjectsHandler(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Integration Tests

For cross-component testing:

```go
func TestControllerWithKubernetes(t *testing.T) {
    // Set up test Kubernetes cluster
    env := envtest.NewEnvironment(...)

    // Create resources
    pipelineConfig := createTestPipeline()

    // Verify behavior
    assert.EventuallyWithT(...)
}
```

### E2E Tests

For user workflows (see [CLAUDE.md](./CLAUDE.md) for E2E test framework):

```typescript
test('User can create and run a pipeline', async ({ page }) => {
    // Login
    await page.goto('/login');

    // Create pipeline
    await page.click('text=Create Pipeline');

    // Verify success
    await expect(page).toHaveURL(/.*pipelines/);
});
```

### Test Coverage

- Aim for **90%+ coverage** on critical code paths
- Use `make test-coverage` to check
- Review coverage report: `go tool cover -html=coverage.out`

---

## Documentation

### Code Documentation

- **Export godoc comments** on all public types/functions
- **Example code** in complex packages
- **README** in package for non-obvious behavior

Example:
```go
// Package dashboard provides HTTP handlers for the web dashboard.
//
// The handlers implement REST APIs for viewing pipeline runs,
// artifacts, and logs. All handlers require authentication via
// the Auth middleware.
package dashboard

// ListProjectsHandler returns all projects accessible to the authenticated user.
//
// GET /api/projects
//
// Returns 401 if user is not authenticated.
// Returns 200 with JSON array of projects.
func ListProjectsHandler(w http.ResponseWriter, r *http.Request) {}
```

### User Documentation

- **Update docs/** files when adding features
- **Keep examples current** with code changes
- **Test documentation** steps manually
- **Link related docs** between files

### Commit Messages

Follow the format in [CLAUDE.md](./CLAUDE.md):

```
[Task ID] Brief description

Longer explanation of changes if needed.
Mention related issues or PRs.

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
```

Examples:
```
[T084] Fix memory leak in pipeline controller

The controller was not properly cleaning up goroutines
when pipelines completed. Add context cancellation.

Fixes: #234
```

---

## Submitting Changes

### Pull Request Checklist

- [ ] Branch is up-to-date with main
- [ ] Code follows style guidelines
- [ ] Tests pass: `make test`
- [ ] Coverage maintained/improved
- [ ] Documentation updated
- [ ] Commit messages follow format
- [ ] No hardcoded passwords/secrets

### Creating a Pull Request

```bash
# Push branch
git push -u origin feature/description

# Create PR via GitHub CLI
gh pr create \
  --title "Brief description" \
  --body "Longer explanation of changes"

# Or create via GitHub web interface
```

### PR Description Template

```markdown
## Summary
Brief explanation of what this PR does.

## Changes
- What was changed
- Why it was changed
- How it was tested

## Related Issues
Fixes #123

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Documentation
- [ ] Code comments updated
- [ ] User docs updated (if needed)
- [ ] Architecture docs updated (if needed)
```

---

## Review Process

### Code Review

All PRs require review from at least one maintainer.

**Review Criteria**:
- Code follows style guidelines
- Tests are adequate (90%+ coverage)
- Documentation is clear
- No security issues
- No breaking changes (without discussion)

### Getting Review Feedback

- Be receptive to suggestions
- Ask for clarification if unclear
- Make requested changes promptly
- Request re-review after changes

### Merging

Once approved:
- Squash commits if many small commits
- Ensure all tests pass
- Ensure branch is up-to-date
- Maintainer merges the PR

---

## Questions?

- **GitHub Issues**: Open an issue for bugs or features
- **GitHub Discussions**: Ask questions in discussions
- **Email**: contact@c8s.dev
- **Slack**: Join our community Slack

---

## License

By contributing, you agree your code is licensed under Apache 2.0 (same as C8S).

Thank you for contributing to C8S! 🎉
