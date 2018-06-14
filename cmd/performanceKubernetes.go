package cmd

import (
	"github.com/spf13/cobra"
)

// performanceKubernetesCmd represents the performanceKubernetes command
var performanceKubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "Convert service graph YAML to manifests for performance testing",
	Run: func(cmd *cobra.Command, args []string) {
		err := runner.run_topologies(args)
		exitIfError(err)
	},
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
