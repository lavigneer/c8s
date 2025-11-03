# S2.6: Authorization System Documentation

**Task**: Create comprehensive authorization documentation
**Status**: ✅ Complete
**Date**: 2025-11-02
**Effort**: 3 hours (documentation creation)

---

## Authorization Architecture Overview

The C8S authorization system implements **Role-Based Access Control (RBAC)** with three tiers of permissions and field-level access control.

### Architecture Diagram

```
Request → Authentication Middleware
  ↓ (extract user from JWT)
Authorization Service (ProjectAccessService)
  ↓ (check user role for project)
Handler Authorization Check
  ↓ (verify action permission)
Field-Level Filtering
  ↓ (remove sensitive fields)
HTTP Response (filtered data)
```

---

## Role Hierarchy

The system uses three roles with strict hierarchy:

```
┌─────────────────────────────────────┐
│         Admin (Level 3)             │
│  - All permissions                  │
│  - Modify settings                  │
│  - Delete resources                 │
│  - Access sensitive fields          │
└──────────────┬──────────────────────┘
               │ inherits from
┌──────────────▼──────────────────────┐
│       Editor (Level 2)              │
│  - Create/update pipelines          │
│  - View/download artifacts          │
│  - Access webhook URLs              │
│  - View author emails               │
└──────────────┬──────────────────────┘
               │ inherits from
┌──────────────▼──────────────────────┐
│       Viewer (Level 1)              │
│  - Read-only access                 │
│  - View pipeline runs               │
│  - Stream logs                      │
│  - Basic project info               │
└─────────────────────────────────────┘
```

### Role Levels

| Role | Level | Can Read | Can Write | Can Delete |
|------|-------|----------|-----------|-----------|
| Viewer | 1 | ✅ | ❌ | ❌ |
| Editor | 2 | ✅ | ✅ | ❌ |
| Admin | 3 | ✅ | ✅ | ✅ |

---

## Component Overview

### 1. ProjectAccessService Interface

**Location**: `pkg/dashboard/project_access.go`

**Purpose**: Central service for checking user access to projects

**Methods**:
```go
// Check if user has any access to project
UserHasProjectAccess(ctx, userID, projectID) (bool, error)

// Get user's specific role for project
GetUserRoleForProject(ctx, userID, projectID) (Role, error)

// Check if user has required role
HasProjectRole(ctx, userID, projectID, requiredRole) (bool, error)

// List all projects user has access to
ListUserProjects(ctx, userID) ([]ProjectDTO, error)
```

**Implementation Details**:
- Integrates with Kubernetes RBAC via ClusterRoleBindings
- Caches results for 5 minutes (TTL) for performance
- Maps K8s ClusterRoles to C8S roles by naming convention
- Supports multiple roles per user per project (takes highest)

### 2. Authorization Helper Functions

**Location**: `cmd/api-server/handlers/authorization_helper.go`

**Purpose**: HTTP handler utilities for enforcement

#### CheckProjectAccess
```go
func CheckProjectAccess(w http.ResponseWriter, r *http.Request,
                       user *User, projectID string,
                       requiredRole dashboard.Role) bool
```

**Behavior**:
- Verifies user has required role for project
- Returns 403 Forbidden if denied
- Returns 500 Internal Server Error if service unavailable
- Logs all authorization decisions for audit trail

**Usage**:
```go
user, ok := GetUserFromContext(r.Context())
if !ok {
    // Not authenticated
    return
}

if !CheckProjectAccess(w, r, user, projectID, dashboard.RoleEditor) {
    // Access denied or error - already sent response
    return
}

// Proceed with operation
```

#### CheckProjectAccessAction
```go
func CheckProjectAccessAction(w http.ResponseWriter, r *http.Request,
                             user *User, projectID string,
                             action AuthorizationAction) bool
```

**Action-to-Role Mapping**:
| Action | Required Role |
|--------|---------------|
| `ActionRead` | RoleViewer |
| `ActionWrite` | RoleEditor |
| `ActionDelete` | RoleAdmin |
| `ActionAdmin` | RoleAdmin |

