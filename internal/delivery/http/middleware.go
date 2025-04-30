package http

import (
	"pollapp/pkg/monitoring"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// PrometheusMiddleware adds middleware for tracking HTTP metrics with Prometheus
func PrometheusMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			path := req.URL.Path
			method := req.Method

			// Track in-flight requests
			monitoring.HTTPRequestsInFlight.Inc()
			defer monitoring.HTTPRequestsInFlight.Dec()

			// Track request duration
			start := time.Now()

			// Execute the request
			err := next(c)

			// Record metrics after request is complete
			duration := time.Since(start).Seconds()
			status := res.Status

			monitoring.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
			monitoring.HTTPRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()

			return err
		}
	}
}
