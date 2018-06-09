package cmd

import (
	"io/ioutil"

	"github.com/Tahler/service-grapher/pkg/graph"
	"github.com/Tahler/service-grapher/pkg/kubernetes"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// performanceKubernetesCmd represents the performanceKubernetes command
var performanceKubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Convert service graph YAML to manifests for performance testing",
	Run: func(cmd *cobra.Command, args []string) {
		inFileName := args[0]
		yamlContents, err := ioutil.ReadFile(inFileName)
		exitIfError(err)

		var serviceGraph graph.ServiceGraph
		exitIfError(yaml.Unmarshal(yamlContents, &serviceGraph))

		serviceGraphManifest, err := kubernetes.ServiceGraphToKubernetesManifests(
			serviceGraph)
		exitIfError(err)

		clientManifest, err = kubernetes.ServiceGraphToFortioClientManifest(
			serviceGraph)
		exitIfError(err)

		exitIfError(writeManifest("service-graph.yaml", serviceGraphManifest))

		exitIfError(writeManifest("client.yaml", clientManifest))

		exitIfError(writeManifest("prometheus.yaml", "here goes prometheus yaml"))
	},
}

func writeManifest(fileName string, manifest string) error {
	return ioutil.WriteFile(fileName, []byte(manifest), 0644)
}

func init() {
	performanceCmd.AddCommand(performanceKubernetesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// performanceKubernetesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// performanceKubernetesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
