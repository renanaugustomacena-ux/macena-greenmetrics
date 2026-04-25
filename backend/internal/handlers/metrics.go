package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler exposes /api/internal/metrics in Prometheus text format.
//
// We use the net/http/httptest.ResponseRecorder from the standard library so
// that promhttp.Handler is driven by a true http.ResponseWriter, then we copy
// the resulting buffer into the Fiber response.
func MetricsHandler(c *fiber.Ctx) error {
	handler := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	handler.ServeHTTP(rec, req)

	// Propagate content-type from the recorder (promhttp picks text or proto).
	ct := rec.Header().Get("Content-Type")
	if ct == "" {
		ct = "text/plain; version=0.0.4"
	}
	c.Set("Content-Type", ct)

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(rec.Body)
	return c.Status(rec.Code).Send(buf.Bytes())
}
