package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Problem is the RFC 7807 problem-details JSON structure.
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code,omitempty"`
}

// ErrorHandler returns a Fiber ErrorHandler that emits RFC 7807 problem-details.
func ErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var fe *fiber.Error
		status := fiber.StatusInternalServerError
		title := "Internal Server Error"
		detail := err.Error()
		if errors.As(err, &fe) {
			status = fe.Code
			title = fiber.ErrInternalServerError.Message
			if fe.Code == fiber.StatusNotFound {
				title = "Not Found"
			} else if fe.Code == fiber.StatusBadRequest {
				title = "Bad Request"
			} else if fe.Code == fiber.StatusUnauthorized {
				title = "Unauthorized"
			} else if fe.Code == fiber.StatusForbidden {
				title = "Forbidden"
			}
			detail = fe.Message
		}
		reqID := c.GetRespHeader("X-Request-ID")
		if status >= 500 {
			logger.Error("request failed",
				zap.Int("status", status),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("request_id", reqID),
				zap.Error(err))
		}
		return c.Status(status).JSON(Problem{
			Type:     "about:blank",
			Title:    title,
			Status:   status,
			Detail:   detail,
			Instance: c.OriginalURL(),
		})
	}
}

// BadRequest returns a formatted 400.
func BadRequest(msg string) error { return fiber.NewError(fiber.StatusBadRequest, msg) }

// Unauthorized returns a formatted 401.
func Unauthorized(msg string) error { return fiber.NewError(fiber.StatusUnauthorized, msg) }

// Forbidden returns a formatted 403.
func Forbidden(msg string) error { return fiber.NewError(fiber.StatusForbidden, msg) }

// NotFound returns a formatted 404.
func NotFound(msg string) error { return fiber.NewError(fiber.StatusNotFound, msg) }

// Internal returns a formatted 500.
func Internal(msg string) error { return fiber.NewError(fiber.StatusInternalServerError, msg) }
