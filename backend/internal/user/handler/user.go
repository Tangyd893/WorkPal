package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/user/service"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/common/pagination"

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

// Register 用户注册
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

// Login 用户登录
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
		"token":       resp.Token,
		"expires_at":  resp.ExpiresAt,
		"user_id":     resp.User.ID,
		"username":    resp.User.Username,
		"nickname":    resp.User.Nickname,
	})
}

// GetMe 获取当前登录用户信息
// GET /api/v1/users/me
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetInt64("userID")
	user, err := h.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.Success(c, user)
}

// UpdateMe 更新当前用户资料
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

// ListUsers 获取用户列表
// GET /api/v1/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	page, pageSize = pagination.GetParams(page, pageSize)

	users, total, err := h.userSvc.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		handleServiceErr(c, err)
		return
	}
	response.SuccessPage(c, users, total, page, pageSize)
}

// RegisterRoutes 注册用户相关路由
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	// 公开路由（无需认证）
	rg.POST("/auth/register", h.Register)
	rg.POST("/auth/login", h.Login)

	// 需要认证的路由
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())
	auth.GET("/users/me", h.GetMe)
	auth.PUT("/users/me", h.UpdateMe)
	auth.GET("/users", h.ListUsers)
}

// handleServiceErr 将 service 层错误转换为 HTTP 响应
func handleServiceErr(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		response.Fail(c, appErr)
		return
	}
	// 未知错误，笼统处理
	response.FailWithMessage(c, http.StatusInternalServerError, "内部错误: "+err.Error())
}
