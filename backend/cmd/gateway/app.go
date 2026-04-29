package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	gatewayServiceName = "gateway"
)

type gatewayApp struct {
	services      []*upstreamService
	routes        []*routeSpec
	limiter       *rateLimiter
	healthTimeout time.Duration
}

type upstreamService struct {
	name              string
	baseURL           string
	healthURL         string
	timeout           time.Duration
	retryMaxAttempts  int
	retryBackoff      time.Duration
	supportsWebSocket bool
	breaker           *circuitBreaker
	proxy             *httputil.ReverseProxy
}

type routeSpec struct {
	name        string
	service     *upstreamService
	methods     []string
	exactPath   string
	prefixPath  string
	suffixPath  string
	timeout     time.Duration
	websocket   bool
	description string
}

type routeSummary struct {
	Name             string   `json:"name"`
	Service          string   `json:"service"`
	Methods          []string `json:"methods"`
	Match            string   `json:"match"`
	TimeoutMS        int64    `json:"timeout_ms"`
	RetryMaxAttempts int      `json:"retry_max_attempts"`
	WebSocket        bool     `json:"websocket"`
	Description      string   `json:"description"`
}

type serviceSummary struct {
	Name              string                 `json:"name"`
	BaseURL           string                 `json:"base_url"`
	HealthURL         string                 `json:"health_url"`
	TimeoutMS         int64                  `json:"timeout_ms"`
	RetryMaxAttempts  int                    `json:"retry_max_attempts"`
	RetryBackoffMS    int64                  `json:"retry_backoff_ms"`
	SupportsWebSocket bool                   `json:"supports_websocket"`
	CircuitBreaker    circuitBreakerSnapshot `json:"circuit_breaker"`
}

func newGatewayApp(cfg *config.Config) (*gatewayApp, error) {
	services, err := buildUpstreamServices(cfg)
	if err != nil {
		return nil, err
	}

	routes, err := buildRouteSpecs(services)
	if err != nil {
		return nil, err
	}

	rateWindow := durationFromMS(cfg.Gateway.RateLimit.WindowMS)
	if rateWindow <= 0 {
		rateWindow = time.Minute
	}

	return &gatewayApp{
		services:      services,
		routes:        routes,
		limiter:       newRateLimiter(maxInt(1, cfg.Gateway.RateLimit.Requests), rateWindow),
		healthTimeout: durationFromMS(cfg.Gateway.HealthTimeoutMS),
	}, nil
}

func (a *gatewayApp) Register(r *gin.Engine) {
	healthChecks := make([]platform.HealthCheck, 0, len(a.services))
	healthTimeout := a.healthTimeout
	if healthTimeout <= 0 {
		healthTimeout = 3 * time.Second
	}
	for _, service := range a.services {
		client := &http.Client{Timeout: healthTimeout}
		healthChecks = append(healthChecks, platform.HTTPHealthCheck(service.name, service.healthURL, client))
	}

	r.Use(requestIDMiddleware())
	r.Use(gatewayAccessLog())
	r.Use(rateLimitMiddleware(a.limiter))

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": gatewayServiceName,
		})
	})
	r.GET("/health/ready", platform.HealthHandler(gatewayServiceName, healthChecks...))
	platform.RegisterHealth(r, gatewayServiceName, healthChecks...)

	r.GET("/gateway/routes", a.handleRoutes)
	r.GET("/gateway/services", a.handleServices)
	r.NoRoute(a.handleProxy)
}

func (a *gatewayApp) handleRoutes(c *gin.Context) {
	out := make([]routeSummary, 0, len(a.routes))
	for _, route := range a.routes {
		out = append(out, route.summary())
	}
	c.JSON(http.StatusOK, gin.H{
		"service": gatewayServiceName,
		"routes":  out,
	})
}

func (a *gatewayApp) handleServices(c *gin.Context) {
	out := make([]serviceSummary, 0, len(a.services))
	for _, service := range a.services {
		out = append(out, service.summary())
	}
	c.JSON(http.StatusOK, gin.H{
		"service":  gatewayServiceName,
		"services": out,
	})
}

func (a *gatewayApp) handleProxy(c *gin.Context) {
	route := a.match(c.Request)
	if route == nil {
		writeGatewayError(c.Writer, http.StatusNotFound, "route not found", requestIDFromContext(c.Request.Context()))
		return
	}
	service := route.service
	if !service.breaker.Allow() {
		writeGatewayError(
			c.Writer,
			http.StatusServiceUnavailable,
			fmt.Sprintf("%s circuit breaker is open", service.name),
			requestIDFromContext(c.Request.Context()),
		)
		return
	}

	if !route.websocket {
		ctx, cancel := context.WithTimeout(c.Request.Context(), route.effectiveTimeout())
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
	}

	c.Header("X-Upstream-Service", service.name)
	c.Header("X-Gateway-Route", route.name)
	service.proxy.ServeHTTP(c.Writer, c.Request)
}

