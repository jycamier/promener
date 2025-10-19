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

	// Enrich all metrics with Go-specific fields
	for i := range data.Namespaces {
		for j := range data.Namespaces[i].Subsystems {
			for k := range data.Namespaces[i].Subsystems[j].Metrics {
				metric := &data.Namespaces[i].Subsystems[j].Metrics[k]
				switch domain.MetricType(metric.Type) {
				case domain.MetricTypeCounter:
					metric.VecType = "CounterVec"
					metric.OptsType = "CounterOpts"
					metric.Constructor = "prometheus.NewCounterVec"
				case domain.MetricTypeGauge:
					metric.VecType = "GaugeVec"
					metric.OptsType = "GaugeOpts"
					metric.Constructor = "prometheus.NewGaugeVec"
				case domain.MetricTypeHistogram:
					metric.VecType = "HistogramVec"
					metric.OptsType = "HistogramOpts"
					metric.Constructor = "prometheus.NewHistogramVec"
				case domain.MetricTypeSummary:
					metric.VecType = "SummaryVec"
					metric.OptsType = "SummaryOpts"
					metric.Constructor = "prometheus.NewSummaryVec"
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
			}
		}
	}

	return data
}
