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
	dotnetNamespace  string
	dotnetGenerateDI bool
	dotnetProvider   = generator.ProviderPrometheus // default to prometheus
)

// dotnetCmd represents the dotnet command
var dotnetCmd = &cobra.Command{
	Use:   "dotnet",
	Short: "Generate .NET code for metrics (Prometheus or OpenTelemetry)",
	Long: `Generate .NET code for metrics from a CUE specification file.
Generates Metrics.cs or MetricsOtel.cs and optionally MetricsExtensions.cs in the output directory.

Examples:
  promener generate dotnet -i metrics.cue -o ./out
  promener generate dotnet -i metrics.cue -o ./out --provider otel
  promener generate dotnet -i metrics.cue -o ./out --di`,
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
		packageName := dotnetNamespace
		if packageName == "" {
			packageName = filepath.Base(outputDir)
		}

		// Create .NET generator
		g, err := generator.NewDotNetGenerator(packageName, outputDir, dotnetProvider)
		if err != nil {
			return fmt.Errorf("failed to create .NET generator: %w", err)
		}

		// Generate the .NET code
		if err := g.GenerateMetrics(spec); err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}

		// Generate DI extensions if requested
		if dotnetGenerateDI {
			if err := g.GenerateDI(spec); err != nil {
				return fmt.Errorf("failed to generate DI extensions: %w", err)
			}
		}

		return nil
	},
}

func init() {
	generateCmd.AddCommand(dotnetCmd)

	dotnetCmd.Flags().StringVarP(&dotnetNamespace, "package", "p", "", "Override namespace (optional)")
	dotnetCmd.Flags().Var(&dotnetProvider, "provider", "Metrics provider implementation (prometheus or otel)")
	dotnetCmd.Flags().BoolVar(&dotnetGenerateDI, "di", false, "Generate dependency injection extensions (optional)")
}
