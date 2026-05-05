package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteMatchingAndCatalog(t *testing.T) {
	cfg := testGatewayConfig()
	app, err := newGatewayApp(cfg)
	require.NoError(t, err)

	tests := []struct {
		name          string
		method        string
		path          string
		wantRoute     string
		wantService   string
		wantTimeout   time.Duration
		wantWebSocket bool
	}{
		{name: "auth", method: http.MethodPost, path: "/api/v1/auth/login", wantRoute: "user-auth", wantService: "user-service", wantTimeout: 5 * time.Second},
		{name: "im ws", method: http.MethodGet, path: "/ws", wantRoute: "gateway-websocket", wantService: "im-service", wantWebSocket: true},
		{name: "conv files", method: http.MethodGet, path: "/api/v1/conversations/12/files", wantRoute: "file-conversation", wantService: "file-service", wantTimeout: 20 * time.Second},
		{name: "tasks", method: http.MethodDelete, path: "/api/v1/tasks/1", wantRoute: "workspace-tasks", wantService: "workspace-service", wantTimeout: 8 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			require.NoError(t, err)
			route := app.match(req)
			require.NotNil(t, route)
			assert.Equal(t, tt.wantRoute, route.name)
			assert.Equal(t, tt.wantService, route.service.name)
			assert.Equal(t, tt.wantTimeout, route.effectiveTimeout())
			assert.Equal(t, tt.wantWebSocket, route.websocket)
		})
	}
}

func TestCircuitBreakerTransitions(t *testing.T) {
	breaker := newCircuitBreaker(2, 5*time.Second)
	now := time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC)
	breaker.now = func() time.Time { return now }

	assert.True(t, breaker.Allow())
	breaker.OnFailure()
	assert.Equal(t, breakerClosed, breaker.Snapshot().State)

	assert.True(t, breaker.Allow())
	breaker.OnFailure()
	assert.Equal(t, breakerOpen, breaker.Snapshot().State)
	assert.False(t, breaker.Allow())

	now = now.Add(6 * time.Second)
	assert.True(t, breaker.Allow())
	assert.Equal(t, breakerHalfOpen, breaker.Snapshot().State)
	breaker.OnSuccess()
	assert.Equal(t, breakerClosed, breaker.Snapshot().State)
}

func TestSearchServiceUsesThirtySecondCircuitBreaker(t *testing.T) {
	app, err := newGatewayApp(testGatewayConfig())
	require.NoError(t, err)

	var search *upstreamService
	for _, service := range app.services {
		if service.name == "search-service" {
			search = service
			break
		}
	}
	require.NotNil(t, search)

	snapshot := search.breaker.Snapshot()
	assert.Equal(t, 5, snapshot.FailureThreshold)
	assert.Equal(t, int64(30000), snapshot.CoolDownMS)
}

func TestRetryTransportRetriesReadRequestsOnGatewayErrors(t *testing.T) {
	base := &stubRoundTripper{
		responses: []*http.Response{
			{StatusCode: http.StatusBadGateway, Header: make(http.Header), Body: http.NoBody},
			{StatusCode: http.StatusOK, Header: make(http.Header), Body: http.NoBody},
		},
	}
	transport := newRetryTransport(base, 2, 0).(*retryTransport)
	transport.sleep = func(time.Duration) {}

	req, err := http.NewRequest(http.MethodGet, "http://gateway.test/api/v1/tasks", nil)
	require.NoError(t, err)

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, base.calls)
	assert.Equal(t, "2", resp.Header.Get("X-Gateway-Attempts"))
}

func TestReverseProxyForwardsTraceHeaders(t *testing.T) {
	var gotTraceID string
	var gotRequestID string
	var gotTraceParent string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceID = r.Header.Get(platform.TraceIDHeader)
		gotRequestID = r.Header.Get(platform.RequestIDHeader)
		gotTraceParent = r.Header.Get(platform.TraceParentHeader)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	service := &upstreamService{
		name:              "user-service",
		cfg:               &config.Config{},
		fallbackBaseURL:   upstream.URL,
		fallbackHealthURL: upstream.URL + "/health",
		retryMaxAttempts:  1,
		breaker:           newCircuitBreaker(1, time.Second),
	}
	proxy, err := newReverseProxy(service)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), requestIDKey{}, "trace-forward-001"))
	req.Header.Set(platform.TraceParentHeader, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	recorder := httptest.NewRecorder()

	proxy.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, "trace-forward-001", gotTraceID)
	assert.Equal(t, "trace-forward-001", gotRequestID)
	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", gotTraceParent)
}

func TestRateLimiterCleansExpiredBuckets(t *testing.T) {
	limiter := newRateLimiter(1, time.Millisecond)

	assert.True(t, limiter.allow("192.0.2.10"))
	require.Len(t, limiter.visits, 1)

	time.Sleep(2 * time.Millisecond)
	assert.True(t, limiter.allow("192.0.2.11"))

	_, exists := limiter.visits["192.0.2.10"]
	assert.False(t, exists)
}

type stubRoundTripper struct {
	responses []*http.Response
	calls     int
}

func (s *stubRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	idx := s.calls
	s.calls++
	if idx >= len(s.responses) {
		return s.responses[len(s.responses)-1], nil
	}
	return s.responses[idx], nil
}

func testGatewayConfig() *config.Config {
	return &config.Config{
		Services: config.ServicesConfig{
			UserURL:      "http://user-service:8081",
			IMURL:        "http://im-service:8082",
			FileURL:      "http://file-service:8083",
			SearchURL:    "http://search-service:8084",
			WorkspaceURL: "http://workspace-service:8085",
		},
		Gateway: config.GatewayConfig{
			Retry: config.GatewayRetryConfig{
				MaxAttempts: 2,
				BackoffMS:   0,
			},
			CircuitBreaker: config.GatewayCircuitBreakerConfig{
				FailureThreshold: 3,
				CoolDownMS:       5000,
			},
			Timeouts: config.GatewayServiceTimeoutsConfig{
				DefaultMS:   8000,
				UserMS:      5000,
				IMMS:        15000,
				FileMS:      20000,
				SearchMS:    5000,
				WorkspaceMS: 8000,
			},
		},
	}
}
