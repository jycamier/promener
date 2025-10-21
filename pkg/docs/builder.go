package docs

import (
	"fmt"
	"os"
)

// Builder is the interface for building multi-service HTML documentation.
type Builder interface {
	// SetDescription sets the description for the overall documentation.
	SetDescription(description string) Builder

	// AddFromSpec merges all services from the given specification into the builder.
	AddFromSpec(spec *Specification) Builder

	// AddService adds a single service with the given name to the builder.
	AddService(name string, service Service) Builder

	// BuildHTML generates the final HTML documentation from all aggregated services.
	BuildHTML() ([]byte, error)

	// BuildHTMLFile generates the HTML documentation and writes it to a file.
	BuildHTMLFile(outputPath string) error

	// GetSpecification returns the underlying specification being built.
	GetSpecification() *Specification
}

// HTMLBuilder helps build multi-service HTML documentation by aggregating
// multiple specifications into a single HTML output.
type HTMLBuilder struct {
	spec *Specification
}

// Ensure HTMLBuilder implements Builder interface.
var _ Builder = (*HTMLBuilder)(nil)

// NewHTMLBuilder creates a new HTML builder with the given title and version.
// The builder starts with an empty set of services.
func NewHTMLBuilder(title, version string) *HTMLBuilder {
	return &HTMLBuilder{
		spec: &Specification{
			Version: "1.0",
			Info: Info{
				Title:   title,
				Version: version,
			},
			Services: make(map[string]Service),
		},
	}
}

// SetDescription sets the description for the overall documentation.
func (b *HTMLBuilder) SetDescription(description string) Builder {
	b.spec.Info.Description = description
	return b
}

// AddFromSpec merges all services from the given specification into the builder.
// If a service with the same name already exists, it will be overwritten.
// This allows aggregating multiple specification files into a single multi-service documentation.
func (b *HTMLBuilder) AddFromSpec(spec *Specification) Builder {
	if spec == nil {
		return b
	}

	for serviceName, service := range spec.Services {
		b.spec.Services[serviceName] = service
	}

	return b
}

// AddService adds a single service with the given name to the builder.
// If a service with the same name already exists, it will be overwritten.
func (b *HTMLBuilder) AddService(name string, service Service) Builder {
	b.spec.Services[name] = service
	return b
}

// BuildHTML generates the final HTML documentation from all aggregated services.
// It validates the specification before generating HTML.
func (b *HTMLBuilder) BuildHTML() ([]byte, error) {
	if len(b.spec.Services) == 0 {
		return nil, fmt.Errorf("no services added to builder")
	}

	if err := b.spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid specification: %w", err)
	}

	return GenerateHTML(b.spec)
}

// BuildHTMLFile generates the HTML documentation and writes it to a file.
func (b *HTMLBuilder) BuildHTMLFile(outputPath string) error {
	html, err := b.BuildHTML()
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, html, 0644); err != nil {
		return fmt.Errorf("failed to write HTML file: %w", err)
	}

	return nil
}

// GetSpecification returns the underlying specification being built.
// This can be useful for inspection or further manipulation.
func (b *HTMLBuilder) GetSpecification() *Specification {
	return b.spec
}
