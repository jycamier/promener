/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
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

		// Create .NET generator
		g, err := generator.NewDotNetGenerator()
		if err != nil {
			return fmt.Errorf("failed to create .NET generator: %w", err)
		}

		// Generate the .NET code
		metricsFile := filepath.Join(outputDir, "Metrics.cs")
		if err := g.GenerateFile(spec, metricsFile); err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}

		fmt.Printf("✓ Generated .NET metrics code: %s\n", metricsFile)

		// Generate DI extensions if requested
		if dotnetGenerateDI {
			diFile := filepath.Join(outputDir, "Metrics.DependencyInjection.cs")
			if err := g.GenerateDIFile(spec, diFile); err != nil {
				return fmt.Errorf("failed to generate DI extensions: %w", err)
			}
			fmt.Printf("✓ Generated DI extensions: %s\n", diFile)
		}

		return nil
	},
}

func init() {
	generateCmd.AddCommand(dotnetCmd)

	dotnetCmd.Flags().StringVarP(&dotnetNamespace, "package", "p", "", "Override namespace (optional)")
	dotnetCmd.Flags().BoolVar(&dotnetGenerateDI, "di", false, "Generate dependency injection extensions (optional)")
}
