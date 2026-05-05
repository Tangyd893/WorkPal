package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	apperrors "github.com/Tangyd893/WorkPal/backend/internal/common/errors"
	"github.com/Tangyd893/WorkPal/backend/internal/common/response"
	"github.com/Tangyd893/WorkPal/backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

// AuthRequired JWT 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.FailWithStatus(c, 401, 40101, "Authorization 格式错误，应为: Bearer <token>")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.FailWithStatus(c, 401, 40102, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息注入 Context，供后续 Handler 使用
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

const InternalTokenHeader = "X-Internal-Token"

func InternalTokenRequired(expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedToken == "" {
			response.FailWithMessage(c, http.StatusInternalServerError, "internal token is not configured")
			c.Abort()
			return
		}
		if subtle.ConstantTimeCompare([]byte(c.GetHeader(InternalTokenHeader)), []byte(expectedToken)) != 1 {
			response.FailWithMessage(c, http.StatusUnauthorized, "invalid internal token")
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Trace-ID, X-Request-ID, traceparent, Idempotency-Key")
		c.Header("Access-Control-Expose-Headers", "X-Trace-ID, X-Request-ID, traceparent")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// RequestID 为每个请求生成唯一 ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	return auth.NewUUID()
}
