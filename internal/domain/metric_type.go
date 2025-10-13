package domain

// MetricType represents the type of Prometheus metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// IsValid checks if the metric type is valid
func (t MetricType) IsValid() bool {
	switch t {
	case MetricTypeCounter, MetricTypeGauge, MetricTypeHistogram, MetricTypeSummary:
		return true
	}
	return false
}
