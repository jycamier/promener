package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jycamier/promener/internal/generator"
	"github.com/jycamier/promener/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	nodejsPackageName string
)

// nodejsCmd represents the nodejs command
var nodejsCmd = &cobra.Command{
	Use:   "nodejs",
	Short: "Generate Node.js code for Prometheus metrics",
	Long: `Generate Node.js/TypeScript code for Prometheus metrics from a CUE specification file.
Generates metrics.ts in the output directory.

Examples:
  promener generate nodejs -i metrics.cue -o ./out
  promener generate nodejs -i metrics.cue -o ./out -p myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get values from Viper
		inputFile := viper.GetString("input")
		outputDir := viper.GetString("output")
		packageName := viper.GetString("nodejs.package")

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Validate and extract the CUE specification
		v := validator.New()
		if rules := viper.GetStringSlice("rules"); len(rules) > 0 {
			v.SetRulesDirs(rules)
		}
		spec, result, err := v.ValidateAndExtract(inputFile)
		threshold := viper.GetString("severity_on_error")

		if err != nil || result.Failed(threshold) {
			if result != nil && result.HasErrors() {
				// Format validation errors
				formatter := validator.NewFormatter(validator.FormatText)
				output, _ := formatter.Format(result)
				fmt.Fprint(os.Stderr, output)
			}
			if result != nil && result.Failed(threshold) {
				return fmt.Errorf("failed to validate specification (threshold: %s)", threshold)
			}
			return fmt.Errorf("failed to validate specification: %w", err)
		}

		// Determine package name
		if packageName == "" {
			packageName = filepath.Base(outputDir)
		}

		// Create Node.js generator
		g, err := generator.NewNodeJSGenerator(packageName, outputDir)
		if err != nil {
			return fmt.Errorf("failed to create Node.js generator: %w", err)
		}

		// Generate the Node.js code
		if err := g.GenerateMetrics(spec); err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}

		return nil
	},
}

func init() {
	generateCmd.AddCommand(nodejsCmd)

	nodejsCmd.Flags().StringVarP(&nodejsPackageName, "package", "p", "", "Override package name (optional)")

	viper.BindPFlag("nodejs.package", nodejsCmd.Flags().Lookup("package"))
}
