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
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	ParentID    int64     `gorm:"default:0" json:"parent_id"`
	LeaderID    int64     `gorm:"default:0" json:"leader_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Department) TableName() string {
	return "departments"
}

type Employee struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         int64          `gorm:"uniqueIndex;not null" json:"user_id"`
	EmployeeNo     string         `gorm:"uniqueIndex;size:32;not null" json:"employee_no"`
	JobTitle       string         `gorm:"size:128" json:"job_title"`
	DepartmentID   int64          `gorm:"index;default:0" json:"department_id"`
	OfficeLocation string         `gorm:"size:128" json:"office_location"`
	HireDate       time.Time      `json:"hire_date"`
	Bio            string         `gorm:"size:512" json:"bio"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Employee) TableName() string {
	return "employees"
}

type DirectoryUser struct {
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	Nickname       string    `json:"nickname"`
	AvatarURL      string    `json:"avatar_url"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Status         int8      `json:"status"`
	DepartmentID   int64     `json:"department_id"`
	DepartmentName string    `json:"department_name"`
	EmployeeID     int64     `json:"employee_id"`
	EmployeeNo     string    `json:"employee_no"`
	JobTitle       string    `json:"job_title"`
	OfficeLocation string    `json:"office_location"`
	Bio            string    `json:"bio"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
