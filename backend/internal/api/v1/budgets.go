// Package v1 — per-endpoint downstream timeout budgets.
//
// Doctrine refs: Rule 36 (timeouts everywhere), Rule 37 (latency budgets).
//
// Use these budgets to derive `context.WithTimeout(ctx, BudgetX)` before
// invoking a downstream call. Budget table is the source of truth; do not
// hard-code timeouts at call sites.
package v1

import "time"

// Downstream call budgets — derive from `c.Context()` with `context.WithTimeout`.
const (
	// DBReadBudget — single-row + small-set SELECTs.
	DBReadBudget = 500 * time.Millisecond

	// DBWriteBudget — INSERT/UPDATE/DELETE single statement.
	DBWriteBudget = 1 * time.Second

	// DBLongAggregationBudget — CAGG-driven aggregated reads, report-time queries.
	DBLongAggregationBudget = 5 * time.Second

	// DBTxBudget — multi-statement Tx (default; report-time uses LongAggregation).
	DBTxBudget = 5 * time.Second

	// ExternalHTTPBudget — outbound HTTP wrapped by breaker + retry.
	ExternalHTTPBudget = 3 * time.Second

	// ModbusPollBudget — single Modbus poll cycle.
	ModbusPollBudget = 3 * time.Second

	// MBusFrameBudget — single M-Bus frame read.
	MBusFrameBudget = 3 * time.Second

	// OCPPMessageBudget — single OCPP message round-trip.
	OCPPMessageBudget = 5 * time.Second

	// JobEnqueueBudget — Asynq client enqueue.
	JobEnqueueBudget = 1 * time.Second
)
