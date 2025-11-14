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
	goPackageName string
	goGenerateDI  bool
	goGenerateFx  bool
	goProvider    = generator.ProviderPrometheus // default to prometheus
)

// goCmd represents the go command
var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Generate Go code for metrics (Prometheus or OpenTelemetry)",
	Long: `Generate Go code for metrics from a CUE specification file.
Generates metrics_interface.go, metrics_validation.go, metrics_prometheus.go or metrics_otel.go,
and optionally metrics_fx.go in the output directory.

Examples:
  promener generate go -i metrics.cue -o ./out
  promener generate go -i metrics.cue -o ./out --provider otel
  promener generate go -i metrics.cue -o ./out --di --fx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate DI flags
		if goGenerateDI && !goGenerateFx {
			return fmt.Errorf("--di requires a DI framework flag (--fx)")
		}

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

		// Determine package name: -p flag or output directory name
		packageName := goPackageName
		if packageName == "" {
			packageName = filepath.Base(outputDir)
		}

		golangGenerator, err := generator.NewGolangGenerator(packageName, outputDir, goProvider)
		if err != nil {
			return err
		}
		err = golangGenerator.GenerateMetrics(spec)
		if err != nil {
			return err
		}
		if goGenerateDI && goGenerateFx {
			err = golangGenerator.GenerateDI(spec)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	generateCmd.AddCommand(goCmd)

	goCmd.Flags().StringVarP(&goPackageName, "package", "p", "", "Override package name (optional)")
	goCmd.Flags().Var(&goProvider, "provider", "Metrics provider implementation (prometheus or otel)")
	goCmd.Flags().BoolVar(&goGenerateDI, "di", false, "Generate dependency injection code (requires a DI framework flag)")
	goCmd.Flags().BoolVar(&goGenerateFx, "fx", false, "Use Uber FX framework for DI (use with --di)")
}
