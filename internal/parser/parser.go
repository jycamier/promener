package parser

import (
	"fmt"
	"os"

	"github.com/jycamier/promener/internal/domain"
	"gopkg.in/yaml.v3"
)

// Parser handles parsing of metric specifications from YAML
type Parser struct{}

// New creates a new Parser instance
func New() *Parser {
	return &Parser{}
}

// ParseFile reads and parses a YAML file into a Specification
func (p *Parser) ParseFile(path string) (*domain.Specification, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses YAML data into a Specification
func (p *Parser) Parse(data []byte) (*domain.Specification, error) {
	var spec domain.Specification
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	for serviceName, service := range spec.Services {
		for key, metric := range service.Metrics {
			if metric.Name == "" {
				metric.Name = key
				service.Metrics[key] = metric
			}
		}
		spec.Services[serviceName] = service
	}

	if err := spec.Validate(); err != nil {
		return nil, fmt.Errorf("invalid specification: %w", err)
	}

	return &spec, nil
}
