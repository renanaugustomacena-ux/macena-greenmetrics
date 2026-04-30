// Package handlers — audit-log writer middleware.
//
// Covers GreenMetrics-GAPS B-04: the audit_log table (0001_init.sql:116-127)
// had no writer. This middleware INSERTs one row per state-changing HTTP
// request (POST/PUT/PATCH/DELETE), keyed on X-Request-ID as correlation_id.
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/repository"
)

// AuditConfig bundles the collaborators needed by the middleware.
type AuditConfig struct {
	Repo   *repository.TimescaleRepository
	Logger *zap.Logger
}

// AuditMiddleware returns a Fiber handler that inserts audit rows after the
// downstream handler returns (so status code is known). Reads fail open: if
// the INSERT errors we log a WARN and continue — audit is best-effort, we
// never break the user-facing request.
func AuditMiddleware(cfg AuditConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()
		mutating := method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch || method == fiber.MethodDelete
		err := c.Next()
		if !mutating {
			return err
		}
		if cfg.Repo == nil || cfg.Repo.Pool() == nil {
			// No DB available (dev fallback); log to stderr at DEBUG.
			if cfg.Logger != nil {
				cfg.Logger.Debug("audit skipped (no repo)", zap.String("path", c.OriginalURL()))
			}
			return err
		}

		email, _ := c.Locals("user_email").(string)
		tenantID, _ := c.Locals("tenant_id").(string)
		reqID := c.GetRespHeader("X-Request-ID")
		action := method + " " + stripQuery(c.OriginalURL())
		bodyHash := hashBody(c.Body())

		entityType, entityID := extractEntity(c.Path())

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		details := map[string]any{
			"status":     c.Response().StatusCode(),
			"method":     method,
			"path":       c.OriginalURL(),
			"body_sha256": bodyHash,
			"user_agent": c.Get("User-Agent"),
			"ip":         c.IP(),
		}

		if _, errIns := cfg.Repo.InsertAudit(ctx, repository.AuditEntry{
			TenantID:      tenantID,
			ActorEmail:    email,
			Action:        action,
			EntityType:    entityType,
			EntityID:      entityID,
			CorrelationID: reqID,
			Details:       details,
		}); errIns != nil && cfg.Logger != nil {
			cfg.Logger.Warn("audit insert failed", zap.Error(errIns), zap.String("path", c.OriginalURL()))
		}
		return err
	}
}

// stripQuery removes the query string for the action label (keeps it stable
// across paginated / filtered calls).
func stripQuery(full string) string {
	if i := strings.IndexByte(full, '?'); i > 0 {
		return full[:i]
	}
	return full
}

// hashBody returns a hex SHA-256 of the body for evidentiary purposes without
// leaking payload contents into the audit row.
func hashBody(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// extractEntity pulls the entity type+id out of well-known REST patterns.
// e.g. /api/v1/meters/abc-123 → ("meter", "abc-123"). Caller treats misses
// as empty strings.
func extractEntity(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// api/v1/<entity>/<id>
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" {
		return singular(parts[2]), parts[3]
	}
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "v1" {
		return singular(parts[2]), ""
	}
	return "", ""
}

func singular(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "s") {
		return s[:len(s)-1]
	}
	return s
}
