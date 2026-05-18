package repo

import (
	"context"

	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	"github.com/Tangyd893/WorkPal/backend/pkg/rbac"
	"gorm.io/gorm"
)

type RBACRepo struct {
	db *gorm.DB
}

func NewRBACRepo(db *gorm.DB) *RBACRepo {
	return &RBACRepo{db: db}
}

func (r *RBACRepo) GetUserPermissions(ctx context.Context, userID int64) ([]rbac.PermissionCode, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("user_roles ur").
		Select("DISTINCT p.code").
		Joins("JOIN role_permissions rp ON rp.role_id = ur.role_id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("ur.user_id = ? AND ur.project_id IS NULL", userID).
		Pluck("p.code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *RBACRepo) GetUserProjectPermissions(ctx context.Context, userID, projectID int64) ([]rbac.PermissionCode, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("project_members pm").
		Select("DISTINCT p.code").
		Joins("JOIN project_role_permissions prp ON prp.project_role_id = pm.project_role_id").
		Joins("JOIN permissions p ON p.id = prp.permission_id").
		Where("pm.user_id = ? AND pm.project_id = ?", userID, projectID).
		Pluck("p.code", &codes).Error
	if err != nil {
		return nil, err
	}

	globalCodes, err := r.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}
	codes = append(codes, globalCodes...)
	return codes, nil
}

func (r *RBACRepo) GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error) {
	var roleName string
	err := r.db.WithContext(ctx).
		Table("project_members pm").
		Select("pr.name").
		Joins("JOIN project_roles pr ON pr.id = pm.project_role_id").
		Where("pm.user_id = ? AND pm.project_id = ?", userID, projectID).
		Pluck("pr.name", &roleName).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return roleName, nil
}

func (r *RBACRepo) ListRoles(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (r *RBACRepo) ListPermissions(ctx context.Context) ([]*model.Permission, error) {
	var perms []*model.Permission
	err := r.db.WithContext(ctx).Find(&perms).Error
	return perms, err
}

func (r *RBACRepo) AssignUserRole(ctx context.Context, userID, roleID int64, projectID *int64) error {
	ur := &model.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		ProjectID: projectID,
	}
	return r.db.WithContext(ctx).Create(ur).Error
}

func (r *RBACRepo) RemoveUserRole(ctx context.Context, userID, roleID int64, projectID *int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ? AND COALESCE(project_id, 0) = COALESCE(?, 0)", userID, roleID, projectID).
		Delete(&model.UserRole{}).Error
}

func (r *RBACRepo) ListProjectRoles(ctx context.Context, projectID int64) ([]*model.ProjectRole, error) {
	var roles []*model.ProjectRole
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&roles).Error
	return roles, err
}

func (r *RBACRepo) CreateProjectRole(ctx context.Context, role *model.ProjectRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RBACRepo) AddProjectMember(ctx context.Context, userID, projectID, projectRoleID int64) error {
	pm := &model.ProjectMember{
		UserID:        userID,
		ProjectID:     projectID,
		ProjectRoleID: projectRoleID,
	}
	return r.db.WithContext(ctx).Create(pm).Error
}

func (r *RBACRepo) RemoveProjectMember(ctx context.Context, userID, projectID int64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Delete(&model.ProjectMember{}).Error
}

func (r *RBACRepo) ListProjectMembers(ctx context.Context, projectID int64) ([]*model.ProjectMember, error) {
	var members []*model.ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&members).Error
	return members, err
}
