package generator

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jycamier/promener/internal/domain"
)

//go:embed templates/dotnet/*.gotmpl
var dotnetTemplatesFS embed.FS

// DotNetGenerator generates .NET/C# code for Prometheus or OpenTelemetry metrics
type DotNetGenerator struct {
	generator *Generator
	provider  ProviderType
}

// Ensure DotNetGenerator implements MetricsGenerator and DIGenerator
var (
	_ MetricsGenerator = (*DotNetGenerator)(nil)
	_ DIGenerator      = (*DotNetGenerator)(nil)
)

func NewDotNetGenerator(packageName string, outputPath string, provider ProviderType) (*DotNetGenerator, error) {
	builder := NewDotNetTemplateDataBuilder()
	generator, err := NewGenerator(dotnetTemplatesFS, "templates/dotnet/*.gotmpl", builder, DotNetEnvTransformer, packageName, outputPath)
	if err != nil {
		return nil, err
	}
	return &DotNetGenerator{
		generator: generator,
		provider:  provider,
	}, nil
}

func (g *DotNetGenerator) GenerateMetrics(spec *domain.Specification) error {
	// Generate interface file
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "interface.gotmpl", "IMetrics.cs")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated interface:", filepath.Join(g.generator.outputPath, "IMetrics.cs"))

	// Generate provider-specific implementation
	var templateFile, outputFile string
	switch g.provider {
	case ProviderPrometheus:
		templateFile = "metrics_prometheus.gotmpl"
		outputFile = "Metrics.cs"
	case ProviderOtel:
		templateFile = "metrics_otel.gotmpl"
		outputFile = "MetricsOtel.cs"
	default:
		return fmt.Errorf("unsupported provider: %s", g.provider)
	}

	err = g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, templateFile, outputFile)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Generated %s metrics: %s\n", g.provider, filepath.Join(g.generator.outputPath, outputFile))

	return nil
}

func (g *DotNetGenerator) GenerateDI(spec *domain.Specification) error {
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "di.gotmpl", "MetricsExtensions.cs")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated DI:", filepath.Join(g.generator.outputPath, "MetricsExtensions.cs"))

	return nil
}
