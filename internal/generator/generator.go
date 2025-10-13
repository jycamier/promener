package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"

	"github.com/jycamier/promener/internal/domain"
)

// Generator handles code generation for Prometheus metrics
type Generator struct {
	tmpl *template.Template
}

// New creates a new Generator instance
func New() (*Generator, error) {
	tmpl, err := template.New("metrics").Funcs(template.FuncMap{
		"toGoCode": func(ev EnvVarValue) string {
			return ev.ToGoCode()
		},
	}).Parse(registryTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &Generator{tmpl: tmpl}, nil
}

// Generate generates Go code from a specification
func (g *Generator) Generate(spec *domain.Specification) ([]byte, error) {
	var buf bytes.Buffer

	// Build template data organized by namespace/subsystem
	data := buildTemplateData(spec)

	if err := g.tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to format generated code: %w", err)
	}

	return formatted, nil
}

// GenerateFile generates Go code and writes it to a file
func (g *Generator) GenerateFile(spec *domain.Specification, outputPath string) error {
	code, err := g.Generate(spec)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, code, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GenerateFxFile generates an FX module file
func (g *Generator) GenerateFxFile(spec *domain.Specification, outputPath string) error {
	var buf bytes.Buffer

	// Parse FX template
	fxTmpl, err := template.New("fx").Parse(fxTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse FX template: %w", err)
	}

	// Build template data
	data := buildTemplateData(spec)

	// Execute template
	if err := fxTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute FX template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format FX code: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write FX file: %w", err)
	}

	return nil
}

