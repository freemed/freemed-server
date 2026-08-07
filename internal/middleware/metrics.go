package middleware

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// httpRequestsTotal counts HTTP requests by method, path, and status code.
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestDurationSeconds is a histogram of HTTP request duration in seconds.
	httpRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// dbPoolConnections tracks the database/sql connection pool stats.
	dbPoolConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_pool_connections",
			Help: "Database connection pool metrics.",
		},
		[]string{"state"},
	)
)

// PrometheusMetrics returns a Gin middleware that tracks HTTP request metrics.
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request.
		c.Next()

		// Record metrics after the request has been handled.
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDurationSeconds.WithLabelValues(method, path, status).Observe(duration)
	}
}

// DBConnectionsGauge is an exported function that registers a background goroutine
// to periodically update the db_pool_connections gauge from the provided sql.DB pool.
// Pass nil to skip DB pool metrics.
func DBConnectionsGauge(db *sql.DB) {
	if db == nil {
		return
	}
	go func() {
		for {
			stats := db.Stats()
			dbPoolConnections.WithLabelValues("open").Set(float64(stats.OpenConnections))
			dbPoolConnections.WithLabelValues("idle").Set(float64(stats.Idle))
			dbPoolConnections.WithLabelValues("in_use").Set(float64(stats.InUse))
			dbPoolConnections.WithLabelValues("max_open").Set(float64(stats.MaxOpenConnections))
			dbPoolConnections.WithLabelValues("wait_count").Set(float64(stats.WaitCount))
			dbPoolConnections.WithLabelValues("wait_duration").Set(float64(stats.WaitDuration.Seconds()))
			time.Sleep(15 * time.Second)
		}
	}()
}
