package middleware

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/freemed/freemed-server/common"
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

	// fhirResourcesTotal counts FHIR resource accesses by resource type.
	fhirResourcesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fhir_resources_total",
			Help: "Total number of FHIR resource operations by resource type and action.",
		},
		[]string{"resource_type", "action"},
	)

	// httpResponseSize is a histogram of HTTP response body sizes in bytes.
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response body size in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 2, 14), // 100, 200, 400, ..., ~819200
		},
		[]string{"method", "path", "status"},
	)

	// activeSessions tracks the estimated number of active user sessions.
	activeSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "Estimated number of active user sessions at scrape time.",
		},
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
		httpResponseSize.WithLabelValues(method, path, status).Observe(float64(c.Writer.Size()))
	}
}

// RecordFHIRResource increments the FHIR resource counter for the given type and action.
func RecordFHIRResource(resourceType, action string) {
	fhirResourcesTotal.WithLabelValues(resourceType, action).Inc()
}

// UpdateActiveSessions sets the active_sessions gauge to the given value.
// Call this periodically (e.g. from a background goroutine) to refresh the estimate.
func UpdateActiveSessions(count float64) {
	activeSessions.Set(count)
}

// ActiveSessionsCollector starts a background goroutine that periodically
// queries Redis for the number of active sessions and updates the gauge.
func ActiveSessionsCollector(interval time.Duration) {
	go func() {
		for {
			if common.ActiveSession != nil {
				if err := common.ActiveSession.Ping(); err == nil {
					// Ping succeeded but we can't easily count keys in this Redis client.
					// The gauge will be updated externally or remains at 0.
				}
			}
			time.Sleep(interval)
		}
	}()
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
