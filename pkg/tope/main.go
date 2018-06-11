package notmain

// import (
// 	"fmt"
// 	"os"

// 	appsv1 "k8s.io/api/apps/v1"
// 	apiv1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/kubernetes"
// 	typedappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
// 	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
// 	"k8s.io/client-go/tools/clientcmd"
// 	helm "k8s.io/helm/pkg/kube"
// )

// // TODO: move this file into cmd/
// func main() {
// 	// Create GKE Cluster : cloud.google.com/go/container/apiv1 : CreateCluster
// 	clientConfig, err := getKubernetesClientConfig()
// 	exitIfError(err)
// 	kubeClient, err := getDeploymentsClient(clientConfig)
// 	exitIfError(err)
// 	promResult, err := deployPrometheus(kubeClient)
// 	fmt.Printf("%v\n", promResult)

// 	// Test with no Istio

// 	// For each test in some test config (YAML)
// 	//   Deploy Istio from YAML : (first try no istio and full istio, then create a dependency tree)
// 	//   Generate Service Graph from YAML
// 	//   Deploy Service Graph
// 	//   Deploy Fortio client job
// 	//   Print data from fortio client logs
// 	//   Print data from prometheus
// }

// func testNoIstio() error {
// 	// Deploy service graph
// 	// Deploy fortio client
// }

// func exitIfError(err error) {
// 	if err != nil {
// 		fmt.Println(err)
// 		os.Exit(1)
// 	}
// }

// func getKubernetesClientConfig() (config clientcmd.ClientConfig, err error) {
// 	config, err = clientcmd.BuildConfigFromFlags("", "")
// 	if err != nil {
// 		return
// 	}
// 	return
// }

// func getDeploymentsClient(clientConfig clientcmd.ClientConfig) (client typedappsv1.DeploymentInterface, err error) {
// 	clientSet, err := kubernetes.NewForConfig(clientConfig)
// 	if err != nil {
// 		return
// 	}
// 	client = clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)
// 	return
// }

// func deployPrometheus(client typedappsv1.DeploymentInterface) (result appsv1.Deployment, err error) {
// 	deployment, err := getPrometheusDeployment()
// 	if err != nil {
// 		return
// 	}
// 	result, err = client.Create(deployment)
// 	if err != nil {
// 		return
// 	}
// 	return
// }

// func getPrometheusDeployment() (deployment appsv1.Deployment) {
// 	deployment = appsv1.Deployment{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "prometheus",
// 		},
// 		Spec: appsv1.DeploymentSpec{
// 			Selector: &metav1.LabelSelector{
// 				MatchLabels: map[string]string{"app": "prometheus"},
// 			},
// 		},
// 	}
// 	return
// }

// func getHelmClient(config clientcmd.ClientConfig) (client helm.Client, err error) {
// 	client, err = helm.New(config)
// 	if err != nil {
// 		return
// 	}
// 	client.
// 	return
// }
