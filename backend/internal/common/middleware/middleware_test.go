package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInternalTokenRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(InternalTokenRequired("secret-token"))
	router.GET("/internal/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	t.Run("rejects missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/internal/ping", nil)
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("accepts matching token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/internal/ping", nil)
		req.Header.Set(InternalTokenHeader, "secret-token")
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}
