# S2.1: Authorization System Design

**Task**: Design authorization service interface and data models for role-based access control
**Status**: ✅ Complete
**Date**: 2025-11-02
**Effort**: 1 hour (S2.1 planning)

---

## Overview

C8S requires a robust authorization system to enforce role-based access control (RBAC) across all handlers. This design builds on the JWT authentication foundation (S1.1-S1.6) to implement fine-grained access control.

## Architecture

### Current State

The codebase has a **partial authorization infrastructure**:

```
├── ✅ Role Definitions (3 roles: admin, editor, viewer)
├── ✅ ProjectAccessService Interface (defined but stubbed)
├── ✅ Role Hierarchy (admin > editor > viewer)
├── ✅ Middleware Stubs (ProjectAccessMiddleware, RoleBasedContextMiddleware)
├── ✅ User Struct with Roles field
├── ❌ K8s RoleBinding Queries (NOT implemented)
├── ❌ Per-Handler Authorization Checks (NOT implemented)
├── ❌ Field-Level Access Control (NOT implemented)
└── ❌ Authorization Tests (NOT implemented)
```

### Design Goals

1. **Layered Authorization**: Authentication → Handler Checks → Field-Level Control
2. **Kubernetes Native**: Use K8s RoleBindings for single source of truth
3. **Flexible**: Support multiple role models (project, namespace, global)
4. **Auditable**: Log all authorization decisions
5. **Testable**: Full unit test coverage for all authorization logic

---

## Core Components

### 1. Role Definitions

**File**: `pkg/dashboard/project_access.go` (existing, reuse)

```go
type Role string

const (
    RoleAdmin  Role = "admin"   // Full access: create, read, update, delete
    RoleEditor Role = "editor"  // Write access: create, read, update
    RoleViewer Role = "viewer"  // Read-only access
)

// RoleLevel returns numeric level for hierarchy comparison
func (r Role) Level() int {
    switch r {
    case RoleAdmin:
        return 3
    case RoleEditor:
        return 2
    case RoleViewer:
        return 1
    default:
        return 0 // Unknown/denied
    }
}

// HasPermission checks if role allows action
func (r Role) HasPermission(action Action) bool {
    // action: Read, Write, Delete, Admin
    // Hierarchy: admin > editor > viewer
}
```

### 2. ProjectAccessService Interface

**File**: `pkg/dashboard/project_access.go` (existing, expand)

Already defined interface with 4 methods:

```go
type ProjectAccessService interface {
    // UserHasProjectAccess checks basic project access
    UserHasProjectAccess(ctx, userID, projectID) (bool, error)

    // GetUserRoleForProject returns user's role (admin/editor/viewer)
    GetUserRoleForProject(ctx, userID, projectID) (Role, error)

    // ListUserProjects returns all accessible projects
    ListUserProjects(ctx, userID) ([]ProjectDTO, error)

    // HasProjectRole checks if user meets minimum role requirement
    HasProjectRole(ctx, userID, projectID, requiredRole) (bool, error)
}
```

**Implementation Strategy**:

1. Query K8s RoleBindings in project namespace
2. Match user ID against RoleBinding subjects
3. Map bound ClusterRole to C8S role (admin/editor/viewer)
4. Cache results for 5-minute TTL to reduce API calls

### 3. Authorization Service Implementation

**New File**: `pkg/dashboard/authorization_service.go` (NEW)

```go
type AuthorizationService struct {
    projectAccess ProjectAccessService
    cache         map[string]CachedAuthResult
    cacheTTL      time.Duration
}

// CheckAccess verifies user can perform action on resource
func (s *AuthorizationService) CheckAccess(ctx, userID, resourceType, resourceID, action) (bool, error) {
    // Route to appropriate checker based on resourceType
    // Types: project, pipeline, step, artifact, log
}

// CheckProjectAccess verifies project-level access
func (s *AuthorizationService) CheckProjectAccess(ctx, userID, projectID, action) (bool, error) {
    role, err := s.projectAccess.GetUserRoleForProject(ctx, userID, projectID)
    if err != nil {
        return false, err
    }
    return role.HasPermission(action), nil
}

// CheckPipelineAccess verifies pipeline access (inherit from project)
func (s *AuthorizationService) CheckPipelineAccess(ctx, userID, projectID, pipelineID, action) (bool, error) {
    // Pipelines inherit project-level access
    return s.CheckProjectAccess(ctx, userID, projectID, action)
}

// CheckArtifactAccess verifies artifact access (inherit from pipeline)
func (s *AuthorizationService) CheckArtifactAccess(ctx, userID, artifactID, action) (bool, error) {
    // Look up artifact → pipeline → project
    // Check access at highest level
}
```

### 4. Action/Permission Model

**New Type**: `pkg/dashboard/permissions.go` (NEW)

