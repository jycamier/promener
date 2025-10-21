package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jycamier/promener/pkg/docs"
)

func TestHTMLBuilder_Basic(t *testing.T) {
	// Create a simple builder
	builder := docs.NewHTMLBuilder("My Platform", "1.0.0")

	// Add a service manually
	service := docs.Service{
		Info: docs.Info{
			Title:   "Test Service",
			Version: "1.0.0",
		},
		Metrics: map[string]docs.Metric{
			"test_metric": {
				Name:      "test_metric",
				Namespace: "test",
				Subsystem: "app",
				Type:      docs.MetricTypeCounter,
				Help:      "A test metric",
			},
		},
	}

	builder.AddService("test-service", service)

	// Build HTML
	html, err := builder.BuildHTML()
	if err != nil {
		t.Fatalf("BuildHTML failed: %v", err)
	}

	if len(html) == 0 {
		t.Fatal("Expected non-empty HTML")
	}

	htmlStr := string(html)
	if !strings.Contains(htmlStr, "My Platform") {
		t.Error("HTML should contain platform title")
	}
}

func TestHTMLBuilder_AddFromSpec(t *testing.T) {
	// Load a test spec
	spec, err := docs.LoadSpec("../../testdata/simple-service.yaml")
	if err != nil {
		t.Fatalf("Failed to load test spec: %v", err)
	}

	// Create builder and add spec
	builder := docs.NewHTMLBuilder("Multi-Service Platform", "2.0.0")
	builder.SetDescription("A platform with multiple services")
	builder.AddFromSpec(spec)

	// Build HTML
	html, err := builder.BuildHTML()
	if err != nil {
		t.Fatalf("BuildHTML failed: %v", err)
	}

	if len(html) == 0 {
		t.Fatal("Expected non-empty HTML")
	}

	htmlStr := string(html)
	if !strings.Contains(htmlStr, "Multi-Service Platform") {
		t.Error("HTML should contain platform title")
	}
}

func TestHTMLBuilder_MultipleSpecs(t *testing.T) {
	// Create two simple specs in memory
	yaml1 := []byte(`
version: "1.0"
info:
  title: "Service 1"
  version: "1.0.0"
services:
  api-gateway:
    info:
      title: "API Gateway"
      version: "1.0.0"
    metrics:
      requests_total:
        namespace: http
        subsystem: server
        type: counter
        help: "Total requests"
`)

	yaml2 := []byte(`
version: "1.0"
info:
  title: "Service 2"
  version: "1.0.0"
services:
  user-service:
    info:
      title: "User Service"
      version: "1.0.0"
    metrics:
      users_total:
        namespace: app
        subsystem: users
        type: gauge
        help: "Total users"
`)

	spec1, err := docs.LoadSpecFromBytes(yaml1)
	if err != nil {
		t.Fatalf("Failed to load spec1: %v", err)
	}

	spec2, err := docs.LoadSpecFromBytes(yaml2)
	if err != nil {
		t.Fatalf("Failed to load spec2: %v", err)
	}

	// Build multi-service HTML
	builder := docs.NewHTMLBuilder("My Platform", "1.0.0")
	builder.AddFromSpec(spec1)
	builder.AddFromSpec(spec2)

	html, err := builder.BuildHTML()
	if err != nil {
		t.Fatalf("BuildHTML failed: %v", err)
	}

	htmlStr := string(html)

	// Verify both services are in the output
	if !strings.Contains(htmlStr, "api-gateway") {
		t.Error("HTML should contain api-gateway service")
	}
	if !strings.Contains(htmlStr, "user-service") {
		t.Error("HTML should contain user-service service")
	}
}

func TestHTMLBuilder_BuildHTMLFile(t *testing.T) {
	// Create a temp directory for output
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.html")

	// Create a simple spec
	yaml := []byte(`
version: "1.0"
info:
  title: "Test"
  version: "1.0.0"
services:
  test:
    info:
      title: "Test Service"
      version: "1.0.0"
    metrics:
      test_metric:
        namespace: test
        subsystem: app
        type: counter
        help: "Test"
`)

	spec, err := docs.LoadSpecFromBytes(yaml)
	if err != nil {
		t.Fatalf("Failed to load spec: %v", err)
	}

	// Build and write to file
	builder := docs.NewHTMLBuilder("Test Platform", "1.0.0")
	builder.AddFromSpec(spec)

	err = builder.BuildHTMLFile(outputPath)
	if err != nil {
		t.Fatalf("BuildHTMLFile failed: %v", err)
	}

	// Verify file exists and has content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("Output file is empty")
	}

	if !strings.Contains(string(content), "Test Platform") {
		t.Error("Output file should contain platform title")
	}
}

func TestHTMLBuilder_NoServices(t *testing.T) {
	builder := docs.NewHTMLBuilder("Empty Platform", "1.0.0")

	// Should fail with no services
	_, err := builder.BuildHTML()
	if err == nil {
		t.Fatal("Expected error when building with no services")
	}

	if !strings.Contains(err.Error(), "no services") {
		t.Errorf("Expected 'no services' error, got: %v", err)
	}
}

func TestHTMLBuilder_Chaining(t *testing.T) {
	yaml := []byte(`
version: "1.0"
info:
  title: "Test"
  version: "1.0.0"
services:
  test:
    info:
      title: "Test Service"
      version: "1.0.0"
    metrics:
      test_metric:
        namespace: test
        subsystem: app
        type: counter
        help: "Test"
`)

	spec, err := docs.LoadSpecFromBytes(yaml)
	if err != nil {
		t.Fatalf("Failed to load spec: %v", err)
	}

	// Test method chaining
	html, err := docs.NewHTMLBuilder("Platform", "1.0.0").
		SetDescription("Test description").
		AddFromSpec(spec).
		BuildHTML()

	if err != nil {
		t.Fatalf("BuildHTML failed: %v", err)
	}

	if len(html) == 0 {
		t.Fatal("Expected non-empty HTML")
	}
}
