package generator

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jycamier/promener/internal/domain"
)

//go:embed templates/nodejs/*.gotmpl
var nodejsTemplatesFS embed.FS

// NodeJSGenerator generates Node.js/TypeScript code for Prometheus or OpenTelemetry metrics
type NodeJSGenerator struct {
	generator *Generator
	provider  ProviderType
}

// Ensure NodeJSGenerator implements MetricsGenerator
var _ MetricsGenerator = (*NodeJSGenerator)(nil)

func NewNodeJSGenerator(packageName string, outputPath string, provider ProviderType) (*NodeJSGenerator, error) {
	builder := NewNodeJSTemplateDataBuilder()
	generator, err := NewGenerator(nodejsTemplatesFS, "templates/nodejs/*.gotmpl", builder, NodeJSEnvTransformer, packageName, outputPath)
	if err != nil {
		return nil, err
	}
	return &NodeJSGenerator{
		generator: generator,
		provider:  provider,
	}, nil
}

func (g *NodeJSGenerator) GenerateMetrics(spec *domain.Specification) error {
	// Generate interface file
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "interface.gotmpl", "metrics_interface.ts")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated interface:", filepath.Join(g.generator.outputPath, "metrics_interface.ts"))

	// Generate provider-specific implementation
	var templateFile, outputFile string
	switch g.provider {
	case ProviderPrometheus:
		templateFile = "metrics_prometheus.gotmpl"
		outputFile = "metrics.ts"
	case ProviderOtel:
		templateFile = "metrics_otel.gotmpl"
		outputFile = "metrics_otel.ts"
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
