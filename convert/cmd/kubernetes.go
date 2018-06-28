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

// kubernetesCmd represents the kubernetes command
var kubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Convert service graph YAML to manifests for performance testing",
	Args:  cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		inPath := args[0]
		serviceGraphOutPath := args[1]
		clientOutPath := args[2]

		defaultNodeSelectorStr := args[3]
		defaultNodeSelector, err := extractNodeSelector(
			defaultNodeSelectorStr)
		exitIfError(err)

		clientNodeSelectorStr := args[4]
		clientNodeSelector, err := extractNodeSelector(clientNodeSelectorStr)
		exitIfError(err)

		yamlContents, err := ioutil.ReadFile(inPath)
		exitIfError(err)

		var serviceGraph graph.ServiceGraph
		exitIfError(yaml.Unmarshal(yamlContents, &serviceGraph))

		serviceImage, err := cmd.PersistentFlags().GetString("service-image")
		exitIfError(err)

		serviceGraphManifest, err := kubernetes.ServiceGraphToKubernetesManifests(
			serviceGraph, defaultNodeSelector, serviceImage)
		exitIfError(err)

		clientImage, err := cmd.PersistentFlags().GetString("client-image")
		exitIfError(err)

		clientArgs, err := cmd.PersistentFlags().GetStringSlice("client-args")
		exitIfError(err)

		clientManifest, err := kubernetes.ServiceGraphToFortioClientManifest(
			serviceGraph, clientNodeSelector, clientImage, clientArgs)
		exitIfError(err)

		exitIfError(writeManifest(serviceGraphOutPath, serviceGraphManifest))

		exitIfError(writeManifest(clientOutPath, clientManifest))
	},
}

func init() {
	rootCmd.AddCommand(kubernetesCmd)
	kubernetesCmd.PersistentFlags().String(
		"service-image", "", "the image to deploy for all services in the graph")
	kubernetesCmd.PersistentFlags().String(
		"client-image", "", "the image to use for the load testing client job")
	kubernetesCmd.PersistentFlags().StringSlice(
		"client-args",
		[]string{},
		"the args to send to the load testing client, separated by comma")
}

func writeManifest(path string, manifest []byte) error {
	return ioutil.WriteFile(path, manifest, 0644)
}

func splitByEquals(s string) (k string, v string, err error) {
	parts := strings.Split(s, "=")
	if len(parts) != 2 {
		err = fmt.Errorf("%s is not a valid node selector", s)
		return
	}
	k = parts[0]
	v = parts[1]
	return
}

func extractNodeSelector(s string) (map[string]string, error) {
	nodeSelector := make(map[string]string, 1)
	k, v, err := splitByEquals(s)
	if err != nil {
		return nodeSelector, err
	}
	nodeSelector[k] = v
	return nodeSelector, nil
}
