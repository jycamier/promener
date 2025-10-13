package domain

import "fmt"

// Specification represents the complete metrics specification (like OpenAPI spec)
type Specification struct {
	Version     string            `yaml:"version"`
	Info        Info              `yaml:"info"`
	Metrics     map[string]Metric `yaml:"metrics"`
	Components  Components        `yaml:"components,omitempty"`
}

// Info contains metadata about the metrics specification
type Info struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version"`
	Package     string `yaml:"package"` // Go package name for generated code
}

// Components contains reusable components (like OpenAPI components)
type Components struct {
	Labels map[string][]string `yaml:"labels,omitempty"`
}

// Validate checks if the specification is valid
func (s *Specification) Validate() error {
	if s.Version == "" {
		return fmt.Errorf("specification version is required")
	}

	if s.Info.Title == "" {
		return fmt.Errorf("info.title is required")
	}

	if s.Info.Version == "" {
		return fmt.Errorf("info.version is required")
	}

	if s.Info.Package == "" {
		return fmt.Errorf("info.package is required")
	}

	if len(s.Metrics) == 0 {
		return fmt.Errorf("at least one metric is required")
	}

	// Validate each metric
	for name, metric := range s.Metrics {
		if metric.Name == "" {
			metric.Name = name
		}
		if err := metric.Validate(); err != nil {
			return fmt.Errorf("invalid metric %s: %w", name, err)
		}
	}

	return nil
}
