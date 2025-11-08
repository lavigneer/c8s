# C8S Improvement Plan - Quick Reference Card

**Print this for your desk while working on the plan**

---

## 3-Phase Overview

```
PHASE 1: CRITICAL FIXES (Weeks 1-2)
├─ Authentication (S1) ............... 6 tasks, 12 hours
├─ Authorization (S2) ................ 6 tasks, 16 hours
├─ Error Handling (C1) ............... 6 tasks, 5.5 hours
├─ Webhook Security (S3) ............ 4 tasks, 5 hours
├─ Request Limits (S4) .............. 2 tasks, 2 hours
├─ Info Disclosure (S5) ............ 3 tasks, 3.5 hours
├─ CORS Fix (S6) .................... 2 tasks, 2.5 hours
├─ Getting Started Docs (D1) ........ 5 tasks, 8 hours
├─ Troubleshooting Docs (D2) ........ 5 tasks, 10 hours
├─ README Update (D3) ............... 3 tasks, 2 hours
└─ Handler Tests (T1) ............... 8 tasks, 22 hours
   TOTAL: 50 tasks, 88.5 hours

PHASE 2: HIGH PRIORITY (Weeks 3-4)
├─ Configuration Docs (D4) .......... 4 tasks, 6.5 hours
├─ Operator Guide (D5) .............. 6 tasks, 11 hours
├─ Dashboard Guide (D6) ............ 7 tasks, 10 hours
├─ Pipeline Syntax (D7) ............ 5 tasks, 10.5 hours
├─ Dashboard Consolidation (D8) ..... 3 tasks, 4.5 hours
├─ Contributing Guide (D9) .......... 2 tasks, 2.5 hours
├─ Remove Test Data (Q1) ............ 4 tasks, 7.5 hours
└─ Refactor Patterns (Q2) ........... 4 tasks, 8.5 hours
   TOTAL: 35 tasks, 61 hours

PHASE 3: ENHANCEMENT (Weeks 5-6)
├─ Architecture Docs (D10) .......... 5 tasks, 9 hours
├─ Security Docs (D11) ............. 6 tasks, 7.5 hours
├─ Testing Guide (D12) ............. 6 tasks, 10 hours
├─ CLI Reference (D13) ............. 4 tasks, 7.5 hours
├─ Webhook Guide (D14) ............. 6 tasks, 9 hours
└─ API Schema (D15) ................ 3 tasks, 7.5 hours
   TOTAL: 30 tasks, 50.5 hours

GRAND TOTAL: 115 tasks, 200 hours (6 weeks @ 33 hrs/week)
```

---

## The 7 Critical Issues to Fix FIRST

| # | Issue | Where | Impact | Time |
|---|-------|-------|--------|------|
| 1 | Auth not functional | `auth_middleware.go` | Any token accepted | 12h (S1) |
| 2 | Authz not enforced | `artifacts.go`, `logs.go` | Unauthorized access | 16h (S2) |
| 3 | Silent SSE failures | `logs.go`, `pipeline_sse.go` | Lost updates | 5.5h (C1) |
| 4 | Unsigned webhooks | `github.go` | Unauthorized triggers | 5h (S3) |
| 5 | No getting started | NEW: `docs/` | User blocked | 8h (D1) |
| 6 | No troubleshooting | NEW: `docs/` | High support load | 10h (D2) |
| 7 | No handler tests | NEW: `tests/unit/` | 1600+ LOC untested | 22h (T1) |

---

## Critical Path (Must Do In Order)

```
START
  ↓
S1.x: Authentication (12h)
  ↓ MUST COMPLETE BEFORE
S2.x: Authorization (16h)
  ↓ PARALLEL (can do same time)
C1.x: Error Handling (5.5h) + D1.x: Getting Started (8h) + D2.x: Troubleshooting (10h)
  ↓ MUST COMPLETE BEFORE
T1.x: Handler Tests (22h)
  ↓
PHASE 1 COMPLETE → Proceed to Phase 2
```

---

## Files to Edit in Phase 1

### Security Fixes
```
cmd/api-server/handlers/
├─ auth_middleware.go          → Implement JWT validation (S1)
├─ authz_middleware.go         → Fix error disclosure (S5)
├─ artifacts.go                → Add auth checks (S2)
├─ logs.go                     → Fix SSE errors (C1)
├─ pipeline_sse.go            → Fix SSE errors (C1)
├─ projects.go                 → Fix JSON errors (C1)
├─ export.go                   → Fix CSV errors (C1)
└─ main.go                     → Add request limits (S4)

cmd/api-server/middleware/
└─ security_headers.go         → Fix CORS (S6)

pkg/webhook/
├─ github.go                   → Require signatures (S3)
├─ gitlab.go                   → Require signatures (S3)
└─ bitbucket.go                → Require signatures (S3)
```

### Documentation to Create
```
docs/
├─ GETTING_STARTED.md          → NEW (D1)
├─ TROUBLESHOOTING.md          → NEW (D2)
└─ CONFIGURATION.md            → NEW (Phase 2)

Root/
├─ README.md                   → Update status (D3)
└─ CONTRIBUTING.md             → NEW (Phase 2)
```

### Tests to Add
```
tests/unit/
├─ handlers/                   → NEW DIRECTORY
│  ├─ auth_middleware_test.go
│  ├─ authz_middleware_test.go
│  ├─ projects_test.go
│  ├─ logs_test.go
│  ├─ artifacts_test.go
│  ├─ pipeline_runs_test.go
│  ├─ export_test.go
│  └─ error_handling_test.go
├─ middleware/
│  ├─ size_limit_test.go       → NEW
│  └─ cors_test.go             → NEW
└─ webhook/
   └─ signature_test.go        → NEW
```

