package generator

import (
	"strings"

	"github.com/jycamier/promener/internal/domain"
)

// Go reserved keywords
var goReservedKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
}

// TemplateData contains all data for template generation
type TemplateData struct {
	PackageName     string
	Info            domain.Info
	Namespaces      []Namespace
	NeedsOsImport   bool
	NeedsHelperFunc bool
}

// Namespace represents a metric namespace
type Namespace struct {
	Name       string
	Subsystems []Subsystem
}

// Subsystem represents a metric subsystem
type Subsystem struct {
	Name    string
	Metrics []MetricData
}

// MetricData contains all information needed to generate a metric
type MetricData struct {
	Name                 string
	Namespace            string
	Subsystem            string
	Type                 string
	Help                 string
	Labels               []string
	LabelDefinitions     []domain.LabelDefinition // Full label definitions with validations
	Buckets              []float64
	Objectives           map[float64]float64
	ConstLabels          map[string]EnvVarValue
	ConstLabelKeys       []string // Sorted keys for consistent iteration
	FieldName            string
	MethodName           string
	MethodParams         string
	MethodArgs           string
	DotNetMethodParams   string
	DotNetMethodArgs     string
	DotNetConstLabelArgs string
	NodeJSMethodParams   string
	NodeJSMethodArgs     string
	NodeJSConstLabelArgs string
	NodeJSType           string
	FullName             string
	VecType              string
	OptsType             string
	Constructor          string
	HasLabels            bool   // true if the metric has labels (uses Vec types)
	SimpleType           string // The simple type without Vec (Counter, Gauge, etc.)
	Deprecated           *domain.Deprecated
}

// toCamelCase converts a snake_case string to CamelCase
func toCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, "")
}

// toLowerCamelCase converts a snake_case string to camelCase
func toLowerCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, "")
}

// escapeGoKeyword adds underscore suffix if the identifier is a Go reserved keyword
func escapeGoKeyword(s string) string {
	if goReservedKeywords[s] {
		return s + "_"
	}
	return s
}

