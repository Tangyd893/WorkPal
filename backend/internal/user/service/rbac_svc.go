package service

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	"github.com/Tangyd893/WorkPal/backend/internal/user/repo"
	"github.com/Tangyd893/WorkPal/backend/pkg/rbac"
)

type RBACService struct {
	rbacRepo *repo.RBACRepo
	engine   *rbac.Engine
}

func NewRBACService(rbacRepo *repo.RBACRepo) *RBACService {
	return &RBACService{
		rbacRepo: rbacRepo,
		engine:   rbac.NewEngine(rbacRepo),
	}
}

func (s *RBACService) Engine() *rbac.Engine { return s.engine }

type RoleDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

type PermissionDTO struct {
	ID           int64  `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ResourceType string `json:"resource_type"`
}

type ProjectRoleDTO struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

type ProjectMemberDTO struct {
	UserID    int64  `json:"user_id"`
	ProjectID int64  `json:"project_id"`
	RoleID    int64  `json:"role_id"`
	RoleName  string `json:"role_name"`
}

type AssignRoleInput struct {
	UserID    int64  `json:"user_id"`
	RoleID    int64  `json:"role_id"`
	ProjectID *int64 `json:"project_id"`
}

type AddProjectMemberInput struct {
	UserID        int64 `json:"user_id"`
	ProjectRoleID int64 `json:"project_role_id"`
}

func (s *RBACService) ListRoles(ctx context.Context) ([]RoleDTO, error) {
	roles, err := s.rbacRepo.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]RoleDTO, 0, len(roles))
	for _, r := range roles {
		out = append(out, RoleDTO{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			IsSystem:    r.IsSystem,
		})
	}
	return out, nil
}

func (s *RBACService) ListPermissions(ctx context.Context) ([]PermissionDTO, error) {
	perms, err := s.rbacRepo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PermissionDTO, 0, len(perms))
	for _, p := range perms {
		out = append(out, PermissionDTO{
			ID:           p.ID,
			Code:         p.Code,
			Name:         p.Name,
			Description:  p.Description,
			ResourceType: p.ResourceType,
		})
	}
	return out, nil
}

func (s *RBACService) AssignRole(ctx context.Context, input AssignRoleInput) error {
	return s.rbacRepo.AssignUserRole(ctx, input.UserID, input.RoleID, input.ProjectID)
}

func (s *RBACService) RemoveRole(ctx context.Context, input AssignRoleInput) error {
	return s.rbacRepo.RemoveUserRole(ctx, input.UserID, input.RoleID, input.ProjectID)
}

func (s *RBACService) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	return s.rbacRepo.GetUserPermissions(ctx, userID)
}

func (s *RBACService) GetUserProjectPermissions(ctx context.Context, userID, projectID int64) ([]string, error) {
	return s.rbacRepo.GetUserProjectPermissions(ctx, userID, projectID)
}

func (s *RBACService) ListProjectRoles(ctx context.Context, projectID int64) ([]ProjectRoleDTO, error) {
	roles, err := s.rbacRepo.ListProjectRoles(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(roles) == 0 {
		s.seedDefaultProjectRoles(ctx, projectID)
		roles, _ = s.rbacRepo.ListProjectRoles(ctx, projectID)
	}
	out := make([]ProjectRoleDTO, 0, len(roles))
	for _, r := range roles {
		out = append(out, ProjectRoleDTO{
			ID:          r.ID,
			ProjectID:   r.ProjectID,
			Name:        r.Name,
			Description: r.Description,
			IsSystem:    r.IsSystem,
		})
	}
	return out, nil
}

func (s *RBACService) CreateProjectRole(ctx context.Context, projectID int64, name, description string) (ProjectRoleDTO, error) {
	role := &model.ProjectRole{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		IsSystem:    false,
	}
	if err := s.rbacRepo.CreateProjectRole(ctx, role); err != nil {
		return ProjectRoleDTO{}, err
	}
	return ProjectRoleDTO{
		ID:          role.ID,
		ProjectID:   role.ProjectID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
	}, nil
}

func (s *RBACService) AddProjectMember(ctx context.Context, projectID int64, input AddProjectMemberInput) error {
	return s.rbacRepo.AddProjectMember(ctx, input.UserID, projectID, input.ProjectRoleID)
}

func (s *RBACService) RemoveProjectMember(ctx context.Context, projectID int64, userID int64) error {
	return s.rbacRepo.RemoveProjectMember(ctx, userID, projectID)
}

func (s *RBACService) ListProjectMembers(ctx context.Context, projectID int64) ([]ProjectMemberDTO, error) {
	members, err := s.rbacRepo.ListProjectMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}
	roles, _ := s.rbacRepo.ListProjectRoles(ctx, projectID)
	roleMap := make(map[int64]string)
	for _, r := range roles {
		roleMap[r.ID] = r.Name
	}
	out := make([]ProjectMemberDTO, 0, len(members))
	for _, m := range members {
		out = append(out, ProjectMemberDTO{
			UserID:    m.UserID,
			ProjectID: m.ProjectID,
			RoleID:    m.ProjectRoleID,
			RoleName:  roleMap[m.ProjectRoleID],
		})
	}
	return out, nil
}

func (s *RBACService) seedDefaultProjectRoles(ctx context.Context, projectID int64) {
	defaults := []struct {
		name   string
		perms  []string
	}{
		{"管理员", []string{
			"project:view", "project:edit", "project:delete",
			"issue:create", "issue:edit", "issue:delete", "issue:view", "issue:change_status",
			"workflow:manage", "board:manage", "member:manage",
		}},
		{"开发者", []string{
			"project:view",
			"issue:create", "issue:edit", "issue:view", "issue:change_status",
			"board:manage",
		}},
		{"观察者", []string{
			"project:view", "issue:view",
		}},
		{"报告者", []string{
			"project:view",
			"issue:create", "issue:view",
		}},
	}

	for _, d := range defaults {
		role := &model.ProjectRole{
			ProjectID:   projectID,
			Name:        d.name,
			IsSystem:    true,
		}
		_ = s.rbacRepo.CreateProjectRole(ctx, role)
	}
}
