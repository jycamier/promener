package generator

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/jycamier/promener/internal/domain"
)

// DotNetGenerator handles code generation for .NET/C#
type DotNetGenerator struct {
	tmpl   *template.Template
	diTmpl *template.Template
}

// NewDotNetGenerator creates a new .NET generator instance
func NewDotNetGenerator() (*DotNetGenerator, error) {
	tmpl, err := template.New("dotnet").Parse(dotnetTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .NET template: %w", err)
	}

	diTmpl, err := template.New("dotnet-di").Parse(dotnetDITemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .NET DI template: %w", err)
	}

	return &DotNetGenerator{
		tmpl:   tmpl,
		diTmpl: diTmpl,
	}, nil
}

// Language returns the language this generator targets
func (g *DotNetGenerator) Language() Language {
	return LanguageDotNet
}

// Generate generates C# code from a specification
func (g *DotNetGenerator) Generate(spec *domain.Specification) ([]byte, error) {
	var buf bytes.Buffer

	// Build template data organized by namespace/subsystem
	data := buildDotNetTemplateData(spec)

	if err := g.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// GenerateFile generates C# code and writes it to a file
func (g *DotNetGenerator) GenerateFile(spec *domain.Specification, outputPath string) error {
	code, err := g.Generate(spec)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, code, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GenerateDIFile generates a dependency injection extension file
func (g *DotNetGenerator) GenerateDIFile(spec *domain.Specification, outputPath string) error {
	var buf bytes.Buffer

	// Build template data
	data := buildDotNetTemplateData(spec)

	// Execute template
	if err := g.diTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute DI template: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write DI file: %w", err)
	}

	return nil
}

// buildDotNetTemplateData builds template data with .NET-specific VecTypes
func buildDotNetTemplateData(spec *domain.Specification) *TemplateData {
	data := buildTemplateData(spec)

	// Update VecType for .NET (prometheus-net uses different names)
	for i := range data.Namespaces {
		for j := range data.Namespaces[i].Subsystems {
			for k := range data.Namespaces[i].Subsystems[j].Metrics {
				metric := &data.Namespaces[i].Subsystems[j].Metrics[k]
				switch metric.Type {
				case "counter":
					metric.VecType = "Counter"
				case "gauge":
					metric.VecType = "Gauge"
				case "histogram":
					metric.VecType = "Histogram"
				case "summary":
					metric.VecType = "Summary"
				}
			}
		}
	}

	return data
}

func init() {
	// Register .NET generator
	RegisterGenerator(LanguageDotNet, func() (CodeGenerator, error) {
		return NewDotNetGenerator()
	})
}
