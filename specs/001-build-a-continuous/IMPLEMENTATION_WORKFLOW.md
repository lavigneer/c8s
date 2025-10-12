# Implementation Workflow: Per-Task Git Commits

**Feature**: 001-build-a-continuous
**Date**: 2025-10-12

## Overview

This document defines the git commit workflow for implementing the 89 tasks in tasks.md. Each completed task will receive a dedicated git commit for clean history and easy rollback.

---

## Commit Strategy

### One Commit Per Task

**Rule**: Each task (T001-T089) gets exactly ONE commit when completed.

**Benefits**:
- Clean, granular git history
- Easy to identify when specific functionality was added
- Easy to rollback individual tasks if needed
- Clear progress tracking via git log
- Supports code review at task level

---

## Commit Message Format

### Standard Format

```
[T###] Short task description

Detailed description:
- What was implemented
- Files created/modified
- Any important decisions made

User Story: [US#] (if applicable)
Phase: # - Phase Name
```

### Examples

**Example 1: Setup Task**
```
[T001] Initialize Go module and project structure

Created initial directory structure and Go module:
- Initialized go.mod with Go 1.25
- Created cmd/, pkg/, tests/, config/, web/, deploy/ directories
- Established standard Go project layout per plan.md

Phase: 1 - Setup & Project Initialization
```

**Example 2: Implementation Task**
```
[T025] Implement Job creation from PipelineStep

Created job_manager.go with Job generation logic:
- CreateJobForStep() generates K8s Job from PipelineStep
- Init container for git clone
- Main container with step.image and commands
- Resource requests/limits from step.resources
- Owner reference for garbage collection
- TTL configured for automatic cleanup

User Story: US1 - Basic Pipeline Execution
Phase: 3 - User Story 1
Files: pkg/controller/job_manager.go
```

**Example 3: Test Task**
```
[T042] Write unit tests for DAG scheduler

Added comprehensive test coverage for scheduler:
- Test cases: empty steps, single step, linear dependencies
- Test cases: parallel steps, complex DAG
- Test cases: circular dependency detection, invalid references
- Uses testify for assertions

User Story: US1 - Basic Pipeline Execution
Phase: 3 - User Story 1
Files: tests/unit/scheduler_test.go
```

---

## Commit Workflow

### For Each Task

