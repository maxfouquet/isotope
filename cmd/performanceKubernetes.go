package cmd

import (
	"io/ioutil"

	"github.com/Tahler/isotope/pkg/graph"
	"github.com/Tahler/isotope/pkg/kubernetes"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// performanceKubernetesCmd represents the performanceKubernetes command
var performanceKubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Convert service graph YAML to manifests for performance testing",
	Run: func(cmd *cobra.Command, args []string) {
		inPath := args[0]
		serviceGraphOutPath := args[1]
		prometheusValuesPath := args[2]
		clientOutPath := args[3]

		yamlContents, err := ioutil.ReadFile(inPath)
		exitIfError(err)

		var serviceGraph graph.ServiceGraph
		exitIfError(yaml.Unmarshal(yamlContents, &serviceGraph))

		labels, err := kubernetes.LabelsFor(inPath)
		exitIfError(err)

		serviceGraphManifest, err := kubernetes.ServiceGraphToKubernetesManifests(
			serviceGraph, labels)
		exitIfError(err)

		promValuesYAML, err := kubernetes.LabelsToPrometheusValuesYAML(labels)

		clientManifest, err := kubernetes.ServiceGraphToFortioClientManifest(
			serviceGraph)
		exitIfError(err)

		exitIfError(writeManifest(serviceGraphOutPath, serviceGraphManifest))

		exitIfError(writeManifest(prometheusValuesPath, promValuesYAML))

		exitIfError(writeManifest(clientOutPath, clientManifest))
	},
}

func writeManifest(path string, manifest []byte) error {
	return ioutil.WriteFile(path, manifest, 0644)
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
