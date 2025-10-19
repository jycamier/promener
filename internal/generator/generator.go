package generator

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jycamier/promener/internal/domain"
)

type Generator struct {
	tmpl        *template.Template
	builder     TemplateDataBuilder
	packageName string
	outputPath  string
}

func NewGenerator(fs embed.FS, pattern string, builder TemplateDataBuilder, envTransformer EnvTransformer, packageName string, outputPath string) (*Generator, error) {
	tmpl, err := template.
		New("default").
		Funcs(template.FuncMap{
			"toCode": envTransformer,
			"toLower": func(s string) string {
				return strings.ToLower(s)
			},
		}).ParseFS(fs, pattern)
	if err != nil {
		return nil, err
	}
	return &Generator{
		tmpl:        tmpl,
		builder:     builder,
		packageName: packageName,
		outputPath:  outputPath,
	}, nil
}

func (g *Generator) GenerateFileFromTemplate(spec *domain.Specification, packageName string, templateName string, fileName string) error {
	var buf bytes.Buffer
	err := g.tmpl.ExecuteTemplate(&buf, templateName, g.builder.BuildTemplateData(spec, packageName))
	if err != nil {
		return err
	}
	file := filepath.Join(g.outputPath, fileName)
	if err := os.WriteFile(file, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
