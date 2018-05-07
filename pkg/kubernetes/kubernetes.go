// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tahler/service-grapher/pkg/graph"
	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceGraphToKubernetesManifests converts a ServiceGraph to Kubernetes
// manifests.
func ServiceGraphToKubernetesManifests(
	serviceGraph graph.ServiceGraph) (yamlDoc []byte, err error) {
	// ConfigMap manifest + 2 manifests per service (i.e. Deployment, Service).
	numManifests := len(serviceGraph.Services)*2 + 1
	manifests := make([]string, numManifests)

	configMap, err := makeConfigMap(serviceGraph)
	if err != nil {
		return
	}
	configMapYAML, err := yaml.Marshal(configMap)
	if err != nil {
		return
	}
	manifests = append(manifests, string(configMapYAML))

	// for _, service := range serviceGraph.Services {
	// }

	yamlDocString := strings.Join(manifests, "---\n")
	fmt.Printf("%+v\n", yamlDocString)
	return []byte(yamlDocString), nil
}

func makeConfigMap(
	graph graph.ServiceGraph) (configMap apiv1.ConfigMap, err error) {
	configMap.ObjectMeta.Name = "scripts"
	configMap.ObjectMeta.CreationTimestamp = metav1.Time{Time: time.Now()}

	data := make(map[string]string)
	for _, service := range graph.Services {
		yml, err := yaml.Marshal(service)
		if err != nil {
			return configMap, err
		}
		data[service.Name] = string(yml)
	}
	configMap.Data = data
	return
}

func makeService(service graph.Service) (k8sService apiv1.Service, err error) {
	k8sService.ObjectMeta.Name = service.Name
	return
}

const containerName = "performance-test"
const containerImage = "istio.gcr.io/performance-test"

func makeDeployment(
	service graph.Service) (k8sDeployment appsv1.Deployment, err error) {
	k8sDeployment.ObjectMeta.Name = service.Name
	k8sDeployment.Spec = appsv1.DeploymentSpec{
		Template: apiv1.PodTemplateSpec{
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  containerName,
						Image: containerImage,
					},
				},
			},
		},
	}
	return
}
