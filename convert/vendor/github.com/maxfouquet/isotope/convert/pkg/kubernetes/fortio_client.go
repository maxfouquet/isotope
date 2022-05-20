// Package kubernetes converts service graphs into Kubernetes manifests.
package kubernetes

import (
	"github.com/maxfouquet/isotope/convert/pkg/consts"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var fortioClientLabels = map[string]string{"app": "client"}

func makeFortioDeployment(
	nodeSelector map[string]string,
	clientImage string) (deployment appsv1.Deployment) {
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
							},
							{
								ContainerPort: consts.FortioMetricsPort,
							},
						},
					},
				},
			},
		},
	}
	timestamp(&deployment.Spec.Template.ObjectMeta)
	return
}

func makeFortioService() (service apiv1.Service) {
	service.APIVersion = "v1"
	service.Kind = "Service"
	service.ObjectMeta.Name = "client"
	service.ObjectMeta.Labels = fortioClientLabels
	timestamp(&service.ObjectMeta)
	service.Spec.Ports = []apiv1.ServicePort{{Port: consts.ServicePort}}
	service.Spec.Selector = fortioClientLabels
	return
}
