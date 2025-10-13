package domain

// Examples contains documentation examples for a metric
type Examples struct {
	PromQL []PromQLExample `yaml:"promql,omitempty"`
	Alerts []AlertExample  `yaml:"alerts,omitempty"`
}

// PromQLExample represents a PromQL query example
type PromQLExample struct {
	Query       string `yaml:"query"`
	Description string `yaml:"description,omitempty"`
}

// AlertExample represents an Alertmanager alert rule example
type AlertExample struct {
	Name        string `yaml:"name"`
	Expr        string `yaml:"expr"`
	Description string `yaml:"description,omitempty"`
	For         string `yaml:"for,omitempty"`
	Severity    string `yaml:"severity,omitempty"`
}
