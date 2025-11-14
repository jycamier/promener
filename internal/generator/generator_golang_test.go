package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jycamier/promener/internal/domain"
)

func TestGolangGenerator_GenerateMetrics(t *testing.T) {
	tests := []struct {
		name    string
		spec    *domain.Specification
		wantErr bool
		checks  []string // Content to verify in generated file
	}{
		{
			name: "simple counter without labels",
			spec: &domain.Specification{
				Info: domain.Info{
					Title:   "Test Metrics",
					Version: "1.0.0",
				},
				Services: map[string]domain.Service{
					"default": {
						Info: domain.Info{
							Title:   "Default Service",
							Version: "1.0.0",
						},
						Metrics: map[string]domain.Metric{
							"requests_total": {
								Name:      "requests_total",
								Namespace: "http",
								Subsystem: "server",
								Type:      domain.MetricTypeCounter,
								Help:      "Total HTTP requests",
								Labels:    []domain.LabelDefinition{},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: []string{
				"package testpackage",
				"prometheus.NewCounter",
				"http_server_requests_total",
			},
		},
		{
			name: "counter with labels",
			spec: &domain.Specification{
				Info: domain.Info{
					Title:   "Test Metrics",
					Version: "1.0.0",
				},
				Services: map[string]domain.Service{
					"default": {
						Info: domain.Info{
							Title:   "Default Service",
							Version: "1.0.0",
						},
						Metrics: map[string]domain.Metric{
							"requests_total": {
								Name:      "requests_total",
								Namespace: "http",
								Subsystem: "server",
								Type:      domain.MetricTypeCounter,
								Help:      "Total HTTP requests",
								Labels: []domain.LabelDefinition{
									{Name: "method", Description: "HTTP method"},
									{Name: "status", Description: "HTTP status code"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: []string{
				"package testpackage",
				"CounterVec",
				"method",
				"status",
				"WithLabelValues",
			},
		},
		{
			name: "gauge metric",
			spec: &domain.Specification{
				Info: domain.Info{
					Title:   "Test Metrics",
					Version: "1.0.0",
				},
				Services: map[string]domain.Service{
					"default": {
						Info: domain.Info{
							Title:   "Default Service",
							Version: "1.0.0",
						},
						Metrics: map[string]domain.Metric{
							"memory_usage": {
								Name:      "memory_usage",
								Namespace: "process",
								Type:      domain.MetricTypeGauge,
								Help:      "Memory usage in bytes",
								Labels:    []domain.LabelDefinition{},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: []string{
				"prometheus.NewGauge",
				"process_memory_usage",
			},
		},
		{
			name: "histogram metric",
			spec: &domain.Specification{
				Info: domain.Info{
					Title:   "Test Metrics",
					Version: "1.0.0",
				},
				Services: map[string]domain.Service{
					"default": {
						Info: domain.Info{
							Title:   "Default Service",
							Version: "1.0.0",
						},
						Metrics: map[string]domain.Metric{
							"request_duration": {
								Name:      "request_duration",
								Namespace: "http",
								Type:      domain.MetricTypeHistogram,
								Help:      "Request duration in seconds",
								Labels:    []domain.LabelDefinition{},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: []string{
				"prometheus.NewHistogram",
				"http_request_duration",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary output directory
			tmpDir, err := os.MkdirTemp("", "promener_test_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Create generator
			gen, err := NewGolangGenerator("testpackage", tmpDir, ProviderPrometheus)
			if err != nil {
				t.Fatalf("NewGolangGenerator() error = %v", err)
			}

			// Generate metrics
			err = gen.GenerateMetrics(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("GolangGenerator.GenerateMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check that metrics_interface.go was created
			interfacePath := filepath.Join(tmpDir, "metrics_interface.go")
			if _, err := os.Stat(interfacePath); os.IsNotExist(err) {
				t.Errorf("Expected metrics_interface.go to be created at %s", interfacePath)
				return
			}

			// Check that metrics_validation.go was created
			validationPath := filepath.Join(tmpDir, "metrics_validation.go")
			if _, err := os.Stat(validationPath); os.IsNotExist(err) {
				t.Errorf("Expected metrics_validation.go to be created at %s", validationPath)
				return
			}

			// Check that metrics_prometheus.go was created
			metricsPath := filepath.Join(tmpDir, "metrics_prometheus.go")
			if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
				t.Errorf("Expected metrics_prometheus.go to be created at %s", metricsPath)
				return
			}

			// Read generated file
			content, err := os.ReadFile(metricsPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			contentStr := string(content)

			// Verify content checks
			for _, check := range tt.checks {
				if !strings.Contains(contentStr, check) {
					t.Errorf("Generated file missing expected content: %q", check)
				}
			}

			// Verify file is not empty
			if len(content) == 0 {
				t.Error("Generated file is empty")
			}
		})
	}
}

func TestGolangGenerator_GenerateDI(t *testing.T) {
	tests := []struct {
		name    string
		spec    *domain.Specification
		wantErr bool
		checks  []string
	}{
		{
			name: "FX DI generation",
			spec: &domain.Specification{
				Info: domain.Info{
					Title:   "Test Metrics",
					Version: "1.0.0",
				},
				Services: map[string]domain.Service{
					"default": {
						Info: domain.Info{
							Title:   "Default Service",
							Version: "1.0.0",
						},
						Metrics: map[string]domain.Metric{
							"requests_total": {
								Name:      "requests_total",
								Namespace: "http",
								Type:      domain.MetricTypeCounter,
								Help:      "Total requests",
								Labels:    []domain.LabelDefinition{},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: []string{
				"package testpackage",
				"fx.Module",
				"fx.Provide",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary output directory
			tmpDir, err := os.MkdirTemp("", "promener_test_*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			// Create generator
			gen, err := NewGolangGenerator("testpackage", tmpDir, ProviderPrometheus)
			if err != nil {
				t.Fatalf("NewGolangGenerator() error = %v", err)
			}

			// Generate DI
			err = gen.GenerateDI(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("GolangGenerator.GenerateDI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check that metrics_fx.go was created
			fxPath := filepath.Join(tmpDir, "metrics_fx.go")
			if _, err := os.Stat(fxPath); os.IsNotExist(err) {
				t.Errorf("Expected metrics_fx.go to be created at %s", fxPath)
				return
			}

			// Read generated file
			content, err := os.ReadFile(fxPath)
			if err != nil {
				t.Fatalf("Failed to read generated file: %v", err)
			}

			contentStr := string(content)

			// Verify content checks
			for _, check := range tt.checks {
				if !strings.Contains(contentStr, check) {
					t.Errorf("Generated file missing expected content: %q", check)
				}
			}

			// Verify file is not empty
			if len(content) == 0 {
				t.Error("Generated file is empty")
			}
		})
	}
}

func TestGolangGenerator_Creation(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		outputPath  string
		wantErr     bool
	}{
		{
			name:        "valid creation",
			packageName: "metrics",
			outputPath:  "/tmp/test",
			wantErr:     false,
		},
		{
			name:        "empty package name",
			packageName: "",
			outputPath:  "/tmp/test",
			wantErr:     false, // Empty package name is handled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGolangGenerator(tt.packageName, tt.outputPath, ProviderPrometheus)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGolangGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gen == nil {
				t.Error("NewGolangGenerator() returned nil generator")
			}
		})
	}
}

func TestGoTemplateDataBuilder_BuildTemplateData(t *testing.T) {
	builder := NewGoTemplateDataBuilder()

	spec := &domain.Specification{
		Info: domain.Info{
			Title:   "Test",
			Version: "1.0.0",
		},
		Services: map[string]domain.Service{
			"default": {
				Info: domain.Info{
					Title:   "Default Service",
					Version: "1.0.0",
				},
				Metrics: map[string]domain.Metric{
					"requests_total": {
						Name:      "requests_total",
						Namespace: "http",
						Subsystem: "server",
						Type:      domain.MetricTypeCounter,
						Help:      "Total requests",
						Labels: []domain.LabelDefinition{
							{Name: "method"},
							{Name: "status"},
						},
					},
				},
			},
		},
	}

	data := builder.BuildTemplateData(spec, "testpkg")

	if data == nil {
		t.Fatal("BuildTemplateData() returned nil")
	}

	if data.PackageName != "testpkg" {
		t.Errorf("PackageName = %q, want %q", data.PackageName, "testpkg")
	}

	if len(data.Namespaces) == 0 {
		t.Error("BuildTemplateData() returned no namespaces")
	}

	// Verify namespace structure
	found := false
	for _, ns := range data.Namespaces {
		if ns.Name == "Http" {
			found = true
			if len(ns.Subsystems) == 0 {
				t.Error("Http namespace has no subsystems")
			}
			for _, ss := range ns.Subsystems {
				if ss.Name == "Server" {
					if len(ss.Metrics) == 0 {
						t.Error("Server subsystem has no metrics")
					}
				}
			}
		}
	}

	if !found {
		t.Error("Http namespace not found in template data")
	}
}
