package srv

import (
	"fmt"
	"net/http"
	"os"
	"strings"
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

	respond := func(status int, paths []string, isLocalErr bool) {
		stampHeader(h.Service.Name, writer.Header(), paths, isLocalErr)

		stopTime := time.Now()
		duration := stopTime.Sub(startTime)
		// TODO: Record size of response payload.
		h.Metrics.RecordResponseSent(duration, 0, status)

		writer.WriteHeader(status)
		err := request.Write(writer)
		if err != nil {
			log.Errf("%s", err)
		}
	}

	if err := h.errorChance(); err != nil {
		respond(http.StatusInternalServerError, nil, true)
		return
	}

	allPaths := make([]string, 0, len(h.Service.Script))
	for _, step := range h.Service.Script {
		forwardableHeader := extractForwardableHeader(request.Header)
		paths, err := execute(step, forwardableHeader, h.ServiceTypes, h.Metrics)
		allPaths = append(allPaths, paths...)
		if err != nil {
			log.Errf("%s", err)
			respond(http.StatusInternalServerError, allPaths, false)
			return
		}
	}

	respond(http.StatusOK, allPaths, false)
}

func stampHeader(
	serviceName string, header http.Header, paths []string, isLocalErr bool) {
	stamp := fmt.Sprintf("%s (%s)", serviceName, hostname)
	if isLocalErr {
		stamp += " (ERROR)"
	}

	var stampedPaths []string
	if len(paths) == 0 {
		stampedPaths = []string{stamp}
	} else {
		stampedPaths = stampPaths(paths, stamp)
	}
	log.Debugf("stamped headers:\n%s", strings.Join(stampedPaths, "\n"))

	header[pathTracesHeaderKey] = stampedPaths
}

func stampPaths(paths []string, stamp string) []string {
	stampedPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		stampedPath := fmt.Sprintf("%s %s", stamp, path)
		stampedPaths = append(stampedPaths, stampedPath)
	}
	return stampedPaths
}

// errorChance randomly returns an error h.Service.ErrorRate percent of the
// time.
func (h Handler) errorChance() (err error) {
	// TODO: Restore once Fortio can ignore errors.
	return nil
	// random := rand.Float64()
	// if random < float64(h.Service.ErrorRate) {
	// 	err = fmt.Errorf("server randomly failed with a chance of %v", h.Service.ErrorRate)
	// }
	// return
}
