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
	goPackageName string
	goGenerateDI  bool
	goGenerateFx  bool
)

// goCmd represents the go command
var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Generate Go code for Prometheus metrics",
	Long: `Generate Go code for Prometheus metrics from a YAML specification file.
Generates metrics.go and optionally metrics_fx.go in the output directory.

Examples:
  promener generate go -i metrics.yaml -o ./out
  promener generate go -i metrics.yaml -o ./out --di --fx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate DI flags
		if goGenerateDI && !goGenerateFx {
			return fmt.Errorf("--di requires a DI framework flag (--fx)")
		}

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

		// Determine package name: -p flag or output directory name
		packageName := goPackageName
		if packageName == "" {
			packageName = filepath.Base(outputDir)
		}

		golangGenerator, err := generator.NewGolangGenerator(packageName, outputDir)
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
	goCmd.Flags().BoolVar(&goGenerateDI, "di", false, "Generate dependency injection code (requires a DI framework flag)")
	goCmd.Flags().BoolVar(&goGenerateFx, "fx", false, "Use Uber FX framework for DI (use with --di)")
}
