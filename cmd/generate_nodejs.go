package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jycamier/promener/internal/generator"
	"github.com/jycamier/promener/internal/logging"
	"github.com/jycamier/promener/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	nodejsPackageName string
)

// nodejsCmd represents the nodejs command
var nodejsCmd = &cobra.Command{
	Use:   "nodejs",
	Short: "Generate Node.js code for Prometheus metrics",
	Long: `Generate Node.js/TypeScript code for Prometheus metrics from a CUE specification file.
Generates metrics.ts in the output directory.

Examples:
  promener generate nodejs -i metrics.cue -o ./out
  promener generate nodejs -i metrics.cue -o ./out -p myapp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get values from Viper
		inputFile := viper.GetString("input")
		outputDir := viper.GetString("output")
		packageName := viper.GetString("nodejs.package")

		logging.Info("generating Node.js code", "input", inputFile, "output", outputDir)
		logging.Debug("generation options", "package", packageName)

		// Create output directory if it doesn't exist
		logging.Debug("creating output directory", "path", outputDir)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			logging.Error("failed to create output directory", "error", err)
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Validate and extract the CUE specification
		logging.Debug("validating specification", "file", inputFile)
		v := validator.New()
		if rules := viper.GetStringSlice("rules"); len(rules) > 0 {
			logging.Debug("loading rego rules", "directories", rules)
			v.SetRulesDirs(rules)
		}
		spec, result, err := v.ValidateAndExtract(inputFile)
		threshold := viper.GetString("severity_on_error")

		if err != nil || result.Failed(threshold) {
			if result != nil && result.HasErrors() {
				logging.Debug("validation errors found",
					"cue_errors", len(result.CueErrors),
					"domain_errors", len(result.DomainErrors),
					"rego_errors", len(result.RegoErrors),
				)
				// Format validation errors
				formatter := validator.NewFormatter(validator.FormatText)
				output, formatErr := formatter.Format(result)
				if formatErr != nil {
					logging.Error("failed to format validation errors", "error", formatErr)
				}
				fmt.Fprint(os.Stderr, output)
			}
			if result != nil && result.Failed(threshold) {
				logging.Error("validation failed threshold", "threshold", threshold)
				return fmt.Errorf("failed to validate specification (threshold: %s)", threshold)
			}
			logging.Error("validation failed", "error", err)
			return fmt.Errorf("failed to validate specification: %w", err)
		}

		// Determine package name
		if packageName == "" {
			packageName = filepath.Base(outputDir)
			logging.Debug("using output directory as package name", "package", packageName)
		}

		// Create Node.js generator
		logging.Debug("creating Node.js generator", "package", packageName)
		g, err := generator.NewNodeJSGenerator(packageName, outputDir)
		if err != nil {
			logging.Error("failed to create generator", "error", err)
			return fmt.Errorf("failed to create Node.js generator: %w", err)
		}

		// Generate the Node.js code
		logging.Debug("generating metrics code")
		if err := g.GenerateMetrics(spec); err != nil {
			logging.Error("failed to generate metrics", "error", err)
			return fmt.Errorf("failed to generate code: %w", err)
		}

		logging.Info("code generation complete", "output", outputDir)
		return nil
	},
}

func init() {
	generateCmd.AddCommand(nodejsCmd)

	nodejsCmd.Flags().StringVarP(&nodejsPackageName, "package", "p", "", "Override package name (optional)")

	viper.BindPFlag("nodejs.package", nodejsCmd.Flags().Lookup("package"))
}
