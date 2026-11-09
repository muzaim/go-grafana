package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GinMetricsMiddleware())
	r.GET("/metrics", PrometheusHandler())
	r.GET("/test-ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// 1. Hit endpoint to trigger metric counter
	req, _ := http.NewRequest(http.MethodGet, "/test-ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2. Query /metrics
	req, _ = http.NewRequest(http.MethodGet, "/metrics", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "uptime_http_requests_total")
}
