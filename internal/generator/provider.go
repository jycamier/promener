package generator

import "fmt"

// ProviderType represents the metrics provider type
type ProviderType string

const (
	ProviderPrometheus ProviderType = "prometheus"
	ProviderOtel       ProviderType = "otel"
)

func (p *ProviderType) String() string {
	return string(*p)
}

func (p *ProviderType) Set(v string) error {
	switch v {
	case "prometheus", "otel":
		*p = ProviderType(v)
		return nil
	default:
		return fmt.Errorf("must be 'prometheus' or 'otel', got: %s", v)
	}
}

func (p *ProviderType) Type() string {
	return "provider"
}
