package platform

import (
	"context"
	"errors"
	"fmt"
	"time"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/user/model"
	userRepo "github.com/Tangyd893/WorkPal/backend/internal/user/repo"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type developmentDepartmentSeed struct {
	Code        string
	Name        string
	Description string
}

type developmentUserSeed struct {
	Username       string
	Password       string
	Nickname       string
	Email          string
	Phone          string
	EmployeeNo     string
	JobTitle       string
	DepartmentCode string
	OfficeLocation string
	Bio            string
	HireDate       time.Time
}

var developmentDepartmentSeeds = []developmentDepartmentSeed{
	{Code: "PMO", Name: "Program Office", Description: "Delivery planning, acceptance, and stakeholder alignment."},
	{Code: "OPS", Name: "Operations", Description: "Launch sequencing, service readiness, and coordination."},
	{Code: "ENG", Name: "Engineering", Description: "Backend delivery, integration, and runtime verification."},
	{Code: "DES", Name: "Design", Description: "Product quality, UX polish, and release review."},
}

var developmentUserSeeds = []developmentUserSeed{
	{
		Username:       "admin",
		Password:       "admin123",
		Nickname:       "Administrator",
		Email:          "admin@workpal.local",
		Phone:          "13800000000",
		EmployeeNo:     "WP-0001",
		JobTitle:       "Workspace Owner",
		DepartmentCode: "PMO",
		OfficeLocation: "Shanghai HQ",
		Bio:            "Coordinates acceptance scope, seeded accounts, and release walkthroughs.",
		HireDate:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.Local),
	},
	{
		Username:       "emma.chen",
		Password:       "workpal123",
		Nickname:       "Emma Chen",
		Email:          "emma.chen@workpal.local",
		Phone:          "13800000001",
		EmployeeNo:     "WP-0207",
		JobTitle:       "Operations Lead",
		DepartmentCode: "OPS",
		OfficeLocation: "Shanghai HQ",
		Bio:            "Owns launch sequencing, announcements, and cross-team follow-through.",
		HireDate:       time.Date(2024, 3, 4, 0, 0, 0, 0, time.Local),
	},
	{
		Username:       "liam.wang",
		Password:       "workpal123",
		Nickname:       "Liam Wang",
		Email:          "liam.wang@workpal.local",
		Phone:          "13800000002",
		EmployeeNo:     "WP-0311",
		JobTitle:       "Platform Engineer",
		DepartmentCode: "ENG",
		OfficeLocation: "Hangzhou Lab",
		Bio:            "Focuses on API stability, WebSocket delivery, and runtime verification.",
		HireDate:       time.Date(2024, 5, 20, 0, 0, 0, 0, time.Local),
	},
	{
		Username:       "sofia.zhao",
		Password:       "workpal123",
		Nickname:       "Sofia Zhao",
		Email:          "sofia.zhao@workpal.local",
		Phone:          "13800000003",
		EmployeeNo:     "WP-0418",
		JobTitle:       "Product Designer",
		DepartmentCode: "DES",
		OfficeLocation: "Shenzhen Studio",
		Bio:            "Reviews interface quality, bilingual polish, and release readiness.",
		HireDate:       time.Date(2024, 7, 8, 0, 0, 0, 0, time.Local),
	},
}

func EnsureDevelopmentUsers(ctx context.Context, db *gorm.DB, userRepoInst *userRepo.UserRepo) error {
	departments, err := ensureDevelopmentDepartments(ctx, db)
	if err != nil {
		return err
	}

	for _, seed := range developmentUserSeeds {
		department, ok := departments[seed.DepartmentCode]
		if !ok {
			return fmt.Errorf("development department %s not found", seed.DepartmentCode)
		}

		if err := ensureDevelopmentUser(ctx, db, userRepoInst, department, seed); err != nil {
			return err
		}
	}

	return nil
}

func DevelopmentUserCount() int {
	return len(developmentUserSeeds)
}

func ensureDevelopmentDepartments(ctx context.Context, db *gorm.DB) (map[string]*model.Department, error) {
	result := make(map[string]*model.Department, len(developmentDepartmentSeeds))

	for _, seed := range developmentDepartmentSeeds {
		var department model.Department
		err := db.WithContext(ctx).Where("code = ?", seed.Code).First(&department).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}

			department = model.Department{
				Code:        seed.Code,
				Name:        seed.Name,
				Description: seed.Description,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := db.WithContext(ctx).Create(&department).Error; err != nil {
				return nil, err
			}
		} else {
			department.Name = seed.Name
			department.Description = seed.Description
			department.UpdatedAt = time.Now()
			if err := db.WithContext(ctx).Save(&department).Error; err != nil {
				return nil, err
			}
		}

		departmentCopy := department
		result[seed.Code] = &departmentCopy
	}

	return result, nil
}

