package prometheus

import (
	"net/http"
	"strconv"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	sizeBuckets = []float64{
		// 1, 10, 100, 1,000, ..., 1,000,000,000
		1e+00, 1e+01, 1e+02, 1e+03, 1e+04, 1e+05, 1e+06, 1e+07, 1e+08, 1e+09}
)

// Metrics encapsulates all the relevant metrics to capture.
type Metrics struct {
	serviceIncomingRequestsTotal  prom.Counter
	serviceOutgoingRequestsTotal  prom.CounterVec
	serviceOutgoingRequestSize    prom.HistogramVec
	serviceRequestDurationSeconds prom.HistogramVec
	serviceResponseSize           prom.HistogramVec
}

// NewMetrics returns an instance of metrics with the configured Prometheus
// metric types.
func NewMetrics() Metrics {
	return Metrics{
		serviceIncomingRequestsTotal: prom.NewCounter(
			prom.CounterOpts{
				Name: "service_incoming_requests_total",
				Help: "Number of requests sent to this service.",
			}),

		serviceOutgoingRequestsTotal: *prom.NewCounterVec(
			prom.CounterOpts{
				Name: "service_outgoing_requests_total",
				Help: "Number of requests sent from this service.",
			}, []string{"destination_service"}),

		serviceOutgoingRequestSize: *prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "service_outgoing_request_size",
				Help:    "Size in bytes of requests sent from this service.",
				Buckets: sizeBuckets,
			}, []string{"destination_service"}),

		serviceRequestDurationSeconds: *prom.NewHistogramVec(
			prom.HistogramOpts{
				Name: "service_request_duration_seconds",
				Help: "Duration in seconds it took to serve requests to this service.",
			}, []string{"code"}),

		serviceResponseSize: *prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "service_response_size",
				Help:    "Size in bytes of responses sent from this service.",
				Buckets: sizeBuckets,
			}, []string{"code"}),
	}
}

// Handler returns an http.Handler which should be attached to a "/metrics"
// endpoint for Prometheus to ingest.
func (m Metrics) Handler() (handler http.Handler, err error) {
	registry := prom.NewRegistry()
	err = registry.Register(m.serviceIncomingRequestsTotal)
	if err != nil {
		return
	}
	err = registry.Register(m.serviceOutgoingRequestsTotal)
	if err != nil {
		return
	}
	err = registry.Register(m.serviceOutgoingRequestSize)
	if err != nil {
		return
	}
	err = registry.Register(m.serviceRequestDurationSeconds)
	if err != nil {
		return
	}
	err = registry.Register(m.serviceResponseSize)
	if err != nil {
		return
	}
	handler = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return
}

// RecordRequestReceived increments the Prometheus counter for incoming
// requests.
func (m Metrics) RecordRequestReceived() {
	m.serviceIncomingRequestsTotal.Inc()
}

// RecordRequestSent increments the Prometheus counter for outgoing requests
// and records an outgoing request size.
func (m Metrics) RecordRequestSent(destinationService string, size uint64) {
	m.serviceOutgoingRequestsTotal.WithLabelValues(destinationService).Inc()
	m.serviceOutgoingRequestSize.WithLabelValues(destinationService).Observe(
		float64(size))
}

// RecordResponseSent observes the time-to-response duration and size for the
// HTTP status code.
func (m Metrics) RecordResponseSent(
	duration time.Duration, size uint64, code int) {
	strCode := strconv.Itoa(code)
	m.serviceRequestDurationSeconds.WithLabelValues(strCode).Observe(
		duration.Seconds())
	m.serviceResponseSize.WithLabelValues(strCode).Observe(float64(size))
}
