package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jycamier/promener/internal/domain"
)

// DotNetTemplateDataBuilder wraps CommonTemplateDataBuilder with .NET-specific logic
type DotNetTemplateDataBuilder struct {
	common *CommonTemplateDataBuilder
}

// NewDotNetTemplateDataBuilder creates a new .NET-specific builder
func NewDotNetTemplateDataBuilder() *DotNetTemplateDataBuilder {
	return &DotNetTemplateDataBuilder{
		common: NewCommonTemplateDataBuilder(),
	}
}

// BuildTemplateData builds template data with .NET-specific enrichment
func (b *DotNetTemplateDataBuilder) BuildTemplateData(spec *domain.Specification, packageName string) *TemplateData {
	data := b.common.BuildTemplateData(spec, packageName)

	// Enrich all metrics with .NET-specific fields using the common helper
	_ = b.common.EnrichMetrics(data, func(metric *MetricData) error {
		// Set VecType for .NET (prometheus-net uses different names)
		switch metric.Type {
		case "counter":
			metric.VecType = "Counter"
		case "gauge":
			metric.VecType = "Gauge"
		case "histogram":
			metric.VecType = "Histogram"
		case "summary":
			metric.VecType = "Summary"
		}

		var params []string
		var args []string
		for _, label := range metric.Labels {
			paramName := toLowerCamelCase(label)
			params = append(params, fmt.Sprintf("string %s", paramName))
			args = append(args, paramName)
		}
		metric.DotNetMethodParams = strings.Join(params, ", ")
		metric.DotNetMethodArgs = strings.Join(args, ", ")

		// Build const label variable names for WithLabels calls
		// Sort keys alphabetically to ensure consistent order
		var constLabelKeys []string
		for key := range metric.ConstLabels {
			constLabelKeys = append(constLabelKeys, key)
		}
		sort.Strings(constLabelKeys)
		metric.DotNetConstLabelArgs = strings.Join(constLabelKeys, ", ")

		return nil
	})

	return data
}
