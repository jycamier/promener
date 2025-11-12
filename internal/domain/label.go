package domain

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// LabelDefinition represents a label with optional description and validations
type LabelDefinition struct {
	Name        string   `yaml:"name,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Validations []string `yaml:"validations,omitempty"`
	Inherited   string   `yaml:"inherited,omitempty"` // Documentation for labels added via relabeling
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
	// Use yaml.Node to preserve key order from YAML file
	if value.Kind == yaml.MappingNode {
		*l = make(Labels, 0, len(value.Content)/2)
		for i := 0; i < len(value.Content); i += 2 {
			keyNode := value.Content[i]
			valueNode := value.Content[i+1]

			name := keyNode.Value

			var detail struct {
				Description string   `yaml:"description"`
				Validations []string `yaml:"validations,omitempty"`
				Inherited   string   `yaml:"inherited,omitempty"`
			}
			if err := valueNode.Decode(&detail); err != nil {
				return fmt.Errorf("invalid label definition for %s: %w", name, err)
			}

			*l = append(*l, LabelDefinition{
				Name:        name,
				Description: detail.Description,
				Validations: detail.Validations,
				Inherited:   detail.Inherited,
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

// IsInherited returns true if the label is inherited (added via relabeling)
func (ld LabelDefinition) IsInherited() bool {
	return ld.Inherited != ""
}

// NonInheritedLabels returns only labels that are not inherited
func (l Labels) NonInheritedLabels() Labels {
	result := make(Labels, 0, len(l))
	for _, label := range l {
		if !label.IsInherited() {
			result = append(result, label)
		}
	}
	return result
}

// InheritedLabels returns only labels that are inherited
func (l Labels) InheritedLabels() Labels {
	result := make(Labels, 0, len(l))
	for _, label := range l {
		if label.IsInherited() {
			result = append(result, label)
		}
	}
	return result
}