```go
type Action string

const (
    ActionRead   Action = "read"   // View resource
    ActionWrite  Action = "write"  // Create/update resource
    ActionDelete Action = "delete" // Delete resource
    ActionAdmin  Action = "admin"  // Manage access/settings
)

// PermissionMatrix defines what roles can do
var PermissionMatrix = map[Role][]Action{
    RoleViewer: {ActionRead},
    RoleEditor: {ActionRead, ActionWrite},
    RoleAdmin:  {ActionRead, ActionWrite, ActionDelete, ActionAdmin},
}

// CanPerform checks if role can perform action
func CanPerform(role Role, action Action) bool {
    for _, allowed := range PermissionMatrix[role] {
        if allowed == action {
            return true
        }
    }
    return false
}
```

### 5. Authorization Result Types

**New Types**: `pkg/dashboard/authz_result.go` (NEW)

```go
type AuthzResult struct {
    Allowed     bool        // Whether action is allowed
    Reason      string      // Why allowed/denied
    Role        Role        // User's role in resource
    Resource    string      // Resource that was checked
    Action      Action      // Action that was checked
    Timestamp   time.Time   // When check occurred
    UserID      string      // User who was checked
}

type CachedAuthResult struct {
    Result    AuthzResult
    ExpiresAt time.Time
}
```

### 6. Handler Integration Pattern

**In Each Handler** (e.g., projects.go, artifacts.go):

```go
func DeleteArtifactHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    artifactID := r.PathValue("artifactId")

    // Check authorization
    authz := AuthorizationFromContext(r.Context())
    allowed, err := authz.CheckArtifactAccess(r.Context(), user.ID, artifactID, ActionDelete)
    if err != nil {
        http.Error(w, "Error checking access", http.StatusInternalServerError)
        return
    }
    if !allowed {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Proceed with delete
}
```

### 7. Middleware Integration

**Enhanced Middleware Chain**:

```
Request
  ↓
[AuthMiddleware] - Validates JWT, extracts user
  ↓
[AuthorizationMiddleware] - Creates AuthorizationService, attaches to context
  ↓
[ProjectAccessMiddleware] - Checks project-level access (optional per route)
  ↓
[RoleBasedContextMiddleware] - Attaches role to context for UI
  ↓
Handler - Uses GetUserFromContext() and AuthorizationFromContext()
```

---

## Implementation Details

### Phase 1: Core Service (S2.2)

**File**: `pkg/dashboard/project_access_impl.go` (NEW)

