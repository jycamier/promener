package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jycamier/promener/internal/domain"
)

func TestNewCueValidator(t *testing.T) {
	v := NewCueValidator()
	if v == nil {
		t.Fatal("NewCueValidator returned nil")
	}
	if v.ctx == nil {
		t.Fatal("CueValidator context is nil")
	}
}

func TestCueValidator_Validate_ValidSpec(t *testing.T) {
	v := NewCueValidator()

	// Create a valid specification
	spec := &domain.Specification{
		Version: "1.0",
		Info: domain.Info{
			Title:       "Test Metrics",
			Description: "Test metrics specification",
			Version:     "1.0.0",
		},
		Services: map[string]domain.Service{
			"default": {
				Info: domain.Info{
					Title:   "Default Service",
					Version: "1.0.0",
				},
				Metrics: map[string]domain.Metric{
					"test_counter": {
						Name:      "test_counter",
						Namespace: "test",
						Subsystem: "api",
						Type:      domain.MetricTypeCounter,
						Help:      "A test counter",
					},
				},
			},
		},
	}

	// Create a temporary CUE schema file
	schemaPath := createTempSchema(t)
	defer os.Remove(schemaPath)

	// Validate
	result, err := v.Validate(spec, schemaPath)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// The spec should be valid according to domain rules
	if result.HasErrors() {
		t.Errorf("Expected no errors, got %d errors", result.TotalErrors())
		for _, e := range result.DomainErrors {
			t.Logf("Domain error: %s", e.Message)
		}
		for _, e := range result.CueErrors {
			t.Logf("CUE error: %s", e.Message)
		}
	}
}

func TestCueValidator_ValidateFile(t *testing.T) {
	v := NewCueValidator()

	// Get the path to the existing test data
	yamlPath := filepath.Join("..", "..", "testdata", "simple-service.yaml")
	schemaPath := createTempSchema(t)
	defer os.Remove(schemaPath)

	// Check if the file exists
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", yamlPath)
	}

	// Validate
	result, err := v.ValidateFile(yamlPath, schemaPath)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// We expect validation to complete (may have errors depending on testdata)
	t.Logf("Validation completed with %d total errors", result.TotalErrors())
}

func TestCueValidator_Validate_InvalidSchema(t *testing.T) {
	v := NewCueValidator()

	spec := &domain.Specification{
		Version: "1.0",
		Info: domain.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Services: map[string]domain.Service{
			"default": {
				Info: domain.Info{
					Title:   "Default",
					Version: "1.0.0",
				},
				Metrics: map[string]domain.Metric{},
			},
		},
	}

	// Use a non-existent schema file
	_, err := v.Validate(spec, "/nonexistent/schema.cue")
	if err == nil {
		t.Fatal("Expected error for non-existent schema file")
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
				Valid:        true,
				CueErrors:    []ValidationError{},
				DomainErrors: []ValidationError{},
			},
			expected: false,
		},
		{
			name: "has CUE errors",
			result: ValidationResult{
				Valid: false,
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
				Valid:     false,
				CueErrors: []ValidationError{},
				DomainErrors: []ValidationError{
					{Message: "test error", Source: "domain"},
				},
			},
			expected: true,
		},
		{
			name: "has both errors",
			result: ValidationResult{
				Valid: false,
				CueErrors: []ValidationError{
					{Message: "cue error", Source: "cue"},
				},
				DomainErrors: []ValidationError{
					{Message: "domain error", Source: "domain"},
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

// createTempSchema creates a temporary CUE schema file for testing
func createTempSchema(t *testing.T) string {
	t.Helper()

	content := `
package test

#Specification: {
	version: string
	info: {
		title:       string
		description?: string
		version:     string
	}
	services: [string]: #Service
}

#Service: {
	info: {
		title:       string
		description?: string
		version:     string
	}
	servers?: [...]
	metrics: [string]: #Metric
}

#Metric: {
	name?:      string
	namespace:  string
	subsystem:  string
	type:       "counter" | "gauge" | "histogram" | "summary"
	help:       string
	labels?:    [string]: _
	constLabels?: [string]: _
	buckets?:   [...number]
}
`

	tmpFile, err := os.CreateTemp("", "test_schema_*.cue")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}
