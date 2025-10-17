package generator

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/jycamier/promener/internal/domain"
)

// TemplateData contains all data for template generation
type TemplateData struct {
	Info              domain.Info
	Namespaces        []Namespace
	NeedsOsImport     bool
	NeedsHelperFunc   bool
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
	Name                string
	Namespace           string
	Subsystem           string
	Type                string
	Help                string
	Labels              []string
	Buckets             []float64
	Objectives          map[float64]float64
	ConstLabels         map[string]EnvVarValue
	FieldName           string
	MethodName          string
	MethodParams        string
	MethodArgs          string
	DotNetMethodParams  string
	DotNetMethodArgs    string
	NodeJSMethodParams  string
	NodeJSType          string
	FullName            string
	VecType             string
	OptsType            string
	Constructor         string
	Deprecated          *domain.Deprecated
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

// toUpperFirst converts first character to uppercase
func toUpperFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// buildTemplateData organizes metrics by namespace and subsystem
func buildTemplateData(spec *domain.Specification) *TemplateData {
	nsMap := make(map[string]map[string][]MetricData)

	// Group metrics by namespace and subsystem
	for key, metric := range spec.Metrics {
		if metric.Name == "" {
			metric.Name = key
		}

		ns := toCamelCase(metric.Namespace)
		ss := toCamelCase(metric.Subsystem)

		if nsMap[ns] == nil {
			nsMap[ns] = make(map[string][]MetricData)
		}

		metricData := MetricData{
			Name:        metric.Name,
			Namespace:   metric.Namespace,
			Subsystem:   metric.Subsystem,
			Type:        string(metric.Type),
			Help:        metric.Help,
			Labels:      metric.GetLabelNames(),
			Buckets:     metric.Buckets,
			Objectives:  metric.Objectives,
			ConstLabels: ParseConstLabelsMap(metric.ConstLabels.ToMap()),
			FieldName:   toLowerCamelCase(metric.Name),
			MethodName:  toCamelCase(metric.Name),
			FullName:    metric.FullName(),
			Deprecated:  metric.Deprecated,
		}

		// Set VecType, OptsType, and Constructor based on metric type
		switch metric.Type {
		case domain.MetricTypeCounter:
			metricData.VecType = "CounterVec"
			metricData.OptsType = "CounterOpts"
			metricData.Constructor = "prometheus.NewCounterVec"
		case domain.MetricTypeGauge:
			metricData.VecType = "GaugeVec"
			metricData.OptsType = "GaugeOpts"
			metricData.Constructor = "prometheus.NewGaugeVec"
		case domain.MetricTypeHistogram:
			metricData.VecType = "HistogramVec"
			metricData.OptsType = "HistogramOpts"
			metricData.Constructor = "prometheus.NewHistogramVec"
		case domain.MetricTypeSummary:
			metricData.VecType = "SummaryVec"
			metricData.OptsType = "SummaryOpts"
			metricData.Constructor = "prometheus.NewSummaryVec"
		}

		// Build method parameters and arguments for Go
		var params []string
		var args []string
		labelNames := metric.GetLabelNames()
		for _, label := range labelNames {
			paramName := toLowerCamelCase(label)
			params = append(params, fmt.Sprintf("%s string", paramName))
			args = append(args, paramName)
		}
		metricData.MethodParams = strings.Join(params, ", ")
		metricData.MethodArgs = strings.Join(args, ", ")

		// Build method parameters and arguments for .NET
		var dotnetParams []string
		var dotnetArgs []string
		for _, label := range labelNames {
			paramName := toLowerCamelCase(label)
			dotnetParams = append(dotnetParams, fmt.Sprintf("string %s", paramName))
			dotnetArgs = append(dotnetArgs, paramName)
		}
		metricData.DotNetMethodParams = strings.Join(dotnetParams, ", ")
		metricData.DotNetMethodArgs = strings.Join(dotnetArgs, ", ")

		// Build method parameters for Node.js/TypeScript
		var nodejsParams []string
		for _, label := range labelNames {
			paramName := toLowerCamelCase(label)
			nodejsParams = append(nodejsParams, fmt.Sprintf("%s: string", paramName))
		}
		metricData.NodeJSMethodParams = strings.Join(nodejsParams, ", ")

		nsMap[ns][ss] = append(nsMap[ns][ss], metricData)
	}

	// Build template data structure
	var namespaces []Namespace
	for nsName, subsystems := range nsMap {
		var ssList []Subsystem
		for ssName, metrics := range subsystems {
			ssList = append(ssList, Subsystem{
				Name:    ssName,
				Metrics: metrics,
			})
		}
		namespaces = append(namespaces, Namespace{
			Name:       nsName,
			Subsystems: ssList,
		})
	}

	// Check if we need os import and helper function
	needsOs := false
	needsHelper := false
	for _, metric := range spec.Metrics {
		constLabelsMap := metric.ConstLabels.ToMap()
		if HasEnvVarsMap(constLabelsMap) {
			needsOs = true
		}
		if NeedsHelperFuncMap(constLabelsMap) {
			needsHelper = true
		}
	}

	return &TemplateData{
		Info:            spec.Info,
		Namespaces:      namespaces,
		NeedsOsImport:   needsOs,
		NeedsHelperFunc: needsHelper,
	}
}
