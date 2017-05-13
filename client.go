// Package instrumented_http provides a drop-in metrics-enabled replacement for
// any http.Client or http.RoundTripper.
package instrumented_http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// InstrumentedRoundTripper is a http.RoundTripper that collects Prometheus
// metrics of every request it processes. It allows to be configured with
// callbacks that process request path and query into a suitable label value.
type InstrumentedRoundTripper struct {
	Transport http.RoundTripper
	Callbacks *Callbacks
}

// Callbacks is a collection of callbacks passed to InstrumentedRoundTripper.
type Callbacks struct {
	PathCallback  func(string) string
	QueryCallback func(string) string
}

const (
	// Metrics created can be identified by this label value.
	handlerName = "instrumented_http"
)

var (
	// RequestDurationSeconds is a Prometheus summary to collect request times.
	RequestDurationSeconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:        "request_duration_microseconds",
			Help:        "The HTTP request latencies in microseconds.",
			Subsystem:   "http",
			ConstLabels: prometheus.Labels{"handler": handlerName},
			Objectives:  map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"scheme", "host", "path", "query", "method", "status"},
	)

	// EliminatingCallback is a callback that returns a blank string on any input.
	EliminatingCallback = func(_ string) string { return "" }
	// IdentityCallback is callback that returns whatever is passed to it.
	IdentityCallback = func(input string) string { return input }
)

// init registers the Prometheus metric globally when the package is loaded.
func init() {
	prometheus.MustRegister(RequestDurationSeconds)
}

// RoundTrip implements http.RoundTripper. It forwards the request to the
// wrapped transport and measures the time it took in Prometheus summary.
func (it *InstrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Remember the current time.
	now := time.Now()

	// Make the request using the wrapped RoundTripper.
	resp, err := it.Transport.RoundTrip(req)

	// Observe the time it took to make the request.
	RequestDurationSeconds.WithLabelValues(
		req.URL.Scheme,
		req.URL.Host,
		it.Callbacks.PathCallback(req.URL.Path),
		it.Callbacks.QueryCallback(req.URL.RawQuery),
		req.Method,
		fmt.Sprintf("%d", resp.StatusCode),
	).Observe(float64(time.Since(now).Nanoseconds() / 1000))

	// return the response and error reported from the wrapped RoundTripper.
	return resp, err
}

// NewInstrumentedClient takes a *http.Client and returns a *http.Client that
// has its RoundTripper wrapped with instrumentation. Optionally, It
// can receive a collection of callbacks that process request path and query
// into a suitable label value.
func NewInstrumentedClient(next *http.Client, callbacks *Callbacks) *http.Client {
	// If wrapped client is not defined we'll use http.DefaultClient.
	if next == nil {
		next = http.DefaultClient
	}

	return &http.Client{Transport: NewInstrumentedRoundTripper(next.Transport, callbacks)}
}

// NewInstrumentedRoundTripper takes a http.RoundTripper, wraps it with
// instrumentation and returns it as a new http.RoundTripper. Optionally, It
// can receive a collection of callbacks that process request path and query
// into a suitable label value.
func NewInstrumentedRoundTripper(next http.RoundTripper, callbacks *Callbacks) http.RoundTripper {
	// If wrapped RoundTripper is not defined we'll use http.DefaultTransport.
	if next == nil {
		next = http.DefaultTransport
	}

	// If callbacks is not defined we'll initilialize it with defaults.
	if callbacks == nil {
		callbacks = &Callbacks{}
	}
	// By default, path and query will be ignored.
	if callbacks.PathCallback == nil {
		callbacks.PathCallback = EliminatingCallback
	}
	if callbacks.QueryCallback == nil {
		callbacks.QueryCallback = EliminatingCallback
	}

	return &InstrumentedRoundTripper{Transport: next, Callbacks: callbacks}
}
