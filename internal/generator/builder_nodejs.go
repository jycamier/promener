package generator

import (
	"fmt"
	"strings"

	"github.com/jycamier/promener/internal/domain"
)

// NodeJSTemplateDataBuilder wraps CommonTemplateDataBuilder with Node.js-specific logic
type NodeJSTemplateDataBuilder struct {
	common *CommonTemplateDataBuilder
}

// NewNodeJSTemplateDataBuilder creates a new Node.js-specific builder
func NewNodeJSTemplateDataBuilder() *NodeJSTemplateDataBuilder {
	return &NodeJSTemplateDataBuilder{
		common: NewCommonTemplateDataBuilder(),
	}
}

// BuildTemplateData builds template data with Node.js-specific enrichment
func (b *NodeJSTemplateDataBuilder) BuildTemplateData(spec *domain.Specification, packageName string) *TemplateData {
	data := b.common.BuildTemplateData(spec, packageName)

	// Enrich all metrics with Node.js-specific fields
	for i := range data.Namespaces {
		for j := range data.Namespaces[i].Subsystems {
			for k := range data.Namespaces[i].Subsystems[j].Metrics {
				metric := &data.Namespaces[i].Subsystems[j].Metrics[k]

				// Set NodeJSType for prom-client
				switch metric.Type {
				case "counter":
					metric.NodeJSType = "Counter"
				case "gauge":
					metric.NodeJSType = "Gauge"
				case "histogram":
					metric.NodeJSType = "Histogram"
				case "summary":
					metric.NodeJSType = "Summary"
				}

				// Build method parameters for dynamic labels
				var params []string
				var args []string
				for _, label := range metric.Labels {
					paramName := toLowerCamelCase(label)
					params = append(params, fmt.Sprintf("%s: string", paramName))
					args = append(args, paramName)
				}
				metric.NodeJSMethodParams = strings.Join(params, ", ")
				metric.NodeJSMethodArgs = strings.Join(args, ", ")

				// Build const label variable names for label object construction
				// (ConstLabelKeys is already sorted by CommonTemplateDataBuilder)
				metric.NodeJSConstLabelArgs = strings.Join(metric.ConstLabelKeys, ", ")
			}
		}
	}

	return data
}
