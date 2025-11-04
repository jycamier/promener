package validator

import (
	"os"
	"path/filepath"
	"testing"

	"cuelang.org/go/cue"
)

func TestCueLoader_LoadAndValidate(t *testing.T) {
	tests := []struct {
		name       string
		cuePath    string
		wantErr    bool
		wantValid  bool
		errContains string
	}{
		{
			name:       "CUE file with modules",
			cuePath:    "../../testdata/with_cue_mod/metrics.cue",
			wantErr:    false,
			wantValid:  true,
		},
		{
			name:        "missing version field",
			cuePath:     createTempCueFile(t, "package test\n\ninfo: {\n\ttitle: \"Test\"\n}"),
			wantErr:     true,
			errContains: "'version' field is required",
		},
		{
			name:        "non-existent file",
			cuePath:     "nonexistent.cue",
			wantErr:     true,
			errContains: "failed to load CUE file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewCueLoader()

			value, result, err := loader.LoadAndValidate(tt.cuePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("CueLoader.LoadAndValidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("CueLoader.LoadAndValidate() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if !value.Exists() {
					t.Error("CueLoader.LoadAndValidate() returned non-existent value")
				}

				if result == nil {
					t.Error("CueLoader.LoadAndValidate() returned nil result")
				}

				if result != nil && result.Valid != tt.wantValid {
					t.Errorf("CueLoader.LoadAndValidate() result.Valid = %v, want %v", result.Valid, tt.wantValid)
				}
			}
		})
	}
}

func TestCueLoader_VersionExtraction(t *testing.T) {
	tests := []struct {
		name        string
		cueContent  string
		wantVersion string
		wantErr     bool
	}{
		{
			name: "version 1.0.0",
			cueContent: `package test
version: "1.0.0"
info: {
	title: "Test"
	version: "1.0.0"
}
services: {}`,
			wantVersion: "1.0.0",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary CUE file
			tmpPath := createTempCueFile(t, tt.cueContent)
			defer func() { _ = os.Remove(tmpPath) }()

			loader := NewCueLoader()
			value, _, err := loader.LoadAndValidate(tmpPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("CueLoader.LoadAndValidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Extract version to verify
				versionValue := value.LookupPath(cue.ParsePath("version"))
				if !versionValue.Exists() {
					t.Error("version field not found in loaded CUE value")
					return
				}

				version, err := versionValue.String()
				if err != nil {
					t.Errorf("Failed to extract version string: %v", err)
					return
				}

				if version != tt.wantVersion {
					t.Errorf("version = %q, want %q", version, tt.wantVersion)
				}
			}
		})
	}
}

func TestCueLoader_SchemaValidation(t *testing.T) {
	tests := []struct {
		name       string
		cueContent string
		wantValid  bool
		wantErrors bool
	}{
		{
			name: "valid specification",
			cueContent: `package test
version: "1.0.0"
info: {
	title: "Test Metrics"
	version: "1.0.0"
}
services: {
	default: {
		info: {
			title: "Default Service"
			version: "1.0.0"
		}
		metrics: {
			test_counter: {
				type: "counter"
				help: "Test counter metric"
			}
		}
	}
}`,
			wantValid:  true,
			wantErrors: false,
		},
		{
			name: "completely invalid CUE syntax",
			cueContent: `package test
version: "1.0.0"
this is not valid CUE syntax {{{
`,
			wantValid:  false,
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpPath := createTempCueFile(t, tt.cueContent)
			defer func() { _ = os.Remove(tmpPath) }()

			loader := NewCueLoader()
			_, result, err := loader.LoadAndValidate(tmpPath)

			if err != nil {
				t.Logf("LoadAndValidate error: %v", err)
			}

			if result == nil {
				if !tt.wantErrors {
					t.Error("Expected result but got nil")
				}
				return
			}

			if result.Valid != tt.wantValid {
				t.Errorf("result.Valid = %v, want %v", result.Valid, tt.wantValid)
			}

			hasErrors := len(result.CueErrors) > 0 || len(result.DomainErrors) > 0
			if hasErrors != tt.wantErrors {
				t.Errorf("hasErrors = %v, want %v (CueErrors: %d, DomainErrors: %d)",
					hasErrors, tt.wantErrors, len(result.CueErrors), len(result.DomainErrors))
			}
		})
	}
}

func TestCueLoader_ModuleSupport(t *testing.T) {
	// Test with the actual with_cue_mod example if it exists
	modPath := "../../testdata/with_cue_mod/metrics.cue"
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		t.Skip("Skipping module test: testdata/with_cue_mod/metrics.cue not found")
	}

	loader := NewCueLoader()
	value, result, err := loader.LoadAndValidate(modPath)

	if err != nil {
		t.Fatalf("LoadAndValidate() error = %v", err)
	}

	if !value.Exists() {
		t.Error("LoadAndValidate() returned non-existent value")
	}

	if result == nil {
		t.Error("LoadAndValidate() returned nil result")
	}

	if result != nil && !result.Valid {
		t.Errorf("LoadAndValidate() result.Valid = false, expected true")
		if len(result.CueErrors) > 0 {
			t.Logf("CUE errors: %+v", result.CueErrors)
		}
		if len(result.DomainErrors) > 0 {
			t.Logf("Domain errors: %+v", result.DomainErrors)
		}
	}
}

func TestNewCueLoader(t *testing.T) {
	loader := NewCueLoader()

	if loader == nil {
		t.Fatal("NewCueLoader() returned nil")
	}

	if loader.ctx == nil {
		t.Error("NewCueLoader() created loader with nil context")
	}
}

// Helper function to create a temporary CUE file for testing
func createTempCueFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.cue")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp CUE file: %v", err)
	}

	return tmpFile
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
