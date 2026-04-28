package handler

import (
	"net/http"
	"strconv"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/pagination"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/user/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *service.UserService
	authSvc *service.AuthService
}

func NewUserHandler(userSvc *service.UserService, authSvc *service.AuthService) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		authSvc: authSvc,
	}
}

// Register
// POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.authSvc.Register(c.Request.Context(), &req)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, user)
}

// Login
// POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	resp, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		handleServiceErr(c, err)
		return
	}

	response.Success(c, gin.H{
		"token":      resp.Token,
		"expires_at": resp.ExpiresAt,
		"user_id":    resp.User.ID,
		"username":   resp.User.Username,
		"nickname":   resp.User.Nickname,
	})
}

// GetMe
// GET /api/v1/users/me
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetInt64("userID")
	user, err := h.userSvc.GetDirectoryByID(c.Request.Context(), userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, user)
}

// UpdateMe
// PUT /api/v1/users/me
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req service.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	user, err := h.userSvc.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, user)
}

// ListUsers returns enriched directory data.
// GET /api/v1/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = pagination.GetParams(page, pageSize)

	departmentID, _ := strconv.ParseInt(c.DefaultQuery("department_id", "0"), 10, 64)
	filter := service.DirectoryFilter{
		Query:        c.Query("q"),
		DepartmentID: departmentID,
	}

	users, total, err := h.userSvc.ListDirectoryUsers(c.Request.Context(), page, pageSize, filter)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.SuccessPage(c, users, total, page, pageSize)
}

// SearchUsers keeps a dedicated search route for compatibility.
// GET /api/v1/users/search?q=keyword
func (h *UserHandler) SearchUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = pagination.GetParams(page, pageSize)

	departmentID, _ := strconv.ParseInt(c.DefaultQuery("department_id", "0"), 10, 64)
	filter := service.DirectoryFilter{
		Query:        c.Query("q"),
		DepartmentID: departmentID,
	}

	users, total, err := h.userSvc.ListDirectoryUsers(c.Request.Context(), page, pageSize, filter)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.SuccessPage(c, users, total, page, pageSize)
}

// ListDepartments
// GET /api/v1/departments
func (h *UserHandler) ListDepartments(c *gin.Context) {
	departments, err := h.userSvc.ListDepartments(c.Request.Context())
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, departments)
}

func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/auth/register", h.Register)
	rg.POST("/auth/login", h.Login)

	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/users/me", h.GetMe)
	auth.PUT("/users/me", h.UpdateMe)
	auth.GET("/users", h.ListUsers)
	auth.GET("/users/search", h.SearchUsers)
	auth.GET("/departments", h.ListDepartments)
}

func handleServiceErr(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		response.Fail(c, appErr)
		return
	}
	response.FailWithMessage(c, http.StatusInternalServerError, "内部错误: "+err.Error())
}
