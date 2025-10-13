package generator

import (
	"regexp"
)

var envVarRegex = regexp.MustCompile(`^\$\{([^:}]+)(?::([^}]*))?\}$`)

// EnvVarValue represents a constant label value that can be a literal or environment variable
type EnvVarValue struct {
	IsEnvVar     bool
	EnvVar       string
	DefaultValue string
	LiteralValue string
}

// ParseEnvVarValue parses a string that can be either:
// - A literal value: "production"
// - An env var: "${REGION}"
// - An env var with default: "${REGION:eu-west-1}"
func ParseEnvVarValue(value string) EnvVarValue {
	matches := envVarRegex.FindStringSubmatch(value)
	if matches == nil {
		// It's a literal value
		return EnvVarValue{
			IsEnvVar:     false,
			LiteralValue: value,
		}
	}

	// It's an environment variable
	return EnvVarValue{
		IsEnvVar:     true,
		EnvVar:       matches[1],
		DefaultValue: matches[2], // Empty string if no default provided
	}
}

// ToGoCode generates the Go code to get the value
func (e EnvVarValue) ToGoCode() string {
	if !e.IsEnvVar {
		return `"` + e.LiteralValue + `"`
	}

	if e.DefaultValue != "" {
		return `getEnvOrDefault("` + e.EnvVar + `", "` + e.DefaultValue + `")`
	}

	return `os.Getenv("` + e.EnvVar + `")`
}

// NeedsOsImport returns true if this value requires the os package
func (e EnvVarValue) NeedsOsImport() bool {
	return e.IsEnvVar
}

// NeedsHelperFunction returns true if this value needs the getEnvOrDefault helper
func (e EnvVarValue) NeedsHelperFunction() bool {
	return e.IsEnvVar && e.DefaultValue != ""
}

// ParseConstLabelsMap parses all const labels from a map and returns their parsed values
func ParseConstLabelsMap(constLabels map[string]string) map[string]EnvVarValue {
	result := make(map[string]EnvVarValue)
	for key, value := range constLabels {
		result[key] = ParseEnvVarValue(value)
	}
	return result
}

// HasEnvVarsMap checks if any const label uses environment variables
func HasEnvVarsMap(constLabels map[string]string) bool {
	for _, value := range constLabels {
		if ParseEnvVarValue(value).IsEnvVar {
			return true
		}
	}
	return false
}

// NeedsHelperFuncMap checks if any const label needs the helper function
func NeedsHelperFuncMap(constLabels map[string]string) bool {
	for _, value := range constLabels {
		if ParseEnvVarValue(value).NeedsHelperFunction() {
			return true
		}
	}
	return false
}
