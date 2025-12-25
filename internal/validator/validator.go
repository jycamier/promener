package validator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jycamier/promener/internal/domain"
)

// Validator validates and extracts Promener specifications from CUE files.
type Validator struct {
	loader    *CueLoader
	extractor *CueExtractor
	rego      *RegoValidator
}

// New creates a new Validator instance.
func New() *Validator {
	return &Validator{
		loader:    NewCueLoader(),
		extractor: NewCueExtractor(),
	}
}

// SetRulesDirs sets the directories containing Rego rules for validation.
func (v *Validator) SetRulesDirs(dirs []string) {
	if len(dirs) > 0 {
		v.rego = NewRegoValidator(dirs)
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
	if result.HasErrors() {
		return nil, result, fmt.Errorf("CUE validation failed")
	}

	// Post CUE export validation (Rego)
	if v.rego != nil {
		var input interface{}
		// Convert CUE value to JSON-compatible structure
		jsonData, err := json.Marshal(cueValue)
		if err == nil {
			if err := json.Unmarshal(jsonData, &input); err == nil {
				regoErrors, err := v.rego.Validate(context.Background(), input)
				if err != nil {
					return nil, nil, fmt.Errorf("rego validation failed: %w", err)
				}
				if len(regoErrors) > 0 {
					result.RegoErrors = append(result.RegoErrors, regoErrors...)
				}
			}
		}
	}

	// Extract the specification
	spec, err := v.extractor.Extract(cueValue)
	if err != nil {
		// Add domain validation errors to the result
		result.DomainErrors = append(result.DomainErrors, ValidationError{
			Path:     "",
			Message:  err.Error(),
			Source:   "domain",
			Severity: "error",
			Line:     0,
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

// ValidationResult contains the combined results of domain, CUE, and Rego validation.
type ValidationResult struct {
	// CueErrors contains errors found during CUE schema validation.
	CueErrors []ValidationError

	// DomainErrors contains errors found during domain validation.
	DomainErrors []ValidationError

	// RegoErrors contains errors found during Rego policy validation.
	RegoErrors []ValidationError
}

// ValidationError represents a single validation error with context.
type ValidationError struct {
	// Path is the JSON path to the field that failed validation (e.g., "services.default.metrics.requests_total").
	Path string

	// Message is the human-readable error message.
	Message string

	// Source indicates where the error came from ("cue", "domain", or "rego").
	Source string

	// Severity indicates the criticality of the error ("error", "warning", "info").
	Severity string

	// Line is the line number in the source file (if available).
	Line int
}

// HasErrors returns true if there are any validation errors.
func (r *ValidationResult) HasErrors() bool {
	return len(r.CueErrors) > 0 || len(r.DomainErrors) > 0 || len(r.RegoErrors) > 0
}

// Failed returns true if any error matches or exceeds the given severity threshold.
func (r *ValidationResult) Failed(threshold string) bool {
	levels := map[string]int{
		"error":   3,
		"warning": 2,
		"info":    1,
		"":        0,
	}

	thresholdLevel := levels[threshold]
	if thresholdLevel == 0 {
		thresholdLevel = 3 // Default to error
	}

	for _, err := range r.CueErrors {
		if levels[err.Severity] >= thresholdLevel {
			return true
		}
	}
	for _, err := range r.DomainErrors {
		if levels[err.Severity] >= thresholdLevel {
			return true
		}
	}
	for _, err := range r.RegoErrors {
		if levels[err.Severity] >= thresholdLevel {
			return true
		}
	}

	return false
}

// TotalErrors returns the total number of validation errors.
func (r *ValidationResult) TotalErrors() int {
	return len(r.CueErrors) + len(r.DomainErrors) + len(r.RegoErrors)
}
