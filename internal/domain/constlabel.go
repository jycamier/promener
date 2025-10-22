package domain

import "gopkg.in/yaml.v3"

// ConstLabelDefinition represents a constant label with its value and optional description
type ConstLabelDefinition struct {
	Name        string
	Value       string
	Description string
}

// ConstLabels is a slice of constant label definitions
type ConstLabels []ConstLabelDefinition

// UnmarshalYAML implements custom YAML unmarshaling for ConstLabels
// Supports both simple map[string]string and detailed map with descriptions
// Uses yaml.Node to preserve key order from YAML file
func (c *ConstLabels) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return nil
	}

	*c = make(ConstLabels, 0, len(value.Content)/2)

	for i := 0; i < len(value.Content); i += 2 {
		keyNode := value.Content[i]
		valueNode := value.Content[i+1]

		name := keyNode.Value

		// Try simple string value first
		if valueNode.Kind == yaml.ScalarNode {
			*c = append(*c, ConstLabelDefinition{
				Name:  name,
				Value: valueNode.Value,
			})
			continue
		}

		// Try detailed format with value and description
		var detail struct {
			Value       string `yaml:"value"`
			Description string `yaml:"description"`
		}
		if err := valueNode.Decode(&detail); err == nil {
			*c = append(*c, ConstLabelDefinition{
				Name:        name,
				Value:       detail.Value,
				Description: detail.Description,
			})
		}
	}

	return nil
}

// ToMap returns a simple map[string]string for code generation (backward compatibility)
func (c ConstLabels) ToMap() map[string]string {
	result := make(map[string]string, len(c))
	for _, label := range c {
		result[label.Name] = label.Value
	}
	return result
}
