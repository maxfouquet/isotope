package srv

import (
	"net/http"
	"os"
	"time"

	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/convert/pkg/graph/svctype"
	"github.com/Tahler/isotope/service/pkg/srv/prometheus"
	"istio.io/fortio/log"
)

// pathTracesHeaderKey is the HTTP header key for path tracing. It must be in
// Train-Case.
const pathTracesHeaderKey = "Path-Traces"

var hostname = os.Getenv("HOSTNAME")

// Handler handles the default endpoint by emulating its Service.
type Handler struct {
	Service      svc.Service
	ServiceTypes map[string]svctype.ServiceType
	Metrics      prometheus.Metrics
}

func (h Handler) ServeHTTP(
	writer http.ResponseWriter, request *http.Request) {
	startTime := time.Now()

	h.Metrics.RecordRequestReceived()

	respond := func(status int) {
		writer.WriteHeader(status)
		err := request.Write(writer)
		if err != nil {
			log.Errf("%s", err)
		}

		stopTime := time.Now()
		duration := stopTime.Sub(startTime)
		// TODO: Record size of response payload.
		h.Metrics.RecordResponseSent(duration, 0, status)
	}

	for _, step := range h.Service.Script {
		forwardableHeader := extractForwardableHeader(request.Header)
		err := execute(step, forwardableHeader, h.ServiceTypes, h.Metrics)
		if err != nil {
			log.Errf("%s", err)
			respond(http.StatusInternalServerError)
			return
		}
	}

	respond(http.StatusOK)
}