**Usage**:
```go
// Handler for DELETE /api/projects/{id}
if !CheckProjectAccessAction(w, r, user, projectID, ActionDelete) {
    return  // 403 Forbidden
}
// Delete the project
```

#### CheckUserExists
```go
func CheckUserExists(w http.ResponseWriter, r *http.Request) (*User, bool)
```

**Purpose**: Verify user is authenticated

**Returns**:
- `(*User, true)` if user in context
- `(nil, false)` with 401 Unauthorized if not

### 3. Field-Level Access Control

**Location**: `pkg/dashboard/field_access.go`

**Purpose**: Apply principle of least privilege by filtering sensitive fields

#### Filter Functions

```go
// Filter single DTO based on role
FilterProjectDTOForRole(dto *ProjectDTO, role Role) *ProjectDTO
FilterPipelineRunDTOForRole(dto *PipelineRunDTO, role Role) *PipelineRunDTO
FilterArtifactDTOForRole(dto *ArtifactDTO, role Role) *ArtifactDTO

// Filter multiple DTOs (batch operation)
FilterProjectDTOsForRole(dtos []*ProjectDTO, role Role) []*ProjectDTO
FilterPipelineRunDTOsForRole(dtos []*PipelineRunDTO, role Role) []*PipelineRunDTO
```

#### Field Visibility Matrix

**ProjectDTO Fields**:
| Field | Viewer | Editor | Admin |
|-------|--------|--------|-------|
| ID | ✅ | ✅ | ✅ |
| Name | ✅ | ✅ | ✅ |
| Description | ✅ | ✅ | ✅ |
| RepoURL | ✅ | ✅ | ✅ |
| **WebhookURL** | ❌ | ✅ | ✅ |
| Namespace | ✅ | ✅ | ✅ |
| **LastRunAt** | ❌ | ✅ | ✅ |
| RunCount | ✅ | ✅ | ✅ |

**PipelineRunDTO Fields**:
| Field | Viewer | Editor | Admin |
|-------|--------|--------|-------|
| ID | ✅ | ✅ | ✅ |
| Status | ✅ | ✅ | ✅ |
| CommitSHA | ✅ | ✅ | ✅ |
| **AuthorEmail** | ❌ | ✅ | ✅ |
| CreatedAt | ✅ | ✅ | ✅ |

**ArtifactDTO Fields**:
| Field | Viewer | Editor | Admin |
|-------|--------|--------|-------|
| ID | ✅ | ✅ | ✅ |
| Name | ✅ | ✅ | ✅ |
| Size | ✅ | ✅ | ✅ |
| **URL** | ❌ | ✅ | ✅ |

---

## Implementation Patterns

### Pattern 1: Requiring Authentication + Authorization

```go
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
    // Step 1: Check authentication
    user, ok := CheckUserExists(w, r)
    if !ok {
        return  // 401 Unauthorized
    }

    projectID := mux.Vars(r)["projectId"]

    // Step 2: Check authorization
    if !CheckProjectAccessAction(w, r, user, projectID, ActionDelete) {
        return  // 403 Forbidden
    }

    // Step 3: Execute operation
    // Delete project...

    dashboard.RespondSuccess(w, http.StatusOK, result)
}
```

### Pattern 2: Field-Level Filtering

```go
func ListProjectsHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := CheckUserExists(w, r)
    if !ok {
        return  // 401
    }

    projects, err := service.GetProjects(r.Context())
    if err != nil {
        dashboard.RespondError(w, http.StatusInternalServerError, ...)
        return
    }

    // Filter by access + apply field-level filtering
    var filtered []*dashboard.ProjectDTO
    for _, project := range projects {
        if !CheckProjectAccess(w, r, user, project.ID, RoleViewer) {
            continue  // User doesn't have access
        }

        role, _ := authzService.GetUserRoleForProject(...)
        filtered = append(filtered,
            dashboard.FilterProjectDTOForRole(project, role))
    }

    dashboard.RespondSuccess(w, http.StatusOK, filtered)
}
```