func (a *gatewayApp) match(req *http.Request) *routeSpec {
	for _, route := range a.routes {
		if route.matches(req) {
			return route
		}
	}
	return nil
}

func buildUpstreamServices(cfg *config.Config) ([]*upstreamService, error) {
	timeoutCfg := cfg.Gateway.Timeouts
	retryCfg := cfg.Gateway.Retry
	breakerCfg := cfg.Gateway.CircuitBreaker

	specs := []struct {
		name              string
		baseURL           string
		timeoutMS         int
		supportsWebSocket bool
	}{
		{name: "user-service", baseURL: cfg.Services.UserURL, timeoutMS: timeoutCfg.UserMS},
		{name: "im-service", baseURL: cfg.Services.IMURL, timeoutMS: timeoutCfg.IMMS, supportsWebSocket: true},
		{name: "file-service", baseURL: cfg.Services.FileURL, timeoutMS: timeoutCfg.FileMS},
		{name: "search-service", baseURL: cfg.Services.SearchURL, timeoutMS: timeoutCfg.SearchMS},
		{name: "workspace-service", baseURL: cfg.Services.WorkspaceURL, timeoutMS: timeoutCfg.WorkspaceMS},
	}

	services := make([]*upstreamService, 0, len(specs))
	for _, spec := range specs {
		timeout := durationFromMS(spec.timeoutMS)
		if timeout <= 0 {
			timeout = durationFromMS(timeoutCfg.DefaultMS)
		}
		service := &upstreamService{
			name:              spec.name,
			baseURL:           strings.TrimRight(spec.baseURL, "/"),
			healthURL:         serviceHealthURL(spec.baseURL),
			timeout:           timeout,
			retryMaxAttempts:  maxInt(1, retryCfg.MaxAttempts),
			retryBackoff:      durationFromMS(retryCfg.BackoffMS),
			supportsWebSocket: spec.supportsWebSocket,
			breaker: newCircuitBreaker(
				maxInt(1, breakerCfg.FailureThreshold),
				durationFromMS(breakerCfg.CoolDownMS),
			),
		}
		proxy, err := newReverseProxy(service)
		if err != nil {
			return nil, err
		}
		service.proxy = proxy
		services = append(services, service)
	}
	return services, nil
}

func buildRouteSpecs(services []*upstreamService) ([]*routeSpec, error) {
	index := make(map[string]*upstreamService, len(services))
	for _, service := range services {
		index[service.name] = service
	}

	resolve := func(name string) (*upstreamService, error) {
		service, ok := index[name]
		if !ok {
			return nil, fmt.Errorf("service %s not found in gateway registry", name)
		}
		return service, nil
	}

	userService, err := resolve("user-service")
	if err != nil {
		return nil, err
	}
	imService, err := resolve("im-service")
	if err != nil {
		return nil, err
	}
	fileService, err := resolve("file-service")
	if err != nil {
		return nil, err
	}
	searchService, err := resolve("search-service")
	if err != nil {
		return nil, err
	}
	workspaceService, err := resolve("workspace-service")
	if err != nil {
		return nil, err
	}

	return []*routeSpec{
		{
			name:        "gateway-websocket",
			service:     imService,
			methods:     []string{http.MethodGet},
			exactPath:   "/ws",
			websocket:   true,
			description: "WebSocket upgrade entry for the IM service.",
		},
		{
			name:        "user-auth",
			service:     userService,
			prefixPath:  "/api/v1/auth",
			description: "Authentication and token issuance endpoints.",
		},
		{
			name:        "user-directory",
			service:     userService,
			prefixPath:  "/api/v1/users",
			description: "Current user, directory, and profile endpoints.",
		},
		{
			name:        "user-departments",
			service:     userService,
			prefixPath:  "/api/v1/departments",
			description: "Department listing endpoints.",
		},
		{
			name:        "file-direct",
			service:     fileService,
			prefixPath:  "/api/v1/files",
			description: "Personal file upload, delete, and share endpoints.",
		},
		{
			name:        "file-conversation",
			service:     fileService,
			prefixPath:  "/api/v1/conversations/",
			suffixPath:  "/files",
			description: "Conversation file listing endpoints.",
		},
		{
			name:        "search-messages",
			service:     searchService,
			prefixPath:  "/api/v1/search",
			description: "Message search endpoints backed by Bleve.",
		},
		{
			name:        "workspace-tasks",
			service:     workspaceService,
			prefixPath:  "/api/v1/tasks",
			description: "Task CRUD and share endpoints.",
		},
		{
			name:        "workspace-schedule",
			service:     workspaceService,
			prefixPath:  "/api/v1/schedule",
			description: "Schedule CRUD and share endpoints.",
		},
		{
			name:        "im-conversations",
			service:     imService,
			prefixPath:  "/api/v1/conversations",
			description: "Conversation, membership, message, and announcement endpoints.",
		},
		{
			name:        "im-messages",
			service:     imService,
			prefixPath:  "/api/v1/messages",
			description: "Dedicated message endpoints.",
		},
	}, nil
}

