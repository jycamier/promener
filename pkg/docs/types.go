package docs

import "github.com/jycamier/promener/internal/domain"

// Specification represents a Prometheus metrics specification.
// It can contain multiple services, each with its own metrics.
type Specification = domain.Specification

// Service represents a microservice with its own metrics.
type Service = domain.Service

// Metric represents a single Prometheus metric definition.
type Metric = domain.Metric

// Info contains metadata about a specification or service.
type Info = domain.Info

// Server represents a server URL where metrics are exposed.
type Server = domain.Server

// Labels represents a collection of metric label definitions.
type Labels = domain.Labels

// LabelDefinition represents a single metric label with optional description.
type LabelDefinition = domain.LabelDefinition

// ConstLabels represents a collection of constant label definitions.
type ConstLabels = domain.ConstLabels

// ConstLabelDefinition represents a constant label with a fixed value.
type ConstLabelDefinition = domain.ConstLabelDefinition

// MetricType represents the type of Prometheus metric.
type MetricType = domain.MetricType

// Metric type constants
const (
	MetricTypeCounter   = domain.MetricTypeCounter
	MetricTypeGauge     = domain.MetricTypeGauge
	MetricTypeHistogram = domain.MetricTypeHistogram
	MetricTypeSummary   = domain.MetricTypeSummary
)

// Components contains reusable components (like OpenAPI components).
type Components = domain.Components

// Examples contains example queries and alerts for a metric.
type Examples = domain.Examples

// PromQLExample represents a PromQL query example.
type PromQLExample = domain.PromQLExample

// AlertExample represents an alert rule example.
type AlertExample = domain.AlertExample

// Deprecated contains deprecation information for a metric.
type Deprecated = domain.Deprecated
