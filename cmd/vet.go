package cmd

import (
	"fmt"
	"os"

	"github.com/jycamier/promener/internal/parser"
	"github.com/jycamier/promener/internal/validator"
	"github.com/spf13/cobra"
)

var (
	vetInputFile  string
	vetSchemaFile string
	vetFormat     string
)

// vetCmd represents the vet command
var vetCmd = &cobra.Command{
	Use:   "vet",
	Short: "Validate a Promener YAML specification against a CUE schema",
	Long: `Validate a Promener metrics specification file against a CUE schema.

This command performs hybrid validation:
  1. Domain validation - checks the specification structure and constraints
  2. CUE schema validation - validates against organizational standards

The validation results can be output in text (human-readable) or JSON format
for integration with CI/CD pipelines.

Examples:
  # Validate with text output
  promener vet -i metrics.yaml -s schema.cue

  # Validate with JSON output for CI/CD
  promener vet -i metrics.yaml -s schema.cue --format json

  # Exit codes:
  #   0 - validation passed
  #   1 - validation failed`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate inputs
		if vetInputFile == "" {
			return fmt.Errorf("input file is required (use -i or --input)")
		}
		if vetSchemaFile == "" {
			return fmt.Errorf("schema file is required (use -s or --schema)")
		}

		// Validate format
		format := validator.OutputFormat(vetFormat)
		if format != validator.FormatText && format != validator.FormatJSON {
			return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", vetFormat)
		}

		// Parse the YAML file
		p := parser.New()
		spec, err := p.ParseFile(vetInputFile)
		if err != nil {
			return fmt.Errorf("failed to parse input file: %w", err)
		}

		// Create validator and perform validation
		v := validator.NewCueValidator()
		result, err := v.Validate(spec, vetSchemaFile)
		if err != nil {
			return fmt.Errorf("validation error: %w", err)
		}

		// Format and display results
		formatter := validator.NewFormatter(format)
		output, err := formatter.Format(result)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		fmt.Print(output)

		// Exit with code 1 if validation failed
		if result.HasErrors() || !result.Valid {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(vetCmd)

	// Define flags
	vetCmd.Flags().StringVarP(&vetInputFile, "input", "i", "", "Input YAML file to validate (required)")
	vetCmd.Flags().StringVarP(&vetSchemaFile, "schema", "s", "", "CUE schema file for validation (required)")
	vetCmd.Flags().StringVarP(&vetFormat, "format", "f", "text", "Output format: text or json")

	// Mark required flags
	vetCmd.MarkFlagRequired("input")
	vetCmd.MarkFlagRequired("schema")
}
