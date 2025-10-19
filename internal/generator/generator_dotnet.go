package generator

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jycamier/promener/internal/domain"
)

//go:embed templates/dotnet/*.gotmpl
var dotnetTemplatesFS embed.FS

type DotNetGenerator struct {
	generator *Generator
}

func NewDotNetGenerator(packageName string, outputPath string) (*DotNetGenerator, error) {
	builder := NewDotNetTemplateDataBuilder()
	generator, err := NewGenerator(dotnetTemplatesFS, "templates/dotnet/*.gotmpl", builder, DotNetEnvTransformer, packageName, outputPath)
	if err != nil {
		return nil, err
	}
	return &DotNetGenerator{
		generator: generator,
	}, nil
}

func (g *DotNetGenerator) GenerateMetrics(spec *domain.Specification) error {
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "metrics.gotmpl", "Metrics.cs")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated metrics:", filepath.Join(g.generator.outputPath, "Metrics.cs"))

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
