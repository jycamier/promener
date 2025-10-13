/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/jycamier/promener/internal/htmlgen"
	"github.com/jycamier/promener/internal/parser"
	"github.com/spf13/cobra"
)

var (
	htmlInputFile  string
	htmlOutputFile string
)

// htmlCmd represents the html command
var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "Generate HTML documentation from YAML specification",
	Long: `Generate beautiful HTML documentation for your Prometheus metrics.

The HTML documentation includes:
- Interactive search and filtering
- Dark mode support
- PromQL query examples with copy button
- Grafana dashboard examples
- Alertmanager alert rule examples
- Detailed label descriptions

Example:
  promener html -i metrics.yaml -o docs/metrics.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse the YAML specification
		p := parser.New()
		spec, err := p.ParseFile(htmlInputFile)
		if err != nil {
			return fmt.Errorf("failed to parse specification: %w", err)
		}

		// Generate HTML
		g, err := htmlgen.New()
		if err != nil {
			return fmt.Errorf("failed to create HTML generator: %w", err)
		}

		if err := g.GenerateFile(spec, htmlOutputFile); err != nil {
			return fmt.Errorf("failed to generate HTML: %w", err)
		}

		fmt.Printf("✓ Generated HTML documentation: %s\n", htmlOutputFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)

	htmlCmd.Flags().StringVarP(&htmlInputFile, "input", "i", "", "Input YAML specification file (required)")
	htmlCmd.Flags().StringVarP(&htmlOutputFile, "output", "o", "", "Output HTML file (required)")

	htmlCmd.MarkFlagRequired("input")
	htmlCmd.MarkFlagRequired("output")
}
