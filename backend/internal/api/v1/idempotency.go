// Package v1 — Idempotency-Key middleware (RFC 9457 / Stripe-style).
//
// Doctrine refs: Rule 14, Rule 35 (idempotency), Rule 39, Rule 42.
// Mitigates: ingest duplicates from retry storms; out-of-order client retries.
//
// Behaviour:
//   - In production, missing Idempotency-Key on POST → 400 Problem.
//   - Lookup `(tenant_id, key)` in `idempotency_keys`.
//   - Hit + same request hash → replay stored response (status, body, headers).
//   - Hit + different request hash → 422 Conflict (Idempotency-Key reuse).
//   - Miss → invoke handler; persist response; return.
//
// Storage: hypertable `idempotency_keys` (migration 00007), 24h retention.
package v1

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// IdempotencyStore is the abstract storage backend (pgx-backed in production,
// in-memory in tests).
type IdempotencyStore interface {
	Get(ctx context.Context, tenantID uuid.UUID, key string) (*IdempotencyRecord, error)
	Put(ctx context.Context, rec *IdempotencyRecord) error
}

// IdempotencyRecord is the stored replay payload.
type IdempotencyRecord struct {
	TenantID    uuid.UUID
	Key         string
	RequestHash [32]byte
	Status      int
	Body        []byte
	Headers     map[string]string
}

// ErrIdempotencyMiss signals not found.
var ErrIdempotencyMiss = errors.New("idempotency: miss")

// IdempotencyConfig governs middleware behaviour per environment.
type IdempotencyConfig struct {
	// RequiredInEnv — list of APP_ENV values that hard-require the header.
	// Default: ["production", "staging"].
	RequiredInEnv []string
	// CurrentEnv — the running env (from cfg.AppEnv).
	CurrentEnv string
}

// IdempotencyMiddleware decodes the Idempotency-Key header, looks up the store,
// replays on hit + matching hash, conflicts on hit + mismatch, otherwise invokes
// the handler and persists the response.
func IdempotencyMiddleware(store IdempotencyStore, cfg IdempotencyConfig) fiber.Handler {
	if len(cfg.RequiredInEnv) == 0 {
		cfg.RequiredInEnv = []string{"production", "staging"}
	}
	required := func() bool {
		for _, e := range cfg.RequiredInEnv {
			if strings.EqualFold(e, cfg.CurrentEnv) {
				return true
			}
		}
		return false
	}()

	return func(c *fiber.Ctx) error {
		// Only POST/PUT/PATCH/DELETE need idempotency.
		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}

		key := strings.TrimSpace(c.Get("Idempotency-Key"))
		if key == "" {
			if required {
				return ProblemDetails(c, fiber.StatusBadRequest,
					"Idempotency-Key required",
					"This endpoint requires an Idempotency-Key header in this environment.",
					"IDEMPOTENCY_KEY_REQUIRED")
			}
			return c.Next()
		}

		// Tenant context comes from JWTMiddleware → c.Locals("tenant_id").
		tenantStr, ok := c.Locals("tenant_id").(string)
		if !ok || tenantStr == "" {
			return ProblemDetails(c, fiber.StatusUnauthorized,
				"Tenant context required",
				"Idempotency requires authenticated tenant context.",
				"TENANT_CONTEXT_REQUIRED")
		}
		tenantID, err := uuid.Parse(tenantStr)
		if err != nil {
			return ProblemDetails(c, fiber.StatusUnauthorized,
				"Invalid tenant id",
				err.Error(),
				"TENANT_INVALID")
		}

		body := c.Body()
		hash := sha256.Sum256(body)

		// Lookup.
		ctx := c.UserContext()
		rec, err := store.Get(ctx, tenantID, key)
		switch {
		case err == nil:
			if rec.RequestHash == hash {
				// Replay.
				for k, v := range rec.Headers {
					c.Set(k, v)
				}
				c.Set("Idempotent-Replay", "true")
				return c.Status(rec.Status).Send(rec.Body)
			}
			return ProblemDetails(c, fiber.StatusUnprocessableEntity,
				"Idempotency-Key conflict",
				"This Idempotency-Key was used with a different request body.",
				"IDEMPOTENCY_CONFLICT")
		case errors.Is(err, ErrIdempotencyMiss):
			// Miss — fall through to handler.
		case errors.Is(err, pgx.ErrNoRows):
			// Same as miss.
		default:
			return ProblemDetails(c, fiber.StatusInternalServerError,
				"Idempotency lookup failed",
				err.Error(),
				"IDEMPOTENCY_LOOKUP_FAILED")
		}

		// Invoke downstream handler.
		if err := c.Next(); err != nil {
			return err
		}

		// Persist successful and client-error responses (4xx are deterministic; 5xx are not).
		status := c.Response().StatusCode()
		if status >= 500 {
			return nil
		}

		respBody := append([]byte(nil), c.Response().Body()...)
		respHeaders := map[string]string{
			"Content-Type": string(c.Response().Header.ContentType()),
		}

		put := &IdempotencyRecord{
			TenantID:    tenantID,
			Key:         key,
			RequestHash: hash,
			Status:      status,
			Body:        respBody,
			Headers:     respHeaders,
		}
		if err := store.Put(ctx, put); err != nil {
			// Don't fail the request on store-write failure — log out-of-band.
			c.Set("Idempotency-Persist-Error", err.Error())
		}
		return nil
	}
}

// IdempotencyKeyHeader returns the request hash of body bytes — exposed for tests.
func IdempotencyKeyHeader(body []byte) string {
	h := sha256.Sum256(body)
	return fmt.Sprintf("%x", h[:8])
}

// MarshalForReplay is a tiny convenience for handlers that need to JSON-encode a
// payload before letting the middleware replay it on subsequent calls.
func MarshalForReplay(v any) ([]byte, error) {
	return json.Marshal(v)
}
