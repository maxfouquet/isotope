package cmd

import (
	"io/ioutil"

	"github.com/Tahler/service-grapher/pkg/graph"
	"github.com/Tahler/service-grapher/pkg/kubernetes"
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

		manifest, err := kubernetes.ServiceGraphToKubernetesManifests(
			serviceGraph)
		exitIfError(err)

		outFileName := args[1]
		err = ioutil.WriteFile(outFileName, []byte(manifest), 0644)
		exitIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(kubernetesCmd)
}
