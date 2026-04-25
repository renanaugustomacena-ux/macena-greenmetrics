// Package resilience — retry policy with exponential backoff + jitter.
//
// Doctrine refs: Rule 36 (failure as normal), Rule 37 (latency budget aware).
//
// Use only on idempotent verbs (or with Idempotency-Key header).
// Never retry on 4xx (except 429 with Retry-After).

package resilience

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// RetryOpts governs retry behaviour.
type RetryOpts struct {
	MaxAttempts     uint64        // default 3
	InitialInterval time.Duration // default 100 ms
	MaxInterval     time.Duration // default 2 s
	Multiplier      float64       // default 2.0
	RandomizationFactor float64   // default 0.1
}

func (o RetryOpts) withDefaults() RetryOpts {
	if o.MaxAttempts == 0 {
		o.MaxAttempts = 3
	}
	if o.InitialInterval == 0 {
		o.InitialInterval = 100 * time.Millisecond
	}
	if o.MaxInterval == 0 {
		o.MaxInterval = 2 * time.Second
	}
	if o.Multiplier == 0 {
		o.Multiplier = 2.0
	}
	if o.RandomizationFactor == 0 {
		o.RandomizationFactor = 0.1
	}
	return o
}

// Do retries fn until it succeeds, the context is cancelled, or MaxAttempts is reached.
// fn must return ErrPermanent (or wrap it) to abort retry.
func Do[T any](ctx context.Context, opts RetryOpts, fn func(ctx context.Context) (T, error)) (T, error) {
	opts = opts.withDefaults()
	bo := backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(opts.InitialInterval),
		backoff.WithMaxInterval(opts.MaxInterval),
		backoff.WithMultiplier(opts.Multiplier),
		backoff.WithRandomizationFactor(opts.RandomizationFactor),
	)
	var attempts uint64
	op := func() (T, error) {
		attempts++
		v, err := fn(ctx)
		if err == nil {
			return v, nil
		}
		if errors.Is(err, ErrPermanent) {
			return v, backoff.Permanent(err)
		}
		if attempts >= opts.MaxAttempts {
			return v, backoff.Permanent(err)
		}
		return v, err
	}
	return backoff.Retry(ctx, op, backoff.WithBackOff(bo))
}

// ErrPermanent — wrap any error caller deems unretryable. backoff.Permanent
// wrapping in Do() honours this signal.
var ErrPermanent = errors.New("permanent error: do not retry")

// IsRetryable returns true if the HTTP status code is safely retried.
func IsRetryable(status int) bool {
	switch status {
	case http.StatusRequestTimeout, // 408
		http.StatusTooManyRequests,                   // 429
		http.StatusInternalServerError,               // 500
		http.StatusBadGateway,                        // 502
		http.StatusServiceUnavailable,                // 503
		http.StatusGatewayTimeout:                    // 504
		return true
	}
	return false
}

// RetryAfterHeader parses the Retry-After header (seconds or HTTP-date) into a
// duration. Returns 0 if absent/unparseable; caller falls back to backoff.
func RetryAfterHeader(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}
