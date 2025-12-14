## Description

<!-- Provide a clear and concise description of the changes -->

## Type of Change

<!-- Check the relevant option(s) -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring
- [ ] Test updates
- [ ] CI/CD changes
- [ ] Dependencies update

## Related Issues

<!-- Link related issues using keywords: Fixes #123, Closes #456, Related to #789 -->

Fixes #

## Changes Made

<!-- List the specific changes made in this PR -->

-
-
-

## Testing

<!-- Describe the tests you ran to verify your changes -->

### Test Coverage

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] E2E tests added/updated
- [ ] Manual testing performed

### Test Commands

```bash
# Commands used to test the changes
make test
make test-integration
npm run test:e2e
```

### Test Results

<!-- Paste relevant test output or link to CI results -->

```
# Test output here
```

## Screenshots/Videos

<!-- If applicable, add screenshots or videos demonstrating the changes -->

## Checklist

### Code Quality

- [ ] My code follows the style guidelines of this project (run `make lint`)
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings or errors
- [ ] I have removed any debug logs or commented-out code

### Documentation

- [ ] I have updated relevant documentation (docs/, README.md, CLAUDE.md)
- [ ] I have updated code comments where necessary
- [ ] I have added/updated API documentation if applicable
- [ ] I have updated the architecture documentation if applicable

### Testing

- [ ] I have added tests that prove my fix is effective or my feature works
- [ ] New and existing unit tests pass locally (`make test`)
- [ ] New and existing integration tests pass locally (`make test-integration`)
- [ ] E2E tests pass if frontend was changed (`npm run test:e2e`)

### Deployment

- [ ] I have updated CRD manifests if API changes were made (`make manifests`)
- [ ] I have updated Helm chart if deployment configuration changed
- [ ] I have verified the changes work with Tilt (`tilt up`)
- [ ] I have considered backward compatibility

### Git

- [ ] My commits follow the project's commit message conventions (see CLAUDE.md)
- [ ] I have rebased my branch on the latest main
- [ ] I have resolved all merge conflicts

## Breaking Changes

<!-- If this is a breaking change, describe the impact and migration path -->

**Impact**:
-

**Migration Guide**:
-

## Additional Notes

<!-- Any additional information for reviewers -->

## Reviewer Notes

<!-- For reviewers: areas to focus on, specific concerns, or questions -->

---

**By submitting this PR, I confirm that**:
- [ ] My contribution is made under the terms of the Apache 2.0 license
- [ ] I have read and agree to the [Contributing Guidelines](./docs/CONTRIBUTING.md)
- [ ] I have tested my changes thoroughly
