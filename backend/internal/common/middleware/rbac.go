package middleware

import (
	"strconv"

	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/pkg/rbac"
	"github.com/gin-gonic/gin"
)

// RequirePermission 要求路由访问权限
func RequirePermission(engine *rbac.Engine, perm rbac.PermissionCode) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		if err := engine.Check(c.Request.Context(), userID, perm); err != nil {
			response.FailWithMessage(c, 403, err.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireProjectPermission 要求项目级别权限
func RequireProjectPermission(engine *rbac.Engine, perm rbac.PermissionCode) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		projectIDStr := c.Param("id")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			response.FailWithMessage(c, 400, "invalid project id")
			c.Abort()
			return
		}
		if err := engine.CheckProject(c.Request.Context(), userID, projectID, perm); err != nil {
			response.FailWithMessage(c, 403, err.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}
