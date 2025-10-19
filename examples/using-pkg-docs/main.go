package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jycamier/promener/pkg/docs"
)

func main() {
	// Example 1: Simplest usage - from file to file
	fmt.Println("Example 1: Generate HTML from YAML file")
	err := docs.GenerateHTMLFromFile(
		"../../testdata/metrics_with_docs.yaml",
		"output-simple.html",
	)
	if err != nil {
		log.Fatalf("Example 1 failed: %v", err)
	}
	fmt.Println("✓ Generated: output-simple.html")

	// Example 2: Load spec, modify it, then generate
	fmt.Println("\nExample 2: Load and inspect spec before generating")
	spec, err := docs.LoadSpec("../../testdata/metrics_with_docs.yaml")
	if err != nil {
		log.Fatalf("Example 2 failed: %v", err)
	}
	fmt.Printf("✓ Loaded spec: %s v%s\n", spec.Info.Title, spec.Info.Version)
	fmt.Printf("  - Services count: %d\n", len(spec.Services))

	totalMetrics := 0
	for _, svc := range spec.Services {
		totalMetrics += len(svc.Metrics)
	}
	fmt.Printf("  - Total metrics count: %d\n", totalMetrics)

	html, err := docs.GenerateHTML(spec)
	if err != nil {
		log.Fatalf("Example 2 failed: %v", err)
	}
	fmt.Printf("✓ Generated HTML: %d bytes\n", len(html))

	// Example 3: Generate from YAML bytes (useful for embedded specs)
	fmt.Println("\nExample 3: Generate from YAML bytes")
	yamlData := []byte(`
version: "1.0"
info:
  title: "Example Metrics"
  description: "Generated from embedded YAML"
  version: "1.0.0"
services:
  example:
    info:
      title: "Example Service"
      version: "1.0.0"
      package: "example"
    metrics:
      example_counter:
        namespace: app
        subsystem: example
        type: counter
        help: "An example counter metric"
`)

	html, err = docs.GenerateHTMLFromBytes(yamlData)
	if err != nil {
		log.Fatalf("Example 3 failed: %v", err)
	}

	err = os.WriteFile("output-embedded.html", html, 0644)
	if err != nil {
		log.Fatalf("Example 3 failed to write file: %v", err)
	}
	fmt.Println("✓ Generated: output-embedded.html")

	// Example 4: Reuse generator for multiple specs
	fmt.Println("\nExample 4: Reuse generator for better performance")
	generator, err := docs.NewHTMLGenerator()
	if err != nil {
		log.Fatalf("Example 4 failed: %v", err)
	}

	specs := []string{
		"../../testdata/metrics_with_docs.yaml",
		"../../examples/ecommerce-platform.yaml",
	}

	for i, specPath := range specs {
		spec, err := docs.LoadSpec(specPath)
		if err != nil {
			log.Printf("  ✗ Failed to load %s: %v", specPath, err)
			continue
		}

		outputPath := fmt.Sprintf("output-batch-%d.html", i+1)
		err = generator.GenerateFile(spec, outputPath)
		if err != nil {
			log.Printf("  ✗ Failed to generate %s: %v", outputPath, err)
			continue
		}
		fmt.Printf("  ✓ Generated: %s from %s\n", outputPath, specPath)
	}

	// Example 5: HTTP server serving dynamic documentation
	fmt.Println("\nExample 5: HTTP server (starting on :8080)")
	fmt.Println("  Visit: http://localhost:8080/docs")
	fmt.Println("  Press Ctrl+C to stop")

	http.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		html, err := docs.GenerateHTMLFromFileToBytes("../../testdata/metrics_with_docs.yaml")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate docs: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Promener Documentation Server</title></head>
<body>
	<h1>Promener Documentation Server</h1>
	<p>Available endpoints:</p>
	<ul>
		<li><a href="/docs">/docs</a> - Metrics documentation</li>
	</ul>
</body>
</html>
`)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