func ensureDevelopmentUser(
	ctx context.Context,
	db *gorm.DB,
	userRepoInst *userRepo.UserRepo,
	department *model.Department,
	seed developmentUserSeed,
) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(seed.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash development password for %s: %w", seed.Username, err)
	}

	existingUser, err := userRepoInst.GetByUsername(ctx, seed.Username)
	if err != nil {
		if !apperrors.Is(err, apperrors.ErrUserNotFound) {
			return err
		}

		existingUser = &model.User{
			Username:     seed.Username,
			PasswordHash: string(passwordHash),
			Nickname:     seed.Nickname,
			Email:        seed.Email,
			Phone:        seed.Phone,
			Status:       1,
			DepartmentID: department.ID,
		}
		if err := userRepoInst.Create(ctx, existingUser); err != nil {
			return err
		}
	} else {
		existingUser.PasswordHash = string(passwordHash)
		existingUser.Nickname = seed.Nickname
		existingUser.Email = seed.Email
		existingUser.Phone = seed.Phone
		existingUser.Status = 1
		existingUser.DepartmentID = department.ID
		if err := userRepoInst.Update(ctx, existingUser); err != nil {
			return err
		}
	}

	return ensureDevelopmentEmployee(ctx, db, existingUser, department, seed)
}

func ensureDevelopmentEmployee(
	ctx context.Context,
	db *gorm.DB,
	user *model.User,
	department *model.Department,
	seed developmentUserSeed,
) error {
	var employee model.Employee
	err := db.WithContext(ctx).Where("user_id = ?", user.ID).First(&employee).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		employee = model.Employee{
			UserID:         user.ID,
			EmployeeNo:     seed.EmployeeNo,
			JobTitle:       seed.JobTitle,
			DepartmentID:   department.ID,
			OfficeLocation: seed.OfficeLocation,
			HireDate:       seed.HireDate,
			Bio:            seed.Bio,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		return db.WithContext(ctx).Create(&employee).Error
	}

	employee.EmployeeNo = seed.EmployeeNo
	employee.JobTitle = seed.JobTitle
	employee.DepartmentID = department.ID
	employee.OfficeLocation = seed.OfficeLocation
	employee.HireDate = seed.HireDate
	employee.Bio = seed.Bio
	employee.UpdatedAt = time.Now()
	return db.WithContext(ctx).Save(&employee).Error
}

func EnsureRBACSeed(ctx context.Context, db *gorm.DB) error {
	roles := []model.Role{
		{Name: "系统管理员", Description: "拥有系统全部管理权限", IsSystem: true},
		{Name: "用户管理员", Description: "管理用户和部门", IsSystem: true},
		{Name: "全局浏览者", Description: "只读浏览权限", IsSystem: true},
	}
	for _, role := range roles {
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&role).Error; err != nil {
			return err
		}
	}

	perms := []model.Permission{
		{Code: "project:create", Name: "创建项目", ResourceType: "project"},
		{Code: "project:delete", Name: "删除项目", ResourceType: "project"},
		{Code: "project:edit", Name: "编辑项目", ResourceType: "project"},
		{Code: "project:view", Name: "查看项目", ResourceType: "project"},
		{Code: "issue:create", Name: "创建事项", ResourceType: "issue"},
		{Code: "issue:edit", Name: "编辑事项", ResourceType: "issue"},
		{Code: "issue:delete", Name: "删除事项", ResourceType: "issue"},
		{Code: "issue:view", Name: "查看事项", ResourceType: "issue"},
		{Code: "issue:change_status", Name: "变更事项状态", ResourceType: "issue"},
		{Code: "workflow:manage", Name: "管理工作流", ResourceType: "workflow"},
		{Code: "board:manage", Name: "管理看板", ResourceType: "board"},
		{Code: "member:manage", Name: "管理项目成员", ResourceType: "member"},
		{Code: "user:manage", Name: "管理用户", ResourceType: "user"},
		{Code: "role:manage", Name: "管理角色", ResourceType: "role"},
	}
	for _, p := range perms {
		if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&p).Error; err != nil {
			return err
		}
	}

	var adminRole model.Role
	if err := db.WithContext(ctx).Where("name = ?", "系统管理员").First(&adminRole).Error; err != nil {
		return err
	}
	var allPerms []model.Permission
	if err := db.WithContext(ctx).Find(&allPerms).Error; err != nil {
		return err
	}
	for _, p := range allPerms {
		rp := model.RolePermission{RoleID: adminRole.ID, PermissionID: p.ID}
		_ = db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rp).Error
	}

	var userManagerRole model.Role
	if err := db.WithContext(ctx).Where("name = ?", "用户管理员").First(&userManagerRole).Error; err != nil {
		return err
	}
	for _, p := range perms {
		if p.Code == "user:manage" || p.Code == "role:manage" {
			rp := model.RolePermission{RoleID: userManagerRole.ID, PermissionID: p.ID}
			_ = db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rp).Error
		}
	}

	var viewerRole model.Role
	if err := db.WithContext(ctx).Where("name = ?", "全局浏览者").First(&viewerRole).Error; err != nil {
		return err
	}
	for _, p := range perms {
		if len(p.Code) > 5 && p.Code[len(p.Code)-5:] == ":view" {
			rp := model.RolePermission{RoleID: viewerRole.ID, PermissionID: p.ID}
			_ = db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rp).Error
		}
	}

	var adminUser model.User
	if err := db.WithContext(ctx).Where("username = ?", "admin").First(&adminUser).Error; err == nil {
		ur := model.UserRole{UserID: adminUser.ID, RoleID: adminRole.ID}
		_ = db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&ur).Error
	}

	return nil
}
