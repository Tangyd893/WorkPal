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
	"sync/atomic"
	"time"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	gatewayServiceName = "gateway"
)

type gatewayApp struct {
	services      []*upstreamService
	routes        []*routeSpec
	limiter       *rateLimiter
	healthTimeout time.Duration
	registry      *platform.ServiceRegistry
	registryStop  context.CancelFunc
	registryRedis *redis.Client
}

type upstreamService struct {
	name              string
	cfg               *config.Config
	registryClient    *redis.Client
	fallbackBaseURL   string
	fallbackHealthURL string
	timeout           time.Duration
	retryMaxAttempts  int
	retryBackoff      time.Duration
	supportsWebSocket bool
	breaker           *circuitBreaker
	proxy             *httputil.ReverseProxy
	requestCounter    uint64
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
	DiscoveryMode     string                 `json:"discovery_mode"`
	DiscoveredCount   int                    `json:"discovered_count"`
	Instances         []serviceInstanceInfo  `json:"instances"`
	TimeoutMS         int64                  `json:"timeout_ms"`
	RetryMaxAttempts  int                    `json:"retry_max_attempts"`
	RetryBackoffMS    int64                  `json:"retry_backoff_ms"`
	SupportsWebSocket bool                   `json:"supports_websocket"`
	CircuitBreaker    circuitBreakerSnapshot `json:"circuit_breaker"`
}

