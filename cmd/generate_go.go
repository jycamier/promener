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
	goPackageName string
	goGenerateDI  bool
	goGenerateFx  bool
)

// goCmd represents the go command
var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Generate Go code for Prometheus metrics",
	Long: `Generate Go code for Prometheus metrics from a CUE specification file.
Generates metrics.go and optionally metrics_fx.go in the output directory.

Examples:
  promener generate go -i metrics.cue -o ./out
  promener generate go -i metrics.cue -o ./out --di --fx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get values from Viper
		inputFile := viper.GetString("input")
		outputDir := viper.GetString("output")
		packageName := viper.GetString("go.package")
		di := viper.GetBool("go.di")
		fx := viper.GetBool("go.fx")

		logging.Info("generating Go code", "input", inputFile, "output", outputDir)
		logging.Debug("generation options",
			"package", packageName,
			"di", di,
			"fx", fx,
		)

		// Validate DI flags
		if di && !fx {
			return fmt.Errorf("--di requires a DI framework flag (--fx)")
		}

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

		// Determine package name: -p flag or output directory name
		if packageName == "" {
			packageName = filepath.Base(outputDir)
			logging.Debug("using output directory as package name", "package", packageName)
		}

		logging.Debug("creating Go generator", "package", packageName)
		golangGenerator, err := generator.NewGolangGenerator(packageName, outputDir)
		if err != nil {
			logging.Error("failed to create generator", "error", err)
			return fmt.Errorf("failed to create Go generator: %w", err)
		}

		logging.Debug("generating metrics code")
		err = golangGenerator.GenerateMetrics(spec)
		if err != nil {
			logging.Error("failed to generate metrics", "error", err)
			return fmt.Errorf("failed to generate metrics code: %w", err)
		}

		if di && fx {
			logging.Debug("generating FX dependency injection code")
			err = golangGenerator.GenerateDI(spec)
			if err != nil {
				logging.Error("failed to generate DI code", "error", err)
				return fmt.Errorf("failed to generate DI code: %w", err)
			}
		}

		logging.Info("code generation complete", "output", outputDir)
		return nil
	},
}

func init() {
	generateCmd.AddCommand(goCmd)

	goCmd.Flags().StringVarP(&goPackageName, "package", "p", "", "Override package name (optional)")
	goCmd.Flags().BoolVar(&goGenerateDI, "di", false, "Generate dependency injection code (requires a DI framework flag)")
	goCmd.Flags().BoolVar(&goGenerateFx, "fx", false, "Use Uber FX framework for DI (use with --di)")

	viper.BindPFlag("go.package", goCmd.Flags().Lookup("package"))
	viper.BindPFlag("go.di", goCmd.Flags().Lookup("di"))
	viper.BindPFlag("go.fx", goCmd.Flags().Lookup("fx"))
}
