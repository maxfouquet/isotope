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

	prometheusValuesYAMLTemplate = `serviceMonitors:
- name: service-graph-monitor
  selector:
    matchLabels:
      app: service-graph
  namespaceSelector:
    matchNames:
    - service-graph
  endpoints:
  - targetPort: 8080
    metricRelabelings:
    {{- range $key, $value := . }}
    - targetLabel: {{ $key }}
      replacement: {{ $value }}
    {{- end }}
- name: istio-mixer-monitor
  selector:
    matchLabels:
      istio: mixer
  namespaceSelector:
    matchNames:
    - istio-system
  endpoints:
  - targetPort: 42422
    metricRelabelings:
    {{- range $key, $value := . }}
    - targetLabel: {{ $key }}
      replacement: {{ $value }}
    {{- end }}
storageSpec:
  volumeClaimTemplate:
    spec:
      # It's necessary to specify "" as the storageClassName
      # so that the default storage class won't be used, see
      # https://kubernetes.io/docs/concepts/storage/persistent-volumes/#class-1
      storageClassName: ""
      volumeName: prometheus-persistent-volume
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 10G
`
)

// LabelsToPrometheusValuesYAML returns the values for coreos/prometheus for
// self-labeling metrics.
func LabelsToPrometheusValuesYAML(labels map[string]string) (
	prometheusValuesYAML []byte, err error) {
	tmpl, err := template.New("prom-values").Parse(prometheusValuesYAMLTemplate)
	if err != nil {
		return
	}
	// TODO: Use yaml.Marshal(labels) and yaml.Marshal(keys(labels))
	var b bytes.Buffer
	if err = tmpl.Execute(&b, labels); err != nil {
		return
	}
	prometheusValuesYAML = b.Bytes()
	return
}

// LabelsFor returns the labels for the topology at topologyPath.
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
