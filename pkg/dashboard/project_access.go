package dashboard

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
)

// Role represents a user's role in a project
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// RoleLevel returns numeric level for role hierarchy comparison
// Higher level = more permissions
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

// ProjectAccessService defines methods for checking user access to projects
type ProjectAccessService interface {
	// UserHasProjectAccess checks if user can access a project
	UserHasProjectAccess(ctx context.Context, userID, projectID string) (bool, error)

	// GetUserRoleForProject returns user's role in project (admin/editor/viewer)
	GetUserRoleForProject(ctx context.Context, userID, projectID string) (Role, error)

	// ListUserProjects returns all projects user has access to
	ListUserProjects(ctx context.Context, userID string) ([]ProjectDTO, error)

	// HasProjectRole checks if user has specific role for project
	HasProjectRole(ctx context.Context, userID, projectID string, requiredRole Role) (bool, error)
}

// CachedRoleResult holds cached role lookup with expiration
type CachedRoleResult struct {
	Role      Role
	ExpiresAt time.Time
}

// ProjectAccessServiceImpl implements ProjectAccessService with K8s RBAC integration
type ProjectAccessServiceImpl struct {
	k8sClient *K8sClient
	cache     map[string]CachedRoleResult
	cacheMu   sync.RWMutex
	cacheTTL  time.Duration
}

// NewProjectAccessService creates a new project access service
func NewProjectAccessService(k8sClient *K8sClient) ProjectAccessService {
	return &ProjectAccessServiceImpl{
		k8sClient: k8sClient,
		cache:     make(map[string]CachedRoleResult),
		cacheTTL:  5 * time.Minute, // 5-minute cache TTL
	}
}

// UserHasProjectAccess checks if user can access a project by checking if they have any role
func (s *ProjectAccessServiceImpl) UserHasProjectAccess(ctx context.Context, userID, projectID string) (bool, error) {
	role, err := s.GetUserRoleForProject(ctx, userID, projectID)
	if err != nil {
		// Log error but don't fail - user doesn't have access if we can't determine role
		log.Printf("Failed to get user role for access check: user=%s project=%s err=%v", userID, projectID, err)
		return false, nil
	}

	// User has access if they have any role (even viewer)
	return role.Level() > 0, nil
}

// GetUserRoleForProject returns user's role in project by querying K8s RoleBindings
func (s *ProjectAccessServiceImpl) GetUserRoleForProject(ctx context.Context, userID, projectID string) (Role, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("user:%s:project:%s", userID, projectID)
	s.cacheMu.RLock()
	if cached, ok := s.cache[cacheKey]; ok && time.Now().Before(cached.ExpiresAt) {
		s.cacheMu.RUnlock()
		return cached.Role, nil
	}
	s.cacheMu.RUnlock()

	// Convert projectID to namespace name (project-name → project-name namespace)
	namespace := projectID

	// Query K8s RoleBindings in the project namespace
	roleBindings, err := s.k8sClient.ListRoleBindings(ctx, namespace)
	if err != nil {
		// Return no access on error (safe fail-closed)
		log.Printf("Failed to list role bindings: namespace=%s err=%v", namespace, err)
		return "", fmt.Errorf("failed to check access: %w", err)
	}

	// Find RoleBindings that include this user
	var highestRole Role
	for _, rb := range roleBindings.Items {
		// Check if user is a subject of this RoleBinding
		userFound := false
		for _, subject := range rb.Subjects {
			if subject.Kind == "User" && subject.Name == userID {
				userFound = true
				break
			}
		}

		if !userFound {
			continue
		}

		// User is in this RoleBinding, get the bound ClusterRole
		clusterRoleName := rb.RoleRef.Name
		clusterRole, err := s.k8sClient.GetClusterRole(ctx, clusterRoleName)
		if err != nil {
			log.Printf("Failed to get cluster role: name=%s err=%v", clusterRoleName, err)
			continue
		}

		// Map ClusterRole to C8S role
		role := s.mapClusterRoleToRole(clusterRole)
		if role.Level() > highestRole.Level() {
			highestRole = role
		}
	}

	// Cache the result
	s.cacheMu.Lock()
	s.cache[cacheKey] = CachedRoleResult{
		Role:      highestRole,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}
	s.cacheMu.Unlock()

	if highestRole == "" {
		return "", fmt.Errorf("user has no role in project: user=%s project=%s", userID, projectID)
	}

	return highestRole, nil
}

// mapClusterRoleToRole maps K8s ClusterRole to C8S role based on naming convention
// In production, this would analyze ClusterRole rules, but for MVP we use role name patterns
func (s *ProjectAccessServiceImpl) mapClusterRoleToRole(clusterRole *rbacv1.ClusterRole) Role {
	if clusterRole == nil {
		return RoleViewer // Default to viewer for safety
	}

	// Map based on ClusterRole name using naming convention
	switch clusterRole.Name {
	case "c8s-admin", "admin":
		return RoleAdmin
	case "c8s-editor", "editor":
		return RoleEditor
	case "c8s-viewer", "viewer":
		return RoleViewer
	default:
		// Default to viewer for unknown roles
		return RoleViewer
	}
}

// ListUserProjects returns all projects where user has any access (read-only MVP)
func (s *ProjectAccessServiceImpl) ListUserProjects(ctx context.Context, userID string) ([]ProjectDTO, error) {
	// TODO: Query K8s ClusterRoleBindings across all namespaces
	// For MVP, return empty list - applications use project listing from API instead
	return []ProjectDTO{}, nil
}

// HasProjectRole checks if user's role meets minimum requirement using role hierarchy
func (s *ProjectAccessServiceImpl) HasProjectRole(ctx context.Context, userID, projectID string, requiredRole Role) (bool, error) {
	actualRole, err := s.GetUserRoleForProject(ctx, userID, projectID)
	if err != nil {
		return false, err
	}

	// Check role hierarchy: admin > editor > viewer
	actualLevel := actualRole.Level()
	requiredLevel := requiredRole.Level()

	return actualLevel >= requiredLevel, nil
}
