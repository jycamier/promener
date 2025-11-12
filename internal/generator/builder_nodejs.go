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

	// Enrich all metrics with Node.js-specific fields using the common helper
	_ = b.common.EnrichMetrics(data, func(metric *MetricData) error {
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

		// Build method parameters for dynamic labels (excluding inherited labels)
		var params []string
		var args []string
		for _, labelDef := range metric.LabelDefinitions {
			if !labelDef.IsInherited() {
				paramName := toLowerCamelCase(labelDef.Name)
				params = append(params, fmt.Sprintf("%s: string", paramName))
				args = append(args, paramName)
			}
		}
		metric.NodeJSMethodParams = strings.Join(params, ", ")
		metric.NodeJSMethodArgs = strings.Join(args, ", ")

		// Build const label variable names for label object construction
		// (ConstLabelKeys is already sorted by CommonTemplateDataBuilder)
		metric.NodeJSConstLabelArgs = strings.Join(metric.ConstLabelKeys, ", ")

		return nil
	})

	return data
}
