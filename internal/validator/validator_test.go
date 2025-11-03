package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidator_ValidateAndExtract_ValidCUE(t *testing.T) {
	// Use the test.cue file from testdata
	cuePath := filepath.Join("..", "..", "testdata", "test.cue")

	// Check if file exists
	if _, err := os.Stat(cuePath); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", cuePath)
	}

	v := New()
	spec, result, err := v.ValidateAndExtract(cuePath)

	if err != nil {
		t.Fatalf("ValidateAndExtract failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if result.HasErrors() {
		t.Errorf("Expected no errors, got %d errors:", result.TotalErrors())
		for _, e := range result.CueErrors {
			t.Logf("  CUE error: %s (path: %s)", e.Message, e.Path)
		}
		for _, e := range result.DomainErrors {
			t.Logf("  Domain error: %s (path: %s)", e.Message, e.Path)
		}
	}

	if spec == nil {
		t.Fatal("Specification is nil")
	}

	// Verify basic spec fields
	if spec.Version == "" {
		t.Error("Specification version is empty")
	}

	if spec.Info.Title == "" {
		t.Error("Specification info.title is empty")
	}

	if len(spec.Services) == 0 {
		t.Error("Specification has no services")
	}
}

func TestValidator_Validate_InvalidCUE(t *testing.T) {
	// Create a temporary invalid CUE file
	tmpFile, err := os.CreateTemp("", "invalid_*.cue")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid CUE
	invalidCUE := `
package main

version: "1.0.0"
info: {
	// Missing required field 'title'
	version: "1.0.0"
}
services: {}
`
	if _, err := tmpFile.WriteString(invalidCUE); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	v := New()
	result, err := v.Validate(tmpFile.Name())

	// We expect validation to fail
	if err == nil && !result.HasErrors() {
		t.Error("Expected validation to fail for invalid CUE")
	}

	if result != nil && result.HasErrors() {
		t.Logf("Got expected validation errors: %d", result.TotalErrors())
	}
}

func TestValidator_Validate_MissingVersion(t *testing.T) {
	// Create a CUE file without version field
	tmpFile, err := os.CreateTemp("", "noversion_*.cue")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// CUE without version
	cueContent := `
package main

info: {
	title: "Test"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Test"
			version: "1.0.0"
		}
		metrics: {}
	}
}
`
	if _, err := tmpFile.WriteString(cueContent); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	v := New()
	_, err = v.Validate(tmpFile.Name())

	// Should fail because version is required
	if err == nil {
		t.Error("Expected error for missing version field")
	}

	if err != nil && err.Error() != "" {
		t.Logf("Got expected error: %v", err)
	}
}

func TestValidator_Validate_NonExistentFile(t *testing.T) {
	v := New()
	_, err := v.Validate("/nonexistent/file.cue")

	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
