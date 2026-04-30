// Package handlers — security headers, CSRF, and rate-limit middleware.
//
// Covers GreenMetrics-GAPS G-02 (headers), G-03 (CSRF), B-03 (rate limit).
package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// SecurityHeaders returns a middleware that applies the v2.0 §12 header set.
//
// Values:
//   - Content-Security-Policy: self only by default; relaxed for dev by caller.
//   - Strict-Transport-Security: 1y preload (the ingress terminates TLS in prod).
//   - X-Content-Type-Options: nosniff
//   - Referrer-Policy: strict-origin-when-cross-origin
//   - Permissions-Policy: denies non-essential powerful features.
//   - X-Frame-Options: DENY (also covered by frame-ancestors in CSP).
func SecurityHeaders(isProd bool) fiber.Handler {
	csp := strings.Join([]string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline'",
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: https:",
		"connect-src 'self' https:",
		"frame-ancestors 'none'",
		"form-action 'self'",
		"base-uri 'self'",
		"object-src 'none'",
	}, "; ")
	permissions := strings.Join([]string{
		"geolocation=()",
		"microphone=()",
		"camera=()",
		"payment=()",
		"usb=()",
		"fullscreen=(self)",
	}, ", ")
	return func(c *fiber.Ctx) error {
		c.Set("Content-Security-Policy", csp)
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", permissions)
		c.Set("X-Frame-Options", "DENY")
		if isProd {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		return c.Next()
	}
}

// --- CSRF ----------------------------------------------------------------

// CSRFConfig holds CSRF middleware knobs.
type CSRFConfig struct {
	CookieName   string
	HeaderName   string
	Secret       string
	TokenTTL     time.Duration
	SafeMethods  map[string]bool
	AllowBearer  bool // if true, requests with Authorization: Bearer bypass CSRF
}

// DefaultCSRFConfig returns sane defaults.
func DefaultCSRFConfig(secret string) CSRFConfig {
	return CSRFConfig{
		CookieName: "gm_csrf",
		HeaderName: "X-CSRF-Token",
		Secret:     secret,
		TokenTTL:   12 * time.Hour,
		SafeMethods: map[string]bool{
			fiber.MethodGet: true, fiber.MethodHead: true, fiber.MethodOptions: true,
		},
		AllowBearer: true,
	}
}

// CSRFMiddleware implements a double-submit-cookie pattern.
//
// For safe HTTP methods or when the client uses an Authorization: Bearer token
// (a non-cookie credential vector that is not replayable via CSRF) we pass
// through. For state-changing requests we require the X-CSRF-Token header to
// match the gm_csrf cookie.
func CSRFMiddleware(cfg CSRFConfig) fiber.Handler {
	if cfg.CookieName == "" {
		cfg.CookieName = "gm_csrf"
	}
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-CSRF-Token"
	}
	if cfg.TokenTTL == 0 {
		cfg.TokenTTL = 12 * time.Hour
	}
	return func(c *fiber.Ctx) error {
		method := c.Method()
		if cfg.SafeMethods[method] {
			// Ensure a cookie is set for subsequent POSTs.
			if c.Cookies(cfg.CookieName) == "" {
				tok, err := newCSRFToken(cfg.Secret)
				if err == nil {
					c.Cookie(&fiber.Cookie{
						Name:     cfg.CookieName,
						Value:    tok,
						Path:     "/",
						HTTPOnly: false, // client JS needs to read it to echo in header
						SameSite: "Lax",
						Secure:   true,
						Expires:  time.Now().Add(cfg.TokenTTL),
					})
				}
			}
			return c.Next()
		}
		if cfg.AllowBearer && strings.HasPrefix(c.Get("Authorization"), "Bearer ") {
			return c.Next()
		}
		cookie := c.Cookies(cfg.CookieName)
		header := c.Get(cfg.HeaderName)
		if cookie == "" || header == "" {
			return Forbidden("csrf token missing")
		}
		if !hmac.Equal([]byte(cookie), []byte(header)) {
			return Forbidden("csrf token mismatch")
		}
		return c.Next()
	}
}

// newCSRFToken returns an HMAC-bound random token.
func newCSRFToken(secret string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(buf)
	tag := mac.Sum(nil)
	return hex.EncodeToString(buf) + "." + hex.EncodeToString(tag[:8]), nil
}

// --- Rate limiter --------------------------------------------------------

// RateLimit builds a limiter.New middleware with per-minute budget.
func RateLimit(perMinute int) fiber.Handler {
	if perMinute <= 0 {
		perMinute = 60
	}
	return limiter.New(limiter.Config{
		Max:        perMinute,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Prefer authenticated subject; fall back to X-Forwarded-For, then IP.
			if sub, ok := c.Locals("user_email").(string); ok && sub != "" {
				return "user:" + sub
			}
			if xff := c.Get("X-Forwarded-For"); xff != "" {
				return "xff:" + xff
			}
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(Problem{
				Type:   "about:blank",
				Title:  "Too Many Requests",
				Status: fiber.StatusTooManyRequests,
				Detail: "rate limit exceeded; slow down",
				Code:   "RATE_LIMITED",
			})
		},
	})
}
