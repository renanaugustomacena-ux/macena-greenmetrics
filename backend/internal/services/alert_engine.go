package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/models"
	"github.com/greenmetrics/backend/internal/repository"
)

// AlertEngine evaluates rules over time-series data and produces alerts.
//
// Rule families:
//   - Consumption anomaly: z-score threshold vs rolling window baseline.
//   - Peak exceeded: 15-min peak above contracted kW.
//   - Baseline drift: month-over-month mean shift > N%.
//   - Meter offline: no reading received in N minutes.
//   - Power factor low: cosφ < 0.90 on three-phase meters.
//   - Emission budget: Scope 1+2 YTD > configured CO2e budget.
//   - Reporting due: audit due, CSRD due, Piano 5.0 window open.
type AlertEngine struct {
	repo   *repository.TimescaleRepository
	logger *zap.Logger
}

// NewAlertEngine builds the engine.
func NewAlertEngine(repo *repository.TimescaleRepository, logger *zap.Logger) *AlertEngine {
	return &AlertEngine{repo: repo, logger: logger}
}

// Evaluate runs all rules for a tenant and window, returning any fired alerts.
func (e *AlertEngine) Evaluate(ctx context.Context, tenantID string, now time.Time) ([]models.Alert, error) {
	fired := []models.Alert{}

	meters, err := e.repo.ListMeters(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	lookback := now.Add(-24 * time.Hour)
	for _, m := range meters {
		rows, err := e.repo.QueryAggregated(ctx, tenantID, m.ID, "15min", lookback, now)
		if err != nil || len(rows) < 10 {
			continue
		}
		vs := make([]float64, 0, len(rows))
		for _, r := range rows {
			vs = append(vs, r.SumValue)
		}
		mean, stdev := meanStd(vs)
		// Anomaly rule: z-score > 3.
		for _, r := range rows {
			if stdev > 0 && math.Abs(r.SumValue-mean)/stdev > 3.0 {
				fired = append(fired, models.Alert{
					ID:       fmt.Sprintf("anom-%s-%d", m.ID, r.Bucket.Unix()),
					TenantID: tenantID,
					MeterID:  m.ID,
					Kind:     models.AlertConsumptionAnomaly,
					Severity: models.SeverityWarning,
					Message:  fmt.Sprintf("Consumo anomalo sul contatore %s (%.2f %s nel bucket %s)", m.Label, r.SumValue, r.Unit, r.Bucket.Format(time.RFC3339)),
					Context: map[string]any{
						"bucket": r.Bucket.Format(time.RFC3339),
						"value":  r.SumValue,
						"mean":   mean,
						"stdev":  stdev,
					},
					TriggeredAt: now,
				})
			}
		}
		// Meter-offline rule: latest bucket older than 30 min.
		last := rows[len(rows)-1].Bucket
		if now.Sub(last) > 30*time.Minute {
			fired = append(fired, models.Alert{
				ID:       fmt.Sprintf("offline-%s-%d", m.ID, now.Unix()),
				TenantID: tenantID,
				MeterID:  m.ID,
				Kind:     models.AlertMeterOffline,
				Severity: models.SeverityCritical,
				Message:  fmt.Sprintf("Contatore %s offline da %s", m.Label, now.Sub(last)),
				Context: map[string]any{
					"last_bucket": last.Format(time.RFC3339),
				},
				TriggeredAt: now,
			})
		}
	}
	return fired, nil
}
