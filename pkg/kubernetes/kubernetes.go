// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"strings"
	"time"

	"github.com/Tahler/service-grapher/pkg/consts"
	"github.com/Tahler/service-grapher/pkg/graph"
	"github.com/Tahler/service-grapher/pkg/graph/svc"
	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var serviceGraphLabels = map[string]string{"app": "service-graph"}

// ServiceGraphToKubernetesManifests converts a ServiceGraph to Kubernetes
// manifests.
func ServiceGraphToKubernetesManifests(
	serviceGraph graph.ServiceGraph) (yamlDoc []byte, err error) {
	numServices := len(serviceGraph.Services)
	numManifests := numManifestsPerService*numServices + numConfigMaps
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
	return []byte(yamlDocString), nil
}

const (
	numConfigMaps          = 1
	numManifestsPerService = 2

	configMapName = "service-graph-config"
)

func makeConfigMap(
	graph graph.ServiceGraph) (configMap apiv1.ConfigMap, err error) {
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.ObjectMeta.Name = configMapName
	timestamp(&configMap.ObjectMeta)

	graphYAMLBytes, err := yaml.Marshal(graph)
	if err != nil {
		return
	}
	configMap.Data = map[string]string{"service-graph": string(graphYAMLBytes)}
	return
}

func makeService(service svc.Service) (k8sService apiv1.Service, err error) {
	k8sService.APIVersion = "v1"
	k8sService.Kind = "Service"
	k8sService.ObjectMeta.Name = service.Name
	timestamp(&k8sService.ObjectMeta)
	k8sService.Spec.Ports = []apiv1.ServicePort{{Port: consts.ServicePort}}
	k8sService.Spec.Selector = serviceGraphLabels
	return
}

func makeDeployment(
	service svc.Service) (k8sDeployment appsv1.Deployment, err error) {
	k8sDeployment.APIVersion = "apps/v1"
	k8sDeployment.Kind = "Deployment"
	k8sDeployment.ObjectMeta.Name = service.Name
	timestamp(&k8sDeployment.ObjectMeta)
	k8sDeployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: serviceGraphLabels,
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: serviceGraphLabels,
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  consts.ServiceContainerName,
						Image: consts.ServiceImageName,
						Env: []apiv1.EnvVar{
							{Name: consts.ServiceNameEnvKey, Value: service.Name},
						},
						VolumeMounts: []apiv1.VolumeMount{
							{
								Name:      "config-volume",
								MountPath: consts.ConfigPath,
							},
						},
					},
				},
				Volumes: []apiv1.Volume{
					{
						Name: "config-volume",
						VolumeSource: apiv1.VolumeSource{
							ConfigMap: &apiv1.ConfigMapVolumeSource{
								LocalObjectReference: apiv1.LocalObjectReference{
									Name: configMapName,
								},
								Items: []apiv1.KeyToPath{
									{
										Key:  consts.ServiceGraphConfigMapKey,
										Path: consts.ServiceGraphYAMLFileName,
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
