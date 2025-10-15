package generator

import "github.com/jycamier/promener/internal/domain"

// CodeGenerator is the interface that all language-specific generators must implement
type CodeGenerator interface {
	// Generate generates code from a specification and returns the raw bytes
	Generate(spec *domain.Specification) ([]byte, error)

	// GenerateFile generates code and writes it to a file
	GenerateFile(spec *domain.Specification, outputPath string) error

	// Language returns the language this generator targets
	Language() Language
}
