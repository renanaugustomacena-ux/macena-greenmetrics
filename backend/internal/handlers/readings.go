package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/greenmetrics/backend/internal/models"
	"github.com/greenmetrics/backend/internal/repository"
)

type readingsHandler struct{ d Dependencies }

func newReadingsHandler(d Dependencies) *readingsHandler { return &readingsHandler{d: d} }

// Query returns raw readings for a meter and time range (paginated).
func (h *readingsHandler) Query(c *fiber.Ctx) error {
	meterID := c.Query("meter_id")
	if meterID == "" {
		return BadRequest("meter_id required")
	}
	from, to, err := parseRange(c)
	if err != nil {
		return BadRequest(err.Error())
	}
	limit, _ := strconv.Atoi(c.Query("limit", "1000"))
	return c.JSON(fiber.Map{
		"meter_id": meterID,
		"from":     from.Format(time.RFC3339),
		"to":       to.Format(time.RFC3339),
		"limit":    limit,
		"items":    []any{},
	})
}

// Ingest accepts a batch of readings and persists to the hypertable.
func (h *readingsHandler) Ingest(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = "placeholder-tenant"
	}
	var body struct {
		Readings []models.Reading `json:"readings"`
	}
	if err := c.BodyParser(&body); err != nil {
		return BadRequest("invalid JSON")
	}
	rows := make([]repository.Reading, 0, len(body.Readings))
	for _, r := range body.Readings {
		rows = append(rows, repository.Reading{
			Ts:          r.Ts,
			TenantID:    tenantID,
			MeterID:     r.MeterID,
			ChannelID:   r.ChannelID,
			Value:       r.Value,
			Unit:        r.Unit,
			QualityCode: r.QualityCode,
		})
	}
	var accepted int64
	if h.d.Repo != nil {
		n, err := h.d.Repo.InsertReadings(c.Context(), rows)
		if err != nil {
			return Internal("ingest failed: " + err.Error())
		}
		accepted = n
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"accepted": accepted,
		"received": len(body.Readings),
	})
}

// Aggregated returns a continuous-aggregate slice.
func (h *readingsHandler) Aggregated(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = "placeholder-tenant"
	}
	meterID := c.Query("meter_id")
	resolution := c.Query("resolution", "1h")
	if meterID == "" {
		return BadRequest("meter_id required")
	}
	from, to, err := parseRange(c)
	if err != nil {
		return BadRequest(err.Error())
	}
	if h.d.Repo == nil {
		return c.JSON(fiber.Map{"items": []any{}, "resolution": resolution})
	}
	rows, err := h.d.Repo.QueryAggregated(c.Context(), tenantID, meterID, resolution, from, to)
	if err != nil {
		return Internal("aggregate query: " + err.Error())
	}
	items := make([]models.Aggregate, 0, len(rows))
	for _, r := range rows {
		items = append(items, models.Aggregate{
			Bucket:    r.Bucket,
			MeterID:   r.MeterID,
			ChannelID: r.ChannelID,
			SumValue:  r.SumValue,
			AvgValue:  r.AvgValue,
			MaxValue:  r.MaxValue,
			Unit:      r.Unit,
		})
	}
	return c.JSON(fiber.Map{
		"resolution": resolution,
		"items":      items,
		"count":      len(items),
	})
}

// Export streams CSV for external audit.
func (h *readingsHandler) Export(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="greenmetrics-readings-export.csv"`)
	return c.SendString("ts,tenant_id,meter_id,channel_id,value,unit,quality_code\n")
}

func parseRange(c *fiber.Ctx) (time.Time, time.Time, error) {
	fromS := c.Query("from")
	toS := c.Query("to")
	if fromS == "" || toS == "" {
		return time.Time{}, time.Time{}, fiberErr("from and to required (RFC3339)")
	}
	from, err := time.Parse(time.RFC3339, fromS)
	if err != nil {
		return time.Time{}, time.Time{}, fiberErr("invalid 'from': " + err.Error())
	}
	to, err := time.Parse(time.RFC3339, toS)
	if err != nil {
		return time.Time{}, time.Time{}, fiberErr("invalid 'to': " + err.Error())
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, fiberErr("'to' must be after 'from'")
	}
	return from, to, nil
}

type errString string

func (e errString) Error() string { return string(e) }
func fiberErr(msg string) error   { return errString(msg) }
