package main

import (
	"net/http"
	"testing"
	"time"

	config "github.com/Tangyd893/WorkPal/backend/configs"
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