Implement ProjectAccessServiceImpl methods:
1. Query K8s for RoleBindings in project namespace
2. Match user to RoleBindings
3. Map ClusterRole to C8S role
4. Implement caching with 5-minute TTL
5. Handle errors gracefully (log but don't fail on K8s errors)

**K8s Query Pattern**:
```go
// Find RoleBindings where user is subject
roleBindings, err := k8sClient.ListRoleBindings(projectNamespace)

// Check each for user as subject
for _, rb := range roleBindings {
    for _, subject := range rb.Subjects {
        if subject.Name == userID && subject.Kind == "User" {
            // Found user in RoleBinding
            // Get bound ClusterRole
            cr, _ := k8sClient.GetClusterRole(rb.RoleRef.Name)
            // Map ClusterRole rules to C8S role
            role := MapK8sRoleToC8sRole(cr)
            return role
        }
    }
}
```

### Phase 2: Handler Authorization (S2.3)

**Strategy**:
1. Add authorization check to every handler that modifies or reads user-specific data
2. Pattern: Check project access → Check resource access → Proceed
3. Log all authorization decisions with user, resource, action, result
4. Return 403 Forbidden for denied access

**Handler Categories**:
- **Public** (no auth needed): `/login`
- **Optional Auth** (better UX with auth): Dashboard, artifact preview
- **Required Auth**: All API endpoints
- **Project-scoped** (check project access): Projects, pipelines, runs
- **Admin Only** (require admin role): Webhook config, project deletion

### Phase 3: Field-Level Access (S2.4)

**Approach**:
- Don't return fields user can't see
- Example: Hide email/internal IDs from non-admins
- Example: Hide artifacts from users without read access
- Use response filtering based on user role

### Phase 4: Testing (S2.5)

**Test Categories**:
1. **Unit Tests** (ProjectAccessService)
   - Valid role queries
   - Invalid users
   - Missing RoleBindings
   - K8s errors

2. **Integration Tests** (Handlers with authorization)
   - Admin can access all
   - Editor can write but not delete
   - Viewer can read only
   - Non-member gets 403

3. **Middleware Tests**
   - ProjectAccessMiddleware enforces
   - RoleBasedContextMiddleware attaches role
   - Authorization propagates through chain

---

## Kubernetes Integration

### RoleBinding Structure

C8S uses standard K8s RBAC RoleBindings:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pipeline-viewer
  namespace: project-alpha
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: c8s-viewer  # Maps to C8S viewer role
subjects:
  - kind: User
    name: "user-123"  # User ID from JWT 'sub' claim
```

### ClusterRole Mapping

C8S defines ClusterRoles:
- `c8s-admin` → C8S admin role (full access)
- `c8s-editor` → C8S editor role (read/write)
- `c8s-viewer` → C8S viewer role (read-only)

### Role Hierarchy

```
admin (3)
  ├─ Can: create, read, update, delete, manage access
  └─ Inherits: all editor and viewer permissions

editor (2)
  ├─ Can: create, read, update (but not delete)
  └─ Inherits: all viewer permissions

viewer (1)
  └─ Can: read only
```

---

## Error Handling

### Authorization Errors

| Scenario | Status | Response | Log |
|----------|--------|----------|-----|
| User not authenticated | 401 | "Unauthorized" | Warning |
| User not authorized | 403 | "Forbidden" | Warning |
| K8s error on check | 500 | "Server error" | Error |
| Invalid role/action | 500 | "Server error" | Error |

### Logging

Every authorization check is logged:
```
level=INFO user=user-123 action=read resource=project:proj-1 allowed=true role=viewer
level=WARN user=user-456 action=delete resource=artifact:art-1 allowed=false role=viewer
level=ERROR action=check resource=project:proj-2 err="k8s_error"
```

---

## Caching Strategy

### TTL-based Cache

```go
type CacheEntry struct {
    Result    AuthzResult
    ExpiresAt time.Time
}

cache := make(map[string]CacheEntry)
cacheTTL := 5 * time.Minute  // Configurable

// Check cache before K8s query
key := fmt.Sprintf("user:%s:project:%s", userID, projectID)
if cached, ok := cache[key]; ok && time.Now().Before(cached.ExpiresAt) {
    return cached.Result
}

// Query K8s if cache miss
role, err := queryK8sRoleBindings(userID, projectID)
if err == nil {
    cache[key] = CacheEntry{
        Result:    buildResult(role),
        ExpiresAt: time.Now().Add(cacheTTL),
    }
}
```

### Cache Invalidation

Cached entries expire after TTL. Manual invalidation on:
- Project deletion
- RoleBinding creation/update
- User deactivation

---

## Files to Create/Modify

### New Files

1. `pkg/dashboard/authorization_service.go` - Core authorization service
2. `pkg/dashboard/authz_result.go` - Result types
3. `pkg/dashboard/permissions.go` - Permission matrix
4. `pkg/dashboard/project_access_impl.go` - K8s integration

### Modified Files

1. `pkg/dashboard/project_access.go` - Add RoleLevel() method
2. `cmd/api-server/handlers/auth_middleware.go` - Add AuthorizationMiddleware
3. `cmd/api-server/handlers/authz_middleware.go` - Update middleware
4. `cmd/api-server/handlers/projects.go` - Add authorization checks
5. `cmd/api-server/handlers/artifacts.go` - Add authorization checks
6. `cmd/api-server/main.go` - Register authorization middleware

### Test Files

1. `tests/unit/auth/authorization_service_test.go` - 15+ unit tests
2. `tests/unit/handlers/authorization_test.go` - 20+ integration tests

---

## Success Criteria (S2.x Complete)

- ✅ ProjectAccessService fully implemented with K8s queries
- ✅ All handlers have authorization checks
- ✅ Field-level access control implemented
- ✅ 35+ unit and integration tests
- ✅ >90% code coverage for authorization code
- ✅ Comprehensive authorization documentation
- ✅ All authorization decisions logged
- ✅ Admin/Editor/Viewer permissions enforced
- ✅ Error handling for K8s failures
- ✅ 5-minute cache with proper invalidation

---

## Next Steps

### S2.2: Implement ProjectAccessService
- Query K8s RoleBindings
- Map to C8S roles
- Implement caching
- Handle errors gracefully

### S2.3: Add Handler Authorization Checks
- Update all handlers
- Check project access first
- Check resource access
- Log decisions

### S2.4: Field-Level Access Control
- Filter response fields
- Hide sensitive data
- Role-based UI rendering

### S2.5: Add Unit Tests
- Test service methods
- Test middleware integration
- Test error scenarios

### S2.6: Create Documentation
- Authorization architecture guide
- Role and permission reference
- K8s integration guide
- Troubleshooting guide

---

## Summary

**S2.1 Complete**: Authorization system designed with:
- Clear role hierarchy (admin > editor > viewer)
- Layered authorization checks
- K8s-native integration
- Comprehensive testing strategy
- Caching for performance

**Status**: Ready to proceed with S2.2 (ProjectAccessService Implementation)

**Files Created**: 1 (this design document)
**Files to Create**: 4 new implementation files
**Files to Modify**: 6 existing files
**Test Files**: 2 new test suites
**Estimated Effort**: S2.2-S2.6 = 16 hours total

---

**Design Date**: 2025-11-02
**Complexity**: Medium
**Risk Level**: Low (based on existing patterns and K8s RBAC)
**Test Coverage Target**: >90%
**Documentation**: Comprehensive
