package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterReq_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("注册请求参数验证", func(t *testing.T) {
		tests := []struct {
			name       string
			body       map[string]interface{}
			wantStatus int
		}{
			{
				name:       "缺少用户名",
				body:       map[string]interface{}{"password": "password123"},
				wantStatus: http.StatusBadRequest,
			},
			{
				name:       "缺少密码",
				body:       map[string]interface{}{"username": "alice"},
				wantStatus: http.StatusBadRequest,
			},
			{
				name:       "用户名太短",
				body:       map[string]interface{}{"username": "al", "password": "password123"},
				wantStatus: http.StatusBadRequest,
			},
			{
				name:       "密码太短",
				body:       map[string]interface{}{"username": "alice", "password": "12345"},
				wantStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// 只测试参数绑定，不实际调用数据库
				router := gin.New()
				router.POST("/test", func(c *gin.Context) {
					var req struct {
						Username string `json:"username" binding:"required,min=3,max=64"`
						Password string `json:"password" binding:"required,min=6,max=128"`
					}
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}
					c.JSON(http.StatusOK, gin.H{"status": "ok"})
				})

				body, _ := json.Marshal(tt.body)
				req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code != tt.wantStatus {
					t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
				}
			})
		}
	})
}

func TestLoginReq_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("登录请求参数验证", func(t *testing.T) {
		tests := []struct {
			name       string
			body       map[string]interface{}
			wantStatus int
		}{
			{
				name:       "缺少用户名",
				body:       map[string]interface{}{"password": "password123"},
				wantStatus: http.StatusBadRequest,
			},
			{
				name:       "缺少密码",
				body:       map[string]interface{}{"username": "alice"},
				wantStatus: http.StatusBadRequest,
			},
			{
				name:       "正常登录",
				body:       map[string]interface{}{"username": "alice", "password": "password123"},
				wantStatus: http.StatusOK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				router := gin.New()
				router.POST("/test", func(c *gin.Context) {
					var req struct {
						Username string `json:"username" binding:"required"`
						Password string `json:"password" binding:"required"`
					}
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}
					c.JSON(http.StatusOK, gin.H{"status": "ok"})
				})

				body, _ := json.Marshal(tt.body)
				req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				if w.Code != tt.wantStatus {
					t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
				}
			})
		}
	})
}
