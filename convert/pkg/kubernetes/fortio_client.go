// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"errors"
	"fmt"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceGraphToFortioClientManifest extracts the entrypoint into the service
// graph and renders a Kubernetes Job manifest to run `Fortio load` on it.
func ServiceGraphToFortioClientManifest(
	serviceGraph graph.ServiceGraph,
	nodeSelector map[string]string,
	clientImage string,
	clientArgs []string) (manifest []byte, err error) {
	entrypoints := make([]svc.Service, 0, 1)
	for _, svc := range serviceGraph.Services {
		if svc.IsEntrypoint {
			entrypoints = append(entrypoints, svc)
		}
	}
	numEntrypoints := len(entrypoints)
	if numEntrypoints > 1 {
		err = fmt.Errorf(
			"cannot create client for service graph with multiple entrypoints: %v",
			entrypoints)
		return
	}
	if numEntrypoints < 1 {
		err = errors.New(
			"cannot create client for service graph with no entrypoints")
		return
	}
	entrypoint := entrypoints[0]
	deployment := entrypointToFortioClientDeployment(
		nodeSelector, clientImage, clientArgs)
	manifest, err = yaml.Marshal(deployment)
	if err != nil {
		return
	}
	return
}

var fortioClientLabels = map[string]string{"app": "client"}

func entrypointToFortioClientDeployment(
	nodeSelector map[string]string,
	clientImage string,
	clientArgs []string) (deployment appsv1.Deployment) {

	deployment.APIVersion = "apps/v1"
	deployment.Kind = "Deployment"
	deployment.ObjectMeta.Name = "client"
	deployment.ObjectMeta.Labels = fortioClientLabels
	timestamp(&deployment.ObjectMeta)
	deployment.Spec = appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: fortioClientLabels,
		},
		Template: apiv1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: fortioClientLabels,
			},
			Spec: apiv1.PodSpec{
				NodeSelector: nodeSelector,
				Containers: []apiv1.Container{
					{
						Name:  "fortio-client",
						Image: clientImage,
						Args:  []string{"server"},
						Ports: []apiv1.ContainerPort{
							{
								ContainerPort: consts.ServicePort,
								ContainerPort: 42422,
							},
						},
					},
				},
			},
		},
	}
	timestamp(&job.Spec.Template.ObjectMeta)
	return
}
