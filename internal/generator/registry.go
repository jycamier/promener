package generator

import (
	"fmt"

	"github.com/jycamier/promener/internal/domain"
)

// generatorRegistry holds all registered code generators
var generatorRegistry = make(map[Language]func() (CodeGenerator, error))

// RegisterGenerator registers a generator factory for a specific language
func RegisterGenerator(lang Language, factory func() (CodeGenerator, error)) {
	generatorRegistry[lang] = factory
}

// NewGeneratorForLanguage creates a new generator for the specified language
func NewGeneratorForLanguage(lang Language) (CodeGenerator, error) {
	factory, ok := generatorRegistry[lang]
	if !ok {
		return nil, fmt.Errorf("no generator registered for language: %s", lang)
	}
	return factory()
}

// GenerateForLanguage is a convenience function to generate code for a specific language
func GenerateForLanguage(spec *domain.Specification, lang Language, outputPath string) error {
	gen, err := NewGeneratorForLanguage(lang)
	if err != nil {
		return err
	}
	return gen.GenerateFile(spec, outputPath)
}

func init() {
	// Register Go generator by default
	RegisterGenerator(LanguageGo, func() (CodeGenerator, error) {
		return NewGoGenerator()
	})
}
