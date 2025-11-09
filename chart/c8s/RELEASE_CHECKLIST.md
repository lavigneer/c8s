# C8S Helm Chart Release Checklist

This checklist ensures quality and consistency when releasing a new version of the C8S Helm chart.

## Pre-Release (1-2 weeks before)

- [ ] Create a release branch: `git checkout -b release/v0.X.Y`
- [ ] Update Chart.yaml with new version
- [ ] Update appVersion in Chart.yaml if app version changed
- [ ] Update CHANGELOG.md with release notes
- [ ] Document any breaking changes
- [ ] Update README.md if needed

## Testing (Before Release)

### Lint & Validation
- [ ] Run `helm lint ./chart/c8s` - should pass with no warnings
- [ ] Run `helm template c8s ./chart/c8s` - should render without errors
- [ ] Template validation: `helm template c8s ./chart/c8s -f values-dev.yaml`
- [ ] Template validation: `helm template c8s ./chart/c8s -f values-staging.yaml`
- [ ] Template validation: `helm template c8s ./chart/c8s -f values-prod.yaml`

### Functional Testing
- [ ] Install with dev values: `helm install c8s ./chart/c8s -f values-dev.yaml`
- [ ] Verify all pods reach Ready state within 5 minutes
- [ ] Test health check hook output
- [ ] Test dashboard accessibility

### Upgrade Testing
- [ ] Deploy previous version
- [ ] Upgrade to new version
- [ ] Verify rolling update completes successfully
- [ ] Verify all components ready after upgrade
- [ ] Test values preservation during upgrade

### Downgrade Testing
- [ ] Deploy new version
- [ ] Rollback to previous version: `helm rollback c8s 1`
- [ ] Verify rollback completes successfully
- [ ] Verify previous configuration restored

### Clean Uninstall
- [ ] Deploy release
- [ ] Uninstall: `helm uninstall c8s`
- [ ] Verify all C8S resources removed
- [ ] Verify configmaps/secrets cleaned up
- [ ] Verify PVCs preserved (intentional)

### Environment Testing
- [ ] Test on k3s (local)
- [ ] Test on kind (local)
- [ ] Test on Docker Desktop Kubernetes
- [ ] Verify on at least one cloud distribution (EKS, GKE, etc.)

### Values Testing
- [ ] Test all component images can be overridden
- [ ] Test replica count overrides
- [ ] Test resource limit overrides
- [ ] Test storage configuration overrides
- [ ] Test log level override

## Documentation

- [ ] README.md is up-to-date
- [ ] CHANGELOG.md has release notes
- [ ] All new parameters documented
- [ ] Examples updated if needed
- [ ] Troubleshooting guide updated
- [ ] Migration guide added (if breaking changes)

## Code Quality

- [ ] No hardcoded values (except defaults in values.yaml)
- [ ] All templates properly indented
- [ ] All helper functions documented
- [ ] RBAC permissions minimal and correct
- [ ] CRD definitions current
- [ ] Security best practices followed

## Release

### Preparation
- [ ] All tests passing
- [ ] All documentation updated
- [ ] Git branch clean (no uncommitted changes)
- [ ] Git history clean

### Version Bump
- [ ] Update Chart.yaml version: `version: 0.X.Y`
- [ ] Update Chart.yaml appVersion if changed
- [ ] Tag git commit: `git tag v0.X.Y`

### Publishing
- [ ] Push tag: `git push origin v0.X.Y`
- [ ] Create GitHub Release with release notes
- [ ] Upload chart tarball to release
- [ ] Package chart: `helm package ./chart/c8s`
- [ ] Create artifact repository (if publishing to repo)

## Post-Release

- [ ] Merge release branch back to main: `git merge release/v0.X.Y`
- [ ] Update main branch version to next dev version
- [ ] Close any associated GitHub issues
- [ ] Update documentation on project site
- [ ] Announce release on channels

## Rollback Plan

If issues are found:

1. **Create hotfix branch**: `git checkout -b hotfix/v0.X.Y+1`
2. **Fix issues**
3. **Test thoroughly**
4. **Release hotfix**: Follow release steps above
5. **Communicate**: Notify users of issue and fix

## Version Numbering

Use [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes
  - Example: New required configuration parameter
  - Requires migration guide

- **MINOR** (0.X.0): Backwards-compatible features
  - Example: New optional parameter, new component
  - No migration needed

- **PATCH** (0.0.X): Bug fixes and minor changes
  - Example: Security patch, template fix
  - No migration needed

## Release Notes Template

```markdown
## Version 0.X.Y - [DATE]

### New Features
- Feature description
- Feature description

### Improvements
- Improvement description
- Improvement description

### Bug Fixes
- Bug fix description
- Bug fix description

### Breaking Changes
- Breaking change description with migration notes
- Breaking change description with migration notes

### Upgrade Instructions
```bash
helm upgrade c8s ./chart/c8s -f values-prod.yaml
```

### Contributors
- Contributor names
```

## Quick Reference

```bash
# One-command release flow:
helm lint ./chart/c8s && \
helm template c8s ./chart/c8s -f values-dev.yaml && \
git add . && \
git commit -m "[Release] v0.X.Y - Release notes" && \
git tag v0.X.Y && \
git push && git push --tags
```

## Common Issues

### Chart linting fails
- Check indentation (must be spaces, not tabs)
- Validate YAML syntax
- Check required fields in Chart.yaml

### Template rendering fails
- Check Helm syntax: `{{ }}`, `{{- -}}`
- Verify all variables exist in values
- Check for missing closing tags

### Pods fail to start
- Check image references in template
- Verify RBAC permissions
- Check resource requests/limits

### Test failures
- Run individual test components
- Check cluster resources
- Review pod events: `kubectl describe pod`
