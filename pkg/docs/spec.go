package docs

import (
	"fmt"
	"os"

	"github.com/jycamier/promener/internal/domain"
	"gopkg.in/yaml.v3"
)

// Specification represents a Prometheus metrics specification.
// It can be in single-service mode or multi-service mode.
type Specification = domain.Specification

// LoadSpec loads and validates a metrics specification from a YAML file.
func LoadSpec(path string) (*Specification, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	return LoadSpecFromBytes(data)
}

// LoadSpecFromBytes loads and validates a metrics specification from YAML bytes.
func LoadSpecFromBytes(data []byte) (*Specification, error) {
	var spec Specification
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid specification: %w", err)
	}

	return &spec, nil
}
