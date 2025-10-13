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
func (c *ConstLabels) UnmarshalYAML(value *yaml.Node) error {
	// Try simple map[string]string first (backward compatibility)
	var simpleMap map[string]string
	if err := value.Decode(&simpleMap); err == nil {
		*c = make(ConstLabels, 0, len(simpleMap))
		for name, val := range simpleMap {
			*c = append(*c, ConstLabelDefinition{
				Name:  name,
				Value: val,
			})
		}
		return nil
	}

	// Try detailed format: map[string]{value: string, description: string}
	var detailedMap map[string]struct {
		Value       string `yaml:"value"`
		Description string `yaml:"description"`
	}
	if err := value.Decode(&detailedMap); err == nil {
		*c = make(ConstLabels, 0, len(detailedMap))
		for name, def := range detailedMap {
			*c = append(*c, ConstLabelDefinition{
				Name:        name,
				Value:       def.Value,
				Description: def.Description,
			})
		}
		return nil
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
