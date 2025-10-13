package domain

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LabelDefinition represents a label with optional description
type LabelDefinition struct {
	Name        string `yaml:"name,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Labels can be either a simple array of strings or a map with descriptions
type Labels []LabelDefinition

// UnmarshalYAML implements custom YAML unmarshaling to support both formats:
// - Simple: ["method", "status"]
// - Detailed: { method: { description: "HTTP method" }, status: { description: "Status code" } }
func (l *Labels) UnmarshalYAML(value *yaml.Node) error {
	// Try to unmarshal as simple string array first
	var simpleLabels []string
	if err := value.Decode(&simpleLabels); err == nil {
		*l = make(Labels, len(simpleLabels))
		for i, name := range simpleLabels {
			(*l)[i] = LabelDefinition{Name: name}
		}
		return nil
	}

	// Try to unmarshal as map[string]LabelDetail
	var detailedLabels map[string]struct {
		Description string `yaml:"description"`
	}
	if err := value.Decode(&detailedLabels); err == nil {
		*l = make(Labels, 0, len(detailedLabels))
		for name, detail := range detailedLabels {
			*l = append(*l, LabelDefinition{
				Name:        name,
				Description: detail.Description,
			})
		}
		return nil
	}

	return fmt.Errorf("labels must be either an array of strings or a map with descriptions")
}

// ToStringSlice returns just the label names as a string slice
func (l Labels) ToStringSlice() []string {
	result := make([]string, len(l))
	for i, label := range l {
		result[i] = label.Name
	}
	return result
}
