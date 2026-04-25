package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SmartMeterClient is the integration layer for Italian DSO data portals:
//   - E-Distribuzione "Servizio di Misura" (portal SMD / Chain2) exposing
//     consumption curves per POD (Point of Delivery) at 15-min granularity,
//     available typically D+1 via the Portale Produttori / Portale Clienti.
//   - Terna "Gaudi" registry + hourly energy bid data for larger producers.
//   - SPD (Servizio Portale Distribuzione) — aggregated API for multi-DSO
//     (Unareti, ARETI, Iren Distribuzione, E-Distribuzione) in production.
//
// This client ships as a documented placeholder: the real integrations require
// qualified credentials (client-certificate for SPD) and are wired in during
// onboarding per-tenant.
type SmartMeterClient struct {
	logger            *zap.Logger
	ternaBase         string
	eDistribuzioneBase string
	spdCertPath       string
	http              *http.Client
}

// NewSmartMeterClient builds the client.
func NewSmartMeterClient(logger *zap.Logger, ternaBase, eDistribuzioneBase, spdCertPath string) *SmartMeterClient {
	return &SmartMeterClient{
		logger:             logger,
		ternaBase:          ternaBase,
		eDistribuzioneBase: eDistribuzioneBase,
		spdCertPath:        spdCertPath,
		http:               &http.Client{Timeout: 15 * time.Second},
	}
}

// FetchEDistribuzioneCurve returns the 15-min curve for a POD over [from, to].
// Placeholder: documented and stubbed.
func (s *SmartMeterClient) FetchEDistribuzioneCurve(ctx context.Context, pod string, from, to time.Time) ([]CurvePoint, error) {
	if pod == "" {
		return nil, errors.New("empty POD")
	}
	// Expected endpoint shape (contract from E-Distribuzione Portale SMD B2B):
	//   GET {eDistribuzioneBase}/pod/{pod}/curva15min?from={ISO8601}&to={ISO8601}
	// Authentication: OAuth2 client_credentials against the DSO SPID-federated IdP.
	// Response: JSON array of {timestamp, active_energy_kwh, quality_code}.
	s.logger.Info("e-distribuzione curve stub",
		zap.String("pod", pod),
		zap.Time("from", from),
		zap.Time("to", to),
	)
	return []CurvePoint{}, nil
}

// FetchTernaNationalMix returns the national hourly mix (placeholder).
func (s *SmartMeterClient) FetchTernaNationalMix(ctx context.Context, day time.Time) (*NationalMix, error) {
	// Terna Download Centre — /download-center/generazione endpoint series.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/transparency-platform/v1/generation/national/%s", s.ternaBase, day.Format("2006-01-02")), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.http.Do(req)
	if err != nil {
		// Non-fatal in stub mode.
		return &NationalMix{Date: day, Renewable: 0.35, Thermal: 0.55, Nuclear: 0.0, Imports: 0.10, Source: "placeholder"}, nil
	}
	defer resp.Body.Close()
	var out NationalMix
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return &NationalMix{Date: day, Renewable: 0.35, Thermal: 0.55, Nuclear: 0.0, Imports: 0.10, Source: "placeholder"}, nil
	}
	return &out, nil
}

// CurvePoint is a 15-min energy sample.
type CurvePoint struct {
	Timestamp      time.Time `json:"timestamp"`
	ActiveKWh      float64   `json:"active_energy_kwh"`
	ReactiveKVArh  float64   `json:"reactive_energy_kvarh"`
	QualityCode    int       `json:"quality_code"`
}

// NationalMix is a daily energy-generation share breakdown (Terna).
type NationalMix struct {
	Date      time.Time `json:"date"`
	Renewable float64   `json:"renewable_share"`
	Thermal   float64   `json:"thermal_share"`
	Nuclear   float64   `json:"nuclear_share"`
	Imports   float64   `json:"imports_share"`
	Source    string    `json:"source"`
}
