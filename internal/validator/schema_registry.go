package validator

import (
	"fmt"
	"regexp"
	"strconv"

	v1 "github.com/jycamier/promener/schema/v1"
)

// schemaRegistry maps major versions to embedded CUE schemas.
var schemaRegistry = map[int]string{
	1: v1.Schema,
}

// versionRegex extracts the major version from a version string like "1.0.0" or "1.2".
var versionRegex = regexp.MustCompile(`^(\d+)\.`)

// GetSchemaForVersion returns the embedded CUE schema for a given version string.
// It extracts the major version (e.g., "1.0.0" → 1) and looks up the corresponding schema.
// Returns an error if the version format is invalid or unsupported.
func GetSchemaForVersion(version string) (string, error) {
	if version == "" {
		return "", fmt.Errorf("version is required in the CUE specification")
	}

	// Extract major version
	matches := versionRegex.FindStringSubmatch(version)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid version format: %s (expected format: X.Y or X.Y.Z)", version)
	}

	majorVersion, err := strconv.Atoi(matches[1])
	if err != nil {
		return "", fmt.Errorf("invalid major version: %s", matches[1])
	}

	// Look up schema
	schema, exists := schemaRegistry[majorVersion]
	if !exists {
		return "", fmt.Errorf("unsupported schema version: v%d (supported: %v)", majorVersion, supportedVersions())
	}

	return schema, nil
}

// supportedVersions returns a list of supported major versions.
func supportedVersions() []int {
	versions := make([]int, 0, len(schemaRegistry))
	for v := range schemaRegistry {
		versions = append(versions, v)
	}
	return versions
}
