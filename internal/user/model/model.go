package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:128" json:"nickname"`
	AvatarURL    string         `gorm:"size:512" json:"avatar_url"`
	Email        string         `gorm:"uniqueIndex;size:255" json:"email"`
	Phone        string         `gorm:"size:32" json:"phone"`
	Status       int8           `gorm:"default:1" json:"status"` // 1=正常 2=禁用
	DepartmentID int64          `gorm:"default:0" json:"department_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

type Department struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string `gorm:"size:128;not null" json:"name"`
	ParentID  int64  `gorm:"default:0" json:"parent_id"`
	LeaderID  int64  `gorm:"default:0" json:"leader_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (Department) TableName() string {
	return "departments"
}
