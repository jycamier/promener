package generator

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jycamier/promener/internal/domain"
)

//go:embed templates/go/*.gotmpl
var templatesFS embed.FS

// GolangGenerator generates Go code for Prometheus or OpenTelemetry metrics
type GolangGenerator struct {
	generator *Generator
	provider  ProviderType
}

// Ensure GolangGenerator implements MetricsGenerator and DIGenerator
var (
	_ MetricsGenerator = (*GolangGenerator)(nil)
	_ DIGenerator      = (*GolangGenerator)(nil)
)

func NewGolangGenerator(packageName string, outputPath string, provider ProviderType) (*GolangGenerator, error) {
	builder := NewGoTemplateDataBuilder()
	generator, err := NewGenerator(templatesFS, "templates/go/*.gotmpl", builder, GoEnvTransformer, packageName, outputPath)
	if err != nil {
		return nil, err
	}
	return &GolangGenerator{
		generator: generator,
		provider:  provider,
	}, nil
}

func (g *GolangGenerator) GenerateMetrics(spec *domain.Specification) error {
	// Generate interface file
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "interface.gotmpl", "metrics_interface.go")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated interface:", filepath.Join(g.generator.outputPath, "metrics_interface.go"))

	// Generate validation file (common to all providers)
	err = g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "validation.gotmpl", "metrics_validation.go")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated validation:", filepath.Join(g.generator.outputPath, "metrics_validation.go"))

	// Generate provider-specific implementation
	var templateFile, outputFile string
	switch g.provider {
	case ProviderPrometheus:
		templateFile = "metrics_prometheus.gotmpl"
		outputFile = "metrics_prometheus.go"
	case ProviderOtel:
		templateFile = "metrics_otel.gotmpl"
		outputFile = "metrics_otel.go"
	default:
		return fmt.Errorf("unsupported provider: %s", g.provider)
	}

	err = g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, templateFile, outputFile)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Generated %s implementation: %s\n", g.provider, filepath.Join(g.generator.outputPath, outputFile))

	// Generate backend file
	var backendTemplateFile, backendOutputFile string
	switch g.provider {
	case ProviderPrometheus:
		backendTemplateFile = "backend_prometheus.gotmpl"
		backendOutputFile = "metrics_backend_prometheus.go"
	case ProviderOtel:
		backendTemplateFile = "backend_otel.gotmpl"
		backendOutputFile = "metrics_backend_otel.go"
	default:
		return fmt.Errorf("unsupported provider: %s", g.provider)
	}

	err = g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, backendTemplateFile, backendOutputFile)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Generated %s backend: %s\n", g.provider, filepath.Join(g.generator.outputPath, backendOutputFile))

	return nil
}

func (g *GolangGenerator) GenerateDI(spec *domain.Specification) error {
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "di_fx.gotmpl", "metrics_fx.go")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated DI:", filepath.Join(g.generator.outputPath, "metrics_fx.go"))

	return nil
}
