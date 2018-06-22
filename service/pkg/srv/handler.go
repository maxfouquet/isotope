package srv

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/Tahler/isotope/service/pkg/srv/prometheus"
	"istio.io/fortio/log"
)

// pathTracesHeaderKey is the HTTP header key for path tracing. It must be in
// Train-Case.
const pathTracesHeaderKey = "Path-Traces"

var hostname = os.Getenv("HOSTNAME")

// Handler handles the default endpoint by emulating its Service.
type Handler struct {
	svc.Service
}

func (h Handler) ServeHTTP(
	writer http.ResponseWriter, request *http.Request) {

	prometheus.RecordRequest()

	respond := func(status int, paths []string, isLocalErr bool) {
		stampHeader(writer.Header(), paths, isLocalErr)

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

	allPaths := make([]string, 0, len(h.Script))
	for _, step := range h.Script {
		forwardableHeader := extractForwardableHeader(request.Header)
		paths, err := execute(step, forwardableHeader)
		allPaths = append(allPaths, paths...)
		if err != nil {
			log.Errf("%s", err)
			respond(http.StatusInternalServerError, allPaths, false)
			return
		}
	}

	respond(http.StatusOK, allPaths, false)
}

func stampHeader(header http.Header, paths []string, isLocalErr bool) {
	stamp := fmt.Sprintf("%s (%s)", Service.Name, hostname)
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

// errorChance randomly returns an error h.ErrorRate percent of the time.
func (h Handler) errorChance() (err error) {
	// TODO: Restore once Fortio can ignore errors.
	return nil
	// random := rand.Float64()
	// if random < float64(h.ErrorRate) {
	// 	err = fmt.Errorf("server randomly failed with a chance of %v", h.ErrorRate)
	// }
	// return
}
