package generator

import (
	"fmt"
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

	// Enrich all metrics with .NET-specific fields
	for i := range data.Namespaces {
		for j := range data.Namespaces[i].Subsystems {
			for k := range data.Namespaces[i].Subsystems[j].Metrics {
				metric := &data.Namespaces[i].Subsystems[j].Metrics[k]

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
				var constLabelVars []string
				for key := range metric.ConstLabels {
					constLabelVars = append(constLabelVars, key)
				}
				metric.DotNetConstLabelArgs = strings.Join(constLabelVars, ", ")
			}
		}
	}

	return data
}
