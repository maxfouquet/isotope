// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	helm "k8s.io/helm/pkg/kube"
)

// topeCmd represents the tope command
var topeCmd = &cobra.Command{
	Use:   "tope",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tope called")
		main2()
	},
}

func init() {
	rootCmd.AddCommand(topeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// topeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// topeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// TODO: move this file into cmd/
func main2() {
	// Create GKE Cluster : cloud.google.com/go/container/apiv1 : CreateCluster
	clientConfig, err := getKubernetesClientConfig()
	exitIfError(err)
	// kubeClient, err := getDeploymentsClient(clientConfig)
	// exitIfError(err)
	// promResult, err := deployPrometheus(kubeClient)
	// fmt.Printf("%v\n", promResult)

	helmClient, err := getHelmClient(clientConfig)
	exitIfError(err)
	err = helmClient.Create("default", "prometheus", 30*time.Second, true)
	exitIfError(err)

	// Test with no Istio

	// For each test in some test config (YAML)
	//   Deploy Istio from YAML : (first try no istio and full istio, then create a dependency tree)
	//   Generate Service Graph from YAML
	//   Deploy Service Graph
	//   Deploy Fortio client job
	//   Print data from fortio client logs
	//   Print data from prometheus
}

func getKubernetesClientConfig() (config clientcmd.ClientConfig, err error) {
	config, err = clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		return
	}
	return
}

func getHelmClient(config clientcmd.ClientConfig) (client helm.Client, err error) {
	client, err = helm.New(config)
	if err != nil {
		return
	}
	return
}
