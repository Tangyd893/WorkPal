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
	"sync"
	"time"

	config "github.com/Tangyd893/WorkPal/backend/configs"
	"github.com/Tangyd893/WorkPal/backend/internal/platform"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	gatewayTimeout = 15 * time.Second
	rateWindow     = time.Minute
	rateLimit      = 180
)

type proxySet struct {
	user      *httputil.ReverseProxy
	im        *httputil.ReverseProxy
	file      *httputil.ReverseProxy
	search    *httputil.ReverseProxy
	workspace *httputil.ReverseProxy
}

func main() {
	cfg, err := platform.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	proxies, err := newProxySet(cfg)
	if err != nil {
		log.Fatalf("create gateway proxies: %v", err)
	}

	limiter := newRateLimiter(rateLimit, rateWindow)
	r := platform.NewRouter(cfg, "gateway")
	r.Use(requestIDMiddleware())
	r.Use(gatewayAccessLog())
	r.Use(rateLimitMiddleware(limiter))
	platform.RegisterHealth(r, nil, nil)
	r.NoRoute(func(c *gin.Context) {
		proxy := proxies.match(c.Request.URL.Path)
		if proxy == nil {
			writeGatewayError(c.Writer, http.StatusNotFound, "route not found", requestIDFromContext(c.Request.Context()))
			return
		}
		if c.Request.URL.Path != "/ws" {
			ctx, cancel := context.WithTimeout(c.Request.Context(), gatewayTimeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	if err := platform.RunHTTP("gateway", cfg.Services.GatewayPort, r, nil); err != nil {
		log.Fatalf("gateway stopped: %v", err)
	}
}

func newProxySet(cfg *config.Config) (*proxySet, error) {
	userProxy, err := newReverseProxy(cfg.Services.UserURL)
	if err != nil {
		return nil, err
	}
	imProxy, err := newReverseProxy(cfg.Services.IMURL)
	if err != nil {
		return nil, err
	}
	fileProxy, err := newReverseProxy(cfg.Services.FileURL)
	if err != nil {
		return nil, err
	}
	searchProxy, err := newReverseProxy(cfg.Services.SearchURL)
	if err != nil {
		return nil, err
	}
	workspaceProxy, err := newReverseProxy(cfg.Services.WorkspaceURL)
	if err != nil {
		return nil, err
	}
	return &proxySet{
		user:      userProxy,
		im:        imProxy,
		file:      fileProxy,
		search:    searchProxy,
		workspace: workspaceProxy,
	}, nil
}

func newReverseProxy(rawURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(rawURL)
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
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("gateway proxy error for %s -> %s: %v", r.URL.Path, rawURL, err)
		status := http.StatusBadGateway
		if r.Context().Err() == context.DeadlineExceeded {
			status = http.StatusGatewayTimeout
		}
		writeGatewayError(w, status, "upstream service unavailable", requestIDFromContext(r.Context()))
	}
	return proxy, nil
}

func (p *proxySet) match(path string) *httputil.ReverseProxy {
	switch {
	case path == "/ws":
		return p.im
	case strings.HasPrefix(path, "/api/v1/auth"):
		return p.user
	case strings.HasPrefix(path, "/api/v1/users"):
		return p.user
	case strings.HasPrefix(path, "/api/v1/departments"):
		return p.user
	case strings.HasPrefix(path, "/api/v1/files"):
		return p.file
	case strings.HasPrefix(path, "/api/v1/conversations/") && strings.HasSuffix(path, "/files"):
		return p.file
	case strings.HasPrefix(path, "/api/v1/search"):
		return p.search
	case strings.HasPrefix(path, "/api/v1/tasks"):
		return p.workspace
	case strings.HasPrefix(path, "/api/v1/schedule"):
		return p.workspace
	case strings.HasPrefix(path, "/api/v1/conversations"):
		return p.im
	case strings.HasPrefix(path, "/api/v1/messages"):
		return p.im
	default:
		return nil
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
			"[gateway] request_id=%s method=%s path=%s status=%d latency=%s client=%s",
			requestIDFromContext(c.Request.Context()),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			clientIP(c.Request),
		)
	}
}

type rateLimiter struct {
	limit  int
	window time.Duration
	visits map[string]*rateBucket
	mu     sync.Mutex
}

type rateBucket struct {
	count     int
	expiresAt time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:  limit,
		window: window,
		visits: make(map[string]*rateBucket),
	}
}

func (l *rateLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, ok := l.visits[key]
	if !ok || now.After(bucket.expiresAt) {
		l.visits[key] = &rateBucket{count: 1, expiresAt: now.Add(l.window)}
		return true
	}
	if bucket.count >= l.limit {
		return false
	}
	bucket.count++
	return true
}

func rateLimitMiddleware(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		if !limiter.allow(clientIP(c.Request)) {
			writeGatewayError(c.Writer, http.StatusTooManyRequests, "too many requests", requestIDFromContext(c.Request.Context()))
			c.Abort()
			return
		}
		c.Next()
	}
}

func writeGatewayError(w http.ResponseWriter, status int, message string, requestID string) {
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
