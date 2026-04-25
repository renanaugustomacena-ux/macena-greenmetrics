// Command worker runs Asynq job consumers for async report generation,
// emission-factor refresh, and other background work.
//
// Doctrine refs: Rule 30 (process boundary), Rule 37 (isolate from OLTP),
//                Rule 41 (concurrency discipline), Rule 60 (incident response).
//
// Process model: separate from cmd/server. Same image, different command.
// Identical config + observability boot path.
//
// Run: `worker` (no args; reads env). Graceful shutdown on SIGTERM.

package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/greenmetrics/backend/internal/config"
	"github.com/greenmetrics/backend/internal/jobs"
	"github.com/greenmetrics/backend/internal/metrics"
	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	metrics.Register(prometheus.DefaultRegisterer)

	concurrency := runtime.GOMAXPROCS(0) * 2
	if concurrency < 4 {
		concurrency = 4
	}

	srv, err := jobs.NewServer(os.Getenv("REDIS_URL"), concurrency)
	if err != nil {
		log.Fatalf("jobs: %v", err)
	}

	// --- handler registration ---
	// Each handler is a thin shim into the domain reporting / factor packages
	// (lands S5 with the DDD split). For now, register stubs that log + return.

	srv.HandleFunc(jobs.TypeReportESRSE1, stubHandler("report:esrs_e1"))
	srv.HandleFunc(jobs.TypeReportPiano5_0, stubHandler("report:piano_5_0"))
	srv.HandleFunc(jobs.TypeReportContoTermico, stubHandler("report:conto_termico"))
	srv.HandleFunc(jobs.TypeReportTEE, stubHandler("report:tee"))
	srv.HandleFunc(jobs.TypeReportAuditDLgs102, stubHandler("report:audit_dlgs102"))
	srv.HandleFunc(jobs.TypeReportMonthlyConsumption, stubHandler("report:monthly_consumption"))
	srv.HandleFunc(jobs.TypeReportCO2Footprint, stubHandler("report:co2_footprint"))
	srv.HandleFunc(jobs.TypeFactorRefreshISPRA, stubHandler("factor:refresh_ispra"))
	srv.HandleFunc(jobs.TypeFactorRefreshTerna, stubHandler("factor:refresh_terna"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("worker starting: env=%s, concurrency=%d", cfg.AppEnv, concurrency)
	if err := srv.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("worker: %v", err)
	}

	// Allow in-flight jobs up to 30s to drain.
	time.Sleep(2 * time.Second)
	log.Printf("worker stopped cleanly")
}

func stubHandler(typ string) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		// TODO renan (S5): wire to internal/domain/reporting/<dossier>.Build.
		log.Printf("[stub] %s job %s payload=%dB", typ, t.ResultWriter().TaskID(), len(t.Payload()))
		return nil
	}
}
