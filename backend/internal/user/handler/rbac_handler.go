package handler

import (
	"net/http"
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/internal/user/service"
	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	rbacSvc *service.RBACService
}

func NewRBACHandler(rbacSvc *service.RBACService) *RBACHandler {
	return &RBACHandler{rbacSvc: rbacSvc}
}

func (h *RBACHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("")
	auth.Use(middleware.AuthRequired())

	auth.GET("/roles", h.ListRoles)
	auth.GET("/permissions", h.ListPermissions)
	auth.POST("/user-roles", h.AssignRole)
	auth.DELETE("/user-roles", h.RemoveRole)
	auth.GET("/users/:id/permissions", h.GetUserPermissions)
	auth.GET("/projects/:id/roles", h.ListProjectRoles)
	auth.POST("/projects/:id/roles", h.CreateProjectRole)
	auth.POST("/projects/:id/members", h.AddProjectMember)
	auth.DELETE("/projects/:id/members/:userID", h.RemoveProjectMember)
	auth.GET("/projects/:id/members", h.ListProjectMembers)
}

func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, err := h.rbacSvc.ListRoles(c.Request.Context())
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, roles)
}

func (h *RBACHandler) ListPermissions(c *gin.Context) {
	perms, err := h.rbacSvc.ListPermissions(c.Request.Context())
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, perms)
}

func (h *RBACHandler) AssignRole(c *gin.Context) {
	var input service.AssignRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.rbacSvc.AssignRole(c.Request.Context(), input); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *RBACHandler) RemoveRole(c *gin.Context) {
	var input service.AssignRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.rbacSvc.RemoveRole(c.Request.Context(), input); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *RBACHandler) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}
	perms, err := h.rbacSvc.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, perms)
}

func (h *RBACHandler) ListProjectRoles(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid project id")
		return
	}
	roles, err := h.rbacSvc.ListProjectRoles(c.Request.Context(), projectID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, roles)
}

func (h *RBACHandler) CreateProjectRole(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid project id")
		return
	}
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	role, err := h.rbacSvc.CreateProjectRole(c.Request.Context(), projectID, input.Name, input.Description)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, role)
}

func (h *RBACHandler) AddProjectMember(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid project id")
		return
	}
	var input service.AddProjectMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.rbacSvc.AddProjectMember(c.Request.Context(), projectID, input); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *RBACHandler) RemoveProjectMember(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid project id")
		return
	}
	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.rbacSvc.RemoveProjectMember(c.Request.Context(), projectID, userID); err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *RBACHandler) ListProjectMembers(c *gin.Context) {
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage(c, http.StatusBadRequest, "invalid project id")
		return
	}
	members, err := h.rbacSvc.ListProjectMembers(c.Request.Context(), projectID)
	if err != nil {
		response.FailWithMessage(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, members)
}
