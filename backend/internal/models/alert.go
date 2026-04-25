package models

import "time"

// AlertSeverity classifies the urgency of an alert.
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// AlertKind enumerates alert categories relevant to energy management.
type AlertKind string

const (
	AlertConsumptionAnomaly AlertKind = "consumption_anomaly"
	AlertPeakExceeded       AlertKind = "peak_exceeded"
	AlertBaselineDrift      AlertKind = "baseline_drift"
	AlertMeterOffline       AlertKind = "meter_offline"
	AlertPowerFactorLow     AlertKind = "power_factor_low"
	AlertEmissionBudget     AlertKind = "emission_budget_exceeded"
	AlertReportingDue       AlertKind = "reporting_due"
)

// Alert is a raised notification.
type Alert struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenant_id"`
	MeterID     string        `json:"meter_id,omitempty"`
	Kind        AlertKind     `json:"kind"`
	Severity    AlertSeverity `json:"severity"`
	Message     string        `json:"message"`
	Context     map[string]any `json:"context,omitempty"`
	TriggeredAt time.Time     `json:"triggered_at"`
	AckedAt     *time.Time    `json:"acked_at,omitempty"`
	AckedBy     string        `json:"acked_by,omitempty"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
}
