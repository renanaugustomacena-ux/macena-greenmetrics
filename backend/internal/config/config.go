// Package config loads 12-factor environment configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// DefaultSentinelJWTSecret is the placeholder that must never reach production.
const DefaultSentinelJWTSecret = "change-me-in-production-do-not-use-default"

// DefaultSentinelGrafanaPassword is the default Grafana password that must be rotated.
const DefaultSentinelGrafanaPassword = "change-me"

// Config centralises all runtime configuration.
type Config struct {
	AppEnv  string
	AppPort string

	DatabaseURL          string
	GrafanaURL           string
	GrafanaAdminUser     string
	GrafanaAdminPassword string

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	OTelServiceName      string
	OTelExporterEndpoint string
	OTelSampleRatio      float64

	ISPRAEmissionFactorsURL string
	TernaAPIBase            string
	EDistribuzioneAPIBase   string

	ModbusTCPTimeoutMS int
	MBusSerialDevice   string
	MBusBaud           int
	MBusPollIntervalS  int
	SPDClientCertPath  string

	ModbusSimulatorAddr string
	ModbusPollIntervalS int
	ModbusSlaveIDs      string

	OCPPCentralSystemURL string
	OCPPChargePointID    string

	SunSpecAddr         string
	SunSpecSlaveID      int
	SunSpecBaseRegister int

	PulseWebhookSecret string
	PulseDedupeWindow  time.Duration

	RateLimitPerMinute      int
	RateLimitLoginPerMinute int
	RateLimitIngestPerMinute int

	LockoutThreshold int
	LockoutWindowMin int

	SessionIdleMinutes    int
	SessionAbsoluteHours  int

	CORSAllowedOrigins string
}

