package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	TraceIDHeader     = "X-Trace-ID"
	RequestIDHeader   = "X-Request-ID"
	TraceParentHeader = "traceparent"
)

type traceIDContextKey struct{}

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "workpal",
			Name:      "http_requests_total",
			Help:      "WorkPal 服务处理的 HTTP 请求总数。",
		},
		[]string{"service", "method", "route", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "workpal",
			Name:      "http_request_duration_seconds",
			Help:      "按服务、方法、路由和状态码统计的 HTTP 请求耗时，单位为秒。",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "route", "status"},
	)
	httpInFlightRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "workpal",
			Name:      "http_in_flight_requests",
			Help:      "按服务和方法统计的当前处理中 HTTP 请求数。",
		},
		[]string{"service", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpInFlightRequests)
}

func ConfigureLogger(serviceName string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler).With("service", serviceName)
	slog.SetDefault(logger)

	stdLogger := slog.NewLogLogger(logger.Handler(), slog.LevelInfo)
	log.SetFlags(0)
	log.SetOutput(stdLogger.Writer())
}

func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID, _ := ctx.Value(traceIDContextKey{}).(string)
	return traceID
}

func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := traceIDFromTraceParent(c.GetHeader(TraceParentHeader))
		if traceID == "" {
			traceID = c.GetHeader(TraceIDHeader)
		}
		if traceID == "" {
			traceID = c.GetHeader(RequestIDHeader)
		}
		if traceID == "" {
			traceID = newTraceID()
		}

		c.Set("trace_id", traceID)
		c.Set("requestID", traceID)
		c.Header(TraceIDHeader, traceID)
		c.Header(RequestIDHeader, traceID)
		if isW3CTraceID(traceID) {
			c.Header(TraceParentHeader, "00-"+traceID+"-0000000000000001-01")
		}

		ctx := context.WithValue(c.Request.Context(), traceIDContextKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RequestMetrics(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		httpInFlightRequests.WithLabelValues(serviceName, method).Inc()
		start := time.Now()
		defer func() {
			httpInFlightRequests.WithLabelValues(serviceName, method).Dec()

			status := strconv.Itoa(c.Writer.Status())
			route := routeLabel(c)
			httpRequestsTotal.WithLabelValues(serviceName, method, route, status).Inc()
			httpRequestDuration.WithLabelValues(serviceName, method, route, status).Observe(time.Since(start).Seconds())
		}()

		c.Next()
	}
}

func StructuredAccessLog(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		attrs := []slog.Attr{
			slog.String("trace_id", TraceIDFromContext(c.Request.Context())),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", routeLabel(c)),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		}
		if userID, ok := c.Get("userID"); ok {
			attrs = append(attrs, slog.Any("user_id", userID))
		}
		if upstream := c.Writer.Header().Get("X-Upstream-Service"); upstream != "" {
			attrs = append(attrs, slog.String("upstream_service", upstream))
		}

		slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "HTTP请求", attrs...)
	}
}

func routeLabel(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}
	if gatewayRoute := c.Writer.Header().Get("X-Gateway-Route"); gatewayRoute != "" {
		return gatewayRoute
	}
	if c.Request != nil && c.Request.URL != nil {
		return c.Request.URL.Path
	}
	return "unknown"
}

func newTraceID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf[:])
}

func traceIDFromTraceParent(value string) string {
	parts := strings.Split(value, "-")
	if len(parts) < 4 {
		return ""
	}
	traceID := strings.ToLower(parts[1])
	if !isW3CTraceID(traceID) {
		return ""
	}
	return traceID
}

func isW3CTraceID(value string) bool {
	if len(value) != 32 || value == "00000000000000000000000000000000" {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}
	return true
}
