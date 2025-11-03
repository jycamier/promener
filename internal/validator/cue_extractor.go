package validator

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/jycamier/promener/internal/domain"
	"gopkg.in/yaml.v3"
)

// CueExtractor extracts domain.Specification from validated CUE values.
type CueExtractor struct{}

// NewCueExtractor creates a new CUE extractor.
func NewCueExtractor() *CueExtractor {
	return &CueExtractor{}
}

// Extract converts a CUE value into a domain.Specification.
// The CUE value should already be validated before calling this method.
func (e *CueExtractor) Extract(value cue.Value) (*domain.Specification, error) {
	// Convert CUE to JSON first, which properly serializes the structure
	jsonBytes, err := value.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CUE to JSON: %w", err)
	}

	// Convert JSON to a generic interface to preserve the structure
	var intermediate interface{}
	if err := json.Unmarshal(jsonBytes, &intermediate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Convert back to YAML bytes (YAML and JSON are compatible)
	yamlBytes, err := yaml.Marshal(intermediate)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Unmarshal using YAML, which will trigger our custom UnmarshalYAML for Labels
	var spec domain.Specification
	if err := yaml.Unmarshal(yamlBytes, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML into Specification: %w", err)
	}

	// Enrich metrics with names from map keys
	for serviceName, service := range spec.Services {
		for key, metric := range service.Metrics {
			if metric.Name == "" {
				metric.Name = key
				service.Metrics[key] = metric
			}
		}
		spec.Services[serviceName] = service
	}

	// Perform domain validation
	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	return &spec, nil
}
