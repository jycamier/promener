package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jycamier/promener/internal/generator"
	"github.com/jycamier/promener/internal/parser"
	"github.com/spf13/cobra"
)

var (
	dotnetNamespace  string
	dotnetGenerateDI bool
)

// dotnetCmd represents the dotnet command
var dotnetCmd = &cobra.Command{
	Use:   "dotnet",
	Short: "Generate .NET code for Prometheus metrics",
	Long: `Generate .NET code for Prometheus metrics from a YAML specification file.
Generates Metrics.cs and optionally Metrics.DependencyInjection.cs in the output directory.

Examples:
  promener generate dotnet -i metrics.yaml -o ./out
  promener generate dotnet -i metrics.yaml -o ./out --di`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Parse the YAML specification
		p := parser.New()
		spec, err := p.ParseFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to parse specification: %w", err)
		}

		// Determine package name
		packageName := dotnetNamespace
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
	dotnetCmd.Flags().BoolVar(&dotnetGenerateDI, "di", false, "Generate dependency injection extensions (optional)")
}
