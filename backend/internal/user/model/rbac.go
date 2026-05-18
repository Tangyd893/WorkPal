package model

import "time"

type Role struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"size:64;not null;uniqueIndex" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Code         string    `gorm:"size:128;not null;uniqueIndex" json:"code"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	Description  string    `gorm:"size:255" json:"description"`
	ResourceType string    `gorm:"size:64;not null;default:global" json:"resource_type"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	RoleID       int64 `gorm:"primaryKey" json:"role_id"`
	PermissionID int64 `gorm:"primaryKey" json:"permission_id"`
}

func (RolePermission) TableName() string { return "role_permissions" }

type UserRole struct {
	UserID    int64     `gorm:"primaryKey" json:"user_id"`
	RoleID    int64     `gorm:"primaryKey" json:"role_id"`
	ProjectID *int64    `gorm:"primaryKey" json:"project_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserRole) TableName() string { return "user_roles" }

type ProjectRole struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectID   int64     `gorm:"not null;uniqueIndex:idx_prj_role" json:"project_id"`
	Name        string    `gorm:"size:64;not null;uniqueIndex:idx_prj_role" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	IsSystem    bool      `gorm:"default:true" json:"is_system"`
}

func (ProjectRole) TableName() string { return "project_roles" }

type ProjectRolePermission struct {
	ProjectRoleID int64 `gorm:"primaryKey" json:"project_role_id"`
	PermissionID  int64 `gorm:"primaryKey" json:"permission_id"`
}

func (ProjectRolePermission) TableName() string { return "project_role_permissions" }

type ProjectMember struct {
	UserID        int64     `gorm:"primaryKey" json:"user_id"`
	ProjectID     int64     `gorm:"primaryKey" json:"project_id"`
	ProjectRoleID int64     `gorm:"not null" json:"project_role_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func (ProjectMember) TableName() string { return "project_members" }