type serviceInstanceInfo struct {
	ID        string            `json:"id"`
	BaseURL   string            `json:"base_url"`
	HealthURL string            `json:"health_url"`
	Version   string            `json:"version"`
	UpdatedAt string            `json:"updated_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type upstreamRequestInfo struct {
	service       string
	baseURL       string
	instanceID    string
	discoveryMode string
}

func newGatewayApp(cfg *config.Config) (*gatewayApp, error) {
	var registryClient *redis.Client
	var registry *platform.ServiceRegistry
	var registryStop context.CancelFunc
	if cfg.Registry.Enabled {
		client, err := platform.OpenRedis(cfg)
		if err != nil {
			log.Printf("[gateway] service registry unavailable, falling back to static upstream catalog: %v", err)
		} else {
			registryClient = client
			reg, stop, err := platform.StartServiceRegistration(cfg, client, gatewayServiceName, map[string]string{
				"role": "ingress",
			})
			if err != nil {
				log.Printf("[gateway] register gateway instance: %v", err)
			} else {
				registry = reg
				registryStop = stop
			}
		}
	}

	services, err := buildUpstreamServices(cfg, registryClient)
	if err != nil {
		if registryStop != nil {
			registryStop()
		}
		if registryClient != nil {
			_ = registryClient.Close()
		}
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
		registry:      registry,
		registryStop:  registryStop,
		registryRedis: registryClient,
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
		healthChecks = append(healthChecks, platform.NamedHealthCheck(service.name, func(ctx context.Context) error {
			return service.checkHealth(ctx, client)
		}))
	}

	r.Use(requestIDMiddleware())
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

func (a *gatewayApp) Shutdown() {
	if a.registry != nil {
		_ = a.registry.Deregister(context.Background())
	}
	if a.registryStop != nil {
		a.registryStop()
	}
	if a.registryRedis != nil {
		_ = a.registryRedis.Close()
	}
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
		out = append(out, service.summary(c.Request.Context()))
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

func buildUpstreamServices(cfg *config.Config, registryClient *redis.Client) ([]*upstreamService, error) {
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
		failureThreshold := maxInt(1, breakerCfg.FailureThreshold)
		coolDown := durationFromMS(breakerCfg.CoolDownMS)
		if spec.name == "search-service" {
			failureThreshold = 5
			coolDown = 30 * time.Second
		}
		service := &upstreamService{
			name:              spec.name,
			cfg:               cfg,
			registryClient:    registryClient,
			fallbackBaseURL:   strings.TrimRight(spec.baseURL, "/"),
			fallbackHealthURL: serviceHealthURL(spec.baseURL),
			timeout:           timeout,
			retryMaxAttempts:  maxInt(1, retryCfg.MaxAttempts),
			retryBackoff:      durationFromMS(retryCfg.BackoffMS),
			supportsWebSocket: spec.supportsWebSocket,
			breaker: newCircuitBreaker(
				failureThreshold,
				coolDown,
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
	target, err := url.Parse(service.fallbackBaseURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		incomingHost := req.Host
		targetInstance, discoveryMode, resolveErr := service.resolveInstance(req.Context())
		targetURL := service.fallbackBaseURL
		if resolveErr != nil {
			log.Printf("[gateway] resolve upstream %s: %v", service.name, resolveErr)
		} else if targetInstance.BaseURL != "" {
			targetURL = targetInstance.BaseURL
		}

		target, err := url.Parse(targetURL)
		if err == nil {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			if target.Path != "" {
				req.URL.Path = joinURLPath(target.Path, req.URL.Path)
			}
		}
		ctx := context.WithValue(req.Context(), upstreamRequestKey{}, upstreamRequestInfo{
			service:       service.name,
			baseURL:       targetURL,
			instanceID:    targetInstance.ID,
			discoveryMode: discoveryMode,
		})
		*req = *req.WithContext(ctx)
		if requestID := requestIDFromContext(req.Context()); requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
			req.Header.Set(platform.TraceIDHeader, requestID)
		}
		if traceParent := req.Header.Get(platform.TraceParentHeader); traceParent != "" {
			req.Header.Set(platform.TraceParentHeader, traceParent)
		}
		req.Header.Set("X-Forwarded-Proto", forwardedProto(req))
		req.Header.Set("X-Forwarded-Host", incomingHost)
	}
	proxy.Transport = newRetryTransport(http.DefaultTransport, service.retryMaxAttempts, service.retryBackoff)
	proxy.ModifyResponse = func(resp *http.Response) error {
		upstreamInfo := upstreamRequestFromContext(resp.Request.Context())
		resp.Header.Set("X-Upstream-Service", service.name)
		if upstreamInfo.instanceID != "" {
			resp.Header.Set("X-Upstream-Instance", upstreamInfo.instanceID)
		}
		if upstreamInfo.discoveryMode != "" {
			resp.Header.Set("X-Upstream-Discovery", upstreamInfo.discoveryMode)
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			service.breaker.OnFailure()
			return nil
		}
		service.breaker.OnSuccess()
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		service.breaker.OnFailure()
		log.Printf("gateway proxy error for %s -> %s: %v", r.URL.Path, upstreamRequestFromContext(r.Context()).baseURL, err)

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

func (s *upstreamService) fallbackInstance() platform.ServiceInstance {
	return platform.ServiceInstance{
		ID:        "static",
		Service:   s.name,
		BaseURL:   s.fallbackBaseURL,
		HealthURL: s.fallbackHealthURL,
		Version:   platform.Version,
	}
}

func (s *upstreamService) resolveDiscoveredInstances(ctx context.Context) ([]platform.ServiceInstance, error) {
	if s.registryClient == nil || s.cfg == nil || !s.cfg.Registry.Enabled {
		return []platform.ServiceInstance{}, nil
	}
	return platform.ListServiceInstances(ctx, s.cfg, s.registryClient, s.name)
}

func (s *upstreamService) resolveInstance(ctx context.Context) (platform.ServiceInstance, string, error) {
	instances, err := s.resolveDiscoveredInstances(ctx)
	if err != nil {
		return s.fallbackInstance(), "static", err
	}
	if len(instances) == 0 {
		return s.fallbackInstance(), "static", nil
	}
	index := atomic.AddUint64(&s.requestCounter, 1)
	selected := instances[(int(index)-1)%len(instances)]
	if selected.HealthURL == "" {
		selected.HealthURL = serviceHealthURL(selected.BaseURL)
	}
	return selected, "registry", nil
}

func (s *upstreamService) checkHealth(ctx context.Context, client *http.Client) error {
	instance, _, err := s.resolveInstance(ctx)
	if err != nil {
		log.Printf("[gateway] service discovery degraded for %s, falling back to health URL %s: %v", s.name, s.fallbackHealthURL, err)
	}
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, instance.HealthURL, nil)
	if reqErr != nil {
		return reqErr
	}
	resp, doErr := client.Do(req)
	if doErr != nil {
		return doErr
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("health endpoint returned HTTP %d", resp.StatusCode)
	}
	return nil
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

func (s *upstreamService) summary(ctx context.Context) serviceSummary {
	instances, err := s.resolveDiscoveredInstances(ctx)
	discoveryMode := "static"
	if len(instances) > 0 {
		discoveryMode = "registry"
	}
	if err != nil {
		log.Printf("[gateway] service summary lookup failed for %s: %v", s.name, err)
		instances = nil
	}

	instanceSummaries := make([]serviceInstanceInfo, 0, maxInt(1, len(instances)))
	if len(instances) == 0 {
		fallback := s.fallbackInstance()
		instanceSummaries = append(instanceSummaries, serviceInstanceInfo{
			ID:        fallback.ID,
			BaseURL:   fallback.BaseURL,
			HealthURL: fallback.HealthURL,
			Version:   fallback.Version,
			UpdatedAt: fallback.UpdatedAt,
		})
	} else {
		for _, instance := range instances {
			instanceSummaries = append(instanceSummaries, serviceInstanceInfo{
				ID:        instance.ID,
				BaseURL:   instance.BaseURL,
				HealthURL: instance.HealthURL,
				Version:   instance.Version,
				UpdatedAt: instance.UpdatedAt,
				Metadata:  instance.Metadata,
			})
		}
	}

	return serviceSummary{
		Name:              s.name,
		BaseURL:           s.fallbackBaseURL,
		HealthURL:         s.fallbackHealthURL,
		DiscoveryMode:     discoveryMode,
		DiscoveredCount:   len(instances),
		Instances:         instanceSummaries,
		TimeoutMS:         s.timeout.Milliseconds(),
		RetryMaxAttempts:  s.retryMaxAttempts,
		RetryBackoffMS:    s.retryBackoff.Milliseconds(),
		SupportsWebSocket: s.supportsWebSocket,
		CircuitBreaker:    s.breaker.Snapshot(),
	}
}

type requestIDKey struct{}
type upstreamRequestKey struct{}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := platform.TraceIDFromContext(c.Request.Context())
		if requestID == "" {
			requestID = c.GetHeader(platform.TraceIDHeader)
		}
		if requestID == "" {
			requestID = c.GetHeader("X-Request-ID")
		}
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Header("X-Request-ID", requestID)
		c.Header(platform.TraceIDHeader, requestID)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDKey{}, requestID))
		c.Next()
	}
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	if requestID == "" {
		requestID = platform.TraceIDFromContext(ctx)
	}
	return requestID
}

func upstreamRequestFromContext(ctx context.Context) upstreamRequestInfo {
	info, _ := ctx.Value(upstreamRequestKey{}).(upstreamRequestInfo)
	return info
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

func joinURLPath(basePath, requestPath string) string {
	baseHasSlash := strings.HasSuffix(basePath, "/")
	requestHasSlash := strings.HasPrefix(requestPath, "/")
	switch {
	case baseHasSlash && requestHasSlash:
		return basePath + requestPath[1:]
	case !baseHasSlash && !requestHasSlash:
		return basePath + "/" + requestPath
	default:
		return basePath + requestPath
	}
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
