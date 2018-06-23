package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/kubernetes"
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
		// Split by '=' (i.e. cloud.google.com/gke-nodepool=client-pool)
		clientNodeSelectorStr := args[4]
		clientNodeSelector, err := extractClientNodeSelector(clientNodeSelectorStr)
		exitIfError(err)

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
			serviceGraph, clientNodeSelector)
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

func extractClientNodeSelector(s string) (map[string]string, error) {
	nodeSelector := make(map[string]string, 1)
	parts := strings.Split(s, "=")
	if len(parts) != 2 {
		return nodeSelector, fmt.Errorf("%s is not a valid node selector", s)
	}
	nodeSelector[parts[0]] = parts[1]
	return nodeSelector, nil
}
