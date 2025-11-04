package htmlgen

import (
	"fmt"
	"os"

	"github.com/jycamier/promener/internal/domain"
)

// Builder helps build multi-service HTML documentation by aggregating
// multiple specifications into a single HTML output.
type Builder struct {
	spec *domain.Specification
	gen  *Generator
}

// NewBuilder creates a new HTML builder with the given title and version.
// The builder starts with an empty set of services.
func NewBuilder(title, version string) *Builder {
	gen, _ := New() // Error is impossible since template is embedded
	return &Builder{
		spec: &domain.Specification{
			Info: domain.Info{
				Title:   title,
				Version: version,
			},
			Services: make(map[string]domain.Service),
		},
		gen: gen,
	}
}

// AddFromSpec merges all services from the given specification into the builder.
// If a service with the same name already exists, it will be overwritten.
// This allows aggregating multiple specification files into a single multi-service documentation.
func (b *Builder) AddFromSpec(spec *domain.Specification) *Builder {
	if spec == nil {
		return b
	}

	for serviceName, service := range spec.Services {
		b.spec.Services[serviceName] = service
	}

	return b
}

// Build generates the HTML documentation and writes it to a file.
func (b *Builder) Build(outputPath string) error {
	if len(b.spec.Services) == 0 {
		return fmt.Errorf("no services added to builder")
	}

	html, err := b.gen.Generate(b.spec)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// NewGenerator creates a new HTML generator that can be used directly
// for single-specification HTML generation.
func NewGenerator() *Generator {
	gen, _ := New() // Error is impossible since template is embedded
	return gen
}