// Load reads env vars (optionally from .env) and returns a populated Config.
func Load() (*Config, error) {
	// .env is best-effort; production relies on real env.
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                  getEnv("APP_ENV", "development"),
		AppPort:                 getEnv("APP_PORT", "8082"),
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://greenmetrics:greenmetrics@localhost:5439/greenmetrics?sslmode=prefer"),
		GrafanaURL:              getEnv("GRAFANA_URL", "http://localhost:3011"),
		GrafanaAdminUser:        getEnv("GRAFANA_ADMIN_USER", "admin"),
		GrafanaAdminPassword:    getEnv("GRAFANA_ADMIN_PASSWORD", DefaultSentinelGrafanaPassword),
		JWTSecret:               getEnv("JWT_SECRET", DefaultSentinelJWTSecret),
		OTelServiceName:         getEnv("OTEL_SERVICE_NAME", "greenmetrics-backend"),
		OTelExporterEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		ISPRAEmissionFactorsURL: getEnv("ISPRA_EMISSION_FACTORS_URL", "https://www.isprambiente.gov.it/it/pubblicazioni/rapporti/fattori-di-emissione"),
		TernaAPIBase:            getEnv("TERNA_API_BASE", "https://api.terna.it"),
		EDistribuzioneAPIBase:   getEnv("E_DISTRIBUZIONE_API_BASE", "https://api.e-distribuzione.it"),
		MBusSerialDevice:        getEnv("MBUS_SERIAL_DEVICE", ""),
		SPDClientCertPath:       getEnv("SPD_CLIENT_CERT_PATH", "/etc/greenmetrics/spd-client.pem"),
		ModbusSimulatorAddr:     getEnv("MODBUS_SIMULATOR_ADDR", ""),
		ModbusSlaveIDs:          getEnv("MODBUS_SLAVE_IDS", "1,2,3,4,5"),
		OCPPCentralSystemURL:    getEnv("OCPP_CENTRAL_SYSTEM_URL", ""),
		OCPPChargePointID:       getEnv("OCPP_CHARGE_POINT_ID", "greenmetrics-cp-0001"),
		SunSpecAddr:             getEnv("SUNSPEC_MODBUS_ADDR", ""),
		PulseWebhookSecret:      getEnv("PULSE_WEBHOOK_SECRET", ""),
		CORSAllowedOrigins:      getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3005,http://localhost:5173"),
	}

	var err error
	cfg.JWTAccessTTL, err = time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}
	cfg.JWTRefreshTTL, err = time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_TTL: %w", err)
	}
	cfg.ModbusTCPTimeoutMS, err = strconv.Atoi(getEnv("MODBUS_TCP_TIMEOUT_MS", "3000"))
	if err != nil {
		return nil, fmt.Errorf("MODBUS_TCP_TIMEOUT_MS: %w", err)
	}
	cfg.ModbusPollIntervalS, err = strconv.Atoi(getEnv("MODBUS_POLL_INTERVAL_S", "30"))
	if err != nil {
		return nil, fmt.Errorf("MODBUS_POLL_INTERVAL_S: %w", err)
	}
	cfg.MBusBaud, err = strconv.Atoi(getEnv("MBUS_BAUD", "2400"))
	if err != nil {
		return nil, fmt.Errorf("MBUS_BAUD: %w", err)
	}
	cfg.MBusPollIntervalS, err = strconv.Atoi(getEnv("MBUS_POLL_INTERVAL_S", "300"))
	if err != nil {
		return nil, fmt.Errorf("MBUS_POLL_INTERVAL_S: %w", err)
	}
	cfg.SunSpecSlaveID, err = strconv.Atoi(getEnv("SUNSPEC_SLAVE_ID", "126"))
	if err != nil {
		return nil, fmt.Errorf("SUNSPEC_SLAVE_ID: %w", err)
	}
	cfg.SunSpecBaseRegister, err = strconv.Atoi(getEnv("SUNSPEC_BASE_REGISTER", "40000"))
	if err != nil {
		return nil, fmt.Errorf("SUNSPEC_BASE_REGISTER: %w", err)
	}
	cfg.PulseDedupeWindow, err = time.ParseDuration(getEnv("PULSE_DEDUPE_WINDOW", "24h"))
	if err != nil {
		return nil, fmt.Errorf("PULSE_DEDUPE_WINDOW: %w", err)
	}
	cfg.RateLimitPerMinute, err = strconv.Atoi(getEnv("RATE_LIMIT_PER_MINUTE", "60"))
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_PER_MINUTE: %w", err)
	}
	cfg.RateLimitLoginPerMinute, err = strconv.Atoi(getEnv("RATE_LIMIT_LOGIN_PER_MINUTE", "5"))
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_LOGIN_PER_MINUTE: %w", err)
	}
	cfg.RateLimitIngestPerMinute, err = strconv.Atoi(getEnv("RATE_LIMIT_INGEST_PER_MINUTE", "300"))
	if err != nil {
		return nil, fmt.Errorf("RATE_LIMIT_INGEST_PER_MINUTE: %w", err)
	}
	cfg.LockoutThreshold, err = strconv.Atoi(getEnv("AUTH_LOCKOUT_THRESHOLD", "5"))
	if err != nil {
		return nil, fmt.Errorf("AUTH_LOCKOUT_THRESHOLD: %w", err)
	}
	cfg.LockoutWindowMin, err = strconv.Atoi(getEnv("AUTH_LOCKOUT_WINDOW_MIN", "15"))
	if err != nil {
		return nil, fmt.Errorf("AUTH_LOCKOUT_WINDOW_MIN: %w", err)
	}
	cfg.SessionIdleMinutes, err = strconv.Atoi(getEnv("SESSION_IDLE_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("SESSION_IDLE_MINUTES: %w", err)
	}
	cfg.SessionAbsoluteHours, err = strconv.Atoi(getEnv("SESSION_ABSOLUTE_HOURS", "12"))
	if err != nil {
		return nil, fmt.Errorf("SESSION_ABSOLUTE_HOURS: %w", err)
	}
	ratioStr := getEnv("OTEL_SAMPLE_RATIO", "0.1")
	cfg.OTelSampleRatio, err = strconv.ParseFloat(ratioStr, 64)
	if err != nil {
		return nil, fmt.Errorf("OTEL_SAMPLE_RATIO: %w", err)
	}

	// Refuse to boot in production with default secrets — hard security gate.
	if strings.EqualFold(cfg.AppEnv, "production") {
		if cfg.JWTSecret == DefaultSentinelJWTSecret || strings.TrimSpace(cfg.JWTSecret) == "" {
			return nil, fmt.Errorf("refusing default JWT_SECRET in production — set a cryptographically random value")
		}
		if cfg.GrafanaAdminPassword == DefaultSentinelGrafanaPassword || strings.TrimSpace(cfg.GrafanaAdminPassword) == "" {
			return nil, fmt.Errorf("refusing default GRAFANA_ADMIN_PASSWORD in production — rotate Grafana admin credentials")
		}
		if strings.Contains(cfg.DatabaseURL, "sslmode=disable") {
			return nil, fmt.Errorf("refusing DATABASE_URL with sslmode=disable in production — require TLS to TimescaleDB")
		}
	}

	// In any env, a JWT_SECRET must be at least 32 bytes effective entropy.
	if len(cfg.JWTSecret) < 32 {
		// Warn — do not hard-fail in dev (orchestrator convenience).
		if strings.EqualFold(cfg.AppEnv, "production") {
			return nil, fmt.Errorf("JWT_SECRET too short (%d bytes); require ≥ 32", len(cfg.JWTSecret))
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
