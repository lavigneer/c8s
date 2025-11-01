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
- 004-create-a-front: **COMPLETED** - Full web dashboard implementation (all 5 user stories)
  - Pipeline history and status visualization (US1)
  - Real-time log streaming with SSE (US2)
  - Advanced filtering with URL state (US3)
  - Project and webhook management UI (US4)
  - Artifact viewing and download (US5)
  - Keyboard shortcuts (FR-013)
  - Authentication and authorization (FR-010)
  - Mobile-responsive design (FR-012)
- 002-i-want-to: Added Go 1.25.0
- 001-build-a-continuous: Added

## Dashboard Implementation Details
**Status**: COMPLETE AND TESTED

### Implementation Summary
- All 5 user stories fully implemented
- 13 feature requirements fulfilled
- 10 commits totaling 600+ lines of code
- Real-time log streaming via SSE
- Advanced filtering with URL persistence
- Keyboard shortcut support
- Artifact management with preview capability
- Responsive design for mobile devices

### Key Files Modified
- `cmd/api-server/handlers/` - 7 new/updated handler files
- `cmd/api-server/templates/` - 10+ template files
- `cmd/api-server/static/js/` - Keyboard shortcuts implementation
- `pkg/dashboard/` - DTOs and logging infrastructure
- `DASHBOARD_IMPLEMENTATION.md` - Complete reference documentation

### Testing Checklist
✅ Authentication flow (login/logout)
✅ Pipeline listing and filtering
✅ Log streaming and viewer
✅ Project creation and management
✅ Artifact download and preview
✅ Keyboard shortcuts functionality
✅ URL state persistence
✅ Responsive design

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
