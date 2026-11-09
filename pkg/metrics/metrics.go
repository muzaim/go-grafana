package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TotalMonitors = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uptime_monitors_total",
		Help: "Total number of registered monitors",
	})

	ActiveMonitors = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "uptime_monitors_active",
		Help: "Number of currently active monitors",
	})

	MonitorStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "uptime_monitor_status",
		Help: "Current status of monitors (1 = UP, 0 = DOWN)",
	}, []string{"monitor_id", "name", "url"})

	ProbeDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "uptime_probe_duration_seconds",
		Help:    "Duration of probe executions in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"protocol", "is_up"})

	HTTPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uptime_http_requests_total",
		Help: "Total number of HTTP requests processed by API",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "uptime_http_request_duration_seconds",
		Help:    "Histogram of response latency for HTTP requests",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})
)

func init() {
	prometheus.MustRegister(
		TotalMonitors,
		ActiveMonitors,
		MonitorStatus,
		ProbeDuration,
		HTTPRequestsTotal,
		HTTPRequestDuration,
	)
}

func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func GinMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