### Pattern 3: Action-Based Authorization

```go
func handlePipelineAction(w http.ResponseWriter, r *http.Request, action string) {
    user, ok := CheckUserExists(w, r)
    if !ok {
        return
    }

    var authAction AuthorizationAction
    switch action {
    case "create":
        authAction = ActionWrite
    case "delete":
        authAction = ActionDelete
    case "view":
        authAction = ActionRead
    }

    projectID := r.URL.Query().Get("project")
    if !CheckProjectAccessAction(w, r, user, projectID, authAction) {
        return
    }

    // Proceed with operation
}
```

---

## HTTP Status Codes

### Authentication & Authorization Response Codes

| Status | Scenario | Response Example |
|--------|----------|------------------|
| **401** | No auth token | `{"code":"UNAUTHORIZED","message":"User not authenticated"}` |
| **403** | Insufficient role | `{"code":"FORBIDDEN","message":"You do not have permission..."}` |
| **400** | Bad request (missing project ID) | `{"code":"BAD_REQUEST","message":"..."}` |
| **500** | Service error | `{"code":"SERVER_ERROR","message":"Failed to verify permissions"}` |

---

## API Response Examples

### Example 1: List Projects (Viewer)

**Request**:
```bash
GET /api/projects
Authorization: Bearer <viewer-token>
```

**Response (200 OK)**:
```json
[
  {
    "id": "proj-1",
    "name": "Frontend App",
    "description": "React frontend",
    "repository_url": "https://github.com/example/frontend",
    "namespace": "production",
    "created_at": "2025-11-01T10:00:00Z",
    "run_count": 42
    // webhook_url: NOT INCLUDED (viewer cannot see)
    // last_run_at: NOT INCLUDED (viewer cannot see)
  }
]
```

### Example 2: List Projects (Admin)

**Request**:
```bash
GET /api/projects
Authorization: Bearer <admin-token>
```

**Response (200 OK)**:
```json
[
  {
    "id": "proj-1",
    "name": "Frontend App",
    "description": "React frontend",
    "repository_url": "https://github.com/example/frontend",
    "webhook_url": "https://c8s.example.com/webhooks/github/proj-1",
    "namespace": "production",
    "created_at": "2025-11-01T10:00:00Z",
    "last_run_at": "2025-11-02T15:30:00Z",
    "run_count": 42
  }
]
```

### Example 3: Delete Project (Viewer)

**Request**:
```bash
DELETE /api/projects/proj-1
Authorization: Bearer <viewer-token>
```

**Response (403 Forbidden)**:
```json
{
  "code": "FORBIDDEN",
  "message": "You do not have permission to perform this action"
}
```

### Example 4: Delete Project (Admin)

**Request**:
```bash
DELETE /api/projects/proj-1
Authorization: Bearer <admin-token>
```

**Response (200 OK)**:
```json
{
  "message": "Project deleted successfully"
}
```

### Example 5: Authentication Failure

**Request**:
```bash
GET /api/projects
// No Authorization header
```

**Response (401 Unauthorized)**:
```json
{
  "code": "UNAUTHORIZED",
  "message": "User not authenticated"
}
```

---

## Kubernetes RBAC Integration

### How K8s Roles Map to C8S Roles

The authorization system integrates with Kubernetes RBAC:

**Naming Convention**:
```
ClusterRole naming: c8s-{project-name}-{role}

Examples:
- c8s-frontend-viewer   → RoleViewer for "frontend" project
- c8s-api-editor       → RoleEditor for "api" project
- c8s-admin-admin      → RoleAdmin for "admin" project
```

**Role Detection**:
1. Query all ClusterRoleBindings for user
2. Filter to roles matching pattern `c8s-*-{role}`
3. Extract project name and role from binding name
4. Cache result for 5 minutes

