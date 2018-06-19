package kubernetes

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	labelsName = "labels"

	istioMetricsYAMLTemplate = `apiVersion: "config.istio.io/v1alpha2"
kind: metric
metadata:
  name: request_count
  namespace: istio-system
spec:
  value: "1"
  dimensions:
    {{- range $key, $value := . }}
    {{ $key }}: {{ $value }}
    {{- end }}
  monitored_resource_type: '"UNSPECIFIED"'
---
apiVersion: "config.istio.io/v1alpha2"
kind: prometheus
metadata:
  name: request_count_handler
  namespace: istio-system
spec:
  metrics:
  - name: service_request_count
    instance_name: request_count.metric.istio-system
    kind: COUNTER
    label_names:
    {{- range $key, $value := . }}
    - {{ $key }}
    {{- end }}
`
)

// LabelsToIstioManifests returns the necessary manifests for self-labeling
// Prometheus metrics for Istio.
func LabelsToIstioManifests(labels map[string]string) (
	istioMetricsManifests []byte, err error) {
	istioMetricsManifests, err = makeIstioMetricsManifests(labels)
	if err != nil {
		return
	}
	return
}

func makeIstioMetricsManifests(
	labels map[string]string) (manifests []byte, err error) {
	tmpl, err := template.New("istio-metrics").Parse(istioMetricsYAMLTemplate)
	if err != nil {
		return
	}
	// TODO: Use yaml.Marshal(labels) and yaml.Marshal(keys(labels))
	var b bytes.Buffer
	if err = tmpl.Execute(&b, labels); err != nil {
		return
	}
	manifests = b.Bytes()
	return
}

// LabelsFor returns the static labels for the topology at topologyPath.
func LabelsFor(topologyPath string) (labels map[string]string, err error) {
	topologyName := getFileNameNoExt(topologyPath)
	topologyHash, err := getHash(topologyPath)
	if err != nil {
		return
	}
	labels = map[string]string{
		"topology_name": topologyName,
		"topology_hash": topologyHash,
	}
	return
}

func getFileNameNoExt(path string) string {
	basename := filepath.Base(path)
	extension := filepath.Ext(path)
	return strings.TrimSuffix(basename, extension)
}

func getHash(path string) (hash string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	h := md5.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return
	}

	hash = fmt.Sprintf("%x", h.Sum(nil))
	return
}
