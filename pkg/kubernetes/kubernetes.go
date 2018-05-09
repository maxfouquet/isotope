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
	manifests := make([]string, 0, numManifests)

	appendManifest := func(manifest interface{}) error {
		yamlDoc, err := yaml.Marshal(manifest)
		if err != nil {
			return err
		}
		manifests = append(manifests, string(yamlDoc))
		return nil
	}

	configMap, err := makeConfigMap(serviceGraph)
	if err != nil {
		return
	}
	appendManifest(configMap)

	for _, service := range serviceGraph.Services {
		k8sDeployment, err := makeDeployment(service)
		if err != nil {
			return nil, err
		}
		appendManifest(k8sDeployment)

		k8sService, err := makeService(service)
		if err != nil {
			return nil, err
		}
		appendManifest(k8sService)
	}

	yamlDocString := strings.Join(manifests, "---\n")
	fmt.Printf("%+v\n", yamlDocString)
	return []byte(yamlDocString), nil
}

func makeConfigMap(
	graph graph.ServiceGraph) (configMap apiv1.ConfigMap, err error) {
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.ObjectMeta.Name = "service-configs"
	timestamp(&configMap.ObjectMeta)

	data := make(map[string]string)
	for _, service := range graph.Services {
		marshallable, err := serviceToMarshallable(service)
		if err != nil {
			return configMap, err
		}
		serviceYAML, err := yaml.Marshal(marshallable)
		if err != nil {
			return configMap, err
		}
		data[service.Name] = string(serviceYAML)
	}
	configMap.Data = data
	return
}

func makeService(service graph.Service) (k8sService apiv1.Service, err error) {
	k8sService.APIVersion = "v1"
	k8sService.Kind = "Service"
	k8sService.ObjectMeta.Name = service.Name
	timestamp(&k8sService.ObjectMeta)
	k8sService.Spec.Ports = []apiv1.ServicePort{{Port: 8080}}
	return
}

const containerName = "perf-test-service"
const containerImage = "tahler/perf-test-service"

func makeDeployment(
	service graph.Service) (k8sDeployment appsv1.Deployment, err error) {
	k8sDeployment.APIVersion = "apps/v1"
	k8sDeployment.Kind = "Deployment"
	k8sDeployment.ObjectMeta.Name = service.Name
	timestamp(&k8sDeployment.ObjectMeta)
	k8sDeployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": service.Name,
			},
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": service.Name,
				},
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  containerName,
						Image: containerImage,
						Env: []apiv1.EnvVar{
							{
								Name: "SERVICE_YAML",
								ValueFrom: &apiv1.EnvVarSource{
									ConfigMapKeyRef: &apiv1.ConfigMapKeySelector{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: "service-configs",
										},
										Key: service.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return
}

func timestamp(objectMeta *metav1.ObjectMeta) {
	objectMeta.CreationTimestamp = metav1.Time{Time: time.Now()}
}
