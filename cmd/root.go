package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "service-grapher",
	Short: "Converts a service graph expressed in YAML to various formats",
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once to
// the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	exitIfError(err)
}

func exitIfError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
