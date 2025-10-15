package domain

// Server represents a server URL where metrics are exposed
type Server struct {
	URL         string `yaml:"url"`
	Description string `yaml:"description,omitempty"`
}

// Service represents a microservice with its own metrics
type Service struct {
	Info    Info              `yaml:"info"`
	Servers []Server          `yaml:"servers,omitempty"`
	Metrics map[string]Metric `yaml:"metrics"`
}
