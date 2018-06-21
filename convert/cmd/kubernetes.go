package cmd

import (
	"io/ioutil"

	"github.com/Tahler/isotope/convert/pkg/graph"
	"github.com/Tahler/isotope/convert/pkg/kubernetes"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// kubernetesCmd represents the kubernetes command
var kubernetesCmd = &cobra.Command{
	Use:   "kubernetes [YAML file] [output file]",
	Short: "Convert a service graph YAML file to Kubernetes manifest YAML",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		inFileName := args[0]
		yamlContents, err := ioutil.ReadFile(inFileName)
		exitIfError(err)

		var serviceGraph graph.ServiceGraph
		err = yaml.Unmarshal(yamlContents, &serviceGraph)
		exitIfError(err)

		labels, err := kubernetes.LabelsFor(inFileName)
		exitIfError(err)

		manifest, err := kubernetes.ServiceGraphToKubernetesManifests(
			serviceGraph, labels)
		exitIfError(err)

		outFileName := args[1]
		err = ioutil.WriteFile(outFileName, []byte(manifest), 0644)
		exitIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(kubernetesCmd)
}
