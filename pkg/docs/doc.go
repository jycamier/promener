// Package docs provides functionality for generating documentation from Prometheus metrics specifications.
//
// This package allows you to load YAML specifications and generate HTML documentation
// without using the CLI. It's designed to be used as a library in other Go applications.
//
// # Quick Start
//
// The simplest way to generate HTML documentation from a YAML file:
//
//	err := docs.GenerateHTMLFromFile("metrics.yaml", "output.html")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Loading and Generating Separately
//
// If you need more control, you can load the spec and generate HTML separately:
//
//	// Load specification
//	spec, err := docs.LoadSpec("metrics.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Generate HTML
//	html, err := docs.GenerateHTML(spec)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Do something with the HTML bytes
//	fmt.Println(string(html))
//
// # Reusing the Generator
//
// If you're generating multiple HTML documents, reuse the generator for better performance:
//
//	generator, err := docs.NewHTMLGenerator()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, specFile := range specFiles {
//	    spec, err := docs.LoadSpec(specFile)
//	    if err != nil {
//	        log.Printf("failed to load %s: %v", specFile, err)
//	        continue
//	    }
//
//	    err = generator.GenerateFile(spec, "output.html")
//	    if err != nil {
//	        log.Printf("failed to generate HTML for %s: %v", specFile, err)
//	    }
//	}
//
// # Working with Bytes
//
// You can also work directly with YAML bytes instead of files:
//
//	yamlData := []byte(`
//	version: "1.0"
//	info:
//	  title: "My Metrics"
//	  version: "1.0.0"
//	  package: "metrics"
//	metrics:
//	  requests_total:
//	    namespace: http
//	    subsystem: server
//	    type: counter
//	    help: "Total HTTP requests"
//	`)
//
//	html, err := docs.GenerateHTMLFromBytes(yamlData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # HTTP Handler Example
//
// Integrate into a web server to serve documentation dynamically:
//
//	http.HandleFunc("/metrics-docs", func(w http.ResponseWriter, r *http.Request) {
//	    html, err := docs.GenerateHTMLFromFile("metrics.yaml")
//	    if err != nil {
//	        http.Error(w, err.Error(), http.StatusInternalServerError)
//	        return
//	    }
//	    w.Header().Set("Content-Type", "text/html; charset=utf-8")
//	    w.Write(html)
//	})
package docs
