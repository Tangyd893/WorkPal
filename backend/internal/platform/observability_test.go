package platform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouterPropagatesTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(&config.Config{}, "test-service")
	router.GET("/probe", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"trace_id": TraceIDFromContext(c.Request.Context()),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set(TraceIDHeader, "trace-test-123")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "trace-test-123", recorder.Header().Get(TraceIDHeader))
	assert.Equal(t, "trace-test-123", recorder.Header().Get(RequestIDHeader))

	var body map[string]string
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, "trace-test-123", body["trace_id"])
}

func TestNewRouterAcceptsTraceParent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(&config.Config{}, "traceparent-test-service")
	router.GET("/probe", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"trace_id": TraceIDFromContext(c.Request.Context()),
		})
	})

	traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set(TraceParentHeader, "00-"+traceID+"-00f067aa0ba902b7-01")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, traceID, recorder.Header().Get(TraceIDHeader))
	assert.Contains(t, recorder.Header().Get(TraceParentHeader), traceID)
}

func TestNewRouterExposesWorkPalMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(&config.Config{}, "metrics-test-service")
	router.GET("/probe", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/probe", nil))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	body := recorder.Body.String()
	assert.True(t, strings.Contains(body, "workpal_http_requests_total"))
	assert.True(t, strings.Contains(body, "workpal_http_request_duration_seconds"))
	assert.True(t, strings.Contains(body, "workpal_http_in_flight_requests"))
}
