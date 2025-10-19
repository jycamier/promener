package generator

import "github.com/jycamier/promener/internal/domain"

// TemplateDataBuilder transforms a domain specification into template data
type TemplateDataBuilder interface {
	BuildTemplateData(spec *domain.Specification, packageName string) *TemplateData
}

// CommonTemplateDataBuilder handles the common logic for building template data
type CommonTemplateDataBuilder struct{}

// NewCommonTemplateDataBuilder creates a new common builder
func NewCommonTemplateDataBuilder() *CommonTemplateDataBuilder {
	return &CommonTemplateDataBuilder{}
}

// BuildTemplateData creates the base template data with namespace/subsystem organization
func (b *CommonTemplateDataBuilder) BuildTemplateData(spec *domain.Specification, packageName string) *TemplateData {
	nsMap := make(map[string]map[string][]MetricData)

	// Group metrics by namespace and subsystem
	for _, service := range spec.Services {
		for key, metric := range service.Metrics {
			if metric.Name == "" {
				metric.Name = key
			}

			ns := toCamelCase(metric.Namespace)
			ss := toCamelCase(metric.Subsystem)

			if nsMap[ns] == nil {
				nsMap[ns] = make(map[string][]MetricData)
			}

			nsMap[ns][ss] = append(nsMap[ns][ss], MetricData{
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
			})
		}
	}

	// Build namespaces structure
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
	for _, service := range spec.Services {
		for _, metric := range service.Metrics {
			constLabelsMap := metric.ConstLabels.ToMap()
			if HasEnvVarsMap(constLabelsMap) {
				needsOs = true
			}
			if NeedsHelperFuncMap(constLabelsMap) {
				needsHelper = true
			}
		}
	}

	return &TemplateData{
		PackageName:     packageName,
		Info:            spec.Info,
		Namespaces:      namespaces,
		NeedsOsImport:   needsOs,
		NeedsHelperFunc: needsHelper,
	}
}
