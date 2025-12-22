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
	dotnetNamespace  string
	dotnetGenerateDI bool
)

// dotnetCmd represents the dotnet command
var dotnetCmd = &cobra.Command{
	Use:   "dotnet",
	Short: "Generate .NET code for Prometheus metrics",
	Long: `Generate .NET code for Prometheus metrics from a CUE specification file.
Generates Metrics.cs and optionally Metrics.DependencyInjection.cs in the output directory.

Examples:
  promener generate dotnet -i metrics.cue -o ./out
  promener generate dotnet -i metrics.cue -o ./out --di`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get values from Viper
		inputFile := viper.GetString("input")
		outputDir := viper.GetString("output")
		packageName := viper.GetString("dotnet.package")
		di := viper.GetBool("dotnet.di")

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
		if packageName == "" {
			packageName = filepath.Base(outputDir)
		}

		// Create .NET generator
		g, err := generator.NewDotNetGenerator(packageName, outputDir)
		if err != nil {
			return fmt.Errorf("failed to create .NET generator: %w", err)
		}

		// Generate the .NET code
		if err := g.GenerateMetrics(spec); err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}

		// Generate DI extensions if requested
		if di {
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
	dotnetCmd.Flags().BoolVar(&dotnetGenerateDI, "di", false, "Generate dependency injection extensions (optional)")

	viper.BindPFlag("dotnet.package", dotnetCmd.Flags().Lookup("package"))
	viper.BindPFlag("dotnet.di", dotnetCmd.Flags().Lookup("di"))
}
