// Package metrics — process-wide Prometheus registry + custom collectors.
//
// Doctrine refs: Rule 18, Rule 40, Rule 58.
// Cardinality budget: tenant_id only on counters (not histograms); use tier/source on histograms.
//
// Wire from cmd/server/main.go:
//
//   metrics.Register(prometheus.DefaultRegisterer)
//
// Then handlers call metrics.IngestReadings.WithLabelValues(...).Inc() etc.
package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// Custom collectors. Names follow gm_* convention.
var (
	IngestReadings = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gm_ingest_readings_total",
			Help: "Number of readings ingested by source/protocol/result.",
		},
		[]string{"tenant_id", "protocol", "result"},
	)

	IngestQueueDepth = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gm_ingest_queue_depth",
		Help: "Current depth of the bounded ingest channel between sources and the batched DB writer.",
	})

	IngestDropped = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gm_ingest_dropped_total",
			Help: "Number of readings dropped due to queue saturation.",
		},
		[]string{"reason"},
	)

	CAGGRefreshDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gm_cagg_refresh_duration_seconds",
			Help:    "TimescaleDB continuous-aggregate refresh duration.",
			Buckets: prometheus.ExponentialBuckets(0.05, 2, 12),
		},
		[]string{"view"},
	)

	CAGGLastRefreshTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gm_cagg_last_refresh_timestamp_seconds",
			Help: "Unix timestamp of the last successful CAGG refresh, per view.",
		},
		[]string{"view"},
	)

	ReportGenerateDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gm_report_generate_duration_seconds",
			Help:    "Report generation duration per dossier type.",
			Buckets: prometheus.ExponentialBuckets(0.5, 2, 10),
		},
		[]string{"type"},
	)

	AlertFired = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gm_alert_fired_total",
			Help: "Number of alerts fired per rule per tenant.",
		},
		[]string{"rule", "tenant_id"},
	)

	JWTVerifyDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "gm_jwt_verify_duration_seconds",
		Help:    "JWT verify duration.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	BreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gm_breaker_state",
			Help: "Circuit breaker state per upstream (0=closed, 1=open, 2=half-open).",
		},
		[]string{"name"},
	)

	DBPoolAcquireDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "gm_db_pool_acquire_duration_seconds",
		Help:    "pgx pool acquire duration.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
	})

	AsyncJobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gm_async_job_duration_seconds",
			Help:    "Async job duration per type per result.",
			Buckets: prometheus.ExponentialBuckets(0.5, 2, 10),
		},
		[]string{"type", "result"},
	)

	LoginFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gm_login_failed_total",
			Help: "Number of failed login attempts per reason.",
		},
		[]string{"reason"},
	)

	ExternalAPIFallback = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gm_external_api_fallback_total",
			Help: "External API call fell back to cached response.",
		},
		[]string{"api"},
	)

	ConfigValidationFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gm_config_validation_failed_total",
		Help: "Boot-time config validation failures (sentinel rejection, missing env, etc.).",
	})

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gm_http_request_duration_seconds",
			Help:    "HTTP request duration per route.",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gm_http_requests_total",
			Help: "HTTP requests per method per path per status.",
		},
		[]string{"method", "path", "status"},
	)
)

var registerOnce sync.Once

// Register attaches all custom collectors to the given registry. Safe to call once;
// subsequent calls are no-ops.
func Register(r prometheus.Registerer) {
	registerOnce.Do(func() {
		r.MustRegister(
			IngestReadings,
			IngestQueueDepth,
			IngestDropped,
			CAGGRefreshDuration,
			CAGGLastRefreshTimestamp,
			ReportGenerateDuration,
			AlertFired,
			JWTVerifyDuration,
			BreakerState,
			DBPoolAcquireDuration,
			AsyncJobDuration,
			LoginFailed,
			ExternalAPIFallback,
			ConfigValidationFailed,
			HTTPRequestDuration,
			HTTPRequests,
		)
	})
}

// SetBreakerState updates the breaker state gauge. Wire as resilience.StateObserver.
func SetBreakerState(name string, state int) {
	BreakerState.WithLabelValues(name).Set(float64(state))
}
