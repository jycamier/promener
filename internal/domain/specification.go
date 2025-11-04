package domain

import "fmt"

// Specification represents the complete metrics specification (like OpenAPI spec)
type Specification struct {
	Version  string             `yaml:"version"`
	Info     Info               `yaml:"info"`
	Services map[string]Service `yaml:"services"`
}

// Info contains metadata about the metrics specification
type Info struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version"`
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

	if len(s.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}

	// Validate each service
	for serviceName, service := range s.Services {
		if service.Info.Title == "" {
			return fmt.Errorf("service %s: info.title is required", serviceName)
		}
		if service.Info.Version == "" {
			return fmt.Errorf("service %s: info.version is required", serviceName)
		}
		if len(service.Metrics) == 0 {
			return fmt.Errorf("service %s: at least one metric is required", serviceName)
		}

		// Validate metrics in service
		for name, metric := range service.Metrics {
			if metric.Name == "" {
				metric.Name = name
			}
			if err := metric.Validate(); err != nil {
				return fmt.Errorf("service %s: invalid metric %s: %w", serviceName, name, err)
			}
		}
	}

	return nil
}
