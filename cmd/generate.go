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
	Short: "Generate Prometheus metrics code from YAML specification",
	Long: `Generate code for Prometheus metrics based on a YAML specification file.
Supports multiple target languages: Go, .NET, and Node.js.

Use subcommands to specify the target language:
  promener generate go -i metrics.yaml -o ./out
  promener generate dotnet -i metrics.yaml -o ./out
  promener generate nodejs -i metrics.yaml -o ./out`,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Persistent flags available to all subcommands
	generateCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input YAML specification file (required)")
	generateCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", "Output directory (required)")

	generateCmd.MarkPersistentFlagRequired("input")
	generateCmd.MarkPersistentFlagRequired("output")
}
