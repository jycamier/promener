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
	nodejsPackageName string
)

// nodejsCmd represents the nodejs command
var nodejsCmd = &cobra.Command{
	Use:   "nodejs",
	Short: "Generate Node.js code for Prometheus metrics",
	Long: `Generate Node.js/TypeScript code for Prometheus metrics from a YAML specification file.
Generates metrics.ts in the output directory.

Examples:
  promener generate nodejs -i metrics.yaml -o ./out
  promener generate nodejs -i metrics.yaml -o ./out -p myapp`,
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
		packageName := nodejsPackageName
		if packageName == "" {
			packageName = filepath.Base(outputDir)
		}

		// Create Node.js generator
		g, err := generator.NewNodeJSGenerator(packageName, outputDir)
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
}
