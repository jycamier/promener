package validator

import (
	"fmt"

	"github.com/jycamier/promener/internal/domain"
)

// Validator validates and extracts Promener specifications from CUE files.
type Validator struct {
	loader    *CueLoader
	extractor *CueExtractor
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{
		loader:    NewCueLoader(),
		extractor: NewCueExtractor(),
	}
}

// ValidateAndExtract loads a CUE file, validates it against the embedded schema,
// and extracts it into a domain.Specification.
// Returns the specification, validation results, and any errors.
func (v *Validator) ValidateAndExtract(cuePath string) (*domain.Specification, *ValidationResult, error) {
	// Load and validate the CUE file
	cueValue, result, err := v.loader.LoadAndValidate(cuePath)
	if err != nil {
		return nil, nil, err
	}

	// If there are CUE validation errors, return early
	if !result.Valid || result.HasErrors() {
		return nil, result, fmt.Errorf("CUE validation failed")
	}

	// Extract the specification
	spec, err := v.extractor.Extract(cueValue)
	if err != nil {
		// Add domain validation errors to the result
		result.Valid = false
		result.DomainErrors = append(result.DomainErrors, ValidationError{
			Path:    "",
			Message: err.Error(),
			Source:  "domain",
			Line:    0,
		})
		return nil, result, err
	}

	return spec, result, nil
}

// Validate validates a CUE file without extracting the specification.
// Useful for the `vet` command which only checks validity.
func (v *Validator) Validate(cuePath string) (*ValidationResult, error) {
	_, result, err := v.ValidateAndExtract(cuePath)
	return result, err
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
