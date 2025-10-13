/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/jycamier/promener/internal/generator"
	"github.com/jycamier/promener/internal/parser"
	"github.com/spf13/cobra"
)

var (
	inputFile   string
	outputFile  string
	packageName string
	generateFx  bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Prometheus metrics code from YAML specification",
	Long: `Generate Go code for Prometheus metrics based on a YAML specification file.

Example:
  promener generate -i metrics.yaml -o metrics.go`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse the YAML specification
		p := parser.New()
		spec, err := p.ParseFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to parse specification: %w", err)
		}

		// Override package name if provided via flag
		if packageName != "" {
			spec.Info.Package = packageName
		}

		// Generate the code
		g, err := generator.New()
		if err != nil {
			return fmt.Errorf("failed to create generator: %w", err)
		}

		if err := g.GenerateFile(spec, outputFile); err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}

		fmt.Printf("✓ Generated metrics code: %s\n", outputFile)

		// Generate FX module if requested
		if generateFx {
			fxFile := outputFile[:len(outputFile)-3] + "_fx.go"
			if err := g.GenerateFxFile(spec, fxFile); err != nil {
				return fmt.Errorf("failed to generate FX module: %w", err)
			}
			fmt.Printf("✓ Generated FX module: %s\n", fxFile)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input YAML specification file (required)")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output Go file (required)")
	generateCmd.Flags().StringVarP(&packageName, "package", "p", "", "Override package name (optional)")
	generateCmd.Flags().BoolVar(&generateFx, "fx", false, "Generate Uber FX module (optional)")

	generateCmd.MarkFlagRequired("input")
	generateCmd.MarkFlagRequired("output")
}
