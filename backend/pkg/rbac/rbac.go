package rbac

import "context"

// PermissionCode 权限代码类型
type PermissionCode = string

const (
	PermProjectCreate      PermissionCode = "project:create"
	PermProjectDelete      PermissionCode = "project:delete"
	PermProjectEdit        PermissionCode = "project:edit"
	PermProjectView        PermissionCode = "project:view"
	PermIssueCreate        PermissionCode = "issue:create"
	PermIssueEdit          PermissionCode = "issue:edit"
	PermIssueDelete        PermissionCode = "issue:delete"
	PermIssueView          PermissionCode = "issue:view"
	PermIssueChangeStatus  PermissionCode = "issue:change_status"
	PermWorkflowManage     PermissionCode = "workflow:manage"
	PermBoardManage        PermissionCode = "board:manage"
	PermMemberManage       PermissionCode = "member:manage"
	PermUserManage         PermissionCode = "user:manage"
	PermRoleManage         PermissionCode = "role:manage"
)

// UserPermStore 用户权限存储接口
type UserPermStore interface {
	GetUserPermissions(ctx context.Context, userID int64) ([]PermissionCode, error)
	GetUserProjectPermissions(ctx context.Context, userID, projectID int64) ([]PermissionCode, error)
	GetUserProjectRole(ctx context.Context, userID, projectID int64) (string, error)
}

// Engine RBAC 权限引擎
type Engine struct {
	store UserPermStore
}

func NewEngine(store UserPermStore) *Engine {
	return &Engine{store: store}
}

func (e *Engine) HasPermission(ctx context.Context, userID int64, perm PermissionCode) bool {
	perms, err := e.store.GetUserPermissions(ctx, userID)
	if err != nil {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

func (e *Engine) HasProjectPermission(ctx context.Context, userID, projectID int64, perm PermissionCode) bool {
	perms, err := e.store.GetUserProjectPermissions(ctx, userID, projectID)
	if err != nil {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}

func (e *Engine) Check(ctx context.Context, userID int64, perm PermissionCode) error {
	if !e.HasPermission(ctx, userID, perm) {
		return NewPermissionDeniedError(perm)
	}
	return nil
}

func (e *Engine) CheckProject(ctx context.Context, userID, projectID int64, perm PermissionCode) error {
	if !e.HasProjectPermission(ctx, userID, projectID, perm) {
		return NewPermissionDeniedError(perm)
	}
	return nil
}

// PermissionDeniedError 权限拒绝错误
type PermissionDeniedError struct {
	Permission string
}

func NewPermissionDeniedError(perm string) *PermissionDeniedError {
	return &PermissionDeniedError{Permission: perm}
}

func (e *PermissionDeniedError) Error() string {
	return "权限不足: " + e.Permission
}
