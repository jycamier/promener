package generator

import (
	"fmt"
	"strings"

	"github.com/jycamier/promener/internal/domain"
)

// GoTemplateDataBuilder wraps CommonTemplateDataBuilder with Go-specific logic
type GoTemplateDataBuilder struct {
	common *CommonTemplateDataBuilder
}

// NewGoTemplateDataBuilder creates a new Go-specific builder
func NewGoTemplateDataBuilder() *GoTemplateDataBuilder {
	return &GoTemplateDataBuilder{
		common: NewCommonTemplateDataBuilder(),
	}
}

// BuildTemplateData builds template data with Go-specific enrichment
func (b *GoTemplateDataBuilder) BuildTemplateData(spec *domain.Specification, packageName string) *TemplateData {
	data := b.common.BuildTemplateData(spec, packageName)

	// Enrich all metrics with Go-specific fields using the common helper
	_ = b.common.EnrichMetrics(data, func(metric *MetricData) error {
		// Check if metric has labels
		metric.HasLabels = len(metric.Labels) > 0

		// Set default values (without Vec)
		switch domain.MetricType(metric.Type) {
		case domain.MetricTypeCounter:
			metric.SimpleType = "Counter"
			metric.OptsType = "CounterOpts"
			metric.VecType = "Counter"
			metric.Constructor = "prometheus.NewCounter"
		case domain.MetricTypeGauge:
			metric.SimpleType = "Gauge"
			metric.OptsType = "GaugeOpts"
			metric.VecType = "Gauge"
			metric.Constructor = "prometheus.NewGauge"
		case domain.MetricTypeHistogram:
			metric.SimpleType = "Histogram"
			metric.OptsType = "HistogramOpts"
			metric.VecType = "Histogram"
			metric.Constructor = "prometheus.NewHistogram"
		case domain.MetricTypeSummary:
			metric.SimpleType = "Summary"
			metric.OptsType = "SummaryOpts"
			metric.VecType = "Summary"
			metric.Constructor = "prometheus.NewSummary"
		}

		// Override with Vec types if metric has labels
		if metric.HasLabels {
			metric.VecType = metric.SimpleType + "Vec"
			metric.Constructor = "prometheus.New" + metric.SimpleType + "Vec"
		}

		// Build method parameters and arguments
		var params []string
		var args []string
		for _, label := range metric.Labels {
			paramName := escapeGoKeyword(toLowerCamelCase(label))
			params = append(params, fmt.Sprintf("%s string", paramName))
			args = append(args, paramName)
		}
		metric.MethodParams = strings.Join(params, ", ")
		metric.MethodArgs = strings.Join(args, ", ")

		return nil
	})

	return data
}
