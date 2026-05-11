package metrics

import (
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Match UUIDs in paths
	uuidPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	// Match numeric IDs in paths
	idPattern = regexp.MustCompile(`/\d+`)
)

// RequestCounter counts HTTP requests by method, path, and status
var RequestCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "pganalytics_http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "path", "status"},
)

// PrometheusMiddleware returns a Gin middleware that records metrics
// for all HTTP requests including response time and request counts.
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics after request completes
		duration := time.Since(start)

		// Normalize path (use gin's FullPath if available, otherwise normalize the URL path)
		path := c.FullPath()
		if path == "" {
			// Fall back to normalizing the actual request path
			path = normalizePath(c.Request.URL.Path)
		}

		// Record to Prometheus
		status := c.Writer.Status()
		RecordAPIResponseTime(c.Request.Method, path, status, duration)

		// Increment request counter
		RequestCounter.WithLabelValues(
			c.Request.Method,
			path,
			strconv.Itoa(status),
		).Inc()
	}
}

// normalizePath replaces UUIDs and numeric IDs with placeholders
// to prevent high cardinality in metrics labels.
func normalizePath(path string) string {
	// Replace UUIDs with :uuid placeholder
	path = uuidPattern.ReplaceAllString(path, ":uuid")
	// Replace numeric IDs with :id placeholder
	path = idPattern.ReplaceAllString(path, "/:id")
	return path
}