func newReverseProxy(service *upstreamService) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(service.baseURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if requestID := requestIDFromContext(req.Context()); requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
		}
		req.Header.Set("X-Forwarded-Proto", forwardedProto(req))
		req.Header.Set("X-Forwarded-Host", req.Host)
	}
	proxy.Transport = newRetryTransport(http.DefaultTransport, service.retryMaxAttempts, service.retryBackoff)
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Set("X-Upstream-Service", service.name)
		if resp.StatusCode >= http.StatusInternalServerError {
			service.breaker.OnFailure()
			return nil
		}
		service.breaker.OnSuccess()
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		service.breaker.OnFailure()
		log.Printf("gateway proxy error for %s -> %s: %v", r.URL.Path, service.baseURL, err)

		status := http.StatusBadGateway
		switch {
		case r.Context().Err() == context.DeadlineExceeded:
			status = http.StatusGatewayTimeout
		case r.Context().Err() == context.Canceled:
			status = 499
		}
		writeGatewayError(w, status, "upstream service unavailable", requestIDFromContext(r.Context()))
	}
	return proxy, nil
}

func (r routeSpec) matches(req *http.Request) bool {
	if len(r.methods) > 0 && !containsMethod(r.methods, req.Method) {
		return false
	}
	path := req.URL.Path
	if r.exactPath != "" && path != r.exactPath {
		return false
	}
	if r.prefixPath != "" && !strings.HasPrefix(path, r.prefixPath) {
		return false
	}
	if r.suffixPath != "" && !strings.HasSuffix(path, r.suffixPath) {
		return false
	}
	return true
}

func (r routeSpec) effectiveTimeout() time.Duration {
	if r.websocket {
		return 0
	}
	if r.timeout > 0 {
		return r.timeout
	}
	return r.service.timeout
}

func (r routeSpec) summary() routeSummary {
	return routeSummary{
		Name:             r.name,
		Service:          r.service.name,
		Methods:          append([]string(nil), r.methods...),
		Match:            r.matchExpression(),
		TimeoutMS:        r.effectiveTimeout().Milliseconds(),
		RetryMaxAttempts: r.service.retryMaxAttempts,
		WebSocket:        r.websocket,
		Description:      r.description,
	}
}

func (r routeSpec) matchExpression() string {
	if r.exactPath != "" {
		return r.exactPath
	}
	if r.prefixPath != "" && r.suffixPath != "" {
		return r.prefixPath + "*" + r.suffixPath
	}
	if r.prefixPath != "" {
		return r.prefixPath + "*"
	}
	return "*"
}

func (s upstreamService) summary() serviceSummary {
	return serviceSummary{
		Name:              s.name,
		BaseURL:           s.baseURL,
		HealthURL:         s.healthURL,
		TimeoutMS:         s.timeout.Milliseconds(),
		RetryMaxAttempts:  s.retryMaxAttempts,
		RetryBackoffMS:    s.retryBackoff.Milliseconds(),
		SupportsWebSocket: s.supportsWebSocket,
		CircuitBreaker:    s.breaker.Snapshot(),
	}
}

type requestIDKey struct{}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Header("X-Request-ID", requestID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDKey{}, requestID))
		c.Next()
	}
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

func gatewayAccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf(
			"[gateway] request_id=%s method=%s path=%s status=%d latency=%s client=%s upstream=%s route=%s",
			requestIDFromContext(c.Request.Context()),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			clientIP(c.Request),
			c.Writer.Header().Get("X-Upstream-Service"),
			c.Writer.Header().Get("X-Gateway-Route"),
		)
	}
}

func writeGatewayError(w http.ResponseWriter, status int, message string, requestID string) {
	if status < http.StatusContinue || status > 999 {
		status = http.StatusBadGateway
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(
		w,
		`{"code":%d,"message":%q,"data":{"request_id":%q}}`,
		status,
		message,
		requestID,
	)
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func forwardedProto(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func serviceHealthURL(baseURL string) string {
	return strings.TrimRight(baseURL, "/") + "/health"
}

func durationFromMS(ms int) time.Duration {
	if ms <= 0 {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}

func containsMethod(methods []string, method string) bool {
	for _, candidate := range methods {
		if strings.EqualFold(candidate, method) {
			return true
		}
	}
	return false
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
