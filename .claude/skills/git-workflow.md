# Git Workflow and Committing

Maintain a clean, trackable history by committing frequently as you progress on features and specs.

## Commit Format
Use the established commit format for all commits:
```
[Txxx] Feature description

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
```

Example:
```
[T042] Add validation to PipelineConfig CRD

🤖 Generated with Claude Code

Co-Authored-By: Claude <noreply@anthropic.com>
```

## When to Commit
- **After completing a feature** - Each logical unit of work gets its own commit
- **After fixing a bug** - Document the fix with a clear message
- **After updating specs** - Commit spec.md, plan.md, or tasks.md changes
- **After code generation** - Always commit generated files (manifests, deepcopy methods)
- **Not during partial implementations** - Don't commit work-in-progress unless documented

## Pre-Commit Checklist
Before committing, verify your changes:
```bash
# In devbox shell
git status              # See all untracked files
git diff --cached --stat # Review what's staged
git log --oneline -5    # Check recent commits for style consistency
```

## Commit Workflow
1. **Make changes** to code, specs, or tests
2. **Stage changes**: `git add <files>` or `git add .`
3. **Review staged changes**: `git diff --cached`
4. **Commit with message**: Use the format above
5. **Verify**: `git log -1` to confirm commit was created

## Spec and Plan Updates
When working with specs and tasks:
- Commit spec.md changes separately: `[Txxx] Update spec: ...`
- Commit tasks.md updates when significant: `[Txxx] Update task list: ...`
- Commit plan.md refinements: `[Txxx] Refine implementation plan: ...`

## Generated Files
Always commit generated files:
- `config/crd/bases/*.yaml` - From `make manifests`
- `*_deepcopy_generated.go` - From `make generate`
- Any output from code generation tools

These changes are part of the feature and should be committed together with their triggering changes.

## Viewing History
```bash
# See recent commits
git log --oneline -10

# See commits for a file
git log --oneline pkg/apis/v1alpha1/pipelineconfig_types.go

# See what changed in last commit
git show

# Compare with main branch
git diff main
```

## Best Practices
- **One feature per commit** - Don't mix multiple features
- **Clear messages** - Describe what changed and why
- **Logical units** - Commits should be reviewable
- **No secrets** - Never commit .env, credentials, or sensitive files
- **Test before committing** - Run `make test && make lint` first
- **Generated files included** - CRDs and generated code are part of features

## Avoiding Common Mistakes
- ❌ Don't manually edit `*_generated.go` files - they'll be overwritten
- ❌ Don't commit work-in-progress without clear labeling
- ❌ Don't mix formatting changes with feature changes in one commit
- ❌ Don't forget to run `make generate && make manifests` before committing CRD changes
- ✅ Do run `make fmt` before committing
- ✅ Do run `make lint && make test` to verify before committing
