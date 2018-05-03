package cmd

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/Tahler/service-grapher/pkg/graphviz"
)

// graphvizCmd represents the graphviz command
var graphvizCmd = &cobra.Command{
	Use:   "graphviz [YAML file] [output file]",
	Short: "Convert a .yaml file to a Graphviz DOT language file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		inFileName := args[0]
		yamlContents, err := ioutil.ReadFile(inFileName)
		exitIfError(err)

		dotLang, err := graphviz.FromYAML(yamlContents)
		exitIfError(err)

		outFileName := args[1]
		err = ioutil.WriteFile(outFileName, []byte(dotLang), 0644)
		exitIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(graphvizCmd)
}
