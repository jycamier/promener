package validator

import (
	"strings"
	"testing"
)

func TestGetSchemaForVersion_Valid(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"v1.0.0", "1.0.0"},
		{"v1.0", "1.0"},
		{"v1.2.3", "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := GetSchemaForVersion(tt.version)
			if err != nil {
				t.Fatalf("GetSchemaForVersion(%s) failed: %v", tt.version, err)
			}

			if schema == "" {
				t.Errorf("GetSchemaForVersion(%s) returned empty schema", tt.version)
			}

			// Schema should contain 'package' keyword (basic CUE syntax check)
			if !strings.Contains(schema, "package") {
				t.Errorf("Schema doesn't appear to be valid CUE (missing 'package')")
			}
		})
	}
}

func TestGetSchemaForVersion_UnsupportedVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"v999", "999.0.0"},
		{"v2", "2.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetSchemaForVersion(tt.version)
			if err == nil {
				t.Errorf("Expected error for unsupported version %s", tt.version)
			}

			if !strings.Contains(err.Error(), "unsupported schema version") {
				t.Errorf("Expected 'unsupported schema version' error, got: %v", err)
			}
		})
	}
}

func TestGetSchemaForVersion_InvalidFormat(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"no_dots", "abc"},
		{"invalid", "v1.0.0"},
		{"just_dots", "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetSchemaForVersion(tt.version)
			if err == nil {
				t.Errorf("Expected error for invalid version format %s", tt.version)
			}
		})
	}
}

func TestGetSchemaForVersion_EmptyVersion(t *testing.T) {
	_, err := GetSchemaForVersion("")
	if err == nil {
		t.Error("Expected error for empty version")
	}

	if !strings.Contains(err.Error(), "version is required") {
		t.Errorf("Expected 'version is required' error, got: %v", err)
	}
}
