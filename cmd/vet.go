package cmd

import (
	"fmt"
	"os"

	"github.com/jycamier/promener/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var vetFormat string

// vetCmd represents the vet command
var vetCmd = &cobra.Command{
	Use:   "vet [file.cue]",
	Short: "Validate a Promener CUE specification",
	Long: `Validate a Promener metrics specification file written in CUE.

This command performs hybrid validation:
  1. CUE schema validation - validates against embedded organizational standards
  2. Domain validation - checks the specification structure and constraints

The schema version is determined automatically from the 'version' field in the CUE file.
The corresponding embedded schema (e.g., v1, v2) is loaded and used for validation.

The validation results can be output in text (human-readable) or JSON format
for integration with CI/CD pipelines.

Examples:
  # Validate with text output
  promener vet metrics.cue

  # Use input from config file
  promener vet

  # Validate with JSON output for CI/CD
  promener vet metrics.cue --format json

  # Exit codes:
  #   0 - validation passed
  #   1 - validation failed`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var cuePath string
		if len(args) > 0 {
			cuePath = args[0]
		} else {
			cuePath = viper.GetString("input")
		}

		if cuePath == "" {
			return fmt.Errorf("input file is required (as argument, via --input flag or config file)")
		}

		// Get format from viper
		formatStr := viper.GetString("vet.format")

		// Validate format
		format := validator.OutputFormat(formatStr)
		if format != validator.FormatText && format != validator.FormatJSON {
			return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", formatStr)
		}

		// Create validator and perform validation
		v := validator.New()
		if rules := viper.GetStringSlice("rules"); len(rules) > 0 {
			v.SetRulesDirs(rules)
		}
		result, err := v.Validate(cuePath)

		// Handle system errors
		if err != nil && (result == nil || !result.HasErrors()) {
			return fmt.Errorf("validation error: %w", err)
		}

		// Format and display results
		formatter := validator.NewFormatter(format)
		output, err := formatter.Format(result)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		fmt.Print(output)

		// Exit with code 1 if validation failed based on severity threshold
		threshold := viper.GetString("severity_on_error")
		if result.Failed(threshold) {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(vetCmd)

	// Define flags
	vetCmd.Flags().StringVarP(&vetFormat, "format", "f", "text", "Output format: text or json")

	viper.BindPFlag("vet.format", vetCmd.Flags().Lookup("format"))
}
