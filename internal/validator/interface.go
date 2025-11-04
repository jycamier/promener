package validator

//go:generate mockgen -source=interface.go -destination=mocks/mock_validator.go -package=mocks

import "github.com/jycamier/promener/internal/domain"

// SpecValidator is the interface for validating and extracting CUE specifications.
// This interface is useful for mocking in tests.
type SpecValidator interface {
	// ValidateAndExtract validates a CUE file and extracts it into a domain.Specification.
	// It returns the specification, validation result, and any error encountered.
	ValidateAndExtract(cuePath string) (*domain.Specification, *ValidationResult, error)
}

// Ensure the concrete Validator type implements the interface.
var _ SpecValidator = (*Validator)(nil)
