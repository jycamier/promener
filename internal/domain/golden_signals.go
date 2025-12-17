package domain

// RecordingRule represents a Prometheus recording rule
type RecordingRule struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

// Thresholds defines threshold values for dashboard visualization
type Thresholds struct {
	Good     string `yaml:"good"`
	Warning  string `yaml:"warning,omitempty"`
	Critical string `yaml:"critical"`
}

// GoldenSignal represents one of the four golden signals
type GoldenSignal struct {
	Description    string          `yaml:"description"`
	Metrics        []string        `yaml:"metrics"`
	RecordingRules []RecordingRule `yaml:"recordingRules,omitempty"`
	Thresholds     *Thresholds     `yaml:"thresholds,omitempty"`
}

// GoldenSignals groups the four golden signals
type GoldenSignals struct {
	Latency    *GoldenSignal `yaml:"latency,omitempty"`
	Errors     *GoldenSignal `yaml:"errors,omitempty"`
	Traffic    *GoldenSignal `yaml:"traffic,omitempty"`
	Saturation *GoldenSignal `yaml:"saturation,omitempty"`
}
