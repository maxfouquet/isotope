package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tahler/isotope/pkg/graph"
	manif "github.com/Tahler/isotope/pkg/kubernetes"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
)

func runTopologies(topologyPaths []string) error {
	for _, path := range topologyPaths {
		serviceGraph, client, err := genManifests(path)
		if err != nil {
			return err
		}
		err = testServiceGraph(serviceGraph, client)
		if err != nil {
			return err
		}
	}
}

func testServiceGraph(serviceGraphManifest, clientManifest []byte) (err error) {

	err := createServiceGraph(serviceGraphManifest)
	if err != nil {
		return
	}

	err := createFromManifest(clientManifest)
	if err != nil {
		return
	}
	return
}

func createServiceGraph(manifest []byte) (err error) {
	err = createNamespace(consts.ServiceGraphNamespace)
	if err != nil {
		return
	}

	err = createFromManifest(serviceGraphManifest)
	if err != nil {
		return
	}

}

func createFromManifest(manifest []byte) (err error) {
	return
}

func genManifests(topologyPath string) (serviceGraphManifest, clientManifest []byte, err error) {
	glog.Infof("generating yaml for %s", topologyPath)

	yamlContents, err := ioutil.ReadFile(topologyPath)
	if err != nil {
		return
	}

	var serviceGraph graph.ServiceGraph
	err = yaml.Unmarshal(yamlContents, &serviceGraph)
	if err != nil {
		return
	}

	serviceGraphManifest, err = manif.ServiceGraphToKubernetesManifests(serviceGraph)
	if err != nil {
		return
	}

	clientManifest, err = manif.ServiceGraphToFortioClientManifest(serviceGraph)
	if err != nil {
		return
	}
	return
}

func main() {

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		namespace := "default"
		pod := "example-xxxxx"
		_, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}

		time.Sleep(10 * time.Second)
	}
}
