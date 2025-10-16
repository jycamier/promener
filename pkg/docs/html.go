package docs

import (
	"fmt"

	"github.com/jycamier/promener/internal/htmlgen"
)

// HTMLGenerator generates HTML documentation from metrics specifications.
type HTMLGenerator struct {
	generator *htmlgen.Generator
}

// NewHTMLGenerator creates a new HTML documentation generator.
func NewHTMLGenerator() (*HTMLGenerator, error) {
	gen, err := htmlgen.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create HTML generator: %w", err)
	}

	return &HTMLGenerator{generator: gen}, nil
}

// Generate generates HTML documentation from a specification.
// Returns the HTML content as bytes.
func (g *HTMLGenerator) Generate(spec *Specification) ([]byte, error) {
	return g.generator.Generate(spec)
}

// GenerateFile generates HTML documentation and writes it to a file.
func (g *HTMLGenerator) GenerateFile(spec *Specification, outputPath string) error {
	return g.generator.GenerateFile(spec, outputPath)
}

// GenerateHTML is a convenience function that creates a generator and generates HTML in one call.
// Use this if you don't need to reuse the generator.
func GenerateHTML(spec *Specification) ([]byte, error) {
	gen, err := NewHTMLGenerator()
	if err != nil {
		return nil, err
	}
	return gen.Generate(spec)
}

// GenerateHTMLFile is a convenience function that creates a generator and writes HTML to a file in one call.
// Use this if you don't need to reuse the generator.
func GenerateHTMLFile(spec *Specification, outputPath string) error {
	gen, err := NewHTMLGenerator()
	if err != nil {
		return err
	}
	return gen.GenerateFile(spec, outputPath)
}

// GenerateHTMLFromFile loads a spec from a YAML file and generates HTML documentation.
// This is the simplest way to generate HTML: just provide input and output paths.
func GenerateHTMLFromFile(specPath, outputPath string) error {
	spec, err := LoadSpec(specPath)
	if err != nil {
		return err
	}

	return GenerateHTMLFile(spec, outputPath)
}

// GenerateHTMLFromFileToBytes loads a spec from a YAML file and returns HTML as bytes.
func GenerateHTMLFromFileToBytes(specPath string) ([]byte, error) {
	spec, err := LoadSpec(specPath)
	if err != nil {
		return nil, err
	}

	return GenerateHTML(spec)
}

// GenerateHTMLFromBytes loads a spec from YAML bytes and generates HTML.
func GenerateHTMLFromBytes(yamlData []byte) ([]byte, error) {
	spec, err := LoadSpecFromBytes(yamlData)
	if err != nil {
		return nil, err
	}

	return GenerateHTML(spec)
}
