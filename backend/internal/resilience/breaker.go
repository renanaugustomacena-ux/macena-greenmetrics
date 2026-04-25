// Package resilience — circuit breakers, retry policy, shared HTTP client.
//
// Doctrine refs: Rule 15, Rule 36, Rule 37.
// Plan ADR: docs/adr/0009-circuit-breakers-sony-gobreaker.md.
// Mitigates: RISK-004, RISK-008.
//
// Per-host breaker pattern: instantiate one breaker per upstream host.
// Settings: open after 5 consecutive failures or 50% over 20 calls; half-open
// after 30s; one trial probe.
package resilience

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sony/gobreaker/v2"
)

// Default settings — production tunable. Override via NewBreaker options.
const (
	DefaultMaxRequests         uint32        = 1
	DefaultInterval            time.Duration = 60 * time.Second
	DefaultTimeout             time.Duration = 30 * time.Second
	DefaultConsecutiveFailures uint32        = 5
	DefaultRatioOver           uint32        = 20
	DefaultRatioFailureRate    float64       = 0.5
)

// Breaker wraps gobreaker.CircuitBreaker with an observability hook.
type Breaker[T any] struct {
	cb       *gobreaker.CircuitBreaker[T]
	name     string
	stateObs StateObserver
}

// State mirrors gobreaker for clarity at call sites.
type State int

const (
	StateClosed   State = iota // 0
	StateOpen                  // 1
	StateHalfOpen              // 2
)

// StateObserver is invoked on every breaker state change. Wire to Prometheus
// gauge `gm_breaker_state{name}`.
type StateObserver func(name string, from, to State)

// Option configures a Breaker.
type Option func(*config)

type config struct {
	settings   gobreaker.Settings
	observer   StateObserver
}

// WithObserver registers a state-change callback (typically the Prometheus gauge updater).
func WithObserver(o StateObserver) Option {
	return func(c *config) { c.observer = o }
}

// WithSettings overrides the default gobreaker.Settings.
func WithSettings(s gobreaker.Settings) Option {
	return func(c *config) { c.settings = s }
}

// NewBreaker returns a new Breaker keyed by name. Name should be the upstream
// host or service identifier (e.g. `ispra_emission_factors`, `modbus_<host>`).
func NewBreaker[T any](name string, opts ...Option) *Breaker[T] {
	if name == "" {
		panic("resilience: breaker name required")
	}
	cfg := &config{
		settings: gobreaker.Settings{
			Name:        name,
			MaxRequests: DefaultMaxRequests,
			Interval:    DefaultInterval,
			Timeout:     DefaultTimeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				if counts.ConsecutiveFailures >= DefaultConsecutiveFailures {
					return true
				}
				if counts.Requests >= DefaultRatioOver {
					ratio := float64(counts.TotalFailures) / float64(counts.Requests)
					return ratio >= DefaultRatioFailureRate
				}
				return false
			},
		},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.settings.Name = name

	cfg.settings.OnStateChange = func(_ string, from, to gobreaker.State) {
		if cfg.observer != nil {
			cfg.observer(name, mapState(from), mapState(to))
		}
	}

	return &Breaker[T]{
		cb:       gobreaker.NewCircuitBreaker[T](cfg.settings),
		name:     name,
		stateObs: cfg.observer,
	}
}

// Execute runs fn through the breaker. On open, returns ErrBreakerOpen.
func (b *Breaker[T]) Execute(fn func() (T, error)) (T, error) {
	v, err := b.cb.Execute(fn)
	if errors.Is(err, gobreaker.ErrOpenState) {
		return v, fmt.Errorf("%w: %s", ErrBreakerOpen, b.name)
	}
	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		return v, fmt.Errorf("%w: %s", ErrBreakerHalfOpen, b.name)
	}
	return v, err
}

// Name returns the breaker identifier.
func (b *Breaker[T]) Name() string { return b.name }

// State returns the current breaker state.
func (b *Breaker[T]) State() State { return mapState(b.cb.State()) }

// Counts exposes underlying gobreaker counts (rolling window).
func (b *Breaker[T]) Counts() gobreaker.Counts { return b.cb.Counts() }

// ErrBreakerOpen indicates the breaker rejected the call (state == open).
var ErrBreakerOpen = errors.New("circuit breaker open")

// ErrBreakerHalfOpen indicates the half-open trial budget is exhausted.
var ErrBreakerHalfOpen = errors.New("circuit breaker half-open: try later")

func mapState(s gobreaker.State) State {
	switch s {
	case gobreaker.StateClosed:
		return StateClosed
	case gobreaker.StateOpen:
		return StateOpen
	case gobreaker.StateHalfOpen:
		return StateHalfOpen
	default:
		return StateClosed
	}
}

// --- Registry --------------------------------------------------------------

// Registry holds named breakers keyed by host. Use NewRegistry on app boot;
// pass to call sites that need a per-host breaker.
type Registry struct {
	mu       sync.RWMutex
	breakers map[string]any   // *Breaker[T] type-erased; lookup-only
	observer StateObserver
}

// NewRegistry returns an empty registry.
func NewRegistry(observer StateObserver) *Registry {
	return &Registry{breakers: make(map[string]any), observer: observer}
}

// GetOrCreate returns the typed breaker for `name`, creating if absent.
// Caller specifies T at the call site to match the wrapped function's return type.
func GetOrCreate[T any](r *Registry, name string, opts ...Option) *Breaker[T] {
	r.mu.RLock()
	if b, ok := r.breakers[name]; ok {
		r.mu.RUnlock()
		return b.(*Breaker[T])
	}
	r.mu.RUnlock()
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.breakers[name]; ok {
		return b.(*Breaker[T])
	}
	if r.observer != nil {
		opts = append(opts, WithObserver(r.observer))
	}
	b := NewBreaker[T](name, opts...)
	r.breakers[name] = b
	return b
}
