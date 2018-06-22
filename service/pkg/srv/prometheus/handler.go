package prometheus

import (
	"net/http"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requests = prom.NewCounter(
		prom.CounterOpts{
			Name: "service_requests_total",
			Help: "Number of requests to this service.",
		})
)

// Handler returns an http.Handler which should be attached to a "/metrics"
// endpoint for Prometheus to ingest.
func Handler() http.Handler {
	prom.MustRegister(requests)
	return promhttp.Handler()
}

// RecordRequest increments the number of requests from serviceName.
func RecordRequest() {
	requests.Inc()
}
