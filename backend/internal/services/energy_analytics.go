// Package services contains the GreenMetrics domain logic.
package services

import (
	"context"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"

	"github.com/greenmetrics/backend/internal/repository"
)

// EnergyAnalytics performs consumption analysis, baselining, and trend decomposition.
type EnergyAnalytics struct {
	repo   *repository.TimescaleRepository
	logger *zap.Logger
}

// NewEnergyAnalytics constructs the analyser.
func NewEnergyAnalytics(repo *repository.TimescaleRepository, logger *zap.Logger) *EnergyAnalytics {
	return &EnergyAnalytics{repo: repo, logger: logger}
}

// ConsumptionStats is the standard summary emitted by the analyser.
type ConsumptionStats struct {
	MeterID         string    `json:"meter_id"`
	TotalKWh        float64   `json:"total_kwh"`
	PeakKW          float64   `json:"peak_kw"`
	AverageKW       float64   `json:"average_kw"`
	LoadFactor      float64   `json:"load_factor"`       // average / peak
	BaselineKWh     float64   `json:"baseline_kwh"`      // rolling baseline
	DeviationPct    float64   `json:"deviation_pct"`     // (actual - baseline) / baseline
	Anomalies       []time.Time `json:"anomalies"`
	DominantShift   string    `json:"dominant_shift"`    // "morning"|"afternoon"|"night"
	From            time.Time `json:"from"`
	To              time.Time `json:"to"`
}

// Compute builds consumption statistics for a meter over a window.
//
// Algorithm (abridged):
//  1. Pull 15-min aggregates via the repository.
//  2. Sum to total energy.
//  3. Compute peak, average, load factor.
//  4. Compute rolling-mean baseline over a reference window.
//  5. Flag anomalies: buckets > baseline + k·σ (z-score).
func (a *EnergyAnalytics) Compute(ctx context.Context, tenantID, meterID string, from, to time.Time) (*ConsumptionStats, error) {
	rows, err := a.repo.QueryAggregated(ctx, tenantID, meterID, "15min", from, to)
	if err != nil {
		return nil, err
	}
	stats := &ConsumptionStats{MeterID: meterID, From: from, To: to}
	if len(rows) == 0 {
		return stats, nil
	}
	var total, peak, sum float64
	values := make([]float64, 0, len(rows))
	for _, r := range rows {
		total += r.SumValue
		values = append(values, r.SumValue)
		if r.SumValue > peak {
			peak = r.SumValue
		}
		sum += r.SumValue
	}
	avg := sum / float64(len(rows))
	stats.TotalKWh = total
	stats.PeakKW = peak * 4.0 // 15-min energy → average kW equivalent
	stats.AverageKW = avg * 4.0
	if peak > 0 {
		stats.LoadFactor = avg / peak
	}
	// baseline = mean of first 30% of the window (rolling baseline).
	baselineLen := int(math.Max(1, float64(len(values))*0.3))
	baselineSamples := values[:baselineLen]
	baselineSum := 0.0
	for _, v := range baselineSamples {
		baselineSum += v
	}
	baseline := baselineSum / float64(len(baselineSamples))
	stats.BaselineKWh = baseline * float64(len(values))
	if baseline > 0 {
		stats.DeviationPct = ((avg - baseline) / baseline) * 100.0
	}
	// Anomalies: z-score > 3 over the window.
	mean, stdev := meanStd(values)
	for i, v := range values {
		if stdev > 0 && math.Abs(v-mean)/stdev > 3.0 {
			stats.Anomalies = append(stats.Anomalies, rows[i].Bucket)
		}
	}
	stats.DominantShift = dominantShift(rows)
	return stats, nil
}

func meanStd(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	m := s / float64(len(xs))
	var sq float64
	for _, x := range xs {
		sq += (x - m) * (x - m)
	}
	return m, math.Sqrt(sq / float64(len(xs)))
}

func dominantShift(rows []repository.AggregateResult) string {
	byShift := map[string]float64{"morning": 0, "afternoon": 0, "evening": 0, "night": 0}
	for _, r := range rows {
		h := r.Bucket.Hour()
		switch {
		case h >= 6 && h < 12:
			byShift["morning"] += r.SumValue
		case h >= 12 && h < 18:
			byShift["afternoon"] += r.SumValue
		case h >= 18 && h < 22:
			byShift["evening"] += r.SumValue
		default:
			byShift["night"] += r.SumValue
		}
	}
	type kv struct {
		k string
		v float64
	}
	out := make([]kv, 0, 4)
	for k, v := range byShift {
		out = append(out, kv{k: k, v: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].v > out[j].v })
	return out[0].k
}

// TrendDecomposition separates consumption into trend/seasonal/residual
// using a naive moving-average filter (placeholder for STL/Prophet).
func (a *EnergyAnalytics) TrendDecomposition(values []float64, window int) (trend, seasonal, residual []float64) {
	if window < 2 || len(values) < window {
		return values, nil, nil
	}
	trend = make([]float64, len(values))
	seasonal = make([]float64, len(values))
	residual = make([]float64, len(values))
	for i := range values {
		lo := i - window/2
		hi := i + window/2
		if lo < 0 {
			lo = 0
		}
		if hi > len(values) {
			hi = len(values)
		}
		var s float64
		for j := lo; j < hi; j++ {
			s += values[j]
		}
		trend[i] = s / float64(hi-lo)
	}
	// Extract weekly seasonal via (value - trend) average per weekday-index, 7-period window.
	period := 7
	if window < period {
		period = window
	}
	avgByIdx := make(map[int]float64)
	cntByIdx := make(map[int]int)
	for i, v := range values {
		diff := v - trend[i]
		avgByIdx[i%period] += diff
		cntByIdx[i%period]++
	}
	for k, n := range cntByIdx {
		if n > 0 {
			avgByIdx[k] /= float64(n)
		}
	}
	for i, v := range values {
		seasonal[i] = avgByIdx[i%period]
		residual[i] = v - trend[i] - seasonal[i]
	}
	return trend, seasonal, residual
}
