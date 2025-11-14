package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jycamier/promener/internal/generator"
	"github.com/jycamier/promener/internal/validator"
	"github.com/spf13/cobra"
)

var (
	nodejsPackageName string
	nodejsProvider    = generator.ProviderPrometheus // default to prometheus
)

// nodejsCmd represents the nodejs command
var nodejsCmd = &cobra.Command{
	Use:   "nodejs",
	Short: "Generate Node.js code for metrics (Prometheus or OpenTelemetry)",
	Long: `Generate Node.js/TypeScript code for metrics from a CUE specification file.
Generates metrics.ts or metrics_otel.ts in the output directory.

Examples:
  promener generate nodejs -i metrics.cue -o ./out
  promener generate nodejs -i metrics.cue -o ./out --provider otel
  promener generate nodejs -i metrics.cue -o ./out -p myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Validate and extract the CUE specification
		v := validator.New()
		spec, result, err := v.ValidateAndExtract(inputFile)
		if err != nil || result.HasErrors() {
			if result != nil && result.HasErrors() {
				// Format validation errors
				formatter := validator.NewFormatter(validator.FormatText)
				output, _ := formatter.Format(result)
				fmt.Fprint(os.Stderr, output)
			}
			return fmt.Errorf("failed to validate specification: %w", err)
		}

		// Determine package name
		packageName := nodejsPackageName
		if packageName == "" {
			packageName = filepath.Base(outputDir)
		}

		// Create Node.js generator
		g, err := generator.NewNodeJSGenerator(packageName, outputDir, nodejsProvider)
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
	nodejsCmd.Flags().Var(&nodejsProvider, "provider", "Metrics provider implementation (prometheus or otel)")
}
