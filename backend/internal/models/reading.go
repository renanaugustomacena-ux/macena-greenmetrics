package models

import "time"

// Reading is an individual time-series data point.
type Reading struct {
	Ts          time.Time       `json:"ts"`
	TenantID    string          `json:"tenant_id"`
	MeterID     string          `json:"meter_id"`
	ChannelID   string          `json:"channel_id"`
	Value       float64         `json:"value"`
	Unit        string          `json:"unit"`
	QualityCode int             `json:"quality_code"`
	RawPayload  map[string]any  `json:"raw_payload,omitempty"`
}

// Aggregate is the continuous-aggregate row returned to clients.
type Aggregate struct {
	Bucket    time.Time `json:"bucket"`
	MeterID   string    `json:"meter_id"`
	ChannelID string    `json:"channel_id"`
	SumValue  float64   `json:"sum_value"`
	AvgValue  float64   `json:"avg_value"`
	MaxValue  float64   `json:"max_value"`
	Unit      string    `json:"unit"`
}

// Resolution is a supported aggregation window.
type Resolution string

const (
	ResolutionRaw     Resolution = "raw"
	Resolution15Min   Resolution = "15min"
	Resolution1Hour   Resolution = "1h"
	Resolution1Day    Resolution = "1d"
)
