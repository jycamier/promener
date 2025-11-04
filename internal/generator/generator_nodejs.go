package generator

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jycamier/promener/internal/domain"
)

//go:embed templates/nodejs/*.gotmpl
var nodejsTemplatesFS embed.FS

// NodeJSGenerator generates Node.js/TypeScript code for Prometheus metrics
type NodeJSGenerator struct {
	generator *Generator
}

// Ensure NodeJSGenerator implements MetricsGenerator
var _ MetricsGenerator = (*NodeJSGenerator)(nil)

func NewNodeJSGenerator(packageName string, outputPath string) (*NodeJSGenerator, error) {
	builder := NewNodeJSTemplateDataBuilder()
	generator, err := NewGenerator(nodejsTemplatesFS, "templates/nodejs/*.gotmpl", builder, NodeJSEnvTransformer, packageName, outputPath)
	if err != nil {
		return nil, err
	}
	return &NodeJSGenerator{
		generator: generator,
	}, nil
}

func (g *NodeJSGenerator) GenerateMetrics(spec *domain.Specification) error {
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "metrics.gotmpl", "metrics.ts")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated metrics:", filepath.Join(g.generator.outputPath, "metrics.ts"))

	return nil
}
