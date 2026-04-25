// Package handlers — account lockout tracker.
//
// Covers GreenMetrics-GAPS B-06 / G-04: no failed-attempt counter + backoff.
// We keep the tracker in-process (not Redis) because the Go runtime is
// single-instance in the MVP; a second-tier operator deploying more than one
// replica needs to front auth with a sticky load balancer or swap this for a
// shared store. That limitation is documented explicitly.
package handlers

import (
	"sync"
	"time"
)

// LockoutEntry captures per-identifier state.
type LockoutEntry struct {
	Failures    int
	LastFailure time.Time
	LockedUntil time.Time
}

// LockoutTracker counts failed logins per email + IP identifier.
type LockoutTracker struct {
	threshold  int
	windowMin  int
	baseBackoff time.Duration
	maxBackoff  time.Duration

	mu      sync.Mutex
	entries map[string]*LockoutEntry
}

// NewLockoutTracker constructs a tracker.
//
//	threshold   — consecutive failures before lockout engages
//	windowMin   — minutes to retain a failure record
func NewLockoutTracker(threshold, windowMin int) *LockoutTracker {
	if threshold <= 0 {
		threshold = 5
	}
	if windowMin <= 0 {
		windowMin = 15
	}
	return &LockoutTracker{
		threshold:   threshold,
		windowMin:   windowMin,
		baseBackoff: 2 * time.Second,
		maxBackoff:  15 * time.Minute,
		entries:     make(map[string]*LockoutEntry),
	}
}

// IsLocked returns true if the identifier is currently locked out.
func (t *LockoutTracker) IsLocked(key string) (bool, time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[key]
	if !ok {
		return false, 0
	}
	now := time.Now()
	if e.LockedUntil.After(now) {
		return true, e.LockedUntil.Sub(now)
	}
	// Window expired → reset.
	if now.Sub(e.LastFailure) > time.Duration(t.windowMin)*time.Minute {
		delete(t.entries, key)
		return false, 0
	}
	return false, 0
}

// RecordFailure increments the counter and engages lockout if threshold hit.
func (t *LockoutTracker) RecordFailure(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[key]
	if !ok {
		e = &LockoutEntry{}
		t.entries[key] = e
	}
	e.Failures++
	e.LastFailure = time.Now()
	if e.Failures >= t.threshold {
		// Exponential backoff: 2^(failures - threshold) * base, capped.
		excess := e.Failures - t.threshold
		back := t.baseBackoff << uint(excess)
		if back <= 0 || back > t.maxBackoff {
			back = t.maxBackoff
		}
		e.LockedUntil = time.Now().Add(back)
	}
}

// RecordSuccess clears the counter for an identifier.
func (t *LockoutTracker) RecordSuccess(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, key)
}
