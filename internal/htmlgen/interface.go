package htmlgen

//go:generate mockgen -source=interface.go -destination=mocks/mock_htmlgen.go -package=mocks

import "github.com/jycamier/promener/internal/domain"

// HTMLGenerator is the interface for generating HTML documentation from specifications.
// This interface is useful for mocking in tests.
type HTMLGenerator interface {
	// Generate generates HTML documentation from a specification.
	Generate(spec *domain.Specification) ([]byte, error)

	// GenerateFile generates HTML and writes to a file.
	GenerateFile(spec *domain.Specification, outputPath string) error
}

// HTMLBuilder is the interface for building multi-service HTML documentation.
// This interface is useful for mocking in tests.
type HTMLBuilder interface {
	// AddFromSpec merges all services from the given specification into the builder.
	AddFromSpec(spec *domain.Specification) *Builder

	// Build generates the HTML documentation and writes it to a file.
	Build(outputPath string) error
}

// Ensure the concrete types implement their interfaces.
var (
	_ HTMLGenerator = (*Generator)(nil)
	_ HTMLBuilder   = (*Builder)(nil)
)
