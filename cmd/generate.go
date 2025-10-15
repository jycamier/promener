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
	generateDI  bool
	targetLang  string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Prometheus metrics code from YAML specification",
	Long: `Generate code for Prometheus metrics based on a YAML specification file.
Supports multiple target languages: Go, .NET, and Node.js.

Examples:
  promener generate -i metrics.yaml -o metrics.go
  promener generate -i metrics.yaml -o metrics.go -l go --fx
  promener generate -i metrics.yaml -o Metrics.cs -l dotnet
  promener generate -i metrics.yaml -o Metrics.cs -l dotnet --di
  promener generate -i metrics.yaml -o metrics.ts -l nodejs`,
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

		// Parse target language
		lang, err := generator.ParseLanguage(targetLang)
		if err != nil {
			return err
		}

		// Generate the code using language-specific generator
		if err := generator.GenerateForLanguage(spec, lang, outputFile); err != nil {
			return fmt.Errorf("failed to generate code: %w", err)
		}

		fmt.Printf("✓ Generated %s metrics code: %s\n", lang, outputFile)

		// Generate FX module if requested (Go only)
		if generateFx {
			if lang != generator.LanguageGo {
				return fmt.Errorf("--fx flag is only supported for Go language")
			}

			g, err := generator.NewGoGenerator()
			if err != nil {
				return fmt.Errorf("failed to create generator: %w", err)
			}

			fxFile := outputFile[:len(outputFile)-3] + "_fx.go"
			if err := g.GenerateFxFile(spec, fxFile); err != nil {
				return fmt.Errorf("failed to generate FX module: %w", err)
			}
			fmt.Printf("✓ Generated FX module: %s\n", fxFile)
		}

		// Generate DI extensions if requested (.NET only)
		if generateDI {
			if lang != generator.LanguageDotNet {
				return fmt.Errorf("--di flag is only supported for .NET language")
			}

			g, err := generator.NewDotNetGenerator()
			if err != nil {
				return fmt.Errorf("failed to create .NET generator: %w", err)
			}

			diFile := outputFile[:len(outputFile)-3] + ".DependencyInjection.cs"
			if err := g.GenerateDIFile(spec, diFile); err != nil {
				return fmt.Errorf("failed to generate DI extensions: %w", err)
			}
			fmt.Printf("✓ Generated DI extensions: %s\n", diFile)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input YAML specification file (required)")
	generateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (required)")
	generateCmd.Flags().StringVarP(&targetLang, "lang", "l", "go", "Target language (go, dotnet, nodejs)")
	generateCmd.Flags().StringVarP(&packageName, "package", "p", "", "Override package/namespace name (optional)")
	generateCmd.Flags().BoolVar(&generateFx, "fx", false, "Generate Uber FX module (Go only, optional)")
	generateCmd.Flags().BoolVar(&generateDI, "di", false, "Generate Dependency Injection extensions (.NET only, optional)")

	generateCmd.MarkFlagRequired("input")
	generateCmd.MarkFlagRequired("output")
}
