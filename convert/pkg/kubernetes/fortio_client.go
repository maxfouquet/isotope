// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"errors"
	"fmt"

	"github.com/Tahler/isotope/convert/pkg/consts"
	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/graph/svc"
	"github.com/ghodss/yaml"
	batchv1 "k8s.io/api/batch/v1"
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
	job := entrypointToFortioClientJob(
		entrypoint, nodeSelector, clientImage, clientArgs)
	manifest, err = yaml.Marshal(job)
	if err != nil {
		return
	}
	return
}

var fortioClientLabels = map[string]string{"app": "client"}

func entrypointToFortioClientJob(
	entrypoint svc.Service,
	nodeSelector map[string]string,
	clientImage string,
	clientArgs []string) (job batchv1.Job) {

	url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%v",
		entrypoint.Name, ServiceGraphNamespace, consts.ServicePort)

	args := make([]string, 0, len(clientArgs)+1)
	args = append(args, clientArgs...)
	args = append(args, url)

	job.APIVersion = "batch/v1"
	job.Kind = "Job"
	job.ObjectMeta.Name = "client"
	timestamp(&job.ObjectMeta)
	job.Spec.Template = apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: combineLabels(serviceGraphAppLabels, fortioClientLabels),
		},
		Spec: apiv1.PodSpec{
			NodeSelector: nodeSelector,
			Containers: []apiv1.Container{
				{
					Name:  "fortio-client",
					Image: clientImage,
					Args:  args,
					Ports: []apiv1.ContainerPort{
						{
							ContainerPort: consts.ServicePort,
						},
					},
				},
			},
			RestartPolicy: apiv1.RestartPolicyNever,
		},
	}
	timestamp(&job.Spec.Template.ObjectMeta)
	return
}
