package main

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimiter struct {
	limit       int
	window      time.Duration
	visits      map[string]*rateBucket
	lastCleanup time.Time
	mu          sync.Mutex
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

	l.cleanupExpired(now)

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

func (l *rateLimiter) cleanupExpired(now time.Time) {
	if l.window <= 0 {
		return
	}
	interval := l.window
	if interval > time.Minute {
		interval = time.Minute
	}
	if !l.lastCleanup.IsZero() && now.Sub(l.lastCleanup) < interval {
		return
	}
	for key, bucket := range l.visits {
		if now.After(bucket.expiresAt) {
			delete(l.visits, key)
		}
	}
	l.lastCleanup = now
}

func rateLimitMiddleware(limiter *rateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if exemptFromRateLimit(c.Request.URL.Path) {
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

func exemptFromRateLimit(path string) bool {
	switch {
	case path == "/metrics":
		return true
	case strings.HasPrefix(path, "/health"):
		return true
	case strings.HasPrefix(path, "/gateway/routes"):
		return true
	case strings.HasPrefix(path, "/gateway/services"):
		return true
	default:
		return false
	}
}
