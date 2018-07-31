package cmd

import (
	"io/ioutil"

	"github.com/Tahler/isotope/automation/convert/pkg/graph"
	"github.com/Tahler/isotope/automation/convert/pkg/graphviz"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
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

		var serviceGraph graph.ServiceGraph
		err = yaml.Unmarshal(yamlContents, &serviceGraph)
		exitIfError(err)

		dotLang, err := graphviz.ServiceGraphToDotLanguage(serviceGraph)
		exitIfError(err)

		outFileName := args[1]
		err = ioutil.WriteFile(outFileName, []byte(dotLang), 0644)
		exitIfError(err)
	},
}

func init() {
	rootCmd.AddCommand(graphvizCmd)
}