1. **Start Task**: Reference tasks.md for task details (T###)
2. **Implement**: Write code, tests, docs per task description
3. **Verify**: Ensure task acceptance criteria met
4. **Stage Changes**: `git add <files>`
5. **Commit**: Use format above with task number
6. **Move to Next**: Proceed to next task in dependency order

### Commands

```bash
# After completing task T001
git add go.mod cmd/ pkg/ tests/ config/ web/ deploy/
git commit -m "[T001] Initialize Go module and project structure

Created initial directory structure and Go module:
- Initialized go.mod with Go 1.25
- Created cmd/, pkg/, tests/, config/, web/, deploy/ directories
- Established standard Go project layout per plan.md

Phase: 1 - Setup & Project Initialization"

# After completing task T025
git add pkg/controller/job_manager.go
git commit -m "[T025] Implement Job creation from PipelineStep

Created job_manager.go with Job generation logic:
- CreateJobForStep() generates K8s Job from PipelineStep
- Init container for git clone
- Main container with step.image and commands
- Resource requests/limits from step.resources
- Owner reference for garbage collection
- TTL configured for automatic cleanup

User Story: US1 - Basic Pipeline Execution
Phase: 3 - User Story 1
Files: pkg/controller/job_manager.go"
```

---

## Checkpoint Commits

At the end of each phase, create a checkpoint commit summarizing progress:

```
[CHECKPOINT] Phase # - Phase Name Complete

Summary:
- # tasks completed (T###-T###)
- Key deliverables: ...
- User Story # deliverable achieved (if applicable)

Next: Phase # - Phase Name
```

**Example**:
```
[CHECKPOINT] Phase 1 - Setup & Project Initialization Complete

Summary:
- 12 tasks completed (T001-T012)
- Project structure established
- Development tooling configured
- CI/CD pipeline ready

Next: Phase 2 - Foundational Infrastructure
```

---

## Git History Structure

Expected git log structure:

```
* [CHECKPOINT] Phase 7 - Parallel Execution Complete
* [T089] Write integration test for matrix execution
* [T088] Write integration test for parallel execution
* [T087] Update dashboard to show matrix executions
...
* [CHECKPOINT] Phase 3 - User Story 1 Complete (MVP)
* [T044] Write integration test for basic pipeline execution
* [T043] Write unit tests for pipeline parser
...
* [CHECKPOINT] Phase 2 - Foundational Infrastructure Complete
* [T022] Setup RBAC manifests for controller
* [T021] Implement pipeline YAML parser
...
* [CHECKPOINT] Phase 1 - Setup Complete
* [T012] Document development workflow
* [T011] Setup envtest for integration testing
...
* [T001] Initialize Go module and project structure
* Initial commit (planning artifacts)
```

---

## Handling Multiple Files

When a task modifies many files, list the key ones in commit message:

```
[T016] Generate CRD manifests and DeepCopy methods

Generated CRD YAML files and DeepCopy implementations:
- Created pkg/apis/v1alpha1/groupversion_info.go
- Generated config/crd/bases/c8s.dev_pipelineconfigs.yaml
- Generated config/crd/bases/c8s.dev_pipelineruns.yaml
- Generated config/crd/bases/c8s.dev_repositoryconnections.yaml
- Generated pkg/apis/v1alpha1/zz_generated.deepcopy.go
- Verified CRDs match data-model.md schemas

Phase: 2 - Foundational Infrastructure
Files: config/crd/bases/*, pkg/apis/v1alpha1/zz_generated.deepcopy.go
```

---

## Handling Parallel Tasks

When multiple tasks marked [P] are done in parallel:

1. Each developer works on separate task branches (optional)
2. Each task gets individual commit
3. Merge/rebase to maintain task order numbering
4. Keep commits atomic (one task = one commit)

**Example Timeline**:
```
Developer A: T023 → commit → T024 → commit
Developer B: T029 → commit → T030 → commit
Developer C: T035 → commit → T036 → commit

All merge to main, maintaining T### ordering
```

---

## Rollback Strategy

### Revert Single Task
```bash
# Revert task T025 if bugs found
git log --grep="T025"  # Find commit hash
git revert <commit-hash>
```

### Rollback to Checkpoint
```bash
# Rollback entire Phase 3
git log --grep="CHECKPOINT.*Phase 2"  # Find Phase 2 checkpoint
git reset --hard <checkpoint-commit-hash>
```

---

## Progress Tracking

### Check Completed Tasks
```bash
# List all completed tasks
git log --oneline --grep="^\[T[0-9]"

# Count completed tasks
git log --oneline --grep="^\[T[0-9]" | wc -l

# Check which phase we're in
git log --oneline --grep="CHECKPOINT" | head -1
```

### Verify All Files Committed
```bash
# Should be clean after each task
git status
```

---

## Automated Workflow (Future Enhancement)

Potential automation with git hooks or task runner:

```bash
# Hypothetical command
./run-task.sh T001

# Would:
# 1. Load task T001 description from tasks.md
# 2. Present task details
# 3. Wait for implementation
# 4. Run tests if task includes tests
# 5. Auto-generate commit message from task
# 6. Prompt for confirmation
# 7. Commit with standard format
```

---

## Best Practices

### DO:
✅ Commit after each task completes
✅ Include task number [T###] in commit message
✅ Add brief description of what was implemented
✅ List key files created/modified
✅ Tag user story if applicable
✅ Create checkpoint commits at phase boundaries
✅ Keep commits atomic (one task, one commit)

### DON'T:
❌ Combine multiple tasks in one commit
❌ Commit partial task implementation (wait until task complete)
❌ Skip task numbers or reorder
❌ Forget to verify tests pass before committing
❌ Commit generated files without documenting generation method
❌ Use vague commit messages

---

## Implementation Start Command

When ready to begin implementation:

```bash
# Ensure we're on feature branch
git checkout 001-build-a-continuous

# Confirm planning artifacts committed
git log --oneline -5

# Begin Phase 1, Task T001
# After completing T001:
git add <files>
git commit -m "[T001] Initialize Go module and project structure
...
"

# Continue with T002, T003, etc.
```

---

## Summary

**89 tasks** = **89 commits** + **7 checkpoint commits** = **96 total commits** for complete implementation

This provides:
- Granular history
- Easy rollback
- Clear progress tracking
- Individual task review capability
- Clean git log for project history

**Ready to begin implementation with clean, tracked commits!** 🚀
