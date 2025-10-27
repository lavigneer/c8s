package dashboard

import "context"

// Role represents a user's role in a project
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

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

// ProjectAccessServiceImpl implements ProjectAccessService
type ProjectAccessServiceImpl struct {
	k8sClient *K8sClient
}

// NewProjectAccessService creates a new project access service
func NewProjectAccessService(k8sClient *K8sClient) ProjectAccessService {
	return &ProjectAccessServiceImpl{
		k8sClient: k8sClient,
	}
}

// UserHasProjectAccess checks if user can access a project
func (s *ProjectAccessServiceImpl) UserHasProjectAccess(ctx context.Context, userID, projectID string) (bool, error) {
	// TODO: Implement actual access check via K8s RoleBinding/ClusterRoleBinding
	// For now, return true to allow all authenticated users
	// This should check K8s RBAC for the project namespace
	return true, nil
}

// GetUserRoleForProject returns user's role in project (admin/editor/viewer)
func (s *ProjectAccessServiceImpl) GetUserRoleForProject(ctx context.Context, userID, projectID string) (Role, error) {
	// TODO: Query K8s RoleBindings to determine user's role
	// Look for RoleBindings in project namespace that include the user
	// and determine role based on the bound ClusterRole

	// For now, default to viewer role
	return RoleViewer, nil
}

// ListUserProjects returns all projects user has access to
func (s *ProjectAccessServiceImpl) ListUserProjects(ctx context.Context, userID string) ([]ProjectDTO, error) {
	// TODO: Query K8s for all projects where user has RoleBinding
	// Return list of accessible projects sorted by name

	return []ProjectDTO{}, nil
}

// HasProjectRole checks if user has specific role for project
func (s *ProjectAccessServiceImpl) HasProjectRole(ctx context.Context, userID, projectID string, requiredRole Role) (bool, error) {
	actualRole, err := s.GetUserRoleForProject(ctx, userID, projectID)
	if err != nil {
		return false, err
	}

	// Check role hierarchy: admin > editor > viewer
	roleHierarchy := map[Role]int{
		RoleAdmin:  3,
		RoleEditor: 2,
		RoleViewer: 1,
	}

	actualLevel := roleHierarchy[actualRole]
	requiredLevel := roleHierarchy[requiredRole]

	return actualLevel >= requiredLevel, nil
}
