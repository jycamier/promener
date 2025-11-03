package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"github.com/jycamier/promener/internal/domain"
	"github.com/jycamier/promener/internal/parser"
)

// CueValidator implements SchemaValidator using CUE for schema validation.
type CueValidator struct {
	ctx *cue.Context
}

// NewCueValidator creates a new CUE-based schema validator.
func NewCueValidator() *CueValidator {
	return &CueValidator{
		ctx: cuecontext.New(),
	}
}

// ValidateFile validates a YAML file against a CUE schema file.
func (v *CueValidator) ValidateFile(yamlPath, schemaPath string) (*ValidationResult, error) {
	// Parse the YAML file
	p := parser.New()
	spec, err := p.ParseFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML file: %w", err)
	}

	return v.Validate(spec, schemaPath)
}

// Validate validates a parsed Specification against a CUE schema file.
func (v *CueValidator) Validate(spec *domain.Specification, schemaPath string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:        true,
		CueErrors:    []ValidationError{},
		DomainErrors: []ValidationError{},
	}

	// First, perform domain validation
	if err := spec.Validate(); err != nil {
		result.Valid = false
		result.DomainErrors = append(result.DomainErrors, ValidationError{
			Path:    "",
			Message: err.Error(),
			Source:  "domain",
			Line:    0,
		})
	}

	// Load the CUE schema
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CUE schema file: %w", err)
	}

	schemaValue := v.ctx.CompileBytes(schemaContent)
	if schemaValue.Err() != nil {
		return nil, fmt.Errorf("failed to compile CUE schema: %w", schemaValue.Err())
	}

	// Convert the specification to JSON-compatible format, then to CUE
	// Note: We need to handle map[float64]float64 (Objectives) which JSON doesn't support
	specJSON, err := marshalSpecToJSON(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal specification to JSON: %w", err)
	}

	specValue := v.ctx.CompileBytes(specJSON)
	if specValue.Err() != nil {
		return nil, fmt.Errorf("failed to compile specification as CUE: %w", specValue.Err())
	}

	// Unify the specification with the schema
	unified := schemaValue.Unify(specValue)

	// Check for validation errors
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		result.Valid = false
		cueErrors := v.extractCueErrors(err)
		result.CueErrors = append(result.CueErrors, cueErrors...)
	}

	return result, nil
}

// extractCueErrors converts CUE errors into ValidationError structs.
func (v *CueValidator) extractCueErrors(err error) []ValidationError {
	var validationErrors []ValidationError

	// CUE errors can be a list of errors
	for _, e := range errors.Errors(err) {
		pos := e.Position()
		path := v.extractPath(e)

		validationErrors = append(validationErrors, ValidationError{
			Path:    path,
			Message: strings.TrimSpace(e.Error()),
			Source:  "cue",
			Line:    pos.Line(),
		})
	}

	return validationErrors
}

// extractPath attempts to extract the field path from a CUE error.
func (v *CueValidator) extractPath(err errors.Error) string {
	// Try to get the path from the error position
	pos := err.Position()
	if pos.IsValid() {
		// The error message often contains the path
		msg := err.Error()
		// Extract path pattern like "services.default.metrics.test_test"
		if idx := strings.Index(msg, ":"); idx > 0 {
			pathPart := msg[:idx]
			pathPart = strings.TrimSpace(pathPart)
			// Remove line/column info if present
			if spaceIdx := strings.Index(pathPart, " "); spaceIdx > 0 {
				pathPart = pathPart[:spaceIdx]
			}
			return pathPart
		}
	}
	return ""
}

// marshalSpecToJSON marshals a specification to JSON, handling the map[float64]float64 issue.
// The Objectives field in metrics uses map[float64]float64 which JSON doesn't support
// (JSON only allows string keys). We convert it to map[string]float64 for JSON serialization.
func marshalSpecToJSON(spec *domain.Specification) ([]byte, error) {
	// Create a deep copy of the spec to avoid modifying the original
	type jsonMetric struct {
		Name        string                 `json:"name,omitempty"`
		Namespace   string                 `json:"namespace"`
		Subsystem   string                 `json:"subsystem"`
		Type        domain.MetricType      `json:"type"`
		Help        string                 `json:"help"`
		Labels      interface{}            `json:"labels,omitempty"`
		Buckets     []float64              `json:"buckets,omitempty"`
		Objectives  map[string]float64     `json:"objectives,omitempty"` // Changed from map[float64]float64
		ConstLabels interface{}            `json:"constLabels,omitempty"`
		Examples    interface{}            `json:"examples,omitempty"`
		Deprecated  interface{}            `json:"deprecated,omitempty"`
	}

	type jsonService struct {
		Info    domain.Info            `json:"info"`
		Servers interface{}            `json:"servers,omitempty"`
		Metrics map[string]jsonMetric  `json:"metrics"`
	}

	type jsonSpec struct {
		Version  string                   `json:"version"`
		Info     domain.Info              `json:"info"`
		Services map[string]jsonService   `json:"services"`
	}

	// Convert the spec
	jsonServices := make(map[string]jsonService)
	for serviceName, service := range spec.Services {
		jsonMetrics := make(map[string]jsonMetric)
		for metricName, metric := range service.Metrics {
			jm := jsonMetric{
				Name:        metric.Name,
				Namespace:   metric.Namespace,
				Subsystem:   metric.Subsystem,
				Type:        metric.Type,
				Help:        metric.Help,
				Labels:      metric.Labels,
				Buckets:     metric.Buckets,
				ConstLabels: metric.ConstLabels,
				Examples:    metric.Examples,
				Deprecated:  metric.Deprecated,
			}

			// Convert map[float64]float64 to map[string]float64
			if len(metric.Objectives) > 0 {
				jm.Objectives = make(map[string]float64)
				for k, v := range metric.Objectives {
					jm.Objectives[fmt.Sprintf("%v", k)] = v
				}
			}

			jsonMetrics[metricName] = jm
		}

		jsonServices[serviceName] = jsonService{
			Info:    service.Info,
			Servers: service.Servers,
			Metrics: jsonMetrics,
		}
	}

	jsonSpecObj := jsonSpec{
		Version:  spec.Version,
		Info:     spec.Info,
		Services: jsonServices,
	}

	return json.Marshal(jsonSpecObj)
}
