# E2E Testing Framework

Run end-to-end tests using Playwright to verify the complete application.

## Running in Devbox (Recommended)
All commands should be run in devbox to ensure consistent environments:
```bash
devbox run npm run test:e2e           # Run all E2E tests
devbox run npm run test:e2e:ui        # Interactive test UI (useful for debugging)
devbox run npm run test:e2e:debug     # Run with Playwright Inspector
devbox run npm run test:e2e:report    # View HTML test report
```

Or enter devbox shell once:
```bash
devbox shell
npm run test:e2e
npm run test:e2e:ui
# etc...
```

## Initial Setup
```bash
devbox run npm install                 # Install dependencies
devbox run npx playwright install      # Install browser binaries
# or from within devbox shell:
npm install
npx playwright install
```

## Test Organization
Located in `tests/e2e/`:
- **specs/**: Test suites organized by feature area
- **pages/**: Page Object Models for maintainability and reusability
- **fixtures/**: Test utilities and helpers (authentication, test data, reporting)

## Best Practices
- Use Page Object Model for interacting with pages
- Tests verify user-visible behavior, not implementation details
- Don't depend on HTMX internals or framework details
- Tests should work the same regardless of frontend implementation
- Use API-based setup for test data when possible

## Running Tests
- **All tests**: `npm run test:e2e`
- **With UI**: `npm run test:e2e:ui` - Watch mode with visual feedback
- **Debug mode**: `npm run test:e2e:debug` - Opens Playwright Inspector
- **View results**: `npm run test:e2e:report` - Opens HTML test report

## CI/CD Integration
- Runs automatically on PR creation and push to main
- Tests run across multiple browser and viewport combinations
- Artifacts captured on failure (screenshots, videos)
- Results commented on PRs

## Debugging Tips
1. Use `npm run test:e2e:ui` to run tests and watch them execute
2. Use `npm run test:e2e:debug` to step through with inspector
3. Check the HTML report for detailed results and traces
4. Look at screenshots/videos in `test-results/` on failure
