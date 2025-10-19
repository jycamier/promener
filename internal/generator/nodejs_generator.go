package generator

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/jycamier/promener/internal/domain"
)

// NodeJSGenerator handles code generation for Node.js/TypeScript
type NodeJSGenerator struct {
	tmpl *template.Template
}

// NewNodeJSGenerator creates a new Node.js generator instance
func NewNodeJSGenerator() (*NodeJSGenerator, error) {
	tmpl, err := template.New("nodejs").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).Parse(nodejsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Node.js template: %w", err)
	}

	return &NodeJSGenerator{tmpl: tmpl}, nil
}

// Generate generates TypeScript code from a specification
func (g *NodeJSGenerator) Generate(spec *domain.Specification) ([]byte, error) {
	var buf bytes.Buffer

	// Build template data organized by namespace/subsystem
	data := buildNodeJSTemplateData(spec)

	if err := g.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateFile generates TypeScript code and writes it to a file
func (g *NodeJSGenerator) GenerateFile(spec *domain.Specification, outputPath string) error {
	code, err := g.Generate(spec)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, code, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// buildNodeJSTemplateData builds template data with Node.js-specific types
func buildNodeJSTemplateData(spec *domain.Specification) *TemplateData {
	data := buildTemplateData(spec, "")

	// Update NodeJSType for prom-client
	for i := range data.Namespaces {
		for j := range data.Namespaces[i].Subsystems {
			for k := range data.Namespaces[i].Subsystems[j].Metrics {
				metric := &data.Namespaces[i].Subsystems[j].Metrics[k]
				switch metric.Type {
				case "counter":
					metric.NodeJSType = "Counter"
				case "gauge":
					metric.NodeJSType = "Gauge"
				case "histogram":
					metric.NodeJSType = "Histogram"
				case "summary":
					metric.NodeJSType = "Summary"
				}
			}
		}
	}

	return data
}
