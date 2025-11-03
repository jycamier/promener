package cmd

import (
	"fmt"
	"os"

	"github.com/jycamier/promener/internal/validator"
	"github.com/spf13/cobra"
)

var vetFormat string

// vetCmd represents the vet command
var vetCmd = &cobra.Command{
	Use:   "vet <file.cue>",
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

  # Validate with JSON output for CI/CD
  promener vet metrics.cue --format json

  # Exit codes:
  #   0 - validation passed
  #   1 - validation failed`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cuePath := args[0]

		// Validate format
		format := validator.OutputFormat(vetFormat)
		if format != validator.FormatText && format != validator.FormatJSON {
			return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", vetFormat)
		}

		// Create validator and perform validation
		v := validator.New()
		result, err := v.Validate(cuePath)

		// Handle validation errors
		var validationFailed bool
		if err != nil {
			// Check if it's a validation error (result exists) or a system error
			if result != nil && result.HasErrors() {
				validationFailed = true
			} else {
				// System error (file not found, invalid CUE, etc.)
				return fmt.Errorf("validation error: %w", err)
			}
		}

		if result.HasErrors() {
			validationFailed = true
		}

		// Format and display results
		formatter := validator.NewFormatter(format)
		output, err := formatter.Format(result)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		fmt.Print(output)

		// Exit with code 1 if validation failed
		if validationFailed || !result.Valid {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(vetCmd)

	// Define flags
	vetCmd.Flags().StringVarP(&vetFormat, "format", "f", "text", "Output format: text or json")
}
