// Package v1 — centralised request binding helper.
//
// Doctrine refs: Rule 14, Rule 34, Rule 39, Rule 42.
// Replaces every manual `c.BodyParser(&req)` + nil-check across handlers.
package v1

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// MaxBodyBytes caps generic request bodies; ingest-specific routes override via per-route option.
const MaxBodyBytes = 4 * 1024 * 1024 // 4 MB
// MaxIngestBodyBytes for /v1/readings/ingest and /v1/pulse/ingest.
const MaxIngestBodyBytes = 16 * 1024 * 1024 // 16 MB

// Bind decodes the request body into dst and validates it. Returns a typed
// fiber.Error suitable for the global error handler to render as RFC 7807
// Problem Details.
//
// Caller must define dst as a struct with `json:"..."` and `validate:"..."` tags.
// Generic via go 1.18+; the dst type appears in the error context for easier debug.
func Bind[T any](c *fiber.Ctx, dst *T) error {
	if dst == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Bind: nil destination")
	}
	if err := c.BodyParser(dst); err != nil {
		return ProblemDetails(c, fiber.StatusBadRequest, "Invalid request body", err.Error(), "BODY_INVALID")
	}
	if err := Validator().Struct(dst); err != nil {
		details := ValidationErrors(err)
		return ProblemDetails(c, fiber.StatusUnprocessableEntity,
			"Validation failed",
			strings.Join(details, "; "),
			"VALIDATION_FAILED")
	}
	return nil
}

// BindQuery decodes query string parameters into dst with validation.
func BindQuery[T any](c *fiber.Ctx, dst *T) error {
	if dst == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "BindQuery: nil destination")
	}
	if err := c.QueryParser(dst); err != nil {
		return ProblemDetails(c, fiber.StatusBadRequest, "Invalid query string", err.Error(), "QUERY_INVALID")
	}
	if err := Validator().Struct(dst); err != nil {
		details := ValidationErrors(err)
		return ProblemDetails(c, fiber.StatusUnprocessableEntity,
			"Validation failed",
			strings.Join(details, "; "),
			"VALIDATION_FAILED")
	}
	return nil
}

// ProblemDetails returns a Fiber error carrying RFC 7807 fields. The global
// error handler in cmd/server/main.go renders this as application/problem+json.
func ProblemDetails(c *fiber.Ctx, status int, title, detail, code string) error {
	return &Problem{
		Status: status,
		Title:  title,
		Detail: detail,
		Code:   code,
		Path:   c.OriginalURL(),
		ReqID:  c.Get(fiber.HeaderXRequestID),
	}
}

// Problem implements the error interface and the fiber error contract.
type Problem struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	Path   string `json:"instance,omitempty"`
	Code   string `json:"code,omitempty"`
	ReqID  string `json:"request_id,omitempty"`
}

func (p *Problem) Error() string {
	if p.Detail != "" {
		return p.Title + ": " + p.Detail
	}
	return p.Title
}

// AsProblem unwraps an error chain looking for a *Problem.
func AsProblem(err error) (*Problem, bool) {
	var p *Problem
	if errors.As(err, &p) {
		return p, true
	}
	return nil, false
}
