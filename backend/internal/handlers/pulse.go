// Package handlers — pulse-counter webhook endpoint.
//
// Field gateways (Raspberry Pi, Teltonika, Elvaco CMe) POST pulse summaries
// here. We validate HMAC-SHA256 against PULSE_WEBHOOK_SECRET and hand off to
// services.PulseIngestor for dedupe and accumulation.
package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/repository"
	"github.com/greenmetrics/backend/internal/services"
)

type pulseHandler struct {
	d        Dependencies
	ingestor *services.PulseIngestor
	secret   string
}

func newPulseHandler(d Dependencies) *pulseHandler {
	ing, err := services.NewPulseIngestor(d.Config.PulseWebhookSecret, d.Config.PulseDedupeWindow, d.Logger)
	if err != nil {
		d.Logger.Info("pulse ingestor disabled", zap.Error(err))
	}
	return &pulseHandler{d: d, ingestor: ing, secret: d.Config.PulseWebhookSecret}
}

// Ingest handles POST /api/v1/pulse/ingest.
func (h *pulseHandler) Ingest(c *fiber.Ctx) error {
	if h.ingestor == nil {
		// No silent stub: surface a typed 501 naming the env var.
		return fiber.NewError(fiber.StatusNotImplemented,
			"pulse webhook not configured; set PULSE_WEBHOOK_SECRET")
	}
	var f services.PulseFrame
	if err := c.BodyParser(&f); err != nil {
		return BadRequest("invalid JSON: " + err.Error())
	}

	if err := verifySignature(h.secret, f, c.Body()); err != nil {
		return Unauthorized("invalid signature: " + err.Error())
	}

	ok, reason := h.ingestor.Accept(c.Context(), f)
	if !ok {
		return fiber.NewError(fiber.StatusConflict, reason)
	}

	value := h.ingestor.ComputeValue(f)

	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = "placeholder-tenant"
	}
	if h.d.Repo != nil && h.d.Repo.Pool() != nil {
		if _, err := h.d.Repo.InsertReadings(c.Context(), []repository.Reading{
			{
				Ts:        f.Timestamp,
				TenantID:  tenantID,
				MeterID:   f.MeterID,
				ChannelID: "pulse_total",
				Value:     value,
				Unit:      f.Unit,
			},
		}); err != nil {
			h.d.Logger.Warn("pulse persist failed", zap.Error(err))
			return Internal("persist failed")
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"meter_id":    f.MeterID,
		"tick_count":  f.TickCount,
		"total_value": value,
		"unit":        f.Unit,
	})
}

// verifySignature reconstructs the canonical body hash and compares.
//
// Canonical form is a fixed ordering of the meaningful fields joined by '|':
//   meter_id|tick_count|pulses_per_unit|unit|timestamp-rfc3339
// The signature is hex(hmac-sha256(secret, canonical)).
func verifySignature(secret string, f services.PulseFrame, raw []byte) error {
	if strings.TrimSpace(secret) == "" {
		return errors.New("server secret empty")
	}
	if f.Signature == "" {
		return errors.New("frame.signature empty")
	}
	// Build canonical string; sort to avoid ambiguity.
	parts := []string{
		"meter_id=" + f.MeterID,
		"tick_count=" + strconv.FormatUint(f.TickCount, 10),
		"pulses_per_unit=" + strconv.FormatFloat(f.PulsesPerUnit, 'f', -1, 64),
		"unit=" + f.Unit,
		"timestamp=" + f.Timestamp.UTC().Format(time.RFC3339),
	}
	sort.Strings(parts)
	canonical := strings.Join(parts, "|")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	want := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(want), []byte(strings.ToLower(f.Signature))) {
		return fmt.Errorf("signature mismatch (canonical bytes=%d, raw=%d)", len(canonical), len(raw))
	}
	return nil
}