**Example K8s Setup**:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: c8s-frontend-viewer
rules:
  - apiGroups: ["c8s.io"]
    resources: ["projects"]
    verbs: ["get", "list"]
    resourceNames: ["frontend"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: alice-frontend-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: c8s-frontend-viewer
subjects:
  - kind: User
    name: alice@example.com
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: c8s-frontend-editor
rules:
  - apiGroups: ["c8s.io"]
    resources: ["projects", "pipelines"]
    verbs: ["get", "list", "create", "update"]
    resourceNames: ["frontend"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: bob-frontend-editor
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: c8s-frontend-editor
subjects:
  - kind: User
    name: bob@example.com
```

### Setting Up K8s RBAC for C8S

**Step 1: Create Custom Resource Definition**

```bash
# Usually pre-created by C8S installation
kubectl apply -f crds/c8s-project.yaml
```

**Step 2: Create Project Namespace**

```bash
kubectl create namespace c8s-projects
```

**Step 3: Create ClusterRole for Viewer**

```bash
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: c8s-myproject-viewer
rules:
  - apiGroups: ["c8s.io"]
    resources: ["projects"]
    verbs: ["get", "list"]
    resourceNames: ["myproject"]
EOF
```

**Step 4: Bind Role to User**

```bash
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: alice-myproject-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: c8s-myproject-viewer
subjects:
  - kind: User
    name: alice@example.com
EOF
```

**Step 5: Verify Setup**

```bash
# Check if user has role binding
kubectl get clusterrolebinding | grep alice

# Check role permissions
kubectl get clusterrole c8s-myproject-viewer -o yaml
```

---

## Authorization Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     HTTP Request                                │
│        GET /api/projects/proj-1/webhook-config                 │
│        Authorization: Bearer <token>                            │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              Authentication Middleware                          │
│         ┌─ Validate JWT signature                               │
│         ├─ Extract user claims (sub, email, roles)             │
│         └─ Put User struct in request context                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              Handler Begins Execution                           │
│  GetWebhookConfigHandler(w, r)                                  │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              Step 1: Check Authentication                       │
│  user, ok := CheckUserExists(w, r)                             │
│  if !ok { return }  // 401 Unauthorized                        │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              Step 2: Check Authorization                        │
│  CheckProjectAccessAction(w, r, user, "proj-1",               │
│      ActionAdmin)  // Requires admin                           │
│  ┌─ Query ProjectAccessService                                 │
│  ├─ Get user role for project (from K8s RBAC)                  │
│  ├─ Compare: RoleAdmin >= RoleAdmin?  YES                      │
│  └─ If NO → return 403 Forbidden                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              Step 3: Execute Operation                          │
│  webhook := getWebhookConfig("proj-1")                          │
│  // Business logic...                                           │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              Step 4: Filter Response Data                       │
│  dto := mapToDTO(webhook)                                       │
│  role, _ := authzService.GetUserRole(...)  // Admin            │
│  filteredDTO := FilterWebhookDTOForRole(dto, role)             │
│  // All fields visible to admin                                │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│              Step 5: Send Response                              │
│  HTTP 200 OK                                                    │
│  {                                                              │
│    "webhook_url": "https://...",                               │
│    "secret": "***",                                            │
│    "events": ["push", "pull_request"]                          │
│  }                                                              │
└─────────────────────────────────────────────────────────────────┘
```

---

## Common Authorization Patterns

### Pattern 1: Admin-Only Endpoint

```go
func ConfigureWebhooksHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := CheckUserExists(w, r)
    if !ok { return }

    projectID := mux.Vars(r)["projectId"]

    // Require admin role
    if !CheckProjectAccessAction(w, r, user, projectID, ActionAdmin) {
        return  // 403
    }

    // Proceed with webhook configuration
}
```

### Pattern 2: Editor-or-Higher Endpoint

```go
func CreatePipelineHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := CheckUserExists(w, r)
    if !ok { return }

    projectID := mux.Vars(r)["projectId"]

    // Require editor role (admin inherits)
    if !CheckProjectAccessAction(w, r, user, projectID, ActionWrite) {
        return  // 403
    }

    // Create pipeline
}
```

### Pattern 3: Viewer-and-Higher Endpoint (with field filtering)

```go
func GetProjectHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := CheckUserExists(w, r)
    if !ok { return }

    projectID := mux.Vars(r)["projectId"]

    // Require viewer role (all roles inherit)
    if !CheckProjectAccessAction(w, r, user, projectID, ActionRead) {
        return  // 403
    }

    // Get project
    project := getProject(projectID)
    dto := mapToProjectDTO(project)

    // Filter fields based on role
    role, _ := authzService.GetUserRoleForProject(...)
    filtered := FilterProjectDTOForRole(dto, role)

    dashboard.RespondSuccess(w, http.StatusOK, filtered)
}
```

---

## Troubleshooting

### User gets 401 Unauthorized

**Causes**:
- Missing `Authorization` header
- Invalid JWT token
- Expired token
- Invalid token signature

**Solution**:
```bash
# Verify token is valid
jwt decode <token>

# Check header is properly formatted
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/projects
```

### User gets 403 Forbidden

**Causes**:
- User doesn't have required role
- Role binding not created in K8s
- Role name doesn't match pattern

**Solution**:
```bash
# Check role bindings for user
kubectl get clusterrolebinding | grep <username>

# Check role exists
kubectl get clusterrole | grep c8s-<project>-<role>

# Verify role has required permissions
kubectl get clusterrole c8s-<project>-<role> -o yaml
```

### Field is visible but shouldn't be

**Cause**: Field filtering not applied in handler

**Solution**:
```go
// Before: Missing filter
dto := mapToProjectDTO(project)
respond(w, http.StatusOK, dto)  // ❌ All fields visible

// After: Apply filter
role, _ := authzService.GetUserRoleForProject(...)
dto := FilterProjectDTOForRole(mapToProjectDTO(project), role)
respond(w, http.StatusOK, dto)  // ✅ Filtered
```

---

## Security Checklist

- ✅ Authentication checked before authorization
- ✅ All admin endpoints require RoleAdmin
- ✅ All write operations require RoleEditor or higher
- ✅ All read operations require RoleViewer or higher
- ✅ Sensitive fields filtered based on role
- ✅ Error messages don't leak details
- ✅ All authorization decisions logged
- ✅ Service errors return 500, not 403
- ✅ Concurrent access is safe
- ✅ K8s RBAC integrated properly

---

## References

### Key Files

- `cmd/api-server/handlers/authorization_helper.go` - Helper functions
- `pkg/dashboard/project_access.go` - Authorization service
- `pkg/dashboard/field_access.go` - Field filtering
- `tests/unit/handlers/authorization_*.go` - 52 tests

### Related Documentation

- `docs/AUTHENTICATION.md` - JWT authentication setup
- `CLAUDE.md` - Development guidelines
- `S2_1_AUTHORIZATION_DESIGN.md` - Design specification
- `S2_2_PROJECTACCESS_IMPLEMENTATION.md` - Service implementation
- `S2_3_HANDLER_AUTHORIZATION.md` - Handler checks
- `S2_4_FIELD_ACCESS_CONTROL.md` - Field filtering
- `S2_5_COMPREHENSIVE_TESTING.md` - Test coverage

---

## Conclusion

The authorization system provides:

✅ **Three-tier RBAC** with clear hierarchy (Admin > Editor > Viewer)
✅ **K8s RBAC integration** for centralized access management
✅ **Field-level access control** with principle of least privilege
✅ **Comprehensive error handling** with proper HTTP status codes
✅ **Audit logging** of all authorization decisions
✅ **Production-ready** with 52 passing tests

Authorization enforcement is complete and production-ready.

---

**Documentation Date**: 2025-11-02
**Status**: ✅ Complete
**Test Coverage**: 52 tests passing
**Code Quality**: Production Ready
