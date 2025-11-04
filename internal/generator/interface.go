package generator

//go:generate mockgen -source=interface.go -destination=mocks/mock_generator.go -package=mocks

import "github.com/jycamier/promener/internal/domain"

// MetricsGenerator is the interface for generating metrics code in any language
type MetricsGenerator interface {
	// GenerateMetrics generates the main metrics code from a specification
	GenerateMetrics(spec *domain.Specification) error
}

// DIGenerator is the interface for generating dependency injection code
type DIGenerator interface {
	// GenerateDI generates dependency injection code (e.g., FX module for Go, DI extensions for .NET)
	GenerateDI(spec *domain.Specification) error
}
