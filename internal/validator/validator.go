package validator

import "github.com/jycamier/promener/internal/domain"

// SchemaValidator validates Promener specifications against CUE schemas.
type SchemaValidator interface {
	// ValidateFile validates a YAML file against a CUE schema file.
	// Returns a ValidationResult containing any validation errors found.
	ValidateFile(yamlPath, schemaPath string) (*ValidationResult, error)

	// Validate validates a parsed Specification against a CUE schema file.
	// Returns a ValidationResult containing any validation errors found.
	Validate(spec *domain.Specification, schemaPath string) (*ValidationResult, error)
}

// ValidationResult contains the combined results of domain and CUE validation.
type ValidationResult struct {
	// Valid indicates whether all validations passed.
	Valid bool

	// CueErrors contains errors found during CUE schema validation.
	CueErrors []ValidationError

	// DomainErrors contains errors found during domain validation.
	DomainErrors []ValidationError
}

// ValidationError represents a single validation error with context.
type ValidationError struct {
	// Path is the JSON path to the field that failed validation (e.g., "services.default.metrics.requests_total").
	Path string

	// Message is the human-readable error message.
	Message string

	// Source indicates where the error came from ("cue" or "domain").
	Source string

	// Line is the line number in the source file (if available).
	Line int
}

// HasErrors returns true if there are any validation errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.CueErrors) > 0 || len(r.DomainErrors) > 0
}

// TotalErrors returns the total number of validation errors.
func (r *ValidationResult) TotalErrors() int {
	return len(r.CueErrors) + len(r.DomainErrors)
}
