// Package resilience — shared HTTP client with bounded keep-alive pool.
//
// Doctrine refs: Rule 36, Rule 42 (resource lifecycle).
//
// Use this client for every outbound HTTP call instead of the default
// `http.Client{}`. Eliminates per-call connection storms.

package resilience

import (
	"net"
	"net/http"
	"time"
)

// HTTPClientOpts governs the shared client.
type HTTPClientOpts struct {
	Timeout              time.Duration // default 10s; per-request override via context
	MaxIdleConns         int           // default 100
	MaxIdleConnsPerHost  int           // default 10
	MaxConnsPerHost      int           // default 50
	IdleConnTimeout      time.Duration // default 90s
	DialTimeout          time.Duration // default 5s
	KeepAlive            time.Duration // default 30s
	TLSHandshakeTimeout  time.Duration // default 5s
	ResponseHeaderTimeout time.Duration // default 5s
	DisableKeepAlives    bool          // default false
}

func (o HTTPClientOpts) withDefaults() HTTPClientOpts {
	if o.Timeout == 0 {
		o.Timeout = 10 * time.Second
	}
	if o.MaxIdleConns == 0 {
		o.MaxIdleConns = 100
	}
	if o.MaxIdleConnsPerHost == 0 {
		o.MaxIdleConnsPerHost = 10
	}
	if o.MaxConnsPerHost == 0 {
		o.MaxConnsPerHost = 50
	}
	if o.IdleConnTimeout == 0 {
		o.IdleConnTimeout = 90 * time.Second
	}
	if o.DialTimeout == 0 {
		o.DialTimeout = 5 * time.Second
	}
	if o.KeepAlive == 0 {
		o.KeepAlive = 30 * time.Second
	}
	if o.TLSHandshakeTimeout == 0 {
		o.TLSHandshakeTimeout = 5 * time.Second
	}
	if o.ResponseHeaderTimeout == 0 {
		o.ResponseHeaderTimeout = 5 * time.Second
	}
	return o
}

// NewHTTPClient returns a tuned http.Client suitable for production outbound calls.
// Use otelhttp.NewTransport in main.go to wrap the returned Transport for tracing.
func NewHTTPClient(opts HTTPClientOpts) *http.Client {
	o := opts.withDefaults()
	dialer := &net.Dialer{Timeout: o.DialTimeout, KeepAlive: o.KeepAlive}
	return &http.Client{
		Timeout: o.Timeout,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			MaxIdleConns:          o.MaxIdleConns,
			MaxIdleConnsPerHost:   o.MaxIdleConnsPerHost,
			MaxConnsPerHost:       o.MaxConnsPerHost,
			IdleConnTimeout:       o.IdleConnTimeout,
			TLSHandshakeTimeout:   o.TLSHandshakeTimeout,
			ExpectContinueTimeout: 1 * time.Second,
			ResponseHeaderTimeout: o.ResponseHeaderTimeout,
			DisableKeepAlives:     o.DisableKeepAlives,
			ForceAttemptHTTP2:     true,
		},
	}
}
