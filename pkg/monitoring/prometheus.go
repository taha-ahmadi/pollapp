package monitoring

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupMonitoringMetrics registers all metrics with Prometheus
func SetupMonitoringMetrics() {
	// HTTP Request metrics
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(HTTPRequestsInFlight)

	// DB metrics
	prometheus.MustRegister(DBQueryDuration)
	prometheus.MustRegister(DBQueriesTotal)

	// Cache metrics
	prometheus.MustRegister(CacheHitsTotal)
	prometheus.MustRegister(CacheMissesTotal)
	prometheus.MustRegister(CacheOperationDuration)

	// Poll metrics
	prometheus.MustRegister(PollCreatedTotal)
	prometheus.MustRegister(VotesReceivedTotal)
}

// MetricsHandler returns an HTTP handler for exposing Prometheus metrics
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// HTTP Metrics
var (
	// HTTPRequestsTotal counts the number of HTTP requests made
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pollapp",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests by method, path, and status code",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration observes the duration of HTTP requests
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pollapp",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds by method and path",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// HTTPRequestsInFlight tracks the number of in-flight HTTP requests
	HTTPRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pollapp",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests being served",
		},
	)
)

// Database Metrics
var (
	// DBQueryDuration observes the duration of database queries
	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pollapp",
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds by operation",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// DBQueriesTotal counts the number of database queries
	DBQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pollapp",
			Subsystem: "db",
			Name:      "queries_total",
			Help:      "Total number of database queries by operation and status",
		},
		[]string{"operation", "status"},
	)
)

// Cache Metrics
var (
	// CacheHitsTotal counts the number of cache hits
	CacheHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pollapp",
			Subsystem: "cache",
			Name:      "hits_total",
			Help:      "Total number of cache hits by operation",
		},
		[]string{"operation"},
	)

	// CacheMissesTotal counts the number of cache misses
	CacheMissesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pollapp",
			Subsystem: "cache",
			Name:      "misses_total",
			Help:      "Total number of cache misses by operation",
		},
		[]string{"operation"},
	)

	// CacheOperationDuration observes the duration of cache operations
	CacheOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "pollapp",
			Subsystem: "cache",
			Name:      "operation_duration_seconds",
			Help:      "Cache operation duration in seconds by operation",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// Poll Metrics
var (
	// PollCreatedTotal counts the number of polls created
	PollCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "pollapp",
			Subsystem: "poll",
			Name:      "created_total",
			Help:      "Total number of polls created",
		},
	)

	// VotesReceivedTotal counts the number of votes received
	VotesReceivedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pollapp",
			Subsystem: "poll",
			Name:      "votes_total",
			Help:      "Total number of votes received by option",
		},
		[]string{"poll_id", "option"},
	)
)

// MeasureDuration is a helper function to measure and record the duration of an operation
func MeasureDuration(histogramVec *prometheus.HistogramVec, labels []string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		histogramVec.WithLabelValues(labels...).Observe(duration)
	}
}
