// Package observability — zap logger with trace context propagation.
//
// Doctrine refs: Rule 18, Rule 40, Rule 58.
// Mitigates: log/trace correlation gap.
//
// Usage in handlers:
//
//   log := obs.Logger(c.Context())
//   log.Info("ingest accepted", zap.Int("count", n))
//
// The returned logger carries `trace_id`, `span_id`, `request_id`, `tenant_id`
// fields automatically.
package observability

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Base is the process-wide structured logger. Set once on boot from
// `cmd/server/main.go`.
var base *zap.Logger

// SetBase records the process logger; call once during boot before any handler runs.
func SetBase(l *zap.Logger) {
	base = l
}

// Logger returns a child of the base logger with trace + tenant fields injected
// from ctx. Safe to call with a nil/background ctx; returns the base logger if
// no trace context is present.
func Logger(ctx context.Context) *zap.Logger {
	if base == nil {
		return zap.NewNop()
	}
	if ctx == nil {
		return base
	}
	fields := contextFields(ctx)
	if len(fields) == 0 {
		return base
	}
	return base.With(fields...)
}

func contextFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 4)
	if span := trace.SpanFromContext(ctx); span != nil {
		sc := span.SpanContext()
		if sc.HasTraceID() {
			fields = append(fields, zap.String("trace_id", sc.TraceID().String()))
		}
		if sc.HasSpanID() {
			fields = append(fields, zap.String("span_id", sc.SpanID().String()))
		}
	}
	if v, ok := ctx.Value(KeyRequestID).(string); ok && v != "" {
		fields = append(fields, zap.String("request_id", v))
	}
	if v, ok := ctx.Value(KeyTenantID).(string); ok && v != "" {
		fields = append(fields, zap.String("tenant_id", v))
	}
	if v, ok := ctx.Value(KeyUserEmail).(string); ok && v != "" {
		fields = append(fields, zap.String("user_email", v))
	}
	return fields
}

// Context keys — opaque to keep type-safe.
type ctxKey int

const (
	KeyRequestID ctxKey = iota
	KeyTenantID
	KeyUserEmail
)

// WithRequestID returns ctx with the request id stamped.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, KeyRequestID, id)
}

// WithTenantID returns ctx with the tenant id stamped.
func WithTenantID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, KeyTenantID, id)
}

// WithUserEmail returns ctx with the user email stamped.
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, KeyUserEmail, email)
}
