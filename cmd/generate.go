package cmd

import (
	"github.com/spf13/cobra"
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
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Persistent flags available to all subcommands
	generateCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input CUE specification file (required)")
	generateCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "Output directory (required)")

	generateCmd.MarkPersistentFlagRequired("input")
	generateCmd.MarkPersistentFlagRequired("output")
}
