// Package jobs — Asynq client + handler registration for async work (reports, refreshes).
//
// Doctrine refs: Rule 30 (process boundaries), Rule 36 (failure as normal),
//                Rule 37 (async to protect OLTP), Rule 40 (observability).
// Plan ADR: docs/adr/0014-async-report-generation-asynq.md.
//
// Handlers registered:
//
//   report:esrs_e1
//   report:piano_5_0
//   report:conto_termico
//   report:tee
//   report:audit_dlgs102
//   report:monthly_consumption
//   report:co2_footprint
//   factor:refresh_ispra
//   factor:refresh_terna
package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/greenmetrics/backend/internal/metrics"
	"github.com/hibiken/asynq"
)

// Job type constants — never inline literal strings at call sites.
const (
	TypeReportESRSE1            = "report:esrs_e1"
	TypeReportPiano5_0          = "report:piano_5_0"
	TypeReportContoTermico      = "report:conto_termico"
	TypeReportTEE               = "report:tee"
	TypeReportAuditDLgs102      = "report:audit_dlgs102"
	TypeReportMonthlyConsumption = "report:monthly_consumption"
	TypeReportCO2Footprint      = "report:co2_footprint"
	TypeFactorRefreshISPRA      = "factor:refresh_ispra"
	TypeFactorRefreshTerna      = "factor:refresh_terna"
)

// Queue identifiers.
const (
	QueueDefault   = "default"
	QueueReports   = "reports"
	QueueFactorRefresh = "factor-refresh"
)

// Client wraps asynq.Client. Construct once on app boot.
type Client struct {
	a *asynq.Client
}

// NewClient returns an Asynq client.
func NewClient(redisURL string) (*Client, error) {
	if redisURL == "" {
		return nil, errors.New("jobs: REDIS_URL required")
	}
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, err
	}
	return &Client{a: asynq.NewClient(opt)}, nil
}

// Close releases Redis resources.
func (c *Client) Close() error { return c.a.Close() }

// Enqueue inserts a job; returns the assigned job ID.
func (c *Client) Enqueue(ctx context.Context, typ string, payload any, opts ...asynq.Option) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	task := asynq.NewTask(typ, body)
	info, err := c.a.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return "", err
	}
	return info.ID, nil
}

// EnqueueReport is a convenience wrapper for report jobs with sensible defaults.
func (c *Client) EnqueueReport(ctx context.Context, typ, tenantID, idempotencyKey string, payload any) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	task := asynq.NewTask(typ, body, asynq.MaxRetry(3), asynq.Timeout(10*time.Minute), asynq.Queue(QueueReports), asynq.TaskID(idempotencyKey))
	info, err := c.a.EnqueueContext(ctx, task)
	if err != nil {
		return "", err
	}
	return info.ID, nil
}

// Server wraps asynq.Server for the worker process (cmd/worker).
type Server struct {
	s *asynq.Server
	mux *asynq.ServeMux
}

// NewServer constructs a worker server with sensible defaults.
func NewServer(redisURL string, concurrency int) (*Server, error) {
	if redisURL == "" {
		return nil, errors.New("jobs: REDIS_URL required")
	}
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, err
	}
	srv := asynq.NewServer(opt, asynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			QueueReports:        6, // weight
			QueueFactorRefresh:  2,
			QueueDefault:        2,
		},
		StrictPriority: true,
	})
	return &Server{s: srv, mux: asynq.NewServeMux()}, nil
}

// HandleFunc registers a handler for a job type.
func (s *Server) HandleFunc(typ string, fn func(ctx context.Context, t *asynq.Task) error) {
	s.mux.HandleFunc(typ, instrument(typ, fn))
}

// Run blocks serving until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	if err := s.s.Start(s.mux); err != nil {
		return err
	}
	<-ctx.Done()
	s.s.Shutdown()
	return ctx.Err()
}

// instrument wraps a handler with timing + result Prometheus metrics.
func instrument(typ string, fn func(ctx context.Context, t *asynq.Task) error) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		err := fn(ctx, t)
		result := "succeeded"
		if err != nil {
			result = "failed"
		}
		metrics.AsyncJobDuration.WithLabelValues(typ, result).Observe(time.Since(start).Seconds())
		return err
	}
}
