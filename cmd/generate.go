package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	inputFile string
	outputDir string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Prometheus metrics code from CUE specification",
	Long: `Generate code for Prometheus metrics based on a CUE specification file.
Supports multiple target languages: Go, .NET, and Node.js.

Use subcommands to specify the target language:
  promener generate go -i metrics.cue -o ./out
  promener generate dotnet -i metrics.cue -o ./out
  promener generate nodejs -i metrics.cue -o ./out`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Only validate if we're running a subcommand
		if cmd.HasSubCommands() {
			return nil
		}

		if viper.GetString("input") == "" {
			return fmt.Errorf("input file is required (via --input flag or config file)")
		}
		if viper.GetString("output") == "" {
			return fmt.Errorf("output directory is required (via --output flag or config file)")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Persistent flags available to all subcommands
	generateCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input CUE specification file")
	generateCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "Output directory")

	viper.BindPFlag("input", generateCmd.PersistentFlags().Lookup("input"))
	viper.BindPFlag("output", generateCmd.PersistentFlags().Lookup("output"))
}
