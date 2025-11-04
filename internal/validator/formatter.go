package validator

//go:generate mockgen -source=formatter.go -destination=mocks/mock_formatter.go -package=mocks ValidationFormatter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputFormat represents the output format for validation results.
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
)

// ValidationFormatter is the interface for formatting validation results.
type ValidationFormatter interface {
	Format(result *ValidationResult) (string, error)
}

// Formatter formats validation results for display.
type Formatter struct {
	format OutputFormat
}

// Ensure Formatter implements ValidationFormatter
var _ ValidationFormatter = (*Formatter)(nil)

// NewFormatter creates a new formatter with the specified output format.
func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
}

// Format formats the validation result according to the configured format.
func (f *Formatter) Format(result *ValidationResult) (string, error) {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(result)
	case FormatText:
		return f.formatText(result), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", f.format)
	}
}

// formatText formats the validation result as human-readable text.
func (f *Formatter) formatText(result *ValidationResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("\n")
	if result.Valid && !result.HasErrors() {
		sb.WriteString("✓ Validation passed\n")
		return sb.String()
	}

	sb.WriteString("✗ Validation failed\n\n")

	// Domain errors
	if len(result.DomainErrors) > 0 {
		sb.WriteString(fmt.Sprintf("Domain Validation Errors (%d):\n", len(result.DomainErrors)))
		for i, err := range result.DomainErrors {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Message))
			if err.Path != "" {
				sb.WriteString(fmt.Sprintf("     Path: %s\n", err.Path))
			}
		}
		sb.WriteString("\n")
	}

	// CUE errors
	if len(result.CueErrors) > 0 {
		sb.WriteString(fmt.Sprintf("CUE Schema Validation Errors (%d):\n", len(result.CueErrors)))
		for i, err := range result.CueErrors {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Message))
			if err.Path != "" {
				sb.WriteString(fmt.Sprintf("     Path: %s\n", err.Path))
			}
			if err.Line > 0 {
				sb.WriteString(fmt.Sprintf("     Line: %d\n", err.Line))
			}
		}
		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString(fmt.Sprintf("Total errors: %d\n", result.TotalErrors()))

	return sb.String()
}

// formatJSON formats the validation result as JSON.
func (f *Formatter) formatJSON(result *ValidationResult) (string, error) {
	type jsonOutput struct {
		Valid        bool              `json:"valid"`
		TotalErrors  int               `json:"total_errors"`
		CueErrors    []ValidationError `json:"cue_errors"`
		DomainErrors []ValidationError `json:"domain_errors"`
	}

	output := jsonOutput{
		Valid:        result.Valid && !result.HasErrors(),
		TotalErrors:  result.TotalErrors(),
		CueErrors:    result.CueErrors,
		DomainErrors: result.DomainErrors,
	}

	// Handle nil slices for cleaner JSON output
	if output.CueErrors == nil {
		output.CueErrors = []ValidationError{}
	}
	if output.DomainErrors == nil {
		output.DomainErrors = []ValidationError{}
	}

	bytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(bytes), nil
}
