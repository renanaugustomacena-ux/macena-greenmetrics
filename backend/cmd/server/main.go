// Package main bootstraps the GreenMetrics Fiber HTTP server.
//
// Responsibilities:
//   - Load configuration from environment (12-factor).
//   - Initialise structured JSON logging via zap.
//   - Initialise OTLP trace exporter (OpenTelemetry).
//   - Construct repositories/services.
//   - Wire routes on a Fiber app.
//   - Start HTTP listener with graceful shutdown on SIGINT/SIGTERM.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/greenmetrics/backend/internal/config"
	"github.com/greenmetrics/backend/internal/handlers"
	"github.com/greenmetrics/backend/internal/repository"
	"github.com/greenmetrics/backend/internal/services"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	// Support --healthcheck CLI flag used by Docker HEALTHCHECK.
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		runHealthcheck()
		return
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config.Load: %w", err)
	}

	logger, err := newLogger(cfg.AppEnv)
	if err != nil {
		return fmt.Errorf("newLogger: %w", err)
	}
	defer func() { _ = logger.Sync() }()
	logger = logger.With(
		zap.String("service", cfg.OTelServiceName),
		zap.String("env", cfg.AppEnv),
		zap.String("version", version),
		zap.String("commit", commit),
	)

	logger.Info("greenmetrics-backend starting",
		zap.String("port", cfg.AppPort),
		zap.String("database", maskDSN(cfg.DatabaseURL)),
	)

	// ---- OpenTelemetry tracer provider ----------------------------------
	shutdownTracer, err := initTracer(cfg, logger)
	if err != nil {
		logger.Warn("tracer init failed, continuing without traces", zap.Error(err))
	}
	defer func() {
		if shutdownTracer == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracer(ctx); err != nil {
			logger.Warn("tracer shutdown error", zap.Error(err))
		}
	}()

	// ---- Repository (TimescaleDB) ---------------------------------------
	repo, err := repository.NewTimescaleRepository(context.Background(), cfg.DatabaseURL, logger)
	if err != nil {
		logger.Warn("timescale connection deferred (placeholder)", zap.Error(err))
	}

	// ---- Services -------------------------------------------------------
	analytics := services.NewEnergyAnalytics(repo, logger)
	carbon := services.NewCarbonCalculator(repo, logger)
	reporter := services.NewReportGenerator(repo, carbon, analytics, logger)
	alerts := services.NewAlertEngine(repo, logger)

	// ---- Ingestor runner (Modbus / M-Bus / SunSpec) ---------------------
	slaveIDs, err := services.ParseSlaveIDs(cfg.ModbusSlaveIDs)
	if err != nil {
		return fmt.Errorf("MODBUS_SLAVE_IDS parse: %w", err)
	}
	ingestorCfg := services.IngestorConfig{
		ModbusAddr:      cfg.ModbusSimulatorAddr,
		ModbusPollEvery: time.Duration(cfg.ModbusPollIntervalS) * time.Second,
		ModbusSlaveIDs:  slaveIDs,
		ModbusTimeout:   time.Duration(cfg.ModbusTCPTimeoutMS) * time.Millisecond,
		MBusDevice:      cfg.MBusSerialDevice,
		MBusBaud:        cfg.MBusBaud,
		MBusInterval:    time.Duration(cfg.MBusPollIntervalS) * time.Second,
	}
	if cfg.SunSpecAddr != "" {
		ingestorCfg.SunSpecConfig = &services.SunSpecConfig{
			Address:      cfg.SunSpecAddr,
			SlaveID:      byte(cfg.SunSpecSlaveID),
			BaseRegister: uint16(cfg.SunSpecBaseRegister),
			PollInterval: 30 * time.Second,
		}
	}
	ingestor := services.NewIngestorRunner(ingestorCfg, repo, logger)
	ingestorCtx, ingestorCancel := context.WithCancel(context.Background())
	defer ingestorCancel()
	if errStart := ingestor.Start(ingestorCtx); errStart != nil {
		if errors.Is(errStart, services.ErrIngestorDisabled) {
			logger.Info("ingestor disabled (no MODBUS_SIMULATOR_ADDR / MBUS_SERIAL_DEVICE / SUNSPEC_MODBUS_ADDR)")
		} else {
			logger.Warn("ingestor start issue", zap.Error(errStart))
		}
	}

	// Optional OCPP client wiring: only constructed if central-system URL set.
	// The Dial is deferred to a goroutine so a missing CS does not block boot.
	if cfg.OCPPCentralSystemURL != "" {
		if _, errOCPP := services.NewOCPPClient(cfg.OCPPCentralSystemURL, cfg.OCPPChargePointID, logger); errOCPP != nil {
			logger.Warn("ocpp client init failed", zap.Error(errOCPP))
		} else {
			logger.Info("ocpp client configured (dial deferred to first use)",
				zap.String("central_system", cfg.OCPPCentralSystemURL),
				zap.String("charge_point_id", cfg.OCPPChargePointID),
			)
		}
	}

	// ---- Fiber app ------------------------------------------------------
	app := fiber.New(fiber.Config{
		AppName:               "greenmetrics-backend",
		DisableStartupMessage: true,
		ErrorHandler:          handlers.ErrorHandler(logger),
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          20 * time.Second,
		IdleTimeout:           60 * time.Second,
	})

	app.Use(recover.New())
	app.Use(requestid.New(requestid.Config{Header: "X-Request-ID"}))
	app.Use(otelfiber.Middleware())
	app.Use(compress.New())
	app.Use(handlers.SecurityHeaders(cfg.AppEnv == "production"))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSAllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Request-ID, X-CSRF-Token",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	// Root-level /metrics for Prometheus scraping (plan §13 + GAPS B-01).
	app.Get("/metrics", handlers.MetricsHandler)

	deps := handlers.Dependencies{
		Config:    cfg,
		Logger:    logger,
		Repo:      repo,
		Analytics: analytics,
		Carbon:    carbon,
		Reporter:  reporter,
		Alerts:    alerts,
		Version:   version,
		Commit:    commit,
		StartedAt: time.Now().UTC(),
	}

	handlers.Register(app, deps)

	// ---- Graceful shutdown ----------------------------------------------
	errCh := make(chan error, 1)
	go func() {
		addr := ":" + cfg.AppPort
		logger.Info("http server listening",
			zap.String("address", addr),
			zap.String("env", cfg.AppEnv),
		)
		if err := app.Listen(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stopCh:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		logger.Error("server error", zap.Error(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("graceful shutdown error", zap.Error(err))
	}
	if repo != nil {
		repo.Close()
	}
	logger.Info("greenmetrics-backend stopped cleanly")
	return nil
}

func newLogger(env string) (*zap.Logger, error) {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "message"
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}
	return cfg.Build()
}

func initTracer(cfg *config.Config, logger *zap.Logger) (func(context.Context) error, error) {
	if cfg.OTelExporterEndpoint == "" {
		logger.Info("OTLP endpoint not set; tracer disabled")
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	exp, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(cfg.OTelExporterEndpoint),
		otlptracegrpc.WithInsecure(),
	))
	if err != nil {
		return nil, err
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.OTelServiceName),
			semconv.ServiceVersion(version),
			semconv.DeploymentEnvironment(cfg.AppEnv),
		),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.OTelSampleRatio))),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

func runHealthcheck() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/api/health")
	if err != nil || resp.StatusCode >= 500 {
		os.Exit(1)
	}
	_ = resp.Body.Close()
}

func maskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	// Just hide credentials — good enough for logs.
	for i := 0; i < len(dsn)-1; i++ {
		if dsn[i] == ':' && i+2 < len(dsn) && dsn[i+1] == '/' && dsn[i+2] == '/' {
			j := i + 3
			for j < len(dsn) && dsn[j] != '@' {
				j++
			}
			if j < len(dsn) {
				return dsn[:i+3] + "***:***" + dsn[j:]
			}
			break
		}
	}
	return dsn
}
