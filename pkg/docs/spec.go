package docs

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

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

// LoadSpecFromURL loads and validates a metrics specification from a URL.
// The URL should return YAML content.
func LoadSpecFromURL(url string) (*Specification, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	return LoadSpecFromReader(resp.Body)
}

// LoadSpecFromReader loads and validates a metrics specification from an io.Reader.
// The reader should provide YAML content.
func LoadSpecFromReader(r io.Reader) (*Specification, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return LoadSpecFromBytes(data)
}
