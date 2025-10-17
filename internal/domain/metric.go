package domain

import (
	"fmt"
	"regexp"
)

var metricNameRegex = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

// Deprecated represents deprecation information for a metric
type Deprecated struct {
	Since      string `yaml:"since,omitempty"`
	ReplacedBy string `yaml:"replacedBy,omitempty"`
	Reason     string `yaml:"reason,omitempty"`
}

// Metric represents a single Prometheus metric definition
type Metric struct {
	Name        string              `yaml:"name,omitempty"`
	Namespace   string              `yaml:"namespace"`
	Subsystem   string              `yaml:"subsystem"`
	Type        MetricType          `yaml:"type"`
	Help        string              `yaml:"help"`
	Labels      Labels              `yaml:"labels,omitempty"`
	Buckets     []float64           `yaml:"buckets,omitempty"`
	Objectives  map[float64]float64 `yaml:"objectives,omitempty"`
	ConstLabels ConstLabels         `yaml:"constLabels,omitempty"`
	Examples    Examples            `yaml:"examples,omitempty"`
	Deprecated  *Deprecated         `yaml:"deprecated,omitempty"`
}

// GetLabelNames returns just the label names as a string slice for backward compatibility
func (m *Metric) GetLabelNames() []string {
	return m.Labels.ToStringSlice()
}

// FullName returns the complete metric name: namespace_subsystem_name
func (m *Metric) FullName() string {
	parts := []string{}
	if m.Namespace != "" {
		parts = append(parts, m.Namespace)
	}
	if m.Subsystem != "" {
		parts = append(parts, m.Subsystem)
	}
	if m.Name != "" {
		parts = append(parts, m.Name)
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "_"
		}
		result += part
	}
	return result
}

// Validate checks if the metric definition is valid
func (m *Metric) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("metric name is required")
	}

	if !metricNameRegex.MatchString(m.Name) {
		return fmt.Errorf("invalid metric name: %s (must match [a-zA-Z_:][a-zA-Z0-9_:]*)", m.Name)
	}

	if m.Namespace == "" {
		return fmt.Errorf("metric namespace is required")
	}

	if m.Subsystem == "" {
		return fmt.Errorf("metric subsystem is required")
	}

	if !m.Type.IsValid() {
		return fmt.Errorf("invalid metric type: %s", m.Type)
	}

	if m.Help == "" {
		return fmt.Errorf("metric help is required")
	}

	// Validate labels
	for _, label := range m.Labels {
		if !metricNameRegex.MatchString(label.Name) {
			return fmt.Errorf("invalid label name: %s", label.Name)
		}
	}

	// Validate const labels
	for _, label := range m.ConstLabels {
		if !metricNameRegex.MatchString(label.Name) {
			return fmt.Errorf("invalid const label name: %s", label.Name)
		}
	}

	// Type-specific validation
	if m.Type == MetricTypeHistogram && len(m.Buckets) == 0 {
		return fmt.Errorf("histogram metrics require buckets")
	}

	return nil
}
