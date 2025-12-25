package validator

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
	}{
		{"text formatter", FormatText},
		{"json formatter", FormatJSON},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.format)
			if f == nil {
				t.Fatal("NewFormatter returned nil")
			}
			if f.format != tt.format {
				t.Errorf("Format = %v, expected %v", f.format, tt.format)
			}
		})
	}
}

func TestFormatter_FormatText_ValidResult(t *testing.T) {
	f := NewFormatter(FormatText)

	result := &ValidationResult{
		CueErrors:    []ValidationError{},
		DomainErrors: []ValidationError{},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(output, "✓ Validation passed") {
		t.Errorf("Output should contain success message, got: %s", output)
	}
}

func TestFormatter_FormatText_WithErrors(t *testing.T) {
	f := NewFormatter(FormatText)

	result := &ValidationResult{
		CueErrors: []ValidationError{
			{
				Path:    "services.default.metrics.test",
				Message: "field not allowed",
				Source:  "cue",
				Line:    10,
			},
		},
		DomainErrors: []ValidationError{
			{
				Path:    "info",
				Message: "title is required",
				Source:  "domain",
			},
		},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Check for failure indicator
	if !strings.Contains(output, "✗ Validation failed") {
		t.Errorf("Output should contain failure message")
	}

	// Check for error sections
	if !strings.Contains(output, "Domain Validation Errors") {
		t.Errorf("Output should contain domain errors section")
	}

	if !strings.Contains(output, "CUE Schema Validation Errors") {
		t.Errorf("Output should contain CUE errors section")
	}

	// Check for error details
	if !strings.Contains(output, "field not allowed") {
		t.Errorf("Output should contain CUE error message")
	}

	if !strings.Contains(output, "title is required") {
		t.Errorf("Output should contain domain error message")
	}

	// Check for path
	if !strings.Contains(output, "services.default.metrics.test") {
		t.Errorf("Output should contain error path")
	}

	// Check for total
	if !strings.Contains(output, "Total errors: 2") {
		t.Errorf("Output should contain total errors count")
	}
}

func TestFormatter_FormatJSON_ValidResult(t *testing.T) {
	f := NewFormatter(FormatJSON)

	result := &ValidationResult{
		CueErrors:    []ValidationError{},
		DomainErrors: []ValidationError{},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Parse JSON to verify it's valid
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Check valid field
	if valid, ok := parsed["valid"].(bool); !ok || !valid {
		t.Errorf("JSON should have valid=true")
	}

	// Check total_errors field
	if totalErrors, ok := parsed["total_errors"].(float64); !ok || totalErrors != 0 {
		t.Errorf("JSON should have total_errors=0")
	}
}

func TestFormatter_FormatJSON_WithErrors(t *testing.T) {
	f := NewFormatter(FormatJSON)

	result := &ValidationResult{
		CueErrors: []ValidationError{
			{
				Path:    "services.default.metrics.test",
				Message: "field not allowed",
				Source:  "cue",
				Line:    10,
			},
		},
		DomainErrors: []ValidationError{
			{
				Path:    "info",
				Message: "title is required",
				Source:  "domain",
			},
		},
	}

	output, err := f.Format(result)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Check valid field
	if valid, ok := parsed["valid"].(bool); !ok || valid {
		t.Errorf("JSON should have valid=false")
	}

	// Check total_errors
	if totalErrors, ok := parsed["total_errors"].(float64); !ok || totalErrors != 2 {
		t.Errorf("JSON should have total_errors=2, got %v", totalErrors)
	}

	// Check error arrays
	cueErrors, ok := parsed["cue_errors"].([]interface{})
	if !ok || len(cueErrors) != 1 {
		t.Errorf("JSON should have 1 CUE error")
	}

	domainErrors, ok := parsed["domain_errors"].([]interface{})
	if !ok || len(domainErrors) != 1 {
		t.Errorf("JSON should have 1 domain error")
	}
}

func TestFormatter_Format_InvalidFormat(t *testing.T) {
	f := NewFormatter(OutputFormat("invalid"))

	result := &ValidationResult{}

	_, err := f.Format(result)
	if err == nil {
		t.Fatal("Expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Error should mention unsupported format, got: %v", err)
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected bool
	}{
		{
			name: "no errors",
			result: ValidationResult{
				CueErrors:    []ValidationError{},
				DomainErrors: []ValidationError{},
			},
			expected: false,
		},
		{
			name: "has CUE errors",
			result: ValidationResult{
				CueErrors: []ValidationError{
					{Message: "test error", Source: "cue"},
				},
				DomainErrors: []ValidationError{},
			},
			expected: true,
		},
		{
			name: "has domain errors",
			result: ValidationResult{
				CueErrors: []ValidationError{},
				DomainErrors: []ValidationError{
					{Message: "test error", Source: "domain"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasErrors(); got != tt.expected {
				t.Errorf("HasErrors() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidationResult_TotalErrors(t *testing.T) {
	result := ValidationResult{
		CueErrors: []ValidationError{
			{Message: "error1", Source: "cue"},
			{Message: "error2", Source: "cue"},
		},
		DomainErrors: []ValidationError{
			{Message: "error3", Source: "domain"},
		},
	}

	expected := 3
	if got := result.TotalErrors(); got != expected {
		t.Errorf("TotalErrors() = %d, expected %d", got, expected)
	}
}
