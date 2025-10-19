package generator

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/jycamier/promener/internal/domain"
)

//go:embed templates/go/*.gotmpl
var templatesFS embed.FS

type GolangGenerator struct {
	generator *Generator
}

func NewGolangGenerator(packageName string, outputPath string) (*GolangGenerator, error) {
	builder := NewGoTemplateDataBuilder()
	generator, err := NewGenerator(templatesFS, "templates/go/*.gotmpl", builder, GoEnvTransformer, packageName, outputPath)
	if err != nil {
		return nil, err
	}
	return &GolangGenerator{
		generator: generator,
	}, nil
}

func (g *GolangGenerator) GenerateMetrics(spec *domain.Specification) error {
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "metrics.gotmpl", "metrics.go")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated metrics:", filepath.Join(g.generator.outputPath, "metrics.go"))

	return nil
}

func (g *GolangGenerator) GenerateDI(spec *domain.Specification) error {
	err := g.generator.GenerateFileFromTemplate(spec, g.generator.packageName, "di_fx.gotmpl", "fx.go")
	if err != nil {
		return err
	}
	fmt.Println("✓ Generated DI:", filepath.Join(g.generator.outputPath, "fx.go"))

	return nil
}
