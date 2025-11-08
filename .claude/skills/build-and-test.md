# Build and Test Workflow

Build all binaries and run tests to verify the codebase is healthy.

## Running in Devbox (Recommended)
All commands should be run in devbox to ensure consistent environments:
```bash
devbox run make all             # Build + test + generate everything
devbox run make build           # Build all binaries
devbox run make test            # Run all unit tests
devbox run make lint            # Run linter
devbox run make generate        # Generate code (DeepCopy, client, etc.)
devbox run make manifests       # Generate CRD manifests
devbox run make test-integration # Run integration tests
```

Or enter devbox shell once:
```bash
devbox shell
make all
make test
make lint
# etc...
```

## Development Workflow
1. Enter devbox: `devbox shell`
2. Make code changes
3. Run `make fmt` to format code
4. Run `make lint` to check for style issues
5. Run `make test` to verify tests pass
6. Run `make generate && make manifests` if you modified CRD types
7. Commit with format: `[Txxx] Description`

## Common Patterns (run in devbox shell or with `devbox run`)
- **Quick iteration**: `make lint && make test`
- **Before committing**: `make fmt && make lint && make test`
- **After CRD changes**: `make generate && make manifests`
- **Full verification**: `make all`

## Test Types
- **Unit tests**: `make test` - Fast, for development
- **Integration tests**: `make test-integration` - Uses envtest, slower
- **Coverage**: `make coverage` - Generates HTML coverage report

## Notes
- Tests run with CGO_ENABLED=1 and -race flag for data race detection
- Timeout: 10 minutes
- Coverage report output: coverage.html
- Linting uses golangci-lint configuration
