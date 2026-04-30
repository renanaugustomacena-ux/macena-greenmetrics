// Package handlers contains HTTP controllers, middleware, and routing.
package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	apiv1 "github.com/greenmetrics/backend/internal/api/v1"
	"github.com/greenmetrics/backend/internal/config"
	"github.com/greenmetrics/backend/internal/repository"
	"github.com/greenmetrics/backend/internal/security"
	"github.com/greenmetrics/backend/internal/services"
)

// Dependencies is the bundle of collaborators injected into handlers.
type Dependencies struct {
	Config    *config.Config
	Logger    *zap.Logger
	Repo      *repository.TimescaleRepository
	Analytics *services.EnergyAnalytics
	Carbon    *services.CarbonCalculator
	Reporter  *services.ReportGenerator
	Alerts    *services.AlertEngine
	Version   string
	Commit    string
	StartedAt time.Time
}

// Register wires all routes onto the Fiber app.
//
//	@title GreenMetrics API
//	@version 1.0
//	@description Energy monitoring + carbon accounting API. CSRD/ESRS E1 + Piano 5.0.
//	@termsOfService https://greenmetrics.it/tos
//	@contact.name GreenMetrics Support
//	@contact.email support@greenmetrics.it
//	@license.name Proprietary
//	@host localhost:8082
//	@BasePath /api
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
func Register(app *fiber.App, d Dependencies) {
	api := app.Group("/api")

	health := newHealthHandler(d)
	api.Get("/health", health.Check)
	api.Get("/ready", health.Ready)
	api.Get("/live", health.Live)

	// v1 endpoints. Default rate limit 60/min (v2.0 §12).
	v1 := api.Group("/v1", RateLimit(d.Config.RateLimitPerMinute))

	auth := newAuthHandler(d)
	// /auth/login and /auth/refresh are the authentication entry points —
	// no session or token exists yet at the moment of the call, so CSRF
	// (which protects an existing session from being replayed) cannot
	// meaningfully apply. They are registered BEFORE `v1.Use(CSRF...)` so
	// that the middleware does not gate the public auth flow.
	// Stricter rate limit for login (brute-force mitigation, v2.0 §12).
	v1.Post("/auth/login", RateLimit(d.Config.RateLimitLoginPerMinute), auth.Login)
	v1.Post("/auth/refresh", RateLimit(d.Config.RateLimitLoginPerMinute), auth.Refresh)

	// CSRF for cookie-based clients on state-changing routes; Bearer-auth
	// clients bypass via the middleware's Authorization-header check.
	v1.Use(CSRFMiddleware(DefaultCSRFConfig(d.Config.JWTSecret)))
	v1.Post("/auth/logout", auth.Logout)

	// Idempotency middleware (Rule 35 / RFC 9457). Backed by idempotency_keys
	// table from migration 00007. The middleware short-circuits on
	// GET/HEAD/OPTIONS; safe to apply at group level.
	//
	// On dev fallback (Repo unavailable), middleware degrades to passthrough so
	// the stack still boots without a DB. CurrentEnv gates "required" enforcement
	// to production + staging only.
	idempMW := func(c *fiber.Ctx) error { return c.Next() }
	if d.Repo != nil && d.Repo.Pool() != nil {
		idempStore := repository.NewIdempotencyRepo(d.Repo.Pool())
		idempMW = apiv1.IdempotencyMiddleware(idempStore, apiv1.IdempotencyConfig{
			CurrentEnv: d.Config.AppEnv,
		})
	}

	// Audit middleware runs AFTER JWTMiddleware so we can attribute the row
	// to an authenticated subject. Idempotency middleware runs after audit so
	// that a replayed response is still audited.
	protected := v1.Group("",
		JWTMiddleware(d),
		AuditMiddleware(AuditConfig{Repo: d.Repo, Logger: d.Logger}),
		idempMW,
	)

	// --- Meters --------------------------------------------------------------
	meters := newMetersHandler(d)
	protected.Get("/meters", security.RequirePermission(security.PermMetersRead), meters.List)
	protected.Post("/meters", security.RequirePermission(security.PermMetersWrite), meters.Create)
	protected.Get("/meters/:id", security.RequirePermission(security.PermMetersRead), meters.Get)
	protected.Put("/meters/:id", security.RequirePermission(security.PermMetersWrite), meters.Update)
	protected.Delete("/meters/:id", security.RequirePermission(security.PermMetersWrite), meters.Delete)
	protected.Post("/meters/:id/probe", security.RequirePermission(security.PermMetersWrite), meters.Probe)

	// --- Readings ------------------------------------------------------------
	readings := newReadingsHandler(d)
	protected.Get("/readings", security.RequirePermission(security.PermReadingsRead), readings.Query)
	// Higher rate for ingest (v2.0 §12 suggests 300/min for ingest endpoints).
	protected.Post("/readings/ingest",
		RateLimit(d.Config.RateLimitIngestPerMinute),
		security.RequirePermission(security.PermReadingsIngest),
		readings.Ingest)
	protected.Get("/readings/aggregated", security.RequirePermission(security.PermReadingsRead), readings.Aggregated)
	protected.Get("/readings/export", security.RequirePermission(security.PermReadingsRead), readings.Export)

	// --- Reports -------------------------------------------------------------
	reports := newReportsHandler(d)
	protected.Get("/reports", security.RequirePermission(security.PermReportsRead), reports.List)
	protected.Post("/reports", security.RequirePermission(security.PermReportsGenerate), reports.Generate)
	protected.Get("/reports/:id", security.RequirePermission(security.PermReportsRead), reports.Get)
	protected.Get("/reports/:id/download", security.RequirePermission(security.PermReportsRead), reports.Download)
	protected.Post("/reports/:id/submit", security.RequirePermission(security.PermReportsGenerate), reports.Submit)

	// --- Tenants -------------------------------------------------------------
	// Self-read available to every authenticated role (no RBAC gate).
	// Self-update gated to tenant admin only.
	tenants := newTenantsHandler(d)
	protected.Get("/tenants/me", tenants.GetSelf)
	protected.Put("/tenants/me", security.RequirePermission(security.PermTenantsAdmin), tenants.UpdateSelf)

	// --- Alerts --------------------------------------------------------------
	alerts := newAlertsHandler(d)
	protected.Get("/alerts", security.RequirePermission(security.PermAlertsRead), alerts.List)
	protected.Post("/alerts/:id/ack", security.RequirePermission(security.PermAlertsAck), alerts.Ack)
	protected.Post("/alerts/:id/resolve", security.RequirePermission(security.PermAlertsAck), alerts.Resolve)

	// --- Emission factors ----------------------------------------------------
	ef := newEmissionFactorsHandler(d)
	protected.Get("/emission-factors", security.RequirePermission(security.PermEmissionFactorsRead), ef.List)
	protected.Post("/emission-factors", security.RequirePermission(security.PermEmissionFactorsWrite), ef.Create)
	protected.Get("/emission-factors/:code", security.RequirePermission(security.PermEmissionFactorsRead), ef.Get)
	protected.Put("/emission-factors/:code", security.RequirePermission(security.PermEmissionFactorsWrite), ef.Update)

	// Pulse-counter webhook (field gateways POST here). Validated inside the
	// handler against PULSE_WEBHOOK_SECRET; rate-limited as ingest.
	protected.Post("/pulse/ingest",
		RateLimit(d.Config.RateLimitIngestPerMinute),
		security.RequirePermission(security.PermPulseIngest),
		newPulseHandler(d).Ingest)

	// OpenAPI / Swagger docs. The spec is generated by `swag init`
	// into docs/openapi.json at build time; here we simply serve it.
	api.Get("/docs/openapi.json", OpenAPIHandler)
	api.Get("/docs", SwaggerUIHandler)

	// Grafana back-channel (service-to-service token).
	api.Get("/internal/metrics", MetricsHandler)
}
