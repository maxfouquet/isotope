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

	serviceIncomingRequestsTotal = prom.NewCounter(
		prom.CounterOpts{
			Name: "service_incoming_requests_total",
			Help: "Number of requests sent to this service.",
		})

	serviceOutgoingRequestsTotal = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "service_outgoing_requests_total",
			Help: "Number of requests sent from this service.",
		}, []string{"destination_service"})

	serviceOutgoingRequestSize = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "service_outgoing_request_size",
			Help:    "Size in bytes of requests sent from this service.",
			Buckets: sizeBuckets,
		}, []string{"destination_service"})

	serviceRequestDurationSeconds = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name: "service_request_duration_seconds",
			Help: "Duration in seconds it took to serve requests to this service.",
		}, []string{"code"})

	serviceResponseSize = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "service_response_size",
			Help:    "Size in bytes of responses sent from this service.",
			Buckets: sizeBuckets,
		}, []string{"code"})
)

// Handler returns an http.Handler which should be attached to a "/metrics"
// endpoint for Prometheus to ingest.
func Handler() http.Handler {
	prom.MustRegister(serviceIncomingRequestsTotal)

	prom.MustRegister(serviceOutgoingRequestsTotal)
	prom.MustRegister(serviceOutgoingRequestSize)

	prom.MustRegister(serviceRequestDurationSeconds)
	prom.MustRegister(serviceResponseSize)

	return promhttp.Handler()
}

// RecordRequestReceived increments the Prometheus counter for incoming
// requests.
func RecordRequestReceived() {
	serviceIncomingRequestsTotal.Inc()
}

// RecordRequestSent increments the Prometheus counter for outgoing requests
// and records an outgoing request size.
func RecordRequestSent(destinationService string, size uint64) {
	serviceOutgoingRequestsTotal.WithLabelValues(destinationService).Inc()
	serviceOutgoingRequestSize.WithLabelValues(destinationService).Observe(
		float64(size))
}

// RecordResponseSent observes the time-to-response duration and size for the
// HTTP status code.
func RecordResponseSent(duration time.Duration, size uint64, code int) {
	strCode := strconv.Itoa(code)
	serviceRequestDurationSeconds.WithLabelValues(strCode).Observe(
		duration.Seconds())
	serviceResponseSize.WithLabelValues(strCode).Observe(float64(size))
}
