# c8s Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-12

## Active Technologies
- (001-build-a-continuous)
- Go 1.25.0 (002-i-want-to)
- N/A (cluster state managed by chosen local K8s distribution) (002-i-want-to)
- Go 1.24.0 (backend API server), HTML5/CSS3/JavaScript (frontend with HTMX) (004-create-a-front)
- Uses existing C8S infrastructure (Kubernetes, S3-compatible object storage for logs/artifacts) (004-create-a-front)

## Project Structure
```
src/
tests/
```

## Commands
# Add commands for 

## Code Style
Follow standard conventions

## Git Workflow
- **Commit after each significant feature** or bug fix
- **Commit format:** `[Txxx] Feature description` (e.g., `[T084] Implement Dashboard Pipeline Runs`)
- **Always include footer:**
  ```
  🤖 Generated with Claude Code

  Co-Authored-By: Claude <noreply@anthropic.com>
  ```
- **Use git status and git diff --cached --stat before committing**
- Commit logical units of work, not partial implementations

## Development Checklist
Before starting work on a feature:
1. Check current `git status`
2. Create a todo list for the task
3. Work on the feature in small, logical steps
4. Test changes thoroughly
5. Commit with meaningful message when feature is complete

After finishing work:
1. Verify `git log` shows commits for completed work
2. Review changes with `git diff HEAD~N` (where N is number of commits)

## Recent Changes
- 004-create-a-front: Added Go 1.24.0 (backend API server), HTML5/CSS3/JavaScript (frontend with HTMX)
- 002-i-want-to: Added Go 1.25.0
- 001-build-a-continuous: Added

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