---

## How to Update PROGRESS_TRACKER.md

### Starting a task
```
- [ ] **S1.2** - Implement JWT (3h) - ⏳ Pending
```
↓ Change to:
```
- [ ] **S1.2** - Implement JWT (3h) - 🔄 In Progress
```

### Completing a task
```
- [ ] **S1.2** - Implement JWT (3h) - 🔄 In Progress
```
↓ Change to:
```
- [x] **S1.2** - Implement JWT (3h) - ✅ Complete
  - Completed: 2025-11-05
  - Actual: 4h (overestimate by 1h)
  - Notes: Added RS256 support
  - PR: #123
```

---

## Effort Estimates by Role

### For 1 Developer (6 weeks)
- Week 1: S1 + S2 start (28h)
- Week 2: S2 finish + C1 + D1 + D2 (35h)
- Week 3: T1 + S3 + S4 + S5 + S6 (30h)
- Week 4: Phase 2 starts (40h)
- Week 5-6: Phase 2 continuation (40h)

### For 2 Developers (4 weeks)
- Team A: Auth/Authz/Security (S1, S2, S3, S4, S5, S6) = 30h
- Team B: Error Handling/Tests/Docs (C1, D1, D2, D3, T1) = 50h
- Total: 80h each over 4 weeks

### For 3 Developers (3 weeks)
- Team A (Security Lead): S1, S2, S3, S4, S5, S6 = 30h
- Team B (Backend Lead): C1, T1 = 27.5h
- Team C (Tech Writer): D1, D2, D3 = 20h
- Total: 77.5h weeks 1-3, then Phase 2

---

## Quick Checklist - First Day

- [ ] Read SYSTEMATIC_REVIEW.md (20 min)
- [ ] Read IMPROVEMENT_PLAN.md sections for your tasks (20 min)
- [ ] Understand the critical path (10 min)
- [ ] Set up test environment
- [ ] Run existing tests: `make test`
- [ ] Run E2E tests: `npm run test:e2e`
- [ ] Start task S1.1: Analyze JWT requirements

---

## Quick Checklist - End of Phase 1

- [ ] All 50 Phase 1 tasks completed
- [ ] 88.5 hours logged
- [ ] 50+ new unit tests added
- [ ] GETTING_STARTED.md published
- [ ] TROUBLESHOOTING.md published
- [ ] Authentication working
- [ ] Authorization working
- [ ] All E2E tests passing
- [ ] No critical issues open
- [ ] Ready for Phase 2

---

## Key Shortcuts

**Run all tests**:
```bash
make test              # Go unit tests
npm run test:e2e       # E2E tests
make test-coverage     # Coverage report
```

**Check code quality**:
```bash
make lint              # golangci-lint
make build             # Verify builds
```

**Development**:
```bash
make install-crds      # Install K8s CRDs
make run-controller    # Run controller
make tilt-up          # Start Tilt dev env
```

**Review changes**:
```bash
git status            # What changed?
git diff              # Show diffs
git diff --cached     # Show staged changes
```

---

## Contact Points in Code

### Where to Add Authentication Checks
- `cmd/api-server/handlers/auth_middleware.go:54-94`
- Look for: `user := &User{ID: "user-id",...}`

### Where to Add Authorization Checks
- `cmd/api-server/handlers/artifacts.go:26, 156, 172` (look for `TODO:`)
- `cmd/api-server/handlers/logs.go` (no access checks)
- `cmd/api-server/handlers/pipeline_runs.go` (no access checks)

### Where to Fix Error Handling
- `cmd/api-server/handlers/logs.go:44, 73, 80, 93`
- `cmd/api-server/handlers/pipeline_sse.go:72`
- `cmd/api-server/handlers/projects.go:105`
- `cmd/api-server/handlers/export.go:83, 147`

### Where to Require Webhook Signatures
- `pkg/webhook/github.go:140-149` (check for `if signature != ""`)

---

## Useful Commands During Development

```bash
# Run tests for your changes
go test ./cmd/api-server/handlers/... -v

# Run single test
go test -run TestAuthenticationFlow ./cmd/api-server/handlers/...

# Run E2E tests
npm run test:e2e:debug

# Check coverage
go test ./... -cover

# Find code locations
grep -r "TODO:" cmd/api-server/handlers/

# Verify no build errors
go build ./cmd/api-server

# Check for unused code
golangci-lint run
```

---

## Red Flags & Dangers

🚩 **Danger**: Modifying auth without tests → **Always test first**
🚩 **Danger**: Changing error handling → **Run full test suite**
🚩 **Danger**: Touching handlers → **Run E2E tests afterward**
🚩 **Danger**: Webhook changes → **Verify signature validation**
🚩 **Danger**: Removing test data → **Update E2E tests**

---

## Success is When...

✅ Authentication validates JWT properly
✅ Authorization blocks unauthorized access
✅ All write operations checked for errors
✅ New handlers have unit tests
✅ Documentation guides new users
✅ Production readiness score 8/10+
✅ E2E test suite passes 100%
✅ No open critical issues

---

**Printed**: 2025-11-02
**Sections**: 13
**Total Words**: ~2000
**Est. Read Time**: 10 minutes
